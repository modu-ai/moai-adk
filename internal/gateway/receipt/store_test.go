package receipt

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func privateDir(t *testing.T) string {
	t.Helper()
	p, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Chmod(p, 0700); e != nil {
		t.Fatal(e)
	}
	return p
}
func TestStorePersistenceConcurrentAndLoss(t *testing.T) {
	if runtime.GOOS == "windows" {
		if _, e := OpenStore(context.Background(), privateDir(t), testUUID, true); !errors.Is(e, ErrUnsupported) {
			t.Fatal(e)
		}
		return
	}
	dir := privateDir(t)
	ctx := context.Background()
	s, e := OpenStore(ctx, dir, testUUID, true)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = OpenStore(ctx, dir, testUUID, true); e == nil {
		t.Fatal("new overwrote store")
	}
	other, e := OpenStore(ctx, dir, testUUID, false)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v := candidate("prefix", "parent", string(rune('A'+i)))
			target := s
			if i%2 == 0 {
				target = other
			}
			errs <- target.Publish(ctx, v)
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	_ = s.Close()
	_ = other.Close()
	resumed, e := OpenStore(ctx, dir, testUUID, false)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := resumed.Close(); err != nil {
			t.Error(err)
		}
	}()
	m, e := resumed.Snapshot(ctx)
	if e != nil || len(m.Candidates()) != 12 {
		t.Fatal(e, len(m.Candidates()))
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if !errors.Is(resumed.Publish(canceled, candidate("x", "y", "z")), context.Canceled) {
		t.Fatal("cancel ignored")
	}
	raw, e := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	if len(raw) > MaxBytes {
		t.Fatal("limit")
	}
	if e = os.Remove(filepath.Join(dir, "manifest.json")); e != nil {
		t.Fatal(e)
	}
	if _, e = resumed.Snapshot(ctx); e == nil {
		t.Fatal("missing reinitialized")
	}
	if _, e = OpenStore(ctx, dir, testUUID, false); e == nil {
		t.Fatal("resume missing accepted")
	}
}
func TestStoreRejectsUnsafePathsAndReplacement(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
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
	manifest := filepath.Join(dir, "manifest.json")
	if e = os.Chmod(manifest, 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Snapshot(ctx); e == nil {
		t.Fatal("public file accepted")
	}
	_ = os.Chmod(manifest, 0600)
	target := filepath.Join(privateDir(t), "target")
	if e = os.WriteFile(target, []byte("do not read"), 0600); e != nil {
		t.Fatal(e)
	}
	_ = os.Remove(manifest)
	if e = os.Symlink(target, manifest); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Snapshot(ctx); e == nil {
		t.Fatal("symlink accepted")
	}
	dir2 := privateDir(t)
	if e = os.Chmod(dir2, 0755); e != nil {
		t.Fatal(e)
	}
	if _, e = OpenStore(ctx, dir2, testUUID, true); e == nil {
		t.Fatal("public root accepted")
	}
	root := privateDir(t)
	link := filepath.Join(privateDir(t), "link")
	if e = os.Symlink(root, link); e != nil {
		t.Fatal(e)
	}
	if _, e = OpenStore(ctx, link, testUUID, true); e == nil {
		t.Fatal("root symlink accepted")
	}
	moved := dir + "-moved"
	if e = os.Rename(dir, moved); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _ = os.RemoveAll(moved) })
	if e = os.Mkdir(dir, 0700); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Snapshot(ctx); e == nil {
		t.Fatal("root replacement accepted")
	}
}

func TestStoreRejectsForeignConversation(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	ctx := context.Background()
	dir := privateDir(t)
	if _, e := OpenStore(ctx, dir, "bad", true); e == nil {
		t.Fatal("invalid authorized UUID")
	}
	s, e := OpenStore(ctx, dir, testUUID, true)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	if _, e = OpenStore(ctx, dir, "00000000-0000-4000-8000-000000000001", false); e == nil {
		t.Fatal("manifest accepted for another conversation")
	}
}
