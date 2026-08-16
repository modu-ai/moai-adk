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
		"moai cc -f 3 --name worker-1",
		"moai cc -f 3 --name worker-2",
		"moai cc -f 3 --name worker-3",
		"moai glm -f 3 --name",
		"/tmp/moai-kanban-abc123",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("lead notice missing %q:\n%s", want, notice)
		}
	}
	// The count must be exact — an off-by-one fan-out misaddresses workers.
	if got := strings.Count(notice, "moai cc -f 3 --name worker-"); got != 3 {
		t.Errorf("launch line count = %d, want 3:\n%s", got, notice)
	}
}

// TestFactoryLeadNoticeWorkerCountDrivesLineCount asserts N drives the line
// count directly (the v1 no-upper-bound rule: any N >= 1 prints N lines).
func TestFactoryLeadNoticeWorkerCountDrivesLineCount(t *testing.T) {
	clearKanbanEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "1")
	t.Setenv(config.EnvMoaiKanbanID, "xyz9")

	if got := strings.Count(factoryBootstrapNotice("", langEnglish), "moai cc -f 1 --name worker-"); got != 1 {
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

// TestFactoryLeadNoticeCarriesLoopDiscipline is the t85 codification AC: the
// lead notice must carry the loop instruction (poll the backlog queue, pick a
// card when a slot is free, dispatch, repeat), the stagger-spawn discipline,
// and the no-model-override rule — with the protocol tokens (queue path,
// command, env names) verbatim in every locale's copy — plus the LIVE data
// lines: the queued-card count read from the backlog and the free-slot list
// read from the worker registry under root.
func TestFactoryLeadNoticeCarriesLoopDiscipline(t *testing.T) {
	clearKanbanEnv(t)

	root := t.TempDir()
	// Two queued cards (plus one picked, one dropped — neither counts).
	if err := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root)).Mutate(func(rec *kanban.BacklogRecord) error {
		rec.Items = []kanban.BacklogItem{
			{ID: "t1", State: kanban.BacklogStateQueued},
			{ID: "t2", State: kanban.BacklogStatePicked},
			{ID: "t3", State: kanban.BacklogStateQueued},
			{ID: "t4", State: kanban.BacklogStateDropped},
		}
		return nil
	}); err != nil {
		t.Fatalf("seed backlog: %v", err)
	}
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
		".moai/state/kanban/backlog.json",
		"moai todo list",
		"SendMessage",
		"started producing output",
		"cache-aware-execution directive 2",
		"No model override",
		"ANTHROPIC_DEFAULT_*_MODEL",
		"Factory backlog: 2 waiting",
		"Free worker slots right now: worker-1, worker-3.",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("lead notice missing loop-discipline token %q:\n%s", want, notice)
		}
	}
}

// TestFactoryLeadNoticeLoopDisciplineKorean asserts the ko locale carries the
// same codification — localized prose around the same verbatim protocol
// tokens, and the same live data lines.
func TestFactoryLeadNoticeLoopDisciplineKorean(t *testing.T) {
	clearKanbanEnv(t)

	root := t.TempDir()
	store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root))
	if _, _, err := store.Add("card one"); err != nil {
		t.Fatalf("seed backlog: %v", err)
	}

	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")

	notice := factoryBootstrapNotice(root, "ko")
	for _, want := range []string{
		".moai/state/kanban/backlog.json",
		"moai todo list",
		"SendMessage",
		"ANTHROPIC_DEFAULT_*_MODEL",
		"cache-aware-execution directive 2",
		"모델 오버라이드 금지",
		"팩토리 백로그: 1장 대기 중",
		"현재 빈 워커 슬롯: worker-1, worker-2.",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("ko lead notice missing loop-discipline token %q:\n%s", want, notice)
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
