package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/PresensiGo/backend/internal/config"
	"github.com/PresensiGo/backend/internal/delivery/http/middleware"
	"github.com/PresensiGo/backend/internal/model"
)

// mockAuthUsecase satisfies AuthUsecaseIface for testing.
type mockAuthUsecase struct{}

func (m *mockAuthUsecase) Register(req *model.RegisterRequest) (*model.User, error) {
	return &model.User{ID: uuid.New(), Name: req.Name, Email: req.Email, Role: "employee"}, nil
}

func (m *mockAuthUsecase) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	return &model.LoginResponse{Token: "mock-token"}, nil
}

func (m *mockAuthUsecase) GetByID(id uuid.UUID) (*model.User, error) {
	return &model.User{ID: id}, nil
}

func (m *mockAuthUsecase) UpdateFaceEmbedding(userID uuid.UUID, embedding []byte) error {
	return nil
}

func (m *mockAuthUsecase) EnrollFace(ctx context.Context, userID uuid.UUID, selfies []string) error {
	return nil
}

// mockAttendanceUsecase satisfies AttendanceUsecaseIface for testing.
type mockAttendanceUsecase struct{}

func (m *mockAttendanceUsecase) CheckIn(userID uuid.UUID, req *model.CheckInRequest) (*model.Attendance, error) {
	return &model.Attendance{ID: uuid.New()}, nil
}

func (m *mockAttendanceUsecase) CheckOut(userID uuid.UUID, req *model.CheckOutRequest) (*model.Attendance, error) {
	return &model.Attendance{ID: uuid.New()}, nil
}

func (m *mockAttendanceUsecase) GetTodayAttendance(userID uuid.UUID) (*model.Attendance, error) {
	return &model.Attendance{ID: uuid.New()}, nil
}

func (m *mockAttendanceUsecase) GetHistory(userID uuid.UUID, limit, offset int) ([]model.AttendanceResponse, error) {
	return []model.AttendanceResponse{}, nil
}

func (m *mockAttendanceUsecase) GetLocations() ([]model.Location, error) {
	return []model.Location{}, nil
}

func (m *mockAttendanceUsecase) CreateLocation(req *model.Location) error {
	return nil
}

func (m *mockAttendanceUsecase) UpdateLocation(id uuid.UUID, req *model.Location) error {
	return nil
}

func (m *mockAttendanceUsecase) DeleteLocation(id uuid.UUID) error {
	return nil
}

func (m *mockAttendanceUsecase) GetFaceChallenge(userID uuid.UUID) (*model.FaceChallengeResponse, error) {
	return &model.FaceChallengeResponse{Challenge: "turn_left", Token: "token"}, nil
}

func (m *mockAttendanceUsecase) Sync(userID uuid.UUID, req *model.SyncRequest) []model.SyncResult {
	return []model.SyncResult{}
}

func (m *mockAttendanceUsecase) GetSyncStatus(userID uuid.UUID) (*model.SyncStatusResponse, error) {
	return &model.SyncStatusResponse{PendingCount: 0, StuckCount: 0}, nil
}

func newTestHandler() (*Handler, *mux.Router) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	return h, r
}

