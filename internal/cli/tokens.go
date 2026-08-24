package cli

// @MX:NOTE: [AUTO] Per-pool token accounting seed: reads a CC session transcript,
// @MX:NOTE: [AUTO] buckets usage per pool (glm/claude/other) split by main vs subagent
// @MX:NOTE: [AUTO] origin, and appends one JSON record to .moai/state/token-accounting.jsonl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/spf13/cobra"
)

// tokensSchemaVersion is the ledger-record schema version for
// .moai/state/token-accounting.jsonl.
const tokensSchemaVersion = 1

// tokensLedgerFilename is the append-only ledger file inside the state dir.
const tokensLedgerFilename = "token-accounting.jsonl"

// tokenUsageTotals is one per-origin token bucket of a pool or model.
type tokenUsageTotals struct {
	InputTokens         int64 `json:"input_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_tokens"`
	CacheReadTokens     int64 `json:"cache_read_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
}

// tokensOriginSplit aggregates one pool (or one exact model) split by origin:
// main conversation vs subagent sidechain.
type tokensOriginSplit struct {
	Main     tokenUsageTotals `json:"main"`
	Subagent tokenUsageTotals `json:"subagent"`
}

// add accumulates a transcript usage object into the given origin bucket.
func (s *tokensOriginSplit) add(sidechain bool, u transcriptUsageMsg) {
	target := &s.Main
	if sidechain {
		target = &s.Subagent
	}
	target.InputTokens += u.InputTokens
	target.CacheCreationTokens += u.CacheCreationInputTokens
	target.CacheReadTokens += u.CacheReadInputTokens
	target.OutputTokens += u.OutputTokens
}

// merge folds another origin split into this one.
func (s *tokensOriginSplit) merge(o *tokensOriginSplit) {
	s.Main.addFrom(o.Main)
	s.Subagent.addFrom(o.Subagent)
}

// addFrom folds raw totals into these totals.
func (t *tokenUsageTotals) addFrom(o tokenUsageTotals) {
	t.InputTokens += o.InputTokens
	t.CacheCreationTokens += o.CacheCreationTokens
	t.CacheReadTokens += o.CacheReadTokens
	t.OutputTokens += o.OutputTokens
}

// tokensMessages counts assistant messages by origin.
type tokensMessages struct {
	Assistant int64 `json:"assistant"`
	Sidechain int64 `json:"sidechain"`
}

// TokensRecord is one ledger line of per-pool token accounting (schema v1).
type TokensRecord struct {
	SchemaVersion int                                `json:"schema_version"`
	RecordedAt    string                             `json:"recorded_at"`
	SessionID     string                             `json:"session_id"`
	Card          string                             `json:"card"`
	Role          string                             `json:"role"`
	Cwd           string                             `json:"cwd"`
	Transcript    string                             `json:"transcript"`
	Pools         map[string]*tokensOriginSplit      `json:"pools"`
	Models        map[string]*tokensOriginSplit      `json:"models"`
	Messages      tokensMessages                     `json:"messages"`
	SkippedLines  int                                `json:"skipped_lines"`
	Context       *statusline.SessionTelemetryRecord `json:"context,omitempty"`
}

// transcriptLine mirrors the countable fields of one CC transcript JSONL line.
// The model and usage live under "message" (verified against real transcripts
// under ~/.claude/projects/), not at the top level.
type transcriptLine struct {
	Type        string             `json:"type"`
	IsSidechain bool               `json:"isSidechain"`
	SessionID   string             `json:"sessionId"`
	Message     *transcriptMessage `json:"message"`
}

type transcriptMessage struct {
	Model string              `json:"model"`
	Usage *transcriptUsageMsg `json:"usage"`
}

type transcriptUsageMsg struct {
	InputTokens              int64 `json:"input_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
}

// transcriptAggregates is the result of one transcript scan, before record
// assembly (timestamps, labels, context snapshot).
type transcriptAggregates struct {
	Pools    map[string]*tokensOriginSplit
	Models   map[string]*tokensOriginSplit
	Messages tokensMessages
	Skipped  int
}

// tokenPoolForModel maps a model name to its accounting pool key: glm* → glm,
// claude* → claude, anything else → the lowercased model string itself.
func tokenPoolForModel(model string) string {
	switch {
	case strings.HasPrefix(model, "glm"):
		return "glm"
	case strings.HasPrefix(model, "claude"):
		return "claude"
	default:
		return strings.ToLower(model)
	}
}

