//go:build unit

package cb_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/velosypedno/genesis-weather-api/pkg/cb"
)

func TestCircuitBreaker(main *testing.T) {
	main.Run("ClosedOnStartup", func(t *testing.T) {
		// Arrange
		cb := cb.NewCircuitBreaker(time.Minute*6, 10)

		// Assert
		assert.True(t, cb.IsClosed())
	})

	main.Run("ClosedBeforeLimit", func(t *testing.T) {
		// Arrange
		cb := cb.NewCircuitBreaker(time.Minute*6, 10)

		// Act
		for i := 0; i < 9; i++ {
			cb.Fail()
		}

		// Assert
		assert.True(t, cb.IsClosed())
	})

	main.Run("OpenAfterLimit", func(t *testing.T) {
		// Arrange
		cb := cb.NewCircuitBreaker(time.Minute*6, 10)

		// Act
		for i := 0; i < 10; i++ {
			cb.Fail()
		}

		// Assert
		assert.False(t, cb.IsClosed())
	})

	main.Run("ClosedAfterTimeout", func(t *testing.T) {
		// Arrange
		currentTime := time.Now()
		cb := cb.NewCircuitBreaker(time.Minute*6, 10)
		cb.Now = func() time.Time {
			return currentTime
		}
		currentTime = currentTime.Add(time.Second * 7)

		// Act
		for i := 0; i < 10; i++ {
			cb.Fail()
		}
		require.False(t, cb.IsClosed())
		currentTime = currentTime.Add(time.Minute * 7)

		// Assert
		assert.True(t, cb.IsClosed())
	})

	main.Run("RestCounterAfterTimeout", func(t *testing.T) {
		// Arrange
		currentTime := time.Now()
		cb := cb.NewCircuitBreaker(time.Minute*6, 10)
		cb.Now = func() time.Time {
			return currentTime
		}
		currentTime = currentTime.Add(time.Second * 7)

		// Act
		for i := 0; i < 10; i++ {
			cb.Fail()
		}
		require.False(t, cb.IsClosed())
		currentTime = currentTime.Add(time.Minute * 7)
		for i := 0; i < 9; i++ {
			cb.Fail()
		}

		// Assert
		require.True(t, cb.IsClosed())
	})

	main.Run("ResetCounterIfNotFailed", func(t *testing.T) {
		// Arrange
		currentTime := time.Now()
		cb := cb.NewCircuitBreaker(time.Minute*6, 10)
		cb.Now = func() time.Time {
			return currentTime
		}
		currentTime = currentTime.Add(time.Second * 7)

		// Act
		for i := 0; i < 9; i++ {
			cb.Fail()
		}
		require.True(t, cb.IsClosed())
		currentTime = currentTime.Add(time.Minute * 7)
		for i := 0; i < 9; i++ {
			cb.Fail()
		}

		// Assert
		assert.True(t, cb.IsClosed())
	})
}
