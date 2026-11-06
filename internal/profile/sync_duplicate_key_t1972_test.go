package profile

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// t1972 N6 (carried from t1966): in a user.yaml that carries the `name` key twice,
// a name splice rewrote only the first row and reported success, leaving a file
// that is still not valid YAML. The card recorded the case as harmless because
// "moai's loader rejects such a file"; measured here, the loader does not refuse it
// (it logs "failed to load user config, using defaults" and carries on), so the
// sync must refuse instead of splicing half of a broken file.
func TestDuplicateKeyUserYAMLIsNeverPartlySpliced(t *testing.T) {
	const dup = "user:\n  name: \"Alice\"\n  name: \"Bob\"\n"

	t.Run("premise: the fixture really is invalid YAML", func(t *testing.T) {
		var doc map[string]any
		if err := yaml.Unmarshal([]byte(dup), &doc); err == nil {
			t.Fatal("the duplicate-key fixture parses; the test would not exercise the case")
		}
	})

	t.Run("a name sync over it fails and leaves the file byte-identical", func(t *testing.T) {
		root := t.TempDir()
		seedUserFixture(t, root, dup)
		err := SyncToProjectConfig(root, ProfilePreferences{UserName: "Carol"})
		if err == nil {
			t.Fatalf("sync over a duplicate-key user.yaml reported success:\n%s", readUserFixture(t, root))
		}
		if got := readUserFixture(t, root); got != dup {
			t.Errorf("a refused sync still changed the file:\nbefore:\n%s\nafter:\n%s", dup, got)
		}
	})

	t.Run("control: a valid user.yaml is still spliced", func(t *testing.T) {
		root := t.TempDir()
		seedUserFixture(t, root, "user:\n  name: \"Alice\"\n")
		if err := SyncToProjectConfig(root, ProfilePreferences{UserName: "Carol"}); err != nil {
			t.Fatalf("sync over a valid user.yaml: %v", err)
		}
		if got := readUserFixture(t, root); got != "user:\n  name: \"Carol\"\n" {
			t.Errorf("valid file not spliced as before:\n%s", got)
		}
	})
}
