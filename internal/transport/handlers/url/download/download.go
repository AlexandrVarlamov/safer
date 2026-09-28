package download

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/aleksandrvarlamov/safer/internal/lib/api/response"
	"github.com/aleksandrvarlamov/safer/internal/lib/logger/ctxlog"
	"github.com/aleksandrvarlamov/safer/internal/lib/logger/sl"
	"github.com/aleksandrvarlamov/safer/internal/lib/validation"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	Bucket string
	Key    string
}

type Response struct {
	Status  string `json:"status"`
	Content []byte `json:"content,omitempty"`
}

type Downloader interface {
	Get(ctx context.Context, bucket, key string) ([]byte, string, error)
}

func Get(downloader Downloader) http.HandlerFunc {
	validate := validation.New()

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "download.Get"

		log := ctxlog.FromContext(r.Context()).With(
			slog.String("op", op),
		)

		bucket := chi.URLParam(r, "bucket")
		key := chi.URLParam(r, "key")

		if bucket == "" || key == "" {
			log.Warn("bucket and key are required")
			render.JSON(w, r, response.Error("bucket and key are required"))
			return
		}

		if err := validate.Struct(r); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				log.Warn("invalid request", sl.Err(err))

				render.JSON(w, r, response.ValidationError(validationErrors))

				return
			}

			log.Error("failed to validate request", sl.Err(err))
			render.JSON(w, r, response.Error(err.Error()))

			return
		}

		data, contentType, err := downloader.Get(r.Context(), bucket, key)
		if err != nil {
			log.Error("failed to get document", sl.Err(err))
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		if contentType == "" {
			contentType = "application/octet-stream"
		}

		w.Header().Set("Content-Disposition", "attachment; filename=\""+key+"\"")
		w.Write(data)
	}
}
