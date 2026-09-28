package ctxlog

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// Inject сохраняет логгер в контексте, чтобы все слои обработки запроса
// писали записи с одинаковыми атрибутами (request_id, method, path).
func Inject(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext возвращает логгер из контекста, а если его там нет
// (фоновые задачи, тесты) — slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
