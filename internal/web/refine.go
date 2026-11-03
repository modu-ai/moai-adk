// refine.go — every text that reaches the Agentic Factory panels passes through
// core.Sanitize, the one sanitizer (Issue #1948, B4). The web only displays what
// the monitor, the claim store and the context model hand it, so it refines on
// the way to the page instead of trusting each source to have masked: an error
// note quoting a path, a persisted detail naming a person, a context item
// owned by an e-mail address. A row that had something masked says so, so a
// reader can tell "[REDACTED]" from "the source wrote that".
package web

import "github.com/modu-ai/moai-adk/internal/agenticfactory/core"

// refineText masks every secret and piece of personal data in s.
func refineText(s string) string {
	out, _ := core.Sanitize(s)
	return out
}

// anyMasked reports whether refining changed (masked something in) any of the
// given raw texts.
func anyMasked(raws ...string) bool {
	for _, s := range raws {
		if _, cats := core.Sanitize(s); len(cats) > 0 {
			return true
		}
	}
	return false
}
