package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGenerateImage_PayloadFormat 验证图片生成请求的 payload 格式与 Python 版本一致
func TestGenerateImage_PayloadFormat(t *testing.T) {
	var receivedPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		if r.URL.Path != "/image_generation" {
			t.Errorf("expected path /image_generation, got %s", r.URL.Path)
		}

		// 验证请求方法
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// 验证 Authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			t.Errorf("expected Authorization 'Bearer test-api-key', got '%s'", auth)
		}

		// 捕获 payload
		json.NewDecoder(r.Body).Decode(&receivedPayload)

		// 返回模拟响应（与 Python 版本一致）
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// 返回一个有效的 base64 编码的 1x1 PNG
		w.Write([]byte(`{
			"base_resp": {"status_code": 0},
			"data": {
				"image_base64": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
			}
		}`))
	}))
	defer server.Close()

	// 替换 API base URL
	origAPIBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = origAPIBase }()

	client := NewClient("test-api-key")
	ctx := context.Background()

	imageData, imageBase64, err := client.GenerateImage(ctx, "a paper boat", "16:9", true)
	if err != nil {
		t.Fatalf("GenerateImage failed: %v", err)
	}

	// 验证返回值
	if len(imageData) == 0 {
		t.Error("expected non-empty image data")
	}
	if imageBase64 == "" {
		t.Error("expected non-empty base64 string")
	}

	// 验证 payload 格式（与 Python 版本一致）
	if receivedPayload["model"] != "image-01" {
		t.Errorf("expected model 'image-01', got '%v'", receivedPayload["model"])
	}
	if receivedPayload["prompt"] != "a paper boat" {
		t.Errorf("expected prompt 'a paper boat', got '%v'", receivedPayload["prompt"])
	}
	if receivedPayload["aspect_ratio"] != "16:9" {
		t.Errorf("expected aspect_ratio '16:9', got '%v'", receivedPayload["aspect_ratio"])
	}
	if receivedPayload["response_format"] != "base64" {
		t.Errorf("expected response_format 'base64', got '%v'", receivedPayload["response_format"])
	}
	if receivedPayload["prompt_optimizer"] != true {
		t.Errorf("expected prompt_optimizer true, got '%v'", receivedPayload["prompt_optimizer"])
	}
}

