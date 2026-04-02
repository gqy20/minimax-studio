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

func (s *Server) HandlePlan(c *gin.Context) {
	var req PlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := s.outputJobDir(jobID)
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
