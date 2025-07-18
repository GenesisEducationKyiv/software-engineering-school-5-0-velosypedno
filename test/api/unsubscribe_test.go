//go:build integration

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	subv1alpha2 "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
)

func TestUnsubscribeFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Clear DB
	t.Log("Clearing the database...")
	clearDB()

	// Step 2: Subscribe via gRPC
	email := "test.unsubscribe.flow@example.com"
	frequency := "daily"
	city := "Kyiv"

	t.Log("Sending Subscribe RPC...")
	_, err := SubGRPCClient.Subscribe(ctx, &subv1alpha2.SubscribeRequest{
		Email:     email,
		Frequency: frequency,
		City:      city,
	})
	require.NoError(t, err, "Subscribe RPC failed")

	// Step 3: Get token from DB
	t.Log("Fetching token from database...")
	var token uuid.UUID
	err = DB.QueryRow("SELECT token FROM subscriptions WHERE email = $1", email).Scan(&token)
	require.NoError(t, err, "Failed to get token from DB")
	t.Logf("Received token: %s", token.String())

	// Step 4: Confirm via gRPC
	t.Log("Sending Confirm RPC...")
	_, err = SubGRPCClient.Confirm(ctx, &subv1alpha2.ConfirmRequest{
		Token: token.String(),
	})
	require.NoError(t, err, "Confirm RPC failed")

	// Step 5: Unsubscribe via gRPC
	t.Log("Sending Unsubscribe RPC...")
	resp, err := SubGRPCClient.Unsubscribe(ctx, &subv1alpha2.UnsubscribeRequest{
		Token: token.String(),
	})
	require.NoError(t, err, "Unsubscribe RPC failed")
	require.NotNil(t, resp, "Unsubscribe response is nil")
	t.Logf("Unsubscribe response: %s", resp.Message)

	// Step 6: Check DB — subscription should be gone
	t.Log("Verifying that subscription was deleted from DB...")
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE email = $1", email).Scan(&count)
	require.NoError(t, err, "DB query failed")
	require.Equal(t, 0, count, "Expected subscription to be deleted, found %d", count)

	t.Log("Subscription successfully deleted.")
}
