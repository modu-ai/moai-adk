package hook

// stale_run_relaunch_test.go — SPEC-FACTORY-STALE-RUN-HEAL-001 M1: the legacy
// stale-run and unbind notices name the way back as executable `moai factory
// relaunch` lines (REQ-SRH-001, REQ-SRH-002, REQ-SRH-016), per the table rows
// R6-R9 of spec.md §D.7.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

const relaunchVerbPrefix = "moai factory relaunch"

// pinFactoryGateBudget widens the gate's measurement budget for one test: the
// production 200 ms budget made these assertions depend on machine load (a
// loaded host answered "context deadline exceeded" for a healthy database).
func pinFactoryGateBudget(t *testing.T) {
	t.Helper()
	prev := factoryGateBudget
	factoryGateBudget = 30 * time.Second
	t.Cleanup(func() { factoryGateBudget = prev })
}

// relaunchLines returns the command lines of a notice: every line that starts
// with the verb, exactly as printed (no trimming — a command is its own line).
func relaunchLines(notice string) []string {
	var out []string
	for _, line := range strings.Split(notice, "\n") {
		if strings.HasPrefix(line, relaunchVerbPrefix) {
			out = append(out, line)
		}
	}
	return out
}

func TestStaleNoticeCarriesExecutableRelaunch(t *testing.T) { // AC-SRH-005 (R6-R9)
	const verb = relaunchVerbPrefix + " --provider "
	rows := []struct {
		name string
		seed func(t *testing.T, root string)
		want func(provider string) []string
	}{
		{"R6 run active", func(t *testing.T, root string) { recordActiveFactoryRun(t, root, "runX") },
			func(p string) []string { return []string{verb + p + " --from-run runX"} }},
		{"R7 retired none active", func(t *testing.T, root string) { recordFactoryRunWithStatus(t, root, "runX", "retired") },
			func(string) []string { return nil }},
		{"R8 retired one active", func(t *testing.T, root string) {
			recordFactoryRunWithStatus(t, root, "runX", "retired")
			recordActiveFactoryRun(t, root, "runY")
		}, func(p string) []string { return []string{verb + p} }},
		{"R9 retired two active", func(t *testing.T, root string) {
			recordFactoryRunWithStatus(t, root, "runX", "retired")
			recordActiveFactoryRun(t, root, "runY")
			recordActiveFactoryRun(t, root, "runZ")
		}, func(p string) []string {
			if p == "codex" {
				return []string{verb + p}
			}
			return []string{verb + p + " --run runY", verb + p + " --run runZ"}
		}},
	}
	backends := []struct{ backend, provider string }{
		{"claude", "cc"}, {"glm", "glm"}, {"gpt", "codex"}, {"", "cc"},
	}
	for _, row := range rows {
		for _, b := range backends {
			t.Run(row.name+" backend="+b.backend, func(t *testing.T) {
				root := t.TempDir()
				row.seed(t, root)
				srlGateEnv(t, "runX", "worker-69")
				t.Setenv(config.EnvMoaiKanbanBackend, b.backend)
				want := row.want(b.provider)

				// Every surface and locale of the notice: the agent-facing peer
				// registration surface (English) and the SessionStart
				// bootstrap surface in all four locales.
				surfaces := map[string]string{
					"peer/en": registerFactoryHookPeer(context.Background(), &HookInput{SessionID: "relaunch-peer", ProjectDir: root}, factoryPeerBindUserPrompt),
				}
				for _, lang := range []string{"en", "ko", "ja", "zh"} {
					surfaces["bootstrap/"+lang] = factoryBootstrapNotice(root, "relaunch-boot-"+lang, lang)
					// The run-id branch of the stale-run message (its own locale
					// field) always prints the R6 prescription line.
					if strings.HasPrefix(row.name, "R6") {
						surfaces["staleRunNotice/"+lang] = staleRunNotice("worker-69", lang)
					}
				}
				for surface, notice := range surfaces {
					if strings.ContainsAny(notice, "<>") {
						t.Errorf("%s: notice carries an angle-bracket placeholder:\n%s", surface, notice)
					}
					if strings.Contains(notice, "runs --retire") {
						t.Errorf("%s: notice still prescribes the retire-then-relaunch prose:\n%s", surface, notice)
					}
					got := relaunchLines(notice)
					if len(got) != len(want) {
						t.Errorf("%s: command lines = %q, want %q\n%s", surface, got, want, notice)
						continue
					}
					for i := range want {
						if got[i] != want[i] {
							t.Errorf("%s: command line %d = %q, want %q", surface, i, got[i], want[i])
						}
					}
				}
				// REQ-SRH-001: the line is the operator's to run from a terminal.
				if peer := surfaces["peer/en"]; len(want) > 0 && !strings.Contains(peer, "operator") {
					t.Errorf("peer notice does not frame the command as the operator's:\n%s", peer)
				}
			})
		}
	}
}
