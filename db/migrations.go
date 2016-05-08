package db

import (
	"sort"

	"github.com/jfmyers9/gotta-track-em-all/db/migrations"
	"github.com/pivotal-golang/lager"
)

func (d *DB) RunMigrations(logger lager.Logger) error {
	sort.Sort(migrations.MigrationsToRun)

	currentVersion, err := d.GetVersion(logger)
	if err != nil {
		if err == ResourceNotFound {
			currentVersion = 0
		} else {
			logger.Error("failed-to-fetch-version", err)
			return err
		}
	}

	for _, migration := range migrations.MigrationsToRun {
		if migration.Version() <= currentVersion {
			continue
		}

		err := migration.Up(logger, d.sqlConn)
		if err != nil {
			logger.Error("failed-running-migration", err, lager.Data{"version": migration.Version()})
			downErr := migration.Down(logger, d.sqlConn)
			if downErr != nil {
				logger.Error("failed-reverting-migration", err, lager.Data{"version": migration.Version()})
			}
			return err
		}

		err = d.SetVersion(logger, migration.Version())
		if err != nil {
			logger.Error("failed-to-set-version", err)
			downErr := migration.Down(logger, d.sqlConn)
			if downErr != nil {
				logger.Error("failed-reverting-migration", err, lager.Data{"version": migration.Version()})
			}
			return err
		}
	}

	return nil
}
