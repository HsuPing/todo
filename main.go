package main

import (
	"log"
	"net/http"
	"os"

	"todo/handler"
	"todo/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := store.NewMemoryStore()
	h := handler.New(s)

	// TODO（加分題 A）：改用 http.Server 並實作 graceful shutdown。
	// TODO（加分題 B）：用 logging middleware 包住 h。
	// TODO（加分題 C）：啟動背景 goroutine 定期印出未完成數量，並用 context 控制結束。
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, h))
}
