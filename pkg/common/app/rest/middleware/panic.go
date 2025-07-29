package middleware

import (
	"net/http"

	pkgErrors "github.com/pkg/errors"

	"tone/agent/pkg/common/app/rest"
)

func PanicAsError(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recoverd := recover()
			if recoverd == nil {
				return
			}

			if err, ok := recoverd.(error); ok {
				switch realErr := pkgErrors.Cause(err).(type) {
				case *rest.APIError:
					rest.RenderError(w, realErr)
					return
				default:
					panic(realErr)
				}
			} else {
				panic(recoverd)
			}
		}()
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

type withPanicOption func(err error) *PanicOption

type PanicOption struct {
	IsPanic    bool
	StatusCode int
	Error      string
}

func NewPanicOption(isPanic bool, statusCode int, err string) *PanicOption {
	return &PanicOption{IsPanic: isPanic, StatusCode: statusCode, Error: err}
}

func PanicAsErrorWithOption(panicOptionFn withPanicOption) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recoverd := recover()
				if recoverd == nil {
					return
				}
				if err, ok := recoverd.(error); ok {
					switch realErr := pkgErrors.Cause(err).(type) {
					case *rest.APIError:
						rest.RenderError(w, realErr)
						return
					default:
						panicOption := panicOptionFn(realErr)
						if panicOption != nil && !panicOption.IsPanic {
							w.WriteHeader(panicOption.StatusCode)
							_, _ = w.Write([]byte(panicOption.Error))
							return
						}
						panic(realErr)
					}
				} else {
					panic(recoverd)
				}
			}()
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
