package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
)

func (s *Server) ListJobs(c *gin.Context) {
	s.jobsMu.RLock()
	jobs := make([]*schemas.Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}
	s.jobsMu.RUnlock()

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].UpdatedAt.After(jobs[j].UpdatedAt)
	})

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

func (s *Server) GetJob(c *gin.Context) {
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

func (s *Server) ServeOutput(c *gin.Context) {
	path := c.Param("path")
	fullPath := filepath.Join(s.outputDir, path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.File(fullPath)
}
