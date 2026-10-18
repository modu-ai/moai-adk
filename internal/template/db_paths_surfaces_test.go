// db_paths_surfaces_test.go — SPEC-MOAI-DB-PATHS-001 (card t1827) M3/M4: the
// instruction surfaces that name a database path must tell the reader to ask
// `moai paths` instead of guessing, and the foreman queue watch must obtain its
// directory from that command and stop visibly when it fails.
//
// Paths are relative to this package directory; the repository root is ../..
package template

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// dbPathsLeadPhrases are the five fixed lead phrases of the instruction block.
// Each must appear exactly once per surface, as a single-line substring.
var dbPathsLeadPhrases = []string{
	"Find the database files with ",
	"Forbidden guessed paths:",
	"Never `sqlite3`-open a guessed path",
	"Direct reads only on a snapshot copy",
	"Prefer the CLI:",
}

// dbPathsPointer is the pointer sentence the drift guard keys on. It contains
// neither the old nor the new queue-command name (card t1641 ordering).
const dbPathsPointer = "Find the database files with `moai paths`"

// dbPathsSurfaces is the twelve-file set S12 of SPEC-MOAI-DB-PATHS-001 § C,
// as paths relative to this package directory.
var dbPathsSurfaces = []string{
	// local tree
	"../../.claude/agents/moai/manager-todo.md",
	"../../.claude/skills/moai-factory-foreman/SKILL.md",
	"../../.claude/skills/moai/SKILL.md",
	"../../.claude/skills/moai/workflows/gtd.md",
	"../../.moai/docs/todo-queue-storage.md",
	"../../.moai/workflows/gtd.md",
	// template tree (the generated Codex TOML is regenerated, never edited)
	"templates/.claude/agents/moai/manager-todo.md",
	"templates/.claude/skills/moai-factory-foreman/SKILL.md",
	"templates/.claude/skills/moai/SKILL.md",
	"templates/.claude/skills/moai/workflows/gtd.md",
	"templates/.moai/docs/todo-queue-storage.md",
	"templates/.codex/agents/moai/manager-todo.toml",
}

// surfaceViolations reports each lead phrase that does not appear exactly once.
func surfaceViolations(content string) []string {
	var bad []string
	for _, phrase := range dbPathsLeadPhrases {
		if n := strings.Count(content, phrase); n != 1 {
			bad = append(bad, phrase+" (count "+strconv.Itoa(n)+")")
		}
	}
	return bad
}

func TestDBPathsBlock_Surfaces(t *testing.T) {
	for _, path := range dbPathsSurfaces {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read surface %s: %v", path, err)
			continue
		}
		if bad := surfaceViolations(string(raw)); len(bad) > 0 {
			t.Errorf("%s lacks the instruction block: %v", path, bad)
		}
	}
	t.Logf("surfaces=%d phrases=%d", len(dbPathsSurfaces), len(dbPathsLeadPhrases))

	// The mutants start from a synthetic well-formed block, so they fail or
	// pass for their own reason whatever state the surfaces are in.
	wellFormed := strings.Join([]string{
		"Find the database files with `moai paths`; never guess them.",
		"Forbidden guessed paths: none listed here.",
		"Never `sqlite3`-open a guessed path: opening a missing file creates an empty one.",
		"Direct reads only on a snapshot copy of `.db`, `.db-wal`, and `.db-shm` together.",
		"Prefer the CLI: `moai todo list --json`.",
	}, "\n")
	if bad := surfaceViolations(wellFormed); len(bad) != 0 {
		t.Fatalf("the synthetic block must pass the check: %v", bad)
	}
	t.Run("mutant-forbidden-line-deleted", func(t *testing.T) {
		var kept []string
		for _, line := range strings.Split(wellFormed, "\n") {
			if !strings.Contains(line, "Forbidden guessed paths:") {
				kept = append(kept, line)
			}
		}
		bad := surfaceViolations(strings.Join(kept, "\n"))
		if len(bad) != 1 || !strings.Contains(bad[0], "Forbidden guessed paths:") {
			t.Errorf("the surface check did not report the deleted forbidden-paths line: %v", bad)
		}
	})
	t.Run("mutant-duplicated-phrase", func(t *testing.T) {
		bad := surfaceViolations(wellFormed + "\nPrefer the CLI: again\n")
		if len(bad) != 1 || !strings.Contains(bad[0], "Prefer the CLI:") {
			t.Errorf("a phrase appearing twice must be reported once: %v", bad)
		}
	})
}

// ---------------------------------------------------------------------------
// REQ-DBP-014 — drift guard over the embedded template text
// ---------------------------------------------------------------------------

// dbPathMention matches a `<project-key>` database path of the todo, factory,
// or broker store.
var dbPathMention = regexp.MustCompile("<project-key>[^\n]*(backlog|factory|broker)\\.db")

