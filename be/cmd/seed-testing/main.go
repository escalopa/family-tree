package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	sql, err := os.ReadFile("scripts/seed-testing.sql")
	if err != nil {
		log.Fatalf("read testing seed SQL: %v", err)
	}

	if _, err := conn.Exec(ctx, string(sql)); err != nil {
		log.Fatalf("execute testing seed SQL: %v", err)
	}

	fmt.Println("testing seed data loaded")
}
