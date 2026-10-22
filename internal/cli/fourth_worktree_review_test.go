package cli

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/testguard"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

func TestCurrentReviewLandingFailurePreservesUnleasedCard(t *testing.T) {
	for _, route := range []string{"selector", "nominated"} {
		t.Run(route, func(t *testing.T) {
			testguard.RedirectHome(t)
			root, store := fcFixture(t)
			fcQueue(t, store, factory.BacklogStatePicked)
			fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked, OwnerLabel: "lane-1"})
			base, err := worktreeroot.Root(root)
			if err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(base, ".moai-project.json")
			if err := os.WriteFile(marker, []byte("{"), 0o600); err != nil {
				t.Fatal(err)
			}
			before := fcCard(t, root, "t1")
			if route == "selector" {
				_, leased, callErr := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
				if callErr == nil || leased {
					t.Errorf("invalid landing owner: leased=%v error=%v", leased, callErr)
				} else if !strings.Contains(callErr.Error(), marker) {
					t.Errorf("refusal lost the owner-marker failure: %v", callErr)
				}
			} else {
				_, callErr := factoryNextNominate(context.Background(), root, fcRun, "lane-1", "t1", "")
				if callErr == nil || !strings.Contains(callErr.Error(), marker) {
					t.Errorf("nominated landing error = %v, want the owner-marker refusal", callErr)
				}
			}
			if after := fcCard(t, root, "t1"); !reflect.DeepEqual(after, before) {
				t.Errorf("landing failure changed card: before=%+v after=%+v", before, after)
			}
			if got := fcQueueItemState(t, store, "t1"); got != factory.BacklogStatePicked {
				t.Errorf("landing failure changed queue state to %s", got)
			}
		})
	}
}

func TestCurrentReviewCodexShortNamePreservesMetadataErrors(t *testing.T) {
	for _, mode := range []string{"corrupt-owner", "unreadable-key", "registered-legacy", "plain-project-legacy"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv(paths.EnvHome, m2Real(t, t.TempDir()))
			for _, key := range []string{"GIT_DIR", "GIT_COMMON_DIR", "GIT_WORK_TREE"} {
				t.Setenv(key, "")
				if err := os.Unsetenv(key); err != nil {
					t.Fatal(err)
				}
			}
			var root string
			if mode == "plain-project-legacy" {
				root = m2Real(t, t.TempDir())
			} else {
				root = m2Repo(t)
			}
			// The plain-project positive is an isolated discovery domain. Cloud
			// may expose protected synthetic metadata above this scratch parent;
			// Git and the metadata-absence proof must stop at the same boundary.
			t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(root))
			legacy := filepath.Join(root, sessionWorktreeSubdir, "card")
			if mode == "plain-project-legacy" {
				if err := os.MkdirAll(legacy, 0o755); err != nil {
					t.Fatal(err)
				}
			} else {
				m3Gitf(t, root, "worktree", "add", "-q", "--detach", legacy, "HEAD")
				base, err := worktreeroot.Root(root)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "corrupt-owner" {
					if err := os.WriteFile(filepath.Join(base, ".moai-project.json"), []byte("{"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "unreadable-key" {
					record := filepath.Join(root, ".git", "moai-project-key")
					if err := os.Remove(record); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(record, 0o755); err != nil {
						t.Fatal(err)
					}
				}
			}
			got, err := resolveCodexWorktreeDir(root, "card")
			if mode == "registered-legacy" || mode == "plain-project-legacy" {
				if err != nil || !worktreeroot.Same(got, legacy) {
					t.Fatalf("normal legacy resolution = %q, %v; want %s", got, err, legacy)
				}
			} else if err == nil || got != "" {
				t.Fatalf("metadata failure fell back to %q, %v", got, err)
			}
		})
	}
}
