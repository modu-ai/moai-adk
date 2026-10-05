package graph

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ShrinkDefect names one removed file-sourced edge together with the source
// file that made its removal unexplainable.
type ShrinkDefect struct {
	// Edge is the edge present in the existing artifact and absent from the
	// rebuild.
	Edge Edge
	// SourceFile is the DECODED, project-relative file part of Edge.Source
	// (splitCodeNode for the compound code-call shape).
	SourceFile string
}

// ShrinkReport is the shrink guard's verdict: the removed edges whose source
// files still exist on disk yet lay outside the rebuild's scanned set. An
// empty report permits the overwrite.
type ShrinkReport struct {
	// Defects is deterministic: existing-artifact order.
	Defects []ShrinkDefect
}

// Empty reports whether the guard found nothing (overwrite permitted).
func (r ShrinkReport) Empty() bool { return len(r.Defects) == 0 }

// Describe renders the refusal: each removed edge and its unscanned source
// file, so a partial rebuild failure is diagnosable rather than silent.
func (r ShrinkReport) Describe() string {
	if len(r.Defects) == 0 {
		return "no unexplained shrink"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "unexplained shrink: %d removed file-sourced edge(s) whose sources exist on disk but were outside the scanned set:", len(r.Defects))
	for _, d := range r.Defects {
		fmt.Fprintf(&b, "\n  %s %s -> %s (source file %s not in scanned set)", d.Edge.Kind, d.Edge.Source, d.Edge.Target, d.SourceFile)
	}
	return b.String()
}

// ShrinkRefusalError is the typed refusal the automatic write paths return
// pre-write: callers distinguish it from mechanical failures to apply the
// fail-safe shape (answer from the existing artifact / skip the refresh)
// rather than a generic error path.
type ShrinkRefusalError struct {
	Report ShrinkReport
}

// Error implements error with the report's own naming.
func (e *ShrinkRefusalError) Error() string { return e.Report.Describe() }

// shrinkGuardedKinds is the guard's kind scope (REQ-GR-008): the file-sourced
// kinds — code-call (compound `file:function` Source) and code-import (bare
// file-path Source) — exactly the kinds whose Source is a project-relative
// file drawn from the SAME extraction walk that produces the scanned set, so
// the existence discriminator ("file exists AND outside the scanned set") is
// well-defined for them.
//
// Every other kind is OUT of scope because its Source is not a file path
// drawn from that walk: doc-import Sources are package/directory names and
// spec-depends Sources are SPEC IDs (their shrinkage is explained by their
// own rebuild inputs — the parsed dependency markdown and SPEC frontmatter);
// mx-spec Sources are file paths but drawn from the mx-index scan universe, a
// DIFFERENT input than the extraction walk, so testing them against the
// extraction's scanned set would misreport ordinary refreshes. Inventing
// per-kind existence tests would fork the guard into divergent
// implementations; the discriminator stays defined only where it is sound.
func shrinkGuardedKinds() map[string]bool {
	return map[string]bool{KindCodeCall: true, KindCodeImport: true}
}

