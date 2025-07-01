//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscribeDuplicateFlow(t *testing.T) {
	// Step 1: Clear the database
	t.Log("Clearing the database...")
	clearDB()

	// Step 2: Send the first subscription request
	toEmail := "test.duplicate.subscribe@example.com"
	payload := fmt.Sprintf(`{
        "email": "%s",
        "frequency": "daily",
        "city": "Kyiv"
    }`, toEmail)
	t.Log("Sending first subscription request...")
	endpoint := "/api/subscribe"
	resp, err := http.Post(apiURL+endpoint, "application/json", strings.NewReader(payload))
	require.NoError(t, err, "Failed to send POST: %v", err)
	resp.Body.Close()

	// Step 3: Send the second (duplicate) subscription request
	t.Log("Sending duplicate subscription request...")
	resp, err = http.Post(apiURL+endpoint, "application/json", strings.NewReader(payload))
	require.NoError(t, err, "Failed to send duplicate POST: %v", err)
	defer resp.Body.Close()

	t.Logf("Received response with status code: %d", resp.StatusCode)
	require.Equal(t, http.StatusConflict, resp.StatusCode, "Expected status 409 Conflict for duplicate subscription, got %d", resp.StatusCode)

	// Step 4: Check that only one subscription exists in the database
	t.Log("Verifying that only one subscription is stored in the database...")
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE email = $1", toEmail).Scan(&count)
	require.NoError(t, err, "Failed to get subscription count from DB: %v", err)
	t.Logf("Found %d subscription(s) in the database for email %s", count, toEmail)
	require.Equal(t, 1, count, "Expected 1 subscription in database, got %d", count)
}
