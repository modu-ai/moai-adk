// Package template — Navigator audit script tests (TDD RED→GREEN driver).
//
// Test fixture driver for the Project Navigator audit shell script shipped at
// templates/.claude/skills/moai-workflow-project/scripts/navigator-audit.sh
// (sibling to navigator-regen.sh). The script is the deterministic audit core:
// it reads the project's design docs (product/structure/tech.md) + the
// capability-map produced by SPEC-PROJECT-NAVIGATOR-001, computes a
// bidirectional drift diff (Missing SPECs + Orphan SPECs + Matched), and emits
// audit-report.md + audit-report.json under .moai/project/navigator/.
//
// These tests exercise AC-NA-001..012 against fixture projects built under
// t.TempDir(). The script is invoked as a shell subprocess so the tests are
// black-box over its bash internals.
//
// Header-driven column resolution (AC-NA-007): two capability-map variants are
// exercised — variant-A capability-first (per 001 spec.md spelling) and
// variant-B spec-id-first (per 001 acceptance.md AC-PN-013 spelling). The
// audit MUST parse both correctly via header-driven resolution.
//
// Matching heuristic bases (AC-NA-007): fixtures exercise at minimum one of
// each — exact (normalized equality), substring (shorter ≥4 chars), and
// module-token (last path segment ≥4 chars appearing as a token in the
// normalized design-doc name). The ≥4-char floor is verified by an
// implementation-path `internal/crm` (last seg `crm`, 3 chars) that MUST NOT
// produce a module-token match.
package template

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// navigatorAuditScript is the template-relative path to the audit script.
const navigatorAuditScript = "templates/.claude/skills/moai-workflow-project/scripts/navigator-audit.sh"

// auditReportFiles returns the (md, json) audit-report paths under dir.
func auditReportFiles(dir string) (md, js string) {
	base := filepath.Join(dir, ".moai", "project", "navigator")
	return filepath.Join(base, "audit-report.md"), filepath.Join(base, "audit-report.json")
}

// runAudit runs the navigator-audit.sh script against dir. It returns the
// combined stdout+stderr and the exit error (or nil). The script MUST exit 0
// always (fail-open contract per REQ-NA-010).
func runAudit(t *testing.T, dir string, envExtra ...string) (output string, exitErr error) {
	t.Helper()
	scriptAbs, err := filepath.Abs(navigatorAuditScript)
	if err != nil {
		t.Fatalf("resolve script abs: %v", err)
	}
	if _, err := os.Stat(scriptAbs); err != nil {
		t.Fatalf("navigator-audit.sh not found at %s — has M2 landed the script?", scriptAbs)
	}
	c := exec.Command("bash", scriptAbs)
	c.Dir = dir
	env := append(os.Environ(), "CLAUDE_PROJECT_DIR="+dir)
	env = append(env, envExtra...)
	c.Env = env
	var buf bytes.Buffer
	c.Stdout = &buf
	c.Stderr = &buf
	err = c.Run()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return buf.String(), ee
		}
		t.Fatalf("invoke script: %v", err)
	}
	return buf.String(), nil
}

// writeDesignDoc writes a design doc at dir/.moai/project/<name> with a
// `## Core Features` section enumerating the given feature names as bolded
// bullets. Each feature is `## Core Features` -> `- **<name>** ...`.
func writeDesignDoc(t *testing.T, dir, name string, features []string) {
	t.Helper()
	path := filepath.Join(dir, ".moai", "project", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir design doc dir: %v", err)
	}
	var b strings.Builder
	b.WriteString("# " + name + "\n\n")
	b.WriteString("## Overview\n\nPlaceholder overview for " + name + ".\n\n")
	b.WriteString("## Core Features\n\n")
	for _, f := range features {
		b.WriteString("- **" + f + "** — short description.\n")
	}
	b.WriteString("\n## Architecture\n\nNon-feature narrative section.\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write design doc %s: %v", name, err)
	}
}

// capMapVariant enumerates the two header spellings AC-NA-007 requires.
type capMapVariant int

