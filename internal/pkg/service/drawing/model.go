package drawing

import (
	"context"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func newToolCallingChatModel(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		Model:   os.Getenv("OPENAI_MODEL_NAME"),
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		ExtraFields: map[string]interface{}{
			"thinking": map[string]string{
				"type": "disabled",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	tools, err := GetTools(ctx)
	if err != nil {
		return nil, err
	}

	// 将工具转换为 ToolInfo 格式
	toolInfos := make([]*schema.ToolInfo, len(tools))
	for i, tool := range tools {
		toolInfo, err := tool.Info(ctx)
		if err != nil {
			return nil, err
		}
		toolInfos[i] = toolInfo
	}

	// 使用 WithTools 方法绑定工具，返回 ToolCallingChatModel
	toolCallingModel, err := chatModel.WithTools(toolInfos)
	if err != nil {
		return nil, err
	}

	return toolCallingModel, nil
}

func GetTools(ctx context.Context) ([]tool.BaseTool, error) {

	toolDDGSearch, err := NewDDGSearch(ctx, nil)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{
		toolDDGSearch,
	}, nil
}

func defaultDDGSearchConfig(ctx context.Context) (*duckduckgo.Config, error) {
	config := &duckduckgo.Config{}
	return config, nil
}

func NewDDGSearch(ctx context.Context, config *duckduckgo.Config) (tn tool.BaseTool, err error) {
	if config == nil {
		config, err = defaultDDGSearchConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	tn, err = duckduckgo.NewTextSearchTool(ctx, config)
	if err != nil {
		return nil, err
	}
	return tn, nil
}
