package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
	"github.com/minimax-ai/minimax-studio/internal/workflows"
)

// Server HTTP API Server
type Server struct {
	engine    *gin.Engine
	jobs      map[string]*schemas.Job
	jobsMu    sync.RWMutex
	outputDir string
	client    *client.MiniMaxClient
}

// NewServer 创建新的 Server
func NewServer(outputDir, apiKey string) *Server {
	if outputDir == "" {
		outputDir = "./output"
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())

	s := &Server{
		engine:    engine,
		jobs:      make(map[string]*schemas.Job),
		outputDir: outputDir,
		client:    client.NewClient(apiKey),
	}

	s.loadJobsFromDisk()
	s.setupRoutes()
	return s
}

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

func (s *Server) setupRoutes() {
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := s.engine.Group("/api/v1")
	{
		v1.GET("/jobs", s.listJobs)
		v1.GET("/jobs/:id", s.getJob)
		v1.POST("/clip", s.handleClip)
		v1.POST("/plan", s.handlePlan)
		v1.POST("/voice", s.handleVoice)
		v1.POST("/music", s.handleMusic)
		v1.POST("/stitch", s.handleStitch)
		v1.POST("/make", s.handleMake)
		v1.GET("/quota", s.handleQuota)
		v1.GET("/output/*path", s.serveOutput)
	}
}

var stepPattern = regexp.MustCompile(`step\s+(\d+)/(\d+)`)

func (s *Server) createJob(jobID, stage string, request interface{}) *schemas.Job {
	now := time.Now().UTC()
	job := &schemas.Job{
		JobID:     jobID,
		Status:    "processing",
		Stage:     stage,
		Progress:  0,
		CreatedAt: now,
		UpdatedAt: now,
		Request:   request,
	}
	job.Logs = []schemas.JobEvent{{
		Time:    now,
		Message: fmt.Sprintf("job created for %s", stage),
	}}

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

	job.Logs = append(job.Logs, schemas.JobEvent{
		Time:    now,
		Message: message,
	})
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
		job.Logs = append(job.Logs, schemas.JobEvent{
			Time:    now,
			Message: "error: " + errStr,
		})
	} else {
		job.Error = ""
		job.Logs = append(job.Logs, schemas.JobEvent{
			Time:    now,
			Message: fmt.Sprintf("job %s", status),
		})
	}
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

	progress := float64(current-1) / float64(total)
	if progress < 0 {
		progress = 0
	}
	return progress, true
}

func collectArtifacts(output interface{}) []schemas.JobArtifact {
	var artifacts []schemas.JobArtifact

	add := func(label, kind, path string) {
		if path == "" {
			return
		}
		artifacts = append(artifacts, schemas.JobArtifact{
			Label: label,
			Kind:  kind,
			Path:  path,
		})
	}

	switch result := output.(type) {
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

func (s *Server) jobFilePath(jobID string) string {
	return filepath.Join(s.outputDir, jobID, "job.json")
}

func (s *Server) persistJobLocked(job *schemas.Job) {
	if job == nil {
		return
	}

	jobPath := s.jobFilePath(job.JobID)
	if err := os.MkdirAll(filepath.Dir(jobPath), 0755); err != nil {
		log.Printf("failed to create job dir for %s: %v", job.JobID, err)
		return
	}

	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		log.Printf("failed to marshal job %s: %v", job.JobID, err)
		return
	}

	if err := os.WriteFile(jobPath, data, 0644); err != nil {
		log.Printf("failed to persist job %s: %v", job.JobID, err)
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

		jobPath := filepath.Join(s.outputDir, entry.Name(), "job.json")
		data, err := os.ReadFile(jobPath)
		if err != nil {
			continue
		}

		var job schemas.Job
		if err := json.Unmarshal(data, &job); err != nil {
			log.Printf("failed to load job metadata %s: %v", jobPath, err)
			continue
		}

		s.jobs[job.JobID] = &job
	}
}

// --- Clip ---

type ClipRequest struct {
	Prompt      string `json:"prompt" binding:"required"`
	Subject     string `json:"subject" binding:"required"`
	AspectRatio string `json:"aspect_ratio"`
	Model       string `json:"model"`
	Duration    int    `json:"duration"`
	Resolution  string `json:"resolution"`
}

func (s *Server) handleClip(c *gin.Context) {
	var req ClipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := filepath.Join(s.outputDir, jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "clip", req)

	go func() {
		ctx := context.Background()
		opts := schemas.ClipOptions{
			ImagePrompt:  req.Prompt,
			VideoPrompt:  req.Subject,
			AspectRatio:  defaultStr(req.AspectRatio, "16:9"),
			VideoModel:   defaultStr(req.Model, "MiniMax-Hailuo-2.3-Fast"),
			Duration:     defaultInt(req.Duration, 5),
			Resolution:   defaultStr(req.Resolution, "720p"),
			PollInterval: 3,
			MaxWait:      300,
			OutputPrefix: filepath.Join(jobDir, "clip"),
		}

		wf := workflows.NewClipWorkflow(s.client)
		result, err := wf.Run(ctx, opts, func(stage string) {
			log.Printf("[job:%s] %s", jobID, stage)
			s.appendJobLog(jobID, stage)
		})

		if err != nil {
			s.updateJob(jobID, "failed", "clip", 0, nil, err.Error())
			return
		}

		s.updateJob(jobID, "completed", "clip", 1.0, result, "")
	}()

	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID, "status": "processing"})
}

