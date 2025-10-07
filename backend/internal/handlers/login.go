package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"backend/internal/storage"
)

const bcryptCost = 10

// LoginHandler handles authentication requests.
type LoginHandler struct {
	store *storage.FileStore
}

// NewLoginHandler constructs a handler backed by the provided store.
func NewLoginHandler(store *storage.FileStore) *LoginHandler {
	return &LoginHandler{
		store: store,
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ServeHTTP validates credentials and manages sign-in or user creation.
func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, loginResponse{
			Success: false,
			Message: "只支持 POST 请求",
		})
		return
	}

	var payload loginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, loginResponse{
			Success: false,
			Message: "请求格式不正确",
		})
		return
	}

	payload.Username = strings.TrimSpace(payload.Username)

	if payload.Username == "" || payload.Password == "" {
		writeJSON(w, http.StatusBadRequest, loginResponse{
			Success: false,
			Message: "用户名和密码均为必填项",
		})
		return
	}

	storedHash, err := h.store.Get(payload.Username)
	switch {
	case err == nil:
		if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(payload.Password)) != nil {
			writeJSON(w, http.StatusUnauthorized, loginResponse{
				Success: false,
				Message: "用户名或密码错误",
			})
			return
		}

		writeJSON(w, http.StatusOK, loginResponse{
			Success: true,
			Message: "登录成功",
		})

	case errors.Is(err, storage.ErrUserNotFound):
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(payload.Password), bcryptCost)
		if hashErr != nil {
			writeJSON(w, http.StatusInternalServerError, loginResponse{
				Success: false,
				Message: "无法处理您的请求，请稍后重试",
			})
			return
		}

		if saveErr := h.store.Upsert(payload.Username, string(hash)); saveErr != nil {
			writeJSON(w, http.StatusInternalServerError, loginResponse{
				Success: false,
				Message: "无法保存用户信息",
			})
			return
		}

		writeJSON(w, http.StatusOK, loginResponse{
			Success: true,
			Message: "账户已创建并登录成功",
		})

	default:
		writeJSON(w, http.StatusInternalServerError, loginResponse{
			Success: false,
			Message: "服务暂时不可用",
		})
	}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
}

func writeJSON(w http.ResponseWriter, status int, payload loginResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
