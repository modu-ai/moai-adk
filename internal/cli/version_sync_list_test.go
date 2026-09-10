// Package cli — version_sync_list_test.go
//
// Guard over the version-stamp list in `.moai/docs/version-management.md`.
//
// The guarantee this check establishes is PARTIAL, and the half it does not
// cover is the half that caused the incident which motivated it.
//
//	CAUGHT — the list names a path that does not exist (a ghost). The v3.1.4
//	         list named `internal/template/templates/.moai/config/config.yaml`,
//	         which has no file behind it.
//	NOT CAUGHT — a real stamp site is absent from the list (an omission). The
//	         v3.1.3 bump rewrote six files and missed `docs-site/hugo.toml`,
//	         leaving the tree carrying two different versions. This check reads
//	         only what the list already names, so it cannot see a site the list
//	         forgot. Closing that direction needs a discriminator between this
//	         repository's own version tokens and everyone else's, which is card
//	         t392's first decision — not this file's.
//
// A silent way to lose the check: rename the stamp subheading. The anchor below
// would then match nothing, the scan would come back empty, and "zero
// violations" would be indistinguishable from "never ran". That is why the
// entry count is asserted against a constant this file holds, rather than
// against anything derived from the parse — a parse compared with itself is
// always equal and asserts nothing.
//
// The check reads the working tree only. It queries no repository object, no
// history, and no other branch.
//
// SPEC-VERSION-STAMP-GUARD-001 (REQ-VSG-004, REQ-VSG-005).

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	// versionSyncDocPath is the document whose list this check reads.
	versionSyncDocPath = ".moai/docs/version-management.md"

	// versionStampAnchor opens the scanned section. The subheading in the
	// document and this constant are one string: change either and the other
	// changes in the same commit, or the check goes quiet.
	versionStampAnchor = "**Version Stamps:**"

	// releaseArtifactAnchor is the sibling subheading, listed here to state
	// that it is deliberately NOT judged. Its entries are placeholders
	// (`.moai/release-notes/vX.Y.Z.ko.md`), and judging them for existence
	// would report a correct entry as a ghost.
	releaseArtifactAnchor = "**Release Artifacts:**"

	// expectedVersionStampEntries is the number of paths a version bump
	// rewrites, held as a constant rather than derived from the parse. When
	// the document's entry count legitimately changes, a human updates this
	// number in the same commit.
	expectedVersionStampEntries = 7
)

// TestVersionSyncListNamesOnlyExistingPaths reads the version-stamp entries and
// makes two independent assertions: that the scan found the expected number of
// entries, and that every path it found exists.
//
// Both report non-fatally. A tree exists in which both fail at once, and a
// fatal first report would end the run before the existence assertion could
// name its path.
func TestVersionSyncListNamesOnlyExistingPaths(t *testing.T) {
	root := repoRootFromCLITest(t)
	docPath := filepath.Join(root, filepath.FromSlash(versionSyncDocPath))

	raw, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read %s: %v", versionSyncDocPath, err)
	}

	entries := parseVersionStampEntries(string(raw))

	// Non-emptiness assertion. A scan that matched nothing is a failure, never
	// a pass — see this file's header comment.
	if len(entries) != expectedVersionStampEntries {
		t.Errorf("version-stamp entries: parsed=%d expected=%d (anchor %q in %s)",
			len(entries), expectedVersionStampEntries, versionStampAnchor, versionSyncDocPath)
	}

	// Existence assertion. Names the offending path so the reader does not have
	// to diff the list to find out which entry rotted.
	for _, entry := range entries {
		if _, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(entry))); statErr != nil {
			t.Errorf("version-sync list names a path that does not exist: %s", entry)
		}
	}
}

// parseVersionStampEntries extracts the path tokens listed under the
// version-stamp anchor. The section runs from the line after the anchor to the
// line before the next bold label or the next `###` heading, whichever comes
// first, so the release-artifact entries below it are never returned.
//
// An entry is a `- ` bullet. The trailing parenthetical comment the document
// uses (`- README.md (Version line)`) is dropped, and backticks are stripped,
// so a later reformatting that wraps paths in backticks does not silently
// change what is judged. Prose lines inside the section — the note explaining
// why `system.yaml.tmpl` is not a stamp — are not bullets and are not returned.
func parseVersionStampEntries(doc string) []string {
	var entries []string
	inSection := false

	for _, line := range strings.Split(doc, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == versionStampAnchor {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}
		if trimmed == releaseArtifactAnchor || strings.HasPrefix(trimmed, "###") || isBoldLabel(trimmed) {
			break
		}
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}

		item := strings.TrimPrefix(trimmed, "- ")
		if idx := strings.Index(item, "("); idx >= 0 {
			item = item[:idx]
		}
		item = strings.TrimSpace(strings.ReplaceAll(item, "`", ""))
		if item != "" {
			entries = append(entries, item)
		}
	}

	return entries
}

// isBoldLabel reports whether the line is a bold label of the
// `**Something:**` form the document uses to open a subheading.
func isBoldLabel(line string) bool {
	return strings.HasPrefix(line, "**") && strings.HasSuffix(line, ":**") && len(line) > len("**:**")
}
