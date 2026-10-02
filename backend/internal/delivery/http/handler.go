package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/PresensiGo/backend/internal/delivery/http/middleware"
	"github.com/PresensiGo/backend/internal/model"
	"github.com/PresensiGo/backend/internal/repository"
	"github.com/PresensiGo/backend/internal/usecase"
)

// AuthUsecaseIface defines the auth operations required by the HTTP handler.
type AuthUsecaseIface interface {
	Register(req *model.RegisterRequest) (*model.User, error)
	Login(req *model.LoginRequest) (*model.LoginResponse, error)
	GetByID(id uuid.UUID) (*model.User, error)
	UpdateFaceEmbedding(userID uuid.UUID, embedding []byte) error
	EnrollFace(ctx context.Context, userID uuid.UUID, selfies []string) error
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
	GetFaceChallenge(userID uuid.UUID) (*model.FaceChallengeResponse, error)
	Sync(userID uuid.UUID, req *model.SyncRequest) []model.SyncResult
	GetSyncStatus(userID uuid.UUID) (*model.SyncStatusResponse, error)
}

// ProfileUsecaseIface defines the profile operations required by the HTTP handler.
type ProfileUsecaseIface interface {
	RegisterUser(ctx context.Context, req *usecase.RegisterUserRequest) (*model.User, string, error)
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*model.User, error)
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, req *usecase.UpdateProfileRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req *usecase.ChangePasswordRequest) error
}

// RedisClientIface defines the Redis operations required by the HTTP handler.
type RedisClientIface interface {
	IsConnected(ctx context.Context) bool
	Close() error
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
}

type DBPinger interface {
	Ping() error
}

type Handler struct {
	authUc      AuthUsecaseIface
	attUc       AttendanceUsecaseIface
	profileUc   ProfileUsecaseIface
	db          DBPinger
	redisClient RedisClientIface
	fraudRepo   *repository.FraudAttemptRepository
	adminRepo   *repository.AdminRepository
}

func NewHandler(authUc AuthUsecaseIface, attUc AttendanceUsecaseIface, profileUc ProfileUsecaseIface, dependencies ...any) *Handler {
	handler := &Handler{authUc: authUc, attUc: attUc, profileUc: profileUc}
	if len(dependencies) > 0 {
		handler.db, _ = dependencies[0].(DBPinger)
	}
	if len(dependencies) > 1 {
		handler.redisClient, _ = dependencies[1].(RedisClientIface)
	}
	if len(dependencies) > 2 {
		handler.fraudRepo, _ = dependencies[2].(*repository.FraudAttemptRepository)
	}
	if len(dependencies) > 3 {
		handler.adminRepo, _ = dependencies[3].(*repository.AdminRepository)
	}
	return handler
}

var validate = validator.New()

func validateRequest(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		requestID := middleware.GetRequestID(r.Context())
		log.Printf("[%s] Failed to decode request body: %v", requestID, err)
		respondValidationError(w, r.Context(), "Invalid request format", nil)
		return err
	}
	if err := validate.Struct(dst); err != nil {
		requestID := middleware.GetRequestID(r.Context())
		details := middleware.SanitizeValidationErrors(err)
		log.Printf("[%s] Validation failed: %v", requestID, err)
		respondValidationError(w, r.Context(), "Validation failed", details)
		return err
	}
	return nil
}

// respondValidationError sends a sanitized validation error response
func respondValidationError(w http.ResponseWriter, ctx context.Context, errorMsg string, details []string) {
	requestID := middleware.GetRequestID(ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]interface{}{
		"error":     errorMsg,
		"requestID": requestID,
	}

	if len(details) > 0 {
		response["details"] = details
	}

	json.NewEncoder(w).Encode(response)
}

