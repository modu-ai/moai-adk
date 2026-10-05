package cli

// handoff.go registers `moai handoff save` / `moai handoff clear` / `moai
// handoff show` — the writer half of the reverse auto-resume handoff
// (SPEC-HANDOFF-AUTORESUME-001 M2) plus the harness-neutral reprint path
// (SPEC-HANDOFF-NEUTRAL-001 M1.2). The pending record is written to the
// project's factory.db via the internal/hook/handoff package; it NEVER
// consumes the SessionEnd flow's separate memory-handoff rows
// (REQ-AUTORESUME-005/007).

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	goalpkg "github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

func init() {
	rootCmd.AddCommand(newHandoffCmd())
}

// newHandoffCmd builds the `moai handoff` command tree (save / clear / show). It is a
// constructor rather than package-level command vars so tests can build isolated
// instances without racing on shared global flag state.
func newHandoffCmd() *cobra.Command {
	handoffCmd := &cobra.Command{
		Use:     "handoff",
		Short:   "Manage the auto-resume handoff pending record",
		GroupID: "tools",
		Long: "Save, clear, or show the reverse auto-resume handoff pending record\n" +
			"(~/.moai/db/<project-key>/factory/factory.db). When handoff.mode=auto, the next\n" +
			"SessionStart on /clear injects the saved record as session context.",
	}

	var projectDir string
	handoffCmd.PersistentFlags().StringVar(&projectDir, "project-dir", "", "project root (default: current working directory)")

	handoffCmd.AddCommand(newHandoffSaveCmd(&projectDir), newHandoffClearCmd(&projectDir), newHandoffShowCmd(&projectDir))
	return handoffCmd
}

// newHandoffSaveCmd builds the `save` subcommand. projectDir points at the
// parent's persistent flag var.
func newHandoffSaveCmd(projectDir *string) *cobra.Command {
	var (
		body       string
		spec       string
		phase      string
		session    string
		lang       string
		ultrathink bool
		ultracode  bool
		goal       string
		useStdin   bool
	)

	saveCmd := &cobra.Command{
		Use:   "save",
		Short: "Save a paste-ready resume body as the pending handoff record",
		RunE: func(cmd *cobra.Command, _ []string) error {
			b := body
			if useStdin {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("handoff save: read stdin: %w", err)
				}
				b = string(data)
			}
			pd, err := handoffProjectDir(*projectDir)
			if err != nil {
				return err
			}
			// SPEC-INFINITE-GOAL-001 REQ-6: embed any live armed goal so the
			// /clear handler re-arms it under the new session-id. Reads the SAME
			// per-session goal file the arm verb wrote. Best-effort: a read error
			// → no embed (the /clear handler then no-ops the rearm).
			var embedded *handoff.EmbeddedGoal
			embedSession := session
			if embedSession == "" {
				embedSession = statusSessionID("")
			}
			if g, loadErr := goalpkg.LoadGoal(pd, embedSession); loadErr == nil && g != nil && g.Status == goalpkg.StatusArmed {
				embedded = &handoff.EmbeddedGoal{
					Condition:   g.Goal,
					MaxTurns:    g.Ceiling.MaxTurns,
					MaxDuration: g.Ceiling.MaxDuration,
					CostCap:     g.Ceiling.CostCap,
				}
			}
			rec := &handoff.PendingRecord{
				SchemaVersion:        handoff.PendingSchemaVersion,
				SpecID:               spec,
				Phase:                phase,
				SavedBySession:       session,
				ConversationLanguage: lang,
				Directives: handoff.Directives{
					Ultrathink: ultrathink,
					Ultracode:  ultracode,
					Goal:       goal,
				},
				EmbeddedGoal: embedded,
				Body:         b,
			}
			path, err := saveHandoff(pd, rec)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "handoff saved: %s\n", path)
			return nil
		},
	}
	saveCmd.Flags().StringVar(&body, "body", "", "resume body (verbatim 6-block paste-ready)")
	saveCmd.Flags().BoolVar(&useStdin, "stdin", false, "read the resume body from stdin instead of --body")
	saveCmd.Flags().StringVar(&spec, "spec", "", "SPEC id this handoff resumes")
	saveCmd.Flags().StringVar(&phase, "phase", "", "phase (plan|run|sync)")
	saveCmd.Flags().StringVar(&session, "session", "", "saved_by_session uuid (attribution)")
	saveCmd.Flags().StringVar(&lang, "lang", "", "conversation_language snapshot")
	saveCmd.Flags().BoolVar(&ultrathink, "ultrathink", false, "record the ultrathink directive (restoration guidance only)")
	saveCmd.Flags().BoolVar(&ultracode, "ultracode", false, "record the ultracode directive (restoration guidance only)")
	saveCmd.Flags().StringVar(&goal, "goal", "", "record a /moai goal condition (restoration guidance only)")
	return saveCmd
}

