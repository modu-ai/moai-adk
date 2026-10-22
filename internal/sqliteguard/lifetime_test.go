package sqliteguard

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

type fakePin struct{ closes atomic.Int32 }

func (p *fakePin) Close() error { p.closes.Add(1); return nil }

type fakeConn struct {
	connectionAPI
	closeErr error
	closes   atomic.Int32
	mode     int
}

func (c *fakeConn) Close() error { c.closes.Add(1); return c.closeErr }
func (c *fakeConn) FileControlPersistWAL(_ string, mode int) (int, error) {
	if mode >= 0 {
		c.mode = mode
	}
	return c.mode, nil
}

type fakeConnector struct {
	conn      driver.Conn
	err       error
	calls     atomic.Int32
	panicOpen bool
}

func (c *fakeConnector) Connect(context.Context) (driver.Conn, error) {
	c.calls.Add(1)
	if c.panicOpen {
		panic("synthetic driver panic")
	}
	return c.conn, c.err
}

func TestDriverPanicRetainsOwnedPins(t *testing.T) {
	c, pin, m, base := testConnector(t, nil, nil, 1)
	base.panicOpen = true
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic lost")
			}
		}()
		_, _ = c.Connect(t.Context())
	}()
	runtime.GC()
	if pin.closes.Load() != 0 || len(m.snapshots()) != 1 || m.snapshots()[0].Stage != "connect-panic" {
		t.Fatalf("panic owner=%+v", m.snapshots())
	}
}

func TestConcurrentAdmissionNeverExceedsReservedCapacity(t *testing.T) {
	m := newManager(3, func(Snapshot) {})
	var wg sync.WaitGroup
	var admitted atomic.Int32
	for i := range 24 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o, err := m.reserve(string(rune('a' + i)))
			if err == nil {
				admitted.Add(1)
				o.retain("synthetic", ErrRetained)
			} else if !errors.Is(err, ErrCapacity) {
				t.Errorf("admission=%v", err)
			}
		}()
	}
	wg.Wait()
	if admitted.Load() != 3 || len(m.snapshots()) != 3 {
		t.Fatalf("admitted=%d owners=%+v", admitted.Load(), m.snapshots())
	}
}
func (*fakeConnector) Driver() driver.Driver { return nil }

func testConnector(t *testing.T, conn driver.Conn, connectErr error, limit int) (*connector, *fakePin, *manager, *fakeConnector) {
	t.Helper()
	pin := &fakePin{}
	m := newManager(limit, func(Snapshot) {})
	base := &fakeConnector{conn: conn, err: connectErr}
	c := &connector{path: "test-canonical.db", base: base, owners: m, prepare: func(_ string, acquired func(io.Closer)) error { acquired(pin); return nil }}
	return c, pin, m, base
}

func TestCloseFailureRetainsBeyondGCAndRefusesReentry(t *testing.T) {
	want := errors.New("physical close uncertain")
	baseConn := &fakeConn{closeErr: want}
	c, pin, m, base := testConnector(t, baseConn, nil, 2)
	conn, err := c.Connect(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); !errors.Is(err, want) {
		t.Fatalf("Close error=%v, want original cause", err)
	}
	conn = nil
	runtime.GC()
	if pin.closes.Load() != 0 {
		t.Fatal("uncertain physical lifetime lost its pin")
	}
	if len(m.snapshots()) != 1 || m.snapshots()[0].Stage != "close-failed" {
		t.Fatalf("owner status=%+v", m.snapshots())
	}
	if _, err := c.Connect(t.Context()); !errors.Is(err, ErrFaulted) {
		t.Fatalf("fault admission=%v", err)
	}
	if base.calls.Load() != 1 {
		t.Fatal("faulted path entered driver again")
	}
	if baseConn.closes.Load() != 1 {
		t.Fatal("implicit Close retry")
	}
}

func TestOpaqueConnectFailureRetainsAndBudgetPrecedesAcquisition(t *testing.T) {
	want := errors.New("opaque open failed")
	c, pin, m, _ := testConnector(t, nil, want, 1)
	if _, err := c.Connect(t.Context()); !errors.Is(err, want) {
		t.Fatalf("Connect=%v", err)
	}
	if pin.closes.Load() != 0 {
		t.Fatal("opaque driver failure released pin")
	}
	if m.snapshots()[0].Stage != "connect-uncertain" {
		t.Fatalf("status=%+v", m.snapshots())
	}
	other := *c
	other.path = "different.db"
	var prepared atomic.Int32
	other.prepare = func(string, func(io.Closer)) error { prepared.Add(1); return nil }
	if _, err := other.Connect(t.Context()); !errors.Is(err, ErrCapacity) {
		t.Fatalf("capacity=%v", err)
	}
	if prepared.Load() != 0 {
		t.Fatal("capacity checked after native acquisition")
	}
}

func TestNormalCloseIsOnceOnlyAndReleasesAfterDriver(t *testing.T) {
	fc := &fakeConn{}
	c, pin, m, _ := testConnector(t, fc, nil, 2)
	conn, err := c.Connect(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if fc.mode != 1 {
		t.Fatal("PERSIST_WAL not configured on physical connection")
	}
	for range 2 {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if fc.closes.Load() != 1 || pin.closes.Load() != 1 || len(m.snapshots()) != 0 {
		t.Fatalf("driver closes=%d pins=%d owners=%+v", fc.closes.Load(), pin.closes.Load(), m.snapshots())
	}
}

func TestBeforeDriverFailureAndCancellationReleaseWithoutFault(t *testing.T) {
	c, pin, m, base := testConnector(t, &fakeConn{}, nil, 2)
	want := errors.New("second pin failed")
	c.prepare = func(_ string, acquired func(io.Closer)) error { acquired(pin); return want }
	if _, err := c.Connect(t.Context()); !errors.Is(err, want) {
		t.Fatalf("prepare=%v", err)
	}
	if base.calls.Load() != 0 || pin.closes.Load() != 1 || len(m.snapshots()) != 0 {
		t.Fatal("safe preparation failure retained or entered driver")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := c.Connect(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation=%v", err)
	}
	if pin.closes.Load() != 1 {
		t.Fatal("cancellation acquired a pin")
	}
}

func TestRawOutstandingResourceMakesCloseUncertain(t *testing.T) {
	physical := &fakeConn{}
	c, pin, m, _ := testConnector(t, physical, nil, 2)
	dc, err := c.Connect(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	conn := dc.(*guardedConn)
	conn.owner.addResource()
	if err := conn.Close(); !errors.Is(err, ErrRetained) {
		t.Fatalf("outstanding raw object Close=%v", err)
	}
	if pin.closes.Load() != 0 || m.snapshots()[0].Stage != "raw-resources" {
		t.Fatalf("resources released or unreported: %+v", m.snapshots())
	}
	if physical.closes.Load() != 0 {
		t.Fatal("Close invalidated TLS while raw child was outstanding")
	}
}
