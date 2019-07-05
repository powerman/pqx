package pqx

import (
	"database/sql"
	"fmt"

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
// Recommended value for suffix is your package's import path.
func EnsureTempDB(log Logger, suffix string, dbCfg Config) (_ *sql.DB, cleanup func(), err error) { //nolint:gocyclo
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

	dbCfg.DBName += "_" + suffix
	sqlDropDB := fmt.Sprintf("DROP DATABASE %s", pq.QuoteIdentifier(dbCfg.DBName))
	sqlCreateDB := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbCfg.DBName))
	if _, err = db.Exec(sqlDropDB); err != nil {
		if e, ok := err.(*pq.Error); !(ok && e.Code.Name() == "invalid_catalog_name") {
			return nil, nil, fmt.Errorf("failed to drop temporary db: %v", err)
		}
	}
	if _, err := db.Exec(sqlCreateDB); err != nil {
		return nil, nil, fmt.Errorf("failed to create temporary db: %v", err)
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
