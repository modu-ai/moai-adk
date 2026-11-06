package yamlpatch

// yamlpatch_anchor_t1974_test.go — card t1974: a key edit that reaches a node
// carrying a YAML anchor (the node itself, or a mapping on the path to it)
// changes every alias of that node, so a hand-written file with a shared anchor
// had a sibling key altered by an edit of another key. The edit is refused with
// an error and the file stays byte-identical; an alias node and an alias parent
// were already refused and stay so.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func patchUser(t *testing.T, in string) (out string, err error) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "user.yaml")
	if werr := os.WriteFile(p, []byte(in), 0o600); werr != nil {
		t.Fatal(werr)
	}
	err = PatchFile(p, []KeyEdit{{Path: []string{"user", "name"}, Value: "New"}})
	data, rerr := os.ReadFile(p)
	if rerr != nil {
		t.Fatal(rerr)
	}
	return string(data), err
}

func TestPatchFileRefusesAnchoredNodesOnThePath(t *testing.T) {
	for name, in := range map[string]string{
		"anchored_scalar_with_sibling_alias":      "user:\n  name: &n Old\n  nick: *n\n",
		"anchored_parent_shared_by_alias":         "user: &u\n  name: Old\nteam: *u\n",
		"upsert_into_an_anchored_parent":          "user: &u\n  other: keep\nteam: *u\n",
		"anchored_scalar_without_any_alias_known": "user:\n  name: &n Old\n",
	} {
		t.Run(name, func(t *testing.T) {
			out, err := patchUser(t, in)
			if err == nil || !strings.Contains(err.Error(), "anchor") {
				t.Fatalf("err = %v, want a refusal naming the anchor", err)
			}
			if out != in {
				t.Errorf("a refused edit changed the file:\n--- before\n%s--- after\n%s", in, out)
			}
		})
	}
}

func TestPatchFileKeepsRefusingAliasNodes(t *testing.T) {
	for name, in := range map[string]string{
		"name_is_an_alias_node":    "base: &b Shared\nuser:\n  name: *b\n  other: keep\n",
		"user_mapping_is_an_alias": "tmpl: &u\n  name: Old\nuser: *u\n",
	} {
		t.Run(name, func(t *testing.T) {
			out, err := patchUser(t, in)
			if err == nil {
				t.Fatal("an edit through an alias node was accepted")
			}
			if out != in {
				t.Errorf("a refused edit changed the file:\n--- before\n%s--- after\n%s", in, out)
			}
		})
	}
}

// Control: an anchor that is NOT on the edited path does not block the edit.
func TestPatchFileEditsBesideAnAnchorNotOnThePath(t *testing.T) {
	in := "defaults: &d\n  x: 1\nuser:\n  name: Old\n  extra: *d\n"
	out, err := patchUser(t, in)
	if err != nil {
		t.Fatalf("an anchor off the path blocked the edit: %v", err)
	}
	if !strings.Contains(out, "name: New") || !strings.Contains(out, "defaults: &d") || !strings.Contains(out, "extra: *d") {
		t.Errorf("unexpected result:\n%s", out)
	}
}
