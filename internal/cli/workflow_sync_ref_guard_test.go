package cli

// workflow_sync_ref_guard_test.go — the workflow-exclusion regression guard
// (SPEC-TRUNK-BACKUP-SYNC-001 REQ-TBS-005, AC-TBS-001/002/010).
//
// The backup verb pushes local main to refs/heads/sync/main. That ref is safe
// only while NO workflow under .github/workflows/ reacts to it: a future
// workflow added with a bare `on: push` would fire the full CI matrix on every
// backup push. This file is the guard that keeps that property observable:
//
//   - TestWorkflowPushExcludesSyncMain sweeps the LIVE workflow set and fails
//     on any push trigger whose branch filter admits sync/main. It also fails
//     when the sweep finds nothing (the empty-sweep guard) — a green over an
//     empty swept set is indistinguishable from a broken enumeration.
//   - TestBranchFilterMatcherMutants runs the in-test mutant fixtures in BOTH
//     directions (admitting and non-admitting), so the guard's failure is
//     observed on known failing inputs, not asserted. The matcher's supported
//     grammar is exactly the REQ-TBS-005 subset; a pattern outside it fails
//     the test loudly rather than reading as non-matching.
//   - TestWorkflowSweepEmptyGuardFails observes the empty-sweep guard's own
//     failure on a directory carrying no push-trigger workflow.
//   - TestGoCodeFilterMapsWorkflows pins the CI scheduling half: a diff
//     touching only .github/workflows/** must classify as Go-suite input, or
//     this guard would not run exactly when its guarded surface changes
//     (AC-TBS-010, verified by YAML parse — never text search).

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// syncGuardRefName is the protected ref the whole guard family evaluates
// against — the backup verb's constant target (REQ-TBS-001).
const syncGuardRefName = "sync/main"

// --- AC-TBS-001: the live sweep ----------------------------------------------

// TestWorkflowPushExcludesSyncMain enumerates every *.yml / *.yaml workflow
// file, reads each on.push trigger, and asserts no push trigger's branch
// filter admits sync/main under GitHub's branch-filter matching. A pattern
// the matcher cannot parse fails the test loudly (fail-loud, never silent
// non-match). Zero push-trigger workflows swept also fails: the guard must
// not report green over an empty set.
func TestWorkflowPushExcludesSyncMain(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	res, err := sweepWorkflowPushAdmission(filepath.Join(root, ".github", "workflows"), syncGuardRefName)
	if err != nil {
		t.Fatalf("WORKFLOW_PUSH_EXCLUSION: sweep failed: %v", err)
	}
	if res.PushTriggers == 0 {
		t.Fatalf("WORKFLOW_PUSH_EXCLUSION: empty sweep — %d workflow files enumerated but none carries an on.push trigger; the guard cannot report green over an empty set", res.WorkflowFiles)
	}
	// The swept count rides the green itself: a pass over N triggers is a
	// statement about N files, and a reader of the log must be able to see N
	// without re-deriving it (a green whose swept set is empty asserts
	// nothing; the count is what makes the green interpretable).
	t.Logf("WORKFLOW_PUSH_EXCLUSION: swept %d workflow files, %d carrying on.push, %d admitting %q", res.WorkflowFiles, res.PushTriggers, len(res.Admitting), syncGuardRefName)
	for _, admitting := range res.Admitting {
		t.Errorf("WORKFLOW_PUSH_EXCLUSION: %s on.push admits %q — a backup push would fire this workflow; exclude the ref from its branch filter", admitting, syncGuardRefName)
	}
}

