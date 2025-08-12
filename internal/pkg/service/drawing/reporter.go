package drawing

import (
	"context"
	"fmt"
	"time"
	"tone/agent/internal/pkg/infra"
	"tone/agent/internal/pkg/model"
	"tone/agent/pkg/common/logger"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func loadReporterMsg(ctx context.Context, name string, opts ...any) (output []*schema.Message, err error) {
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		sysPrompt, err := infra.GetDirPromptTemplate(ctx, "draw", name)
		if err != nil {
			logger.Errorf(ctx, "get prompt template fail: %v", err)
			return err
		}

		promptTemp := prompt.FromMessages(schema.Jinja2,
			schema.SystemMessage(sysPrompt),
			schema.MessagesPlaceholder("user_input", true),
		)

		msg := []*schema.Message{}
		// Build a concise, structured input describing the story and scenes with image URLs
		// Title and intro
		msg = append(msg,
			schema.UserMessage(fmt.Sprintf("# Story Title\n\n%v\n\n# Story Introduction\n\n%v", state.CurrentPlan.Title, state.CurrentPlan.Thought)),
			schema.SystemMessage("IMPORTANT: Optimize and format the story layout according to the reporter prompt. Produce polished Markdown in the specified locale. Each scene must include its image (embed by URL), a short italic caption, and concise narrative paragraphs. Do not invent content or URLs. Output raw Markdown only."),
		)

		// Aggregate scenes from storyteller output with their drawer image URLs
		for _, step := range state.CurrentPlan.Steps {
			if len(step.StorytellerScene) == 0 {
				continue
			}
			for _, sc := range step.StorytellerScene {
				DrawerOutput := sc.DrawerOutput // string
				// Provide per-scene payload succinctly
				msg = append(msg, schema.UserMessage(
					fmt.Sprintf("Scene %d\nTitle: %s\nDetails: %s\nDrawerOutput: %s",
						sc.SceneIndex, sc.Title, sc.StoryDetails, DrawerOutput),
				))
			}
		}
		variables := map[string]any{
			"locale":              state.Locale,
			"max_step_num":        state.MaxStepNum,
			"max_plan_iterations": state.MaxPlanIterations,
			"CURRENT_TIME":        time.Now().Format("2006-01-02 15:04:05"),
			"user_input":          msg,
		}
		output, err = promptTemp.Format(context.Background(), variables)
		return err
	})
	return output, err
}

func routerReporter(ctx context.Context, input *schema.Message, opts ...any) (output string, err error) {
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		defer func() {
			output = state.Goto
		}()
		logger.Infof(ctx, "report_end: %v", input.Content)
		state.Goto = compose.END
		return nil
	})
	return output, nil
}

func NewReporter[I, O any](ctx context.Context) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()

	_ = cag.AddLambdaNode("load", compose.InvokableLambdaWithOption(loadReporterMsg))
	_ = cag.AddChatModelNode("agent", infra.ChatModel)
	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(routerReporter))

	_ = cag.AddEdge(compose.START, "load")
	_ = cag.AddEdge("load", "agent")
	_ = cag.AddEdge("agent", "router")
	_ = cag.AddEdge("router", compose.END)
	return cag
}
