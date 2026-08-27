package cli

// codex_init.go — SPEC-CODEX-INIT-001 M1+M2 (REQ-CI-001..004, REQ-CI-009,
// REQ-CI-010): the init-offer gate both launch verbs pass through right
// before launching. The gate's ONLY input is the wiring state the launcher's
// classifier returns, consumed through a seam — it never re-derives the
// judgement from disk (AC-CI-002), and its decision section performs zero
// filesystem calls of its own (AC-CI-001).
//
// Flow (plan §C.2): wired → return (the caller launches); incomplete →
// report state+remedy; prompt-incapable → report and exit non-success
// WITHOUT issuing any prompt (REQ-CI-009); declined → write nothing and exit
// with the cancel code (REQ-CI-003); accepted → invoke the existing
// generator exactly once with the codex agent selection (REQ-CI-004), then
// secure the instruction contract; a failure at either step stops the launch
// and no contract step runs after a failed generator (REQ-CI-010).
//
// The gate is deliberately spawn-blind: it takes no --spawn parameter, so a
// spawn-only bypass of the gate cannot exist (REQ-CI-002 — one function, all
// four verb×spawn combinations).

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/spf13/cobra"
)

// codexInitDeclinedExitCode is the CANCEL exit code — distinct from every
// error code: the operator was offered the remedy and declined (AC-CI-003's
// "a value that means cancellation, not error"). 130 follows the 128+n
// operator-abort convention; the tests compare against this constant, not
// the literal.
const codexInitDeclinedExitCode = 130

// codexInitFailureExitCode is the non-success code for the
// non-interactive refusal and the generator/contract failures.
const codexInitFailureExitCode = 1

// codexGeneratorAgentCodex is the agent selection the gate passes to the
// generator: exactly the codex selection (REQ-CI-004 — not "both", which
// would also lay the unrequested claude wiring).
const codexGeneratorAgentCodex = "codex"

// codexGatePrintf emits a best-effort diagnostic. A failed write on a
// refusal report must never mask the refusal it accompanies, so the write
// error is observed and deliberately dropped.
func codexGatePrintf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		return
	}
}

// codexGateSeams — every judgement, prompt, and write the gate performs
// travels through one of these seams, which is what lets the acceptance
// drive every path without the codex binary or a real prompt.
var (
	// codexWiringClassifierFn is the classifier seam. The default delegates
	// to the launcher's single wiring judgement (REQ-CL-006) — the gate
	// CONSUMES it and never re-implements it. AC-CI-002 injects
	// contradictory states here and holds the gate to the returned value.
	codexWiringClassifierFn = func(projectRoot string) codexWiringInfo {
		return classifyCodexWiring(projectRoot)
	}

	// codexPromptCapableFn reports whether a prompt can be answered at all.
	// Its return value is the ONLY source of the interactive/non-interactive
	// decision — the gate never queries stdin's kind itself (AC-CI-004's
	// contradiction injection relies on this).
	codexPromptCapableFn = defaultCodexPromptCapable

	// codexOfferPromptFn issues the initialization offer and reports the
	// operator's answer. The default reads one line from in; the tests
	// replace it to count issuance (exactly 1 when capable, 0 when not) and
	// to fix the answer.
	codexOfferPromptFn = defaultCodexOfferPrompt

	// codexInitGeneratorFn is the generator seam: the gate invokes the
	// EXISTING generator through it exactly once on acceptance, with the
	// codex agent selection captured in the request (AC-CI-004).
	codexInitGeneratorFn = defaultCodexInitGenerator

	// codexContractFn is the instruction-contract seam. The gate calls it
	// AFTER a successful generator (AC-CI-010 counts zero calls on a
	// generator failure); the request is captured so the spawn-pair
	// comparison can prove the contract step is spawn-invariant.
	codexContractFn = secureCodexInstructionContract
)

