// session_start_guide_i18n_render_test.go — SPEC-SESSION-START-GUIDE-I18N-001
// M4 (AC-003, AC-006, AC-008; the M1 content requirements REQ-001/002/005 in
// their full-table form): the four new guide fields (laneSpawnAuthority,
// autoModeGuideLeader, autoModeGuideLane, docsPointer) are complete in every
// locale, carry their pinned literals verbatim, assemble into the notices
// without breaking the layout invariants, and the English fallback delivers a
// complete notice for every advertised-but-untranslated locale, an empty
// language, and a nil configuration.
//
// The three TestRed tests carry the en-side block-scoped assertions (AC-001,
// AC-002, AC-005); this file extends the same predicates to the whole table
// and to the fallback contract.

package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// guideEnGuideFragment is an English-only clause of the en auto-mode guide; a
// ko/ja/zh field carrying it is a mixed-locale block (REQ-009).
const guideEnGuideFragment = "the two-mode split names"

// TestFactoryGuideNewFieldsCompleteInEveryLocale (REQ-001/REQ-002/REQ-006,
// acceptance §D.5 edge): every locale's four new fields are non-empty, the
// authority entries carry the matrix pointer path verbatim, en keeps the
// canonical sentence, and the ko/ja/zh authority and guide fields carry no
// English sentence prose (the protocol tokens are exempt).
func TestFactoryGuideNewFieldsCompleteInEveryLocale(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		fm := factoryMessagesFor(lang)
		for name, field := range map[string]string{
			"laneSpawnAuthority":  fm.laneSpawnAuthority,
			"autoModeGuideLeader": fm.autoModeGuideLeader,
			"autoModeGuideLane":   fm.autoModeGuideLane,
			"docsPointer":         fm.docsPointer,
		} {
			if strings.TrimSpace(field) == "" {
				t.Errorf("%s: locale %q field %s is empty — a single empty locale entry renders an empty block, not the fallback", lang, lang, name)
			}
		}

		if !strings.Contains(fm.laneSpawnAuthority, factoryMatrixPath) {
			t.Errorf("%s laneSpawnAuthority lost the matrix pointer path %q", lang, factoryMatrixPath)
		}
		for _, field := range []string{fm.autoModeGuideLeader, fm.autoModeGuideLane} {
			for _, want := range []string{"--auto-leader", "--auto-lane", designSurfaceMarker} {
				if !strings.Contains(field, want) {
					t.Errorf("%s auto-mode guide missing %q:\n%s", lang, want, field)
				}
			}
		}
		if !strings.Contains(fm.docsPointer, "https://adk.mo.ai.kr") {
			t.Errorf("%s docsPointer lost the docs URL:\n%s", lang, fm.docsPointer)
		}

		// REQ-009 mixed-locale proxy: the non-en authority and guide fields
		// must not carry the English prose (the en canonical authority
		// sentence, its body fragment, or the en guide clause). The matrix
		// path, the flag names, the marker literal, and the URL are protocol
		// tokens and are exempt.
		if lang != langEnglish {
			for name, field := range map[string]string{
				"laneSpawnAuthority":  fm.laneSpawnAuthority,
				"autoModeGuideLeader": fm.autoModeGuideLeader,
				"autoModeGuideLane":   fm.autoModeGuideLane,
			} {
				for _, banned := range []string{"Standing spawn authority:", authorityBodyFragment, guideEnGuideFragment} {
					if strings.Contains(field, banned) {
						t.Errorf("%s %s mixes English prose into the locale block (%q):\n%s", lang, name, banned, field)
					}
				}
			}
		}
	}

	// The en authority keeps the canonical sentence verbatim (the lane
	// spawn-authority regression markers depend on it).
	if en := factoryMessagesFor(langEnglish).laneSpawnAuthority; !strings.Contains(en, "Standing spawn authority:") || !strings.Contains(en, authorityBodyFragment) {
		t.Errorf("en laneSpawnAuthority lost the canonical sentence:\n%s", en)
	}
}

// TestFactoryGuideAutoModeMarkerBlocksEveryLocale extends the en-only RED
// block-scope assertion (AC-005) to every locale: in BOTH notices the block
// carrying --auto-leader also carries the design-surface marker, and the lane
// notice carries its own lane-variant block with both flags.
func TestFactoryGuideAutoModeMarkerBlocksEveryLocale(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		leader := factoryLeaderNotice("guide"+lang, 2, lang)
		flagBlock := ""
		for _, block := range strings.Split(leader, "\n\n") {
			if strings.Contains(block, "--auto-leader") {
				flagBlock = block
				break
			}
		}
		if flagBlock == "" {
			t.Errorf("%s leader notice has no --auto-leader block", lang)
			continue
		}
		if !strings.Contains(flagBlock, designSurfaceMarker) {
			t.Errorf("%s leader --auto-leader block lost the design-surface marker %q", lang, designSurfaceMarker)
		}
		if !strings.Contains(flagBlock, "--auto-lane") {
			t.Errorf("%s leader --auto-leader block does not name the lane mode's flag", lang)
		}

		lane := factoryLaneNotice("lane-1", 2, lang)
		laneBlock := ""
		for _, block := range strings.Split(lane, "\n\n") {
			if strings.Contains(block, "--auto-lane") {
				laneBlock = block
				break
			}
		}
		if laneBlock == "" {
			t.Errorf("%s lane notice has no --auto-lane block", lang)
			continue
		}
		if !strings.Contains(laneBlock, designSurfaceMarker) {
			t.Errorf("%s lane --auto-lane block lost the design-surface marker %q", lang, designSurfaceMarker)
		}
		if !strings.Contains(laneBlock, "--auto-leader") {
			t.Errorf("%s lane --auto-lane block does not name the leader mode's flag", lang)
		}
	}
}

