package client

import (
	"testing"
)

func TestParsePlanJSON_Valid(t *testing.T) {
	input := `{"title":"纸船","visual_style":"cinematic","narration":"一艘纸船缓缓前行","music_prompt":"gentle piano","scenes":[{"name":"scene1","image_prompt":"a paper boat","video_prompt":"drifts slowly"}]}`

	plan, err := parsePlanJSON(input, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Title != "纸船" {
		t.Errorf("expected title '纸船', got '%s'", plan.Title)
	}
	if len(plan.Scenes) != 1 {
		t.Errorf("expected 1 scene, got %d", len(plan.Scenes))
	}
	if plan.Scenes[0].ImagePrompt != "a paper boat" {
		t.Errorf("expected image_prompt 'a paper boat', got '%s'", plan.Scenes[0].ImagePrompt)
	}
}

func TestParsePlanJSON_WithThinking(t *testing.T) {
	input := `<thinking>Let me think about this</thinking>{"title":"纸船","visual_style":"cinematic","narration":"test","music_prompt":"piano","scenes":[{"name":"s1","image_prompt":"boat","video_prompt":"drift"}]}`

	plan, err := parsePlanJSON(input, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Title != "纸船" {
		t.Errorf("expected title '纸船', got '%s'", plan.Title)
	}
}

func TestParsePlanJSON_WithCodeFence(t *testing.T) {
	input := "```json\n{\"title\":\"test\",\"visual_style\":\"v\",\"narration\":\"n\",\"music_prompt\":\"m\",\"scenes\":[{\"name\":\"s1\",\"image_prompt\":\"i\",\"video_prompt\":\"v\"}]}\n```"

	plan, err := parsePlanJSON(input, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Title != "test" {
		t.Errorf("expected title 'test', got '%s'", plan.Title)
	}
}

func TestParsePlanJSON_WrongSceneCount(t *testing.T) {
	input := `{"title":"t","visual_style":"v","narration":"n","music_prompt":"m","scenes":[{"name":"s1","image_prompt":"i","video_prompt":"v"}]}`

	_, err := parsePlanJSON(input, 3)
	if err == nil {
		t.Fatal("expected error for wrong scene count")
	}
}

func TestParsePlanJSON_EmbeddedInText(t *testing.T) {
	input := `Here is the plan:
{"title":"embedded","visual_style":"v","narration":"n","music_prompt":"m","scenes":[{"name":"s1","image_prompt":"i","video_prompt":"v"}]}
Hope this helps!`

	plan, err := parsePlanJSON(input, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Title != "embedded" {
		t.Errorf("expected title 'embedded', got '%s'", plan.Title)
	}
}

func TestParsePlanJSON_NoJSON(t *testing.T) {
	input := "This is just plain text without any JSON."

	_, err := parsePlanJSON(input, 1)
	if err == nil {
		t.Fatal("expected error for no JSON")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input   string
		maxLen  int
		wantLen int
		suffix  string
	}{
		{"short", 10, 5, ""},
		{"this is a long string", 10, 13, "..."},
		{"exact", 5, 5, ""},
	}
	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if len(result) != tt.wantLen {
			t.Errorf("truncate(%q, %d): got len %d, want %d", tt.input, tt.maxLen, len(result), tt.wantLen)
		}
		if tt.suffix != "" && result[len(result)-3:] != tt.suffix {
			t.Errorf("truncate(%q, %d): got %q, want suffix %q", tt.input, tt.maxLen, result, tt.suffix)
		}
	}
}