// TestWorkflowSweepEmptyGuardFails observes the empty-sweep guard's failure
// directly (verification-completeness: a guard whose red has not been seen has
// proven nothing). A directory with no workflow files and a directory whose
// workflows carry no push trigger must both refuse to report a verdict.
func TestWorkflowSweepEmptyGuardFails(t *testing.T) {
	t.Parallel()

	empty := t.TempDir()
	if _, err := sweepWorkflowPushAdmission(empty, syncGuardRefName); err == nil {
		t.Errorf("empty dir: want the sweep to refuse (no workflow files enumerated), got nil error")
	}

	noPush := t.TempDir()
	if err := os.WriteFile(filepath.Join(noPush, "dispatch-only.yml"), []byte("on: workflow_dispatch\njobs:\n  x:\n    runs-on: ubuntu-latest\n    steps: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := sweepWorkflowPushAdmission(noPush, syncGuardRefName); err == nil {
		t.Errorf("no-push-trigger dir: want the sweep to refuse (zero push triggers — a green over an empty swept set is uninterpreted output), got nil error")
	}
}

// --- AC-TBS-002: mutant detection, both directions ----------------------------

// TestBranchFilterMatcherMutants runs the in-test mutant fixtures — the guard's
// observed-failure evidence. Every case is a workflow fragment parsed by the
// SAME parser the live sweep uses, so a matcher that passes the live sweep but
// misses a mutant, or one that flags ['*'] as admitting, fails here.
func TestBranchFilterMatcherMutants(t *testing.T) {
	t.Parallel()

	push := func(body string) string { return "on:\n  push:\n" + body }
	cases := []struct {
		name       string
		yaml       string
		wantPush   bool // whether the workflow carries an on.push trigger at all
		wantAdmits bool
		wantErr    bool
	}{
		// A bare `on: push` block with no branch filter — the canonical
		// future-workflow accident: fires on EVERY branch push.
		{name: "bare scalar push admits", yaml: "on: push\n", wantPush: true, wantAdmits: true},
		{name: "bare mapping push admits", yaml: "on:\n  push:\n", wantPush: true, wantAdmits: true},
		{name: "list form carries push", yaml: "on:\n  - push\n  - pull_request\n", wantPush: true, wantAdmits: true},

		// Non-push triggers are not push triggers.
		{name: "scalar pull_request is no push trigger", yaml: "on: pull_request\n", wantPush: false},
		{name: "list form without push", yaml: "on:\n  - pull_request\n  - workflow_dispatch\n", wantPush: false},

		// A branches list containing sync/**.
		{name: "sync glob admits", yaml: push("    branches:\n      - 'sync/**'\n"), wantPush: true, wantAdmits: true},
		{name: "literal sync/main admits", yaml: push("    branches:\n      - main\n      - sync/main\n"), wantPush: true, wantAdmits: true},

		// The bidirectional set of REQ-TBS-005.
		{name: "double star admits (crosses /)", yaml: push("    branches:\n      - '**'\n"), wantPush: true, wantAdmits: true},
		{name: "single star does not cross /", yaml: push("    branches:\n      - '*'\n"), wantPush: true, wantAdmits: false},
		{name: "last-match negation wins", yaml: push("    branches:\n      - '**'\n      - '!sync/main'\n"), wantPush: true, wantAdmits: false},
		{name: "negation then positive also last-match-wins", yaml: push("    branches:\n      - '!sync/main'\n      - '**'\n"), wantPush: true, wantAdmits: true},
		{name: "character class matches", yaml: push("    branches:\n      - 'sync/m[a-z]in'\n"), wantPush: true, wantAdmits: true},
		{name: "character class range that excludes", yaml: push("    branches:\n      - 'sync/m[x-z]in'\n"), wantPush: true, wantAdmits: false},

		// `**/` in segment position (including a leading `**/` prefix)
		// matches ZERO or more path segments (amended REQ-TBS-005,
		// sync-audit round 1 S2): `**/sync/main` and `sync/**/main` both
		// admit sync/main, while the anchor segment still has to match.
		{name: "leading globstar slash matches zero segments", yaml: push("    branches:\n      - '**/sync/main'\n"), wantPush: true, wantAdmits: true},
		{name: "mid globstar slash matches zero segments", yaml: push("    branches:\n      - 'sync/**/main'\n"), wantPush: true, wantAdmits: true},
		{name: "globstar slash still requires its anchor segment", yaml: push("    branches:\n      - 'x/**/main'\n"), wantPush: true, wantAdmits: false},
		{name: "globstar slash consumes one segment", yaml: push("    branches:\n      - '**/main'\n"), wantPush: true, wantAdmits: true},
		{name: "main does not admit sync", yaml: push("    branches:\n      - main\n"), wantPush: true, wantAdmits: false},
		{name: "ci style filter does not admit", yaml: push("    branches:\n      - main\n      - 'ci/**'\n"), wantPush: true, wantAdmits: false},

		// tags semantics (REQ-TBS-005): tags-only fires on tag pushes only;
		// with branches also defined, the branches list decides branch pushes.
		{name: "tags only does not fire on branch push", yaml: push("    tags:\n      - 'v*'\n"), wantPush: true, wantAdmits: false},
		{name: "tags and branches: branches decide", yaml: push("    branches:\n      - main\n    tags:\n      - 'v*'\n"), wantPush: true, wantAdmits: false},

		// branches-ignore only: fires on every branch not ignored.
		{name: "branches-ignore of main admits sync", yaml: push("    branches-ignore:\n      - main\n"), wantPush: true, wantAdmits: true},
		{name: "branches-ignore matching the ref excludes", yaml: push("    branches-ignore:\n      - 'sync/**'\n"), wantPush: true, wantAdmits: false},
		{name: "branches and matching ignore excludes", yaml: push("    branches:\n      - '**'\n    branches-ignore:\n      - 'sync/**'\n"), wantPush: true, wantAdmits: false},

		// Out-of-subset grammar fails loudly, never silently non-matching.
		{name: "question mark is out of subset", yaml: push("    branches:\n      - 'sync/ma?in'\n"), wantPush: true, wantErr: true},
		{name: "braces are out of subset", yaml: push("    branches:\n      - '{sync}/main'\n"), wantPush: true, wantErr: true},
		{name: "triple star is out of subset", yaml: push("    branches:\n      - '***'\n"), wantPush: true, wantErr: true},
		{name: "unclosed character class", yaml: push("    branches:\n      - 'sync/m[a-zin'\n"), wantPush: true, wantErr: true},
		{name: "empty pattern is unparseable", yaml: push("    branches:\n      - ''\n"), wantPush: true, wantErr: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			trig, err := parseWorkflowPushTrigger([]byte(tc.yaml), "fixture.yml")
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			if (trig != nil) != tc.wantPush {
				t.Fatalf("push trigger present = %v, want %v (fixture: %q)", trig != nil, tc.wantPush, tc.yaml)
			}
			if trig == nil {
				return
			}
			admits, err := pushTriggerAdmitsBranch(trig, syncGuardRefName)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want a loud parse/match error for the out-of-subset pattern, got admits=%v nil error", admits)
				}
				return
			}
			if err != nil {
				t.Fatalf("match: %v", err)
			}
			if admits != tc.wantAdmits {
				t.Errorf("admits %q = %v, want %v", syncGuardRefName, admits, tc.wantAdmits)
			}
		})
	}
}

