package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
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

	waitReporter := startReporter(ctx, s, 10*time.Second)

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

	waitReporter()
	log.Println("shutdown complete")
}
