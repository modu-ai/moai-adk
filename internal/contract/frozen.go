package contract

// FrozenInstructionFiles is this package's copy of the frozen instruction
// basenames guarded by the PreToolUse hook (frozenInstructionFiles in
// internal/hook/pre_tool.go). The verification core may not import
// internal/hook, so it carries the copy; a test inside internal/hook pins the
// two lists together. The `frozen-files` invariant emits each basename as the
// glob `**/<basename>` so the basename-anywhere semantics survive (design.md
// § Frozen Files).
var FrozenInstructionFiles = []string{"CLAUDE.md", "CLAUDE.local.md"}

// Invariant tokens (design.md § Contract Schema, invariants).
const (
	// InvariantFrozenFiles denotes the frozen-files union.
	InvariantFrozenFiles = "frozen-files"
	// InvariantConstitutionPrefix introduces a glob over registry rule IDs.
	InvariantConstitutionPrefix = "constitution:"
)
