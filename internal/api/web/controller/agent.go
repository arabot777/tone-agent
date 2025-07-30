package controller

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"tone/agent/internal/api/service"
	"tone/agent/internal/pkg/common/code"
	"tone/agent/internal/pkg/service/journal"
	"tone/agent/pkg/common/logger"

	"github.com/cloudwego/eino/flow/agent/multiagent/host"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/gin-gonic/gin"
)

//go:embed ui
var webContent embed.FS

type AgentController struct {
	agentService *service.AgentService
}

func NewAgentController() *AgentController {
	return &AgentController{
		agentService: service.NewAgentService(),
	}
}

func (c *AgentController) Ok(g *gin.Context) {
	g.JSON(http.StatusOK, c.agentService.Ok())
}

func (c *AgentController) WebUI(g *gin.Context) {
	content, err := webContent.ReadFile("ui/index.html")
	if err != nil {
		g.String(consts.StatusNotFound, "File not found")
		return
	}
	g.Header("Content-Type", "text/html")
	g.Data(http.StatusOK, "text/html", content)
}

func (c *AgentController) WebUIFile(g *gin.Context) {
	file := g.Param("file")
	content, err := webContent.ReadFile("ui/" + file)
	if err != nil {
		g.String(consts.StatusNotFound, "File not found")
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(file))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	g.Header("Content-Type", contentType)
	g.Data(http.StatusOK, contentType, content)
}

func (c *AgentController) Einoagent(g *gin.Context) {
	id := g.Query("id")
	message := g.Query("message")
	if id == "" || message == "" {
		g.JSON(http.StatusBadRequest, code.ReqParseErr.Msg("missing id or message"))
		return
	}
	// ctx := g.Request.Context()
	// 创建带有更长超时的新 context
	ctx, cancel := context.WithTimeout(g.Request.Context(), 100*time.Minute)
	defer cancel()
	logger.Infof(ctx, "[Chat] Starting chat with ID: %s, Message: %s", id, message)

	sr, err := c.agentService.Einoagent(ctx, id, message)
	if err != nil {
		logger.Errorf(ctx, "[Chat] Error running agent: %v", err)
		g.JSON(consts.StatusInternalServerError, err)
		return
	}

	// 设置 SSE 响应头
	g.Header("Content-Type", "text/event-stream")
	g.Header("Cache-Control", "no-cache")
	g.Header("Connection", "keep-alive")
	g.Header("Access-Control-Allow-Origin", "*")

	defer func() {
		sr.Close()
		// if flusher, ok := g.Writer.(http.Flusher); ok {
		// 	flusher.Flush()
		// }

		logger.Infof(ctx, "[Chat] Finished chat with ID: %s", id)
	}()

outer:
	for {
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "[Chat] Context done for chat ID: %s", id)
			return
		default:
			msg, err := sr.Recv()
			if errors.Is(err, io.EOF) {
				logger.Infof(ctx, "[Chat] EOF received for chat ID: %s", id)
				break outer
			}
			if err != nil {
				logger.Infof(ctx, "[Chat] Error receiving message: %v", err)
				break outer
			}
			// 发送 SSE 格式数据
			// g.Writer.WriteString("data: " + msg.Content + "\n\n")
			// flusher.Flush()
			g.SSEvent("data", msg)
			g.Writer.Flush()
			// err = s.Publish(&sse.Event{
			// 	Data: []byte(msg.Content),
			// })
			// if err != nil {
			// 	logger.Errorf(g, "[Chat] Error publishing message: %v", err)
			// 	break outer
			// }
		}
	}
}

func (c *AgentController) Journal(g *gin.Context) {
	// TODO SSE 返回

	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	openAIBaseURL := os.Getenv("OPENAI_BASE_URL")
	openAIModelName := os.Getenv("OPENAI_MODEL_NAME")

	ctx := g.Request.Context()

	h, err := journal.NewJournal(ctx, openAIBaseURL, openAIAPIKey, openAIModelName)
	if err != nil {
		panic(err)
	}

	writer, err := journal.NewWriteJournalSpecialist(ctx, openAIBaseURL, openAIAPIKey, openAIModelName)
	if err != nil {
		panic(err)
	}

	reader, err := journal.NewReadJournalSpecialist(ctx)
	if err != nil {
		panic(err)
	}

	answerer, err := journal.NewAnswerWithJournalSpecialist(ctx, openAIBaseURL, openAIAPIKey, openAIModelName)
	if err != nil {
		panic(err)
	}

	hostMA, err := host.NewMultiAgent(ctx, &host.MultiAgentConfig{
		Host: *h,
		Specialists: []*host.Specialist{
			writer,
			reader,
			answerer,
		},
	})
	if err != nil {
		panic(err)
	}

	cb := &logCallback{}

	for { // 多轮对话，除非用户输入了 "exit"，否则一直循环
		println("\n\nYou: ") // 提示轮到用户输入了

		var message string
		scanner := bufio.NewScanner(os.Stdin) // 获取用户在命令行的输入
		for scanner.Scan() {
			message += scanner.Text()
			break
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		if message == "exit" {
			return
		}

		msg := &schema.Message{
			Role:    schema.User,
			Content: message,
		}

		out, err := hostMA.Stream(ctx, []*schema.Message{msg}, host.WithAgentCallbacks(cb))
		if err != nil {
			panic(err)
		}

		println("\nAnswer:")

		for {
			msg, err := out.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
			}

			print(msg.Content)
		}

		out.Close()
	}
}

type logCallback struct{}

func (l *logCallback) OnHandOff(ctx context.Context, info *host.HandOffInfo) context.Context {
	logger.Infof(ctx, "HandOff to %s with argument %s", info.ToAgentName, info.Argument)
	return ctx
}
