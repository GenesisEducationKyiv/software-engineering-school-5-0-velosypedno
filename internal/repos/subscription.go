package repos

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

const (
	PGUniqueViolationCode = "23505"
)

var (
	ErrEmailAlreadyExists = errors.New("subscription with this email already exists")
	ErrTokenNotFound      = errors.New("subscription with this token not found")
	ErrInternal           = errors.New("internal error")
)

type SubscriptionDBRepo struct {
	db *sql.DB
}

func NewSubscriptionDBRepo(db *sql.DB) *SubscriptionDBRepo {
	return &SubscriptionDBRepo{
		db: db,
	}
}

func (r *SubscriptionDBRepo) Create(subscription domain.Subscription) error {
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

		err = fmt.Errorf("subscription repo: failed to create subscription, err:%v ", err)
		log.Println(err)
		return ErrInternal
	}

	return nil
}

func (r *SubscriptionDBRepo) Activate(token uuid.UUID) error {
	res, err := r.db.Exec("UPDATE subscriptions SET activated = true WHERE token = $1", token)
	if err != nil {
		err = fmt.Errorf("subscription repo: failed to activate subscription, err:%v ", err)
		log.Println(err)
		return ErrInternal
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		err := fmt.Errorf("subscription repo: failed to activate subscription, err:%v ", err)
		log.Println(err)
		return ErrInternal
	}
	if rowsAffected == 0 {
		return ErrTokenNotFound
	}
	return nil
}

func (r *SubscriptionDBRepo) DeleteByToken(token uuid.UUID) error {
	res, err := r.db.Exec("DELETE FROM subscriptions WHERE token = $1", token)
	if err != nil {
		err = fmt.Errorf("subscription repo: failed to delete subscription, err:%v ", err)
		log.Println(err)
		return ErrInternal
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		err := fmt.Errorf("subscription repo: failed to delete subscription, err:%v ", err)
		log.Println(err)
		return ErrInternal
	}
	if rowsAffected == 0 {
		return ErrTokenNotFound
	}
	return nil
}

func (r *SubscriptionDBRepo) GetActivatedByFreq(freq domain.Frequency) ([]domain.Subscription, error) {
	rows, err := r.db.Query("SELECT * FROM subscriptions WHERE activated = true AND frequency = $1", freq)
	if err != nil {
		err = fmt.Errorf("subscription repo: failed to get subscriptions, err:%v ", err)
		log.Println(err)
		return nil, ErrInternal
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()
	var result []domain.Subscription
	for rows.Next() {
		var subscription domain.Subscription
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
		log.Println(err)
		return nil, ErrInternal
	}
	return result, nil
}
