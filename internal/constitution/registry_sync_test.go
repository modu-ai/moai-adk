package constitution_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// Registry sync guard — SPEC-ZONE-REGISTRY-RESYNC-001 (M2, plan.md §F).
//
// This test is the BLOCKING guard for the zone registry: it rides the ordinary
// `go test ./...` CI job, which is the only blocking path (AC-ZRR-008 / D6 —
// the separate `constitution-check` workflow job is continue-on-error and can
// never block a PR). For EACH of the two mirrors it checks:
//
//  1. Validate result must not be Skipped (MOAI_CONSTITUTION_SKIP_VALIDATE=1
//     must make this guard FAIL, not pass — REQ-ZRR-010 / D7).
//  2. Validate DriftCount must be 0, naming every failing entry ID.
//  3. Anchor resolution (test-side interpreter, validator untouched — D1).
//  4. Literal clause presence with NO normalization (grep -F equivalent),
//     strictly stronger than the validator's whitespace-normalized check.
//  5. No entry's file: may reference the registry itself (D13).
//
// Option C (spec.md §1.2 v0.5.0): the 4 [SUPERSEDED …] retired entries are
// exempt from the clause/literal checks (their clause is an immutable audit
// record of withdrawn doctrine) but MUST still pass the anchor check — their
// anchor points at the successor section, so anchor resolution is the one
// mechanical check the retirement marker keeps.
//
// Evaluated-entry counts are asserted, not merely loaded (AC-ZRR-007
// partial-traversal mutant: a guard with an early return or exclusion list
// passes every mutation that lands outside its set unless the count itself is
// pinned). The passing output reports clause-checks / retired-skip /
// anchor-checks separately per mirror — the two clause/anchor counts differing
// IS the option-C contract.

const (
	// wantRegistryEntries pins the entry-set size (REQ-ZRR-005 / AC-ZRR-006).
	// Deliberate registry growth updates this constant in the same change.
	wantRegistryEntries = 101
	// wantRetiredExempt pins the number of [SUPERSEDED …] clause-exempt
	// entries under option C (spec.md §1.2 v0.5.0).
	wantRetiredExempt = 4
	// skipValidateEnv is the documented validation bypass. The guard treats
	// its presence as a FAILURE (REQ-ZRR-010 / D7) — never as "skip the
	// checks and pass".
	skipValidateEnv = "MOAI_CONSTITUTION_SKIP_VALIDATE"
)

// registrySyncMirror describes one registry surface under test.
type registrySyncMirror struct {
	name       string // subtest name
	registry   string // zone-registry.md path, relative to the package dir
	projectDir string // ProjectDir that resolves each entry's file: path
}

// repoRootForSync locates the repository root by walking up from the test
// working directory (the package dir) until go.mod is found — robust to the
// package moving within the tree, unlike a fixed "..", ".." join.
func repoRootForSync(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found walking up from %s", dir)
		}
		dir = parent
	}
}

func registrySyncMirrors(t *testing.T) []registrySyncMirror {
	t.Helper()
	repoRoot := repoRootForSync(t)
	return []registrySyncMirror{
		{
			name:       "local",
			registry:   filepath.Join(repoRoot, ".claude", "rules", "moai", "core", "zone-registry.md"),
			projectDir: repoRoot,
		},
		{
			name:       "template",
			registry:   filepath.Join(repoRoot, "internal", "template", "templates", ".claude", "rules", "moai", "core", "zone-registry.md"),
			projectDir: filepath.Join(repoRoot, "internal", "template", "templates"),
		},
	}
}

