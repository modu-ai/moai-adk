// Output-style prefix-diet guards (SPEC-PREFIX-DIET-001, card t1450).
//
//   - TestOutputStylesCharBudget      whole-file UTF-16 budget per deployed style (REQ-PFD-001/002)
//   - TestOutputStyleBindingLedger    complete binding ledger, no rewrite, no lost binding token
//     (REQ-PFD-003/004)
//   - TestOutputStyleHandoffUnitsFrozen   the handoff sections stay byte-identical (REQ-PFD-005)
//   - TestOutputStyleLocalizationTableParity   localization tables stay cell-identical (REQ-PFD-006)
//
// The fixtures live in testdata/output_style_*.json, recorded at anchor 5d5ff1aae. They are test
// inputs only and never ship in the binary.
package template

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

// Whole-file UTF-16 budgets of the deployed output styles. A constant may only go down: it never
// exceeds the anchor size recorded in the ledger (REQ-PFD-002).
const (
	dietBudgetMoai      = 61362
	dietBudgetMoaiEasy  = 23586
	dietBudgetMoaiLearn = 28517
)

var dietBudgets = []struct {
	file   string
	name   string
	budget int
}{
	{"moai.md", "moai", dietBudgetMoai},
	{"moai-easy.md", "moai-easy", dietBudgetMoaiEasy},
	{"moai-learn.md", "moai-learn", dietBudgetMoaiLearn},
}

func TestOutputStylesCharBudget(t *testing.T) {
	deployed := dietReadDeployed(t)
	var led dietLedger
	dietLoadJSON(t, "output_style_ledger.json", &led)
	for _, b := range dietBudgets {
		t.Run(b.name, func(t *testing.T) {
			size := dietUTF16Len(deployed[b.file])
			t.Logf("output-style=%s %d", b.name, size)
			if size > b.budget {
				t.Errorf("output style %s is %d UTF-16 units, over its budget %d", b.file, size, b.budget)
			}
			meta := led.Files[b.file]
			if meta == nil {
				t.Fatalf("ledger has no entry for %s", b.file)
			}
			if b.budget > meta.AnchorFileUTF16 {
				t.Errorf("budget constant %d for %s exceeds its anchor size %d", b.budget, b.file, meta.AnchorFileUTF16)
			}
		})
	}
}

