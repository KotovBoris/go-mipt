//go:build !solution

package requestlog

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/felixge/httpsnoop"
	"go.uber.org/zap"
)

var requestIDCounter uint64

func Log(l *zap.Logger) func(next http.Handler) http.Handler {
	logger := l

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := atomic.AddUint64(&requestIDCounter, 1)
			reqIDStr := strconv.FormatUint(reqID, 10)

			path := r.URL.Path
			method := r.Method
			commonFields := []zap.Field{
				zap.String("path", path),
				zap.String("method", method),
				zap.String("request_id", reqIDStr),
			}

			logger.Info("request started", commonFields...)

			start := time.Now()

			var capturedStatusCode int
			var panicked bool = true

			defer func() {
				duration := time.Since(start)
				endFields := append(commonFields, zap.Duration("duration", duration))

				if r := recover(); r != nil {
					logger.Error("request panicked", append(endFields, zap.Any("panic", r))...)
					panic(r)
				}

				if !panicked {
					endFields = append(endFields, zap.Int("status_code", capturedStatusCode))
					logger.Info("request finished", endFields...)
				}
			}()

			metrics := httpsnoop.CaptureMetrics(next, w, r)

			panicked = false
			capturedStatusCode = metrics.Code
		})
	}
}

func Log_Alternative(l *zap.Logger) func(next http.Handler) http.Handler {
	logger := l

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := atomic.AddUint64(&requestIDCounter, 1)
			reqIDStr := strconv.FormatUint(reqID, 10)
			path := r.URL.Path
			method := r.Method
			commonFields := []zap.Field{
				zap.String("path", path),
				zap.String("method", method),
				zap.String("request_id", reqIDStr),
			}

			logger.Info("request started", commonFields...)
			start := time.Now()

			wrappedWriter := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			defer func() {
				duration := time.Since(start)
				endFields := append(commonFields, zap.Duration("duration", duration))

				if rec := recover(); rec != nil {
					endFields = append(endFields, zap.Int("status_code", wrappedWriter.statusCode))
					logger.Error("request panicked", append(endFields, zap.Any("panic", rec))...)
					panic(rec)
				} else {
					endFields = append(endFields, zap.Int("status_code", wrappedWriter.statusCode))
					logger.Info("request finished", endFields...)
				}
			}()

			next.ServeHTTP(wrappedWriter, r)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (sr *statusRecorder) WriteHeader(code int) {
	if !sr.wroteHeader {
		sr.statusCode = code
		sr.wroteHeader = true
	}
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if !sr.wroteHeader {
		sr.WriteHeader(http.StatusOK)
	}
	return sr.ResponseWriter.Write(b)
}

func (sr *statusRecorder) Flush() {
	if flusher, ok := sr.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
