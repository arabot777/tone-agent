package http

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Client() *http.Client {
	return otelhttp.DefaultClient
}
