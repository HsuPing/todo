package store

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
)

// ptr 回傳 v 的指標，方便產生 Update、List 需要的 *string、*bool 參數。
func ptr[T any](v T) *T { return &v }

func TestMemoryStore_Create(t *testing.T) {
	tests := []struct {
		name      string // 案例名稱，失敗時會印出來
		title     string // 輸入
		wantTitle string // 預期存進去的 title
		wantErr   error  // 預期的錯誤；nil 表示預期成功
	}{
		{name: "正常新增", title: "Buy Milk", wantTitle: "Buy Milk"},
		{name: "中文", title: "買牛奶", wantTitle: "買牛奶"},
		{name: "去除前後空白", title: "  買牛奶  ", wantTitle: "買牛奶"},
		{name: "剛好 100 個中文字", title: strings.Repeat("中", 100), wantTitle: strings.Repeat("中", 100)},
		{name: "100 個中文字加前後空白", title: "  " + strings.Repeat("中", 100) + "  ", wantTitle: strings.Repeat("中", 100)},
		{name: "title 為空字串", title: "", wantErr: ErrInvalidTitle},
		{name: "title 全是空白", title: " \t\n ", wantErr: ErrInvalidTitle},
		{name: "101 個中文字", title: strings.Repeat("中", 101), wantErr: ErrInvalidTitle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemoryStore() // 每個案例用新的 store，互不影響

			got, err := s.Create(context.Background(), tt.title)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v，預期 %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return // 預期失敗的案例，錯誤對了就結束
			}
			if got.ID != 1 {
				t.Errorf("ID = %d，預期 1", got.ID)
			}
			if got.Title != tt.wantTitle {
				t.Errorf("Title = %q，預期 %q", got.Title, tt.wantTitle)
			}
			if got.Done {
				t.Errorf("Done = true，預期預設為 false")
			}
			if got.CreatedAt.IsZero() {
				t.Errorf("CreatedAt 沒有設定")
			}

			// 確認真的有存進去，而不只是回傳值正確
			stored, err := s.Get(context.Background(), got.ID)
			if err != nil {
				t.Fatalf("Get(%d) err = %v，預期找得到", got.ID, err)
			}
			if stored.Title != tt.wantTitle {
				t.Errorf("存進去的 Title = %q，預期 %q", stored.Title, tt.wantTitle)
			}
		})
	}
}

func TestMemoryStore_IDNotReused(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	for want := 1; want <= 3; want++ {
		got, _ := s.Create(ctx, "todo")
		if got.ID != want {
			t.Fatalf("第 %d 筆的 ID = %d，預期 %d", want, got.ID, want)
		}
	}

	// 刪掉最後一筆，下一筆仍然要拿到 4，不能重用 3
	if err := s.Delete(ctx, 3); err != nil {
		t.Fatalf("Delete(3) err = %v", err)
	}
	got, _ := s.Create(ctx, "todo")
	if got.ID != 4 {
		t.Errorf("刪除後新增的 ID = %d，預期 4", got.ID)
	}
}

func TestMemoryStore_Get(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	created, _ := s.Create(ctx, "買牛奶")

	got, err := s.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get(%d) err = %v", created.ID, err)
	}
	if got != created {
		t.Errorf("Get = %+v，預期 %+v", got, created)
	}

	if _, err := s.Get(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(999) err = %v，預期 ErrNotFound", err)
	}
}

func TestMemoryStore_Update(t *testing.T) {
	const original = "原本的標題"

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
		{name: "只改 done", id: 1, done: ptr(true), wantTitle: original, wantDone: true},
		{name: "兩個都改", id: 1, title: ptr("新標題"), done: ptr(true), wantTitle: "新標題", wantDone: true},
		{name: "兩個都不改", id: 1, wantTitle: original},
		{name: "title 去除前後空白", id: 1, title: ptr("  新標題  "), wantTitle: "新標題"},
		// 以下是失敗案例：store 裡的資料必須完全不變
		{name: "title 全是空白", id: 1, title: ptr("   "), done: ptr(true), wantTitle: original, wantErr: ErrInvalidTitle},
		{name: "title 101 個中文字", id: 1, title: ptr(strings.Repeat("中", 101)), done: ptr(true), wantTitle: original, wantErr: ErrInvalidTitle},
		{name: "ID 不存在", id: 999, done: ptr(true), wantTitle: original, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := NewMemoryStore()
			s.Create(ctx, original) // ID 是 1

			got, err := s.Update(ctx, tt.id, tt.title, tt.done)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v，預期 %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && (got.Title != tt.wantTitle || got.Done != tt.wantDone) {
				t.Errorf("回傳 Title=%q Done=%v，預期 Title=%q Done=%v",
					got.Title, got.Done, tt.wantTitle, tt.wantDone)
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

func TestMemoryStore_Delete(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	created, _ := s.Create(ctx, "買牛奶")

	if err := s.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete err = %v", err)
	}
	if _, err := s.Get(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("刪除後 Get err = %v，預期 ErrNotFound", err)
	}
	if err := s.Delete(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("重複刪除 err = %v，預期 ErrNotFound", err)
	}
	if err := s.Delete(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete(999) err = %v，預期 ErrNotFound", err)
	}
}

func TestMemoryStore_List(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	// 建立 6 筆，偶數 ID 標成已完成
	for i := 1; i <= 6; i++ {
		s.Create(ctx, "todo")
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

func TestMemoryStore_ListEmpty(t *testing.T) {
	todos, err := NewMemoryStore().List(context.Background(), nil)
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

func TestMemoryStore_ListReturnsCopy(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	s.Create(ctx, "原本的標題")

	todos, _ := s.List(ctx, nil)
	todos[0].Title = "被呼叫端改掉"

	stored, _ := s.Get(ctx, 1)
	if stored.Title != "原本的標題" {
		t.Errorf("修改 List 的結果影響了 store：Title = %q", stored.Title)
	}
}

// 用 go test -race 執行，才能偵測到漏加鎖的問題。
func TestMemoryStore_Concurrent(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	const n = 100

	var wg sync.WaitGroup
	for range n {
		wg.Add(3)
		go func() { defer wg.Done(); s.Create(ctx, "todo") }()
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
