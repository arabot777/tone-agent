package drawing

import "github.com/cloudwego/eino/schema"

type UserMessage struct {
	ID      string            `json:"id"`
	Query   string            `json:"query"`
	History []*schema.Message `json:"history"`
}

// 搜索查询列表
type SearchQueryList struct {
	Query     []string `json:"query"`     // 搜索查询列表
	Rationale string   `json:"rationale"` // 查询相关性说明
}

// 画图提示
type DrawingPrompt struct {
	Prompt string `json:"prompt"` // 详细的画图提示
}

// 画图评审结果
type DrawingReview struct {
	IsSatisfactory bool   `json:"is_satisfactory"` // 是否满足要求
	Feedback       string `json:"feedback"`        // 详细反馈
}

// 单个扩展创意
type ExtensionIdea struct {
	Title       string `json:"title"`       // 扩展创意标题
	Description string `json:"description"` // 详细描述
}

// 创意扩展集合
type ExtensionIdeas struct {
	Ideas []ExtensionIdea `json:"ideas"` // 创意扩展列表
}

// 深度画图配置
type Configuration struct {
	// 模型配置
	ReasoningModel string `json:"reasoning_model"` // 推理模型
	DrawingModel   string `json:"drawing_model"`   // 画图模型

	// 搜索和研究配置
	MaxResearchLoops         int `json:"max_research_loops"`           // 最大研究循环次数
	NumberOfInitialQueries   int `json:"number_of_initial_queries"`    // 初始查询数量
	MaxSearchResultsPerQuery int `json:"max_search_results_per_query"` // 每个查询的最大搜索结果数

	// 画图配置
	MaxDrawingAttempts int    `json:"max_drawing_attempts"` // 最大画图尝试次数
	DrawingQuality     string `json:"drawing_quality"`      // 画图质量
	DrawingStyle       string `json:"drawing_style"`        // 画图风格

	// 评审配置
	SatisfactionThreshold float64 `json:"satisfaction_threshold"` // 满意度阈值

	// 扩展配置
	MaxExtensions            int  `json:"max_extensions"`             // 最大扩展数量
	EnableCreativeExtensions bool `json:"enable_creative_extensions"` // 是否启用创意扩展
}

// 默认配置
func DefaultConfiguration() *Configuration {
	return &Configuration{
		ReasoningModel:           "doubao-1-5-lite-32k-250115",
		DrawingModel:             "doubao-1-5-lite-32k-250115",
		MaxResearchLoops:         3,
		NumberOfInitialQueries:   3,
		MaxSearchResultsPerQuery: 5,
		MaxDrawingAttempts:       2,
		DrawingQuality:           "high",
		DrawingStyle:             "",
		SatisfactionThreshold:    0.7,
		MaxExtensions:            3,
		EnableCreativeExtensions: true,
	}
}

// 工作流状态
type WorkflowState struct {
	CurrentStage        string           `json:"current_stage"`         // 当前阶段
	DrawingTopic        string           `json:"drawing_topic"`         // 画图主题
	SearchQueries       *SearchQueryList `json:"search_queries"`        // 搜索查询
	ResearchContext     string           `json:"research_context"`      // 研究上下文
	DrawingPrompt       *DrawingPrompt   `json:"drawing_prompt"`        // 画图提示
	DrawingResult       string           `json:"drawing_result"`        // 画图结果
	DrawingReview       *DrawingReview   `json:"drawing_review"`        // 画图评审
	ExtensionIdeas      *ExtensionIdeas  `json:"extension_ideas"`       // 扩展创意
	FinalOutput         string           `json:"final_output"`          // 最终输出
	WorkflowComplete    bool             `json:"workflow_complete"`     // 工作流是否完成
	ResearchLoopCount   int              `json:"research_loop_count"`   // 研究循环计数
	DrawingAttemptCount int              `json:"drawing_attempt_count"` // 画图尝试计数
}
