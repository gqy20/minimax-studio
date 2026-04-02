package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
	"github.com/minimax-ai/minimax-studio/internal/workflows"
)

func (s *Server) HandleStitch(c *gin.Context) {
	var req StitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := s.outputJobDir(jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "stitch", req)

	go func() {
		opts := schemas.StitchOptions{
			VideoPaths:    req.Videos,
			NarrationPath: req.Narration,
			MusicPath:     req.Music,
			OutputPath:    joinPath(jobDir, "final.mp4"),
		}

		wf := workflows.NewStitchWorkflow()
		result, err := wf.Run(opts, func(stage string) {
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
