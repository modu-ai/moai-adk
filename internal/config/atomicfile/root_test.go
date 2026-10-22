package atomicfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

func TestRootWritePreservesBytesModesAndForcedMode(t *testing.T) {
	dir := t.TempDir()
	r, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	})
	path := filepath.Join(dir, "section.yaml")
	if err := Write("section.yaml", []byte("first"), defs.FilePerm, RootOptions{Root: r}); err != nil {
		t.Fatal(err)
	}
	for _, forced := range []bool{false, true} {
		if err := os.Chmod(path, 0o640); err != nil {
			t.Fatal(err)
		}
		before, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		options := RootOptions{Root: r}
		want := before.Mode().Perm()
		if forced {
			options.ForceMode = 0o600
			if err := os.Chmod(path, options.ForceMode); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			want = info.Mode().Perm()
			if err := os.Chmod(path, 0o640); err != nil {
				t.Fatal(err)
			}
		}
		if err := Write("section.yaml", []byte("complete replacement"), defs.FilePerm, options); err != nil {
			t.Fatal(err)
		}
		if got, err := os.ReadFile(path); err != nil || string(got) != "complete replacement" {
			t.Fatalf("rooted payload = %q, %v", got, err)
		}
		if info, err := os.Stat(path); err != nil || info.Mode().Perm() != want {
			t.Fatalf("rooted mode differs from native fixture mode %o: %v", want, err)
		}
	}
	if got := countOrphanTemps(t, dir); got != 0 {
		t.Fatalf("rooted success left %d staging files", got)
	}
}

func TestRootWriteRefusesPlainPathFallbackAndCleansPublicationFailure(t *testing.T) {
	dir, outside := t.TempDir(), filepath.Join(t.TempDir(), "outside.yaml")
	if err := os.WriteFile(outside, []byte("outside bytes"), defs.FilePerm); err != nil {
		t.Fatal(err)
	}
	r, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, tc := range []struct {
		path string
		root *os.Root
	}{{outside, nil}, {outside, r}, {"../outside.yaml", r}} {
		if err := Write(tc.path, []byte("wrong bytes"), defs.FilePerm, RootOptions{Root: tc.root}); err == nil {
			t.Errorf("unsafe rooted write %q accepted", tc.path)
		}
	}
	if got, err := os.ReadFile(outside); err != nil || string(got) != "outside bytes" {
		t.Fatalf("outside payload changed: %q, %v", got, err)
	}
	refusal := errors.New("destination policy changed")
	err = Write("section.yaml", []byte("new"), defs.FilePerm, RootOptions{Root: r, BeforeRename: func() error { return refusal }})
	var stage *WriteError
	if !errors.Is(err, refusal) || !errors.As(err, &stage) || stage.Op != "stat" {
		t.Fatalf("publication policy failure = %v; want wrapped stat refusal", err)
	}
	if _, err := os.Lstat(filepath.Join(dir, "section.yaml")); !os.IsNotExist(err) {
		t.Fatalf("refused payload was published: %v", err)
	}
	if got := countOrphanTemps(t, dir); got != 0 {
		t.Fatalf("rooted failure left %d staging files", got)
	}
}
