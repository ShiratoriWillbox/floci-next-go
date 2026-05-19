package main

import (
	"context"
	"log"
	"os"

	"github.com/floci-next-go/back/internal/database"
	"github.com/floci-next-go/back/internal/handlers"
	"github.com/floci-next-go/back/internal/infra/repository"
	"github.com/floci-next-go/back/internal/infra/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	addr := os.Getenv("BACK_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	db, err := database.OpenFromEnv()
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	ctx := context.Background()
	s3Client, err := storage.NewS3Client(ctx)
	if err != nil {
		log.Fatalf("s3 client: %v", err)
	}

	h := &handlers.TodoHandler{
		Todos:    repository.NewTodoRepository(db),
		Files:    repository.NewFileRepository(db),
		S3Client: s3Client,
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		// ブラウザが API を別オリジンへ直接叩く場合の CORS（Next の rewrite 経由なら同一オリジンだが、開発用に幅広く許可）
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		if rq := c.Request.Header.Get("Access-Control-Request-Headers"); rq != "" {
			c.Writer.Header().Set("Access-Control-Allow-Headers", rq)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With")
		}
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		api.GET("/todos", h.List)
		api.GET("/todos/:id", h.Get)
		api.POST("/todos", h.Create)
		api.PUT("/todos/:id", h.Update)
		api.DELETE("/todos/:id", h.Delete)
		api.POST("/todos/:id/files", h.CreateTodoFileUpload)
		api.PUT("/todos/:id/files/:file_id", h.AttachTodoFile)
		api.GET("/todos/:id/files/:file_id", h.GetTodoFileDownload)
	}

	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
