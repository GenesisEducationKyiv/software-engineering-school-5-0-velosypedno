//go:build unit

package cb_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/cb"
)

func TestCircuitBreaker(main *testing.T) {
	main.Run("ClosedOnStartup", func(t *testing.T) {
		// Arrange
		breaker := cb.NewCircuitBreaker(time.Minute*6, 10, 1)

		// Assert
		require.Equal(t, breaker.State(), cb.Closed)
		assert.True(t, breaker.Allowed())
	})

	main.Run("ClosedBeforeLimit", func(t *testing.T) {
		// Arrange
		breaker := cb.NewCircuitBreaker(time.Minute*6, 10, 1)

		// Act
		for i := 0; i < 9; i++ {
			breaker.Fail()
		}

		// Assert
		require.Equal(t, breaker.State(), cb.Closed)
		assert.True(t, breaker.Allowed())
	})

	main.Run("OpenAfterLimit", func(t *testing.T) {
		// Arrange
		breaker := cb.NewCircuitBreaker(time.Minute*6, 10, 1)

		// Act
		for i := 0; i < 10; i++ {
			breaker.Fail()
		}

		// Assert
		require.Equal(t, breaker.State(), cb.Open)
		assert.False(t, breaker.Allowed())
	})

	main.Run("HalfOpenAfterTimeout", func(t *testing.T) {
		// Arrange
		currentTime := time.Now()
		breaker := cb.NewCircuitBreaker(time.Minute*6, 10, 1)
		breaker.Now = func() time.Time {
			return currentTime
		}
		currentTime = currentTime.Add(time.Second * 7)

		// Act
		for i := 0; i < 10; i++ {
			breaker.Fail()
		}
		require.False(t, breaker.Allowed())
		require.Equal(t, breaker.State(), cb.Open)
		currentTime = currentTime.Add(time.Minute * 7)
		require.Equal(t, breaker.State(), cb.HalfOpen)
		for i := 0; i < 9; i++ {
			currentTime = currentTime.Add(time.Second * 1)
			breaker.Fail()
			currentTime = currentTime.Add(time.Second * 1)
			require.Equal(t, breaker.State(), cb.Open)
		}
		currentTime = currentTime.Add(time.Minute * 7)

		// Assert
		require.Equal(t, cb.HalfOpen, breaker.State())
		require.True(t, breaker.Allowed())
	})

	main.Run("RecoverAfterSuccess", func(t *testing.T) {
		// Arrange
		currentTime := time.Now()
		breaker := cb.NewCircuitBreaker(time.Minute*6, 10, 3)
		breaker.Now = func() time.Time {
			return currentTime
		}
		currentTime = currentTime.Add(time.Second * 7)

		// Act
		for i := 0; i < 10; i++ {
			breaker.Fail()
		}
		require.Equal(t, breaker.State(), cb.Open)
		require.False(t, breaker.Allowed())
		currentTime = currentTime.Add(time.Minute * 7)
		for i := 0; i < 3; i++ {
			breaker.Success()
		}

		// Assert
		require.Equal(t, breaker.State(), cb.Closed)
		assert.True(t, breaker.Allowed())
	})

	main.Run("OpenAgainDueToFailure", func(t *testing.T) {
		// Arrange
		currentTime := time.Now()
		breaker := cb.NewCircuitBreaker(time.Minute*6, 10, 3)
		breaker.Now = func() time.Time {
			return currentTime
		}
		currentTime = currentTime.Add(time.Second * 7)

		// Act
		for i := 0; i < 10; i++ {
			breaker.Fail()
		}
		require.Equal(t, breaker.State(), cb.Open)
		require.False(t, breaker.Allowed())
		currentTime = currentTime.Add(time.Minute * 7)
		require.Equal(t, breaker.State(), cb.HalfOpen)
		require.True(t, breaker.Allowed())
		breaker.Fail()

		// Assert
		require.Equal(t, breaker.State(), cb.Open)
		assert.False(t, breaker.Allowed())
	})
}
