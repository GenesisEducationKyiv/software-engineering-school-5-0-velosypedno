//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSubscribeSuccessFlow(t *testing.T) {
	// Step 1: Clear the database
	t.Log("Clearing the database...")
	clearDB()

	// Step 2: Send the gRPC Subscribe request
	toEmail := "test.subscribe.success@example.com"
	t.Logf("Sending gRPC subscription request for email: %s", toEmail)

	req := &pb.SubscribeRequest{
		Email:     toEmail,
		Frequency: "daily",
		City:      "Kyiv",
	}

	_, err := SubGRPCClient.Subscribe(context.Background(), req)
	require.NoError(t, err, "Failed to send gRPC Subscribe request")

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
		require.NoError(t, err, "Failed to query MailHog API: %v", err)

		var result smtpAPISearchResult
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		require.NoError(t, err, "Failed to parse MailHog API response: %v", err)

		if result.Total >= 1 {
			t.Logf("Found %d email(s) in MailHog", result.Total)
			break
		}

		require.Less(t, time.Since(start), timeout, "Timeout reached while waiting for email")

		t.Log("No email found yet, retrying...")
		time.Sleep(interval)
	}

	// Step 4: Check that the subscription was saved in the database
	t.Log("Checking subscription entry in the database...")
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE email = $1", toEmail).Scan(&count)
	require.NoError(t, err, "Failed to query subscription count: %v", err)
	t.Logf("Found %d subscription(s) in the database for email %s", count, toEmail)
	require.Equal(t, 1, count, "Expected 1 subscription in database, got %d", count)

	var activated bool
	err = DB.QueryRow("SELECT activated FROM subscriptions WHERE email = $1", toEmail).Scan(&activated)
	require.NoError(t, err, "Failed to query subscription activation status: %v", err)
	t.Logf("Subscription activated status: %v", activated)
	require.False(t, activated, "Expected subscription to be not activated, got activated = true")
}

func TestSubscribeDuplicateFlow(t *testing.T) {
	ctx := context.Background()

	// Step 1: Clear the database
	t.Log("Clearing the database...")
	clearDB()

	// Step 2: Send the first subscription request
	toEmail := "test.duplicate.subscribe@example.com"
	t.Log("Sending first gRPC subscription request...")
	req := &pb.SubscribeRequest{
		Email:     toEmail,
		Frequency: "daily",
		City:      "Kyiv",
	}
	_, err := SubGRPCClient.Subscribe(ctx, req)
	require.NoError(t, err, "Failed to send first Subscribe request: %v", err)

	// Step 3: Send the second (duplicate) subscription request
	t.Log("Sending duplicate gRPC subscription request...")
	_, err = SubGRPCClient.Subscribe(ctx, req)
	require.Error(t, err, "Expected error on duplicate Subscribe request")

	st, ok := status.FromError(err)
	require.True(t, ok, "Expected a gRPC status error")
	require.Equal(t, codes.AlreadyExists, st.Code(), "Expected AlreadyExists code for duplicate subscription, got %v", st.Code())

	// Step 4: Check that only one subscription exists in the database
	t.Log("Verifying that only one subscription is stored in the database...")
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE email = $1", toEmail).Scan(&count)
	require.NoError(t, err, "Failed to get subscription count from DB: %v", err)
	t.Logf("Found %d subscription(s) in the database for email %s", count, toEmail)
	require.Equal(t, 1, count, "Expected 1 subscription in database, got %d", count)
}
