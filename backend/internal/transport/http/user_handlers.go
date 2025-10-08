package transporthttp

import (
	"net/http"

	"backend/internal/service"
)

type updateProfileRequest struct {
	DisplayName string `json:"displayName"`
	Headline    string `json:"headline"`
	Bio         string `json:"bio"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type toggleAdminRequest struct {
	Grant bool `json:"grant"`
}

func (h *Handler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	profile, err := h.services.Users.GetProfile(r.Context(), userID)
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapUser(profile))
}

func (h *Handler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	var req updateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_payload", "请求格式不正确")
		return
	}

	updated, err := h.services.Users.UpdateProfile(r.Context(), userID, service.ProfileUpdateInput{
		DisplayName: req.DisplayName,
		Headline:    req.Headline,
		Bio:         req.Bio,
	})
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, mapUser(updated))
}

func (h *Handler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	var req changePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_payload", "请求格式不正确")
		return
	}

	if err := h.services.Users.ChangePassword(r.Context(), userID, service.PasswordChangeInput{
		Current: req.CurrentPassword,
		New:     req.NewPassword,
	}); err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "密码已更新"})
}

func (h *Handler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 20)

	result, err := h.services.Users.ListUsers(r.Context(), service.ListUsersInput{
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	users := make([]userDTO, 0, len(result.Items))
	for _, item := range result.Items {
		users = append(users, mapUser(item))
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"items":    users,
		"total":    result.Total,
		"page":     result.Page,
		"pageSize": result.PageSize,
	})
}

func (h *Handler) handleToggleAdmin(w http.ResponseWriter, r *http.Request) {
	operatorID, ok := CurrentUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "未授权访问")
		return
	}

	targetID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_id", "用户 ID 不合法")
		return
	}

	var req toggleAdminRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_payload", "请求格式不正确")
		return
	}

	if err := h.services.Users.ToggleAdmin(r.Context(), operatorID, targetID, req.Grant); err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"grant": req.Grant})
}
