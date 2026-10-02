package web

// SPEC-WEB-SETTINGS-SAVE-001 M1 — observed-RED reproduction for scope ①
// (AC-WSS-001, REQ-WSS-101). The operator could not save the git-worktree or
// audit tab from the moai web console: toggling workflow.worktree.auto_create
// / auto_merge (and editing workflow.audit.* fields) and pressing Save
// produced no visible reaction and nothing on disk.
//
// The reproduction is browser-faithful by construction: it renders the real
// GET /settings page, extracts from the rendered HTML exactly what a browser
// would submit for #settings-form (all panels stay in the DOM, so the POST
// carries every cross-tab field — AC-WSS-001's "전체 폼 POST" shape), applies
// the operator's edits on top of that submission, and POSTs to /save. A
// hand-built field subset would not carry the boundary where this defect is
// expected to live (full-form × schema parser), and the minimal-form overlay
// probe of plan-audit iter 1 reported a clean save for exactly that smaller
// shape.

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// workflowReproFixture mirrors the distributed template's workflow.yaml
// shape for the two failing tabs: comments, an unmodeled sibling key, the
// worktree block the operator toggled, and an audit block with an editable
// claude effort pin. The unmodeled key and comments double as the AC-WSS-002
// preserve assertions once the fix lands.
const workflowReproFixture = `workflow:
    # top-level comment (user-maintained)
    default_mode: ""
    # unmodeled sibling key — must survive a save untouched
    custom_note: keep-me
    worktree:
        # worktree automation comment
        auto_create: false
        auto_merge: false
        auto_cleanup: false
        tmux_preferred: true
    audit:
        model: multi
        gates:
            claude: required
            codex: required
            glm: advisory
        claude:
            effort: medium
`

// extractBrowserSubmission builds the form a real browser submits for
// #settings-form from the rendered page HTML. It reproduces the submission
// semantics of native form controls, scoped to the form subtree (plus any
// out-of-form control carrying form="settings-form"):
//
//   - hidden inputs submit always;
//   - radio / checkbox inputs submit only when checked (an unchecked group
//     submits nothing — the empty=preserve contract);
//   - text-like inputs submit their value (empty string included);
//   - selects submit the first selected option, or the first option when
//     none carries selected (native auto-select-first behavior);
//   - textareas submit their content.
func extractBrowserSubmission(t *testing.T, page string) url.Values {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatalf("parse rendered /settings: %v", err)
	}
	form := url.Values{}

	var walk func(n *html.Node, insideForm bool)
	walk = func(n *html.Node, insideForm bool) {
		if n.Type == html.ElementNode {
			if n.Data == "form" && attrValue(n, "id") == settingsFormID {
				insideForm = true
			}
			if owner, hasOwner := attrLookup(n, "form"); hasOwner {
				insideForm = owner == settingsFormID
			}
			if insideForm {
				switch n.Data {
				case "input":
					submitInputNode(form, n)
				case "select":
					submitSelectNode(t, form, n)
				case "textarea":
					if name := attrValue(n, "name"); name != "" {
						form.Add(name, n.FirstChild.Data)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, insideForm)
		}
	}
	walk(doc, false)
	return form
}

// submitInputNode records one input element's contribution to the submission.
func submitInputNode(form url.Values, n *html.Node) {
	name := attrValue(n, "name")
	if name == "" {
		return
	}
	value := attrValue(n, "value")
	typ := strings.ToLower(attrValue(n, "type"))
	switch typ {
	case "hidden":
		form.Add(name, value)
	case "radio", "checkbox":
		if _, checked := attrLookup(n, "checked"); checked {
			form.Add(name, value)
		}
	case "submit", "button", "image", "reset":
		// only the activating button's name ever submits
	default:
		form.Add(name, value)
	}
}

// submitSelectNode records one select element's contribution: the first
// selected option, or the first option when none is marked selected.
func submitSelectNode(t *testing.T, form url.Values, n *html.Node) {
	t.Helper()
	name := attrValue(n, "name")
	if name == "" {
		return
	}
	firstOpt, firstOK := "", false
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "option" {
			continue
		}
		val := attrValue(c, "value")
		if !firstOK {
			firstOpt, firstOK = val, true
		}
		if _, selected := attrLookup(c, "selected"); selected {
			form.Add(name, val)
			return
		}
	}
	if firstOK {
		form.Add(name, firstOpt)
	}
}

