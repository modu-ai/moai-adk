package cli

// codex_contract_link_test.go — SPEC-CODEX-INIT-001 M5+M6 cells:
// AC-CI-005 (link creation, 10), AC-CI-006 (three-run idempotency, 12),
// AC-CI-007 (local-file reachability, 4).
//
// Disciplines: fixture-specific EXPECTED BYTE SEQUENCES compared for FULL
// equality on every cell that touches an existing file (partial-substring
// checks pass an implementation that appends a provenance block it was
// never asked for); renames counted PER FILE (a cell writing both files
// must rename twice); idempotency proven by 1↔2 AND 2↔3 byte comparisons
// (a run that rewrites once and then stabilizes is not idempotent); the
// local file's content reached by a closure WALK over executing imports,
// never by a filename grep.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const codexSentinelLocal = "SENTINEL-LOCAL-7q7"

// ─── fixtures — instruction files (acceptance common fixture table) ───────

type codexLinkFixture struct {
	name   string
	agents []byte // nil = absent
	claude []byte // nil = absent
	local  []byte // nil = absent (all I-fixtures; L1 sets it)
	// expectations
	wantAgentsRenames int
	wantClaudeRenames int
	// full expected byte sequences for cells that TOUCH an existing file
	// (nil = the file is created, judged by the lighter creation assertions)
	wantAgentsExact []byte
	wantClaudeExact []byte
}

var userAgentsBody = []byte("# user agents notes\n\ninstruction prose for agents\n")
var userClaudeBody = []byte("# user claude notes\n\ninstruction prose for claude\n")

func codexLinkFixtures() []codexLinkFixture {
	return []codexLinkFixture{
		{
			name:              "i1_both_absent",
			wantAgentsRenames: 1, wantClaudeRenames: 1,
		},
		{
			name:              "i2_agents_only",
			agents:            userAgentsBody,
			wantAgentsRenames: 0, wantClaudeRenames: 1,
			wantAgentsExact: userAgentsBody,
		},
		{
			name:              "i3_claude_only",
			claude:            userClaudeBody,
			wantAgentsRenames: 1, wantClaudeRenames: 1,
			wantClaudeExact: append(append([]byte(nil), userClaudeBody...), []byte(codexLinkAgentsDirective+"\n")...),
		},
		{
			name:              "i4_both_no_link",
			agents:            userAgentsBody,
			claude:            userClaudeBody,
			wantAgentsRenames: 0, wantClaudeRenames: 1,
			wantAgentsExact: userAgentsBody,
			wantClaudeExact: append(append([]byte(nil), userClaudeBody...), []byte(codexLinkAgentsDirective+"\n")...),
		},
		{
			name:              "i5a_line_front",
			agents:            userAgentsBody,
			claude:            []byte("@AGENTS.md\n\ntitle body follows\n"),
			wantAgentsRenames: 0, wantClaudeRenames: 0,
			wantAgentsExact: userAgentsBody,
			wantClaudeExact: []byte("@AGENTS.md\n\ntitle body follows\n"),
		},
		{
			name:              "i5b_line_middle",
			agents:            userAgentsBody,
			claude:            []byte("# title\n\n@AGENTS.md\n\nbody\n"),
			wantAgentsRenames: 0, wantClaudeRenames: 0,
			wantAgentsExact: userAgentsBody,
			wantClaudeExact: []byte("# title\n\n@AGENTS.md\n\nbody\n"),
		},
		{
			name:              "i5c_line_end",
			agents:            userAgentsBody,
			claude:            []byte("# title\n\nbody\n@AGENTS.md\n"),
			wantAgentsRenames: 0, wantClaudeRenames: 0,
			wantAgentsExact: userAgentsBody,
			wantClaudeExact: []byte("# title\n\nbody\n@AGENTS.md\n"),
		},
		{
			name:              "i5d_crlf",
			agents:            userAgentsBody,
			claude:            []byte("# title\r\n\r\n@AGENTS.md\r\n\r\nbody\r\n"),
			wantAgentsRenames: 0, wantClaudeRenames: 0,
			wantAgentsExact: userAgentsBody,
			wantClaudeExact: []byte("# title\r\n\r\n@AGENTS.md\r\n\r\nbody\r\n"),
		},
		{
			name:              "i6_fenced_only",
			claude:            []byte("# title\n\n```\n@AGENTS.md\n```\n\nbody\n"),
			wantAgentsRenames: 1, wantClaudeRenames: 1,
			wantClaudeExact: append(append([]byte(nil), []byte("# title\n\n```\n@AGENTS.md\n```\n\nbody\n")...), []byte(codexLinkAgentsDirective+"\n")...),
		},
		{
			name:              "i7_quoted_only",
			claude:            []byte("# title\n\n> @AGENTS.md\n\nbody\n"),
			wantAgentsRenames: 1, wantClaudeRenames: 1,
			wantClaudeExact: append(append([]byte(nil), []byte("# title\n\n> @AGENTS.md\n\nbody\n")...), []byte(codexLinkAgentsDirective+"\n")...),
		},
	}
}

