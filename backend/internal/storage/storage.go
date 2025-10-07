package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrUserNotFound 表示指定的用户不存在。
var ErrUserNotFound = errors.New("user not found")

// Credential 描述了登录凭据信息。
type Credential struct {
	UserID        uuid.UUID
	Username      string
	PasswordHash  string
	HashAlgorithm string
	HashCost      int
}

// CredentialStore 定义了登录凭据的读写接口。
type CredentialStore interface {
	GetCredential(ctx context.Context, username string) (Credential, error)
	CreateUserWithPassword(ctx context.Context, username, passwordHash, algorithm string, cost int) (Credential, error)
	RecordSuccessfulLogin(ctx context.Context, userID uuid.UUID) error
}
