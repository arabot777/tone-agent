package auth

import (
	"fmt"
	"strings"
	"time"
	"tone/agent/pkg/common/logger"
	"tone/agent/pkg/kin"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const SECRET_KEY = "t&*^%$one#v587&"

type jWTClaims struct {
	jwt.RegisteredClaims
	UserId string `json:"userId"`
	OrgId  string `json:"orgId"`
	Admin  bool   `json:"admin"`
	Token  string `json:"token"`
}

func SignToken(userId string, orgId string, admin bool, evegenToken string) (string, error) {
	claims := &jWTClaims{
		UserId: userId,
		OrgId:  orgId,
		Admin:  admin,
		Token:  evegenToken,
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now()),                         // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 过期时间 7天
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Bearer %s", signed), nil
}

const contextJWTClaims = "ContextJWTClaims"

func ValidateJWT(c *gin.Context) {
	ctx := kin.NewCtx(c)
	tokenString := c.GetHeader("Authorization")
	if !strings.HasPrefix(tokenString, "Bearer ") {
		ctx.ReplyUnauthorized()
		logger.Infof(c, "token is invalid, tokenString is: %v", tokenString)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	jwtClaims := new(jWTClaims)
	token, err := jwt.ParseWithClaims(tokenString, jwtClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SECRET_KEY), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		logger.Warnf(c, "token is invalid: %v. tokenString is: %s", err.Error(), tokenString)
		ctx.ReplyUnauthorized()
		return
	}
	if claims, ok := token.Claims.(*jWTClaims); ok && token.Valid {
		c.Set(contextJWTClaims, claims)
	} else {
		logger.Errorf(c, "token is invalid: %v, tokenString is: %s", err, tokenString)
		ctx.ReplyUnauthorized()
		return
	}
}

func ValidateJWTAdmin(c *gin.Context) {
	ctx := kin.NewCtx(c)
	tokenString := c.GetHeader("Authorization")
	if !strings.HasPrefix(tokenString, "Bearer ") {
		ctx.ReplyUnauthorized()
		logger.Infof(c, "token is invalid, tokenString is: %v", tokenString)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	jwtClaims := new(jWTClaims)
	token, err := jwt.ParseWithClaims(tokenString, jwtClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SECRET_KEY), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		logger.Warnf(c, "token is invalid: %v. tokenString is: %s", err.Error(), tokenString)
		ctx.ReplyUnauthorized()
		return
	}
	if claims, ok := token.Claims.(*jWTClaims); ok && token.Valid {
		if !claims.Admin {
			logger.Errorf(c, "user is not admin, tokenString is: %s", tokenString)
			ctx.ReplyUnauthorized()
			return
		}
		c.Set(contextJWTClaims, claims)
	} else {
		logger.Errorf(c, "token is invalid: %v, tokenString is: %s", err, tokenString)
		ctx.ReplyUnauthorized()
		return
	}
}

func GetUserInfo(c *gin.Context) *jWTClaims {
	if v, ok := c.Get(contextJWTClaims); ok {
		if value, ok := v.(*jWTClaims); ok {
			return value
		}
	}
	return nil
}