// codexLayLinkFixture writes the fixture files into proj and returns the
// CLAUDE.local.md bytes (nil when the fixture has none).
func codexLayLinkFixture(t *testing.T, proj string, fx codexLinkFixture) []byte {
	t.Helper()
	for name, content := range map[string][]byte{
		codexAgentsRelPath: fx.agents, codexClaudeRelPath: fx.claude, codexLocalInstructionName: fx.local,
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

// ─── AC-CI-005 — link creation (10 cells) ──────────────────────────────────

// TestCodexContractLinkCreation: each I-fixture is initialized once; cells
// that touch an existing file compare the WHOLE file against the expected
// byte sequence (original + exactly one appended link line, nothing else);
// renames are counted per target file; created files carry exactly one
// executing import.
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
			if got := codexRenamesOf(rec, codexClaudeRelPath); got != fx.wantClaudeRenames {
				t.Errorf("CLAUDE.md renames = %d, want %d", got, fx.wantClaudeRenames)
			}
			if got := codexRenamesOf(rec, codexLocalInstructionName); got != 0 {
				t.Errorf("CLAUDE.local.md renames = %d, want 0 — the local file is never written", got)
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
			if fx.wantClaudeExact != nil {
				got, rerr := os.ReadFile(filepath.Join(proj, codexClaudeRelPath))
				if rerr != nil {
					t.Fatalf("read CLAUDE.md: %v", rerr)
				}
				if string(got) != string(fx.wantClaudeExact) {
					t.Errorf("CLAUDE.md bytes diverge from the expected sequence:\n got  = %q\n want = %q", got, fx.wantClaudeExact)
				}
			}

			// every fixture: exactly one executing import of AGENTS.md
			if got := codexTestExecImports(t, filepath.Join(proj, codexClaudeRelPath), codexLinkAgentsDirective); got != 1 {
				t.Errorf("executing @AGENTS.md imports = %d, want 1", got)
			}
			// no local file in the I-fixtures → nothing may reference it
			if got := codexTestExecImports(t, filepath.Join(proj, codexAgentsRelPath), codexLinkLocalDirective); got != 0 {
				t.Errorf("executing @CLAUDE.local.md imports = %d, want 0 (no local file exists)", got)
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

			// I6/I7: the inactive line survives byte-for-byte AND stays
			// inactive — raw occurrences 2, executing 1.
			if fx.name == "i6_fenced_only" || fx.name == "i7_quoted_only" {
				claudePath := filepath.Join(proj, codexClaudeRelPath)
				if got := codexTestRawOccurrences(t, claudePath, codexLinkAgentsDirective); got != 2 {
					t.Errorf("raw @AGENTS.md occurrences = %d, want 2 (the fenced/quoted one preserved + the appended link)", got)
				}
				if got := codexTestExecImports(t, claudePath, codexLinkAgentsDirective); got != 1 {
					t.Errorf("executing @AGENTS.md imports = %d, want 1", got)
				}
			}
		})
	}
}

// ─── AC-CI-006 — idempotency, three runs (12 cells) ────────────────────────

