package repos

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/models"
)

const (
	PGUniqueViolationCode = "23505"
)

var ErrEmailAlreadyExists = errors.New("subscription with this email already exists")
var ErrTokenNotFound = errors.New("subscription with this token not found")

type SubscriptionDBRepo struct {
	db *sql.DB
}

func NewSubscriptionDBRepo(db *sql.DB) *SubscriptionDBRepo {
	return &SubscriptionDBRepo{
		db: db,
	}
}

func (r *SubscriptionDBRepo) CreateSubscription(subscription models.Subscription) error {
	_, err := r.db.Exec(`
		INSERT INTO subscriptions (id, email, frequency, city, activated, token)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
		subscription.ID,
		subscription.Email,
		subscription.Frequency,
		subscription.City,
		subscription.Activated,
		subscription.Token,
	)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == PGUniqueViolationCode {
			return ErrEmailAlreadyExists
		}
		return err
	}

	return nil
}

func (r *SubscriptionDBRepo) ActivateSubscription(token uuid.UUID) error {
	res, err := r.db.Exec("UPDATE subscriptions SET activated = true WHERE token = $1", token)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTokenNotFound
	}
	return nil
}

func (r *SubscriptionDBRepo) DeleteSubscriptionByToken(token uuid.UUID) error {
	res, err := r.db.Exec("DELETE FROM subscriptions WHERE token = $1", token)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTokenNotFound
	}
	return nil
}

func (r *SubscriptionDBRepo) GetActivatedSubscriptionsByFreq(freq models.Frequency) ([]models.Subscription, error) {
	rows, err := r.db.Query("SELECT * FROM subscriptions WHERE activated = true AND frequency = $1", freq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []models.Subscription
	for rows.Next() {
		var subscription models.Subscription
		if err := rows.Scan(
			&subscription.ID,
			&subscription.Email,
			&subscription.Frequency,
			&subscription.City,
			&subscription.Activated,
			&subscription.Token,
		); err != nil {
			return nil, err
		}
		result = append(result, subscription)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
