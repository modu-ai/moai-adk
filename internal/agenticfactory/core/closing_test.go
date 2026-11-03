package core

import (
	"fmt"
	"testing"
)

func refsOf(texts ...string) string {
	var out []string
	for _, r := range ClosingTargets(texts...) {
		out = append(out, fmt.Sprintf("%s#%d", r.Repo, r.Issue))
	}
	return fmt.Sprint(out)
}

// The nine closing keywords GitHub documents, in the documented forms.
func TestEveryDocumentedKeywordIsRecognised(t *testing.T) {
	for _, kw := range []string{"close", "closes", "closed", "fix", "fixes", "fixed", "resolve", "resolves", "resolved"} {
		for _, form := range []string{"%s #10", "%s: #10", "%s:#10", "%s  #10"} {
			text := fmt.Sprintf(form, kw)
			if got := refsOf(text); got != "[#10]" {
				t.Errorf("%q -> %s", text, got)
			}
		}
	}
}

func TestCaseDoesNotMatter(t *testing.T) {
	for _, text := range []string{"CLOSES #5", "Closes #5", "cLoSeS #5", "FIXES: #5", "Resolved #5"} {
		if got := refsOf(text); got != "[#5]" {
			t.Errorf("%q -> %s", text, got)
		}
	}
}

func TestOtherRepositoriesAndUrlsAreCaught(t *testing.T) {
	cases := map[string]string{
		"Fixes octo-org/octo-repo#100":                          "[octo-org/octo-repo#100]",
		"Resolves #10, resolves #123, resolves octo-org/r#100":  "[#10 #123 octo-org/r#100]",
		"closes https://github.com/octo-org/octo-repo/issues/7": "[octo-org/octo-repo#7]",
	}
	for text, want := range cases {
		if got := refsOf(text); got != want {
			t.Errorf("%q -> %s, want %s", text, got, want)
		}
	}
}

func TestSeveralTextsAreScannedTogether(t *testing.T) {
	got := refsOf("PR body\n\nFixes #1", "commit one: closes #2", "no keyword here #3")
	if got != "[#1 #2]" {
		t.Errorf("got %s", got)
	}
}

// Plain references and look-alike words do not close anything.
func TestNonClosingTextIsLeftAlone(t *testing.T) {
	for _, text := range []string{
		"Refs #12", "Related to #12", "See #12", "Part of #12", "see issue 12",
		"prefix #12", "suffixes #12", "disclose #12", "enclosed #12", "unfixed #12",
		"fixed in #12", "closes the gap", "resolves to #12 later", "fixes", "#12 closes",
		"", "closes #", "closes #abc",
	} {
		if got := refsOf(text); got != "[]" {
			t.Errorf("%q must not be a closing reference, got %s", text, got)
		}
	}
}

func TestEachReferenceIsReportedOnceInOrder(t *testing.T) {
	got := ClosingTargets("fixes #4", "Fixes #4", "closes #2")
	if len(got) != 2 || got[0].Issue != 4 || got[1].Issue != 2 {
		t.Errorf("got %+v", got)
	}
	if got[0].Keyword == "" || got[0].Text == "" {
		t.Errorf("the matched keyword and text must be kept for the message: %+v", got[0])
	}
}