// --- AC-TBS-010: the guard is SCHEDULED on workflow-only diffs ----------------

// TestGoCodeFilterMapsWorkflows pins .github/test-input-filters.yml's go_code
// filter to map .github/workflows/** — without it a diff touching only
// workflow files classifies as not-Go and skips the Go suite carrying this
// guard, so the guard would stop firing exactly when its guarded surface
// changes. The mapping is verified by PARSING the YAML and asserting the
// sequence contains the glob as a list entry — never by text search, which a
// comment mentioning the glob also satisfies (in-tree parsing precedent:
// mappingSequenceStrings, internal/template/dotgithub_template_test.go).
func TestGoCodeFilterMapsWorkflows(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, ".github", "test-input-filters.yml"))
	if err != nil {
		t.Fatalf("read .github/test-input-filters.yml: %v", err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse .github/test-input-filters.yml: %v", err)
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		t.Fatalf("expected a top-level mapping in test-input-filters.yml")
	}
	entries := syncGuardSequenceStrings(t, doc.Content[0], "go_code")
	const want = ".github/workflows/**"
	for _, e := range entries {
		if e == want {
			return
		}
	}
	t.Errorf("GO_CODE_FILTER: go_code = %v, want it to contain %q as a parsed list entry — a workflow-only diff would classify as not-Go and skip the Go suite carrying the sync/main guard", entries, want)
}

