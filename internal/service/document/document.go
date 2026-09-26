package document

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/aleksandrvarlamov/safer/internal/lib/logger/sl"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage interface {
	Put(ctx context.Context, bucket, key string, body io.Reader, size *int64, contentType string) (*s3.PutObjectOutput, error)
	Get(ctx context.Context, bucket, key string) ([]byte, string, error)
}

type Service struct {
	log     *slog.Logger
	storage Storage
}

func New(log *slog.Logger, storage Storage) *Service {
	return &Service{
		log:     log,
		storage: storage,
	}
}

func (s *Service) Save(ctx context.Context, bucket, key string, body io.Reader, size *int64, contentType string) (string, error) {
	const op = "document.Save"

	log := s.log.With(
		slog.String("op", op),
	)

	result, err := s.storage.Put(ctx, bucket, key, body, size, contentType)
	if err != nil {
		log.Error("failed to save document", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	etag := aws.ToString(result.ETag)

	log.Info("document saved", slog.String("etag", etag))

	return etag, nil
}

func (s *Service) Get(ctx context.Context, bucket, key string) ([]byte, string, error) {
	const op = "document.Get"

	log := s.log.With(
		slog.String("op", op),
	)

	content, contentType, err := s.storage.Get(ctx, bucket, key)
	if err != nil {
		log.Error("failed to get document", sl.Err(err))
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("document retrieved", slog.Int("size", len(content)), slog.String("content_type", contentType))

	return content, contentType, nil
}
