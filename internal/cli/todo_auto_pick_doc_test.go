// todo_auto_pick_doc_test.go — SPEC-TODO-AUTO-PICK-001 M5 (AC-TAU-003, -007,
// -008, -010): the `--auto` doctrine states that the invoked session chooses
// cards on its own judgment, each only through a lease and never a keep-set
// card; every surface that stated the old serial-only authority carries the
// new sentences and none keeps the old ones; and the live copies agree with
// their template mirrors byte for byte, except the one pre-existing
// live-only sentence of factory-dispatch.md, which is preserved, not absorbed.
//
// Literals are matched after the shared normalization of the sibling doc
// tests (backticks dropped, whitespace collapsed) so a reflow cannot fake a
// pass or a failure. The literals that the acceptance ledger also greps in
// the raw bytes (rawContain) are additionally checked on the raw text, so a
// phrase that wraps across a hard line break, or that carries an interior
// backtick, fails here exactly as it fails the grep.
package cli

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// autoPickSurface is one document the card-pick doctrine lands on. live is a
// repository-relative path; mirror is derived from it unless templateOnly is
// set (a generated artifact that has no live twin).
type autoPickSurface struct {
	name         string
	live         string
	templateOnly bool
	contain      []string // after normalization
	absent       []string // after normalization, case-sensitive
	rawContain   []string // raw bytes
}

func autoPickMirrorPath(live string) string {
	return filepath.Join("internal", "template", "templates", live)
}