// Health returns liveness check - simple ok response
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log.Printf("[%s] Health check request from %s", requestID, r.RemoteAddr)
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HealthReady returns readiness check - verifies Redis and database connectivity
func (h *Handler) HealthReady(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	log.Printf("[%s] Health readiness check request from %s", requestID, r.RemoteAddr)

	// Check database connectivity
	dbOK := h.db.Ping() == nil

	// Check Redis connectivity using context
	ctx := r.Context()
	redisOK := h.redisClient.IsConnected(ctx)

	// Both must be OK for readiness
	ready := dbOK && redisOK

	if ready {
		log.Printf("[%s] Health readiness: OK (DB: %v, Redis: %v)", requestID, dbOK, redisOK)
		respondJSON(w, http.StatusOK, map[string]bool{"ready": true})
	} else {
		// Return 503 Service Unavailable if not ready
		log.Printf("[%s] Health readiness: NOT READY (DB: %v, Redis: %v)", requestID, dbOK, redisOK)
		w.WriteHeader(http.StatusServiceUnavailable)
		respondJSON(w, http.StatusServiceUnavailable, map[string]bool{"ready": false})
	}
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
	r.HandleFunc("/api/attendance/sync", h.SyncAttendance).Methods("POST")
	r.HandleFunc("/api/attendance/sync/status", h.GetSyncStatus).Methods("GET")

	// Location routes
	r.HandleFunc("/api/locations", h.GetLocations).Methods("GET")
	r.HandleFunc("/api/locations", h.CreateLocation).Methods("POST")
	r.HandleFunc("/api/locations/{id}", h.UpdateLocation).Methods("PUT")
	r.HandleFunc("/api/locations/{id}", h.DeleteLocation).Methods("DELETE")

	// User profile route
	r.HandleFunc("/api/profile", h.GetProfile).Methods("GET")
	r.HandleFunc("/api/profile/face-enrollment", h.EnrollFace).Methods("POST")
	r.HandleFunc("/api/face/challenge", h.GetFaceChallenge).Methods("POST")
	r.HandleFunc("/api/security/location-attempts", h.ReportFraudAttempt).Methods("POST")
	r.HandleFunc("/api/admin/security/location-alerts", h.GetFraudAlerts).Methods("GET")
	h.registerAdminRoutes(r)
}

const repeatedFraudThreshold = 3

func (h *Handler) ReportFraudAttempt(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if h.fraudRepo == nil {
		respondError(w, http.StatusServiceUnavailable, "fraud reporting unavailable")
		return
	}
	var req model.FraudAttemptRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}
	attempt, err := h.fraudRepo.Create(userID, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to record attempt")
		return
	}
	count, err := h.fraudRepo.CountRecent(userID)
	if err != nil {
		log.Printf("failed to count fraud attempts for user %s: %v", userID, err)
	}
	if count >= repeatedFraudThreshold {
		log.Printf("SECURITY ALERT: user %s has %d rejected location attempts in 24 hours", userID, count)
	}
	respondJSON(w, http.StatusCreated, map[string]interface{}{"attempt": attempt, "alerted": count >= repeatedFraudThreshold})
}

func (h *Handler) GetFraudAlerts(w http.ResponseWriter, r *http.Request) {
	if middleware.GetRoleFromContext(r.Context()) != "admin" {
		respondError(w, http.StatusForbidden, "admin access required")
		return
	}
	if h.fraudRepo == nil {
		respondError(w, http.StatusServiceUnavailable, "fraud alerts unavailable")
		return
	}
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}
	alerts, err := h.fraudRepo.ListAlerts(limit, repeatedFraudThreshold)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load fraud alerts")
		return
	}
	respondJSON(w, http.StatusOK, alerts)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	var req usecase.RegisterUserRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}

	user, token, err := h.profileUc.RegisterUser(r.Context(), &req)
	if err != nil {
		log.Printf("[%s] Registration failed for email %s: %v", requestID, req.Email, err)

		switch err.Error() {
		case usecase.ErrEmailAlreadyExists:
			respondError(w, http.StatusConflict, err.Error(), requestID)
		case usecase.ErrWeakPassword,
			usecase.ErrPasswordMismatch,
			usecase.ErrInvalidEmailFormat,
			usecase.ErrInvalidPhone,
			usecase.ErrInvalidName,
			usecase.ErrTermsNotAccepted:
			respondValidationError(w, r.Context(), err.Error(), nil)
		default:
			respondError(w, http.StatusInternalServerError, "failed to register user", requestID)
		}
		return
	}

	log.Printf("[%s] Registration successful for user %s", requestID, user.ID)
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "registration successful",
		"token":   token,
		"user":    user,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	var req model.LoginRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}

	resp, err := h.authUc.Login(&req)
	if err != nil {
		log.Printf("[%s] Login failed: invalid credentials for email %s", requestID, req.Email)
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	log.Printf("[%s] Login successful for user %s", requestID, resp.User.ID)
	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized check-in attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized", requestID)
		return
	}

	// A 1 MB image grows by roughly 33% when base64 encoded.
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	var req model.CheckInRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}

	att, err := h.attUc.CheckIn(userID, &req)
	if err != nil {
		log.Printf("[%s] Check-in failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[%s] Check-in successful for user %s at location", requestID, userID)
	respondJSON(w, http.StatusOK, att)
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized check-out attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized", requestID)
		return
	}

	var req model.CheckOutRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}

	att, err := h.attUc.CheckOut(userID, &req)
	if err != nil {
		log.Printf("[%s] Check-out failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[%s] Check-out successful for user %s", requestID, userID)
	respondJSON(w, http.StatusOK, att)
}

