// contract_sign_guard_rule_test.go: AC-AP-013 of the contract-sign-guard SPEC.
// Verifies that the paths:-scoped rule documenting the contract-sign guard and
// its internal/template/templates/ mirror both exist, both state that the deny
// is mode-independent (active under `guided`), both state that the companion
// "nothing changes under `guided`" promise is scoped in words to the escalation
// detector, are byte-identical apart from the front-matter `paths:` value, and
// carry no card id, SPEC id, internal date, or commit SHA.
//
// Both files are created by the same change, so every limb reads a surface
// this change owns: a missing file is a RED, never a vacuous pass.
//
// Sentinel on failure: CONTRACT_SIGN_GUARD_RULE_DRIFT
package template_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var contractSignGuardRuleRel = filepath.Join(
	".claude", "rules", "moai", "workflow", "contract-sign-guard.md")

// contractSignGuardRuleStatements are the required mode-independence limbs:
// each must appear in BOTH the local rule file and its template mirror.
var contractSignGuardRuleStatements = []string{
	// The deny is mode-independent and active under guided.
	"mode-independent",
	"guided",
	// The companion promise is scoped in words to the escalation detector.
	"escalation detector",
	"nothing changes under",
	// The guard behavior is documented: deny sentinel and role marker.
	"CONTRACT_SIGN_AGENT_VIOLATION:",
	"MOAI_FACTORY_ROLE",
}

// contractSignGuardNeutralityPatterns are the forbidden internal-content
// classes over BOTH files: SPEC ids, card ids, internal dates, commit SHAs.
var contractSignGuardNeutralityPatterns = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"SPEC id", regexp.MustCompile(`SPEC-[A-Z]`)},
	{"card id", regexp.MustCompile(`\bt[0-9]{3,4}\b`)},
	{"internal date", regexp.MustCompile(`[0-9]{4}-[0-9]{2}-[0-9]{2}`)},
	{"commit SHA", regexp.MustCompile(`\b[0-9a-f]{7,40}\b`)},
}

// TestContractSignGuardRuleDocumentsModeIndependence reads both the local rule
// file and its template mirror from disk and asserts parity and neutrality
// together (AC-AP-013).
func TestContractSignGuardRuleDocumentsModeIndependence(t *testing.T) {
	repoRoot := findProjectRootForMirrorTest(t)

	localPath := filepath.Join(repoRoot, contractSignGuardRuleRel)
	mirrorPath := filepath.Join(repoRoot,
		"internal", "template", "templates", contractSignGuardRuleRel)

	local, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("CONTRACT_SIGN_GUARD_RULE_DRIFT: local rule file unreadable %s: %v", localPath, err)
	}
	mirror, err := os.ReadFile(mirrorPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("CONTRACT_SIGN_GUARD_RULE_DRIFT: template mirror missing at %s; "+
				"run 'cp %s %s' and stage both files (Template-First)",
				mirrorPath, localPath, mirrorPath)
		}
		t.Fatalf("mirror file unreadable %s: %v", mirrorPath, err)
	}

	for _, content := range []struct {
		name string
		data []byte
	}{{"local rule", local}, {"template mirror", mirror}} {
		text := string(content.data)
		for _, statement := range contractSignGuardRuleStatements {
			if !strings.Contains(text, statement) {
				t.Errorf("%s: required statement %q not found", content.name, statement)
			}
		}
		for _, np := range contractSignGuardNeutralityPatterns {
			if m := np.pattern.FindString(text); m != "" {
				t.Errorf("%s: neutrality violation (%s): %q", content.name, np.name, m)
			}
		}
	}

	// Parity limb: byte-identical apart from the front-matter paths: value.
	if norm, normMirror := withoutPathsLine(local), withoutPathsLine(mirror); !bytes.Equal(norm, normMirror) {
		t.Errorf("RULE_TEMPLATE_MIRROR_DRIFT: local rule and template mirror differ apart from "+
			"the front-matter paths: value (local %d bytes, mirror %d bytes after normalization); "+
			"edit both trees in the same commit", len(norm), len(normMirror))
	}
}

// withoutPathsLine strips the front-matter `paths:` line so the two trees
// compare byte-identical apart from their deliberately differing paths values.
func withoutPathsLine(content []byte) []byte {
	var out []byte
	for _, line := range bytes.Split(content, []byte("\n")) {
		if bytes.HasPrefix(bytes.TrimSpace(line), []byte("paths:")) {
			continue
		}
		out = append(out, line...)
		out = append(out, '\n')
	}
	// Drop the trailing newline added after the final line to keep the
	// comparison a pure line-filter of the original bytes.
	if n := len(out); n > 0 && out[n-1] == '\n' {
		out = out[:n-1]
	}
	return out
}
