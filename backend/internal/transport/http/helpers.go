package transporthttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/repository"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func (h *Handler) respondServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		respondError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		respondError(w, http.StatusUnauthorized, "invalid_credentials", "用户名或密码错误")
	case errors.Is(err, service.ErrUnauthorized):
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
	case errors.Is(err, service.ErrForbidden):
		respondError(w, http.StatusForbidden, "forbidden", "权限不足")
	case errors.Is(err, service.ErrNotFound), errors.Is(err, repository.ErrNotFound):
		respondError(w, http.StatusNotFound, "not_found", "资源不存在")
	default:
		respondError(w, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
		h.log.Error("service error", zap.Error(err))
	}
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	raw := chi.URLParam(r, name)
	return uuid.Parse(strings.TrimSpace(raw))
}

func queryInt(r *http.Request, key string, fallback int) int {
	val := strings.TrimSpace(r.URL.Query().Get(key))
	if val == "" {
		return fallback
	}
	if num, err := strconv.Atoi(val); err == nil {
		return num
	}
	return fallback
}
