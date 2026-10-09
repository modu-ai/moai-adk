package wizard

// The coexistence guard (SPEC participation REQ-023, AC-024): asking for
// consent to a feature that cannot act would record a consent the user
// cannot exercise. The wizard's participation question may therefore ship
// ONLY alongside the sender package — the guard fails when the question
// exists while internal/feedback/publish does not. The release-ordering
// rationale: a cherry-pick of the consent question alone onto a tree
// without the sender trips this guard at integration.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// participationGuardViolation reports why the wizard's participation
// question cannot ship against the given sender package directory; nil
// means the pair coexists (or no question is asked, which needs no sender).
func participationGuardViolation(questionIDs []string, senderDir string) error {
	asked := false
	for _, id := range questionIDs {
		if id == ParticipationQuestionID {
			asked = true
			break
		}
	}
	if !asked {
		return nil
	}
	entries, err := os.ReadDir(senderDir)
	if err != nil {
		return fmt.Errorf("participation question ships but the sender package is unreadable at %s: %w", senderDir, err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			return nil // the sender exists: the pair coexists
		}
	}
	return fmt.Errorf("participation question ships but no sender implementation exists under %s — asking for consent to a feature that cannot act records a consent the user cannot exercise", senderDir)
}

// wizardParticipationQuestionIDs returns the IDs of the questions the init
// wizard actually asks (the real tree's source of the question).
func wizardParticipationQuestionIDs(t *testing.T) []string {
	t.Helper()
	questions := Page3Questions("")
	ids := make([]string, 0, len(questions))
	for _, q := range questions {
		ids = append(ids, q.ID)
	}
	return ids
}

// TestParticipationQuestionRequiresSender: on the real tree the question
// and the sender package coexist — the guard passes.
func TestParticipationQuestionRequiresSender(t *testing.T) {
	senderDir := filepath.Join("..", "..", "feedback", "publish")
	if err := participationGuardViolation(wizardParticipationQuestionIDs(t), senderDir); err != nil {
		t.Fatalf("the real tree fails the coexistence guard: %v", err)
	}
}

// TestParticipationQuestionGuardCatchesMissingSender: the guard is not
// vacuous — a synthetic tree that carries the question but lacks the sender
// fails it. The red is observed here, before the guard is adopted.
func TestParticipationQuestionGuardCatchesMissingSender(t *testing.T) {
	// The synthetic tree: the question IS present (the real wizard's set),
	// the sender directory is an empty temporary stand-in.
	err := participationGuardViolation([]string{ParticipationQuestionID}, t.TempDir())
	if err == nil {
		t.Fatal("the guard passed a synthetic tree that asks the participation question with no sender package — it proves nothing")
	}
	// The no-question arm: without the question, no sender is required.
	if err := participationGuardViolation([]string{"other_question"}, t.TempDir()); err != nil {
		t.Fatalf("the guard failed a tree that asks no participation question: %v", err)
	}
}
