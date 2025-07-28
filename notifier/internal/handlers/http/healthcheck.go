package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func NewHealthcheckGETHandler(ch *amqp.Channel, queueNames []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, queueName := range queueNames {
			_, err := ch.QueueDeclarePassive(queueName, true, false, false, false, nil)
			if err != nil {
				log.Println(fmt.Errorf("healthcheck handler: %v", err))
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
