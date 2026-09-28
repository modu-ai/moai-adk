// served_model.go — observation of the model that actually answered a subagent.
//
// The PreToolUse agent-model guard (agent_model_guard.go) compares the model a
// spawn DECLARED with the model the profile RESOLVES. Neither is the model that
// answered: that only exists once the subagent has run, as the
// `.message.model` value of its transcript's assistant rows. A spawn declaring
// `opus` whose every response came from another backend is recorded `ok` by
// the PreToolUse guard. This file reads the answer back.
//
// One classification implementation serves two surfaces: the SubagentStop
// observer (served_model_stop.go) and the read-only `moai doctor` sweep. Both
// call ObserveServedModel, so the two cannot disagree about a transcript.
//
// Every uncertain state reads `unknown`, never `ok`: an absent transcript, no
// usable assistant row, a transcript past the read budget, or no expectation
// to compare against. Evidence absent is not evidence of success.
package hook

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// Served verdicts. `ok` and `unmapped` share their spelling with the
// PreToolUse verdicts on purpose: both rows land in the same audit log.
const (
	ServedVerdictOK       = "ok"
	ServedVerdictDrift    = "served_drift"
	ServedVerdictUnknown  = "unknown"
	ServedVerdictUnmapped = "unmapped"
)

// Causes carried by an `unknown` observation, so a reader can tell why the
// transcript could not be judged.
const (
	servedCauseTranscriptUnreadable = "transcript absent or unreadable"
	servedCauseNoAssistantModel     = "no assistant row carries a model value"
	servedCauseOverBudget           = "transcript exceeds the read budget"
	servedCauseNoExpectation        = "no declared model — the subagent inherits the session model"
)

// syntheticServedModel is the model value the runtime writes on rows it
// synthesized itself; it names no backend and is excluded from the served set.
const syntheticServedModel = "<synthetic>"

// claudeModelFamilyPrefix is the leading hyphen token of a Claude model
// identifier (claude-<family>-<version>). A bare family alias such as the
// resolver's `opus` matches only identifiers carrying this prefix.
const claudeModelFamilyPrefix = "claude"

// servedReadBudget bounds one transcript read. The SubagentStop hook runs under
// a 5-second wrapper timeout, so a transcript past the budget is `unknown`
// rather than a hook that never returns.
type servedReadBudget struct {
	MaxBytes int64
	MaxLines int
}

// defaultServedReadBudget admits the largest transcripts observed in practice
// (single-digit MiB) with an order of magnitude to spare.
var defaultServedReadBudget = servedReadBudget{MaxBytes: 32 << 20, MaxLines: 200_000}

// ServedObservation is the classification of one subagent transcript.
type ServedObservation struct {
	AgentType     string
	DeclaredModel string
	ResolvedModel string
	ExpectedModel string
	ServedModels  []string // sorted, distinct, never carries <synthetic>
	Verdict       string
	Cause         string // non-empty for `unknown`
}

// subagentMeta is the sibling agent-<id>.meta.json the runtime writes.
type subagentMeta struct {
	AgentType string `json:"agentType"`
	Model     string `json:"model"`
}

// SubagentMetaPath returns the meta.json sibling of a subagent transcript.
func SubagentMetaPath(transcriptPath string) string {
	return strings.TrimSuffix(transcriptPath, ".jsonl") + ".meta.json"
}

func readSubagentMeta(transcriptPath string) subagentMeta {
	var m subagentMeta
	data, err := os.ReadFile(SubagentMetaPath(transcriptPath))
	if err != nil {
		return m
	}
	_ = json.Unmarshal(data, &m)
	return m
}

// expectationState reports how expectedServedModel arrived at its answer.
type expectationState int

const (
	expectationKnown  expectationState = iota // expected model is set
	expectationAbsent                         // no declaration — the subagent inherits the session model
)

// expectedServedModel returns the model a subagent's responses are expected
// to come from, plus the recorded expectation basis. With the profile matrix
// gone (SPEC-AGENT-MODEL-INHERIT-001 M5), the expectation is purely
// declaration-based: a declared model is the expectation, and no declaration
// means the subagent inherits the main session's model — which this hook
// cannot see — so there is no expectation and the caller records `unknown`.
// cfg is accepted for signature stability (the @MX:ANCHOR below) and no
// longer participates.
//
// @MX:ANCHOR: [AUTO] the single coupling point between served-model observation and the notion of an expected model
// @MX:REASON: fan_in = SubagentStop observer + doctor served-model sweep via ObserveServedModel; replacing the declared-model expectation (for example with a session-inherited or gate-agent expectation) must change only this function
func expectedServedModel(declared, agentType string, cfg *config.Config) (expected, resolved string, state expectationState) {
	_ = cfg
	_ = agentType
	// A declared "inherit" names the parent session's model, which this hook
	// cannot see, so it is treated like no declaration at all.
	if declared = normalizeModelDeclaration(declared); declared != "" && declared != template.ModelInherit {
		return declared, declared, expectationKnown
	}
	return "", "inherit", expectationAbsent
}

