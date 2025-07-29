package middleware

import (
	"errors"
	"net/http"
)

type bufferWriter struct {
	http.ResponseWriter

	buffer []byte
}

func (bw *bufferWriter) Write(data []byte) (int, error) {
	if bw.buffer != nil {
		return 0, errors.New("server rendered twice")
	}

	bw.buffer = make([]byte, len(data))
	copy(bw.buffer, data)

	return len(data), nil
}

func (bw *bufferWriter) Flush() {
	if bw.buffer != nil {
		_, _ = bw.ResponseWriter.Write(bw.buffer)
	}
}

func RenderOnce(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		bw := &bufferWriter{
			ResponseWriter: w,
			buffer:         nil,
		}
		h.ServeHTTP(bw, r)
		bw.Flush()
	}
	return http.HandlerFunc(fn)
}
