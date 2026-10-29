package core

import "regexp"

// rule is one masking pattern. A rule with keepLead matches one character of
// left context (so a path is only masked where it starts); that character is
// put back.
type rule struct {
	re       *regexp.Regexp
	keepLead bool
}

var secretRules = []rule{
	{re: regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`)},
	{re: regexp.MustCompile(`github_pat_[A-Za-z0-9_]{20,}`)},
	{re: regexp.MustCompile(`sk-[A-Za-z0-9_-]{16,}`)},
	{re: regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/=-]{8,}`)},
	{re: regexp.MustCompile(`(?i)(token|api[_-]?key|secret|password|cookie)\s*[:=]\s*[^\s",}]+`)},
	// An e-mail address, including the user@host part of an ssh remote.
	{re: regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}`)},
	// An absolute home directory and everything under it.
	{re: regexp.MustCompile(`(?:^|[^A-Za-z0-9_./-])/(?:Users|home)/[^\s"'/\\]+(?:/[^\s"']*)?`), keepLead: true},
	{re: regexp.MustCompile(`[A-Za-z]:\\Users\\[^\s"'\\]+(?:\\[^\s"']*)?`)},
}

// Redact masks tokens, keys, passwords, cookies, e-mail addresses and absolute
// home paths, and reports how many values it masked. Text that leaves the
// process (an Issue comment, a stored error, a context ledger) passes through
// it first.
func Redact(s string) (string, int) {
	n := 0
	for _, r := range secretRules {
		r := r
		s = r.re.ReplaceAllStringFunc(s, func(m string) string {
			n++
			if r.keepLead && m != "" && m[0] != '/' {
				return m[:1] + "[REDACTED]"
			}
			return "[REDACTED]"
		})
	}
	return s, n
}
