package repos

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	pgUniqueViolationCode = "23505"
)

type DBRepo struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewDBRepo(logger *zap.Logger, db *sql.DB) *DBRepo {
	return &DBRepo{
		logger: logger.With(zap.String("repo", "DBRepo")),
		db:     db,
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

		r.logger.Error(
			"Failed to create subscription",
			zap.Error(err),
			zap.String("method", "Create"),
			zap.String("city", subscription.City),
			zap.String("frequency", subscription.Frequency),
			zap.String("email_hash", logging.HashEmail(subscription.Email)),
		)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}

	return nil
}

func (r *DBRepo) Activate(token uuid.UUID) error {
	logger := r.logger.With(
		zap.String("method", "Activate"),
		zap.String("token", token.String()),
	)
	res, err := r.db.Exec("UPDATE subscriptions SET activated = true WHERE token = $1", token)
	if err != nil {
		logger.Error(
			"Failed to activate subscription",
			zap.Error(err),
		)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("subscription repo: activate: %v\n", err)
		logger.Error(
			"Failed to activate subscription",
			zap.Error(err),
		)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	if rowsAffected == 0 {
		logger.Warn(
			"Subscription not found",
		)
		return fmt.Errorf("subscription repo: %w", domain.ErrSubNotFound)
	}
	return nil
}

func (r *DBRepo) DeleteByToken(token uuid.UUID) error {
	logger := r.logger.With(
		zap.String("method", "DeleteByToken"),
		zap.String("token", token.String()),
	)
	res, err := r.db.Exec("DELETE FROM subscriptions WHERE token = $1", token)
	if err != nil {
		logger.Error(
			"Failed to delete subscription",
			zap.Error(err),
		)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logger.Error(
			"Failed to delete subscription",
			zap.Error(err),
		)
		return fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	if rowsAffected == 0 {
		logger.Warn(
			"Subscription not found",
		)
		return fmt.Errorf("subscription repo: %w", domain.ErrSubNotFound)
	}
	return nil
}

func (r *DBRepo) GetActivatedByFreq(freq domain.Frequency) ([]domain.Subscription, error) {
	logger := r.logger.With(
		zap.String("method", "GetActivatedByFreq"),
		zap.String("frequency", string(freq)),
	)
	rows, err := r.db.Query("SELECT * FROM subscriptions WHERE activated = true AND frequency = $1", freq)
	if err != nil {
		r.logger.Error(
			"Failed to get activated subscriptions",
			zap.Error(err),
			zap.String("method", "GetActivatedByFreq"),
			zap.String("frequency", string(freq)),
		)
		return nil, fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error(
				"Failed to close rows",
				zap.Error(err),
			)
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
			logger.Error(
				"Failed to scan subscription",
				zap.Error(err),
			)
			return nil, err
		}
		result = append(result, subscription)
	}
	if err := rows.Err(); err != nil {
		logger.Error(
			"Failed to get activated subscriptions",
			zap.Error(err),
		)
		return nil, fmt.Errorf("subscription repo: %w", domain.ErrInternal)
	}
	return result, nil
}
