package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	BaseURL             = "https://api.minimax.chat"
	OpenPlatformBaseURL = "https://www.minimaxi.com/v1/api/openplatform"
	AnthropicMessagesURL = "https://api.minimaxi.com/anthropic/v1/messages"
)

// MiniMaxClient MiniMax API 客户端
type MiniMaxClient struct {
	APIKey     string
	GroupID    string
	HTTPClient *http.Client
	BaseURL    string
}

// NewClient 创建新的客户端
func NewClient(apiKey, groupID string) *MiniMaxClient {
	return &MiniMaxClient{
		APIKey:  apiKey,
		GroupID: groupID,
		HTTPClient: &http.Client{
			Timeout: 600 * time.Second,
		},
		BaseURL: BaseURL,
	}
}

// GetAPIKey 从环境变量获取 API Key
func GetAPIKey() string {
	return os.Getenv("MINIMAX_API_KEY")
}

// GetGroupID 从环境变量获取 Group ID
func GetGroupID() string {
	return os.Getenv("MINIMAX_GROUP_ID")
}

// ImageResponse 图片生成响应
type ImageResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ImageBase64 string `json:"image_base64"`
		ImageURL    string `json:"image_url"`
	} `json:"data"`
}

// GenerateImage 生成图片
func (c *MiniMaxClient) GenerateImage(ctx context.Context, prompt, aspectRatio string, promptOptimizer bool) ([]byte, error) {
	url := fmt.Sprintf("%s/v1/images/generations", c.BaseURL)

	payload := map[string]interface{}{
		"model":            "image-01",
		"prompt":           prompt,
		"aspect_ratio":     aspectRatio,
		"prompt_optimizer": promptOptimizer,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkError(resp); err != nil {
		return nil, err
	}

	var result ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return base64.StdEncoding.DecodeString(result.Data.ImageBase64)
}

// VideoTaskResponse 视频任务响应
type VideoTaskResponse struct {
	Success bool `json:"success"`
	Data    struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
}

// CreateVideoTask 创建视频生成任务
func (c *MiniMaxClient) CreateVideoTask(ctx context.Context, imageBase64, prompt, model string, duration int, resolution string) (string, error) {
	url := fmt.Sprintf("%s/v1/video_generations", c.BaseURL)

	payload := map[string]interface{}{
		"model": model,
		"input": map[string]interface{}{
			"image_base64": imageBase64,
			"prompt":       prompt,
		},
		"config": map[string]interface{}{
			"duration":   duration,
			"resolution": resolution,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkError(resp); err != nil {
		return "", err
	}

	var result VideoTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data.TaskID, nil
}

// VideoStatusResponse 视频任务状态响应
type VideoStatusResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Status      string `json:"status"`
		FileID      string `json:"file_id"`
		DownloadURL string `json:"download_url"`
	} `json:"data"`
}

// PollVideoTask 轮询视频任务状态
func (c *MiniMaxClient) PollVideoTask(ctx context.Context, taskID string, intervalSeconds, maxWaitSeconds int, onStatus func(string)) (string, error) {
	url := fmt.Sprintf("%s/v1/video_generations/%s", c.BaseURL, taskID)

	deadline := time.Now().Add(time.Duration(maxWaitSeconds) * time.Second)

	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		c.setHeaders(req)
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to send request: %w", err)
		}

		var result VideoStatusResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return "", fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		if onStatus != nil {
			onStatus(result.Data.Status)
		}

		switch result.Data.Status {
		case "success":
			return result.Data.FileID, nil
		case "failed":
			return "", fmt.Errorf("video generation failed")
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(intervalSeconds) * time.Second):
		}
	}

	return "", fmt.Errorf("timeout waiting for video generation")
}

// DownloadURLResponse 下载 URL 响应
type DownloadURLResponse struct {
	Success bool `json:"success"`
	Data    struct {
		URL string `json:"url"`
	} `json:"data"`
}

