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

func (s *Server) HandleClip(c *gin.Context) {
	var req ClipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := s.outputJobDir(jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "clip", req)

	go func() {
		ctx := context.Background()
		opts := schemas.ClipOptions{
			ImagePrompt:  req.Prompt,
			VideoPrompt:  req.Subject,
			AspectRatio: defaultStr(req.AspectRatio, "16:9"),
			VideoModel:   defaultStr(req.Model, "MiniMax-Hailuo-2.3-Fast"),
			Duration:     defaultInt(req.Duration, 5),
			Resolution:   defaultStr(req.Resolution, "720p"),
			PollInterval: 3,
			MaxWait:      300,
			OutputPrefix: joinPath(jobDir, "clip"),
		}

		wf := workflows.NewClipWorkflow(s.client)
		result, err := wf.Run(ctx, opts, func(stage string) {
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
