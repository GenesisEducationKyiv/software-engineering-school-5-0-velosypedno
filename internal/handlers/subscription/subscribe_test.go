package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	subh "github.com/velosypedno/genesis-weather-api/internal/handlers/subscription"
	subsrv "github.com/velosypedno/genesis-weather-api/internal/services/subscription"
)

func extractField(jsonStr, field string) string {
	var m map[string]string
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		log.Fatal(err)
	}
	return m[field]
}

type mockSubscriber struct {
	mock.Mock
}

func (m *mockSubscriber) Subscribe(input subsrv.SubscriptionInput) error {
	args := m.Called(input)
	return args.Error(0)
}

func TestSubscribePOSTHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockSrvErr     error
		expectedStatus int
	}{
		{
			name:           "invalid json (missing fields)",
			body:           `{"email": "test@example.com"}`,
			mockSrvErr:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid email",
			body:           `{"email": "invalid", "frequency": "daily", "city": "Kyiv"}`,
			mockSrvErr:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "conflict - email already exists",
			body:           `{"email": "test@example.com", "frequency": "daily", "city": "Kyiv"}`,
			mockSrvErr:     subsrv.ErrSubAlreadyExists,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "internal error during subscribe",
			body:           `{"email": "test@example.com", "frequency": "daily", "city": "Kyiv"}`,
			mockSrvErr:     errors.New("db failure"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "successful subscription",
			body:           `{"email": "test@example.com", "frequency": "hourly", "city": "Lviv"}`,
			mockSrvErr:     nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockSubscriber)

			if tt.expectedStatus != http.StatusBadRequest {
				input := subsrv.SubscriptionInput{
					Email:     extractField(tt.body, "email"),
					Frequency: extractField(tt.body, "frequency"),
					City:      extractField(tt.body, "city"),
				}
				mockService.On("Subscribe", input).Return(tt.mockSrvErr)
			}

			route := gin.New()
			route.POST("/subscribe", subh.NewSubscribePOSTHandler(mockService))
			req := httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewBuffer([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			route.ServeHTTP(resp, req)
			assert.Equal(t, tt.expectedStatus, resp.Code)
			mockService.AssertExpectations(t)
		})
	}
}
