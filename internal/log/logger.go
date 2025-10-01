package log

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"
)

// AccessHTTPMiddleware is a middleware which outputs accept request log
// and finish request log.
func AccessHTTPMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if logger == nil {
				logger = slog.Default()
			}
			ctx := r.Context()
			logger.InfoContext(ctx, "accept request", slog.Group("request", slog.String("method", r.Method), slog.String("endpoint", r.URL.Path)))
			next.ServeHTTP(w, r)
			logger.InfoContext(ctx, "done request", slog.Group("request", slog.String("method", r.Method), slog.String("endpoint", r.URL.Path)))
		})
	}
}

var defaultLoggerOptions = slog.HandlerOptions{
	AddSource: true,
	ReplaceAttr: func(_ []string, att slog.Attr) slog.Attr {
		if att.Key == slog.SourceKey {
			source, _ := att.Value.Any().(*slog.Source)
			if source != nil {
				source.File = filepath.Base(source.File)
			}
		}
		if att.Key == slog.TimeKey {
			att.Key = "ts"
		}
		return att
	},
}

func NewLogger(w io.Writer, opts *slog.HandlerOptions, logLevel slog.Level) *slog.Logger {
	if opts == nil {
		opts = &defaultLoggerOptions
	}

	opts.Level = logLevel
	return slog.New(NewTraceIDHandler(slog.NewJSONHandler(w, opts)))
}

var generator rand.Rand

// newTraceID generates a 10-digit random number.
func newTraceID() string {
	n := generator.Uint32()
	return fmt.Sprintf("%010d", n)
}

type traceIDKey string

var defaultTraceIDKey traceIDKey = "traceID"

// WithTraceIDContext return a copy of parent which is assosiated
// a random trace ID.
func WithTraceIDContext(parent context.Context) context.Context {
	return context.WithValue(parent, defaultTraceIDKey, newTraceID())
}

// ExtractTraceID return the trace ID assosiated to the context.
func ExtractTraceID(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("context is nil pointer")
	}
	v := ctx.Value(defaultTraceIDKey)
	traceID, ok := v.(string)
	if !ok {
		return "", errors.New("trace ID is not found")
	}
	return traceID, nil
}

// TraceIDMiddleware is a middleware which inject a trace ID into
// the request's context.
func TraceIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(WithTraceIDContext(r.Context()))
			next.ServeHTTP(w, r)
		})
	}
}

// TraceIDHandler is an extend slog.Handler that add trace ID in Records.
type TraceIDHandler struct {
	parent slog.Handler
}

// NewTraceIDHandler creates a TraceIDHandler using the given parent slog.Handler.
// If a parent is nil, panic occurs.
func NewTraceIDHandler(parent slog.Handler) *TraceIDHandler {
	if parent == nil {
		panic("panic no parent handler")
	}
	return &TraceIDHandler{parent}
}

var _ slog.Handler = (*TraceIDHandler)(nil)

// Enabled call parent.Enabled without doing anything.
func (h *TraceIDHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.parent.Enabled(ctx, l)
}

// Handle add trace ID in context to Record and call parent.Handle.
func (h *TraceIDHandler) Handle(ctx context.Context, r slog.Record) error {
	traceID, err := ExtractTraceID(ctx)
	if err != nil {
		return h.parent.Handle(ctx, r)
	}
	r.AddAttrs(slog.String("traceID", traceID))
	return h.parent.Handle(ctx, r)
}

// WithAttrs call parent.WithAttrs without doing anything.
func (h *TraceIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.parent.WithAttrs(attrs)
}

// WithGroup call parent.WithGroup without doing anything.
func (h *TraceIDHandler) WithGroup(name string) slog.Handler {
	return h.parent.WithGroup(name)
}

// WithTraceIDLogger return a copy of parent logger which is assosiated
// a random trace ID.
func WithTraceIDLogger(parent *slog.Logger) *slog.Logger {
	return parent.With(slog.String("traceID", newTraceID()))
}

func init() {
	generator = *rand.New(rand.NewSource(time.Now().UnixNano()))
}
