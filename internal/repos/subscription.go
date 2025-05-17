package repos

import (
	"database/sql"
	"errors"

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

func (r *SubscriptionDBRepo) ActivateSubscription(token string) error {
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
