package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/PresensiGo/backend/internal/model"
)

type AttendanceRepository struct {
	db *sql.DB
}

func NewAttendanceRepository(db *sql.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) FindNearestLocation(lat, lng float64) (*model.Location, error) {
	loc := &model.Location{}
	query := `
		SELECT id, name, address, latitude, longitude, radius_meters, created_at, updated_at
		FROM locations
		ORDER BY geom <-> ST_SetSRID(ST_MakePoint($1, $2), 4326)
		LIMIT 1`

	err := r.db.QueryRow(query, lng, lat).Scan(
		&loc.ID, &loc.Name, &loc.Address, &loc.Latitude,
		&loc.Longitude, &loc.RadiusMeters, &loc.CreatedAt, &loc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return loc, nil
}

func (r *AttendanceRepository) CheckGeofence(locationID uuid.UUID, lat, lng float64) (bool, error) {
	var inside bool
	query := `
		SELECT ST_DWithin(
			geom::geography,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
			(SELECT radius_meters FROM locations WHERE id = $3)
		)`

	err := r.db.QueryRow(query, lng, lat, locationID).Scan(&inside)
	if err != nil {
		return false, err
	}
	return inside, nil
}

func (r *AttendanceRepository) CreateCheckIn(att *model.Attendance) error {
	query := `
		INSERT INTO attendances (
			id, user_id, location_id, check_in_time, check_in_location,
			selfie_url, status, is_late, device_uuid, hmac_signature, synced
		) VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326), $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at`

	return r.db.QueryRow(query,
		att.ID, att.UserID, att.LocationID, att.CheckInTime,
		att.CheckInLocation[1], att.CheckInLocation[0], // lat, lng reversed for PostGIS
		att.SelfieURL, att.Status, att.IsLate,
		att.DeviceUUID, att.HMACSignature, att.Synced,
	).Scan(&att.CreatedAt, &att.UpdatedAt)
}

func (r *AttendanceRepository) CreateCheckOut(att *model.Attendance) error {
	query := `
		UPDATE attendances 
		SET check_out_time = $1, 
			check_out_location = ST_SetSRID(ST_MakePoint($2, $3), 4326),
			updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at`

	return r.db.QueryRow(query,
		att.CheckOutTime,
		att.CheckOutLocation[1], att.CheckOutLocation[0],
		att.ID,
	).Scan(&att.UpdatedAt)
}

func (r *AttendanceRepository) FindTodayByUser(userID uuid.UUID) (*model.Attendance, error) {
	att := &model.Attendance{}
	query := `
		SELECT id, user_id, location_id, check_in_time, check_out_time,
			CASE WHEN check_in_location IS NOT NULL THEN 
				ARRAY[ST_Y(check_in_location), ST_X(check_in_location)]
			ELSE NULL END,
			CASE WHEN check_out_location IS NOT NULL THEN 
				ARRAY[ST_Y(check_out_location), ST_X(check_out_location)]
			ELSE NULL END,
			selfie_url, status, is_late, notes, device_uuid, hmac_signature, synced, created_at, updated_at
		FROM attendances 
		WHERE user_id = $1 AND DATE(check_in_time) = CURRENT_DATE
		ORDER BY created_at DESC
		LIMIT 1`

	err := r.db.QueryRow(query, userID).Scan(
		&att.ID, &att.UserID, &att.LocationID, &att.CheckInTime, &att.CheckOutTime,
		&att.CheckInLocation, &att.CheckOutLocation,
		&att.SelfieURL, &att.Status, &att.IsLate, &att.Notes,
		&att.DeviceUUID, &att.HMACSignature, &att.Synced,
		&att.CreatedAt, &att.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return att, nil
}

func (r *AttendanceRepository) GetHistory(userID uuid.UUID, limit, offset int) ([]model.AttendanceResponse, error) {
	query := `
		SELECT a.id, a.user_id, a.location_id, a.check_in_time, a.check_out_time,
			CASE WHEN a.check_in_location IS NOT NULL THEN 
				ARRAY[ST_Y(a.check_in_location), ST_X(a.check_in_location)]
			ELSE NULL END,
			CASE WHEN a.check_out_location IS NOT NULL THEN 
				ARRAY[ST_Y(a.check_out_location), ST_X(a.check_out_location)]
			ELSE NULL END,
			a.selfie_url, a.status, a.is_late, a.notes, a.device_uuid, a.hmac_signature, a.synced, 
			a.created_at, a.updated_at, u.name, l.name
		FROM attendances a
		JOIN users u ON a.user_id = u.id
		JOIN locations l ON a.location_id = l.id
		WHERE a.user_id = $1
		ORDER BY a.check_in_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.AttendanceResponse
	for rows.Next() {
		var att model.AttendanceResponse
		err := rows.Scan(
			&att.ID, &att.UserID, &att.LocationID, &att.CheckInTime, &att.CheckOutTime,
			&att.CheckInLocation, &att.CheckOutLocation,
			&att.SelfieURL, &att.Status, &att.IsLate, &att.Notes,
			&att.DeviceUUID, &att.HMACSignature, &att.Synced,
			&att.CreatedAt, &att.UpdatedAt, &att.UserName, &att.LocationName,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, att)
	}
	return results, nil
}

func (r *AttendanceRepository) CreateLocation(loc *model.Location) error {
	query := `
		INSERT INTO locations (id, name, address, latitude, longitude, radius_meters, geom)
		VALUES ($1, $2, $3, $4, $5, $6, ST_SetSRID(ST_MakePoint($4, $5), 4326))
		RETURNING created_at, updated_at`

	return r.db.QueryRow(query,
		loc.ID, loc.Name, loc.Address, loc.Latitude,
		loc.Longitude, loc.RadiusMeters,
	).Scan(&loc.CreatedAt, &loc.UpdatedAt)
}

func (r *AttendanceRepository) GetLocations() ([]model.Location, error) {
	query := `
		SELECT id, name, address, latitude, longitude, radius_meters, created_at, updated_at
		FROM locations
		ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []model.Location
	for rows.Next() {
		var loc model.Location
		err := rows.Scan(
			&loc.ID, &loc.Name, &loc.Address, &loc.Latitude,
			&loc.Longitude, &loc.RadiusMeters, &loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	return locations, nil
}
