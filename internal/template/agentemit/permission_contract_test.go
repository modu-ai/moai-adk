// permission_contract_test.go — the role permission contract.
//
// Every Codex role carries a contract over seven axes. Each axis restriction
// a role's contract requires maps to exactly one of `enforced` (a Codex field
// the emitter writes, with a measured or documented basis) or `UNSUPPORTED`
// (a host-expressivity reason, with a measured, documented, or unmeasured
// basis). An unmeasured basis is never `enforced`, a required restriction
// with no mapping fails emission, and an `UNSUPPORTED` axis is never counted
// as passing.
package agentemit_test

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

// contractAxes is the axis set the contract must cover, written here from the
// requirement rather than read back from the package under test.
var contractAxes = []string{"sandbox", "write-path-scope", "shell", "mcp-server", "mcp-tool", "subagent", "web"}

type roleAxis struct{ role, axis, kind string }

func (r roleAxis) String() string { return r.role + "/" + r.axis + "/" + r.kind }

func mappingFor(t *testing.T, man agentemit.Manifest, axis, kind string) (agentemit.AxisMapping, bool) {
	t.Helper()
	for _, m := range man.PermissionContract.Axes {
		if m.Axis == axis && m.Kind == kind {
			return m, true
		}
	}
	return agentemit.AxisMapping{}, false
}

func contractSandbox(man agentemit.Manifest, role string) string {
	if v, ok := man.PermissionContract.RoleSandbox[role]; ok {
		return v
	}
	return man.PermissionContract.DefaultSandbox
}

