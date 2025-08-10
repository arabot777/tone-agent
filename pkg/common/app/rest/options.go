package rest

import (
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
)

type Option func(*HTTPBundle)

func Port(port int) Option {
	return func(bundle *HTTPBundle) {
		bundle.port = port
	}
}

func WithoutHTTP2(w bool) Option {
	return func(bundle *HTTPBundle) {
		bundle.withoutHTTP2 = w
	}
}

func Timeout(t time.Duration) Option {
	return func(bundle *HTTPBundle) {
		bundle.timeout = t
	}
}

func WithRouter(router http.Handler) Option {
	return func(bundle *HTTPBundle) {
		bundle.router = router
	}
}

func ReadTimeout(t time.Duration) Option {
	return func(bundle *HTTPBundle) {
		bundle.readTimeout = t
	}
}

func WriteTimeout(t time.Duration) Option {
	return func(bundle *HTTPBundle) {
		bundle.writeTimeout = t
	}
}

func WithHertzServer(hertzServer *server.Hertz) Option {
	return func(bundle *HTTPBundle) {
		bundle.hertzServer = hertzServer
	}
}
