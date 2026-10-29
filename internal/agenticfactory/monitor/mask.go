package monitor

import "regexp"

const redacted = "[REDACTED]"

type maskRule struct {
	re   *regexp.Regexp
	repl string
}

var maskRules = []maskRule{
	{regexp.MustCompile(`(?i)(authorization\s*:\s*(?:bearer|basic|token)\s+)\S+`), "${1}" + redacted},
	{regexp.MustCompile(`(?i)((?:set-)?cookie\s*:\s*)[^\n]+`), "${1}" + redacted},
	{regexp.MustCompile(`(?i)(https?://[^/\s:@]+:)[^@\s/]+(@)`), "${1}" + redacted + "${2}"},
	{regexp.MustCompile(`(?i)\b([a-z0-9_]*(?:token|secret|passw(?:or)?d|api_?key|key|credential)[a-z0-9_]*)=("[^"]*"|'[^']*'|\S+)`), "${1}=" + redacted},
	{regexp.MustCompile(`(?i)("[a-z_]*(?:api_?key|token|secret|passw(?:or)?d)[a-z_]*"\s*:\s*")[^"]*(")`), "${1}" + redacted + "${2}"},
	{regexp.MustCompile(`github_pat_[A-Za-z0-9_]{20,}`), redacted},
	{regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`), redacted},
	{regexp.MustCompile(`sk-[A-Za-z0-9_\-]{16,}`), redacted},
	{regexp.MustCompile(`\b(?:AKIA|ASIA)[A-Z0-9]{16}\b`), redacted},
}

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

var (
	blobRe = regexp.MustCompile(`[A-Za-z0-9_+=\-]{32,}`)
	hexRe  = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
)

// Mask replaces known credential shapes. The bool reports whether anything was replaced.
func Mask(s string) (string, bool) {
	out := s
	for _, r := range maskRules {
		out = r.re.ReplaceAllString(out, r.repl)
	}
	return out, out != s
}

// looksOpaque reports a long mixed letter/digit run that is not a full commit SHA. Such text could
// be a credential in a shape Mask does not know, so the caller drops it (fail-closed).
func looksOpaque(s string) bool {
	for _, b := range blobRe.FindAllString(s, -1) {
		if hexRe.MatchString(b) {
			continue
		}
		var lo, up, dg bool
		for _, c := range b {
			switch {
			case c >= 'a' && c <= 'z':
				lo = true
			case c >= 'A' && c <= 'Z':
				up = true
			case c >= '0' && c <= '9':
				dg = true
			}
		}
		if lo && dg || up && dg {
			return true
		}
	}
	return false
}
