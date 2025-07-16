package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	subsrv "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/subscription"
	"github.com/gin-gonic/gin"
)

type subReqBody struct {
	Email     string `json:"email" binding:"required,email"`
	Frequency string `json:"frequency" binding:"required,oneof=daily hourly"`
	City      string `json:"city" binding:"required"`
}

type subscriber interface {
	Subscribe(subInput subsrv.SubscriptionInput) error
}

func NewSubscribePOSTHandler(service subscriber) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body subReqBody
		if err := c.ShouldBindJSON(&body); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		input := subsrv.SubscriptionInput{
			Email:     body.Email,
			Frequency: body.Frequency,
			City:      body.City,
		}

		err := service.Subscribe(input)
		if errors.Is(err, domain.ErrSubAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already subscribed"})
			return
		}
		if errors.Is(err, domain.ErrInternal) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
			return
		}
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Subscription successful. Confirmation email sent."})
	}
}
