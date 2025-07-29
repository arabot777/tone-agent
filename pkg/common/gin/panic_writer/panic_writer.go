package panic_writer

import (
	"context"
	"io"
)

type Logger func(ctx context.Context, format string, v ...interface{})

type PanicWriter struct {
	io.Writer
	Logger
}

func (w *PanicWriter) Write(b []byte) (int, error) {
	w.Logger(context.Background(), string(b))
	return len(b), nil
}
