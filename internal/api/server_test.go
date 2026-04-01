package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key")

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
	s := NewServer(t.TempDir(), "test-key")

	req := httptest.NewRequest("GET", "/api/v1/jobs/nonexistent", nil)
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestClip_MissingPrompt(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key")

	req := httptest.NewRequest("POST", "/api/v1/clip", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestClip_ValidRequest(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key")

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
	s := NewServer(t.TempDir(), "test-key")

	req := httptest.NewRequest("POST", "/api/v1/plan", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPlan_ValidRequest(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key")

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
	s := NewServer(t.TempDir(), "test-key")

	req := httptest.NewRequest("POST", "/api/v1/voice", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVoice_ValidRequest(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key")

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
	s := NewServer(t.TempDir(), "test-key")

	req := httptest.NewRequest("POST", "/api/v1/music", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestMusic_ValidRequest(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key")

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
	s := NewServer(t.TempDir(), "test-key")

	req := httptest.NewRequest("POST", "/api/v1/stitch", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestMake_MissingTheme(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key")

	req := httptest.NewRequest("POST", "/api/v1/make", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestMake_ValidRequest(t *testing.T) {
	s := NewServer(t.TempDir(), "test-key")

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
	s := NewServer(t.TempDir(), "test-key")

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
	s := NewServer(t.TempDir(), "test-key")

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
