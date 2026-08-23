package feedback

import (
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// advisoriesURL is the private reporting channel SECURITY.md names. It is the
// one thing a blocked user has to act on, so it is carried verbatim rather than
// described.
const advisoriesURL = "https://github.com/modu-ai/moai-adk/security/advisories/new"

// The two sentences below are quoted from SECURITY.md ("Reporting a
// Vulnerability" → "How to Report"). They are quoted rather than paraphrased on
// purpose: a restatement drifts from the policy it claims to carry, and the
// user acting on this message is being told what the project's policy says, not
// what this package thinks it says.
const (
	securityPolicyDoNotOpen = "Do NOT open a public GitHub issue for security vulnerabilities."
	securityPolicyReportVia = "Email the security report to the maintainers via GitHub Security Advisories."
)

// blockedLead states the verdict before the quotation, so the message reads as
// a decision with a citation rather than as a bare policy excerpt.
const blockedLead = "This report is classified as a security vulnerability disclosure and must not be submitted to the public issue tracker."

// blockedReason is the whole Result.Reason for a blocked report.
const blockedReason = blockedLead + " " + securityPolicyDoNotOpen + " " + securityPolicyReportVia + " " + advisoriesURL

// @MX:NOTE: [AUTO] the threshold is deliberately biased toward false positives
//
// vulnerabilityScoreThreshold is the combined score at which signals 2 and 3
// block. The two error directions do not cost the same: a miss publishes a
// vulnerability on a public tracker and cannot be taken back, a false positive
// sends one user down the manual advisory route. The threshold is set where a
// single weak signal passes and any two agree.
const vulnerabilityScoreThreshold = 2

// identifierWeight is what a formal advisory identifier scores. A CVE, CWE or
// GHSA number is not vocabulary a bug report reaches for by accident, so it
// carries the threshold on its own.
const identifierWeight = 2

// vulnerabilityIdentifierPattern matches the formal advisory identifiers.
var vulnerabilityIdentifierPattern = regexp.MustCompile(`(?i)\b(?:CVE-\d{4}-\d{4,}|CWE-\d{1,4}|GHSA-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4})\b`)

// vulnerabilityTerms is signal 3: the vocabulary that suggests a report is a
// disclosure rather than a defect report. Terms that ordinary bug reports use
// ("security", "token", "sandbox") are deliberately absent — they would move
// the score on reports this classifier must let through.
//
// The list is English-only, which is a known limitation rather than an
// oversight: signals 1 and 2 are language-independent and cover the highest-risk
// inputs, and a vocabulary spanning every supported locale is not something a
// fixed list can claim to do.
var vulnerabilityTerms = []string{
	"vulnerability",
	"vulnerabilities",
	"exploit",
	"exploited",
	"exploitable",
	"zero-day",
	"zero day",
	"0day",
	"remote code execution",
	"arbitrary code execution",
	"arbitrary file write",
	"code injection",
	"command injection",
	"sql injection",
	"cross-site scripting",
	"cross site scripting",
	"xss",
	"csrf",
	"ssrf",
	"rce",
	"path traversal",
	"directory traversal",
	"buffer overflow",
	"use-after-free",
	"privilege escalation",
	"authentication bypass",
	"auth bypass",
	"sandbox escape",
	"security flaw",
	"security hole",
	"security issue",
	"security bug",
	"security vulnerability",
	"responsible disclosure",
	"security advisory",
	"proof of concept exploit",
	"insecure deserialization",
	"timing attack",
	"credential leak",
	"secret leak",
	"token leak",
}

// vulnerabilityTermPattern is the vocabulary compiled into one alternation.
var vulnerabilityTermPattern = compileTermPattern(vulnerabilityTerms)

func compileTermPattern(terms []string) *regexp.Regexp {
	quoted := make([]string, 0, len(terms))
	for _, t := range terms {
		quoted = append(quoted, regexp.QuoteMeta(t))
	}
	return regexp.MustCompile(`(?i)\b(?:` + strings.Join(quoted, "|") + `)\b`)
}

// pathTokenPattern carves path-shaped tokens out of prose so the policy's
// file-class patterns can be applied to them. The policy patterns are anchored
// as file paths (`\.pem$`, `\.ssh/.*`), so running them over a whole paragraph
// would answer a different question than the one they were written for.
var pathTokenPattern = regexp.MustCompile(`[A-Za-z0-9_.~/\\-]+`)

// pathTokenTrimCutset drops the sentence punctuation that clings to a path when
// it is written inline in prose, so a trailing period does not defeat a `$`
// anchored pattern.
const pathTokenTrimCutset = ".,-"

// classify decides whether a report may go to a public channel, reading the RAW
// input — the title and the body as the user wrote them, before any masking.
//
// [HARD] The ordering is not an implementation detail. Masking removes exactly
// what signals 1 and 2 read: the credential itself and the key-file path. A
// classifier fed the masked text sees an ordinary report and answers "ok", which
// is a silent false negative — nothing fails, the report is simply published.
// Scrub calls this first for that reason, and AC-F-013 pins the ordering.
//
// Signal 1 (a credential in the text) blocks on its own: whatever the report is
// about, it is carrying a live secret. Signals 2 (a secret-bearing file class is
// named) and 3 (vulnerability vocabulary) combine against a threshold.
func classify(in Input, opt Options) (verdict, reason string) {
	raw := in.Title + "\n" + in.Body

	if matchesAnySecret(raw, rewritePatterns(opt.Policy)) {
		return VerdictBlocked, blockedReason
	}

	if keyFileMentionScore(raw, policyOf(opt))+vocabularyScore(raw) >= vulnerabilityScoreThreshold {
		return VerdictBlocked, blockedReason
	}

	return VerdictOK, ""
}

// policyOf resolves the detector policy a classification runs against.
func policyOf(opt Options) *hook.SecurityPolicy {
	if opt.Policy != nil {
		return opt.Policy
	}
	return hook.DefaultSecurityPolicy()
}

// matchesAnySecret reports whether the raw text carries a credential.
//
// It runs the REWRITE form of the pattern set, not the detector form: the
// detector's case-insensitive compile matches lowercase prose that merely has a
// key's shape, and blocking a report on that would be a false positive with no
// credential behind it.
func matchesAnySecret(s string, patterns []rewritePattern) bool {
	for _, p := range patterns {
		if p.re.MatchString(s) {
			return true
		}
	}
	return false
}

// keyFileMentionScore counts the DISTINCT path tokens in the text that name a
// file class the policy refuses to let anything write — private keys, SSH
// material, cloud credential stores.
//
// Distinct tokens rather than distinct patterns: one path can satisfy several
// policy patterns at once (`~/.ssh/id_rsa` matches two), and counting per
// pattern would let a single mention reach the threshold on its own, which is
// not what "signals 2 and 3 combine" means.
func keyFileMentionScore(s string, policy *hook.SecurityPolicy) int {
	seen := make(map[string]bool)
	score := 0

	for _, token := range pathTokenPattern.FindAllString(s, -1) {
		token = strings.Trim(token, pathTokenTrimCutset)
		if token == "" || seen[token] {
			continue
		}
		seen[token] = true
		for _, re := range policy.DenyPatterns {
			if re.MatchString(token) {
				score++
				break
			}
		}
	}
	return score
}

// vocabularyScore counts the distinct vulnerability terms and adds the weight of
// a formal advisory identifier when one is present.
func vocabularyScore(s string) int {
	seen := make(map[string]bool)
	for _, m := range vulnerabilityTermPattern.FindAllString(s, -1) {
		seen[strings.ToLower(m)] = true
	}
	score := len(seen)

	if vulnerabilityIdentifierPattern.MatchString(s) {
		score += identifierWeight
	}
	return score
}
