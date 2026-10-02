package hook

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The two leader notices (Kanban and Factory) carry one sentence that names the
// batch gate summary and points at its canonical home. This file pins that
// sentence: its two address tokens, what it must not contain, where it must not
// appear, and that the product sources stay free of the question-tool token.

const (
	// leaderGatePointer is the address of the canonical section. The path and the
	// section mark are not translated: they are an address, like the command
	// tokens elsewhere in the notices.
	leaderGatePointer = ".claude/rules/moai/workflow/auto-semantics.md` §9.2"
	// leaderGateName is the name of the presentation form, kept verbatim in every
	// locale.
	leaderGateName = "batch gate summary"
)

var (
	leaderGateLocales = []string{"en", "ko", "ja", "zh"}
	leaderGateKinds   = []string{"kanban", "factory"}

	leaderGateCardIDRe = regexp.MustCompile(`\bt\d{3,5}\b`)
	leaderGateDateRe   = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	leaderGateLeadRe   = regexp.MustCompile(`(?i)\blead\b`)
	leaderGateEpicRe   = regexp.MustCompile(`(?i)epic`)
)

// leaderGateGrid maps a leader kind to a locale to the rendered notice.
type leaderGateGrid map[string]map[string]string

// renderLeaderGateGrid renders both leader notices in every locale. The kanban
// leader notice takes its run id from the argument and the factory leader notice
// from the same; an empty root degrades the queue and slot summaries inside the
// notices rather than failing.
func renderLeaderGateGrid(t *testing.T) leaderGateGrid {
	t.Helper()
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	clearKanbanEnv(t)
	grid := leaderGateGrid{"kanban": {}, "factory": {}}
	for _, lang := range leaderGateLocales {
		grid["kanban"][lang] = kanbanLeaderNotice("tjgate", "", lang)
		grid["factory"][lang] = factoryLeaderNotice("tjgate", 2, "", lang)
	}
	return grid
}

// cloneLeaderGateGrid returns a deep copy so a mutant never touches the real grid.
func cloneLeaderGateGrid(g leaderGateGrid) leaderGateGrid {
	out := leaderGateGrid{}
	for kind, byLang := range g {
		out[kind] = map[string]string{}
		for lang, notice := range byLang {
			out[kind][lang] = notice
		}
	}
	return out
}

// gateLine returns the line of a notice that carries the sentence: the line
// holding the name token, or failing that the one holding the pointer. The
// forbidden-token checks run on that line only, because the surrounding notice
// legitimately carries tokens (the todo command, the SPEC line) the sentence
// itself must not.
func gateLine(notice string) string {
	for _, line := range strings.Split(notice, "\n") {
		if strings.Contains(line, leaderGateName) {
			return line
		}
	}
	for _, line := range strings.Split(notice, "\n") {
		if strings.Contains(line, leaderGatePointer) {
			return line
		}
	}
	return ""
}

// leaderNoticeViolations checks one rendered notice and returns its violation
// codes. The forbidden-token checks run on the sentence line only.
func leaderNoticeViolations(n string) []string {
	var codes []string
	if !strings.Contains(n, leaderGatePointer) {
		codes = append(codes, "omit-pointer")
	}
	if !strings.Contains(n, leaderGateName) {
		codes = append(codes, "omit-name-token")
	}
	if strings.Contains(n, "AskUserQuestion") || strings.Contains(n, "mcp__askuser") {
		codes = append(codes, "question-tool-name")
	}
	line := gateLine(n)
	switch {
	case leaderGateCardIDRe.MatchString(line):
		codes = append(codes, "card-id")
	case strings.Contains(line, "SPEC-"):
		codes = append(codes, "spec-id")
	case leaderGateDateRe.MatchString(line):
		codes = append(codes, "iso-date")
	case strings.Contains(line, "moai todo"):
		codes = append(codes, "moai-todo")
	case leaderGateLeadRe.MatchString(line) || strings.Contains(line, "리드"):
		codes = append(codes, "role-word-lead")
	case leaderGateEpicRe.MatchString(line):
		codes = append(codes, "epic")
	case strings.Contains(line, "%"):
		codes = append(codes, "format-verb")
	}
	return codes
}

