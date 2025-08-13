package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"tone/agent/internal/pkg/model"
	"tone/agent/pkg/common/logger"
	"tone/agent/pkg/utils"

	"github.com/cloudwego/eino/callbacks"
	ecmodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
)

type LoggerCallback struct {
	callbacks.HandlerBuilder // 可以用 callbacks.HandlerBuilder 来辅助实现 callback

	ID  string
	SSE *sse.Writer
	Out chan string
	// Agent 缓存当前agent名称，避免在流式回调中上下文缺失导致识别不到
	Agent string
	mu    sync.RWMutex
}

// 线程安全地设置/获取 Agent
func (cb *LoggerCallback) setAgent(a string) {
	cb.mu.Lock()
	cb.Agent = a
	cb.mu.Unlock()
}

func (cb *LoggerCallback) getAgent() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.Agent
}

func (cb *LoggerCallback) pushF(ctx context.Context, event string, data *model.ChatResp) error {
	dataByte, err := json.Marshal(data)
	if err != nil {
		logger.Errorf(ctx, "json_marshal_error, data:%v, err:%v", data, err)
		return err
	}
	if cb.SSE != nil {
		if err = cb.SSE.WriteEvent("", event, dataByte); err != nil {
			logger.Errorf(ctx, "sse_write_error, event:%s, err:%v", event, err)
		}
	}
	if cb.Out != nil {
		// 避免阻塞：若无人消费则丢弃并告警
		select {
		case cb.Out <- data.Content:
		default:
			logger.Warnf(ctx, "logger_out_channel_blocked, drop, event:%s", event)
		}
	}
	return nil
}

func (cb *LoggerCallback) pushMsg(ctx context.Context, msgID string, msg *schema.Message) error {
	if msg == nil {
		return nil
	}

	// 优先从状态机读取；读取失败则使用缓存的 cb.Agent（加锁读）
	agentName := cb.getAgent()
	_ = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		agentName = state.Goto
		// 同步更新缓存，便于后续帧使用
		cb.setAgent(agentName)
		return nil
	})

	fr := ""
	if msg.ResponseMeta != nil {
		fr = msg.ResponseMeta.FinishReason
	}
	data := &model.ChatResp{
		ThreadID:      cb.ID,
		Agent:         agentName,
		ID:            msgID,
		Role:          "assistant",
		Content:       msg.Content,
		FinishReason:  fr,
		MessageChunks: msg.Content,
	}

	if msg.Role == schema.Tool {
		data.ToolCallID = msg.ToolCallID
		return cb.pushF(ctx, "tool_call_result", data)
	}

	if len(msg.ToolCalls) > 0 {
		event := "tool_call_chunks"
		if len(msg.ToolCalls) != 1 {
			logger.Warnf(ctx, "sse_tool_calls, raw:%v", msg)
			return nil
		}

		ts := []model.ToolResp{}
		tcs := []model.ToolChunkResp{}
		fn := msg.ToolCalls[0].Function.Name
		if len(fn) > 0 {
			event = "tool_calls"
			if strings.HasSuffix(fn, "search") {
				fn = "web_search"
			}
			ts = append(ts, model.ToolResp{
				Name: fn,
				Args: map[string]interface{}{},
				Type: "tool_call",
				ID:   msg.ToolCalls[0].ID,
			})
		}
		tcs = append(tcs, model.ToolChunkResp{
			Name: fn,
			Args: msg.ToolCalls[0].Function.Arguments,
			Type: "tool_call_chunk",
			ID:   msg.ToolCalls[0].ID,
		})
		data.ToolCalls = ts
		data.ToolCallChunks = tcs
		return cb.pushF(ctx, event, data)
	}
	return cb.pushF(ctx, "message_chunk", data)
}

func (cb *LoggerCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	if inputStr, ok := input.(string); ok {
		if cb.Out != nil {
			cb.Out <- "\n==================\n"
			cb.Out <- fmt.Sprintf(" [OnStart] %s ", inputStr)
			cb.Out <- "\n==================\n"
		}
	}
	// 在开始时尽力识别当前agent并缓存，供后续pushMsg使用
	_ = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		cb.setAgent(state.Goto)
		return nil
	})
	if cb.getAgent() == "" && info != nil {
		// 回退使用回调信息里的节点名称
		cb.setAgent(info.Name)
	}
	return ctx
}

func (cb *LoggerCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	//fmt.Println("=========[OnEnd]=========", info.Name, "|", info.Component, "|", info.Type)
	//outputStr, _ := json.MarshalIndent(output, "", "  ") // nolint: byted_s_returned_err_check
	//	outputStr = outputStr[:200]
	//}
	//fmt.Println(string(outputStr))
	return ctx
}

func (cb *LoggerCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	fmt.Println("=========[OnError]=========")
	fmt.Println(err)
	return ctx
}

func (cb *LoggerCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	msgID := utils.RandStr(20)
	go func() {
		defer output.Close() // remember to close the stream in defer
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf(ctx, "[OnEndStream]panic_recover, msgID:%s, err:%v", msgID, err)
			}
		}()
		for {
			frame, err := output.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				logger.Errorf(ctx, "[OnEndStream] recv_error:%v", err)
				return
			}

			switch v := frame.(type) {
			case *schema.Message:
				_ = cb.pushMsg(ctx, msgID, v)
			case *ecmodel.CallbackOutput:
				_ = cb.pushMsg(ctx, msgID, v.Message)
			case []*schema.Message:
				for _, m := range v {
					_ = cb.pushMsg(ctx, msgID, m)
				}
			//case string:
			//	ilog.EventInfo(ctx, "frame_type", "type", "str", "v", v)
			default:
				//ilog.EventInfo(ctx, "frame_type", "type", "unknown", "v", v)
			}
		}

	}()
	return ctx
}

func (cb *LoggerCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	defer input.Close()
	return ctx
}
