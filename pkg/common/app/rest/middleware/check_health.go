package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CheckHealth() gin.HandlerFunc {
	f := func(g *gin.Context) {
		if g.Request.Method == http.MethodGet &&
			strings.EqualFold(g.Request.URL.Path, "/health") {
			g.String(http.StatusOK, "ok")
			return
		}
		g.Next()
	}

	return f
}
