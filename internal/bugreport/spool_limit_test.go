package bugreport

// The spool read's memory bound (card-review finding, P2): the bounded read
// loaded the WHOLE file into memory before checking the 64KiB cap — an
// 8MiB spool file allocated ~8.4MB before being refused, and larger files
// scale to OOM. The read is bounded AT the cap (io.LimitReader) before the
// size judgment: an oversized spool allocates at most cap+1 bytes, ever.

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestSpoolReadBoundsAllocationAtTheCap(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	t.Setenv("CI", "")

	spoolDir := filepath.Join(home, filepath.FromSlash(BugreportStoreDir))
	if err := os.MkdirAll(spoolDir, 0o700); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}
	// An 8MiB spool file — 128x over the cap.
	oversized := filepath.Join(spoolDir, "spool.jsonl")
	f, err := os.Create(oversized)
	if err != nil {
		t.Fatalf("create oversized spool: %v", err)
	}
	line := []byte(`{"kind":"panic","verdict":"moai","frames":["internal/cli.Execute"]}` + "\n")
	for i := 0; i < 8*1024*1024/len(line); i++ {
		if _, err := f.Write(line); err != nil {
			t.Fatalf("write oversized spool: %v", err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close oversized spool: %v", err)
	}

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	_, _, err = ReadSpoolConsumable()
	runtime.ReadMemStats(&after)
	allocated := after.TotalAlloc - before.TotalAlloc

	if err == nil {
		t.Fatal("an oversized spool read without the over-cap error")
	}
	if allocated > 2*1024*1024 {
		t.Fatalf("the spool read allocated %d bytes for an 8MiB file — the cap must bound the read itself (%s), not judge after the fact", allocated, fmt.Sprintf("cap %d", config.DefaultBugreportSpoolMaxBytes))
	}
}
