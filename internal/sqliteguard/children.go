package sqliteguard

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
)

func (c *guardedConn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}
func (c *guardedConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	defer c.owner.retainPanic("prepare-panic")
	if err := c.owner.addResource(); err != nil {
		return nil, err
	}
	s, err := c.connectionAPI.PrepareContext(ctx, query)
	if err != nil {
		return nil, rejectErrorChild(c.owner, s, c.owner.resourceDone, err)
	}
	api, ok := s.(statementAPI)
	if !ok {
		return nil, rejectErrorChild(c.owner, s, c.owner.resourceDone, errors.New("SQLite statement public API changed"))
	}
	return &statement{statementAPI: api, owner: c.owner, token: c.owner.holdChild(api)}, nil
}

func (c *guardedConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	defer c.owner.retainPanic("query-panic")
	if err := c.owner.addResource(); err != nil {
		return nil, err
	}
	r, err := c.connectionAPI.Query(query, args)
	return wrapRows(c.owner, r, err, nil)
}
func (c *guardedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	defer c.owner.retainPanic("query-panic")
	if err := c.owner.addResource(); err != nil {
		return nil, err
	}
	r, err := c.connectionAPI.QueryContext(ctx, query, args)
	return wrapRows(c.owner, r, err, nil)
}

type statement struct {
	statementAPI
	owner    *lifetime
	mu       sync.Mutex
	rows     int
	closed   bool
	once     sync.Once
	closeErr error
	token    uint64
}

func (s *statement) Close() error {
	s.once.Do(func() {
		defer s.owner.retainClosePanic("statement-close-panic", &s.closeErr)
		s.mu.Lock()
		s.closed = true
		pending := s.rows
		s.mu.Unlock()
		if pending != 0 {
			s.closeErr = s.owner.retain("statement-live-rows", ErrRetained)
			return
		}
		s.closeErr = s.statementAPI.Close()
		if s.closeErr != nil {
			// The registry keeps the joined diagnostic; return the original
			// backend error unchanged to the caller of this child Close.
			_ = s.owner.retain("statement-close-failed", s.closeErr)
			return
		}
		s.owner.resourceDone()
		s.owner.dropChild(s.token)
	})
	return s.closeErr
}
func (s *statement) startRows() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("SQLite statement closed")
	}
	if err := s.owner.addResource(); err != nil {
		return err
	}
	s.rows++
	return nil
}
func (s *statement) rowsDone() { s.mu.Lock(); s.rows--; s.mu.Unlock() }
func (s *statement) Query(args []driver.Value) (driver.Rows, error) {
	defer s.owner.retainPanic("statement-query-panic")
	if err := s.startRows(); err != nil {
		return nil, err
	}
	r, err := s.statementAPI.Query(args)
	return wrapRows(s.owner, r, err, s.rowsDone)
}
func (s *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	defer s.owner.retainPanic("statement-query-panic")
	if err := s.startRows(); err != nil {
		return nil, err
	}
	r, err := s.statementAPI.QueryContext(ctx, args)
	return wrapRows(s.owner, r, err, s.rowsDone)
}

type rows struct {
	rowsAPI
	owner    *lifetime
	done     func()
	once     sync.Once
	closeErr error
	token    uint64
}

func wrapRows(o *lifetime, r driver.Rows, err error, done func()) (driver.Rows, error) {
	finish := func() {
		o.resourceDone()
		if done != nil {
			done()
		}
	}
	if err != nil {
		return nil, rejectErrorChild(o, r, finish, err)
	}
	api, ok := r.(rowsAPI)
	if !ok {
		return nil, rejectErrorChild(o, r, finish, errors.New("SQLite rows public API changed"))
	}
	return &rows{rowsAPI: api, owner: o, done: finish, token: o.holdChild(api)}, nil
}
func (r *rows) Close() error {
	r.once.Do(func() {
		defer r.owner.retainClosePanic("rows-close-panic", &r.closeErr)
		r.closeErr = r.rowsAPI.Close()
		if r.closeErr != nil {
			// Keep the original rows Close error identity at the public API.
			// The joined retention diagnostic remains observable in the registry.
			_ = r.owner.retain("rows-close-failed", r.closeErr)
			return
		}
		r.done()
		r.owner.dropChild(r.token)
	})
	return r.closeErr
}

func rejectErrorChild(o *lifetime, child io.Closer, done func(), cause error) error {
	defer o.retainPanic("error-child-close-panic")
	if child != nil {
		token := o.holdChild(child)
		if err := child.Close(); err != nil {
			return errors.Join(cause, o.retain("error-child-close-failed", err))
		}
		o.dropChild(token)
	}
	done()
	return cause
}
