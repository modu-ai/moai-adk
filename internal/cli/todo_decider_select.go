// todo_decider_select.go — SPEC-TCD-LLM-DECIDER-001 M1: the standing-decider
// selector (REQ-TLD-002).
//
// The add surfaces — the CLI `todo add` path and the MCP `todo_add` tool,
// which share runTodoAddAppendRoot — resolve their judgment backend through
// this selector, read ONCE per add invocation. Unset/empty/"default" keep
// the deterministic default decider; "llm" selects the LLM-backed decider;
// ANY other value is a usage refusal (exit 2) with nothing written — an
// operator-authored misconfiguration fails loud rather than silently
// reverting to the default (the EnvClaudeBin precedent, plan OD-A). The
// --classification-file flag outranks this selection (REQ-TLD-002): where
// both are supplied, the file's judgment wins and the LLM decider is never
// invoked.
package cli

import (
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoDeciderFromEnv resolves the standing decider from the
// MOAI_TODO_DECIDER selector value (config.EnvTodoDecider). The accepted
// set is exactly {default, llm} — the jev identities are refused with the
// parent SPEC's named refusal wording, and any other value with a usage
// refusal naming the accepted set. The accepted-set spellings reuse the
// classification constants (C3: one spelling, never re-literalized).
func todoDeciderFromEnv() (kanban.CardDecider, error) {
	switch v := os.Getenv(config.EnvTodoDecider); v {
	case "", kanban.DeciderIdentityDefault:
		return todoCardDecider, nil
	case kanban.DeciderIdentityLLM:
		return newLLMCardDecider(), nil
	case kanban.DeciderIdentityJev, kanban.DeciderIdentityLLMJev:
		return nil, &exitCodeError{
			code: 2,
			msg: fmt.Sprintf("todo add: %s value %q refused: jev is never a product classification decider (accepted values: %s, %s; nothing written)",
				config.EnvTodoDecider, v, kanban.DeciderIdentityDefault, kanban.DeciderIdentityLLM),
		}
	default:
		return nil, &exitCodeError{
			code: 2,
			msg: fmt.Sprintf("todo add: %s value %q refused: accepted values are %s, %s (nothing written)",
				config.EnvTodoDecider, v, kanban.DeciderIdentityDefault, kanban.DeciderIdentityLLM),
		}
	}
}
