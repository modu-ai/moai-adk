package cellguard

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// repoRoot resolves the repository root from this package's directory
// (internal/harness/cellguard) and asserts it is the tree this guard means to
// measure. Without that assertion the whole package could run green against the
// wrong directory and report nothing, which is the shape of a vacuous pass.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := filepath.Join(wd, "..", "..", "..")
	for _, marker := range []string{"go.mod", filepath.Join("docs-site", "content")} {
		if _, err := os.Stat(filepath.Join(root, marker)); err != nil {
			t.Fatalf("resolved repo root %s does not carry %s: %v", root, marker, err)
		}
	}
	return root
}

// locales is the four-locale set every registered page exists in. The cell
// values are byte-identical across them (measured), but the line numbers are
// not — ko prints the agent table at line 27 and the other three at line 36 —
// which is why every site is read by content.
var locales = []string{"ko", "en", "ja", "zh"}

// sites is the registry. It lives in code rather than behind a glob so a reader
// can see WHAT is guarded; widening it is a deliberate edit, not an accident of
// a wildcard. The anti-vacuity sweep below stops the opposite failure — a new
// page carrying cells and never being registered here.
func sites() []Site {
	var out []Site
	for _, l := range locales {
		out = append(out,
			Site{
				ID:       "faq/" + l,
				Path:     "docs-site/content/" + l + "/getting-started/faq.md",
				Kind:     KindAgentMatrix,
				WantRows: 12,
				Complete: false,
				SubsetReason: "the page splits the roster across a Manager table and an " +
					"Evaluator/Builder/Advisor/Specialist table, and states the built-in " +
					"Explore in the sentence below them rather than as a row. Twelve rows " +
					"plus one prose statement is the whole roster; asserting complete " +
					"membership on the tables alone would break a correct page.",
			},
			Site{
				ID:       "profile-matrix/agents/" + l,
				Path:     "docs-site/content/" + l + "/advanced/profile-matrix.md",
				Kind:     KindAgentMatrix,
				WantRows: 13,
				Complete: true,
			},
			Site{
				ID:       "profile-matrix/harness/" + l,
				Path:     "docs-site/content/" + l + "/advanced/profile-matrix.md",
				Kind:     KindHarnessClass,
				WantRows: 7,
				Complete: true,
			},
			// This page was not named by the card that produced this package —
			// the sweep below found it, which is the sweep earning its keep.
			// Its cells already agreed with the matrix; it was unguarded, not
			// wrong, and an unguarded correct page is one edit from a wrong one.
			Site{
				ID:       "model-policy/" + l,
				Path:     "docs-site/content/" + l + "/multi-llm/model-policy.md",
				Kind:     KindAgentMatrix,
				WantRows: 13,
				Complete: true,
			},
		)
	}
	return out
}

// population returns the row keys a kind draws from, as a set and as the
// canonical ordered slice used for the membership comparison.
func population(kind TableKind) (map[string]bool, []string) {
	var names []string
	switch kind {
	case KindAgentMatrix:
		names = template.ProfileMatrixAgents()
	case KindHarnessClass:
		names = template.HarnessClasses()
	}
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	return set, names
}

