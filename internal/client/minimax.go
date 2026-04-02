package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	APIBase              = "https://api.minimaxi.com/v1"
	OpenPlatformBase     = "https://www.minimaxi.com/v1/api/openplatform"
	AnthropicMessagesURL = "https://api.minimaxi.com/anthropic/v1/messages"
)

const (
	DefaultImageModel = "image-01"
	DefaultVideoModel = "MiniMax-Hailuo-2.3-Fast"
	DefaultTTSModel   = "speech-2.8-hd"
	DefaultMusicModel = "music-2.5"
	DefaultTextModel  = "MiniMax-M2.7-highspeed"
	DefaultVoiceID    = "male-qn-qingse"
)

// MiniMaxClient MiniMax API 客户端
type MiniMaxClient struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewClient 创建新的客户端
func NewClient(apiKey string) *MiniMaxClient {
	return &MiniMaxClient{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 600 * time.Second,
		},
	}
}

// GetAPIKey 从环境变量获取 API Key
func GetAPIKey() string {
	return os.Getenv("MINIMAX_API_KEY")
}

// --- 通用请求方法 ---

func (c *MiniMaxClient) requestJSON(ctx context.Context, method, path, label string, params, payload map[string]interface{}) (map[string]interface{}, error) {
	url := APIBase + path

	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to marshal request: %w", label, err)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", label, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", label, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read response: %w", label, err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s: HTTP %d: %s", label, resp.StatusCode, string(respBody))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("%s: failed to decode JSON: %w", label, err)
	}

	// 检查 base_resp
	baseResp, _ := data["base_resp"].(map[string]interface{})
	statusCode, _ := baseResp["status_code"].(float64)
	if int(statusCode) != 0 {
		statusMsg, _ := baseResp["status_msg"].(string)
		return nil, fmt.Errorf("%s failed: %d %s", label, int(statusCode), statusMsg)
	}

	return data, nil
}

func (c *MiniMaxClient) requestOpenPlatform(ctx context.Context, method, path, label string, params map[string]interface{}) (map[string]interface{}, error) {
	url := OpenPlatformBase + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", label, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", label, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read response: %w", label, err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s: HTTP %d: %s", label, resp.StatusCode, string(respBody))
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("%s: failed to decode JSON: %w", label, err)
	}

	baseResp, _ := data["base_resp"].(map[string]interface{})
	statusCode, _ := baseResp["status_code"].(float64)
	if int(statusCode) != 0 {
		statusMsg, _ := baseResp["status_msg"].(string)
		return nil, fmt.Errorf("%s failed: %d %s", label, int(statusCode), statusMsg)
	}

	return data, nil
}

// --- 图片生成 ---

// GenerateImage 生成图片，返回图片二进制数据
func (c *MiniMaxClient) GenerateImage(ctx context.Context, prompt, aspectRatio string, promptOptimizer bool) ([]byte, string, error) {
	payload := map[string]interface{}{
		"model":            DefaultImageModel,
		"prompt":           prompt,
		"aspect_ratio":     aspectRatio,
		"response_format":  "base64",
		"prompt_optimizer": promptOptimizer,
	}

	data, err := c.requestJSON(ctx, "POST", "/image_generation", "image_generation", nil, payload)
	if err != nil {
		return nil, "", err
	}

	dataField, _ := data["data"].(map[string]interface{})
	imageBase64, err := extractImageBase64(dataField["image_base64"])
	if err != nil {
		return nil, "", fmt.Errorf("image_generation returned invalid image_base64: %w", err)
	}
	if imageBase64 == "" {
		return nil, "", fmt.Errorf("image_generation returned no image_base64: %v", data)
	}

	imageBytes, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return nil, imageBase64, fmt.Errorf("failed to decode base64 image: %w", err)
	}

	return imageBytes, imageBase64, nil
}

func extractImageBase64(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []string:
		if len(v) == 0 {
			return "", nil
		}
		return v[0], nil
	case []interface{}:
		if len(v) == 0 {
			return "", nil
		}
		first, ok := v[0].(string)
		if !ok {
			return "", fmt.Errorf("first image entry is %T", v[0])
		}
		return first, nil
	default:
		return "", fmt.Errorf("unsupported type %T", value)
	}
}

// --- 视频生成 ---

