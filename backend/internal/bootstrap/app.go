package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/logger"
	"backend/internal/repository"
	"backend/internal/service"
	httptransport "backend/internal/transport/http"
)

// Application 聚合服务所需的全部依赖并负责启动 HTTP 服务。
type Application struct {
	cfg    config.Config
	log    *zap.Logger
	db     *sql.DB
	server *http.Server
}

// New 构造应用实例。
func New(ctx context.Context) (*Application, error) {
	if err := loadDotEnv(".env"); err != nil {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log, err := logger.New()
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}
	logger.ReplaceGlobals(log)

	dbConn, err := openDatabase(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := db.RunMigrations(ctx, dbConn); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	repos := repository.NewRegistry(dbConn)
	services := service.NewRegistry(cfg, repos, log)

	router := httptransport.NewRouter(cfg, services, log)

	server := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	return &Application{
		cfg:    cfg,
		log:    log,
		db:     dbConn,
		server: server,
	}, nil
}

// Run 启动 HTTP 服务。
func (a *Application) Run() error {
	a.log.Info("server starting", zap.String("addr", a.server.Addr))
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown 优雅关闭。
func (a *Application) Shutdown(ctx context.Context) {
	if a == nil {
		return
	}

	if a.server != nil {
		_ = a.server.Shutdown(ctx)
	}
	if a.db != nil {
		_ = a.db.Close()
	}
	logger.SyncOnShutdown(a.log)
}

func openDatabase(ctx context.Context, cfg config.DBConfig) (*sql.DB, error) {
	conn, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}
