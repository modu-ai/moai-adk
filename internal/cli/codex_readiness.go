package cli

// codex_readiness.go — SPEC-CODEX-LAUNCHER-001 M2 (REQ-CL-004/005/006/007):
// the readiness-readout assembly. CODEX_HOME interpretation, project wiring
// judgement, and the fixed six-row readout (plan §C.4) live here as ONE
// struct; the binary/version/auth rows come from the SHARED probe
// (ProbeCodexSetup) so no second auth-classification or wiring-validation
// path is forked (REQ-CL-007).
//
// The readout is bounded (§E): exactly six rows, one line each, and every
// probe failure degrades to that row's unknown token while the remaining rows
// still render (fail-open). The launcher never writes — nowhere here is a
// directory created or a file mutated (REQ-CL-013).

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// Row labels — the closed label set of AC-CL-011.
const (
	codexRowLabelCodex   = "codex"
	codexRowLabelHome    = "home"
	codexRowLabelAuth    = "auth"
	codexRowLabelWiring  = "wiring"
	codexRowLabelAgents  = "agents"
	codexRowLabelHarness = "harness"
)

// codexRowLabelWidth pads every label to a common width so the value column
// lines up. "harness" is the longest label.
const codexRowLabelWidth = len(codexRowLabelHarness)

// Wiring status tokens — the closed status set of AC-CL-004.
const (
	codexWiringStatusNotWired = "not wired"
	codexWiringStatusPartial  = "partial"
	codexWiringStatusWired    = "wired"
	codexWiringStatusInvalid  = "invalid"
)

// codexWiringAction is the remediation phrase REQ-CL-006 requires on EVERY
// incomplete wiring state (absent, empty, either partial, invalid) and forbids
// on the wired state.
const codexWiringAction = "run moai init --agent codex"

// codexHomeMissingAction is the remediation phrase when CODEX_HOME points at a
// path that does not exist (plan §C.3). The launcher reports the gap; logging
// in is what creates the directory, never the readout.
const codexHomeMissingAction = "run codex login"

// codexBinaryNotFound is the binary-row token when the codex binary is absent
// (AC-CL-011): the readout form succeeds precisely where the launch verbs
// refuse, so the diagnostic surface must work when the thing it diagnoses is
// missing (REQ-CL-012).
const codexBinaryNotFound = "not found"

// codexHarnessCommand is the fixed last row — the runtime dispatcher entry the
// whole wiring exists to serve (spec §A.4).
const codexHarnessCommand = "moai hook --harness codex"

// codexAgentsRelDir is the generated agent-TOML directory inside a project
// root (spec §A.4: 11 templates ship at .codex/agents/moai/*.toml).
const codexAgentsRelDir = ".codex/agents/moai"

// codexHomeInfo is the CODEX_HOME interpretation (REQ-CL-005): the resolved
// value, which source supplied it, and whether the path currently exists.
type codexHomeInfo struct {
	Path   string
	Source string // codexHomeSourceEnv | codexHomeSourceDefault
	Exists bool
}

// codexWiringInfo is the project wiring judgement (REQ-CL-006). Status is the
// closed set above; Detail names the files that decided it.
type codexWiringInfo struct {
	Status string
	Detail string
}

// codexReadiness is the assembled readout: one field per dynamic row of the
// fixed six (plan §C.4). The harness row is a constant and carries no field.
type codexReadiness struct {
	BinaryFound  bool
	BinaryPath   string
	Version      string
	Home         codexHomeInfo
	AuthProvider string
	Wiring       codexWiringInfo
	AgentsTOMLs  int
}

// codexSetupProbe is the shared-probe seam (REQ-CL-007): the readout consumes
// ProbeCodexSetup through this var, so tests can drive the whole probe with
// sentinel values (AC-CL-007) without forking the classification path.
var codexSetupProbe = ProbeCodexSetup

// codexValidateHooksConfig is the whitelist-validation seam. The wiring
// judgement calls the adapter's measured whitelist (REQ-CL-007 — no local
// reimplementation); the var exists so a test can stub a sentinel violation
// and watch it surface in the wiring row (AC-CL-007).
var codexValidateHooksConfig = codexadapter.ValidateConfig

// probeCodexReadiness assembles the readout from the shared probe plus the
// three launcher-owned probes (home, wiring, agents). Every probe degrades on
// its own — a failed binary lookup never suppresses the wiring row.
func probeCodexReadiness(ctx context.Context) codexReadiness {
	s := codexSetupProbe(ctx)
	root := projectDirResolver()
	return codexReadiness{
		BinaryFound:  s.Installed,
		BinaryPath:   s.Binary,
		Version:      s.Version,
		Home:         resolveCodexHomeInfo(),
		AuthProvider: s.AuthProvider,
		Wiring:       classifyCodexWiring(root),
		AgentsTOMLs:  countCodexAgentTOMLs(root),
	}
}

