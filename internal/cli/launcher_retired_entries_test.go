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
	"sort"
	"strings"
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