// dbPathPointerAllowlist exempts a template file that names such a path
// without being an instruction surface. Every entry carries a reason.
var dbPathPointerAllowlist = map[string]string{}

// pointerViolations reports each file that names a `<project-key>` database
// path, lacks the pointer sentence, and is not on the allowlist.
func pointerViolations(files map[string]string, allow map[string]string) []string {
	var bad []string
	for path, content := range files {
		if _, ok := allow[path]; ok {
			continue
		}
		if dbPathMention.MatchString(content) && !strings.Contains(content, dbPathsPointer) {
			bad = append(bad, path)
		}
	}
	sort.Strings(bad)
	return bad
}

// templateTextFiles reads every text file below templates/.
func templateTextFiles(t *testing.T) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir("templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if utf8.Valid(raw) && !bytes.Contains(raw, []byte{0}) {
			files[filepath.ToSlash(path)] = string(raw)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk templates: %v", err)
	}
	return files
}

func TestTemplateDBPathSurfacesCarryPointer(t *testing.T) {
	files := templateTextFiles(t)

	t.Run("tree", func(t *testing.T) {
		if bad := pointerViolations(files, dbPathPointerAllowlist); len(bad) > 0 {
			t.Errorf("template files name a <project-key> database path without the pointer sentence %q: %v", dbPathsPointer, bad)
		}
		for path, reason := range dbPathPointerAllowlist {
			if strings.TrimSpace(reason) == "" {
				t.Errorf("allowlist entry %s has no reason", path)
			}
			if _, ok := files[path]; !ok {
				t.Errorf("allowlist entry %s names no template file", path)
			}
		}
		// The guard must be able to see the surfaces it protects.
		if !dbPathMention.MatchString(files["templates/.claude/agents/moai/manager-todo.md"]) {
			t.Errorf("the detector no longer recognises the manager-todo database mention")
		}
	})
	t.Run("mutant-pointer-removed", func(t *testing.T) {
		const target = "templates/.claude/agents/moai/manager-todo.md"
		// home-guard:allow synthetic template text asserted as a string, never resolved
		intact := "State: `~/.moai/db/<project-key>/todo/backlog.db`.\n" + dbPathsPointer + "; never guess them.\n"
		if bad := pointerViolations(map[string]string{target: intact}, nil); len(bad) != 0 {
			t.Fatalf("a file carrying the pointer must pass: %v", bad)
		}
		mutated := map[string]string{target: strings.ReplaceAll(intact, dbPathsPointer, "Look for them")}
		if bad := pointerViolations(mutated, nil); len(bad) != 1 || bad[0] != target {
			t.Errorf("the guard did not name %s: %v", target, bad)
		}
	})
	t.Run("mutant-new-file", func(t *testing.T) {
		const target = "templates/.claude/skills/new/SKILL.md"
		mutated := map[string]string{target: "State lives in `<project-key>/todo/backlog.db`.\n"}
		if bad := pointerViolations(mutated, nil); len(bad) != 1 || bad[0] != target {
			t.Errorf("the guard did not name %s: %v", target, bad)
		}
	})
}

// ---------------------------------------------------------------------------
// REQ-DBP-012 — the foreman queue watch obtains its directory from the command
// ---------------------------------------------------------------------------

// foremanWatchBlock returns the first fenced sh block after the Queue watch
// step of a foreman SKILL.md, with the fence indentation stripped.
func foremanWatchBlock(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(string(raw), "\n")
	start := -1
	for i, ln := range lines {
		if strings.Contains(ln, "**Queue watch.**") {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("%s: no Queue watch step", path)
	}
	open, indent := -1, ""
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "```sh" {
			open, indent = i, lines[i][:strings.Index(lines[i], "```")]
			break
		}
	}
	if open < 0 {
		t.Fatalf("%s: no sh block after the Queue watch step", path)
	}
	var body []string
	for i := open + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "```" {
			return strings.Join(body, "\n")
		}
		body = append(body, strings.TrimPrefix(lines[i], indent))
	}
	t.Fatalf("%s: unterminated sh block", path)
	return ""
}

