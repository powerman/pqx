// +build integration

package pqx

import (
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/powerman/gotest/testinit"
)

const testDBSuffix = "github.com/powerman/pqx"

var testDB *sqlx.DB

func init() { testinit.Setup(2, setupIntegration) }

func setupIntegration() {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	db, cleanup, err := EnsureTempDB(logger, testDBSuffix, Config{ConnectTimeout: 3 * time.Second})
	if err != nil {
		if err, ok := err.(*pq.Error); !ok || err.Code.Class().Name() == "invalid_authorization_specification" {
			logger.Print("set environment variables to allow connection to postgresql:\nhttps://www.postgresql.org/docs/current/libpq-envars.html")
		}
		testinit.Fatal(err)
	}
	testinit.Teardown(cleanup)
	testDB = sqlx.NewDb(db, "pqx")
}
