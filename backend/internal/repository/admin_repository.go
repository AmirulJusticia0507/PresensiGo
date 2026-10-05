package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/PresensiGo/backend/internal/model"
	"github.com/google/uuid"
)

type AdminRepository struct{ db *sql.DB }

func NewAdminRepository(db *sql.DB) *AdminRepository { return &AdminRepository{db: db} }

func attendanceWhere(filter model.AttendanceFilter) (string, []any) {
	parts := []string{"1=1"}
	args := make([]any, 0, 5)
	add := func(clause string, value any) {
		args = append(args, value)
		parts = append(parts, fmt.Sprintf(clause, len(args)))
	}
	if filter.UserID != nil {
		add("a.user_id = $%d", *filter.UserID)
	}
	if filter.DateFrom != nil {
		add("a.check_in_time >= $%d", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		add("a.check_in_time < $%d", filter.DateTo.Add(24*time.Hour))
	}
	if filter.Status != "" {
		add("a.status = $%d", filter.Status)
	}
	if filter.LocationID != nil {
		add("a.location_id = $%d", *filter.LocationID)
	}
	return strings.Join(parts, " AND "), args
}

func (r *AdminRepository) ListAttendances(ctx context.Context, filter model.AttendanceFilter) (*model.PaginatedAttendances, error) {
	where, args := attendanceWhere(filter)
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM attendances a WHERE "+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	query := `SELECT a.id, a.user_id, u.name, a.location_id, l.name, a.check_in_time, a.check_out_time,
		a.status, a.is_late, a.notes FROM attendances a JOIN users u ON u.id=a.user_id
		JOIN locations l ON l.id=a.location_id WHERE ` + where +
		fmt.Sprintf(" ORDER BY a.check_in_time DESC NULLS LAST LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, filter.Limit, filter.Offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]model.AttendanceListItem, 0)
	for rows.Next() {
		var item model.AttendanceListItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.LocationID, &item.LocationName,
			&item.CheckInTime, &item.CheckOutTime, &item.Status, &item.IsLate, &item.Notes); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &model.PaginatedAttendances{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset}, rows.Err()
}

func (r *AdminRepository) ListUsers(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,email,role,device_uuid,phone,emergency_contact_name,
		emergency_contact_phone,address,profile_picture_url,terms_accepted_at,face_enrolled_at,created_at,updated_at
		FROM users ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	users := make([]model.User, 0)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.DeviceUUID, &u.Phone, &u.EmergencyContactName,
			&u.EmergencyContactPhone, &u.Address, &u.ProfilePictureUrl, &u.TermsAcceptedAt, &u.FaceEnrolledAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *AdminRepository) UpdateUser(ctx context.Context, id uuid.UUID, req model.AdminUserUpdate) error {
	if req.Role != nil {
		if _, err := r.db.ExecContext(ctx, `UPDATE users SET role=$1, updated_at=NOW() WHERE id=$2`, *req.Role, id); err != nil {
			return err
		}
	}
	if req.ResetDevice {
		_, err := r.db.ExecContext(ctx, `UPDATE users SET device_uuid=NULL, updated_at=NOW() WHERE id=$1`, id)
		return err
	}
	return nil
}

func (r *AdminRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1 AND role <> 'admin'`, id)
	return err
}

func (r *AdminRepository) UpsertSchedule(ctx context.Context, s *model.WorkSchedule) error {
	if s.ID != uuid.Nil {
		result, err := r.db.ExecContext(ctx, `UPDATE work_schedules SET location_id=$1,user_id=$2,
			day_of_week=$3,start_time=$4,end_time=$5,late_after_minutes=$6,active=$7,updated_at=NOW()
			WHERE id=$8`, s.LocationID, s.UserID, s.DayOfWeek, s.StartTime, s.EndTime,
			s.LateAfterMinute, s.Active, s.ID)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if count == 0 {
			return sql.ErrNoRows
		}
		return nil
	}
	s.ID = uuid.New()
	conflict := `(user_id,day_of_week) WHERE user_id IS NOT NULL`
	if s.UserID == nil {
		conflict = `(location_id,day_of_week) WHERE user_id IS NULL`
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO work_schedules
		(id,location_id,user_id,day_of_week,start_time,end_time,late_after_minutes,active)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT `+conflict+` DO UPDATE SET
		location_id=EXCLUDED.location_id,start_time=EXCLUDED.start_time,end_time=EXCLUDED.end_time,
		late_after_minutes=EXCLUDED.late_after_minutes,active=EXCLUDED.active,updated_at=NOW()`,
		s.ID, s.LocationID, s.UserID, s.DayOfWeek, s.StartTime, s.EndTime, s.LateAfterMinute, s.Active)
	return err
}

func (r *AdminRepository) ListSchedules(ctx context.Context) ([]model.WorkSchedule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,location_id,user_id,day_of_week,start_time::text,end_time::text,late_after_minutes,active FROM work_schedules ORDER BY day_of_week`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]model.WorkSchedule, 0)
	for rows.Next() {
		var s model.WorkSchedule
		if err := rows.Scan(&s.ID, &s.LocationID, &s.UserID, &s.DayOfWeek, &s.StartTime, &s.EndTime, &s.LateAfterMinute, &s.Active); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *AdminRepository) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM work_schedules WHERE id=$1`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *AdminRepository) CreateLeave(ctx context.Context, userID uuid.UUID, req model.CreateLeaveRequest) error {
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	if end.Before(start) {
		return fmt.Errorf("end_date must not be before start_date")
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO leave_requests(id,user_id,start_date,end_date,type,reason) VALUES($1,$2,$3,$4,$5,$6)`, uuid.New(), userID, start, end, req.Type, req.Reason)
	return err
}

func (r *AdminRepository) ListLeaves(ctx context.Context, userID *uuid.UUID) ([]model.LeaveRequest, error) {
	query := `SELECT lr.id,lr.user_id,u.name,lr.start_date,lr.end_date,lr.type,lr.reason,lr.status,lr.reviewed_by,lr.created_at FROM leave_requests lr JOIN users u ON u.id=lr.user_id`
	args := []any{}
	if userID != nil {
		query += " WHERE lr.user_id=$1"
		args = append(args, *userID)
	}
	query += " ORDER BY lr.created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]model.LeaveRequest, 0)
	for rows.Next() {
		var item model.LeaveRequest
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.StartDate, &item.EndDate, &item.Type, &item.Reason, &item.Status, &item.ReviewedBy, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AdminRepository) ReviewLeave(ctx context.Context, id, reviewer uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE leave_requests SET status=$1,reviewed_by=$2,reviewed_at=NOW() WHERE id=$3 AND status='pending'`, status, reviewer, id)
	return err
}
