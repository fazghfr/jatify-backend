package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"job-tracker/internal/config"
	"job-tracker/internal/dto"
	"job-tracker/internal/entity"
	notion_client "job-tracker/internal/notion"
	"job-tracker/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jomei/notionapi"
	"gorm.io/gorm"
)

// httpDoer allows injecting a custom HTTP client for testing.
type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// notionQuerier is the subset of Notion API operations used by this service.
type notionQuerier interface {
	ListDatabases(ctx context.Context) ([]notionapi.Database, error)
	QueryDatabase(ctx context.Context, databaseID string) ([]notionapi.Page, error)
}

var (
	ErrNotionNotConnected   = errors.New("notion integration not connected")
	ErrNotionDatabaseNotSet = errors.New("notion database not configured")
)

type NotionService interface {
	Connect(userID int) (string, error)
	Callback(code, state string) (*entity.NotionIntegration, error)
	Status(userID int) (*dto.NotionStatusResponse, error)
	Configure(userID int, databaseID string) error
	ListDatabases(userID int) ([]dto.NotionDatabaseItem, error)
	Sync(userID int) (*dto.NotionSyncResult, error)
	Disconnect(userID int) error
	RawPages(userID int) ([]notionapi.Page, error)
}

type notionService struct {
	cfg                 *config.Config
	notionRepo          repository.NotionIntegrationRepository
	appRepo             repository.ApplicationRepository
	jobRepo             repository.JobRepository
	statusRepo          repository.StatusRepository
	historyRepo         repository.StatusHistoryRepository
	httpClient          httpDoer
	notionClientFactory func(accessToken string) notionQuerier
}

func NewNotionService(
	cfg *config.Config,
	notionRepo repository.NotionIntegrationRepository,
	appRepo repository.ApplicationRepository,
	jobRepo repository.JobRepository,
	statusRepo repository.StatusRepository,
	historyRepo repository.StatusHistoryRepository,
) NotionService {
	return &notionService{
		cfg:         cfg,
		notionRepo:  notionRepo,
		appRepo:     appRepo,
		jobRepo:     jobRepo,
		statusRepo:  statusRepo,
		historyRepo: historyRepo,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
		notionClientFactory: func(tok string) notionQuerier {
			return notion_client.New(tok)
		},
	}
}

// Connect builds a state JWT and returns the Notion OAuth authorization URL.
func (s *notionService) Connect(userID int) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"purpose": "notion_oauth",
		"exp":     now.Add(10 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	state, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"https://api.notion.com/v1/oauth/authorize?client_id=%s&response_type=code&owner=user&redirect_uri=%s&state=%s",
		s.cfg.NotionClientID,
		s.cfg.NotionRedirectURI,
		state,
	)
	return url, nil
}

