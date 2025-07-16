//go:build unit

package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	subh "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/http/subscription"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSubscriptionDeactivator struct {
	mock.Mock
}

func (m *mockSubscriptionDeactivator) Unsubscribe(token uuid.UUID) error {
	args := m.Called(token)
	return args.Error(0)
}

func TestUnsubscribeGETHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validUUID := uuid.New()
	invalidUUIDStr := "invalid-uuid"

	tests := []struct {
		name           string
		token          string
		mockErr        error
		expectedStatus int
	}{
		{
			name:    "InvalidToken",
			token:   invalidUUIDStr,
			mockErr: nil,

			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "TokenNotFound",
			token:   validUUID.String(),
			mockErr: domain.ErrSubNotFound,

			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "ErrInternal",
			token:   validUUID.String(),
			mockErr: errors.New("internal error"),

			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "Success",
			token:   validUUID.String(),
			mockErr: nil,

			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(mockSubscriptionDeactivator)
			if tt.mockErr != nil || tt.expectedStatus != http.StatusBadRequest {
				tokenUUID, err := uuid.Parse(tt.token)
				if err == nil {
					mockService.On("Unsubscribe", tokenUUID).Return(tt.mockErr)
				}
			}
			route := gin.New()
			route.GET("/unsubscribe/:token", subh.NewUnsubscribeGETHandler(mockService))

			// Act
			req := httptest.NewRequest(http.MethodGet, "/unsubscribe/"+tt.token, nil)
			resp := httptest.NewRecorder()
			route.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.Code)
		})
	}
}
