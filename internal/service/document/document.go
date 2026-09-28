package document

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aleksandrvarlamov/safer/internal/lib/logger/ctxlog"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage interface {
	Put(ctx context.Context, bucket, key string, body io.Reader, size *int64, contentType string) (*s3.PutObjectOutput, error)
	Get(ctx context.Context, bucket, key string) ([]byte, string, error)
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Save(ctx context.Context, bucket, key string, body io.Reader, size *int64, contentType string) (string, error) {
	const op = "document.Save"

	log := ctxlog.FromContext(ctx).With(
		slog.String("op", op),
	)

	result, err := s.storage.Put(ctx, bucket, key, body, size, contentType)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	etag := aws.ToString(result.ETag)

	attrs := []slog.Attr{slog.String("etag", etag)}
	if size != nil {
		attrs = append(attrs, slog.Int64("size", *size))
	}
	log.LogAttrs(ctx, slog.LevelInfo, "document saved", attrs...)

	return etag, nil
}

func (s *Service) Get(ctx context.Context, bucket, key string) ([]byte, string, error) {
	const op = "document.Get"

	log := ctxlog.FromContext(ctx).With(
		slog.String("op", op),
	)

	content, contentType, err := s.storage.Get(ctx, bucket, key)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("document retrieved",
		slog.Int("size", len(content)),
		slog.String("content_type", contentType),
	)

	return content, contentType, nil
}