// dietValidateLedger returns one violation line per defect. Every line starts with a stable
// keyword so the mutation subtests can assert which check fired.
func dietValidateLedger(led *dietLedger, frozen *dietFrozenDoc, deployed map[string]string) []string {
	var v []string
	files := make([]string, 0, len(led.Files))
	for f := range led.Files {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		meta := led.Files[f]
		rows := map[int]*dietRow{}
		for i := range led.Rows {
			r := &led.Rows[i]
			if r.File != f {
				continue
			}
			if _, dup := rows[r.Index]; dup {
				v = append(v, fmt.Sprintf("DUPLICATE_ROW %s unit %d", f, r.Index))
			}
			rows[r.Index] = r
		}
		frozenAt := map[int]dietFrozenSection{}
		for _, s := range frozen.Sections {
			if s.File == f {
				frozenAt[s.FirstIndex] = s
			}
		}
		// Reconstruct the anchor body and the expected deployed body from the rows.
		var before, after strings.Builder
		for i := 0; i < meta.AnchorUnits; {
			if s, ok := frozenAt[i]; ok {
				before.WriteString(s.Text)
				after.WriteString(s.Text)
				i = s.LastIndex + 1
				continue
			}
			r, ok := rows[i]
			if !ok {
				v = append(v, fmt.Sprintf("MISSING_ROW %s anchor unit %d has no ledger row", f, i))
				i++
				continue
			}
			before.WriteString(r.BeforeText)
			after.WriteString(r.AfterText)
			i++
		}
		if got := dietSHA256(before.String()); got != meta.AnchorBodySHA256 {
			v = append(v, fmt.Sprintf("ANCHOR_SHA %s anchor body rebuilt from the ledger hashes to %s, want %s", f, got, meta.AnchorBodySHA256))
		}
		anchorUnits := dietExtractUnits(before.String())
		if len(anchorUnits) != meta.AnchorUnits {
			v = append(v, fmt.Sprintf("UNIT_COUNT %s extractor finds %d units in the rebuilt anchor, want %d", f, len(anchorUnits), meta.AnchorUnits))
		}
		for i, r := range rows {
			if i < len(anchorUnits) && r.BeforeText != anchorUnits[i] {
				v = append(v, fmt.Sprintf("UNIT_MISMATCH %s row %s before_text differs from extracted unit %d", f, r.ID, i))
			}
			c := dietTokenCounts(r.BeforeText)
			switch r.Kind {
			case "binding", "normative", "rationale", "example":
			default:
				v = append(v, fmt.Sprintf("KIND_UNKNOWN %s row %s kind %q", f, r.ID, r.Kind))
			}
			if dietHasToken(c) != (r.Kind == "binding") {
				v = append(v, fmt.Sprintf("KIND_MISMATCH %s row %s kind %q but binding tokens %v", f, r.ID, r.Kind, c))
			}
			switch r.Treatment {
			case "verbatim":
				if r.AfterText != r.BeforeText {
					v = append(v, fmt.Sprintf("VERBATIM_DIFF %s row %s after_text differs from before_text", f, r.ID))
				}
				if dietTokenCounts(r.AfterText) != c {
					v = append(v, fmt.Sprintf("ROW_TOKEN %s row %s token counts differ", f, r.ID))
				}
			case "dropped":
				if r.Kind != "rationale" && r.Kind != "example" {
					v = append(v, fmt.Sprintf("DROPPED_KIND %s row %s kind %q is a binding or normative row and must not be dropped", f, r.ID, r.Kind))
				}
				if strings.TrimSpace(r.Note) == "" {
					v = append(v, fmt.Sprintf("DROPPED_NOTE %s row %s has no reason in note", f, r.ID))
				}
				if r.SurvivorFile == "" || r.SurvivorAnchor == "" {
					v = append(v, fmt.Sprintf("DROPPED_SURVIVOR %s row %s has no survivor pointer (survivor_file + survivor_anchor)", f, r.ID))
				}
				if dietHasToken(c) {
					v = append(v, fmt.Sprintf("DROPPED_TOKENS %s row %s carries binding tokens %v", f, r.ID, c))
				}
				if r.AfterText != "" {
					v = append(v, fmt.Sprintf("DROPPED_AFTER %s row %s keeps after_text", f, r.ID))
				}
				if raw, err := os.ReadFile(r.SurvivorFile); r.SurvivorFile == "" || r.SurvivorAnchor == "" || err != nil || !strings.Contains(string(raw), r.SurvivorAnchor) {
					v = append(v, fmt.Sprintf("SURVIVOR_UNRESOLVED %s row %s survivor pointer %q + %q does not resolve", f, r.ID, r.SurvivorFile, r.SurvivorAnchor))
				}
			default:
				v = append(v, fmt.Sprintf("TREATMENT %s row %s treatment %q is neither verbatim nor dropped", f, r.ID, r.Treatment))
			}
		}
		if got := dietTokenCounts(after.String()); got != meta.AnchorTokens {
			v = append(v, fmt.Sprintf("TOKEN_TOTAL %s binding token total after the change is %v, anchor total is %v", f, got, meta.AnchorTokens))
		}
		head, body, err := dietSplitFrontmatter(deployed[f])
		if err != nil {
			v = append(v, fmt.Sprintf("DEPLOYED_PARSE %s %v", f, err))
			continue
		}
		if dietSHA256(head) != meta.AnchorFrontmatterSHA256 {
			v = append(v, fmt.Sprintf("FRONTMATTER_DIFF %s deployed frontmatter differs from the anchor", f))
		}
		if body != after.String() {
			v = append(v, fmt.Sprintf("DEPLOYED_DIFF %s deployed body differs from the body the ledger accounts for (undocumented edit)", f))
		}
	}
	return v
}

