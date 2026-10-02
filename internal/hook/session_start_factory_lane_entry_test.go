package hook

// session_start_factory_lane_entry_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M4
// (AC-010, REQ-009): the factory leader SessionStart notice states how to
// start a lane in one place and carries no number — no lane count, no
// per-lane launch line, no numbered lane label, no free-slot list — in each of
// the four locales; the stale-run rebind hint names `moai cc -l`.

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// laneEntryLocales is the complete conversation-language set the notice is
// authored in.
var laneEntryLocales = []string{langEnglish, "ko", "ja", "zh"}

// laneEntryCommands are the three lane-start commands the notice names.
var laneEntryCommands = []string{"moai cc -l", "moai glm -l", "moai codex -l"}

// removedFreeSlotLines are the free-slot line prefixes the notice used to
// carry, one per locale (the string-table fields are gone; the text must not
// survive anywhere in the rendered notice).
var removedFreeSlotLines = []string{
	"Free lane slots",
	"현재 빈 레인 슬롯",
	"現在の空きレーンスロット",
	"当前空闲泳道",
}

// numberedLaneLabel matches a concrete or placeholder numbered lane label.
var numberedLaneLabel = regexp.MustCompile(`lane-[0-9<]`)

// launcherEntryFlag extracts the flag a `moai <entry> <flag>` mention carries.
var launcherEntryFlag = regexp.MustCompile("moai (?:cc|glm|codex) (-[-A-Za-z]+)")

// renderLeaderNotice renders the factory leader notice for a declared lane
// count through the real bootstrap builder, with every factory/kanban
// environment axis fixed first.
func renderLeaderNotice(t *testing.T, lang string, lanes int) string {
	t.Helper()
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	t.Setenv(config.EnvMoaiFactoryWorkers, strconv.Itoa(lanes))
	t.Setenv(config.EnvMoaiKanbanID, "abc123")
	return factoryBootstrapNotice("", "", lang)
}

// TestFactoryLeadNoticePrintsLaneCommandOnce pins the lane-start sentence: the
// three `-l` commands appear exactly once each in every locale, and nothing in
// the notice teaches a removed entry form or states a number.
func TestFactoryLeadNoticePrintsLaneCommandOnce(t *testing.T) {
	for _, lang := range laneEntryLocales {
		t.Run(lang, func(t *testing.T) {
			notice := renderLeaderNotice(t, lang, 3)
			if notice == "" {
				t.Fatal("leader notice rendered empty")
			}
			for _, cmd := range laneEntryCommands {
				if got := strings.Count(notice, cmd); got != 1 {
					t.Errorf("%q appears %d times, want exactly 1:\n%s", cmd, got, notice)
				}
			}
			for _, removed := range []string{"-f lane", "lane-<n>", "lane-1..lane-"} {
				if strings.Contains(notice, removed) {
					t.Errorf("notice still teaches the removed form %q:\n%s", removed, notice)
				}
			}
			if numberedLaneLabel.MatchString(notice) {
				t.Errorf("notice carries a numbered lane label:\n%s", notice)
			}
			for _, line := range strings.Split(notice, "\n") {
				if strings.HasPrefix(line, "moai ") {
					t.Errorf("notice carries a per-lane launch line %q", line)
				}
			}
			for _, gone := range removedFreeSlotLines {
				if strings.Contains(notice, gone) {
					t.Errorf("notice still carries the free-slot line %q:\n%s", gone, notice)
				}
			}
			// The entry guide names -l and -f only: every launcher flag the
			// notice mentions is one of the two.
			for _, m := range launcherEntryFlag.FindAllStringSubmatch(notice, -1) {
				if m[1] != "-l" && m[1] != "-f" {
					t.Errorf("notice names launcher flag %q beside -l and -f:\n%s", m[1], notice)
				}
			}
		})
	}
}

// TestFactoryLeadNoticeIsLaneCountIndependent pins that the lane guidance is
// the same bytes at declared lane counts 1, 3, and 8 in every locale, and for
// every launch provenance — a notice that prints the count (or a per-lane
// line) only above one lane cannot satisfy it.
func TestFactoryLeadNoticeIsLaneCountIndependent(t *testing.T) {
	for _, lang := range laneEntryLocales {
		t.Run(lang, func(t *testing.T) {
			want := renderLeaderNotice(t, lang, 1)
			if want == "" {
				t.Fatal("leader notice rendered empty at one lane")
			}
			for _, lanes := range []int{3, 8} {
				if got := renderLeaderNotice(t, lang, lanes); got != want {
					t.Errorf("notice at %d lanes differs from the notice at 1 lane:\n--- 1 ---\n%s\n--- %d ---\n%s", lanes, want, lanes, got)
				}
			}
			for _, provider := range []string{"glm", "gpt", "claude", "other"} {
				renderLeaderNotice(t, lang, 1) // fixes every other axis again
				t.Setenv(config.EnvMoaiLaunchProvider, provider)
				if got := factoryBootstrapNotice("", "", lang); got != want {
					t.Errorf("notice under launch provider %q differs from the default:\n%s", provider, got)
				}
			}
		})
	}
}

// TestStaleRunRebindHintNamesLaneEntry pins the stale-run rebind hint: it
// names `moai cc -l` (the next free slot) and no removed form, in four
// locales.
func TestStaleRunRebindHintNamesLaneEntry(t *testing.T) {
	for _, lang := range laneEntryLocales {
		hint := staleRunMessagesFor(lang).laneLabelUnbindRebind
		if !strings.Contains(hint, "moai cc -l") {
			t.Errorf("%s rebind hint does not name 'moai cc -l': %s", lang, hint)
		}
		for _, removed := range []string{"-f lane", "lane-<n>"} {
			if strings.Contains(hint, removed) {
				t.Errorf("%s rebind hint still teaches %q: %s", lang, removed, hint)
			}
		}
	}
}
