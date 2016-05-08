package db

import (
	"database/sql"
	"errors"

	"github.com/pivotal-golang/lager"
)

var ResourceNotFound = errors.New("resource-not-found")

type DB struct {
	sqlConn *sql.DB
}

func NewDB(sqlConn *sql.DB) *DB {
	return &DB{sqlConn}
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