func TestRegistrySyncGuard(t *testing.T) {
	// REQ-ZRR-010 / D7 — guard clause FIRST: a bypassed validation is itself a
	// failure. Even if Validate's own Skipped return were removed upstream,
	// this direct environment check keeps the guard from passing under
	// MOAI_CONSTITUTION_SKIP_VALIDATE=1.
	if os.Getenv(skipValidateEnv) == "1" {
		t.Fatalf("validation skipped: %s=1 present in the test environment — the registry-sync guard must fail rather than pass (REQ-ZRR-010 / AC-ZRR-010)", skipValidateEnv)
	}

	mirrors := registrySyncMirrors(t)
	for _, m := range mirrors {
		t.Run(m.name, func(t *testing.T) {
			// --- 1. Production validator over this mirror (REQ-ZRR-008) ---
			result, err := constitution.Validate(constitution.ValidateOptions{
				RegistryPath: m.registry,
				ProjectDir:   m.projectDir,
			})
			if err != nil {
				t.Fatalf("Validate(%s mirror): %v", m.name, err)
			}
			if result.Skipped {
				t.Fatalf("validation was bypassed (MOAI_CONSTITUTION_SKIP_VALIDATE=1): the guard must not report success when validation was skipped (REQ-ZRR-010 / D7)")
			}
			if result.DriftCount != 0 || result.Status == constitution.ValidateStatusDrift {
				for _, e := range result.Entries {
					t.Errorf("validate [%s mirror]: [%s] %s @ %s %s — %s", m.name, e.SentinelKey, e.ID, e.File, e.Anchor, e.Detail)
				}
				t.Fatalf("validate [%s mirror]: drift/errors found (drift_count=%d)", m.name, result.DriftCount)
			}

			// --- 2. Registry data checks (anchor + literal + self-reference) ---
			reg, err := constitution.LoadRegistry(m.registry, m.projectDir)
			if err != nil {
				t.Fatalf("LoadRegistry(%s mirror): %v", m.name, err)
			}
			if len(reg.Entries) != wantRegistryEntries {
				t.Fatalf("[%s mirror] entry count = %d, want %d (REQ-ZRR-005: the repair must not add or delete entries; deliberate growth updates wantRegistryEntries in the same change)", m.name, len(reg.Entries), wantRegistryEntries)
			}

			registryRel, err := filepath.Rel(m.projectDir, m.registry)
			if err != nil {
				t.Fatalf("filepath.Rel: %v", err)
			}

			// sourceCache mirrors the validator's: one read per cited file.
			sourceCache := make(map[string]string)

			clauseChecked, retiredSkipped, anchorChecked := 0, 0, 0
			hitOnce, hitZero, hitMulti, selfReference := 0, 0, 0, 0
			for _, entry := range reg.Entries {
				// D13 — self-referential file: would satisfy its own clause by
				// definition and resolve anchors against the registry's ~50
				// headings, passing the guard with zero repair.
				if filepath.Clean(entry.File) == filepath.Clean(registryRel) {
					selfReference++
					t.Errorf("[%s mirror] %s: file: points at the registry itself (D13): %s", m.name, entry.ID, entry.File)
				}

				sourcePath := filepath.Join(m.projectDir, entry.File)
				raw, cached := sourceCache[sourcePath]
				if !cached {
					data, readErr := os.ReadFile(sourcePath) // #nosec G304 -- registry-controlled, project-scoped
					if readErr != nil {
						t.Fatalf("[%s mirror] %s: read source %s: %v", m.name, entry.ID, entry.File, readErr)
					}
					raw = string(data)
					sourceCache[sourcePath] = raw
				}

				// --- anchor check: ALL 101 entries, retired included (option C) ---
				if !anchorResolves(raw, entry.Anchor) {
					t.Errorf("[%s mirror] %s: anchor %q resolves to no heading in %s (six-step slug rule, REQ-ZRR-002/012)", m.name, entry.ID, entry.Anchor, entry.File)
				}
				anchorChecked++

				// --- clause check: live entries only (option C exemption) ---
				if constitution.IsRetiredClause(entry.Clause) {
					retiredSkipped++
					continue
				}
				n := literalHitCount(raw, entry.Clause)
				switch n {
				case 1:
					hitOnce++
				case 0:
					hitZero++
				default:
					hitMulti++
				}
				if n != 1 {
					t.Errorf("[%s mirror] %s: clause hits %d lines in %s, want exactly 1 (verbatim single-line span, AC-ZRR-002/003): %q", m.name, entry.ID, n, entry.File, truncateFor(entry.Clause, 60))
				}
				clauseChecked++
			}

			// --- 3. Evaluated-count assertion (AC-ZRR-007 partial-traversal mutant) ---
			if retiredSkipped != wantRetiredExempt {
				t.Errorf("[%s mirror] retired-skip count = %d, want %d (option-C audit records)", m.name, retiredSkipped, wantRetiredExempt)
			}
			if wantClause := wantRegistryEntries - wantRetiredExempt; clauseChecked != wantClause {
				t.Errorf("[%s mirror] clause checks completed = %d, want %d (partial traversal / exclusion list?)", m.name, clauseChecked, wantClause)
			}
			if anchorChecked != wantRegistryEntries {
				t.Errorf("[%s mirror] anchor checks completed = %d, want %d (partial traversal / exclusion list?)", m.name, anchorChecked, wantRegistryEntries)
			}
			// P4 (guard-failure-scenario.md): the passing output reports the two
			// counts separately per mirror — clause 97 / anchor 101.
			t.Logf("[%s mirror] evaluated: clause-checks=%d retired-skip=%d anchor-checks=%d of %d entries", m.name, clauseChecked, retiredSkipped, anchorChecked, len(reg.Entries))

			// Bucket assertion (plan.md §F M2 literal check): once=97 zero=0
			// multi=0 retired_exempt=4 self_reference=0 per mirror. The
			// per-entry errors above already name offenders; these pins carry
			// the measured shape into the passing output too, so a summary
			// reader sees the same buckets M1 measured.
			if hitZero != 0 || hitMulti != 0 || selfReference != 0 {
				t.Errorf("[%s mirror] literal buckets must be zero=0 multi=0 self_reference=0, got zero=%d multi=%d self_reference=%d", m.name, hitZero, hitMulti, selfReference)
			}
			t.Logf("[%s mirror] clause literal buckets: once=%d zero=%d multi=%d retired_exempt=%d self_reference=%d", m.name, hitOnce, hitZero, hitMulti, retiredSkipped, selfReference)
		})
	}

	// --- SKIP-env discipline (AC-ZRR-010, BOTH directions) ---
	// The mirror subtests above fatal on Validate's Skipped return; these
	// subtests pin that the Skipped return actually happens in both tree
	// states. The clean direction catches an implementation that ignores the
	// bypass entirely; the mutated direction catches "skip, then pass because
	// the tree happens to be clean" — both must end in the guard's explicit
	// skip failure, never a pass.
	t.Run("skip-env clean tree fails", func(t *testing.T) {
		t.Setenv(skipValidateEnv, "1")
		m := mirrors[0]
		result, err := constitution.Validate(constitution.ValidateOptions{
			RegistryPath: m.registry,
			ProjectDir:   m.projectDir,
		})
		if err != nil {
			t.Fatalf("Validate(clean tree, SKIP=1): %v", err)
		}
		if !result.Skipped {
			t.Fatalf("clean tree + SKIP=1: Validate did not report Skipped — the guard would PASS on a bypassed validation (REQ-ZRR-010 / AC-ZRR-010)")
		}
		t.Logf("clean tree + SKIP=1 → Validate Skipped=true — the guard fails on this, never passes")
	})
	t.Run("skip-env mutated tree still fails", func(t *testing.T) {
		t.Setenv(skipValidateEnv, "1")
		regPath := writeMutatedRegistryCopy(t, mirrors[0].registry)
		result, err := constitution.Validate(constitution.ValidateOptions{
			RegistryPath: regPath,
			ProjectDir:   filepath.Dir(regPath),
		})
		if err != nil {
			t.Fatalf("Validate(mutated tree, SKIP=1): %v", err)
		}
		if !result.Skipped {
			t.Fatalf("mutated tree + SKIP=1: Validate did not report Skipped — the guard could still evaluate (or wave through) a broken tree under a bypass (REQ-ZRR-010 / AC-ZRR-010)")
		}
		t.Logf("mutated tree + SKIP=1 → Validate Skipped=true — the guard still fails, never passes")
	})
}

