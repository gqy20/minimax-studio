package handlers

// --- Request types ---

type ImageRequest struct {
	Prompt      string `form:"prompt" binding:"required"`
	AspectRatio string `form:"aspect_ratio"`
}

type ClipRequest struct {
	Prompt      string `json:"prompt" binding:"required"`
	Subject     string `json:"subject"`
	AspectRatio string `json:"aspect_ratio"`
	Model       string `json:"model"`
	Duration    int    `json:"duration"`
	Resolution  string `json:"resolution"`
}

type PlanRequest struct {
	Theme         string `json:"theme" binding:"required"`
	SceneCount    int    `json:"scene_count"`
	SceneDuration int    `json:"scene_duration"`
	Language      string `json:"language"`
}

type VoiceRequest struct {
	Text        string `json:"text" binding:"required"`
	VoiceID     string `json:"voice_id"`
	Model       string `json:"model"`
	AudioFormat string `json:"audio_format"`
}

type MusicRequest struct {
	Prompt      string `json:"prompt" binding:"required"`
	Model       string `json:"model"`
	AudioFormat string `json:"audio_format"`
}

type StitchRequest struct {
	Videos    []string `json:"videos" binding:"required"`
	Narration string   `json:"narration" binding:"required"`
	Music     string   `json:"music"`
}

type MakeRequest struct {
	Theme         string `json:"theme" binding:"required"`
	SceneCount    int    `json:"scene_count"`
	SceneDuration int    `json:"scene_duration"`
	Language      string `json:"language"`
	InputVideo    string `json:"input_video"`
}
