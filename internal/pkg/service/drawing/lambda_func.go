package drawing

import (
	"context"
	"encoding/json"
	"time"
	"tone/agent/pkg/common/logger"

	"github.com/cloudwego/eino/schema"
)

// newUserMessagePreLambda 提取用户消息 - 输出 map[string]interface{} 用于模板
func newUserMessagePreLambda(ctx context.Context, input *UserMessage, opts ...any) (output map[string]any, err error) {
	output = map[string]any{
		"drawing_topic":  input.Query,
		"number_queries": 5, // 假设生成5个查询
		"current_date":   time.Now().Format("2006-01-02"),
	}
	return output, nil
}

// newSearchQueryPreLambda 处理搜索查询预处理 - 输出 map[string]interface{} 用于下一个模板
func newSearchQueryPreLambda(ctx context.Context, input *schema.Message, opts ...any) (output map[string]any, err error) {
	var queryList SearchQueryList
	if err := json.Unmarshal([]byte(input.Content), &queryList); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "处理搜索查询预处理: %v", queryList)
	// 模拟处理查询列表，增加查询上下文

	return map[string]interface{}{
		"query_content":    input.Content,
		"current_stage":    "query_generation",
		"research_context": queryList.Query,
	}, nil
}

// researchProcessor 处理研究结果 - 从 WorkflowState 输出 map[string]interface{}
func researchProcessor(ctx context.Context, input *WorkflowState) (map[string]interface{}, error) {
	// 模拟研究处理，增加研究上下文
	newContext := input.ResearchContext + "\n研究阶段完成，获得相关信息"

	return map[string]interface{}{
		"drawing_topic":       input.DrawingTopic,
		"search_queries":      input.SearchQueries,
		"research_context":    newContext,
		"research_loop_count": input.ResearchLoopCount + 1,
		"current_stage":       "research",
	}, nil
}

// drawingPromptProcessor 处理画图提示生成结果 - 输出 map[string]interface{}
func drawingPromptProcessor(ctx context.Context, input *schema.Message) (map[string]interface{}, error) {
	var drawingPrompt DrawingPrompt
	if err := json.Unmarshal([]byte(input.Content), &drawingPrompt); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"drawing_prompt":   drawingPrompt.Prompt,
		"current_stage":    "drawing_prompt",
		"research_context": "生成画图提示: " + drawingPrompt.Prompt,
	}, nil
}

// drawingCreatorProcessor 处理画图创建结果 - 输出 WorkflowState 用于最终答案
func drawingCreatorProcessor(ctx context.Context, input *schema.Message) (*WorkflowState, error) {
	// 从 Agent 响应中提取图片信息
	state := &WorkflowState{
		CurrentStage:        "drawing_creation",
		DrawingResult:       input.Content,
		DrawingAttemptCount: 1,
		ResearchContext:     "完成图片生成",
	}
	return state, nil
}

// drawingReviewProcessor 处理画图评审结果 - 输出 map[string]interface{}
func drawingReviewProcessor(ctx context.Context, input *schema.Message) (map[string]interface{}, error) {
	var review DrawingReview
	if err := json.Unmarshal([]byte(input.Content), &review); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"review_feedback":  review.Feedback,
		"current_stage":    "drawing_review",
		"research_context": "完成图片评审: " + review.Feedback,
	}, nil
}

// extensionGeneratorProcessor 处理扩展生成结果 - 输出 WorkflowState 用于最终答案
func extensionGeneratorProcessor(ctx context.Context, input *schema.Message) (*WorkflowState, error) {
	var extensions ExtensionIdeas
	if err := json.Unmarshal([]byte(input.Content), &extensions); err != nil {
		return nil, err
	}

	state := &WorkflowState{
		CurrentStage:     "extension_generation",
		ExtensionIdeas:   &extensions,
		ResearchContext:  "生成扩展想法",
		WorkflowComplete: true, // 标记工作流完成
	}
	return state, nil
}

// finalAnswerFormatter 格式化最终答案 - 从 WorkflowState 输出最终消息
func finalAnswerFormatter(ctx context.Context, input *WorkflowState) (*schema.Message, error) {
	// 构建最终回复
	finalResponse := map[string]interface{}{
		"drawing_topic":         input.DrawingTopic,
		"workflow_completed":    true,
		"current_stage":         input.CurrentStage,
		"search_queries":        input.SearchQueries,
		"research_context":      input.ResearchContext,
		"drawing_prompt":        input.DrawingPrompt,
		"drawing_result":        input.DrawingResult,
		"drawing_review":        input.DrawingReview,
		"extension_ideas":       input.ExtensionIdeas,
		"final_output":          input.FinalOutput,
		"research_loop_count":   input.ResearchLoopCount,
		"drawing_attempt_count": input.DrawingAttemptCount,
		"status":                "success",
	}

	responseBytes, err := json.Marshal(finalResponse)
	if err != nil {
		return nil, err
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: string(responseBytes),
	}, nil
}

// 添加状态转换函数，将 map[string]interface{} 转换为 WorkflowState
func mapToWorkflowState(ctx context.Context, input map[string]interface{}) (*WorkflowState, error) {
	state := &WorkflowState{}

	if topic, ok := input["drawing_topic"].(string); ok {
		state.DrawingTopic = topic
	}
	if stage, ok := input["current_stage"].(string); ok {
		state.CurrentStage = stage
	}
	if context, ok := input["research_context"].(string); ok {
		state.ResearchContext = context
	}
	if count, ok := input["research_loop_count"].(int); ok {
		state.ResearchLoopCount = count
	}

	return state, nil
}
