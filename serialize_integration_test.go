// +build integration

package pqx_test

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"github.com/powerman/check"

	"github.com/powerman/pqx"
)

func TestSerialize(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()

	_, err := testDB.Exec(`CREATE TABLE serialize (class INT, value INT)`)
	t.Nil(err)
	_, err = testDB.Exec(`INSERT INTO serialize (class, value) VALUES (1,10),(1,20),(2,100),(2,200)`)
	t.Nil(err)

	var runs int32
	errc := make(chan error, 2)
	go func() { errc <- pqx.Serialize(func() error { atomic.AddInt32(&runs, 1); return conflict(1) }) }()
	go func() { errc <- pqx.Serialize(func() error { atomic.AddInt32(&runs, 1); return conflict(2) }) }()
	t.Nil(<-errc)
	t.Nil(<-errc)
	t.Greater(runs, 2)

	type row struct {
		Class int
		Value int
	}
	rows := []row{}
	t.Nil(testDB.Select(&rows, `SELECT class, value FROM serialize`))
	t.Equal(rows[len(rows)-1].Value, 330)
}

func conflict(class int) (err error) {
	var sum int
	tx, err := testDB.BeginTxx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err == nil {
		defer tx.Rollback()
		err = tx.Get(&sum, `SELECT SUM(value) FROM serialize WHERE class=$1`, class)
	}
	if err == nil {
		time.Sleep(testSecond / 10)
		_, err = tx.Exec(`INSERT INTO serialize (class,value) VALUES ($1,$2)`, 3-class, sum)
	}
	if err == nil {
		err = tx.Commit()
	}
	return err
}
