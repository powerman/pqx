package pqx_test

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/powerman/gotest/testinit"

	"github.com/powerman/pqx"
)

const testDBSuffix = "github.com/powerman/pqx"

var testDB *sqlx.DB

func init() { testinit.Setup(2, setupIntegration) }

func setupIntegration() {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	if os.Getenv("GO_INTEGRATION_TEST") == "" {
		logger.Print("skipping integration tests; set GO_INTEGRATION_TEST to run them")
		return
	}
	db, cleanup, err := pqx.EnsureTempDB(logger, testDBSuffix, pqx.Config{ConnectTimeout: 3 * time.Second})
	if err != nil {
		var e *pq.Error
		if !errors.As(err, &e) || e.Code.Class().Name() == "invalid_authorization_specification" {
			logger.Print("set environment variables to allow connection to postgresql:\nhttps://www.postgresql.org/docs/current/libpq-envars.html")
		}
		testinit.Fatal(err)
	}
	testinit.Teardown(cleanup)
	testDB = sqlx.NewDb(db, "pqx")
}
