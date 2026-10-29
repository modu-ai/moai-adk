package core

import "regexp"

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`github_pat_[A-Za-z0-9_]{20,}`),
	regexp.MustCompile(`sk-[A-Za-z0-9_-]{16,}`),
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/=-]{8,}`),
	regexp.MustCompile(`(?i)(token|api[_-]?key|secret|password|cookie)\s*[:=]\s*[^\s",}]+`),
}

// Redact masks token, key, password and cookie shapes and reports how many
// values it masked. Text that leaves the process (an Issue comment, a stored
// error) passes through it first.
func Redact(s string) (string, int) {
	n := 0
	for _, re := range secretPatterns {
		s = re.ReplaceAllStringFunc(s, func(string) string { n++; return "[REDACTED]" })
	}
	return s, n
}
