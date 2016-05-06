package models

import "time"

type User struct {
	AccountID       string
	TrackerAPIToken string
	LastProcessedAt time.Time
	Pokemon         []int
}