func TestOutputStyleBindingLedger(t *testing.T) {
	var led dietLedger
	var frozen dietFrozenDoc
	dietLoadJSON(t, "output_style_ledger.json", &led)
	dietLoadJSON(t, "output_style_frozen.json", &frozen)
	deployed := dietReadDeployed(t)

	if got := dietValidateLedger(&led, &frozen, deployed); len(got) != 0 {
		for _, line := range got {
			t.Error(line)
		}
	}

	hasRewrite := false
	for _, r := range led.Rows {
		if r.Treatment == "rewrite" {
			hasRewrite = true
		}
	}
	if hasRewrite {
		t.Error("ledger carries a rewrite row (leader decision D3: verbatim or dropped only)")
	}

	// Each mutation must be caught by the check it targets. A mutant that validates clean means
	// the guard cannot discriminate.
	pick := func(t *testing.T, l *dietLedger, pred func(r *dietRow) bool) *dietRow {
		t.Helper()
		for i := range l.Rows {
			if pred(&l.Rows[i]) {
				return &l.Rows[i]
			}
		}
		t.Fatalf("no ledger row satisfies the mutation precondition")
		return nil
	}
	expect := func(t *testing.T, l *dietLedger, keyword string) {
		t.Helper()
		got := dietValidateLedger(l, &frozen, deployed)
		if !slices.ContainsFunc(got, func(s string) bool { return strings.HasPrefix(s, keyword+" ") }) {
			t.Fatalf("mutant was not caught by %s; violations: %v", keyword, got)
		}
	}
	binding := func(r *dietRow) bool { return r.Kind == "binding" && r.Treatment == "verbatim" }

	t.Run("missing_anchor_unit_row", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, func(r *dietRow) bool { return r.Kind == "normative" })
		id := r.ID
		l.Rows = slices.DeleteFunc(l.Rows, func(x dietRow) bool { return x.ID == id })
		expect(t, l, "MISSING_ROW")
	})
	t.Run("binding_unit_relabeled", func(t *testing.T) {
		l := dietClone(t, &led)
		pick(t, l, binding).Kind = "normative"
		expect(t, l, "KIND_MISMATCH")
	})
	t.Run("binding_row_dropped", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, binding)
		r.Treatment, r.AfterText, r.Note, r.Survivor = "dropped", "", "mutant", "mutant"
		expect(t, l, "DROPPED_KIND")
	})
	t.Run("normative_row_dropped", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, func(r *dietRow) bool { return r.Kind == "normative" && r.Treatment == "verbatim" })
		r.Treatment, r.AfterText, r.Note, r.Survivor = "dropped", "", "mutant", "mutant"
		expect(t, l, "DROPPED_KIND")
	})
	t.Run("dropped_without_note", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, func(r *dietRow) bool { return r.Kind == "rationale" || r.Kind == "example" })
		r.Treatment, r.AfterText, r.Note, r.Survivor = "dropped", "", "", "somewhere"
		expect(t, l, "DROPPED_NOTE")
	})
	t.Run("dropped_survivor_anchor_missing", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, func(r *dietRow) bool { return r.Treatment == "dropped" && r.SurvivorFile != "" })
		r.SurvivorAnchor = "no-such-anchor-in-that-file-7f3a"
		expect(t, l, "SURVIVOR_UNRESOLVED")
	})
	t.Run("dropped_survivor_file_missing", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, func(r *dietRow) bool { return r.Treatment == "dropped" && r.SurvivorFile != "" })
		r.SurvivorFile = "templates/no/such/file.md"
		expect(t, l, "SURVIVOR_UNRESOLVED")
	})
	t.Run("dropped_without_survivor", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, func(r *dietRow) bool { return r.Kind == "rationale" || r.Kind == "example" })
		r.Treatment, r.AfterText, r.Note, r.SurvivorFile, r.SurvivorAnchor = "dropped", "", "a reason", "", ""
		expect(t, l, "DROPPED_SURVIVOR")
	})
	t.Run("verbatim_text_changed", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, func(r *dietRow) bool { return r.Kind == "normative" && r.Treatment == "verbatim" })
		r.AfterText = r.BeforeText + "x"
		expect(t, l, "VERBATIM_DIFF")
	})
	t.Run("dropped_with_tokens", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, binding)
		r.Kind, r.Treatment, r.AfterText, r.Note, r.Survivor = "rationale", "dropped", "", "mutant", "mutant"
		expect(t, l, "DROPPED_TOKENS")
	})
	t.Run("file_total_token_lost", func(t *testing.T) {
		l := dietClone(t, &led)
		r := pick(t, l, binding)
		r.AfterText = strings.Replace(r.AfterText, "[HARD]", "[hard]", 1)
		expect(t, l, "TOKEN_TOTAL")
	})
	t.Run("rewrite_row_rejected", func(t *testing.T) {
		l := dietClone(t, &led)
		pick(t, l, func(r *dietRow) bool { return r.Kind == "normative" }).Treatment = "rewrite"
		expect(t, l, "TREATMENT")
	})
}

