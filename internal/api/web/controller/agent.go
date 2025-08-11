package controller

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"os"
	"path/filepath"
	"time"
	"tone/agent/internal/api/service"
	"tone/agent/internal/pkg/common/code"
	"tone/agent/internal/pkg/infra"
	"tone/agent/internal/pkg/model"
	"tone/agent/internal/pkg/service/deep"
	"tone/agent/internal/pkg/service/drawing"
	"tone/agent/internal/pkg/service/journal"
	"tone/agent/pkg/common/logger"
	"tone/agent/pkg/utils"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/multiagent/host"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
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

func (c *AgentController) Ok(ctx context.Context, req *app.RequestContext) {
	req.JSON(consts.StatusOK, c.agentService.Ok())
}

func (c *AgentController) WebUI(ctx context.Context, req *app.RequestContext) {
	content, err := webContent.ReadFile("ui/index.html")
	if err != nil {
		req.String(consts.StatusNotFound, "File not found")
		return
	}
	req.Response.Header.Set("Content-Type", "text/html")
	req.Data(consts.StatusOK, "text/html", content)
}

func (c *AgentController) WebUIFile(ctx context.Context, req *app.RequestContext) {
	file := req.Param("file")
	content, err := webContent.ReadFile("ui/" + file)
	if err != nil {
		req.String(consts.StatusNotFound, "File not found")
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(file))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req.Response.Header.Set("Content-Type", contentType)
	req.Data(consts.StatusOK, contentType, content)
}

