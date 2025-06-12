package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
	"github.com/velosypedno/genesis-weather-api/internal/services"
)

type subReqBody struct {
	Email     string `json:"email" binding:"required,email"`
	Frequency string `json:"frequency" binding:"required,oneof=daily hourly"`
	City      string `json:"city" binding:"required"`
}

type subscriber interface {
	Subscribe(subscription services.SubscriptionInput) error
}

func NewSubscribePOSTHandler(service subscriber) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body subReqBody
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		input := services.SubscriptionInput{
			Email:     body.Email,
			Frequency: body.Frequency,
			City:      body.City,
		}
		err := service.Subscribe(input)
		if err != nil {
			if errors.Is(err, repos.ErrEmailAlreadyExists) {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already subscribed"})
				return
			}
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Subscription successful. Confirmation email sent."})
	}
}
