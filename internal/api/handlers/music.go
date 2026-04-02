package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
	"github.com/minimax-ai/minimax-studio/internal/workflows"
)

func (s *Server) HandleMusic(c *gin.Context) {
	var req MusicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := s.outputJobDir(jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "music", req)

	go func() {
		ctx := context.Background()
		format := defaultStr(req.AudioFormat, "mp3")
		opts := schemas.MusicOptions{
			Prompt:      req.Prompt,
			OutputPath:  joinPath(jobDir, "music."+format),
			Model:       defaultStr(req.Model, "music-2.5"),
			AudioFormat: format,
		}

		wf := workflows.NewMusicWorkflow(s.client)
		result, err := wf.Run(ctx, opts, func(stage string) {
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
