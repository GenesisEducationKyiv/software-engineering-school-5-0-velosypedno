package mock

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func NewVisualCrossingAPI() *httptest.Server {
	handler := gin.Default()

	handler.GET("/:city/today", func(c *gin.Context) {
		city := c.Param("city")

		if city == CityDoesNotExist {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "city not found",
			})
			return
		}

		temp := 18.3
		humidity := 75
		c.JSON(http.StatusOK, gin.H{
			"currentConditions": gin.H{
				"temp":       temp,
				"humidity":   humidity,
				"conditions": "Partly cloudy",
			},
		})
	})

	httpServer := httptest.NewServer(handler)
	return httpServer
}
