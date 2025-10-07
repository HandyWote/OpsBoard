package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"backend/internal/db"
	"backend/internal/handlers"
	"backend/internal/storage"
)

func main() {
	addr := pickAddr()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbCfg := loadDBConfig()
	conn, err := openDatabase(ctx, dbCfg)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer conn.Close()

	if err := db.RunMigrations(context.Background(), conn); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	store := storage.NewPostgresStore(conn)

	mux := http.NewServeMux()
	mux.Handle("/api/login", handlers.NewLoginHandler(store))

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	dbLabel := dbCfg.DSN
	if dbCfg.Host != "" {
		dbLabel = fmt.Sprintf("%s:%d/%s", dbCfg.Host, dbCfg.Port, dbCfg.Name)
	}

	log.Printf("登录服务已启动，监听地址 %s，数据库 %s", addr, dbLabel)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器运行异常: %v", err)
	}
}

func pickAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":9012"
}

type databaseConfig struct {
	DSN      string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func loadDBConfig() databaseConfig {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return databaseConfig{
			DSN: dsn,
		}
	}

	host := getenvDefault("DB_HOST", "127.0.0.1")
	port := getenvDefault("DB_PORT", "5432")
	user := getenvDefault("DB_USER", "opsboard")
	password := getenvDefault("DB_PASSWORD", "admin")
	name := getenvDefault("DB_NAME", "opsboard")
	sslMode := getenvDefault("DB_SSLMODE", "disable")

	portNum, _ := strconv.Atoi(port)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, name, sslMode,
	)

	return databaseConfig{
		DSN:      dsn,
		Host:     host,
		Port:     portNum,
		User:     user,
		Password: password,
		Name:     name,
		SSLMode:  sslMode,
	}
}

func openDatabase(ctx context.Context, cfg databaseConfig) (*sql.DB, error) {
	conn, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	conn.SetMaxIdleConns(4)
	conn.SetMaxOpenConns(12)
	conn.SetConnMaxLifetime(30 * time.Minute)

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func getenvDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
