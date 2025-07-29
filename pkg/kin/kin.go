package kin

import (
	"context"
	"net/http"
	"strings"
	"time"

	"tone/agent/pkg/common/gin/code"
	"tone/agent/pkg/common/gin/code/rpccode"

	"github.com/gin-gonic/gin"
)

const CodeOK = 0

type Context interface {
	Shortcut
	GetContext() context.Context
	context.Context
}

type Shortcut interface {
	ReplyOK()
	Reply(data interface{})
	ReplyErr(code int, hints ...string)
	ReplyErrWithStatusCode(statusCode int, code int, hints ...string)
	ReplyErrCoder(err error)
	ReplyForbidden()
	ReplyUnauthorized()
	InternalErr()
	Notfound()
	ReplyRequestErr(hints ...string)
	ReplyCoder(err error, data interface{})

	Param(key string) string
}

func NewCtx(c *gin.Context, ctxs ...context.Context) Context {
	if len(ctxs) > 0 {
		return &ctx{Context: c, stdCtx: ctxs[0]}
	}
	return &ctx{Context: c, stdCtx: context.Background()}
}

type ctx struct {
	*gin.Context
	stdCtx context.Context
}

func (c *ctx) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *ctx) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *ctx) Err() error {
	return c.Context.Err()
}

func (c *ctx) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

func (c *ctx) GetContext() context.Context {
	return c.Context
}

func (c *ctx) ReplyForbidden() {
	c.Status(http.StatusForbidden)
	c.Abort()
}

func (c *ctx) ReplyUnauthorized() {
	c.Status(http.StatusUnauthorized)
	c.Abort()
}

func (c *ctx) Reply(data interface{}) {
	c.JSON(http.StatusOK, &Message{
		Code: CodeOK,
		Data: data,
	})
}

func (c *ctx) ReplyCoder(err error, data interface{}) {
	codeValue, _ := rpccode.RenderCoder(err)
	c.JSON(http.StatusOK, &Message{
		Code: codeValue,
		Data: data,
	})
}

func (c *ctx) ReplyErrWithStatusCode(statusCode, code int, hints ...string) {
	msg := &MessageError{
		Code:  code,
		Error: rpccode.Render(code),
	}
	if len(hints) > 0 {
		msg.Error.Msg = strings.Join(hints, ", ")
	}
	c.JSON(statusCode, msg)
}

func (c *ctx) ReplyOK() {
	c.JSON(http.StatusOK, &Message{
		Code: CodeOK,
	})
}

func (c *ctx) ReplyErr(code int, hints ...string) {
	msg := &MessageError{
		Code:  code,
		Error: rpccode.Render(code),
	}
	if len(hints) > 0 {
		msg.Error.Msg = strings.Join(hints, ", ")
	}
	c.JSON(http.StatusOK, msg)
}

func (c *ctx) ReplyErrCoder(err error) {
	codeValue, errInfo := rpccode.RenderCoder(err)
	msg := &MessageError{
		Code:  codeValue,
		Error: errInfo,
	}
	c.JSON(http.StatusOK, msg)
}

func (c *ctx) ReplyRequestErr(hints ...string) {
	msg := &MessageError{
		Code:  code.ErrBadParams,
		Error: rpccode.Render(code.ErrBadParams),
	}
	if len(hints) > 0 {
		msg.Error.Msg = strings.Join(hints, ", ")
	}
	c.JSON(http.StatusOK, msg)
}

func (c *ctx) InternalErr() {
	c.JSON(http.StatusOK, &MessageError{
		Code:  code.ErrInternal,
		Error: rpccode.Render(code.ErrInternal),
	})
}

func (c *ctx) Notfound() {
	c.JSON(http.StatusOK, &MessageError{
		Code:  code.ErrNotFound,
		Error: rpccode.Render(code.ErrNotFound),
	})
}

type Message struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
}

type MessageError struct {
	Code  int             `json:"code"`
	Error rpccode.ErrInfo `json:"error,omitempty"`
}
