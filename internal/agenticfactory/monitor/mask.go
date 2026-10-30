package monitor

import (
	"regexp"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Mask hides every secret and piece of personal data the one core sanitizer knows. The bool reports
// whether anything was replaced.
func Mask(s string) (string, bool) {
	out, cats := core.Sanitize(s)
	return out, len(cats) > 0
}

// MaskField masks one displayed value the way every field of a Row is masked: what the core sanitizer
// knows is replaced, and text that still looks like an unknown credential (a long mixed letter/digit run
// that is not a full commit SHA) is dropped altogether. redacted reports that anything was replaced or
// dropped; a dropped value is the empty string.
func MaskField(s string) (value string, redacted bool) {
	out, cats := core.Sanitize(s)
	for _, c := range cats {
		if c == core.CategoryOpaqueToken {
			return "", true
		}
	}
	return out, len(cats) > 0
}
