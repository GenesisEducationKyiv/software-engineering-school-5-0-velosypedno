package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
)

type subscriptionActivator interface {
	ActivateSubscription(token string) error
}

func NewConfirmGETHandler(service subscriptionActivator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		token := c.Param("token")
		_, err = uuid.Parse(token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			return
		}

		err = service.ActivateSubscription(token)
		if err == repos.ErrTokenNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate subscription"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed successfully"})
	}
}