// TestCreateVideoTask_PayloadFormat 验证视频任务创建的 payload 格式与 Python 版本一致
func TestCreateVideoTask_PayloadFormat(t *testing.T) {
	var receivedPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/video_generation" {
			t.Errorf("expected path /video_generation, got %s", r.URL.Path)
		}

		json.NewDecoder(r.Body).Decode(&receivedPayload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"base_resp": {"status_code": 0},
			"task_id": "task_12345"
		}`))
	}))
	defer server.Close()

	origAPIBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = origAPIBase }()

	client := NewClient("test-api-key")
	ctx := context.Background()

	taskID, err := client.CreateVideoTask(ctx, "base64imagedata", "drift slowly", "MiniMax-Hailuo-2.3-Fast", 5, "720p")
	if err != nil {
		t.Fatalf("CreateVideoTask failed: %v", err)
	}

	if taskID != "task_12345" {
		t.Errorf("expected task_id 'task_12345', got '%s'", taskID)
	}

	// 验证 payload：Python 版本使用 first_frame_image（带 data:image/jpeg;base64, 前缀）
	if receivedPayload["model"] != "MiniMax-Hailuo-2.3-Fast" {
		t.Errorf("expected model 'MiniMax-Hailuo-2.3-Fast', got '%v'", receivedPayload["model"])
	}
	if receivedPayload["prompt"] != "drift slowly" {
		t.Errorf("expected prompt, got '%v'", receivedPayload["prompt"])
	}

	// 关键：Python 版使用 first_frame_image，不是 input.image_base64
	ffi, _ := receivedPayload["first_frame_image"].(string)
	if ffi != "data:image/jpeg;base64,base64imagedata" {
		t.Errorf("expected first_frame_image with data URI prefix, got '%s'", ffi)
	}

	if receivedPayload["duration"] != float64(5) {
		t.Errorf("expected duration 5, got '%v'", receivedPayload["duration"])
	}
	if receivedPayload["resolution"] != "720p" {
		t.Errorf("expected resolution '720p', got '%v'", receivedPayload["resolution"])
	}
}

// TestPollVideoTask_UsesQueryParam 验证轮询使用 task_id 查询参数（与 Python 版本一致）
func TestPollVideoTask_UsesQueryParam(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Python 版本使用 GET /query/video_generation?task_id=xxx
		if r.URL.Path != "/query/video_generation" {
			t.Errorf("expected path /query/video_generation, got %s", r.URL.Path)
		}

		taskID := r.URL.Query().Get("task_id")
		if taskID != "task_test" {
			t.Errorf("expected task_id param 'task_test', got '%s'", taskID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)

		if callCount < 3 {
			w.Write([]byte(`{"base_resp": {"status_code": 0}, "status": "Processing"}`))
		} else {
			w.Write([]byte(`{"base_resp": {"status_code": 0}, "status": "Success", "file_id": "file_abc"}`))
		}
	}))
	defer server.Close()

	origAPIBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = origAPIBase }()

	client := NewClient("test-api-key")
	ctx := context.Background()

	fileID, err := client.PollVideoTask(ctx, "task_test", 0, 30, nil)
	if err != nil {
		t.Fatalf("PollVideoTask failed: %v", err)
	}

	if fileID != "file_abc" {
		t.Errorf("expected file_id 'file_abc', got '%s'", fileID)
	}
}

// TestFetchDownloadURL_UsesQueryParam 验证获取下载 URL 使用 file_id 查询参数
func TestFetchDownloadURL_UsesQueryParam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Python 版本使用 GET /files/retrieve?file_id=xxx
		if r.URL.Path != "/files/retrieve" {
			t.Errorf("expected path /files/retrieve, got %s", r.URL.Path)
		}

		fileID := r.URL.Query().Get("file_id")
		if fileID != "file_test" {
			t.Errorf("expected file_id param 'file_test', got '%s'", fileID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// Python 版本返回 file.download_url
		w.Write([]byte(`{
			"base_resp": {"status_code": 0},
			"file": {"download_url": "https://cdn.example.com/video.mp4"}
		}`))
	}))
	defer server.Close()

	origAPIBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = origAPIBase }()

	client := NewClient("test-api-key")
	ctx := context.Background()

	url, err := client.FetchDownloadURL(ctx, "file_test")
	if err != nil {
		t.Fatalf("FetchDownloadURL failed: %v", err)
	}

	if url != "https://cdn.example.com/video.mp4" {
		t.Errorf("expected download URL, got '%s'", url)
	}
}

// TestSynthesizeSpeech_PayloadFormat 验证 TTS payload 格式与 Python 版本一致
func TestSynthesizeSpeech_PayloadFormat(t *testing.T) {
	var receivedPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/t2a_v2" {
			t.Errorf("expected path /t2a_v2, got %s", r.URL.Path)
		}

		json.NewDecoder(r.Body).Decode(&receivedPayload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		// Python 版本返回 data.audio（hex 编码）
		w.Write([]byte(fmt.Sprintf(`{
			"base_resp": {"status_code": 0},
			"data": {"audio": "%s"}
		}`, "48656c6c6f"))) // "Hello" in hex
	}))
	defer server.Close()

	origAPIBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = origAPIBase }()

	client := NewClient("test-api-key")
	ctx := context.Background()

	audioData, err := client.SynthesizeSpeech(ctx, "这是一段旁白", "male-qn-qingse", "speech-2.8-hd", "mp3")
	if err != nil {
		t.Fatalf("SynthesizeSpeech failed: %v", err)
	}

	if string(audioData) != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", string(audioData))
	}

	// 验证 payload 格式
	if receivedPayload["model"] != "speech-2.8-hd" {
		t.Errorf("expected model 'speech-2.8-hd', got '%v'", receivedPayload["model"])
	}
	if receivedPayload["text"] != "这是一段旁白" {
		t.Errorf("expected text, got '%v'", receivedPayload["text"])
	}
	if receivedPayload["stream"] != false {
		t.Errorf("expected stream false, got '%v'", receivedPayload["stream"])
	}

	// 验证 voice_setting（Python 版本有完整结构）
	vs, _ := receivedPayload["voice_setting"].(map[string]interface{})
	if vs["voice_id"] != "male-qn-qingse" {
		t.Errorf("expected voice_id 'male-qn-qingse', got '%v'", vs["voice_id"])
	}

	// 验证 audio_setting（Python 版本有完整结构）
	as, _ := receivedPayload["audio_setting"].(map[string]interface{})
	if as["format"] != "mp3" {
		t.Errorf("expected format 'mp3', got '%v'", as["format"])
	}

	if receivedPayload["language_boost"] != "Chinese" {
		t.Errorf("expected language_boost 'Chinese', got '%v'", receivedPayload["language_boost"])
	}
}

// TestGenerateMusic_PayloadFormat 验证音乐生成 payload 格式与 Python 版本一致
func TestGenerateMusic_PayloadFormat(t *testing.T) {
	var receivedPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/music_generation" {
			t.Errorf("expected path /music_generation, got %s", r.URL.Path)
		}

		json.NewDecoder(r.Body).Decode(&receivedPayload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`{
			"base_resp": {"status_code": 0},
			"data": {"audio": "%s"}
		}`, "48656c6c6f")))
	}))
	defer server.Close()

	origAPIBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = origAPIBase }()

	client := NewClient("test-api-key")
	ctx := context.Background()

	audioData, err := client.GenerateMusic(ctx, "warm piano", "music-2.5", "mp3")
	if err != nil {
		t.Fatalf("GenerateMusic failed: %v", err)
	}

	if string(audioData) != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", string(audioData))
	}

	// 验证 music-2.5 特有字段（Python 版本）
	if receivedPayload["lyrics"] != "[Inst]" {
		t.Errorf("expected lyrics '[Inst]', got '%v'", receivedPayload["lyrics"])
	}
	if receivedPayload["lyrics_optimizer"] != false {
		t.Errorf("expected lyrics_optimizer false, got '%v'", receivedPayload["lyrics_optimizer"])
	}
	if receivedPayload["output_format"] != "hex" {
		t.Errorf("expected output_format 'hex', got '%v'", receivedPayload["output_format"])
	}
	if receivedPayload["aigc_watermark"] != false {
		t.Errorf("expected aigc_watermark false, got '%v'", receivedPayload["aigc_watermark"])
	}
}

// TestPlanVideo_AnthropicFormat 验证分镜规划使用 Anthropic 兼容接口
func TestPlanVideo_AnthropicFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证使用 Anthropic 兼容端点
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("expected x-api-key header, got '%s'", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version '2023-06-01', got '%s'", r.Header.Get("anthropic-version"))
		}

		// 返回 Anthropic 格式响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"base_resp": {"status_code": 0},
			"content": [
				{"type": "text", "text": "{\"title\":\"纸船\",\"visual_style\":\"cinematic\",\"narration\":\"一艘纸船缓缓前行\",\"music_prompt\":\"gentle piano\",\"scenes\":[{\"name\":\"scene1\",\"image_prompt\":\"paper boat\",\"video_prompt\":\"drifts\"}]}"}
			]
		}`))
	}))
	defer server.Close()

	// 修改 AnthropicMessagesURL
	origURL := AnthropicMessagesURL
	AnthropicMessagesURL = server.URL
	defer func() { AnthropicMessagesURL = origURL }()

	client := NewClient("test-api-key")
	ctx := context.Background()

	plan, err := client.PlanVideo(ctx, PlanVideoRequest{
		Theme:         "纸船晨光",
		SceneCount:    1,
		SceneDuration: 6,
		Language:      "zh",
		TextModel:     "MiniMax-M2.7-highspeed",
		TextMaxTokens: 4096,
	})
	if err != nil {
		t.Fatalf("PlanVideo failed: %v", err)
	}

	if plan.Title != "纸船" {
		t.Errorf("expected title '纸船', got '%s'", plan.Title)
	}
	if len(plan.Scenes) != 1 {
		t.Errorf("expected 1 scene, got %d", len(plan.Scenes))
	}
	if plan.Scenes[0].ImagePrompt != "paper boat" {
		t.Errorf("expected image_prompt 'paper boat', got '%s'", plan.Scenes[0].ImagePrompt)
	}
}

// TestGetQuota_UsesOpenPlatform 验证额度查询使用 OpenPlatform 端点
func TestGetQuota_UsesOpenPlatform(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Python 版本使用 OpenPlatform 端点
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"base_resp": {"status_code": 0},
			"model_remains": [
				{"model_name": "test-model", "remains_time": 1000}
			]
		}`))
	}))
	defer server.Close()

	origBase := OpenPlatformBase
	OpenPlatformBase = server.URL
	defer func() { OpenPlatformBase = origBase }()

	client := NewClient("test-api-key")
	ctx := context.Background()

	result, err := client.GetQuota(ctx)
	if err != nil {
		t.Fatalf("GetQuota failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result))
	}
}

// TestBaseRespErrorHandling 验证 base_resp 错误处理
func TestBaseRespErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"base_resp": {"status_code": 1001, "status_msg": "invalid api key"}
		}`))
	}))
	defer server.Close()

	origAPIBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = origAPIBase }()

	client := NewClient("bad-key")
	ctx := context.Background()

	_, _, err := client.GenerateImage(ctx, "test", "16:9", false)
	if err == nil {
		t.Fatal("expected error for non-zero base_resp status_code")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}
