package wrapper

import (
	"net/http"
	"strings"

	"tone/agent/pkg/common/gin/code"
	"tone/agent/pkg/common/gin/code/rpccode"

	"github.com/gin-gonic/gin"
)

type Message struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
}

type MessageError struct {
	Code  int             `json:"code"`
	Error rpccode.ErrInfo `json:"error,omitempty"`
}

func Reply(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, &Message{
		Code: code.OK,
		Data: data,
	})
}

func ReplyErrWithStatusCode(ctx *gin.Context, statusCode, code int, hints ...string) {
	msg := &MessageError{
		Code: code,
	}
	if len(hints) > 0 {
		msg.Error.Msg = strings.Join(hints, ", ")
	}
	ctx.JSON(statusCode, msg)
}

func ReplyOK(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &Message{
		Code: code.OK,
	})
}

func ReplyErr(ctx *gin.Context, code int, hints ...string) {
	msg := &MessageError{
		Code: code,
	}
	if len(hints) > 0 {
		msg.Error.Msg = strings.Join(hints, ", ")
	}
	ctx.JSON(http.StatusOK, msg)
}

func ReplyErrCoder(ctx *gin.Context, err error) {
	codeValue, errInfo := rpccode.RenderCoder(err)
	if codeValue == code.ErrSignDecode {
		ctx.JSON(http.StatusUnauthorized, errInfo)
		return
	}
	msg := &MessageError{
		Code:  codeValue,
		Error: errInfo,
	}
	ctx.JSON(http.StatusOK, msg)
}
