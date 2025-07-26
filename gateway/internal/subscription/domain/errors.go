package domain

import "errors"

var (
	ErrInternal         = errors.New("internal error")
	ErrSubNotFound      = errors.New("subscription not found")
	ErrSubInvalid       = errors.New("invalid")
	ErrSubAlreadyExists = errors.New("subscription already exists")
)
