// goreleaser_prerelease_test.go: AC-GFD-009 / REQ-GFD-009. A tag with a
// pre-release suffix must publish a GitHub release flagged prerelease. GoReleaser
// does that when `release.prerelease` is `auto`; the unset default marks every
// release final. The assertion reads the YAML path, not a string, so the key in
// another block (or a different value) does not satisfy it.
package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGoreleaserPrereleaseAuto(t *testing.T) {
	t.Parallel()

	path := filepath.Join(findProjectRootForMirrorTest(t), ".goreleaser.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read .goreleaser.yml: %v", err)
	}
	var cfg struct {
		Release struct {
			Prerelease *string `yaml:"prerelease"`
		} `yaml:"release"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse .goreleaser.yml: %v", err)
	}
	if cfg.Release.Prerelease == nil {
		t.Fatalf("release.prerelease is not set: a -rc.N tag would publish a final GitHub release")
	}
	if got := *cfg.Release.Prerelease; got != "auto" {
		t.Errorf("release.prerelease = %q, want %q", got, "auto")
	}
}
