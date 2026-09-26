// Package cli — codex_role_fingerprint.go
//
// The Codex role-load predicate (SPEC-ROLE-LOAD-PREDICATE-001): a contract-
// neutral replacement for the nonce-based load check. Instead of asking a
// role to echo back a nonce (which two roles — manager-lead, mission-governor
// — are contractually obliged to refuse for a delegation naming no work),
// this predicate reads the role's own session record (rollout JSONL) and
// compares the sha256 of its "developer" response_item body against the
// role's own developer_instructions, extracted from the role TOML by a
// TOML-specification parse (never a regular-expression extraction —
// REQ-RLP-003).
//
// @MX:ANCHOR: [AUTO] codexRoleLoadPredicate is the SSOT for role-load
// decisions across every role, including the two contract-refusal roles.
// @MX:REASON: replaces AC-DHR-012's nonce predicate, which measured contract
// compliance rather than role load for manager-lead and mission-governor.
package cli

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// codexRoleBodyKey is the TOML key this predicate extracts from a role file.
const codexRoleBodyKey = "developer_instructions"

// ErrCodexRoleBodyKeyAbsent reports that a role TOML source carries no
// developer_instructions key at all. Extraction failure is always reported
// as a non-nil error — never as a silently-returned empty string
// (REQ-RLP-003 negative arm).
var ErrCodexRoleBodyKeyAbsent = errors.New("codex role body: developer_instructions key absent")

// codexRoleBodyExtractTOML extracts the developer_instructions value from a
// role TOML source per the TOML 1.0 multi-line-literal-string specification:
// a newline immediately following the opening ''' delimiter is trimmed, and
// no such trimming happens when the string's content begins on the same
// line as the opening delimiter. This function never uses the regexp
// package — it is a TOML-specification parse, not a regular-expression
// extraction (REQ-RLP-003).
func codexRoleBodyExtractTOML(src string) (string, error) {
	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			break // a table header ends the top-level key scan
		}
		key, rhs, ok := strings.Cut(trimmed, "=")
		if !ok {
			continue
		}
		if strings.TrimSpace(key) != codexRoleBodyKey {
			continue
		}
		val := strings.TrimSpace(rhs)
		if !strings.HasPrefix(val, "'''") {
			return "", fmt.Errorf("codex role body: %s value is not a TOML multi-line literal string", codexRoleBodyKey)
		}
		afterOpen := val[len("'''"):]
		rest := afterOpen
		if i+1 < len(lines) {
			rest += "\n" + strings.Join(lines[i+1:], "\n")
		}
		// TOML rule: trim exactly one newline immediately following the
		// opening delimiter — but ONLY when nothing else followed the
		// delimiter on the same line.
		if afterOpen == "" {
			rest = strings.TrimPrefix(rest, "\n")
		}
		end := strings.Index(rest, "'''")
		if end < 0 {
			return "", fmt.Errorf("codex role body: %s literal is not closed", codexRoleBodyKey)
		}
		run := 3
		for end+run < len(rest) && rest[end+run] == '\'' && run < 5 {
			run++
		}
		end += run - 3 // up to two extra apostrophes may precede the closing delimiter
		return rest[:end], nil
	}
	return "", ErrCodexRoleBodyKeyAbsent
}

// codexRoleBodySHA256Hex returns the lowercase-hex sha256 of body.
func codexRoleBodySHA256Hex(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

// --- Expectation table (AC-RLP-002) -----------------------------------

// codexRoleExpectationTable maps a role's own developer_instructions body
// sha256 to its name, built from a directory of role TOML files. It is
// eligible only when every file's body is non-empty and every body hash is
// distinct from every other (REQ-RLP-005).
type codexRoleExpectationTable struct {
	ByHash map[string]string // sha256 hex -> role name
	ByRole map[string]string // role name -> sha256 hex
}

// Three distinct ineligibility reason codes an expectation-table build can
// report (AC-RLP-002).
const (
	codexRoleTableReasonEmptyDir  = "empty_dir"
	codexRoleTableReasonEmptyBody = "empty_body"
	codexRoleTableReasonCollision = "collision"
)

// codexBuildRoleExpectationTable builds an expectation table from every
// *.toml file directly inside dir. On success it returns a populated table
// and an empty reason string. On ineligibility it returns a zero table and
// one of the three reason codes above — never both a table and a reason.
func codexBuildRoleExpectationTable(dir string) (codexRoleExpectationTable, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return codexRoleExpectationTable{}, "", err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".toml") {
			files = append(files, e.Name())
		}
	}
	if len(files) == 0 {
		return codexRoleExpectationTable{}, codexRoleTableReasonEmptyDir, nil
	}
	sort.Strings(files)

	table := codexRoleExpectationTable{ByHash: map[string]string{}, ByRole: map[string]string{}}
	for _, name := range files {
		role := strings.TrimSuffix(name, ".toml")
		src, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return codexRoleExpectationTable{}, "", err
		}
		body, err := codexRoleBodyExtractTOML(string(src))
		if err != nil {
			return codexRoleExpectationTable{}, "", fmt.Errorf("role %q: %w", role, err)
		}
		if body == "" {
			return codexRoleExpectationTable{}, codexRoleTableReasonEmptyBody, nil
		}
		h := codexRoleBodySHA256Hex(body)
		if _, collided := table.ByHash[h]; collided {
			return codexRoleExpectationTable{}, codexRoleTableReasonCollision, nil
		}
		table.ByHash[h] = role
		table.ByRole[role] = h
	}
	return table, "", nil
}

