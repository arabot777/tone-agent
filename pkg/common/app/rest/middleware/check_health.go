package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func CheckHealth() app.HandlerFunc {
	f := func(ctx context.Context, c *app.RequestContext) {
		if string(c.Method()) == "GET" &&
			strings.EqualFold(string(c.Path()), "/health") {
			c.String(consts.StatusOK, "ok")
			return
		}
		c.Next(ctx)
	}

	return f
}
