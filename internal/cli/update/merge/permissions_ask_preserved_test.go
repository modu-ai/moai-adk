package merge

// permissions_ask_preserved_test.go pins REQ-RC30-024 (U-15, AC-RC30-16): the
// settings merge keeps a user array under permissions.* that the template no
// longer carries. Hook arrays are out of scope for rc.30.
//
// The assertion runs through the production path (MergeUserFiles, via
// mergeSettings) on the canonical snapshot base, the same route as
// settings_snapshot_base_test.go.

import (
	"reflect"
	"testing"
)

// The base and the user copy both carry permissions.ask, unchanged. The
// template render keeps permissions.allow and has retired ask, so the merge
// must keep the user's ask list.
func TestPermissionsAskPreserved(t *testing.T) {
	base := `{"permissions":{"allow":["Bash(git status)"],"ask":["sudo","chmod","chown"]}}`
	render := `{"permissions":{"allow":["Bash(git status)"]}}`

	root := newSnapshotProject(t)
	writeRel(t, root, canonicalSnapshotRel, base)
	doc, _ := mergeSettings(t, root, base, render)

	want := []any{"sudo", "chmod", "chown"}
	if got := leaf(doc, "permissions", "ask"); !reflect.DeepEqual(got, want) {
		t.Errorf("permissions.ask = %v, want %v", got, want)
	}
}
