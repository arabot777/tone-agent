package auth_middleware

import (
	"github.com/gin-gonic/gin"
)

func AuthPayloadToHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// payload, err := auth.GetCurrentUserV2(c)
		// if err != nil {
		// 	logger.Errorf(c, "get current user error: %v", err)
		// 	_ = c.AbortWithError(http.StatusUnauthorized, err)
		// 	return
		// }
		// c.Request.Header.Set(auth.UserIdHeaderKey, strconv.FormatUint(payload.UserID, 10))
		// c.Request.Header.Set(auth.OrgIdHeaderKey, strconv.FormatUint(payload.OrgID, 10))
		c.Next()
	}
}

func TryAuthPayloadToHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// payload, err := auth.GetCurrentUserV2(c)
		// if err != nil {
		// 	return
		// }
		// c.Request.Header.Set(auth.UserIdHeaderKey, strconv.FormatUint(payload.UserID, 10))
		// c.Request.Header.Set(auth.OrgIdHeaderKey, strconv.FormatUint(payload.OrgID, 10))
		c.Next()
	}
}
