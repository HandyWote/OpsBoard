package service

import (
	"backend/internal/config"
	"backend/internal/repository"

	"go.uber.org/zap"
)

// Registry 汇总所有业务服务。
type Registry struct {
	Auth  *AuthService
	Users *UserService
	Tasks *TaskService
}

// NewRegistry 初始化服务依赖。
func NewRegistry(cfg config.Config, repos repository.Registry, log *zap.Logger) Registry {
	userService := NewUserService(cfg.Auth, repos.User, log)
	authService := NewAuthService(cfg.Auth, repos.User, log)
	taskService := NewTaskService(repos.Task, log)

	return Registry{
		Auth:  authService,
		Users: userService,
		Tasks: taskService,
	}
}
