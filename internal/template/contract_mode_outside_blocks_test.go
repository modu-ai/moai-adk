// contract_mode_outside_blocks_test.go — card t2009: a base-ref guard that can
// judge a base which ALREADY carries contract-mode blocks.
//
// TestContractModeGuidedPreservation requires the current copy with its blocks
// stripped to equal the base copy byte for byte, which presumes a base that has
// no blocks. Against a base that already carries them (any main after the
// rewiring landed) it fails on every block-bearing copy whatever was edited, so
// it cannot say whether a later change touched text outside the blocks. This
// companion strips the blocks from BOTH sides, so the same question stays
// decidable: no outside-block change passes, an outside-block change fails. It
// does not move the base, widen any list, or replace the original guard.
package template_test

import (
	"strings"
	"testing"
)

func TestContractModeOutsideBlocksVsBase(t *testing.T) {
	block := grBlockText("a", grCondition+" x")

	t.Run("falsifier/unchanged-with-blocks-on-both-sides-passes", func(t *testing.T) {
		base := "alpha\n" + block + "\nbeta\n"
		cur := base
		if f := grPreservationFindings("fixture.md", grStrip(cur), grStrip(base)); len(f) != 0 {
			t.Fatalf("an unchanged copy was rejected against a base that already carries the block: %v", f)
		}
	})
	t.Run("falsifier/block-body-change-passes", func(t *testing.T) {
		base := "alpha\n" + block + "\nbeta\n"
		cur := "alpha\n" + grBlockText("a", grCondition+" y") + "\nbeta\n"
		if f := grPreservationFindings("fixture.md", grStrip(cur), grStrip(base)); len(f) != 0 {
			t.Fatalf("a change inside the block was reported as an outside-block change: %v", f)
		}
	})
	t.Run("falsifier/outside-change-fails", func(t *testing.T) {
		base := "alpha\n" + block + "\nbeta\n"
		cur := "alpha\n" + block + "\nbetta\n"
		if f := grPreservationFindings("fixture.md", grStrip(cur), grStrip(base)); len(f) == 0 {
			t.Fatal("a changed word outside the block was accepted")
		}
	})
	t.Run("tree", func(t *testing.T) {
		base := grBase(t)
		root := grRoot(t)
		stripped := 0
		for _, rel := range grAllCopies() {
			cur := grRead(t, root, rel)
			blocks, _ := grParse(rel, cur)
			stripped += len(blocks)
			baseText, ok := grShow(t, root, base, rel)
			if !ok {
				t.Errorf("%s: absent at the base ref", rel)
				continue
			}
			got := grStrip(cur)
			want := grStrip(baseText)
			// A relocation is undone only when the base predates it. A base that
			// already carries the pointer line holds the relocated form, so
			// restoring the moved lines would manufacture a difference.
			if r, prefix, ok := grRelocationFor(rel); ok && !strings.Contains(want, r.pointer) {
				var rf []string
				got, rf = grRestore(rel, got, grRead(t, root, prefix+r.dst), r)
				for _, f := range rf {
					t.Error(f)
				}
			}
			for _, f := range grPreservationFindings(rel, got, want) {
				t.Error(f)
			}
		}
		if stripped == 0 {
			t.Fatal("no block was stripped — empty sweep")
		}
		t.Logf("compared the outside-block text of %d copies (%d blocks stripped on the current side)", len(grAllCopies()), stripped)
	})
}
