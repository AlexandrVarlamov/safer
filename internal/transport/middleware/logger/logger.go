package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/aleksandrvarlamov/safer/internal/lib/logger/ctxlog"
	"github.com/go-chi/chi/v5/middleware"
)

// New логирует завершение каждого запроса одним JSON-событием
// (уровень зависит от HTTP-статуса) и кладёт логгер запроса в контекст
// для остальных слоёв.
func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/logger"),
		)

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				status := ww.Status()
				level := slog.LevelInfo
				switch {
				case status >= 500:
					level = slog.LevelError
				case status >= 400:
					level = slog.LevelWarn
				}
				entry.LogAttrs(r.Context(), level, "request completed",
					slog.Int("status", status),
					slog.Int64("bytes", int64(ww.BytesWritten())),
					slog.Float64("duration_ms", float64(time.Since(t1).Microseconds())/1000.0),
				)
			}()

			next.ServeHTTP(ww, r.WithContext(ctxlog.Inject(r.Context(), entry)))
		}

		return http.HandlerFunc(fn)
	}
}
