package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/bootstrap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := bootstrap.New(ctx)
	if err != nil {
		log.Fatalf("初始化应用失败: %v", err)
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Run(); err != nil {
			log.Fatalf("服务运行异常: %v", err)
		}
	}()

	<-shutdownCh
	log.Println("接收到退出信号，正在优雅关闭…")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	app.Shutdown(shutdownCtx)
}
