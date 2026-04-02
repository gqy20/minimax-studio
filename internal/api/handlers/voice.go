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

func (s *Server) HandleVoice(c *gin.Context) {
	var req VoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := s.outputJobDir(jobID)
	os.MkdirAll(jobDir, 0755)

	s.createJob(jobID, "voice", req)

	go func() {
		ctx := context.Background()
		format := defaultStr(req.AudioFormat, "mp3")
		opts := schemas.VoiceOptions{
			Text:        req.Text,
			OutputPath:  joinPath(jobDir, "voice."+format),
			VoiceID:     defaultStr(req.VoiceID, "male-qn-qingse"),
			TTSModel:    defaultStr(req.Model, "speech-2.8-hd"),
			AudioFormat: format,
		}

		wf := workflows.NewVoiceWorkflow(s.client)
		result, err := wf.Run(ctx, opts, func(stage string) {
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
