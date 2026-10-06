package store

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// MemoryStore 是以記憶體儲存資料的 Store 實作。
//
// TODO: 自行設計需要的欄位（資料容器、下一個 ID、鎖……）。
type MemoryStore struct {
	mu     sync.Mutex
	todos  map[int]Todo
	nextID int
}

// 編譯期檢查：確保 *MemoryStore 有實作 Store 介面。請保留這一行。
var _ Store = (*MemoryStore)(nil)

// NewMemoryStore 建立一個空的 MemoryStore。請勿修改函式簽章。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		todos:  make(map[int]Todo),
		nextID: 1,
	}
}

func (m *MemoryStore) List(ctx context.Context, done *bool) ([]Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	todos := make([]Todo, 0, len(m.todos))
	for _, todo := range m.todos {
		if done != nil && *done != todo.Done {
			continue
		}
		todos = append(todos, todo)
	}
	sort.Slice(todos, func(i, j int) bool {
		return todos[i].ID < todos[j].ID
	})
	return todos, nil
}

func (m *MemoryStore) Get(ctx context.Context, id int) (Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	todo, ok := m.todos[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	return todo, nil
}

func (m *MemoryStore) Create(ctx context.Context, title string) (Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > MaxTitleLen {
		return Todo{}, ErrInvalidTitle
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	t := Todo{
		ID:        m.nextID,
		Title:     title,
		CreatedAt: time.Now(),
	}

	m.todos[m.nextID] = t
	m.nextID++

	return t, nil
}

func (m *MemoryStore) Update(ctx context.Context, id int, title *string, done *bool) (Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	todo, ok := m.todos[id]
	if !ok {
		return Todo{}, ErrNotFound
	}

	if title != nil {
		t := strings.TrimSpace(*title)
		if t == "" || utf8.RuneCountInString(t) > MaxTitleLen {
			return Todo{}, ErrInvalidTitle
		}
		todo.Title = t
	}

	if done != nil {
		todo.Done = *done
	}

	m.todos[id] = todo
	return todo, nil
}

func (m *MemoryStore) Delete(ctx context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.todos[id]
	if !ok {
		return ErrNotFound
	}

	delete(m.todos, id)
	return nil
}
