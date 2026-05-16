package spec

import (
	"os"
	"testing"
)

// TestStatusGitConsistency_TerminalStateExemption은 SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
// Wave 4 요구사항(REQ-SDF-007)을 검증한다:
// terminal lifecycle state(superseded, archived)를 가진 SPEC은
// git-implied status와 mismatch가 발생해도 StatusGitConsistency finding을 emit하지 않아야 한다.
//
// Pattern D: superseded frontmatter, git-implied 'completed'
// Pattern E: superseded frontmatter, git-implied 'implemented'
// Pattern F: archived frontmatter, git-implied 'implemented'
// Pattern G: archived frontmatter, git-implied 'in-progress'

// TestStatusGitConsistency_SupersededIsTerminal_Completed는 Pattern D를 검증한다.
// frontmatter status='superseded', git-implied='completed' → finding 없음
func TestStatusGitConsistency_SupersededIsTerminal_Completed(t *testing.T) {
	// git 픽스처: sync(docs) commit → git-implied 'completed'
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

// TestStatusGitConsistency_SupersededIsTerminal_Implemented는 Pattern E를 검증한다.
// frontmatter status='superseded', git-implied='implemented' → finding 없음
func TestStatusGitConsistency_SupersededIsTerminal_Implemented(t *testing.T) {
	// git 픽스처: feat commit → git-implied 'implemented'
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

// TestStatusGitConsistency_ArchivedIsTerminal_Implemented는 Pattern F를 검증한다.
// frontmatter status='archived', git-implied='implemented' → finding 없음
func TestStatusGitConsistency_ArchivedIsTerminal_Implemented(t *testing.T) {
	// git 픽스처: feat commit → git-implied 'implemented'
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

// TestStatusGitConsistency_ArchivedIsTerminal_InProgress는 Pattern G를 검증한다.
// frontmatter status='archived', git-implied='in-progress' → finding 없음
func TestStatusGitConsistency_ArchivedIsTerminal_InProgress(t *testing.T) {
	// git 픽스처: ci commit → git-implied 'in-progress'
	// (chore(spec): 은 shouldSkipCommitTitle 로 필터되므로 ci: 를 사용)
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
