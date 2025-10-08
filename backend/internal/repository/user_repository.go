package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/domain/user"
)

// ErrNotFound 表示未找到记录。
var ErrNotFound = errors.New("repository: not found")

// Credential 包含登录凭据字段。
type Credential struct {
	UserID        uuid.UUID
	Username      string
	DisplayName   string
	PasswordHash  string
	HashAlgorithm string
	HashCost      int
	Roles         []user.Role
}

// UserRepository 定义用户与身份相关的数据库操作。
type UserRepository interface {
	GetCredential(ctx context.Context, username string) (Credential, error)
	CreateWithPassword(ctx context.Context, username, passwordHash, algorithm string, cost int) (user.User, error)
	RecordLogin(ctx context.Context, userID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (user.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, displayName, headline, bio string) (user.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash, algorithm string, cost int) error
	ListUsers(ctx context.Context, keyword string, limit, offset int) ([]user.User, int, error)
	ToggleRole(ctx context.Context, targetID, operatorID uuid.UUID, role user.Role, grant bool) error
	GetRoles(ctx context.Context, id uuid.UUID) ([]user.Role, error)
	CreateSession(ctx context.Context, session user.Session) error
	GetSessionByHash(ctx context.Context, hash string) (user.Session, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository 构造用户仓储。
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetCredential(ctx context.Context, username string) (Credential, error) {
	const query = `
SELECT
	u.id,
	u.username,
	u.display_name,
	cred.password_hash,
	cred.hash_algorithm,
	cred.hash_cost,
	COALESCE(
		JSONB_AGG(r.role_key) FILTER (WHERE r.role_key IS NOT NULL),
		'[]'::jsonb
	) AS roles
FROM users u
JOIN user_credentials cred ON cred.user_id = u.id
LEFT JOIN user_roles r ON r.user_id = u.id
WHERE lower(u.username) = lower($1)
GROUP BY
	u.id,
	u.username,
	u.display_name,
	cred.password_hash,
	cred.hash_algorithm,
	cred.hash_cost
LIMIT 1
`

	var cred Credential
	var rolesRaw []byte
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&cred.UserID,
		&cred.Username,
		&cred.DisplayName,
		&cred.PasswordHash,
		&cred.HashAlgorithm,
		&cred.HashCost,
		&rolesRaw,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Credential{}, ErrNotFound
	}
	if err != nil {
		return Credential{}, err
	}

	roles, err := parseRolesJSON(rolesRaw)
	if err != nil {
		return Credential{}, fmt.Errorf("decode credential roles: %w", err)
	}
	cred.Roles = roles

	return cred, nil
}

func (r *userRepository) CreateWithPassword(ctx context.Context, username, passwordHash, algorithm string, cost int) (user.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return user.User{}, err
	}
	defer tx.Rollback()

	const firstUserQuery = `SELECT COUNT(*) = 0 FROM users`
	var isFirstUser bool
	if err := tx.QueryRowContext(ctx, firstUserQuery).Scan(&isFirstUser); err != nil {
		return user.User{}, err
	}

	now := time.Now().UTC()
	id := uuid.New()
	displayName := username

	const insertUser = `
INSERT INTO users (id, username, display_name, email, headline, bio, avatar_url, status, created_at, updated_at)
VALUES ($1, $2, $3, NULL, NULL, NULL, NULL, 'active', $4, $4)
RETURNING id, username, display_name, COALESCE(email, ''), COALESCE(headline, ''), COALESCE(bio, ''), COALESCE(avatar_url, ''), status, last_login_at, created_at, updated_at
`
	var u user.User
	err = tx.QueryRowContext(ctx, insertUser, id, username, displayName, now).Scan(
		&u.ID,
		&u.Username,
		&u.DisplayName,
		&u.Email,
		&u.Headline,
		&u.Bio,
		&u.AvatarURL,
		&u.Status,
		&u.LastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return user.User{}, err
	}

	const insertCredential = `
INSERT INTO user_credentials (user_id, password_hash, hash_algorithm, hash_cost, password_updated_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $5, $5)
`
	if _, err := tx.ExecContext(ctx, insertCredential, u.ID, passwordHash, algorithm, cost, now); err != nil {
		return user.User{}, err
	}

	const assignRole = `
INSERT INTO user_roles (user_id, role_key, assigned_at)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING
`
	rolesToAssign := []user.Role{user.RoleMember}
	if isFirstUser {
		rolesToAssign = append(rolesToAssign, user.RoleAdmin)
	}
	for _, role := range rolesToAssign {
		if _, err := tx.ExecContext(ctx, assignRole, u.ID, string(role), now); err != nil {
			return user.User{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return user.User{}, err
	}

	u.Roles = rolesToAssign
	return u, nil
}

func (r *userRepository) RecordLogin(ctx context.Context, userID uuid.UUID) error {
	const query = `
UPDATE users
SET last_login_at = $2,
	updated_at    = $2
WHERE id = $1
`
	_, err := r.db.ExecContext(ctx, query, userID, time.Now().UTC())
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	const query = `
SELECT
	u.id,
	u.username,
	u.display_name,
	COALESCE(u.email, ''),
	COALESCE(u.headline, ''),
	COALESCE(u.bio, ''),
	COALESCE(u.avatar_url, ''),
	u.status,
	u.last_login_at,
	u.created_at,
	u.updated_at,
	COALESCE(
		(
			SELECT JSONB_AGG(role_key)
			FROM user_roles
			WHERE user_id = u.id
		),
		'[]'::jsonb
	) AS roles
FROM users u
WHERE u.id = $1
`
	var rolesRaw []byte
	var u user.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Username,
		&u.DisplayName,
		&u.Email,
		&u.Headline,
		&u.Bio,
		&u.AvatarURL,
		&u.Status,
		&u.LastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
		&rolesRaw,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}

	roles, err := parseRolesJSON(rolesRaw)
	if err != nil {
		return user.User{}, fmt.Errorf("decode user roles: %w", err)
	}
	u.Roles = roles
	return u, nil
}

func (r *userRepository) UpdateProfile(ctx context.Context, id uuid.UUID, displayName, headline, bio string) (user.User, error) {
	const query = `
UPDATE users
SET display_name = $2,
	headline = $3,
	bio = $4,
	updated_at = $5
WHERE id = $1
RETURNING
	id,
	username,
	display_name,
	COALESCE(email, ''),
	COALESCE(headline, ''),
	COALESCE(bio, ''),
	COALESCE(avatar_url, ''),
	status,
	last_login_at,
	created_at,
	updated_at
`

	now := time.Now().UTC()
	var u user.User
	err := r.db.QueryRowContext(ctx, query, id, displayName, headline, bio, now).Scan(
		&u.ID,
		&u.Username,
		&u.DisplayName,
		&u.Email,
		&u.Headline,
		&u.Bio,
		&u.AvatarURL,
		&u.Status,
		&u.LastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}

	roles, err := r.GetRoles(ctx, u.ID)
	if err != nil {
		return user.User{}, err
	}
	u.Roles = roles
	return u, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash, algorithm string, cost int) error {
	const query = `
UPDATE user_credentials
SET password_hash = $2,
	hash_algorithm = $3,
	hash_cost = $4,
	password_updated_at = $5,
	updated_at = $5
WHERE user_id = $1
`
	_, err := r.db.ExecContext(ctx, query, id, passwordHash, algorithm, cost, time.Now().UTC())
	return err
}

func (r *userRepository) ListUsers(ctx context.Context, keyword string, limit, offset int) ([]user.User, int, error) {
	var args []any
	var conditions []string
	args = append(args, limit, offset)

	baseQuery := `
SELECT
	u.id,
	u.username,
	u.display_name,
	COALESCE(u.email, ''),
	COALESCE(u.headline, ''),
	COALESCE(u.bio, ''),
	COALESCE(u.avatar_url, ''),
	u.status,
	u.last_login_at,
	u.created_at,
	u.updated_at,
	COALESCE(
		JSONB_AGG(r.role_key) FILTER (WHERE r.role_key IS NOT NULL),
		'[]'::jsonb
	) AS roles
FROM users u
LEFT JOIN user_roles r ON r.user_id = u.id
`
	if strings.TrimSpace(keyword) != "" {
		conditions = append(conditions, "(LOWER(u.username) LIKE LOWER($3) OR LOWER(u.display_name) LIKE LOWER($3))")
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(keyword))+"%")
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := baseQuery + where + `
GROUP BY u.id
ORDER BY u.created_at DESC
LIMIT $1 OFFSET $2
`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]user.User, 0)
	for rows.Next() {
		var rolesRaw []byte
		var u user.User
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.DisplayName,
			&u.Email,
			&u.Headline,
			&u.Bio,
			&u.AvatarURL,
			&u.Status,
			&u.LastLoginAt,
			&u.CreatedAt,
			&u.UpdatedAt,
			&rolesRaw,
		); err != nil {
			return nil, 0, err
		}
		roles, err := parseRolesJSON(rolesRaw)
		if err != nil {
			return nil, 0, fmt.Errorf("decode user roles: %w", err)
		}
		u.Roles = roles
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countQuery := `SELECT COUNT(*) FROM users u ` + where
	var total int
	switch len(conditions) {
	case 0:
		if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
			return nil, 0, err
		}
	default:
		if err := r.db.QueryRowContext(ctx, countQuery, args[2:]...).Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	return users, total, nil
}

