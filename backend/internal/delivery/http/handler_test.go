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
