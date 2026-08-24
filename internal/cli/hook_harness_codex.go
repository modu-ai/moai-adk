package cli

// hook_harness_codex.go — SPEC-CODEX-WIRING-001 M3, the `--harness codex`
// runtime mode of the `moai hook` dispatcher (REQ-CW-007).
//
// The seam lives HERE, in the CLI dispatcher layer — in front of and behind
// the dispatcher's own decision logic — and nothing under internal/hook is
// modified (M3 REQ-7 spirit): MapOutput rewrites the serialized output,
// RecordDiscards persists undeliverables, Resolve cross-checks the payload's
// event name against the invoked subcommand, and the exit code / stderr pass
// through untouched.

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// codexHarnessFlagValue is the single non-default --harness value.
const codexHarnessFlagValue = "codex"

// harnessModeIsCodex reads the --harness flag. Empty means claude (the
// default — flag-absent behavior is byte-identical to today); any value other
// than claude/codex fails loud with the valid set named.
func harnessModeIsCodex(cmd *cobra.Command) (bool, error) {
	switch v := getStringFlag(cmd, "harness"); v {
	case "":
		return false, nil
	case codexHarnessFlagValue:
		return true, nil
	case "claude":
		return false, nil
	default:
		return false, fmt.Errorf("invalid --harness value %q: must be one of: claude, codex", v)
	}
}

// validateCodexHarnessEvent cross-checks the payload's hook_event_name
// against the invoked subcommand via codexadapter.Resolve (REQ-CW-007 second
// clause): the hooks.json the generator emits and the runtime command Codex
// runs must agree, and a mismatch is refused with a diagnostic rather than
// dispatched into the wrong handler.
func validateCodexHarnessEvent(event hook.EventType, input *hook.HookInput) error {
	if input == nil || input.HookEventName == "" {
		// Nothing to cross-check — the dispatcher's own injection (subcommand
		// event) fills an absent name, which by construction matches.
		return nil
	}
	payloadArg, err := codexadapter.Resolve(input.HookEventName)
	if err != nil {
		return fmt.Errorf("codex harness: rejecting payload: %w", err)
	}
	subArg, err := codexadapter.Resolve(string(event))
	if err != nil {
		return fmt.Errorf("codex harness: refusing this subcommand: %w", err)
	}
	if payloadArg != subArg {
		return fmt.Errorf("codex harness: payload hook_event_name %q maps to dispatcher %q but this subcommand dispatches %q — the wiring table and the runtime command disagree",
			input.HookEventName, payloadArg, subArg)
	}
	return nil
}

// writeHookOutputCodex maps one hook output through the codex adapter
// (REQ-CW-007 first clause): continue:false becomes decision:block (reason
// filled), UserPromptSubmit systemMessage routes to additionalContext,
// undeliverables are recorded to the adapter's diagnostic sink, and the
// mapped bytes go to stdout. The hook's own exit code and stderr are not
// touched here — the exit-2 path stays in runHookEvent, and hook stderr was
// already written by the handlers themselves.
func writeHookOutputCodex(event hook.EventType, output *hook.HookOutput) error {
	var raw bytes.Buffer
	if err := deps.HookProtocol.WriteOutput(&raw, output); err != nil {
		return fmt.Errorf("serialize hook output for codex mapping: %w", err)
	}
	mapped, discards, err := codexadapter.MapOutput(event, raw.Bytes())
	if err != nil {
		return fmt.Errorf("map hook output for codex: %w", err)
	}

	// hookBlocked mirrors RecordDiscards' own contract: when the underlying
	// hook exited 2, stderr carries the blocking reason and must not gain a
	// diagnostic line — but the sink record is still written.
	hookBlocked := output != nil && output.ExitCode == 2
	if err := codexadapter.RecordDiscards(resolveHookProjectRoot(), discards, hookBlocked, os.Stderr); err != nil {
		// The sink record is the durable half of the no-silence obligation,
		// but losing the console copy must not lose the hook's own output.
		_, _ = fmt.Fprintf(os.Stderr, "codex harness: record discards: %v\n", err)
	}

	if _, err := os.Stdout.Write(append(mapped, '\n')); err != nil {
		return fmt.Errorf("write mapped hook output: %w", err)
	}
	return nil
}
