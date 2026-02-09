package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sublink/models"
	"sublink/utils"
	"time"
)

// httpClientTimeout is the timeout duration for LLM API requests
const httpClientTimeout = 120 * time.Second

// ChatMessage represents a message in the chat completion request
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents an OpenAI-compatible chat completion request
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatChoice represents a choice in the chat completion response
type ChatChoice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
}

// ChatResponse represents an OpenAI-compatible chat completion response
type ChatResponse struct {
	ID      string       `json:"id"`
	Choices []ChatChoice `json:"choices"`
}

// LLMConfig holds the LLM API configuration
type LLMConfig struct {
	APIUrl string `json:"apiUrl"`
	APIKey string `json:"apiKey"`
	Model  string `json:"model"`
}

// GetConfig retrieves the LLM configuration from system settings
func GetConfig() (*LLMConfig, error) {
	apiUrl, _ := models.GetSetting("llm_api_url")
	apiKey, _ := models.GetSetting("llm_api_key")
	model, _ := models.GetSetting("llm_model")

	if apiUrl == "" {
		return nil, fmt.Errorf("LLM API URL 未配置")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("LLM API Key 未配置")
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	return &LLMConfig{
		APIUrl: apiUrl,
		APIKey: apiKey,
		Model:  model,
	}, nil
}

// buildEndpointURL constructs the chat completions endpoint URL from the base API URL
func buildEndpointURL(apiUrl string) string {
	if strings.HasSuffix(apiUrl, "/chat/completions") {
		return apiUrl
	}
	if strings.HasSuffix(apiUrl, "/v1") {
		return apiUrl + "/chat/completions"
	}
	return strings.TrimRight(apiUrl, "/") + "/v1/chat/completions"
}

// callAPI sends a chat completion request to the OpenAI-compatible API
func callAPI(config *LLMConfig, messages []ChatMessage) (string, error) {
	reqBody := ChatRequest{
		Model:       config.Model,
		Messages:    messages,
		Temperature: 0.7,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	apiEndpoint := buildEndpointURL(config.APIUrl)

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{Timeout: httpClientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求LLM API失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API返回错误 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM API未返回有效结果")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// NodeInfo represents simplified node info for LLM processing
type NodeInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Country  string `json:"country"`
	Group    string `json:"group"`
}

// OrganizeNodes uses the LLM to organize and categorize nodes
func OrganizeNodes(nodes []NodeInfo, instruction string) (string, error) {
	config, err := GetConfig()
	if err != nil {
		return "", err
	}

	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		return "", fmt.Errorf("序列化节点信息失败: %v", err)
	}

	systemPrompt := `你是一个代理节点整理助手。你的任务是根据用户的指令，对代理节点进行分类、整理和建议。
请以JSON格式返回结果。

返回格式要求:
{
  "groups": [
    {
      "name": "分组名称",
      "nodeIds": [1, 2, 3],
      "description": "分组说明"
    }
  ],
  "suggestions": "整理建议和说明"
}

注意：
- nodeIds必须使用原始节点的id
- 只返回JSON，不要包含其他内容
- 分组名称应该简洁明了`

	userPrompt := fmt.Sprintf("以下是需要整理的节点列表：\n%s\n\n用户指令：%s", string(nodesJSON), instruction)
	if instruction == "" {
		userPrompt = fmt.Sprintf("以下是需要整理的节点列表：\n%s\n\n请按照地区和协议对节点进行分组整理。", string(nodesJSON))
	}

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	result, err := callAPI(config, messages)
	if err != nil {
		return "", err
	}

	utils.Info("LLM节点整理完成")
	return result, nil
}

// GenerateRules uses the LLM to generate subscription rules based on selected nodes
func GenerateRules(nodes []NodeInfo, clientType string, instruction string) (string, error) {
	config, err := GetConfig()
	if err != nil {
		return "", err
	}

	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		return "", fmt.Errorf("序列化节点信息失败: %v", err)
	}

	var formatDesc string
	switch clientType {
	case "clash":
		formatDesc = `Clash/Mihomo YAML格式的rules部分。示例格式:
rules:
  - DOMAIN-SUFFIX,google.com,节点分组名
  - GEOIP,CN,DIRECT
  - MATCH,节点分组名`
	case "surge":
		formatDesc = `Surge规则格式。示例格式:
[Rule]
DOMAIN-SUFFIX,google.com,节点分组名
GEOIP,CN,DIRECT
FINAL,节点分组名`
	default:
		formatDesc = "通用代理规则格式"
	}

	systemPrompt := fmt.Sprintf(`你是一个代理订阅规则生成助手。根据用户提供的节点信息和需求，生成合适的%s订阅规则。

规则格式要求：%s

请以JSON格式返回结果：
{
  "rules": "生成的规则内容（字符串形式）",
  "proxyGroups": [
    {
      "name": "分组名称",
      "type": "select/url-test/fallback",
      "nodeIds": [1, 2, 3]
    }
  ],
  "description": "规则说明"
}

注意：
- 只返回JSON，不要包含其他内容
- 规则应该包含常用的分流规则（如国内直连、国外代理等）
- 代理组名称应该简洁明了
- nodeIds必须使用原始节点的id`, clientType, formatDesc)

	userPrompt := fmt.Sprintf("以下是可用的节点列表：\n%s\n\n", string(nodesJSON))
	if instruction != "" {
		userPrompt += fmt.Sprintf("用户需求：%s", instruction)
	} else {
		userPrompt += "请根据节点的地区和类型，生成合适的代理分流规则。包含国内直连、国外代理、流媒体分流等常用规则。"
	}

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	result, err := callAPI(config, messages)
	if err != nil {
		return "", err
	}

	utils.Info("LLM规则生成完成")
	return result, nil
}

// TestConnection tests the LLM API connection
func TestConnection(config *LLMConfig) error {
	messages := []ChatMessage{
		{Role: "user", Content: "请回复 ok"},
	}

	result, err := callAPI(config, messages)
	if err != nil {
		return err
	}

	if result == "" {
		return fmt.Errorf("LLM API返回空结果")
	}

	return nil
}
