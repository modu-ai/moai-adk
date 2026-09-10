package cli

// codex_launcher_guards_test.go — SPEC-CODEX-LAUNCHER-001 M3 static guards,
// widened to the full 12-file non-test codex CLI set by
// SPEC-CODEX-E2E-GUARD-001 M3/M4:
//
//   - AC-CL-013 — neutrality: every string/[]string field of codexCmd and
//     every flag usage string carries no internal-content marker and ZERO
//     non-ASCII characters
//   - AC-CL-014 — zero OS build tags, zero syscall imports, zero
//     process-replacement identifiers across ALL 12 non-test codex CLI files
//   - AC-CL-016 — source axis of the launched-executable closed set: every
//     process-start primitive hands its file's EXPECTED FIRST ARGUMENT
//     (per-file table — req.Program / binaryPath / "git"); the 9 files
//     outside the table carry ZERO process-start primitives; the shared tmux
//     diagnostic literal appears exactly once across non-test Go sources
//     (AC-CL-003)
//   - REQ-CEG-010 — string-literal neutrality of the 12-file set's sources
//     (same forbidden-pattern classes as AC-CL-013, comments excluded via
//     go/ast, with a positive-control canary)
//
// Surface reconciliation (no double-claim): AC-CL-007 owns the
// sentinel-VALUE classification of the launch surface (command strings the
// cobra walk collects); the REQ-CEG-010 literal scan owns the string
// LITERALS of these files' sources. Different surfaces — values walked from
// the command object vs text scanned from the source files.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// codexGuardFiles is the non-test codex CLI file set the static judgments
// sweep (SPEC-CODEX-E2E-GUARD-001 M3, widened from SPEC-CODEX-LAUNCHER-001's
// original 2-file codexSpecFiles). mcp_codex.go's diff parts are covered by
// the AC-CL-007 closed-set judgment elsewhere.
//
// The set is length-pinned by sweepGuardFiles: the population is exactly 12
// (independently cross-checked at run phase: `ls internal/cli/*codex*.go |
// grep -v _test | wc -l` → 12). A future codex CLI file REDs the swept-count
// assertion and forces a deliberate update here — that friction is the guard
// working (acceptance.md §D.2 edge 6).
var codexGuardFiles = []string{
	"codex_contract.go",
	"codex_init.go",
	"codex_job_control.go",
	"codex_jobs.go",
	"codex_launcher.go",
	"codex_readiness.go",
	"codex_review_gate.go",
	"codex_task.go",
	"doctor_codex.go",
	"hook_harness_codex.go",
	"mcp_codex.go",
	"update_codex_wiring.go",
}

// sweepGuardFiles pins the swept count before any verdict is read: a green
// over an accidentally-short set is a vacuous green (verification-
// completeness §1.1 — a pass whose swept set is empty asserts nothing).
func sweepGuardFiles(t *testing.T) []string {
	t.Helper()
	if len(codexGuardFiles) != 12 {
		t.Fatalf("guard swept %d files, want 12 — vacuous green", len(codexGuardFiles))
	}
	return codexGuardFiles
}

// codexNeutralityPatterns is the forbidden-pattern class table of the
// template isolation doctrine (SPEC- ids, REQ- ids, card ids, ISO dates,
// long hex SHAs, absolute home paths, CLAUDE.local, .moai/reports). Both
// neutrality scans (the AC-CL-013 cobra-command scan and the REQ-CEG-010
// literal scan) judge against THIS one table.
func codexNeutralityPatterns() []struct {
	name string
	re   *regexp.Regexp
} {
	return []struct {
		name string
		re   *regexp.Regexp
	}{
		{"SPEC- id", regexp.MustCompile(`SPEC-`)},
		{"REQ- id", regexp.MustCompile(`REQ-`)},
		{"card id (t+digits)", regexp.MustCompile(`\bt[0-9]+\b`)},
		{"ISO date", regexp.MustCompile(`\b[0-9]{4}-[0-9]{2}-[0-9]{2}\b`)},
		{"commit SHA (7+ hex)", regexp.MustCompile(`\b[0-9a-f]{7,}\b`)},
		{"absolute home path /Users/", regexp.MustCompile(`/Users/`)},
		{"absolute home path /home/", regexp.MustCompile(`/home/`)},
		{"CLAUDE.local", regexp.MustCompile(`CLAUDE\.local`)},
		{".moai/reports", regexp.MustCompile(`\.moai/reports`)},
	}
}

func codexReadSpecFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}

// ─── AC-CL-013 — neutrality ────────────────────────────────────────────────

// codexCommandStrings walks codexCmd (and any subcommands — there are none,
// but the walk makes that structural) collecting every string and []string
// field the cobra surface can carry, plus every flag's usage string.
func codexCommandStrings(t *testing.T) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		v := reflect.ValueOf(c).Elem()
		typ := v.Type()
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if !f.IsExported() {
				continue
			}
			switch f.Type.Kind() {
			case reflect.String:
				out[c.Name()+"."+f.Name] = append(out[c.Name()+"."+f.Name], v.Field(i).String())
			case reflect.Slice:
				if f.Type.Elem().Kind() == reflect.String {
					out[c.Name()+"."+f.Name] = append(out[c.Name()+"."+f.Name], v.Field(i).Interface().([]string)...)
				}
			}
		}
		c.Flags().VisitAll(func(fl *pflag.Flag) {
			out[c.Name()+".flag:"+fl.Name+".usage"] = append(out[c.Name()+".flag:"+fl.Name+".usage"], fl.Usage)
		})
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(codexCmd)
	return out
}

// TestCodexCommand_NeutralityScan — the forbidden-pattern classes of the
// template isolation doctrine and the non-ASCII count are all ZERO across
// the collected command strings.
func TestCodexCommand_NeutralityScan(t *testing.T) {
	for where, values := range codexCommandStrings(t) {
		for _, val := range values {
			for _, p := range codexNeutralityPatterns() {
				if p.re.MatchString(val) {
					t.Errorf("%s = %q matches forbidden pattern %q", where, val, p.name)
				}
			}
			for i, r := range val {
				if r > 127 {
					t.Errorf("%s contains non-ASCII rune %q at %d: %q", where, r, i, val)
				}
			}
		}
	}
}

// ─── AC-CL-014 — build tags and syscall ────────────────────────────────────

// TestCodexSpecFiles_NoBuildTagsOrSyscall — zero OS build tags (no
// windows/darwin/linux/unix tokens in //go:build lines), zero syscall
// imports, zero process-replacement identifiers across all 12 non-test codex
// CLI files. The only way to satisfy this alongside the Windows compile
// gates is platform-common APIs.
func TestCodexSpecFiles_NoBuildTagsOrSyscall(t *testing.T) {
	buildTag := regexp.MustCompile(`(?m)^//go:build\s+(.*)$`)
	osToken := regexp.MustCompile(`\b(windows|darwin|linux|unix)\b`)
	for _, name := range sweepGuardFiles(t) {
		src := codexReadSpecFile(t, name)
		for _, m := range buildTag.FindAllStringSubmatch(src, -1) {
			if osToken.MatchString(m[1]) {
				t.Errorf("%s: OS build tag %q", name, m[0])
			}
		}
		if strings.Contains(src, `"syscall"`) {
			t.Errorf("%s: syscall import", name)
		}
		for _, ident := range []string{"syscall.Exec", "unix.Exec", "golang.org/x/sys/unix"} {
			if strings.Contains(src, ident) {
				t.Errorf("%s: process-replacement identifier %q", name, ident)
			}
		}
		if strings.HasSuffix(name, "_windows.go") || strings.HasSuffix(name, "_unix.go") || strings.HasSuffix(name, "_darwin.go") {
			t.Errorf("%s: GOOS-suffixed file", name)
		}
	}
}

// ─── AC-CL-016 axis 2 — process-start primitives ───────────────────────────

// codexExecFirstArg is the per-file expected-first-argument table (spec.md
// §A; every row's call site was re-read at run phase): each listed file's
// process-start primitives must hand the named executable expression as
// FIRST argument. A blanket req.Program expectation would false-RED on
// codex_review_gate.go's legitimate `git` porcelain probe, and mcp_codex.go
// runs the codex binary under the binaryPath parameter at three sites —
// hence the per-file rows.
var codexExecFirstArg = map[string]string{
	"codex_launcher.go":    "req.Program",
	"mcp_codex.go":         "binaryPath",
	"codex_review_gate.go": `"git"`,
}

