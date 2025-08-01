package domain

import "github.com/google/uuid"

type Frequency string

const (
	FreqDaily  Frequency = "daily"
	FreqHourly Frequency = "hourly"
)

type Subscription struct {
	ID        uuid.UUID
	Email     string
	Frequency string
	City      string
	Activated bool
	Token     uuid.UUID
}

type Weather struct {
	Temperature float64
	Humidity    float64
	Description string
}
