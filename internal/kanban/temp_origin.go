// temp_origin.go — SPEC-TODO-HOME-TEMP-GUARD-001 M1: the temporary-origin
// discriminant, as a PURE function plus its injection seam.
//
// M1 ships this UNWIRED: nothing in this package calls TempOriginReason yet.
// The wiring into the home-fallback branch of todo_root.go is M2's scope, and
// keeping the two apart is the plan's ordering constraint (§F) — the shape of
// the discriminant is the decision most likely to change, so it is pinned
// first and the wiring follows.
//
// Why the discriminant is not a prefix test (spec.md §4): os.TempDir() returns
// the UNRESOLVED spelling on macOS (/var/folders/...), while a caller path that
// passed through filepath.EvalSymlinks carries the /private/var/folders/...
// spelling. Comparing one resolved side against one unresolved side fails
// SILENTLY — the guard simply never fires. Both sides are therefore normalized
// by one rule, and the comparison is path-component containment rather than a
// raw string prefix so that a sibling like /tmpfoo is not read as living under
// /tmp.
package kanban

import (
	"os"
	"path/filepath"
	"strings"
)

// TempRootsFn is this package's temp-root-set injection seam
// (REQ-THG-009), the same class of seam as HomeDirFn above it: an exported
// package-level function variable, READ AT CALL TIME by TempOriginReason.
//
// Exported, and read at call time, for one measured reason each:
//
//   - Exported, because this repository's [HARD] test-isolation discipline puts
//     every fixture under t.TempDir() — which is BY DEFINITION inside
//     os.TempDir(). Without a seam there is no way to construct a
//     "non-temporary non-git base" fixture at all, so the guard would ship
//     unexercisable and REQ-THG-005 would lose its only producing acceptance
//     criterion. internal/cli and internal/web tests need the same escape, so
//     the seam cannot be package-private.
//   - Read at call time, because a test stub is always installed AFTER package
//     initialization. An implementation that snapshotted the set into a value
//     at init time would ignore every stub silently, and that single mistake
//     would make a whole family of acceptance assertions vacuous.
//
// A stub replaces the whole set; restore the original in a t.Cleanup.
var TempRootsFn = defaultTempRoots

// defaultTempRoots is the production temp-root set (REQ-THG-002).
//
// /tmp is NOT redundant with os.TempDir(): on this project's reference machine
// os.TempDir() reports /var/folders/..., so the one production contamination
// recovered by key inversion (~/.moai/todo/t203-probe-d7a16ea2, whose origin
// re-derives as sha256("/tmp/t203-probe")[:4]) would NOT have been caught by
// os.TempDir() alone.
//
// /var/tmp is deliberately EXCLUDED (spec.md §8): it survives reboots, so the
// "the directory later disappears" premise that makes an orphaned home queue
// unreachable is weak there, while the chance of a long-lived real project
// living under it is higher than under /tmp. The exclusion is fail-open —
// a /var/tmp launch keeps today's home fallback.
func defaultTempRoots() []string {
	return []string{os.TempDir(), "/tmp", "/var/folders"}
}

// TempOriginReason classifies base as a temporary-directory origin, reporting
// the matched temp root alongside the verdict so a caller's guidance can name
// what it matched (REQ-THG-002, REQ-THG-006). It is pure apart from the
// Lstat/EvalSymlinks reads normalization performs: it creates, moves, and
// writes nothing.
//
// Fail-open (REQ-THG-004): an empty base, a base whose absolute form cannot be
// derived, or any root that cannot be made absolute is reported NOT temporary.
// The two misclassification directions are not symmetric — a false negative
// leaves today's behaviour exactly as it is, while a false positive withdraws a
// real project's queue.
func TempOriginReason(base string) (reason string, isTemp bool) {
	if strings.TrimSpace(base) == "" {
		return "", false
	}
	abs, err := filepath.Abs(base)
	if err != nil {
		return "", false
	}
	normalizedBase := normalizeForTempCompare(abs)

	for _, root := range TempRootsFn() {
		if strings.TrimSpace(root) == "" {
			continue
		}
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		// Both spellings of the root are anchors: the normalized one (which on
		// macOS resolves /tmp -> /private/tmp) and the lexical one. A base that
		// does not exist normalizes lexically per the rule below, so it can only
		// meet a root that is spelled the same way — and the measured
		// contamination arrived in exactly the unresolved spelling. Widening the
		// anchor set only makes a genuinely-inside path easier to recognize; it
		// never admits a path that lies outside both spellings.
		for _, anchor := range []string{normalizeForTempCompare(rootAbs), filepath.Clean(rootAbs)} {
			if pathWithin(normalizedBase, anchor) {
				return root, true
			}
		}
	}
	return "", false
}

// normalizeForTempCompare is the single normalization rule both sides of every
// comparison pass through (REQ-THG-003): filepath.EvalSymlinks when the path
// exists, lexical filepath.Clean when it does not.
//
// The rule is adopted verbatim from internal/cli/launcher.go's resolveSymlinks
// for its documented GOOS determinism: filepath.EvalSymlinks on a NON-existent
// path diverges across platforms (windows partially resolves the existing
// prefix, expanding 8.3 short names), which breaks containment matching. This
// package keeps its own copy rather than reaching into internal/cli, because
// that function is unexported there; consolidating the two is deliberately out
// of this SPEC's scope.
func normalizeForTempCompare(path string) string {
	if _, err := os.Lstat(path); err != nil {
		return filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

// pathWithin reports whether p is dir itself or lies beneath it, compared by
// path COMPONENT rather than by raw string prefix (REQ-THG-002): "/tmpfoo" is
// not inside "/tmp", though a strings.HasPrefix test would say it is.
func pathWithin(p, dir string) bool {
	rel, err := filepath.Rel(dir, p)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
