// Package store 定義 Todo 的資料模型與儲存介面。
//
// ⚠️ 本檔案為「固定介面」，請勿修改型別名稱、欄位、函式簽章與錯誤變數，
// 自動批改程式會依照這裡的定義呼叫你的程式。
package store

import (
	"context"
	"errors"
	"time"
)

// MaxTitleLen 是 title 的最大長度，以 rune（字元）計算，中文一個字算 1。
const MaxTitleLen = 100

// 由 Store 實作回傳的標準錯誤。呼叫端應使用 errors.Is 判斷。
var (
	// ErrNotFound 表示指定 ID 的 todo 不存在。
	ErrNotFound = errors.New("todo not found")
	// ErrInvalidTitle 表示 title 為空白，或超過 MaxTitleLen 個字元。
	ErrInvalidTitle = errors.New("invalid title")
)

// Todo 是一筆待辦事項。
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// Store 是 todo 的儲存介面。所有實作都必須可被多個 goroutine 同時安全呼叫。
type Store interface {
	// List 回傳所有 todo，依 ID 由小到大排序。
	// done 為 nil 表示不篩選；不為 nil 則只回傳 Done == *done 的項目。
	// 沒有資料時回傳空 slice（非 nil）。
	// 回傳的 slice 被呼叫端修改時，不得影響 Store 內部資料。
	List(ctx context.Context, done *bool) ([]Todo, error)

	// Get 回傳指定 ID 的 todo；不存在時回傳 ErrNotFound。
	Get(ctx context.Context, id int) (Todo, error)

	// Create 建立新的 todo 並回傳。
	// ID 從 1 開始遞增、不可重複（刪除後也不可重用）；Done 預設 false；
	// CreatedAt 為建立當下時間。
	// title 去除前後空白後為空字串，或長度超過 MaxTitleLen 時回傳 ErrInvalidTitle。
	// 儲存的 title 為去除前後空白後的結果。
	Create(ctx context.Context, title string) (Todo, error)

	// Update 更新指定 ID 的 todo，nil 參數代表該欄位不變更，回傳更新後的結果。
	// 不存在時回傳 ErrNotFound；title 不合法時回傳 ErrInvalidTitle（且不做任何變更）。
	Update(ctx context.Context, id int, title *string, done *bool) (Todo, error)

	// Delete 刪除指定 ID 的 todo；不存在時回傳 ErrNotFound。
	Delete(ctx context.Context, id int) error
}
