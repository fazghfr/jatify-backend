package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"job-tracker/internal/dto"
	"job-tracker/internal/entity"
	"job-tracker/internal/middleware"
	"job-tracker/internal/service"

	"github.com/gin-gonic/gin"
)

// ---- mock service ----

type mockNotionSvc struct {
	connectURL    string
	connectErr    error
	callbackRes   *entity.NotionIntegration
	callbackErr   error
	statusRes     *dto.NotionStatusResponse
	statusErr     error
	configureErr  error
	listDBRes     []dto.NotionDatabaseItem
	listDBErr     error
	syncRes       *dto.NotionSyncResult
	syncErr       error
	disconnectErr error
}

func (m *mockNotionSvc) Connect(_ int) (string, error) {
	return m.connectURL, m.connectErr
}
func (m *mockNotionSvc) Callback(_, _ string) (*entity.NotionIntegration, error) {
	return m.callbackRes, m.callbackErr
}
func (m *mockNotionSvc) Status(_ int) (*dto.NotionStatusResponse, error) {
	return m.statusRes, m.statusErr
}
func (m *mockNotionSvc) Configure(_ int, _ string) error { return m.configureErr }
func (m *mockNotionSvc) ListDatabases(_ int) ([]dto.NotionDatabaseItem, error) {
	return m.listDBRes, m.listDBErr
}
func (m *mockNotionSvc) Sync(_ int) (*dto.NotionSyncResult, error) {
	return m.syncRes, m.syncErr
}
func (m *mockNotionSvc) Disconnect(_ int) error { return m.disconnectErr }

// ---- router setup ----

type apiResp struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func makeRouter(svc service.NotionService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, 1)
		c.Next()
	})
	h := NewNotionHandler(svc)
	r.GET("/notion/connect", h.Connect)
	r.GET("/notion/callback", h.Callback)
	r.GET("/notion/status", h.Status)
	r.POST("/notion/configure", h.Configure)
	r.GET("/notion/databases", h.ListDatabases)
	r.POST("/notion/sync", h.Sync)
	r.DELETE("/notion/disconnect", h.Disconnect)
	return r
}

func decodeResp(t *testing.T, w *httptest.ResponseRecorder) apiResp {
	t.Helper()
	var resp apiResp
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

// ---- Connect ----

func TestConnect_Handler_Success(t *testing.T) {
	r := makeRouter(&mockNotionSvc{connectURL: "https://notion.com/oauth?foo=bar"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/connect", nil))

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	resp := decodeResp(t, w)
	if !resp.Success {
		t.Error("expected success=true")
	}
	var data map[string]string
	json.Unmarshal(resp.Data, &data)
	if data["url"] == "" {
		t.Error("expected url key in data")
	}
}

func TestConnect_Handler_ServiceError(t *testing.T) {
	r := makeRouter(&mockNotionSvc{connectErr: errors.New("oops")})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/connect", nil))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// ---- Callback ----

func TestCallback_MissingCode(t *testing.T) {
	r := makeRouter(&mockNotionSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/callback?state=abc", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestCallback_MissingState(t *testing.T) {
	r := makeRouter(&mockNotionSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/callback?code=xyz", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestCallback_ServiceError(t *testing.T) {
	r := makeRouter(&mockNotionSvc{callbackErr: errors.New("bad state")})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/callback?code=x&state=y", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestCallback_Success(t *testing.T) {
	integration := &entity.NotionIntegration{
		WorkspaceID:   "ws1",
		WorkspaceName: "My WS",
	}
	r := makeRouter(&mockNotionSvc{callbackRes: integration})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/callback?code=x&state=y", nil))
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	resp := decodeResp(t, w)
	var data map[string]interface{}
	json.Unmarshal(resp.Data, &data)
	if data["connected"] != true {
		t.Errorf("expected connected=true, got %v", data["connected"])
	}
}

// ---- Status ----

func TestStatus_Handler_Success(t *testing.T) {
	statusResp := &dto.NotionStatusResponse{Connected: true, WorkspaceID: "ws1"}
	r := makeRouter(&mockNotionSvc{statusRes: statusResp})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/status", nil))
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	resp := decodeResp(t, w)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestStatus_Handler_ServiceError(t *testing.T) {
	r := makeRouter(&mockNotionSvc{statusErr: errors.New("db error")})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/status", nil))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// ---- Configure ----

func TestConfigure_MissingBody(t *testing.T) {
	r := makeRouter(&mockNotionSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/notion/configure", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestConfigure_NotConnected(t *testing.T) {
	r := makeRouter(&mockNotionSvc{configureErr: service.ErrNotionNotConnected})
	body := bytes.NewBufferString(`{"database_id":"test-db"}`)
	req := httptest.NewRequest(http.MethodPost, "/notion/configure", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestConfigure_Success(t *testing.T) {
	r := makeRouter(&mockNotionSvc{})
	body := bytes.NewBufferString(`{"database_id":"test-db"}`)
	req := httptest.NewRequest(http.MethodPost, "/notion/configure", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

// ---- ListDatabases ----

func TestListDatabases_NotConnected(t *testing.T) {
	r := makeRouter(&mockNotionSvc{listDBErr: service.ErrNotionNotConnected})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/databases", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestListDatabases_Success(t *testing.T) {
	dbs := []dto.NotionDatabaseItem{{ID: "db1", Title: "My DB"}}
	r := makeRouter(&mockNotionSvc{listDBRes: dbs})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notion/databases", nil))
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	resp := decodeResp(t, w)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

// ---- Sync ----

func TestSync_NotConnected(t *testing.T) {
	r := makeRouter(&mockNotionSvc{syncErr: service.ErrNotionNotConnected})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/notion/sync", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestSync_DatabaseNotSet(t *testing.T) {
	r := makeRouter(&mockNotionSvc{syncErr: service.ErrNotionDatabaseNotSet})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/notion/sync", nil))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestSync_Success(t *testing.T) {
	syncRes := &dto.NotionSyncResult{Created: 2, Updated: 1, Skipped: 0, Errors: []string{}}
	r := makeRouter(&mockNotionSvc{syncRes: syncRes})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/notion/sync", nil))
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	resp := decodeResp(t, w)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

// ---- Disconnect ----

func TestDisconnect_Handler_Success(t *testing.T) {
	r := makeRouter(&mockNotionSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/notion/disconnect", nil))
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestDisconnect_Handler_Error(t *testing.T) {
	r := makeRouter(&mockNotionSvc{disconnectErr: errors.New("db error")})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/notion/disconnect", nil))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}