// FetchDownloadURL 获取下载 URL
func (c *MiniMaxClient) FetchDownloadURL(ctx context.Context, fileID string) (string, error) {
	url := fmt.Sprintf("%s/v1/files/%s/download", c.BaseURL, fileID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkError(resp); err != nil {
		return "", err
	}

	var result DownloadURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data.URL, nil
}

// DownloadFile 下载文件
func (c *MiniMaxClient) DownloadFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// SynthesizeSpeech 语音合成
func (c *MiniMaxClient) SynthesizeSpeech(ctx context.Context, text, voiceID, model, audioFormat string) ([]byte, error) {
	url := fmt.Sprintf("%s/v1/t2a_v2", c.BaseURL)

	payload := map[string]interface{}{
		"model":        model,
		"text":         text,
		"voice_id":     voiceID,
		"audio_format": audioFormat,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkError(resp); err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

// GenerateMusic 生成音乐
func (c *MiniMaxClient) GenerateMusic(ctx context.Context, prompt, model, audioFormat string) (string, error) {
	url := fmt.Sprintf("%s/v1/music_generations", c.BaseURL)

	payload := map[string]interface{}{
		"model":        model,
		"prompt":       prompt,
		"audio_format": audioFormat,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkError(resp); err != nil {
		return "", err
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			TaskID string `json:"task_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data.TaskID, nil
}

// QuotaResponse 额度查询响应
type QuotaResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ModelInfos []struct {
			ModelName                 string `json:"model_name"`
			StartTime                 int64  `json:"start_time"`
			EndTime                   int64  `json:"end_time"`
			RemainsTime               int64  `json:"remains_time"`
			CurrentIntervalTotalCount int    `json:"current_interval_total_count"`
			CurrentIntervalUsageCount int    `json:"current_interval_usage_count"`
			WeeklyStartTime           int64  `json:"weekly_start_time"`
			WeeklyEndTime             int64  `json:"weekly_end_time"`
			CurrentWeeklyTotalCount   int    `json:"current_weekly_total_count"`
			CurrentWeeklyUsageCount   int    `json:"current_weekly_usage_count"`
			WeeklyRemainsTime         int64  `json:"weekly_remains_time"`
		} `json:"model_infos"`
	} `json:"data"`
}

// GetQuota 查询额度
func (c *MiniMaxClient) GetQuota(ctx context.Context) (*QuotaResponse, error) {
	url := fmt.Sprintf("%s/v1/coding_plan/remains", c.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkError(resp); err != nil {
		return nil, err
	}

	var result QuotaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AnthropicResponse Anthropic 兼容接口响应
type AnthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	BaseResp struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}

// PlanVideoRequest 分镜规划请求参数
type PlanVideoRequest struct {
	Theme         string
	SceneCount    int
	SceneDuration int
	Language      string
	TextModel     string
	TextMaxTokens int
}

// PlanVideoResponse 分镜规划响应
type PlanVideoResponse struct {
	Title       string `json:"title"`
	VisualStyle string `json:"visual_style"`
	Narration   string `json:"narration"`
	MusicPrompt string `json:"music_prompt"`
	Scenes      []struct {
		Name        string `json:"name"`
		ImagePrompt string `json:"image_prompt"`
		VideoPrompt string `json:"video_prompt"`
	} `json:"scenes"`
}

// PlanVideo 调用 MiniMax 文本模型生成分镜规划
func (c *MiniMaxClient) PlanVideo(ctx context.Context, req PlanVideoRequest) (*PlanVideoResponse, error) {
	systemPrompt := "You are a video creative planner. Return valid JSON only. Do not use markdown fences. Do not include explanations. Provide concise, production-ready prompts."

	maxChars := max(18, req.SceneCount*req.SceneDuration*5)

	userPrompt := fmt.Sprintf(`为主题"%s"生成一个短视频制作方案。

要求：
1. 输出 JSON 对象，字段必须完整。
2. scenes 数组长度必须等于 %d。
3. 每个 scene 的画面提示词和运动提示词用英文，适合 AI 图片/视频生成。
4. narration 使用 %s，必须简短自然，总长度控制在 %d 个字符以内，适配总时长约 %d 秒。
5. music_prompt 用英文，描述纯音乐，不要人声。
6. 风格要统一，适合短视频成片。

JSON Schema:
{
  "title": "string",
  "visual_style": "string",
  "narration": "string",
  "music_prompt": "string",
  "scenes": [
    {
      "name": "string",
      "image_prompt": "string",
      "video_prompt": "string"
    }
  ]
}`, req.Theme, req.SceneCount, req.Language, maxChars, req.SceneCount*req.SceneDuration)

	payload := map[string]interface{}{
		"model":      req.TextModel,
		"max_tokens": req.TextMaxTokens,
		"system":     systemPrompt,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "text", "text": userPrompt},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", AnthropicMessagesURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if anthropicResp.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("anthropic_messages failed: %d %s", anthropicResp.BaseResp.StatusCode, anthropicResp.BaseResp.StatusMsg)
	}

	// 提取文本内容
	var textParts []string
	for _, block := range anthropicResp.Content {
		if block.Type == "text" && block.Text != "" {
			textParts = append(textParts, block.Text)
		}
	}

	if len(textParts) == 0 {
		return nil, fmt.Errorf("anthropic_messages returned no text content")
	}

	content := joinNonEmpty(textParts)

	// 解析 JSON
	return parsePlanJSON(content, req.SceneCount)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func joinNonEmpty(parts []string) string {
	var result string
	for i, p := range parts {
		if p != "" {
			if i > 0 {
				result += "\n"
			}
			result += p
		}
	}
	return result
}

func parsePlanJSON(text string, expectedSceneCount int) (*PlanVideoResponse, error) {
	// 移除 thinking 标签
	re := regexp.MustCompile(`<thinking>.*?</thinking>`)
	text = re.ReplaceAllString(text, "")

	// 移除 markdown code fences
	text = strings.TrimSpace(text)
	re2 := regexp.MustCompile("(?s)^```[a-zA-Z0-9_-]*\n?(.*?)\n?```$")
	if matches := re2.FindStringSubmatch(text); len(matches) > 1 {
		text = matches[1]
	}

	// 尝试直接解析
	var plan PlanVideoResponse
	if err := json.Unmarshal([]byte(text), &plan); err != nil {
		// 尝试提取 JSON 对象
		re3 := regexp.MustCompile(`(?s)\{.*\}`)
		match := re3.FindString(text)
		if match == "" {
			return nil, fmt.Errorf("failed to locate JSON object in response")
		}
		if err := json.Unmarshal([]byte(match), &plan); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	// 验证场景数量
	if expectedSceneCount > 0 && len(plan.Scenes) != expectedSceneCount {
		return nil, fmt.Errorf("plan returned %d scenes, expected %d", len(plan.Scenes), expectedSceneCount)
	}

	return &plan, nil
}

func (c *MiniMaxClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	if c.GroupID != "" {
		req.Header.Set("GroupId", c.GroupID)
	}
}

func (c *MiniMaxClient) checkError(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