// aggregateTranscriptUsage scans a CC session transcript (JSONL, one JSON
// object per line) and aggregates assistant-message usage per pool, per exact
// model, and by origin. Lines that are not assistant messages carrying a
// usage object are ignored; malformed lines (e.g. a truncated trailing write)
// are counted in Skipped and never abort the scan (fail-open).
func aggregateTranscriptUsage(r io.Reader) (*transcriptAggregates, error) {
	agg := &transcriptAggregates{
		Pools:  make(map[string]*tokensOriginSplit),
		Models: make(map[string]*tokensOriginSplit),
	}

	ensure := func(m map[string]*tokensOriginSplit, key string) *tokensOriginSplit {
		if m[key] == nil {
			m[key] = &tokensOriginSplit{}
		}
		return m[key]
	}

	reader := bufio.NewReader(r)
	for {
		line, readErr := reader.ReadString('\n')
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed != "" {
			var tl transcriptLine
			if err := json.Unmarshal([]byte(trimmed), &tl); err != nil {
				agg.Skipped++
			} else if tl.Type == "assistant" && tl.Message != nil &&
				tl.Message.Model != "" && tl.Message.Usage != nil {
				agg.Messages.Assistant++
				if tl.IsSidechain {
					agg.Messages.Sidechain++
				}
				ensure(agg.Pools, tokenPoolForModel(tl.Message.Model)).add(tl.IsSidechain, *tl.Message.Usage)
				ensure(agg.Models, tl.Message.Model).add(tl.IsSidechain, *tl.Message.Usage)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return nil, fmt.Errorf("read transcript: %w", readErr)
		}
	}
	return agg, nil
}

// mergeFrom folds another scan's aggregates into this one (primary transcript
// plus sibling subagent files).
func (a *transcriptAggregates) mergeFrom(src *transcriptAggregates) {
	merge := func(dst, other map[string]*tokensOriginSplit) {
		for k, v := range other {
			if dst[k] == nil {
				dst[k] = &tokensOriginSplit{}
			}
			dst[k].merge(v)
		}
	}
	merge(a.Pools, src.Pools)
	merge(a.Models, src.Models)
	a.Messages.Assistant += src.Messages.Assistant
	a.Messages.Sidechain += src.Messages.Sidechain
	a.Skipped += src.Skipped
}

// subagentTranscriptPaths returns CC's sibling sidechain transcripts for a
// session transcript: <dir>/<session-stem>/subagents/*.jsonl. CC 2.1.23x
// stores subagent sidechains as separate files rather than inlining them in
// the main session transcript (measured: a main file carries 0
// isSidechain:true lines while its subagents/ siblings carry all of them).
// Returns nil when no sibling directory exists (older CC layouts inline the
// sidechain lines, which the per-line isSidechain flag already covers).
func subagentTranscriptPaths(transcriptPath string) []string {
	dir := filepath.Dir(transcriptPath)
	stem := strings.TrimSuffix(filepath.Base(transcriptPath), ".jsonl")
	matches, err := filepath.Glob(filepath.Join(dir, stem, "subagents", "*.jsonl"))
	if err != nil || len(matches) == 0 {
		return nil
	}
	return matches
}

// transcriptResolver maps a session UUID to its transcript path.
type transcriptResolver func(sessionID string) (string, error)

// defaultTranscriptResolver resolves a session UUID to
// ~/.claude/projects/<path-derived-dir>/<uuid>.jsonl via glob.
func defaultTranscriptResolver(sessionID string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	pattern := filepath.Join(home, ".claude", "projects", "*", sessionID+".jsonl")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob transcripts: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no transcript found for session %s under %s", sessionID, filepath.Join(home, ".claude", "projects"))
	}
	return matches[0], nil
}

// tokensRecordOpts carries the record command inputs; Resolver, Stdout, and
// StateDir exist so tests can inject fakes instead of touching $HOME.
type tokensRecordOpts struct {
	Transcript string
	Session    string
	Card       string
	Role       string
	AsJSON     bool
	Resolver   transcriptResolver
	Stdout     io.Writer
	StateDir   string
}

