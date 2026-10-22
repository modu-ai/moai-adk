package privatefile_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/testguard"
)

func TestSecureExistingFilePreservesPayloadAndRejectsLink(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload")
	if err := privatefile.WriteFile(path, []byte("preserved")); err != nil {
		t.Fatal(err)
	}
	testguard.MakeFilePublic(t, path)
	if err := privatefile.Check(path); err == nil {
		t.Fatal("fixture must begin publicly readable")
	}
	if err := privatefile.SecureFile(path); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.Check(path); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "preserved" {
		t.Fatalf("payload=%q err=%v", raw, err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.SecureFile(link); err == nil {
		t.Fatal("symlink accepted")
	}
}

func TestSecureDirectoryAppliesNativePrivatePolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "directory")
	if err := privatefile.Mkdir(path); err != nil {
		t.Fatal(err)
	}
	testguard.MakeFilePublic(t, path)
	if err := privatefile.CheckDirectory(path); err == nil {
		t.Fatal("fixture directory must begin broadly readable")
	}
	if err := privatefile.SecureDirectory(path); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.CheckDirectory(path); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.SecureDirectory(link); err == nil {
		t.Fatal("symlink directory accepted")
	}
	if err := privatefile.CheckDirectory(link); err == nil {
		t.Fatal("symlink counted as a private directory")
	}
}

func TestOpenAppendNarrowsExistingWithoutTruncation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log")
	if err := privatefile.WriteFile(path, []byte("first\n")); err != nil {
		t.Fatal(err)
	}
	testguard.MakeFilePublic(t, path)
	f, err := privatefile.OpenAppend(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := privatefile.Check(path); err != nil {
		_ = f.Close()
		t.Fatalf("not private before append: %v", err)
	}
	if _, err := f.WriteString("second\n"); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "first\nsecond\n" {
		t.Fatalf("records=%q err=%v", raw, err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if f, err := privatefile.OpenAppend(link); err == nil {
		_ = f.Close()
		t.Fatal("symlink append accepted")
	}
}

func TestOpenAppendConcurrentRecordsStayIntact(t *testing.T) {
	const writers, records = 8, 20
	path := filepath.Join(t.TempDir(), "concurrent-log")
	errors := make(chan error, writers)
	var done sync.WaitGroup
	for writer := range writers {
		done.Add(1)
		go func() {
			defer done.Done()
			f, err := privatefile.OpenAppend(path)
			if err != nil {
				errors <- err
				return
			}
			for seq := range records {
				if _, err := fmt.Fprintf(f, "writer=%d seq=%d\n", writer, seq); err != nil {
					_ = f.Close()
					errors <- err
					return
				}
			}
			if err := f.Close(); err != nil {
				errors <- err
			}
		}()
	}
	done.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
	if t.Failed() {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		seen[line]++
	}
	if len(seen) != writers*records {
		t.Fatalf("unique records=%d, want %d", len(seen), writers*records)
	}
	for writer := range writers {
		for seq := range records {
			line := fmt.Sprintf("writer=%d seq=%d", writer, seq)
			if seen[line] != 1 {
				t.Errorf("record %q count=%d, want 1", line, seen[line])
			}
		}
	}
	if err := privatefile.Check(path); err != nil {
		t.Fatal(err)
	}
}

func TestMkdirIsPrivateBeforeReceivingPayload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private-directory")
	if err := privatefile.Mkdir(path); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.CheckDirectory(path); err != nil {
		t.Fatal(err)
	}
	if err := privatefile.Mkdir(path); !os.IsExist(err) {
		t.Fatalf("existing directory error=%v", err)
	}
}
