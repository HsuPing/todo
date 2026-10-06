// Package handler 實作 Todo 的 HTTP API。
package handler

import (
	"net/http"

	"todo/store"
)

// New 建立並回傳處理所有 /todos 路由的 http.Handler。
// handler 只能依賴 store.Store 介面，不可直接使用 *store.MemoryStore。
//
// ⚠️ 請勿修改函式簽章，自動批改程式會呼叫 handler.New(s)。
func New(s store.Store) http.Handler {
	mux := http.NewServeMux()

	// TODO: 註冊路由，例如（Go 1.22+ 語法）：
	//   mux.HandleFunc("GET /todos", ...)
	//   mux.HandleFunc("GET /todos/{id}", ...)   // 用 r.PathValue("id") 取值
	//   mux.HandleFunc("POST /todos", ...)
	//   mux.HandleFunc("PUT /todos/{id}", ...)
	//   mux.HandleFunc("DELETE /todos/{id}", ...)
	_ = s

	return mux
}
