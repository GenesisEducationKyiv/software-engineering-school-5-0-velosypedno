package models

import "github.com/google/uuid"

type Subscription struct {
	ID        uuid.UUID
	Email     string
	Frequency string
	City      string
	Activated bool
	Token     uuid.UUID
}
