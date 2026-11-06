package profile

// sync_anchor_t1974_test.go — card t1974: `moai profile set` of the user name
// on a hand-written user.yaml whose name sits under a YAML anchor/alias. The
// edit is refused with an error and the file is unchanged (the disposition the
// clear path got in card t1966), instead of silently changing a sibling key.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncToProjectConfigUserNameRefusesAnchorsAndAliases(t *testing.T) {
	for name, in := range map[string]string{
		"anchored_name_with_sibling_alias": "user:\n  name: &n original\n  nick: *n\n",
		"anchored_user_shared_by_alias":    "user: &u\n  name: original\nteam: *u\n",
		"name_absent_in_anchored_user":     "user: &u\n  other: keep\nteam: *u\n",
		"name_is_an_alias_node":            "base: &b Shared\nuser:\n  name: *b\n",
	} {
		t.Run(name, func(t *testing.T) {
			projectRoot := t.TempDir()
			setupProjectConfig(t, projectRoot)
			path := filepath.Join(projectRoot, ".moai", "config", "sections", "user.yaml")
			if err := os.WriteFile(path, []byte(in), 0o644); err != nil {
				t.Fatal(err)
			}
			err := SyncToProjectConfig(projectRoot, ProfilePreferences{UserName: "newuser"})
			if err == nil || !strings.Contains(err.Error(), "sync user name") {
				t.Fatalf("err = %v, want a refusal from the user-name sync", err)
			}
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				t.Fatal(rerr)
			}
			if string(data) != in {
				t.Errorf("a refused sync changed user.yaml:\n--- before\n%s--- after\n%s", in, data)
			}
		})
	}
}
