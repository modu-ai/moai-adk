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

// unsetLaneEnvForCodexHook removes the eleven lane launch keys from this hook
// process and returns the function that restores them. A hook running under
// --harness codex is a Codex session's hook, never a Claude lane's, so it must
// not register, bind, or rotate a factory peer in whatever lane environment
// it inherited (SPEC-CODEX-FACTORY-RETIRE-001 REQ-CFR-022). The hook package
// has no harness field; this boundary is where the harness is known.
func unsetLaneEnvForCodexHook() func() {
	restores := make([]func(), 0, len(codexLaneLaunchEnvKeys))
	for _, key := range codexLaneLaunchEnvKeys {
		restores = append(restores, captureEnvState(key))
		_ = os.Unsetenv(key)
	}
	return func() {
		for i := len(restores) - 1; i >= 0; i-- {
			restores[i]()
		}
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
	if isPermissionRequestDeny(event, output) {
		return writeCodexPermissionRequestDeny(output)
	}
	var raw bytes.Buffer
	if err := deps.HookProtocol.WriteOutput(&raw, output); err != nil {
		err = fmt.Errorf("serialize hook output for codex mapping: %w", err)
		if codexadapter.IsDecisionBearing(event) && (output == nil || output.ExitCode != 2) {
			return writeCodexFailClosed(event, err)
		}
		return err
	}
	mapped, discards, err := codexadapter.MapOutput(event, raw.Bytes())
	if err != nil {
		err = fmt.Errorf("map hook output for codex: %w", err)
		// Unparseable output on a decision-bearing event is a fault, answered
		// fail-closed (SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2c, REQ-HPR-009).
		// Under exit 2 the exit code already carries the block.
		if codexadapter.IsDecisionBearing(event) && (output == nil || output.ExitCode != 2) {
			return writeCodexFailClosed(event, err)
		}
		return err
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

// isPermissionRequestDeny reports whether output is the Claude-shape
// PermissionRequest deny (hookSpecificOutput.decision.behavior "deny"). The
// Claude schema has no reason slot there, so the handler puts its reason in
// systemMessage — a key Codex ignores — and the deny would reach Codex without
// a reason (SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2e, REQ-HPR-011).
func isPermissionRequestDeny(event hook.EventType, output *hook.HookOutput) bool {
	return event == hook.EventPermissionRequest && output != nil && output.ExitCode != 2 &&
		output.HookSpecificOutput != nil && output.HookSpecificOutput.Decision != nil &&
		output.HookSpecificOutput.Decision.Behavior == "deny"
}

// writeCodexPermissionRequestDeny renders a PermissionRequest deny through the
// translation table, which carries the handler's reason in decision.message,
// so the deny reaches Codex as a deny with a non-empty reason (AC-HPR-010).
func writeCodexPermissionRequestDeny(output *hook.HookOutput) error {
	rendered, discards, err := codexadapter.TranslateCodex(hook.EventPermissionRequest, codexadapter.DecisionDeny, output.SystemMessage)
	if err != nil {
		return writeCodexFailClosed(hook.EventPermissionRequest, fmt.Errorf("render PermissionRequest deny: %w", err))
	}
	if err := codexadapter.RecordDiscards(resolveHookProjectRoot(), discards, false, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "codex harness: record discards: %v\n", err)
	}
	if _, err := os.Stdout.Write(append(rendered, '\n')); err != nil {
		return fmt.Errorf("write mapped hook output: %w", err)
	}
	return nil
}