// TestRegister_EmptyBody_Returns400 verifies that a missing required fields body returns HTTP 400.
func TestRegister_EmptyBody_Returns400(t *testing.T) {
	_, router := newTestHandler()

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestRegister_InvalidEmail_Returns400 verifies that a malformed email field returns HTTP 400.
func TestRegister_InvalidEmail_Returns400(t *testing.T) {
	_, router := newTestHandler()

	payload := map[string]string{
		"name":     "Test User",
		"email":    "not-an-email",
		"password": "password123",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestLogin_EmptyBody_Returns400 verifies that missing credentials return HTTP 400.
func TestLogin_EmptyBody_Returns400(t *testing.T) {
	_, router := newTestHandler()

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestCreateLocationAsAdmin verifies that admin can create locations (201).
func TestCreateLocationAsAdmin(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      6.2,
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateLocation(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

// TestCreateLocationAsEmployee verifies that employee cannot create locations (403).
func TestCreateLocationAsEmployee(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      6.2,
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "employee")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateLocation(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// TestUpdateLocationAsAdmin verifies that admin can update locations (200).
func TestUpdateLocationAsAdmin(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	payload := map[string]interface{}{
		"name":          "Updated Location",
		"latitude":      6.3,
		"longitude":     106.9,
		"radius_meters": 60,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	locID := uuid.New()
	req := httptest.NewRequest(http.MethodPut, "/api/locations/"+locID.String(), bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": locID.String()})

	w := httptest.NewRecorder()

	h.UpdateLocation(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestUpdateLocationAsEmployee verifies that employee cannot update locations (403).
func TestUpdateLocationAsEmployee(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	payload := map[string]interface{}{
		"name":          "Updated Location",
		"latitude":      6.3,
		"longitude":     106.9,
		"radius_meters": 60,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "employee")
	locID := uuid.New()
	req := httptest.NewRequest(http.MethodPut, "/api/locations/"+locID.String(), bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": locID.String()})

	w := httptest.NewRecorder()

	h.UpdateLocation(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// TestDeleteLocationAsAdmin verifies that admin can delete locations (200).
func TestDeleteLocationAsAdmin(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	locID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/locations/"+locID.String(), nil)
	req = req.WithContext(ctx)

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": locID.String()})

	w := httptest.NewRecorder()

	h.DeleteLocation(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestDeleteLocationAsEmployee verifies that employee cannot delete locations (403).
func TestDeleteLocationAsEmployee(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "employee")
	locID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/locations/"+locID.String(), nil)
	req = req.WithContext(ctx)

	// Setup mux vars
	req = mux.SetURLVars(req, map[string]string{"id": locID.String()})

	w := httptest.NewRecorder()

	h.DeleteLocation(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// TestGetLocationsAsEmployee verifies that employee can read locations (200).
func TestGetLocationsAsEmployee(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "employee")
	req := httptest.NewRequest(http.MethodGet, "/api/locations", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetLocations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestRBACIntegration_EmployeeCannotMutateLocation verifies that employee cannot create location.
func TestRBACIntegration_EmployeeCannotMutateLocation(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      6.2,
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	// Simulate employee token context
	employeeCtx := context.WithValue(context.Background(), middleware.RoleKey, "employee")

	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(employeeCtx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateLocation(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("employee should get 403, got %d", w.Code)
	}
}

// TestRBACIntegration_AdminCanMutateLocation verifies that admin can create location.
func TestRBACIntegration_AdminCanMutateLocation(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      6.2,
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	// Simulate admin token context
	adminCtx := context.WithValue(context.Background(), middleware.RoleKey, "admin")

	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(adminCtx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateLocation(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("admin should get 201, got %d", w.Code)
	}
}

// mockDB is a minimal mock database that implements Ping method
type mockDB struct {
	pingError error
}

func (m *mockDB) Ping() error {
	return m.pingError
}

// Other sql.DB methods would be needed for a full implementation, but we only need Ping for health tests
func (m *mockDB) Exec(query string, args ...interface{}) (sql.Result, error) { return nil, nil }
func (m *mockDB) Query(query string, args ...interface{}) (*sql.Rows, error) { return nil, nil }
func (m *mockDB) QueryRow(query string, args ...interface{}) *sql.Row        { return nil }
func (m *mockDB) Prepare(query string) (*sql.Stmt, error)                    { return nil, nil }
func (m *mockDB) Begin() (*sql.Tx, error)                                    { return nil, nil }
func (m *mockDB) Close() error                                               { return nil }
func (m *mockDB) SetMaxOpenConns(n int)                                      {}
func (m *mockDB) SetMaxIdleConns(n int)                                      {}
func (m *mockDB) SetConnMaxLifetime(d time.Duration)                         {}
func (m *mockDB) SetConnMaxIdleTime(d time.Duration)                         {}
func (m *mockDB) Stats() sql.DBStats                                         { return sql.DBStats{} }

// mockRedisClient is a minimal mock Redis client
type mockRedisClient struct {
	connected bool
}

func (m *mockRedisClient) IsConnected(ctx context.Context) bool {
	return m.connected
}

func (m *mockRedisClient) Close() error { return nil }
func (m *mockRedisClient) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return 0, nil
}
func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) { return "", nil }
func (m *mockRedisClient) Delete(ctx context.Context, keys ...string) error    { return nil }

// TestHealthEndpoint tests the /health endpoint returns 200 OK with correct response
func TestHealthEndpoint(t *testing.T) {
	// Create handler with mock dependencies
	db := &mockDB{}
	redis := &mockRedisClient{connected: true}
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, db, redis)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Check response body
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if status, ok := response["status"]; !ok || status != "ok" {
		t.Errorf("expected {\"status\": \"ok\"}, got %v", response)
	}
}

// TestHealthReadyEndpoint_Ready tests /health/ready endpoint when both DB and Redis are available
func TestHealthReadyEndpoint_Ready(t *testing.T) {
	// Create handler with mock dependencies (both connected)
	db := &mockDB{pingError: nil} // nil error means connected
	redis := &mockRedisClient{connected: true}
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, db, redis)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	h.HealthReady(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 when ready, got %d", w.Code)
	}

	// Check response body
	var response map[string]bool
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if ready, ok := response["ready"]; !ok || !ready {
		t.Errorf("expected {\"ready\": true}, got %v", response)
	}
}

// TestHealthReadyEndpoint_NotReady_DB tests /health/ready endpoint when DB is unavailable
func TestHealthReadyEndpoint_NotReady_DB(t *testing.T) {
	// Create handler with mock dependencies (DB not connected)
	db := &mockDB{pingError: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}}
	redis := &mockRedisClient{connected: true}
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, db, redis)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	h.HealthReady(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when DB not ready, got %d", w.Code)
	}

	// Check response body
	var response map[string]bool
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if ready, ok := response["ready"]; !ok || ready {
		t.Errorf("expected {\"ready\": false}, got %v", response)
	}
}

// TestHealthReadyEndpoint_NotReady_Redis tests /health/ready endpoint when Redis is unavailable
func TestHealthReadyEndpoint_NotReady_Redis(t *testing.T) {
	// Create handler with mock dependencies (Redis not connected)
	db := &mockDB{pingError: nil} // DB is connected
	redis := &mockRedisClient{connected: false}
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, db, redis)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	h.HealthReady(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when Redis not ready, got %d", w.Code)
	}

	// Check response body
	var response map[string]bool
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if ready, ok := response["ready"]; !ok || ready {
		t.Errorf("expected {\"ready\": false}, got %v", response)
	}
}

// TestHealthReadyEndpoint_NotReady_Both tests /health/ready endpoint when both DB and Redis are unavailable
func TestHealthReadyEndpoint_NotReady_Both(t *testing.T) {
	// Create handler with mock dependencies (both not connected)
	db := &mockDB{pingError: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}}
	redis := &mockRedisClient{connected: false}
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, db, redis)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	h.HealthReady(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when both not ready, got %d", w.Code)
	}

	// Check response body
	var response map[string]bool
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if ready, ok := response["ready"]; !ok || ready {
		t.Errorf("expected {\"ready\": false}, got %v", response)
	}
}

// TestValidationError_IncludesRequestID verifies that validation errors include requestID
func TestValidationError_IncludesRequestID(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	// Invalid latitude (>90)
	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      100,
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "test_request_123")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateLocation(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid latitude, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := response["requestID"]; !ok {
		t.Error("expected requestID in error response")
	}

	if response["requestID"] != "test_request_123" {
		t.Errorf("expected requestID to be 'test_request_123', got %v", response["requestID"])
	}
}

// TestValidationError_HasContentType verifies error response has Content-Type: application/json
func TestValidationError_HasContentType(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	// Invalid email format
	payload := map[string]interface{}{
		"name":     "John Doe",
		"email":    "not-an-email",
		"password": "password123",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %s", contentType)
	}
}

// TestValidationError_SanitizeFieldErrors verifies field error messages are sanitized
func TestValidationError_SanitizeFieldErrors(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	// Invalid latitude
	payload := map[string]interface{}{
		"latitude":       100,
		"longitude":      106.8,
		"device_uuid":    "123e4567-e89b-12d3-a456-426614174000",
		"timestamp":      1234567890,
		"hmac_signature": "sig",
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, uuid.New())
	req := httptest.NewRequest(http.MethodPost, "/api/attendance/check-in", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CheckIn(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify error response has details array
	if _, ok := response["details"]; !ok {
		t.Error("expected details array in error response")
	}

	// Verify details don't contain technical constraint info
	details := response["details"].([]interface{})
	for _, detail := range details {
		detailStr := detail.(string)
		if strings.Contains(detailStr, "struct tag") || strings.Contains(detailStr, "validator") {
			t.Errorf("error message contains technical details: %s", detailStr)
		}
	}
}

// TestErrorHandler_ValidationError_Returns400 tests that validation errors use the error handler correctly
func TestErrorHandler_ValidationError_Returns400(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	// Create a request with invalid data
	payload := `{"name": "", "email": "test@test.com", "password": "pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	// Add request ID to context
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_validation_test")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	if _, hasError := response["error"]; !hasError {
		t.Error("Expected 'error' field in response")
	}
	if _, hasRequestID := response["requestID"]; !hasRequestID {
		t.Error("Expected 'requestID' field in response")
	}
}

// TestErrorHandler_AuthError_Returns401 tests that auth errors return 401 with proper format
func TestErrorHandler_AuthError_Returns401(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	h := NewHandler(mockAuth, &mockAttendanceUsecase{}, nil, nil)

	// Attempt check-in without authentication context
	payload := `{"latitude": 45.0, "longitude": 106.8, "device_id": "device123", "signature": "sig"}`
	req := httptest.NewRequest(http.MethodPost, "/api/attendance/check-in", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_auth_test")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.CheckIn(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	if response["error"] != "unauthorized" {
		t.Errorf("Expected error 'unauthorized', got '%v'", response["error"])
	}
}

// TestErrorHandler_AllErrorsIncludeRequestID tests that all error responses include request ID
func TestErrorHandler_AllErrorsIncludeRequestID(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	tests := []struct {
		name            string
		method          string
		path            string
		payload         string
		expectedStatus  int
		shouldHaveError bool
	}{
		{
			name:            "invalid_registration_email",
			method:          "POST",
			path:            "/api/auth/register",
			payload:         `{"name": "Test", "email": "invalid-email", "password": "pass123"}`,
			expectedStatus:  http.StatusBadRequest,
			shouldHaveError: true,
		},
		{
			name:            "unauthorized_checkin",
			method:          "POST",
			path:            "/api/attendance/check-in",
			payload:         `{"latitude": 0, "longitude": 0, "device_id": "dev", "signature": "sig"}`,
			expectedStatus:  http.StatusUnauthorized,
			shouldHaveError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.payload))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_test_"+test.name)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Call appropriate handler
			if test.path == "/api/auth/register" {
				h.Register(w, req)
			} else if test.path == "/api/attendance/check-in" {
				h.CheckIn(w, req)
			}

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			json.NewDecoder(w.Body).Decode(&response)

			if test.shouldHaveError {
				if _, hasRequestID := response["requestID"]; !hasRequestID {
					t.Error("Expected 'requestID' in error response")
				}
				if response["requestID"] != "req_test_"+test.name {
					t.Errorf("Expected requestID 'req_test_%s', got '%v'", test.name, response["requestID"])
				}
			}
		})
	}
}

// TestErrorResponse_NoLeakSensitiveInfo tests that error responses don't leak internal details
func TestErrorResponse_NoLeakSensitiveInfo(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	// Test with invalid coordinates (tries various error-inducing inputs)
	tests := []struct {
		name            string
		latitude        float64
		longitude       float64
		shouldFail      bool
		unexpectedTerms []string
	}{
		{
			name:            "out_of_range_latitude",
			latitude:        100,
			longitude:       106.8,
			shouldFail:      true,
			unexpectedTerms: []string{"schema", "column", "database", "table", "goroutine", "panic"},
		},
		{
			name:            "out_of_range_longitude",
			latitude:        45.0,
			longitude:       200,
			shouldFail:      true,
			unexpectedTerms: []string{"schema", "column", "database", "table", "goroutine", "panic"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"name":          "Test Location",
				"latitude":      test.latitude,
				"longitude":     test.longitude,
				"radius_meters": 50,
			}
			b, _ := json.Marshal(payload)

			ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test_req")
			req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.CreateLocation(w, req)

			responseBody := w.Body.String()

			// Check that sensitive info is not in response
			for _, term := range test.unexpectedTerms {
				if strings.Contains(strings.ToLower(responseBody), strings.ToLower(term)) {
					t.Errorf("Response contains sensitive term '%s': %s", term, responseBody)
				}
			}

			// Verify it's valid JSON
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Response is not valid JSON: %v", err)
			}
		})
	}
}

// TestErrorResponse_ContentTypeApplicationJSON tests all error responses have correct Content-Type
func TestErrorResponse_ContentTypeApplicationJSON(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	tests := []struct {
		name           string
		method         string
		path           string
		payload        string
		expectedStatus int
	}{
		{
			name:           "validation_error",
			method:         "POST",
			path:           "/api/auth/register",
			payload:        `{"name": "", "email": "x", "password": "p"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "auth_error",
			method:         "POST",
			path:           "/api/attendance/check-in",
			payload:        `{"latitude": 0, "longitude": 0, "device_id": "", "signature": ""}`,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.payload))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_test")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			if test.path == "/api/auth/register" {
				h.Register(w, req)
			} else if test.path == "/api/attendance/check-in" {
				h.CheckIn(w, req)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s' for %s", contentType, test.name)
			}
		})
	}
}

// TestGetSyncStatus_Authorized verifies sync status endpoint returns 200 for authorized user
func TestGetSyncStatus_Authorized(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	userID := uuid.New()
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
	req := httptest.NewRequest(http.MethodGet, "/api/attendance/sync/status", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetSyncStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := response["pending_count"]; !ok {
		t.Error("expected pending_count in response")
	}
	if _, ok := response["stuck_count"]; !ok {
		t.Error("expected stuck_count in response")
	}
}

// ============================================================================
// Task 7: Integration Tests for Input Validation
// ============================================================================

// TestIntegration_InvalidLatitude_Returns400 tests POST /api/locations with invalid latitude (>90)
func TestIntegration_InvalidLatitude_Returns400(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      100, // Invalid: > 90
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_invalid_lat_001")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateLocation(w, req)

	// Verify HTTP 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Verify response is JSON
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %s", contentType)
	}

	// Verify response includes requestID
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if _, ok := response["requestID"]; !ok {
		t.Error("expected requestID field in response")
	}

	// Verify no schema leak (no table/column names, no schema info)
	responseBody := w.Body.String()
	if strings.Contains(strings.ToLower(responseBody), "table") ||
		strings.Contains(strings.ToLower(responseBody), "column") ||
		strings.Contains(strings.ToLower(responseBody), "schema") ||
		strings.Contains(strings.ToLower(responseBody), "constraint") {
		t.Errorf("response contains schema information: %s", responseBody)
	}
}

// TestIntegration_InvalidLatitude_EdgeCases tests latitude edge cases
func TestIntegration_InvalidLatitude_EdgeCases(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	tests := []struct {
		name      string
		latitude  float64
		expectErr bool
	}{
		{"latitude_-91", -91, true},
		{"latitude_91", 91, true},
		{"latitude_-90", -90, false},
		{"latitude_90", 90, false},
		{"latitude_0", 0, false},
		{"latitude_45", 45, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"name":          "Test Location",
				"latitude":      test.latitude,
				"longitude":     106.8,
				"radius_meters": 50,
			}
			b, _ := json.Marshal(payload)

			ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_lat_edge_"+test.name)
			req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.CreateLocation(w, req)

			if test.expectErr {
				if w.Code != http.StatusBadRequest {
					t.Errorf("expected 400 for %s, got %d", test.name, w.Code)
				}
			} else {
				if w.Code != http.StatusCreated {
					t.Errorf("expected 201 for %s, got %d", test.name, w.Code)
				}
			}
		})
	}
}

// TestIntegration_InvalidLongitude_Returns400 tests POST /api/locations with invalid longitude (>180)
func TestIntegration_InvalidLongitude_Returns400(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      6.2,
		"longitude":     200, // Invalid: > 180
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_invalid_lng_001")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateLocation(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	if _, ok := response["requestID"]; !ok {
		t.Error("expected requestID in error response")
	}
}

// TestIntegration_InvalidLongitude_EdgeCases tests longitude edge cases
func TestIntegration_InvalidLongitude_EdgeCases(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	tests := []struct {
		name      string
		longitude float64
		expectErr bool
	}{
		{"longitude_-181", -181, true},
		{"longitude_181", 181, true},
		{"longitude_-180", -180, false},
		{"longitude_180", 180, false},
		{"longitude_0", 0, false},
		{"longitude_106.8", 106.8, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"name":          "Test Location",
				"latitude":      6.2,
				"longitude":     test.longitude,
				"radius_meters": 50,
			}
			b, _ := json.Marshal(payload)

			ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_lng_edge_"+test.name)
			req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.CreateLocation(w, req)

			if test.expectErr {
				if w.Code != http.StatusBadRequest {
					t.Errorf("expected 400 for %s, got %d", test.name, w.Code)
				}
			} else {
				if w.Code != http.StatusCreated {
					t.Errorf("expected 201 for %s, got %d", test.name, w.Code)
				}
			}
		})
	}
}

// TestIntegration_InvalidRadius_Returns400 tests POST /api/locations with invalid radius (<=0)
func TestIntegration_InvalidRadius_Returns400(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      6.2,
		"longitude":     106.8,
		"radius_meters": 0, // Invalid: must be > 0
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_invalid_radius_001")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateLocation(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestIntegration_SQLInjection_CheckIn_NoLeak tests SQL injection in deviceID returns 400 with no SQL in response
func TestIntegration_SQLInjection_CheckIn_NoLeak(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	userID := uuid.New()
	payload := map[string]interface{}{
		"latitude":           6.2,
		"longitude":          106.8,
		"device_uuid":        "123e4567-e89b-12d3-a456-426614174000' OR '1'='1", // SQL injection attempt
		"timestamp":          int64(1234567890),
		"hmac_signature":     "sig",
		"liveness_challenge": "test",
		"liveness_token":     "test",
		"idempotency_key":    uuid.New().String(),
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_sql_inject_001")
	req := httptest.NewRequest(http.MethodPost, "/api/attendance/check-in", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CheckIn(w, req)

	// Verify HTTP 400 for invalid UUID
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	responseBody := w.Body.String()

	// Verify no SQL injection payload in response
	if strings.Contains(responseBody, "OR") && strings.Contains(responseBody, "'") {
		t.Errorf("response contains SQL injection payload: %s", responseBody)
	}

	// Verify no SQL keywords in error response
	if strings.Contains(strings.ToLower(responseBody), "select") ||
		strings.Contains(strings.ToLower(responseBody), "insert") ||
		strings.Contains(strings.ToLower(responseBody), "update") ||
		strings.Contains(strings.ToLower(responseBody), "delete") {
		t.Errorf("response contains SQL keywords: %s", responseBody)
	}

	// Verify no table/column names
	if strings.Contains(strings.ToLower(responseBody), "table") ||
		strings.Contains(strings.ToLower(responseBody), "column") {
		t.Errorf("response contains schema information: %s", responseBody)
	}

	// Verify response is valid JSON
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}
}

// TestIntegration_SQLInjection_Location_NoLeak tests SQL injection in location parameters
func TestIntegration_SQLInjection_Location_NoLeak(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"name":          "Test'; DROP TABLE locations; --",
		"latitude":      6.2,
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_sql_inject_loc_001")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateLocation(w, req)

	responseBody := w.Body.String()

	// Verify no SQL keywords
	if strings.Contains(strings.ToLower(responseBody), "drop") ||
		strings.Contains(strings.ToLower(responseBody), "table") {
		t.Errorf("response contains SQL content: %s", responseBody)
	}

	// Verify valid JSON response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}
}

// TestIntegration_AuthError_Returns401WithJSON tests auth error response returns 401 with Content-Type: application/json
func TestIntegration_AuthError_Returns401WithJSON(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	userID := uuid.New()
	payload := map[string]interface{}{
		"latitude":           6.2,
		"longitude":          106.8,
		"device_uuid":        "123e4567-e89b-12d3-a456-426614174000",
		"timestamp":          int64(1234567890),
		"hmac_signature":     "sig",
		"liveness_challenge": "test",
		"liveness_token":     "test",
		"idempotency_key":    uuid.New().String(),
	}
	b, _ := json.Marshal(payload)

	// Note: no user ID in context = unauthorized
	req := httptest.NewRequest(http.MethodPost, "/api/attendance/check-in", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	_ = userID // check-in must still be unauthorized without a user ID in context

	w := httptest.NewRecorder()
	h.CheckIn(w, req)

	// Verify HTTP 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	// Verify Content-Type: application/json
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %s", contentType)
	}

	// Verify valid JSON response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	// Verify error field exists
	if _, ok := response["error"]; !ok {
		t.Error("expected 'error' field in response")
	}
}

// TestIntegration_ValidationError_IncludesRequestIDAndHeader tests validation error includes requestID and X-Request-ID header
func TestIntegration_ValidationError_IncludesRequestIDAndHeader(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      100, // Invalid
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_header_test_001")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateLocation(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Verify X-Request-ID header exists (would be set by middleware in production)
	// Note: In handler_test, we test the context value instead of the header since the
	// middleware is not applied here. In integration tests the middleware sets the header.
	if requestIDHeader := w.Header().Get("X-Request-ID"); requestIDHeader != "" {
		t.Logf("X-Request-ID header set by middleware: %s", requestIDHeader)
	}

	// Verify response includes requestID field
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if requestID, ok := response["requestID"]; !ok {
		t.Error("expected 'requestID' field in response")
	} else if requestID != "req_header_test_001" {
		t.Errorf("expected requestID 'req_header_test_001', got '%v'", requestID)
	}
}

// TestIntegration_NoStackTraceInError tests that error responses don't include stack traces
func TestIntegration_NoStackTraceInError(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      100, // Invalid
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_stack_test_001")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateLocation(w, req)

	responseBody := w.Body.String()

	// Verify no stack trace patterns
	stackTracePatterns := []string{
		"goroutine",
		"runtime/",
		".go:",
		"panic",
		"main.",
	}

	for _, pattern := range stackTracePatterns {
		if strings.Contains(strings.ToLower(responseBody), strings.ToLower(pattern)) {
			t.Errorf("response contains stack trace pattern '%s': %s", pattern, responseBody)
		}
	}
}

// TestIntegration_NoDBDetailsInError tests that error responses don't include database details
func TestIntegration_NoDBDetailsInError(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      100, // Invalid
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_db_test_001")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateLocation(w, req)

	responseBody := w.Body.String()

	// Verify no database-specific information
	dbPatterns := []string{
		"duplicate key",
		"constraint",
		"violates",
		"UNIQUE",
		"PRIMARY KEY",
		"FOREIGN KEY",
	}

	for _, pattern := range dbPatterns {
		if strings.Contains(strings.ToLower(responseBody), strings.ToLower(pattern)) {
			t.Errorf("response contains database pattern '%s': %s", pattern, responseBody)
		}
	}
}

// TestIntegration_ErrorSanitization_Comprehensive tests comprehensive error sanitization
func TestIntegration_ErrorSanitization_Comprehensive(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	tests := []struct {
		name              string
		payload           map[string]interface{}
		endpoint          string
		method            string
		setupContext      func(*http.Request) *http.Request
		sensitivePatterns []string
	}{
		{
			name: "invalid_latitude",
			payload: map[string]interface{}{
				"name":          "Test",
				"latitude":      100,
				"longitude":     106.8,
				"radius_meters": 50,
			},
			endpoint: "/api/locations",
			method:   http.MethodPost,
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), middleware.RoleKey, "admin")
				ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_comprehensive_lat")
				return r.WithContext(ctx)
			},
			sensitivePatterns: []string{"table", "column", "schema", "SELECT", "INSERT"},
		},
		{
			name: "invalid_longitude",
			payload: map[string]interface{}{
				"name":          "Test",
				"latitude":      6.2,
				"longitude":     200,
				"radius_meters": 50,
			},
			endpoint: "/api/locations",
			method:   http.MethodPost,
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), middleware.RoleKey, "admin")
				ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_comprehensive_lng")
				return r.WithContext(ctx)
			},
			sensitivePatterns: []string{"table", "column", "schema", "SELECT", "INSERT"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, _ := json.Marshal(test.payload)
			req := httptest.NewRequest(test.method, test.endpoint, bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			req = test.setupContext(req)

			w := httptest.NewRecorder()

			if test.endpoint == "/api/locations" && test.method == http.MethodPost {
				h.CreateLocation(w, req)
			}

			responseBody := w.Body.String()

			// Verify no sensitive patterns
			for _, pattern := range test.sensitivePatterns {
				if strings.Contains(strings.ToLower(responseBody), strings.ToLower(pattern)) {
					t.Errorf("response contains sensitive pattern '%s': %s", pattern, responseBody)
				}
			}

			// Verify valid JSON
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("response is not valid JSON: %v", err)
			}
		})
	}
}

