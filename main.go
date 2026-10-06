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
	h := logging(handler.New(s))

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

// logging 包住 next，每個請求都印出方法、路徑、狀態碼和花費時間。
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// handler 沒呼叫 WriteHeader 時，Go 會自動回 200，所以預設值是 200
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// statusRecorder 包住 http.ResponseWriter，把 handler 寫入的狀態碼記下來。
type statusRecorder struct {
	http.ResponseWriter // 嵌入：Header()、Write() 直接沿用原本的
	status              int
}

// WriteHeader 先記下狀態碼，再轉交給原本的 ResponseWriter。
func (rec *statusRecorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
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
