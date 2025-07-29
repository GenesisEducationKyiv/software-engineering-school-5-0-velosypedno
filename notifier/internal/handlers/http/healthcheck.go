package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func NewHealthcheckGETHandler(logger *zap.Logger, ch *amqp.Channel, queueNames []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("healthcheck request")
		for _, queueName := range queueNames {
			_, err := ch.QueueDeclarePassive(queueName, true, false, false, false, nil)
			if err != nil {
				logger.Warn("queue not available",
					zap.String("queue", queueName),
					zap.Error(err),
				)
			} else {
				logger.Debug("queue available", zap.String("queue", queueName))
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
