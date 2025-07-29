package req

import (
	"context"
	"errors"

	"github.com/imroc/req/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func NewTraceClient() *req.Client {
	tr := otel.Tracer("reqClient")
	c := req.C()
	setTracer(c, tr)
	return c
}

type apiNameType int

const apiNameKey apiNameType = iota

var BizError = errors.New("biz error")

func setTracer(c *req.Client, tracer trace.Tracer) {
	c.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(req *req.Request) (resp *req.Response, err error) {
			savedCtx := req.Context()
			defer func() {
				req.SetContext(savedCtx)
			}()
			apiName, ok := savedCtx.Value(apiNameKey).(string)
			if !ok {
				apiName = req.URL.Path
			}
			ctx, span := tracer.Start(savedCtx, apiName, trace.WithSpanKind(trace.SpanKindClient))
			defer span.End()
			prop := propagation.TraceContext{}
			prop.Inject(ctx, propagation.HeaderCarrier(req.Headers))
			req.SetContext(ctx)
			span.SetAttributes(
				attribute.String("http.url", req.URL.String()),
				attribute.String("http.method", req.Method),
				attribute.String("http.req.header", req.HeaderToString()),
			)
			if len(req.Body) > 0 {
				span.SetAttributes(attribute.String("http.req.body", string(req.Body)))
			}
			resp, err = rt.RoundTrip(req)
			if err != nil {
				span.RecordError(err)
				if errors.Is(err, BizError) {
					span.SetStatus(codes.Unset, err.Error())
				} else {
					span.SetStatus(codes.Error, err.Error())
				}
			}
			if resp.Response != nil {
				span.SetAttributes(
					attribute.Int("http.status_code", resp.Response.StatusCode),
					attribute.String("http.resp.header", resp.HeaderToString()),
					attribute.String("http.resp.body", resp.String()),
				)
			}
			return
		}
	})
}

func WithAPIName(ctx context.Context, name string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, apiNameKey, name)
}
