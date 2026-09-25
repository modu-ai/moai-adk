// codex_audit_derive_test.go — the single derivation of the audit route
// fields (acceptance §A "경로 필드의 유도"). The deterministic criterion
// (TestCodexAuditEvidenceDerivation) and the LIVE evidence writer
// (TestCodexRoleLiveLoadAndReadOnly) both call deriveCodexAuditEvidence, so
// the LIVE evidence never carries a hand-written route field.
package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// codexAuditEvidenceInput is everything one audit item left behind.
type codexAuditEvidenceInput struct {
	Role          string            // the audit role the item ran
	Rollouts      [][]byte          // every session record created during the item
	LaunchRecord  *codexAuditRecord // the launcher's record for the item, if any
	VerdictSHA256 string            // sha256 of the verdict file the item produced
	ProbeCommand  string            // the exact shell command the role was told to run
	ProbeExists   bool              // whether the probe file exists after the item
}

// codexAuditEvidence is the derived field set.
type codexAuditEvidence struct {
	SessionSandbox       string `json:"session_sandbox"`
	TopLevel             bool   `json:"top_level"`
	UsedSpawnAgent       bool   `json:"used_spawn_agent"`
	Route                string `json:"route"`
	VerdictWriter        string `json:"verdict_writer"`
	ProbeCommandExecuted bool   `json:"probe_command_executed"`
	ProbeExitCode        *int   `json:"probe_exit_code"`
	WriteDenied          bool   `json:"write_denied"`
}

// derivedRollout is the part of one session record the derivation reads.
type derivedRollout struct {
	topLevel     bool   // first session_meta has source "exec"
	subagentRole string // first session_meta's subagent thread_spawn.agent_role
	sandboxes    map[string]bool
	spawnAgent   bool
	execInputs   map[string]string // custom_tool_call "exec" input by call_id
	execOrder    []string
	outputs      map[string]string // custom_tool_call_output text by call_id
}

var (
	derivedCmdPattern  = regexp.MustCompile(`cmd:"((?:[^"\\]|\\.)*)"`)
	derivedExitPattern = regexp.MustCompile(`"exit_code":\s*(-?\d+)`)
)

func deriveRollout(data []byte) derivedRollout {
	r := derivedRollout{sandboxes: map[string]bool{}, execInputs: map[string]string{}, outputs: map[string]string{}}
	seenMeta := false
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for sc.Scan() {
		var ln struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if json.Unmarshal(sc.Bytes(), &ln) != nil {
			continue
		}
		var p map[string]any
		if json.Unmarshal(ln.Payload, &p) != nil {
			continue
		}
		switch ln.Type {
		case "session_meta":
			if seenMeta {
				continue // a subagent record repeats its parent's meta after its own
			}
			seenMeta = true
			switch src := p["source"].(type) {
			case string:
				r.topLevel = src == "exec"
			case map[string]any:
				if sub, ok := src["subagent"].(map[string]any); ok {
					if ts, ok := sub["thread_spawn"].(map[string]any); ok {
						r.subagentRole, _ = ts["agent_role"].(string)
					}
				}
			}
		case "turn_context":
			if sp, ok := p["sandbox_policy"].(map[string]any); ok {
				if t, _ := sp["type"].(string); t != "" {
					r.sandboxes[t] = true
				}
			}
		case "response_item":
			id, _ := p["call_id"].(string)
			switch p["type"] {
			case "function_call":
				if p["name"] == "spawn_agent" {
					r.spawnAgent = true
				}
			case "custom_tool_call":
				if p["name"] == "exec" {
					in, _ := p["input"].(string)
					r.execInputs[id] = in
					r.execOrder = append(r.execOrder, id)
				}
			case "custom_tool_call_output":
				r.outputs[id] += liveText(p["output"])
			}
		}
	}
	return r
}

// deriveCodexAuditEvidence applies the derivation rule to one audit item.
func deriveCodexAuditEvidence(in codexAuditEvidenceInput) codexAuditEvidence {
	ev := codexAuditEvidence{Route: "unattributed", VerdictWriter: "unattributed"}
	var parsed []derivedRollout
	for _, data := range in.Rollouts {
		r := deriveRollout(data)
		if r.spawnAgent {
			ev.UsedSpawnAgent = true
		}
		parsed = append(parsed, r)
	}

	// The audit session: a subagent of the audit role, else the top-level
	// session the launch record names for that role. Ambiguity attributes nothing.
	var audit *derivedRollout
	var subs, tops []int
	for i, r := range parsed {
		if r.subagentRole == in.Role && in.Role != "" {
			subs = append(subs, i)
		}
		if r.topLevel {
			tops = append(tops, i)
		}
	}
	recordMatches := in.LaunchRecord != nil && in.LaunchRecord.Role == in.Role && in.Role != "" &&
		in.LaunchRecord.ExitCode == 0 && in.LaunchRecord.VerdictSHA256 != nil &&
		*in.LaunchRecord.VerdictSHA256 != "" && *in.LaunchRecord.VerdictSHA256 == in.VerdictSHA256
	switch {
	case len(subs) == 1:
		audit = &parsed[subs[0]]
	case len(subs) == 0 && len(tops) == 1 && in.LaunchRecord != nil && in.LaunchRecord.Role == in.Role:
		audit = &parsed[tops[0]]
	}
	if audit == nil {
		return ev
	}

	if len(audit.sandboxes) == 1 {
		for s := range audit.sandboxes {
			ev.SessionSandbox = s
		}
	}
	ev.TopLevel = audit.topLevel && audit.subagentRole == ""

	switch {
	case audit.subagentRole != "":
		ev.Route = "spawn_agent"
	case ev.TopLevel && !ev.UsedSpawnAgent && recordMatches:
		ev.Route = "launcher"
	}
	if recordMatches {
		ev.VerdictWriter = "launcher"
	}

	for _, id := range audit.execOrder {
		m := derivedCmdPattern.FindStringSubmatch(audit.execInputs[id])
		if m == nil || !strings.Contains(audit.execInputs[id], "tools.exec_command") {
			continue
		}
		cmd, err := strconv.Unquote(`"` + m[1] + `"`)
		if err != nil || cmd != in.ProbeCommand {
			continue
		}
		ev.ProbeCommandExecuted = true
		if em := derivedExitPattern.FindStringSubmatch(audit.outputs[id]); em != nil {
			if n, err := strconv.Atoi(em[1]); err == nil {
				ev.ProbeExitCode = &n
			}
		}
		break
	}
	ev.WriteDenied = ev.ProbeCommandExecuted && ev.ProbeExitCode != nil && *ev.ProbeExitCode != 0 && !in.ProbeExists
	return ev
}
