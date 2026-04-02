package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
)

func (s *Server) HandleImage(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := s.outputJobDir(jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "image", req)

	go func() {
		ctx := context.Background()
		s.appendJobLog(jobID, "generating image...")

		imageData, _, err := s.client.GenerateImage(ctx, req.Prompt, defaultStr(req.AspectRatio, "16:9"), false)
		if err != nil {
			s.updateJob(jobID, "failed", "image", 0, nil, err.Error())
			return
		}

		outputPath := joinPath(jobDir, "image.jpg")
		if err := os.WriteFile(outputPath, imageData, 0644); err != nil {
			s.updateJob(jobID, "failed", "image", 0, nil, err.Error())
			return
		}

		s.appendJobLog(jobID, "image saved to: "+outputPath)
		s.updateJob(jobID, "completed", "image", 1.0, &schemas.ImageResult{ImagePath: outputPath}, "")
	}()

	c.JSON(http.StatusAccepted, gin.H{"job_id": jobID, "status": "processing"})
}
