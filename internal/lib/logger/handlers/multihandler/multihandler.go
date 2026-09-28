package multihandler

import (
	"context"
	"log/slog"
)

// Multi рассылает каждую запись сразу в несколько хендлеров
// (например, pretty в консоль + JSON в файл для Alloy).
type Multi struct {
	handlers []slog.Handler
}

func New(handlers ...slog.Handler) *Multi {
	return &Multi{handlers: handlers}
}

func (m *Multi) Enabled(ctx context.Context, l slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, l) {
			return true
		}
	}
	return false
}

func (m *Multi) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *Multi) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return New(handlers...)
}

func (m *Multi) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return New(handlers...)
}
