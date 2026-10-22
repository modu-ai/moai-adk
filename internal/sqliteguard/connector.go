package sqliteguard

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/modu-ai/moai-adk/internal/bugreport"

	"modernc.org/sqlite"
)

type connector struct {
	path    string
	base    driver.Connector
	owners  *manager
	prepare func(string, func(io.Closer)) error
}

func (c *connector) Driver() driver.Driver { return c.base.Driver() }

func (c *connector) Connect(ctx context.Context) (result driver.Conn, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	o, err := c.owners.reserve(c.path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			bugreport.Capture(bugreport.KindPanic, nil, "", nil)
			// Report and retain ownership, then propagate the original panic
			// value rather than replacing it with the diagnostic error.
			_ = o.retain("connect-panic", errors.New("driver operation panicked"))
			panic(p)
		}
	}()
	err = c.prepare(c.path, func(pin io.Closer) { o.m.mu.Lock(); o.pins = append(o.pins, pin); o.m.mu.Unlock() })
	if err != nil {
		return nil, errors.Join(err, o.release())
	}
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(err, o.release())
	}
	o.m.mu.Lock()
	o.stage = "connecting"
	o.m.mu.Unlock()
	raw, err := c.base.Connect(ctx)
	o.m.mu.Lock()
	o.conn = raw
	o.m.mu.Unlock()
	if err != nil {
		if raw == nil {
			return nil, o.retain("connect-uncertain", err)
		}
		return nil, errors.Join(err, closeKnown(o))
	}
	if raw == nil {
		return nil, o.retain("connect-uncertain", errors.New("driver returned no connection"))
	}
	api, ok := raw.(connectionAPI)
	if !ok {
		return nil, errors.Join(errors.New("SQLite connection public API changed"), closeKnown(o))
	}
	mode, err := api.FileControlPersistWAL("main", 1)
	if err == nil && mode != 1 {
		err = fmt.Errorf("PERSIST_WAL set returned %d", mode)
	}
	if err == nil {
		mode, err = api.FileControlPersistWAL("main", -1)
		if err == nil && mode != 1 {
			err = fmt.Errorf("PERSIST_WAL readback returned %d", mode)
		}
	}
	if err != nil {
		return nil, errors.Join(err, closeKnown(o))
	}
	o.m.mu.Lock()
	o.stage = "live"
	o.m.mu.Unlock()
	return &guardedConn{connectionAPI: api, owner: o}, nil
}

func closeKnown(o *lifetime) error {
	if err := o.conn.Close(); err != nil {
		return o.retain("close-failed", err)
	}
	return o.release()
}

type guardedConn struct {
	connectionAPI
	owner    *lifetime
	once     sync.Once
	closeErr error
}

func (c *guardedConn) Close() error {
	c.once.Do(func() {
		defer func() {
			if p := recover(); p != nil {
				bugreport.Capture(bugreport.KindPanic, nil, "", nil)
				c.closeErr = c.owner.retain("close-panic", errors.New("physical Close panicked"))
				panic(p)
			}
		}()
		// modernc frees its TLS after close_v2 succeeds, even if SQLite
		// would still have zombie children. Do not enter Close with raw
		// children or an opaque Backup lifetime outstanding.
		if !c.owner.beginClose() {
			c.closeErr = c.owner.retain("raw-resources", ErrRetained)
			return
		}
		c.closeErr = closeKnown(c.owner)
	})
	return c.closeErr
}

func (c *guardedConn) NewBackup(uri string) (*sqlite.Backup, error) {
	defer c.owner.retainPanic("backup-panic")
	c.owner.m.mu.Lock()
	c.owner.sticky = true
	c.owner.m.mu.Unlock()
	// Status/Close expose sticky retention; preserve the original
	// concrete Backup result and error from this API call.
	_ = c.owner.retain("raw-backup", ErrRetained)
	return c.connectionAPI.NewBackup(uri)
}
func (c *guardedConn) NewRestore(uri string) (*sqlite.Backup, error) {
	defer c.owner.retainPanic("restore-panic")
	c.owner.m.mu.Lock()
	c.owner.sticky = true
	c.owner.m.mu.Unlock()
	// Status/Close expose sticky retention while this call returns the
	// original concrete Restore result and error.
	_ = c.owner.retain("raw-restore", ErrRetained)
	return c.connectionAPI.NewRestore(uri)
}
