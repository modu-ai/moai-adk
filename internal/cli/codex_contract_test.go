package cli

// codex_contract_test.go — SPEC-CODEX-INIT-001 run-phase tests, contract
// side. This file holds the path-containment axis (AC-CI-011: 84 + 60 + 36
// = 180 cells); the link-creation, idempotency, and reachability cells
// (AC-CI-005/006/007) live in codex_contract_link_test.go.
//
// Disciplines: refusal is judged by STRUCTURE (closed-set IsRegular, not a
// kind enumeration), the refused path is never READ (read-seam 0 — Lstat is
// a component probe and excluded from the count), every axis is crossed
// with verb × spawn (a guard reachable through only one launch path is the
// defect definition 1 kills), and skipped real-fixture kinds are LISTED.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// codexSentinelOutside is the acceptance-mandated marker proving an outside
// file was neither read nor written through a refused path.
const codexSentinelOutside = "SENTINEL-OUTSIDE-3k3"

var errCodexFixtureUnsupported = errors.New("fixture kind unsupported on this platform")

// ─── shared cell assertions ────────────────────────────────────────────────

// codexWantPathRefusal holds the assertions every containment-refusal cell
// shares: non-success exit naming the refused path, zero launches (both
// sites), and zero read-mode seam calls toward the refused path.
func codexWantPathRefusal(t *testing.T, h *codexGateHarness, refusedRel string, stdout, stderr string, exitCode int) {
	t.Helper()
	if exitCode != codexInitFailureExitCode {
		t.Errorf("exit code = %d, want %d (non-success)", exitCode, codexInitFailureExitCode)
	}
	combined := stdout + stderr
	if !strings.Contains(combined, refusedRel) {
		t.Errorf("diagnostic does not name the refused path %q: %q", refusedRel, combined)
	}
	if n := h.launch.count(); n != 0 {
		t.Errorf("launches = %d, want 0 on a refused path", n)
	}
	if n := h.fs.readsOf(refusedRel); n != 0 {
		t.Errorf("read-mode seam calls toward %q = %d, want 0 (refusal precedes any read)", refusedRel, n)
	}
	if h.contractCalls != 1 {
		t.Errorf("contract calls = %d, want 1 (the guard is the contract's first act)", h.contractCalls)
	}
}

// codexContainmentOutcome is the per-cell observation the spawn/verb
// invariance assertions compare.
type codexContainmentOutcome struct {
	exitCode int
	named    bool
	launches int
	readsOfP int
}

// ─── axis 1 — Lstat mode injection (84 cells) ──────────────────────────────

// codexFakeFileInfo injects an Lstat result the real filesystem cannot
// produce in a portable test (device nodes, sockets at arbitrary names...).
type codexFakeFileInfo struct {
	mode os.FileMode
}

func (f codexFakeFileInfo) Name() string       { return "injected" }
func (f codexFakeFileInfo) Size() int64        { return 0 }
func (f codexFakeFileInfo) Mode() os.FileMode  { return f.mode }
func (f codexFakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f codexFakeFileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f codexFakeFileInfo) Sys() any           { return nil }

// TestCodexPathGuardInjectedModes: every non-regular mode bit × each of the
// three instruction paths × 2 verbs × 2 spawn. The judgement must be the
// CLOSED SET (IsRegular) — any enumeration of refused kinds leaves one of
// these seven alive.
func TestCodexPathGuardInjectedModes(t *testing.T) {
	modes := []struct {
		name string
		mode os.FileMode
	}{
		{"dir", os.ModeDir},
		{"symlink", os.ModeSymlink},
		{"fifo", os.ModeNamedPipe},
		{"socket", os.ModeSocket},
		{"device", os.ModeDevice},
		{"chardev", os.ModeDevice | os.ModeCharDevice},
		{"irregular", os.ModeIrregular},
	}
	files := defaultCodexInstructionRelPaths()
	verbs := []string{"cli", "app"}
	type key struct {
		mode, file string
	}
	outcomes := map[key][]codexContainmentOutcome{}
	for _, m := range modes {
		for _, fileName := range files {
			for _, verb := range verbs {
				for _, spawn := range []bool{false, true} {
					name := fmt.Sprintf("mode=%s/file=%s/verb=%s/spawn=%t", m.name, fileName, verb, spawn)
					t.Run(name, func(t *testing.T) {
						_, proj := codexNewSandbox(t)
						codexLayWiringState(t, proj, stateNotWiredMissing)
						h := withCodexGateHarness(t, proj, codexHarnessOpts{
							capable: true, promptAccepts: true, contractReal: true,
						})
						// Overlay the mode injection ON TOP of the recorder —
						// cleanup order (LIFO) restores the recorder first.
						prev := codexLstatFn
						codexLstatFn = func(path string) (os.FileInfo, error) {
							if strings.HasSuffix(path, string(filepath.Separator)+fileName) {
								return codexFakeFileInfo{mode: m.mode}, nil
							}
							return prev(path)
						}
						t.Cleanup(func() { codexLstatFn = prev })

						stdout, stderr, code := runCodexInitCell(t, verb, spawn)
						codexWantPathRefusal(t, h, fileName, stdout, stderr, code)

						outcomes[key{m.name, fileName}] = append(outcomes[key{m.name, fileName}], codexContainmentOutcome{
							exitCode: code,
							named:    strings.Contains(stdout+stderr, fileName),
							launches: h.launch.count(),
							readsOfP: h.fs.readsOf(fileName),
						})
					})
				}
			}
		}
	}
	// Same mode × file must refuse identically across all four verb×spawn
	// combinations — the guard is reachable through ONE function only.
	for k, outs := range outcomes {
		if len(outs) != 4 {
			t.Fatalf("mode=%s file=%s: %d outcomes, want 4", k.mode, k.file, len(outs))
		}
		for i, o := range outs[1:] {
			if o != outs[0] {
				t.Errorf("mode=%s file=%s: outcome %d differs from the first (%+v vs %+v) — the refusal split by verb or spawn", k.mode, k.file, i+1, o, outs[0])
			}
		}
	}
}

