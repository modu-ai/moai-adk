package hook

// shell_tool.go — the one place that names the shell tools.
//
// Claude Code ships two tools that run shell command text: Bash, and the
// PowerShell tool (enabled with CLAUDE_CODE_USE_POWERSHELL_TOOL=1). Both send a
// hook payload whose tool_input carries the command string under "command".
// Every hook branch that inspects shell-command text decides "is this a shell
// call?" through IsShellTool, so a guard written for Bash does not silently
// fall through for the same command sent through PowerShell.
//
// The set is closed on purpose. Adding a tool name here widens every guard at
// once, so it is an operator decision, not a drive-by edit. A source guard
// (hmp_source_guard_test.go) fails when a "bash" literal appears in this
// package or in internal/cli/hook.go outside this file.

const (
	// toolNameBash is the Bash tool's name as it arrives in tool_name.
	toolNameBash = "Bash"
	// toolNamePowerShell is the PowerShell tool's name as it arrives in
	// tool_name. The comparison is exact: the vendor name is case-sensitive.
	toolNamePowerShell = "PowerShell"
)

// IsShellTool reports whether name is a tool that runs shell command text —
// exactly Bash or PowerShell.
//
// @MX:ANCHOR: [AUTO] shell-tool predicate — every guard and evidence branch that inspects command text routes through it
// @MX:REASON: [AUTO] fan_in >= 3 (pre_tool.go guards, post_tool.go evidence, evidence_writer.go, internal/cli/hook.go); widening it widens every shell guard at once
func IsShellTool(name string) bool {
	return name == toolNameBash || name == toolNamePowerShell
}

// isPowerShellTool reports whether name is the PowerShell tool.
func isPowerShellTool(name string) bool {
	return name == toolNamePowerShell
}
