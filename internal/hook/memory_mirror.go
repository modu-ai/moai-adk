// memory_mirror.go — write-time agent-memory mirror entry for the
// PostToolUse hook (SPEC-AGENT-MEMORY-DRAIN-001 M2).
//
// The reconciliation core lives in agentmemory.go; this file is only the
// hook-facing wrapper: extract the Write/Edit target path, mirror it, and
// turn every failure into a stderr notice. The hook's observation-only
// contract is preserved — a mirror failure never blocks or errors the
// originating tool call (REQ-AM-005).
package hook

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// mirrorAgentMemory mirrors the just-written agent-memory file of a
// Write/Edit tool call into the primary checkout's store. Observation-shaped:
// it returns nothing, exits nothing, blocks nothing. Every failure —
// unparseable input, unresolvable primary, copy error — is a [memory-mirror]
// stderr notice, and the host tool call proceeds untouched.
//
// @MX:NOTE: [AUTO] SPEC-AGENT-MEMORY-DRAIN-001 M2 — write-time mirror wrapper; the strict path anchor lives in agentmemory.go (plan D7).
func mirrorAgentMemory(input *HookInput) {
	var parsed map[string]any
	if err := json.Unmarshal(input.ToolInput, &parsed); err != nil {
		return
	}
	filePath, ok := parsed["file_path"].(string)
	if !ok || filePath == "" {
		return
	}

	mirrored, err := MirrorAgentMemoryFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[memory-mirror] %v (fail-open; the write itself succeeded)\n", err)
		return
	}
	if mirrored {
		slog.Debug("memory mirror: copied to primary store", "path", filePath)
	}
}
