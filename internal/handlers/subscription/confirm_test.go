package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	subh "github.com/velosypedno/genesis-weather-api/internal/handlers/subscription"
	subsrv "github.com/velosypedno/genesis-weather-api/internal/services/subscription"
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
		name           string
		token          string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "invalid UUID token",
			token:          invalidUUIDStr,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "token not found",
			token:          validUUID.String(),
			mockErr:        subsrv.ErrSubNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "internal error on activation",
			token:          validUUID.String(),
			mockErr:        errors.New("some internal error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "successful activation",
			token:          validUUID.String(),
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockSubscriptionActivator)
			if tt.mockErr != nil || tt.expectedStatus != http.StatusBadRequest {
				tokenUUID, err := uuid.Parse(tt.token)
				if err == nil {
					mockService.On("Activate", tokenUUID).Return(tt.mockErr)
				}
			}

			route := gin.New()
			route.GET("/confirm/:token", subh.NewConfirmGETHandler(mockService))
			req := httptest.NewRequest(http.MethodGet, "/confirm/"+tt.token, nil)
			resp := httptest.NewRecorder()
			route.ServeHTTP(resp, req)
			assert.Equal(t, tt.expectedStatus, resp.Code)
			mockService.AssertExpectations(t)
		})
	}
}
