//go:build integration
// +build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestUnsubscribeFlow(t *testing.T) {
	// Step 1: Clear DB
	t.Log("Clearing the database...")
	ClearDB()

	// Step 2: Subscribe
	toEmail := "test.unsubscribe.flow@example.com"
	payload := fmt.Sprintf(`{
    	"email": "%s",
        "frequency": "daily",
        "city": "Kyiv"
    }`, toEmail)

	t.Log("Sending subscribe request...")
	resp, err := http.Post(TestServer.URL+"/api/subscribe", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("Failed to send POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code from subscribe: %d", resp.StatusCode)
	}

	// Step 3: Get token from DB
	t.Log("Fetching token from database...")
	var token uuid.UUID
	err = DB.QueryRow("SELECT token FROM subscriptions WHERE email = $1", toEmail).Scan(&token)
	if err != nil {
		t.Fatalf("Failed to get token from DB: %v", err)
	}
	t.Logf("Received token: %s", token.String())

	// Step 4: Activate subscription
	t.Log("Activating subscription...")
	confirmResp, err := http.Get(TestServer.URL + "/api/confirm/" + token.String())
	if err != nil {
		t.Fatalf("Failed to confirm subscription: %v", err)
	}
	defer confirmResp.Body.Close()
	if confirmResp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status from confirmation: %d", confirmResp.StatusCode)
	}

	// Step 5: Unsubscribe
	t.Log("Sending unsubscribe request...")
	unsubResp, err := http.Get(TestServer.URL + "/api/unsubscribe/" + token.String())
	if err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}
	defer unsubResp.Body.Close()

	var unsubBody map[string]string
	if err := json.NewDecoder(unsubResp.Body).Decode(&unsubBody); err != nil {
		t.Fatalf("Failed to decode unsubscribe response: %v", err)
	}
	t.Logf("Unsubscribe response: %v", unsubBody)

	if unsubResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK from unsubscribe, got %d", unsubResp.StatusCode)
	}

	// Step 6: Check DB — subscription should be gone
	t.Log("Checking that subscription no longer exists in DB...")
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE email = $1", toEmail).Scan(&count)
	if err != nil {
		t.Fatalf("DB query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected subscription to be deleted, found %d", count)
	} else {
		t.Log("Subscription successfully deleted.")
	}
}
