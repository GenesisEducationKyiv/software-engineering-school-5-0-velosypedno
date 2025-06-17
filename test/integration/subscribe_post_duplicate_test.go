//go:build integration
// +build integration

package integration_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestSubscribeDuplicateFlow(t *testing.T) {
	// Step 1: Clear the database
	t.Log("Clearing the database...")
	ClearDB()

	// Step 2: Send the first subscription request
	toEmail := "test.duplicate.subscribe@example.com"
	payload := fmt.Sprintf(`{
        "email": "%s",
        "frequency": "daily",
        "city": "Kyiv"
    }`, toEmail)
	t.Log("Sending first subscription request...")
	resp, err := http.Post(TestServer.URL+"/api/subscribe", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("Failed to send first POST request: %v", err)
	}
	resp.Body.Close()

	// Step 3: Send the second (duplicate) subscription request
	t.Log("Sending duplicate subscription request...")
	resp, err = http.Post(TestServer.URL+"/api/subscribe", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("Failed to send duplicate POST request: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Received response with status code: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status 409 Conflict for duplicate subscription, got %d", resp.StatusCode)
	}

	// Step 4: Check that only one subscription exists in the database
	t.Log("Verifying that only one subscription is stored in the database...")
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE email = $1", toEmail).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query subscription count: %v", err)
	}
	t.Logf("Found %d subscription(s) in the database for email %s", count, toEmail)
	if count != 1 {
		t.Errorf("Expected 1 subscription in database, got %d", count)
	}
}
