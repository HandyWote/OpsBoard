package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PostgresStore 基于 PostgreSQL 的用户凭据存储实现。
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore 构造存储实例。
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// GetCredential 根据用户名读取凭据。
func (p *PostgresStore) GetCredential(ctx context.Context, username string) (Credential, error) {
	query := `
SELECT
	u.id,
	u.username,
	cred.password_hash,
	cred.hash_algorithm,
	cred.hash_cost
FROM users AS u
JOIN user_credentials AS cred ON cred.user_id = u.id
WHERE lower(u.username) = lower($1)
`

	var cred Credential
	err := p.db.QueryRowContext(ctx, query, username).Scan(
		&cred.UserID,
		&cred.Username,
		&cred.PasswordHash,
		&cred.HashAlgorithm,
		&cred.HashCost,
	)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Credential{}, ErrUserNotFound
	case err != nil:
		return Credential{}, err
	default:
		return cred, nil
	}
}

// CreateUserWithPassword 创建新用户并保存密码哈希。
func (p *PostgresStore) CreateUserWithPassword(ctx context.Context, username, passwordHash, algorithm string, cost int) (Credential, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return Credential{}, errors.New("username cannot be empty")
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return Credential{}, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now().UTC()
	userID := uuid.New()

	insertUser := `
INSERT INTO users (
	id,
	username,
	display_name,
	status,
	metadata,
	created_at,
	updated_at
) VALUES ($1, $2, $3, 'active', '{}'::jsonb, $4, $4)
`

	if _, err := tx.ExecContext(ctx, insertUser, userID, username, username, now); err != nil {
		return Credential{}, err
	}

	insertCredential := `
INSERT INTO user_credentials (
	user_id,
	password_hash,
	hash_algorithm,
	hash_cost,
	password_updated_at,
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $5, $5)
`

	if _, err := tx.ExecContext(ctx, insertCredential, userID, passwordHash, algorithm, cost, now); err != nil {
		return Credential{}, err
	}

	if err := tx.Commit(); err != nil {
		return Credential{}, err
	}

	return Credential{
		UserID:        userID,
		Username:      username,
		PasswordHash:  passwordHash,
		HashAlgorithm: algorithm,
		HashCost:      cost,
	}, nil
}

// RecordSuccessfulLogin 记录用户成功登录时间。
func (p *PostgresStore) RecordSuccessfulLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
UPDATE users
SET last_login_at = $2,
    updated_at    = $2
WHERE id = $1
`
	_, err := p.db.ExecContext(ctx, query, userID, time.Now().UTC())
	return err
}
