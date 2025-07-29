package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"tone/agent/pkg/common/logger"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body      *bytes.Buffer
	maxSize   int
	truncated bool
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	// 先写入到原始的 ResponseWriter
	size, err := w.ResponseWriter.Write(b)
	if err != nil {
		return size, err
	}

	// 如果 buffer 还没达到最大大小，继续写入
	if w.body.Len() < w.maxSize {
		remaining := w.maxSize - w.body.Len()
		if len(b) > remaining {
			w.body.Write(b[:remaining])
			w.truncated = true
		} else {
			w.body.Write(b)
		}
	}

	return size, nil
}

func LogWithWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete {
			// 检查 Content-Length 头部
			contentLength := c.Request.ContentLength
			contentType := c.Request.Header.Get("Content-Type")

			if contentLength > 0 && contentLength < 1024*4 && !strings.Contains(contentType, "multipart/form-data") {
				// 请求体
				var bodyBytes []byte
				if c.Request.Body != nil {
					bodyBytes, _ = c.GetRawData()
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					m := make(map[string]interface{})
					_ = json.Unmarshal(bodyBytes, &m)
					delete(m, "password")
					delete(m, "old_password")
					delete(m, "new_password")
					jsonBytes, _ := json.Marshal(&m)
					logger.Infof(c, "request url: %s request body: %s", c.Request.RequestURI, string(jsonBytes))
				}
			} else {
				logger.Infof(c, "request url: %s request body: [too large or binary content]", c.Request.RequestURI)
			}
		}
		// 存储 put 接口一般是文件上传，不打印请求体
		if c.Request.Method == http.MethodPut {
			logger.Infof(c, "request url: %s", c.Request.RequestURI)
		}
		// 使用限制大小的 buffer 来记录响应
		const maxLogSize = 8 * 1024 // 4KB
		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(make([]byte, 0, maxLogSize)),
			maxSize:        maxLogSize,
		}
		c.Writer = blw
		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.Request.Host

		body := "skipLogging"
		if skip, exists := c.Get("skipLogging"); !exists || !skip.(bool) {
			// 检查响应的 Content-Type 和请求路径
			if blw.body.Len() < 1024*4 {
				body = blw.body.String()
			} else {
				body = "[too large or binary content]"
			}
		}

		logger.Infof(c, "response-info: %3d | %13v | %15s | %s | %s | %s|",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
			body,
		)
	}
}