// dietCheckFrozen compares the hash of every frozen section of the deployed files with the
// fixture. A change names the section and both hashes.
func dietCheckFrozen(deployed map[string]string, doc *dietFrozenDoc) []string {
	var v []string
	for _, s := range doc.Sections {
		_, body, err := dietSplitFrontmatter(deployed[s.File])
		if err != nil {
			v = append(v, fmt.Sprintf("FROZEN_PARSE %s: %v", s.File, err))
			continue
		}
		units := dietExtractUnits(body)
		first, last, ok := dietSectionRange(units, s.Title, s.Level)
		if !ok {
			v = append(v, fmt.Sprintf("FROZEN_MISSING %s section %q not found", s.File, s.Title))
			continue
		}
		got := dietSHA256(strings.Join(units[first:last+1], ""))
		if got != s.SHA256 {
			v = append(v, fmt.Sprintf("FROZEN_CHANGED %s section %q hash %s, anchor hash %s", s.File, s.Title, got, s.SHA256))
		}
	}
	return v
}

func TestOutputStyleHandoffUnitsFrozen(t *testing.T) {
	var doc dietFrozenDoc
	dietLoadJSON(t, "output_style_frozen.json", &doc)
	deployed := dietReadDeployed(t)
	if len(doc.Sections) != 3 {
		t.Fatalf("frozen fixture carries %d sections, want 3 (moai.md two, moai-easy.md one)", len(doc.Sections))
	}
	for _, line := range dietCheckFrozen(deployed, &doc) {
		t.Error(line)
	}

	t.Run("one_char_mutation_in_frozen_section_is_caught", func(t *testing.T) {
		for _, s := range doc.Sections {
			mutated := map[string]string{}
			for k, val := range deployed {
				mutated[k] = val
			}
			idx := strings.Index(mutated[s.File], s.Text)
			if idx < 0 {
				t.Fatalf("frozen text of %q not found verbatim in %s", s.Title, s.File)
			}
			// Mutate one character below the heading line so the section is still located.
			nl := strings.Index(s.Text, "\n") + 1
			bad := s.Text[:nl] + strings.Replace(s.Text[nl:], "e", "E", 1)
			mutated[s.File] = mutated[s.File][:idx] + bad + mutated[s.File][idx+len(s.Text):]
			got := dietCheckFrozen(mutated, &doc)
			found := slices.ContainsFunc(got, func(l string) bool {
				return strings.HasPrefix(l, "FROZEN_CHANGED "+s.File) && strings.Contains(l, s.Title) && strings.Contains(l, s.SHA256)
			})
			if !found {
				t.Errorf("one-character mutation of %q in %s was not caught; got %v", s.Title, s.File, got)
			}
		}
	})
}

