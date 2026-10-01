package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

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
