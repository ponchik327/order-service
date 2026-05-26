package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// loggingResponseWriter перехватывает статус и опционально тело ответа.
// Body буферизуется только когда статус >= 400 — на успешных ответах
// мы тело не логируем, поэтому нет смысла платить копией каждого
// Write'а в обёрнутый буфер (на больших ответах это 100KB+ на запрос).
type loggingResponseWriter struct {
	http.ResponseWriter
	status      int
	captureBody bool
	body        bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.captureBody = code >= 400
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.captureBody {
		lrw.body.Write(b)
	}
	return lrw.ResponseWriter.Write(b)
}

// reset подготавливает обёртку к следующему использованию из пула.
func (lrw *loggingResponseWriter) reset(w http.ResponseWriter) {
	lrw.ResponseWriter = w
	lrw.status = http.StatusOK
	lrw.captureBody = false
	lrw.body.Reset()
}

// lrwPool переиспользует обёртку между запросами. lrw уходит из стека
// на кучу из-за конверсии в http.ResponseWriter (interface) при вызове
// next.ServeHTTP, и в pprof виден отдельной аллокацией в горячем пути.
var lrwPool = sync.Pool{
	New: func() any { return &loggingResponseWriter{} },
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := lrwPool.Get().(*loggingResponseWriter)
		lrw.reset(w)
		defer func() {
			// Очищаем ссылку на внешний writer, чтобы он не удерживался
			// дольше нужного, и возвращаем обёртку в пул.
			lrw.ResponseWriter = nil
			lrwPool.Put(lrw)
		}()

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
