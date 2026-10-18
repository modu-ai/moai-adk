package cli

// migrate_ownership_case_host_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M4
// (card t1865), AC-OWR-11 host half: what THIS machine's file system must give
// for a path that differs from its manifest key only in letter case.
//
// The test probes the host's case mode inside its own temporary directory — the
// operational probe (probeCaseSensitiveFS) writes into a project root, is not
// called by any operational code, and is not used here — and then asserts the
// outcome that mode requires. Neither mode is skipped. It touches only existing
// identifiers (rows are read through Class/Reason and the report line), so it
// builds and judges the migration as it stands.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	caseRegisteredRel = ".claude/skills/moai-ref-owasp-checklist/SKILL.md"
	caseDiskRel       = ".claude/skills/moai-ref-owasp-checklist/skill.md"
)

// probeHostCaseInsensitive reports whether the file system of dir folds letter
// case, using a probe file inside dir (the test's own temporary directory).
func probeHostCaseInsensitive(t *testing.T, dir string) bool {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "case-probe"), []byte("p"), 0o644); err != nil {
		t.Fatalf("write the case probe: %v", err)
	}
	_, err := os.Stat(filepath.Join(dir, "CASE-PROBE"))
	return err == nil
}

// operationalProbeCalls counts the call expressions of probeCaseSensitiveFS in
// the given sources (slash file name -> source), outside its own declaration.
func operationalProbeCalls(t *testing.T, files map[string]string) []string {
	t.Helper()
	var calls []string
	fset := token.NewFileSet()
	for name, src := range files {
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "probeCaseSensitiveFS" {
					calls = append(calls, name)
				}
			}
			return true
		})
	}
	return calls
}

func TestCaseVariantOnHostFilesystem(t *testing.T) {
	t.Run("the host's own case mode decides the expectation", func(t *testing.T) {
		insensitive := probeHostCaseInsensitive(t, t.TempDir())
		home := installMigrationUserCounterparts(t)
		ownProfileSandbox(t)
		root := t.TempDir()
		data := ownEmbeddedBytes(t, caseRegisteredRel)
		ownAddRegisteredManaged(t, root, caseRegisteredRel, data)
		dir := filepath.Dir(filepath.Join(root, filepath.FromSlash(caseRegisteredRel)))
		if insensitive {
			// One physical file, written under the registered spelling and
			// renamed to the lower-case spelling.
			if err := os.Rename(filepath.Join(dir, "SKILL.md"), filepath.Join(dir, "skill.md")); err != nil {
				t.Fatal(err)
			}
		} else {
			// Two physical files: the registered copy and a separate lower-case
			// copy with the same bytes.
			ownWriteFile(t, root, caseDiskRel, data)
		}
		before := ownTreeSnapshot(t, root, ownTreeOpts{})
		// The removal switch stays off: the files stay, so the tree before and
		// after shows the update did not rename anything.
		r := runOwnershipMigration(t, root, home)
		if diff := ownTreeDiff(before, ownTreeSnapshot(t, root, ownTreeOpts{SkipPrefixes: []string{".moai-backups/"}})); len(diff) != 0 {
			t.Errorf("the update changed the tree: %v", diff)
		}
		rowsFor := func(rel string) []remainingRow {
			var out []remainingRow
			for _, rr := range r.rc.remaining {
				if rr.Rel == rel {
					out = append(out, rr)
				}
			}
			return out
		}
		reportFor := func(rel string) string {
			for _, l := range r.report {
				if strings.Contains(l, rel) {
					return l
				}
			}
			return ""
		}
		entries, _ := os.ReadDir(dir)
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		if insensitive {
			if len(names) != 1 || names[0] != "skill.md" {
				t.Errorf("case-insensitive host: the directory holds %v, want the single file skill.md", names)
			}
			rows := rowsFor(caseDiskRel)
			if len(rows) != 1 {
				t.Fatalf("case-insensitive host: %d rows for the one physical file, want exactly 1 (%v)", len(rows), r.rc.remaining)
			}
			if rows[0].Class != ownershipManaged {
				t.Errorf("case-insensitive host: class %q for the lower-case spelling, want %q (the registered entry's verdict applies to the one file)", rows[0].Class, ownershipManaged)
			}
			if line := reportFor(caseDiskRel); !strings.Contains(line, "case-variant") {
				t.Errorf("case-insensitive host: the spelling difference is not reported: %q", line)
			}
			return
		}
		if len(names) != 2 {
			t.Errorf("case-sensitive host: the directory holds %v, want both spellings as separate files", names)
		}
		if rows := rowsFor(caseDiskRel); len(rows) != 1 || rows[0].Class != ownershipUnregistered {
			t.Errorf("case-sensitive host: rows %+v for %s, want one unregistered row", rows, caseDiskRel)
		}
		if rows := rowsFor(caseRegisteredRel); len(rows) != 1 || rows[0].Class != ownershipManaged {
			t.Errorf("case-sensitive host: rows %+v for %s, want one managed row", rows, caseRegisteredRel)
		}
	})

	t.Run("the operational code never runs the project-root probe", func(t *testing.T) {
		matches, err := filepath.Glob("*.go")
		if err != nil || len(matches) == 0 {
			t.Fatalf("list the package sources: %v (%d files)", err, len(matches))
		}
		files := map[string]string{}
		for _, m := range matches {
			if strings.HasSuffix(m, "_test.go") {
				continue
			}
			data, readErr := os.ReadFile(m)
			if readErr != nil {
				t.Fatal(readErr)
			}
			files[m] = string(data)
		}
		if _, ok := files["update_cleanup.go"]; !ok {
			t.Fatal("update_cleanup.go is not among the scanned sources; the scan would be blind to the probe's home file")
		}
		if calls := operationalProbeCalls(t, files); len(calls) != 0 {
			t.Errorf("probeCaseSensitiveFS (it writes into the project root) is called from operational code: %v", calls)
		}
		mutated := map[string]string{}
		for k, v := range files {
			mutated[k] = v
		}
		mutated["migrate_ownership_case.go"] = "package cli\n\nfunc zzMutant(root string) { _, _ = probeCaseSensitiveFS(root) }\n"
		if got := operationalProbeCalls(t, mutated); len(got) == 0 {
			t.Errorf("the scan stayed green in front of an operational call to the probe")
		}
	})
}
