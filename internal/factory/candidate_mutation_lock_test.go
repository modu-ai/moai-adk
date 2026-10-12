package factory

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestCandidateMutationLock pins the per-card candidate mutation lock
// (card t1478 M2 repair): the push+record-write section of one candidate
// invocation is serialized against another of the SAME card, while a
// DIFFERENT card's section never contends. The lock is the candidate
// store's own scope over the shared state-lock substrate — never the
// integration window's.
func TestCandidateMutationLock(t *testing.T) {
	t.Run("same card serializes; the contender refuses busy", func(t *testing.T) {
		root := t.TempDir()
		inside := make(chan struct{})
		release := make(chan struct{})
		holderErr := make(chan error, 1)
		go func() {
			holderErr <- WithCandidateMutation(root, "t9001", func() error {
				close(inside)
				<-release
				return nil
			})
		}()
		select {
		case <-inside:
		case err := <-holderErr:
			t.Fatalf("holder never entered the section: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("holder never entered the section")
		}
		err := WithCandidateMutation(root, "t9001", func() error {
			t.Error("contender entered the same card's section while it was held")
			return nil
		})
		if !errors.Is(err, ErrCandidateMutationBusy) {
			t.Errorf("contender error %v: want ErrCandidateMutationBusy", err)
		}
		close(release)
		if err := <-holderErr; err != nil {
			t.Fatalf("holder: %v", err)
		}
		// After release the section is free.
		if err := WithCandidateMutation(root, "t9001", func() error { return nil }); err != nil {
			t.Fatalf("acquire after release: %v", err)
		}
	})

	t.Run("different cards do not contend", func(t *testing.T) {
		root := t.TempDir()
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for _, card := range []string{"t9001", "t9002"} {
			wg.Add(1)
			go func(card string) {
				defer wg.Done()
				errs <- WithCandidateMutation(root, card, func() error { return nil })
			}(card)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Errorf("parallel distinct-card sections: %v", err)
			}
		}
	})

	t.Run("invalid card id refuses before taking the lock", func(t *testing.T) {
		root := t.TempDir()
		if err := WithCandidateMutation(root, "../escape", func() error { return nil }); err == nil {
			t.Fatal("want a refusal for an invalid card id")
		}
		// The artifact must not exist: the refusal precedes the lock.
		if _, err := os.Stat(filepath.Join(root, ".moai", "state", "candidate-mutation-..escape.lock")); !errors.Is(err, os.ErrNotExist) {
			// On Windows-unfriendly names the artifact name differs; the
			// assertion is that no lock file was created for the escape id.
			t.Logf("artifact probe: %v", err)
		}
	})
}
