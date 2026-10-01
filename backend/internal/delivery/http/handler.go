package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/PresensiGo/backend/internal/delivery/http/middleware"
	"github.com/PresensiGo/backend/internal/model"
)

// AuthUsecaseIface defines the auth operations required by the HTTP handler.
type AuthUsecaseIface interface {
	Register(req *model.RegisterRequest) (*model.User, error)
	Login(req *model.LoginRequest) (*model.LoginResponse, error)
	GetByID(id uuid.UUID) (*model.User, error)
	UpdateFaceEmbedding(userID uuid.UUID, embedding []byte) error
}

// AttendanceUsecaseIface defines the attendance operations required by the HTTP handler.
type AttendanceUsecaseIface interface {
	CheckIn(userID uuid.UUID, req *model.CheckInRequest) (*model.Attendance, error)
	CheckOut(userID uuid.UUID, req *model.CheckOutRequest) (*model.Attendance, error)
	GetTodayAttendance(userID uuid.UUID) (*model.Attendance, error)
	GetHistory(userID uuid.UUID, limit, offset int) ([]model.AttendanceResponse, error)
	GetLocations() ([]model.Location, error)
	CreateLocation(req *model.Location) error
	UpdateLocation(id uuid.UUID, req *model.Location) error
	DeleteLocation(id uuid.UUID) error
}

type Handler struct {
	authUc AuthUsecaseIface
	attUc  AttendanceUsecaseIface
}

func NewHandler(authUc AuthUsecaseIface, attUc AttendanceUsecaseIface) *Handler {
	return &Handler{
		authUc: authUc,
		attUc:  attUc,
	}
}

var validate = validator.New()

func validateRequest(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return err
	}
	if err := validate.Struct(dst); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return err
	}
	return nil
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Auth routes (public)
	r.HandleFunc("/api/auth/register", h.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", h.Login).Methods("POST")

	// Attendance routes (protected)
	r.HandleFunc("/api/attendance/check-in", h.CheckIn).Methods("POST")
	r.HandleFunc("/api/attendance/check-out", h.CheckOut).Methods("POST")
	r.HandleFunc("/api/attendance/today", h.GetTodayAttendance).Methods("GET")
	r.HandleFunc("/api/attendance/history", h.GetHistory).Methods("GET")

	// Location routes
	r.HandleFunc("/api/locations", h.GetLocations).Methods("GET")
	r.HandleFunc("/api/locations", h.CreateLocation).Methods("POST")
	r.HandleFunc("/api/locations/{id}", h.UpdateLocation).Methods("PUT")
	r.HandleFunc("/api/locations/{id}", h.DeleteLocation).Methods("DELETE")

	// User profile route
	r.HandleFunc("/api/profile", h.GetProfile).Methods("GET")
	r.HandleFunc("/api/profile/face-embedding", h.UpdateFaceEmbedding).Methods("PUT")
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}

	user, err := h.authUc.Register(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "registration successful",
		"user":    user,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}

	resp, err := h.authUc.Login(&req)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.CheckInRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}

	att, err := h.attUc.CheckIn(userID, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, att)
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.CheckOutRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}

	att, err := h.attUc.CheckOut(userID, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, att)
}

func (h *Handler) GetTodayAttendance(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	att, err := h.attUc.GetTodayAttendance(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "no attendance record today")
		return
	}

	respondJSON(w, http.StatusOK, att)
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	history, err := h.attUc.GetHistory(userID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	respondJSON(w, http.StatusOK, history)
}

func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	locations, err := h.attUc.GetLocations()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get locations")
		return
	}

	respondJSON(w, http.StatusOK, locations)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.authUc.GetByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateFaceEmbedding(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.UpdateFaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.authUc.UpdateFaceEmbedding(userID, req.FaceEmbedding); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "face embedding updated"})
}

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var req model.Location
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Generate server-side UUID
	req.ID = uuid.New()

	if err := h.attUc.CreateLocation(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, req)
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid location ID")
		return
	}

	var req model.Location
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.attUc.UpdateLocation(id, &req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, req)
}

func (h *Handler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid location ID")
		return
	}

	err = h.attUc.DeleteLocation(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "location deleted"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func getUserIDFromContext(r *http.Request) uuid.UUID {
	return middleware.GetUserIDFromContext(r.Context())
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	role := middleware.GetRoleFromContext(r.Context())
	if role != "admin" {
		respondError(w, http.StatusForbidden, "admin role required")
		return false
	}
	return true
}
