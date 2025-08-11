package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"tone/agent/internal/pkg/service/einoagent"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type AgentService struct {
}

func NewAgentService() *AgentService {
	return &AgentService{}
}

func (s *AgentService) Ok() string {
	return "agent"
}

func (s *AgentService) Einoagent(ctx context.Context, id string, msg string) (*schema.StreamReader[*schema.Message], error) {

	runner, err := einoagent.BuildeinoagentAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build agent graph: %w", err)
	}

	// conversation := memory.GetConversation(id, true)

	userMessage := &einoagent.UserMessage{
		ID:    id,
		Query: msg,
		// History: conversation.GetMessages(),
	}

	sr, err := runner.Stream(ctx, userMessage, compose.WithCallbacks())
	if err != nil {
		return nil, fmt.Errorf("failed to stream: %w", err)
	}

	srs := sr.Copy(2)

	go func() {
		// for save to memory
		fullMsgs := make([]*schema.Message, 0)

		defer func() {
			// close stream if you used it
			srs[1].Close()

			// add user input to history
			// conversation.Append(schema.UserMessage(msg))

			// fullMsg, err := schema.ConcatMessages(fullMsgs)
			// if err != nil {
			// 	fmt.Println("error concatenating messages: ", err.Error())
			// }
			// add agent response to history
			// conversation.Append(fullMsg)
		}()

	outer:
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context done", ctx.Err())
				return
			default:
				chunk, err := srs[1].Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break outer
					}
				}

				fullMsgs = append(fullMsgs, chunk)
			}
		}
	}()

	return srs[0], nil
}

func (s *AgentService) Journal() string {
	// TODO SSE 返回
	return "journal stream"
}

type LogCallbackConfig struct {
	Detail bool
	Debug  bool
	Writer io.Writer
}

func LogCallback(config *LogCallbackConfig) callbacks.Handler {
	if config == nil {
		config = &LogCallbackConfig{
			Detail: true,
			Writer: os.Stdout,
		}
	}
	if config.Writer == nil {
		config.Writer = os.Stdout
	}
	builder := callbacks.NewHandlerBuilder()
	builder.OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
		fmt.Fprintf(config.Writer, "[view]: start [%s:%s:%s]\n", info.Component, info.Type, info.Name)
		if config.Detail {
			var b []byte
			if config.Debug {
				b, _ = json.MarshalIndent(input, "", "  ")
			} else {
				b, _ = json.Marshal(input)
			}
			fmt.Fprintf(config.Writer, "%s\n", string(b))
		}
		return ctx
	})
	builder.OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		fmt.Fprintf(config.Writer, "[view]: end [%s:%s:%s]\n", info.Component, info.Type, info.Name)
		return ctx
	})
	return builder.Build()
}

// func (s *AgentService) Drawing(ctx context.Context, id string, msg string) (*schema.StreamReader[*schema.Message], error) {
// 	// conversation := memory.GetConversation(id, true)

// 	userMessage := &drawing.UserMessage{
// 		ID:    id,
// 		Query: msg,
// 		// History: conversation.GetMessages(),
// 	}

// 	runner, err := drawing.BuildDrawingAgent(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to build agent graph: %w", err)
// 	}

// 	sr, err := runner.Stream(ctx, userMessage, compose.WithCallbacks())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to stream: %w", err)
// 	}

// 	srs := sr.Copy(2)

// 	go func() {
// 		// for save to memory
// 		fullMsgs := make([]*schema.Message, 0)

// 		defer func() {
// 			srs[1].Close()
// 		}()

// 	outer:
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				fmt.Println("context done", ctx.Err())
// 				return
// 			default:
// 				chunk, err := srs[1].Recv()
// 				if err != nil {
// 					if errors.Is(err, io.EOF) {
// 						break outer
// 					}
// 				}

// 				fullMsgs = append(fullMsgs, chunk)
// 			}
// 		}
// 	}()

// 	return srs[0], nil
// }
