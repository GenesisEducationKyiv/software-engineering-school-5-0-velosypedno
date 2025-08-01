//go:build unit

package decorator_test

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/repos/decorator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	repo := decorator.NewLogDecorator(mock, "MockRepo", logger)

	// Act
	result, err := repo.GetCurrent(context.Background(), "Kyiv")

	// Assert
	require.NoError(t, err)
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
	repo := decorator.NewLogDecorator(mock, "MockRepo", logger)

	// Act
	result, err := repo.GetCurrent(context.Background(), "kmaTop")

	// Assert
	require.Error(t, err)
	assert.Equal(t, domain.Weather{}, result)
	assert.True(t, mock.Called)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "MockRepo")
	assert.Contains(t, logOutput, "kmaTop")
	assert.Contains(t, logOutput, domain.ErrInternal.Error())
}