// want resolves the source-of-truth cell for one row under one profile.
func want(t *testing.T, kind TableKind, key, profile string) Cell {
	t.Helper()
	switch kind {
	case KindAgentMatrix:
		me, ok := template.DefaultProfileMatrix()[profile][key]
		if !ok {
			t.Fatalf("profile %q has no matrix cell for agent %q", profile, key)
		}
		return Cell{Model: me.Model, Effort: me.Effort}
	case KindHarnessClass:
		me, known := template.ResolveHarnessAgentModelEffort(config.LLMConfig{Profile: profile}, key)
		if !known {
			t.Fatalf("purpose class %q is not recognized by ResolveHarnessAgentModelEffort", key)
		}
		return Cell{Model: me.Model, Effort: me.Effort}
	}
	t.Fatalf("unknown table kind %q", kind)
	return Cell{}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

// TestProfileCellsMatchTheMatrix is the guard. Each registered site is read,
// its rows extracted by content, and every cell compared against the Go
// structure that decides it.
func TestProfileCellsMatchTheMatrix(t *testing.T) {
	root := repoRoot(t)

	registered := sites()
	if len(registered) == 0 {
		t.Fatal("the site registry is empty; an empty registry guards nothing and would pass vacuously")
	}

	for _, s := range registered {
		t.Run(s.ID, func(t *testing.T) {
			if !s.Complete && strings.TrimSpace(s.SubsetReason) == "" {
				t.Fatalf("site %s declares a subset and states no reason; an unexplained subset is indistinguishable from a stale page", s.ID)
			}

			set, canonical := population(s.Kind)
			rows, err := Extract(read(t, root, s.Path), set)
			if err != nil {
				t.Fatalf("%s: %v", s.Path, err)
			}

			// Anti-vacuity: an extractor that read no table agrees with
			// everything, so the count is asserted before the cells are.
			if len(rows) != s.WantRows {
				t.Fatalf("%s: extracted %d rows, want %d — extracting nothing is not agreement, it is a guard that read no table", s.Path, len(rows), s.WantRows)
			}

			if s.Complete {
				missing, extra := SetDiff(rows, canonical)
				for _, m := range missing {
					t.Errorf("%s: %s omits %q, which the %s population carries", s.Path, s.ID, m, s.Kind)
				}
				for _, e := range extra {
					t.Errorf("%s: %s lists %q, which the %s population does not carry", s.Path, s.ID, e, s.Kind)
				}
			}

			for _, r := range rows {
				for _, p := range Profiles {
					got := r.Cells[p]
					w := want(t, s.Kind, r.Key, p)
					if got != w {
						t.Errorf("%s:%d %s under %q = %q, the matrix says %q", s.Path, r.Line, r.Key, p, got, w)
					}
				}
			}
		})
	}
}

// TestEveryCellTableIsRegistered is the anti-vacuity sweep. A page that grows a
// profile cell table and is never added to the registry would be guarded by
// nothing, and nothing would say so — which is exactly how the two pages this
// package exists for went unguarded. The sweep walks docs-site and fails on any
// file yielding rows of a registered kind from an unregistered path.
func TestEveryCellTableIsRegistered(t *testing.T) {
	root := repoRoot(t)

	known := map[string]bool{}
	for _, s := range sites() {
		known[s.Path] = true
	}

	contentDir := filepath.Join(root, "docs-site", "content")
	var unregistered []string
	var scanned int

	err := filepath.WalkDir(contentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		scanned++
		relOS, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		rel := filepath.ToSlash(relOS)
		if known[rel] {
			return nil
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, kind := range []TableKind{KindAgentMatrix, KindHarnessClass} {
			set, _ := population(kind)
			rows, exErr := Extract(string(b), set)
			if exErr != nil {
				// A malformed cell in an UNREGISTERED file is itself a
				// finding: the page writes something cell-shaped that this
				// guard cannot read, so it is reported rather than skipped.
				unregistered = append(unregistered, rel+" ("+string(kind)+": "+exErr.Error()+")")
				continue
			}
			if len(rows) > 0 {
				unregistered = append(unregistered, rel+" ("+string(kind)+": "+strconv.Itoa(len(rows))+" rows)")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", contentDir, err)
	}

	// The sweep's own premise: it must actually have read files. A walk that
	// scanned nothing reports no unregistered pages for the wrong reason.
	if scanned == 0 {
		t.Fatalf("the sweep scanned 0 markdown files under %s; an empty sweep finds nothing and proves nothing", contentDir)
	}
	t.Logf("sweep scanned %d markdown files, %d registered paths", scanned, len(known))

	for _, u := range unregistered {
		t.Errorf("%s carries a profile cell table and is not in the registry; register it or the cells it prints are guarded by nothing", u)
	}
}
