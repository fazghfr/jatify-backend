package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"job-tracker/internal/config"
	"job-tracker/internal/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jomei/notionapi"
	"gorm.io/gorm"
)

// ---- mock repos ----

type mockNotionRepo struct {
	integration *entity.NotionIntegration
	findErr     error
	upsertErr   error
	deleteErr   error
	upserted    *entity.NotionIntegration
}

func (m *mockNotionRepo) Upsert(integration *entity.NotionIntegration) error {
	m.upserted = integration
	return m.upsertErr
}
func (m *mockNotionRepo) FindByUserID(_ int) (*entity.NotionIntegration, error) {
	return m.integration, m.findErr
}
func (m *mockNotionRepo) DeleteByUserID(_ int) error { return m.deleteErr }

// mockAppRepo returns a fixed app for FindByNotionPageID; all other methods no-op.
type mockAppRepo struct {
	app       *entity.Application
	findErr   error
	createErr error
	updateErr error
	created   *entity.Application
	updated   *entity.Application
}

func (m *mockAppRepo) Create(app *entity.Application) error {
	m.created = app
	if app.ID == 0 {
		app.ID = 99
	}
	return m.createErr
}
func (m *mockAppRepo) FindAllByUserID(_ int) ([]entity.Application, error) { return nil, nil }
func (m *mockAppRepo) FindPageByUserID(_, _, _ int) ([]entity.Application, int64, error) {
	return nil, 0, nil
}
func (m *mockAppRepo) FindByID(_ int) (*entity.Application, error) { return nil, nil }
func (m *mockAppRepo) FindByNotionPageID(_ string) (*entity.Application, error) {
	return m.app, m.findErr
}
func (m *mockAppRepo) Update(app *entity.Application) error {
	m.updated = app
	return m.updateErr
}
func (m *mockAppRepo) Delete(_ int) error { return nil }

type mockJobRepo struct {
	job       *entity.Job
	findErr   error
	createErr error
	created   *entity.Job
}

func (m *mockJobRepo) Create(job *entity.Job) error {
	m.created = job
	if job.ID == 0 {
		job.ID = 1
	}
	return m.createErr
}
func (m *mockJobRepo) FindAllByUserID(_ int) ([]entity.Job, error) { return nil, nil }
func (m *mockJobRepo) FindByID(_ int) (*entity.Job, error)         { return nil, nil }
func (m *mockJobRepo) FindByCompanyPositionUserID(_ int, _, _ string) (*entity.Job, error) {
	return m.job, m.findErr
}
func (m *mockJobRepo) Update(_ *entity.Job) error { return nil }
func (m *mockJobRepo) Delete(_ int) error         { return nil }

type mockStatusRepo struct {
	statuses []entity.Status
	err      error
}

func (m *mockStatusRepo) FindAll() ([]entity.Status, error) { return m.statuses, m.err }

type mockHistoryRepo struct {
	created []*entity.StatusHistory
	err     error
}

func (m *mockHistoryRepo) Create(h *entity.StatusHistory) error {
	m.created = append(m.created, h)
	return m.err
}

// ---- mock HTTP client ----

type mockHTTP struct {
	statusCode int
	body       string
	err        error
}

func (m *mockHTTP) Do(_ *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.body)),
	}, nil
}

// ---- mock Notion querier ----

type mockNotionQuerier struct {
	pages []notionapi.Page
	dbs   []notionapi.Database
	err   error
}

func (m *mockNotionQuerier) ListDatabases(_ context.Context) ([]notionapi.Database, error) {
	return m.dbs, m.err
}
func (m *mockNotionQuerier) QueryDatabase(_ context.Context, _ string) ([]notionapi.Page, error) {
	return m.pages, m.err
}

// ---- helpers ----

const testSecret = "test-jwt-secret"

func newTestCfg() *config.Config {
	return &config.Config{
		JWTSecret:          testSecret,
		NotionClientID:     "client-id",
		NotionClientSecret: "client-secret",
		NotionRedirectURI:  "http://localhost/callback",
	}
}

func makeStateJWT(t *testing.T, userID int, purpose string, expiry time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"purpose": purpose,
		"exp":     expiry.Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("makeStateJWT: %v", err)
	}
	return s
}

