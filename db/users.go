package db

import (
	"database/sql"
	"strings"
	"time"

	"github.com/jfmyers9/gotta-track-em-all/models"
	"github.com/pivotal-golang/lager"
)

func (d *DB) CreateUser(logger lager.Logger, username string, trackerAPIToken string) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger.Info("inserting-user", lager.Data{"username": username})
		_, err := tx.Exec(`
		  INSERT INTO users(username,pokemon,last_processed_at,tracker_api_token) VALUES($1,$2,$3,$4);`,
			username,
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

func (d *DB) GetUser(logger lager.Logger, username string) (*models.User, error) {
	row := d.sqlConn.QueryRow("SELECT pokemon,last_processed_at,tracker_api_token FROM users WHERE username = $1;", username)

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
		Username:        username,
		Pokemon:         pokemon,
		LastProcessedAt: time.Unix(0, int64(lastProcessedAt)),
		TrackerAPIToken: trackerAPIToken,
	}, nil
}

func parsePokemonString(pokemonString string) ([]string, error) {
	result := strings.Split(pokemonString, ",")
	return result, nil
}

func (d *DB) Users(logger lager.Logger) ([]*models.User, error) {
	rows, err := d.sqlConn.Query("SELECT username,pokemon,last_processed_at,tracker_api_token FROM users;")
	if err != nil {
		logger.Error("failed-to-fetch-users", err)
		return nil, err
	}

	users := []*models.User{}

	for rows.Next() {
		var username, pokemonString, trackerAPIToken string
		var lastProcessedAt int

		err := rows.Scan(&username, &pokemonString, &lastProcessedAt, &trackerAPIToken)
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
			Username:        username,
			Pokemon:         pokemon,
			LastProcessedAt: time.Unix(0, int64(lastProcessedAt)),
			TrackerAPIToken: trackerAPIToken,
		})
	}

	return users, nil
}

func (d *DB) UpdateUser(logger lager.Logger, username string, trackerAPIToken string) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger.Info("updating-user", lager.Data{"username": username})

		_, err := tx.Exec(`
		  UPDATE users SET tracker_api_token = $2 WHERE username = $3;`,
			trackerAPIToken,
			username,
		)
		if err != nil {
			logger.Error("failed-inserting-user", err)
			return err
		}
		return nil
	})
}

func marshalPokemon(pokemon []string) string {
	return strings.Join(pokemon, ",")
}

func (d *DB) AddUserPokemon(logger lager.Logger, username string, newPokemon []string, lastProcessedAt time.Time) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		logger.Info("updating-user", lager.Data{"username": username})

		_, err := tx.Exec(`
		  UPDATE users SET pokemon=$1,last_processed_at=$2 WHERE username = $3;`,
			marshalPokemon(newPokemon),
			lastProcessedAt.UnixNano(),
			username,
		)
		if err != nil {
			logger.Error("failed-inserting-user", err)
			return err
		}
		return nil
	})
}

func (d *DB) DeleteUser(logger lager.Logger, username string) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		_, err := tx.Exec(`
		  DELETE FROM users WHERE username = $1;`,
			username,
		)
		if err != nil {
			logger.Error("failed-inserting-user", err)
			return err
		}
		return nil
	})
}
