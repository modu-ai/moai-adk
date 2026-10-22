package privatefile

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestCreateExclusivePrivateBeforePayloadAndNoOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.db")
	var f *os.File
	defer func() {
		if f != nil {
			if err := f.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	if err := CreateExclusive(path, func(owned *os.File) { f = owned }); err != nil {
		t.Fatal(err)
	}
	if err := Check(path); err != nil {
		t.Fatal(err)
	}
	info, err := f.Stat()
	if err != nil || info.Size() != 0 {
		t.Fatalf("not empty before payload: %v", err)
	}
	if _, err := f.WriteString("original"); err != nil {
		t.Fatal(err)
	}
	called := false
	if err := CreateExclusive(path, func(owned *os.File) {
		called = true
		if err := owned.Close(); err != nil {
			t.Error(err)
		}
	}); !errors.Is(err, os.ErrExist) {
		t.Fatalf("existing name error=%v", err)
	}
	if called {
		t.Fatal("existing file acquired")
	}
	after, err := f.Stat()
	if err != nil || !os.SameFile(info, after) {
		t.Fatalf("identity changed: %v", err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "original" {
		t.Fatalf("existing payload changed: %v", err)
	}
	if err := Check(path); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	f = nil
	if err := os.Remove(path); err != nil {
		t.Fatalf("completed owner did not release handle: %v", err)
	}
}

func TestCreateExclusiveCompetingCreators(t *testing.T) {
	path := filepath.Join(t.TempDir(), "single.db")
	start := make(chan struct{})
	result := make(chan error, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			var f *os.File
			err := CreateExclusive(path, func(owned *os.File) { f = owned })
			if f != nil {
				err = errors.Join(err, f.Close())
			}
			result <- err
		}()
	}
	close(start)
	wg.Wait()
	close(result)
	winners := 0
	for err := range result {
		if err == nil {
			winners++
		} else if !errors.Is(err, os.ErrExist) {
			t.Fatal(err)
		}
	}
	if winners != 1 {
		t.Fatalf("exclusive winners=%d", winners)
	}
	if err := Check(path); err != nil {
		t.Fatal(err)
	}
}

func TestCreateExclusiveRequiresOwnerBeforeCreation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unused.db")
	if err := CreateExclusive(path, nil); err == nil {
		t.Fatal("missing owner accepted")
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("missing owner created output: %v", err)
	}
}
