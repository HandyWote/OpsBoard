package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/config"
	"backend/internal/domain/user"
	"backend/internal/repository"

	"go.uber.org/zap"
)

// AuthMetadata 捕获发起请求的客户端信息。
type AuthMetadata struct {
	UserAgent string
	IP        string
}

// AuthResult 是登录或刷新后返回给客户端的响应数据。
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         user.User
}

// AccessTokenClaims 声明访问令牌的载荷。
type AccessTokenClaims struct {
	Roles       []string `json:"roles"`
	DisplayName string   `json:"name"`
	jwt.RegisteredClaims
}

// AuthService 提供身份认证与令牌签发逻辑。
type AuthService struct {
	cfg  config.AuthConfig
	repo repository.UserRepository
	log  *zap.Logger
}

// NewAuthService 构造身份服务实例。
func NewAuthService(cfg config.AuthConfig, repo repository.UserRepository, log *zap.Logger) *AuthService {
	if log == nil {
		log = zap.NewNop()
	}
	return &AuthService{cfg: cfg, repo: repo, log: log}
}

// Login 校验凭据并签发令牌。如配置允许，在首次登录时自动创建用户。
func (s *AuthService) Login(ctx context.Context, username, password string, meta AuthMetadata) (AuthResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return AuthResult{}, fmt.Errorf("%w: username and password are required", ErrValidation)
	}

	cred, err := s.repo.GetCredential(ctx, username)
	switch {
	case err == nil:
		// continue
	case errors.Is(err, repository.ErrNotFound) && s.cfg.AllowAutoUserCreation:
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(password), s.passwordCost())
		if hashErr != nil {
			return AuthResult{}, fmt.Errorf("hash password: %w", hashErr)
		}
		var created user.User
		created, err = s.repo.CreateWithPassword(ctx, username, string(hash), "bcrypt", s.passwordCost())
		if err != nil {
			return AuthResult{}, fmt.Errorf("create user: %w", err)
		}
		cred, err = s.repo.GetCredential(ctx, username)
		if err != nil {
			return AuthResult{}, fmt.Errorf("reload credential: %w", err)
		}
		s.log.Info("auto created user via login", zap.String("username", username), zap.String("user_id", created.ID.String()))
	default:
		return AuthResult{}, fmt.Errorf("lookup credential: %w", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(password)) != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	if err := s.repo.RecordLogin(ctx, cred.UserID); err != nil {
		s.log.Warn("record login failed", zap.String("user_id", cred.UserID.String()), zap.Error(err))
	}

	currentUser, err := s.repo.GetByID(ctx, cred.UserID)
	if err != nil {
		return AuthResult{}, fmt.Errorf("load user: %w", err)
	}

	return s.issueTokens(ctx, currentUser, meta)
}

// Refresh 根据刷新令牌换取新的一对令牌。
func (s *AuthService) Refresh(ctx context.Context, refreshToken string, meta AuthMetadata) (AuthResult, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return AuthResult{}, fmt.Errorf("%w: refresh token missing", ErrValidation)
	}

	hash := s.hashRefreshToken(refreshToken)
	session, err := s.repo.GetSessionByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return AuthResult{}, ErrUnauthorized
		}
		return AuthResult{}, fmt.Errorf("load session: %w", err)
	}

	if session.RevokedAt != nil {
		return AuthResult{}, ErrUnauthorized
	}
	if time.Now().After(session.ExpiresAt) {
		return AuthResult{}, ErrUnauthorized
	}

	u, err := s.repo.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return AuthResult{}, ErrUnauthorized
		}
		return AuthResult{}, fmt.Errorf("load user: %w", err)
	}

	if err := s.repo.RevokeSession(ctx, session.ID); err != nil {
		s.log.Warn("revoke session failed", zap.String("session_id", session.ID.String()), zap.Error(err))
	}

	return s.issueTokens(ctx, u, meta)
}

// Logout 撤销刷新令牌。
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return fmt.Errorf("%w: refresh token missing", ErrValidation)
	}

	hash := s.hashRefreshToken(refreshToken)
	session, err := s.repo.GetSessionByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("load session: %w", err)
	}
	return s.repo.RevokeSession(ctx, session.ID)
}

// ParseAccessToken 解析访问令牌并返回声明。
func (s *AuthService) ParseAccessToken(token string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, ErrUnauthorized
	}
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrUnauthorized
	}
	if claims.Subject == "" {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

func (s *AuthService) issueTokens(ctx context.Context, usr user.User, meta AuthMetadata) (AuthResult, error) {
	accessToken, err := s.signAccessToken(usr)
	if err != nil {
		return AuthResult{}, err
	}

	refreshToken, session, err := s.generateRefreshToken(usr, meta)
	if err != nil {
		return AuthResult{}, err
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return AuthResult{}, fmt.Errorf("create session: %w", err)
	}

	return AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         usr,
	}, nil
}

func (s *AuthService) signAccessToken(usr user.User) (string, error) {
	now := time.Now().UTC()
	roles := make([]string, 0, len(usr.Roles))
	for _, role := range usr.Roles {
		roles = append(roles, string(role))
	}

	claims := AccessTokenClaims{
		Roles:       roles,
		DisplayName: usr.DisplayName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   usr.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (s *AuthService) generateRefreshToken(usr user.User, meta AuthMetadata) (string, user.Session, error) {
	tokenBytes := make([]byte, 48)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", user.Session{}, fmt.Errorf("generate refresh token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(tokenBytes)

	now := time.Now().UTC()
	session := user.Session{
		ID:              uuid.New(),
		UserID:          usr.ID,
		RefreshTokenSHA: s.hashRefreshToken(raw),
		ExpiresAt:       now.Add(s.cfg.RefreshTokenTTL),
		UserAgent:       meta.UserAgent,
		IP:              meta.IP,
		CreatedAt:       now,
	}

	return raw, session, nil
}

func (s *AuthService) hashRefreshToken(token string) string {
	key := s.cfg.RefreshTokenHashKey
	sum := sha256.Sum256([]byte(key + token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (s *AuthService) passwordCost() int {
	if s.cfg.PasswordHashCost >= bcrypt.MinCost && s.cfg.PasswordHashCost <= bcrypt.MaxCost {
		return s.cfg.PasswordHashCost
	}
	return bcrypt.DefaultCost
}
