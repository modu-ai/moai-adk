//go:build !windows

package receipt

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type receiptNonceCallback struct {
	source   io.Reader
	callback func()
}

func (r *receiptNonceCallback) Read(b []byte) (int, error) {
	if r.callback != nil {
		cb := r.callback
		r.callback = nil
		cb()
	}
	return r.source.Read(b)
}
func atReceiptNonce(t *testing.T, callback func()) {
	t.Helper()
	original := rand.Reader
	t.Cleanup(func() { rand.Reader = original })
	rand.Reader = &receiptNonceCallback{source: original, callback: func() { rand.Reader = original; callback() }}
}
func TestPrecommitCancellationPreservesSnapshot(t *testing.T) {
	s, e := OpenStore(context.Background(), privateDir(t), testUUID, true)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	atReceiptNonce(t, cancel)
	if e = s.Publish(ctx, candidate("p", "prior", "A")); !errors.Is(e, context.Canceled) {
		t.Fatal("precommit cancellation", e)
	}
	m, e := s.Snapshot(context.Background())
	if e != nil || len(m.Candidates()) != 0 {
		t.Fatal("cancel published", e)
	}
}
func TestLockReplacementRejectsStalePublisher(t *testing.T) {
	ctx := context.Background()
	dir := privateDir(t)
	a, e := OpenStore(ctx, dir, testUUID, true)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := a.Close(); err != nil {
			t.Error(err)
		}
	}()
	b, e := OpenStore(ctx, dir, testUUID, false)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := b.Close(); err != nil {
			t.Error(err)
		}
	}()
	next := candidate("p", "prior", "B")
	atReceiptNonce(t, func() {
		if e := os.Rename(filepath.Join(dir, ".lock"), filepath.Join(dir, ".lock-old")); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(filepath.Join(dir, ".lock"), nil, 0600); e != nil {
			t.Fatal(e)
		}
		if e := b.Publish(ctx, next); e != nil {
			t.Fatal(e)
		}
	})
	if e = a.Publish(ctx, candidate("p", "prior", "A")); e == nil {
		t.Fatal("stale publisher acknowledged")
	}
	m, e := b.Snapshot(ctx)
	if e != nil || len(m.Candidates()) != 1 || m.Candidates()[0] != next {
		t.Fatal("acknowledged publication lost", e)
	}
}
func TestSnapshotReplacementRejectsStalePublisher(t *testing.T) {
	ctx := context.Background()
	dir := privateDir(t)
	s, e := OpenStore(ctx, dir, testUUID, true)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	next := candidate("p", "prior", "B")
	replacement, _ := New(testUUID)
	_ = replacement.Publish(next)
	raw, _ := replacement.Marshal()
	atReceiptNonce(t, func() {
		if e := os.WriteFile(filepath.Join(dir, "manifest.json"), raw, 0600); e != nil {
			t.Fatal(e)
		}
	})
	if e = s.Publish(ctx, candidate("p", "prior", "A")); e == nil {
		t.Fatal("changed snapshot overwritten")
	}
	m, e := s.Snapshot(ctx)
	if e != nil || len(m.Candidates()) != 1 || m.Candidates()[0] != next {
		t.Fatal("changed snapshot lost", e)
	}
}

func TestWriteCancellationBeforeAndAfterCommit(t *testing.T) {
	for _, tc := range []struct {
		name       string
		at         int
		wantCount  int
		wantCancel bool
	}{{"before-rename", 2, 0, true}, {"final-guard", 3, 0, true}, {"after-rename", 4, 1, false}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			s, e := OpenStore(ctx, privateDir(t), testUUID, true)
			if e != nil {
				t.Fatal(e)
			}
			defer func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			}()
			e = s.transaction(ctx, false, func(guard func() error) error {
				m, err := s.read()
				if err != nil {
					return err
				}
				raw, err := m.Marshal()
				if err != nil {
					return err
				}
				expected := Hash(raw)
				if err = m.Publish(candidate("p", "prior", "A")); err != nil {
					return err
				}
				calls := 0
				return s.write(ctx, m, func() error {
					calls++
					if calls == tc.at {
						cancel()
					}
					return guard()
				}, &expected)
			})
			if tc.wantCancel && !errors.Is(e, context.Canceled) {
				t.Fatal("precommit cancellation", e)
			}
			if !tc.wantCancel && e != nil {
				t.Fatal("committed record failed", e)
			}
			m, err := s.Snapshot(context.Background())
			if err != nil || len(m.Candidates()) != tc.wantCount {
				t.Fatal("commit boundary", err)
			}
		})
	}
}