func (r *userRepository) ToggleRole(ctx context.Context, targetID, operatorID uuid.UUID, role user.Role, grant bool) error {
	if role != user.RoleAdmin && role != user.RoleMember {
		return fmt.Errorf("unsupported role %s", role)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if grant {
		const insert = `
INSERT INTO user_roles (user_id, role_key, assigned_at, assigned_by)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING
`
		if _, err := tx.ExecContext(ctx, insert, targetID, string(role), time.Now().UTC(), operatorID); err != nil {
			return err
		}
	} else {
		if role == user.RoleAdmin {
			const adminCountQuery = `SELECT COUNT(*) FROM user_roles WHERE role_key = 'admin'`
			var count int
			if err := tx.QueryRowContext(ctx, adminCountQuery).Scan(&count); err != nil {
				return err
			}
			if count <= 1 {
				return fmt.Errorf("cannot revoke last admin")
			}
		}
		const delete = `
DELETE FROM user_roles
WHERE user_id = $1 AND role_key = $2
`
		if _, err := tx.ExecContext(ctx, delete, targetID, string(role)); err != nil {
			return err
		}
	}

	const audit = `
INSERT INTO audit_logs (user_id, action, resource, resource_id, metadata, created_at)
VALUES ($1, $2, 'user', $3, $4, $5)
`
	action := "role_revoke"
	if grant {
		action = "role_grant"
	}
	meta := fmt.Sprintf(`{"role":"%s"}`, role)
	if _, err := tx.ExecContext(ctx, audit, operatorID, action, targetID.String(), meta, time.Now().UTC()); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *userRepository) GetRoles(ctx context.Context, id uuid.UUID) ([]user.Role, error) {
	const query = `
SELECT role_key FROM user_roles WHERE user_id = $1
`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]user.Role, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, user.Role(role))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *userRepository) CreateSession(ctx context.Context, session user.Session) error {
	const query = `
INSERT INTO user_sessions (id, user_id, refresh_token_sha, expires_at, user_agent, ip_address, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshTokenSHA,
		session.ExpiresAt,
		session.UserAgent,
		session.IP,
		session.CreatedAt,
	)
	return err
}

func (r *userRepository) GetSessionByHash(ctx context.Context, hash string) (user.Session, error) {
	const query = `
SELECT id, user_id, refresh_token_sha, expires_at, user_agent, ip_address, revoked_at, created_at
FROM user_sessions
WHERE refresh_token_sha = $1
LIMIT 1
`
	var session user.Session
	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenSHA,
		&session.ExpiresAt,
		&session.UserAgent,
		&session.IP,
		&session.RevokedAt,
		&session.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return user.Session{}, ErrNotFound
	}
	return session, err
}

func (r *userRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	const query = `
UPDATE user_sessions
SET revoked_at = $2
WHERE id = $1
`
	_, err := r.db.ExecContext(ctx, query, sessionID, time.Now().UTC())
	return err
}

func parseRolesJSON(raw []byte) ([]user.Role, error) {
	if len(raw) == 0 {
		return []user.Role{}, nil
	}

	var roleKeys []string
	if err := json.Unmarshal(raw, &roleKeys); err != nil {
		return nil, err
	}

	roles := make([]user.Role, 0, len(roleKeys))
	for _, key := range roleKeys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		roles = append(roles, user.Role(key))
	}
	return roles, nil
}
