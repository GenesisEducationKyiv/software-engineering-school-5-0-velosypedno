//go:build integration

package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSubscribeConfirmFlow(t *testing.T) {
	// Step 1: Clear DB and insert a test subscription
	t.Log("Clearing DB and inserting a fake subscription...")
	clearDB()

	email := "test.confirm@example.com"
	frequency := "daily"
	city := "Kyiv"
	token := uuid.New()

	_, err := DB.Exec(`
		INSERT INTO subscriptions (id, email, frequency, city, activated, token)
		VALUES ($1, $2, $3, $4, false, $5)
	`, uuid.New(), email, frequency, city, token)

	require.NoError(t, err, "Failed to insert test subscription")
	t.Logf("Inserted subscription with token: %s", token)

	// Step 2: Make GET request to /api/confirm/:token
	url := fmt.Sprintf("%s/api/confirm/%s", apiURL, token.String())
	t.Logf("Sending GET request to: %s", url)

	resp, err := http.Get(url)
	require.NoError(t, err, "Failed to send GET request to confirm subscription")
	defer resp.Body.Close()

	t.Logf("Response status code: %d", resp.StatusCode)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK, got %d", resp.StatusCode)

	// Step 3: Verify that subscription is now activated
	var activated bool
	err = DB.QueryRow("SELECT activated FROM subscriptions WHERE token = $1", token).Scan(&activated)
	require.NoError(t, err, "Failed to query activation status: %v", err)

	require.True(t, activated, "Expected subscription to be activated, but it was not")
	t.Logf("Activated status in DB: %v", activated)
}
