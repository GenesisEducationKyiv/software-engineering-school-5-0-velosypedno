package handlers

import (
	"errors"
	"net/http"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type subscriptionDeactivator interface {
	Unsubscribe(token uuid.UUID) error
}

func NewUnsubscribeGETHandler(logger *zap.Logger, service subscriptionDeactivator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")

		parsedToken, err := uuid.Parse(token)
		if err != nil {
			logger.Warn("Invalid token format",
				zap.String("token", token),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			return
		}

		err = service.Unsubscribe(parsedToken)

		switch {
		case errors.Is(err, domain.ErrSubNotFound):
			logger.Info("Unsubscribe attempt for non-existent token",
				zap.String("token", parsedToken.String()),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return

		case errors.Is(err, domain.ErrInternal):
			logger.Error("Internal error during unsubscribe",
				zap.String("token", parsedToken.String()),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe"})
			return

		case err != nil:
			logger.Error("Unexpected error during unsubscribe",
				zap.String("token", parsedToken.String()),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe"})
			return
		}

		logger.Info("Successfully unsubscribed",
			zap.String("token", parsedToken.String()),
		)
		c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
	}
}
