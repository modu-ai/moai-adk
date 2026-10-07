package cli

// codex_contract_link_test.go — SPEC-CODEX-INIT-001 M5+M6 cells:
// AC-CI-005 (creation), AC-CI-006 (three-run idempotency), AC-CI-007
// (local-file separation).
//
// The AGENTS.md-primary product reshapes the contract: AGENTS.md is the sole
// instruction file (created when absent, byte-untouched when present), the
// local file AGENTS.local.md is never written — Codex reads it through the
// launcher's developer_instructions injection — and a legacy CLAUDE.md is
// INERT: byte-untouched whether or not it carries the old import lines.
//
// Disciplines: fixture-specific EXPECTED BYTE SEQUENCES compared for FULL
// equality on every cell that touches an existing file (partial-substring
// checks pass an implementation that appends a provenance block it was
// never asked for); renames counted PER FILE (a cell writing both files
// must rename twice); idempotency proven by 1↔2 AND 2↔3 byte comparisons
// (a run that rewrites once and then stabilizes is not idempotent); the
// local file isolation proven by a closure WALK over executing imports,
// never by a filename grep.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const codexSentinelLocal = "SENTINEL-LOCAL-7q7"
const codexTestLocalImportDirective = "@AGENTS.local.md"

// ─── fixtures — instruction files (acceptance common fixture table) ───────

type codexLinkFixture struct {
	name   string
	agents []byte // nil = absent
	claude []byte // nil = absent (legacy file; the contract must leave it inert)
	local  []byte // nil = absent (all I-fixtures; L1 sets it)
	// expectations
	wantAgentsRenames int
	// full expected byte sequences for cells that TOUCH an existing file
	// (nil = the file is created, judged by the lighter creation assertions)
	wantAgentsExact []byte
}

var userAgentsBody = []byte("# user agents notes\n\ninstruction prose for agents\n")
var userClaudeBody = []byte("# user claude notes\n\ninstruction prose for claude\n")

func codexLinkFixtures() []codexLinkFixture {
	return []codexLinkFixture{
		{
			name:              "i1_both_absent",
			wantAgentsRenames: 1,
		},
		{
			name:              "i2_agents_only",
			agents:            userAgentsBody,
			wantAgentsRenames: 0,
			wantAgentsExact:   userAgentsBody,
		},
		{
			// Legacy inertness: an existing CLAUDE.md is never created,
			// linked, or rewritten — AGENTS.md is the sole instruction file.
			name:              "i3_claude_only",
			claude:            userClaudeBody,
			wantAgentsRenames: 1,
		},
		{
			name:              "i4_agents_and_claude",
			agents:            userAgentsBody,
			claude:            userClaudeBody,
			wantAgentsRenames: 0,
			wantAgentsExact:   userAgentsBody,
		},
		{
			// A legacy CLAUDE.md carrying the old two-import shape stays
			// byte-for-byte as the user left it.
			name:              "i5_claude_legacy_links",
			agents:            userAgentsBody,
			claude:            []byte("# title\n\n@AGENTS.md\n\nbody\n@AGENTS.local.md\n"),
			wantAgentsRenames: 0,
			wantAgentsExact:   userAgentsBody,
		},
	}
}

