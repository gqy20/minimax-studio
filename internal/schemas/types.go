package schemas

import "time"

// ScenePlan 单个分镜
type ScenePlan struct {
	Name        string `json:"name"`
	ImagePrompt string `json:"image_prompt"`
	VideoPrompt string `json:"video_prompt"`
}

// VideoPlan 完整分镜规划
type VideoPlan struct {
	Title       string      `json:"title"`
	VisualStyle string      `json:"visual_style"`
	Narration   string      `json:"narration"`
	MusicPrompt string      `json:"music_prompt"`
	Scenes      []ScenePlan `json:"scenes"`
}

// ClipOptions 图生视频选项
type ClipOptions struct {
	ImagePrompt          string `json:"image_prompt"`
	VideoPrompt          string `json:"video_prompt"`
	AspectRatio          string `json:"aspect_ratio"`
	VideoModel           string `json:"video_model"`
	Duration             int    `json:"duration"`
	Resolution           string `json:"resolution"`
	PollInterval         int    `json:"poll_interval"`
	MaxWait              int    `json:"max_wait"`
	OutputPrefix         string `json:"output_prefix"`
	ImagePromptOptimizer bool   `json:"image_prompt_optimizer"`
}

// ClipResult 图生视频结果
type ClipResult struct {
	ImagePath   string `json:"image_path"`
	VideoPath   string `json:"video_path"`
	TaskID      string `json:"task_id"`
	FileID      string `json:"file_id"`
	DownloadURL string `json:"download_url"`
}

// ImageResult 单图生成结果
type ImageResult struct {
	ImagePath string `json:"image_path"`
}

// PlanOptions 分镜规划选项
type PlanOptions struct {
	Theme         string `json:"theme"`
	SceneCount    int    `json:"scene_count"`
	SceneDuration int    `json:"scene_duration"`
	Language      string `json:"language"`
	TextModel     string `json:"text_model"`
	TextMaxTokens int    `json:"text_max_tokens"`
	OutputDir     string `json:"output_dir"`
}

// PlanResult 分镜规划结果
type PlanResult struct {
	OutputDir     string `json:"output_dir"`
	PlanPath      string `json:"plan_path"`
	NarrationPath string `json:"narration_path"`
}

// VoiceOptions 语音合成选项
type VoiceOptions struct {
	Text        string `json:"text"`
	OutputPath  string `json:"output_path"`
	VoiceID     string `json:"voice_id"`
	TTSModel    string `json:"tts_model"`
	AudioFormat string `json:"audio_format"`
}

// VoiceResult 语音合成结果
type VoiceResult struct {
	OutputPath string `json:"output_path"`
}

// MusicOptions 音乐生成选项
type MusicOptions struct {
	Prompt      string `json:"prompt"`
	OutputPath  string `json:"output_path"`
	Model       string `json:"model"`
	AudioFormat string `json:"audio_format"`
}

// MusicResult 音乐生成结果
type MusicResult struct {
	OutputPath string `json:"output_path"`
}

// StitchOptions 素材合成选项
type StitchOptions struct {
	VideoPaths    []string `json:"video_paths"`
	NarrationPath string   `json:"narration_path"`
	OutputPath    string   `json:"output_path"`
	MusicPath     string   `json:"music_path,omitempty"`
}

// StitchResult 素材合成结果
type StitchResult struct {
	StitchedVideoPath string `json:"stitched_video_path"`
	PaddedVideoPath   string `json:"padded_video_path"`
	FinalVideoPath    string `json:"final_video_path"`
}

// QuotaInfo 额度信息
type QuotaInfo struct {
	ModelName                 string    `json:"model_name"`
	StartTime                 time.Time `json:"start_time"`
	EndTime                   time.Time `json:"end_time"`
	RemainsTimeMs             int64     `json:"remains_time_ms"`
	CurrentIntervalTotalCount int       `json:"current_interval_total_count"`
	CurrentIntervalUsageCount int       `json:"current_interval_usage_count"`
	WeeklyStartTime           time.Time `json:"weekly_start_time"`
	WeeklyEndTime             time.Time `json:"weekly_end_time"`
	CurrentWeeklyTotalCount   int       `json:"current_weekly_total_count"`
	CurrentWeeklyUsageCount   int       `json:"current_weekly_usage_count"`
	WeeklyRemainsTimeMs       int64     `json:"weekly_remains_time_ms"`
}

// QuotaResult 额度查询结果
type QuotaResult struct {
	Entries []QuotaInfo `json:"entries"`
}

// JobEvent 任务日志事件
type JobEvent struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

// JobArtifact 任务产物
type JobArtifact struct {
	Label string `json:"label"`
	Kind  string `json:"kind"`
	Path  string `json:"path"`
}

// Job 任务状态
type Job struct {
	JobID     string        `json:"job_id"`
	Status    string        `json:"status"` // pending, processing, completed, failed
	Progress  float64       `json:"progress"`
	Stage     string        `json:"stage"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Request   interface{}   `json:"request,omitempty"`
	Output    interface{}   `json:"output,omitempty"`
	Error     string        `json:"error,omitempty"`
	Logs      []JobEvent    `json:"logs,omitempty"`
	Artifacts []JobArtifact `json:"artifacts,omitempty"`
}
