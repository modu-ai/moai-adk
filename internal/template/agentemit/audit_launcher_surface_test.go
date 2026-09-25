// audit_launcher_surface_test.go — the instruction surface names the audit
// launcher, by the one route measured to reach the model, for every role whose
// permission contract is read-only.
//
// testdata/measured-route.txt is the committed copy of the measured route
// (`shell` or `mcp`). The surface must name only that route's launcher form
// and never the other one, must not tell the parent to start these roles with
// spawn_agent, and must say the launcher writes the file with the returned text.
package agentemit_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// launcherForms maps a measured route to the launcher form the surface names.
var launcherForms = map[string]string{
	"mcp":   "`codex_role_audit`",
	"shell": "`moai codex audit`",
}

const (
	launcherWritesMarker = "with exactly the returned text"
	launcherNameMarker   = "launcher"
)

// noSpawnClause is the explicit refusal of the spawn_agent route.
var noSpawnClause = regexp.MustCompile("never (started )?through `spawn_agent`")

// subagentClaims are phrasings that tell a reader a read-only role runs as a
// spawned subagent — the claim the launcher route retires.
var subagentClaims = regexp.MustCompile(`(?i)runs? as (a )?subagents?|started (as|by) (a )?subagents?`)

func TestAuditRoleLauncherInstructionSurface(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "measured-route.txt"))
	if err != nil {
		t.Fatalf("read measured route: %v", err)
	}
	route := strings.TrimSpace(string(raw))
	want, ok := launcherForms[route]
	if !ok {
		t.Fatalf("measured route %q is not a launchable route", route)
	}
	var unmeasured []string
	for r, form := range launcherForms {
		if r != route {
			unmeasured = append(unmeasured, form)
		}
	}

	man, err := agentemit.LoadManifest()
	if err != nil {
		t.Fatal(err)
	}
	pub := emitRealSet(t)
	var readOnly []string
	emitted := map[string]string{}
	for path, data := range pub.CodexTOML {
		doc, err := decodeTOML(string(data))
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		name, _ := doc["name"].(string)
		emitted[name] = path
		if man.PermissionContract.ContractSandbox(name) == "read-only" {
			readOnly = append(readOnly, name)
		}
	}
	sort.Strings(readOnly)
	if len(readOnly) == 0 {
		t.Fatal("no read-only contract role — vacuous")
	}

	check := func(where, text string) {
		t.Helper()
		for _, need := range []string{want, launcherNameMarker, launcherWritesMarker} {
			if !strings.Contains(text, need) {
				t.Errorf("%s lacks %q", where, need)
			}
		}
		if !noSpawnClause.MatchString(text) {
			t.Errorf("%s does not refuse the spawn_agent route", where)
		}
		for _, bad := range unmeasured {
			if strings.Contains(text, bad) {
				t.Errorf("%s names the unmeasured launcher form %s", where, bad)
			}
		}
		if m := subagentClaims.FindString(text); m != "" {
			t.Errorf("%s still says the role runs as a subagent: %q", where, m)
		}
	}

	// Every read-only role's addendum, in the manifest and in the committed TOML.
	for _, role := range readOnly {
		add, ok := man.CodexRoleAddenda[role]
		if !ok {
			t.Errorf("%s: read-only contract role carries no Codex addendum", role)
			continue
		}
		check(role+" addendum", add)
		committed, err := os.ReadFile(committedTOMLPath(emitted[role]))
		if err != nil {
			t.Fatalf("read committed %s: %v", emitted[role], err)
		}
		if !strings.Contains(string(committed), strings.TrimRight(add, "\n")) {
			t.Errorf("%s: committed TOML does not carry the manifest addendum — run make agents-emit", role)
		}
	}

	// The parent-side row on the surface Codex reads natively.
	contract, err := os.ReadFile(filepath.Join(templatesDir, "AGENTS.md.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	var row string
	for _, line := range strings.Split(string(contract), "\n") {
		if strings.HasPrefix(line, "| audit-verdict-file |") {
			row = line
		}
	}
	if row == "" {
		t.Fatal("AGENTS.md template has no audit-verdict-file row")
	}
	const refusalInstruction = "When the call is refused with `MCP tool call requires approval, but approval policy is never`, do not skip the audit and do not fall back to `spawn_agent`; return a blocker that quotes the refusal text"
	if !strings.HasSuffix(row, ". "+refusalInstruction+" |") {
		t.Errorf("audit-verdict-file row must end with the fixed approval-refusal instruction: %s", row)
	}
	if count := strings.Count(string(contract), refusalInstruction); count != 1 {
		t.Errorf("approval-refusal instruction occurs %d times, want exactly 1", count)
	}
	check("AGENTS.md audit-verdict-file row", row)
	for _, role := range readOnly {
		if !strings.Contains(row, "`"+role+"`") {
			t.Errorf("audit-verdict-file row does not name read-only role %s", role)
		}
	}
}
