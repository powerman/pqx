package pqx_test

import (
	"context"
	"database/sql"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/powerman/check"

	"github.com/powerman/pqx"
)

func TestSerialize(tt *testing.T) {
	if os.Getenv("GO_INTEGRATION_TEST") == "" {
		tt.Skip("skipping integration test; set GO_INTEGRATION_TEST to run it")
	}
	tt.Parallel()
	t := check.T(tt)

	_, err := testDB.ExecContext(t.Context(), `CREATE TABLE serialize (class INT, value INT)`)
	t.Nil(err)
	_, err = testDB.ExecContext(t.Context(), `INSERT INTO serialize (class, value) VALUES (1,10),(1,20),(2,100),(2,200)`)
	t.Nil(err)

	var runs int32
	errc := make(chan error, 2)
	go func() {
		errc <- pqx.Serialize(func() error { atomic.AddInt32(&runs, 1); return conflict(t.Context(), 1) })
	}()
	go func() {
		errc <- pqx.Serialize(func() error { atomic.AddInt32(&runs, 1); return conflict(t.Context(), 2) })
	}()
	t.Nil(<-errc)
	t.Nil(<-errc)
	t.Greater(runs, 2)

	type row struct {
		Class int `db:"class"`
		Value int `db:"value"`
	}
	rows := []row{}
	t.Nil(testDB.Select(&rows, `SELECT class, value FROM serialize`))
	t.Equal(rows[len(rows)-1].Value, 330)
}

func conflict(ctx context.Context, class int) (err error) {
	var sum int
	tx, err := testDB.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err == nil {
		defer tx.Rollback()
		err = tx.Get(&sum, `SELECT SUM(value) FROM serialize WHERE class=$1`, class)
	}
	if err == nil {
		time.Sleep(time.Second / 10)
		_, err = tx.ExecContext(ctx, `INSERT INTO serialize (class,value) VALUES ($1,$2)`, 3-class, sum)
	}
	if err == nil {
		err = tx.Commit()
	}
	return err
}
