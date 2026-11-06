package yamlpatch

// yamlpatch_remove_anchor_t2109_test.go — card t2109 (follow-up of the t1974
// audit): key REMOVAL did not get the anchor refusal that t1974 gave to setting a
// value. Deleting the key that defines an anchor leaves its aliases pointing at
// nothing ("unknown anchor", the file no longer parses), and deleting a key inside
// an anchored mapping deletes it from every alias of that mapping. The removal is
// refused with an error naming the anchor and the file stays byte-identical.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func patchRemove(t *testing.T, in string, path ...string) (out string, err error) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "user.yaml")
	if werr := os.WriteFile(p, []byte(in), 0o600); werr != nil {
		t.Fatal(werr)
	}
	err = PatchFile(p, []KeyEdit{{Path: path, Remove: true}})
	data, rerr := os.ReadFile(p)
	if rerr != nil {
		t.Fatal(rerr)
	}
	return string(data), err
}

func TestPatchFileRefusesRemovingAnchoredKeys(t *testing.T) {
	cases := []struct {
		name string
		in   string
		path []string
	}{
		{"key_whose_mapping_value_defines_an_anchor", "base: &b\n  x: 1\nuser:\n  name: Old\n  other: *b\n", []string{"base"}},
		{"key_whose_scalar_value_defines_an_anchor", "user:\n  name: &n Old\n  nick: *n\n", []string{"user", "name"}},
		{"key_inside_an_anchored_mapping_shared_by_an_alias", "user: &u\n  name: Old\n  nick: k\nteam: *u\n", []string{"user", "name"}},
		{"removed_subtree_holds_an_anchor_definition", "user:\n  nested:\n    k: &a V\n  nick: *a\n", []string{"user", "nested"}},
		// audit t2109 F2: two wrong implementations passed the first four cases
		{"the_key_node_itself_carries_the_anchor", "user:\n  &k name: Old\n  other: *k\n", []string{"user", "name"}},
		{"an_anchored_ancestor_above_the_direct_parent", "a: &x\n  b:\n    c: 1\n    d: 2\ncopy: *x\n", []string{"a", "b", "c"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := patchRemove(t, tc.in, tc.path...)
			if err == nil || !strings.Contains(err.Error(), "anchor") {
				t.Fatalf("err = %v, want a refusal naming the anchor", err)
			}
			if out != tc.in {
				t.Errorf("a refused removal changed the file:\n--- before\n%s--- after\n%s", tc.in, out)
			}
		})
	}
}

// Controls: removal that touches no anchor keeps working, and so does removal of an
// absent key, so the refusal is not a blanket ban on Remove.
func TestPatchFileStillRemovesUnanchoredKeys(t *testing.T) {
	t.Run("plain_key", func(t *testing.T) {
		out, err := patchRemove(t, "user:\n  name: Old\n  nick: x\n", "user", "name")
		if err != nil {
			t.Fatalf("plain removal refused: %v", err)
		}
		if strings.Contains(out, "name:") || !strings.Contains(out, "nick: x") {
			t.Errorf("removal result wrong:\n%s", out)
		}
	})
	t.Run("an_anchor_elsewhere_in_the_file_is_unrelated", func(t *testing.T) {
		in := "base: &b\n  x: 1\nuser:\n  name: Old\n  other: *b\n"
		out, err := patchRemove(t, in, "user", "name")
		if err != nil {
			t.Fatalf("removal next to an unrelated anchor refused: %v", err)
		}
		if strings.Contains(out, "name:") || !strings.Contains(out, "&b") || !strings.Contains(out, "*b") {
			t.Errorf("removal result wrong:\n%s", out)
		}
	})
	// audit t2109 F1: removing a key that is not there changes nothing, so it must not be
	// refused just because its parent mapping carries an anchor. This is the shape of the
	// launcher's write: one real edit elsewhere forces the re-encode path, and the legacy
	// alias keys it removes are already gone from an anchored llm.glm.models.
	t.Run("absent_key_under_an_anchored_mapping_beside_a_real_edit", func(t *testing.T) {
		in := "llm:\n  team_mode: a\n  glm:\n    models: &m\n      high: x\ncopy: *m\n"
		p := filepath.Join(t.TempDir(), "llm.yaml")
		if err := os.WriteFile(p, []byte(in), 0o600); err != nil {
			t.Fatal(err)
		}
		err := PatchFile(p, []KeyEdit{
			{Path: []string{"llm", "added_key"}, Value: "v"},
			{Path: []string{"llm", "glm", "models", "opus"}, Remove: true},
			{Path: []string{"llm", "glm", "models", "sonnet"}, Remove: true},
		})
		if err != nil {
			t.Fatalf("a write whose removals are absent keys was refused: %v", err)
		}
		data, _ := os.ReadFile(p)
		out := string(data)
		if !strings.Contains(out, "added_key") || !strings.Contains(out, "&m") || !strings.Contains(out, "*m") || !strings.Contains(out, "high: x") {
			t.Errorf("result lost the edit or the anchor:\n%s", out)
		}
	})
	t.Run("absent_key", func(t *testing.T) {
		in := "user:\n  name: Old\n"
		out, err := patchRemove(t, in, "user", "missing")
		if err != nil {
			t.Fatalf("removing an absent key failed: %v", err)
		}
		if out != in {
			t.Errorf("removing an absent key changed the file:\n%s", out)
		}
	})
}
