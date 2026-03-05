package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const traceHeader = "X-Trace-ID"

const TraceIDKey contextKey = "trace_id"

func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(traceHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		r = r.WithContext(ctx)
		r.Header.Set(traceHeader, traceID)
		w.Header().Set(traceHeader, traceID)

		next.ServeHTTP(w, r)
	})
}

func TraceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(TraceIDKey).(string)
	return v
}
