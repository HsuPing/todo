package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
)

func newTestMySQLStore(t *testing.T) *MySQLStore {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN not set")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}

	db.SetMaxOpenConns(10)
	err = db.PingContext(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { db.Close() })

	_, err = db.ExecContext(t.Context(), "TRUNCATE TABLE todos")
	if err != nil {
		t.Fatal(err)
	}

	return NewMySQLStore(db)
}

func TestMySQLStore_IDNotReused(t *testing.T) {
	ctx := context.Background()
	s := newTestMySQLStore(t)

	for want := 1; want <= 3; want++ {
		todo, err := s.Create(ctx, "todo")

		if err != nil {
			t.Fatalf("Create err = %v", err)
		}

		if todo.ID != want {
			t.Fatalf("第 %d 筆的 ID = %d，預期 %d", want, todo.ID, want)
		}
	}

	// 刪掉最後一筆，下一筆仍然要拿到 4，不能重用 3
	if err := s.Delete(ctx, 3); err != nil {
		t.Fatalf("Delete(3) err = %v", err)
	}
	todo, err := s.Create(ctx, "todo")

	if err != nil {
		t.Fatalf("Create err = %v", err)
	}

	if todo.ID != 4 {
		t.Errorf("刪除後新增的 ID = %d，預期 4", todo.ID)
	}
}

func TestMySQLStore_Update(t *testing.T) {
	const originalTitle = "原本的標題"

	tests := []struct {
		name      string
		id        int
		title     *string // nil 表示不改
		done      *bool   // nil 表示不改
		wantTitle string  // 呼叫後 store 裡的 title
		wantDone  bool    // 呼叫後 store 裡的 done
		wantErr   error
	}{
		{name: "只改 title", id: 1, title: ptr("新標題"), wantTitle: "新標題"},
		{name: "只改 done", id: 1, done: ptr(true), wantTitle: originalTitle, wantDone: true},
		{name: "兩個都改", id: 1, title: ptr("新標題"), done: ptr(true), wantTitle: "新標題", wantDone: true},
		{name: "兩個都不改", id: 1, wantTitle: originalTitle},
		{name: "title 去除前後空白", id: 1, title: ptr("  新標題  "), wantTitle: "新標題"},
		// 以下是失敗案例：store 裡的資料必須完全不變
		{name: "title 全是空白", id: 1, title: ptr("   "), done: ptr(true), wantTitle: originalTitle, wantErr: ErrInvalidTitle},
		{name: "title 101 個中文字", id: 1, title: ptr(strings.Repeat("中", 101)), done: ptr(true), wantTitle: originalTitle, wantErr: ErrInvalidTitle},
		{name: "ID 不存在", id: 999, done: ptr(true), wantTitle: originalTitle, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := newTestMySQLStore(t)
			todo, err := s.Create(ctx, originalTitle) // ID 是 1
			if err != nil {
				t.Fatalf("Create err = %v", err)
			}

			todo, err = s.Update(ctx, tt.id, tt.title, tt.done)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v，預期 %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && (todo.Title != tt.wantTitle || todo.Done != tt.wantDone) {
				t.Errorf("回傳 Title=%q Done=%v，預期 Title=%q Done=%v",
					todo.Title, todo.Done, tt.wantTitle, tt.wantDone)
			}

			// 不管成功或失敗，都檢查 store 裡實際存的資料
			stored, _ := s.Get(ctx, 1)
			if stored.Title != tt.wantTitle || stored.Done != tt.wantDone {
				t.Errorf("store 裡是 Title=%q Done=%v，預期 Title=%q Done=%v",
					stored.Title, stored.Done, tt.wantTitle, tt.wantDone)
			}
		})
	}
}

