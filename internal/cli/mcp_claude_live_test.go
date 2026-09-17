package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func snapshotClaudeAuditFixture(t *testing.T, root string) map[string][sha256.Size]byte {
	t.Helper()
	snapshot := map[string][sha256.Size]byte{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		snapshot[rel] = sha256.Sum256(data)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot fixture: %v", err)
	}
	return snapshot
}

func TestLiveClaudeAudit_GPTAndGLMOriginsUseSubscriptionReadOnly_AC_CLA_003_005_006_009_010_015(t *testing.T) {
	if os.Getenv("MOAI_LIVE_CLAUDE_AUDIT") != "1" {
		t.Skip("set MOAI_LIVE_CLAUDE_AUDIT=1 for the explicit subscription acceptance")
	}

	root := newClaudeReviewTree(t, true)
	before := snapshotClaudeAuditFixture(t, root)
	t.Setenv("ANTHROPIC_BASE_URL", "http://127.0.0.1:1")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "live-gateway-token-sentinel")
	t.Setenv("ANTHROPIC_MODEL", "gpt-6-astra")
	t.Setenv("Z_AI_API_KEY", "live-glm-token-sentinel")
	t.Setenv(config.EnvMoaiLaunchProvider, "gpt")
	t.Setenv(claudeNestedSessionEnv, "1")
	t.Setenv("CLAUDE_CODE_FUTURE_ROUTING", "parent-session")

	for _, origin := range []string{"gpt", "glm"} {
		t.Run(origin, func(t *testing.T) {
			result := runMultiAudit(
				context.Background(),
				ReviewOutput{Verdict: "fail", Summary: "caller-anchor-must-be-ignored", Findings: []Finding{}, NextSteps: []string{}},
				codexTargetUncommitted,
				"correctness and command-injection resistance",
				MultiAuditConfig{
					Gates:          config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateOff, GLM: config.AuditGateOff},
					ProjectRoot:    root,
					OriginProvider: origin,
				},
				nil,
			)

			if len(result.PerBackendVerdicts) != 1 {
				t.Fatalf("per-backend count = %d, want one Claude audit", len(result.PerBackendVerdicts))
			}
			claude := result.PerBackendVerdicts[0]
			if claude.Backend != BackendClaude || claude.Source != "mcp_claude_audit" {
				t.Fatalf("Claude source = %+v", claude)
			}
			if claude.Provenance == nil || claude.Provenance.AuthMode != "subscription" || claude.Provenance.Transport != claudeAuditTransport {
				t.Fatalf("live Claude provenance = %+v", claude.Provenance)
			}
			if claude.Verdict == VerdictInconclusive {
				if claude.Provenance.ErrorCode != claudeErrCapacityUnavailable {
					t.Fatalf("live Claude inconclusive = %+v", claude.Provenance)
				}
				t.Logf("%s origin reached verified Claude subscription transport; provider capacity was unavailable", origin)
			} else {
				if claude.Verdict != "pass" && claude.Verdict != "fail" {
					t.Fatalf("live Claude verdict = %q, summary=%q provenance=%+v", claude.Verdict, claude.Summary, claude.Provenance)
				}
				if !strings.HasPrefix(strings.ToLower(claude.Provenance.ResolvedModel), "claude-") {
					t.Fatalf("resolved model = %q, want Claude family", claude.Provenance.ResolvedModel)
				}
			}
			if claude.Provenance.ToolSurface != "none" || claude.Provenance.SessionPersisted {
				t.Fatalf("live isolation provenance = %+v", claude.Provenance)
			}
			encoded, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			for _, forbidden := range []string{"caller-anchor-must-be-ignored", "live-gateway-token-sentinel", "live-glm-token-sentinel"} {
				if strings.Contains(string(encoded), forbidden) {
					t.Fatalf("forbidden caller/provider data %q leaked into result", forbidden)
				}
			}
		})
	}

	after := snapshotClaudeAuditFixture(t, root)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("fixture files changed during read-only Claude audits: before=%v after=%v", before, after)
	}
}
