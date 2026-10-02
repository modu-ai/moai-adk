package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestFactoryGuideNamesWorkerJoinInEveryLocale pins the lane-join guidance of
// the factory notice tables (SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-009): in every
// locale the lane-start sentence names the three `-l` entries and the entry
// guide names the two `-f` leader entries, and neither teaches the removed
// `-f lane` / `-f lane-<n>` forms.
func TestFactoryGuideNamesWorkerJoinInEveryLocale(t *testing.T) {
	for lang, m := range factoryLocales {
		for _, want := range []string{"`moai cc -l`", "`moai glm -l`", "`moai codex -l`"} {
			if !strings.Contains(m.leaderManual, want) {
				t.Errorf("%s leaderManual missing %q:\n%s", lang, want, m.leaderManual)
			}
		}
		for _, want := range []string{"`moai cc -f`", "`moai glm -f`"} {
			if !strings.Contains(m.entryGuide, want) {
				t.Errorf("%s entryGuide missing %q:\n%s", lang, want, m.entryGuide)
			}
		}
		for field, text := range map[string]string{"leaderManual": m.leaderManual, "entryGuide": m.entryGuide} {
			for _, banned := range []string{"-f lane", "lane-<n>"} {
				if strings.Contains(text, banned) {
					t.Errorf("%s %s still teaches %q:\n%s", lang, field, banned, text)
				}
			}
		}
	}
}

// TestFactoryBootstrapNoticeSilentForOrdinarySession is the blast-radius
// case: a session that is not part of a factory run is completely unaffected.
func TestFactoryBootstrapNoticeSilentForOrdinarySession(t *testing.T) {
	clearFactoryEnv(t)

	if got := factoryBootstrapNotice("", "", langEnglish); got != "" {
		t.Errorf("ordinary session must get no factory notice, got:\n%s", got)
	}
}

// TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide is the AC bundle
// for the lead notice (t118, re-pinned by SPEC-LAUNCHER-ENTRY-FLAGS-001 M4):
// the one lane-start sentence naming the three `-l` entries (no per-lane
// launch line), the entry-point guidance naming the `-f` leader entries, the
// per-lane fan-out line, the leader socket path, and the run id alongside the
// session name that must match it.
func TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide(t *testing.T) {
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	clearFactoryEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	t.Setenv(config.EnvFactoryRunID, "abc123")
	t.Setenv(config.EnvFactoryLeadAddr, "/tmp/moai-socket-factory/abc123")

	notice := factoryBootstrapNotice("", "", langEnglish)
	for _, want := range []string{
		"run abc123",
		// The lead is named by its bare role (t133): the run id lives in the
		// header line above, not in the session name.
		"named leader.",
		"start a lane, enter `moai cc -l`",
		"`moai glm -l`",
		"`moai codex -l`",
		"`moai cc -f` starts a Claude factory leader",
		"`moai glm -f` a GLM leader",
		"Every lane can run up to 10 agents concurrently in parallel.",
		"/tmp/moai-socket-factory/abc123",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("lead notice missing %q:\n%s", want, notice)
		}
	}
	// The count-independence and the absence of numbered lines are pinned by
	// TestFactoryLeadNoticeIsLaneCountIndependent (the line-count test this
	// replaced printed one launch line per declared lane).
}

// TestFactoryLeadNoticeEmptyWithoutRunID asserts the fail-open shape: a lead
// with no run id (or a nonsensical count) emits nothing rather than a notice
// addressing an unnamed run.
func TestFactoryLeadNoticeEmptyWithoutRunID(t *testing.T) {
	clearFactoryEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	if got := factoryBootstrapNotice("", "", langEnglish); got != "" {
		t.Errorf("lead without a run id must emit nothing, got:\n%s", got)
	}
}