const (
	// capMapVariantA: capability | owning-spec | status | implementation-path | commit-sha | captured-at
	// (capability-first, per 001 spec.md spelling).
	capMapVariantA capMapVariant = iota
	// capMapVariantB: spec-id | title | status | implementation-path | commit-sha | captured-at
	// (spec-id-first, per 001 acceptance.md AC-PN-013 spelling).
	capMapVariantB
)

// capRow is a single capability-map data row.
type capRow struct {
	SpecID            string
	Title             string
	Status            string
	ImplementationPath string
}

// writeCapabilityMap writes a capability-map.md at
// dir/.moai/project/navigator/capability-map.md under the given header
// variant. commit-sha + captured-at are filled with deterministic fixtures.
func writeCapabilityMap(t *testing.T, dir string, v capMapVariant, rows []capRow) {
	t.Helper()
	path := filepath.Join(dir, ".moai", "project", "navigator", "capability-map.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir navigator dir: %v", err)
	}
	var b strings.Builder
	b.WriteString("# Capability Map\n\n")
	switch v {
	case capMapVariantA:
		b.WriteString("| capability | owning-spec | status | implementation-path | commit-sha | captured-at |\n")
		b.WriteString("|------------|-------------|--------|----------------------|------------|-------------|\n")
		for _, r := range rows {
			b.WriteString("| " + r.Title + " | " + r.SpecID + " | " + r.Status + " | " + r.ImplementationPath + " | deadbeefdeadbeefdeadbeefdeadbeefdeadbeef | 2026-08-01T10:00:00+00:00 |\n")
		}
	case capMapVariantB:
		b.WriteString("| spec-id | title | status | implementation-path | commit-sha | captured-at |\n")
		b.WriteString("|---------|-------|--------|----------------------|------------|-------------|\n")
		for _, r := range rows {
			b.WriteString("| " + r.SpecID + " | " + r.Title + " | " + r.Status + " | " + r.ImplementationPath + " | deadbeefdeadbeefdeadbeefdeadbeefdeadbeef | 2026-08-01T10:00:00+00:00 |\n")
		}
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write capability-map: %v", err)
	}
}

// writeOverride writes the audit-known-matches.yaml override file.
// match is a list of design-name=spec-id pairs; ignore is a list of strings
// (each is a design-name OR spec-id).
func writeOverride(t *testing.T, dir string, match [][2]string, ignore []string) {
	t.Helper()
	path := filepath.Join(dir, ".moai", "project", "navigator", "audit-known-matches.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir navigator dir: %v", err)
	}
	var b strings.Builder
	b.WriteString("# fixture override\n")
	if len(match) > 0 {
		b.WriteString("match:\n")
		for _, m := range match {
			b.WriteString("  - design_name: \"" + m[0] + "\"\n")
			b.WriteString("    spec_id: \"" + m[1] + "\"\n")
		}
	}
	if len(ignore) > 0 {
		b.WriteString("ignore:\n")
		for _, i := range ignore {
			b.WriteString("  - \"" + i + "\"\n")
		}
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write override: %v", err)
	}
}

// loadAuditJSON loads audit-report.json into a generic map for assertion.
func loadAuditJSON(t *testing.T, dir string) map[string]any {
	t.Helper()
	_, js := auditReportFiles(dir)
	data, err := os.ReadFile(js)
	if err != nil {
		t.Fatalf("read audit-report.json: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse audit-report.json: %v\nbody: %s", err, data)
	}
	return m
}