// TestRegistrySyncMirrorsIdentical pins AC-ZRR-011: the local registry and the
// template registry are byte-identical, so every user `moai init` receives the
// same catalog this repository validates locally.
func TestRegistrySyncMirrorsIdentical(t *testing.T) {
	ms := registrySyncMirrors(t)
	local, err := os.ReadFile(ms[0].registry)
	if err != nil {
		t.Fatalf("read local mirror: %v", err)
	}
	template, err := os.ReadFile(ms[1].registry)
	if err != nil {
		t.Fatalf("read template mirror: %v", err)
	}
	if string(local) != string(template) {
		t.Fatalf("registry mirrors are not byte-identical (%d vs %d bytes) — repair one mirror only and the parity is gone (AC-ZRR-011)", len(local), len(template))
	}
	t.Logf("mirrors byte-identical: %d bytes", len(local))
}

// writeMutatedRegistryCopy copies the real registry into a temp dir with the
// R1 scenario mutation applied (guard-failure-scenario.md §1): one character
// inserted mid-span inside the quoted clause value of CONST-V3R2-004 — the
// smallest edit a real rules edit could make, never touching the YAML key or
// quoting. The clause text is read from the registry at run time, so the
// fixture tracks the entry through future rewordings and does not collide
// with a scratch mutation left on the tree.
func writeMutatedRegistryCopy(t *testing.T, srcRegistry string) string {
	t.Helper()
	data, err := os.ReadFile(srcRegistry)
	if err != nil {
		t.Fatalf("read registry for mutation copy: %v", err)
	}
	re := regexp.MustCompile(`(- id: CONST-V3R2-004\n(?:  [a-z_]+: [^\n]*\n)*?  clause: ")([^"\n]+)(")`)
	m := re.FindStringSubmatch(string(data))
	if m == nil {
		t.Fatalf("R1 mutation fixture drifted: no single-line double-quoted clause found for CONST-V3R2-004 in %s", srcRegistry)
	}
	clause := m[2]
	mutatedClause := clause[:len(clause)/2] + "x" + clause[len(clause)/2:]
	dir := t.TempDir()
	regPath := filepath.Join(dir, "zone-registry.md")
	out := strings.Replace(string(data), m[0], m[1]+mutatedClause+m[3], 1)
	if err := os.WriteFile(regPath, []byte(out), 0o600); err != nil {
		t.Fatalf("write mutated registry copy: %v", err)
	}
	return regPath
}

