package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aleksandrvarlamov/safer/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2 struct {
	client *s3.Client
}

func NewR2(ctx context.Context, cfg *config.Config) (*R2, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID, cfg.SecretAccessKey, "",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
		o.UsePathStyle = true
	})

	return &R2{client: client}, nil
}

func (r *R2) Put(ctx context.Context, bucket, key string, body io.Reader, size *int64, contentType string) (*s3.PutObjectOutput, error) {
	result, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
		ContentLength: size,
		ContentType:   aws.String(contentType),
	})
	return result, err
}

func (r *R2) Get(ctx context.Context, bucket, key string) ([]byte, error) {
	resp, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}