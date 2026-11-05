package core

import (
	"regexp"
	"strconv"
	"strings"
)

// ClosingRef is one place where text would close an Issue automatically.
type ClosingRef struct {
	Repo    string // "owner/name" as written; empty means the PR's own repository
	Issue   int
	Keyword string // the matched keyword, as written
	Text    string // the matched phrase, for the message
}

// closingRe finds a closing keyword followed by an Issue reference.
//
// GitHub documents nine keywords (close, closes, closed, fix, fixes, fixed,
// resolve, resolves, resolved) and the forms `KEYWORD #N` and
// `KEYWORD OWNER/REPO#N`, with an optional colon and upper case accepted
// (https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue,
// read 2026-10-29 as a summary, not byte for byte). Mixed case and the URL form
// are not described there; this scan accepts both, because refusing a harmless
// reference costs a rewording and missing a real one closes someone else's Issue.
var closingRe = regexp.MustCompile(`(?i)\b(close[sd]?|fix(?:es|ed)?|resolve[sd]?)\b[ \t]*:?[ \t]*(?:https?://github\.com/([\w.-]+/[\w.-]+)/issues/(\d+)|([\w.-]+/[\w.-]+)?#(\d+))`)

// ClosingTargets returns the Issues the given texts (a PR description, commit
// messages) would close, each once, in order of first appearance.
func ClosingTargets(texts ...string) []ClosingRef {
	var out []ClosingRef
	seen := map[string]bool{}
	for _, text := range texts {
		for _, m := range closingRe.FindAllStringSubmatch(text, -1) {
			repo, num := m[4], m[5]
			if m[3] != "" {
				repo, num = m[2], m[3]
			}
			n, err := strconv.Atoi(num)
			if err != nil || n <= 0 {
				continue
			}
			key := strings.ToLower(repo) + "#" + num
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, ClosingRef{Repo: repo, Issue: n, Keyword: m[1], Text: strings.TrimSpace(m[0])})
		}
	}
	return out
}
