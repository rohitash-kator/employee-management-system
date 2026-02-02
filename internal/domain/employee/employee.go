package employee

import (
	"time"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type Employee struct {
	ID         string
	FirstName  string
	LastName   string
	Email      string
	Department string
	Position   string
	Salary     float64
	Status     Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