func TestMySQLStore_Delete(t *testing.T) {
	ctx := context.Background()
	s := newTestMySQLStore(t)
	todo, err := s.Create(ctx, "買牛奶")

	if err != nil {
		t.Fatalf("Create err = %v", err)
	}

	if err := s.Delete(ctx, todo.ID); err != nil {
		t.Fatalf("Delete err = %v", err)
	}
	if _, err := s.Get(ctx, todo.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("刪除後 Get err = %v，預期 ErrNotFound", err)
	}
	if err := s.Delete(ctx, todo.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("重複刪除 err = %v，預期 ErrNotFound", err)
	}
	if err := s.Delete(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete(999) err = %v，預期 ErrNotFound", err)
	}
}

func TestMySQLStore_List(t *testing.T) {
	ctx := context.Background()
	s := newTestMySQLStore(t)
	// 建立 6 筆，偶數 ID 標成已完成
	for i := 1; i <= 6; i++ {
		_, err := s.Create(ctx, "todo")
		if err != nil {
			t.Fatalf("Create err = %v", err)
		}
		if i%2 == 0 {
			s.Update(ctx, i, nil, ptr(true))
		}
	}

	tests := []struct {
		name    string
		done    *bool
		wantIDs []int // 同時檢查篩選結果和排序
	}{
		{name: "不篩選", done: nil, wantIDs: []int{1, 2, 3, 4, 5, 6}},
		{name: "只要已完成", done: ptr(true), wantIDs: []int{2, 4, 6}},
		{name: "只要未完成", done: ptr(false), wantIDs: []int{1, 3, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todos, err := s.List(ctx, tt.done)
			if err != nil {
				t.Fatalf("List err = %v", err)
			}
			var todoIDs []int
			for _, todo := range todos {
				todoIDs = append(todoIDs, todo.ID)
			}
			if !slices.Equal(todoIDs, tt.wantIDs) {
				t.Errorf("IDs = %v，預期 %v", todoIDs, tt.wantIDs)
			}
		})
	}
}

func TestMySQLStore_ListEmpty(t *testing.T) {
	todos, err := newTestMySQLStore(t).List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List err = %v", err)
	}
	// nil slice 轉成 JSON 是 null，空 slice 才是 []
	if todos == nil {
		t.Errorf("List 回傳 nil，預期空 slice")
	}
	if len(todos) != 0 {
		t.Errorf("len = %d，預期 0", len(todos))
	}
}

func TestMySQLStore_ListReturnsCopy(t *testing.T) {
	ctx := context.Background()
	s := newTestMySQLStore(t)
	_, err := s.Create(ctx, "原本的標題")
	if err != nil {
		t.Fatalf("Create err = %v", err)
	}

	todos, err := s.List(ctx, nil)
	if err != nil {
		t.Fatalf("List err = %v", err)
		return
	}

	todos[0].Title = "被呼叫端改掉"

	stored, err := s.Get(ctx, 1)
	if err != nil {
		t.Fatalf("Get err = %v", err)
		return
	}

	if stored.Title != "原本的標題" {
		t.Errorf("修改 List 的結果影響了 store：Title = %q", stored.Title)
	}
}

// 用 go test -race ./store 執行，才能偵測到漏加鎖的問題。
func TestMySQLStore_Concurrent(t *testing.T) {
	ctx := context.Background()
	s := newTestMySQLStore(t)
	const n = 100

	var wg sync.WaitGroup
	for range n {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_, err := s.Create(ctx, "todo")
			if err != nil {
				t.Errorf("Create err = %v", err)
			}
		}()
		go func() { defer wg.Done(); s.Get(ctx, 1) }()
		go func() { defer wg.Done(); s.List(ctx, nil) }()
	}
	wg.Wait()

	// 同時新增 100 筆，ID 必須剛好是 1~100，不能重複或跳號
	todos, _ := s.List(ctx, nil)
	if len(todos) != n {
		t.Fatalf("len = %d，預期 %d", len(todos), n)
	}
	for i, todo := range todos {
		if todo.ID != i+1 {
			t.Fatalf("第 %d 筆 ID = %d，預期 %d", i, todo.ID, i+1)
		}
	}
}
