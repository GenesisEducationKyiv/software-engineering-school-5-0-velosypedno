package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
)

type subscriptionDeactivator interface {
	Unsubscribe(token uuid.UUID) error
}

func NewUnsubscribeGETHandler(service subscriptionDeactivator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		parsedToken, err := uuid.Parse(token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			return
		}
		err = service.Unsubscribe(parsedToken)
		if err != nil {
			if errors.Is(err, repos.ErrTokenNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successful"})
	}
}
