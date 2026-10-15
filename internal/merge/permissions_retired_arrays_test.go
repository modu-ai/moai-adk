package merge

// permissions_retired_arrays_test.go — card t1619 (SPEC-UPDATE-REPAIR-RC30-001
// REQ-RC30-024): when the template retires a user array under permissions.*,
// the merge keeps the user's array instead of dropping it. These cases cover
// keepRetiredArrays and its caller in deepMergeMap (strategies.go lines 445-447
// and 496-516), which internal/merge's own tests did not reach.

import (
	"reflect"
	"testing"
)

// TestDeepMergeMapKeepsUserArrayRetiredByTemplate: the user left permissions
// unchanged, and the template dropped the deny array. The user's deny array is
// restored and reported as retained at permissions.deny.
func TestDeepMergeMapKeepsUserArrayRetiredByTemplate(t *testing.T) {
	base := map[string]any{
		"permissions": map[string]any{"allow": []any{"Bash(git:*)"}, "deny": []any{"Read(.env)"}},
	}
	current := map[string]any{
		"permissions": map[string]any{"allow": []any{"Bash(git:*)"}, "deny": []any{"Read(.env)"}},
	}
	updated := map[string]any{
		"permissions": map[string]any{"allow": []any{"Bash(git:*)"}},
	}

	merged, conflicts, retained := deepMergeMap(base, current, updated, "")

	if len(conflicts) != 0 {
		t.Errorf("a template-only change must not conflict: %+v", conflicts)
	}
	want := map[string]any{
		"allow": []any{"Bash(git:*)"},
		"deny":  []any{"Read(.env)"},
	}
	if got := merged["permissions"]; !reflect.DeepEqual(got, want) {
		t.Errorf("permissions = %#v, want the template's allow plus the user's retired deny %#v", got, want)
	}
	if wantRetained := []string{"permissions.deny"}; !reflect.DeepEqual(retained, wantRetained) {
		t.Errorf("retained = %v, want %v", retained, wantRetained)
	}
}

// TestDeepMergeMapTakesTemplatePermissionsWhenNotAMap: the template replaces the
// permissions object with a scalar. keepRetiredArrays has no map to compare and
// returns the template's value unchanged, with nothing retained.
func TestDeepMergeMapTakesTemplatePermissionsWhenNotAMap(t *testing.T) {
	base := map[string]any{
		"permissions": map[string]any{"allow": []any{"Bash(git:*)"}},
	}
	current := map[string]any{
		"permissions": map[string]any{"allow": []any{"Bash(git:*)"}},
	}
	updated := map[string]any{
		"permissions": "none",
	}

	merged, _, retained := deepMergeMap(base, current, updated, "")

	if got := merged["permissions"]; got != "none" {
		t.Errorf("permissions = %#v, want the template's scalar %q", got, "none")
	}
	if len(retained) != 0 {
		t.Errorf("retained = %v, want none", retained)
	}
}
