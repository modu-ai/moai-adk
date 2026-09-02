package cli

// codex_launcher_guards_test.go — SPEC-CODEX-LAUNCHER-001 M3 static guards:
//
//   - AC-CL-013 — neutrality: every string/[]string field of codexCmd and
//     every flag usage string carries no internal-content marker and ZERO
//     non-ASCII characters
//   - AC-CL-014 — zero OS build tags, zero syscall imports, zero
//     process-replacement identifiers in this SPEC's files
//   - AC-CL-016 — source axis of the launched-executable closed set: every
//     process-start primitive in this SPEC's files hands the codex path
//     variable as its first argument; the shared tmux diagnostic literal
//     appears exactly once across non-test Go sources (AC-CL-003)

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// codexSpecFiles is the "files added by this SPEC" set the static judgments
// scope to (acceptance 판정 어휘): the two new files. mcp_codex.go's diff
// parts are covered by the AC-CL-007 closed-set judgment elsewhere.
var codexSpecFiles = []string{"codex_launcher.go", "codex_readiness.go"}

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
// template isolation doctrine (SPEC- ids, REQ- ids, card ids, ISO dates,
// long hex SHAs, absolute home paths, CLAUDE.local, .moai/reports) and the
// non-ASCII count are all ZERO across the collected command strings.
func TestCodexCommand_NeutralityScan(t *testing.T) {
	patterns := []struct {
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
	for where, values := range codexCommandStrings(t) {
		for _, val := range values {
			for _, p := range patterns {
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
// imports, zero process-replacement identifiers. The only way to satisfy
// this alongside the Windows compile gates is platform-common APIs.
func TestCodexSpecFiles_NoBuildTagsOrSyscall(t *testing.T) {
	buildTag := regexp.MustCompile(`(?m)^//go:build\s+(.*)$`)
	osToken := regexp.MustCompile(`\b(windows|darwin|linux|unix)\b`)
	for _, name := range codexSpecFiles {
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

// TestCodexSpecFiles_ExecPrimitivesCodexOnly — every process-start primitive
// call in this SPEC's files hands the codex path variable as its FIRST
// argument (here: req.Program, the launch request's resolved binary). A
// tmux/open/other executable introduced in these files dies on this scan;
// spawn.go owns the tmux primitive and is not this SPEC's file.
func TestCodexSpecFiles_ExecPrimitivesCodexOnly(t *testing.T) {
	call := regexp.MustCompile(`\b(?:exec\.Command|exec\.CommandContext|os\.StartProcess)\s*\(([^,)]*)`)
	for _, name := range codexSpecFiles {
		src := codexReadSpecFile(t, name)
		var matches [][]string
		for _, line := range strings.Split(src, "\n") {
			// Comment-only lines are skipped: the doc comments mention the
			// tmux primitive spawn.go owns; only real call sites judge.
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			if m := call.FindStringSubmatch(line); m != nil {
				matches = append(matches, m)
			}
		}
		if name == "codex_readiness.go" && len(matches) != 0 {
			t.Errorf("codex_readiness.go starts processes: %v", matches)
		}
		for _, m := range matches {
			first := strings.TrimSpace(m[1])
			if first != "req.Program" {
				t.Errorf("%s: process-start first argument %q is not the codex path variable (req.Program)", name, first)
			}
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