// leaderGateViolations checks a grid and returns the violation codes of the
// first stage that finds any, so each mutant reports exactly its own code. The
// stages run in a fixed order: a locale missing from the grid, then a leader
// whose notice carries the sentence in no locale while the other leader does,
// then the per-notice checks.
func leaderGateViolations(g leaderGateGrid) []string {
	var codes []string
	add := func(code string) {
		for _, c := range codes {
			if c == code {
				return
			}
		}
		codes = append(codes, code)
	}

	for _, kind := range leaderGateKinds {
		for _, lang := range leaderGateLocales {
			if _, ok := g[kind][lang]; !ok {
				add("omit-locale")
			}
		}
	}
	if len(codes) > 0 {
		return codes
	}

	has := map[string]bool{}
	for _, kind := range leaderGateKinds {
		for _, lang := range leaderGateLocales {
			n := g[kind][lang]
			if strings.Contains(n, leaderGatePointer) || strings.Contains(n, leaderGateName) {
				has[kind] = true
			}
		}
	}
	if has["kanban"] != has["factory"] {
		return []string{"one-sided-leader"}
	}

	for _, kind := range leaderGateKinds {
		for _, lang := range leaderGateLocales {
			for _, code := range leaderNoticeViolations(g[kind][lang]) {
				add(code)
			}
		}
	}
	return codes
}

