package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
	"github.com/minimax-ai/minimax-studio/internal/workflows"
)

// H is a shorthand for gin.H
type H = gin.H

/* ── Server (thin orchestrator) ── */

type Server struct {
	engine      *gin.Engine
	jobs        map[string]*schemas.Job
	jobsMu      sync.RWMutex
	outputDir   string
	client      *client.MiniMaxClient
	frontendDir string
}

func NewServer(outputDir, apiKey, frontendDir string) *Server {
	if outputDir == "" {
		outputDir = "./output"
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())

	s := &Server{
		engine:      engine,
		jobs:        make(map[string]*schemas.Job),
		outputDir:   outputDir,
		client:      client.NewClient(apiKey),
		frontendDir: frontendDir,
	}

	s.loadJobsFromDisk()
	s.setupRoutes()
	registerFrontend(engine, s.frontendDir)
	return s
}

func (s *Server) Run(addr string) error { return s.engine.Run(addr) }

/* ── Routing ── */

func (s *Server) setupRoutes() {
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, H{"status": "ok"})
	})

	v1 := s.engine.Group("/api/v1")
	{
		v1.GET("/jobs", s.ListJobs)
		v1.GET("/jobs/:id", s.GetJob)
		v1.POST("/image", s.HandleImage)
		v1.POST("/clip", s.HandleClip)
		v1.POST("/plan", s.HandlePlan)
		v1.POST("/voice", s.HandleVoice)
		v1.POST("/music", s.HandleMusic)
		v1.POST("/stitch", s.HandleStitch)
		v1.POST("/make", s.HandleMake)
		v1.GET("/quota", s.HandleQuota)
		v1.GET("/output/*path", s.ServeOutput)
	}
}

/* ── Frontend serving ── */

func registerFrontend(engine *gin.Engine, frontendDir string) {
	if frontendDir == "" {
		engine.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, H{"error": "frontend not built — run `make frontend-build` and rebuild"})
		})
		return
	}

	entries, err := os.ReadDir(frontendDir)
	if err != nil || len(entries) == 0 {
		engine.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, H{"error": "frontend not built — run `make frontend-build` and rebuild"})
		})
		return
	}

	fileServer := http.FileServer(http.Dir(frontendDir))

	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasSuffix(path, "/") {
			ext := filepath.Ext(path)
			if ext != "" {
				cleanPath := strings.TrimPrefix(path, "/")
				fullPath := filepath.Join(frontendDir, cleanPath)
				if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
					fileServer.ServeHTTP(c.Writer, c.Request)
					return
				}
			}
		}
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}

/* ── CORS ── */

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

/* ── Job CRUD helpers (used by all handlers) ── */

var stepPattern = regexp.MustCompile(`step\s+(\d+)/(\d+)`)

func (s *Server) createJob(jobID, stage string, request interface{}) *schemas.Job {
	now := time.Now().UTC()
	job := &schemas.Job{
		JobID: jobID, Status: "processing", Stage: stage,
		Progress: 0, CreatedAt: now, UpdatedAt: now, Request: request,
	}
	job.Logs = []schemas.JobEvent{{Time: now, Message: fmt.Sprintf("job created for %s", stage)}}

	s.jobsMu.Lock()
	s.jobs[jobID] = job
	s.persistJobLocked(job)
	s.jobsMu.Unlock()

	return job
}

func (s *Server) appendJobLog(jobID, message string) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return
	}

	now := time.Now().UTC()
	job.Stage = message
	job.UpdatedAt = now
	if progress, ok := parseProgress(message); ok && progress > job.Progress {
		job.Progress = progress
	}

	job.Logs = append(job.Logs, schemas.JobEvent{Time: now, Message: message})
	if len(job.Logs) > 50 {
		job.Logs = append([]schemas.JobEvent(nil), job.Logs[len(job.Logs)-50:]...)
	}

	s.persistJobLocked(job)
}

func (s *Server) updateJob(jobID, status, stage string, progress float64, output interface{}, errStr string) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return
	}

	now := time.Now().UTC()
	job.Status = status
	job.Stage = stage
	job.Progress = progress
	job.Output = output
	job.UpdatedAt = now
	job.Artifacts = collectArtifacts(output)
	if errStr != "" {
		job.Error = errStr
	} else {
		job.Error = ""
	}

	job.Logs = append(job.Logs, schemas.JobEvent{
		Time: now, Message: fmt.Sprintf("job %s", status),
	})
	if len(job.Logs) > 50 {
		job.Logs = append([]schemas.JobEvent(nil), job.Logs[len(job.Logs)-50:]...)
	}

	s.persistJobLocked(job)
}

func parseProgress(message string) (float64, bool) {
	matches := stepPattern.FindStringSubmatch(strings.ToLower(message))
	if len(matches) != 3 {
		return 0, false
	}
	var current, total int
	if _, err := fmt.Sscanf(matches[0], "step %d/%d", &current, &total); err != nil || total <= 0 {
		return 0, false
	}
	return float64(current-1) / float64(total), true
}

func collectArtifacts(output interface{}) []schemas.JobArtifact {
	var artifacts []schemas.JobArtifact
	add := func(label, kind, path string) {
		if path != "" {
			artifacts = append(artifacts, schemas.JobArtifact{Label: label, Kind: kind, Path: path})
		}
	}

	switch result := output.(type) {
	case *schemas.ImageResult:
		add("image", "image", result.ImagePath)
	case *schemas.ClipResult:
		add("image", "image", result.ImagePath)
		add("video", "video", result.VideoPath)
	case *schemas.PlanResult:
		add("plan", "json", result.PlanPath)
		add("narration_text", "text", result.NarrationPath)
	case *schemas.VoiceResult:
		add("voice", "audio", result.OutputPath)
	case *schemas.MusicResult:
		add("music", "audio", result.OutputPath)
	case *schemas.StitchResult:
		add("stitched_video", "video", result.StitchedVideoPath)
		add("timed_video", "video", result.PaddedVideoPath)
		add("final_video", "video", result.FinalVideoPath)
	case *workflows.MakeResult:
		add("plan", "json", result.PlanPath)
		add("narration", "audio", result.NarrationPath)
		add("music", "audio", result.MusicPath)
		add("final_video", "video", result.FinalVideoPath)
	}
	return artifacts
}

/* ── Persistence ── */

func (s *Server) jobFilePath(jobID string) string {
	return filepath.Join(s.outputDir, jobID, "job.json")
}

func (s *Server) persistJobLocked(job *schemas.Job) {
	if job == nil {
		return
	}
	p := s.jobFilePath(job.JobID)
	os.MkdirAll(filepath.Dir(p), 0755)
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		log.Printf("failed to marshal job %s: %v", job.JobID, err)
		return
	}
	if err := os.WriteFile(p, data, 0644); err != nil {
		log.Printf("failed to write job %s: %v", job.JobID, err)
	}
}

func (s *Server) loadJobsFromDisk() {
	entries, err := os.ReadDir(s.outputDir)
	if err != nil {
		return
	}

	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		p := filepath.Join(s.outputDir, entry.Name(), "job.json")
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var job schemas.Job
		if err := json.Unmarshal(data, &job); err != nil {
			continue
		}
		s.jobs[job.JobID] = &job
	}
}

/* ── Path helpers ── */

func (s *Server) outputJobDir(jobID string) string { return filepath.Join(s.outputDir, jobID) }

func joinPath(dir, file string) string { return filepath.Join(dir, file) }

func defaultStr(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func defaultInt(val, def int) int {
	if val == 0 {
		return def
	}
	return val
}