// runTokensRecord resolves the transcript, aggregates usage, embeds the
// context snapshot when present, and either appends one JSON line to the
// ledger (default) or prints the record JSON (AsJSON) without writing.
func runTokensRecord(opts tokensRecordOpts) (*TokensRecord, error) {
	if (opts.Transcript == "") == (opts.Session == "") {
		return nil, fmt.Errorf("exactly one of --transcript / --session is required")
	}

	transcriptPath := opts.Transcript
	sessionID := opts.Session
	if sessionID != "" {
		resolver := opts.Resolver
		if resolver == nil {
			resolver = defaultTranscriptResolver
		}
		resolved, err := resolver(sessionID)
		if err != nil {
			return nil, fmt.Errorf("resolve transcript for session %s: %w", sessionID, err)
		}
		transcriptPath = resolved
	} else {
		sessionID = strings.TrimSuffix(filepath.Base(transcriptPath), ".jsonl")
	}

	f, err := os.Open(transcriptPath)
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer func() { _ = f.Close() }()

	agg, err := aggregateTranscriptUsage(f)
	if err != nil {
		return nil, fmt.Errorf("aggregate transcript %s: %w", transcriptPath, err)
	}
	for _, subPath := range subagentTranscriptPaths(transcriptPath) {
		subFile, err := os.Open(subPath)
		if err != nil {
			return nil, fmt.Errorf("open subagent transcript: %w", err)
		}
		subAgg, err := aggregateTranscriptUsage(subFile)
		_ = subFile.Close()
		if err != nil {
			return nil, fmt.Errorf("aggregate subagent transcript %s: %w", subPath, err)
		}
		agg.mergeFrom(subAgg)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	stateDir := opts.StateDir
	if stateDir == "" {
		stateDir, err = resolveTokensStateDir()
		if err != nil {
			return nil, fmt.Errorf("resolve state dir: %w", err)
		}
	}

	rec := &TokensRecord{
		SchemaVersion: tokensSchemaVersion,
		RecordedAt:    time.Now().UTC().Format(time.RFC3339),
		SessionID:     sessionID,
		Card:          opts.Card,
		Role:          opts.Role,
		Cwd:           cwd,
		Transcript:    transcriptPath,
		Pools:         agg.Pools,
		Models:        agg.Models,
		Messages:      agg.Messages,
		SkippedLines:  agg.Skipped,
		Context:       readTokensContextSnapshot(stateDir),
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return nil, fmt.Errorf("marshal record: %w", err)
	}

	if opts.AsJSON {
		out := opts.Stdout
		if out == nil {
			out = os.Stdout
		}
		if _, err := fmt.Fprintf(out, "%s\n", data); err != nil {
			return nil, fmt.Errorf("print record: %w", err)
		}
		return rec, nil
	}

	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, fmt.Errorf("create state dir: %w", err)
	}
	ledgerPath := filepath.Join(stateDir, tokensLedgerFilename)
	lf, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open ledger: %w", err)
	}
	defer func() { _ = lf.Close() }()
	if _, err := lf.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("append ledger: %w", err)
	}
	return rec, nil
}

// resolveTokensStateDir returns the ledger's .moai/state directory: the shared
// project-state resolution (findStateDir, state.go) when a project is found,
// else <cwd>/.moai/state (created lazily on append) so a fresh checkout records
// without pre-scaffolding.
func resolveTokensStateDir() (string, error) {
	if dir, err := findStateDir(); err == nil {
		announceResolvedRoot(os.Stderr, dir)
		return dir, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	fallback := filepath.Join(normalizeDir(cwd), ".moai", "state")
	announceResolvedRoot(os.Stderr, fallback)
	return fallback, nil
}

// readTokensContextSnapshot embeds the statusline session telemetry record
// when it exists and parses; any failure yields nil (fail-open, never blocks
// the accounting record). The record is read through the statusline package's
// single exported reader (SPEC-SESSION-TELEMETRY-001 REQ-ST-006) rather than
// through a second declaration of the same schema.
func readTokensContextSnapshot(stateDir string) *statusline.SessionTelemetryRecord {
	rec, err := statusline.ReadSessionTelemetry(statusline.SessionTelemetryPath(stateDir))
	if err != nil {
		return nil
	}
	return rec
}

// newTokensCmd creates the root of the tokens command tree.
func newTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tokens",
		Short:   "Record per-pool token usage",
		Long:    "Record per-pool token accounting for a Claude Code session at card or session close",
		GroupID: "tools",
	}
	cmd.AddCommand(newTokensRecordCmd())
	return cmd
}

// newTokensRecordCmd creates the tokens record subcommand.
func newTokensRecordCmd() *cobra.Command {
	var transcript, session, card, role string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "record",
		Short: "Append a token-usage record to the ledger",
		Long: "Aggregate token usage per pool (glm/claude/other) and per origin (main conversation vs subagent sidechain) " +
			"from a Claude Code session transcript, embed the context-usage snapshot when present, and append one JSON " +
			"record to .moai/state/token-accounting.jsonl",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))
			rec, err := runTokensRecord(tokensRecordOpts{
				Transcript: transcript,
				Session:    session,
				Card:       card,
				Role:       role,
				AsJSON:     asJSON,
			})
			if err != nil {
				return err
			}
			if !asJSON {
				stateDir, dirErr := resolveTokensStateDir()
				if dirErr != nil {
					return fmt.Errorf("resolve state dir: %w", dirErr)
				}
				p.Info("Recorded session %s (%d pools, %d assistant messages) to %s",
					rec.SessionID, len(rec.Pools), rec.Messages.Assistant,
					filepath.Join(stateDir, tokensLedgerFilename))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&transcript, "transcript", "", "Path to a CC session transcript (.jsonl)")
	cmd.Flags().StringVar(&session, "session", "", "Session UUID; resolves the transcript under ~/.claude/projects/")
	cmd.Flags().StringVar(&card, "card", "", "Kanban card id label (e.g. t86)")
	cmd.Flags().StringVar(&role, "role", "", "Lane role label (e.g. lead, plan, run, sync)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the record JSON to stdout without writing the ledger")
	cmd.MarkFlagsMutuallyExclusive("transcript", "session")
	cmd.MarkFlagsOneRequired("transcript", "session")

	return cmd
}
