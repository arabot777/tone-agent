package journal

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/multiagent/host"
	"github.com/cloudwego/eino/schema"
)

func getJournalFilePath(dateStr string) (string, error) {
	// generate the unique file path for today's journal file
	filePath := fmt.Sprintf("journal_%s.txt", dateStr)

	// find the file path for today's journal file
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// if not found, create the journal file with the file path
		file, err := os.Create(filePath)
		if err != nil {
			return "", err
		}
		file.Close()
	}

	// return the file path
	return filePath, nil
}

func NewWriteJournalSpecialist(ctx context.Context, baseURL, model, apiKey string) (*host.Specialist, error) {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   model,
		APIKey:  apiKey,
	})
	if err != nil {
		return nil, err
	}

	// use a chat model to rewrite user query to journal entry
	// for example, the user query might be:
	//
	// write: I got up at 7:00 in the morning.
	//
	// should be rewritten to:
	//
	// I got up at 7:00 in the morning.
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
		systemMsg := &schema.Message{
			Role:    schema.System,
			Content: "You are responsible for preparing the user query for insertion into journal. The user's query is expected to contain the actual text the user want to write to journal, as well as convey the intention that this query should be written to journal. You job is to remove that intention from the user query, while preserving as much as possible the user's original query, and output ONLY the text to be written into journal",
		}
		return append([]*schema.Message{systemMsg}, input...), nil
	})).
		AppendChatModel(chatModel).
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			err := appendJournal(input.Content)
			if err != nil {
				return nil, err
			}
			return &schema.Message{
				Role:    schema.Assistant,
				Content: "Journal written successfully: " + input.Content,
			}, nil
		}))

	r, err := chain.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &host.Specialist{
		AgentMeta: host.AgentMeta{
			Name:        "write_journal",
			IntendedUse: "treat the user query as a sentence of a journal entry, append it to the right journal file",
		},
		Invokable: func(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.Message, error) {
			return r.Invoke(ctx, input, agent.GetComposeOptions(opts...)...)
		},
	}, nil
}
