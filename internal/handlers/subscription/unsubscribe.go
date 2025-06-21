package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	subsrv "github.com/velosypedno/genesis-weather-api/internal/services/subscription"
)

type subscriptionDeactivator interface {
	Unsubscribe(token uuid.UUID) error
}

func NewUnsubscribeGETHandler(service subscriptionDeactivator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		parsedToken, err := uuid.Parse(token)
		if err != nil {
			err = fmt.Errorf("unsubscribe subscription handler: failed to parse token: %v", err)
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			return
		}
		err = service.Unsubscribe(parsedToken)
		if errors.Is(err, subsrv.ErrSubNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return
		}
		if errors.Is(err, subsrv.ErrInternal) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe"})
			return
		}
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unsubscribe"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successful"})
	}
}
