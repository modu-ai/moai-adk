package web

// mcp_audit_surface_test.go — SPEC-MOAI-MCP-SERVER-001 M4 (REQ-MCP-015 /
// AC-MCP-021). Verifies the audit selection surfaces in the web console
// schema, AND that the web console does NOT fork the audit interpreter — it
// reuses the M3 typed config (config.AuditConfig). The per-agent resolution
// sentinels are scoped by SPEC-WEB-AGENTFM-RESTORE-001 to the restored
// agentfm surface files (TestWebConsole_NoPerAgentModelResolver below); the
// console still derives nothing of its own — it calls the template resolver.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestSchemaSurfaces_AuditSelection verifies AC-MCP-021: the audit_model +
// per-auditor audit_gate selection surface as schema fields in the Workflow
// section (so the web console renders them via the same schema-driven form +
// yamlpatch seam worktree/branch_guard use). The fields reuse the M3 typed
// config yaml paths (workflow.audit.model / .gates.*) — the identical
// interpreter.
func TestSchemaSurfaces_AuditSelection(t *testing.T) {
	wantNames := map[string]bool{
		"workflow.audit.model":        true,
		"workflow.audit.gates.claude": true,
		"workflow.audit.gates.codex":  true,
		"workflow.audit.gates.glm":    true,
	}
	got := map[string]bool{}
	for _, f := range settings.AllFields() {
		if f.Section == settings.SectionWorkflow {
			got[f.Name] = true
		}
	}
	for name := range wantNames {
		if !got[name] {
			t.Errorf("workflow schema field %q missing (AC-MCP-021 — audit selection must surface in the web console)", name)
		}
	}
}

// TestWebConsole_AuditNoForkedInterpreter verifies the "identical interpreter"
// clause of AC-MCP-021: the web console MUST NOT define its own audit-backend
// resolver. The audit selection surfaces via the schema-driven form (which
// reads/writes the M3 config.AuditConfig yaml paths); the model/effort
// resolution for the audit backends stays in the MCP handlers (which call the
// shared SSOT). A second ResolveAuditPlan / audit-model resolver in
// internal/web would be the fork this test forbids.
//
// The sentinel is only meaningful while it names a live symbol:
// TestWebConsole_AuditSentinelExists is the positive control that fails when
// config.ResolveAuditPlan is renamed or removed, so this guard cannot go
// vacuous.
func TestWebConsole_AuditNoForkedInterpreter(t *testing.T) {
	// Collect every non-test .go file in internal/web.
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob internal/web *.go: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if strings.Contains(string(data), sentinel) {
			t.Errorf("internal/web/%s defines/references %q — the web console must NOT fork the audit interpreter (AC-MCP-021); it reads the M3 config.AuditConfig via the schema seam", f, sentinel)
		}
	}
}

// sentinel is the live audit-model resolver symbol the web console must not
// fork: config.ResolveAuditPlan, the single consumer of the workflow.audit.model
// token. Both the guard and its positive control read this one constant.
const sentinel = "ResolveAuditPlan"

// declaresFunc reports how many non-test .go files in dir declare
// `func <name>(`. It returns an error when dir holds no non-test source at
// all, so an empty sweep is never read as a verdict.
func declaresFunc(dir, name string) (int, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return 0, err
	}
	var scanned, hits int
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			return 0, err
		}
		scanned++
		if strings.Contains(string(data), "func "+name+"(") {
			hits++
		}
	}
	if scanned == 0 {
		return 0, os.ErrNotExist
	}
	return hits, nil
}

// TestWebConsole_AuditSentinelExists is the positive control for
// TestWebConsole_AuditNoForkedInterpreter: a guard whose sentinel names a
// deleted symbol passes vacuously, so the sentinel must be declared in the
// non-test source of internal/config. The second subtest proves the probe can
// report absence.
func TestWebConsole_AuditSentinelExists(t *testing.T) {
	t.Run("sentinel is declared in internal/config non-test source", func(t *testing.T) {
		hits, err := declaresFunc(filepath.Join("..", "config"), sentinel)
		if err != nil {
			t.Fatalf("scan internal/config: %v", err)
		}
		if hits == 0 {
			t.Errorf("func %s( is not declared in internal/config non-test source — the no-forked-interpreter guard names no live symbol and passes vacuously", sentinel)
		}
	})
	t.Run("probe reports an absent symbol as absent", func(t *testing.T) {
		hits, err := declaresFunc(filepath.Join("..", "config"), "NoSuchAuditResolverSymbol")
		if err != nil {
			t.Fatalf("scan internal/config: %v", err)
		}
		if hits != 0 {
			t.Errorf("probe found %d declarations of a symbol that does not exist — the control cannot fail", hits)
		}
	})
}

