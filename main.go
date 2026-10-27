package main

import (
	"context"
	"database/sql"
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

	mysql_dsn := os.Getenv("MYSQL_DSN")

	var s store.Store
	if mysql_dsn != "" {
		db, err := sql.Open("mysql", mysql_dsn)
		if err != nil {
			log.Fatal(err)
		}

		db.SetMaxOpenConns(10)
		defer db.Close()

		err = db.PingContext(ctx)
		if err != nil {
			log.Fatal(err)
		}
		s = store.NewMySQLStore(db)
	} else {
		s = store.NewMemoryStore()
	}
	h := handler.New(s)

	h = logging(h)

	waitReporter := startReporter(ctx, s, 10*time.Second)
	runServer(ctx, h)

	waitReporter()
	log.Println("shutdown complete")
}
