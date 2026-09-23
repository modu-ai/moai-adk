// audit_role_exception_test.go — the Codex-only audit-role exception.
//
// On Codex the two audit roles run with a read-only sandbox and return their
// verdict text; the parent lane orchestrator writes the verdict or report
// file with exactly that text. The exception is Codex-only: the Claude agent
// definitions (the local copy under the repository .claude tree and the
// neutral template copy) carry no trace of it, and the Claude audit workflow
// is untouched.
package agentemit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// auditReturnMarker is the Codex-only instruction sentence the emitter appends
// to the two audit roles' developer_instructions.
const auditReturnMarker = "Return the complete verdict or report text as your final response"

// parentWriteRowMarker and parentWriteMarker identify the parent-side
// instruction on the surface Codex reads natively (the deployed AGENTS.md).
const (
	parentWriteRowMarker = "| audit-verdict-file |"
	parentWriteMarker    = "writes the verdict or report file with exactly the returned text"
)

// codexAuditRoles are the roles the exception covers, by name.
var codexAuditRoles = map[string]bool{"plan-auditor": true, "sync-auditor": true}

// repoRoot is the module root relative to this package directory.
const repoRoot = "../../.."

func TestCodexAuditRolesReadOnlyScopedException(t *testing.T) {
	pub := emitRealSet(t)

	// Emitted + committed artifacts: read-only sandbox and the return-text
	// instruction on exactly the two audit roles.
	withMarker := map[string]bool{}
	for path, data := range pub.CodexTOML {
		doc, err := decodeTOML(string(data))
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		name, _ := doc["name"].(string)
		body, _ := doc["developer_instructions"].(string)
		if strings.Contains(body, auditReturnMarker) {
			withMarker[name] = true
		}
		if codexAuditRoles[name] {
			if got, _ := doc["sandbox_mode"].(string); got != "read-only" {
				t.Errorf("%s: emitted sandbox_mode = %q, want read-only", name, got)
			}
			committed, err := os.ReadFile(committedTOMLPath(path))
			if err != nil {
				t.Fatalf("read committed %s: %v", path, err)
			}
			if !strings.Contains(string(committed), "sandbox_mode = \"read-only\"") {
				t.Errorf("%s: committed artifact does not carry sandbox_mode = \"read-only\"", path)
			}
			if !strings.Contains(string(committed), auditReturnMarker) {
				t.Errorf("%s: committed artifact does not carry the return-text instruction", path)
			}
		}
	}
	for role := range codexAuditRoles {
		if !withMarker[role] {
			t.Errorf("%s: emitted developer_instructions lack the Codex-only return-text instruction", role)
		}
	}
	for role := range withMarker {
		if !codexAuditRoles[role] {
			t.Errorf("%s: carries the audit return-text instruction but is not a Codex audit role", role)
		}
	}

	// Claude path unchanged: neither the neutral template copy (C2) nor the
	// local copy (C1) carries the Codex-only instruction.
	for role := range codexAuditRoles {
		for _, p := range []string{
			filepath.Join(templatesDir, filepath.FromSlash(agentMDRoot), role+".md"),
			filepath.Join(repoRoot, ".claude", "agents", "moai", role+".md"),
		} {
			raw, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read Claude definition %s: %v", p, err)
			}
			if strings.Contains(string(raw), auditReturnMarker) {
				t.Errorf("%s: Claude definition carries the Codex-only return-text instruction", p)
			}
		}
	}

	// Parent-side instruction on the surface Codex reads (deployed AGENTS.md).
	contract, err := os.ReadFile(filepath.Join(templatesDir, "AGENTS.md.tmpl"))
	if err != nil {
		t.Fatalf("read AGENTS.md template: %v", err)
	}
	var row string
	for _, line := range strings.Split(string(contract), "\n") {
		if strings.HasPrefix(line, parentWriteRowMarker) {
			row = line
			break
		}
	}
	if row == "" {
		t.Errorf("AGENTS.md template carries no %q capability row", parentWriteRowMarker)
	} else if !strings.Contains(row, parentWriteMarker) {
		t.Errorf("AGENTS.md audit-verdict-file row lacks the parent-write instruction %q: %s", parentWriteMarker, row)
	}

	// Manifest mutants the emitter must refuse: an audit role dropped from
	// the emitted sandbox overrides, and the addendum reassigned away.
	src, err := os.ReadFile("agents-codex.yaml")
	if err != nil {
		t.Fatalf("read manifest source: %v", err)
	}
	for label, mutate := range map[string]func(string) string{
		"plan-auditor dropped from role_values": func(s string) string {
			return strings.Replace(s, "      plan-auditor: read-only\n", "", 1)
		},
		"sync-auditor emitted workspace-write": func(s string) string {
			return strings.Replace(s, "      sync-auditor: read-only\n", "      sync-auditor: workspace-write\n", 1)
		},
	} {
		mutated := mutate(string(src))
		if mutated == string(src) {
			t.Errorf("%s: mutant did not apply — the manifest no longer carries the expected line", label)
			continue
		}
		man, err := agentemit.ParseManifest([]byte(mutated))
		if err != nil {
			continue // refused at manifest validation
		}
		if _, err := agentemit.EmitAll(os.DirFS(templatesDir), agentMDRoot, man); err == nil {
			t.Errorf("%s: emitter accepted the mutant; want refusal", label)
		}
	}
}
