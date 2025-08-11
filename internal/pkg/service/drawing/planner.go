package drawing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"tone/agent/internal/pkg/enum"
	"tone/agent/internal/pkg/infra"
	"tone/agent/internal/pkg/model"
	"tone/agent/pkg/common/logger"

	"github.com/RanFeng/ilog"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func loadPlannerMsg(ctx context.Context, name string, opts ...any) (output []*schema.Message, err error) {
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		sysPrompt, err := infra.GetDirPromptTemplate(ctx, "draw", name)
		if err != nil {
			ilog.EventInfo(ctx, "get prompt template fail")
			return err
		}

		var promptTemp *prompt.DefaultChatTemplate
		if state.EnableBackgroundInvestigation && len(state.BackgroundInvestigationResults) > 0 {
			promptTemp = prompt.FromMessages(schema.Jinja2,
				schema.SystemMessage(sysPrompt),
				schema.MessagesPlaceholder("user_input", true),
				schema.UserMessage(fmt.Sprintf("background investigation results of user query: \n %s", state.BackgroundInvestigationResults)),
			)
		} else {
			promptTemp = prompt.FromMessages(schema.Jinja2,
				schema.SystemMessage(sysPrompt),
				schema.MessagesPlaceholder("user_input", true),
			)
		}

		variables := map[string]any{
			"locale":              state.Locale,
			"max_step_num":        state.MaxStepNum,
			"max_plan_iterations": state.MaxPlanIterations,
			"CURRENT_TIME":        time.Now().Format("2006-01-02 15:04:05"),
			"user_input":          state.Messages,
		}
		output, err = promptTemp.Format(context.Background(), variables)
		return err
	})
	return output, err
}

func routerPlanner(ctx context.Context, input *schema.Message, opts ...any) (output string, err error) {
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		defer func() {
			output = state.Goto
		}()
		state.Goto = compose.END
		state.CurrentPlan = &model.Plan{}
		// TODO fix 一些 ```
		err = json.Unmarshal([]byte(input.Content), state.CurrentPlan)
		if err != nil {
			logger.Errorf(ctx, "gen_plan_fail, input.Content: %v, err: %v", input.Content, err)
			if state.PlanIterations > 0 {
				state.Goto = enum.Reporter
				return nil
			}
			return nil
		}
		logger.Infof(ctx, "gen_plan_ok, current_plan: %+#v", state.CurrentPlan)
		state.PlanIterations++
		if state.CurrentPlan.HasEnoughContext {
			state.Goto = enum.Reporter
			return nil
		}

		// state.Goto = enum.Human // TODO 改成 human_feedback
		state.Goto = enum.DrawerTeam
		return nil
	})
	return output, nil
}

func NewPlanner[I, O any](ctx context.Context) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()

	_ = cag.AddLambdaNode("load", compose.InvokableLambdaWithOption(loadPlannerMsg))
	_ = cag.AddChatModelNode("agent", infra.PlanModel)
	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(routerPlanner))

	_ = cag.AddEdge(compose.START, "load")
	_ = cag.AddEdge("load", "agent")
	_ = cag.AddEdge("agent", "router")
	_ = cag.AddEdge("router", compose.END)
	return cag
}
