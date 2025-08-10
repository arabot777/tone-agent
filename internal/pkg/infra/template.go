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

package infra

import (
	"context"
	"embed"
	"fmt"
)

//go:embed prompts/*.md
var promptsFS embed.FS

// GetPromptTemplate 加载并返回一个提示模板
func GetPromptTemplate(ctx context.Context, promptName string) (string, error) {
	// 使用 embed.FS 读取嵌入的模板文件
	filePath := fmt.Sprintf("prompts/%s.md", promptName)
	content, err := promptsFS.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取模板文件 %s 失败: %w", promptName, err)
	}

	return string(content), nil
}
