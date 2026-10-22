package sqliteguard

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
)

type errorStatement struct {
	statementAPI
	closes   int
	closeErr error
}

func (s *errorStatement) Close() error { s.closes++; return s.closeErr }

type errorRows struct {
	rowsAPI
	closes   int
	closeErr error
}

func (r *errorRows) Close() error { r.closes++; return r.closeErr }

type errorChildConn struct {
	*fakeConn
	stmt     driver.Stmt
	rows     driver.Rows
	childErr error
}

func (c *errorChildConn) PrepareContext(context.Context, string) (driver.Stmt, error) {
	return c.stmt, c.childErr
}
func (c *errorChildConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return c.rows, c.childErr
}

func TestNonnullErrorChildIsClosedBeforeAccountingRelease(t *testing.T) {
	for _, kind := range []string{"statement", "rows"} {
		t.Run(kind, func(t *testing.T) {
			want := errors.New("original child error")
			s := &errorStatement{}
			r := &errorRows{}
			physical := &errorChildConn{fakeConn: &fakeConn{}, stmt: s, rows: r, childErr: want}
			c, pin, m, _ := testConnector(t, physical, nil, 1)
			dc, err := c.Connect(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			api := dc.(connectionAPI)
			if kind == "statement" {
				_, err = api.PrepareContext(t.Context(), "synthetic")
			} else {
				_, err = api.QueryContext(t.Context(), "synthetic", nil)
			}
			if !errors.Is(err, want) {
				t.Fatalf("original error lost: %v", err)
			}
			closed := s.closes
			if kind == "rows" {
				closed = r.closes
			}
			if closed != 1 {
				t.Errorf("nonnull error child Close count=%d, want 1", closed)
			}
			if m.snapshots()[0].Outstanding != 0 {
				t.Errorf("successfully closed child still accounted: %+v", m.snapshots())
			}
			if err := dc.Close(); err != nil {
				t.Fatal(err)
			}
			if pin.closes.Load() != 1 {
				t.Fatal("confirmed termination did not release pin")
			}
		})
	}
}

func TestNonnullErrorChildCloseFailureRetainsOwner(t *testing.T) {
	for _, kind := range []string{"statement", "rows"} {
		t.Run(kind, func(t *testing.T) {
			want := errors.New("original child error")
			closeErr := errors.New("uncertain child Close")
			s := &errorStatement{closeErr: closeErr}
			r := &errorRows{closeErr: closeErr}
			physical := &errorChildConn{fakeConn: &fakeConn{}, stmt: s, rows: r, childErr: want}
			c, pin, m, _ := testConnector(t, physical, nil, 1)
			dc, err := c.Connect(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			api := dc.(connectionAPI)
			if kind == "statement" {
				_, err = api.PrepareContext(t.Context(), "synthetic")
			} else {
				_, err = api.QueryContext(t.Context(), "synthetic", nil)
			}
			if !errors.Is(err, want) || !errors.Is(err, closeErr) {
				t.Errorf("error identity not preserved: %v", err)
			}
			if err := dc.Close(); !errors.Is(err, ErrRetained) {
				t.Errorf("uncertain child physical Close falsely succeeded: %v", err)
			}
			if physical.closes.Load() != 0 || pin.closes.Load() != 0 || !m.snapshots()[0].Faulted {
				t.Errorf("uncertain child owner lost: %+v", m.snapshots())
			}
		})
	}
}
