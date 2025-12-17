package aiagent

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
	"strings"
)

// AnalyzeReport 使用 LLM 对扫描报告进行风险评估
func AnalyzeReport(reportFile string, apiKey string) (string, error) {
	//1，读取gs-scout 生成的JSON报告 exporter.go
	content, err := os.ReadFile(reportFile)
	if err != nil {
		return "", fmt.Errorf("failed to read report file: %w", err)
	}
	// 2. 构造 Prompt (指令工程)
	// 关键步骤：告诉 AI 它的角色和目标。
	//systemInstruction := "你是一位经验丰富的网络安全工程师，请根据用户提供的端口扫描JSON报告，给出详细的风险评估和修复建议。"

	const userPrompt = `
你是一名专业的网络安全分析专家。
我会给你一份 JSON 格式的端口扫描报告。
请你分析这些端口的风险，并【必须】严格按照以下 JSON 数组格式返回结果，不要包含任何多余的解释文字、代码块标记（如 ` + "```json" + `）：

[
  {
    "port": 端口号,
    "service": "推断的服务名称",
    "level": "高/中/低",
    "risk": "简短的风险描述",
    "suggestion": "修复建议"
  }
]

扫描报告内容如下：
`
	// 3. 初始化 OpenAI 客户端
	config := openai.DefaultConfig(apiKey)
	// 设置 DeepSeek 的 Base URL
	config.BaseURL = "https://api.deepseek.com/v1"
	client := openai.NewClientWithConfig(config)

	//4. 调用 Chat API
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "deepseek-chat", //使用模型
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "你是一个只输出 JSON 格式结果的安全分析助手。不准输出任何非 JSON 文本.",
				},
				{Role: openai.ChatMessageRoleUser,
					Content: userPrompt + string(content),
				},
			},

			// 如果 DeepSeek 库支持，可以尝试开启 ResponseFormat
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to analyze report: %w", err)
	}
	// 5. 返回 AI 的回复内容
	// 4. 清洗结果 (防止 AI 带上 ```json 这样的 Markdown 标记)
	rawResult := resp.Choices[0].Message.Content
	cleanResult := strings.TrimPrefix(rawResult, "```json")
	cleanResult = strings.TrimPrefix(cleanResult, "```")
	cleanResult = strings.TrimSuffix(cleanResult, "```")
	cleanResult = strings.TrimSpace(cleanResult)
	return rawResult, nil

}
