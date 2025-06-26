package repos

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

const (
	pgUniqueViolationCode = "23505"
)

type DBRepo struct {
	db *sql.DB
}

func NewDBRepo(db *sql.DB) *DBRepo {
	return &DBRepo{
		db: db,
	}
}

func (r *DBRepo) Create(subscription domain.Subscription) error {
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
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgUniqueViolationCode {
			return domain.ErrSubAlreadyExists
		}

		log.Printf("subscription repo: create: %v\n", err)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}

	return nil
}

func (r *DBRepo) Activate(token uuid.UUID) error {
	res, err := r.db.Exec("UPDATE subscriptions SET activated = true WHERE token = $1", token)
	if err != nil {
		log.Printf("subscription repo: activate: %v\n", err)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("subscription repo: activate: %v\n", err)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subscription repo: %w", domain.ErrSubNotFound)
	}
	return nil
}

func (r *DBRepo) DeleteByToken(token uuid.UUID) error {
	res, err := r.db.Exec("DELETE FROM subscriptions WHERE token = $1", token)
	if err != nil {
		log.Printf("subscription repo: delete: %v\n", err)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("subscription repo: delete: %v\n", err)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subscription repo: %w", domain.ErrSubNotFound)
	}
	return nil
}

func (r *DBRepo) GetActivatedByFreq(freq domain.Frequency) ([]domain.Subscription, error) {
	rows, err := r.db.Query("SELECT * FROM subscriptions WHERE activated = true AND frequency = $1", freq)
	if err != nil {
		log.Printf("subscription repo: select: %v\n", err)
		return nil, fmt.Errorf("subscription repo: %w", domain.ErrInternal)
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
		log.Printf("subscription repo: select: %v\n", err)
		return nil, fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	return result, nil
}
