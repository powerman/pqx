// Package pqx provides helpers for use with github.com/lib/pq (Postgres
// driver for the database/sql package).
//
// It also registers own driver for the database/sql package named "pqx".
// This driver is a thin wrapper around github.com/lib/pq driver used to
// make it easier to control connections opened by github.com/lib/pq.
// By default it'll just enable TCP keepalives, but you can set Dial to
// your own hook and have full control over opened connections.
package pqx
