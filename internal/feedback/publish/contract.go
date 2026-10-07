package publish

// contract.go — the issue contract (REQ-ANON-019, AC-020): the title key,
// the machine-readable marker blocks, and the untrusted-input rule.
//
// The contract's authority chain: the TITLE KEY (`[auto-report] <kind>
// <fingerprint>`) and the ISSUE BODY's marker block are authoritative; a
// COMMENT's marker is untrusted input. Anyone can post a comment shaped
// like a marker, so a consumer re-derives fingerprint, kind, version,
// commit, and os_arch from the title key and the body block and ignores
// any marker field that disagrees — while a marker-shaped comment still
// counts toward the advisory occurrence total (design section 7: the count
// is advisory; a forged marker can inflate it, and that is accepted).

import (
	"strings"

	"github.com/modu-ai/moai-adk/internal/feedback"
)

// markerPrefix opens every machine-readable block the pipeline renders.
const markerPrefix = "<!-- moai-bugreport:v1 "

// occurrenceToken tags an occurrence comment's marker block; the issue
// body's block carries no token (its first field is schema=).
const occurrenceToken = "occurrence"

// ParseTitleKey parses the contract title `[auto-report] <kind>
// <fingerprint>` back to its two fields.
func ParseTitleKey(title string) (kind, fingerprint string, ok bool) {
	const prefix = "[auto-report] "
	rest, found := strings.CutPrefix(title, prefix)
	if !found {
		return "", "", false
	}
	kind, fingerprint, found = strings.Cut(rest, " ")
	if !found || kind == "" || fingerprint == "" || strings.Contains(fingerprint, " ") {
		return "", "", false
	}
	return kind, fingerprint, true
}

// MarkerFields are the key=value fields of one marker block.
type MarkerFields map[string]string

// ParseMarker parses one `<!-- moai-bugreport:v1 ... -->` block out of a
// body: the first field may be the bare occurrence token; the rest are
// key=value pairs. Ok is false when no well-formed block exists — untrusted
// input never parses into partial truth.
func ParseMarker(body string) (MarkerFields, bool) {
	start := strings.Index(body, markerPrefix)
	if start < 0 {
		return nil, false
	}
	rest := body[start+len(markerPrefix):]
	end := strings.Index(rest, "-->")
	if end < 0 {
		return nil, false
	}
	fields := MarkerFields{}
	for i, tok := range strings.Fields(rest[:end]) {
		if i == 0 && tok == occurrenceToken {
			continue
		}
		k, v, found := strings.Cut(tok, "=")
		if !found || k == "" {
			return nil, false
		}
		fields[k] = v
	}
	if fields["schema"] == "" {
		return nil, false
	}
	return fields, true
}

// IsOccurrenceComment reports whether a comment body carries the
// occurrence marker SHAPE. Shape match only: the fields are untrusted and
// never consulted (they exist for human readers and downstream consumers,
// which re-derive from the title key).
func IsOccurrenceComment(body string) bool {
	fields, ok := ParseMarker(body)
	if !ok {
		return false
	}
	// ParseMarker skips a leading bare token; its presence distinguishes
	// the occurrence block from the issue body's block. Re-detect it here:
	// the first field after the prefix must be the bare token.
	start := strings.Index(body, markerPrefix)
	rest := body[start+len(markerPrefix):]
	end := strings.Index(rest, "-->")
	if end < 0 {
		return false
	}
	toks := strings.Fields(rest[:end])
	return len(toks) > 0 && toks[0] == occurrenceToken && fields != nil
}

// OccurrenceCount counts the marker-shaped comments in a remote issue's
// comment array — the advisory occurrence total is one (the issue itself)
// plus this count. Forged markers count; that is the accepted
// approximation (design section 7).
func OccurrenceCount(comments []RemoteComment) int {
	n := 0
	for _, c := range comments {
		if IsOccurrenceComment(c.Body) {
			n++
		}
	}
	return n
}

// OccurrenceComment renders one occurrence comment's body for a queued
// item: the item's OWN marker fields (recovered from the queued body — the
// one render, reused), re-tagged as an occurrence. An item whose body
// carries no marker block (a manual-flow item that never belongs on this
// path) renders empty and the sender refuses to comment.
func OccurrenceComment(item feedback.QueueItem) string {
	fields, ok := ParseMarker(item.Body)
	if !ok {
		return ""
	}
	var b strings.Builder
	b.WriteString(markerPrefix)
	b.WriteString(occurrenceToken)
	for _, k := range markerFieldOrder {
		if v, ok := fields[k]; ok {
			b.WriteString(" " + k + "=" + v)
		}
	}
	b.WriteString(" -->")
	return b.String()
}

// markerFieldOrder pins the field order of rendered blocks (stable output
// for the golden and for grep-able issues).
var markerFieldOrder = []string{"schema", "fingerprint", "kind", "version", "commit", "os_arch", "frames", "detail"}

// ContractFields derives a public issue's authoritative fields from the
// TITLE KEY and the ISSUE BODY's marker block — never from any comment.
// This is the function a consumer runs to re-derive what the title
// asserts; marker fields that disagree with it are ignored by
// construction because they are never read.
func ContractFields(title, body string) (MarkerFields, bool) {
	fields := MarkerFields{}
	kind, fingerprint, ok := ParseTitleKey(title)
	if !ok {
		return nil, false
	}
	fields["kind"] = kind
	fields["fingerprint"] = fingerprint
	bodyFields, ok := ParseMarker(body)
	if !ok {
		return nil, false
	}
	for k, v := range bodyFields {
		if _, fromTitle := fields[k]; !fromTitle {
			fields[k] = v
		}
	}
	return fields, true
}
