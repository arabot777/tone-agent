package drawing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"tone/agent/internal/pkg/enum"
	"tone/agent/internal/pkg/infra"
	"tone/agent/internal/pkg/model"
	"tone/agent/pkg/common/logger"

	"github.com/RanFeng/ilog"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

func loadStorytellerMsg(ctx context.Context, name string, opts ...any) (output []*schema.Message, err error) {
	logger.Infof(ctx, "loadStorytellerMsg, name: %s", name)
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		sysPrompt, err := infra.GetDirPromptTemplate(ctx, "draw", name)
		if err != nil {
			logger.Errorf(ctx, "get prompt template error: %v", err)
			return err
		}

		promptTemp := prompt.FromMessages(schema.Jinja2,
			schema.SystemMessage(sysPrompt),
			schema.MessagesPlaceholder("user_input", true),
		)

		var curStep *model.Step
		for i := range state.CurrentPlan.Steps {
			step := state.CurrentPlan.Steps[i]
			if len(step.StorytellerScene) == 0 {
				curStep = &step
				break
			}
		}

		if curStep == nil {
			panic("no step found")
		}

		// 获取历史的故事
		var storyContext strings.Builder
		storyContext.WriteString("# Story Context\n\n")
		for _, step := range state.CurrentPlan.Steps {
			if step.StepType == enum.Storyteller && step.StorytellerScene != nil {
				storyContext.WriteString(fmt.Sprintf("## %s\n\n", step.Title))
				for _, scene := range step.StorytellerScene {
					storyContext.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", scene.Title, scene.StoryDetails))
				}
			}
		}

		msg := []*schema.Message{}
		msg = append(msg,
			schema.UserMessage(fmt.Sprintf("#Task\n\n##title\n\n %v \n\n##description\n\n %v \n\n##locale\n\n %v", curStep.Title, curStep.Description, state.Locale)),
		)
		variables := map[string]any{
			"locale":        state.Locale,
			"CURRENT_TIME":  time.Now().Format("2006-01-02 15:04:05"),
			"user_input":    msg,
			"story_context": storyContext.String(),
		}
		output, err = promptTemp.Format(context.Background(), variables)
		return err
	})
	return output, err
}

func routerStoryteller(ctx context.Context, input *schema.Message, opts ...any) (output string, err error) {
	logger.Infof(ctx, "routerStoryteller, input: %+#v", input)
	last := input
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		defer func() {
			output = state.Goto
		}()
		// Always store raw storyteller output to the first pending step
		for i, step := range state.CurrentPlan.Steps {
			if len(step.StorytellerScene) == 0 {
				str := strings.Clone(last.Content)
				state.CurrentPlan.Steps[i].ExecutionRes = &str
				var payload []model.StorytellerScene
				if err := json.Unmarshal([]byte(last.Content), &payload); err != nil {
					// Not a JSON payload; fall back to original flow
					logger.Warnf(ctx, "storyteller JSON parse failed, fallback to DrawerTeam, err: %v", err)
					return nil
				}
				state.CurrentPlan.Steps[i].StorytellerScene = payload

				break
			}
		}

		// Default goto
		state.Goto = enum.DrawerTeam

		logger.Infof(ctx, "routerStoryteller, plan: %v", state.CurrentPlan)
		return nil
	})
	return output, nil
}

func modifyStorytellerfunc(ctx context.Context, input []*schema.Message) []*schema.Message {
	sum := 0
	maxLimit := 50000
	for i := range input {
		if input[i] == nil {
			ilog.EventWarn(ctx, "modify_inputfunc_nil", "input", input[i])
			continue
		}
		l := len(input[i].Content)
		if l > maxLimit {
			ilog.EventWarn(ctx, "modify_inputfunc_clip", "raw_len", l)
			input[i].Content = input[i].Content[l-maxLimit:]
		}
		sum += len(input[i].Content)
	}
	// ilog.EventInfo(ctx, "modify_inputfunc", "sum", sum, "input_len", input)
	return input
}

func NewStorytellerNode[I, O any](ctx context.Context) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		MaxStep:               1000,
		ToolCallingModel:      infra.ChatModel,
		MessageModifier:       modifyStorytellerfunc,
		StreamToolCallChecker: toolCallChecker,
	})
	if err != nil {
		logger.Errorf(ctx, "storyteller agent error: %v", err)
		panic(err)
	}

	agentLambda, err := compose.AnyLambda(agent.Generate, agent.Stream, nil, nil)
	if err != nil {
		logger.Errorf(ctx, "storyteller agent error: %v", err)
		panic(err)
	}

	_ = cag.AddLambdaNode("load", compose.InvokableLambdaWithOption(loadStorytellerMsg))
	_ = cag.AddLambdaNode("agent", agentLambda)
	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(routerStoryteller))

	_ = cag.AddEdge(compose.START, "load")
	_ = cag.AddEdge("load", "agent")
	_ = cag.AddEdge("agent", "router")
	_ = cag.AddEdge("router", compose.END)
	return cag
}
