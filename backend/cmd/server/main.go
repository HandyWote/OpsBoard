package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"backend/internal/handlers"
	"backend/internal/storage"
)

func main() {
	addr := pickAddr()
	storePath := pickStorePath()

	store, err := storage.NewFileStore(storePath)
	if err != nil {
		log.Fatalf("初始化用户存储失败: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/login", handlers.NewLoginHandler(store))

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("登录服务已启动，监听地址 %s，数据文件 %s", addr, storePath)

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

func pickStorePath() string {
	if path := os.Getenv("USER_STORE_PATH"); path != "" {
		return path
	}
	return filepath.Join("data", "users.json")
}