// normalizeModelDeclaration trims whitespace and drops a trailing bracketed
// context-window suffix such as "[1m]": "opus[1m]" and "claude-opus-5[1m]"
// select the same model as their bare forms, and responses never carry the
// suffix, so comparing it verbatim reported every such spawn as drift.
func normalizeModelDeclaration(model string) string {
	model = strings.TrimSpace(model)
	if strings.HasSuffix(model, "]") {
		if i := strings.LastIndexByte(model, '['); i > 0 {
			model = strings.TrimSpace(model[:i])
		}
	}
	return model
}

// servedMatches reports whether one served model satisfies the expectation:
// equal ignoring case, or the expectation is a bare family alias and the served
// value is a Claude model identifier of that family.
func servedMatches(expected, served string) bool {
	if strings.EqualFold(expected, served) {
		return true
	}
	alias := strings.ToLower(strings.TrimSpace(expected))
	if alias == "" || strings.ContainsAny(alias, "-.0123456789") {
		return false
	}
	tokens := strings.Split(strings.ToLower(served), "-")
	if len(tokens) < 2 || tokens[0] != claudeModelFamilyPrefix {
		return false
	}
	return slices.Contains(tokens[1:], alias)
}

// errServedOverBudget reports a transcript read that stopped at the budget.
var errServedOverBudget = errors.New("served-model read budget exceeded")

// assistantMarker pre-filters lines before JSON decoding: a row that does not
// even contain the word cannot be an assistant row.
var assistantMarker = []byte(`"assistant"`)

// readServedModels streams a transcript and returns the distinct served models
// of its assistant rows. Rows that are not JSON are skipped; a line or byte
// count past the budget returns errServedOverBudget.
//
// @MX:NOTE: [AUTO] streaming read — the transcript is never loaded whole; the budget is what keeps the SubagentStop hook inside its wrapper timeout
func readServedModels(path string, budget servedReadBudget) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	seen := map[string]bool{}
	r := bufio.NewReaderSize(f, 64<<10)
	var total int64
	lines := 0
	for {
		line, err := r.ReadBytes('\n')
		total += int64(len(line))
		if len(line) > 0 {
			lines++
		}
		if total > budget.MaxBytes || lines > budget.MaxLines {
			return nil, errServedOverBudget
		}
		if len(line) > 0 && bytes.Contains(line, assistantMarker) {
			var row struct {
				Type    string `json:"type"`
				Message struct {
					Model string `json:"model"`
				} `json:"message"`
			}
			if json.Unmarshal(line, &row) == nil && row.Type == "assistant" {
				m := strings.TrimSpace(row.Message.Model)
				if m != "" && m != syntheticServedModel {
					seen[m] = true
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}
	out := make([]string, 0, len(seen))
	for m := range seen {
		out = append(out, m)
	}
	sort.Strings(out)
	return out, nil
}

// ObserveServedModel classifies one subagent transcript with the default read
// budget. agentTypeHint is the hook input's agent type; when empty, the
// sibling meta.json supplies it. cfg nil means no configuration was reachable.
func ObserveServedModel(transcriptPath, agentTypeHint string, cfg *config.Config) ServedObservation {
	return observeServedModel(transcriptPath, agentTypeHint, cfg, defaultServedReadBudget)
}

func observeServedModel(transcriptPath, agentTypeHint string, cfg *config.Config, budget servedReadBudget) ServedObservation {
	meta := readSubagentMeta(transcriptPath)
	obs := ServedObservation{AgentType: agentTypeHint, DeclaredModel: strings.TrimSpace(meta.Model)}
	if obs.AgentType == "" {
		obs.AgentType = meta.AgentType
	}

	expected, resolved, state := expectedServedModel(obs.DeclaredModel, obs.AgentType, cfg)
	obs.ExpectedModel, obs.ResolvedModel = expected, resolved

	served, err := readServedModels(transcriptPath, budget)
	switch {
	case errors.Is(err, errServedOverBudget):
		return unknownObservation(obs, servedCauseOverBudget)
	case err != nil:
		return unknownObservation(obs, servedCauseTranscriptUnreadable)
	}
	obs.ServedModels = served
	if len(served) == 0 {
		return unknownObservation(obs, servedCauseNoAssistantModel)
	}

	switch state {
	case expectationAbsent:
		return unknownObservation(obs, servedCauseNoExpectation)
	}
	obs.Verdict = ServedVerdictOK
	for _, m := range served {
		if !servedMatches(expected, m) {
			obs.Verdict = ServedVerdictDrift
			break
		}
	}
	return obs
}

func unknownObservation(obs ServedObservation, cause string) ServedObservation {
	obs.Verdict = ServedVerdictUnknown
	obs.Cause = cause
	return obs
}