// ─── axis 2 — real fixtures (60 cells) ─────────────────────────────────────

type codexRealKind string

const (
	kindExternalSymlink codexRealKind = "external_symlink"
	kindInternalSymlink codexRealKind = "internal_symlink"
	kindDirectory       codexRealKind = "directory"
	kindFIFO            codexRealKind = "fifo"
	kindSocket          codexRealKind = "unix_socket"
)

// codexMakeRealFixture places the given kind AT the instruction path.
// It returns a non-empty skip reason when the platform cannot create the
// kind — the caller skips the cell and records the kind in the skip list.
func codexMakeRealFixture(t *testing.T, sbx, proj, name string, kind codexRealKind) string {
	t.Helper()
	target := filepath.Join(proj, name)
	switch kind {
	case kindDirectory:
		if err := os.MkdirAll(target, 0o755); err != nil {
			t.Fatal(err)
		}
		return ""
	case kindExternalSymlink:
		outside := filepath.Join(sbx, "outside-"+name)
		if err := os.WriteFile(outside, []byte(codexSentinelOutside), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, target); err != nil {
			return fmt.Sprintf("symlink unsupported: %v", err)
		}
		return ""
	case kindInternalSymlink:
		inner := filepath.Join(proj, name+".inner.txt")
		if err := os.WriteFile(inner, []byte("internal regular target"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(inner, target); err != nil {
			return fmt.Sprintf("symlink unsupported: %v", err)
		}
		return ""
	case kindFIFO:
		if err := makeCodexFIFOFixture(target); err != nil {
			if errors.Is(err, errCodexFixtureUnsupported) {
				return "fifo unsupported on this platform"
			}
			t.Fatal(err)
		}
		return ""
	case kindSocket:
		cleanup, err := makeCodexSocketFixtureAt(target)
		if err != nil {
			if errors.Is(err, errCodexFixtureUnsupported) {
				return "unix socket unsupported on this platform"
			}
			t.Fatal(err)
		}
		t.Cleanup(cleanup)
		return ""
	default:
		t.Fatalf("unknown real fixture kind %q", kind)
		return ""
	}
}

// TestCodexPathGuardRealFixtures: five real path kinds × 3 files × 2 verbs
// × 2 spawn. Kinds the platform cannot create are skipped AND listed.
func TestCodexPathGuardRealFixtures(t *testing.T) {
	kinds := []codexRealKind{kindExternalSymlink, kindInternalSymlink, kindDirectory, kindFIFO, kindSocket}
	files := defaultCodexInstructionRelPaths()
	verbs := []string{"cli", "app"}
	var skippedKinds []string
	for _, kind := range kinds {
		for _, fileName := range files {
			for _, verb := range verbs {
				for _, spawn := range []bool{false, true} {
					name := fmt.Sprintf("kind=%s/file=%s/verb=%s/spawn=%t", kind, fileName, verb, spawn)
					t.Run(name, func(t *testing.T) {
						sbx, proj := codexNewSandbox(t)
						codexLayWiringState(t, proj, stateNotWiredMissing)
						if reason := codexMakeRealFixture(t, sbx, proj, fileName, kind); reason != "" {
							skippedKinds = appendUniqueKind(skippedKinds, string(kind))
							t.Skip(reason)
						}
						h := withCodexGateHarness(t, proj, codexHarnessOpts{
							capable: true, promptAccepts: true, contractReal: true,
						})
						stdout, stderr, code := runCodexInitCell(t, verb, spawn)
						codexWantPathRefusal(t, h, fileName, stdout, stderr, code)

						combined := stdout + stderr + fmt.Sprint(code)
						if kind == kindExternalSymlink {
							outside := filepath.Join(sbx, "outside-"+fileName)
							data, rerr := os.ReadFile(outside)
							if rerr != nil {
								t.Fatalf("outside target unreadable: %v", rerr)
							}
							if string(data) != codexSentinelOutside {
								t.Errorf("outside target changed through the refused link: %q", data)
							}
							if strings.Contains(combined, codexSentinelOutside) {
								t.Errorf("the outside sentinel leaked into the diagnostics — reading the refused path is also forbidden")
							}
						}
					})
				}
			}
		}
	}
	// The skip list: quiet skips read as passes, so every skipped kind is
	// named. An empty list with zero skips is equally honest.
	t.Logf("skipped real fixture kinds on this platform: %s", strings.Join(skippedOrNone(skippedKinds), ", "))
}

func appendUniqueKind(list []string, kind string) []string {
	for _, k := range list {
		if k == kind {
			return list
		}
	}
	return append(list, kind)
}

func skippedOrNone(list []string) []string {
	if len(list) == 0 {
		return []string{"(none)"}
	}
	return list
}

// ─── axis 3 — parent-component escapes (36 cells) ──────────────────────────

// TestCodexPathGuardParentEscape: the leaf name alone does not bound the
// path. Three batches — a parent symlink to outside, a `..` escape through
// plain directories, and the lexical trap (parent symlink + `..` that only
// LOOKS inside if cleaned before resolution) — each × 3 files × 2 verbs ×
// 2 spawn. The variant paths travel through the path-table seam; the cells
// still run the real launch verbs end to end.
func TestCodexPathGuardParentEscape(t *testing.T) {
	batches := []struct {
		name    string
		variant func(t *testing.T, sbx, proj, fileName string) (rel string, outsideTarget string)
	}{
		{
			name: "parent_symlink_out",
			variant: func(t *testing.T, sbx, proj, fileName string) (string, string) {
				outsideDir := filepath.Join(sbx, "outside-dir")
				if err := os.MkdirAll(outsideDir, 0o755); err != nil {
					t.Fatal(err)
				}
				outsideTarget := filepath.Join(outsideDir, fileName)
				if err := os.WriteFile(outsideTarget, []byte(codexSentinelOutside), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outsideDir, filepath.Join(proj, "docs")); err != nil {
					t.Skipf("symlink unsupported: %v", err)
				}
				return "docs/" + fileName, outsideTarget
			},
		},
		{
			name: "dotdot_escape",
			variant: func(t *testing.T, sbx, proj, fileName string) (string, string) {
				outsideTarget := filepath.Join(sbx, "outside-"+fileName)
				if err := os.WriteFile(outsideTarget, []byte(codexSentinelOutside), 0o644); err != nil {
					t.Fatal(err)
				}
				return "../outside-" + fileName, outsideTarget
			},
		},
		{
			name: "parent_symlink_dotdot",
			variant: func(t *testing.T, sbx, proj, fileName string) (string, string) {
				outsideDir := filepath.Join(sbx, "outside-dir")
				if err := os.MkdirAll(outsideDir, 0o755); err != nil {
					t.Fatal(err)
				}
				outsideTarget := filepath.Join(outsideDir, fileName)
				if err := os.WriteFile(outsideTarget, []byte(codexSentinelOutside), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outsideDir, filepath.Join(proj, "docs")); err != nil {
					t.Skipf("symlink unsupported: %v", err)
				}
				// The lexical trap: docs/../AGENTS.md cleans to AGENTS.md at
				// the project root, but the KERNEL walks docs first.
				return "docs/../" + fileName, outsideTarget
			},
		},
	}
	files := defaultCodexInstructionRelPaths()
	verbs := []string{"cli", "app"}
	for _, batch := range batches {
		for _, fileName := range files {
			for _, verb := range verbs {
				for _, spawn := range []bool{false, true} {
					name := fmt.Sprintf("batch=%s/file=%s/verb=%s/spawn=%t", batch.name, fileName, verb, spawn)
					t.Run(name, func(t *testing.T) {
						sbx, proj := codexNewSandbox(t)
						codexLayWiringState(t, proj, stateNotWiredMissing)
						variantRel, outsideTarget := batch.variant(t, sbx, proj, fileName)

						// Override ONLY the current file's table entry so the
						// refusal lands on this cell's own file axis.
						table := defaultCodexInstructionRelPaths()
						for i, f := range table {
							if f == fileName {
								table[i] = variantRel
							}
						}
						prevTable := codexInstructionRelPathsFn
						codexInstructionRelPathsFn = func() []string { return append([]string(nil), table...) }
						t.Cleanup(func() { codexInstructionRelPathsFn = prevTable })

						h := withCodexGateHarness(t, proj, codexHarnessOpts{
							capable: true, promptAccepts: true, contractReal: true,
						})
						stdout, stderr, code := runCodexInitCell(t, verb, spawn)
						codexWantPathRefusal(t, h, variantRel, stdout, stderr, code)

						data, rerr := os.ReadFile(outsideTarget)
						if rerr != nil {
							t.Fatalf("outside target unreadable: %v", rerr)
						}
						if string(data) != codexSentinelOutside {
							t.Errorf("outside target changed through the refused path: %q", data)
						}
						combined := stdout + stderr + fmt.Sprint(code)
						if strings.Contains(combined, codexSentinelOutside) {
							t.Errorf("the outside sentinel leaked into the diagnostics")
						}
					})
				}
			}
		}
	}
}
