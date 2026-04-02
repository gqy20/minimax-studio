package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minimax-ai/minimax-studio/internal/workflows"
)

func (s *Server) HandleMake(c *gin.Context) {
	var req MakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobID := uuid.New().String()
	jobDir := s.outputJobDir(jobID)
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
