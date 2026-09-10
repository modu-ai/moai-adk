package spec

import "testing"

// Card t565 — which heading anchors the acceptance criteria section, and where
// that section ends.
//
// Before the fix the anchor was the first heading starting with "##" whose
// lower-cased text contained "acceptance", so a heading that only pointed at the
// sibling file acceptance.md, or an Out of Scope heading, won over the real
// section, while a Korean or "Success Criteria" heading never anchored at all.
// The section then ended at the next line starting with "##", so the section's
// own ### subheadings cut it short while a # heading did not end it.
//
// Every fixture places the decoy AC-HDG-09 where only the rule under test keeps
// it out, and the real AC-HDG-01 where only the rule under test lets it in.

const t565RealAC = "- AC-HDG-01: Given a SPEC, When it is parsed, Then this criterion is read\n"

const t565DecoyAC = "- AC-HDG-09: Given a decoy, When it is parsed, Then it is not read\n"

func t565Expect(t *testing.T, md string, want, notWant []string) {
	t.Helper()
	ids := t528RootIDs(t, md)
	for _, id := range want {
		if !t528Has(ids, id) {
			t.Errorf("%s was NOT collected; got ids %v", id, ids)
		}
	}
	for _, id := range notWant {
		if t528Has(ids, id) {
			t.Errorf("%s WAS collected; got ids %v", id, ids)
		}
	}
}

func TestT565AnchorSkipsFileMentionOnlyHeading(t *testing.T) {
	md := "# Fixture\n\n" +
		"## 3. Plan — the details live in acceptance.md\n\n" + t565DecoyAC + "\n" +
		"## 6. Acceptance Criteria\n\n" + t565RealAC
	t565Expect(t, md, []string{"AC-HDG-01"}, []string{"AC-HDG-09"})
}

func TestT565AnchorSkipsOutOfScopeHeading(t *testing.T) {
	for name, heading := range map[string]string{
		// Names the section, not the file: only the negative exclusion keeps it out.
		"english": "### Out of Scope — rewriting acceptance criteria",
		// Corpus shape (SPEC-AC-COUNT-DISCRIMINATOR-001). Once Korean is vocabulary,
		// only the negative exclusion keeps it out.
		"korean": "### Out of Scope — 수락 기준 저작 방식 일반",
	} {
		t.Run(name, func(t *testing.T) {
			md := "# Fixture\n\n## 2. Scope\n\n" + heading + "\n\n" + t565DecoyAC + "\n" +
				"## 6. Acceptance Criteria\n\n" + t565RealAC
			t565Expect(t, md, []string{"AC-HDG-01"}, []string{"AC-HDG-09"})
		})
	}
}

func TestT565AnchorVocabulary(t *testing.T) {
	for _, heading := range []string{
		"## 6. Acceptance Criteria",
		// Names the file AND the section: the file-name exclusion must not reach it.
		"## 3. 수락 기준 — acceptance.md (Tier M)",
		"## 수락 기준",
		"## 인수 기준",
		"## 검수 기준",
		"## §F AC Matrix",
		"## §E Success Criteria",
		"## §E. Success criteria",
		"## §D 수용 기준",
		"## §H. 성공 기준 (요약)",
		// Corpus shape (SPEC-SYNC-AUDIT-FALSIFICATION-001): with the file name
		// removed, only "ac summary" names the section.
		"## §H AC summary (full GWT in acceptance.md)",
	} {
		t.Run(heading, func(t *testing.T) {
			md := "# Fixture\n\n## 1. Overview\n\nProse only.\n\n" + heading + "\n\n" + t565RealAC
			t565Expect(t, md, []string{"AC-HDG-01"}, nil)
		})
	}
}

// Corpus shape (SPEC-LEARN-CHANNEL-SCOPE-001): an empty summary section that
// names the criteria comes before the real one. The anchor is the first
// section-naming heading whose section the parser reads lines from.
func TestT565AnchorSkipsEmptySection(t *testing.T) {
	md := "# Fixture\n\n" +
		"## §F. Success Criteria\n\nProse only, no declarations.\n\n" +
		"## §G. Out of Scope\n\nProse.\n\n" +
		"## §I. Acceptance Criteria\n\n" + t565RealAC
	t565Expect(t, md, []string{"AC-HDG-01"}, nil)
}

// When every section-naming heading is empty, the document still has an
// acceptance criteria section: no "section not found" error.
func TestT565AnchorAllEmptySectionsStillAnchor(t *testing.T) {
	md := "# Fixture\n\n## §F. Success Criteria\n\nProse only.\n\n## §I. Acceptance Criteria\n\nProse only.\n"
	criteria, errs := ParseAcceptanceCriteria(md, false)
	if len(criteria) != 0 {
		t.Errorf("expected no criteria from empty sections; got %+v", criteria)
	}
	if len(errs) != 0 {
		t.Errorf("empty sections must still anchor (no section-not-found error); got %v", errs)
	}
}

func TestT565AnchorVocabularyIsNotWide(t *testing.T) {
	md := "# Fixture\n\n## 7. Review Criteria\n\n" + t565DecoyAC
	criteria, errs := ParseAcceptanceCriteria(md, false)
	if len(criteria) != 0 {
		t.Errorf("a heading outside the vocabulary anchored the section; got %+v", criteria)
	}
	if len(errs) == 0 {
		t.Errorf("expected the section-not-found error for a heading outside the vocabulary; got none")
	}
}

func TestT565SectionReadsItsOwnSubheadings(t *testing.T) {
	md := "# Fixture\n\n" +
		"## 6. Acceptance Criteria\n\n" +
		"### 6.1 Parser\n\n" + t565RealAC + "\n" +
		"#### 6.1.1 Edge\n\n" +
		"- AC-HDG-02: Given a deeper subheading, When it is parsed, Then it is read\n\n" +
		"## 7. Notes\n\n" + t565DecoyAC
	t565Expect(t, md, []string{"AC-HDG-01", "AC-HDG-02"}, []string{"AC-HDG-09"})
}

func TestT565SectionEndsAtSameOrHigherLevel(t *testing.T) {
	t.Run("h3-anchor", func(t *testing.T) {
		md := "# Fixture\n\n## 6. Verification\n\n" +
			"### 6.1 Acceptance Criteria\n\n" + t565RealAC + "\n" +
			"#### 6.1.1 Edge\n\n" +
			"- AC-HDG-02: Given a deeper subheading, When it is parsed, Then it is read\n\n" +
			"### 6.2 Benchmarks\n\n" + t565DecoyAC
		t565Expect(t, md, []string{"AC-HDG-01", "AC-HDG-02"}, []string{"AC-HDG-09"})
	})
	t.Run("h1-ends-h2", func(t *testing.T) {
		md := "# Fixture\n\n## 6. Acceptance Criteria\n\n" + t565RealAC + "\n" +
			"# Appendix\n\n" + t565DecoyAC
		t565Expect(t, md, []string{"AC-HDG-01"}, []string{"AC-HDG-09"})
	})
}