// resolveCodexHomeInfo interprets CODEX_HOME (resolveCodexHomeDir) and checks
// existence WITHOUT creating anything (REQ-CL-013). An unresolvable home
// stays Path="" / Exists=false and the row renders the gap.
func resolveCodexHomeInfo() codexHomeInfo {
	path, source := resolveCodexHomeDir()
	info := codexHomeInfo{Path: path, Source: source}
	if path == "" {
		return info
	}
	if _, err := os.Stat(path); err == nil {
		info.Exists = true
	}
	return info
}

// classifyCodexWiring judges the project's .codex wiring by the FILE SET
// (spec §A.4): an absent or empty .codex/ is not wired — directory existence
// alone would report an empty scaffold as wired — one file is partial naming
// the missing one, and present hooks.json must additionally clear the
// adapter's measured whitelist to count as wired.
//
// @MX:NOTE: [AUTO] this is the single definition of codex wiring completeness
// (REQ-CL-006) — SPEC-CODEX-INIT-001 consumes this judgement rather than
// restating it.
func classifyCodexWiring(projectRoot string) codexWiringInfo {
	hooksPath := filepath.Join(projectRoot, filepath.FromSlash(codexwiring.HooksRelPath))
	cfgPath := filepath.Join(projectRoot, filepath.FromSlash(codexwiring.ConfigRelPath))
	hooksExists := codexPathExists(hooksPath)
	cfgExists := codexPathExists(cfgPath)

	switch {
	case !hooksExists && !cfgExists:
		return codexWiringInfo{
			Status: codexWiringStatusNotWired,
			Detail: "(" + codexwiring.HooksRelPath + " and " + codexwiring.ConfigRelPath + " absent)",
		}
	case hooksExists && !cfgExists:
		return codexWiringInfo{
			Status: codexWiringStatusPartial,
			Detail: "(" + codexwiring.ConfigRelPath + " missing)",
		}
	case !hooksExists && cfgExists:
		return codexWiringInfo{
			Status: codexWiringStatusPartial,
			Detail: "(" + codexwiring.HooksRelPath + " missing)",
		}
	}

	// Both files present: hooks.json must clear the adapter's whitelist.
	raw, err := os.ReadFile(hooksPath) //nolint:gosec // fixed wiring path under the project root
	if err != nil {
		return codexWiringInfo{Status: codexWiringStatusInvalid, Detail: "(hooks.json unreadable)"}
	}
	violations, verr := codexValidateHooksConfig(raw)
	if verr != nil {
		return codexWiringInfo{Status: codexWiringStatusInvalid, Detail: "(hooks.json unparseable)"}
	}
	if len(violations) > 0 {
		names := make([]string, len(violations))
		for i, v := range violations {
			names[i] = v.Error()
		}
		return codexWiringInfo{
			Status: codexWiringStatusInvalid,
			Detail: "(hooks.json whitelist: " + strings.Join(names, "; ") + ")",
		}
	}
	return codexWiringInfo{
		Status: codexWiringStatusWired,
		Detail: "(" + codexwiring.HooksRelPath + ", " + codexwiring.ConfigRelPath + ")",
	}
}

// codexPathExists reports path presence; any stat error counts as absent for
// readout purposes (fail-open to the simpler state).
func codexPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// countCodexAgentTOMLs counts the generated agent TOMLs under
// .codex/agents/moai/. Read-only; an unreadable directory is a count of 0.
func countCodexAgentTOMLs(projectRoot string) int {
	matches, err := filepath.Glob(filepath.Join(projectRoot, filepath.FromSlash(codexAgentsRelDir), "*.toml"))
	if err != nil {
		return 0
	}
	return len(matches)
}

// codexWiringRowValue composes the wiring row's value: status, the deciding
// detail, and — for every state except wired — the remediation action.
func codexWiringRowValue(w codexWiringInfo) string {
	value := w.Status + " " + w.Detail
	if w.Status != codexWiringStatusWired {
		value += " — " + codexWiringAction
	}
	return value
}

// rows renders the fixed six rows (plan §C.4): label padded to the common
// width, then the value. Exactly one line per row; a failed probe degrades to
// its row's token rather than dropping the row.
func (r codexReadiness) rows() [6]string {
	var binaryValue string
	switch {
	case !r.BinaryFound:
		binaryValue = codexBinaryNotFound
	case r.Version != "":
		binaryValue = r.Version + " (" + r.BinaryPath + ")"
	default:
		binaryValue = "(" + r.BinaryPath + ")"
	}

	homeValue := r.Home.Path + " (" + r.Home.Source + ")"
	if !r.Home.Exists {
		homeValue += ", missing — " + codexHomeMissingAction
	}

	format := "%-" + strconv.Itoa(codexRowLabelWidth) + "s  %s"
	return [6]string{
		fmt.Sprintf(format, codexRowLabelCodex, binaryValue),
		fmt.Sprintf(format, codexRowLabelHome, homeValue),
		fmt.Sprintf(format, codexRowLabelAuth, r.AuthProvider),
		fmt.Sprintf(format, codexRowLabelWiring, codexWiringRowValue(r.Wiring)),
		fmt.Sprintf(format, codexRowLabelAgents, strconv.Itoa(r.AgentsTOMLs)+" TOML"),
		fmt.Sprintf(format, codexRowLabelHarness, codexHarnessCommand),
	}
}