// TestCodexSpecFiles_ExecPrimitivesCodexOnly — over all 12 non-test codex
// CLI files, every process-start primitive call hands its file's expected
// first argument (see codexExecFirstArg). The 9 files OUTSIDE the table are
// pinned to ZERO process-start primitives — a positive assertion, so any new
// primitive in them must add a table row (that friction is the guard
// working). A tmux/open/other executable introduced in these files dies on
// this scan; spawn.go owns the tmux primitive and is not in the set.
func TestCodexSpecFiles_ExecPrimitivesCodexOnly(t *testing.T) {
	// exec.CommandContext takes ctx FIRST and the binary SECOND, so its
	// binary argument is captured as the second group; plain exec.Command
	// and os.StartProcess take the binary first.
	callCtx := regexp.MustCompile(`\bexec\.CommandContext\s*\(\s*[^,]+,\s*([^,)]*)`)
	callPlain := regexp.MustCompile(`\b(?:exec\.Command|os\.StartProcess)\s*\(([^,)]*)`)
	for _, name := range sweepGuardFiles(t) {
		src := codexReadSpecFile(t, name)
		var matches [][]string
		for _, line := range strings.Split(src, "\n") {
			// Comment-only lines are skipped: the doc comments mention the
			// tmux primitive spawn.go owns (codex_launcher.go:136,
			// mcp_codex.go:339); only real call sites judge.
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			if m := callCtx.FindStringSubmatch(line); m != nil {
				matches = append(matches, m)
			} else if m := callPlain.FindStringSubmatch(line); m != nil {
				matches = append(matches, m)
			}
		}
		want, hasRow := codexExecFirstArg[name]
		if hasRow {
			// The row is the EXPECTATION, not an allowance: a table row
			// whose file has lost every call site must also fail — otherwise
			// a file could drift to zero sites (or to a differently-shaped
			// one) and the stale row would pass vacuously. This generalizes
			// t495's launcher floor (3840ff028): a zero-match scan on a
			// table row is exactly the vacuous-green shape that floor
			// guarded, now enforced for every table row.
			if len(matches) == 0 {
				t.Errorf("%s: expected-first-argument row %s found no process-start primitive — the table must track the source (a zero-match scan is the vacuous-green shape t495's floor guarded)", name, want)
				continue
			}
			for _, m := range matches {
				if first := strings.TrimSpace(m[1]); first != want {
					t.Errorf("%s: process-start first argument %q is not the expected %s", name, first, want)
				}
			}
			continue
		}
		// Zero-call-site files are positive assertions — the t495-era
		// "codex_readiness.go starts processes" check generalized to every
		// file outside the table.
		if len(matches) != 0 {
			t.Errorf("%s starts processes but has no expected-first-argument row in codexExecFirstArg: %v — update the table", name, matches)
		}
	}
}

// TestCodexSpawn_TmuxDiagnosticSingleSource — the shared tmux-absent
// diagnostic literal appears EXACTLY once across non-test Go sources of this
// package: a copied second instance would pass a by-eye byte comparison and
// fail single-source (AC-CL-003).
func TestCodexSpawn_TmuxDiagnosticSingleSource(t *testing.T) {
	needle := "tmux session required for"
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	hits := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src := codexReadSpecFile(t, f)
		hits += strings.Count(src, needle)
	}
	if hits != 1 {
		t.Errorf("shared tmux diagnostic literal appears %d times across non-test Go sources, want exactly 1", hits)
	}
}

// ─── REQ-CEG-010 — string-literal neutrality over the 12-file set ──────────

// codexGuardFileLiterals extracts the string literals (interpreted and raw,
// comments excluded — comments are not BasicLit nodes) of one file via
// go/parser. Returns the literal values with their 1-based line numbers.
func codexGuardFileLiterals(t *testing.T, name string) []struct {
	line int
	val  string
} {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	var out []struct {
		line int
		val  string
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			out = append(out, struct {
				line int
				val  string
			}{fset.Position(lit.Pos()).Line, lit.Value})
		}
		return true
	})
	return out
}

