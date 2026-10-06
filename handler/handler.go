// Package handler 實作 Todo 的 HTTP API。
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"todo/store"
)

// writeJSON 把 v 轉成 JSON 寫進回應，狀態碼是 status。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError 回傳 {"error": "訊息"} 格式的錯誤。
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// writeStoreError 把 store 回傳的錯誤轉成對應的狀態碼。
func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrInvalidTitle):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

// New 建立並回傳處理所有 /todos 路由的 http.Handler。
// handler 只能依賴 store.Store 介面，不可直接使用 *store.MemoryStore。
//
// ⚠️ 請勿修改函式簽章，自動批改程式會呼叫 handler.New(s)。
func New(s store.Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /todos", func(w http.ResponseWriter, r *http.Request) {
		var done *bool
		if v := r.URL.Query().Get("done"); v != "" {
			b, err := strconv.ParseBool(v)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid done")
				return
			}
			done = &b
		}

		todos, err := s.List(r.Context(), done)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, todos)
	})

	mux.HandleFunc("GET /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		todo, err := s.Get(r.Context(), id)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, todo)
	})

	mux.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		var todo store.Todo
		err := json.NewDecoder(r.Body).Decode(&todo)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		todo, err = s.Create(r.Context(), todo.Title)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, todo)
	})

	mux.HandleFunc("PUT /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Title *string `json:"title"`
			Done  *bool   `json:"done"`
		}

		id, err := strconv.Atoi(r.PathValue("id"))

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		err = json.NewDecoder(r.Body).Decode(&body)

		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		todo, err := s.Update(r.Context(), id, body.Title, body.Done)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, todo)
	})

	mux.HandleFunc("DELETE /todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		err = s.Delete(r.Context(), id)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}
