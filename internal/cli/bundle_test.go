// bundle_test.go — the `moai bundle add|remove` command (SPEC-USER-ASSET-
// INSTALL-001 M2, pulled forward from M3 per the leader's mid-run sequencing
// decision — the recovery command must exist while the catalog exclusion
// takes deployment effect; AC-018).
//
// Every test runs against a temp HOME (plan §D).
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/userassets"
)

// userassetsLoadManifest loads the test HOME's manifest.
func userassetsLoadManifest(t *testing.T) (*userassets.Manifest, error) {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return userassets.Load(userassets.ManifestPath(home))
}

// TestBundleAddInstallsBundleEntries covers AC-018's add arm at the command
// level: `moai bundle add <name>` installs exactly that bundle's catalog
// entries and records the selection in the manifest.
func TestBundleAddInstallsBundleEntries(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))

	cmd := newBundleTestCmd()
	if err := runBundleAdd(cmd, []string{"ops-tools"}); err != nil {
		t.Fatalf("bundle add ops-tools: %v", err)
	}
	home, _ := os.UserHomeDir()
	// ops-tools carries moai-workflow-loop (an unselected-bundle skill before
	// this call — the leader's P2 reproduction now has its remedy).
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "moai-workflow-loop", "SKILL.md")); err != nil {
		t.Errorf("bundle add did not land moai-workflow-loop: %v", err)
	}
	m, err := userassetsLoadManifest(t)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, b := range m.Bundles {
		if b == "ops-tools" {
			found = true
		}
	}
	if !found {
		t.Errorf("manifest bundle selection = %v, want ops-tools recorded", m.Bundles)
	}
}

// TestBundleRemoveTakesComplementOnly covers the E3 shared-asset arm: removing
// the devops pack keeps its L0 trio (owasp-checklist / cross-model-audit /
// secops — catalog members of core AND devops) with report notes, and removes
// only the complement (llm-security, supply-chain — neither is a declared
// dependency of any preserved non-dispatcher entry).
func TestBundleRemoveTakesComplementOnly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))

	cmd := newBundleTestCmd()
	if err := runBundleAdd(cmd, []string{"devops"}); err != nil {
		t.Fatalf("bundle add devops: %v", err)
	}
	if err := runBundleRemove(cmd, []string{"devops"}); err != nil {
		t.Fatalf("bundle remove devops: %v", err)
	}
	home, _ := os.UserHomeDir()

	// The E3 shared trio survives.
	for _, skill := range []string{"moai-ref-owasp-checklist", "moai-ref-cross-model-audit", "moai-ref-secops"} {
		if _, err := os.Stat(filepath.Join(home, ".claude", "skills", skill, "SKILL.md")); err != nil {
			t.Errorf("E3 shared L0 skill %s was deleted by bundle remove: %v", skill, err)
		}
	}
	// The complement is gone.
	for _, skill := range []string{"moai-ref-llm-security", "moai-ref-supply-chain"} {
		if _, err := os.Stat(filepath.Join(home, ".claude", "skills", skill, "SKILL.md")); !os.IsNotExist(err) {
			t.Errorf("complement skill %s survived the bundle remove: %v", skill, err)
		}
	}
	// The selection no longer names devops.
	m, err := userassetsLoadManifest(t)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range m.Bundles {
		if b == "devops" {
			t.Errorf("manifest still records devops after remove: %v", m.Bundles)
		}
	}
}

// TestBundleRemoveHonorsDivergence covers the REQ-023 removal arm at the
// command level: a tracked file whose hash matches neither its manifest hash
// nor the shipped bytes is preserved and reported, never deleted.
func TestBundleRemoveHonorsDivergence(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))

	cmd := newBundleTestCmd()
	if err := runBundleAdd(cmd, []string{"ops-tools"}); err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	target := filepath.Join(home, ".claude", "skills", "moai-workflow-loop", "SKILL.md")
	edited := []byte("user edit\n")
	if err := os.WriteFile(target, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runBundleRemove(cmd, []string{"ops-tools"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("divergent file removed: %v", err)
	}
	if string(data) != string(edited) {
		t.Errorf("divergent file rewritten: %q", data)
	}
}

// TestBundleUnknownNameRefused covers C4: an unknown bundle name refuses with
// an actionable error naming the valid bundles.
func TestBundleUnknownNameRefused(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))

	cmd := newBundleTestCmd()
	err := runBundleAdd(cmd, []string{"not-a-bundle"})
	if err == nil {
		t.Fatal("unknown bundle name accepted")
	}
	if !strings.Contains(err.Error(), "not-a-bundle") {
		t.Errorf("refusal does not name the offending bundle: %v", err)
	}
}

// TestBundleRemoveDefersDependency covers R-f-② at the command level:
// removing a bundle whose entry is a declared dependency of a PRESERVED
// asset defers the deletion (kept + reported). The case: `commands` stays
// opted in; its moai-e2e entry declares e2e-tester (consult) as an agent
// dependency — removing consult must DEFER e2e-tester's deletion while the
// other consult entries go. The dispatcher's own dep list is EXCLUDED from
// the deferral rule (its bundle-row edges are the derivation matrix's
// conditional class — their absence is handled by the remediation/refusal
// pattern, and including them would defeat the D28 selection-based prune:
// AC-018's shipped-but-deselected arm requires the prune to function).
func TestBundleRemoveDefersDependency(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", os.Getenv("HOME"))

	cmd := newBundleTestCmd()
	if err := runBundleAdd(cmd, []string{"commands"}); err != nil {
		t.Fatal(err)
	}
	if err := runBundleAdd(cmd, []string{"consult"}); err != nil {
		t.Fatal(err)
	}
	if err := runBundleRemove(cmd, []string{"consult"}); err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	// DEFERRED: e2e-tester is declared by the preserved moai-e2e.
	if _, err := os.Stat(filepath.Join(home, ".claude", "agents", "e2e-tester.md")); err != nil {
		t.Errorf("dependency-of-preserved deletion was not deferred: %v", err)
	}
	// REMOVED: no preserved entry declares these.
	for _, agent := range []string{"manager-design", "manager-todo", "super-advisor"} {
		if _, err := os.Stat(filepath.Join(home, ".claude", "agents", agent+".md")); !os.IsNotExist(err) {
			t.Errorf("non-dependency consult agent %s survived the removal: %v", agent, err)
		}
	}
}
