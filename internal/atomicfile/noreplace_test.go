//go:build linux || darwin || windows

package atomicfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestRenameNoReplacePreservesOccupiedDestination(t *testing.T) {
	for _, content := range []string{"", "another writer"} {
		t.Run(fmt.Sprintf("size-%d", len(content)), func(t *testing.T) {
			dir := t.TempDir()
			source := filepath.Join(dir, "stage")
			dest := filepath.Join(dir, "destination")
			if err := os.WriteFile(source, []byte("snapshot"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dest, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			if err := RenameNoReplace(source, dest); !errors.Is(err, os.ErrExist) {
				t.Fatalf("occupied error=%v", err)
			}
			for path, want := range map[string]string{source: "snapshot", dest: content} {
				if got, err := os.ReadFile(path); err != nil || string(got) != want {
					t.Fatalf("file changed: %v", err)
				}
			}
		})
	}
}

func TestRenameNoReplaceCompetingPublications(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "destination")
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	for _, path := range []string{a, b} {
		if err := os.WriteFile(path, []byte(filepath.Base(path)), 0600); err != nil {
			t.Fatal(err)
		}
	}
	result := make(chan error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, path := range []string{a, b} {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; result <- RenameNoReplace(path, dest) }()
	}
	close(start)
	wg.Wait()
	close(result)
	won := 0
	for err := range result {
		if err == nil {
			won++
		} else if !errors.Is(err, os.ErrExist) {
			t.Fatal(err)
		}
	}
	if won != 1 {
		t.Fatalf("publication winners=%d", won)
	}
	got, err := os.ReadFile(dest)
	if err != nil || (string(got) != "a" && string(got) != "b") {
		t.Fatalf("destination content: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("expected winner and intact losing stage: %v", err)
	}
}
