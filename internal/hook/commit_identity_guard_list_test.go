package hook

// commit_identity_guard_list_test.go — the built-in deny-list drift test
// (REQ-CIG-008 / AC-CIG-010).
//
// This file's NAME matters: it matches internal/hook/commit_identity_guard*_test.go,
// the exclusion glob of the enumeration predicate. The predicate excludes the
// guard's own test files because they carry out-of-list control values
// (dev@real-host.invalid, ops-bot@corp.invalid, …) that must never be
// promoted into the built-in list; the drift test re-runs the SAME predicate,
// so it must not see those controls either.
//
// The predicate below reproduces plan.md §B.1's measured pipeline byte for
// byte: grep -rhoE '<extraction>' --include='*_test.go'
// --exclude='commit_identity_guard*_test.go' internal pkg cmd
//   | sed -E 's/.*(=|", *")//; s/"$//' | grep '@' | sort -u
//
// If the predicate and the list drift apart (someone's new test writes a new
// fixture email, or a list entry is removed), this test fails naming the
// missing literal — the remedy is to add it to
// builtinCommitIdentityDenyEmails in commit_identity_guard.go, NOT to weaken
// this test. Empty-set passes are prohibited: the test asserts it scanned at
// least one file and found at least one literal (AC-CIG-010).

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
)

// fixtureEmailExtractionPattern is the grep -oE extraction of plan.md §B.1.
var fixtureEmailExtractionPattern = regexp.MustCompile(
	`(user\.email"?,? *"[^"]+"|user\.email=[^" ,)]+|GIT_(AUTHOR|COMMITTER)_EMAIL=[^" ,)]+)`)

// qCommaSpaceQuotePattern reproduces the sed alternation branch `", *"`.
var qCommaSpaceQuotePattern = regexp.MustCompile(`", *"`)

// sedToValue applies `s/.*(=|", *")//; s/"$//` to one extraction match: the
// greedy `.*` takes the LAST position where `=` or `", *"` matches, strips
// through it, then strips one trailing double quote.
func sedToValue(m string) string {
	lastEq := strings.LastIndex(m, "=")
	cut, n := -1, 0
	for _, loc := range qCommaSpaceQuotePattern.FindAllStringIndex(m, -1) {
		if loc[0] > cut {
			cut, n = loc[0], loc[1]-loc[0]
		}
	}
	if lastEq > cut {
		cut, n = lastEq, 1
	}
	if cut >= 0 {
		m = m[cut+n:]
	}
	return strings.TrimSuffix(m, `"`)
}

// moduleRootFromCaller walks up from this file's directory to the directory
// holding go.mod.
func moduleRootFromCaller(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed; cannot locate the module root")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no go.mod above %s", filepath.Dir(thisFile))
		}
		dir = parent
	}
}

// enumerateFixtureEmails re-runs the plan.md §B.1 predicate in-process and
// returns the literal set S, the number of *_test.go files scanned, and the
// number of '@'-bearing literals found.
func enumerateFixtureEmails(t *testing.T, root string) (set []string, filesScanned, literals int) {
	t.Helper()
	for _, top := range []string{"internal", "pkg", "cmd"} {
		dir := filepath.Join(root, top)
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			name := d.Name()
			if !strings.HasSuffix(name, "_test.go") {
				return nil
			}
			if ok, _ := filepath.Match("commit_identity_guard*_test.go", name); ok {
				return nil // the exclusion glob of the predicate
			}
			filesScanned++
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			for _, line := range strings.Split(string(data), "\n") {
				for _, m := range fixtureEmailExtractionPattern.FindAllString(line, -1) {
					v := sedToValue(m)
					if !strings.Contains(v, "@") {
						continue // the `grep '@'` stage
					}
					literals++
					if !slices.Contains(set, v) {
						set = append(set, v)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("enumerate %s: %v", dir, err)
		}
	}
	sort.Strings(set)
	return set, filesScanned, literals
}

// TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration is AC-CIG-010:
// the built-in list ⊇ S, the sweep is nonempty on both axes, and a control
// value living only in the excluded guard test files is NOT in the list.
func TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration(t *testing.T) {
	root := moduleRootFromCaller(t)
	set, filesScanned, literals := enumerateFixtureEmails(t, root)

	if filesScanned < 1 {
		t.Fatal("the enumeration swept 0 files; a green here is vacuous (AC-CIG-010 forbids empty-set passes)")
	}
	if literals < 1 {
		t.Fatal("the enumeration found 0 literals; a green here is vacuous (AC-CIG-010 forbids empty-set passes)")
	}

	var missing []string
	for _, email := range set {
		if !isDenyListed(email, builtinCommitIdentityDenyEmails) {
			missing = append(missing, email)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("fixture email literals found in the repository's *_test.go but missing from "+
			"builtinCommitIdentityDenyEmails (internal/hook/commit_identity_guard.go): %v — "+
			"add them to the built-in list; do NOT weaken this test or change the enumeration predicate", missing)
	}

	// Control: a value that exists only in the excluded guard test files
	// (this file and commit_identity_guard_test.go carry it) must NOT have
	// been promoted into the built-in list by the enumeration.
	for _, control := range []string{"dev@real-host.invalid", "ops-bot@corp.invalid"} {
		if isDenyListed(control, builtinCommitIdentityDenyEmails) {
			t.Fatalf("control value %q leaked into the built-in list; the exclusion glob is broken", control)
		}
	}

	// The incident identity and the plan snapshot values stay covered even if
	// their defining test files are ever renamed away.
	for _, anchor := range []string{"t@t.t", "t@t.test"} {
		if !isDenyListed(anchor, builtinCommitIdentityDenyEmails) {
			t.Fatalf("anchor literal %q is not in the built-in list (SPEC-COMMIT-IDENTITY-GUARD-001 root cause / plan.md §B.1)", anchor)
		}
	}
}
