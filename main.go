package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"todo/handler"
	"todo/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := store.NewMemoryStore()
	h := handler.New(s)
	/*
		handler.go 的 http.Handler 被 middleware.go logging 包住
		請求會先經過 logging，再交給 handler
		logging 使用的 statusRecorder 實作了 http.ResponseWriter interface 的方法
		所以 handler.go 在呼叫 WriteHeader 時，會執行 statusRecorder 所定義的行爲
	*/
	h = logging(h)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		reportPending(ctx, s, 10*time.Second)
	}()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	wg.Wait()
	log.Println("shutdown complete")
}

// reportPending 每隔 interval 印出一次未完成的 todo 數量，ctx 被取消時結束。
func reportPending(ctx context.Context, s store.Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	notDone := false
	for {
		select {
		case <-ctx.Done():
			log.Println("reporter stopped")
			return
		case <-ticker.C:
			todos, err := s.List(ctx, &notDone)
			if err != nil {
				log.Printf("report error: %v", err)
				continue
			}
			log.Printf("pending todos: %d", len(todos))
		}
	}
}