func dietCells(row string) []string {
	parts := strings.Split(row, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

// dietCheckLocalization requires every anchor table to appear in its file as a contiguous block of
// identical rows. When no block matches, the closest candidate is reported cell by cell.
func dietCheckLocalization(deployed map[string]string, doc *dietLocalDoc) []string {
	var v []string
	for _, tb := range doc.Tables {
		_, body, err := dietSplitFrontmatter(deployed[tb.File])
		if err != nil {
			v = append(v, fmt.Sprintf("LOCALIZATION_PARSE %s: %v", tb.File, err))
			continue
		}
		lines := strings.Split(body, "\n")
		bestStart, bestDiff := -1, 1<<30
		matched := false
		for k := 0; k+len(tb.Rows) <= len(lines) && !matched; k++ {
			diff := 0
			for r := range tb.Rows {
				if lines[k+r] != tb.Rows[r] {
					diff++
				}
			}
			if diff == 0 {
				matched = true
			} else if lines[k] == tb.Rows[0] && diff < bestDiff {
				bestStart, bestDiff = k, diff
			}
		}
		if matched {
			continue
		}
		if bestStart < 0 {
			v = append(v, fmt.Sprintf("LOCALIZATION_TABLE_MISSING %s table %s header row not found", tb.File, tb.ID))
			continue
		}
		for r := range tb.Rows {
			if lines[bestStart+r] == tb.Rows[r] {
				continue
			}
			want, got := dietCells(tb.Rows[r]), dietCells(lines[bestStart+r])
			for c := range want {
				if c >= len(got) || want[c] != got[c] {
					have := ""
					if c < len(got) {
						have = got[c]
					}
					v = append(v, fmt.Sprintf("LOCALIZATION_CELL %s table %s row %d cell %d is %q, anchor %q", tb.File, tb.ID, r, c, have, want[c]))
				}
			}
		}
	}
	return v
}

func TestOutputStyleLocalizationTableParity(t *testing.T) {
	var doc dietLocalDoc
	dietLoadJSON(t, "output_style_localization.json", &doc)
	deployed := dietReadDeployed(t)
	if len(doc.Tables) == 0 {
		t.Fatal("localization fixture carries no tables")
	}
	for _, line := range dietCheckLocalization(deployed, &doc) {
		t.Error(line)
	}

	t.Run("one_cell_mutation_is_caught", func(t *testing.T) {
		tb := doc.Tables[0]
		row := tb.Rows[len(tb.Rows)-1]
		cells := dietCells(row)
		cell := cells[len(cells)-2] // last non-empty cell
		mutated := map[string]string{}
		for k, val := range deployed {
			mutated[k] = val
		}
		mutated[tb.File] = strings.Replace(mutated[tb.File], row, strings.Replace(row, cell, cell+"x", 1), 1)
		got := dietCheckLocalization(mutated, &doc)
		if !slices.ContainsFunc(got, func(l string) bool { return strings.HasPrefix(l, "LOCALIZATION_CELL "+tb.File) }) {
			t.Errorf("one-cell mutation was not caught; got %v", got)
		}
	})
	t.Run("removed_row_is_caught", func(t *testing.T) {
		tb := doc.Tables[0]
		row := tb.Rows[len(tb.Rows)-1]
		mutated := map[string]string{}
		for k, val := range deployed {
			mutated[k] = val
		}
		mutated[tb.File] = strings.Replace(mutated[tb.File], row+"\n", "", 1)
		if got := dietCheckLocalization(mutated, &doc); len(got) == 0 {
			t.Errorf("removing a localization row was not caught")
		}
	})
}
