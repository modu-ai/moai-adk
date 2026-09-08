package hook

import (
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// TestFactoryBootstrapNoticeSilentForOrdinarySession is the blast-radius
// case: a session that is not part of a factory run is completely unaffected.
func TestFactoryBootstrapNoticeSilentForOrdinarySession(t *testing.T) {
	clearKanbanEnv(t)

	if got := factoryBootstrapNotice("", langEnglish); got != "" {
		t.Errorf("ordinary session must get no factory notice, got:\n%s", got)
	}
}

// TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide is the AC bundle
// for the lead notice (t118): N lane launch lines carrying the incremental
// `-f lane-<i>` form, the entry-point guidance naming `moai glm -f <N>` and
// the `-f lane-<n>` form, the per-lane fan-out line, the leader socket
// path, and the run id alongside the session name that must match it.
func TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")
	t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-socket-factory/abc123")

	notice := factoryBootstrapNotice("", langEnglish)
	for _, want := range []string{
		"run abc123",
		// The lead is named by its bare role (t133): the run id lives in the
		// header line above, not in the session name.
		"named lead.",
		"moai cc -f lane-1",
		"moai cc -f lane-2",
		"moai cc -f lane-3",
		"moai glm -f 3",
		"moai cc -f lane-<n>",
		"one-lane default",
		"Every lane can run up to 10 agents concurrently in parallel.",
		"/tmp/moai-socket-factory/abc123",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("lead notice missing %q:\n%s", want, notice)
		}
	}
	// The count must be exact — an off-by-one fan-out misaddresses lanes.
	// The launch lines are NUMBERED; the entry guide's generic
	// `moai cc -f lane-<n>` mention is not a launch line, so the count
	// matches digits only.
	if got := len(factoryLaunchLineRe.FindAllString(notice, -1)); got != 3 {
		t.Errorf("launch line count = %d, want 3:\n%s", got, notice)
	}
}

// factoryLaunchLineRe matches exactly the per-lane launch lines of the
// factory lead notice (`moai cc -f lane-<i>`, i a number) — the entry
// guide's `lane-<n>` placeholder is excluded by the digit class.
var factoryLaunchLineRe = regexp.MustCompile(`moai cc -f lane-[0-9]+`)

// TestFactoryLeadNoticeWorkerCountDrivesLineCount asserts N drives the line
// count directly (the v1 no-upper-bound rule: any N >= 1 prints N lines).
func TestFactoryLeadNoticeWorkerCountDrivesLineCount(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	t.Setenv(config.EnvMoaiKanbanID, "xyz9")

	if got := len(factoryLaunchLineRe.FindAllString(factoryBootstrapNotice("", langEnglish), -1)); got != 1 {
		t.Errorf("N=1 launch line count = %d, want 1", got)
	}
}

// TestFactoryLeadNoticeEmptyWithoutRunID asserts the fail-open shape: a lead
// with no run id (or a nonsensical count) emits nothing rather than a notice
// addressing an unnamed run.
func TestFactoryLeadNoticeEmptyWithoutRunID(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	if got := factoryBootstrapNotice("", langEnglish); got != "" {
		t.Errorf("lead without a run id must emit nothing, got:\n%s", got)
	}
}