func newSvc(
	notionRepo *mockNotionRepo,
	appRepo *mockAppRepo,
	jobRepo *mockJobRepo,
	statusRepo *mockStatusRepo,
	historyRepo *mockHistoryRepo,
	httpCl *mockHTTP,
	querier *mockNotionQuerier,
) *notionService {
	return &notionService{
		cfg:         newTestCfg(),
		notionRepo:  notionRepo,
		appRepo:     appRepo,
		jobRepo:     jobRepo,
		statusRepo:  statusRepo,
		historyRepo: historyRepo,
		httpClient:  httpCl,
		notionClientFactory: func(_ string) notionQuerier {
			return querier
		},
	}
}

// ---- Connect ----

func TestConnect_BuildsCorrectURL(t *testing.T) {
	svc := newSvc(&mockNotionRepo{}, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})

	url, err := svc.Connect(42)
	if err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	if !strings.Contains(url, "client_id=client-id") {
		t.Errorf("URL missing client_id: %s", url)
	}
	if !strings.Contains(url, "redirect_uri=") {
		t.Errorf("URL missing redirect_uri: %s", url)
	}
	if !strings.Contains(url, "state=") {
		t.Errorf("URL missing state: %s", url)
	}

	// Extract and validate state JWT
	parts := strings.SplitN(url, "state=", 2)
	if len(parts) < 2 {
		t.Fatal("no state param in URL")
	}
	tok, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil || !tok.Valid {
		t.Fatalf("state JWT invalid: %v", err)
	}
	claims, _ := tok.Claims.(jwt.MapClaims)
	if claims["purpose"] != "notion_oauth" {
		t.Errorf("purpose: got %v, want notion_oauth", claims["purpose"])
	}
	if int(claims["user_id"].(float64)) != 42 {
		t.Errorf("user_id: got %v, want 42", claims["user_id"])
	}
}

// ---- Callback ----

