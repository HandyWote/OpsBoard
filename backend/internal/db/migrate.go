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
		headline TEXT,
		bio TEXT,
		avatar_url TEXT,
		status TEXT NOT NULL DEFAULT 'active',
		last_login_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT chk_users_status CHECK (status IN ('active','disabled'))
	);`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_unique ON users (LOWER(username));`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (LOWER(email)) WHERE email IS NOT NULL;`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS headline TEXT;`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT;`,
	`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;`,

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

	// 角色映射
	`CREATE TABLE IF NOT EXISTS user_roles (
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role_key TEXT NOT NULL,
		assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		assigned_by UUID REFERENCES users(id),
		PRIMARY KEY (user_id, role_key),
		CONSTRAINT chk_user_role_key CHECK (role_key IN ('member','admin'))
	);`,

	// 刷新 Token / 会话
	`CREATE TABLE IF NOT EXISTS user_sessions (
		id UUID PRIMARY KEY,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		refresh_token_sha TEXT NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		user_agent TEXT,
		ip_address INET,
		revoked_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`,
	`CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions (user_id);`,
	`CREATE INDEX IF NOT EXISTS idx_user_sessions_refresh ON user_sessions (refresh_token_sha);`,

	// 审计日志
	`CREATE TABLE IF NOT EXISTS audit_logs (
		id BIGSERIAL PRIMARY KEY,
		user_id UUID REFERENCES users(id) ON DELETE SET NULL,
		action TEXT NOT NULL,
		resource TEXT,
		resource_id TEXT,
		metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
		ip_address INET,
		user_agent TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`,
	`CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs (user_id);`,

	// 任务表
	`CREATE TABLE IF NOT EXISTS tasks (
		id UUID PRIMARY KEY,
		title TEXT NOT NULL,
		description_html TEXT NOT NULL,
		description_plain TEXT NOT NULL,
		bounty BIGINT NOT NULL DEFAULT 0,
		priority TEXT NOT NULL,
		status TEXT NOT NULL,
		deadline TIMESTAMPTZ,
		created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
		published_by UUID REFERENCES users(id) ON DELETE SET NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT chk_tasks_priority CHECK (priority IN ('critical','high','medium','low')),
		CONSTRAINT chk_tasks_status CHECK (status IN ('draft','available','claimed','completed','archived'))
	);`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_deadline ON tasks (deadline);`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_created_by ON tasks (created_by);`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks (created_at DESC);`,
	`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks (deleted_at);`,

	// 标签
	`CREATE TABLE IF NOT EXISTS task_tags (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`,

	`CREATE TABLE IF NOT EXISTS task_tag_map (
		task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		tag_id INTEGER NOT NULL REFERENCES task_tags(id) ON DELETE CASCADE,
		PRIMARY KEY (task_id, tag_id)
	);`,

	// 任务领取记录
	`CREATE TABLE IF NOT EXISTS task_assignments (
		id BIGSERIAL PRIMARY KEY,
		task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		status TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		completed_at TIMESTAMPTZ,
		released_at TIMESTAMPTZ,
	CONSTRAINT chk_task_assign_status CHECK (status IN ('claimed','completed','released'))
	);`,
	`CREATE INDEX IF NOT EXISTS idx_task_assign_task ON task_assignments (task_id);`,
	`CREATE INDEX IF NOT EXISTS idx_task_assign_user ON task_assignments (user_id);`,

	// 任务全文检索索引（使用 simple 配置，兼容默认 PostgreSQL）
	`CREATE INDEX IF NOT EXISTS idx_tasks_search ON tasks USING GIN (to_tsvector('simple', title || ' ' || description_plain));`,

	// 扩展任务与领取记录的状态约束，支持提交待验收流程
	`ALTER TABLE tasks DROP CONSTRAINT IF EXISTS chk_tasks_status;`,
	`ALTER TABLE tasks ADD CONSTRAINT chk_tasks_status CHECK (status IN ('draft','available','claimed','submitted','completed','archived'));`,
	`ALTER TABLE task_assignments DROP CONSTRAINT IF EXISTS chk_task_assign_status;`,
	`ALTER TABLE task_assignments ADD CONSTRAINT chk_task_assign_status CHECK (status IN ('claimed','submitted','completed','released'));`,
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