// TestFactoryWorkerNoticeNamesLabel asserts the join ack names the label the
// session actually launched under — the reliable surface for a bumped number,
// since the launcher's stderr note is gone by the time the TUI takes the
// screen. The exact-sentence assertion also pins the (label, count) argument
// order of the workerJoin format: the pre-t118 formats carried %d before %s
// while the call passed the label first, rendering %!d(string=lane-4) —
// a Contains("lane-4") assertion passed right through that garbage, so the
// whole sentence is asserted here.
func TestFactoryWorkerNoticeNamesLabel(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorker, "lane-4")
	t.Setenv(config.EnvMoaiFactoryWorkers, "3")

	// Card t224: the notice is the join line PLUS the standing spawn authority
	// (appended, never substituted) — assert the join line as a prefix-presence
	// rather than whole-output equality.
	got := factoryBootstrapNotice("", langEnglish)
	if !strings.HasPrefix(got, "Factory Mode: joined a 3-lane run as lane-4.") {
		t.Errorf("lane notice missing the join line prefix:\n%s", got)
	}
	if !strings.Contains(got, "Standing spawn authority") {
		t.Errorf("lane notice lost the standing spawn authority:\n%s", got)
	}

	// The incremental `-f lane-<n>` form carries no count (workers=0); the
	// count-less sentence must render, not fabricate a fan-out size and not
	// leak a bad verb.
	t.Setenv(config.EnvMoaiFactoryWorkers, "0")
	if got := factoryBootstrapNotice("", langEnglish); !strings.HasPrefix(got,
		"Factory Mode: joined the factory run as lane-4.") {
		t.Errorf("count-less lane notice missing the join line prefix:\n%s", got)
	}

	// A malformed label emits nothing (fail-open, mirroring the companion
	// branch) — no error, no notice.
	t.Setenv(config.EnvMoaiFactoryWorker, "not-a-lane-label")
	if got := factoryBootstrapNotice("", langEnglish); got != "" {
		t.Errorf("malformed lane label must emit nothing, got:\n%s", got)
	}
}

// TestFactoryWorkerNoticeLocaleWordOrders pins the explicit-index contract on
// the two sentence shapes across all four locales: every rendering must name
// the label and the count (or just the label, count-less) with no %!verb
// artifact, whatever order the locale's natural prose puts them in.
func TestFactoryWorkerNoticeLocaleWordOrders(t *testing.T) {
	clearKanbanEnv(t)

	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		got := factoryWorkerNotice("lane-2", 5, lang)
		if !strings.Contains(got, "lane-2") || !strings.Contains(got, "5") ||
			strings.Contains(got, "%!") {
			t.Errorf("locale %q worker join rendered wrong: %q", lang, got)
		}
		gotNoCount := factoryWorkerNotice("lane-2", 0, lang)
		if !strings.Contains(gotNoCount, "lane-2") || strings.Contains(gotNoCount, "%!") {
			t.Errorf("locale %q count-less join rendered wrong: %q", lang, gotNoCount)
		}
	}
}

// TestFactoryBootstrapNoticeStartupOnly asserts the re-entry gating shared
// with the kanban notice: resume / clear / compact / fork re-emit nothing,
// because the operator's lane terminals are already open by then.
func TestFactoryBootstrapNoticeStartupOnly(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	for _, source := range []string{"resume", "clear", "compact", "fork", "upgrade-mystery"} {
		if got := factoryBootstrapNoticeForSource(source, "", langEnglish); got != "" {
			t.Errorf("source %q must not re-announce the bootstrap, got:\n%s", source, got)
		}
	}
	if got := factoryBootstrapNoticeForSource("startup", "", langEnglish); got == "" {
		t.Error("source startup must announce the bootstrap")
	}
	if got := factoryBootstrapNoticeForSource("", "", langEnglish); got == "" {
		t.Error("an empty source is treated as startup and must announce")
	}
}

// TestKanbanNoticeSuppressedUnderFactoryEnv asserts the insurance guard: a
// hand-exported kanban environment on top of a factory session must not stack
// the three-role kanban notice under the factory one.
func TestKanbanNoticeSuppressedUnderFactoryEnv(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")
	t.Setenv(config.EnvMoaiFactoryWorkers, "2")

	if got := kanbanBootstrapNotice("", langEnglish); got != "" {
		t.Errorf("kanban notice must yield to the factory notice, got:\n%s", got)
	}
}

// TestFactoryMessagesLocaleFallback asserts the locale table resolves the
// four conversation languages and falls back to English for anything else.
func TestFactoryMessagesLocaleFallback(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		if factoryMessagesFor(lang).leadHeader == "" {
			t.Errorf("locale %q resolved to an empty message set", lang)
		}
	}
	if factoryMessagesFor("fr").leadHeader != factoryMessagesFor(langEnglish).leadHeader {
		t.Error("an unknown locale must fall back to English")
	}
}