// TestFactoryWorkerNoticeNamesLabel asserts the join ack names the label the
// session actually launched under — the reliable surface for a bumped number,
// since the launcher's stderr note is gone by the time the TUI takes the
// screen. The exact-sentence assertion also pins the (label, count) argument
// order of the laneJoin format: the pre-t118 formats carried %d before %s
// while the call passed the label first, rendering %!d(string=lane-4) —
// a Contains("lane-4") assertion passed right through that garbage, so the
// whole sentence is asserted here.
func TestFactoryWorkerNoticeNamesLabel(t *testing.T) {
	clearFactoryEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorker, "lane-4")
	t.Setenv(config.EnvMoaiFactoryWorkers, "3")

	// Card t224: the notice is the join line PLUS the standing spawn authority
	// (appended, never substituted) — assert the join line as a prefix-presence
	// rather than whole-output equality.
	got := factoryBootstrapNotice("", "", langEnglish)
	if !strings.HasPrefix(got, "Factory Mode: joined the leader's 3-lane run as lane-4.") {
		t.Errorf("lane notice missing the join line prefix:\n%s", got)
	}
	if !strings.Contains(got, "Standing spawn authority") {
		t.Errorf("lane notice lost the standing spawn authority:\n%s", got)
	}

	// The incremental `-f worker-<n>` form carries no count (workers=0); the
	// count-less sentence must render, not fabricate a fan-out size and not
	// leak a bad verb.
	t.Setenv(config.EnvMoaiFactoryWorkers, "0")
	if got := factoryBootstrapNotice("", "", langEnglish); !strings.HasPrefix(got,
		"Factory Mode: joined the leader's factory run as lane-4.") {
		t.Errorf("count-less lane notice missing the join line prefix:\n%s", got)
	}

	// A malformed label emits nothing (fail-open, mirroring the companion
	// branch) — no error, no notice.
	t.Setenv(config.EnvMoaiFactoryWorker, "not-a-worker-label")
	if got := factoryBootstrapNotice("", "", langEnglish); got != "" {
		t.Errorf("malformed lane label must emit nothing, got:\n%s", got)
	}
}

// TestFactoryWorkerNoticeLocaleWordOrders pins the explicit-index contract on
// the two sentence shapes across all four locales: every rendering must name
// the label and the count (or just the label, count-less) with no %!verb
// artifact, whatever order the locale's natural prose puts them in.
// Card t1335 (M3) adds the two checks the composition tests lacked: the zh
// label-first order pin (the one locale whose %[n] order flips, asserted
// against the en count-first contrast) and leading/trailing-newline hygiene
// on every join-line i18n field in both tables.
func TestFactoryWorkerNoticeLocaleWordOrders(t *testing.T) {
	clearFactoryEnv(t)

	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		got := factoryLaneNotice("lane-2", 5, lang)
		if !strings.Contains(got, "lane-2") || !strings.Contains(got, "5") ||
			strings.Contains(got, "%!") {
			t.Errorf("locale %q worker join rendered wrong: %q", lang, got)
		}
		gotNoCount := factoryLaneNotice("lane-2", 0, lang)
		if !strings.Contains(gotNoCount, "lane-2") || strings.Contains(gotNoCount, "%!") {
			t.Errorf("locale %q count-less join rendered wrong: %q", lang, gotNoCount)
		}
	}

	// zh label-first order pin: zh is the one locale whose laneJoin places
	// %[1]s (the label) BEFORE %[2]d (the count); en pins the flipped
	// contrast (count first). A later format-string reorder in either
	// direction fails here.
	zhJoin := factoryLaneNotice("lane-2", 5, "zh")
	if labelAt, countAt := strings.Index(zhJoin, "lane-2"), strings.Index(zhJoin, "5"); labelAt < 0 || countAt < 0 || labelAt >= countAt {
		t.Errorf("zh join must render the label before the count (label at %d, count at %d):\n%s", labelAt, countAt, zhJoin)
	}
	enJoin := factoryLaneNotice("lane-2", 5, "en")
	if labelAt, countAt := strings.Index(enJoin, "lane-2"), strings.Index(enJoin, "5"); labelAt < 0 || countAt < 0 || countAt >= labelAt {
		t.Errorf("en join must render the count before the label (label at %d, count at %d):\n%s", labelAt, countAt, enJoin)
	}

	// Newline hygiene: no join-line i18n field in either table may lead or
	// trail with a newline — the builders join with "\n\n" themselves, so a
	// stray newline in a field would double the separator.
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		fm := factoryMessagesFor(lang)
		for name, field := range map[string]string{
			"laneJoin":        fm.laneJoin,
			"laneJoinNoCount": fm.laneJoinNoCount,
		} {
			if strings.HasPrefix(field, "\n") || strings.HasSuffix(field, "\n") {
				t.Errorf("factory %s for locale %q has a leading/trailing newline: %q", name, lang, field)
			}
		}
	}
}

