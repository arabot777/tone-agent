package drawing

import (
	"context"
	"tone/agent/internal/pkg/enum"
	"tone/agent/internal/pkg/model"
	"tone/agent/pkg/common/logger"

	"github.com/cloudwego/eino/compose"
)

func routerDrawerTeam(ctx context.Context, input string, opts ...any) (output string, err error) {
	logger.Infof(ctx, "drawer_team router select next node by input: %s", input)
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		defer func() {
			output = state.Goto
		}()
		state.Goto = enum.Planner
		if state.CurrentPlan == nil {
			return nil
		}
		logger.Infof(ctx, "routerDrawerTeam, plan: %+#v", state.CurrentPlan)
		for i, step := range state.CurrentPlan.Steps {
			logger.Infof(ctx, "drawer_team router select next node by input: %s, step_type: %s, scene count: %d, index: %d",
				input, step.StepType, len(step.StorytellerScene), i)

			if gotoDrawer(step) {
				state.Goto = enum.Drawer
				return nil
			}

			if step.ExecutionRes != nil && *step.ExecutionRes != "" {
				continue
			}
			logger.Infof(ctx, "drawer_team_step, step: %v, index: %d", step, i)
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

func gotoDrawer(step model.Step) bool {
	if step.StepType != enum.Storyteller {
		return false
	}
	if len(step.StorytellerScene) == 0 {
		return false
	}
	for _, scene := range step.StorytellerScene {
		if scene.DrawerOutput == "" {
			return true
		}
	}
	return false
}

func NewDrawTeamNode[I, O any](ctx context.Context) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()
	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(routerDrawerTeam))

	_ = cag.AddEdge(compose.START, "router")
	_ = cag.AddEdge("router", compose.END)

	return cag
}
