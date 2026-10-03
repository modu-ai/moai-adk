// releasecheck_test.go — SPEC-PLUGIN-MARKETPLACE-001 M4, AC-024: the release
// coupling lives with the generator that writes the plugin version. The script
// compares the committed plugin manifest with a release tag; the workflow must
// call it inside verify-provenance.
package pluginemit_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/pkg/version"
)

const versionCheckScript = "scripts/check-plugin-version.sh"

// runVersionScript runs the script from the repository root with args and
// returns its combined output and exit code.
func runVersionScript(t *testing.T, args ...string) (string, int) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the script is a POSIX shell script")
	}
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh on PATH")
	}
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, versionCheckScript)); err != nil {
		t.Errorf("%s is absent: %v", versionCheckScript, err)
		return "", -1
	}
	cmd := exec.Command(sh, append([]string{versionCheckScript}, args...)...)
	cmd.Dir = root
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	exit := 0
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run script: %v", err)
		}
		exit = ee.ExitCode()
	}
	return out.String(), exit
}

func TestPluginVersionScript(t *testing.T) {
	t.Run("ssot-tag-accepted", func(t *testing.T) {
		// The tag is the version SSOT itself, read here, so the criterion holds
		// no version literal and survives a bump.
		tag := version.Version
		if !strings.HasPrefix(tag, "v") {
			t.Fatalf("version.Version %q has no leading v; the leading-v strip would go untested", tag)
		}
		out, exit := runVersionScript(t, tag)
		if exit != 0 {
			t.Errorf("tag %s: exit %d, want 0; output:\n%s", tag, exit, out)
		}
	})
	t.Run("bare-tag-accepted", func(t *testing.T) {
		out, exit := runVersionScript(t, strings.TrimPrefix(version.Version, "v"))
		if exit != 0 {
			t.Errorf("exit %d, want 0; output:\n%s", exit, out)
		}
	})
	t.Run("other-tag-rejected-names-both", func(t *testing.T) {
		out, exit := runVersionScript(t, "v9.9.9")
		if exit != 1 {
			t.Errorf("exit %d, want 1; output:\n%s", exit, out)
		}
		committed := strings.TrimPrefix(version.Version, "v")
		for _, want := range []string{"9.9.9", committed} {
			if !strings.Contains(out, want) {
				t.Errorf("output lacks %q:\n%s", want, out)
			}
		}
	})
	t.Run("missing-argument-exit-2", func(t *testing.T) {
		out, exit := runVersionScript(t)
		if exit != 2 {
			t.Errorf("exit %d, want 2; output:\n%s", exit, out)
		}
	})
}

// TestReleaseWorkflowCallsPluginVersionCheck parses release.yml and requires a
// non-comment line, inside a step of job verify-provenance, that runs the
// script with the tag.
func TestReleaseWorkflowCallsPluginVersionCheck(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(repoRoot, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var wf struct {
		Jobs map[string]struct {
			Steps []struct {
				Name string `yaml:"name"`
				Run  string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(raw, &wf); err != nil {
		t.Fatalf("parse release.yml: %v", err)
	}
	job, ok := wf.Jobs["verify-provenance"]
	if !ok || len(job.Steps) == 0 {
		t.Fatal("job verify-provenance is missing or has no steps; an empty sweep asserts nothing")
	}
	calls := func(steps []struct {
		Name string `yaml:"name"`
		Run  string `yaml:"run"`
	}) bool {
		for _, s := range steps {
			for _, line := range strings.Split(s.Run, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "#") {
					continue
				}
				if strings.Contains(line, "scripts/check-plugin-version.sh") && strings.Contains(line, "${TAG}") {
					return true
				}
			}
		}
		return false
	}
	if !calls(job.Steps) {
		t.Error("no non-comment line of a verify-provenance step runs scripts/check-plugin-version.sh with ${TAG}")
	}
	for name, j := range wf.Jobs {
		if name != "verify-provenance" && calls(j.Steps) {
			t.Logf("note: job %s also calls the script", name)
		}
	}
}