// --- Plan ---

type PlanRequest struct {
	Theme         string `json:"theme" binding:"required"`
	SceneCount    int    `json:"scene_count"`
	SceneDuration int    `json:"scene_duration"`
	Language      string `json:"language"`
}

func (s *Server) handlePlan(c *gin.Context) {
	var req PlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := filepath.Join(s.outputDir, jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "plan", req)

	go func() {
		ctx := context.Background()
		opts := schemas.PlanOptions{
			Theme:         req.Theme,
			SceneCount:    defaultInt(req.SceneCount, 1),
			SceneDuration: defaultInt(req.SceneDuration, 6),
			Language:      defaultStr(req.Language, "zh"),
			TextModel:     "MiniMax-M2.7-highspeed",
			TextMaxTokens: 4096,
			OutputDir:     jobDir,
		}

		wf := workflows.NewPlanWorkflow(s.client)
		result, err := wf.Run(ctx, opts, func(stage string) {
			log.Printf("[job:%s] %s", jobID, stage)
			s.appendJobLog(jobID, stage)
		})

		if err != nil {
			s.updateJob(jobID, "failed", "plan", 0, nil, err.Error())
			return
		}

		s.updateJob(jobID, "completed", "plan", 1.0, result, "")
	}()

	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID, "status": "processing"})
}

// --- Voice ---

type VoiceRequest struct {
	Text        string `json:"text" binding:"required"`
	VoiceID     string `json:"voice_id"`
	Model       string `json:"model"`
	AudioFormat string `json:"audio_format"`
}

func (s *Server) handleVoice(c *gin.Context) {
	var req VoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := filepath.Join(s.outputDir, jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "voice", req)

	go func() {
		ctx := context.Background()
		format := defaultStr(req.AudioFormat, "mp3")
		opts := schemas.VoiceOptions{
			Text:        req.Text,
			OutputPath:  filepath.Join(jobDir, "voice."+format),
			VoiceID:     defaultStr(req.VoiceID, "male-qn-qingse"),
			TTSModel:    defaultStr(req.Model, "speech-2.8-hd"),
			AudioFormat: format,
		}

		wf := workflows.NewVoiceWorkflow(s.client)
		result, err := wf.Run(ctx, opts, func(stage string) {
			log.Printf("[job:%s] %s", jobID, stage)
			s.appendJobLog(jobID, stage)
		})

		if err != nil {
			s.updateJob(jobID, "failed", "voice", 0, nil, err.Error())
			return
		}

		s.updateJob(jobID, "completed", "voice", 1.0, result, "")
	}()

	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID, "status": "processing"})
}

// --- Music ---

type MusicRequest struct {
	Prompt      string `json:"prompt" binding:"required"`
	Model       string `json:"model"`
	AudioFormat string `json:"audio_format"`
}

func (s *Server) handleMusic(c *gin.Context) {
	var req MusicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := filepath.Join(s.outputDir, jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "music", req)

	go func() {
		ctx := context.Background()
		format := defaultStr(req.AudioFormat, "mp3")
		opts := schemas.MusicOptions{
			Prompt:      req.Prompt,
			OutputPath:  filepath.Join(jobDir, "music."+format),
			Model:       defaultStr(req.Model, "music-2.5"),
			AudioFormat: format,
		}

		wf := workflows.NewMusicWorkflow(s.client)
		result, err := wf.Run(ctx, opts, func(stage string) {
			log.Printf("[job:%s] %s", jobID, stage)
			s.appendJobLog(jobID, stage)
		})

		if err != nil {
			s.updateJob(jobID, "failed", "music", 0, nil, err.Error())
			return
		}

		s.updateJob(jobID, "completed", "music", 1.0, result, "")
	}()

	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID, "status": "processing"})
}

