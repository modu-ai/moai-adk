package template

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The PowerShell tool has its own permission namespace: a Bash(...) deny
// rule does not stop the same command issued through the PowerShell tool.
// Every Bash deny therefore either carries a PowerShell counterpart with the
// same pattern, or sits on the exclusion list below with a stated reason.
// The guard is closed-world: a new Bash deny with neither fails the test.

// psDenyExclusions lists the Bash deny rules that intentionally have no
// PowerShell counterpart, with the reason.
var psDenyExclusions = map[string]string{
	// Filesystem-root removals: Claude Code's built-in Remove-Item / cmd
	// system-path and wildcard checks already deny these on the PowerShell
	// tool in every permission mode.
	"Bash(rm -rf /:*)":                        "builtin",
	"Bash(rm -rf /\\* *)":                     "builtin",
	"Bash(rm -rf ~:*)":                        "builtin",
	"Bash(rm -rf ~/\\* *)":                    "builtin",
	"Bash(rm -rf C:/:*)":                      "builtin",
	"Bash(rm -rf C:/\\* *)":                   "builtin",
	"Bash(del /S /Q C:/:*)":                   "builtin",
	"Bash(rmdir /S /Q C:/:*)":                 "builtin",
	"Bash(Remove-Item -Recurse -Force C:/:*)": "builtin",
	// "kill" is a PowerShell alias of Stop-Process on Windows; after alias
	// canonicalization a rule written as "kill -9" may never match.
	"Bash(kill -9:*)": "alias head, unmeasured",
	// PowerShell rules match case-insensitively, so a TRUNCATE counterpart
	// would also block the lowercase `truncate` file utility.
	"Bash(TRUNCATE:*)": "case-fold over-block",
}

// psCounterpart returns the PowerShell rule mirroring a Bash rule: the same
// pattern text in the PowerShell namespace.
func psCounterpart(bashRule string) string {
	return "PowerShell(" + strings.TrimPrefix(bashRule, "Bash(")
}

// psParityViolations checks a deny list against the exclusion list and
// returns one message per violation, each naming the offending rule.
func psParityViolations(deny []string, exclusions map[string]string) []string {
	present := make(map[string]bool, len(deny))
	bash := make(map[string]bool)
	for _, r := range deny {
		present[r] = true
		if strings.HasPrefix(r, "Bash(") {
			bash[r] = true
		}
	}
	var out []string
	for _, r := range deny {
		switch {
		case strings.HasPrefix(r, "Bash("):
			_, excluded := exclusions[r]
			hasPS := present[psCounterpart(r)]
			if excluded && hasPS {
				out = append(out, "excluded Bash deny also has a PowerShell counterpart: "+r)
			}
			if !excluded && !hasPS {
				out = append(out, "Bash deny has neither a PowerShell counterpart nor an exclusion: "+r)
			}
		case strings.HasPrefix(r, "PowerShell("):
			if !bash["Bash("+strings.TrimPrefix(r, "PowerShell(")] {
				out = append(out, "PowerShell deny maps to no Bash deny: "+r)
			}
			if strings.Contains(r, "\\:") {
				out = append(out, "PowerShell deny escapes ':' and cannot match a real drive path: "+r)
			}
			if strings.HasSuffix(r, ":*)") && strings.Contains(strings.TrimSuffix(r, ":*)"), "*") {
				out = append(out, "PowerShell deny mixes a non-trailing '*' with ':*': "+r)
			}
		}
	}
	for r := range exclusions {
		if !bash[r] {
			out = append(out, "exclusion names a rule absent from the deny list: "+r)
		}
	}
	sort.Strings(out)
	return out
}

// psRuleMatches models the documented PowerShell rule semantics: ":*" equals
// a trailing " *"; a trailing " *" matches the bare prefix or the prefix
// followed by a space and anything; any other "*" (including one glued to
// text at the end) matches any character sequence; matching is
// case-insensitive. Alias canonicalization and compound-command splitting are
// not modelled.
func psRuleMatches(rule, command string) bool {
	pattern := strings.TrimSuffix(strings.TrimPrefix(rule, "PowerShell("), ")")
	if strings.HasSuffix(pattern, ":*") {
		pattern = strings.TrimSuffix(pattern, ":*") + " *"
	}
	tail := ""
	if strings.HasSuffix(pattern, " *") {
		pattern = strings.TrimSuffix(pattern, " *")
		tail = "(?: .*)?"
	}
	parts := strings.Split(pattern, "*")
	for i, p := range parts {
		parts[i] = regexp.QuoteMeta(p)
	}
	re := regexp.MustCompile("(?is)^" + strings.Join(parts, ".*") + tail + "$")
	return re.MatchString(command)
}