// codexLayLinkFixture writes the fixture files into proj and returns the
// AGENTS.local.md bytes (nil when the fixture has none).
func codexLayLinkFixture(t *testing.T, proj string, fx codexLinkFixture) []byte {
	t.Helper()
	for name, content := range map[string][]byte{
		codexAgentsRelPath: fx.agents, "CLAUDE.md": fx.claude, codexLocalInstructionName: fx.local,
	} {
		if content == nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(proj, name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return fx.local
}

func codexRenamesOf(rec *codexFSRecorder, name string) int {
	n := 0
	for _, c := range rec.calls {
		if c.Kind == "rename" && filepath.Base(c.Path2) == name {
			n++
		}
	}
	return n
}

// ─── AC-CI-005 — creation and legacy inertness ─────────────────────────────

// TestCodexContractLinkCreation: each I-fixture is initialized once; cells
// that touch an existing file compare the WHOLE file against the expected
// byte sequence; renames are counted per target file; a legacy CLAUDE.md is
// byte-identical to its laid body after the run; a created AGENTS.md carries
// at least one non-space character and no local import.
func TestCodexContractLinkCreation(t *testing.T) {
	for _, fx := range codexLinkFixtures() {
		t.Run("fixture="+fx.name, func(t *testing.T) {
			_, proj := codexNewSandbox(t)
			codexLayLinkFixture(t, proj, fx)
			rec := withCodexFSRecorder(t, nil)

			err := secureCodexInstructionContract(codexContractRequest{ProjectRoot: proj})
			if err != nil {
				t.Fatalf("contract failed: %v", err)
			}

			// per-file rename counts
			if got := codexRenamesOf(rec, codexAgentsRelPath); got != fx.wantAgentsRenames {
				t.Errorf("AGENTS.md renames = %d, want %d", got, fx.wantAgentsRenames)
			}
			if got := codexRenamesOf(rec, "CLAUDE.md"); got != 0 {
				t.Errorf("CLAUDE.md renames = %d, want 0 — a legacy CLAUDE.md is inert", got)
			}
			if got := codexRenamesOf(rec, codexLocalInstructionName); got != 0 {
				t.Errorf("AGENTS.local.md renames = %d, want 0 — the local file is never written", got)
			}

			// exact byte expectations for touched files
			if fx.wantAgentsExact != nil {
				got, rerr := os.ReadFile(filepath.Join(proj, codexAgentsRelPath))
				if rerr != nil {
					t.Fatalf("read AGENTS.md: %v", rerr)
				}
				if string(got) != string(fx.wantAgentsExact) {
					t.Errorf("AGENTS.md bytes diverge from the expected sequence:\n got  = %q\n want = %q", got, fx.wantAgentsExact)
				}
			}

			// legacy inertness: CLAUDE.md byte-identical to its laid body
			if fx.claude != nil {
				got, rerr := os.ReadFile(filepath.Join(proj, "CLAUDE.md"))
				if rerr != nil {
					t.Fatalf("read CLAUDE.md: %v", rerr)
				}
				if string(got) != string(fx.claude) {
					t.Errorf("legacy CLAUDE.md changed across initialization:\n got  = %q\n want = %q", got, fx.claude)
				}
			}

			// the neutral AGENTS.md never imports the local file — the local
			// file reaches Codex through the launcher injection only.
			if got := codexTestExecImports(t, filepath.Join(proj, codexAgentsRelPath), codexTestLocalImportDirective); got != 0 {
				t.Errorf("executing @AGENTS.local.md imports in AGENTS.md = %d, want 0", got)
			}

			// created AGENTS.md: non-empty with at least one non-space char
			if fx.agents == nil {
				got, rerr := os.ReadFile(filepath.Join(proj, codexAgentsRelPath))
				if rerr != nil {
					t.Fatalf("created AGENTS.md unreadable: %v", rerr)
				}
				if len(got) == 0 || len(strings.TrimSpace(string(got))) == 0 {
					t.Errorf("created AGENTS.md is empty or blank (%d bytes)", len(got))
				}
			}
		})
	}
}

// ─── AC-CI-006 — idempotency, three runs ───────────────────────────────────

// codexIdempotencyFixtures = the I-fixtures plus L1 (local present) and
// L2 (local absent).
func codexIdempotencyFixtures() []codexLinkFixture {
	fixtures := codexLinkFixtures()
	fixtures = append(fixtures,
		codexLinkFixture{name: "l1_local_present", local: []byte("local guidance " + codexSentinelLocal + "\n")},
		codexLinkFixture{name: "l2_local_absent"},
	)
	return fixtures
}

// TestCodexContractIdempotent: three consecutive initializations per
// fixture. Every instruction file — the legacy CLAUDE.md included — must be
// byte-identical across run 1→2 AND run 2→3; the I2/I4/I5 family (nothing
// to create) must additionally be byte-identical to its PRE-RUN state after
// run 1 (a rewrite-once-then-stabilize implementation is not idempotent).
func TestCodexContractIdempotent(t *testing.T) {
	for _, fx := range codexIdempotencyFixtures() {
		t.Run("fixture="+fx.name, func(t *testing.T) {
			_, proj := codexNewSandbox(t)
			codexLayLinkFixture(t, proj, fx)

			snap := func() map[string]string {
				out := map[string]string{}
				for _, name := range []string{codexAgentsRelPath, "CLAUDE.md", codexLocalInstructionName} {
					data, err := os.ReadFile(filepath.Join(proj, name))
					if err != nil {
						if os.IsNotExist(err) {
							continue
						}
						t.Fatalf("read %s: %v", name, err)
					}
					out[name] = string(data)
				}
				return out
			}

			before := snap()
			s1 := codexRunContractSnap(t, proj, snap)
			s2 := codexRunContractSnap(t, proj, snap)
			s3 := codexRunContractSnap(t, proj, snap)

			// 1↔2 and 2↔3
			for label, pair := range map[string][2]map[string]string{"1-2": {s1, s2}, "2-3": {s2, s3}} {
				for name, want := range pair[0] {
					if pair[1][name] != want {
						t.Errorf("runs %s: %s changed (len %d → %d)", label, name, len(want), len(pair[1][name]))
					}
				}
				for name := range pair[1] {
					if _, ok := pair[0][name]; !ok {
						t.Errorf("runs %s: %s appeared", label, name)
					}
				}
			}

			// nothing-to-create family: run 1 must not touch ANY file
			if fx.agents != nil {
				for name, want := range before {
					if s1[name] != want {
						t.Errorf("%s: run 1 rewrote %s (len %d → %d) — already-present files must be untouched", fx.name, name, len(want), len(s1[name]))
					}
				}
				for name := range s1 {
					if _, ok := before[name]; !ok {
						t.Errorf("%s: run 1 created %s", fx.name, name)
					}
				}
			}

			// the neutral AGENTS.md never imports the local file
			for i, s := range []map[string]string{s1, s2, s3} {
				if got := codexTestCountImportsInString(t, s[codexAgentsRelPath], codexTestLocalImportDirective); got != 0 {
					t.Errorf("snapshot %d: executing @AGENTS.local.md imports in AGENTS.md = %d, want 0", i+1, got)
				}
			}

			// the local file itself is never rewritten
			if fx.local != nil {
				if s3[codexLocalInstructionName] != string(fx.local) {
					t.Errorf("AGENTS.local.md was rewritten across the runs")
				}
			}
		})
	}
}

func codexRunContractSnap(t *testing.T, proj string, snap func() map[string]string) map[string]string {
	t.Helper()
	if err := secureCodexInstructionContract(codexContractRequest{ProjectRoot: proj}); err != nil {
		t.Fatalf("contract run failed: %v", err)
	}
	return snap()
}

// ─── AC-CI-007 — local-file separation ─────────────────────────────────────

// TestCodexLocalSeparation: from AGENTS.md, walk the transitive closure of
// EXECUTING imports. The local file is reachable through none — the neutral
// contract must not pull a local file into Codex's discovered chain; the
// launcher injects it as developer_instructions instead. The launcher test
// separately proves that the same bytes reach Codex.
func TestCodexLocalSeparation(t *testing.T) {
	type reachCase struct {
		name  string
		local []byte
	}
	cases := []reachCase{
		{name: "l1_local_present", local: []byte("local guidance " + codexSentinelLocal + "\n")},
		{name: "l2_local_absent"},
	}
	for _, rc := range cases {
		t.Run(fmt.Sprintf("fixture=%s/entry=%s", rc.name, codexAgentsRelPath), func(t *testing.T) {
			_, proj := codexNewSandbox(t)
			localBytes := rc.local
			if localBytes != nil {
				if err := os.WriteFile(filepath.Join(proj, codexLocalInstructionName), localBytes, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := secureCodexInstructionContract(codexContractRequest{ProjectRoot: proj}); err != nil {
				t.Fatalf("contract failed: %v", err)
			}

			contents, directiveTargets := codexTestWalkClosure(t, proj, codexAgentsRelPath)

			for _, content := range contents {
				if strings.Contains(content, codexSentinelLocal) {
					t.Errorf("entry %s: local sentinel reachable — the neutral contract must not import the local file", codexAgentsRelPath)
				}
			}

			// no directive points at the local file across the WHOLE closure
			localAbs := filepath.Join(proj, codexLocalInstructionName)
			if got := directiveTargets[localAbs]; got != 0 {
				t.Errorf("entry %s: directives pointing at %s = %d, want 0", codexAgentsRelPath, codexLocalInstructionName, got)
			}

			// the local file is byte-untouched
			if localBytes != nil {
				got, rerr := os.ReadFile(localAbs)
				if rerr != nil {
					t.Fatalf("read AGENTS.local.md: %v", rerr)
				}
				if string(got) != string(localBytes) {
					t.Errorf("AGENTS.local.md changed across initialization")
				}
			}
		})
	}
}

// codexTestWalkClosure walks the executing-import closure of entry within
// proj. It returns the collected file contents (rel name → bytes) and the
// count of directive lines per RESOLVED ABSOLUTE target path.
func codexTestWalkClosure(t *testing.T, proj, entry string) (map[string]string, map[string]int) {
	t.Helper()
	contents := map[string]string{}
	directiveTargets := map[string]int{}
	visited := map[string]bool{}

	var walk func(rel string)
	walk = func(rel string) {
		if visited[rel] {
			return
		}
		visited[rel] = true
		abs := filepath.Join(proj, rel)
		data, err := os.ReadFile(abs)
		if err != nil {
			return // a dangling directive contributes no content
		}
		contents[rel] = string(data)
		for _, name := range codexTestExecDirectives(t, data) {
			targetAbs := filepath.Join(filepath.Dir(abs), name)
			directiveTargets[targetAbs]++
			next, rerr := filepath.Rel(proj, targetAbs)
			if rerr != nil || strings.HasPrefix(next, "..") {
				continue
			}
			walk(filepath.ToSlash(next))
		}
	}
	walk(entry)
	return contents, directiveTargets
}

// codexTestExecDirectives lists the directive FILENAMES of every executing
// import line (independent implementation, same definition-5 semantics).
func codexTestExecDirectives(t *testing.T, content []byte) []string {
	t.Helper()
	var out []string
	fence, comment := false, false
	for _, ln := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(ln, "```") || strings.HasPrefix(ln, "~~~") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		body := strings.TrimRight(ln, "\r")
		if comment {
			if i := strings.Index(body, "-->"); i >= 0 {
				comment = false
				body = body[i+3:]
			} else {
				continue
			}
		}
		if i := strings.Index(body, "<!--"); i >= 0 {
			if j := strings.Index(body[i:], "-->"); j >= 0 {
				body = body[i+j+3:]
			} else {
				comment = true
				continue
			}
		}
		if strings.HasPrefix(body, ">") {
			continue
		}
		if !strings.HasPrefix(body, "@") {
			continue
		}
		token := strings.TrimRight(body, " \t")
		if token == body && len(token) > 1 && !strings.ContainsAny(token, " \t") {
			out = append(out, token[1:])
		}
	}
	return out
}

// codexTestCountImportsInString counts executing imports in raw content.
func codexTestCountImportsInString(t *testing.T, content, directive string) int {
	t.Helper()
	var n int
	fence, comment := false, false
	for _, ln := range strings.Split(content, "\n") {
		if strings.HasPrefix(ln, "```") || strings.HasPrefix(ln, "~~~") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		body := strings.TrimRight(ln, "\r")
		if comment {
			if i := strings.Index(body, "-->"); i >= 0 {
				comment = false
				body = body[i+3:]
			} else {
				continue
			}
		}
		if i := strings.Index(body, "<!--"); i >= 0 {
			if j := strings.Index(body[i:], "-->"); j >= 0 {
				body = body[i+j+3:]
			} else {
				comment = true
				continue
			}
		}
		if strings.HasPrefix(body, ">") {
			continue
		}
		if strings.HasPrefix(body, "@") && strings.TrimRight(body, " \t") == directive {
			n++
		}
	}
	return n
}
