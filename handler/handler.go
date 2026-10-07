// Package handler 實作 Todo 的 HTTP API。
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"todo/store"
)

type api struct {
	store store.Store
}

// New 建立並回傳處理所有 /todos 路由的 http.Handler。
// handler 只能依賴 store.Store 介面，不可直接使用 *store.MemoryStore。
//
// ⚠️ 請勿修改函式簽章，自動批改程式會呼叫 handler.New(s)。
func New(s store.Store) http.Handler {
	a := &api{store: s}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", a.getAll)
	mux.HandleFunc("GET /todos/{id}", a.get)
	mux.HandleFunc("POST /todos", a.create)
	mux.HandleFunc("PUT /todos/{id}", a.update)
	mux.HandleFunc("DELETE /todos/{id}", a.delete)

	return mux
}

// Get all todos
func (a *api) getAll(w http.ResponseWriter, r *http.Request) {
	var done *bool
	if v := r.URL.Query().Get("done"); v != "" {
		if v != "true" && v != "false" {
			writeError(w, http.StatusBadRequest, "invalid done")
			return
		}
		b := v == "true"
		done = &b
	}

	todos, err := a.store.List(r.Context(), done)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

// Get a todo by id
func (a *api) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	todo, err := a.store.Get(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

// Create a todo
func (a *api) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	todo, err := a.store.Create(r.Context(), body.Title)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, todo)
}

// Update a todo
func (a *api) update(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title *string `json:"title"`
		Done  *bool   `json:"done"`
	}

	id, ok := parseID(r)

	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if body.Title == nil && body.Done == nil {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	todo, err := a.store.Update(r.Context(), id, body.Title, body.Done)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

// Delete a todo
func (a *api) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err := a.store.Delete(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseID 從路徑取出 id，必須是正整數。
func parseID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

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
		writeError(w, http.StatusNotFound, "todo not found")
	case errors.Is(err, store.ErrInvalidTitle):
		writeError(w, http.StatusBadRequest, "invalid title")
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
