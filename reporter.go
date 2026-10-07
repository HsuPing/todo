package main

import (
	"context"
	"log"
	"sync"
	"time"

	"todo/store"
)

func startReporter(ctx context.Context, s store.Store, interval time.Duration) func() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		reportPending(ctx, s, interval)
	}()

	return func() {
		wg.Wait()
	}
}

// reportPending 每隔 interval 印出一次未完成的 todo 數量，ctx 被取消時結束。
func reportPending(ctx context.Context, s store.Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	notDone := false
	for {
		select {
		// 按下 Ctrl+C 時，ctx 被取消，結束 reportPending
		case <-ctx.Done():
			log.Println("reporter stopped")
			return
		// 每隔 interval 印出一次未完成的 todo 數量
		case <-ticker.C:
			todos, err := s.List(ctx, &notDone)
			if err != nil {
				log.Printf("report error: %v", err)
				continue
			}
			log.Printf("pending todos: %d", len(todos))
		}
	}
}