// TestFactoryGuideDocsPointerInBothNotices (AC-003): every locale's leader and
// lane notice each carry the docs URL at least once.
func TestFactoryGuideDocsPointerInBothNotices(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		if n := strings.Count(factoryLeaderNotice("docs"+lang, 2, lang), "https://adk.mo.ai.kr"); n < 1 {
			t.Errorf("%s leader notice carries the docs URL %d times, want >= 1", lang, n)
		}
		if n := strings.Count(factoryLaneNotice("lane-1", 2, lang), "https://adk.mo.ai.kr"); n < 1 {
			t.Errorf("%s lane notice carries the docs URL %d times, want >= 1", lang, n)
		}
	}
}

// TestFactoryGuideFallbackCompleteNotice (AC-006): an advertised locale the
// table does not carry (fr), the empty string, and a nil configuration all
// resolve to the complete English notice — never an empty notice and never a
// half-rendered one.
func TestFactoryGuideFallbackCompleteNotice(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvFactoryRunID, "guidefallback")
	t.Setenv(config.EnvFactoryLeadAddr, "/tmp/moai-socket-factory/guidefallback")

	enNotice := factoryBootstrapNotice("", "guide-session", langEnglish)

	for name, lang := range map[string]string{"fr": "fr", "empty": ""} {
		if got := factoryBootstrapNotice("", "guide-session", lang); got != enNotice {
			t.Errorf("%s conversation_language must render the complete English notice byte-identically", name)
		}
		fm, en := factoryMessagesFor(lang), factoryMessagesFor(langEnglish)
		for _, pair := range [][2]string{
			{fm.laneSpawnAuthority, en.laneSpawnAuthority},
			{fm.autoModeGuideLeader, en.autoModeGuideLeader},
			{fm.autoModeGuideLane, en.autoModeGuideLane},
			{fm.docsPointer, en.docsPointer},
		} {
			if pair[0] != pair[1] {
				t.Errorf("%s fallback resolved a non-English guide field", name)
			}
		}
	}

	// A nil configuration provider resolves en (operatorLang's fail-open) and
	// renders the same complete notice.
	if got := factoryBootstrapNoticeForSource("startup", "", "guide-session", operatorLang(nil)); got != enNotice {
		t.Errorf("a nil configuration must render the complete English notice byte-identically")
	}
}

// TestFactoryGuideNewFieldsLayoutInvariants (AC-008): the new fields carry no
// leading or trailing newline and no internal blank-line separator (the
// builders join lines within a block and blank-separate the blocks), and the
// assembled notices have no empty or newline-padded blocks.
func TestFactoryGuideNewFieldsLayoutInvariants(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		fm := factoryMessagesFor(lang)
		for name, field := range map[string]string{
			"laneSpawnAuthority":  fm.laneSpawnAuthority,
			"autoModeGuideLeader": fm.autoModeGuideLeader,
			"autoModeGuideLane":   fm.autoModeGuideLane,
			"docsPointer":         fm.docsPointer,
		} {
			if strings.HasPrefix(field, "\n") || strings.HasSuffix(field, "\n") {
				t.Errorf("%s %s has a leading/trailing newline: %q", lang, name, field)
			}
			if strings.Contains(field, "\n\n") {
				t.Errorf("%s %s carries an internal blank-line separator (the builder owns block separation)", lang, name)
			}
		}

		// The assembled notices: no block is empty or newline-padded. The
		// leader notice carries one trailing newline as its terminator (the
		// builder's established shape: strings.Join(blocks, "\n\n") + "\n"),
		// so it is trimmed before the block split; the lane notice carries
		// none.
		leader := strings.TrimSuffix(factoryLeaderNotice("layout"+lang, 2, lang), "\n")
		for i, block := range strings.Split(leader, "\n\n") {
			if strings.HasPrefix(block, "\n") || strings.HasSuffix(block, "\n") || strings.TrimSpace(block) == "" {
				t.Errorf("%s leader notice block %d is empty or newline-padded: %q", lang, i, block)
			}
		}
		lane := factoryLaneNotice("lane-1", 2, lang)
		for i, block := range strings.Split(lane, "\n\n") {
			if strings.HasPrefix(block, "\n") || strings.HasSuffix(block, "\n") || strings.TrimSpace(block) == "" {
				t.Errorf("%s lane notice block %d is empty or newline-padded: %q", lang, i, block)
			}
		}
	}
}