// newHandoffClearCmd builds the `clear` subcommand.
func newHandoffClearCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove the pending handoff record",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pd, err := handoffProjectDir(*projectDir)
			if err != nil {
				return err
			}
			path, err := clearHandoff(pd)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "handoff cleared: %s\n", path)
			return nil
		},
	}
}

// handoffProjectDir returns explicit if non-empty, else the current working dir.
func handoffProjectDir(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("handoff: resolve project dir: %w", err)
	}
	return wd, nil
}

// saveHandoff writes rec to factory.db and returns the database path.
// Extracted from the cobra RunE so it is directly unit-testable.
func saveHandoff(projectDir string, rec *handoff.PendingRecord) (string, error) {
	if rec == nil || strings.TrimSpace(rec.Body) == "" {
		return "", fmt.Errorf("handoff save: body is required (use --body or --stdin)")
	}
	if err := handoff.SavePending(projectDir, rec); err != nil {
		return "", fmt.Errorf("handoff save: %w", err)
	}
	return handoff.PendingPath(projectDir), nil
}

// clearHandoff marks the pending resume row cleared and returns the database path.
func clearHandoff(projectDir string) (string, error) {
	if err := handoff.ClearPending(projectDir); err != nil {
		return "", fmt.Errorf("handoff clear: %w", err)
	}
	return handoff.PendingPath(projectDir), nil
}

// --- SPEC-HANDOFF-NEUTRAL-001 M1.2: `moai handoff show` (P3 path) ---
//
// show is the harness-neutral consumer: it re-prints the stored 6-block body
// verbatim so a session of ANY harness (or a human terminal) can consume a
// handoff by paste. Read-only by contract (REQ-HN-002): no claim, no
// consumption, no state transition — safe to run beside the auto-inject flow.

// handoffShowSource values are protocol tokens rendered verbatim in every
// locale (identifiers, not prose).
const (
	handoffSourcePending  = "pending"
	handoffSourceConsumed = "consumed"
)

// newHandoffShowCmd builds the `show` subcommand.
func newHandoffShowCmd(projectDir *string) *cobra.Command {
	var asJSON bool

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Reprint the saved handoff body (harness-neutral paste path)",
		Long: "Reprint the stored resume handoff: the pending row first, falling back to\n" +
			"the latest consumed row. The 6-block body is printed verbatim for copy-paste.\n" +
			"Read-only — never claims or consumes the record.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pd, err := handoffProjectDir(*projectDir)
			if err != nil {
				return err
			}
			rec, source, err := readHandoffForShow(pd)
			if err != nil {
				return err
			}
			if rec == nil {
				return fmt.Errorf("no saved handoff found — 저장된 핸드오프 없음 (save one with `moai handoff save --stdin`)")
			}
			if asJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(handoffShowJSON{Source: source, Record: rec})
			}
			printHandoffShow(cmd.OutOrStdout(), rec, source)
			return nil
		},
	}
	showCmd.Flags().BoolVar(&asJSON, "json", false, "print the full record as JSON with its source (pending/consumed)")
	return showCmd
}

// handoffShowJSON is the --json envelope: the record plus its source.
type handoffShowJSON struct {
	Source string                 `json:"source"`
	Record *handoff.PendingRecord `json:"record"`
}

// readHandoffForShow resolves the show source (REQ-HN-001): the pending row
// first (ReadPending), then the latest consumed row (consumed_at DESC). A
// pending-read error falls through to the consumed fallback; only when both
// paths fail (or error with nothing to show) does it return an error.
func readHandoffForShow(pd string) (*handoff.PendingRecord, string, error) {
	rec, present, pendingErr := handoff.ReadPending(pd)
	if pendingErr == nil && present {
		return rec, handoffSourcePending, nil
	}

	db, err := homestate.OpenFactory(pd)
	if err != nil {
		if pendingErr != nil {
			return nil, "", fmt.Errorf("handoff show: read pending: %v; open consumed history: %w", pendingErr, err)
		}
		return nil, "", nil
	}
	defer func() { _ = db.Close() }()
	row, present, err := db.ReadLatestConsumedResume(context.Background())
	if err != nil {
		if pendingErr != nil {
			return nil, "", fmt.Errorf("handoff show: read pending: %v; read consumed history: %w", pendingErr, err)
		}
		return nil, "", fmt.Errorf("handoff show: read consumed history: %w", err)
	}
	if !present {
		if pendingErr != nil {
			return nil, "", fmt.Errorf("handoff show: read pending: %w", pendingErr)
		}
		return nil, "", nil
	}
	consumed, convErr := consumedShowRecord(row)
	if convErr != nil {
		return nil, "", fmt.Errorf("handoff show: %w", convErr)
	}
	return consumed, handoffSourceConsumed, nil
}