func (h *Handler) SyncAttendance(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 25<<20)
	var req model.SyncRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"results": h.attUc.Sync(userID, &req)})
}

func (h *Handler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized get-sync-status attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	status, err := h.attUc.GetSyncStatus(userID)
	if err != nil {
		log.Printf("[%s] GetSyncStatus failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusInternalServerError, "failed to get sync status")
		return
	}

	respondJSON(w, http.StatusOK, status)
}

func (h *Handler) GetTodayAttendance(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized get-today-attendance attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	att, err := h.attUc.GetTodayAttendance(userID)
	if err != nil {
		log.Printf("[%s] GetTodayAttendance failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusNotFound, "no attendance record today")
		return
	}

	respondJSON(w, http.StatusOK, att)
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized get-history attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	history, err := h.attUc.GetHistory(userID, limit, offset)
	if err != nil {
		log.Printf("[%s] GetHistory failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	respondJSON(w, http.StatusOK, history)
}

func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	locations, err := h.attUc.GetLocations()
	if err != nil {
		log.Printf("[%s] GetLocations failed: %v", requestID, err)
		respondError(w, http.StatusInternalServerError, "failed to get locations")
		return
	}

	respondJSON(w, http.StatusOK, locations)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized get-profile attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.profileUc.GetUserProfile(r.Context(), userID)
	if err != nil {
		log.Printf("[%s] GetProfile failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":                        user.ID,
		"name":                      user.Name,
		"email":                     user.Email,
		"role":                      user.Role,
		"device_uuid":               user.DeviceUUID,
		"phone":                     user.Phone,
		"emergency_contact_name":    user.EmergencyContactName,
		"emergency_contact_phone":   user.EmergencyContactPhone,
		"address":                   user.Address,
		"profile_picture_url":       user.ProfilePictureUrl,
		"face_enrolled":             len(user.FaceEmbedding) > 0,
		"face_enrolled_at":          user.FaceEnrolledAt,
		"face_similarity_threshold": user.FaceSimilarityThreshold,
		"created_at":                user.CreatedAt,
		"updated_at":                user.UpdatedAt,
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized update-profile attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized", requestID)
		return
	}

	var req usecase.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[%s] Failed to decode update-profile request: %v", requestID, err)
		respondValidationError(w, r.Context(), "Invalid request format", nil)
		return
	}

	// Call ProfileUsecase to handle update
	user, err := h.profileUc.UpdateUserProfile(r.Context(), userID, &req)
	if err != nil {
		log.Printf("[%s] UpdateProfile failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[%s] Profile updated successfully for user %s", requestID, userID)
	respondJSON(w, http.StatusOK, user)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized change-password attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized", requestID)
		return
	}

	var req usecase.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[%s] Failed to decode change-password request: %v", requestID, err)
		respondValidationError(w, r.Context(), "Invalid request format", nil)
		return
	}

	// Call ProfileUsecase to handle password change
	if err := h.profileUc.ChangePassword(r.Context(), userID, &req); err != nil {
		log.Printf("[%s] ChangePassword failed for user %s: %v", requestID, userID, err)
		// Use generic error message (don't reveal password-specific details)
		respondError(w, http.StatusBadRequest, "Invalid credentials")
		return
	}

	log.Printf("[%s] Password changed successfully for user %s", requestID, userID)
	respondJSON(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}

func (h *Handler) EnrollFace(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized face-enrollment attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	var req model.EnrollFaceRequest
	if err := validateRequest(w, r, &req); err != nil {
		return
	}
	if err := h.authUc.EnrollFace(r.Context(), userID, req.Selfies); err != nil {
		log.Printf("[%s] Face enrollment failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	log.Printf("[%s] Face enrollment successful for user %s with %d samples", requestID, userID, len(req.Selfies))
	respondJSON(w, http.StatusOK, map[string]any{"message": "face enrollment completed", "samples": len(req.Selfies)})
}

func (h *Handler) GetFaceChallenge(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized face-challenge attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	challenge, err := h.attUc.GetFaceChallenge(userID)
	if err != nil {
		log.Printf("[%s] GetFaceChallenge failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, challenge)
}

func (h *Handler) UpdateFaceEmbedding(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	userID := getUserIDFromContext(r)
	if userID == uuid.Nil {
		log.Printf("[%s] Unauthorized update-face-embedding attempt", requestID)
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.UpdateFaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[%s] Failed to decode update-face-embedding request: %v", requestID, err)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.authUc.UpdateFaceEmbedding(userID, req.FaceEmbedding); err != nil {
		log.Printf("[%s] UpdateFaceEmbedding failed for user %s: %v", requestID, userID, err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[%s] Face embedding updated successfully for user %s", requestID, userID)
	respondJSON(w, http.StatusOK, map[string]string{"message": "face embedding updated"})
}

func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	if !requireAdmin(w, r) {
		log.Printf("[%s] Unauthorized location-creation attempt (non-admin user)", requestID)
		return
	}

	var input model.CreateLocationRequest
	if err := validateRequest(w, r, &input); err != nil {
		return
	}

	req := model.Location{
		ID:           uuid.New(),
		Name:         input.Name,
		Address:      input.Address,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		RadiusMeters: input.RadiusMeters,
	}

	if err := h.attUc.CreateLocation(&req); err != nil {
		log.Printf("[%s] CreateLocation failed: %v", requestID, err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[%s] Location created successfully with ID %s", requestID, req.ID)
	respondJSON(w, http.StatusCreated, req)
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	if !requireAdmin(w, r) {
		log.Printf("[%s] Unauthorized location-update attempt (non-admin user)", requestID)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		log.Printf("[%s] Failed to parse location ID: %v", requestID, err)
		respondError(w, http.StatusBadRequest, "invalid location ID")
		return
	}

	var req model.Location
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[%s] Failed to decode update-location request: %v", requestID, err)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.attUc.UpdateLocation(id, &req); err != nil {
		log.Printf("[%s] UpdateLocation failed for ID %s: %v", requestID, id, err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[%s] Location updated successfully with ID %s", requestID, id)
	respondJSON(w, http.StatusOK, req)
}

func (h *Handler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	if !requireAdmin(w, r) {
		log.Printf("[%s] Unauthorized location-deletion attempt (non-admin user)", requestID)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		log.Printf("[%s] Failed to parse location ID: %v", requestID, err)
		respondError(w, http.StatusBadRequest, "invalid location ID")
		return
	}

	err = h.attUc.DeleteLocation(id)
	if err != nil {
		log.Printf("[%s] DeleteLocation failed for ID %s: %v", requestID, id, err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("[%s] Location deleted successfully with ID %s", requestID, id)
	respondJSON(w, http.StatusOK, map[string]string{"message": "location deleted"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
func respondError(w http.ResponseWriter, status int, message string, requestIDs ...string) {
	response := map[string]string{"error": message}
	if len(requestIDs) > 0 {
		response["requestID"] = requestIDs[0]
	}
	respondJSON(w, status, response)
}

// respondWithError uses the centralized error handler for consistent error responses
func (h *Handler) respondWithError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := middleware.GetRequestID(r.Context())
	middleware.HandleError(w, err, requestID)
}

func getUserIDFromContext(r *http.Request) uuid.UUID {
	return middleware.GetUserIDFromContext(r.Context())
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	requestID := middleware.GetRequestID(r.Context())
	role := middleware.GetRoleFromContext(r.Context())
	if role != "admin" {
		log.Printf("[%s] Authorization failed: admin role required", requestID)
		respondError(w, http.StatusForbidden, "admin role required", requestID)
		return false
	}
	return true
}