func TestRolePermissionContractNoSilentDrop(t *testing.T) {
	man := mustManifest(t)
	if man.PermissionContract == nil {
		t.Fatal("manifest carries no permission_contract")
	}

	// Axis table: all seven axes covered; each row exactly one of the two
	// mappings with a basis; never an unmeasured enforced row.
	covered := map[string]bool{}
	kinds := map[string]map[string]bool{}
	for _, m := range man.PermissionContract.Axes {
		covered[m.Axis] = true
		if kinds[m.Axis] == nil {
			kinds[m.Axis] = map[string]bool{}
		}
		if kinds[m.Axis][m.Kind] {
			t.Errorf("%s/%s: mapped more than once", m.Axis, m.Kind)
		}
		kinds[m.Axis][m.Kind] = true
		switch m.Mapping {
		case "enforced":
			if m.Field == "" {
				t.Errorf("%s/%s: enforced mapping names no Codex field", m.Axis, m.Kind)
			}
			if m.Basis != "measured" && m.Basis != "documented" {
				t.Errorf("%s/%s: enforced mapping basis %q, want measured or documented", m.Axis, m.Kind, m.Basis)
			}
		case "UNSUPPORTED":
			if strings.TrimSpace(m.Reason) == "" {
				t.Errorf("%s/%s: UNSUPPORTED mapping names no host-expressivity reason", m.Axis, m.Kind)
			}
			if m.Basis != "measured" && m.Basis != "documented" && m.Basis != "unmeasured" {
				t.Errorf("%s/%s: UNSUPPORTED mapping basis %q invalid", m.Axis, m.Kind, m.Basis)
			}
		default:
			t.Errorf("%s/%s: mapping %q is neither enforced nor UNSUPPORTED", m.Axis, m.Kind, m.Mapping)
		}
		if strings.TrimSpace(m.Evidence) == "" {
			t.Errorf("%s/%s: mapping carries no evidence text", m.Axis, m.Kind)
		}
	}
	for _, axis := range contractAxes {
		if !covered[axis] {
			t.Errorf("axis %s: absent from the permission contract", axis)
		}
	}
	if !kinds["mcp-server"]["grant"] || !kinds["mcp-server"]["deny"] {
		t.Errorf("mcp-server must be mapped once per restriction kind (grant and deny); have %v", kinds["mcp-server"])
	}

	// Real set: every requirement of every role is mapped, and the emitted
	// sandbox_mode equals the contract's value for that role.
	pub := emitRealSet(t)
	report, err := agentemit.BuildPermissionReport(os.DirFS(templatesDir), agentMDRoot, man)
	if err != nil {
		t.Fatalf("BuildPermissionReport: %v", err)
	}
	roles := map[string]int{}
	for _, v := range report.Verdicts {
		roles[v.Role]++
		if _, ok := mappingFor(t, man, v.Axis, v.Kind); !ok {
			t.Errorf("%s/%s/%s: report verdict with no contract mapping", v.Role, v.Axis, v.Kind)
		}
	}
	if len(roles) != 12 {
		t.Errorf("report covers %d roles, want 12", len(roles))
	}
	for path, data := range pub.CodexTOML {
		doc, err := decodeTOML(string(data))
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		name, _ := doc["name"].(string)
		if got, want := doc["sandbox_mode"], contractSandbox(man, name); got != want {
			t.Errorf("%s: emitted sandbox_mode %v, contract says %q", name, got, want)
		}
		if roles[name] == 0 {
			t.Errorf("%s: no contract verdicts in the report", name)
		}
	}

	// Mutants the generator must refuse.
	fsys := os.DirFS(templatesDir)
	dropped := mustManifest(t)
	var kept []agentemit.AxisMapping
	for _, m := range dropped.PermissionContract.Axes {
		if m.Axis != "web" {
			kept = append(kept, m)
		}
	}
	dropped.PermissionContract.Axes = kept
	if _, err := agentemit.EmitAll(fsys, agentMDRoot, dropped); err == nil || !strings.Contains(err.Error(), "web") {
		t.Errorf("mapping removed for axis web: emitter must refuse naming the axis, got %v", err)
	}

	broad := mustManifest(t)
	broad.Fields["sandbox_mode"].RoleValues["mission-governor"] = "workspace-write"
	if _, err := agentemit.EmitAll(fsys, agentMDRoot, broad); err == nil || !strings.Contains(err.Error(), "mission-governor") {
		t.Errorf("read-only contract role emitted workspace-write: emitter must refuse, got %v", err)
	}

	src, err := os.ReadFile("agents-codex.yaml")
	if err != nil {
		t.Fatalf("read manifest source: %v", err)
	}
	for label, tc := range map[string]struct{ old, new, want string }{
		"unmeasured axis marked enforced": {
			old:  "    - axis: shell\n      kind: deny\n      mapping: UNSUPPORTED\n",
			new:  "    - axis: shell\n      kind: deny\n      mapping: enforced\n      field: sandbox_mode\n",
			want: "unmeasured",
		},
		"axis row deleted": {
			old:  "    - axis: subagent\n",
			new:  "    - axis: subagent-renamed\n",
			want: "subagent",
		},
	} {
		mutated := strings.Replace(string(src), tc.old, tc.new, 1)
		if mutated == string(src) {
			t.Errorf("%s: mutant did not apply", label)
			continue
		}
		if _, err := agentemit.ParseManifest([]byte(mutated)); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: manifest validation must refuse mentioning %q, got %v", label, tc.want, err)
		}
	}

	// An enforced mapping naming a field the emitter never writes is refused
	// in memory too (bypassing manifest validation).
	fake := mustManifest(t)
	for i, m := range fake.PermissionContract.Axes {
		if m.Axis == "shell" {
			fake.PermissionContract.Axes[i].Mapping = "enforced"
			fake.PermissionContract.Axes[i].Basis = "measured"
			fake.PermissionContract.Axes[i].Field = "shell_enabled"
		}
	}
	if _, err := agentemit.EmitAll(fsys, agentMDRoot, fake); err == nil || !strings.Contains(err.Error(), "shell") {
		t.Errorf("enforced mapping through a field the emitter never writes must be refused, got %v", err)
	}
}