// DetectUnexplainedShrink compares a rebuild's edge set against the existing
// artifact and names the removed edges whose removal the rebuild cannot
// explain (REQ-GR-008, the graphify #1116 pattern).
//
// The trigger is the SET DIFFERENCE `existing − rebuilt` — evaluated per edge
// identity (Kind, Source, Target; positional and annotation fields are not
// identity) regardless of total counts, so an equal-cardinality substitution
// that loses one edge while adding another cannot evade evaluation.
//
// For each removed edge in the kind scope above, the file part of the
// compound Source is DECODED first (splitCodeNode — the undecoded
// `file:function` string is never stat'ed) and validated as project-relative
// (no `..`, absolute, or symlink escape) BEFORE the existence test, which
// resolves the path under projectRoot. The refusal fires only when the
// source file BOTH still exists on disk AND lies outside scannedSources (the
// file list the rebuild's own extraction loop actually processed): a file
// absent from disk is a genuine deletion (file gone ⇒ not in the scanned set
// ⇒ not a shrink defect), and a file inside the scanned set had its
// relationships observed by the scan itself. Paths that fail validation, or
// whose existence cannot be determined, are skipped — never probed outside
// the project tree.
//
// @MX:ANCHOR: [AUTO] single choke point for every automatic write path's shrink evaluation — refresh, build, and the deferred path all consume this one verdict
// @MX:REASON: forking the discriminator per path would let one path accept a shrink another refuses; the graphify #1116 incident shape this guard exists to close
// @MX:SPEC:SPEC-GRAPH-REPORT-001
func DetectUnexplainedShrink(existing, rebuilt []Edge, scannedSources map[string]bool, projectRoot string) ShrinkReport {
	guarded := shrinkGuardedKinds()

	// Rebuilt identities: kind+source+target (relationship identity).
	rebuiltIDs := make(map[string]bool, len(rebuilt))
	for _, e := range rebuilt {
		rebuiltIDs[edgeID(e)] = true
	}

	resolvedRoot, rootErr := filepath.EvalSymlinks(projectRoot)
	if rootErr != nil {
		resolvedRoot = projectRoot // unresolvable root: fall through to the plain join; validation still gates the stat
	}

	report := ShrinkReport{}
	for _, e := range existing {
		if !guarded[e.Kind] || rebuiltIDs[edgeID(e)] {
			continue
		}
		file := decodeSourceFile(e.Kind, e.Source)
		if !isSafeProjectRelative(file) {
			continue
		}
		exists, determinate := fileExistsUnderRoot(resolvedRoot, file)
		if !determinate || !exists || scannedSources[file] {
			continue
		}
		report.Defects = append(report.Defects, ShrinkDefect{Edge: e, SourceFile: file})
	}
	return report
}

// edgeID is the relationship identity behind the set difference: Kind, Source,
// Target. Line numbers shift on edits and annotation fields change with
// layer disagreement — neither removes the relationship.
func edgeID(e Edge) string {
	return e.Kind + "\x00" + e.Source + "\x00" + e.Target
}

// decodeSourceFile extracts the file part of a Source: the code-call compound
// `file:function` shape via splitCodeNode; file-sourced kinds whose Source is
// already a bare path pass through. The result is slash-normalized so the
// scanned-set comparison is separator-consistent (the extraction emits
// slash-normalized repo-relative paths).
func decodeSourceFile(kind, source string) string {
	if kind == KindCodeCall {
		file, _ := splitCodeNode(source) // the function part is not part of the existence test
		return filepath.ToSlash(file)
	}
	return filepath.ToSlash(source)
}

// isSafeProjectRelative validates a slash-separated path as a clean
// project-relative file path: not empty, not absolute, no `..` segment, and
// no empty segment (which a leading or doubled separator would leave).
func isSafeProjectRelative(file string) bool {
	if file == "" || strings.HasPrefix(file, "/") {
		return false
	}
	if filepath.IsAbs(filepath.FromSlash(file)) { // windows drive letters
		return false
	}
	for _, seg := range strings.Split(file, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
	}
	return true
}

// fileExistsUnderRoot reports whether root/file exists on disk, resolving
// symlinks and refusing to answer for a path that escapes the root. ok=false
// means the existence test was indeterminate (validation-grade failure) — the
// caller skips the edge rather than probing outside the tree.
func fileExistsUnderRoot(root, file string) (exists, ok bool) {
	abs := filepath.Join(root, filepath.FromSlash(file))
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, true // determinate: the file is genuinely gone (a legitimate deletion)
		}
		return false, false
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false, false
	}
	if resolved != resolvedRoot && !strings.HasPrefix(resolved, resolvedRoot+string(os.PathSeparator)) {
		return false, false // symlink escape: outside the project tree — never evaluated
	}
	return true, true
}
