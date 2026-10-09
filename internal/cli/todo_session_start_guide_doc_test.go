// todo_session_start_guide_doc_test.go — SPEC-SESSION-START-GUIDE-I18N-001
// M3 (AC-007): the doctrine surfaces that describe the standing spawn
// authority agree with the shipped runtime behavior — the SessionStart join
// notice renders the authority sentence per conversation_language (the
// canonical English sentence for `en`), no longer English-only, with the
// Status Transition Ownership Matrix pointer path verbatim in every locale.
//
// The test pins the TEMPLATE mirrors (the deployed copies). The live copies
// under .claude/rules/ at the repository root are deliberately not read here:
// this card ships the template tree, and the live harness copies follow
// through the mirror-sync pass (the divergence is recorded in the card's
// progress record, not silently assumed).
//
// Pinned literals are checked on the RAW text (the todo_auto_doc_test.go
// precedent): a phrase that wraps across a hard line break fails the literal
// grep. Unpinned clauses compare on whitespace-normalized text.

package cli

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// guideLocalityPhrase is the wording the amended doctrine surfaces carry in
// the standing-authority passage: the sentence renders per locale instead of
// English-only. Both surfaces carry it, so a mutant that reverts either
// surface to the English-only claim (or deletes the locality statement)
// fails.
const guideLocalityPhrase = "localized per `conversation_language`"

// guideAuthorityMatrixPath is the matrix pointer path the doctrine passage
// must keep verbatim (same path the runtime authority sentence carries in
// every locale).
const guideAuthorityMatrixPath = ".claude/rules/moai/development/spec-frontmatter-schema.md"

// guideDocStaleClaims are the pre-amendment claims, quoted verbatim from the
// passages they used to live in. Each asserts the authority sentence is
// carried verbatim — true only under the English-only rule this card amended.
// A mutant that reverts the wording keeps these and fails.
var guideDocStaleClaims = []string{
	"the SessionStart join notice carries it verbatim",
	"the SessionStart join notice carries the authority sentence verbatim",
}

// TestAutoRankSessionStartGuideAuthorityParity (AC-007, M3): both template
// mirrors of the standing-spawn-authority doctrine describe the localized
// render — the locality phrase and the matrix pointer path are present, the
// pre-amendment English-only claims are gone, and the two surfaces agree on
// the locality wording. The amended lines stay template-neutral (no SPEC id,
// requirement token, ISO date, or hex run).
func TestAutoRankSessionStartGuideAuthorityParity(t *testing.T) {
	root := autoDocRepoRoot(t)

	surfaces := []autoDocSurface{
		{"template factory-dispatch.md", filepath.Join("internal", "template", "templates", ".claude", "rules", "moai", "workflow", "factory-dispatch.md")},
		{"template factory-dispatch-detail.md", filepath.Join("internal", "template", "templates", ".claude", "rules", "moai", "workflow", "factory-dispatch-detail.md")},
	}

	// The neutrality contract of the template tree (same expression family as
	// the sibling guards), applied to the changed passage: extract the text
	// between the passage's opening bold marker and the end of the line that
	// carries the locality phrase.
	neutral := regexp.MustCompile(`SPEC-[A-Z0-9-]+-[0-9]{3}|REQ-[A-Z]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b`)

	var normalizedPassages []string
	for _, s := range surfaces {
		doc := autoDocRead(t, root, s.path)

		t.Run("locality phrase present/"+s.name, func(t *testing.T) {
			if !strings.Contains(doc, guideLocalityPhrase) {
				t.Errorf("%s does not state the localized render %q", s.name, guideLocalityPhrase)
			}
		})
		t.Run("matrix path verbatim/"+s.name, func(t *testing.T) {
			if !strings.Contains(doc, guideAuthorityMatrixPath) {
				t.Errorf("%s lost the matrix pointer path %q", s.name, guideAuthorityMatrixPath)
			}
		})
		for _, stale := range guideDocStaleClaims {
			t.Run("stale English-only claim absent/"+s.name, func(t *testing.T) {
				if strings.Contains(doc, stale) {
					t.Errorf("%s still carries the pre-amendment English-only claim %q", s.name, stale)
				}
			})
		}

		// Neutrality of the amended line(s): the line carrying the locality
		// phrase must stay free of internal content tokens.
		t.Run("amended line neutral/"+s.name, func(t *testing.T) {
			for i, line := range strings.Split(doc, "\n") {
				if strings.Contains(line, guideLocalityPhrase) {
					if hit := neutral.FindString(line); hit != "" {
						t.Errorf("%s amended line %d carries internal content %q", s.name, i+1, hit)
					}
					normalizedPassages = append(normalizedPassages, autoDocNormalize(line))
					break
				}
			}
		})
	}

	// The two mirror surfaces agree on the locality wording once normalized —
	// a rewording on one surface only splits the doctrine the same way a
	// live/mirror divergence would. The normalized form drops backticks, so
	// the phrase is compared in its normalized form too.
	t.Run("surfaces agree on the locality wording", func(t *testing.T) {
		if len(normalizedPassages) != len(surfaces) {
			t.Fatalf("locality phrase found on %d of %d surfaces", len(normalizedPassages), len(surfaces))
		}
		wantPhrase := autoDocNormalize(guideLocalityPhrase)
		for _, p := range normalizedPassages[1:] {
			if !strings.Contains(p, wantPhrase) {
				t.Errorf("a surface lost the locality phrase after normalization: %q", p)
			}
		}
	})
}