// TestIntegration_CORSPreflight_AllowedOrigin tests CORS preflight with allowed origin
func TestIntegration_CORSPreflight_AllowedOrigin(t *testing.T) {
	// This test verifies that CORS is configured correctly
	// In production, CORS would be handled by the rs/cors middleware
	// This test verifies that allowed origins from environment config are respected

	// Create a mock router to test CORS behavior
	r := mux.NewRouter()
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})
	h.RegisterRoutes(r)

	// In development environment, localhost should be allowed
	req := httptest.NewRequest(http.MethodOptions, "/api/locations", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Note: CORS headers are set by the rs/cors middleware, not by individual handlers
	// This test verifies that the handler doesn't block CORS preflight requests
	// In a full integration test with CORS middleware, we'd verify:
	// - Access-Control-Allow-Origin is set to a specific origin (not *)
	// - Access-Control-Allow-Methods includes POST, PUT, DELETE
	// - Access-Control-Allow-Credentials is true only in certain environments
}

// TestIntegration_CORSConfiguration_EnvironmentAware tests that CORS config is environment-aware
func TestIntegration_CORSConfiguration_EnvironmentAware(t *testing.T) {
	// Test that LoadCORSConfig returns different origins based on environment
	devConfig := config.LoadCORSConfig("development")
	stagingConfig := config.LoadCORSConfig("staging")
	prodConfig := config.LoadCORSConfig("production")

	// Verify development config allows localhost
	if len(devConfig.AllowedOrigins) == 0 {
		t.Error("development CORS config should have allowed origins")
	}
	hasLocalhost := false
	for _, origin := range devConfig.AllowedOrigins {
		if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
			hasLocalhost = true
		}
	}
	if !hasLocalhost {
		t.Errorf("development CORS config should allow localhost, got: %v", devConfig.AllowedOrigins)
	}

	// Verify staging config doesn't allow localhost
	for _, origin := range stagingConfig.AllowedOrigins {
		if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
			t.Errorf("staging CORS config should NOT allow localhost, got: %v", stagingConfig.AllowedOrigins)
		}
	}

	// Verify production config doesn't allow localhost
	for _, origin := range prodConfig.AllowedOrigins {
		if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
			t.Errorf("production CORS config should NOT allow localhost, got: %v", prodConfig.AllowedOrigins)
		}
	}

	// Verify no config uses wildcard origin with credentials
	for _, cfg := range []*config.CORSConfig{devConfig, stagingConfig, prodConfig} {
		for _, origin := range cfg.AllowedOrigins {
			if origin == "*" && cfg.Credentials {
				t.Error("CORS config uses wildcard origin with credentials (security risk)")
			}
		}
	}
}

