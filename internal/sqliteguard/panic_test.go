package sqliteguard

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
)

type panicPrepareConn struct{ *fakeConn }

func (*panicPrepareConn) PrepareContext(context.Context, string) (driver.Stmt, error) {
	panic("prepare uncertain")
}

func TestPreparePanicReportsFaultBeforePoolClose(t *testing.T) {
	physical := &panicPrepareConn{fakeConn: &fakeConn{}}
	c, pin, m, _ := testConnector(t, physical, nil, 1)
	conn, err := c.Connect(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic lost")
			}
		}()
		_, _ = conn.Prepare("SELECT 1")
	}()
	if m.snapshots()[0].Stage != "prepare-panic" || !m.snapshots()[0].Faulted {
		t.Fatalf("panic not reported: %+v", m.snapshots())
	}
	if err := conn.Close(); !errors.Is(err, ErrRetained) {
		t.Fatalf("Close=%v", err)
	}
	if physical.closes.Load() != 0 || pin.closes.Load() != 0 {
		t.Fatal("uncertain prepared resources were closed")
	}
}

type panicRows struct {
	rowsAPI
	closes int
}

func (r *panicRows) Close() error { r.closes++; panic("row Close uncertain") }

func TestChildClosePanicPreservesFirstOutcomeAndPin(t *testing.T) {
	c, pin, m, _ := testConnector(t, &fakeConn{}, nil, 1)
	dc, err := c.Connect(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	owner := dc.(*guardedConn).owner
	if err := owner.addResource(); err != nil {
		t.Fatal(err)
	}
	physical := &panicRows{}
	r, err := wrapRows(owner, physical, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic lost")
			}
		}()
		_ = r.Close()
	}()
	if err := r.Close(); !errors.Is(err, ErrRetained) {
		t.Fatalf("second Close falsely succeeded: %v", err)
	}
	if physical.closes != 1 || pin.closes.Load() != 0 || !m.snapshots()[0].Faulted {
		t.Fatalf("retry/release after panic: %+v", m.snapshots())
	}
}
