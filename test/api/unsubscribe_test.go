//go:build integration

package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnsubscribeFlow(t *testing.T) {
	// Step 1: Clear DB
	t.Log("Clearing the database...")
	clearDB()

	// Step 2: Subscribe
	toEmail := "test.unsubscribe.flow@example.com"
	payload := fmt.Sprintf(`{
    	"email": "%s",
        "frequency": "daily",
        "city": "Kyiv"
    }`, toEmail)

	t.Log("Sending subscribe request...")
	endpoint := "/api/subscribe"

	t.Logf("Sending subscribe request to: %s", endpoint)
	resp, err := http.Post(apiURL+endpoint, "application/json", strings.NewReader(payload))
	require.NoError(t, err, "Failed to send POST: %v", err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK, got %d", resp.StatusCode)

	// Step 3: Get token from DB
	t.Log("Fetching token from database...")
	var token uuid.UUID
	err = DB.QueryRow("SELECT token FROM subscriptions WHERE email = $1", toEmail).Scan(&token)
	require.NoError(t, err, "Failed to get token from DB: %v", err)
	t.Logf("Received token: %s", token.String())

	// Step 4: Activate subscription
	t.Log("Activating subscription...")
	confirmResp, err := http.Get(apiURL + "/api/confirm/" + token.String())
	require.NoError(t, err, "Failed to confirm subscription: %v", err)
	defer confirmResp.Body.Close()
	require.Equal(t, http.StatusOK, confirmResp.StatusCode, "Expected status 200 OK, got %d", confirmResp.StatusCode)

	// Step 5: Unsubscribe
	t.Log("Sending unsubscribe request...")
	unsubResp, err := http.Get(apiURL + "/api/unsubscribe/" + token.String())
	require.NoError(t, err, "Failed to unsubscribe: %v", err)
	defer unsubResp.Body.Close()

	var unsubBody map[string]string
	require.NoError(t, err, "Failed to decode unsubscribe response: %v", err)
	t.Logf("Unsubscribe response: %v", unsubBody)
	require.Equal(t, http.StatusOK, unsubResp.StatusCode, "Expected status 200 OK, got %d", unsubResp.StatusCode)

	// Step 6: Check DB — subscription should be gone
	t.Log("Checking that subscription no longer exists in DB...")
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE email = $1", toEmail).Scan(&count)
	require.NoError(t, err, "DB query failed: %v", err)
	require.Equal(t, 0, count, "Expected subscription to be deleted, found %d", count)
	t.Log("Subscription successfully deleted.")
}
