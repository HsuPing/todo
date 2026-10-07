package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo/handler"
	"todo/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := store.NewMemoryStore()
	h := handler.New(s)
	/*
		handler.go 的 http.Handler 被 middleware.go logging 包住
		請求會先經過 logging，再交給 handler
		logging 使用的 statusRecorder 實作了 http.ResponseWriter interface 的方法
		所以 handler.go 在呼叫 WriteHeader 時，會執行 statusRecorder 所定義的行爲
	*/
	h = logging(h)

	waitReporter := startReporter(ctx, s, 10*time.Second)
	runServer(ctx, h)

	waitReporter()
	log.Println("shutdown complete")
}
