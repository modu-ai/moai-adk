package escalation

import "regexp"

// masked replaces every credential value the detector would otherwise write.
const masked = "***"

// Credential shapes masked in shell commands and failure text before either
// reaches a record, the card log, or the state file.
var (
	// scheme://user:secret@host — the whole userinfo, since a token can sit
	// in the user part as well.
	urlUserinfoRe = regexp.MustCompile(`(\b[A-Za-z][A-Za-z0-9+.-]*://)[^/\s@'"]+@`)
	// Authorization: [scheme] value (header argument or echoed header).
	authHeaderRe = regexp.MustCompile(`(?i)(\bauthorization\s*:\s*)((?:bearer|basic|token|digest)\s+)?[^\s'"]+`)
	// name=value / name: value where the name carries a credential word.
	secretAssignRe = regexp.MustCompile(`(?i)(\b[A-Za-z0-9_.-]*(?:token|password|passwd|secret|api[_-]?key|access[_-]?key|credential)[A-Za-z0-9_.-]*\s*[=:]\s*)("[^"]*"|'[^']*'|[^\s'"&;|]+)`)
	// --password value / -token value (flag followed by a separate value).
	secretFlagRe = regexp.MustCompile(`(?i)(\s--?[A-Za-z0-9-]*(?:token|password|passwd|secret|api-?key|credential)[A-Za-z0-9-]*\s+)([^\s'"-][^\s'"]*)`)
	// Well-known token prefixes, wherever they appear.
	tokenPrefixRe = regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr|github_pat|glpat|xox[abprs])[-_][A-Za-z0-9_-]{8,}`)
)

// MaskCommand masks credentials in a shell command or failure text before it
// is recorded (URL userinfo, Authorization values, token/password-style
// assignments and flags, known token prefixes). The result is deterministic:
// the same command shape masks to the same text whatever the secret, so a key
// built from it does not split.
//
// @MX:NOTE: [AUTO] class 6 and class 8 name the observed command; every such write passes through this mask first, including the class 8 streak key kept in the state file
func MaskCommand(s string) string {
	s = authHeaderRe.ReplaceAllString(s, "${1}${2}"+masked)
	s = urlUserinfoRe.ReplaceAllString(s, "${1}"+masked+"@")
	s = secretAssignRe.ReplaceAllString(s, "${1}"+masked)
	s = secretFlagRe.ReplaceAllString(s, "${1}"+masked)
	return tokenPrefixRe.ReplaceAllString(s, masked)
}
