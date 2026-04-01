package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

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
func NewServer(outputDir, apiKey, groupID string) *Server {
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
		client:    client.NewClient(apiKey, groupID),
	}

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

func (s *Server) updateJob(jobID, status, stage string, progress float64, output interface{}, errStr string) {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()

	if job, ok := s.jobs[jobID]; ok {
		job.Status = status
		job.Stage = stage
		job.Progress = progress
		job.Output = output
		if errStr != "" {
			job.Error = errStr
		}
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

	job := &schemas.Job{JobID: jobID, Status: "processing", Stage: "clip", Progress: 0}
	s.jobsMu.Lock()
	s.jobs[jobID] = job
	s.jobsMu.Unlock()

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

	job := &schemas.Job{JobID: jobID, Status: "processing", Stage: "plan", Progress: 0}
	s.jobsMu.Lock()
	s.jobs[jobID] = job
	s.jobsMu.Unlock()

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

	job := &schemas.Job{JobID: jobID, Status: "processing", Stage: "voice", Progress: 0}
	s.jobsMu.Lock()
	s.jobs[jobID] = job
	s.jobsMu.Unlock()

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

	job := &schemas.Job{JobID: jobID, Status: "processing", Stage: "music", Progress: 0}
	s.jobsMu.Lock()
	s.jobs[jobID] = job
	s.jobsMu.Unlock()

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

	job := &schemas.Job{JobID: jobID, Status: "processing", Stage: "stitch", Progress: 0}
	s.jobsMu.Lock()
	s.jobs[jobID] = job
	s.jobsMu.Unlock()

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

	job := &schemas.Job{JobID: jobID, Status: "processing", Stage: "make", Progress: 0}
	s.jobsMu.Lock()
	s.jobs[jobID] = job
	s.jobsMu.Unlock()

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

func (s *Server) getJob(c *gin.Context) {
	jobID := c.Param("id")

	s.jobsMu.RLock()
	job, exists := s.jobs[jobID]
	s.jobsMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
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
