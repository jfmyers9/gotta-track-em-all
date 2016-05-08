package migrations

import (
	"database/sql"

	"github.com/pivotal-golang/lager"
)

var MigrationsToRun Migrations

type Migrations []Migration

type Migration interface {
	Up(logger lager.Logger, sqlConn *sql.DB) error
	Down(logger lager.Logger, sqlConn *sql.DB) error
	Version() int
}

// For Sorting
func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Less(i, j int) bool { return m[i].Version() < m[j].Version() }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

func AppendMigration(migration Migration) {
	for _, m := range MigrationsToRun {
		if migration.Version() == m.Version() {
			panic("Cannot have 2 migrations with the same verison")
		}
	}

	MigrationsToRun = append(MigrationsToRun, migration)
}
