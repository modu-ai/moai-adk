package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestLauncherShortName_AmbiguityStopsBackend(t *testing.T) {
	for _, backend := range []string{"cc", "glm"} {
		t.Run(backend, func(t *testing.T) {
			w := m3NewWorld(t)
			got := m2LaunchFixture(t, w.repo)
			t.Setenv(config.DefaultGLMEnvVar, "fixture-key")
			cmd := ccCmd
			if backend == "glm" {
				cmd = glmCmd
			}
			err := cmd.RunE(cmd, []string{"-w", "t6"})
			if err == nil {
				t.Fatal("ambiguous legacy name launched instead of refusing")
			}
			for _, text := range []string{w.moaiT6, w.claudeT6, "absolute path"} {
				if !strings.Contains(err.Error(), text) {
					t.Errorf("ambiguity must identify candidates and selection remedy %q: %v", text, err)
				}
			}
			if *got != nil {
				t.Fatalf("backend was called with %v after ambiguous selection", *got)
			}
		})
	}
}

func TestFactoryLandingWithoutBaseIgnoresRelativeCardDirectory(t *testing.T) {
	root, _ := fcFixture(t)
	t.Chdir(t.TempDir())
	if err := os.Mkdir("t1", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := factoryRefuseForeignWorktree(context.Background(), root, "t1"); err != nil {
		t.Fatalf("missing resolver base must not treat cwd/t1 as a candidate: %v", err)
	}
	legacy := filepath.Join(root, sessionWorktreeSubdir, "t1")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := factoryRefuseForeignWorktree(context.Background(), root, "t1"); err == nil || !strings.Contains(err.Error(), legacy) {
		t.Fatalf("legacy foreign tree must still be refused: %v", err)
	}
}

func TestAutoOwnerAliveUsesRecordedCustomTree(t *testing.T) {
	for _, channel := range []string{"registry", "process"} {
		t.Run(channel, func(t *testing.T) {
			root, _ := fcFixture(t)
			custom := filepath.Join(t.TempDir(), "operator-tree")
			if err := os.MkdirAll(filepath.Join(custom, "subdir"), 0o755); err != nil {
				t.Fatal(err)
			}
			fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked, WorktreePath: custom})
			lv := autoLiveness{
				registryEntries: func() ([]autoRegistryEntry, error) {
					if channel == "registry" {
						return []autoRegistryEntry{{Cwd: filepath.Join(custom, "subdir"), PID: 123}}, nil
					}
					return nil, nil
				},
				processCWDs: func() ([]string, error) {
					if channel == "process" {
						return []string{custom}, nil
					}
					return nil, nil
				},
				pidAlive: func(int) bool { return true },
			}
			alive, _, err := lv.ownerAlive(root, "t1")
			if err != nil || !alive {
				t.Fatalf("live %s owner in recorded custom path was missed: alive=%v err=%v", channel, alive, err)
			}
			alive, _, err = lv.ownerAlive(root, "t2")
			if err != nil || alive {
				t.Fatalf("another card must not inherit t1's owner: alive=%v err=%v", alive, err)
			}
		})
	}
}
