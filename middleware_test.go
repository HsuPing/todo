package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc // 模擬的 handler
		wantCode int              // 客戶端應該收到的狀態碼
		wantLog  string           // 紀錄裡必須包含的文字
	}{
		{
			name:     "handler 有呼叫 WriteHeader",
			handler:  func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusCreated) },
			wantCode: http.StatusCreated,
			wantLog:  "POST /todos 201",
		},
		{
			name:     "handler 沒呼叫 WriteHeader 時記成 200",
			handler:  func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) },
			wantCode: http.StatusOK,
			wantLog:  "POST /todos 200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)            // 把 log 改寫進 buf，才能檢查內容
			defer log.SetOutput(os.Stderr) // 測完改回來

			rec := httptest.NewRecorder()
			logging(tt.handler).ServeHTTP(rec, httptest.NewRequest("POST", "/todos", nil))

			if rec.Code != tt.wantCode {
				t.Errorf("客戶端收到 %d，預期 %d", rec.Code, tt.wantCode)
			}
			if !strings.Contains(buf.String(), tt.wantLog) {
				t.Errorf("紀錄 = %q，預期包含 %q", buf.String(), tt.wantLog)
			}
		})
	}
}
