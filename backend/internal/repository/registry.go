package repository

import (
	"database/sql"
)

// Registry 聚合仓储接口实例。
type Registry struct {
	User UserRepository
	Task TaskRepository
}

// NewRegistry 根据数据库连接创建仓储实例。
func NewRegistry(db *sql.DB) Registry {
	return Registry{
		User: NewUserRepository(db),
		Task: NewTaskRepository(db),
	}
}
