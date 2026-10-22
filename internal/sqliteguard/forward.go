package sqliteguard

import (
	"context"
	"database/sql/driver"

	"modernc.org/sqlite"
)

// Calls remain on the original connection, preserving driver error identity.
// The guard owns resources before entry and quarantines a panicking backend.
func invoke[T any](o *lifetime, call func() (T, error)) (result T, err error) {
	defer o.retainPanic("backend-panic")
	if err := o.available(); err != nil {
		return result, err
	}
	return call()
}
func (c *guardedConn) Begin() (driver.Tx, error) { return invoke(c.owner, c.connectionAPI.Begin) }
func (c *guardedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return invoke(c.owner, func() (driver.Tx, error) { return c.connectionAPI.BeginTx(ctx, opts) })
}
func (c *guardedConn) ResetSession(ctx context.Context) error {
	_, err := invoke(c.owner, func() (struct{}, error) { return struct{}{}, c.connectionAPI.ResetSession(ctx) })
	return err
}
func (c *guardedConn) Ping(ctx context.Context) error {
	_, err := invoke(c.owner, func() (struct{}, error) { return struct{}{}, c.connectionAPI.Ping(ctx) })
	return err
}
func (c *guardedConn) IsValid() bool {
	defer c.owner.retainPanic("validator-panic")
	return c.owner.available() == nil && c.connectionAPI.IsValid()
}
func (c *guardedConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return invoke(c.owner, func() (driver.Result, error) { return c.connectionAPI.Exec(query, args) })
}
func (c *guardedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return invoke(c.owner, func() (driver.Result, error) { return c.connectionAPI.ExecContext(ctx, query, args) })
}
func (c *guardedConn) ColumnInfo(query string) ([]sqlite.ColumnInfo, error) {
	return invoke(c.owner, func() ([]sqlite.ColumnInfo, error) { return c.connectionAPI.ColumnInfo(query) })
}
func (c *guardedConn) Serialize() ([]byte, error) { return invoke(c.owner, c.connectionAPI.Serialize) }
func (c *guardedConn) Deserialize(data []byte) error {
	_, err := invoke(c.owner, func() (struct{}, error) { return struct{}{}, c.connectionAPI.Deserialize(data) })
	return err
}
func (c *guardedConn) IsReadOnly(schema string) (bool, error) {
	return invoke(c.owner, func() (bool, error) { return c.connectionAPI.IsReadOnly(schema) })
}
func (c *guardedConn) FileControlPersistWAL(schema string, mode int) (int, error) {
	return invoke(c.owner, func() (int, error) { return c.connectionAPI.FileControlPersistWAL(schema, mode) })
}
func (c *guardedConn) FileControlDataVersion(schema string) (uint32, error) {
	return invoke(c.owner, func() (uint32, error) { return c.connectionAPI.FileControlDataVersion(schema) })
}
func (c *guardedConn) Status(op sqlite.DBStatusOp, reset bool) (int, int, error) {
	defer c.owner.retainPanic("status-panic")
	if err := c.owner.available(); err != nil {
		return 0, 0, err
	}
	return c.connectionAPI.Status(op, reset)
}
func (c *guardedConn) RegisterPreUpdateHook(fn sqlite.PreUpdateHookFn) {
	defer c.owner.retainPanic("hook-panic")
	if err := c.owner.available(); err != nil {
		panic(err)
	}
	c.connectionAPI.RegisterPreUpdateHook(fn)
}
func (c *guardedConn) RegisterCommitHook(fn sqlite.CommitHookFn) {
	defer c.owner.retainPanic("hook-panic")
	if err := c.owner.available(); err != nil {
		panic(err)
	}
	c.connectionAPI.RegisterCommitHook(fn)
}
func (c *guardedConn) RegisterRollbackHook(fn sqlite.RollbackHookFn) {
	defer c.owner.retainPanic("hook-panic")
	if err := c.owner.available(); err != nil {
		panic(err)
	}
	c.connectionAPI.RegisterRollbackHook(fn)
}

func (s *statement) Exec(args []driver.Value) (driver.Result, error) {
	return invoke(s.owner, func() (driver.Result, error) { return s.statementAPI.Exec(args) })
}
func (s *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return invoke(s.owner, func() (driver.Result, error) { return s.statementAPI.ExecContext(ctx, args) })
}
func (r *rows) Next(dest []driver.Value) error {
	defer r.owner.retainPanic("rows-next-panic")
	return r.rowsAPI.Next(dest)
}
