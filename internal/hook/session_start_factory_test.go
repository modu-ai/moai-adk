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

// TestFactoryLeadNoticeCarriesWorkerLinesSocketAndEntryGuide is the AC bundle
// for the lead notice (t118): N worker launch lines carrying the incremental
// `-f worker-<i>` form, the entry-point guidance naming `moai glm -f <N>` and
// the `-f worker-<n>` form, the per-lane fan-out line, the leader socket
// path, and the run id alongside the session name that must match it.
func TestFactoryLeadNoticeCarriesWorkerLinesSocketAndEntryGuide(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")
	t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-socket-factory/abc123")

	notice := factoryBootstrapNotice("", langEnglish)
	for _, want := range []string{
		"run abc123",
		"leader-abc123",
		"moai cc -f worker-1",
		"moai cc -f worker-2",
		"moai cc -f worker-3",
		"moai glm -f 3",
		"moai cc -f worker-<n>",
		"one-worker default",
		"Every worker can run up to 10 agents concurrently in parallel.",
		"/tmp/moai-socket-factory/abc123",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("lead notice missing %q:\n%s", want, notice)
		}
	}
	// The count must be exact — an off-by-one fan-out misaddresses workers.
	// The launch lines are NUMBERED; the entry guide's generic
	// `moai cc -f worker-<n>` mention is not a launch line, so the count
	// matches digits only.
	if got := len(factoryLaunchLineRe.FindAllString(notice, -1)); got != 3 {
		t.Errorf("launch line count = %d, want 3:\n%s", got, notice)
	}
}

// factoryLaunchLineRe matches exactly the per-worker launch lines of the
// factory lead notice (`moai cc -f worker-<i>`, i a number) — the entry
// guide's `worker-<n>` placeholder is excluded by the digit class.
var factoryLaunchLineRe = regexp.MustCompile(`moai cc -f worker-[0-9]+`)

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
// while the call passed the label first, rendering %!d(string=worker-4) —
// a Contains("worker-4") assertion passed right through that garbage, so the
// whole sentence is asserted here.
func TestFactoryWorkerNoticeNamesLabel(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorker, "worker-4")
	t.Setenv(config.EnvMoaiFactoryWorkers, "3")

	if got, want := factoryBootstrapNotice("", langEnglish),
		"Factory Mode: joined a 3-worker run as worker-4."; got != want {
		t.Errorf("worker notice = %q, want %q", got, want)
	}

	// The incremental `-f worker-<n>` form carries no count (workers=0); the
	// count-less sentence must render, not fabricate a fan-out size and not
	// leak a bad verb.
	t.Setenv(config.EnvMoaiFactoryWorkers, "0")
	if got, want := factoryBootstrapNotice("", langEnglish),
		"Factory Mode: joined the factory run as worker-4."; got != want {
		t.Errorf("count-less worker notice = %q, want %q", got, want)
	}

	// A malformed label emits nothing (fail-open, mirroring the companion
	// branch) — no error, no notice.
	t.Setenv(config.EnvMoaiFactoryWorker, "not-a-worker-label")
	if got := factoryBootstrapNotice("", langEnglish); got != "" {
		t.Errorf("malformed worker label must emit nothing, got:\n%s", got)
	}
}

// TestFactoryWorkerNoticeLocaleWordOrders pins the explicit-index contract on
// the two sentence shapes across all four locales: every rendering must name
// the label and the count (or just the label, count-less) with no %!verb
// artifact, whatever order the locale's natural prose puts them in.
func TestFactoryWorkerNoticeLocaleWordOrders(t *testing.T) {
	clearKanbanEnv(t)

	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		got := factoryWorkerNotice("worker-2", 5, lang)
		if !strings.Contains(got, "worker-2") || !strings.Contains(got, "5") ||
			strings.Contains(got, "%!") {
			t.Errorf("locale %q worker join rendered wrong: %q", lang, got)
		}
		gotNoCount := factoryWorkerNotice("worker-2", 0, lang)
		if !strings.Contains(gotNoCount, "worker-2") || strings.Contains(gotNoCount, "%!") {
			t.Errorf("locale %q count-less join rendered wrong: %q", lang, gotNoCount)
		}
	}
}

// TestFactoryBootstrapNoticeStartupOnly asserts the re-entry gating shared
// with the kanban notice: resume / clear / compact / fork re-emit nothing,
// because the operator's worker terminals are already open by then.
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
// (v1.2.0): the lead notice carries the FACTORY-specific dispatch discipline —
// card-class routing (A/B wholesale, C serial 3-stage), the foreman handoff
// line, the fan-out-only stagger rule (workflow auto-stagger explicitly
// excluded), and the no-model-override rule — plus the live free-slot line.
// It deliberately does NOT teach queue polling: that loop is the kanban
// foreman's (t96), and a second polling protocol here would conflict.
func TestFactoryLeadNoticeCarriesDispatchDiscipline(t *testing.T) {
	clearKanbanEnv(t)

	root := t.TempDir()
	// worker-2 claimed by THIS test process — a pid that is genuinely alive,
	// so slot 2 reads busy without a probe seam.
	if err := kanban.SaveFactoryRegistry(kanban.FactoryRegistryPath(root), map[string]kanban.FactoryWorkerEntry{
		"worker-2": kanban.NewFactoryWorkerEntry(),
	}); err != nil {
		t.Fatalf("seed registry: %v", err)
	}

	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	notice := factoryBootstrapNotice(root, langEnglish)
	for _, want := range []string{
		"A-class and B-class cards are fanned out WHOLESALE",
		"serial 3-stage path",
		"plan -> run -> sync",
		"kanban foreman loop (bare `/loop`)",
		"started producing output",
		"cache-aware-execution directive 2",
		"FACTORY fan-out only",
		"CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS",
		"No model override",
		"ANTHROPIC_DEFAULT_*_MODEL",
		"Free worker slots right now: worker-1, worker-3.",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("lead notice missing dispatch-discipline token %q:\n%s", want, notice)
		}
	}
	// The t96-absorbed content must NOT come back: no queue-polling protocol.
	for _, gone := range []string{"moai todo list", ".moai/state/kanban/backlog.json", "poll the backlog queue"} {
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
		"모델 오버라이드 금지",
		"ANTHROPIC_DEFAULT_*_MODEL",
		"현재 빈 워커 슬롯: worker-1, worker-2.",
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
		reg[kanban.FactoryWorkerLabel(i)] = kanban.NewFactoryWorkerEntry()
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
