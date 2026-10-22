package hook

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewLifecycleRegistryFailuresWarnWithoutBlocking(t *testing.T) {
	for _, boundary := range []string{"launcher", "create", "remove"} {
		t.Run(boundary, func(t *testing.T) {
			root := t.TempDir()
			markMoaiProjectRoot(t, root)
			t.Setenv("CLAUDE_PROJECT_DIR", root)
			state := worktreeStateFile(root)
			if err := os.MkdirAll(filepath.Dir(state), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(state, []byte("broken registry"), 0600); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			old := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&out, nil)))
			t.Cleanup(func() { slog.SetDefault(old) })
			switch boundary {
			case "launcher":
				RegisterLauncherWorktree(root, "/tree", "WT-tree", "test")
			case "create":
				(&worktreeCreateHandler{}).registerEntry(&HookInput{CWD: root}, "/tree", "WT-tree")
			case "remove":
				if _, err := (&worktreeRemoveHandler{}).Handle(context.Background(), &HookInput{CWD: root, WorktreePath: "/tree"}); err != nil {
					t.Fatalf("lifecycle becomes blocking: %v", err)
				}
			}
			if !strings.Contains(out.String(), "level=WARN") || !strings.Contains(out.String(), "error=") {
				t.Fatalf("%s silently discarded registry failure: %s", boundary, out.String())
			}
			raw, err := os.ReadFile(state)
			if err != nil || string(raw) != "broken registry" {
				t.Fatalf("warning boundary replaced invalid state: %q %v", raw, err)
			}
		})
	}
}