// workflowOperatorFixture mirrors the operator's actual project file shape
// (primary checkout, mtime 2026-09-30 17:46, the state the 10-01 save attempt
// failed to change): empty claude pin scalars (YAML null), a persisted codex
// pin, and user-maintained unmodeled keys. The empty-pin upsert path (absent
// value → first pin) is the exact edit family the audit-tab failure report
// named.
const workflowOperatorFixture = `workflow:
    default_mode: ""
    worktree:
        auto_create: true
        auto_merge: true
        auto_cleanup: true
        tmux_preferred: true
        # session_name_pattern: declared but not read — no code builds a session
        # name from this value (reserved). Retained as a placeholder only.
        session_name_pattern: "moai-{ProjectName}-{SPEC-ID}"
    audit:
        claude:
            model:
            effort:
        codex:
            model: gpt-6.1-sol
            effort: high
`

// TestFullFormSavePersistsAuditPinUpsertOnOperatorFileShape pins the
// operator's exact disk state: raising the empty claude effort pin to a value
// through a full-form POST must upsert the previously-null scalar. The
// operator's 2026-10-01 report is this edit family on this file shape.
func TestFullFormSavePersistsAuditPinUpsertOnOperatorFileShape(t *testing.T) {
	root := t.TempDir()
	seedSectionFile(t, root, "workflow", workflowOperatorFixture)
	a := newAppWithRoot(t, root)

	page := renderSettingsGET(a)
	form := extractBrowserSubmission(t, page)
	form.Set("workflow.audit.claude.effort", "high")

	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("full-form save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	got := readSeededSectionFile(t, root, "workflow")
	if !strings.Contains(got, "effort: high") {
		t.Errorf("claude effort pin upsert not persisted on the operator file shape; file:\n%s", got)
	}
	// The persisted codex pin survives unedited.
	if !strings.Contains(got, "model: gpt-6.1-sol") {
		t.Errorf("persisted codex pin lost after save; file:\n%s", got)
	}
	// User-maintained unmodeled keys survive.
	if !strings.Contains(got, "session_name_pattern") {
		t.Errorf("unmodeled session_name_pattern key lost after save; file:\n%s", got)
	}
}

// TestFullFormSavePersistsWorktreeAndAuditEdits is the AC-WSS-001
// reproduction: the operator's effective payload — the full cross-tab form
// POST — carrying the git-worktree toggle edits and an audit pin edit must
// persist all three to disk.
//
// Observed-RED expectation on the pre-fix tree: the toggles' new values do
// NOT reach workflow.yaml (the operator's "nothing happened" symptom), which
// this test asserts as a failure before the repair (M2) turns it GREEN.
func TestFullFormSavePersistsWorktreeAndAuditEdits(t *testing.T) {
	root := t.TempDir()
	seedSectionFile(t, root, "workflow", workflowReproFixture)
	a := newAppWithRoot(t, root)

	page := renderSettingsGET(a)
	form := extractBrowserSubmission(t, page)

	// The operator's edits (2026-10-01 screenshot): both worktree toggles
	// switched on, and the audit claude effort pin raised to high.
	form.Set("workflow.worktree.auto_create", "1")
	form.Set("workflow.worktree.auto_merge", "1")
	form.Set("workflow.audit.claude.effort", "high")

	rec := servePost(t, a.routes(), "/save", form)
	if rec.Code != http.StatusOK {
		t.Fatalf("full-form save status = %d, want 200; body:\n%s", rec.Code, rec.Body.String())
	}

	// A silent failure must be distinguishable from a reported one: the
	// re-rendered page must not carry an error banner when the save succeeds.
	body := rec.Body.String()
	for _, marker := range []string{"Validation failed", "could not", "failed"} {
		if strings.Contains(body, marker) {
			t.Errorf("save response carries failure marker %q; body:\n%s", marker, body)
		}
	}

	got := readSeededSectionFile(t, root, "workflow")
	if !strings.Contains(got, "auto_create: true") {
		t.Errorf("workflow.worktree.auto_create edit not persisted to disk; file:\n%s", got)
	}
	if !strings.Contains(got, "auto_merge: true") {
		t.Errorf("workflow.worktree.auto_merge edit not persisted to disk; file:\n%s", got)
	}
	if !strings.Contains(got, "effort: high") {
		t.Errorf("workflow.audit.claude.effort edit not persisted to disk; file:\n%s", got)
	}
	// Unedited neighbours stay untouched (no collateral rewrite).
	if !strings.Contains(got, "auto_cleanup: false") || !strings.Contains(got, "tmux_preferred: true") {
		t.Errorf("unedited worktree neighbours changed after save; file:\n%s", got)
	}
	// Lossless contract: comments and the unmodeled sibling key survive.
	if !strings.Contains(got, "# worktree automation comment") {
		t.Errorf("worktree comment lost after save; file:\n%s", got)
	}
	if !strings.Contains(got, "custom_note: keep-me") {
		t.Errorf("unmodeled sibling key lost after save; file:\n%s", got)
	}
}
