package http

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PresensiGo/backend/internal/delivery/http/middleware"
	"github.com/PresensiGo/backend/internal/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) registerAdminRoutes(r *mux.Router) {
	r.HandleFunc("/api/attendance/history/page", h.GetHistoryPage).Methods("GET")
	r.HandleFunc("/api/leaves", h.CreateLeave).Methods("POST")
	r.HandleFunc("/api/leaves", h.ListMyLeaves).Methods("GET")
	r.HandleFunc("/api/admin/attendances", h.AdminAttendances).Methods("GET")
	r.HandleFunc("/api/admin/attendances.csv", h.ExportAttendancesCSV).Methods("GET")
	r.HandleFunc("/api/admin/users", h.AdminUsers).Methods("GET")
	r.HandleFunc("/api/admin/users/{id}", h.AdminUpdateUser).Methods("PATCH")
	r.HandleFunc("/api/admin/users/{id}", h.AdminDeleteUser).Methods("DELETE")
	r.HandleFunc("/api/admin/schedules", h.AdminSchedules).Methods("GET", "POST")
	r.HandleFunc("/api/admin/schedules/{id}", h.AdminDeleteSchedule).Methods("DELETE")
	r.HandleFunc("/api/admin/leaves", h.AdminLeaves).Methods("GET")
	r.HandleFunc("/api/admin/leaves/{id}", h.AdminReviewLeave).Methods("PATCH")
}

func (h *Handler) needAdminRepo(w http.ResponseWriter, r *http.Request, admin bool) bool {
	if admin && !requireAdmin(w, r) {
		return false
	}
	if h.adminRepo == nil {
		respondError(w, http.StatusServiceUnavailable, "admin service unavailable")
		return false
	}
	return true
}

func parseAttendanceFilter(r *http.Request) (model.AttendanceFilter, error) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	f := model.AttendanceFilter{Limit: limit, Offset: offset, Status: q.Get("status")}
	if f.Status != "" && f.Status != "present" && f.Status != "late" && f.Status != "absent" && f.Status != "leave" {
		return f, fmt.Errorf("invalid status")
	}
	for key, dst := range map[string]**time.Time{"date_from": &f.DateFrom, "date_to": &f.DateTo} {
		if raw := q.Get(key); raw != "" {
			v, err := time.Parse("2006-01-02", raw)
			if err != nil {
				return f, fmt.Errorf("invalid %s", key)
			}
			*dst = &v
		}
	}
	if raw := q.Get("location_id"); raw != "" {
		v, err := uuid.Parse(raw)
		if err != nil {
			return f, fmt.Errorf("invalid location_id")
		}
		f.LocationID = &v
	}
	return f, nil
}

func (h *Handler) GetHistoryPage(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, false) {
		return
	}
	uid := middleware.GetUserIDFromContext(r.Context())
	if uid == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	f, err := parseAttendanceFilter(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	f.UserID = &uid
	result, err := h.adminRepo.ListAttendances(r.Context(), f)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get history")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *Handler) AdminAttendances(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, true) {
		return
	}
	f, err := parseAttendanceFilter(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.adminRepo.ListAttendances(r.Context(), f)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get attendances")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (h *Handler) ExportAttendancesCSV(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, true) {
		return
	}
	f, err := parseAttendanceFilter(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	f.Limit = 10000
	f.Offset = 0
	result, err := h.adminRepo.ListAttendances(r.Context(), f)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to export attendances")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="attendances.csv"`)
	writer := csv.NewWriter(w)
	defer writer.Flush()
	_ = writer.Write([]string{"id", "user", "location", "check_in", "check_out", "status", "is_late"})
	for _, a := range result.Items {
		in, out := "", ""
		if a.CheckInTime != nil {
			in = a.CheckInTime.Format(time.RFC3339)
		}
		if a.CheckOutTime != nil {
			out = a.CheckOutTime.Format(time.RFC3339)
		}
		_ = writer.Write([]string{a.ID.String(), a.UserName, a.LocationName, in, out, a.Status, strconv.FormatBool(a.IsLate)})
	}
}

