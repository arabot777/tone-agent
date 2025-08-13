package drawing

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
	"tone/agent/internal/pkg/enum"
	"tone/agent/internal/pkg/infra"
	"tone/agent/internal/pkg/model"
	"tone/agent/pkg/common/logger"

	"github.com/RanFeng/ilog"
	"github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

func loadDrawerMsg(ctx context.Context, name string, opts ...any) (output []*schema.Message, err error) {
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		sysPrompt, err := infra.GetDirPromptTemplate(ctx, "draw", name)
		if err != nil {
			ilog.EventInfo(ctx, "get prompt template fail")
			return err
		}

		promptTemp := prompt.FromMessages(schema.Jinja2,
			schema.SystemMessage(sysPrompt),
			schema.MessagesPlaceholder("user_input", true),
		)

		// Find the next storyteller step with a pending scene (scene.DrawerOutput == "")
		var curStep *model.Step
		var sceneIdx int = -1
		for i := range state.CurrentPlan.Steps {
			st := &state.CurrentPlan.Steps[i]
			if st.StepType != enum.Storyteller || len(st.StorytellerScene) == 0 {
				continue
			}
			for j := range st.StorytellerScene {
				if st.StorytellerScene[j].DrawerOutput == "" {
					curStep = st
					sceneIdx = j
					break
				}
			}
			if curStep != nil {
				break
			}
		}

		if curStep == nil || sceneIdx < 0 {
			panic("no pending scene found for drawing")
		}

		// Collect finished storyteller results as Story Context
		var storyCtx strings.Builder
		storyCtx.WriteString("# Story Context\n\n")
		for _, step := range state.CurrentPlan.Steps {
			if step.StepType == enum.Storyteller && step.ExecutionRes != nil && *step.ExecutionRes != "" {
				storyCtx.WriteString(fmt.Sprintf("## %s\n\n%s\n\n", step.Title, *step.ExecutionRes))
			}
		}

		msg := []*schema.Message{}
		if storyCtx.Len() > len("# Story Context\n\n") {
			msg = append(msg, schema.UserMessage(storyCtx.String()))
		}

		// Build Current Drawing Task from structured storyteller scenes
		sc := curStep.StorytellerScene[sceneIdx]
		sceneLocale := state.Locale
		globalTitle := curStep.Title

		var sb strings.Builder
		sb.WriteString("# Current Drawing Task\n\n")
		if globalTitle != "" {
			sb.WriteString("## Story Title\n\n" + globalTitle + "\n\n")
		}
		sb.WriteString("## Scene Title\n\n" + sc.Title + "\n\n")
		if sc.StoryDetails != "" {
			sb.WriteString("## StoryDetails\n\n" + sc.StoryDetails + "\n\n")
		}
		if sc.SceneIndex > 0 {
			sb.WriteString(fmt.Sprintf("## SceneIndex\n\n%v\n\n", sc.SceneIndex))
		}
		sb.WriteString("## Locale\n\n" + sceneLocale)

		msg = append(msg,
			schema.UserMessage(sb.String()),
			schema.SystemMessage("IMPORTANT: Generate exactly ONE image for the Current Drawing Task. Use the Story Context strictly as reference. Illustrate ONLY the current scene specified. You MUST call tools to generate the image. Output ONLY the image URL as a single plain-text line (no extra text, no JSON, no markdown)."),
		)
		variables := map[string]any{
			"locale": state.Locale,
			// "max_step_num":        state.MaxStepNum,
			// "max_plan_iterations": state.MaxPlanIterations,
			"CURRENT_TIME": time.Now().Format("2006-01-02 15:04:05"),
			"user_input":   msg,
		}
		output, err = promptTemp.Format(context.Background(), variables)
		return err
	})
	return output, err
}

func modifyInputfunc(ctx context.Context, input []*schema.Message) []*schema.Message {
	sum := 0
	maxLimit := 50000
	for i := range input {
		if input[i] == nil {
			logger.Warnf(ctx, "modify_inputfunc_nil input=%v", input[i])
			continue
		}
		l := len(input[i].Content)
		if l > maxLimit {
			logger.Warnf(ctx, "modify_inputfunc_clip raw_len=%d", l)
			input[i].Content = input[i].Content[l-maxLimit:]
		}
		sum += len(input[i].Content)
	}
	logger.Infof(ctx, "modify_inputfunc sum=%d input_len=%d", sum, len(input))
	return input
}

func routerDrawer(ctx context.Context, input *schema.Message, opts ...any) (output string, err error) {
	//ilog.EventInfo(ctx, "routerResearcher", "input", input)
	last := input
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		defer func() {
			output = state.Goto
		}()
		// Mark the current pending scene as completed by storing the drawer result
		for i := range state.CurrentPlan.Steps {
			st := &state.CurrentPlan.Steps[i]
			if st.StepType != enum.Storyteller || len(st.StorytellerScene) == 0 {
				continue
			}
			for j := range st.StorytellerScene {
				if st.StorytellerScene[j].DrawerOutput == "" {
					st.StorytellerScene[j].DrawerOutput = strings.Clone(last.Content)
					goto DrawerTeam
				}
			}
		}
	DrawerTeam:
		// Hand off routing to drawer_team so team can coordinate next steps (continue drawing or escalate to reporter)
		state.Goto = enum.DrawerTeam

		logger.Infof(ctx, "routerDrawer, handoff_to=%s, plan: %v", state.Goto, state.CurrentPlan)
		return nil
	})
	return output, nil
}

func toolCallChecker(_ context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
	defer sr.Close()

	for {
		msg, err := sr.Recv()
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		if len(msg.ToolCalls) > 0 {
			return true, nil
		}
	}
}

func NewDrawerNode[I, O any](ctx context.Context) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()

	drawerTools := []tool.BaseTool{}
	for _, cli := range infra.MCPServer {
		ts, err := mcp.GetTools(ctx, &mcp.Config{Cli: cli})
		if err != nil {
			logger.Errorf(ctx, "builder_error, err: %v", err)
		}
		drawerTools = append(drawerTools, ts...)
	}
	logger.Infof(ctx, "drawer_init, tools_loaded: %v", len(drawerTools))

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		MaxStep:               1000,
		ToolCallingModel:      infra.ChatModel,
		ToolsConfig:           compose.ToolsNodeConfig{Tools: drawerTools},
		MessageModifier:       modifyInputfunc,
		StreamToolCallChecker: toolCallChecker,
	})
	if err != nil {
		logger.Errorf(ctx, "drawer builder_error, err: %v", err)
	}

	agentLambda, err := compose.AnyLambda(agent.Generate, agent.Stream, nil, nil)
	if err != nil {
		logger.Errorf(ctx, "drawer builder_error, err: %v", err)
		panic(err)
	}

	_ = cag.AddLambdaNode("load", compose.InvokableLambdaWithOption(loadDrawerMsg))
	_ = cag.AddLambdaNode("agent", agentLambda)
	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(routerDrawer))

	_ = cag.AddEdge(compose.START, "load")
	_ = cag.AddEdge("load", "agent")
	_ = cag.AddEdge("agent", "router")
	_ = cag.AddEdge("router", compose.END)
	return cag
}
