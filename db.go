package pqx

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/lib/pq"
)

// Logger interface consumed by this package.
type Logger interface {
	Print(...interface{})
}

// EnsureTempDB will drop/create new temporary db with suffix in db name
// and return opened temporary db together with cleanup func which will
// close and drop temporary db. It will use dbCfg to connect to existing
// database first, and then reuse same dbCfg with modified DBName to
// connect to temporary db.
//
// It'll also create schema with name set to dbCfg.User in temporary db.
//
// Recommended value for suffix is your package's import path.
func EnsureTempDB(log Logger, suffix string, dbCfg Config) (_ *sql.DB, cleanup func(), err error) { //nolint:gocyclo,funlen,gocognit,cyclop // Not sure is it make sense to split.
	onErr := func(f func()) {
		if err != nil {
			f()
		}
	}

	db, err := sql.Open("pqx", dbCfg.FormatDSN())
	if err != nil {
		return nil, nil, err
	}
	closeDB := func() {
		if err := db.Close(); err != nil {
			log.Print("failed to close db: ", err)
		}
	}
	defer onErr(closeDB)

	err = db.Ping()
	if err != nil {
		return nil, nil, err
	}

	if dbCfg.DBName == "" {
		dbCfg.DBName = os.Getenv("PGDATABASE")
	}
	dbCfg.DBName += "_" + suffix
	sqlDropDB := fmt.Sprintf("DROP DATABASE %s", pq.QuoteIdentifier(dbCfg.DBName))
	sqlCreateDB := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbCfg.DBName))
	if _, err = db.Exec(sqlDropDB); err != nil {
		if e := new(pq.Error); !(errors.As(err, &e) && e.Code.Name() == "invalid_catalog_name") {
			return nil, nil, fmt.Errorf("failed to drop temporary db: %w", err)
		}
	}
	if _, err := db.Exec(sqlCreateDB); err != nil {
		return nil, nil, fmt.Errorf("failed to create temporary db: %w", err)
	}
	dropTempDB := func() {
		if _, err := db.Exec(sqlDropDB); err != nil {
			log.Print("failed to drop temporary db: ", err)
		}
	}
	defer onErr(dropTempDB)

	dbTemp, err := sql.Open("pqx", dbCfg.FormatDSN())
	if err != nil {
		return nil, nil, err
	}
	closeTempDB := func() {
		if err := dbTemp.Close(); err != nil {
			log.Print("failed to close temporary db: ", err)
		}
	}
	defer onErr(closeTempDB)

	if dbCfg.User == "" {
		dbCfg.User = os.Getenv("PGUSER")
	}
	sqlCreateSchema := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION %s", pq.QuoteIdentifier(dbCfg.User))
	if _, err := dbTemp.Exec(sqlCreateSchema); err != nil {
		return nil, nil, fmt.Errorf("failed to create schema in temporary db: %w", err)
	}

	err = dbTemp.Ping()
	if err != nil {
		return nil, nil, err
	}

	cleanup = func() {
		closeTempDB()
		dropTempDB()
		closeDB()
	}
	return dbTemp, cleanup, nil
}
