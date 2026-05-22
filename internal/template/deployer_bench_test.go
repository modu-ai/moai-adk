// SPEC-V3R3-UPDATE-CLEANUP-001 — NFR-UPC-P1 benchmark gate.
// Compares baseline (direct os.WriteFile) vs atomic write (write-to-tmp + rename).
// The gate passes if delta ≤ 5% (benchstat analysis on ubuntu-latest CI).

package template

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// setupSyntheticProject builds a temp directory with numFiles regular files and
// numDeprecated deprecated-path stubs. Returns the project root and an FS
// with the same files for deployment.
func setupSyntheticProject(b *testing.B, numFiles, numDeprecated, _ int) (root string, syntheticFS fstest.MapFS) {
	b.Helper()
	root = b.TempDir()
	syntheticFS = make(fstest.MapFS)

	for i := 0; i < numFiles; i++ {
		name := fmt.Sprintf(".claude/agents/core/agent-%03d.md", i)
		data := []byte(fmt.Sprintf("# Agent %d\ncontent line 1\ncontent line 2\n", i))
		syntheticFS[name] = &fstest.MapFile{Data: data}
	}
	for i := 0; i < numDeprecated; i++ {
		name := fmt.Sprintf(".claude/commands/agency/cmd-%02d.md", i)
		data := []byte(fmt.Sprintf("# Deprecated cmd %d\n", i))
		syntheticFS[name] = &fstest.MapFile{Data: data}
		// Create the file on disk to simulate a user project with deprecated paths
		abs := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			b.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(abs, data, 0o644); err != nil {
			b.Fatalf("WriteFile deprecated: %v", err)
		}
	}
	// Initialise .moai dir
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		b.Fatalf("MkdirAll .moai: %v", err)
	}
	return root, syntheticFS
}

// Benchmark_UpdateCleanup_Baseline simulates deployment WITHOUT atomic write.
// This is the v2.14.0-equivalent: direct os.WriteFile (non-atomic).
// Used as the baseline for benchstat comparison.
func Benchmark_UpdateCleanup_Baseline(b *testing.B) {
	root, syntheticFS := setupSyntheticProject(b, 100, 5, 0)
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		b.Fatalf("manifest Load: %v", err)
	}

	// Use direct WriteFile deployer (non-atomic baseline)
	d := &baselineDeployer{fsys: syntheticFS}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := d.deployBaseline(ctx, root, mgr); err != nil {
			b.Fatalf("baseline deploy: %v", err)
		}
	}
}

// Benchmark_UpdateCleanup_AtomicWrite simulates deployment WITH atomic write.
// This is the SPEC-V3R3-UPDATE-CLEANUP-001 implementation.
func Benchmark_UpdateCleanup_AtomicWrite(b *testing.B) {
	root, syntheticFS := setupSyntheticProject(b, 100, 5, 0)
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		b.Fatalf("manifest Load: %v", err)
	}

	d := NewDeployerWithForceUpdate(syntheticFS, true)

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := d.Deploy(ctx, root, mgr, nil); err != nil {
			b.Fatalf("atomic deploy: %v", err)
		}
	}
}

// baselineDeployer is a minimal non-atomic deployer for benchmarking the
// pre-SPEC write pattern (direct os.WriteFile without tmp+rename).
type baselineDeployer struct {
	fsys fstest.MapFS
}

func (d *baselineDeployer) deployBaseline(ctx context.Context, root string, m manifest.Manager) error {
	for name, file := range d.fsys {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		destPath := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(destPath, file.Data, 0o644); err != nil {
			return err
		}
		hash := manifest.HashBytes(file.Data)
		_ = m.Track(name, manifest.TemplateManaged, hash)
	}
	return nil
}
