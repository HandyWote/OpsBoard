package transporthttp

import (
	"net/http"
	"strings"
	"time"

	"backend/internal/domain/user"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *Handler) requestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			h.log.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.String("duration", duration.String()),
			)
		})
	}
}

func (h *Handler) authRequired() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r.Header.Get("Authorization"))
			if token == "" {
				respondError(w, http.StatusUnauthorized, "unauthorized", "缺少访问令牌")
				return
			}

			claims, err := h.services.Auth.ParseAccessToken(token)
			if err != nil {
				h.log.Debug("access token invalid", zap.Error(err))
				respondError(w, http.StatusUnauthorized, "unauthorized", "访问令牌无效")
				return
			}

			userID, err := uuidFromString(claims.Subject)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "unauthorized", "访问令牌无效")
				return
			}

			roles := make([]string, len(claims.Roles))
			copy(roles, claims.Roles)

			ctx := WithUser(r.Context(), userID, claims.DisplayName, roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *Handler) adminRequired() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles := CurrentUserRoles(r.Context())
			for _, role := range roles {
				if role == string(user.RoleAdmin) {
					next.ServeHTTP(w, r)
					return
				}
			}
			respondError(w, http.StatusForbidden, "forbidden", "权限不足")
		})
	}
}

func extractBearer(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func uuidFromString(val string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(val))
}