// syncGuardMappingValue returns the value node for key in a YAML mapping node
// (local copy of the internal/template precedent; the `on:`-as-boolean quirk
// is why node values, not struct fields, are compared).
func syncGuardMappingValue(t *testing.T, mapping *yaml.Node, key string) *yaml.Node {
	t.Helper()
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		t.Fatalf("expected a mapping node to look up %q, got kind %v", key, mapping.Kind)
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	t.Fatalf("key %q not found in mapping", key)
	return nil
}

// syncGuardSequenceStrings returns the scalar strings of the sequence node at
// key in a YAML mapping node.
func syncGuardSequenceStrings(t *testing.T, mapping *yaml.Node, key string) []string {
	t.Helper()
	seq := syncGuardMappingValue(t, mapping, key)
	if seq.Kind != yaml.SequenceNode {
		t.Fatalf("expected a sequence node at %q, got kind %v", key, seq.Kind)
	}
	out := make([]string, 0, len(seq.Content))
	for _, item := range seq.Content {
		out = append(out, item.Value)
	}
	return out
}

// sweepWorkflowPushAdmission, parseWorkflowPushTrigger, and
// pushTriggerAdmitsBranch are implemented below (GREEN step; the RED run's
// undefined-symbol compile failure is the pre-GREEN evidence).

// --- The matcher (REQ-TBS-005 grammar, pinned subset) --------------------------

// globTokKind enumerates the matcher's token kinds. The supported grammar is
// EXACTLY the REQ-TBS-005 subset: '*' (one path segment, never crossing '/'),
// '**' (crosses '/'), '[...]' character classes (members, a-z ranges, leading
// '!' or '^' negation inside the class), and literals. A leading '!' is list
// negation and is handled by branchListAdmits, not here. Anything else — '?',
// braces, escapes, runs of three or more '*' — errors: fail-loud, never a
// silent non-match (a matcher that quietly undermatches would pass the live
// sweep while admitting the ref in production).
type globTokKind int

const (
	globLiteral  globTokKind = iota
	globStar                 // '*' — zero or more characters, no '/'
	globStarStar             // '**' — zero or more characters, '/' included
	globSeg                  // '**/' in segment position — zero or more whole segments
	globClass                // '[...]' — exactly one character
)

type globTok struct {
	kind    globTokKind
	ch      byte
	class   [256]bool
	classOn bool // class semantics are negated when the class leads with ! or ^
}

