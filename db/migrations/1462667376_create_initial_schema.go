package migrations

import (
	"database/sql"

	"github.com/pivotal-golang/lager"
)

func init() {
	AppendMigration(NewCreateInitialSchema())
}

type createInitialSchema struct{}

func NewCreateInitialSchema() *createInitialSchema {
	return &createInitialSchema{}
}

func (c *createInitialSchema) Up(logger lager.Logger, sqlConn *sql.DB) error {
	createTables := []string{
		createConfigurationTable,
		createUsersTable,
	}

	for _, stmt := range createTables {
		_, err := sqlConn.Exec(stmt)
		if err != nil {
			logger.Error("failed-creating-table", err)
			return err
		}
	}

	return nil
}

func (c *createInitialSchema) Down(logger lager.Logger, sqlConn *sql.DB) error {
	dropTables := []string{
		dropConfigurationTable,
		dropUsersTable,
	}

	for _, stmt := range dropTables {
		_, err := sqlConn.Exec(stmt)
		if err != nil {
			logger.Error("failed-dropping-table", err)
		}
	}

	return nil
}

func (c *createInitialSchema) Version() int {
	return 1462667376
}

var createConfigurationTable = `CREATE TABLE configuration (
	name VARCHAR(255) PRIMARY KEY,
	value VARCHAR(255) NOT NULL
)`

var dropConfigurationTable = `DROP TABLE configuration;`

var createUsersTable = `CREATE TABLE users (
	username VARCHAR(255) PRIMARY KEY,
	pokemon TEXT,
	last_processed_at BIGINT,
	tracker_api_token VARCHAR(255)
)`

var dropUsersTable = `DROP TABLE users;`
