package aiagent

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
)

// AnalyzeReport 使用 LLM 对扫描报告进行风险评估
func AnalyzeReport(jsonFilePath string, apiKey string) (string, error) {
	//1，读取gs-scout 生成的JSON报告 exporter.go
	reportData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read report file: %w", err)
	}
	// 2. 构造 Prompt (指令工程)
	// 关键步骤：告诉 AI 它的角色和目标。
	systemInstruction := "你是一位经验丰富的网络安全工程师，请根据用户提供的端口扫描JSON报告，给出详细的风险评估和修复建议。"

	userPrompt := fmt.Sprintf(`请对以下端口扫描报告进行分析，关注开放的 Web 服务（如 Server: Apache/Nginx）：
     1. 列出发现的开放服务和指纹。
        2. 对每个开放端口给出中/高/低风险评级。
        3. 给出针对性的修复建议。
        
        JSON报告内容如下:
        ---
        %s
        ---`, string(reportData))
	// 3. 初始化 OpenAI 客户端
	//client := openai.NewClient(apiKey)
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
					Content: systemInstruction,
				},
				{Role: openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to analyze report: %w", err)
	}
	// 5. 返回 AI 的回复内容
	return resp.Choices[0].Message.Content, nil
}
