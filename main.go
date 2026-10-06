package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"todo/handler"
	"todo/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := store.NewMemoryStore()
	h := logging(handler.New(s))

	// TODO（加分題 A）：改用 http.Server 並實作 graceful shutdown。
	// TODO（加分題 C）：啟動背景 goroutine 定期印出未完成數量，並用 context 控制結束。
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, h))
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