func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, true) {
		return
	}
	items, err := h.adminRepo.ListUsers(r.Context())
	if err != nil {
		respondError(w, 500, "failed to get users")
		return
	}
	respondJSON(w, 200, items)
}
func (h *Handler) AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, true) {
		return
	}
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, 400, "invalid user ID")
		return
	}
	var req model.AdminUserUpdate
	if validateRequest(w, r, &req) != nil {
		return
	}
	if h.adminRepo.UpdateUser(r.Context(), id, req) != nil {
		respondError(w, 500, "failed to update user")
		return
	}
	respondJSON(w, 200, map[string]string{"message": "user updated"})
}
func (h *Handler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, true) {
		return
	}
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, 400, "invalid user ID")
		return
	}
	if id == middleware.GetUserIDFromContext(r.Context()) {
		respondError(w, 400, "cannot delete your own account")
		return
	}
	if h.adminRepo.DeleteUser(r.Context(), id) != nil {
		respondError(w, 500, "failed to delete user")
		return
	}
	respondJSON(w, 200, map[string]string{"message": "user deleted"})
}

func (h *Handler) AdminSchedules(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, true) {
		return
	}
	if r.Method == http.MethodGet {
		items, err := h.adminRepo.ListSchedules(r.Context())
		if err != nil {
			respondError(w, 500, "failed to get schedules")
			return
		}
		respondJSON(w, 200, items)
		return
	}
	var req model.WorkSchedule
	if validateRequest(w, r, &req) != nil {
		return
	}
	if req.LocationID == nil && req.UserID == nil {
		respondError(w, 400, "location_id or user_id is required")
		return
	}
	if h.adminRepo.UpsertSchedule(r.Context(), &req) != nil {
		respondError(w, 500, "failed to save schedule")
		return
	}
	respondJSON(w, 201, req)
}

func (h *Handler) AdminDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, true) {
		return
	}
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid schedule ID")
		return
	}
	if err := h.adminRepo.DeleteSchedule(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, "schedule not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "schedule deleted"})
}

func (h *Handler) CreateLeave(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, false) {
		return
	}
	uid := middleware.GetUserIDFromContext(r.Context())
	if uid == uuid.Nil {
		respondError(w, 401, "unauthorized")
		return
	}
	var req model.CreateLeaveRequest
	if validateRequest(w, r, &req) != nil {
		return
	}
	if err := h.adminRepo.CreateLeave(r.Context(), uid, req); err != nil {
		respondError(w, 400, err.Error())
		return
	}
	respondJSON(w, 201, map[string]string{"message": "leave request submitted"})
}
func (h *Handler) ListMyLeaves(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, false) {
		return
	}
	uid := middleware.GetUserIDFromContext(r.Context())
	if uid == uuid.Nil {
		respondError(w, 401, "unauthorized")
		return
	}
	items, err := h.adminRepo.ListLeaves(r.Context(), &uid)
	if err != nil {
		respondError(w, 500, "failed to get leave requests")
		return
	}
	respondJSON(w, 200, items)
}
func (h *Handler) AdminLeaves(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, true) {
		return
	}
	items, err := h.adminRepo.ListLeaves(r.Context(), nil)
	if err != nil {
		respondError(w, 500, "failed to get leave requests")
		return
	}
	respondJSON(w, 200, items)
}
func (h *Handler) AdminReviewLeave(w http.ResponseWriter, r *http.Request) {
	if !h.needAdminRepo(w, r, true) {
		return
	}
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, 400, "invalid leave ID")
		return
	}
	var req model.ReviewLeaveRequest
	if validateRequest(w, r, &req) != nil {
		return
	}
	reviewer := middleware.GetUserIDFromContext(r.Context())
	if h.adminRepo.ReviewLeave(r.Context(), id, reviewer, req.Status) != nil {
		respondError(w, 500, "failed to review leave")
		return
	}
	respondJSON(w, 200, map[string]string{"message": "leave reviewed"})
}
