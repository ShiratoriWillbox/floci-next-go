//go:generate mockgen -source=s3.go -destination=s3_mock.go -package storage IS3Client
package storage

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
)

// DefaultPutExpires is the lifetime of generated PUT pre-signed URLs.
const DefaultPutExpires = 15 * time.Minute

// DefaultGetExpires is the lifetime of generated GET pre-signed URLs.
const DefaultGetExpires = 15 * time.Minute

// PresignedPutRequest is the result of presigning a PUT to an object key.
type PresignedPutRequest struct {
	URL          string
	Method       string
	SignedHeader http.Header
}

// PresignedGetRequest is the result of presigning a GET for an object key.
type PresignedGetRequest struct {
	URL          string
	Method       string
	SignedHeader http.Header
}

type IS3Client interface {
	PresignPutObject(ctx context.Context, objectKey string, expires time.Duration) (*PresignedPutRequest, error)
	PresignGetObject(ctx context.Context, objectKey string, expires time.Duration) (*PresignedGetRequest, error)
}

// S3Client wraps the AWS S3 API client (SDK v1) and presigning helpers.
type S3Client struct {
	api    *awss3.S3
	bucket string
}

// NewS3Client builds an S3 client from environment variables.
//
// Required: AWS_REGION (or AWS_DEFAULT_REGION), S3_BUCKET
// Optional: S3_ENDPOINT — base URL for S3-compatible APIs (e.g. LocalStack). Path-style addressing is used when set.
func NewS3Client(ctx context.Context) (*S3Client, error) {
	_ = ctx // session init is synchronous; callers pass context for API calls

	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}

	endpoint := os.Getenv("S3_ENDPOINT")

	awsCfg := &aws.Config{
		Region: aws.String(region),
	}
	if endpoint != "" {
		awsCfg.Endpoint = aws.String(endpoint)
		awsCfg.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, err
	}

	return &S3Client{
		api:    awss3.New(sess),
		bucket: bucket,
	}, nil
}

var _ IS3Client = (*S3Client)(nil)

// API returns the underlying S3 service client (HeadBucket, PutObject, etc.).
func (c *S3Client) API() *awss3.S3 {
	return c.api
}

// Bucket returns the configured default bucket name.
func (c *S3Client) Bucket() string {
	return c.bucket
}

// PresignPutObject generates a presigned HTTP PUT URL and signing headers for the default bucket.
func (c *S3Client) PresignPutObject(ctx context.Context, objectKey string, expires time.Duration) (*PresignedPutRequest, error) {
	req, _ := c.api.PutObjectRequest(&awss3.PutObjectInput{
		Bucket:  aws.String(c.bucket),
		Key:     aws.String(objectKey),
		Tagging: aws.String("tag=test"),
	})
	req.SetContext(ctx)

	urlStr, hdr, err := req.PresignRequest(expires)
	if err != nil {
		return nil, err
	}

	return &PresignedPutRequest{
		URL:          urlStr,
		Method:       "PUT",
		SignedHeader: hdr,
	}, nil
}

// PresignGetObject generates a pre-signed HTTP GET URL for the default bucket.
func (c *S3Client) PresignGetObject(ctx context.Context, objectKey string, expires time.Duration) (*PresignedGetRequest, error) {
	req, _ := c.api.GetObjectRequest(&awss3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	})
	req.SetContext(ctx)

	urlStr, hdr, err := req.PresignRequest(expires)
	if err != nil {
		return nil, err
	}

	return &PresignedGetRequest{
		URL:          urlStr,
		Method:       "GET",
		SignedHeader: hdr,
	}, nil
}

// S3TargetSummary describes the configured S3 target for logs (no secrets).
func S3TargetSummary() string {
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "(missing)"
	}
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		return fmt.Sprintf("bucket=%s region=%s endpoint=(default)", bucket, region)
	}
	return fmt.Sprintf("bucket=%s region=%s endpoint=%s", bucket, region, endpoint)
}