// TestFactoryBootstrapNoticeStartupOnly asserts the re-entry gating:
// resume / clear / compact / fork re-emit nothing, because the operator's lane
// terminals are already open by then.
func TestFactoryBootstrapNoticeStartupOnly(t *testing.T) {
	clearFactoryEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvFactoryRunID, "abc123")

	for _, source := range []string{"resume", "clear", "compact", "fork", "upgrade-mystery"} {
		if got := factoryBootstrapNoticeForSource(source, "", "", langEnglish); got != "" {
			t.Errorf("source %q must not re-announce the bootstrap, got:\n%s", source, got)
		}
	}
	if got := factoryBootstrapNoticeForSource("startup", "", "", langEnglish); got == "" {
		t.Error("source startup must announce the bootstrap")
	}
	if got := factoryBootstrapNoticeForSource("", "", "", langEnglish); got == "" {
		t.Error("an empty source is treated as startup and must announce")
	}
}

// TestFactoryMessagesLocaleFallback asserts the locale table resolves the
// four conversation languages and falls back to English for anything else.
func TestFactoryMessagesLocaleFallback(t *testing.T) {
	t.Parallel()

	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		if factoryMessagesFor(lang).leaderHeader == "" {
			t.Errorf("locale %q resolved to an empty message set", lang)
		}
	}
	if factoryMessagesFor("fr").leaderHeader != factoryMessagesFor(langEnglish).leaderHeader {
		t.Error("an unknown locale must fall back to English")
	}
}

// TestFactoryLeadNoticeCarriesDispatchDiscipline is the t85 codification AC
// (v1.2.0, reworded for the t118 final design): the lead notice carries the
// FACTORY-specific dispatch discipline — whole-card routing (every card to
// ONE lane, which runs the serial plan -> run -> sync path in-session), the
// factory foreman handoff line, the fan-out-only stagger rule (workflow
// auto-stagger explicitly excluded), and the no-model-override rule. The
// free-slot line is gone with the lane count (SPEC-LAUNCHER-ENTRY-FLAGS-001
// REQ-009, M4).
// It deliberately does NOT teach queue polling: that loop is the foreman's
// (t96), and a second polling protocol here would conflict.
func TestFactoryLeadNoticeCarriesDispatchDiscipline(t *testing.T) {
	clearFactoryEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "3")
	t.Setenv(config.EnvFactoryRunID, "abc123")

	notice := factoryBootstrapNotice("", "", langEnglish)
	for _, want := range []string{
		"every card is routed WHOLE to one lane",
		"serial 3-stage path",
		"plan -> run -> sync",
		"factory foreman loop (bare `/loop`)",
		"started producing output",
		"cache-aware-execution directive 2",
		"FACTORY fan-out only",
		"CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("lead notice missing dispatch-discipline token %q:\n%s", want, notice)
		}
	}
	// The t96-absorbed content must NOT come back: no queue-polling protocol.
	// The no-model-override discipline is likewise retired from the notice
	// (operator request): the rule itself lives in the factory skill docs.
	for _, gone := range []string{"moai todo list", ".moai/state/kanban/backlog.json", "poll the backlog queue", "No model override", "ANTHROPIC_DEFAULT_*_MODEL", "kanban foreman"} {
		if strings.Contains(notice, gone) {
			t.Errorf("lead notice re-teaches foreman-owned polling or the old foreman name (%q):\n%s", gone, notice)
		}
	}
}

// TestFactoryLeadNoticeDispatchDisciplineKorean asserts the ko locale carries
// the same codification — localized prose around the same verbatim protocol
// tokens, with the foreman named in its factory rendering.
func TestFactoryLeadNoticeDispatchDisciplineKorean(t *testing.T) {
	clearFactoryEnv(t)

	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvFactoryRunID, "abc123")

	notice := factoryBootstrapNotice("", "", "ko")
	for _, want := range []string{
		"`/loop`",
		"plan -> run -> sync",
		"CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS",
		"팩토리 포어맨",
	} {
		if !strings.Contains(notice, want) {
			t.Errorf("ko lead notice missing dispatch-discipline token %q:\n%s", want, notice)
		}
	}
	if strings.Contains(notice, "칸반 포어맨") {
		t.Errorf("ko lead notice still names the kanban foreman:\n%s", notice)
	}
}
