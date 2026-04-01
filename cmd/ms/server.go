package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
	"github.com/spf13/cobra"
)

var (
	serverPort      string
	serverOutputDir string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP API server",
	Long:  `Start the MiniMax Studio HTTP API server for frontend integration.`,
	RunE:  runServer,
}

func init() {
	serverCmd.Flags().StringVar(&serverPort, "port", "8080", "Server port")
	serverCmd.Flags().StringVar(&serverOutputDir, "output-dir", "./output", "Output directory")

	RootCmd.AddCommand(serverCmd)
}

type Server struct {
	engine    *gin.Engine
	jobs      map[string]*schemas.Job
	jobsMu    sync.RWMutex
	outputDir string
}

func newServer(outputDir string) *Server {
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
		v1.POST("/jobs", s.createJob)
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

func (s *Server) createJob(c *gin.Context) {
	var req struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()

	s.jobsMu.Lock()
	s.jobs[jobID] = &schemas.Job{
		JobID:    jobID,
		Status:   "pending",
		Progress: 0,
		Stage:    "created",
	}
	s.jobsMu.Unlock()

	c.JSON(http.StatusAccepted, s.jobs[jobID])
}

func (s *Server) handleClip(c *gin.Context) {
	jobID := uuid.New().String()

	s.jobsMu.Lock()
	s.jobs[jobID] = &schemas.Job{
		JobID:    jobID,
		Status:   "processing",
		Progress: 0.25,
		Stage:    "generating_image",
	}
	s.jobsMu.Unlock()

	c.JSON(http.StatusAccepted, s.jobs[jobID])
}

func (s *Server) handlePlan(c *gin.Context) {
	jobID := uuid.New().String()

	s.jobsMu.Lock()
	s.jobs[jobID] = &schemas.Job{
		JobID:    jobID,
		Status:   "processing",
		Progress: 0.25,
		Stage:    "planning",
	}
	s.jobsMu.Unlock()

	c.JSON(http.StatusAccepted, s.jobs[jobID])
}

func (s *Server) handleVoice(c *gin.Context) {
	jobID := uuid.New().String()

	s.jobsMu.Lock()
	s.jobs[jobID] = &schemas.Job{
		JobID:    jobID,
		Status:   "processing",
		Progress: 0.5,
		Stage:    "synthesizing",
	}
	s.jobsMu.Unlock()

	c.JSON(http.StatusAccepted, s.jobs[jobID])
}

func (s *Server) handleMusic(c *gin.Context) {
	jobID := uuid.New().String()

	s.jobsMu.Lock()
	s.jobs[jobID] = &schemas.Job{
		JobID:    jobID,
		Status:   "processing",
		Progress: 0.5,
		Stage:    "generating_music",
	}
	s.jobsMu.Unlock()

	c.JSON(http.StatusAccepted, s.jobs[jobID])
}

func (s *Server) handleStitch(c *gin.Context) {
	jobID := uuid.New().String()

	s.jobsMu.Lock()
	s.jobs[jobID] = &schemas.Job{
		JobID:    jobID,
		Status:   "processing",
		Progress: 0.75,
		Stage:    "stitching",
	}
	s.jobsMu.Unlock()

	c.JSON(http.StatusAccepted, s.jobs[jobID])
}

func (s *Server) handleMake(c *gin.Context) {
	jobID := uuid.New().String()

	s.jobsMu.Lock()
	s.jobs[jobID] = &schemas.Job{
		JobID:    jobID,
		Status:   "processing",
		Progress: 0,
		Stage:    "making",
	}
	s.jobsMu.Unlock()

	c.JSON(http.StatusAccepted, s.jobs[jobID])
}

func (s *Server) handleQuota(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "quota endpoint not yet implemented"})
}

func (s *Server) serveOutput(c *gin.Context) {
	path := c.Param("path")
	fullPath := filepath.Join(s.outputDir, path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.File(fullPath)
}

func runServer(cmd *cobra.Command, args []string) error {
	s := newServer(serverOutputDir)
	addr := ":" + serverPort
	log.Printf("Starting API server on %s", addr)
	return s.engine.Run(addr)
}
