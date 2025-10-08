package transporthttp

import (
	"net/http"

	"backend/internal/service"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type authResponse struct {
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
	User         userDTO `json:"user"`
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_payload", "请求格式不正确")
		return
	}

	result, err := h.services.Auth.Login(r.Context(), req.Username, req.Password, service.AuthMetadata{
		UserAgent: r.Header.Get("User-Agent"),
		IP:        h.clientIP(r),
	})
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         mapUser(result.User),
	})
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_payload", "请求格式不正确")
		return
	}

	result, err := h.services.Auth.Refresh(r.Context(), req.RefreshToken, service.AuthMetadata{
		UserAgent: r.Header.Get("User-Agent"),
		IP:        h.clientIP(r),
	})
	if err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User:         mapUser(result.User),
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_payload", "请求格式不正确")
		return
	}

	if err := h.services.Auth.Logout(r.Context(), req.RefreshToken); err != nil {
		h.respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "已退出"})
}
