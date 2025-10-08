package transporthttp

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	contextKeyUserID   contextKey = "user_id"
	contextKeyUserName contextKey = "user_name"
	contextKeyRoles    contextKey = "roles"
)

// WithUser 注入当前用户信息。
func WithUser(ctx context.Context, id uuid.UUID, name string, roles []string) context.Context {
	ctx = context.WithValue(ctx, contextKeyUserID, id)
	ctx = context.WithValue(ctx, contextKeyUserName, name)
	ctx = context.WithValue(ctx, contextKeyRoles, roles)
	return ctx
}

// CurrentUserID 返回请求绑定的用户 ID。
func CurrentUserID(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(contextKeyUserID)
	if id, ok := val.(uuid.UUID); ok {
		return id, true
	}
	return uuid.Nil, false
}

// CurrentUserRoles 返回当前用户角色列表。
func CurrentUserRoles(ctx context.Context) []string {
	val := ctx.Value(contextKeyRoles)
	if roles, ok := val.([]string); ok {
		return roles
	}
	return nil
}
