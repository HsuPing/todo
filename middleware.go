package main

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder 包住 http.ResponseWriter，把 handler 寫入的狀態碼記下來，給 logging 印出。
type statusRecorder struct {
	http.ResponseWriter // 嵌入：Header()、Write()、WriteHeader() 直接沿用原本的
	status              int
}

// 此為 http.ResponseWriter interface 方法實作
func (rec *statusRecorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
