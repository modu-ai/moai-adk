package fix

// failopen_test.go — M3.6 fail-open coverage (SPEC-NAVIGATOR-SYNC-005,
// REQ-NS5-009 / AC-NS5-009). A single table-driven test asserts ALL 8 fail-open
// modes exit 0 (Run never errors / never panics) with the expected empty-or-partial
// output AND one diagnostic log line each.
//
// The 8 rows are the verbatim AC-NS5-009 table from acceptance.md §D:
// 009a work-items absent · 009b detect absent/empty · 009c nav-graph absent ·
// 009d baseline unresolvable · 009e unparseable JSON · 009f schema-invalid ·
// 009g empty diff-scope · 009h no-LLM-runtime.
//
// Test isolation: every temp tree lives under t.TempDir() (auto-cleaned).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// failOpenCase is one row of the AC-NS5-009 table.
type failOpenCase struct {
	name string // subtest name
	// setup writes the fixture inputs for this mode into root. The baseline
	// resolution + scope computation proceeds from whatever setup leaves on disk.
	setup func(t *testing.T, root string)
	// compareTo is the --compare-to flag value (empty → unset).
	compareTo string
	// wantWritten: whether request.json is expected to be written.
	// 009a/b/c/e/f/g/h write; 009d does NOT.
	wantWritten bool
	// wantStatus is the expected Result.Status.
	wantStatus string
	// logToken is a substring that MUST appear in the navigator-sync.log after
	// the run (the "one log line each" the AC table requires).
	logToken string
}

// TestFixFailOpen_All8Modes exercises AC-NS5-009: every fail-open mode exits 0
// (Run returns a Result, never an error/panic) with the expected output shape
// and one diagnostic log line. Run cannot return an error by contract, so
// "exit 0" is asserted as "Run returned without panic" + the expected Result.
func TestFixFailOpen_All8Modes(t *testing.T) {
	t.Parallel()

	// graphBindingUnused binds a path NO input touches → used by 009g to force
	// an empty (but graph-present) diff-scope.
	const graphBindingUnused = `{
  "provenance": {"extract_commit_sha": "b1", "captured_at": "2026-01-01T00:00:00+00:00"},
  "nodes": [{"entity_type": "symbol", "identifier": "pkg.Unused", "display_name": "Unused"}],
  "edges": [{"edge_type": "sym-edge", "source_node": "symbol:pkg.Unused", "target_node": "symbol:pkg.Unused", "source_path": "src/unused.go", "line_number": 1}]
}`

	cases := []failOpenCase{
		{
			name: "009a_work_items_absent",
			setup: func(t *testing.T, root string) {
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
				fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)
				// No work-items.json.
			},
			wantWritten: true,
			wantStatus:  "ready",
			logToken:    "work-items",
		},
		{
			name: "009b_detect_absent",
			setup: func(t *testing.T, root string) {
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), workItemsFixture)
				// No detect state dir.
			},
			wantWritten: true,
			wantStatus:  "ready",
			logToken:    "detect",
		},
		{
			name: "009c_nav_graph_absent",
			setup: func(t *testing.T, root string) {
				// Real git repo so HEAD~1 resolves (baseline degrades to HEAD~1).
				_ = initGitRepo(t, root)
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), workItemsFixture)
				fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)
				gitAddCommit(t, root, "first")
				fixWrite(t, filepath.Join(root, "src", "x", "a.go"), "package x\n")
				gitAddCommit(t, root, "second")
				// No nav-graph.json.
			},
			wantWritten: true,
			wantStatus:  "consistent",
			logToken:    "nav-graph",
		},
		{
			name: "009d_baseline_unresolvable",
			setup: func(t *testing.T, root string) {
				// Non-git dir, no nav-graph, no compareTo → baseline unresolvable.
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), workItemsFixture)
				fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)
			},
			wantWritten: false,
			wantStatus:  "skipped",
			logToken:    "baseline unresolvable",
		},
		{
			name: "009e_unparseable_json",
			setup: func(t *testing.T, root string) {
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
				// Malformed work-items.json — skipped, well-formed detect still seeds.
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), `{not valid json`)
				fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)
			},
			wantWritten: true,
			wantStatus:  "ready",
			logToken:    "unparseable",
		},
		{
			name: "009f_schema_invalid",
			setup: func(t *testing.T, root string) {
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
				// Valid JSON but MISSING the work_items[] array → schema-invalid.
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), `{"provenance":{"route_commit_sha":"x"},"no_work_items":true}`)
				fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)
			},
			wantWritten: true,
			wantStatus:  "ready",
			logToken:    "schema-invalid",
		},
		{
			name: "009g_empty_diff_scope",
			setup: func(t *testing.T, root string) {
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), graphBindingUnused)
				// Inputs touch a DIFFERENT, non-graph-bound path → empty scope.
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"),
					`{"work_items":[{"source_kind":"detect","owner_path":"src/other.go","action":"verify"}]}`)
				fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s.jsonl"),
					`{"changed_path":"src/other.go","changed_at":"2026-01-02T00:00:00Z"}`)
			},
			wantWritten: true,
			wantStatus:  "consistent",
			logToken:    "consistent",
		},
		{
			name: "009h_no_llm_runtime",
			setup: func(t *testing.T, root string) {
				// Full happy-path inputs — the CLI produces request.json (layer 1
				// complete); the AI draft (layer 2) is the orchestrator's job and
				// cannot fire from the Go CLI. 009h is the "layer 1 done, layer 2
				// deferred" degraded mode — true for EVERY ready run by construction.
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"), navGraphFixture)
				fixWrite(t, filepath.Join(root, ".moai", "project", "navigator", "work-items.json"), workItemsFixture)
				fixWrite(t, filepath.Join(root, ".moai", "state", "navigator-detect", "s1.jsonl"), detectJSONLFixture)
			},
			wantWritten: true,
			wantStatus:  "ready",
			logToken:    "draft-request produced",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			tc.setup(t, root)

			// Run never returns an error and never panics (fail-open contract).
			// "exit 0" is asserted as "returned a Result without panicking".
			res := Run(Options{ProjectRoot: root, CompareTo: tc.compareTo})

			if res.Written != tc.wantWritten {
				t.Fatalf("Written = %v, want %v (status=%q msg=%q)", res.Written, tc.wantWritten, res.Status, res.Message)
			}
			if res.Status != tc.wantStatus {
				t.Errorf("Status = %q, want %q", res.Status, tc.wantStatus)
			}

			// 009h: layer 2 cannot fire from the Go CLI → NO draft files exist
			// at fix-drafts/<id>/draft/ after Run (only request.json is written).
			if tc.name == "009h_no_llm_runtime" && res.Written {
				draftDir := filepath.Join(filepath.Dir(res.DraftRequestPath), "draft")
				if _, err := os.Stat(draftDir); !os.IsNotExist(err) {
					t.Errorf("009h: draft/ dir must NOT exist (layer 2 cannot fire from CLI): %v", err)
				}
			}

			// Every row emits exactly one (>=1) diagnostic log line carrying the
			// row-specific token (the AC table's "one log line" requirement).
			logPath := filepath.Join(root, ".moai", "logs", "navigator-sync.log")
			logBytes, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read navigator-sync.log: %v", err)
			}
			logTxt := string(logBytes)
			if !strings.Contains(logTxt, tc.logToken) {
				t.Errorf("log missing token %q;\nlog contents:\n%s", tc.logToken, logTxt)
			}
		})
	}
}