func TestRolePermissionUnsupportedNeverPass(t *testing.T) {
	man := mustManifest(t)
	if man.PermissionContract == nil {
		t.Fatal("manifest carries no permission_contract")
	}
	fsys := os.DirFS(templatesDir)
	report, err := agentemit.BuildPermissionReport(fsys, agentMDRoot, man)
	if err != nil {
		t.Fatalf("BuildPermissionReport: %v", err)
	}

	// Expected set, computed from the contract: (role, axis, kind) where the
	// role's contract requires the restriction and its axis maps UNSUPPORTED.
	entries, err := os.ReadDir(templatesDir + "/" + agentMDRoot)
	if err != nil {
		t.Fatalf("read agent sources: %v", err)
	}
	want := map[roleAxis]bool{}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(templatesDir + "/" + agentMDRoot + "/" + e.Name())
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		doc, err := agentemit.ParseAgentDoc(e.Name(), raw)
		if err != nil {
			t.Fatalf("parse %s: %v", e.Name(), err)
		}
		reqs, err := agentemit.RoleRequirements(doc, man)
		if err != nil {
			t.Fatalf("RoleRequirements %s: %v", doc.Name, err)
		}
		for _, r := range reqs {
			m, ok := mappingFor(t, man, r.Axis, r.Kind)
			if !ok {
				t.Fatalf("%s/%s/%s: requirement with no mapping", doc.Name, r.Axis, r.Kind)
			}
			if m.Mapping == "UNSUPPORTED" {
				want[roleAxis{doc.Name, r.Axis, r.Kind}] = true
			}
		}
	}
	if len(want) == 0 {
		t.Fatal("expected UNSUPPORTED set is empty — the computation is blind")
	}

	got := map[roleAxis]bool{}
	for _, v := range report.Unsupported {
		if v.Mapping != "UNSUPPORTED" {
			t.Errorf("%s/%s/%s: listed as UNSUPPORTED with mapping %q", v.Role, v.Axis, v.Kind, v.Mapping)
		}
		got[roleAxis{v.Role, v.Axis, v.Kind}] = true
	}
	var missing, extra []string
	for k := range want {
		if !got[k] {
			missing = append(missing, k.String())
		}
	}
	for k := range got {
		if !want[k] {
			extra = append(extra, k.String())
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 || len(extra) > 0 {
		t.Errorf("UNSUPPORTED set differs from the contract: missing %v, extra %v", missing, extra)
	}

	for _, v := range report.Pass {
		if v.Mapping != "enforced" {
			t.Errorf("%s/%s/%s: counted as PASS with mapping %q", v.Role, v.Axis, v.Kind, v.Mapping)
		}
		if got[roleAxis{v.Role, v.Axis, v.Kind}] {
			t.Errorf("%s/%s/%s: counted as both PASS and UNSUPPORTED", v.Role, v.Axis, v.Kind)
		}
	}
	if len(report.Pass)+len(report.Unsupported) != len(report.Verdicts) {
		t.Errorf("PASS %d + UNSUPPORTED %d != verdicts %d", len(report.Pass), len(report.Unsupported), len(report.Verdicts))
	}

	// Anchors that keep the computation honest in both directions.
	for _, k := range []roleAxis{
		{"mission-governor", "shell", "deny"},
		{"manager-develop", "write-path-scope", "path-scope"},
		{"plan-auditor", "mcp-tool", "subset"},
	} {
		if !want[k] {
			t.Errorf("%s: expected in the UNSUPPORTED set", k)
		}
	}
	for _, k := range []roleAxis{
		{"manager-lead", "subagent", "deny"},
		{"plan-auditor", "write-path-scope", "path-scope"},
		{"sync-auditor", "write-path-scope", "path-scope"},
	} {
		if want[k] {
			t.Errorf("%s: must not be in the UNSUPPORTED set", k)
		}
	}
}