// consumedShowRecord maps a consumed homestate row onto the show payload. It
// mirrors hook/handoff's row→record mapping on the presentation side so the
// show verb adds no write-path surface to the hook package.
func consumedShowRecord(row *homestate.ResumeHandoff) (*handoff.PendingRecord, error) {
	if row == nil {
		return nil, fmt.Errorf("nil consumed row")
	}
	rec := &handoff.PendingRecord{
		ID:                   row.ID,
		SchemaVersion:        row.SchemaVersion,
		SpecID:               row.SpecID,
		Phase:                row.Phase,
		SavedAt:              row.SavedAt,
		SavedBySession:       row.SavedBySession,
		ConversationLanguage: row.ConversationLanguage,
		Body:                 row.Body,
	}
	if row.DirectivesJSON != "" {
		if err := json.Unmarshal([]byte(row.DirectivesJSON), &rec.Directives); err != nil {
			return nil, fmt.Errorf("parse directives: %w", err)
		}
	}
	if row.EmbeddedGoalJSON != nil {
		var goal handoff.EmbeddedGoal
		if err := json.Unmarshal([]byte(*row.EmbeddedGoalJSON), &goal); err != nil {
			return nil, fmt.Errorf("parse embedded goal: %w", err)
		}
		rec.EmbeddedGoal = &goal
	}
	return rec, nil
}

// printHandoffShow renders the human output: a localized one-line header
// (REQ-HN-004 — ko/en/ja/zh per the stored ConversationLanguage, en fallback)
// followed by the stored Body verbatim. The body is never localized or
// reformatted (REQ-HN-010).
func printHandoffShow(w io.Writer, rec *handoff.PendingRecord, source string) {
	s := handoffShowLocaleStrings(rec.ConversationLanguage)
	consumedWord := s.no
	if source == handoffSourceConsumed {
		consumedWord = s.yes
	}
	spec := rec.SpecID
	if spec == "" {
		spec = "-"
	}
	phase := rec.Phase
	if phase == "" {
		phase = "-"
	}
	lang := rec.ConversationLanguage
	if lang == "" {
		lang = "-"
	}
	_, _ = fmt.Fprintf(w, "%s\n%s: %s | %s: %s | %s: %s | %s: %s | %s: %s | %s: %s\n%s\n%s\n",
		s.title,
		s.source, source,
		s.spec, spec,
		s.phase, phase,
		s.lang, lang,
		s.savedAt, rec.SavedAt.Format(time.RFC3339),
		s.consumed, consumedWord,
		s.separator,
		rec.Body)
}

// handoffShowStrings carries the localized header labels. The source VALUES
// (pending/consumed) stay verbatim in every locale — they are protocol tokens.
// Follows the handoffLocaleStrings convention of the injector render
// (internal/hook/handoff_inject_render.go): 4 locales, en fallback.
type handoffShowStrings struct {
	title     string
	source    string
	spec      string
	phase     string
	lang      string
	savedAt   string
	consumed  string
	yes       string
	no        string
	separator string
}

// handoffShowLocale normalizes a conversation_language to {ko, en, ja, zh},
// falling back to en (same rule as the injector render).
func handoffShowLocale(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "ko", "en", "ja", "zh":
		return strings.ToLower(strings.TrimSpace(lang))
	default:
		return "en"
	}
}

// handoffShowLocaleStrings returns the localized header labels for lang.
func handoffShowLocaleStrings(lang string) handoffShowStrings {
	switch handoffShowLocale(lang) {
	case "ko":
		return handoffShowStrings{
			title:     "[MoAI 핸드오프 — 저장된 인계 재출력]",
			source:    "출처",
			spec:      "SPEC",
			phase:     "단계",
			lang:      "언어",
			savedAt:   "저장시각",
			consumed:  "소비됨",
			yes:       "예",
			no:        "아니오",
			separator: "──── 저장된 인계 본문 (그대로 복사) ────",
		}
	case "ja":
		return handoffShowStrings{
			title:     "[MoAI ハンドオフ — 保存された引継ぎの再出力]",
			source:    "出所",
			spec:      "SPEC",
			phase:     "フェーズ",
			lang:      "言語",
			savedAt:   "保存時刻",
			consumed:  "消費済み",
			yes:       "はい",
			no:        "いいえ",
			separator: "──── 保存された引継ぎ本文（そのままコピー） ────",
		}
	case "zh":
		return handoffShowStrings{
			title:     "[MoAI 交接 — 重新输出已保存的交接]",
			source:    "来源",
			spec:      "SPEC",
			phase:     "阶段",
			lang:      "语言",
			savedAt:   "保存时间",
			consumed:  "已消费",
			yes:       "是",
			no:        "否",
			separator: "──── 已保存的交接正文（原样复制） ────",
		}
	default: // en
		return handoffShowStrings{
			title:     "[MoAI handoff — saved handoff reprint]",
			source:    "Source",
			spec:      "SPEC",
			phase:     "Phase",
			lang:      "Lang",
			savedAt:   "Saved",
			consumed:  "Consumed",
			yes:       "yes",
			no:        "no",
			separator: "──── saved handoff body (copy verbatim) ────",
		}
	}
}