// headingSlug implements the six-step heading-to-slug rule declared by
// SPEC-ZONE-REGISTRY-RESYNC-001 spec.md §2.2 (REQ-ZRR-012):
//
//  1. lines inside code fences are excluded from heading candidacy
//  2. strip the leading '#' prefix and trim
//  3. strip backticks
//  4. lowercase
//  5. drop characters outside [a-z0-9 space '-']
//  6. collapse whitespace runs to a single '-' and prefix '#'
//
// This is the rule under which the 17 anchor failures were measured at guard
// landing time (independent analyzer .moai/reports/t232/analyze.py, tree
// e0afbb53c era, 2026-08-25; all 17 repaired in M1). The rule is declared here
// — not inferred — because a different rule yields a different failure count,
// so any change to it is a requirements change, not an implementation detail.
func headingSlug(headingLine string) string {
	h := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(headingLine), "#"))
	h = strings.ReplaceAll(h, "`", "")
	h = strings.ToLower(h)

	var b strings.Builder
	for _, r := range h {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '-':
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteRune(' ')
		}
	}

	return "#" + strings.Join(strings.Fields(b.String()), "-")
}

// anchorResolves reports whether anchor matches some heading slug of raw, the
// RAW file content (fence lines toggle heading candidacy off, mirroring the
// analyzer; the validator's stripCodeFences plays no part here).
func anchorResolves(raw, anchor string) bool {
	if anchor == "" {
		return false
	}
	inFence := false
	for ln := range strings.SplitSeq(raw, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "```") {
			inFence = !inFence
			continue
		}
		if !inFence && strings.HasPrefix(ln, "#") {
			if headingSlug(ln) == anchor {
				return true
			}
		}
	}
	return false
}

// literalHitCount is the grep -F -c equivalent: the number of LINES in raw
// that contain clause as a literal substring, with no normalization of either
// side. Exactly 1 is a real verbatim quote; 0 is a miss; 2+ means the clause
// is too short to be a quote. Strictly stronger than the validator's
// whitespace-normalized containment (AC-ZRR-002/003).
func literalHitCount(raw, clause string) int {
	if clause == "" {
		return 0
	}
	n := 0
	for ln := range strings.SplitSeq(raw, "\n") {
		if strings.Contains(ln, clause) {
			n++
		}
	}
	return n
}

func truncateFor(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
