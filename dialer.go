package pqx

import (
	"database/sql"
	"database/sql/driver"
	"net"
	"time"

	"github.com/lib/pq"
)

//nolint:gochecknoinits // By design.
func init() {
	sql.Register("pqx", &sqlDriver{})
}

type sqlDriver struct{}

// Open returns a new SQL driver connection using Dial hook.
func (*sqlDriver) Open(name string) (driver.Conn, error) {
	return pq.DialOpen(&dialer{}, name)
}

type dialer struct{}

// Dial implements pq.Dialer interface.
func (dialer) Dial(network, address string) (net.Conn, error) {
	return Dial(network, address, 0) //nolint:revive // False positive.
}

// DialTimeout implements pq.Dialer interface.
func (dialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return Dial(network, address, timeout)
}

// DialFunc used to open new connections to PostgreSQL server.
type DialFunc func(network, address string, timeout time.Duration) (net.Conn, error)

// Dial is a hook which should be set before connecting to PostgreSQL
// server using "pqx" driver.
//
//nolint:gochecknoglobals // By design.
var Dial = KeepAliveDial(time.Minute)

// KeepAliveDial returns hook which adds TCP keepalives.
func KeepAliveDial(keepAlivePeriod time.Duration) DialFunc {
	return func(network, address string, timeout time.Duration) (net.Conn, error) {
		dialer := &net.Dialer{
			KeepAlive: keepAlivePeriod,
			Timeout:   timeout,
		}
		return dialer.Dial(network, address)
	}
}
