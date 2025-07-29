package handlers

import (
	"errors"
	"net/http"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/services"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type subReqBody struct {
	Email     string `json:"email" binding:"required,email"`
	Frequency string `json:"frequency" binding:"required,oneof=daily hourly"`
	City      string `json:"city" binding:"required"`
}

type subscriber interface {
	Subscribe(subInput services.SubscriptionInput) error
}

func NewSubscribePOSTHandler(service subscriber, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body subReqBody
		if err := c.ShouldBindJSON(&body); err != nil {
			logger.Warn("Invalid subscription request body",
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		input := services.SubscriptionInput{
			Email:     body.Email,
			Frequency: body.Frequency,
			City:      body.City,
		}

		err := service.Subscribe(input)
		switch {
		case errors.Is(err, domain.ErrSubAlreadyExists):

			logger.Info("Attempt to subscribe already subscribed email",
				zap.String("email_hash", logging.HashEmail(body.Email)),
			)
			c.JSON(http.StatusConflict, gin.H{"error": "Email already subscribed"})
			return

		case errors.Is(err, domain.ErrInternal):
			logger.Error("Internal error on subscription creation",
				zap.String("email_hash", logging.HashEmail(body.Email)),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
			return

		case err != nil:
			logger.Error("Unexpected error on subscription creation",
				zap.String("email_hash", logging.HashEmail(body.Email)),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
			return
		}

		logger.Info("Subscription created successfully",
			zap.String("email_hash", logging.HashEmail(body.Email)),
			zap.String("city", body.City),
			zap.String("frequency", body.Frequency),
		)

		c.JSON(http.StatusOK, gin.H{"message": "Subscription successful. Confirmation email sent."})
	}
}