// lexBranchGlob parses one branch-filter pattern into tokens, erroring on any
// element outside the supported subset.
func lexBranchGlob(p string) ([]globTok, error) {
	var toks []globTok
	for i := 0; i < len(p); {
		switch c := p[i]; c {
		case '*':
			n := 1
			for i+n < len(p) && p[i+n] == '*' {
				n++
			}
			if n > 2 {
				return nil, fmt.Errorf("run of %d consecutive '*' is outside the supported grammar (*, **, [...] classes only)", n)
			}
			// A `**/` in SEGMENT position — preceded by the pattern start
			// or a '/', followed by '/' — is one token (amended
			// REQ-TBS-005, sync-audit round 1 S2): it matches zero or more
			// whole path segments, so `**/sync/main` and `sync/**/main`
			// both match sync/main. A `**` in any other position keeps the
			// zero-or-more-characters-including-'/' reading.
			if n == 2 && i+2 < len(p) && p[i+2] == '/' && (i == 0 || p[i-1] == '/') {
				toks = append(toks, globTok{kind: globSeg})
				i += 3 // the trailing '/' is part of the token
				continue
			}
			if n == 2 {
				toks = append(toks, globTok{kind: globStarStar})
			} else {
				toks = append(toks, globTok{kind: globStar})
			}
			i += n
		case '[':
			j := i + 1
			neg := false
			if j < len(p) && (p[j] == '!' || p[j] == '^') {
				neg = true
				j++
			}
			start := j
			var set [256]bool
			for j < len(p) && p[j] != ']' {
				// A range "a-z" spans its endpoints inclusive; a '-' with no
				// room for two endpoints around it is a literal member.
				if j+2 < len(p) && p[j+1] == '-' && p[j+2] != ']' {
					lo, hi := p[j], p[j+2]
					if lo > hi {
						return nil, fmt.Errorf("character class range %q is reversed", p[j:j+3])
					}
					for k := int(lo); k <= int(hi); k++ {
						set[k] = true
					}
					j += 3
				} else {
					set[p[j]] = true
					j++
				}
			}
			if j >= len(p) {
				return nil, fmt.Errorf("unclosed '[' character class")
			}
			if p[start:j] == "" {
				return nil, fmt.Errorf("empty '[' character class")
			}
			toks = append(toks, globTok{kind: globClass, class: set, classOn: neg})
			i = j + 1
		case '?', '{', '}', '\\', '+', '(', ')':
			return nil, fmt.Errorf("metacharacter %q is outside the supported grammar (*, **, [...] classes, leading ! only)", string(c))
		default:
			toks = append(toks, globTok{kind: globLiteral, ch: c})
			i++
		}
	}
	return toks, nil
}

// globMatchToks reports whether the token stream matches name. '*' extends
// only within a path segment; '**' extends across '/'; a class consumes
// exactly one character. Plain recursion: branch-filter patterns and branch
// names are tens of characters, not adversarial inputs.
func globMatchToks(toks []globTok, s string) bool {
	if len(toks) == 0 {
		return s == ""
	}
	switch tok := toks[0]; tok.kind {
	case globLiteral:
		return len(s) > 0 && s[0] == tok.ch && globMatchToks(toks[1:], s[1:])
	case globClass:
		if len(s) == 0 {
			return false
		}
		in := tok.class[s[0]]
		if tok.classOn {
			in = !in
		}
		return in && globMatchToks(toks[1:], s[1:])
	case globStar:
		for k := 0; k <= len(s); k++ {
			if k > 0 && s[k-1] == '/' {
				break // '*' never consumes '/'
			}
			if globMatchToks(toks[1:], s[k:]) {
				return true
			}
		}
		return false
	case globStarStar:
		for k := 0; k <= len(s); k++ {
			if globMatchToks(toks[1:], s[k:]) {
				return true
			}
		}
		return false
	case globSeg:
		// "**/" — zero or more complete segments, each WITH its trailing
		// '/'. Zero segments matches the empty string; one or more matches
		// any prefix that ends at a '/' boundary.
		if globMatchToks(toks[1:], s) {
			return true
		}
		for k := 0; k < len(s); k++ {
			if s[k] == '/' && globMatchToks(toks[1:], s[k+1:]) {
				return true
			}
		}
		return false
	}
	return false
}

// matchBranchGlob matches one branch-filter pattern (no leading '!') against
// a branch name.
func matchBranchGlob(pattern, name string) (bool, error) {
	if pattern == "" {
		return false, fmt.Errorf("empty pattern")
	}
	toks, err := lexBranchGlob(pattern)
	if err != nil {
		return false, err
	}
	return globMatchToks(toks, name), nil
}