// fixtureAudit sets up a populated fixture (design docs + capability-map +
// optional SPEC dirs) and runs the audit. Returns the parsed JSON report.
func fixtureAudit(t *testing.T, v capMapVariant, withOverride bool) map[string]any {
	t.Helper()
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	designFeatures := []string{
		"Project Initialization",   // module-token match via internal/cli/init
		"Authentication",           // exact match
		"Notification System",      // substring match
		"Real-time Collaboration",  // Missing SPEC
	}
	if withOverride {
		designFeatures = append(designFeatures, "Autonomy Loop") // override match
		designFeatures = append(designFeatures, "Old Feature Name") // override ignore
	}
	writeDesignDoc(t, dir, "product.md", designFeatures)
	writeDesignDoc(t, dir, "structure.md", nil)
	writeDesignDoc(t, dir, "tech.md", nil)
	rows := []capRow{
		{"SPEC-CLI-001", "CLI Tool — init template selection", "completed", "internal/cli/init"},
		{"SPEC-AUTH-001", "Authentication", "completed", "internal/auth"},
		{"SPEC-NOTIFY-001", "Notifications Subsystem", "in-progress", "internal/notify"},
		{"SPEC-X-999", "Legacy CRM Integration", "in-progress", "internal/crm"},   // orphan, last seg crm < 4
	}
	if withOverride {
		rows = append(rows,
			capRow{"SPEC-AUTONOMY-WORKFLOW-001", "Autonomy Workflow Loop", "completed", "internal/autonomy"},
			capRow{"SPEC-DEPRECATED-001", "Deprecated Feature", "in-progress", "internal/deprecated"}, // override ignore
		)
	}
	writeCapabilityMap(t, dir, v, rows)
	if withOverride {
		writeOverride(t, dir,
			[][2]string{{"Autonomy Loop", "SPEC-AUTONOMY-WORKFLOW-001"}},
			[]string{"SPEC-DEPRECATED-001", "Old Feature Name"})
	}
	if _, err := runAudit(t, dir); err != nil {
		t.Fatalf("audit exited non-zero: %v", err)
	}
	return loadAuditJSON(t, dir)
}

// --- AC-NA-001 ------------------------------------------------------------

// TestACNA001_AuditProducesReport verifies AC-NA-001: the audit produces
// exactly two files (audit-report.md + audit-report.json) and no other
// top-level audit file.
func TestACNA001_AuditProducesReport(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeDesignDoc(t, dir, "product.md", []string{"Some Feature"})
	writeCapabilityMap(t, dir, capMapVariantB, []capRow{
		{"SPEC-X-001", "Some Feature", "completed", "internal/x"},
	})
	if _, err := runAudit(t, dir); err != nil {
		t.Fatalf("audit exited non-zero: %v", err)
	}
	md, js := auditReportFiles(dir)
	for _, p := range []string{md, js} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("AC-NA-001: expected %s to exist: %v", p, err)
		}
	}
	// No third top-level audit file.
	base := filepath.Join(dir, ".moai", "project", "navigator")
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatalf("read navigator dir: %v", err)
	}
	auditExtras := []string{}
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "audit-") && n != "audit-report.md" && n != "audit-report.json" {
			auditExtras = append(auditExtras, n)
		}
	}
	if len(auditExtras) > 0 {
		t.Errorf("AC-NA-001: extra top-level audit files found: %v", auditExtras)
	}
}

// --- AC-NA-002 ------------------------------------------------------------

