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
	"time"
)

const (
	BaseURL = "https://api.minimax.chat"
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
