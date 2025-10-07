package db

import (
	"context"
	"database/sql"
	"fmt"
)

var migrationStatements = []string{
	// 用户主表
	`CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		username TEXT NOT NULL,
		display_name TEXT NOT NULL,
		email TEXT UNIQUE,
		phone TEXT UNIQUE,
		dept_code TEXT,
		avatar_url TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
		last_login_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_unique ON users (LOWER(username));`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (LOWER(email)) WHERE email IS NOT NULL;`,

	// 密码表
	`CREATE TABLE IF NOT EXISTS user_credentials (
		user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		password_hash TEXT NOT NULL,
		hash_algorithm TEXT NOT NULL,
		hash_cost INTEGER NOT NULL,
		password_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`,

	// 第三方身份信息
	`CREATE TABLE IF NOT EXISTS user_identities (
		id BIGSERIAL PRIMARY KEY,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		provider TEXT NOT NULL,
		subject TEXT NOT NULL,
		attrs JSONB NOT NULL DEFAULT '{}'::jsonb,
		last_synced_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (provider, subject)
	);`,
	`CREATE INDEX IF NOT EXISTS idx_user_identities_user ON user_identities (user_id);`,

	// 角色映射
	`CREATE TABLE IF NOT EXISTS user_roles (
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role_key TEXT NOT NULL,
		assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		assigned_by UUID REFERENCES users(id),
		PRIMARY KEY (user_id, role_key)
	);`,

	// 审计日志
	`CREATE TABLE IF NOT EXISTS user_audit_logs (
		id BIGSERIAL PRIMARY KEY,
		user_id UUID REFERENCES users(id) ON DELETE SET NULL,
		action TEXT NOT NULL,
		ip_address INET,
		user_agent TEXT,
		metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`,
	`CREATE INDEX IF NOT EXISTS idx_user_audit_logs_user ON user_audit_logs (user_id);`,
}

// RunMigrations 会依次执行所有迁移语句，保证幂等。
func RunMigrations(ctx context.Context, db *sql.DB) error {
	for _, stmt := range migrationStatements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migration failed for statement %q: %w", stmt, err)
		}
	}
	return nil
}
