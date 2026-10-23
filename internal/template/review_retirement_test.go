package template_test

// SPEC-CI-CODEX-REVIEW-001 M4 — retirement of the earlier reviewers (AC-CCR-014, AC-CCR-015).
//
// Both tests read the repository's own files (the real workflow directory, the real scripts/ci
// directory, the real .coderabbit.yaml); neither reads a template tree.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func rrWorkflowFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	for _, pat := range []string{"*.yml", "*.yaml"} {
		m, err := filepath.Glob(filepath.Join(root, ".github", "workflows", pat))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, m...)
	}
	sort.Strings(files)
	return files
}

// rrTriggers returns the event names an `on:` value declares, whatever its shape (scalar, list, map).
func rrTriggers(on any) []string {
	var out []string
	switch v := on.(type) {
	case string:
		out = append(out, v)
	case []any:
		for _, e := range v {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
	case map[string]any:
		for k := range v {
			out = append(out, k)
		}
	}
	return out
}

// AC-CCR-014 (REQ-CCR-014) — no workflow runs anthropics/claude-code-action, carries the name
// "Claude Code Review", or triggers on check_run; the three retired workflow files do not exist.
// claude-code-review.yml is remote-only: this census turns red if a merge brings it back.
//
// Mutant: restoring claude.yml, or merging the remote claude-code-review.yml back in, turns a row red.
func TestNoClaudeReviewWorkflows(t *testing.T) {
	t.Parallel()
	root := findProjectRootForMirrorTest(t)
	files := rrWorkflowFiles(t, root)
	t.Run("no_claude_code_action", func(t *testing.T) {
		for _, f := range files {
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), "anthropics/claude-code-action") {
				t.Errorf("%s runs anthropics/claude-code-action", filepath.Base(f))
			}
		}
	})
	parse := func(t *testing.T, f string) (name string, triggers []string) {
		t.Helper()
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var wf struct {
			Name string `yaml:"name"`
			On   any    `yaml:"on"`
		}
		if err := yaml.Unmarshal(raw, &wf); err != nil {
			t.Fatalf("parse %s: %v", filepath.Base(f), err)
		}
		return wf.Name, rrTriggers(wf.On)
	}
	t.Run("no_workflow_named_claude_code_review", func(t *testing.T) {
		for _, f := range files {
			if name, _ := parse(t, f); name == "Claude Code Review" {
				t.Errorf("%s is named %q", filepath.Base(f), name)
			}
		}
	})
	t.Run("no_check_run_trigger", func(t *testing.T) {
		for _, f := range files {
			_, triggers := parse(t, f)
			for _, tr := range triggers {
				if tr == "check_run" {
					t.Errorf("%s triggers on check_run", filepath.Base(f))
				}
			}
		}
	})
	for _, tc := range []struct{ name, file string }{
		{"claude_yml_absent", "claude.yml"},
		{"review_quality_gate_yml_absent", "review-quality-gate.yml"},
		{"claude_code_review_yml_absent", "claude-code-review.yml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := os.Stat(filepath.Join(root, ".github", "workflows", tc.file)); err == nil {
				t.Fatalf(".github/workflows/%s exists", tc.file)
			}
		})
	}
	t.Run("sweep_is_nonempty", func(t *testing.T) {
		if len(files) < 18 {
			t.Fatalf("the census swept %d workflow files, want at least 18 (21 before the retirement, 19 after)", len(files))
		}
	})
}

// AC-CCR-015 (REQ-CCR-015) — .coderabbit.yaml stays, with automatic review and automatic chat reply off, and
// nothing under .github/workflows or scripts/ci reads CodeRabbit.
//
// Mutant: deleting the file, leaving enabled: true, or reading the CodeRabbit status in a workflow turns a row red.
func TestCodeRabbitDisabledAndUnread(t *testing.T) {
	t.Parallel()
	root := findProjectRootForMirrorTest(t)
	var cfg struct {
		Reviews struct {
			AutoReview struct {
				Enabled *bool `yaml:"enabled"`
			} `yaml:"auto_review"`
		} `yaml:"reviews"`
		Chat struct {
			AutoReply *bool `yaml:"auto_reply"`
		} `yaml:"chat"`
	}
	raw, readErr := os.ReadFile(filepath.Join(root, ".coderabbit.yaml"))
	if readErr == nil {
		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			t.Fatalf("parse .coderabbit.yaml: %v", err)
		}
	}
	t.Run("config_present", func(t *testing.T) {
		if readErr != nil {
			t.Fatalf(".coderabbit.yaml is not readable: %v (the file is disabled, never deleted)", readErr)
		}
	})
	show := func(b *bool) string {
		if b == nil {
			return "<unset>"
		}
		return fmt.Sprint(*b)
	}
	t.Run("auto_review_false", func(t *testing.T) {
		if cfg.Reviews.AutoReview.Enabled == nil || *cfg.Reviews.AutoReview.Enabled {
			t.Fatalf("reviews.auto_review.enabled = %s, want false", show(cfg.Reviews.AutoReview.Enabled))
		}
	})
	t.Run("auto_reply_false", func(t *testing.T) {
		if cfg.Chat.AutoReply == nil || *cfg.Chat.AutoReply {
			t.Fatalf("chat.auto_reply = %s, want false", show(cfg.Chat.AutoReply))
		}
	})
	sweep := func(t *testing.T, dir string, match func(string) bool) int {
		t.Helper()
		n := 0
		err := filepath.Walk(filepath.Join(root, filepath.FromSlash(dir)), func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if !match(p) {
				return nil
			}
			b, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			n++
			if strings.Contains(strings.ToLower(string(b)), "coderabbit") {
				t.Errorf("%s mentions CodeRabbit", filepath.ToSlash(strings.TrimPrefix(p, root+string(filepath.Separator))))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	var wfSwept, scriptsSwept int
	t.Run("no_coderabbit_in_workflows", func(t *testing.T) {
		wfSwept = sweep(t, ".github/workflows", func(p string) bool {
			return strings.HasSuffix(p, ".yml") || strings.HasSuffix(p, ".yaml")
		})
	})
	t.Run("no_coderabbit_in_scripts_ci", func(t *testing.T) {
		scriptsSwept = sweep(t, "scripts/ci", func(string) bool { return true })
	})
	t.Run("sweep_is_nonempty", func(t *testing.T) {
		if wfSwept == 0 || scriptsSwept == 0 {
			t.Fatalf("swept %d workflow files and %d scripts/ci files; both must be non-empty", wfSwept, scriptsSwept)
		}
	})
}
