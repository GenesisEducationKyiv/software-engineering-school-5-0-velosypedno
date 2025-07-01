//go:build unit

package repos_test

import (
	"database/sql"
	"errors"
	"log"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	subr "github.com/velosypedno/genesis-weather-api/internal/repos/subscription"
)

const (
	pgUniqueViolationCode = "23505"
)

func closeDB(mock sqlmock.Sqlmock, db *sql.DB, t *testing.T) {
	mock.ExpectClose()
	if err := db.Close(); err != nil {
		t.Log(err)
	}
}

func TestCreateSubscription_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer closeDB(mock, db, t)
	repo := subr.NewDBRepo(db)
	sub := domain.Subscription{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Frequency: string(domain.FreqDaily),
		City:      "Kyiv",
		Activated: false,
		Token:     uuid.New(),
	}
	mock.ExpectExec(
		regexp.QuoteMeta(
			`INSERT INTO subscriptions (id, email, frequency, city, activated, token)`+
				` VALUES ($1, $2, $3, $4, $5, $6)`),
	).
		WithArgs(sub.ID, sub.Email, sub.Frequency, sub.City, sub.Activated, sub.Token).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Act
	err = repo.Create(sub)

	// Assert
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateSubscription_EmailExists(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer closeDB(mock, db, t)
	repo := subr.NewDBRepo(db)
	sub := domain.Subscription{
		ID:        uuid.New(),
		Email:     "exists@example.com",
		Frequency: string(domain.FreqDaily),
		City:      "Lviv",
		Activated: false,
		Token:     uuid.New(),
	}
	pqErr := &pq.Error{Code: pgUniqueViolationCode}

	mock.ExpectExec(
		regexp.QuoteMeta(
			`INSERT INTO subscriptions (id, email, frequency, city, activated, token) `+
				` VALUES ($1, $2, $3, $4, $5, $6)`,
		),
	).
		WithArgs(sub.ID, sub.Email, sub.Frequency, sub.City, sub.Activated, sub.Token).
		WillReturnError(pqErr)

	// Act
	err = repo.Create(sub)

	// Assert
	assert.ErrorIs(t, err, domain.ErrSubAlreadyExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestActivateSubscription_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer closeDB(mock, db, t)

	repo := subr.NewDBRepo(db)
	token := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE subscriptions SET activated = true WHERE token = $1`)).
		WithArgs(token).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err = repo.Activate(token)

	// Assert
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestActivateSubscription_TokenNotFound(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer closeDB(mock, db, t)

	repo := subr.NewDBRepo(db)
	token := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE subscriptions SET activated = true WHERE token = $1`)).
		WithArgs(token).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Act
	err = repo.Activate(token)

	// Assert
	assert.ErrorIs(t, err, domain.ErrSubNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteSubscriptionByToken_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer closeDB(mock, db, t)

	repo := subr.NewDBRepo(db)
	token := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM subscriptions WHERE token = $1`)).
		WithArgs(token).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err = repo.DeleteByToken(token)

	// Assert
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteSubscriptionByToken_TokenNotFound(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer closeDB(mock, db, t)

	repo := subr.NewDBRepo(db)
	token := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM subscriptions WHERE token = $1`)).
		WithArgs(token).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Act
	err = repo.DeleteByToken(token)

	// Assert
	assert.ErrorIs(t, err, domain.ErrSubNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetActivatedSubscriptionsByFreq_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer closeDB(mock, db, t)

	repo := subr.NewDBRepo(db)

	freq := domain.FreqDaily

	rows := sqlmock.NewRows([]string{"id", "email", "frequency", "city", "activated", "token"}).
		AddRow(uuid.New(), "user1@example.com", freq, "Kyiv", true, uuid.New()).
		AddRow(uuid.New(), "user2@example.com", freq, "Lviv", true, uuid.New())

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM subscriptions WHERE activated = true AND frequency = $1`),
	).
		WithArgs(freq).
		WillReturnRows(rows)

	// Act
	subs, err := repo.GetActivatedByFreq(freq)

	// Assert
	require.NoError(t, err)
	assert.Len(t, subs, 2)
}

func TestGetActivatedSubscriptionsByFreq_QueryError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer closeDB(mock, db, t)

	repo := subr.NewDBRepo(db)
	freq := domain.FreqDaily
	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM subscriptions WHERE activated = true AND frequency = $1`),
	).
		WithArgs(freq).
		WillReturnError(errors.New("query error"))

	// Act
	subs, err := repo.GetActivatedByFreq(freq)

	// Assert
	require.Error(t, err)
	assert.Nil(t, subs)
}
