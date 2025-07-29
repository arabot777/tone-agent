# 名称字典替换功能实现

## 概述

本功能实现了在流式会话处理过程中，自动将英文名称替换为中文名称的功能。通过硬编码在代码中的名称映射关系，系统会在处理 `rawMsg` 时自动进行字符串替换。

## 实现位置

主要实现在 `internal/api/service/agent_session_stream.go` 文件中。

## 核心功能

### 1. 全局名称字典

```go
var (
    // 全局名称字典，用于字符串替换
    nameDict = map[string]string{
        "Requirement_Analyst":   "需求分析师",
        "Data_Validator":        "数据验证者",
        "Data_Arbiter":          "数据层仲裁者",
        "Ambiguity_Negotiator":  "歧义协商者",
        "Quality_Guardian":      "质量守门人",
        "Four_Causes_Tracer":    "四因追溯者",
        "Integration_Architect": "整合架构师",
        "Dialectic_Engine":      "辩证矛盾引擎",
        "Cognitive_Sentinel":    "认知免疫哨兵",
        "Ethical_Arbiter":       "伦理层仲裁者",
        "Report_Generator":      "报告生成师",
        "Fact_Checker":          "事实核查官",
        "fetch_webpage":         "网页获取",
        "google_search":         "谷歌搜索",
    }
)
```

### 2. 字符串替换函数

```go
func replaceWithNameDict(content string) string {
    result := content
    for key, value := range nameDict {
        result = strings.ReplaceAll(result, key, value)
    }
    return result
}
```

## 应用场景

### 1. 字节数组类型的 rawMsg

```go
if rawMsg, ok := resp.([]byte); ok {
    // 对rawMsg进行字符串替换
    replacedMsg := replaceWithNameDict(string(rawMsg))
    // 添加换行符，确保每条消息独立一行
    data := append([]byte(replacedMsg), '\n')
    a.buffer = data
}
```

### 2. Go 结构体类型的响应

```go
} else {
    // 将收到的 Go 结构体重新序列化为 JSON 字符串
    data, err := json.Marshal(resp)
    if err != nil {
        // 如果序列化失败，退回到字符串表示
        data = []byte(fmt.Sprintf("%v", resp))
    }
    // 对序列化后的数据进行字符串替换
    replacedData := replaceWithNameDict(string(data))
    // 添加换行符，确保每条消息独立一行
    data = append([]byte(replacedData), '\n')
    a.buffer = data
}
```

### 3. agent_session.go 中的 source 字段

```go
if config, ok := msg["config"].(map[string]interface{}); ok {
    source, _ := config["source"].(string)
    msgType, _ := config["type"].(string)
    id, _ := msg["id"].(float64)

    // 对source应用名称字典替换
    replacedSource := replaceWithNameDict(source)

    m := bo.Message{
        ID: fmt.Sprintf("%.0f", id),
        Role: func() string {
            if source == "user" {
                return "user"
            }
            return "assistant"
        }(),
        Content: config["content"],
        Type:    msgType,
        Source:  replacedSource,  // 使用替换后的source
    }
}
```

## 名称字典配置

名称映射关系直接硬编码在代码中，包含以下映射：

```go
{
  "Requirement_Analyst": "需求分析师",
  "Data_Validator": "数据验证者",
  "Data_Arbiter": "数据层仲裁者",
  "Ambiguity_Negotiator": "歧义协商者",
  "Quality_Guardian": "质量守门人",
  "Four_Causes_Tracer": "四因追溯者",
  "Integration_Architect": "整合架构师",
  "Dialectic_Engine": "辩证矛盾引擎",
  "Cognitive_Sentinel": "认知免疫哨兵",
  "Ethical_Arbiter": "伦理层仲裁者",
  "Report_Generator": "报告生成师",
  "Fact_Checker": "事实核查官",
  "fetch_webpage": "网页获取",
  "google_search": "谷歌搜索"
}
```

## 测试验证

创建了完整的测试用例来验证功能：

- `TestNameDictReplacement`: 基础字符串替换测试
- `TestStreamJSONReplacement`: 流式JSON数据替换测试
- `TestComplexJSONReplacement`: 复杂JSON结构替换测试

运行测试：
```bash
go test ./test -v
```

## 功能特点

1. **硬编码配置**: 名称字典直接硬编码在代码中，无需外部文件
2. **高性能**: 避免了文件读取和JSON解析的开销
3. **简单可靠**: 不依赖外部文件，减少了出错的可能性
4. **全局应用**: 适用于所有类型的流式响应（字节数组和结构体）
5. **实时替换**: 在流式处理过程中实时进行字符串替换

## 使用示例

### 输入示例
```json
{"type": "start", "agent": "Requirement_Analyst", "status": "running"}
```

### 输出示例
```json
{"type": "start", "agent": "需求分析师", "status": "running"}
```

## 扩展说明

如需添加新的名称映射，只需在 `internal/api/service/agent_session_stream.go` 文件中的 `nameDict` 变量中添加新的键值对即可，系统会自动识别并应用新的映射关系。 