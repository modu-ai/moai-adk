package wizard

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// t1972 F6 (carried from t1966): selectHeight counted one row per option, so an
// option whose "label - description" text wraps (long text, or East Asian wide
// characters that take two columns each) was under-counted and the select
// viewport hid options.

func rowsLowerBound(width int, text string) int {
	w := runewidth.StringWidth(text)
	if w <= width {
		return 1
	}
	return (w + width - 1) / width
}

func TestSelectHeightCountsWrappedOptionRows(t *testing.T) {
	const optionWidth = selectFieldUsableWidth - 2 // the cursor column of an option row
	oneLine := func(s string) int {
		return lipgloss.Height(lipgloss.NewStyle().Width(selectFieldUsableWidth).Render(s))
	}
	cases := []struct {
		name string
		q    Question
		// the rows the option set needs at least, by display width
		minOptionRows int
	}{
		{
			name:          "control: short options take one row each",
			q:             Question{Description: "pick one", Options: []Option{{Label: "A", Value: "a", Desc: "short"}, {Label: "B", Value: "b", Desc: "short"}}},
			minOptionRows: 2,
		},
		{
			name:          "a long ASCII description wraps",
			q:             Question{Description: "pick one", Options: []Option{{Label: "A", Value: "a", Desc: strings.Repeat("word ", 30)}, {Label: "B", Value: "b", Desc: "short"}}},
			minOptionRows: 3 + 1,
		},
		{
			// 40 characters but 80 columns: counting runes or bytes says one row
			name:          "a CJK description of 40 characters is 80 columns and wraps",
			q:             Question{Description: "pick one", Options: []Option{{Label: "A", Value: "a", Desc: strings.Repeat("한", 40)}, {Label: "B", Value: "b", Desc: "short"}}},
			minOptionRows: 2 + 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := tc.q
			want := 1 + oneLine(q.Description) + tc.minOptionRows
			if tc.minOptionRows == 2 && len(q.Options) == 2 {
				// control: exactly the old formula
				if got := selectHeight(&q); got != want {
					t.Errorf("selectHeight = %d, want %d", got, want)
				}
				return
			}
			if got := selectHeight(&q); got < want {
				t.Errorf("selectHeight = %d, want at least %d (wrapped option rows are not counted)", got, want)
			}
		})
	}

	// every real select question, in every locale: the height must cover the rows
	// its aligned option texts need at the option width
	checked := 0
	root := t.TempDir()
	for _, set := range [][]Question{ReconfigureQuestions(root), InitQuestions(root)} {
		for i := range set {
			q := &set[i]
			if q.Type != QuestionTypeSelect {
				continue
			}
			for loc := range translations {
				lq := GetLocalizedQuestion(q, loc)
				rows := 0
				for _, o := range alignOptionLabels(lq.Options) {
					rows += rowsLowerBound(optionWidth, o.Key)
				}
				need := 1 + oneLine(lq.Description) + rows
				checked++
				if got := selectHeight(q); got < need {
					t.Errorf("question %q locale %q: selectHeight = %d, need at least %d", q.ID, loc, got, need)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no select question was checked; the scan would pass vacuously")
	}
}
