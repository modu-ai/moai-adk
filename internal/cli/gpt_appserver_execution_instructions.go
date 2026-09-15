package cli

import "github.com/modu-ai/moai-adk/internal/codextools"

func managedGPTExecutionInstructions(instructions string, tools []codextools.Definition) string {
	if len(tools) != 0 {
		instructions += "\n\nPermission boundary: the App Server native filesystem sandbox is distinct from the external Claude Code tools listed for this request. Claude Code permissions are enforced by the host when those tools run. Do not infer that a listed Write, Edit, Bash, or EnterWorktree tool is forbidden from the native sandbox policy alone. For an authorized user task, invoke only tools present in this request and respect the actual tool result or host approval decision. If a needed tool is absent, ask for the required capability rather than inventing it. Never bypass a denial, approval requirement, or native sandbox restriction."
	}
	for _, tool := range tools {
		if tool.Name == "Agent" {
			return instructions + "\n\nClaude Code execution boundary: Agent tasks may run asynchronously. After useful independent work, provide user-visible progress and end your response so completion notifications can drive the next turn. Do not use native sleep, setTimeout loops, or native agent-wait tools merely to poll Claude Agent tasks. This does not prohibit waiting when the user explicitly requests waiting, or using the matching wait tool to receive a yielded execution result. Never report an agent task as completed before its result is received."
		}
	}
	return instructions
}
