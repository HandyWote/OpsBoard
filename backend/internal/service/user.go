package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/config"
	"backend/internal/domain/user"
	"backend/internal/repository"

	"go.uber.org/zap"
)

// UserService 负责用户资料与权限管理。
type UserService struct {
	repo   repository.UserRepository
	auth   config.AuthConfig
	logger *zap.Logger
}

// NewUserService 构造用户服务。
func NewUserService(authCfg config.AuthConfig, repo repository.UserRepository, log *zap.Logger) *UserService {
	if log == nil {
		log = zap.NewNop()
	}
	return &UserService{repo: repo, auth: authCfg, logger: log}
}

// ProfileUpdateInput 描述可更新的资料字段。
type ProfileUpdateInput struct {
	DisplayName string
	Headline    string
	Bio         string
}

// PasswordChangeInput 描述密码更新请求。
type PasswordChangeInput struct {
	Current string
	New     string
}

// ListUsersInput 控制用户列表查询。
type ListUsersInput struct {
	Keyword  string
	Page     int
	PageSize int
}

// ListUsersResult 返回分页结果。
type ListUsersResult struct {
	Items    []user.User
	Total    int
	Page     int
	PageSize int
}

// GetProfile 返回指定用户的资料。
func (s *UserService) GetProfile(ctx context.Context, id uuid.UUID) (user.User, error) {
	usr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return user.User{}, ErrNotFound
		}
		return user.User{}, err
	}
	return usr, nil
}

// UpdateProfile 更新昵称、签名与自我介绍。
func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, input ProfileUpdateInput) (user.User, error) {
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		return user.User{}, fmt.Errorf("%w: display name required", ErrValidation)
	}
	if utf8.RuneCountInString(displayName) > 40 {
		return user.User{}, fmt.Errorf("%w: display name too long", ErrValidation)
	}

	headline := strings.TrimSpace(input.Headline)
	if utf8.RuneCountInString(headline) > 80 {
		return user.User{}, fmt.Errorf("%w: headline too long", ErrValidation)
	}

	bio := strings.TrimSpace(input.Bio)
	if utf8.RuneCountInString(bio) > 400 {
		return user.User{}, fmt.Errorf("%w: bio too long", ErrValidation)
	}

	usr, err := s.repo.UpdateProfile(ctx, id, displayName, headline, bio)
	if err != nil {
		if err == repository.ErrNotFound {
			return user.User{}, ErrNotFound
		}
		return user.User{}, err
	}
	return usr, nil
}

// ChangePassword 更新当前用户的密码。
func (s *UserService) ChangePassword(ctx context.Context, id uuid.UUID, input PasswordChangeInput) error {
	if strings.TrimSpace(input.Current) == "" || strings.TrimSpace(input.New) == "" {
		return fmt.Errorf("%w: password required", ErrValidation)
	}
	if len(input.New) < 8 {
		return fmt.Errorf("%w: password too short", ErrValidation)
	}

	usr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrNotFound
		}
		return err
	}

	cred, err := s.repo.GetCredential(ctx, usr.Username)
	if err != nil {
		if err == repository.ErrNotFound {
			return ErrUnauthorized
		}
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(input.Current)) != nil {
		return ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.New), s.passwordCost())
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, usr.ID, string(hash), "bcrypt", s.passwordCost())
}

// ListUsers 返回分页用户列表。
func (s *UserService) ListUsers(ctx context.Context, input ListUsersInput) (ListUsersResult, error) {
	pageSize := input.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	page := input.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	items, total, err := s.repo.ListUsers(ctx, strings.TrimSpace(input.Keyword), pageSize, offset)
	if err != nil {
		return ListUsersResult{}, err
	}

	return ListUsersResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ToggleAdmin 切换管理员角色。
func (s *UserService) ToggleAdmin(ctx context.Context, operatorID, targetID uuid.UUID, grant bool) error {
	if operatorID == uuid.Nil {
		return ErrUnauthorized
	}
	if err := s.repo.ToggleRole(ctx, targetID, operatorID, user.RoleAdmin, grant); err != nil {
		return err
	}
	return nil
}

func (s *UserService) passwordCost() int {
	if s.auth.PasswordHashCost >= bcrypt.MinCost && s.auth.PasswordHashCost <= bcrypt.MaxCost {
		return s.auth.PasswordHashCost
	}
	return bcrypt.DefaultCost
}
