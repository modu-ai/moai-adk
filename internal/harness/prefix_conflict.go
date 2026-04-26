// Package harness — my-harness-* prefix conflict detector.
package harness

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Conflict describes one semantic name collision between a user-area
// `my-harness-*` skill and a moai-managed `moai-*` skill.
type Conflict struct {
	// MyHarnessSkill is the user-area skill directory name (e.g. "my-harness-foundation-core").
	MyHarnessSkill string
	// MoaiSkill is the conflicting moai-managed skill directory name.
	MoaiSkill string
	// Reason explains the type of conflict (suffix match or close edit distance).
	Reason string
}

// DetectPrefixConflicts scans skillsDir (typically `.claude/skills/`) for
// `my-harness-*` directories and flags any whose suffix collides with an
// existing `moai-*` skill (REQ-PH-009 indirect — diagnostic warning).
//
// Conflict heuristic:
//  1. Strip "my-harness-" prefix → suffix S
//  2. If a "moai-S" directory exists in the same skillsDir → conflict (suffix match)
//  3. Else if any "moai-*" name has Levenshtein distance ≤ 2 from "moai-S" → conflict (close name)
//
// The function never errors when skillsDir is missing — it returns an empty
// slice. Other errors (permission etc.) are returned to the caller.
func DetectPrefixConflicts(skillsDir string) ([]Conflict, error) {
	if skillsDir == "" {
		return nil, errors.New("DetectPrefixConflicts: empty skills dir")
	}
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var moaiNames []string
	var myHarnessNames []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		switch {
		case strings.HasPrefix(name, "my-harness-"):
			myHarnessNames = append(myHarnessNames, name)
		case strings.HasPrefix(name, "moai-"):
			moaiNames = append(moaiNames, name)
		}
	}
	var conflicts []Conflict
	for _, mh := range myHarnessNames {
		suffix := strings.TrimPrefix(mh, "my-harness-")
		expected := "moai-" + suffix
		// Pass 1: exact suffix match
		exactFound := false
		for _, m := range moaiNames {
			if m == expected {
				conflicts = append(conflicts, Conflict{
					MyHarnessSkill: mh,
					MoaiSkill:      m,
					Reason:         "exact suffix match (semantic name overlap)",
				})
				exactFound = true
				break
			}
		}
		if exactFound {
			continue
		}
		// Pass 2: Levenshtein distance ≤ 2 against "moai-suffix"
		for _, m := range moaiNames {
			d := levenshtein(m, expected)
			if d > 0 && d <= 2 {
				conflicts = append(conflicts, Conflict{
					MyHarnessSkill: mh,
					MoaiSkill:      m,
					Reason:         "close name (edit distance " + itoa(d) + ")",
				})
				break
			}
		}
	}
	// Stable: discovery order from os.ReadDir is sorted on most platforms,
	// but explicit filepath join not required since we only return names.
	_ = filepath.Join // keep import used when building paths in callers
	return conflicts, nil
}

// levenshtein computes the edit distance between two strings using a simple
// dynamic programming approach. Suitable for short skill name comparisons.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// itoa is a tiny zero-alloc int→string formatter for small positive ints
// used in conflict reasons. We avoid strconv import to keep the file lean.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
