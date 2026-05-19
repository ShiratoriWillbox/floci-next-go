package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/floci-next-go/back/internal/database"
	"github.com/floci-next-go/back/internal/infra/dto"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.OpenFromEnv()
	if err != nil {
		log.Fatalf("open: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("sql db: %v", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	var n int64
	if err := db.WithContext(ctx).Model(&dto.Todo{}).Count(&n).Error; err != nil {
		log.Fatalf("count todos: %v", err)
	}

	fmt.Printf("db ok: driver=postgres target=%s, ping ok, todos.count=%d\n", database.PostgresTargetSummary(), n)
}
