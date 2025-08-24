package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b) // сохраняем копию
	return lrw.ResponseWriter.Write(b)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}

		log.Printf("➡️  %s %s", r.Method, r.RequestURI)

		defer func() {
			if rec := recover(); rec != nil {
				// Перехватываем панику и отдаем JSON ошибку
				log.Printf("🔥 Panic: %v", rec)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "Internal Server Error",
				})
			}
		}()

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		if lrw.status >= 400 {
			// Ошибочный ответ → логируем тело
			log.Printf("⬅️  %s %s %d (%s)\nResponse: %s",
				r.Method, r.RequestURI, lrw.status, duration, lrw.body.String())
		} else {
			log.Printf("⬅️  %s %s %d (%s)",
				r.Method, r.RequestURI, lrw.status, duration)
		}
	})
}