type notionTokenResponse struct {
	AccessToken   string `json:"access_token"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	BotID         string `json:"bot_id"`
	Error         string `json:"error"`
}

// Callback validates the state JWT, exchanges the OAuth code for a token, and upserts the integration.
func (s *notionService) Callback(code, state string) (*entity.NotionIntegration, error) {
	// Validate state JWT
	token, err := jwt.Parse(state, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired state token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid state token claims")
	}
	purpose, _ := claims["purpose"].(string)
	if purpose != "notion_oauth" {
		return nil, errors.New("invalid state token purpose")
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user_id in state token")
	}
	userID := int(userIDFloat)

	// Exchange code for token
	tokenResp, err := s.exchangeCode(code)
	if err != nil {
		return nil, err
	}

	// Upsert integration
	existing, err := s.notionRepo.FindByUserID(userID)
	var integration *entity.NotionIntegration
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		integration = &entity.NotionIntegration{
			UserID: userID,
		}
	} else if err != nil {
		return nil, err
	} else {
		integration = existing
	}

	integration.AccessToken = tokenResp.AccessToken
	integration.WorkspaceID = tokenResp.WorkspaceID
	integration.WorkspaceName = tokenResp.WorkspaceName
	integration.BotID = tokenResp.BotID

	if err := s.notionRepo.Upsert(integration); err != nil {
		return nil, err
	}

	return integration, nil
}

func (s *notionService) exchangeCode(code string) (*notionTokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": s.cfg.NotionRedirectURI,
	})

	req, err := http.NewRequest(http.MethodPost, "https://api.notion.com/v1/oauth/token", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.cfg.NotionClientID, s.cfg.NotionClientSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp notionTokenResponse
	if err := json.Unmarshal(raw, &tokenResp); err != nil {
		return nil, err
	}
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("notion OAuth error: %s", tokenResp.Error)
	}
	return &tokenResp, nil
}

// Status returns the current Notion integration status for the user.
func (s *notionService) Status(userID int) (*dto.NotionStatusResponse, error) {
	integration, err := s.notionRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.NotionStatusResponse{Connected: false}, nil
		}
		return nil, err
	}

	return &dto.NotionStatusResponse{
		Connected:     true,
		WorkspaceID:   integration.WorkspaceID,
		WorkspaceName: integration.WorkspaceName,
		DatabaseID:    integration.DatabaseID,
		LastSyncAt:    integration.LastSyncAt,
	}, nil
}

// Configure sets the Notion database ID for sync.
func (s *notionService) Configure(userID int, databaseID string) error {
	integration, err := s.notionRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotionNotConnected
		}
		return err
	}
	integration.DatabaseID = databaseID
	return s.notionRepo.Upsert(integration)
}

// ListDatabases returns all databases accessible via the integration.
func (s *notionService) ListDatabases(userID int) ([]dto.NotionDatabaseItem, error) {
	integration, err := s.notionRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotionNotConnected
		}
		return nil, err
	}

	nc := s.notionClientFactory(integration.AccessToken)
	dbs, err := nc.ListDatabases(context.Background())
	if err != nil {
		return nil, err
	}

	items := make([]dto.NotionDatabaseItem, 0, len(dbs))
	for _, db := range dbs {
		title := ""
		if len(db.Title) > 0 {
			title = db.Title[0].PlainText
		}
		items = append(items, dto.NotionDatabaseItem{
			ID:    string(db.ID),
			Title: title,
		})
	}
	return items, nil
}

// Sync pulls all pages from the configured Notion database and creates/updates applications.
func (s *notionService) Sync(userID int) (*dto.NotionSyncResult, error) {
	// 1. Load integration
	integration, err := s.notionRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotionNotConnected
		}
		return nil, err
	}

	// 2. Verify database ID is set
	if integration.DatabaseID == "" {
		return nil, ErrNotionDatabaseNotSet
	}

	// 3. Build Notion client
	nc := s.notionClientFactory(integration.AccessToken)

	// 4. Load all statuses into a map (lowercase text → id)
	statuses, err := s.statusRepo.FindAll()
	if err != nil {
		return nil, err
	}
	statusMap := make(map[string]int, len(statuses))
	defaultStatusID := 1
	for _, st := range statuses {
		statusMap[strings.ToLower(st.Text)] = st.ID
		if defaultStatusID == 0 {
			defaultStatusID = st.ID
		}
	}
	if len(statuses) > 0 {
		defaultStatusID = statuses[0].ID
	}

	// 5. Paginate Notion database
	pages, err := nc.QueryDatabase(context.Background(), integration.DatabaseID)
	if err != nil {
		return nil, err
	}

	result := &dto.NotionSyncResult{
		Errors: []string{},
	}

	for _, page := range pages {
		jf, af := notion_client.MapPage(page)

		// Skip if Company or Position is empty
		if jf.Company == "" || jf.Position == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("page %s missing Company or Position", af.NotionPageID))
			continue
		}

		// Find or create job
		job, err := s.jobRepo.FindByCompanyPositionUserID(userID, jf.Company, jf.Position)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				job = &entity.Job{
					UserID:      userID,
					Company:     jf.Company,
					Position:    jf.Position,
					Description: jf.Description,
					UUID:        uuid.New(),
				}
				if createErr := s.jobRepo.Create(job); createErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("page %s: failed to create job: %v", af.NotionPageID, createErr))
					continue
				}
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("page %s: job lookup error: %v", af.NotionPageID, err))
				continue
			}
		}

		// Resolve status ID
		statusID := defaultStatusID
		if af.Status != "" {
			if sid, ok := statusMap[strings.ToLower(af.Status)]; ok {
				statusID = sid
			}
		}

		// Find application by Notion page ID
		existing, err := s.appRepo.FindByNotionPageID(af.NotionPageID)
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new application
			pageID := af.NotionPageID
			app := &entity.Application{
				UserID:       userID,
				JobID:        job.ID,
				ResumeID:     nil,
				Text:         af.CoverLetter,
				StatusID:     statusID,
				NotionPageID: &pageID,
				UUID:         uuid.New(),
			}
			if createErr := s.appRepo.Create(app); createErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("page %s: failed to create application: %v", af.NotionPageID, createErr))
				continue
			}
			if af.ApplicationDate != nil {
				_ = s.appRepo.UpdateTimestamps(app.ID, *af.ApplicationDate)
			}
			_ = s.historyRepo.Create(&entity.StatusHistory{
				ApplicationID: app.ID,
				StatusID:      app.StatusID,
			})
			result.Created++
		} else if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("page %s: application lookup error: %v", af.NotionPageID, err))
			continue
		} else {
			// Diff and update if changed
			changed := false
			if existing.DeletedAt.Valid {
				existing.DeletedAt = gorm.DeletedAt{}
				changed = true
			}
			if existing.StatusID != statusID {
				prevStatusID := existing.StatusID
				existing.StatusID = statusID
				_ = s.historyRepo.Create(&entity.StatusHistory{
					ApplicationID: existing.ID,
					StatusID:      prevStatusID,
				})
				changed = true
			}
			if existing.Text != af.CoverLetter {
				existing.Text = af.CoverLetter
				changed = true
			}
			if existing.JobID != job.ID {
				existing.JobID = job.ID
				changed = true
			}

			if changed {
				if updateErr := s.appRepo.Update(existing); updateErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("page %s: failed to update application: %v", af.NotionPageID, updateErr))
					continue
				}
				if af.ApplicationDate != nil {
					_ = s.appRepo.UpdateTimestamps(existing.ID, *af.ApplicationDate)
				}
				result.Updated++
			} else {
				result.Skipped++
			}
		}
	}

	// 6. Update LastSyncAt
	now := time.Now()
	integration.LastSyncAt = &now
	_ = s.notionRepo.Upsert(integration)

	return result, nil
}

// Disconnect removes the Notion integration for the user.
func (s *notionService) Disconnect(userID int) error {
	return s.notionRepo.DeleteByUserID(userID)
}

// RawPages returns the raw Notion pages from the configured database without any mapping.
func (s *notionService) RawPages(userID int) ([]notionapi.Page, error) {
	integration, err := s.notionRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotionNotConnected
		}
		return nil, err
	}
	if integration.DatabaseID == "" {
		return nil, ErrNotionDatabaseNotSet
	}
	nc := s.notionClientFactory(integration.AccessToken)
	return nc.QueryDatabase(context.Background(), integration.DatabaseID)
}
