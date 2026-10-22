// Package sqliteguard owns native privacy pins across physical SQLite lifetimes.
package sqliteguard

import (
	"database/sql/driver"

	"modernc.org/sqlite"
)

// Keep the driver's declared public connection methods, including its Raw
// extensions. Do not add NamedValueChecker, which modernc does not implement.
type connectionAPI interface {
	driver.Conn
	driver.SessionResetter
	driver.Validator
	driver.Execer
	driver.Queryer
	driver.Pinger
	driver.ConnBeginTx
	driver.ConnPrepareContext
	driver.ExecerContext
	driver.QueryerContext
	sqlite.FileControl
	ColumnInfo(string) ([]sqlite.ColumnInfo, error)
	Serialize() ([]byte, error)
	Deserialize([]byte) error
	NewBackup(string) (*sqlite.Backup, error)
	NewRestore(string) (*sqlite.Backup, error)
	IsReadOnly(string) (bool, error)
	Status(sqlite.DBStatusOp, bool) (int, int, error)
	RegisterPreUpdateHook(sqlite.PreUpdateHookFn)
	RegisterCommitHook(sqlite.CommitHookFn)
	RegisterRollbackHook(sqlite.RollbackHookFn)
}

type statementAPI interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

type rowsAPI interface {
	driver.Rows
	driver.RowsColumnTypeDatabaseTypeName
	driver.RowsColumnTypeLength
	driver.RowsColumnTypeNullable
	driver.RowsColumnTypePrecisionScale
	driver.RowsColumnTypeScanType
}
