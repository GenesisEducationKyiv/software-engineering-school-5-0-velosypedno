package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type subscriptionActivator interface {
	Activate(token uuid.UUID) error
}

func NewConfirmGETHandler(service subscriptionActivator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		parsedToken, err := uuid.Parse(token)
		if err != nil {
			err = fmt.Errorf("confirm subscription handler: failed to parse token: %v", err)
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			return
		}

		err = service.Activate(parsedToken)
		if errors.Is(err, domain.ErrSubNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
			return
		}
		if errors.Is(err, domain.ErrInternal) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate subscription"})
			return
		}
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate subscription"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed successfully"})
	}
}
