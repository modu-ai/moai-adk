package gateway

import (
	"errors"
	"strings"

	"github.com/modu-ai/moai-adk/internal/codexbridge"
)

// appServerScopeReason emits fixed diagnostic codes, never exception text.
// Public responses retain the stable appserver_scope_mismatch contract.
func appServerScopeReason(err error) string {
	if !errors.Is(err, codexbridge.ErrScope) {
		return ""
	}
	for cause := err; cause != nil; cause = errors.Unwrap(cause) {
		switch strings.TrimPrefix(cause.Error(), codexbridge.ErrScope.Error()+": ") {
		case "active model mismatch":
			return "active_model_mismatch"
		case "working directory mismatch":
			return "working_directory_mismatch"
		case "history prefix mismatch":
			return "history_prefix_mismatch"
		case "new input arrived while tools are pending":
			return "input_while_tools_pending"
		case "tool result count mismatch":
			return "tool_result_count_mismatch"
		case "unknown or duplicate tool result":
			return "unknown_or_duplicate_tool_result"
		case "unsupported tool result content":
			return "unsupported_tool_result_content"
		case "empty tool result image":
			return "empty_tool_result_image"
		case "unexpected tool result":
			return "unexpected_tool_result"
		case "empty turn input":
			return "empty_turn_input"
		case "prepare owner binding":
			return "prepare_owner_binding"
		case "prepare owner length":
			return "prepare_owner_length"
		case "prepare active model or cwd":
			return "prepare_active_model_or_cwd"
		case "prepare public delta":
			return "prepare_public_delta"
		}
	}
	return "unclassified_scope"
}