// TestLeaderNoticeBatchGatePointer pins the batch gate summary sentence of the
// two leader notices.
func TestLeaderNoticeBatchGatePointer(t *testing.T) {
	grid := renderLeaderGateGrid(t)

	t.Run("real_grid", func(t *testing.T) {
		if v := leaderGateViolations(grid); len(v) != 0 {
			t.Fatalf("real_grid violations=%v", v)
		}
		t.Logf("real_grid violations=0")
	})

	// Each of the eight leader x locale combinations must carry the sentence on
	// its own, so a table edit that drops one locale names the combination.
	for _, kind := range leaderGateKinds {
		for _, lang := range leaderGateLocales {
			t.Run(kind+"/"+lang, func(t *testing.T) {
				n := grid[kind][lang]
				if v := leaderNoticeViolations(n); len(v) != 0 {
					t.Errorf("%s notice (%s) violates %v; sentence line %q", kind, lang, v, gateLine(n))
				}
			})
		}
	}

	// Mutants: every one must be rejected, and rejected under its own name.
	mutants := []struct {
		name   string
		mutate func(g leaderGateGrid)
	}{
		{"omit-locale", func(g leaderGateGrid) { delete(g["factory"], "ja") }},
		{"question-tool-name", func(g leaderGateGrid) {
			g["kanban"]["en"] = strings.Replace(g["kanban"]["en"], leaderGateName, leaderGateName+" (AskUserQuestion)", 1)
		}},
		{"omit-pointer", func(g leaderGateGrid) {
			g["kanban"]["ko"] = strings.Replace(g["kanban"]["ko"], leaderGatePointer, "", 1)
		}},
		{"omit-name-token", func(g leaderGateGrid) {
			g["factory"]["zh"] = strings.Replace(g["factory"]["zh"], leaderGateName, "", 1)
		}},
		{"card-id", func(g leaderGateGrid) {
			g["factory"]["en"] = strings.Replace(g["factory"]["en"], leaderGateName, leaderGateName+" t1344", 1)
		}},
		{"one-sided-leader", func(g leaderGateGrid) {
			for _, lang := range leaderGateLocales {
				n := g["kanban"][lang]
				n = strings.ReplaceAll(n, leaderGatePointer, "")
				n = strings.ReplaceAll(n, leaderGateName, "")
				g["kanban"][lang] = n
			}
		}},
	}
	for _, m := range mutants {
		t.Run("mutant_"+m.name, func(t *testing.T) {
			g := cloneLeaderGateGrid(grid)
			m.mutate(g)
			got := leaderGateViolations(g)
			if len(got) != 1 || got[0] != m.name {
				t.Fatalf("mutant %s reported %v, want exactly [%s]", m.name, got, m.name)
			}
			t.Logf("mutant %s rejected: reported=%s", m.name, got[0])
		})
	}

	// Lane, companion and stale-run style notices never carry the pointer: a lane
	// holds one card and so forms no cross-card batch.
	t.Run("lane_and_companion_lack_pointer", func(t *testing.T) {
		checked := 0
		for _, lang := range leaderGateLocales {
			for name, n := range map[string]string{
				"kanban companion":      kanbanCompanionNotice("plan", lang),
				"factory lane":          factoryLaneNotice("lane-1", 2, lang),
				"factory lane no cnt":   factoryLaneNotice("lane-1", 0, lang),
				"factory lane rule":     factoryMessagesFor(lang).laneNextCardRule,
				"factory owned rule":    factoryMessagesFor(lang).laneOwnedCardRule,
				"factory manual rule":   factoryMessagesFor(lang).laneManualDispatchRule,
				"kanban companion join": kanbanMessagesFor(lang).companionJoin,
			} {
				checked++
				if strings.Contains(n, leaderGatePointer) || strings.Contains(n, leaderGateName) {
					t.Errorf("%s (%s) carries the leader-only sentence:\n%s", name, lang, n)
				}
			}
		}
		if checked == 0 {
			t.Fatal("no lane or companion notice was checked")
		}
	})

	// The four product sources stay free of the question-tool token outside
	// comments, the same scope and exclusion form as TestNoUserInteraction.
	t.Run("source_scan", func(t *testing.T) {
		files := []string{
			"session_start_kanban.go",
			"session_start_kanban_i18n.go",
			"session_start_factory.go",
			"session_start_factory_i18n.go",
		}
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", f, err)
			}
			for i, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "//") {
					continue
				}
				if strings.Contains(line, "AskUserQuestion") || strings.Contains(line, "mcp__askuser") {
					t.Errorf("%s:%d references a user-interaction token (subagent boundary): %s", f, i+1, strings.TrimSpace(line))
				}
			}
		}
	})

	// Handler level: both leaders, the English agent copy (additionalContext) and
	// the operator copy in a non-English locale (systemMessage) carry the
	// sentence.
	t.Run("handler_level", func(t *testing.T) {
		for _, kind := range leaderGateKinds {
			t.Run(kind, func(t *testing.T) {
				t.Setenv(config.EnvMoaiLaunchProvider, "")
				clearKanbanEnv(t)
				t.Setenv(config.EnvMoaiKanbanID, "tjgate")
				if kind == "kanban" {
					t.Setenv(config.EnvMoaiKanban, "1")
				} else {
					t.Setenv(config.EnvMoaiFactoryWorkers, "2")
				}
				projectDir := newKanbanProjectDir(t)
				out, err := NewSessionStartHandler(configWithLang("ko")).Handle(context.Background(), &HookInput{
					SessionID:  "uuid-leader-gate-" + kind,
					CWD:        projectDir,
					ProjectDir: projectDir,
				})
				if err != nil {
					t.Fatalf("Handle: %v", err)
				}
				ac := out.HookSpecificOutput.AdditionalContext
				for channel, text := range map[string]string{"additionalContext": ac, "systemMessage": out.SystemMessage} {
					if !strings.Contains(text, leaderGatePointer) || !strings.Contains(text, leaderGateName) {
						t.Errorf("%s leader %s lacks the sentence:\n%s", kind, channel, text)
					}
				}
				if strings.Contains(ac, "운영자") {
					t.Errorf("%s leader additionalContext leaked the operator locale", kind)
				}
			})
		}
	})
}
