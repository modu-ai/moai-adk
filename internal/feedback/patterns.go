package feedback

import (
	"log/slog"
	"regexp"
	"strings"

	ghsecret "github.com/modu-ai/moai-adk/internal/github"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// googleAPIKeyPattern is the one addition this package makes to the hook's
// sensitive-content set. The pattern lives only in the ast-grep rule set, so
// the two collections are not in a containment relation and the scrubber takes
// their union. It is added here, on the rewrite side, and deliberately NOT in
// the hook's own list: adding it there would widen the Write/Edit deny verdict,
// which is a behaviour change outside this scope.
const googleAPIKeyPattern = `AIza[0-9A-Za-z_-]{35}`

// caseInsensitiveFlag is the inline flag the detector prepends to every
// sensitive-content pattern (see compilePatterns in internal/hook).
const caseInsensitiveFlag = "(?i)"

// armorHeaderPrefix opens a PEM-style armored block. It is a block-structure
// marker, not a credential pattern: it only ever extends the span of a match
// the policy already made.
const armorHeaderPrefix = "-----BEGIN"

// armorTerminator closes a PEM-style armored block.
var armorTerminator = regexp.MustCompile(`-----END[^-\n]*-----`)

// rewritePattern is one sensitive-content pattern prepared for the REWRITE
// path, as opposed to the detector path it came from.
type rewritePattern struct {
	re *regexp.Regexp

	// blockAnchored marks a pattern that matches only the opening line of a
	// multi-line block. Masking just that line leaves the key material itself
	// in the text, so the replacement span runs through the block terminator —
	// or to the end of the input when the block is truncated.
	blockAnchored bool
}

// rewritePatterns adapts a detector policy into rewrite patterns.
//
// The policy object is taken whole rather than copied pattern by pattern, so a
// project's security.extra_sensitive_content_patterns extensions reach the
// scrubber through the same merge contract the hook already guarantees.
//
// [HARD] The detector compiles every pattern case-insensitively. That widening
// is free for a detector — a false positive costs one refused write — but a
// rewriter pays for it in destroyed prose: a lowercase run of ordinary words
// that happens to fit an uppercase key shape gets masked out of the report. The
// rewrite path therefore recompiles each pattern WITHOUT the inline flag.
// Credentials are emitted in their canonical case by the systems that issue
// them, so the case-sensitive form still matches every real key.
func rewritePatterns(policy *hook.SecurityPolicy) []rewritePattern {
	if policy == nil {
		policy = hook.DefaultSecurityPolicy()
	}

	sources := make([]string, 0, len(policy.SensitiveContentPatterns)+1)
	for _, re := range policy.SensitiveContentPatterns {
		sources = append(sources, strings.TrimPrefix(re.String(), caseInsensitiveFlag))
	}
	sources = append(sources, googleAPIKeyPattern)

	seen := make(map[string]bool, len(sources))
	out := make([]rewritePattern, 0, len(sources))
	for _, src := range sources {
		if seen[src] {
			continue
		}
		seen[src] = true

		re, err := regexp.Compile(src)
		if err != nil {
			// Same posture as the detector: an unusable pattern is skipped, it
			// does not abort the scrub.
			slog.Warn("feedback: skipping uncompilable sensitive-content pattern", "pattern", src, "error", err)
			continue
		}
		out = append(out, rewritePattern{re: re, blockAnchored: strings.HasPrefix(src, armorHeaderPrefix)})
	}
	return out
}

// maskSecrets applies every rewrite pattern to s and returns the masked text
// plus the number of spans replaced.
func maskSecrets(s string, patterns []rewritePattern) (string, int) {
	total := 0
	for _, p := range patterns {
		masked, n := maskPattern(s, p)
		s = masked
		total += n
	}
	return s, total
}

// maskPattern replaces every span p matches in s with the adopted masker's
// output.
func maskPattern(s string, p rewritePattern) (string, int) {
	var b strings.Builder
	count := 0
	idx := 0

	for idx <= len(s) {
		loc := p.re.FindStringIndex(s[idx:])
		if loc == nil {
			break
		}
		start := idx + loc[0]
		end := idx + loc[1]
		if p.blockAnchored {
			end = blockEnd(s, end)
		}
		if end <= start {
			break
		}

		b.WriteString(s[idx:start])
		b.WriteString(ghsecret.MaskSecret(s[start:end]))
		count++
		idx = end
	}

	if count == 0 {
		return s, 0
	}
	b.WriteString(s[idx:])
	return b.String(), count
}

// blockEnd returns the offset just past the armored block's terminator, or the
// end of the input when the block is truncated. A truncated block is the more
// dangerous case, not the safer one: the key material is still there, only the
// closing line is missing.
func blockEnd(s string, from int) int {
	loc := armorTerminator.FindStringIndex(s[from:])
	if loc == nil {
		return len(s)
	}
	return from + loc[1]
}
