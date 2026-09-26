package save

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aleksandrvarlamov/safer/internal/lib/api/response"
	"github.com/aleksandrvarlamov/safer/internal/lib/logger/sl"
	"github.com/aleksandrvarlamov/safer/internal/lib/validation"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	Bucket string `json:"bucket" validate:"required,bucket"`
	Key    string `json:"key" validate:"required,key"`
}

type Response struct {
	Status response.Response `json:"status"`
	Key    string            `json:"key"`
	Size   int64             `json:"size,omitempty"`
	ETag   string            `json:"etag,omitempty"`
}

type Saver interface {
	Save(ctx context.Context, bucket, key string, body io.Reader, size *int64, contentType string) (string, error)
}

func New(log *slog.Logger, saver Saver) http.HandlerFunc {
	validate := validation.New()

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			log.Error("failed to parse multipart form", sl.Err(err))
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		var req Request
		req.Bucket = r.FormValue("bucket")
		req.Key = r.FormValue("key")

		if err := validate.Struct(req); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				log.Error("invalid request", sl.Err(err))

				render.JSON(w, r, response.ValidationError(validationErrors))

				return
			}

			log.Error("failed to validate request", sl.Err(err))
			render.JSON(w, r, response.Error(err.Error()))

			return
		}

		files := r.MultipartForm.File["file"]
		if len(files) == 0 {
			log.Error("no files in request")

			render.JSON(w, r, response.Error("no files provided"))

			return
		}

		saved := make([]Response, 0, len(files))
		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				log.Error("failed to open uploaded file", sl.Err(err))

				render.JSON(w, r, response.Error(err.Error()))

				return
			}

			contentType := fh.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			etag, err := saver.Save(r.Context(), req.Bucket, req.Key, f, &fh.Size, contentType)
			f.Close()
			if err != nil {
				log.Error("failed to save document", sl.Err(err))

				render.JSON(w, r, response.Error(err.Error()))

				return
			}

			saved = append(saved, Response{
				Status: response.OK(),
				Key:    req.Key,
				Size:   fh.Size,
				ETag:   strings.Trim(aws.ToString(&etag), `"`),
			})
		}

		render.JSON(w, r, saved)
	}
}
