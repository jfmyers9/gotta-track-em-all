package models

import "time"

type User struct {
	Username        string
	TrackerAPIToken string
	LastProcessedAt time.Time
	Pokemon         []int
}