func TestCallback_InvalidStateJWT(t *testing.T) {
	svc := newSvc(&mockNotionRepo{}, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	_, err := svc.Callback("code123", "not-a-valid-jwt")
	if err == nil {
		t.Fatal("expected error for invalid state JWT")
	}
}

func TestCallback_ExpiredStateJWT(t *testing.T) {
	state := makeStateJWT(t, 1, "notion_oauth", time.Now().Add(-1*time.Hour))
	svc := newSvc(&mockNotionRepo{}, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	_, err := svc.Callback("code", state)
	if err == nil {
		t.Fatal("expected error for expired state JWT")
	}
}

func TestCallback_WrongPurpose(t *testing.T) {
	state := makeStateJWT(t, 1, "wrong_purpose", time.Now().Add(10*time.Minute))
	svc := newSvc(&mockNotionRepo{}, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	_, err := svc.Callback("code", state)
	if err == nil {
		t.Fatal("expected error for wrong purpose")
	}
}

func TestCallback_Success(t *testing.T) {
	tokenBody, _ := json.Marshal(map[string]string{
		"access_token":   "tok123",
		"workspace_id":   "ws1",
		"workspace_name": "My WS",
		"bot_id":         "bot1",
	})
	state := makeStateJWT(t, 5, "notion_oauth", time.Now().Add(10*time.Minute))
	repo := &mockNotionRepo{findErr: gorm.ErrRecordNotFound}
	svc := newSvc(repo, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{},
		&mockHTTP{statusCode: 200, body: string(tokenBody)},
		&mockNotionQuerier{})

	integration, err := svc.Callback("authcode", state)
	if err != nil {
		t.Fatalf("Callback error: %v", err)
	}
	if integration.AccessToken != "tok123" {
		t.Errorf("AccessToken: got %q, want tok123", integration.AccessToken)
	}
	if repo.upserted == nil {
		t.Error("expected Upsert to be called")
	}
}

// ---- Status ----

func TestStatus_NotConnected(t *testing.T) {
	svc := newSvc(&mockNotionRepo{findErr: gorm.ErrRecordNotFound}, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	status, err := svc.Status(1)
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}
	if status.Connected {
		t.Error("expected Connected=false")
	}
}

func TestStatus_Connected(t *testing.T) {
	integration := &entity.NotionIntegration{
		UserID:        1,
		WorkspaceID:   "ws1",
		WorkspaceName: "My WS",
		DatabaseID:    "db1",
	}
	svc := newSvc(&mockNotionRepo{integration: integration}, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	status, err := svc.Status(1)
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}
	if !status.Connected {
		t.Error("expected Connected=true")
	}
	if status.WorkspaceID != "ws1" {
		t.Errorf("WorkspaceID: got %q, want ws1", status.WorkspaceID)
	}
	if status.DatabaseID != "db1" {
		t.Errorf("DatabaseID: got %q, want db1", status.DatabaseID)
	}
}

// ---- Configure ----

func TestConfigure_NotConnected(t *testing.T) {
	svc := newSvc(&mockNotionRepo{findErr: gorm.ErrRecordNotFound}, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	err := svc.Configure(1, "db-id")
	if !errors.Is(err, ErrNotionNotConnected) {
		t.Errorf("expected ErrNotionNotConnected, got %v", err)
	}
}

func TestConfigure_Success(t *testing.T) {
	integration := &entity.NotionIntegration{UserID: 1}
	repo := &mockNotionRepo{integration: integration}
	svc := newSvc(repo, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	if err := svc.Configure(1, "new-db-id"); err != nil {
		t.Fatalf("Configure error: %v", err)
	}
	if repo.upserted == nil || repo.upserted.DatabaseID != "new-db-id" {
		t.Errorf("expected DatabaseID=new-db-id, got %+v", repo.upserted)
	}
}

// ---- Sync ----

func TestSync_NotConnected(t *testing.T) {
	svc := newSvc(&mockNotionRepo{findErr: gorm.ErrRecordNotFound}, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	_, err := svc.Sync(1)
	if !errors.Is(err, ErrNotionNotConnected) {
		t.Errorf("expected ErrNotionNotConnected, got %v", err)
	}
}

func TestSync_DatabaseNotSet(t *testing.T) {
	integration := &entity.NotionIntegration{UserID: 1, AccessToken: "tok", DatabaseID: ""}
	svc := newSvc(&mockNotionRepo{integration: integration}, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	_, err := svc.Sync(1)
	if !errors.Is(err, ErrNotionDatabaseNotSet) {
		t.Errorf("expected ErrNotionDatabaseNotSet, got %v", err)
	}
}

func TestSync_MissingCompanyOrPosition(t *testing.T) {
	integration := &entity.NotionIntegration{UserID: 1, AccessToken: "tok", DatabaseID: "db1"}
	// Page with no properties → Company and Position both ""
	page := notionapi.Page{ID: "page-bad", Properties: notionapi.Properties{}}
	querier := &mockNotionQuerier{pages: []notionapi.Page{page}}
	statusRepo := &mockStatusRepo{statuses: []entity.Status{{ID: 1, Text: "applied"}}}

	svc := newSvc(&mockNotionRepo{integration: integration}, &mockAppRepo{}, &mockJobRepo{},
		statusRepo, &mockHistoryRepo{}, &mockHTTP{}, querier)

	result, err := svc.Sync(1)
	if err != nil {
		t.Fatalf("Sync error: %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one error for missing Company/Position")
	}
	if result.Created != 0 {
		t.Errorf("expected Created=0, got %d", result.Created)
	}
}

func TestSync_CreatesNewJobAndApp(t *testing.T) {
	integration := &entity.NotionIntegration{UserID: 1, AccessToken: "tok", DatabaseID: "db1"}
	page := notionapi.Page{
		ID: "page-001",
		Properties: notionapi.Properties{
			"Company":  &notionapi.TitleProperty{Title: []notionapi.RichText{{PlainText: "Acme"}}},
			"Position": &notionapi.RichTextProperty{RichText: []notionapi.RichText{{PlainText: "Engineer"}}},
			"Status":   &notionapi.SelectProperty{Select: notionapi.Option{Name: "applied"}},
		},
	}
	querier := &mockNotionQuerier{pages: []notionapi.Page{page}}
	statusRepo := &mockStatusRepo{statuses: []entity.Status{{ID: 1, Text: "applied"}}}
	jobRepo := &mockJobRepo{findErr: gorm.ErrRecordNotFound}
	appRepo := &mockAppRepo{findErr: gorm.ErrRecordNotFound}
	historyRepo := &mockHistoryRepo{}

	svc := newSvc(&mockNotionRepo{integration: integration}, appRepo, jobRepo,
		statusRepo, historyRepo, &mockHTTP{}, querier)

	result, err := svc.Sync(1)
	if err != nil {
		t.Fatalf("Sync error: %v", err)
	}
	if result.Created != 1 {
		t.Errorf("expected Created=1, got %d (errors: %v)", result.Created, result.Errors)
	}
	if jobRepo.created == nil {
		t.Error("expected job to be created")
	}
	if appRepo.created == nil {
		t.Error("expected application to be created")
	}
	if len(historyRepo.created) == 0 {
		t.Error("expected status history record to be created")
	}
}

func TestSync_UpdatesStatus(t *testing.T) {
	integration := &entity.NotionIntegration{UserID: 1, AccessToken: "tok", DatabaseID: "db1"}
	pageID := "page-002"
	page := notionapi.Page{
		ID: notionapi.ObjectID(pageID),
		Properties: notionapi.Properties{
			"Company":  &notionapi.TitleProperty{Title: []notionapi.RichText{{PlainText: "Corp"}}},
			"Position": &notionapi.RichTextProperty{RichText: []notionapi.RichText{{PlainText: "Dev"}}},
			"Status":   &notionapi.SelectProperty{Select: notionapi.Option{Name: "offer"}},
		},
	}
	querier := &mockNotionQuerier{pages: []notionapi.Page{page}}
	statusRepo := &mockStatusRepo{statuses: []entity.Status{
		{ID: 1, Text: "applied"},
		{ID: 2, Text: "offer"},
	}}
	existingJob := &entity.Job{ID: 10}
	existingApp := &entity.Application{ID: 20, JobID: 10, StatusID: 1, Text: ""}
	historyRepo := &mockHistoryRepo{}

	svc := newSvc(&mockNotionRepo{integration: integration},
		&mockAppRepo{app: existingApp},
		&mockJobRepo{job: existingJob},
		statusRepo, historyRepo, &mockHTTP{}, querier)

	result, err := svc.Sync(1)
	if err != nil {
		t.Fatalf("Sync error: %v", err)
	}
	if result.Updated != 1 {
		t.Errorf("expected Updated=1, got %d (errors: %v)", result.Updated, result.Errors)
	}
	if len(historyRepo.created) == 0 {
		t.Error("expected status history record on status change")
	}
}

func TestSync_SkipsUnchanged(t *testing.T) {
	integration := &entity.NotionIntegration{UserID: 1, AccessToken: "tok", DatabaseID: "db1"}
	pageID := "page-003"
	page := notionapi.Page{
		ID: notionapi.ObjectID(pageID),
		Properties: notionapi.Properties{
			"Company":  &notionapi.TitleProperty{Title: []notionapi.RichText{{PlainText: "Stable Inc"}}},
			"Position": &notionapi.RichTextProperty{RichText: []notionapi.RichText{{PlainText: "QA"}}},
			"Status":   &notionapi.SelectProperty{Select: notionapi.Option{Name: "applied"}},
		},
	}
	querier := &mockNotionQuerier{pages: []notionapi.Page{page}}
	statusRepo := &mockStatusRepo{statuses: []entity.Status{{ID: 1, Text: "applied"}}}
	existingJob := &entity.Job{ID: 5}
	// same StatusID and empty Text (matches page's empty cover letter)
	existingApp := &entity.Application{ID: 30, JobID: 5, StatusID: 1, Text: ""}

	svc := newSvc(&mockNotionRepo{integration: integration},
		&mockAppRepo{app: existingApp},
		&mockJobRepo{job: existingJob},
		statusRepo, &mockHistoryRepo{}, &mockHTTP{}, querier)

	result, err := svc.Sync(1)
	if err != nil {
		t.Fatalf("Sync error: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("expected Skipped=1, got %d (errors: %v)", result.Skipped, result.Errors)
	}
}

func TestSync_UpdatesLastSyncAt(t *testing.T) {
	integration := &entity.NotionIntegration{UserID: 1, AccessToken: "tok", DatabaseID: "db1"}
	querier := &mockNotionQuerier{pages: []notionapi.Page{}} // no pages, quick sync
	repo := &mockNotionRepo{integration: integration}
	statusRepo := &mockStatusRepo{statuses: []entity.Status{{ID: 1, Text: "applied"}}}

	svc := newSvc(repo, &mockAppRepo{}, &mockJobRepo{},
		statusRepo, &mockHistoryRepo{}, &mockHTTP{}, querier)

	_, err := svc.Sync(1)
	if err != nil {
		t.Fatalf("Sync error: %v", err)
	}
	if repo.upserted == nil || repo.upserted.LastSyncAt == nil {
		t.Error("expected LastSyncAt to be set after sync")
	}
}

// ---- Disconnect ----

func TestDisconnect(t *testing.T) {
	repo := &mockNotionRepo{}
	svc := newSvc(repo, &mockAppRepo{}, &mockJobRepo{},
		&mockStatusRepo{}, &mockHistoryRepo{}, &mockHTTP{}, &mockNotionQuerier{})
	if err := svc.Disconnect(1); err != nil {
		t.Fatalf("Disconnect error: %v", err)
	}
}
