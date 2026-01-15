package logging

import (
	"context"
	"log/slog"

	"github.com/pupload/pupload/internal/models"
)

type CollectHandler struct {
	Inner   slog.Handler
	Records *[]models.LogRecord
}

func (h *CollectHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Inner.Enabled(ctx, level)
}

func (h *CollectHandler) Handle(ctx context.Context, r slog.Record) error {
	err := h.Inner.Handle(ctx, r)
	if err != nil {
		return err
	}

	rec := models.LogRecord{
		Time:   r.Time,
		Level:  r.Level.String(),
		Msg:    r.Message,
		Fields: map[string]string{},
	}

	r.Attrs(func(a slog.Attr) bool {
		rec.Fields[a.Key] = a.Value.String()
		return true
	})

	*h.Records = append(*h.Records, rec)

	return nil
}

func (h *CollectHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CollectHandler{
		Inner:   h.Inner.WithAttrs(attrs),
		Records: h.Records,
	}
}

func (h *CollectHandler) WithGroup(name string) slog.Handler {
	return &CollectHandler{
		Inner:   h.Inner.WithGroup(name),
		Records: h.Records,
	}
}

type ctxKeyLogger struct{}

func LoggerFromCtx(ctx context.Context) *slog.Logger {
	if v := ctx.Value(ctxKeyLogger{}); v != nil {
		if l, ok := v.(*slog.Logger); ok {
			return l
		}
	}
	// fallback: global logger or no-op
	return slog.Default()
}

func CtxWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger{}, l)
}
