package models

import "time"

type User struct {
	Username        string
	TrackerAPIToken string
	LastProcessedAt time.Time
	Pokemon         []string
}

type PokemonEntry struct {
	Index  int
	Name   string
	Weight float64
}