// agentfmSurfaceFiles are the restored console surface files the four
// per-agent resolution sentinels are ALLOWED in (SPEC-WEB-AGENTFM-RESTORE-001,
// plan §D.5 amend row — the blueprint-measured sentinel-bearing surfaces:
// app.go :3100 patchAgentFM wiring, handlers.go :6488 the template resolver
// call, schemaform.go :7098/:7115 EffectiveProfile/AgentOverrides reads, and
// the re-ported agentfm.go which calls template.ResolveAgentModelEffort).
// The set tracks reality per milestone: M3 added agentfm.go when it re-ported
// the parse/persist helpers that call template.ResolveAgentModelEffort.
// Every OTHER non-test internal/web file — the generated fieldsets_templ.go
// included — stays sentinel-free.
var agentfmSurfaceFiles = map[string]bool{
	"agentfm.go":    true,
	"app.go":        true,
	"handlers.go":   true,
	"schemaform.go": true,
}

// TestWebConsole_NoPerAgentModelResolver verifies the narrowed per-agent
// resolution contract (SPEC-WEB-AGENTFM-RESTORE-001 plan §D.5, amending the
// SPEC-AGENT-MODEL-INHERIT-001-era blanket ban). Three clauses:
//
//  1. DEFINITION BAN (unconditional, every non-test web file): internal/web
//     must not DEFINE its own per-agent resolver — the single derivation
//     lives in template.ResolveAgentModelEffort (REQ-AFR-010 "no second
//     derivation"). A `func ResolveAgentModelEffort` anywhere in this
//     package is the fork this clause forbids.
//  2. REFERENCE BAN outside the restored surface files: the four sentinels
//     (ResolveAgentModelEffort / ProfileMatrixAgents / EffectiveProfile /
//     AgentOverrides) may appear ONLY in agentfmSurfaceFiles; every other
//     non-test file stays sentinel-free.
//  3. NON-EMPTY EXCLUSION SET: if the allowed set shrinks — a surface file
//     deleted without the guard being revisited — the guard FAILS rather
//     than silently passing on a surface that no longer exists.
func TestWebConsole_NoPerAgentModelResolver(t *testing.T) {
	if len(agentfmSurfaceFiles) == 0 {
		t.Fatal("the agentfm surface allowlist is empty — the restored surface files must be named here or the guard is vacuous")
	}
	for f := range agentfmSurfaceFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("allowed surface file %s is missing from internal/web — revisit this guard before dropping it from the allowlist", f)
		}
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob internal/web *.go: %v", err)
	}
	var scanned int
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		scanned++
		if strings.Contains(string(data), "func ResolveAgentModelEffort") {
			t.Errorf("internal/web/%s defines ResolveAgentModelEffort — the web console must not fork the resolver; call template.ResolveAgentModelEffort (single derivation, REQ-AFR-010)", f)
		}
		if agentfmSurfaceFiles[f] {
			continue
		}
		for _, sym := range []string{"ResolveAgentModelEffort", "ProfileMatrixAgents", "EffectiveProfile", "AgentOverrides"} {
			if strings.Contains(string(data), sym) {
				t.Errorf("internal/web/%s references %s — per-agent resolution sentinels are allowed only in the restored agentfm surface files (%v)", f, sym, surfaceFileNames())
			}
		}
	}
	if scanned == 0 {
		t.Fatal("scanned no internal/web source file — the guard read nothing")
	}
}

// surfaceFileNames returns the allowlist keys in a stable order for messages.
func surfaceFileNames() []string {
	names := make([]string, 0, len(agentfmSurfaceFiles))
	for f := range agentfmSurfaceFiles {
		names = append(names, f)
	}
	sort.Strings(names)
	return names
}
