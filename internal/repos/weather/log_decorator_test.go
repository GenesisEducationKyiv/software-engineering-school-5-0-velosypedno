package repos_test

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	weathr "github.com/velosypedno/genesis-weather-api/internal/repos/weather"
)

type mockWeatherRepo struct {
	Response domain.Weather
	Err      error
	Called   bool
}

func (m *mockWeatherRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	m.Called = true
	return m.Response, m.Err
}

func TestLoggingWeatherRepo_Success(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	mock := &mockWeatherRepo{
		Response: domain.Weather{Temperature: 25.0, Humidity: 60.0, Description: "Clear"},
		Err:      nil,
	}
	repo := weathr.NewLogDecorator(mock, "MockRepo", logger)

	// Act
	result, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 25.0, result.Temperature)
	assert.True(t, mock.Called)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "MockRepo")
	assert.Contains(t, logOutput, "Kyiv")
}

func TestLoggingWeatherRepo_Error(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	mock := &mockWeatherRepo{
		Response: domain.Weather{},
		Err:      domain.ErrInternal,
	}
	repo := weathr.NewLogDecorator(mock, "MockRepo", logger)

	// Act
	result, err := repo.GetCurrent(context.Background(), "kmaTop")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.Weather{}, result)
	assert.True(t, mock.Called)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "MockRepo")
	assert.Contains(t, logOutput, "kmaTop")
	assert.Contains(t, logOutput, domain.ErrInternal.Error())
}
