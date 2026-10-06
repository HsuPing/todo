package store

import (
	"context"
	"errors"
)

// errNotImplemented 僅供骨架使用，完成實作後請刪除。
var errNotImplemented = errors.New("not implemented")

// MemoryStore 是以記憶體儲存資料的 Store 實作。
//
// TODO: 自行設計需要的欄位（資料容器、下一個 ID、鎖……）。
type MemoryStore struct {
}

// 編譯期檢查：確保 *MemoryStore 有實作 Store 介面。請保留這一行。
var _ Store = (*MemoryStore)(nil)

// NewMemoryStore 建立一個空的 MemoryStore。請勿修改函式簽章。
func NewMemoryStore() *MemoryStore {
	// TODO
	return &MemoryStore{}
}

func (m *MemoryStore) List(ctx context.Context, done *bool) ([]Todo, error) {
	// TODO
	return nil, errNotImplemented
}

func (m *MemoryStore) Get(ctx context.Context, id int) (Todo, error) {
	// TODO
	return Todo{}, errNotImplemented
}

func (m *MemoryStore) Create(ctx context.Context, title string) (Todo, error) {
	// TODO
	return Todo{}, errNotImplemented
}

func (m *MemoryStore) Update(ctx context.Context, id int, title *string, done *bool) (Todo, error) {
	// TODO
	return Todo{}, errNotImplemented
}

func (m *MemoryStore) Delete(ctx context.Context, id int) error {
	// TODO
	return errNotImplemented
}
