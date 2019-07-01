package pqx

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SSLMode determines whether or with what priority a secure SSL TCP/IP
// connection will be negotiated with the server..
type SSLMode string

// SSL modes.
const (
	SSLDisable    SSLMode = "disable"     // Only try a non-SSL connection.
	SSLRequire    SSLMode = "require"     // Only try an SSL connection. If a root CA file is present, verify the certificate in the same way as if verify-ca was specified.
	SSLVerifyCA   SSLMode = "verify-ca"   // Only try an SSL connection, and verify that the server certificate is issued by a trusted certificate authority (CA).
	SSLVerifyFull SSLMode = "verify-full" // Only try an SSL connection, verify that the server certificate is issued by a trusted CA and that the requested server host name matches that in the certificate.
)

// Config described connection parameters for github.com/lib/pq.
type Config struct {
	DBName                          string
	User                            string
	Pass                            string
	Host                            string
	Port                            int
	FallbackApplicationName         string
	ConnectTimeout                  time.Duration // Round to seconds.
	SSLMode                         SSLMode
	SSLCert                         string             // PEM file path.
	SSLKey                          string             // PEM file path.
	SSLRootCert                     string             // PEM file path.
	SearchPath                      string             // Specifies the order in which schemas are searched.
	DefaultTransactionIsolation     sql.IsolationLevel // One of: LevelDefault, LevelReadUncommitted, LevelReadCommitted, LevelRepeatableRead, LevelSerializable.
	StatementTimeout                time.Duration      // Round to milliseconds.
	LockTimeout                     time.Duration      // Round to milliseconds.
	IdleInTransactionSessionTimeout time.Duration      // Round to milliseconds.
	Other                           map[string]string  // Any other parameters from https://www.postgresql.org/docs/current/runtime-config-client.html.
}

// FormatDSN returns dataSourceName string with properly escaped
// connection parameters suitable for sql.Open.
func (c Config) FormatDSN() string {
	// Borrowed from pq.ParseURL.
	var kvs []string
	escaper := strings.NewReplacer(` `, `\ `, `'`, `\'`, `\`, `\\`)
	accrue := func(k, v string) {
		if v != "" {
			kvs = append(kvs, k+"="+escaper.Replace(v))
		}
	}

	accrue("dbname", c.DBName)
	accrue("user", c.User)
	accrue("password", c.Pass)
	accrue("host", c.Host)
	if c.Port > 0 {
		accrue("port", strconv.Itoa(c.Port))
	}
	accrue("fallback_application_name", c.FallbackApplicationName)
	accrue("connect_timeout", timeoutSeconds(c.ConnectTimeout))
	accrue("sslmode", string(c.SSLMode))
	accrue("sslcert", c.SSLCert)
	accrue("sslkey", c.SSLKey)
	accrue("sslrootcert", c.SSLRootCert)
	accrue("search_path", c.SearchPath)
	switch c.DefaultTransactionIsolation {
	case sql.LevelDefault:
	case sql.LevelReadUncommitted:
		accrue("default_transaction_isolation", "read uncommitted")
	case sql.LevelReadCommitted:
		accrue("default_transaction_isolation", "read committed")
	case sql.LevelRepeatableRead:
		accrue("default_transaction_isolation", "repeatable read")
	case sql.LevelSerializable:
		accrue("default_transaction_isolation", "serializable")
	default:
		panic(fmt.Sprintf("invalid DefaultTransactionIsolation: %s", c.DefaultTransactionIsolation))
	}
	accrue("statement_timeout", timeoutMilliseconds(c.StatementTimeout))
	accrue("lock_timeout", timeoutMilliseconds(c.LockTimeout))
	accrue("idle_in_transaction_session_timeout", timeoutMilliseconds(c.IdleInTransactionSessionTimeout))

	customPos := len(kvs)
	for k, v := range c.Other {
		accrue(k, v)
	}
	sort.Strings(kvs[customPos:]) // For testing.

	return strings.Join(kvs, " ")
}

func timeoutSeconds(t time.Duration) string {
	switch {
	case t == 0:
		return ""
	case 0 < t && t < time.Second:
		return "1"
	default:
		return strconv.Itoa(int(t.Round(time.Second).Seconds()))
	}
}

func timeoutMilliseconds(t time.Duration) string {
	switch {
	case t == 0:
		return ""
	case 0 < t && t < time.Millisecond:
		return "1"
	default:
		return strconv.Itoa(int(t.Round(time.Millisecond) / time.Millisecond))
	}
}