// CreateVideoTask 创建视频生成任务
func (c *MiniMaxClient) CreateVideoTask(ctx context.Context, imageBase64, prompt, model string, duration int, resolution string) (string, error) {
	payload := map[string]interface{}{
		"model":             model,
		"prompt":            prompt,
		"first_frame_image": "data:image/jpeg;base64," + imageBase64,
		"duration":          duration,
		"resolution":        resolution,
	}

	data, err := c.requestJSON(ctx, "POST", "/video_generation", "video_generation", nil, payload)
	if err != nil {
		return "", err
	}

	// 响应: task_id (顶层字段)
	taskID, _ := data["task_id"].(string)
	if taskID == "" {
		return "", fmt.Errorf("video_generation returned no task_id: %v", data)
	}

	return taskID, nil
}

// PollVideoTask 轮询视频任务状态
func (c *MiniMaxClient) PollVideoTask(ctx context.Context, taskID string, intervalSeconds, maxWaitSeconds int, onStatus func(string)) (string, error) {
	deadline := time.Now().Add(time.Duration(maxWaitSeconds) * time.Second)

	for time.Now().Before(deadline) {
		data, err := c.requestJSON(ctx, "GET", "/query/video_generation", "query_video_generation",
			map[string]interface{}{"task_id": taskID}, nil)
		if err != nil {
			return "", err
		}

		status, _ := data["status"].(string)
		if onStatus != nil {
			onStatus(fmt.Sprintf("video task %s status: %s", taskID, status))
		}

		switch status {
		case "Success":
			fileID, _ := data["file_id"].(string)
			if fileID == "" {
				return "", fmt.Errorf("video task succeeded without file_id: %v", data)
			}
			return fileID, nil
		case "Fail", "Failed", "Expired":
			return "", fmt.Errorf("video task ended with status=%s: %v", status, data)
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Duration(intervalSeconds) * time.Second):
		}
	}

	return "", fmt.Errorf("video task %s timed out after %d seconds", taskID, maxWaitSeconds)
}

// FetchDownloadURL 获取文件下载 URL
func (c *MiniMaxClient) FetchDownloadURL(ctx context.Context, fileID string) (string, error) {
	data, err := c.requestJSON(ctx, "GET", "/files/retrieve", "files_retrieve",
		map[string]interface{}{"file_id": fileID}, nil)
	if err != nil {
		return "", err
	}

	fileObj, _ := data["file"].(map[string]interface{})
	downloadURL, _ := fileObj["download_url"].(string)
	if downloadURL == "" {
		return "", fmt.Errorf("files_retrieve returned no download_url: %v", data)
	}

	return downloadURL, nil
}

// DownloadFile 下载文件
func (c *MiniMaxClient) DownloadFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// --- 语音合成 ---

// SynthesizeSpeech 语音合成，返回音频二进制数据
func (c *MiniMaxClient) SynthesizeSpeech(ctx context.Context, text, voiceID, ttsModel, audioFormat string) ([]byte, error) {
	payload := map[string]interface{}{
		"model": ttsModel,
		"text":  text,
		"stream": false,
		"voice_setting": map[string]interface{}{
			"voice_id": voiceID,
			"speed":    1,
			"vol":      1,
			"pitch":    0,
		},
		"language_boost": "Chinese",
		"audio_setting": map[string]interface{}{
			"sample_rate": 32000,
			"bitrate":     128000,
			"format":      audioFormat,
			"channel":     1,
		},
		"subtitle_enable": false,
	}

	data, err := c.requestJSON(ctx, "POST", "/t2a_v2", "t2a_v2", nil, payload)
	if err != nil {
		return nil, err
	}

	// 响应: data.audio (hex string)
	dataField, _ := data["data"].(map[string]interface{})
	audioHex, _ := dataField["audio"].(string)
	if audioHex == "" {
		return nil, fmt.Errorf("t2a_v2 returned no audio: %v", data)
	}

	audioBytes, err := hex.DecodeString(audioHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex audio: %w", err)
	}

	return audioBytes, nil
}

// --- 音乐生成 ---

