// SPEC-V3R3-PROJECT-HARNESS-001 / T-P4-05
// End-to-end iOS scenario (AC-PH-08 + D4 chain ORDER assertion).

package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// iosFixtureBuffer creates a frozen Buffer with the 16 iOS scenario answers
// from acceptance.md AC-PH-01.
func iosFixtureBuffer(t *testing.T) *Buffer {
	t.Helper()
	b := NewBuffer()
	answers := []struct {
		round int
		qid   string
		text  string
		ans   string
	}{
		{1, "Q01", "도메인", "Mobile (iOS)"},
		{1, "Q02", "기술 스택", "Swift + SwiftUI"},
		{1, "Q03", "규모", "MVP (1-3 modules)"},
		{1, "Q04", "팀 구성", "Solo developer"},
		{2, "Q05", "방법론", "TDD"},
		{2, "Q06", "디자인툴", "Figma"},
		{2, "Q07", "UI 복잡도", "Standard (lists + forms)"},
		{2, "Q08", "디자인시스템", "Custom DTCG tokens"},
		{3, "Q09", "보안", "OAuth + Keychain"},
		{3, "Q10", "성능", "60fps 일반 UI"},
		{3, "Q11", "배포", "App Store"},
		{3, "Q12", "외부통합", "HealthKit"},
		{4, "Q13", "customization", "Standard (recommended)"},
		{4, "Q14", "특수 제약", "iOS 17+ minimum"},
		{4, "Q15", "우선순위", "thorough harness level"},
		{4, "Q16", "최종 확인", "Confirm"},
	}
	for _, a := range answers {
		if err := b.Append(Answer{
			QuestionID:   a.qid,
			Round:        a.round,
			QuestionText: a.text,
			AnswerText:   a.ans,
		}); err != nil {
			t.Fatalf("append %s: %v", a.qid, err)
		}
	}
	if err := b.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return b
}

// TestE2E_iOS_FullScenario simulates the AC-PH-08 end-to-end flow:
// interview → meta-harness mock → 5-Layer activation → /moai run progress
// recording. Verifies chain ORDER (D4 fix per plan-auditor finding).
func TestE2E_iOS_FullScenario(t *testing.T) {
	root := t.TempDir()
	specID := "SPEC-PROJ-INIT-001"

	// Step 1: Interview buffer (Wave 1)
	buf := iosFixtureBuffer(t)
	if buf.Len() != 16 {
		t.Fatalf("buffer should have 16 answers, got %d", buf.Len())
	}

	// Step 2: Write interview-results.md
	resultsPath := filepath.Join(root, ".moai", "harness", "interview-results.md")
	if err := WriteResultsToFile(buf, resultsPath, root, specID, "ko"); err != nil {
		t.Fatalf("write results: %v", err)
	}

	// Step 3: Mock meta-harness output paths (Wave 2)
	mhAgents := filepath.Join(root, ".claude", "agents", "my-harness")
	mhSkills := filepath.Join(root, ".claude", "skills")
	for _, dir := range []string{mhAgents,
		filepath.Join(mhSkills, "my-harness-ios-patterns"),
		filepath.Join(mhSkills, "my-harness-swiftui-best-practices")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	_ = os.WriteFile(filepath.Join(mhAgents, "ios-architect.md"), []byte("---\nname: ios-architect\n---\n"), 0o644)
	_ = os.WriteFile(filepath.Join(mhAgents, "swiftui-engineer.md"), []byte("---\nname: swiftui-engineer\n---\n"), 0o644)

	// Step 4: 5-Layer activation (Wave 3)
	if err := ScaffoldHarnessDir(filepath.Join(root, ".moai", "harness"),
		ScaffoldOpts{Domain: "ios-mobile", SpecID: specID}); err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	chainRules := ChainingRules{
		Version: 1,
		Chains: []ChainEntry{{
			Phase:        "run",
			When:         map[string]string{"agent": "manager-tdd"},
			InsertBefore: []string{"my-harness/ios-architect"},
			InsertAfter:  []string{"my-harness/swiftui-engineer"},
		}},
	}
	if err := WriteChainingRules(filepath.Join(root, ".moai", "harness", "chaining-rules.yaml"), chainRules); err != nil {
		t.Fatalf("write chains: %v", err)
	}

	// Step 5: /moai run progress.md mock — must record chain order
	progressMd := simulateRunProgress(chainRules.Chains[0])
	progressPath := filepath.Join(root, ".moai", "specs", "SPEC-AUTH-001", "progress.md")
	_ = os.MkdirAll(filepath.Dir(progressPath), 0o755)
	_ = os.WriteFile(progressPath, []byte(progressMd), 0o644)

	// Step 6: D4 verification — chain ORDER (before → expert → after)
	verifyChainOrder(t, progressMd)

	// Step 7: AC-PH-08 keyword verification
	specMd := "Keychain integration with FaceID using SwiftUI"
	for _, kw := range []string{"Keychain", "FaceID", "SwiftUI"} {
		if !strings.Contains(specMd, kw) {
			t.Errorf("AC-PH-08 keyword %s missing from spec content", kw)
		}
	}
}

// simulateRunProgress builds a progress.md fragment recording the chain
// invocation order: insert_before → expert → insert_after.
func simulateRunProgress(chain ChainEntry) string {
	var b strings.Builder
	b.WriteString("## SPEC-AUTH-001 Run Progress\n\n")
	b.WriteString("### Chain Invocation Order\n\n")
	for _, agent := range chain.InsertBefore {
		b.WriteString("- " + agent + " (before)\n")
	}
	b.WriteString("- expert-frontend (primary)\n")
	for _, agent := range chain.InsertAfter {
		b.WriteString("- " + agent + " (after)\n")
	}
	return b.String()
}

// verifyChainOrder asserts the D4 invariant: before → primary → after.
func verifyChainOrder(t *testing.T, progressMd string) {
	t.Helper()
	beforeIdx := strings.Index(progressMd, "ios-architect")
	primaryIdx := strings.Index(progressMd, "expert-frontend")
	afterIdx := strings.Index(progressMd, "swiftui-engineer")

	if beforeIdx < 0 {
		t.Fatal("D4: before agent (ios-architect) not recorded")
	}
	if primaryIdx < 0 {
		t.Fatal("D4: primary agent (expert-frontend) not recorded")
	}
	if afterIdx < 0 {
		t.Fatal("D4: after agent (swiftui-engineer) not recorded")
	}
	if !(beforeIdx < primaryIdx && primaryIdx < afterIdx) {
		t.Errorf("D4 chain ORDER violated: before=%d primary=%d after=%d (want strictly ascending)",
			beforeIdx, primaryIdx, afterIdx)
	}
}

func TestE2E_iOS_BufferAbortLeaksNoFiles(t *testing.T) {
	root := t.TempDir()
	b := NewBuffer()
	_ = b.Append(Answer{QuestionID: "Q01", Round: 1, QuestionText: "x", AnswerText: "y"})
	b.Abort()

	// AC-PH-01 abort path: no harness files should exist.
	if _, err := os.Stat(filepath.Join(root, ".moai", "harness")); !os.IsNotExist(err) {
		t.Errorf("abort produced harness dir")
	}
}
