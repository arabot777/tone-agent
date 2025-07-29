package rest

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type HTTPBundle struct {
	name       string
	router     http.Handler
	httpServer http.Server

	timeout      time.Duration // 针对 client 端的限制，到时间返回，并不能到时间停止服务端 request
	writeTimeout time.Duration // 对应 http server 的 writeTimeout
	readTimeout  time.Duration // 对应 http server 的 readTimeout
	withoutHTTP2 bool          // http server 关闭 http2

	port int
}

func New(opts ...Option) *HTTPBundle {
	defaults := getDefaults()
	api := &HTTPBundle{
		name:         defaults.Name,
		port:         defaults.port,
		timeout:      defaults.timeout,
		readTimeout:  defaults.readTimeOut,
		writeTimeout: defaults.writeTimeout,
	}

	for _, opt := range opts {
		opt(api)
	}

	return api
}

func (s *HTTPBundle) Type() string {
	return "HTTP"
}

func (s *HTTPBundle) Name() string {
	return s.name
}

func (s *HTTPBundle) Run(ctx context.Context) error {
	handler := s.router
	if s.timeout != 0 {
		handler = http.TimeoutHandler(s.router, s.timeout, "")
	}

	s.httpServer = http.Server{
		Addr:         ":" + strconv.Itoa(s.port),
		Handler:      handler,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		TLSNextProto: func() map[string]func(*http.Server, *tls.Conn, http.Handler) {
			if s.withoutHTTP2 {
				return make(map[string]func(*http.Server, *tls.Conn, http.Handler))
			}
			return nil
		}(),
	}

	if err := s.httpServer.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}

	return nil
}

func (s *HTTPBundle) Stop() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			fmt.Println(ctx, err)
		}
		cancel()
	}()

	return ctx
}