// codexGeneratorRequest is the captured generator invocation: the project
// root, the agent selection (exactly codex), and the output streams.
type codexGeneratorRequest struct {
	ProjectRoot string
	Agent       string
	Out         io.Writer
	Err         io.Writer
}

// codexContractRequest is the captured contract invocation (AC-CI-004's
// pair comparison reads these).
type codexContractRequest struct {
	ProjectRoot string
	Out         io.Writer
	Err         io.Writer
}

// defaultCodexPromptCapable reports whether stdin can carry an interactive
// answer. It lives INSIDE the seam so the gate itself stays stdin-blind.
func defaultCodexPromptCapable() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// defaultCodexOfferPrompt prints the offer naming the state and the remedy,
// then reads a one-line answer. Only y/yes (any case) accepts; anything else
// — including EOF — declines.
func defaultCodexOfferPrompt(out io.Writer, in io.Reader, info codexWiringInfo) bool {
	codexGatePrintf(out, "codex wiring is %s %s — initialize now with `%s`? [y/N] ",
		info.Status, info.Detail, codexWiringAction)
	line, _ := bufio.NewReader(in).ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

// defaultCodexInitGenerator delegates to the existing wiring generator —
// the same one `moai init --agent codex` calls. The gate adds no path of
// its own that writes a wiring file (REQ-CI-004).
func defaultCodexInitGenerator(req codexGeneratorRequest) error {
	if _, err := codexwiring.Wire(req.ProjectRoot, req.Out, req.Err); err != nil {
		return fmt.Errorf("codex wiring generator: %w", err)
	}
	return nil
}

// codexGateReport names the incomplete state and its remedy — every
// incomplete gate outcome prints this line (AC-CI-003/004 output duty).
func codexGateReport(w io.Writer, info codexWiringInfo) {
	codexGatePrintf(w, "codex wiring is %s %s — %s\n", info.Status, info.Detail, codexWiringAction)
}

// @MX:NOTE: [AUTO] single judgement source: the gate acts ONLY on the
// classifier's returned state; wired alone passes, every other state is
// incomplete (REQ-CI-001 — calling and re-deciding is the defect AC-CI-002
// kills).
func codexInitOfferGate(cmd *cobra.Command, projectRoot string) error {
	info := codexWiringClassifierFn(projectRoot)
	if info.Status == codexWiringStatusWired {
		return nil
	}

	errW := cmd.ErrOrStderr()
	codexGateReport(errW, info)

	if !codexPromptCapableFn() {
		// Non-interactive: report and stop, issuing NO prompt at all — a
		// prompt nobody can answer hangs the automation (REQ-CI-009).
		codexGatePrintf(errW, "non-interactive session — nothing launched; %s\n", codexWiringAction)
		return &exitCodeError{code: codexInitFailureExitCode}
	}

	if !codexOfferPromptFn(cmd.OutOrStdout(), os.Stdin, info) {
		// Declined: write nothing, launch nothing, cancel exit code.
		codexGatePrintf(errW, "declined — nothing written; %s\n", codexWiringAction)
		return &exitCodeError{code: codexInitDeclinedExitCode}
	}

	if err := codexInitGeneratorFn(codexGeneratorRequest{
		ProjectRoot: projectRoot,
		Agent:       codexGeneratorAgentCodex,
		Out:         cmd.OutOrStdout(),
		Err:         errW,
	}); err != nil {
		// Generator failure stops everything: no contract step, no launch
		// (REQ-CI-010 — a launch here delivers the unwired session).
		codexGatePrintf(errW, "codex init failed: %v — nothing launched; %s\n", err, codexWiringAction)
		return &exitCodeError{code: codexInitFailureExitCode}
	}

	if err := codexContractFn(codexContractRequest{
		ProjectRoot: projectRoot,
		Out:         cmd.OutOrStdout(),
		Err:         errW,
	}); err != nil {
		codexGatePrintf(errW, "codex init failed: %v — nothing launched; %s\n", err, codexWiringAction)
		return &exitCodeError{code: codexInitFailureExitCode}
	}
	return nil
}
