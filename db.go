package pqx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/lib/pq"
)

// Logger interface consumed by this package.
type Logger interface {
	Print(v ...any)
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
func EnsureTempDB(log Logger, suffix string, dbCfg Config) (_ *sql.DB, cleanup func(), err error) { //nolint:gocyclo,funlen,cyclop // Not sure is it make sense to split.
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
		err := db.Close()
		if err != nil {
			log.Print("failed to close db: ", err)
		}
	}
	defer onErr(closeDB)

	err = db.PingContext(context.Background())
	if err != nil {
		return nil, nil, err
	}

	if dbCfg.DBName == "" {
		dbCfg.DBName = os.Getenv("PGDATABASE")
	}
	dbCfg.DBName += "_" + suffix
	sqlDropDB := fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", pq.QuoteIdentifier(dbCfg.DBName))
	sqlCreateDB := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbCfg.DBName))
	_, err = db.ExecContext(context.Background(), sqlDropDB) //nolint:gosec // pq.QuoteIdentifier ensures safe identifier quoting.
	if err != nil {
		e := new(pq.Error)
		if !errors.As(err, &e) || e.Code.Name() != "invalid_catalog_name" {
			return nil, nil, fmt.Errorf("failed to drop temporary db: %w", err)
		}
	}
	_, err = db.ExecContext(context.Background(), sqlCreateDB) //nolint:gosec // pq.QuoteIdentifier ensures safe identifier quoting.
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temporary db: %w", err)
	}
	dropTempDB := func() {
		_, e := db.ExecContext(context.Background(), sqlDropDB) //nolint:gosec // pq.QuoteIdentifier ensures safe identifier quoting.
		if e != nil {
			log.Print("failed to drop temporary db: ", e)
		}
	}
	defer onErr(dropTempDB)

	dbTemp, err := sql.Open("pqx", dbCfg.FormatDSN())
	if err != nil {
		return nil, nil, err
	}
	closeTempDB := func() {
		err := dbTemp.Close()
		if err != nil {
			log.Print("failed to close temporary db: ", err)
		}
	}
	defer onErr(closeTempDB)

	if dbCfg.User == "" {
		dbCfg.User = os.Getenv("PGUSER")
	}
	sqlCreateSchema := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION %s", pq.QuoteIdentifier(dbCfg.User))
	_, err = dbTemp.ExecContext(context.Background(), sqlCreateSchema) //nolint:gosec // pq.QuoteIdentifier ensures safe identifier quoting.
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create schema in temporary db: %w", err)
	}

	err = dbTemp.PingContext(context.Background())
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
