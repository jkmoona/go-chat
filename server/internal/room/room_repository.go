package room

import (
	"context"
	"database/sql"
	"time"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type repository struct {
	db DBTX
}

func NewRepository(db DBTX) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, room *Room) error {
	query := `INSERT INTO rooms (id, name, creator_id, pin_hash, ttl_minutes, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	var creatorID interface{} = room.CreatorID
	if room.CreatorID == 0 {
		creatorID = nil
	}
	_, err := r.db.ExecContext(ctx, query,
		room.ID, room.Name, creatorID, room.PinHash, room.TTL, room.ExpiresAt,
	)
	return err
}

func (r *repository) GetByID(ctx context.Context, id string) (*Room, error) {
	query := `SELECT id, name, creator_id, pin_hash, ttl_minutes, created_at, expires_at, is_active
		FROM rooms WHERE id = $1 AND is_active = true`

	var room Room
	var creatorID sql.NullInt64
	var pinHash sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&room.ID, &room.Name, &creatorID, &pinHash,
		&room.TTL, &room.CreatedAt, &room.ExpiresAt, &room.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}
	room.CreatorID = creatorID.Int64
	if pinHash.Valid {
		room.PinHash = pinHash.String
	}
	return &room, nil
}

func (r *repository) ListActive(ctx context.Context) ([]*Room, error) {
	query := `SELECT id, name, creator_id, pin_hash, ttl_minutes, created_at, expires_at, is_active
		FROM rooms WHERE is_active = true AND expires_at > NOW()
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*Room
	for rows.Next() {
		var room Room
		var creatorID sql.NullInt64
		var pinHash sql.NullString
		if err := rows.Scan(
			&room.ID, &room.Name, &creatorID, &pinHash,
			&room.TTL, &room.CreatedAt, &room.ExpiresAt, &room.IsActive,
		); err != nil {
			return nil, err
		}
		room.CreatorID = creatorID.Int64
		if pinHash.Valid {
			room.PinHash = pinHash.String
		}
		rooms = append(rooms, &room)
	}
	return rooms, rows.Err()
}

func (r *repository) Deactivate(ctx context.Context, id string) error {
	query := `UPDATE rooms SET is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *repository) UpdateExpiresAt(ctx context.Context, id string, expiresAt time.Time) error {
	query := `UPDATE rooms SET expires_at = $1 WHERE id = $2 AND is_active = true`
	_, err := r.db.ExecContext(ctx, query, expiresAt, id)
	return err
}
