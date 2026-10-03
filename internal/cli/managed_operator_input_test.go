package cli

// managed_operator_input_test.go — SPEC-FACTORY-MANAGED-CARD-CHILD-001
// AC-CC-010 (unit level): the lane loop's operator-input pump hands each
// managed session one line at a time, latches an adapter shut after the line
// that ends its session (/exit or /quit), and holds every later line for the
// next session.

import (
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

const pumpWatchdog = 5 * time.Second

// pumpRead performs one Read of up to size bytes bounded by the watchdog.
func pumpRead(t *testing.T, r io.Reader, size int) (string, error) {
	t.Helper()
	type result struct {
		s   string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		buf := make([]byte, size)
		n, err := r.Read(buf)
		ch <- result{string(buf[:n]), err}
	}()
	select {
	case res := <-ch:
		return res.s, res.err
	case <-time.After(pumpWatchdog):
		t.Fatalf("Read did not return within %s", pumpWatchdog)
		return "", nil
	}
}

func pumpWriteAsync(pw *io.PipeWriter, chunk string) {
	go func() { _, _ = pw.Write([]byte(chunk)) }()
}

func TestManagedOperatorInputPumpDetachesEndedSession(t *testing.T) {
	newPump := func(t *testing.T) (*managedOperatorPump, *io.PipeWriter) {
		t.Helper()
		pr, pw := io.Pipe()
		t.Cleanup(func() { _ = pw.Close() })
		return newManagedOperatorPump(pr), pw
	}
	mustLine := func(t *testing.T, r io.Reader, want string) {
		t.Helper()
		got, err := pumpRead(t, r, 64)
		if err != nil || got != want {
			t.Fatalf("Read = %q, %v; want %q, nil", got, err, want)
		}
	}
	mustEOF := func(t *testing.T, r io.Reader) {
		t.Helper()
		got, err := pumpRead(t, r, 64)
		if got != "" || !errors.Is(err, io.EOF) {
			t.Fatalf("Read = %q, %v; want \"\", io.EOF", got, err)
		}
	}

	t.Run("single_line_per_read", func(t *testing.T) {
		p, pw := newPump(t)
		a := p.attach()
		defer func() { _ = a.Close() }()
		pumpWriteAsync(pw, "a\nb\n")
		mustLine(t, a, "a\n")
		mustLine(t, a, "b\n")
		// A caller buffer smaller than the line gets the remainder next.
		pumpWriteAsync(pw, "xy\n")
		if got, err := pumpRead(t, a, 1); err != nil || got != "x" {
			t.Fatalf("1-byte Read = %q, %v; want \"x\", nil", got, err)
		}
		mustLine(t, a, "y\n")
	})

	t.Run("multi_line_chunk_after_exit", func(t *testing.T) {
		p, pw := newPump(t)
		a := p.attach()
		pumpWriteAsync(pw, "a\n/exit\nnext\n")
		mustLine(t, a, "a\n")
		mustLine(t, a, "/exit\n")
		mustEOF(t, a)
		_ = a.Close()
		b := p.attach()
		defer func() { _ = b.Close() }()
		mustLine(t, b, "next\n")
	})

	t.Run("quit_token", func(t *testing.T) {
		p, pw := newPump(t)
		a := p.attach()
		pumpWriteAsync(pw, "/quit\nnext\n")
		mustLine(t, a, "/quit\n")
		mustEOF(t, a)
		_ = a.Close()
		b := p.attach()
		defer func() { _ = b.Close() }()
		mustLine(t, b, "next\n")
	})

	t.Run("buffer_saturation", func(t *testing.T) {
		p, pw := newPump(t)
		a := p.attach()
		lines := make(chan string, managedOperatorInputBuffer)
		go readManagedOperatorInput(a, lines)
		chunk := ""
		for i := 0; i < 10; i++ {
			chunk += "l\n"
		}
		pumpWriteAsync(pw, chunk+"/exit\ntail\n")
		// The reader goroutine fills the driver-sized channel and blocks; only
		// then does the test drain it, as a busy driver eventually would.
		time.Sleep(100 * time.Millisecond)
		got := 0
		last := ""
		deadline := time.After(pumpWatchdog)
	drain:
		for {
			select {
			case line, ok := <-lines:
				if !ok {
					break drain
				}
				got++
				last = line
			case <-deadline:
				t.Fatal("reader goroutine never finished after the end token")
			}
		}
		if got != 11 || last != "/exit" {
			t.Fatalf("session A received %d lines ending in %q, want 11 ending in /exit", got, last)
		}
		_ = a.Close()
		b := p.attach()
		defer func() { _ = b.Close() }()
		mustLine(t, b, "tail\n")
	})

	t.Run("close_unblocks_read", func(t *testing.T) {
		p, _ := newPump(t)
		a := p.attach()
		done := make(chan error, 1)
		go func() {
			_, err := a.Read(make([]byte, 8))
			done <- err
		}()
		time.Sleep(50 * time.Millisecond)
		_ = a.Close()
		select {
		case err := <-done:
			if !errors.Is(err, io.EOF) {
				t.Fatalf("blocked Read after Close = %v, want io.EOF", err)
			}
		case <-time.After(time.Second):
			t.Fatal("Close did not unblock the Read within 1s")
		}
	})

	t.Run("source_eof", func(t *testing.T) {
		p, pw := newPump(t)
		a := p.attach()
		// An unterminated final line is only a line once the source ends.
		go func() {
			_, _ = pw.Write([]byte("x\ny"))
			_ = pw.Close()
		}()
		mustLine(t, a, "x\n")
		mustLine(t, a, "y")
		mustEOF(t, a)
		_ = a.Close()
		b := p.attach()
		defer func() { _ = b.Close() }()
		mustEOF(t, b)
	})

	t.Run("fatal_end_residual", func(t *testing.T) {
		p, pw := newPump(t)
		a := p.attach()
		pumpWriteAsync(pw, "got\n")
		mustLine(t, a, "got\n") // already received: A's, never recovered
		_ = a.Close()           // session ended with no end token
		pumpWriteAsync(pw, "after\n")
		b := p.attach()
		defer func() { _ = b.Close() }()
		mustLine(t, b, "after\n")
	})
}

// pumpCountingSource serves n zero bytes (no newline) and counts what it served.
type pumpCountingSource struct {
	left int
	read atomic.Int64
}

func (c *pumpCountingSource) Read(b []byte) (int, error) {
	if c.left == 0 {
		return 0, io.EOF
	}
	n := min(len(b), c.left)
	clear(b[:n])
	c.left -= n
	c.read.Add(int64(n))
	return n, nil
}

// The driver's scanner stops at bufio.MaxScanTokenSize; the pump mirrors that
// bound, so a newline-free stream cannot grow launcher memory.
func TestManagedOperatorInputPumpBoundsLineLength(t *testing.T) {
	src := &pumpCountingSource{left: 2 * managedOperatorLineLimit}
	p := newManagedOperatorPump(src)
	a := p.attach()
	defer func() { _ = a.Close() }()
	got, err := pumpRead(t, a, 64)
	if got != "" || !errors.Is(err, io.EOF) {
		t.Fatalf("Read = %d bytes, %v; want no line and io.EOF (the driver's scanner delivers nothing for an oversize token)", len(got), err)
	}
	if n := src.read.Load(); n > int64(managedOperatorLineLimit) {
		t.Fatalf("pump consumed %d bytes from the source; bound is %d", n, managedOperatorLineLimit)
	}
}
