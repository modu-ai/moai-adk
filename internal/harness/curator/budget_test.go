package curator

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// --- AC-HEV2-012: WriteManagedBlock rejects a block whose rendered bullet
// body exceeds the 3,000-char digest budget, returning ErrDigestBudgetExceeded
// WITHOUT touching the file (REQ-HEV2-008). ---

func TestWriteManagedBlock_BudgetExceeded_ErrDigestBudgetExceeded(t *testing.T) {
	path := writeFixture(t, "# Project\n")
	before := mustRead(t, path)

	// Build a bullet body that exceeds MaxDigestBlockChars (3000) while
	// staying within the bullet cap (MaxDigestBullets = 20). 20 bullets × 160
	// chars text → body = 20 × (2 + 160 + 1) = 3260 > 3000.
	bullets := make([]Bullet, 0, MaxDigestBullets)
	chunk := strings.Repeat("x", 160)
	for i := 0; i < MaxDigestBullets; i++ {
		bullets = append(bullets, Bullet{Text: chunk})
	}
	if bodyLen := len(renderBullets(bullets)); bodyLen <= MaxDigestBlockChars {
		t.Fatalf("fixture mis-sized: body=%d must exceed %d", bodyLen, MaxDigestBlockChars)
	}

	err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{Bullets: bullets})

	if !errors.Is(err, ErrDigestBudgetExceeded) {
		t.Fatalf("expected ErrDigestBudgetExceeded, got %v", err)
	}

	// The file MUST be untouched (REQ-HEV2-008: "NOT touch the file").
	after := mustRead(t, path)
	if string(before) != string(after) {
		t.Errorf("file modified despite budget rejection:\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
}

// --- AC-HEV2-012 (boundary): a block whose body is exactly at the budget
// (3000 chars) is ADMITTED; one byte over is rejected. ---

func TestWriteManagedBlock_BudgetBoundary(t *testing.T) {
	t.Run("exactly_at_budget_admitted", func(t *testing.T) {
		path := writeFixture(t, "# Project\n")
		// Body == renderBullets output. Each bullet line = "- " + text + "\n"
		// (3 bytes overhead). To hit exactly MaxDigestBlockChars (3000) with
		// 20 bullets: text length = 3000/20 - 3 = 147.
		text := strings.Repeat("a", 147)
		bullets := make([]Bullet, 0, MaxDigestBullets)
		for i := 0; i < MaxDigestBullets; i++ {
			bullets = append(bullets, Bullet{Text: text})
		}
		if bodyLen := len(renderBullets(bullets)); bodyLen != MaxDigestBlockChars {
			t.Fatalf("test fixture mis-sized: body=%d want=%d", bodyLen, MaxDigestBlockChars)
		}
		if err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{Bullets: bullets}); err != nil {
			t.Fatalf("block at exact budget should be admitted, got %v", err)
		}
	})

	t.Run("one_byte_over_rejected", func(t *testing.T) {
		path := writeFixture(t, "# Project\n")
		// 19 bullets of 147 chars + 1 bullet of 148 chars → body = 3001
		// (stays within the 20-bullet cap so the budget check, not the cap,
		// is what rejects).
		bullets := make([]Bullet, 0, MaxDigestBullets)
		for i := 0; i < 19; i++ {
			bullets = append(bullets, Bullet{Text: strings.Repeat("a", 147)})
		}
		bullets = append(bullets, Bullet{Text: strings.Repeat("a", 148)})
		if bodyLen := len(renderBullets(bullets)); bodyLen != MaxDigestBlockChars+1 {
			t.Fatalf("fixture mis-sized: body=%d want=%d", bodyLen, MaxDigestBlockChars+1)
		}
		err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{Bullets: bullets})
		if !errors.Is(err, ErrDigestBudgetExceeded) {
			t.Fatalf("block one byte over budget should be rejected, got %v", err)
		}
	})
}

// --- AC-HEV2-013: measureAlwaysLoaded per-section attribution — the
// MOAI:LEARNED-WORKFLOW block's contribution is measurable distinctly via
// config.MeasureAlwaysLoadedSection (REQ-HEV2-008 anti-fabrication: budget
// enforcement must verify the ACTUAL measured contribution, not an assumed
// value). ---

