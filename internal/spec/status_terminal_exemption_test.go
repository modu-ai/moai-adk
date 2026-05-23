package spec

import (
	"os"
	"testing"
)

// TestStatusGitConsistency_TerminalStateExemption verifies the SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
// Wave 4 requirement (REQ-SDF-007):
// SPECs in a terminal lifecycle state (superseded, archived) must not emit a
// StatusGitConsistency finding even when their git-implied status mismatches.
//
// Pattern D: superseded frontmatter, git-implied 'completed'
// Pattern E: superseded frontmatter, git-implied 'implemented'
// Pattern F: archived frontmatter, git-implied 'implemented'
// Pattern G: archived frontmatter, git-implied 'in-progress'

// TestStatusGitConsistency_SupersededIsTerminal_Completed verifies Pattern D.
// frontmatter status='superseded', git-implied='completed' -> no finding
func TestStatusGitConsistency_SupersededIsTerminal_Completed(t *testing.T) {
	// git fixture: sync(docs) commit -> git-implied 'completed'
	dir := setupGitFixture(t, []fixtureCommit{
		{title: "docs(sync): SPEC-TERMINAL-D001 closeout"},
	})

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd 실패: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir 실패: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	doc := &SPECDoc{
		Path: "fake/SPEC-TERMINAL-D001/spec.md",
		Frontmatter: SPECFrontmatter{
			ID:     "SPEC-TERMINAL-D001",
			Status: "superseded",
		},
	}

	rule := &StatusGitConsistencyRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 0 {
		t.Errorf("terminal state 'superseded'는 finding을 emit하지 않아야 한다; got %d finding(s): %v", len(findings), findings)
	}
}

// TestStatusGitConsistency_SupersededIsTerminal_Implemented verifies Pattern E.
// frontmatter status='superseded', git-implied='implemented' -> no finding
func TestStatusGitConsistency_SupersededIsTerminal_Implemented(t *testing.T) {
	// git fixture: feat commit -> git-implied 'implemented'
	dir := setupGitFixture(t, []fixtureCommit{
		{title: "feat(spec): SPEC-TERMINAL-E001 구현 완료"},
	})

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd 실패: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir 실패: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	doc := &SPECDoc{
		Path: "fake/SPEC-TERMINAL-E001/spec.md",
		Frontmatter: SPECFrontmatter{
			ID:     "SPEC-TERMINAL-E001",
			Status: "superseded",
		},
	}

	rule := &StatusGitConsistencyRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 0 {
		t.Errorf("terminal state 'superseded'는 finding을 emit하지 않아야 한다; got %d finding(s): %v", len(findings), findings)
	}
}

// TestStatusGitConsistency_ArchivedIsTerminal_Implemented verifies Pattern F.
// frontmatter status='archived', git-implied='implemented' -> no finding
func TestStatusGitConsistency_ArchivedIsTerminal_Implemented(t *testing.T) {
	// git fixture: feat commit -> git-implied 'implemented'
	dir := setupGitFixture(t, []fixtureCommit{
		{title: "feat(spec): SPEC-TERMINAL-F001 구현 완료"},
	})

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd 실패: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir 실패: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	doc := &SPECDoc{
		Path: "fake/SPEC-TERMINAL-F001/spec.md",
		Frontmatter: SPECFrontmatter{
			ID:     "SPEC-TERMINAL-F001",
			Status: "archived",
		},
	}

	rule := &StatusGitConsistencyRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 0 {
		t.Errorf("terminal state 'archived'는 finding을 emit하지 않아야 한다; got %d finding(s): %v", len(findings), findings)
	}
}

// TestStatusGitConsistency_ArchivedIsTerminal_InProgress verifies Pattern G.
// frontmatter status='archived', git-implied='in-progress' -> no finding
func TestStatusGitConsistency_ArchivedIsTerminal_InProgress(t *testing.T) {
	// git fixture: ci commit -> git-implied 'in-progress'
	// (chore(spec): is filtered by shouldSkipCommitTitle, so use ci: instead)
	dir := setupGitFixture(t, []fixtureCommit{
		{title: "ci: SPEC-TERMINAL-G001 파이프라인 수정"},
	})

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd 실패: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir 실패: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	doc := &SPECDoc{
		Path: "fake/SPEC-TERMINAL-G001/spec.md",
		Frontmatter: SPECFrontmatter{
			ID:     "SPEC-TERMINAL-G001",
			Status: "archived",
		},
	}

	rule := &StatusGitConsistencyRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 0 {
		t.Errorf("terminal state 'archived'는 finding을 emit하지 않아야 한다; got %d finding(s): %v", len(findings), findings)
	}
}
