package sqliteguard

import (
	"errors"
	"io"
	"runtime"
	"testing"
)

func TestPartialPreparationPanicKeepsAcquiredPins(t *testing.T) {
	c, pin, m, base := testConnector(t, nil, nil, 1)
	c.prepare = func(_ string, acquired func(io.Closer)) error {
		acquired(pin)
		panic("second native preparation uncertain")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic lost")
			}
		}()
		_, _ = c.Connect(t.Context())
	}()
	runtime.GC()
	if base.calls.Load() != 0 || pin.closes.Load() != 0 || len(m.snapshots()) != 1 || m.snapshots()[0].Pins != 1 || !m.snapshots()[0].Faulted {
		t.Fatalf("partial preparation root=%+v", m.snapshots())
	}
}

type failedCleanupPin struct {
	err    error
	closes int
}

func (p *failedCleanupPin) Close() error { p.closes++; return p.err }

func TestFailedPreparationCleanupRemainsReservedAndReported(t *testing.T) {
	c, _, m, base := testConnector(t, nil, nil, 1)
	prepareErr := errors.New("native policy refused")
	cleanupErr := errors.New("native handle Close uncertain")
	pin := &failedCleanupPin{err: cleanupErr}
	c.prepare = func(_ string, acquired func(io.Closer)) error { acquired(pin); return prepareErr }
	_, err := c.Connect(t.Context())
	if !errors.Is(err, prepareErr) || !errors.Is(err, cleanupErr) {
		t.Fatalf("error identity=%v", err)
	}
	if pin.closes != 1 || base.calls.Load() != 0 || len(m.snapshots()) != 1 || m.snapshots()[0].Pins != 1 || !m.snapshots()[0].Faulted {
		t.Fatalf("cleanup ownership=%+v", m.snapshots())
	}
	if _, err := c.Connect(t.Context()); !errors.Is(err, ErrFaulted) {
		t.Fatalf("cleanup fault admission=%v", err)
	}
	if pin.closes != 1 {
		t.Fatal("unrequested native cleanup retry")
	}
}
