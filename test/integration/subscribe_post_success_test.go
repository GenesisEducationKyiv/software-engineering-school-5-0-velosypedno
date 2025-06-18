//go:build integration
// +build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSubscribeSuccessFlow(t *testing.T) {
	// Step 1: Clear the database
	t.Log("Clearing the database...")
	clearDB()

	// Step 2: Send the POST request to /api/subscribe
	toEmail := "test.subscribe.success@example.com"
	payload := fmt.Sprintf(`{
        "email": "%s",
        "frequency": "daily",
        "city": "Kyiv"
    }`, toEmail)
	t.Logf("Sending subscription request for email: %s", toEmail)
	resp, err := http.Post(TestServer.URL+"/api/subscribe", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Received response with status code: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	// Step 3: Wait for the email to appear in MailHog
	t.Log("Waiting for confirmation email to appear in MailHog...")
	var (
		smtpAPIUrl = "http://localhost:8025/api/v2/search"
		searchUrl  = smtpAPIUrl + "?kind=to&query=" + toEmail
		timeout    = 5 * time.Second
		interval   = 300 * time.Millisecond
		start      = time.Now()
	)
	type smtpAPISearchResult struct {
		Total int `json:"total"`
	}

	for {
		t.Logf("Checking MailHog API: %s", searchUrl)
		resp, err := http.Get(searchUrl)
		if err != nil {
			t.Fatalf("Failed to query MailHog API: %v", err)
		}

		var result smtpAPISearchResult
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil {
			t.Fatalf("Failed to parse MailHog API response: %v", err)
		}

		if result.Total >= 1 {
			t.Logf("Found %d email(s) in MailHog", result.Total)
			break
		}

		if time.Since(start) > timeout {
			t.Errorf("Expected 1 email in MailHog, got %d after %v", result.Total, timeout)
			break
		}

		t.Log("No email found yet, retrying...")
		time.Sleep(interval)
	}

	// Step 4: Check that the subscription was saved in the database
	t.Log("Checking subscription entry in the database...")
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE email = $1", toEmail).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query subscription count: %v", err)
	}
	t.Logf("Found %d subscription(s) in the database for email %s", count, toEmail)
	if count != 1 {
		t.Errorf("Expected 1 subscription in database, got %d", count)
	}

	var activated bool
	err = DB.QueryRow("SELECT activated FROM subscriptions WHERE email = $1", toEmail).Scan(&activated)
	if err != nil {
		t.Fatalf("Failed to query subscription activation status: %v", err)
	}
	t.Logf("Subscription activated status: %v", activated)
	if activated {
		t.Errorf("Expected subscription to be not activated, got activated = true")
	}
}