// codexIdempotencyFixtures = the 10 I-fixtures plus L1 (local present) and
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
// fixture. Every instruction file must be byte-identical across run 1→2 AND
// run 2→3; the I5 family must additionally be byte-identical to its PRE-RUN
// state after run 1 (a rewrite-once-then-stabilize implementation is not
// idempotent); executing-import counts stay 1 at all three points — for a
// CRLF file too.
func TestCodexContractIdempotent(t *testing.T) {
	for _, fx := range codexIdempotencyFixtures() {
		t.Run("fixture="+fx.name, func(t *testing.T) {
			_, proj := codexNewSandbox(t)
			codexLayLinkFixture(t, proj, fx)

			snap := func() map[string]string {
				out := map[string]string{}
				for _, name := range []string{codexAgentsRelPath, codexClaudeRelPath, codexLocalInstructionName} {
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

			// I5 family: run 1 must not touch ANY file
			if strings.HasPrefix(fx.name, "i5") {
				for name, want := range before {
					if s1[name] != want {
						t.Errorf("%s: run 1 rewrote %s (len %d → %d) — already-linked files must be untouched", fx.name, name, len(want), len(s1[name]))
					}
				}
				for name := range s1 {
					if _, ok := before[name]; !ok {
						t.Errorf("%s: run 1 created %s", fx.name, name)
					}
				}
			}

			// executing-import counts stay 1 at all three points
			for i, s := range []map[string]string{s1, s2, s3} {
				claude := s[codexClaudeRelPath]
				if got := codexTestCountImportsInString(t, claude, codexLinkAgentsDirective); got != 1 {
					t.Errorf("snapshot %d: executing @AGENTS.md imports = %d, want 1", i+1, got)
				}
				if s[codexLocalInstructionName] != "" {
					agents := s[codexAgentsRelPath]
					if got := codexTestCountImportsInString(t, agents, codexLinkLocalDirective); got != 1 {
						t.Errorf("snapshot %d: executing @CLAUDE.local.md imports = %d, want 1", i+1, got)
					}
				}
			}

			// the local file itself is never rewritten
			if fx.local != nil {
				if s3[codexLocalInstructionName] != string(fx.local) {
					t.Errorf("CLAUDE.local.md was rewritten across the runs")
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

// ─── AC-CI-007 — local-file reachability (4 cells) ─────────────────────────

// TestCodexLocalReachability: from EACH entry file, walk the transitive
// closure of EXECUTING imports and collect contents. The sentinel must be
// reachable from both entries; the only file CONTAINING it must be
// CLAUDE.local.md itself (a copy-into-AGENTS.md implementation loads it
// twice); exactly ONE directive in the whole closure may point at it
// (counted on the directive-resolved absolute path); with no local file,
// zero directives may point at it.
func TestCodexLocalReachability(t *testing.T) {
	type reachCase struct {
		name        string
		local       []byte
		wantDirects int
	}
	cases := []reachCase{
		{name: "l1_local_present", local: []byte("local guidance " + codexSentinelLocal + "\n"), wantDirects: 1},
		{name: "l2_local_absent", wantDirects: 0},
	}
	for _, rc := range cases {
		for _, entry := range []string{codexAgentsRelPath, codexClaudeRelPath} {
			t.Run(fmt.Sprintf("fixture=%s/entry=%s", rc.name, entry), func(t *testing.T) {
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

				contents, directiveTargets := codexTestWalkClosure(t, proj, entry)

				// sentinel reachable from this entry
				reachable := false
				containFiles := map[string]bool{}
				for rel, content := range contents {
					if strings.Contains(content, codexSentinelLocal) {
						reachable = true
						containFiles[rel] = true
					}
				}
				if localBytes != nil && !reachable {
					t.Errorf("entry %s: sentinel not reachable through executing imports", entry)
				}
				if localBytes != nil && len(containFiles) != 1 || (localBytes != nil && !containFiles[codexLocalInstructionName]) {
					t.Errorf("entry %s: sentinel-carrying files = %v, want exactly {CLAUDE.local.md}", entry, containFiles)
				}

				// directive count toward the local file across the WHOLE closure
				localAbs := filepath.Join(proj, codexLocalInstructionName)
				if got := directiveTargets[localAbs]; got != rc.wantDirects {
					t.Errorf("entry %s: directives pointing at %s = %d, want %d", entry, codexLocalInstructionName, got, rc.wantDirects)
				}

				// the local file is byte-untouched
				if localBytes != nil {
					got, rerr := os.ReadFile(localAbs)
					if rerr != nil {
						t.Fatalf("read CLAUDE.local.md: %v", rerr)
					}
					if string(got) != string(localBytes) {
						t.Errorf("CLAUDE.local.md changed across initialization")
					}
				}
			})
		}
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
