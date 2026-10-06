package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"todo/store"
)

// newHandler 建立一個新的 handler，裡面預先放一筆 todo：ID 1「買牛奶」。
func newHandler(t *testing.T) http.Handler {
	t.Helper() // 失敗時，錯誤訊息的行號會指向呼叫 newHandler 的那一行
	s := store.NewMemoryStore()
	if _, err := s.Create(context.Background(), "買牛奶"); err != nil {
		t.Fatalf("準備資料失敗：%v", err)
	}
	return New(s)
}

// do 對 h 送出一個 HTTP 請求，回傳錄下來的回應。
// 不需要真的啟動 server，整個過程都在記憶體裡完成。
func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// decode 把回應的 JSON body 解析進 v。
func decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("回應不是合法的 JSON：%v（body：%s）", err, rec.Body.String())
	}
}

func TestHandler_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string // 回應裡必須包含的文字；空字串表示不檢查
	}{
		// GET /todos
		{name: "列出全部", method: "GET", path: "/todos", wantStatus: http.StatusOK, wantBody: `"title":"買牛奶"`},
		{name: "done 不是布林值", method: "GET", path: "/todos?done=abc", wantStatus: http.StatusBadRequest, wantBody: `"error"`},

		// GET /todos/{id}
		{name: "查詢存在的 id", method: "GET", path: "/todos/1", wantStatus: http.StatusOK, wantBody: `"id":1`},
		{name: "查詢不存在的 id", method: "GET", path: "/todos/999", wantStatus: http.StatusNotFound, wantBody: `"error"`},
		{name: "查詢 id 不是數字", method: "GET", path: "/todos/abc", wantStatus: http.StatusBadRequest, wantBody: `"error"`},

		// POST /todos
		{name: "新增成功", method: "POST", path: "/todos", body: `{"title":"買豆漿"}`, wantStatus: http.StatusCreated, wantBody: `"title":"買豆漿"`},
		{name: "新增 title 空白", method: "POST", path: "/todos", body: `{"title":"   "}`, wantStatus: http.StatusBadRequest, wantBody: `"error"`},
		{name: "新增沒有 title", method: "POST", path: "/todos", body: `{}`, wantStatus: http.StatusBadRequest, wantBody: `"error"`},
		{name: "新增 body 不是 JSON", method: "POST", path: "/todos", body: `not json`, wantStatus: http.StatusBadRequest, wantBody: `"error"`},

		// PUT /todos/{id}
		{name: "更新成功", method: "PUT", path: "/todos/1", body: `{"done":true}`, wantStatus: http.StatusOK, wantBody: `"done":true`},
		{name: "更新 title 空白", method: "PUT", path: "/todos/1", body: `{"title":""}`, wantStatus: http.StatusBadRequest, wantBody: `"error"`},
		{name: "更新不存在的 id", method: "PUT", path: "/todos/999", body: `{"done":true}`, wantStatus: http.StatusNotFound, wantBody: `"error"`},
		{name: "更新 id 不是數字", method: "PUT", path: "/todos/abc", body: `{"done":true}`, wantStatus: http.StatusBadRequest, wantBody: `"error"`},
		{name: "更新 body 不是 JSON", method: "PUT", path: "/todos/1", body: `not json`, wantStatus: http.StatusBadRequest, wantBody: `"error"`},

		// DELETE /todos/{id}
		{name: "刪除成功", method: "DELETE", path: "/todos/1", wantStatus: http.StatusNoContent},
		{name: "刪除不存在的 id", method: "DELETE", path: "/todos/999", wantStatus: http.StatusNotFound, wantBody: `"error"`},
		{name: "刪除 id 不是數字", method: "DELETE", path: "/todos/abc", wantStatus: http.StatusBadRequest, wantBody: `"error"`},

		// 路由本身
		{name: "不支援的方法", method: "PATCH", path: "/todos/1", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(t) // 每個案例用新的 handler，互不影響

			rec := do(h, tt.method, tt.path, tt.body)

			if rec.Code != tt.wantStatus {
				t.Errorf("狀態碼 = %d，預期 %d（body：%s）", rec.Code, tt.wantStatus, rec.Body.String())
			}
			// strings.Contains(任何字串, "") 一定是 true，所以 wantBody 空白時等於不檢查
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %s，預期包含 %s", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestHandler_Create(t *testing.T) {
	h := newHandler(t)

	rec := do(h, "POST", "/todos", `{"title":"  買豆漿  "}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("狀態碼 = %d，預期 201", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q，預期 application/json", ct)
	}
	var got store.Todo
	decode(t, rec, &got)
	if got.ID != 2 || got.Title != "買豆漿" || got.Done {
		t.Errorf("回傳 %+v，預期 ID=2 Title=買豆漿 Done=false", got)
	}

	// 確認真的存進去了
	if rec := do(h, "GET", "/todos/2", ""); rec.Code != http.StatusOK {
		t.Errorf("新增後 GET /todos/2 狀態碼 = %d，預期 200", rec.Code)
	}
}

// 只送部分欄位時，沒送的欄位必須維持原值。
func TestHandler_UpdateKeepsOtherFields(t *testing.T) {
	h := newHandler(t)
	var got store.Todo

	// 只送 done：title 不能被清空
	decode(t, do(h, "PUT", "/todos/1", `{"done":true}`), &got)
	if got.Title != "買牛奶" || !got.Done {
		t.Errorf("只改 done 後 = %+v，預期 Title=買牛奶 Done=true", got)
	}

	// 只送 title：done 不能被改回 false
	decode(t, do(h, "PUT", "/todos/1", `{"title":"買豆漿"}`), &got)
	if got.Title != "買豆漿" || !got.Done {
		t.Errorf("只改 title 後 = %+v，預期 Title=買豆漿 Done=true", got)
	}
}

func TestHandler_DeleteThenGet(t *testing.T) {
	h := newHandler(t)

	rec := do(h, "DELETE", "/todos/1", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE 狀態碼 = %d，預期 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("204 不應該有 body，得到 %s", rec.Body.String())
	}

	if rec := do(h, "GET", "/todos/1", ""); rec.Code != http.StatusNotFound {
		t.Errorf("刪除後 GET 狀態碼 = %d，預期 404", rec.Code)
	}
}

func TestHandler_ListFilter(t *testing.T) {
	h := newHandler(t)                         // ID 1「買牛奶」
	do(h, "POST", "/todos", `{"title":"寫作業"}`) // ID 2
	do(h, "POST", "/todos", `{"title":"運動"}`)  // ID 3
	do(h, "PUT", "/todos/2", `{"done":true}`)

	tests := []struct {
		name    string
		path    string
		wantIDs []int
	}{
		{name: "不篩選", path: "/todos", wantIDs: []int{1, 2, 3}},
		{name: "只要已完成", path: "/todos?done=true", wantIDs: []int{2}},
		{name: "只要未完成", path: "/todos?done=false", wantIDs: []int{1, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(h, "GET", tt.path, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("狀態碼 = %d，預期 200", rec.Code)
			}
			var todos []store.Todo
			decode(t, rec, &todos)

			var gotIDs []int
			for _, todo := range todos {
				gotIDs = append(gotIDs, todo.ID)
			}
			if !slices.Equal(gotIDs, tt.wantIDs) {
				t.Errorf("IDs = %v，預期 %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

// 沒有資料時要回傳 []，不能是 null。
func TestHandler_ListEmpty(t *testing.T) {
	h := New(store.NewMemoryStore()) // 空的 store

	rec := do(h, "GET", "/todos", "")

	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("body = %s，預期 []", body)
	}
}