// --- Session-record scanning (AC-RLP-003) -----------------------------

// codexRolloutLine is the shape common to every rollout JSONL record: a
// type discriminator, an ordinal, and a payload whose shape depends on Type.
type codexRolloutLine struct {
	Type    string          `json:"type"`
	Ordinal int             `json:"ordinal"`
	Payload json.RawMessage `json:"payload"`
}

// codexRolloutSessionMetaPayload is the payload of a "session_meta" line.
type codexRolloutSessionMetaPayload struct {
	Source json.RawMessage `json:"source"`
}

// codexRolloutSubagentSource is the shape of payload.source when the session
// is a subagent thread spawn (as opposed to a bare string like "exec" for a
// parent session).
type codexRolloutSubagentSource struct {
	Subagent struct {
		ThreadSpawn struct {
			AgentRole string `json:"agent_role"`
		} `json:"thread_spawn"`
	} `json:"subagent"`
}

// codexRolloutResponseItemPayload is the payload of a "response_item" line
// whose payload.role is "developer".
type codexRolloutResponseItemPayload struct {
	Role    string `json:"role"`
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// codexRoleSessionLabel reads a rollout JSONL file's session_meta line and
// returns the agent_role label recorded at
// payload.source.subagent.thread_spawn.agent_role. A parent session — whose
// payload.source is a bare string, e.g. "exec" — carries no label
// (hasLabel == false), never an error (REQ-RLP-015: selection is by label
// presence only).
func codexRoleSessionLabel(path string) (role string, hasLabel bool, err error) {
	f, err := os.Open(path) //nolint:gosec // fixture/test-controlled path
	if err != nil {
		return "", false, err
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var rl codexRolloutLine
		if err := json.Unmarshal(line, &rl); err != nil {
			continue
		}
		if rl.Type != "session_meta" {
			continue
		}
		var meta codexRolloutSessionMetaPayload
		if err := json.Unmarshal(rl.Payload, &meta); err != nil {
			return "", false, nil
		}
		var sub codexRolloutSubagentSource
		if err := json.Unmarshal(meta.Source, &sub); err != nil {
			return "", false, nil // payload.source is a bare string: parent session
		}
		agentRole := sub.Subagent.ThreadSpawn.AgentRole
		return agentRole, agentRole != "", nil
	}
	if err := sc.Err(); err != nil {
		return "", false, err
	}
	return "", false, nil
}

// codexRoleFingerprintDerive scans every response_item whose payload.role is
// "developer", hashes each item's concatenated text, and returns the set of
// role names from table whose expected body hash was observed in this
// session record. Its input is the session record path and an expectation
// table — never the model's free-form reply text (REQ-RLP-001).
func codexRoleFingerprintDerive(path string, table codexRoleExpectationTable) (map[string]bool, error) {
	f, err := os.Open(path) //nolint:gosec // fixture/test-controlled path
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	matched := map[string]bool{}
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var rl codexRolloutLine
		if err := json.Unmarshal(line, &rl); err != nil {
			continue
		}
		if rl.Type != "response_item" {
			continue
		}
		var item codexRolloutResponseItemPayload
		if err := json.Unmarshal(rl.Payload, &item); err != nil {
			continue
		}
		if item.Role != "developer" {
			continue
		}
		var sb strings.Builder
		for _, c := range item.Content {
			sb.WriteString(c.Text)
		}
		h := codexRoleBodySHA256Hex(sb.String())
		if role, ok := table.ByHash[h]; ok {
			matched[role] = true
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return matched, nil
}

// --- Load predicate + structural independence (AC-RLP-004, AC-RLP-006) --

// codexRoleLoadInput is codexRoleLoadPredicate's ONLY input type. It carries
// no field derived from the model's free-form reply text and no field named
// (case-insensitively) "behaviour" or "nonce" — reading such a value inside
// codexRoleLoadPredicate's body is therefore a compile error, not a
// convention (REQ-RLP-011).
type codexRoleLoadInput struct {
	Role    string
	Matched map[string]bool
}

// codexRoleLoadPredicate decides role load from Role's presence in Matched —
// the SAME function for every role, regardless of that role's delegation
// contract (REQ-RLP-002). It never branches on the role name.
func codexRoleLoadPredicate(in codexRoleLoadInput) bool {
	return in.Matched[in.Role]
}

// codexRoleBehaviour groups model-response-derived facts sampled purely for
// diagnostic/analysis purposes (e.g. whether a contract-compliant refusal
// was observed). codexRoleLoadInput MUST NOT embed or reference this type,
// and codexRoleLoadPredicate MUST NOT read from it (REQ-RLP-010,
// REQ-RLP-011); AC-RLP-006 enumerates its fields by reflection and mutates
// each independently to prove the load predicate is unaffected by any of
// them.
type codexRoleBehaviour struct {
	NonceReturned           bool
	ContractRefusalObserved bool
}
