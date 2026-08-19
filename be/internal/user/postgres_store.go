package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

const userColumns = `id, username, display_name, avatar_url, password_hash, role, all_houses, active, created_at, updated_at`
const emptyHouseIDs = `, '{}'::text[]`

func (s *PostgresStore) Create(ctx context.Context, item User) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin user create: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	query := `INSERT INTO users (` + userColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING ` + userColumns + emptyHouseIDs
	created, err := scanUser(tx.QueryRow(ctx, query, item.ID, item.Username, item.DisplayName, item.AvatarURL, item.PasswordHash, item.Role, item.AllHouses, item.Active, item.CreatedAt, item.UpdatedAt))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, ErrUsernameExists
	}
	if err != nil {
		return User{}, err
	}
	created.HouseIDs = item.HouseIDs
	if err = replaceHouseScope(ctx, tx, created); err != nil {
		return User{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit user create: %w", err)
	}
	return created, nil
}

func (s *PostgresStore) List(ctx context.Context) ([]User, error) {
	rows, err := s.pool.Query(ctx, `SELECT u.id,u.username,u.display_name,u.avatar_url,u.password_hash,u.role,u.all_houses,u.active,u.created_at,u.updated_at,
		COALESCE(array_agg(hm.house_id ORDER BY hm.house_id) FILTER (WHERE hm.house_id IS NOT NULL), '{}'::text[])
		FROM users u LEFT JOIN spirit_house_members hm ON hm.user_id=u.id
		GROUP BY u.id ORDER BY u.display_name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	items := make([]User, 0)
	for rows.Next() {
		item, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresStore) UpdateAccount(ctx context.Context, item User, passwordHash string) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin account update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `UPDATE users SET username=$2, display_name=$3, role=$4, all_houses=$5, password_hash=COALESCE(NULLIF($6, ''), password_hash), updated_at=$7 WHERE id=$1 RETURNING ` + userColumns + emptyHouseIDs
	updated, err := scanUser(tx.QueryRow(ctx, query, item.ID, item.Username, item.DisplayName, item.Role, item.AllHouses, passwordHash, item.UpdatedAt))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, ErrUsernameExists
	}
	if err != nil {
		return User{}, err
	}
	updated.HouseIDs = item.HouseIDs
	if err := replaceHouseScope(ctx, tx, updated); err != nil {
		return User{}, err
	}
	if passwordHash != "" {
		if _, err := tx.Exec(ctx, `DELETE FROM user_sessions WHERE user_id=$1`, item.ID); err != nil {
			return User{}, fmt.Errorf("revoke sessions after password reset: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit account update: %w", err)
	}
	return updated, nil
}

func (s *PostgresStore) FindByUsername(ctx context.Context, username string) (User, error) {
	return scanUser(s.pool.QueryRow(ctx, `SELECT `+userColumns+emptyHouseIDs+` FROM users WHERE username = $1`, username))
}

func (s *PostgresStore) CreateSession(ctx context.Context, session Session) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO user_sessions (token_hash, user_id, expires_at, created_at) VALUES ($1,$2,$3,$4)`, session.TokenHash, session.UserID, session.ExpiresAt, session.CreatedAt)
	return err
}

func (s *PostgresStore) UserBySession(ctx context.Context, tokenHash string) (User, error) {
	query := `SELECT u.` + strings.ReplaceAll(userColumns, `, `, `, u.`) + emptyHouseIDs + ` FROM users u JOIN user_sessions s ON s.user_id = u.id WHERE s.token_hash = $1 AND s.expires_at > NOW() AND u.active = true`
	return scanUser(s.pool.QueryRow(ctx, query, tokenHash))
}

func (s *PostgresStore) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM user_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (s *PostgresStore) ChangePassword(ctx context.Context, userID, passwordHash string, updatedAt time.Time, keepTokenHash string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password change: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	result, err := tx.Exec(ctx, `UPDATE users SET password_hash=$2, updated_at=$3 WHERE id=$1`, userID, passwordHash, updatedAt)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM user_sessions WHERE user_id=$1 AND token_hash<>$2`, userID, keepTokenHash); err != nil {
		return fmt.Errorf("revoke other sessions: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit password change: %w", err)
	}
	return nil
}

func (s *PostgresStore) UpdateProfile(ctx context.Context, item User) (User, error) {
	query := `UPDATE users SET username=$2, display_name=$3, updated_at=$4 WHERE id=$1 RETURNING ` + userColumns + emptyHouseIDs
	updated, err := scanUser(s.pool.QueryRow(ctx, query, item.ID, item.Username, item.DisplayName, item.UpdatedAt))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, ErrUsernameExists
	}
	return updated, err
}

func (s *PostgresStore) UpdateAvatar(ctx context.Context, item User) (User, error) {
	return scanUser(s.pool.QueryRow(ctx, `UPDATE users SET avatar_url=$2, updated_at=$3 WHERE id=$1 RETURNING `+userColumns+emptyHouseIDs, item.ID, item.AvatarURL, item.UpdatedAt))
}

func replaceHouseScope(ctx context.Context, tx pgx.Tx, item User) error {
	if _, err := tx.Exec(ctx, `DELETE FROM spirit_house_members WHERE user_id=$1`, item.ID); err != nil {
		return fmt.Errorf("clear user house scope: %w", err)
	}
	if item.AllHouses || item.Role == RoleAdmin || len(item.HouseIDs) == 0 {
		return nil
	}
	_, err := tx.Exec(ctx, `INSERT INTO spirit_house_members(user_id,house_id)
		SELECT $1,house_id FROM unnest($2::text[]) AS house_id`, item.ID, item.HouseIDs)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return fmt.Errorf("%w: spirit house does not exist", ErrInvalidInput)
	}
	if err != nil {
		return fmt.Errorf("set user house scope: %w", err)
	}
	return nil
}

type scanner interface{ Scan(dest ...any) error }

func scanUser(row scanner) (User, error) {
	var item User
	if err := row.Scan(&item.ID, &item.Username, &item.DisplayName, &item.AvatarURL, &item.PasswordHash, &item.Role, &item.AllHouses, &item.Active, &item.CreatedAt, &item.UpdatedAt, &item.HouseIDs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	return item, nil
}
