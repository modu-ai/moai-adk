package cli

// launcher_retired_entries_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M5a (card
// t1399), AC-011 and AC-013. `-k` / `--kanban` is a retired entry on cc, glm
// and codex: every shape is refused with one line before any branch resolves,
// nothing is launched and nothing is written. The Codex lane child no longer
// carries the retired lane-label marker and still launches and identifies
// itself through MOAI_FACTORY_WORKER.
//
// Every test sets or clears every env axis it reads (netScrubLaneEnv through
// the lane fixtures) and runs through the launch seams, so no session starts
// and no real home state is written.

import (
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// retiredLaneLabelMarker is the spelling of the Codex lane-label marker that
// M5a retired from the child environment. It is written out here, not read
// from internal/config, so the assertion survives the constant's deletion.
const retiredLaneLabelMarker = "MOAI_KANBAN_LABEL"

// retiredSpecMarker is the SPEC marker no launcher publishes any more, written
// out for the same reason.
const retiredSpecMarker = "MOAI_KANBAN_SPEC"

// retiredLeaderMarker is the kanban-leader signal, deleted from internal/config
// at M5b with its last reader. It is written out for the same reason, so the
// tests that assert a factory session never carries it, and the ambient
// scrub that gives them a clean slate, keep the name after the constant is gone.
const retiredLeaderMarker = "MOAI_KANBAN"

// retiredEntryShapes are AC-011's seven `-k` shapes.
var retiredEntryShapes = [][]string{
	{"-k"},
	{"-k", "SPEC-X-001"},
	{"-k", "3"},
	{"-k", "--name", "plan"},
	{"-k", "--name", "lane-1"},
	{"--kanban"},
	{"-k=3"},
}

// TestKanbanEntryRefused — AC-011: every `-k` shape on cc, glm and codex is
// refused with one line stating the mode is retired and naming `-f` (lead) and
// `-l` (join as a lane); on codex the line also names `moai codex -l` and
// `moai cc -f` / `moai glm -f`. A `-k` after the `--` marker is forwarded.
func TestKanbanEntryRefused(t *testing.T) {
	for _, verb := range laneVerbs {
		for _, shape := range retiredEntryShapes {
			t.Run(verb.name+"_"+strings.Join(shape, "_"), func(t *testing.T) {
				err := m2Refusal(t, verb.backend, verb.entry, shape)
				if err == nil {
					return
				}
				requireTokens(t, shape, err.Error(), "retired", "-f", "-l")
			})
		}
	}
	for _, shape := range retiredEntryShapes {
		t.Run("codex_"+strings.Join(shape, "_"), func(t *testing.T) {
			line := m3CodexRefusal(t, shape)
			requireTokens(t, shape, line, "retired", "-f", "-l", "moai codex -l", "moai cc -f", "moai glm -f")
		})
	}

	t.Run("passthrough", func(t *testing.T) {
		args := []string{"--", "-k", "3", "--kanban"}
		entry, err := parseLauncherEntry(args)
		if err != nil {
			t.Fatalf("parseLauncherEntry(%v) = %v, want the tokens past -- forwarded untouched", args, err)
		}
		if entry.FactoryEnabled || !slices.Equal(entry.Rest, args) {
			t.Errorf("parseLauncherEntry(%v): factory=%v rest=%v, want no entry and the args verbatim", args, entry.FactoryEnabled, entry.Rest)
		}
	})
}

// laneKeyNames returns the sorted MOAI_KANBAN* / MOAI_FACTORY_* names an
// environment carries with a non-empty value.
func laneKeyNames(env map[string]string) []string {
	var names []string
	for name, value := range env {
		if value == "" {
			continue
		}
		if strings.HasPrefix(name, "MOAI_KANBAN") || strings.HasPrefix(name, "MOAI_FACTORY_") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// TestCodexLaneChildEnvOmitsLabelMarker — AC-013, first half: the per-card
// child of the Codex relaunch loop carries the lane role, the lane label under
// MOAI_FACTORY_WORKER, the backend, the card id and the dispatch mode the
// launcher's own lane stamp exports, and no other lane-family key — in
// particular no MOAI_KANBAN_LABEL, under that name or any other.
func TestCodexLaneChildEnvOmitsLabelMarker(t *testing.T) {
	lane := netCodexLaneChildFor(t, "-l")
	if _, present := lane.env[retiredLaneLabelMarker]; present {
		t.Errorf("the Codex lane child carries %s=%q; the label marker is retired", retiredLaneLabelMarker, lane.env[retiredLaneLabelMarker])
	}
	want := []string{
		config.EnvFactoryAutoDispatch,
		config.EnvFactoryRole,
		config.EnvMoaiFactoryWorker,
		config.EnvFactoryBackend,
		config.EnvFactoryCard,
	}
	sort.Strings(want)
	if got := laneKeyNames(lane.env); !slices.Equal(got, want) {
		t.Errorf("Codex lane child lane-family keys = %v, want exactly %v", got, want)
	}
	if lane.env[config.EnvFactoryRole] != config.FactoryRoleLane || lane.env[config.EnvMoaiFactoryWorker] == "" {
		t.Errorf("the child does not identify itself as a lane: role=%q worker=%q", lane.env[config.EnvFactoryRole], lane.env[config.EnvMoaiFactoryWorker])
	}
}

// adoptChildLaneEnv makes this process's lane-family environment exactly the
// child's: every key the launcher stamped or the child carries is cleared,
// then the child's values are set. The returned function puts each key back to
// its prior presence.
func adoptChildLaneEnv(child map[string]string) func() {
	keys := map[string]bool{}
	for _, key := range codexLaneLaunchEnvKeys {
		keys[key] = true
	}
	for key := range child {
		if strings.HasPrefix(key, "MOAI_KANBAN") || strings.HasPrefix(key, "MOAI_FACTORY_") {
			keys[key] = true
		}
	}
	var restores []func()
	for key := range keys {
		restores = append(restores, captureEnvState(key))
		_ = os.Unsetenv(key)
	}
	for key := range keys {
		if value, ok := child[key]; ok {
			_ = os.Setenv(key, value)
		}
	}
	return func() {
		for _, restore := range restores {
			restore()
		}
	}
}

// TestFactoryCardVerbsResolveLaneFromWorkerMarker — AC-013, second half: with
// the process environment set to exactly the Codex lane child's, the factory
// card verbs resolve the lane label and admission from MOAI_FACTORY_WORKER and
// MOAI_FACTORY_ROLE alone and the card reaches merge-ready. A launcher that
// stopped publishing the worker marker along with the label fails here.
func TestFactoryCardVerbsResolveLaneFromWorkerMarker(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	t.Chdir(root)
	netScrubLaneEnv(t)

	sessions := 0
	prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
	codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		sessions++
		child := sdEnvOf(t, c.Env)
		restore := adoptChildLaneEnv(child)
		defer restore()

		if !factoryLaneAdmission() {
			t.Errorf("lane admission is false in the child environment (role=%q)", child[config.EnvFactoryRole])
		}
		label, err := factoryLaneLabelFromEnv("stage")
		if err != nil || label != child[config.EnvMoaiFactoryWorker] {
			t.Errorf("lane label from the child environment = (%q, %v), want %q", label, err, child[config.EnvMoaiFactoryWorker])
		}
		sdCodexSessionWork(t, root, child[config.EnvFactoryCard])
		return nil
	}
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })

	if _, _, err := runCodexCmd(t, "-l"); err != nil {
		t.Fatalf("codex lane: %v", err)
	}
	if sessions != 1 {
		t.Fatalf("the substituted Codex session ran %d times, want 1", sessions)
	}
	if card := fcCard(t, root, "t1"); card.State != homestate.CardMergeReady {
		t.Errorf("t1 ended at %s, want merge-ready (the card verbs ran from the child environment alone)", card.State)
	}
}
