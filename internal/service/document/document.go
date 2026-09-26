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