// codexLiteralNeutralityPatterns is the NARROWED class set the literal scan
// judges (plan.md M4.2: baseline measured first, then narrowed with the
// justification recorded — never by weakening TestCodexCommand_NeutralityScan).
//
// Kept from codexNeutralityPatterns — the internal-content-LEAKAGE classes,
// all observed CLEAN on the 12-file baseline (run phase, this tree):
// SPEC- id, REQ- id, card id (t+digits), ISO date, commit SHA (7+ hex),
// /Users/, /home/.
//
// Narrowed OUT, with the baseline observations that forced each narrowing:
//
//   - "CLAUDE.local" — codex_contract.go:31/37 carry the literals
//     "CLAUDE.local.md" / "@CLAUDE.local.md": the contract check itself
//     NAMES the dev-only files it inspects. In source, referencing internal
//     paths is the implementation's job; the neutrality concern is the
//     DISTRIBUTED template content, which the AC-CL-013 command-surface scan
//     still judges with the full table.
//   - ".moai/reports" — codex_review_gate.go:38 defines the constant
//     ".moai/reports/": that IS the runtime-managed-path filter the review
//     gate must exclude from reviewable changes. Same reasoning as above.
//   - non-ASCII count — the baseline carries deliberate bilingual runtime
//     copy (codex_task.go / mcp_codex.go Korean progress messages) and
//     em-dash typography across error strings. moai's user-facing runtime
//     copy is intentionally localized (ko/en); the non-ASCII ZERO stays
//     owned by AC-CL-013 on the cobra command surface (Use/Short/flag
//     usage), which is English by contract.
func codexLiteralNeutralityPatterns() []struct {
	name string
	re   *regexp.Regexp
} {
	return []struct {
		name string
		re   *regexp.Regexp
	}{
		{"SPEC- id", regexp.MustCompile(`SPEC-`)},
		{"REQ- id", regexp.MustCompile(`REQ-`)},
		{"card id (t+digits)", regexp.MustCompile(`\bt[0-9]+\b`)},
		{"ISO date", regexp.MustCompile(`\b[0-9]{4}-[0-9]{2}-[0-9]{2}\b`)},
		{"commit SHA (7+ hex)", regexp.MustCompile(`\b[0-9a-f]{7,}\b`)},
		{"absolute home path /Users/", regexp.MustCompile(`/Users/`)},
		{"absolute home path /home/", regexp.MustCompile(`/home/`)},
	}
}

// TestCodexCommand_GuardFileLiteralsNeutral — the narrowed forbidden-pattern
// classes (codexLiteralNeutralityPatterns — the leakage classes of the
// AC-CL-013 table, narrowed per plan.md M4.2) are all ZERO across the string
// LITERALS of the 12-file guard set's sources. This is the source-text
// complement of the cobra-command scan: AC-CL-007 owns sentinel-VALUE
// classification, this owns literal neutrality — no double-claim.
//
// Positive-control canary: the scan must OBSERVE the known existing literal
// `"git"` in codex_review_gate.go (the porcelain probe at hasReviewableChanges).
// A scanning failure that finds nothing reads RED here, not green — a pass
// whose swept set is empty asserts nothing (verification-completeness §1.1).
func TestCodexCommand_GuardFileLiteralsNeutral(t *testing.T) {
	sawCanary := false
	total := 0
	for _, name := range sweepGuardFiles(t) {
		for _, lit := range codexGuardFileLiterals(t, name) {
			total++
			if name == "codex_review_gate.go" && lit.val == `"git"` {
				sawCanary = true
			}
			for _, p := range codexLiteralNeutralityPatterns() {
				if p.re.MatchString(lit.val) {
					t.Errorf("%s:%d literal %s matches forbidden pattern %q", name, lit.line, lit.val, p.name)
				}
			}
		}
	}
	if total == 0 {
		t.Errorf("literal scan swept 0 literals across %d files — scanning failed, the green would be vacuous", len(sweepGuardFiles(t)))
	}
	if !sawCanary {
		t.Errorf("positive-control canary not observed: the scan did not see the \"git\" literal in codex_review_gate.go (hasReviewableChanges) — a scanning failure must read RED, not green")
	}
}
