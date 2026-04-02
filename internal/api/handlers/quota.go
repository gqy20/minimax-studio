package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandleQuota(c *gin.Context) {
	ctx := context.Background()
	result, err := s.client.GetQuota(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