// TestFactoryLeadNoticeCarriesDispatchDiscipline is the t85 codification AC
// (v1.2.0, reworded for the t118 final design): the lead notice carries the
// FACTORY-specific dispatch discipline — whole-card routing (every card to
// ONE lane, which runs the serial plan -> run -> sync path in-session), the
// foreman handoff line, the fan-out-only stagger rule (workflow auto-stagger
// explicitly excluded), and the no-model-override rule — plus the live
// free-slot line.
// It deliberately does NOT teach queue polling: that loop is the kanban
// foreman's (t96), and a second polling protocol here would conflict.
func TestFactoryLeadNoticeCarriesDispatchDiscipline(t *testing.T) {
	clearKanbanEnv(t)

	root := t.TempDir()
	// lane-2 claimed by THIS test process — a pid that is genuinely alive,
	// so slot 2 reads busy without a probe seam.
	if err := kanban.SaveFactoryRegistry(kanban.FactoryRegistryPath(root), map[string]kanban.FactoryWorkerEntry{
		"lane-2": kanban.NewFactoryWorkerEntry(),
	}); err != nil {
		t.Fatalf("seed registry: %v", err)
	}

	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	notice := factoryBootstrapNotice(root, langEnglish)
	for _, want := range []string{
		"every card is routed WHOLE to one lane",
		"serial 3-stage path",
		"plan -> run -> sync",
		"kanban foreman loop (bare `/loop`)",
		"started producing output",
		"cache-aware-execution directive 2",
		"FACTORY fan-out only",
		"CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS",
		"Free lane slots right now: lane-1, lane-3.",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("lead notice missing dispatch-discipline token %q:\n%s", want, notice)
		}
	}
	// The t96-absorbed content must NOT come back: no queue-polling protocol.
	// The no-model-override discipline is likewise retired from the notice
	// (operator request): the rule itself lives in the factory skill docs.
	for _, gone := range []string{"moai todo list", ".moai/state/kanban/backlog.json", "poll the backlog queue", "No model override", "ANTHROPIC_DEFAULT_*_MODEL"} {
		if strings.Contains(notice, gone) {
			t.Errorf("lead notice re-teaches foreman-owned polling (%q) — t96 absorbs it:\n%s", gone, notice)
		}
	}
}

// TestFactoryLeadNoticeDispatchDisciplineKorean asserts the ko locale carries
// the same codification — localized prose around the same verbatim protocol
// tokens, and the same free-slot line.
func TestFactoryLeadNoticeDispatchDisciplineKorean(t *testing.T) {
	clearKanbanEnv(t)

	root := t.TempDir()

	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	notice := factoryBootstrapNotice(root, "ko")
	for _, want := range []string{
		"`/loop`",
		"plan -> run -> sync",
		"CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS",
		"현재 빈 레인 슬롯: lane-1, lane-2.",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("ko lead notice missing dispatch-discipline token %q:\n%s", want, notice)
		}
	}
}

// TestFactoryLeadNoticeAllSlotsClaimed asserts the none-form renders when
// every slot is held by a live session — the lead must see capacity 0, not a
// silently missing line.
func TestFactoryLeadNoticeAllSlotsClaimed(t *testing.T) {
	clearKanbanEnv(t)

	root := t.TempDir()
	reg := map[string]kanban.FactoryWorkerEntry{}
	for i := 1; i <= 2; i++ {
		reg[kanban.FactoryLaneLabel(i)] = kanban.NewFactoryWorkerEntry()
	}
	if err := kanban.SaveFactoryRegistry(kanban.FactoryRegistryPath(root), reg); err != nil {
		t.Fatalf("seed registry: %v", err)
	}

	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	notice := factoryBootstrapNotice(root, langEnglish)
	if !strings.Contains(notice, "every slot is held by a live session") {
		t.Errorf("all-claimed notice must render the none-form:\n%s", notice)
	}
}
