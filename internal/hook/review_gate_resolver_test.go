package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReviewGateResolverRegularExecutablePrecedence(t *testing.T) {
	requireBash(t)
	repo := filepath.Join(mustGetwd(t), "..", "..")
	for _, gate := range []string{"codex", "multi"} {
		for _, mirror := range []string{"", "internal/template/templates"} {
			for _, tc := range []struct{ name, dev, path, home, old, want string }{
				{"dev-first", "file", "file", "file", "", "dev"},
				{"path-second", "dir", "file", "file", "dir", "path"},
				{"home-third", "dir", "dir", "file", "dir", "home"},
				{"all-directories", "dir", "dir", "dir", "dir", ""},
				{"old-path-ignored", "", "file", "file", "file", "path"},
				{"non-executable", "noexec", "", "file", "", "home"},
				{"unavailable", "", "", "", "", ""},
			} {
				t.Run(gate+"/"+mirror+"/"+tc.name, func(t *testing.T) {
					base := t.TempDir()
					project := filepath.Join(base, "workspace", "project")
					home := filepath.Join(base, "home")
					pathBin := filepath.Join(base, "path")
					cfg := filepath.Join(project, ".moai", "config", "sections")
					for _, dir := range []string{cfg, home, pathBin} {
						if err := os.MkdirAll(dir, 0755); err != nil {
							t.Fatal(err)
						}
					}
					if err := os.WriteFile(filepath.Join(cfg, "workflow.yaml"), []byte(reviewGateWorkflowYAML(gate, true)), 0644); err != nil {
						t.Fatal(err)
					}
					for _, candidate := range []struct{ path, kind, label string }{{filepath.Join(project, "bin", "moai"), tc.dev, "dev"}, {filepath.Join(pathBin, "moai"), tc.path, "path"}, {filepath.Join(home, "go", "bin", "moai"), tc.home, "home"}, {filepath.Join(base, "moai"), tc.old, "old"}} {
						if candidate.kind == "" {
							continue
						}
						if err := os.MkdirAll(filepath.Dir(candidate.path), 0755); err != nil {
							t.Fatal(err)
						}
						if candidate.kind == "dir" {
							if err := os.Mkdir(candidate.path, 0755); err != nil {
								t.Fatal(err)
							}
							continue
						}
						mode := os.FileMode(0755)
						if candidate.kind == "noexec" {
							mode = 0644
						}
						if err := os.WriteFile(candidate.path, []byte("#!/bin/sh\nprintf '%s' '"+candidate.label+"'\n"), mode); err != nil {
							t.Fatal(err)
						}
					}
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					cmd := exec.CommandContext(ctx, "bash", filepath.Join(repo, mirror, ".claude", "hooks", "moai", "handle-"+gate+"-review-gate.sh"))
					cmd.Dir = project
					cmd.Env = []string{"PATH=" + pathBin + ":/usr/bin:/bin", "HOME=" + home, "CLAUDE_PROJECT_DIR=" + project}
					cmd.Stdin = strings.NewReader(`{}`)
					out, err := cmd.CombinedOutput()
					if err != nil || string(out) != tc.want {
						t.Errorf("resolver exit=%v output=%q want=%q", err, out, tc.want)
					}
				})
			}
		}
	}
}

func TestReviewGateResolverMirrorEquality(t *testing.T) {
	repo := filepath.Join(mustGetwd(t), "..", "..")
	var expected string
	for _, gate := range []string{"codex", "multi"} {
		for _, mirror := range []string{"internal/template/templates", ""} {
			path := filepath.Join(repo, mirror, ".claude", "hooks", "moai", "handle-"+gate+"-review-gate.sh")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, resolver, found := strings.Cut(string(raw), "# Resolve the moai binary")
			if !found {
				t.Fatal("resolver missing")
			}
			resolver, _, found = strings.Cut(resolver, "# Fail-open:")
			if !found {
				t.Fatal("resolver boundary missing")
			}
			if expected == "" {
				expected = resolver
			}
			if resolver != expected {
				t.Errorf("resolver differs: %s", path)
			}
		}
	}
}
