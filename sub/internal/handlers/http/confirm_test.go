//go:build unit

package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"

	handlers "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/handlers/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSubscriptionActivator struct {
	mock.Mock
}

func (m *mockSubscriptionActivator) Activate(token uuid.UUID) error {
	args := m.Called(token)
	return args.Error(0)
}

func TestConfirmGETHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validUUID := uuid.New()
	invalidUUIDStr := "not-a-uuid"

	tests := []struct {
		name    string
		token   string
		mockErr error

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
			mockErr: errors.New("some internal error"),

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
			mockService := new(mockSubscriptionActivator)
			if tt.mockErr != nil || tt.expectedStatus != http.StatusBadRequest {
				tokenUUID, err := uuid.Parse(tt.token)
				if err == nil {
					mockService.On("Activate", tokenUUID).Return(tt.mockErr)
				}
			}
			route := gin.New()
			route.GET("/confirm/:token", handlers.NewConfirmGETHandler(mockService))

			// Act
			req := httptest.NewRequest(http.MethodGet, "/confirm/"+tt.token, nil)
			resp := httptest.NewRecorder()
			route.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, resp.Code)
			mockService.AssertExpectations(t)
		})
	}
}
