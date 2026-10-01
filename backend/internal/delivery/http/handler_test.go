package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
		"name":           "Test Location",
		"latitude":       6.2,
		"longitude":      106.8,
		"radius_meters":  50,
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
		"name":           "Test Location",
		"latitude":       6.2,
		"longitude":      106.8,
		"radius_meters":  50,
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
		"name":           "Updated Location",
		"latitude":       6.3,
		"longitude":      106.9,
		"radius_meters":  60,
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
		"name":           "Updated Location",
		"latitude":       6.3,
		"longitude":      106.9,
		"radius_meters":  60,
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
		"name":           "Test Location",
		"latitude":       6.2,
		"longitude":      106.8,
		"radius_meters":  50,
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
		"name":           "Test Location",
		"latitude":       6.2,
		"longitude":      106.8,
		"radius_meters":  50,
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
