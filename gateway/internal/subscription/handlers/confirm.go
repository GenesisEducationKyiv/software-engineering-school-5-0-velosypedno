package handlers

import (
	"errors"
	"net/http"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type subscriptionActivator interface {
	Activate(token uuid.UUID) error
}

func NewConfirmGETHandler(logger *zap.Logger, service subscriptionActivator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")

		parsedToken, err := uuid.Parse(token)
		if err != nil {
			logger.Warn("Invalid UUID in confirm subscription request",
				zap.String("raw_token", token),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			return
		}

		err = service.Activate(parsedToken)
		switch {
		case errors.Is(err, domain.ErrSubNotFound):
			logger.Info("Subscription token not found",
				zap.String("token", parsedToken.String()),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return

		case errors.Is(err, domain.ErrInternal):
			logger.Error("Internal error during subscription activation",
				zap.String("token", parsedToken.String()),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate subscription"})
			return

		case err != nil:
			logger.Error("Unexpected error during subscription activation",
				zap.String("token", parsedToken.String()),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate subscription"})
			return
		}

		logger.Info("Subscription successfully confirmed",
			zap.String("token", parsedToken.String()),
		)
		c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed successfully"})
	}
}
