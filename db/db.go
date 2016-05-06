package db

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

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
	pokemon TEXT,
	last_processed_at BIGINT,
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

func (d *DB) CreateUser(logger lager.Logger, accountID string, trackerAPIToken string) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger.Info("inserting-user", lager.Data{"account_id": accountID})
		_, err := tx.Exec(`
		  INSERT INTO users(account_id,pokemon,last_processed_at,tracker_api_token) VALUES($1,$2,$3,$4);`,
			accountID,
			"",
			0,
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
	row := d.sqlConn.QueryRow("SELECT pokemon,last_processed_at,tracker_api_token FROM users WHERE account_id = $1;", accountID)

	var pokemonString, trackerAPIToken string
	var lastProcessedAt int

	err := row.Scan(&pokemonString, &lastProcessedAt, &trackerAPIToken)
	if err != nil {
		logger.Error("failed-to-fetch-user", err)
		return nil, err
	}

	pokemon, err := parsePokemonString(pokemonString)
	if err != nil {
		logger.Error("failed-to-fetch-user", err)
		return nil, err
	}

	return &models.User{
		AccountID:       accountID,
		Pokemon:         pokemon,
		LastProcessedAt: time.Unix(0, int64(lastProcessedAt)),
		TrackerAPIToken: trackerAPIToken,
	}, nil
}

func parsePokemonString(pokemonString string) ([]int, error) {
	result := []int{}

	nums := strings.Split(pokemonString, ",")
	for _, num := range nums {
		if num == "" {
			continue
		}
		i, err := strconv.Atoi(num)
		if err != nil {
			return nil, err
		}

		result = append(result, i)
	}

	return result, nil
}

func (d *DB) Users(logger lager.Logger) ([]*models.User, error) {
	rows, err := d.sqlConn.Query("SELECT account_id,pokemon,last_processed_at,tracker_api_token FROM users;")
	if err != nil {
		logger.Error("failed-to-fetch-users", err)
		return nil, err
	}

	users := []*models.User{}

	for rows.Next() {
		var accountID, pokemonString, trackerAPIToken string
		var lastProcessedAt int

		err := rows.Scan(&accountID, &pokemonString, &lastProcessedAt, &trackerAPIToken)
		if err != nil {
			logger.Error("failed-to-fetch-user", err)
			return nil, err
		}

		pokemon, err := parsePokemonString(pokemonString)
		if err != nil {
			logger.Error("failed-to-fetch-user", err)
			return nil, err
		}

		users = append(users, &models.User{
			AccountID:       accountID,
			Pokemon:         pokemon,
			LastProcessedAt: time.Unix(0, int64(lastProcessedAt)),
			TrackerAPIToken: trackerAPIToken,
		})
	}

	return users, nil
}

func (d *DB) UpdateUser(logger lager.Logger, accountID string, trackerAPIToken string) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger.Info("updating-user", lager.Data{"account_id": accountID})

		_, err := tx.Exec(`
		  UPDATE users SET tracker_api_token = $2 WHERE account_id = $3;`,
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

func marshalPokemon(pokemon []int) string {
	result := []string{}
	for _, i := range pokemon {
		s := strconv.Itoa(i)
		result = append(result, s)
	}
	return strings.Join(result, ",")
}

func (d *DB) AddUserPokemon(logger lager.Logger, accountID string, newPokemon []int, lastProcessedAt time.Time) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger.Info("updating-user", lager.Data{"account_id": accountID})

		_, err := tx.Exec(`
		  UPDATE users SET pokemon=$1,last_processed_at=$2 WHERE account_id = $3;`,
			marshalPokemon(newPokemon),
			lastProcessedAt.UnixNano(),
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