// branchListAdmits evaluates a branches / branches-ignore list against a
// branch name under GitHub's last-match-wins semantics: patterns are tried in
// order, each MATCHING pattern sets the verdict (positive patterns admit,
// '!'-negated patterns deny), and the last match decides. A list where
// nothing matches denies — a branches filter is an allowlist. A '!'-negated
// pattern is evaluated by its core (the '!' only flips the verdict it sets,
// it does not widen what matches).
func branchListAdmits(patterns []string, name string) (bool, error) {
	allow := false
	for _, p := range patterns {
		core := strings.TrimPrefix(p, "!")
		ok, err := matchBranchGlob(core, name)
		if err != nil {
			return false, fmt.Errorf("pattern %q: %w", p, err)
		}
		if ok {
			allow = !strings.HasPrefix(p, "!")
		}
	}
	return allow, nil
}

// --- The on.push trigger parser -------------------------------------------------

// workflowPushTrigger is the parsed filter set of one workflow's on.push
// trigger. A nil slice means the filter key is absent; a present-but-empty
// list stays a zero-length non-nil slice and denies (fires on nothing).
type workflowPushTrigger struct {
	File           string
	Branches       []string
	BranchesIgnore []string
	Tags           []string
}

// parseWorkflowPushTrigger reads one workflow document and returns its
// on.push trigger, or nil when the workflow carries no push trigger. The
// `on:` key parses as a boolean scalar under YAML 1.1 rules, but the yaml.Node
// form keeps the raw key text, so keys are compared by node value (the
// in-tree precedent, internal/template/dotgithub_template_test.go).
func parseWorkflowPushTrigger(data []byte, name string) (*workflowPushTrigger, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("%s: parse YAML: %w", name, err)
	}
	if len(doc.Content) == 0 {
		return nil, fmt.Errorf("%s: empty document", name)
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s: workflow root is not a mapping (kind %v)", name, root.Kind)
	}
	onNode := syncGuardLookup(root, "on")
	if onNode == nil {
		return nil, nil // no event keys at all
	}
	switch onNode.Kind {
	case yaml.ScalarNode:
		if onNode.Value != "push" {
			return nil, nil
		}
		return &workflowPushTrigger{File: name}, nil
	case yaml.SequenceNode:
		for _, item := range onNode.Content {
			if item.Kind == yaml.ScalarNode && item.Value == "push" {
				return &workflowPushTrigger{File: name}, nil
			}
		}
		return nil, nil
	case yaml.MappingNode:
		pushNode := syncGuardLookup(onNode, "push")
		if pushNode == nil {
			return nil, nil
		}
		switch pushNode.Kind {
		case yaml.MappingNode:
			trig := &workflowPushTrigger{File: name}
			for _, spec := range []struct {
				key    string
				target *[]string
			}{
				{"branches", &trig.Branches},
				{"branches-ignore", &trig.BranchesIgnore},
				{"tags", &trig.Tags},
			} {
				node := syncGuardLookup(pushNode, spec.key)
				if node == nil {
					continue
				}
				vals, err := syncGuardStringSeq(node, name, spec.key)
				if err != nil {
					return nil, err
				}
				*spec.target = vals
			}
			return trig, nil
		case yaml.ScalarNode:
			// `push:` with nothing (null) or a bare scalar: a push trigger
			// with no filters. Boolean false is not a workflow form GitHub
			// documents; treating it as absent would silently drop a trigger
			// this guard exists to see, so it reads as present like every
			// other scalar.
			return &workflowPushTrigger{File: name}, nil
		default:
			return nil, fmt.Errorf("%s: on.push is a %v, not a filter mapping", name, pushNode.Kind)
		}
	default:
		return nil, fmt.Errorf("%s: on is a %v, not an event list or mapping", name, onNode.Kind)
	}
}

