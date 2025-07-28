//go:build integration

package consumers_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

func TestSubscribeEventConsumed(t *testing.T) {
	email := "test.queue.subscribe@gmail.com"
	token := uuid.New()
	event := messaging.SubscribeEvent{
		Email: email,
		Token: token.String(),
	}
	body, err := json.Marshal(event)
	require.NoError(t, err, "Failed to marshal subscribe event: %v", err)
	err = RMQChannel.Publish(
		messaging.ExchangeName,
		messaging.SubscribeRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	require.NoError(t, err, "Failed to publish subscribe event: %v", err)

	var (
		smtpAPIUrl = "http://localhost:8025/api/v2/search"
		searchUrl  = smtpAPIUrl + "?kind=to&query=" + email
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

		require.Less(t, time.Since(start), timeout, "Timeout reached")

		t.Log("No email found yet, retrying...")
		time.Sleep(interval)
	}
}
