package drawing

import (
	"context"
	"tone/agent/internal/pkg/enum"
	"tone/agent/internal/pkg/model"
	"tone/agent/pkg/common/logger"

	"github.com/cloudwego/eino/compose"
)

func routerResearchTeam(ctx context.Context, input string, opts ...any) (output string, err error) {
	//ilog.EventInfo(ctx, "routerResearchTeam", "input", input)
	logger.Infof(ctx, "routerResearchTeam, input: %v", input)
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		defer func() {
			output = state.Goto
		}()
		state.Goto = enum.Planner
		if state.CurrentPlan == nil {
			return nil
		}
		for i, step := range state.CurrentPlan.Steps {
			if step.ExecutionRes != nil {
				continue
			}
			logger.Infof(ctx, "research_team_step, step: %v, index: %d", step, i)
			switch step.StepType {
			case enum.Drawer:
				state.Goto = enum.Drawer
				return nil
			case enum.Storyteller:
				state.Goto = enum.Storyteller
				return nil
			}
		}
		if state.PlanIterations >= state.MaxPlanIterations {
			state.Goto = enum.Reporter
			return nil
		}
		return nil
	})
	return output, nil
}

func NewDrawTeamNode[I, O any](ctx context.Context) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()
	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(routerResearchTeam))

	_ = cag.AddEdge(compose.START, "router")
	_ = cag.AddEdge("router", compose.END)

	return cag
}