func autoPickSurfaces() []autoPickSurface {
	return []autoPickSurface{
		{
			name: "factory-dispatch.md",
			live: filepath.Join(".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
			contain: []string{
				"Outside an --auto authorization the leader never picks for the operator",
				"authorizes the invoked session to take cards from the queue on its own judgment",
				"moai factory next — bare, or --card <id>",
			},
			absent: []string{
				"Promotion is the operator's act, always.",
				"The leader never picks for the operator",
				"consumption of the queue and nothing else",
			},
			rawContain: []string{
				"Outside an --auto authorization the leader never picks for the operator",
				"authorizes the invoked session to take cards from the queue on its own judgment",
				"factory next --card",
			},
		},
		{
			name: "gtd.md",
			live: filepath.Join(".claude", "skills", "moai", "workflows", "gtd.md"),
			contain: []string{
				"on its own judgment",
				"a lane session exercises the --auto authorization through moai factory next",
				"factory next --card",
				"keep-set",
				"payments, secrets, or irreversible external-shared work",
				"decision record:",
				"ladder_path=gate-row card pick",
				"unmeasured",
				"demotes a card in the serial cycle and excludes it on the lease path",
				"evidence, not the decision board",
				"adds an input without amending the keep-set or the lease path",
			},
			absent: []string{
				"consumption of the queue and nothing else",
			},
			rawContain: []string{
				"exercises the --auto authorization through",
				"ladder_path=gate-row card pick",
				"factory next --card",
				"demotes a card in the serial cycle and excludes it on the lease path",
			},
		},
		{
			name: "manager-todo.md",
			live: filepath.Join(".claude", "agents", "moai", "manager-todo.md"),
			contain: []string{
				"on its own judgment",
				"keep-set",
			},
			absent: []string{
				"process cards in queue order",
				"consumption of the queue and nothing else",
			},
			rawContain: []string{
				"factory next --card",
			},
		},
		{
			// Generated from the template manager-todo.md by `make agents-emit`;
			// never hand-edited, so the same literals follow the source.
			name:         "manager-todo.toml",
			live:         filepath.Join(".codex", "agents", "moai", "manager-todo.toml"),
			templateOnly: true,
			contain: []string{
				"on its own judgment",
				"keep-set",
			},
			absent: []string{
				"process cards in queue order",
				"consumption of the queue and nothing else",
			},
		},
		{
			name: "moai-factory-foreman/SKILL.md",
			live: filepath.Join(".claude", "skills", "moai-factory-foreman", "SKILL.md"),
			contain: []string{
				"on its own judgment",
				"outside a batch authorization",
			},
			absent: []string{
				"consumption in queue order is authorized",
				"not yours to pick.",
			},
			rawContain: []string{
				"factory next --card",
			},
		},
		{
			name: "factory-dispatch-detail.md",
			live: filepath.Join(".claude", "rules", "moai", "workflow", "factory-dispatch-detail.md"),
			contain: []string{
				"a pull request or landed state is a skip input for a queued candidate the session chose and is report-only for an operator-picked card",
			},
			rawContain: []string{
				"report-only for an operator-picked card",
			},
		},
		{
			name: "auto-semantics.md",
			live: filepath.Join(".claude", "rules", "moai", "workflow", "auto-semantics.md"),
			contain: []string{
				"ladder_path=gate-row card pick",
				"adds an input without amending the keep-set or the lease path",
				"unmeasured",
				"evidence, not the decision board",
				"no party re-reads the card-pick record",
			},
			absent: []string{
				"authorizes serial queue consumption and nothing else",
			},
			rawContain: []string{
				"ladder_path=gate-row card pick",
				"factory next --card",
			},
		},
	}
}

// autoPickCopies returns the copies of one surface the doctrine must reach.
func autoPickCopies(s autoPickSurface) []autoDocSurface {
	if s.templateOnly {
		return []autoDocSurface{{"template " + s.name, autoPickMirrorPath(s.live)}}
	}
	return []autoDocSurface{
		{"live " + s.name, s.live},
		{"template " + s.name, autoPickMirrorPath(s.live)},
	}
}

// TestAutoPickDocDoctrine (AC-TAU-003, -007, -008): every must-contain literal
// of the contract is on every copy, every old literal is gone from every copy
// — live and mirror, so an edit that adds the new wording and leaves the old
// fails — and the record, the input set and the evidence-not-board framing
// are stated without binding the relation records to one store.
func TestAutoPickDocDoctrine(t *testing.T) {
	root := autoDocRepoRoot(t)

	for _, s := range autoPickSurfaces() {
		for _, c := range autoPickCopies(s) {
			doc := autoDocRead(t, root, c.path)
			norm := autoDocNormalize(doc)
			for _, want := range s.contain {
				t.Run("contains/"+c.name+"/"+want, func(t *testing.T) {
					if !strings.Contains(norm, autoDocNormalize(want)) {
						t.Errorf("%s does not state %q", c.name, want)
					}
				})
			}
			for _, old := range s.absent {
				t.Run("absent/"+c.name+"/"+old, func(t *testing.T) {
					if strings.Contains(norm, autoDocNormalize(old)) {
						t.Errorf("%s still carries the replaced sentence %q", c.name, old)
					}
				})
			}
			for _, want := range s.rawContain {
				t.Run("raw/"+c.name+"/"+want, func(t *testing.T) {
					if !strings.Contains(doc, want) {
						t.Errorf("%s does not carry %q in its raw bytes (a hard wrap or an interior backtick breaks the grep)", c.name, want)
					}
				})
			}
		}
	}

	// auto-semantics.md carries a section numbered 9.3 for the card pick.
	heading := regexp.MustCompile(`(?m)^### 9\.3 `)
	for _, c := range []autoDocSurface{
		{"live auto-semantics.md", filepath.Join(".claude", "rules", "moai", "workflow", "auto-semantics.md")},
		{"template auto-semantics.md", autoPickMirrorPath(filepath.Join(".claude", "rules", "moai", "workflow", "auto-semantics.md"))},
	} {
		doc := autoDocRead(t, root, c.path)
		t.Run("9.3 heading/"+c.name, func(t *testing.T) {
			if !heading.MatchString(doc) {
				t.Errorf("%s has no section numbered 9.3", c.name)
			}
		})
	}

	// The record form names the inputs it carries, in the section 9.3 of the
	// rule and in the gtd.md `--auto` section a lane actually reads, and no
	// sentence there binds the relation records to one store.
	gtd := filepath.Join(".claude", "skills", "moai", "workflows", "gtd.md")
	sem := filepath.Join(".claude", "rules", "moai", "workflow", "auto-semantics.md")
	regions := []struct{ name, path, start, end string }{
		{"live gtd.md --auto section", gtd, "### `--auto` — the serial batch consumption", "## Standing sources"},
		{"template gtd.md --auto section", autoPickMirrorPath(gtd), "### `--auto` — the serial batch consumption", "## Standing sources"},
		{"live auto-semantics.md 9.3", sem, "### 9.3 ", "## 10. "},
		{"template auto-semantics.md 9.3", autoPickMirrorPath(sem), "### 9.3 ", "## 10. "},
	}
	inputs := []string{"relation records", "pull-request", "worktree presence", "file overlap", "unmeasured"}
	for _, r := range regions {
		region := autoDocNormalize(autoDocMirrorPassage(t, r.name, autoDocRead(t, root, r.path), r.start, r.end))
		for _, in := range inputs {
			t.Run("record input/"+r.name+"/"+in, func(t *testing.T) {
				if !strings.Contains(region, in) {
					t.Errorf("%s does not name the record input %q", r.name, in)
				}
			})
		}
		for _, store := range []string{"gtd_relations", "queue-findings"} {
			t.Run("not store-bound/"+r.name+"/"+store, func(t *testing.T) {
				if strings.Contains(region, store) {
					t.Errorf("%s names %q as the relation source; the input set is open", r.name, store)
				}
			})
		}
	}
}

// autoPickMirrorPairs are the seven live files whose template mirror the
// card-pick amendment edits.
func autoPickMirrorPairs() []string {
	return []string{
		filepath.Join(".claude", "rules", "moai", "workflow", "factory-dispatch.md"),
		filepath.Join(".claude", "rules", "moai", "workflow", "factory-dispatch-detail.md"),
		filepath.Join(".claude", "rules", "moai", "workflow", "auto-semantics.md"),
		filepath.Join(".claude", "skills", "moai", "workflows", "gtd.md"),
		filepath.Join(".claude", "agents", "moai", "manager-todo.md"),
		filepath.Join(".claude", "skills", "moai-factory-foreman", "SKILL.md"),
		filepath.Join(".claude", "rules", "moai", "core", "moai-mcp-tools-catalogue.md"),
	}
}

// autoPickLiveOnlyMarker identifies the one pre-existing live-only sentence of
// factory-dispatch.md; the copies differ by exactly the line that carries it.
const autoPickLiveOnlyMarker = "moai worktree sweep"

// TestAutoPickMirrorParity (AC-TAU-010): every edited live file is
// byte-identical to its template mirror, except factory-dispatch.md, which
// differs by exactly the one live-only line it differed by before and by
// nothing else. The drift is preserved, not absorbed in either direction.
func TestAutoPickMirrorParity(t *testing.T) {
	root := autoDocRepoRoot(t)
	driftFile := filepath.Join(".claude", "rules", "moai", "workflow", "factory-dispatch.md")

	for _, live := range autoPickMirrorPairs() {
		liveDoc := autoDocRead(t, root, live)
		mirrorDoc := autoDocRead(t, root, autoPickMirrorPath(live))
		if liveDoc == "" || mirrorDoc == "" {
			t.Fatalf("%s: an empty copy agrees with anything — refusing to compare", live)
		}
		if live == driftFile {
			continue
		}
		t.Run("byte-identical/"+live, func(t *testing.T) {
			if liveDoc != mirrorDoc {
				t.Errorf("%s differs from its template mirror (%d vs %d bytes)", live, len(liveDoc), len(mirrorDoc))
			}
		})
	}

	t.Run("factory-dispatch.md differs by exactly the live-only sentence", func(t *testing.T) {
		liveLines := strings.Split(autoDocRead(t, root, driftFile), "\n")
		mirrorLines := strings.Split(autoDocRead(t, root, autoPickMirrorPath(driftFile)), "\n")
		if len(liveLines) != len(mirrorLines) {
			t.Fatalf("factory-dispatch.md live has %d lines, the mirror %d: the copies differ by more than one line", len(liveLines), len(mirrorLines))
		}

		// The live-only sentence is the tail of one paragraph line: the mirror
		// line is a strict prefix of the live line, and every other line agrees.
		differing := 0
		for i := range liveLines {
			if liveLines[i] == mirrorLines[i] {
				continue
			}
			differing++
			if !strings.Contains(liveLines[i], autoPickLiveOnlyMarker) {
				t.Errorf("line %d differs and the live line does not carry %q", i+1, autoPickLiveOnlyMarker)
			}
			if strings.Contains(mirrorLines[i], autoPickLiveOnlyMarker) {
				t.Errorf("line %d: the mirror absorbed the live-only sentence", i+1)
			}
			if !strings.HasPrefix(liveLines[i], mirrorLines[i]) {
				t.Errorf("line %d differs by more than the live-only tail", i+1)
			}
		}
		if differing != 1 {
			t.Errorf("factory-dispatch.md live and mirror differ on %d lines, want exactly 1", differing)
		}
	})
}