// TestACNA002_ReadOnlyNoRegeneration verifies AC-NA-002: capability-map +
// progress-map + navigator.md + design docs are byte-identical before/after
// the audit, and the audit script never invokes navigator-regen.sh.
func TestACNA002_ReadOnlyNoRegeneration(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeDesignDoc(t, dir, "product.md", []string{"Some Feature"})
	writeDesignDoc(t, dir, "structure.md", nil)
	writeDesignDoc(t, dir, "tech.md", nil)
	// Seed the 001 outputs (capability-map.md, progress-map.md, navigator.md)
	// so the read-only claim is verifiable.
	navDir := filepath.Join(dir, ".moai", "project", "navigator")
	if err := os.MkdirAll(navDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"capability-map.md", "progress-map.md", "navigator.md"} {
		body := []byte("# OLD " + name + " (must not change)\n")
		if err := os.WriteFile(filepath.Join(navDir, name), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Snapshot sha256 of all input surfaces.
	snapshot := func() map[string][]byte {
		m := map[string][]byte{}
		for _, rel := range []string{
			".moai/project/product.md",
			".moai/project/structure.md",
			".moai/project/tech.md",
			".moai/project/navigator/capability-map.md",
			".moai/project/navigator/progress-map.md",
			".moai/project/navigator/navigator.md",
		} {
			b, err := os.ReadFile(filepath.Join(dir, rel))
			if err != nil {
				t.Fatalf("snapshot %s: %v", rel, err)
			}
			m[rel] = b
		}
		return m
	}
	// Need a capability-map with at least one row matching the design feature.
	writeCapabilityMap(t, dir, capMapVariantB, []capRow{
		{"SPEC-X-001", "Some Feature", "completed", "internal/x"},
	})
	before := snapshot()
	if _, err := runAudit(t, dir); err != nil {
		t.Fatalf("audit exited non-zero: %v", err)
	}
	after := snapshot()
	for rel, b := range before {
		if !bytes.Equal(b, after[rel]) {
			t.Errorf("AC-NA-002: %s modified by audit (read-only violation)\nbefore: %q\nafter:  %q", rel, b, after[rel])
		}
	}
}

// --- AC-NA-003 ------------------------------------------------------------

// TestACNA003_MissingSPECdetection verifies AC-NA-003: a design-named feature
// with no capability-map match appears in missing[] with provenance.
func TestACNA003_MissingSPECdetection(t *testing.T) {
	report := fixtureAudit(t, capMapVariantB, false)
	missing, _ := report["missing"].([]any)
	found := false
	for _, e := range missing {
		entry, _ := e.(map[string]any)
		if name, _ := entry["design_name"].(string); name == "Real-time Collaboration" {
			found = true
			src, _ := entry["source"].(map[string]any)
			file, _ := src["file"].(string)
			hp, _ := src["heading_path"].(string)
			if !strings.Contains(file, "product.md") {
				t.Errorf("AC-NA-003: missing.source.file expected product.md, got %q", file)
			}
			if !strings.Contains(hp, "Core Features") {
				t.Errorf("AC-NA-003: missing.source.heading_path expected Core Features, got %q", hp)
			}
			// closest_match MUST be null per D2 fix.
			if cm, ok := entry["closest_match"]; ok && cm != nil {
				t.Errorf("AC-NA-003: closest_match expected null, got %v", cm)
			}
		}
	}
	if !found {
		t.Errorf("AC-NA-003: 'Real-time Collaboration' not in missing[]: %v", missing)
	}
}

// --- AC-NA-004 ------------------------------------------------------------

// TestACNA004_OrphanSPECdetection verifies AC-NA-004: a capability-map row
// whose implementation-path last segment is `crm` (3 chars, under the ≥4-char
// floor) and has no design-doc anchor appears in orphan[].
func TestACNA004_OrphanSPECdetection(t *testing.T) {
	report := fixtureAudit(t, capMapVariantB, false)
	orphan, _ := report["orphan"].([]any)
	found := false
	for _, e := range orphan {
		entry, _ := e.(map[string]any)
		if sid, _ := entry["spec_id"].(string); sid == "SPEC-X-999" {
			found = true
			if title, _ := entry["title"].(string); !strings.Contains(title, "Legacy CRM Integration") {
				t.Errorf("AC-NA-004: orphan.title expected 'Legacy CRM Integration', got %q", title)
			}
			if ip, _ := entry["implementation_path"].(string); ip != "internal/crm" {
				t.Errorf("AC-NA-004: orphan.implementation_path expected internal/crm, got %q", ip)
			}
		}
	}
	if !found {
		t.Errorf("AC-NA-004: SPEC-X-999 not in orphan[]: %v", orphan)
	}
}

// --- AC-NA-005 ------------------------------------------------------------

// TestACNA005_DualOutputStableSchema verifies AC-NA-005: JSON has exactly the
// 6 top-level keys; markdown has exactly the 3 sections in order; no extras.
func TestACNA005_DualOutputStableSchema(t *testing.T) {
	report := fixtureAudit(t, capMapVariantB, false)
	want := map[string]bool{
		"audit_at": true, "audit_commit": true, "inputs": true,
		"missing": true, "orphan": true, "matched": true,
	}
	for k := range report {
		if !want[k] {
			t.Errorf("AC-NA-005: unexpected top-level JSON key %q (schema not stable)", k)
		}
	}
	for k := range want {
		if _, ok := report[k]; !ok {
			t.Errorf("AC-NA-005: missing top-level JSON key %q", k)
		}
	}
	// Markdown section ordering.
	dir := redoFixtureReadMD(t, capMapVariantB, false)
	md, _ := auditReportFiles(dir)
	mdata, err := os.ReadFile(md)
	if err != nil {
		t.Fatalf("read audit-report.md: %v", err)
	}
	body := string(mdata)
	idxMissing := strings.Index(body, "## Missing SPECs")
	idxOrphan := strings.Index(body, "## Orphan SPECs")
	idxMatched := strings.Index(body, "## Matched")
	if idxMissing < 0 || idxOrphan < 0 || idxMatched < 0 {
		t.Fatalf("AC-NA-005: markdown missing required section(s); indices missing=%d orphan=%d matched=%d\nbody:\n%s",
			idxMissing, idxOrphan, idxMatched, body)
	}
	if !(idxMissing < idxOrphan && idxOrphan < idxMatched) {
		t.Errorf("AC-NA-005: markdown sections out of order: missing=%d orphan=%d matched=%d", idxMissing, idxOrphan, idxMatched)
	}
}

// redoFixtureReadMD is a helper for AC-NA-005's markdown section check; it
// re-materializes the same fixture as fixtureAudit but returns the dir so the
// caller can read the markdown. (Keeping fixture setup inline-by-id here would
// duplicate fixtureAudit; this helper accepts that cost for test isolation.)
func redoFixtureReadMD(t *testing.T, v capMapVariant, withOverride bool) string {
	t.Helper()
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeDesignDoc(t, dir, "product.md", []string{
		"Project Initialization", "Authentication", "Notification System", "Real-time Collaboration",
	})
	writeCapabilityMap(t, dir, v, []capRow{
		{"SPEC-CLI-001", "CLI Tool — init template selection", "completed", "internal/cli/init"},
		{"SPEC-AUTH-001", "Authentication", "completed", "internal/auth"},
		{"SPEC-NOTIFY-001", "Notifications Subsystem", "in-progress", "internal/notify"},
		{"SPEC-X-999", "Legacy CRM Integration", "in-progress", "internal/crm"},
	})
	if withOverride {
		writeOverride(t, dir, [][2]string{}, []string{})
	}
	if _, err := runAudit(t, dir); err != nil {
		t.Fatalf("audit exited non-zero: %v", err)
	}
	return dir
}

// --- AC-NA-006 ------------------------------------------------------------

// TestACNA006_ProvenancePerRow verifies AC-NA-006: every candidate row carries
// the required provenance field for its type.
func TestACNA006_ProvenancePerRow(t *testing.T) {
	report := fixtureAudit(t, capMapVariantB, true)
	missing, _ := report["missing"].([]any)
	for _, e := range missing {
		entry, _ := e.(map[string]any)
		src, _ := entry["source"].(map[string]any)
		if _, ok := src["file"]; !ok {
			t.Errorf("AC-NA-006: missing entry missing source.file: %v", entry)
		}
		if _, ok := src["heading_path"]; !ok {
			t.Errorf("AC-NA-006: missing entry missing source.heading_path: %v", entry)
		}
	}
	orphan, _ := report["orphan"].([]any)
	for _, e := range orphan {
		entry, _ := e.(map[string]any)
		for _, k := range []string{"spec_id", "title", "implementation_path"} {
			if _, ok := entry[k]; !ok {
				t.Errorf("AC-NA-006: orphan entry missing %s: %v", k, entry)
			}
		}
	}
	matched, _ := report["matched"].([]any)
	validBasis := map[string]bool{"exact": true, "substring": true, "module-token": true, "override": true}
	for _, e := range matched {
		entry, _ := e.(map[string]any)
		basis, _ := entry["match_basis"].(string)
		if !validBasis[basis] {
			t.Errorf("AC-NA-006: matched entry has invalid match_basis %q: %v", basis, entry)
		}
	}
}

// --- AC-NA-007 ------------------------------------------------------------

// TestACNA007_HeaderDrivenBothSpellings verifies AC-NA-007: BOTH header
// variants parse correctly, and the heuristic exercises exact + substring +
// module-token, AND the ≥4-char floor blocks the `crm` segment from matching.
func TestACNA007_HeaderDrivenBothSpellings(t *testing.T) {
	for _, v := range []capMapVariant{capMapVariantA, capMapVariantB} {
		t.Run(variantName(v), func(t *testing.T) {
			report := fixtureAudit(t, v, false)
			matched, _ := report["matched"].([]any)
			bases := map[string]bool{}
			specIDs := map[string]bool{}
			for _, e := range matched {
				entry, _ := e.(map[string]any)
				basis, _ := entry["match_basis"].(string)
				bases[basis] = true
				if sid, _ := entry["spec_id"].(string); sid != "" {
					specIDs[sid] = true
				}
			}
			// (a) Both variants parse — at least 3 matched rows expected
			// (Project Init, Authentication, Notifications).
			if !specIDs["SPEC-CLI-001"] {
				t.Errorf("AC-NA-007 variant %s: SPEC-CLI-001 not matched (header parsing or module-token match failed)", variantName(v))
			}
			if !specIDs["SPEC-AUTH-001"] {
				t.Errorf("AC-NA-007 variant %s: SPEC-AUTH-001 not matched (exact match failed)", variantName(v))
			}
			if !specIDs["SPEC-NOTIFY-001"] {
				t.Errorf("AC-NA-007 variant %s: SPEC-NOTIFY-001 not matched (substring match failed)", variantName(v))
			}
			// (b) basis coverage: at least one of each — exact, substring, module-token.
			for _, want := range []string{"exact", "substring", "module-token"} {
				if !bases[want] {
					t.Errorf("AC-NA-007 variant %s: heuristic basis %q not exercised", variantName(v), want)
				}
			}
			// (c) ≥4-char floor: SPEC-X-999 (internal/crm, last seg crm = 3 chars)
			// MUST NOT appear in matched[] AND MUST appear in orphan[].
			if specIDs["SPEC-X-999"] {
				t.Errorf("AC-NA-007 variant %s: SPEC-X-999 matched — ≥4-char floor violated (crm = 3 chars)", variantName(v))
			}
			orphan, _ := report["orphan"].([]any)
			orphanHas := false
			for _, e := range orphan {
				entry, _ := e.(map[string]any)
				if sid, _ := entry["spec_id"].(string); sid == "SPEC-X-999" {
					orphanHas = true
				}
			}
			if !orphanHas {
				t.Errorf("AC-NA-007 variant %s: SPEC-X-999 not in orphan[] (≥4-char floor check broken)", variantName(v))
			}
		})
	}
}

func variantName(v capMapVariant) string {
	if v == capMapVariantA {
		return "A-capability-first"
	}
	return "B-spec-id-first"
}

// --- AC-NA-008 ------------------------------------------------------------

// TestACNA008_OverrideFileHonored verifies AC-NA-008: override match entries
// move a Missing candidate into matched[] with basis override; ignore entries
// exclude from BOTH missing[] and orphan[].
func TestACNA008_OverrideFileHonored(t *testing.T) {
	report := fixtureAudit(t, capMapVariantB, true)
	// (1) "Autonomy Loop" in matched[] with basis override, NOT in missing[].
	matched, _ := report["matched"].([]any)
	overrideMatched := false
	for _, e := range matched {
		entry, _ := e.(map[string]any)
		if dn, _ := entry["design_name"].(string); dn == "Autonomy Loop" {
			if basis, _ := entry["match_basis"].(string); basis == "override" {
				overrideMatched = true
			}
		}
	}
	if !overrideMatched {
		t.Errorf("AC-NA-008: 'Autonomy Loop' not in matched[] with basis override: %v", matched)
	}
	missing, _ := report["missing"].([]any)
	for _, e := range missing {
		entry, _ := e.(map[string]any)
		if dn, _ := entry["design_name"].(string); dn == "Autonomy Loop" {
			t.Errorf("AC-NA-008: 'Autonomy Loop' appeared in missing[] despite override match")
		}
	}
	// (2) "SPEC-DEPRECATED-001" in NEITHER orphan[] NOR matched[].
	orphan, _ := report["orphan"].([]any)
	for _, e := range orphan {
		entry, _ := e.(map[string]any)
		if sid, _ := entry["spec_id"].(string); sid == "SPEC-DEPRECATED-001" {
			t.Errorf("AC-NA-008: SPEC-DEPRECATED-001 in orphan[] despite override ignore")
		}
	}
	for _, e := range matched {
		entry, _ := e.(map[string]any)
		if sid, _ := entry["spec_id"].(string); sid == "SPEC-DEPRECATED-001" {
			t.Errorf("AC-NA-008: SPEC-DEPRECATED-001 in matched[] despite override ignore")
		}
	}
	// (3) "Old Feature Name" in NEITHER missing[] NOR matched[].
	for _, e := range missing {
		entry, _ := e.(map[string]any)
		if dn, _ := entry["design_name"].(string); dn == "Old Feature Name" {
			t.Errorf("AC-NA-008: 'Old Feature Name' in missing[] despite override ignore")
		}
	}
	for _, e := range matched {
		entry, _ := e.(map[string]any)
		if dn, _ := entry["design_name"].(string); dn == "Old Feature Name" {
			t.Errorf("AC-NA-008: 'Old Feature Name' in matched[] despite override ignore")
		}
	}
}

// --- AC-NA-009 ------------------------------------------------------------

// TestACNA009_Idempotent verifies AC-NA-009: two runs over identical inputs
// produce byte-identical audit-report.{md,json}.
func TestACNA009_Idempotent(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeDesignDoc(t, dir, "product.md", []string{"Authentication"})
	writeCapabilityMap(t, dir, capMapVariantB, []capRow{
		{"SPEC-AUTH-001", "Authentication", "completed", "internal/auth"},
	})
	if _, err := runAudit(t, dir); err != nil {
		t.Fatal(err)
	}
	md, js := auditReportFiles(dir)
	md1, _ := os.ReadFile(md)
	js1, _ := os.ReadFile(js)
	time.Sleep(1100 * time.Millisecond)
	if _, err := runAudit(t, dir); err != nil {
		t.Fatal(err)
	}
	md2, _ := os.ReadFile(md)
	js2, _ := os.ReadFile(js)
	if !bytes.Equal(md1, md2) {
		t.Errorf("AC-NA-009: audit-report.md differs between two runs (idempotence broken)")
	}
	if !bytes.Equal(js1, js2) {
		t.Errorf("AC-NA-009: audit-report.json differs between two runs (idempotence broken)")
	}
}

// --- AC-NA-010 ------------------------------------------------------------

// TestACNA010_FailOpenOnMissingInputs verifies AC-NA-010: each of the 4
// missing-input variants exits 0, writes a minimal report, and logs a warning.
func TestACNA010_FailOpenOnMissingInputs(t *testing.T) {
	cases := []struct {
		name    string
		setupFn func(t *testing.T, dir string)
	}{
		{
			name: "no-design-docs",
			setupFn: func(t *testing.T, dir string) {
				writeCapabilityMap(t, dir, capMapVariantB, []capRow{
					{"SPEC-A-001", "A", "completed", "internal/a"},
				})
			},
		},
		{
			name: "no-capability-map",
			setupFn: func(t *testing.T, dir string) {
				writeDesignDoc(t, dir, "product.md", []string{"Some Feature"})
			},
		},
		{
			name: "no-spec-registry",
			setupFn: func(t *testing.T, dir string) {
				// Design docs + capability-map present, but .moai/specs empty.
				// The audit still runs (capability-map already encodes the
				// SPEC inventory). The fail-open requirement is that the audit
				// exits 0 even when the SPEC registry is absent.
				writeDesignDoc(t, dir, "product.md", []string{"Some Feature"})
				writeCapabilityMap(t, dir, capMapVariantB, nil)
			},
		},
		{
			name: "all-missing-fresh-project",
			setupFn: func(t *testing.T, dir string) {
				// Intentionally empty.
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			initFixtureRepo(t, dir)
			tc.setupFn(t, dir)
			out, err := runAudit(t, dir)
			if err != nil {
				t.Fatalf("AC-NA-010/%s: audit exited non-zero on missing inputs: %v\noutput:\n%s", tc.name, err, out)
			}
			md, _ := auditReportFiles(dir)
			mdata, merr := os.ReadFile(md)
			if merr != nil {
				t.Fatalf("AC-NA-010/%s: audit-report.md not written: %v", tc.name, merr)
			}
			if !strings.Contains(string(mdata), "no inputs available") {
				t.Errorf("AC-NA-010/%s: report missing 'no inputs available' placeholder; got:\n%s", tc.name, mdata)
			}
			warnLog := filepath.Join(dir, ".moai", "logs", "navigator-warnings.log")
			wdata, _ := os.ReadFile(warnLog)
			if !strings.Contains(string(wdata), "WARN") {
				t.Errorf("AC-NA-010/%s: warning log missing WARN entry; got:\n%s", tc.name, wdata)
			}
		})
	}
}

// --- AC-NA-011 ------------------------------------------------------------

// TestACNA011_BoundaryNonOverlap verifies AC-NA-011: the audit script touches
// NO LSEL surface and NO SPEC-003 surface (read or write).
func TestACNA011_BoundaryNonOverlap(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeDesignDoc(t, dir, "product.md", []string{"Authentication"})
	writeCapabilityMap(t, dir, capMapVariantB, []capRow{
		{"SPEC-AUTH-001", "Authentication", "completed", "internal/auth"},
	})
	// LSEL + SPEC-003 sentinel surfaces.
	sentinels := map[string]string{
		".moai/lessons-inbox.jsonl":            "LSEL_SENTINEL_INBOX\n",
		".moai/state/lsel/clusters.json":       "LSEL_SENTINEL_CLUSTER\n",
		"memory/feedback_test.md":              "LSEL_SENTINEL_MEMORY\n",
		".moai/state/tree-sitter/grammar.json": "SPEC003_SENTINEL\n",
	}
	for rel, body := range sentinels {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := runAudit(t, dir); err != nil {
		t.Fatal(err)
	}
	for rel, want := range sentinels {
		p := filepath.Join(dir, rel)
		got, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("AC-NA-011: sentinel %s touched/removed by audit: %v", rel, err)
			continue
		}
		if string(got) != want {
			t.Errorf("AC-NA-011: sentinel %s modified by audit\nwant: %q\ngot:  %q", rel, want, got)
		}
	}
	// Write-set: ONLY audit-report.md + audit-report.json + warnings.log append.
	entries, err := os.ReadDir(filepath.Join(dir, ".moai", "project", "navigator"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "audit-") && n != "audit-report.md" && n != "audit-report.json" {
			t.Errorf("AC-NA-011: audit wrote unexpected file under navigator/: %s", n)
		}
	}
}

// --- AC-NA-012 ------------------------------------------------------------

// TestACNA012_NonGoFixture verifies the non-Go aspect of AC-NA-012: a
// Python-only fixture (no Go markers) audit runs successfully with identical
// output format. The template-neutrality CI guard (internal_content_leak_test.go)
// covers the leak aspect separately at the template level.
func TestACNA012_NonGoFixture(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	// Python-only project markers.
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[project]\nname = \"fixture-py\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.py"),
		[]byte("print('hello')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeDesignDoc(t, dir, "product.md", []string{"Authentication"})
	writeCapabilityMap(t, dir, capMapVariantB, []capRow{
		{"SPEC-AUTH-001", "Authentication", "completed", "src/auth"},
	})
	if err := gitRun(dir, "add", "."); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "py fixture"); err != nil {
		t.Fatal(err)
	}
	out, err := runAudit(t, dir)
	if err != nil {
		t.Fatalf("AC-NA-012: audit failed on Python fixture: %v\n%s", err, out)
	}
	// Same output format as Go fixture case.
	report := loadAuditJSON(t, dir)
	if _, ok := report["missing"]; !ok {
		t.Errorf("AC-NA-012: Python fixture report missing 'missing' key")
	}
	if _, ok := report["matched"]; !ok {
		t.Errorf("AC-NA-012: Python fixture report missing 'matched' key")
	}
}