func (c *AgentController) Einoagent(ctx context.Context, req *app.RequestContext) {
	id := string(req.Query("id"))
	message := string(req.Query("message"))
	if id == "" || message == "" {
		req.JSON(consts.StatusBadRequest, code.ReqParseErr.Msg("missing id or message"))
		return
	}
	// 创建带有更长超时的新 context
	ctx, cancel := context.WithTimeout(ctx, 100*time.Minute)
	defer cancel()
	logger.Infof(ctx, "[Chat] Starting chat with ID: %s, Message: %s", id, message)

	sr, err := c.agentService.Einoagent(ctx, id, message)
	if err != nil {
		logger.Errorf(ctx, "[Chat] Error running agent: %v", err)
		req.JSON(consts.StatusInternalServerError, err)
		return
	}

	// 设置 SSE 响应头
	req.Response.Header.Set("Content-Type", "text/event-stream")
	req.Response.Header.Set("Cache-Control", "no-cache")
	req.Response.Header.Set("Connection", "keep-alive")
	req.Response.Header.Set("Access-Control-Allow-Origin", "*")

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
			req.SetBodyString("data: " + msg.Content + "\n\n")
			// TODO: 需要适配 Hertz 的 SSE 流式响应
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

func (a *AgentController) Drawing(ctx context.Context, c *app.RequestContext) {
	// 设置响应头（NewStream 会自动设置部分头，但建议显式声明）
	c.SetContentType("text/event-stream; charset=utf-8")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")
	c.Response.Header.Set("Access-Control-Allow-Origin", "*")

	c.SetStatusCode(http.StatusOK)
	// 初始化一个sse writer
	w := sse.NewWriter(c)
	defer w.Close()

	// 请求体校验
	req := new(model.ChatRequest)
	err := c.BindAndValidate(req)
	if err != nil {
		return
	}
	logger.Infof(ctx, "ChatStream_begin, req: %v", req)

	// 根据前端参数生成Graph State
	genFunc := func(ctx context.Context) *model.State {
		return &model.State{
			MaxPlanIterations:             req.MaxPlanIterations,
			MaxStepNum:                    req.MaxStepNum,
			Messages:                      req.Messages,
			Goto:                          deep.Coordinator,
			EnableBackgroundInvestigation: req.EnableBackgroundInvestigation,
		}
	}
	// Build Graph
	r := drawing.Builder[string, string, *model.State](ctx, genFunc)

	// Run Graph
	_, err = r.Stream(ctx, deep.Coordinator,
		compose.WithCheckPointID(req.ThreadID), // 指定Graph的CheckPointID
		// 中断后，获取用户的edit_plan信息
		compose.WithStateModifier(func(ctx context.Context, path compose.NodePath, state any) error {
			s := state.(*model.State)
			s.InterruptFeedback = req.InterruptFeedback
			if req.InterruptFeedback == "edit_plan" {
				s.Messages = append(s.Messages, req.Messages...)
			}
			logger.Infof(ctx, "ChatStream_modf", "path", path.GetPath(), "state", state)
			return nil
		}),
		// 连接LoggerCallback
		compose.WithCallbacks(&infra.LoggerCallback{
			ID:  req.ThreadID,
			SSE: w,
		}),
	)

	// 将interrupt信号传递到前端
	if info, ok := compose.ExtractInterruptInfo(err); ok {
		logger.Infof(ctx, "ChatStream_interrupt", "info", info)
		data := &model.ChatResp{
			ThreadID:     req.ThreadID,
			ID:           "human_feedback:" + utils.RandStr(20),
			Role:         "assistant",
			Content:      "检查计划",
			FinishReason: "interrupt",
			Options: []map[string]any{
				{
					"text":  "编辑计划",
					"value": "edit_plan",
				},
				{
					"text":  "开始执行",
					"value": "accepted",
				},
			},
		}
		dB, _ := json.Marshal(data)
		w.WriteEvent("", "interrupt", dB)
	}
	if err != nil {
		logger.Errorf(ctx, "ChatStream_error, err: %v", err)
	}
}

func (a *AgentController) Researcher(ctx context.Context, c *app.RequestContext) {
	// 设置响应头（NewStream 会自动设置部分头，但建议显式声明）
	c.SetContentType("text/event-stream; charset=utf-8")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")
	c.Response.Header.Set("Access-Control-Allow-Origin", "*")

	c.SetStatusCode(http.StatusOK)
	// 初始化一个sse writer
	w := sse.NewWriter(c)
	defer w.Close()

	// 请求体校验
	req := new(model.ChatRequest)
	err := c.BindAndValidate(req)
	if err != nil {
		return
	}
	logger.Infof(ctx, "ChatStream_begin, req: %v", req)

	// 根据前端参数生成Graph State
	genFunc := func(ctx context.Context) *model.State {
		return &model.State{
			MaxPlanIterations:             req.MaxPlanIterations,
			MaxStepNum:                    req.MaxStepNum,
			Messages:                      req.Messages,
			Goto:                          deep.Coordinator,
			EnableBackgroundInvestigation: req.EnableBackgroundInvestigation,
		}
	}

	// Build Graph
	r := deep.Builder[string, string, *model.State](ctx, genFunc)

	// Run Graph
	_, err = r.Stream(ctx, deep.Coordinator,
		compose.WithCheckPointID(req.ThreadID), // 指定Graph的CheckPointID
		// 中断后，获取用户的edit_plan信息
		compose.WithStateModifier(func(ctx context.Context, path compose.NodePath, state any) error {
			s := state.(*model.State)
			s.InterruptFeedback = req.InterruptFeedback
			if req.InterruptFeedback == "edit_plan" {
				s.Messages = append(s.Messages, req.Messages...)
			}
			logger.Infof(ctx, "ChatStream_modf", "path", path.GetPath(), "state", state)
			return nil
		}),
		// 连接LoggerCallback
		compose.WithCallbacks(&infra.LoggerCallback{
			ID:  req.ThreadID,
			SSE: w,
		}),
	)

	// 将interrupt信号传递到前端
	if info, ok := compose.ExtractInterruptInfo(err); ok {
		logger.Infof(ctx, "ChatStream_interrupt", "info", info)
		data := &model.ChatResp{
			ThreadID:     req.ThreadID,
			ID:           "human_feedback:" + utils.RandStr(20),
			Role:         "assistant",
			Content:      "检查计划",
			FinishReason: "interrupt",
			Options: []map[string]any{
				{
					"text":  "编辑计划",
					"value": "edit_plan",
				},
				{
					"text":  "开始执行",
					"value": "accepted",
				},
			},
		}
		dB, _ := json.Marshal(data)
		w.WriteEvent("", "interrupt", dB)
	}
	if err != nil {
		logger.Errorf(ctx, "ChatStream_error, err: %v", err)
	}
}

func (c *AgentController) Journal(ctx context.Context, req *app.RequestContext) {
	// TODO SSE 返回

	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	openAIBaseURL := os.Getenv("OPENAI_BASE_URL")
	openAIModelName := os.Getenv("OPENAI_MODEL_NAME")

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
