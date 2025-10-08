package user

import (
	"time"

	"github.com/google/uuid"
)

// Role 定义系统角色。
type Role string

const (
	RoleMember Role = "member"
	RoleAdmin  Role = "admin"
)

// User 描述用户主信息。
type User struct {
	ID          uuid.UUID
	Username    string
	DisplayName string
	Email       string
	Headline    string
	Bio         string
	AvatarURL   string
	Status      string
	LastLoginAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Roles       []Role
}

// Session 记录刷新 Token。
type Session struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	RefreshTokenSHA string
	ExpiresAt       time.Time
	UserAgent       string
	IP              string
	RevokedAt       *time.Time
	CreatedAt       time.Time
}
