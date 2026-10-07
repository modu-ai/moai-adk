package template_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestAutoMergeRequiredChecks executes the live required-checks step. Empty
// or malformed observations must never authorize a merge.
func TestAutoMergeRequiredChecks(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("GitHub's merge step runs Bash on Ubuntu")
	}
	for _, tool := range []string{"bash", "jq"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not available", tool)
		}
	}
	root := findProjectRootForMirrorTest(t)
	raw, err := os.ReadFile(filepath.Join(root, ".github/workflows/auto-merge.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		Jobs map[string]struct {
			Steps []struct{ Name, Run string }
		}
	}
	if err := yaml.Unmarshal(raw, &workflow); err != nil {
		t.Fatal(err)
	}
	var script string
	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			if step.Name == "Check all required CI checks passed" {
				if script != "" {
					t.Fatal("duplicate required-checks step")
				}
				script = step.Run
			}
		}
	}
	if script == "" {
		t.Fatal("required-checks step missing")
	}
	for _, tc := range []struct {
		name, json, want, status string
	}{
		{"empty array", `[]`, "false", "0"},
		{"wrong shape", `{}`, "false", "0"},
		{"malformed", `not-json`, "false", "0"},
		{"missing bucket", `[{"name":"Lint","state":"SUCCESS"}]`, "false", "0"},
		{"failure", `[{"name":"Lint","state":"FAILURE","bucket":"fail"}]`, "false", "0"},
		{"unknown bucket", `[{"name":"Lint","bucket":"unknown"}]`, "false", "0"},
		{"missing name", `[{"bucket":"pass"}]`, "false", "0"},
		{"pending status without pending checks", `[{"name":"Lint","state":"SUCCESS","bucket":"pass"}]`, "false", "8"},
		{"missing state", `[{"name":"Lint","bucket":"pass"}]`, "false", "0"},
		{"failed status", `[{"name":"Lint","bucket":"fail"}]`, "false", "1"},
		{"failed state cannot pass", `[{"name":"Lint","state":"FAILURE","bucket":"pass"}]`, "false", "0"},
		{"pending state cannot pass", `[{"name":"Lint","state":"PENDING","bucket":"pass"}]`, "false", "0"},
		{"unknown state cannot pass", `[{"name":"Lint","state":"NOT_A_STATE","bucket":"pass"}]`, "false", "0"},
		{"neutral skips", `[{"name":"Lint","state":"NEUTRAL","bucket":"skipping"}]`, "false", "0"},
		{"skipped skips", `[{"name":"Lint","state":"SKIPPED","bucket":"skipping"}]`, "false", "0"},
		{"passed", `[{"name":"Lint","state":"SUCCESS","bucket":"pass"}]`, "true", "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			write := func(name, body string, mode os.FileMode) string {
				path := filepath.Join(dir, name)
				if err := os.WriteFile(path, []byte(body), mode); err != nil {
					t.Fatal(err)
				}
				return path
			}
			write("gh", "#!/bin/sh\ncat \"$CHECKS_FIXTURE\"\nexit \"$MOCK_CHECKS_EXIT\"\n", 0o755)
			fixture := write("checks.json", tc.json, 0o600)
			output := write("output", "", 0o600)
			body := write("step.sh", script, 0o600)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "bash", "-e", body)
			cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"PR_NUMBER=1", "GITHUB_OUTPUT="+output, "CHECKS_FIXTURE="+fixture, "MOCK_CHECKS_EXIT="+tc.status)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("step failed: %v\n%s", err, out)
			}
			if strings.HasSuffix(tc.name, "skips") && !strings.Contains(string(out), "Some checks did not pass") {
				t.Fatalf("valid skipping state was not classified: %s", out)
			}
			result, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(result)) != "should_merge="+tc.want {
				t.Fatalf("output=%q, want should_merge=%s\n%s", result, tc.want, out)
			}
		})
	}
}
