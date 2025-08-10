package infra

import (
	"context"
	"os"
	"tone/agent/internal/pkg/model"

	"github.com/cloudwego/eino-ext/components/model/openai"
	openai3 "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/getkin/kin-openapi/openapi3gen"
)

var (
	ChatModel *openai.ChatModel
	PlanModel *openai.ChatModel
)

func InitModel() {
	config := &openai.ChatModelConfig{
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		Model:   os.Getenv("OPENAI_MODEL_NAME"),
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		ExtraFields: map[string]interface{}{
			"thinking": map[string]string{
				"type": "disabled",
			},
		},
	}
	ChatModel, _ = openai.NewChatModel(context.Background(), config)
	planSchema, _ := openapi3gen.NewSchemaRefForValue(&model.Plan{}, nil)

	planconfig := &openai.ChatModelConfig{
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		Model:   os.Getenv("OPENAI_MODEL_NAME"),
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		ResponseFormat: &openai3.ChatCompletionResponseFormat{
			Type: openai3.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai3.ChatCompletionResponseFormatJSONSchema{
				Name:   "plan",
				Strict: false,
				Schema: planSchema.Value,
			},
		},
		ExtraFields: map[string]interface{}{
			"thinking": map[string]string{
				"type": "disabled",
			},
		},
	}
	PlanModel, _ = openai.NewChatModel(context.Background(), planconfig)
}