func TestMeasureAlwaysLoaded_PerSectionAttribution(t *testing.T) {
	// Write a CLAUDE.md with a populated LearnedWorkflow block carrying a
	// known body of distilled bullets.
	dir := t.TempDir()
	claudeMd := filepath.Join(dir, "CLAUDE.md")
	body := strings.Join([]string{
		"# Project",
		"",
		"## MOAI:LEARNED-WORKFLOW",
		"<!-- moai:learned-start -->",
		"- distilled rule one",
		"- distilled rule two",
		"<!-- moai:learned-end -->",
		"",
		"## Other Section",
		"",
		"Unrelated prose that is NOT part of the digest block.",
		"",
	}, "\n")
	if err := os.WriteFile(claudeMd, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	spec := markerRegistry[BlockTypeLearnedWorkflow]
	chars, tokens, found, err := config.MeasureAlwaysLoadedSection(
		claudeMd, spec.startPrefix, spec.endMarker)
	if err != nil {
		t.Fatalf("MeasureAlwaysLoadedSection: %v", err)
	}
	if !found {
		t.Fatal("section not found; per-section attribution failed to locate the block")
	}

	// The section body is the bytes strictly between startMarker and endMarker.
	// For the fixture above that is:
	//   " -->\n- distilled rule one\n- distilled rule two\n"
	// (the remainder of the start comment line through the last bullet's newline).
	wantFragments := []string{"distilled rule one", "distilled rule two"}
	gotSection := body[strings.Index(body, spec.startPrefix)+len(spec.startPrefix):]
	gotSection = gotSection[:strings.Index(gotSection, spec.endMarker)]
	if chars != len(gotSection) {
		t.Errorf("chars = %d, want %d (actual section byte length)", chars, len(gotSection))
	}
	for _, frag := range wantFragments {
		if !strings.Contains(gotSection, frag) {
			t.Errorf("attributed section missing %q", frag)
		}
	}
	if tokens != chars/4 {
		t.Errorf("tokens = %d, want %d (chars/4)", tokens, chars/4)
	}

	// The "Unrelated prose" outside the block MUST NOT be counted in the
	// attributed section (distinct attribution).
	if strings.Contains(gotSection, "Unrelated prose") {
		t.Error("per-section attribution leaked prose from outside the block")
	}
}

// --- AC-HEV2-013 (negative): attribution returns found=false for an absent
// block (graceful, no error). ---

func TestMeasureAlwaysLoaded_SectionAbsent(t *testing.T) {
	dir := t.TempDir()
	claudeMd := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(claudeMd, []byte("# Project\nNo block here.\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	spec := markerRegistry[BlockTypeLearnedWorkflow]
	chars, tokens, found, err := config.MeasureAlwaysLoadedSection(
		claudeMd, spec.startPrefix, spec.endMarker)
	if err != nil {
		t.Fatalf("unexpected error on absent section: %v", err)
	}
	if found {
		t.Error("expected found=false for absent block")
	}
	if chars != 0 || tokens != 0 {
		t.Errorf("absent section should report 0/0, got chars=%d tokens=%d", chars, tokens)
	}
}

// --- AC-HEV2-014: WriteManagedBlock rejects a block whose bullet count
// exceeds MaxDigestBullets (20), returning ErrBulletCapExceeded WITHOUT
// touching the file (REQ-HEV2-009). ---

func TestWriteManagedBlock_BulletCapExceeded(t *testing.T) {
	path := writeFixture(t, "# Project\n")
	before := mustRead(t, path)

	// 21 bullets — one over the cap.
	bullets := make([]Bullet, 0, MaxDigestBullets+1)
	for i := 0; i < MaxDigestBullets+1; i++ {
		bullets = append(bullets, Bullet{Text: "short rule"})
	}

	err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{Bullets: bullets})
	if !errors.Is(err, ErrBulletCapExceeded) {
		t.Fatalf("expected ErrBulletCapExceeded for %d bullets, got %v", len(bullets), err)
	}

	// The file MUST be untouched.
	after := mustRead(t, path)
	if string(before) != string(after) {
		t.Error("file modified despite bullet-cap rejection")
	}
}

// --- AC-HEV2-014 (boundary): exactly 20 bullets admitted, 21 rejected. ---

func TestWriteManagedBlock_BulletCapBoundary(t *testing.T) {
	t.Run("exactly_20_admitted", func(t *testing.T) {
		path := writeFixture(t, "# Project\n")
		bullets := make([]Bullet, 0, MaxDigestBullets)
		for i := 0; i < MaxDigestBullets; i++ {
			bullets = append(bullets, Bullet{Text: "rule"})
		}
		if err := WriteManagedBlock(path, BlockTypeLearnedWorkflow, BlockContent{Bullets: bullets}); err != nil {
			t.Fatalf("exactly %d bullets should be admitted, got %v", MaxDigestBullets, err)
		}
	})
}
