package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/floci-next-go/back/internal/infra/storage"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := storage.NewS3Client(ctx)
	if err != nil {
		log.Fatalf("s3 client: %v", err)
	}

	api := client.API()
	bucket := client.Bucket()

	_, err = api.HeadBucketWithContext(ctx, &awss3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Fatalf("head bucket: %v", err)
	}

	key := fmt.Sprintf("__debug__/presign-check-%d", time.Now().UnixNano())
	req, err := client.PresignPutObject(ctx, key, 30*time.Second)
	if err != nil {
		log.Fatalf("presign put: %v", err)
	}
	if req.URL == "" {
		log.Fatalf("presign put: empty URL")
	}

	fmt.Printf(
		"s3 ok: target=%s, head bucket ok, presign put ok (sample key=%q)\n",
		storage.S3TargetSummary(),
		key,
	)
}
