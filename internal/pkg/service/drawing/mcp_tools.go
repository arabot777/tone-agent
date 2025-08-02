package drawing

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
)

// GetDrawingTools 获取画图工具
// 暂时返回空工具列表，专注于完成工作流架构
func GetDrawingTools(ctx context.Context) ([]tool.BaseTool, error) {
	// TODO: 实现真正的 Wavespeed MCP 工具集成
	// 当前返回空列表以完成编译
	return []tool.BaseTool{}, nil
}
