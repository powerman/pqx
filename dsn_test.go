package pqx_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/powerman/check"

	"github.com/powerman/pqx"
)

func TestConfig(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()

	testCases := []struct {
		cfg       pqx.Config
		wantDSN   string
		wantURL   string
		wantPanic string
	}{
		{
			pqx.Config{},
			``,
			`postgres:///`,
			``,
		},
		{
			pqx.Config{
				DBName: "db", User: "u", Pass: "p", Host: "h", Port: 1,
				FallbackApplicationName: "a", ConnectTimeout: 2 * time.Second,
				SSLMode: pqx.SSLVerifyFull, SSLCert: "crt", SSLKey: "key", SSLRootCert: "ca",
				SearchPath:                      "public",
				DefaultTransactionIsolation:     sql.LevelSerializable,
				StatementTimeout:                1 * time.Second,
				LockTimeout:                     2 * time.Second,
				IdleInTransactionSessionTimeout: 3 * time.Second,
				Other:                           map[string]string{"a": "A", "b": "B"},
			},
			`dbname=db user=u password=p host=h port=1 ` +
				`fallback_application_name=a connect_timeout=2 ` +
				`sslmode=verify-full sslcert=crt sslkey=key sslrootcert=ca ` +
				`search_path=public default_transaction_isolation=serializable ` +
				`statement_timeout=1000 lock_timeout=2000 ` +
				`idle_in_transaction_session_timeout=3000 ` +
				`a=A b=B`,
			`postgres://u:p@h:1/db?a=A&b=B&connect_timeout=2&` +
				`default_transaction_isolation=serializable&` +
				`fallback_application_name=a&` +
				`idle_in_transaction_session_timeout=3000&lock_timeout=2000&` +
				`search_path=public&sslcert=crt&sslkey=key&sslmode=verify-full&` +
				`sslrootcert=ca&statement_timeout=1000`,
			``,
		},
		{
			pqx.Config{Pass: `' very \special `},
			`password=\'\ very\ \\special\ `,
			`postgres://:%27%20very%20%5Cspecial%20@/`,
			``,
		},
		{
			pqx.Config{ConnectTimeout: time.Second * 3 / 2},
			`connect_timeout=2`,
			`postgres:///?connect_timeout=2`,
			``,
		},
		{
			pqx.Config{ConnectTimeout: time.Second / 2},
			`connect_timeout=1`,
			`postgres:///?connect_timeout=1`,
			``,
		},
		{
			pqx.Config{ConnectTimeout: time.Second / 10},
			`connect_timeout=1`,
			`postgres:///?connect_timeout=1`,
			``,
		},
		{
			pqx.Config{SearchPath: `"$user", public`},
			`search_path="$user",\ public`,
			`postgres:///?search_path=%22%24user%22%2C+public`,
			``,
		},
		{
			pqx.Config{DefaultTransactionIsolation: sql.LevelDefault},
			``,
			`postgres:///`,
			``,
		},
		{
			pqx.Config{DefaultTransactionIsolation: sql.LevelReadUncommitted},
			`default_transaction_isolation=read\ uncommitted`,
			`postgres:///?default_transaction_isolation=read+uncommitted`,
			``,
		},
		{
			pqx.Config{DefaultTransactionIsolation: sql.LevelReadCommitted},
			`default_transaction_isolation=read\ committed`,
			`postgres:///?default_transaction_isolation=read+committed`,
			``,
		},
		{
			pqx.Config{DefaultTransactionIsolation: sql.LevelWriteCommitted},
			``, ``, `invalid.*Write Committed`,
		},
		{
			pqx.Config{DefaultTransactionIsolation: sql.LevelRepeatableRead},
			`default_transaction_isolation=repeatable\ read`,
			`postgres:///?default_transaction_isolation=repeatable+read`,
			``,
		},
		{
			pqx.Config{DefaultTransactionIsolation: sql.LevelSnapshot},
			``, ``, `invalid.*Snapshot`,
		},
		{
			pqx.Config{DefaultTransactionIsolation: sql.LevelSerializable},
			`default_transaction_isolation=serializable`,
			`postgres:///?default_transaction_isolation=serializable`,
			``,
		},
		{
			pqx.Config{DefaultTransactionIsolation: sql.LevelLinearizable},
			``, ``, `invalid.*Linearizable`,
		},
		{
			pqx.Config{StatementTimeout: time.Millisecond * 3 / 2},
			`statement_timeout=2`,
			`postgres:///?statement_timeout=2`,
			``,
		},
		{
			pqx.Config{StatementTimeout: time.Millisecond / 2},
			`statement_timeout=1`,
			`postgres:///?statement_timeout=1`,
			``,
		},
		{
			pqx.Config{LockTimeout: time.Millisecond * 3 / 2},
			`lock_timeout=2`,
			`postgres:///?lock_timeout=2`,
			``,
		},
		{
			pqx.Config{LockTimeout: time.Millisecond / 2},
			`lock_timeout=1`,
			`postgres:///?lock_timeout=1`,
			``,
		},
		{
			pqx.Config{IdleInTransactionSessionTimeout: time.Millisecond * 3 / 2},
			`idle_in_transaction_session_timeout=2`,
			`postgres:///?idle_in_transaction_session_timeout=2`,
			``,
		},
		{
			pqx.Config{IdleInTransactionSessionTimeout: time.Millisecond / 2},
			`idle_in_transaction_session_timeout=1`,
			`postgres:///?idle_in_transaction_session_timeout=1`,
			``,
		},
		{
			pqx.Config{Other: map[string]string{"a": ""}},
			``,
			`postgres:///`,
			``,
		},
		{
			pqx.Config{Other: map[string]string{"a": " "}},
			`a=\ `,
			`postgres:///?a=+`,
			``,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run("", func(tt *testing.T) {
			t := check.T(tt)
			if tc.wantPanic != "" {
				t.PanicMatch(func() { tc.cfg.FormatDSN() }, tc.wantPanic)
			} else {
				t.Equal(tc.cfg.FormatDSN(), tc.wantDSN)
				t.Equal(tc.cfg.FormatURL(), tc.wantURL)
			}
		})
	}
}
