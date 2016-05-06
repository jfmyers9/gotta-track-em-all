package db

import (
	"database/sql"
	"strings"

	"github.com/jfmyers9/gotta-track-em-all/models"
	"github.com/pivotal-golang/lager"
)

type DB struct {
	sqlConn *sql.DB
}

func NewDB(sqlConn *sql.DB) *DB {
	return &DB{sqlConn}
}

var createUsersTable = `CREATE TABLE users (
	account_id VARCHAR(255) PRIMARY KEY,
	project_ids TEXT,
	tracker_api_token VARCHAR(255)
)`

func (d *DB) CreateInitialSchema(logger lager.Logger) error {
	createTables := []string{
		createUsersTable,
	}

	for _, stmt := range createTables {
		_, err := d.sqlConn.Exec(stmt)
		if err != nil {
			logger.Error("failed-creating-table", err)
		}
	}

	return nil
}

func (d *DB) CreateUser(logger lager.Logger, accountID string, projectIDs []string, trackerAPIToken string) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger.Info("inserting-user", lager.Data{"account_id": accountID, "projectIDs": projectIDs})
		_, err := tx.Exec(`
		  INSERT INTO users(account_id,project_ids,tracker_api_token) VALUES($1,$2,$3);`,
			accountID,
			strings.Join(projectIDs, ","),
			trackerAPIToken,
		)
		if err != nil {
			logger.Error("failed-inserting-user", err)
			return err
		}
		return nil
	})
}

func (d *DB) GetUser(logger lager.Logger, accountID string) (*models.User, error) {
	row := d.sqlConn.QueryRow("SELECT project_ids,tracker_api_token FROM users WHERE account_id = $1;", accountID)

	var projectIDs, trackerAPIToken string

	err := row.Scan(&projectIDs, &trackerAPIToken)
	if err != nil {
		logger.Error("failed-to-fetch-user", err)
		return nil, err
	}

	return &models.User{
		AccountID:       accountID,
		ProjectIDs:      strings.Split(projectIDs, ","),
		TrackerAPIToken: trackerAPIToken,
	}, nil
}

func (d *DB) UpdateUser(logger lager.Logger, accountID string, projectIDs []string, trackerAPIToken string) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger.Info("updating-user", lager.Data{"account_id": accountID})

		_, err := tx.Exec(`
		  UPDATE users SET project_ids = $1, tracker_api_token = $2 WHERE account_id = $3;`,
			strings.Join(projectIDs, ","),
			trackerAPIToken,
			accountID,
		)
		if err != nil {
			logger.Error("failed-inserting-user", err)
			return err
		}
		return nil
	})
}

func (d *DB) DeleteUser(logger lager.Logger, accountID string) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		_, err := tx.Exec(`
		  DELETE FROM users WHERE account_id = $1;`,
			accountID,
		)
		if err != nil {
			logger.Error("failed-inserting-user", err)
			return err
		}
		return nil
	})
}

func (d *DB) transact(logger lager.Logger, f func(logger lager.Logger, tx *sql.Tx) error) error {
	tx, err := d.sqlConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = f(logger, tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}