// Benign commands no PowerShell deny may block, and destructive commands
// that must stay blocked.
var (
	psBenignSample = []string{
		"Remove-Item ./build -Recurse -Force",
		"Get-ChildItem C:/",
		"git push origin HEAD",
		"git clean -n",
		"git status",
		"Format-Table",
		"truncate -s 0 app.log",
		"redis-cli GET key",
	}
	psKnownDeny = []string{
		"git push --force origin main",
		"git clean -fdx",
		"git -C repo reset --hard HEAD",
		"Format-Volume -DriveLetter D",
		"redis-cli FLUSHALL",
		"psql -c DROP TABLE t",
	}
)

// psMatchViolations reports benign commands that some PowerShell deny
// matches, and known-deny commands that none matches.
func psMatchViolations(deny []string, match func(rule, command string) bool) []string {
	var ps []string
	for _, r := range deny {
		if strings.HasPrefix(r, "PowerShell(") {
			ps = append(ps, r)
		}
	}
	matchedBy := func(cmd string) string {
		for _, r := range ps {
			if match(r, cmd) {
				return r
			}
		}
		return ""
	}
	var out []string
	for _, cmd := range psBenignSample {
		if r := matchedBy(cmd); r != "" {
			out = append(out, "benign command "+cmd+" is blocked by "+r)
		}
	}
	for _, cmd := range psKnownDeny {
		if matchedBy(cmd) == "" {
			out = append(out, "destructive command "+cmd+" is not blocked by any PowerShell deny")
		}
	}
	return out
}

func renderedDeny(t *testing.T, platform string) []string {
	t.Helper()
	var settings struct {
		Permissions struct {
			Deny []string `json:"deny"`
		} `json:"permissions"`
	}
	rendered := renderTemplate(t, ".claude/settings.json.tmpl", testContext(platform))
	if err := json.Unmarshal([]byte(rendered), &settings); err != nil {
		t.Fatalf("rendered settings.json: %v", err)
	}
	return settings.Permissions.Deny
}

func TestSettingsTemplatePowerShellDenyParity(t *testing.T) {
	for _, platform := range []string{"darwin", "linux", "windows"} {
		t.Run(platform, func(t *testing.T) {
			for _, v := range psParityViolations(renderedDeny(t, platform), psDenyExclusions) {
				t.Error(v)
			}
		})
	}
}

func TestSettingsTemplatePowerShellDenyNoOverBlock(t *testing.T) {
	for _, v := range psMatchViolations(renderedDeny(t, "windows"), psRuleMatches) {
		t.Error(v)
	}
}

// The guard proves it can fail: each mutation of the rendered deny list must
// produce a violation naming the offending rule.
func TestSettingsTemplatePowerShellDenyGuardDetectsMutations(t *testing.T) {
	base := renderedDeny(t, "windows")
	without := func(drop string) []string {
		var out []string
		for _, r := range base {
			if r != drop {
				out = append(out, r)
			}
		}
		return out
	}
	with := func(add string) []string {
		return append(append([]string(nil), base...), add)
	}
	cases := []struct {
		name, offender string
		deny           []string
	}{
		{"counterpart removed", "neither a PowerShell counterpart nor an exclusion: Bash(git clean -fdx:*)", without("PowerShell(git clean -fdx:*)")},
		{"unmapped Bash deny added", "neither a PowerShell counterpart nor an exclusion: Bash(wipefs:*)", with("Bash(wipefs:*)")},
		{"escaped colon in PowerShell rule", `escapes ':' and cannot match a real drive path: PowerShell(rm -rf C\:/:*)`, with(`PowerShell(rm -rf C\:/:*)`)},
		{"middle wildcard mixed with :*", "mixes a non-trailing '*' with ':*': PowerShell(git * clean -fdx:*)", with("PowerShell(git * clean -fdx:*)")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := psParityViolations(tc.deny, psDenyExclusions)
			if !strings.Contains(strings.Join(got, "\n"), tc.offender) {
				t.Errorf("guard did not name %q; violations: %v", tc.offender, got)
			}
		})
	}
	t.Run("always-false matcher", func(t *testing.T) {
		got := psMatchViolations(base, func(string, string) bool { return false })
		if len(got) != len(psKnownDeny) {
			t.Errorf("an always-false matcher must fail every known-deny control; got %v", got)
		}
	})
}
