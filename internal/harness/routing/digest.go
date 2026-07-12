package routing

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// digestPrefix labels the request digest so consumers can recognize the hash
// algorithm without parsing.
const digestPrefix = "sha256:"

// digestHexLen is the truncated hex length of the request digest. 12 hex chars
// = 48 bits — enough to distinguish requests for aggregation, far too short to
// invert (and the raw text is never persisted anyway; REQ-HEV-005).
const digestHexLen = 12

// RequestDigest returns a privacy-preserving digest of a raw request string:
// digestPrefix + the first digestHexLen hex chars of the SHA-256 of the
// normalized request. The raw text is NEVER persisted by this package
// (REQ-HEV-005) — only this digest and the coarse class (see ClassifyRequest)
// are recorded. The digest is computed in-process; callers pass raw request
// text via stdin and discard it after hashing.
func RequestDigest(raw string) string {
	sum := sha256.Sum256([]byte(normalizeRequest(raw)))
	return digestPrefix + fmt.Sprintf("%x", sum)[:digestHexLen]
}

// normalizeRequest lowercases and collapses whitespace so trivial formatting
// differences map to the same digest/class.
func normalizeRequest(raw string) string {
	return strings.ToLower(strings.Join(strings.Fields(raw), " "))
}

// ClassifyRequest derives the coarse, non-verbatim request class (REQ-HEV-005,
// pinned user decision 2). The class is one of the closed enum
// feature|bugfix|refactor|docs|question|pipeline|other and is keyword-derived —
// it never carries raw request text. Classification is deterministic (fixed
// precedence) and general (token/phrase matching, not fixture-specific).
func ClassifyRequest(raw string) string {
	norm := normalizeRequest(raw)

	tokens := make(map[string]bool)
	for _, t := range strings.Fields(norm) {
		tokens[strings.Trim(t, ".,!?;:\"'()[]{}")] = true
	}
	hasWord := func(ws ...string) bool {
		for _, w := range ws {
			if tokens[w] {
				return true
			}
		}
		return false
	}
	hasPhrase := func(ps ...string) bool {
		for _, p := range ps {
			if strings.Contains(norm, p) {
				return true
			}
		}
		return false
	}

	switch {
	case hasWord("refactor", "restructure", "cleanup", "rename", "extract", "dedupe") || hasPhrase("clean up"):
		return "refactor"
	case hasWord("fix", "bug", "broken", "regression", "crash", "failing", "error", "hotfix"):
		return "bugfix"
	case hasWord("doc", "docs", "readme", "changelog", "documentation", "comment"):
		return "docs"
	case hasWord("pipeline") || hasPhrase("/moai run", "/moai sync", "/moai plan"):
		return "pipeline"
	case hasWord("how", "why", "what", "which", "explain", "?") || strings.HasSuffix(strings.TrimSpace(norm), "?"):
		return "question"
	case hasWord("add", "implement", "feature", "create", "new", "build", "support"):
		return "feature"
	default:
		return "other"
	}
}
