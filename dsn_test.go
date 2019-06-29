package pqx

import (
	"database/sql"
	"testing"
	"time"

	"github.com/powerman/check"
)

func TestConfig(tt *testing.T) {
	t := check.T(tt)
	t.Parallel()

	testCases := []struct {
		cfg       Config
		want      string
		wantPanic string
	}{
		{Config{}, ``, ``},
		{Config{
			DBName: "db", User: "u", Pass: "p", Host: "h", Port: 1,
			FallbackApplicationName: "a", ConnectTimeout: 2 * time.Second,
			SSLMode: SSLVerifyFull, SSLCert: "crt", SSLKey: "key", SSLRootCert: "ca",
			DefaultTransactionIsolation:     sql.LevelSerializable,
			StatementTimeout:                1 * time.Second,
			LockTimeout:                     2 * time.Second,
			IdleInTransactionSessionTimeout: 3 * time.Second,
			Other:                           map[string]string{"a": "A", "b": "B"},
		}, `dbname=db user=u password=p host=h port=1 ` +
			`fallback_application_name=a connect_timeout=2 ` +
			`sslmode=verify-full sslcert=crt sslkey=key sslrootcert=ca ` +
			`default_transaction_isolation=serializable ` +
			`statement_timeout=1000 lock_timeout=2000 ` +
			`idle_in_transaction_session_timeout=3000 ` +
			`a=A b=B`, ``},
		{Config{Pass: `' very \special `},
			`password=\'\ very\ \\special\ `, ``},
		{Config{ConnectTimeout: time.Second * 3 / 2},
			`connect_timeout=2`, ``},
		{Config{ConnectTimeout: time.Second / 2},
			`connect_timeout=1`, ``},
		{Config{ConnectTimeout: time.Second / 10},
			`connect_timeout=1`, ``},
		{Config{DefaultTransactionIsolation: sql.LevelDefault},
			``, ``},
		{Config{DefaultTransactionIsolation: sql.LevelReadUncommitted},
			`default_transaction_isolation=read\ uncommitted`, ``},
		{Config{DefaultTransactionIsolation: sql.LevelReadCommitted},
			`default_transaction_isolation=read\ committed`, ``},
		{Config{DefaultTransactionIsolation: sql.LevelWriteCommitted},
			``, `invalid.*Write Committed`},
		{Config{DefaultTransactionIsolation: sql.LevelRepeatableRead},
			`default_transaction_isolation=repeatable\ read`, ``},
		{Config{DefaultTransactionIsolation: sql.LevelSnapshot},
			``, `invalid.*Snapshot`},
		{Config{DefaultTransactionIsolation: sql.LevelSerializable},
			`default_transaction_isolation=serializable`, ``},
		{Config{DefaultTransactionIsolation: sql.LevelLinearizable},
			``, `invalid.*Linearizable`},
		{Config{StatementTimeout: time.Millisecond * 3 / 2},
			`statement_timeout=2`, ``},
		{Config{StatementTimeout: time.Millisecond / 2},
			`statement_timeout=1`, ``},
		{Config{LockTimeout: time.Millisecond * 3 / 2},
			`lock_timeout=2`, ``},
		{Config{LockTimeout: time.Millisecond / 2},
			`lock_timeout=1`, ``},
		{Config{IdleInTransactionSessionTimeout: time.Millisecond * 3 / 2},
			`idle_in_transaction_session_timeout=2`, ``},
		{Config{IdleInTransactionSessionTimeout: time.Millisecond / 2},
			`idle_in_transaction_session_timeout=1`, ``},
		{Config{Other: map[string]string{"a": ""}},
			``, ``},
		{Config{Other: map[string]string{"a": " "}},
			`a=\ `, ``},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run("", func(tt *testing.T) {
			t := check.T(tt)
			if tc.wantPanic != "" {
				t.PanicMatch(func() { tc.cfg.FormatDSN() }, tc.wantPanic)
			} else {
				t.Equal(tc.cfg.FormatDSN(), tc.want)
			}
		})
	}
}
