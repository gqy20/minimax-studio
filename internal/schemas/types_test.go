package schemas

import (
	"encoding/json"
	"testing"
)

func TestVideoPlan_JSON(t *testing.T) {
	plan := VideoPlan{
		Title:       "纸船",
		VisualStyle: "cinematic",
		Narration:   "一艘纸船缓缓前行",
		MusicPrompt: "gentle piano",
		Scenes: []ScenePlan{
			{Name: "scene1", ImagePrompt: "a paper boat", VideoPrompt: "drifts slowly"},
		},
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded VideoPlan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Title != "纸船" {
		t.Errorf("expected title '纸船', got '%s'", decoded.Title)
	}
	if len(decoded.Scenes) != 1 {
		t.Errorf("expected 1 scene, got %d", len(decoded.Scenes))
	}
	if decoded.Scenes[0].ImagePrompt != "a paper boat" {
		t.Errorf("expected image_prompt 'a paper boat', got '%s'", decoded.Scenes[0].ImagePrompt)
	}
}

func TestJob_JSON(t *testing.T) {
	job := Job{
		JobID:    "abc123",
		Status:   "processing",
		Progress: 0.5,
		Stage:    "clip",
		Output:   map[string]string{"url": "http://example.com"},
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Job
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.JobID != "abc123" {
		t.Errorf("expected job_id 'abc123', got '%s'", decoded.JobID)
	}
	if decoded.Status != "processing" {
		t.Errorf("expected status 'processing', got '%s'", decoded.Status)
	}
	if decoded.Progress != 0.5 {
		t.Errorf("expected progress 0.5, got %f", decoded.Progress)
	}
}

func TestJob_EmptyOmit(t *testing.T) {
	job := Job{JobID: "x", Status: "pending", Progress: 0, Stage: ""}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// output and error should be omitted when nil/empty
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, exists := raw["output"]; exists {
		t.Error("expected 'output' to be omitted when nil")
	}
	if _, exists := raw["error"]; exists {
		t.Error("expected 'error' to be omitted when empty")
	}
}

func TestClipOptions_Defaults(t *testing.T) {
	opts := ClipOptions{
		ImagePrompt:  "test",
		VideoPrompt:  "subject",
		AspectRatio:  "16:9",
		VideoModel:   "MiniMax-Hailuo-2.3-Fast",
		Duration:     5,
		Resolution:   "720p",
		PollInterval: 3,
		MaxWait:      300,
		OutputPrefix: "/tmp/test",
	}

	if opts.AspectRatio != "16:9" {
		t.Errorf("expected 16:9, got %s", opts.AspectRatio)
	}
	if opts.Duration != 5 {
		t.Errorf("expected 5, got %d", opts.Duration)
	}
}

func TestStitchOptions_MusicOmit(t *testing.T) {
	opts := StitchOptions{
		VideoPaths:    []string{"/tmp/a.mp4"},
		NarrationPath: "/tmp/voice.mp3",
		OutputPath:    "/tmp/final.mp4",
		// MusicPath is empty
	}

	data, _ := json.Marshal(opts)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, exists := raw["music_path"]; exists {
		t.Error("expected 'music_path' to be omitted when empty")
	}
}

func TestStitchOptions_MusicPresent(t *testing.T) {
	opts := StitchOptions{
		VideoPaths:    []string{"/tmp/a.mp4"},
		NarrationPath: "/tmp/voice.mp3",
		OutputPath:    "/tmp/final.mp4",
		MusicPath:     "/tmp/music.mp3",
	}

	data, _ := json.Marshal(opts)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, exists := raw["music_path"]; !exists {
		t.Error("expected 'music_path' to be present when set")
	}
}
