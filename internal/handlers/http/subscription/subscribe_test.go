//go:build unit

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
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	subh "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/subscription"
	subsrv "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/subscription"
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
			name:       "InvalidJson",
			body:       `{"email": "test@example.com"}`,
			mockSrvErr: nil,

			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "InvalidEmail",
			body:       `{"email": "invalid", "frequency": "daily", "city": "Kyiv"}`,
			mockSrvErr: nil,

			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "AlreadyExists",
			body:       `{"email": "test@example.com", "frequency": "daily", "city": "Kyiv"}`,
			mockSrvErr: domain.ErrSubAlreadyExists,

			expectedStatus: http.StatusConflict,
		},
		{
			name:       "ErrInternal",
			body:       `{"email": "test@example.com", "frequency": "daily", "city": "Kyiv"}`,
			mockSrvErr: errors.New("db failure"),

			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			body:       `{"email": "test@example.com", "frequency": "hourly", "city": "Lviv"}`,
			mockSrvErr: nil,

			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
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

			// Act
			req := httptest.NewRequest(http.MethodPost, "/subscribe", bytes.NewBuffer([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			route.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.Code)
			mockService.AssertExpectations(t)
		})
	}
}
