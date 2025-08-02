package drawing

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func BuildDrawingAgent(ctx context.Context) (r compose.Runnable[*UserMessage, *schema.Message], err error) {
	// 定义节点名称常量
	const (
		UserMesssagePre      = "UserMessagePre"
		UserMessagePrompt    = "UserMessagePrompt"
		UserMessageExtractor = "UserMessageExtractor"

		SearchQueryPre       = "SearchQueryPre"
		SearchQueryPrompt    = "SearchQueryPrompt"
		SearchQueryExtractor = "SearchQueryExtractor"
	)

	// 创建图
	g := compose.NewGraph[*UserMessage, *schema.Message](compose.WithGenLocalState(func(ctx context.Context) *WorkflowState {
		return &WorkflowState{}
	}))

	llm, err := newToolCallingChatModel(ctx)
	if err != nil {
		return nil, err
	}

	// 1. 用户消息提取器

	_ = g.AddLambdaNode(UserMesssagePre,
		compose.InvokableLambdaWithOption[*UserMessage](newUserMessagePreLambda),
		compose.WithNodeName("用户消息预处理"))

	_ = g.AddChatTemplateNode(UserMessagePrompt,
		newQueryWriterTemplate(ctx),
		compose.WithNodeName("用户消息处理"))

	err = g.AddChatModelNode(UserMessageExtractor,
		llm,
		compose.WithNodeName("用户消息提取器"))
	if err != nil {
		return nil, err
	}

	// web query
	_ = g.AddLambdaNode(SearchQueryPre,
		compose.InvokableLambdaWithOption[*schema.Message](newSearchQueryPreLambda),
		compose.WithNodeName("搜索查询预处理"))
	_ = g.AddChatTemplateNode(SearchQueryPrompt,
		newWebSearcherTemplate(ctx),
		compose.WithNodeName("搜索查询生成"))
	err = g.AddChatModelNode(SearchQueryExtractor,
		llm,
		compose.WithNodeName("搜索查询提取器"))
	if err != nil {
		return nil, err
	}

	_ = g.AddEdge(compose.START, UserMesssagePre)
	_ = g.AddEdge(UserMesssagePre, UserMessagePrompt)
	_ = g.AddEdge(UserMessagePrompt, UserMessageExtractor)
	_ = g.AddEdge(UserMessageExtractor, SearchQueryPre)
	_ = g.AddEdge(SearchQueryPre, SearchQueryPrompt)
	_ = g.AddEdge(SearchQueryPrompt, SearchQueryExtractor)
	_ = g.AddEdge(SearchQueryExtractor, compose.END)

	r, err = g.Compile(ctx, compose.WithGraphName("DrawingAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