// pushTriggerAdmitsBranch reports whether the trigger fires for a push of the
// branch name, under GitHub's filter semantics (REQ-TBS-005):
//   - branches defined: the branch list must admit; an additionally defined
//     branches-ignore must NOT match (both defined = AND).
//   - branches-ignore only: fires on every branch the ignore list denies.
//   - tags only: fires on tag pushes only — never a branch push.
//   - neither: fires on every branch push (the bare `on: push` shape).
func pushTriggerAdmitsBranch(trig *workflowPushTrigger, branch string) (bool, error) {
	switch {
	case trig.Branches != nil:
		allow, err := branchListAdmits(trig.Branches, branch)
		if err != nil || !allow {
			return false, err
		}
		if trig.BranchesIgnore != nil {
			ignored, err := branchListAdmits(trig.BranchesIgnore, branch)
			if err != nil {
				return false, err
			}
			if ignored {
				return false, nil
			}
		}
		return true, nil
	case trig.BranchesIgnore != nil:
		ignored, err := branchListAdmits(trig.BranchesIgnore, branch)
		if err != nil {
			return false, err
		}
		return !ignored, nil
	case trig.Tags != nil:
		return false, nil // tags-only: not a branch-push trigger
	default:
		return true, nil // no filters: fires on every branch push
	}
}

// --- The sweep -------------------------------------------------------------------

// workflowSweepResult reports what one sweep observed.
type workflowSweepResult struct {
	WorkflowFiles int      // *.yml / *.yaml files enumerated
	PushTriggers  int      // workflows carrying an on.push trigger
	Admitting     []string // workflow file names whose push trigger admits the branch
}

// sweepWorkflowPushAdmission enumerates every *.yml / *.yaml file under
// workflowsDir (GitHub reads only the top level) and evaluates each on.push
// trigger's admission of branch. Zero workflow files, zero push triggers, and
// any parse or matcher error are all errors — the sweep never reports a
// verdict over a set it did not actually read (the empty-sweep guard).
func sweepWorkflowPushAdmission(workflowsDir, branch string) (workflowSweepResult, error) {
	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		return workflowSweepResult{}, fmt.Errorf("enumerate %s: %w", workflowsDir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".yml") || strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return workflowSweepResult{}, fmt.Errorf("no *.yml / *.yaml workflow files under %s — the sweep refuses to report over an empty set", workflowsDir)
	}
	sort.Strings(names)
	res := workflowSweepResult{WorkflowFiles: len(names)}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(workflowsDir, name))
		if err != nil {
			return workflowSweepResult{}, fmt.Errorf("read %s: %w", name, err)
		}
		trig, err := parseWorkflowPushTrigger(data, name)
		if err != nil {
			return workflowSweepResult{}, err
		}
		if trig == nil {
			continue
		}
		res.PushTriggers++
		admits, err := pushTriggerAdmitsBranch(trig, branch)
		if err != nil {
			return workflowSweepResult{}, fmt.Errorf("%s: %w", name, err)
		}
		if admits {
			res.Admitting = append(res.Admitting, name)
		}
	}
	if res.PushTriggers == 0 {
		return workflowSweepResult{}, fmt.Errorf("%d workflow files enumerated, none carries an on.push trigger — the sweep refuses to report over an empty push-trigger set", res.WorkflowFiles)
	}
	return res, nil
}

// syncGuardLookup returns the value node for key in a mapping node, or nil.
func syncGuardLookup(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// syncGuardStringSeq reads a filter's sequence of scalar strings. A non-
// sequence or non-scalar item is an error: filters this guard cannot read are
// failures, not passes.
func syncGuardStringSeq(node *yaml.Node, file, key string) ([]string, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("%s: on.push.%s is a %v, not a list", file, key, node.Kind)
	}
	out := make([]string, 0, len(node.Content))
	for _, item := range node.Content {
		if item.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("%s: on.push.%s contains a non-scalar entry (kind %v)", file, key, item.Kind)
		}
		out = append(out, item.Value)
	}
	return out, nil
}
