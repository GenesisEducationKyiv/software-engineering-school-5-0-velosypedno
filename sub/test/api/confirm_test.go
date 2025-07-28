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

func TestSubscribeConfirmFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	// Step 2: Call Confirm RPC method
	t.Logf("Calling gRPC Confirm method with token: %s", token)

	resp, err := SubGRPCClient.Confirm(ctx, &subv1alpha2.ConfirmRequest{
		Token: token.String(),
	})
	require.NoError(t, err, "gRPC Confirm call failed")
	require.NotNil(t, resp, "Confirm response is nil")

	t.Logf("Confirm response message: %s", resp.Message)

	// Step 3: Verify that subscription is now activated
	var activated bool
	err = DB.QueryRow("SELECT activated FROM subscriptions WHERE token = $1", token).Scan(&activated)
	require.NoError(t, err, "Failed to query activation status: %v", err)

	require.True(t, activated, "Expected subscription to be activated, but it was not")
	t.Logf("Activated status in DB: %v", activated)
}
