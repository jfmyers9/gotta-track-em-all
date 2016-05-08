package db

import (
	"database/sql"
	"strconv"

	"github.com/pivotal-golang/lager"
)

const VersionName = "version"

func (d *DB) GetVersion(logger lager.Logger) (int, error) {
	row := d.sqlConn.QueryRow(`SELECT value FROM configuration WHERE name = $1;`, VersionName)

	var version string
	err := row.Scan(&version)
	if err != nil {
		return -1, ResourceNotFound
	}

	return strconv.Atoi(version)
}

func (d *DB) SetVersion(logger lager.Logger, version int) error {
	return d.transact(logger, func(logger lager.Logger, tx *sql.Tx) error {
		result, err := tx.Exec(`
			UPDATE configuration SET value=$1 WHERE name=$2;`,
			strconv.Itoa(version),
			VersionName,
		)

		if err != nil {
			logger.Error("failed-updating-recourd", err)
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			panic(err)
		}

		if rowsAffected <= 0 {
			_, err := tx.Exec(`
				INSERT INTO configuration(name,value) VALUES($1,$2);`,
				VersionName,
				strconv.Itoa(version),
			)

			if err != nil {
				logger.Error("failed-inserting-version", err)
				return err
			}
		}

		return nil
	})
}
