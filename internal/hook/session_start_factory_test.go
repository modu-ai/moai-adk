package hook

import (
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

// TestFactoryLeadNoticeCarriesWorkerLinesSocketAndGLMSubstitute is the AC
// bundle for the lead notice: N worker launch lines carrying -f <N>, the
// leader socket path, the GLM substitute guidance naming `moai glm -f <N>`,
// and the run id alongside the session name that must match it.
func TestFactoryLeadNoticeCarriesWorkerLinesSocketAndGLMSubstitute(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")
	t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-abc123")

	notice := factoryBootstrapNotice("", langEnglish)
	for _, want := range []string{
		"run abc123",
		"lead-abc123",
		"moai cc -k 3 --name worker-1",
		"moai cc -k 3 --name worker-2",
		"moai cc -k 3 --name worker-3",
		"moai glm -k 3 --name",
		"/tmp/moai-kanban-abc123",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("lead notice missing %q:\n%s", want, notice)
		}
	}
	// The count must be exact — an off-by-one fan-out misaddresses workers.
	if got := strings.Count(notice, "moai cc -k 3 --name worker-"); got != 3 {
		t.Errorf("launch line count = %d, want 3:\n%s", got, notice)
	}
}

// TestFactoryLeadNoticeWorkerCountDrivesLineCount asserts N drives the line
// count directly (the v1 no-upper-bound rule: any N >= 1 prints N lines).
func TestFactoryLeadNoticeWorkerCountDrivesLineCount(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	t.Setenv(config.EnvMoaiKanbanID, "xyz9")

	if got := strings.Count(factoryBootstrapNotice("", langEnglish), "moai cc -k 1 --name worker-"); got != 1 {
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
// screen.
func TestFactoryWorkerNoticeNamesLabel(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorker, "worker-4")
	t.Setenv(config.EnvMoaiFactoryWorkers, "3")

	notice := factoryBootstrapNotice("", langEnglish)
	if !strings.Contains(notice, "worker-4") {
		t.Errorf("worker notice must name the (possibly bumped) label:\n%s", notice)
	}

	// A malformed label emits nothing (fail-open, mirroring the companion
	// branch) — no error, no notice.
	t.Setenv(config.EnvMoaiFactoryWorker, "not-a-worker-label")
	if got := factoryBootstrapNotice("", langEnglish); got != "" {
		t.Errorf("malformed worker label must emit nothing, got:\n%s", got)
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
// the four-role kanban notice under the factory one.
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