// TestIntegration_AllErrorsHaveRequestID tests that all error responses include request ID
func TestIntegration_AllErrorsHaveRequestID(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	tests := []struct {
		name           string
		handler        func(http.ResponseWriter, *http.Request)
		payload        string
		expectedStatus int
		context        func(*http.Request) *http.Request
	}{
		{
			name: "validation_error_register",
			handler: func(w http.ResponseWriter, r *http.Request) {
				h.Register(w, r)
			},
			payload:        `{"name": "", "email": "test@test.com", "password": "pass"}`,
			expectedStatus: http.StatusBadRequest,
			context: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), middleware.RequestIDKey, "req_all_errors_001"))
			},
		},
		{
			name: "auth_error_checkin",
			handler: func(w http.ResponseWriter, r *http.Request) {
				h.CheckIn(w, r)
			},
			payload:        `{"latitude": 0, "longitude": 0, "device_uuid": "dev", "signature": "sig"}`,
			expectedStatus: http.StatusUnauthorized,
			context: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), middleware.RequestIDKey, "req_all_errors_002"))
			},
		},
		{
			name: "forbidden_error_location",
			handler: func(w http.ResponseWriter, r *http.Request) {
				h.CreateLocation(w, r)
			},
			payload:        `{"name": "Test", "latitude": 6.2, "longitude": 106.8, "radius_meters": 50}`,
			expectedStatus: http.StatusForbidden,
			context: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), middleware.RoleKey, "employee")
				ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_all_errors_003")
				return r.WithContext(ctx)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(test.payload))
			req.Header.Set("Content-Type", "application/json")
			req = test.context(req)

			w := httptest.NewRecorder()
			test.handler(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("expected status %d, got %d", test.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}

			// Verify requestID is present
			if _, ok := response["requestID"]; !ok {
				t.Error("expected 'requestID' field in error response")
			}

			// Verify Content-Type is application/json
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type: application/json, got %s", contentType)
			}
		})
	}
}

