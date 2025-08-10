/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package deep

import (
	"context"
	"tone/agent/internal/pkg/model"

	"github.com/RanFeng/ilog"
	"github.com/cloudwego/eino/compose"
)

//type I = string
//type O = string

// 子图流转函数，由上一个子图决定接下来流转到哪个agent
// 并将其name写入 state.Goto ，该函数读取 state.Goto 并将控制权交给对应agent
func agentHandOff(ctx context.Context, input string) (next string, err error) {
	defer func() {
		ilog.EventInfo(ctx, "agent_hand_off", "input", input, "next", next)
	}()
	_ = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		next = state.Goto
		return nil
	})
	return next, nil
}

// Builder 初始化全部子图并连接
func Builder[I, O, S any](ctx context.Context, genFunc compose.GenLocalState[S]) compose.Runnable[I, O] {
	//tools := map[string]Tool{}
	//for _, cli := range llms.MCPServer {
	//	ts, err := mcp.GetTools(ctx, &mcp.Config{Cli: cli})
	//	if err != nil {
	//		ilog.EventError(ctx, err, "builder_error")
	//	}
	//	for _, t := range ts {
	//		v := Tool{}
	//		v.Schema, err = t.Info(ctx)
	//		v.CallAble = t.(tool.InvokableTool)
	//		tools[v.Schema.Name] = v
	//	}
	//}
	//ilog.EventInfo(ctx, "builder", "tools", tools)

	g := compose.NewGraph[I, O](
		compose.WithGenLocalState(genFunc),
	)

	outMap := map[string]bool{
		Coordinator:            true,
		Planner:                true,
		Reporter:               true,
		ResearchTeam:           true,
		Researcher:             true,
		Coder:                  true,
		BackgroundInvestigator: true,
		Human:                  true,
		compose.END:            true,
	}

	coordinatorGraph := NewCAgent[I, O](ctx)
	plannerGraph := NewPlanner[I, O](ctx)
	reporterGraph := NewReporter[I, O](ctx)
	researchTeamGraph := NewResearchTeamNode[I, O](ctx)
	researcherGraph := NewResearcher[I, O](ctx)
	bIGraph := NewBAgent[I, O](ctx)
	coder := NewCoder[I, O](ctx)
	human := NewHumanNode[I, O](ctx)

	_ = g.AddGraphNode(Coordinator, coordinatorGraph, compose.WithNodeName(Coordinator))
	_ = g.AddGraphNode(Planner, plannerGraph, compose.WithNodeName(Planner))
	_ = g.AddGraphNode(Reporter, reporterGraph, compose.WithNodeName(Reporter))
	_ = g.AddGraphNode(ResearchTeam, researchTeamGraph, compose.WithNodeName(ResearchTeam))
	_ = g.AddGraphNode(Researcher, researcherGraph, compose.WithNodeName(Researcher))
	_ = g.AddGraphNode(Coder, coder, compose.WithNodeName(Coder))
	_ = g.AddGraphNode(BackgroundInvestigator, bIGraph, compose.WithNodeName(BackgroundInvestigator))
	_ = g.AddGraphNode(Human, human, compose.WithNodeName(Human))

	_ = g.AddBranch(Coordinator, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(Planner, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(Reporter, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(ResearchTeam, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(Researcher, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(Coder, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(BackgroundInvestigator, compose.NewGraphBranch(agentHandOff, outMap))
	_ = g.AddBranch(Human, compose.NewGraphBranch(agentHandOff, outMap))

	_ = g.AddEdge(compose.START, Coordinator)

	r, err := g.Compile(ctx,
		compose.WithGraphName("EinoDeer"),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
		compose.WithCheckPointStore(model.NewDeerCheckPoint(ctx)), // 指定Graph CheckPointStore
	)
	if err != nil {
		ilog.EventError(ctx, err, "compile failed")
	}
	return r
}