// --- Stitch ---

type StitchRequest struct {
	Videos    []string `json:"videos" binding:"required"`
	Narration string   `json:"narration" binding:"required"`
	Music     string   `json:"music"`
}

func (s *Server) handleStitch(c *gin.Context) {
	var req StitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := filepath.Join(s.outputDir, jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "stitch", req)

	go func() {
		opts := schemas.StitchOptions{
			VideoPaths:    req.Videos,
			NarrationPath: req.Narration,
			MusicPath:     req.Music,
			OutputPath:    filepath.Join(jobDir, "final.mp4"),
		}

		wf := workflows.NewStitchWorkflow()
		result, err := wf.Run(opts, func(stage string) {
			log.Printf("[job:%s] %s", jobID, stage)
			s.appendJobLog(jobID, stage)
		})

		if err != nil {
			s.updateJob(jobID, "failed", "stitch", 0, nil, err.Error())
			return
		}

		s.updateJob(jobID, "completed", "stitch", 1.0, result, "")
	}()

	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID, "status": "processing"})
}

// --- Make ---

type MakeRequest struct {
	Theme         string `json:"theme" binding:"required"`
	SceneCount    int    `json:"scene_count"`
	SceneDuration int    `json:"scene_duration"`
	Language      string `json:"language"`
	InputVideo    string `json:"input_video"`
}

func (s *Server) handleMake(c *gin.Context) {
	var req MakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := filepath.Join(s.outputDir, jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "make", req)

	go func() {
		ctx := context.Background()
		opts := workflows.MakeOptions{
			Theme:         req.Theme,
			SceneCount:    defaultInt(req.SceneCount, 1),
			SceneDuration: defaultInt(req.SceneDuration, 6),
			Language:      defaultStr(req.Language, "zh"),
			AspectRatio:   "16:9",
			Resolution:    "720p",
			TextModel:     "MiniMax-M2.7-highspeed",
			TextMaxTokens: 4096,
			VideoModel:    "MiniMax-Hailuo-2.3-Fast",
			TTSModel:      "speech-2.8-hd",
			MusicModel:    "music-2.5",
			MusicMode:     "optional",
			VoiceID:       "male-qn-qingse",
			AudioFormat:   "mp3",
			PollInterval:  3,
			MaxWait:       300,
			OutputDir:     jobDir,
			InputVideo:    req.InputVideo,
		}

		wf := workflows.NewMakeWorkflow(s.client)
		result, err := wf.Run(ctx, opts, func(stage string) {
			log.Printf("[job:%s] %s", jobID, stage)
			s.appendJobLog(jobID, stage)
		})

		if err != nil {
			s.updateJob(jobID, "failed", "make", 0, nil, err.Error())
			return
		}

		s.updateJob(jobID, "completed", "make", 1.0, result, "")
	}()

	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID, "status": "processing"})
}

// --- Quota ---

func (s *Server) handleQuota(c *gin.Context) {
	ctx := context.Background()
	result, err := s.client.GetQuota(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// --- Jobs ---

func (s *Server) listJobs(c *gin.Context) {
	s.jobsMu.RLock()
	jobs := make([]*schemas.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	s.jobsMu.RUnlock()

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].UpdatedAt.After(jobs[j].UpdatedAt)
	})

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

func (s *Server) getJob(c *gin.Context) {
	jobID := c.Param("id")

	s.jobsMu.RLock()
	job, exists := s.jobs[jobID]
	s.jobsMu.RUnlock()

	if !exists {
		s.loadJobsFromDisk()
		s.jobsMu.RLock()
		job, exists = s.jobs[jobID]
		s.jobsMu.RUnlock()
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
	}

	c.JSON(http.StatusOK, job)
}

// --- Output files ---

func (s *Server) serveOutput(c *gin.Context) {
	path := c.Param("path")
	fullPath := filepath.Join(s.outputDir, path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.File(fullPath)
}

// --- Helpers ---

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

// Run 启动 HTTP 服务器
func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}
