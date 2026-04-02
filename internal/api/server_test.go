package api

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/minimax-ai/minimax-studio/internal/schemas"
)

// asyncTestDir creates a temp dir that is cleaned up after a short delay,
// giving background goroutines time to finish writing.
func asyncTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Cleanup(func() { time.Sleep(500 * time.Millisecond) })
	return dir
}

func TestHealthEndpoint(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
}

func TestGetJob_NotFound(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	req := httptest.NewRequest("GET", "/api/v1/jobs/nonexistent", nil)
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestListJobs(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")
	s.createJob("older", "clip", map[string]string{"theme": "a"})
	s.createJob("newer", "make", map[string]string{"theme": "b"})
	s.updateJob("older", "completed", "clip", 1.0, nil, "")
	time.Sleep(10 * time.Millisecond)
	s.updateJob("newer", "processing", "step 1/7: planning storyboard...", 0.1, nil, "")

	req := httptest.NewRequest("GET", "/api/v1/jobs", nil)
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Jobs []schemas.Job `json:"jobs"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode jobs: %v", err)
	}

	if len(resp.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(resp.Jobs))
	}
	if resp.Jobs[0].JobID != "newer" {
		t.Fatalf("expected newest job first, got %s", resp.Jobs[0].JobID)
	}
}

func TestGetJob_LoadsPersistedJob(t *testing.T) {
	outputDir := t.TempDir()
	s := NewServer(outputDir, "test-key", "")
	s.createJob("persisted", "make", map[string]string{"theme": "paper boat"})
	s.appendJobLog("persisted", "step 1/7: planning storyboard...")
	s.updateJob("persisted", "completed", "make", 1.0, nil, "")

	reloaded := NewServer(outputDir, "test-key", "")
	req := httptest.NewRequest("GET", "/api/v1/jobs/persisted", nil)
	w := httptest.NewRecorder()
	reloaded.engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var job schemas.Job
	if err := json.Unmarshal(w.Body.Bytes(), &job); err != nil {
		t.Fatalf("failed to decode job: %v", err)
	}

	if job.JobID != "persisted" {
		t.Fatalf("expected persisted job, got %s", job.JobID)
	}
	if len(job.Logs) == 0 {
		t.Fatal("expected persisted logs")
	}

	jobPath := filepath.Join(outputDir, "persisted", "job.json")
	if _, err := os.Stat(jobPath); err != nil {
		t.Fatalf("expected job.json to exist: %v", err)
	}
}

func TestClip_MissingPrompt(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	req := httptest.NewRequest("POST", "/api/v1/clip", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestClip_ValidRequest(t *testing.T) {
	s := NewServer(asyncTestDir(t), "test-key", "")

	body := `{"prompt":"a paper boat","subject":"drifts slowly"}`
	req := httptest.NewRequest("POST", "/api/v1/clip", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Fatalf("expected 202, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "processing" {
		t.Errorf("expected status processing, got %v", resp["status"])
	}
	if resp["job_id"] == nil || resp["job_id"] == "" {
		t.Error("expected non-empty job_id")
	}
}

func TestPlan_MissingTheme(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	req := httptest.NewRequest("POST", "/api/v1/plan", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPlan_ValidRequest(t *testing.T) {
	s := NewServer(asyncTestDir(t), "test-key", "")

	body := `{"theme":"纸船"}`
	req := httptest.NewRequest("POST", "/api/v1/plan", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Fatalf("expected 202, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["job_id"] == nil || resp["job_id"] == "" {
		t.Error("expected non-empty job_id")
	}
}

func TestVoice_MissingText(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	req := httptest.NewRequest("POST", "/api/v1/voice", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVoice_ValidRequest(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	body := `{"text":"这是一段旁白"}`
	req := httptest.NewRequest("POST", "/api/v1/voice", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Fatalf("expected 202, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestMusic_MissingPrompt(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	req := httptest.NewRequest("POST", "/api/v1/music", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestMusic_ValidRequest(t *testing.T) {
	s := NewServer(asyncTestDir(t), "test-key", "")

	body := `{"prompt":"warm piano"}`
	req := httptest.NewRequest("POST", "/api/v1/music", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Fatalf("expected 202, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestStitch_MissingVideos(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	req := httptest.NewRequest("POST", "/api/v1/stitch", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestMake_MissingTheme(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	req := httptest.NewRequest("POST", "/api/v1/make", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestMake_ValidRequest(t *testing.T) {
	s := NewServer(asyncTestDir(t), "test-key", "")

	body := `{"theme":"纸船","scene_count":1}`
	req := httptest.NewRequest("POST", "/api/v1/make", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Fatalf("expected 202, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestCORSMiddleware(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	// OPTIONS preflight
	req := httptest.NewRequest("OPTIONS", "/api/v1/clip", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("expected 204 for OPTIONS, got %d", w.Code)
	}

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected CORS origin *, got %s", origin)
	}
}

func TestOutput_NotFound(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key", "")

	req := httptest.NewRequest("GET", "/api/v1/output/nonexistent/video.mp4", nil)
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDefaultHelpers(t *testing.T) {
	if defaultStr("", "fallback") != "fallback" {
		t.Error("expected fallback for empty string")
	}
	if defaultStr("value", "fallback") != "value" {
		t.Error("expected value when non-empty")
	}
	if defaultInt(0, 42) != 42 {
		t.Error("expected fallback for zero int")
	}
	if defaultInt(7, 42) != 7 {
		t.Error("expected value when non-zero")
	}
}
