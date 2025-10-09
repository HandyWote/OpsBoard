package transporthttp

import (
	"net/http"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// Handler 负责 HTTP 层 orchestrate。
type Handler struct {
	cfg      config.Config
	services service.Registry
	log      *zap.Logger
}

// NewRouter 构建完整的 HTTP 路由。
func NewRouter(cfg config.Config, services service.Registry, log *zap.Logger) http.Handler {
	h := &Handler{cfg: cfg, services: services, log: log}

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(h.requestLogger())

	corsOpts := cors.Options{
		AllowedOrigins:   cfg.Server.AllowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	if len(corsOpts.AllowedOrigins) == 0 {
		corsOpts.AllowedOrigins = []string{"*"}
	}
	r.Use(cors.Handler(corsOpts))

	r.Get("/healthz", h.handleHealth)

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/login", h.handleLogin)
		api.Post("/auth/refresh", h.handleRefresh)
		api.Post("/auth/logout", h.handleLogout)

		api.Group(func(priv chi.Router) {
			priv.Use(h.authRequired())

			priv.Get("/users/me", h.handleGetProfile)
			priv.Patch("/users/me/profile", h.handleUpdateProfile)
			priv.Patch("/users/me/password", h.handleChangePassword)

			priv.Get("/tasks", h.handleListTasks)
			priv.Get("/tasks/{id}", h.handleGetTask)
			priv.Post("/tasks/{id}/claim", h.handleClaimTask)
			priv.Post("/tasks/{id}/release", h.handleReleaseTask)
			priv.Post("/tasks/{id}/submit", h.handleSubmitTask)
			priv.Post("/tasks/{id}/complete", h.handleCompleteTask)

			priv.Group(func(admin chi.Router) {
				admin.Use(h.adminRequired())
				admin.Post("/tasks", h.handleCreateTask)
				admin.Patch("/tasks/{id}", h.handleUpdateTask)
				admin.Delete("/tasks/{id}", h.handleDeleteTask)
				admin.Post("/tasks/{id}/publish", h.handlePublishTask)
				admin.Post("/tasks/{id}/archive", h.handleArchiveTask)

				admin.Get("/users", h.handleListUsers)
				admin.Post("/users/{id}/toggle-admin", h.handleToggleAdmin)
			})
		})
	})

	return r
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) clientIP(r *http.Request) string {
	if ip := r.Header.Get(h.cfg.Server.TrustedProxyHeader); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