// GenerateMusic 生成音乐，返回音频二进制数据
func (c *MiniMaxClient) GenerateMusic(ctx context.Context, prompt, musicModel, audioFormat string) ([]byte, error) {
	payload := map[string]interface{}{
		"model":         musicModel,
		"prompt":        prompt,
		"stream":        false,
		"output_format": "hex",
		"aigc_watermark": false,
		"audio_setting": map[string]interface{}{
			"sample_rate": 44100,
			"bitrate":     256000,
			"format":      audioFormat,
		},
	}

	if musicModel == "music-2.5" {
		payload["lyrics"] = "[Inst]"
		payload["lyrics_optimizer"] = false
	} else {
		payload["is_instrumental"] = true
	}

	data, err := c.requestJSON(ctx, "POST", "/music_generation", "music_generation", nil, payload)
	if err != nil {
		return nil, err
	}

	// 响应: data.audio (hex string)
	dataField, _ := data["data"].(map[string]interface{})
	audioHex, _ := dataField["audio"].(string)
	if audioHex == "" {
		return nil, fmt.Errorf("music_generation returned no audio: %v", data)
	}

	audioBytes, err := hex.DecodeString(audioHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex audio: %w", err)
	}

	return audioBytes, nil
}

// --- 额度查询 ---

// GetQuota 查询额度
func (c *MiniMaxClient) GetQuota(ctx context.Context) ([]interface{}, error) {
	data, err := c.requestOpenPlatform(ctx, "GET", "/coding_plan/remains", "coding_plan_remains", nil)
	if err != nil {
		return nil, err
	}

	modelRemains, _ := data["model_remains"].([]interface{})
	return modelRemains, nil
}

// --- 分镜规划（Anthropic 兼容接口） ---

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

// PlanVideoRequest 分镜规划请求参数
type PlanVideoRequest struct {
	Theme         string
	SceneCount    int
	SceneDuration int
	Language      string
	TextModel     string
	TextMaxTokens int
}

// PlanVideo 调用 MiniMax 文本模型生成分镜规划
func (c *MiniMaxClient) PlanVideo(ctx context.Context, req PlanVideoRequest) (*PlanVideoResponse, error) {
	systemPrompt := "You are a video creative planner. Return valid JSON only. Do not use markdown fences. Do not include explanations. Provide concise, production-ready prompts."

	maxChars := maxInt(18, req.SceneCount*req.SceneDuration*5)

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
		return nil, fmt.Errorf("plan: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", AnthropicMessagesURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("plan: failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("plan: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("plan: failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("plan: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// 检查 base_resp
	var rawResp map[string]interface{}
	if err := json.Unmarshal(respBody, &rawResp); err != nil {
		return nil, fmt.Errorf("plan: failed to decode response: %w", err)
	}

	baseResp, _ := rawResp["base_resp"].(map[string]interface{})
	statusCode, _ := baseResp["status_code"].(float64)
	if int(statusCode) != 0 {
		statusMsg, _ := baseResp["status_msg"].(string)
		return nil, fmt.Errorf("plan failed: %d %s", int(statusCode), statusMsg)
	}

	// 提取文本内容
	contentBlocks, _ := rawResp["content"].([]interface{})
	var textParts []string
	for _, block := range contentBlocks {
		b, _ := block.(map[string]interface{})
		if b["type"] == "text" {
			if text, ok := b["text"].(string); ok && text != "" {
				textParts = append(textParts, text)
			}
		}
	}

	if len(textParts) == 0 {
		return nil, fmt.Errorf("plan: anthropic_messages returned no text content: %s", string(respBody))
	}

	content := strings.Join(textParts, "\n")
	return parsePlanJSON(content, req.SceneCount)
}

// --- JSON 解析辅助 ---

func parsePlanJSON(text string, expectedSceneCount int) (*PlanVideoResponse, error) {
	// 移除 thinking 标签
	re := regexp.MustCompile(`(?s)<thinking>.*?</thinking>`)
	text = re.ReplaceAllString(text, "")
	text = strings.TrimSpace(text)

	// 移除 markdown code fences
	if strings.HasPrefix(text, "```") {
		re2 := regexp.MustCompile("(?s)^```[a-zA-Z0-9_-]*\n?(.*?)\n?```$")
		if matches := re2.FindStringSubmatch(text); len(matches) > 1 {
			text = matches[1]
		}
	}

	// 尝试直接解析
	var plan PlanVideoResponse
	if err := json.Unmarshal([]byte(text), &plan); err != nil {
		// 尝试提取 JSON 对象
		re3 := regexp.MustCompile(`(?s)\{.*\}`)
		match := re3.FindString(text)
		if match == "" {
			return nil, fmt.Errorf("failed to locate JSON object in response: %s", truncate(text, 200))
		}
		if err := json.Unmarshal([]byte(match), &plan); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	if expectedSceneCount > 0 && len(plan.Scenes) != expectedSceneCount {
		return nil, fmt.Errorf("plan returned %d scenes, expected %d", len(plan.Scenes), expectedSceneCount)
	}

	return &plan, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