func TestForemanWatch_FailsVisibly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the queue watch is a POSIX sh fragment")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skipf("no sh on PATH: %v", err)
	}
	for name, path := range map[string]string{
		"local":    "../../.claude/skills/moai-factory-foreman/SKILL.md",
		"template": "templates/.claude/skills/moai-factory-foreman/SKILL.md",
	} {
		t.Run(name, func(t *testing.T) {
			block := foremanWatchBlock(t, path)
			if strings.Contains(block, "sha256sum") || strings.Contains(block, "shasum") {
				t.Errorf("the watch still computes the project key by hand")
			}
			dir := t.TempDir()
			stubDir := filepath.Join(dir, "bin")
			if err := os.MkdirAll(stubDir, 0o700); err != nil {
				t.Fatal(err)
			}
			stub := "#!/bin/sh\nexit 2\n"
			if err := os.WriteFile(filepath.Join(stubDir, "moai"), []byte(stub), 0o700); err != nil {
				t.Fatal(err)
			}
			script := filepath.Join(dir, "watch.sh")
			if err := os.WriteFile(script, []byte(block+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			// The pre-repair watch never exits on its own; the deadline turns that
			// into a failure instead of a hang.
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "sh", script)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "PATH="+stubDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			var out bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &out
			err := cmd.Run()
			if ctx.Err() != nil {
				t.Errorf("the watch kept running although `moai paths` failed (output %q)", out.String())
			} else if err == nil {
				t.Errorf("the watch exited 0 although `moai paths` failed (output %q)", out.String())
			}
			if strings.Contains(out.String(), "backlog changed") {
				t.Errorf("the watch emitted a change event without a queue directory: %q", out.String())
			}
			if strings.TrimSpace(out.String()) == "" {
				t.Errorf("the watch stopped silently; the failure must be visible")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// REQ-DBP-013 — the four READMEs and eight docs-site pages
// ---------------------------------------------------------------------------

// dbPathsUserFacingAnchors are language-neutral anchors in code spans; the
// locale prose around them is judged by review, their presence by this test.
var dbPathsUserFacingAnchors = []string{
	"moai paths",
	// home-guard:allow the anchor is the forbidden guessed path, asserted as text and never resolved
	"~/.moai/db/todo.db",
	"sqlite3",
	".db-wal",
	"moai todo list --json",
}

var dbPathsUserFacingPages = func() []string {
	pages := []string{"../../README.md", "../../README.ko.md", "../../README.ja.md", "../../README.zh.md"}
	for _, l := range []string{"en", "ko", "ja", "zh"} {
		for _, p := range []string{"advanced/factory-mode.md", "utility-commands/moai-todo.md"} {
			pages = append(pages, "../../docs-site/content/"+l+"/"+p)
		}
	}
	return pages
}()

// anchorsMissingInSection reports the anchors absent from the heading-delimited
// section that holds the page's pre-existing database mention (the section
// containing `project-key` or `backlog.db`). A page with no such section
// reports every anchor.
func anchorsMissingInSection(content string) []string {
	lines := strings.Split(content, "\n")
	mention := -1
	for i, ln := range lines {
		if strings.Contains(ln, "project-key") || strings.Contains(ln, "backlog.db") {
			mention = i
			break
		}
	}
	if mention < 0 {
		return append([]string(nil), dbPathsUserFacingAnchors...)
	}
	isHeading := func(ln string) bool { return strings.HasPrefix(ln, "#") && strings.Contains(ln, " ") }
	start := 0
	for i := mention; i >= 0; i-- {
		if isHeading(lines[i]) {
			start = i
			break
		}
	}
	end := len(lines)
	for i := mention + 1; i < len(lines); i++ {
		if isHeading(lines[i]) {
			end = i
			break
		}
	}
	section := strings.Join(lines[start:end], "\n")
	var missing []string
	for _, a := range dbPathsUserFacingAnchors {
		if !strings.Contains(section, a) {
			missing = append(missing, a)
		}
	}
	return missing
}

func TestDBPathsBlock_UserFacingPages(t *testing.T) {
	for _, path := range dbPathsUserFacingPages {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read page %s: %v", path, err)
			continue
		}
		if missing := anchorsMissingInSection(string(raw)); len(missing) > 0 {
			t.Errorf("%s: the section holding the database mention lacks the anchors %v", path, missing)
		}
	}
	t.Logf("pages=%d anchors=%d", len(dbPathsUserFacingPages), len(dbPathsUserFacingAnchors))

	t.Run("mutant-anchor-dropped", func(t *testing.T) {
		// home-guard:allow synthetic page text asserted as a string, never resolved against a home
		page := "## Storage\n\nThe queue is `backlog.db`. Use `moai paths`, never `~/.moai/db/todo.db`; `sqlite3` creates files; copy `.db-wal`; `moai todo list --json`.\n"
		if m := anchorsMissingInSection(page); len(m) != 0 {
			t.Fatalf("the synthetic page must pass: %v", m)
		}
		mutated := strings.ReplaceAll(page, ".db-wal", ".wal")
		if m := anchorsMissingInSection(mutated); len(m) != 1 || m[0] != ".db-wal" {
			t.Errorf("the check did not report the dropped .db-wal anchor: %v", m)
		}
	})
	t.Run("mutant-anchors-in-other-section", func(t *testing.T) {
		// home-guard:allow synthetic page text asserted as a string, never resolved against a home
		page := "## Storage\n\nThe queue is `backlog.db`.\n\n## Elsewhere\n\n`moai paths` `~/.moai/db/todo.db` `sqlite3` `.db-wal` `moai todo list --json`\n"
		if m := anchorsMissingInSection(page); len(m) != len(dbPathsUserFacingAnchors) {
			t.Errorf("anchors outside the mention's section must not count: missing %v", m)
		}
	})
}