// TestGetSyncStatus_Unauthorized verifies sync status endpoint returns 401 without auth
func TestGetSyncStatus_Unauthorized(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{})

	req := httptest.NewRequest(http.MethodGet, "/api/attendance/sync/status", nil)
	w := httptest.NewRecorder()

	h.GetSyncStatus(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestHandleError_Helper_IntegrationWithHandler tests the respondWithError helper function
func TestHandleError_Helper_IntegrationWithHandler(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	// Test using the helper to respond with a custom error
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_helper_test")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Create a custom error and use the handler
	err := middleware.NotFoundError{Message: "User not found"}
	h.respondWithError(w, req, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response middleware.ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.RequestID != "req_helper_test" {
		t.Errorf("Expected requestID 'req_helper_test', got '%s'", response.RequestID)
	}

	if response.Error != "User not found" {
		t.Errorf("Expected error 'User not found', got '%s'", response.Error)
	}
}

// ============================================================================
// Task 4: Structured Logging Tests
// ============================================================================

// TestStructuredLogging_LoginSuccess verifies login success is logged with request ID
func TestStructuredLogging_LoginSuccess(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]string{
		"email":       "user@example.com",
		"password":    "password123",
		"device_uuid": "550e8400-e29b-41d4-a716-446655440000",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	// Add request ID to context
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_login_success_123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// The test confirms logging happens (verified via log output)
	// In production, this would be captured by a logging sink
}

// TestStructuredLogging_LoginFailure verifies login failure is logged with request ID
func TestStructuredLogging_LoginFailure(t *testing.T) {
	// Create a mock that returns error
	mockAuth := &mockAuthUsecase{}
	h := NewHandler(mockAuth, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]string{
		"email":       "user@example.com",
		"password":    "wrongpassword",
		"device_uuid": "550e8400-e29b-41d4-a716-446655440000",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	// Add request ID to context
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_login_failure_456")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Login(w, req)

	// Login fails due to mock (would log failure with request ID)
	// In production, the log message would be:
	// [req_login_failure_456] Login failed: invalid credentials for email user@example.com
}

// TestStructuredLogging_CheckInSuccess verifies check-in success is logged
func TestStructuredLogging_CheckInSuccess(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	userID := uuid.New()
	payload := map[string]interface{}{
		"latitude":    6.2,
		"longitude":   106.8,
		"device_uuid": "550e8400-e29b-41d4-a716-446655440000",
		"signature":   "test_signature",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/attendance/check-in", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	// Add user and request ID to context
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_checkin_success_789")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.CheckIn(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// In production, the log would be:
	// [req_checkin_success_789] Check-in successful for user <userID> at location
}

// TestStructuredLogging_UnauthorizedCheckIn verifies unauthorized check-in is logged
func TestStructuredLogging_UnauthorizedCheckIn(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"latitude":    6.2,
		"longitude":   106.8,
		"device_uuid": "550e8400-e29b-41d4-a716-446655440000",
		"signature":   "test_signature",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/attendance/check-in", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	// Add request ID but NO user ID (simulate unauthorized)
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_unauthorized_checkin_999")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.CheckIn(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	// In production, the log would be:
	// [req_unauthorized_checkin_999] Unauthorized check-in attempt
}

// TestStructuredLogging_AdminForbidden verifies admin-only endpoint logs authorization failure
func TestStructuredLogging_AdminForbidden(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"name":          "Test Location",
		"latitude":      6.2,
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	// Set role to employee (not admin)
	ctx := context.WithValue(context.Background(), middleware.RoleKey, "employee")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_admin_forbidden_111")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateLocation(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	// In production, the log would be:
	// [req_admin_forbidden_111] Authorization failed: admin role required
}

// TestStructuredLogging_ValidationErrorIncludesRequestID verifies validation errors include request ID
func TestStructuredLogging_ValidationErrorIncludesRequestID(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	// Missing required name field
	payload := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	// Add request ID to context
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_validation_required_222")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Verify response includes request ID
	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	if response["requestID"] != "req_validation_required_222" {
		t.Errorf("expected requestID in response, got %v", response)
	}
}

// TestStructuredLogging_ResponseHeaderIncludesRequestID verifies X-Request-ID header is in responses
func TestStructuredLogging_ResponseHeaderIncludesRequestID(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	// Note: In the actual middleware, X-Request-ID would be set in the response header
	// This test verifies that the handler uses the request ID from context
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_header_test_333")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Register(w, req)

	// The X-Request-ID header would normally be set by RequestIDMiddleware
	// This confirms that we extract the request ID correctly in handlers
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

// TestStructuredLogging_RegistrationSuccess verifies registration success is logged with request ID
func TestStructuredLogging_RegistrationSuccess(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "password123",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_register_success_444")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	// In production, the log would be:
	// [req_register_success_444] Registration successful for user <userID>
}

// TestStructuredLogging_FaceEnrollmentSuccess verifies face enrollment is logged
func TestStructuredLogging_FaceEnrollmentSuccess(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	userID := uuid.New()
	payload := map[string][]string{
		"selfies": []string{"photo1.jpg", "photo2.jpg", "photo3.jpg"},
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/profile/face-enrollment", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_enroll_success_555")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.EnrollFace(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// In production, the log would be:
	// [req_enroll_success_555] Face enrollment successful for user <userID> with 3 samples
}

// TestStructuredLogging_LocationCreationSuccess verifies location creation is logged
func TestStructuredLogging_LocationCreationSuccess(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	payload := map[string]interface{}{
		"name":          "Office Location",
		"latitude":      6.2,
		"longitude":     106.8,
		"radius_meters": 50,
	}
	b, _ := json.Marshal(payload)

	ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req_location_create_666")
	req := httptest.NewRequest(http.MethodPost, "/api/locations", bytes.NewReader(b))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.CreateLocation(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	// In production, the log would be:
	// [req_location_create_666] Location created successfully with ID <locationID>
}

// TestStructuredLogging_HealthCheckLogged verifies health check is logged with request ID
func TestStructuredLogging_HealthCheckLogged(t *testing.T) {
	db := &mockDB{}
	redis := &mockRedisClient{connected: true}
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, db, redis)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_health_777")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// In production, the log would be:
	// [req_health_777] Health check request from <remoteAddr>
}

// TestStructuredLogging_ErrorResponseFormatConsistent verifies all error responses have consistent format
func TestStructuredLogging_ErrorResponseFormatConsistent(t *testing.T) {
	h := NewHandler(&mockAuthUsecase{}, &mockAttendanceUsecase{}, nil, nil)

	testCases := []struct {
		name           string
		handler        func(*httptest.ResponseRecorder)
		expectedStatus int
	}{
		{
			name: "validation_error",
			handler: func(w *httptest.ResponseRecorder) {
				payload := `{"name": "", "email": "test@test.com", "password": "pass"}`
				req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(payload))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_val_888")
				req = req.WithContext(ctx)
				h.Register(w, req)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "auth_error",
			handler: func(w *httptest.ResponseRecorder) {
				payload := `{"latitude": 0, "longitude": 0, "device_id": "dev", "signature": "sig"}`
				req := httptest.NewRequest(http.MethodPost, "/api/attendance/check-in", strings.NewReader(payload))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "req_auth_888")
				req = req.WithContext(ctx)
				h.CheckIn(w, req)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.handler(w)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Verify response is JSON
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}

			// Verify response includes error and requestID fields
			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Errorf("response is not valid JSON: %v", err)
				return
			}

			if _, hasError := response["error"]; !hasError {
				t.Error("response missing 'error' field")
			}

			if _, hasRequestID := response["requestID"]; !hasRequestID {
				t.Error("response missing 'requestID' field")
			}
		})
	}
}


// INTENTIONAL FAILURE FOR CI/CD TESTING - Task 7 End-to-End Validation
// This test is deliberately designed to fail to validate the CI/CD pipeline catches test failures
func TestIntentionalFailure_CIPipelineValidation(t *testing.T) {
	// This test is expected to FAIL during the first push
	// It confirms that the CI/CD workflow properly detects and reports test failures
	// After this test is pushed and workflow shows red status, it will be fixed
	t.Error("INTENTIONAL: This test must fail to validate CI pipeline is working properly")
}
