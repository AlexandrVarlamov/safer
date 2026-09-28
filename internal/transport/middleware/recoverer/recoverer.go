package recoverer

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/aleksandrvarlamov/safer/internal/lib/logger/ctxlog"
)

// New перехватывает панику в обработчиках: пишет её в slog в JSON-формате
// (со стеком и request_id) и возвращает клиенту 500.
func New() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
					ctxlog.FromContext(r.Context()).Error("panic recovered",
						slog.Any("panic", rvr),
						slog.String("stack", string(debug.Stack())),
					)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
