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
			name:           "unauthorized_checkin",
			method:         "POST",
			path:           "/api/attendance/check-in",
			payload:        `{"latitude": 0, "longitude": 0, "device_id": "dev", "signature": "sig"}`,
			expectedStatus: http.StatusUnauthorized,
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
			name:             "out_of_range_latitude",
			latitude:         100,
			longitude:        106.8,
			shouldFail:       true,
			unexpectedTerms: []string{"schema", "column", "database", "table", "goroutine", "panic"},
		},
		{
			name:             "out_of_range_longitude",
			latitude:         45.0,
			longitude:        200,
			shouldFail:       true,
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
