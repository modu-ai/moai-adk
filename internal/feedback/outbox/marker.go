package outbox

// marker.go — the marker field format shared by every renderer and parser
// of the `<!-- moai-bugreport:v1 ... -->` blocks (IssueMarker in this
// package; publish's ParseMarker and occurrence renderer on the other
// side). Field values are space-separated key=value tokens; a value that
// itself carries a space — the hook detail's canonical
// "event=<E> handler=<H>" token — is rendered quoted and backslash-escaped,
// and the tokenizer restores it as ONE token, so the marker round-trips
// every closed-set field losslessly (review-gate finding: the handler half
// used to split off into its own key and the truncated detail failed
// ParseDetail).

import "strings"

// QuoteMarkerValue renders one marker field's value: bare when it carries
// no space, quote, or backslash (the common closed-set values — byte-stable
// for the golden and for simple-field greps); otherwise a quoted, escaped
// form SplitMarkerTokens restores losslessly.
func QuoteMarkerValue(v string) string {
	if !strings.ContainsAny(v, " \"\\") {
		return v
	}
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range v {
		if r == '"' || r == '\\' {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}

// SplitMarkerTokens splits a marker block's content into field tokens,
// honoring quotes and backslash escapes: a quoted span is ONE token with
// the quotes consumed and escapes resolved, so a quoted multi-field detail
// survives as a single key=value token. Ok is false on an unterminated
// quote or a dangling escape — untrusted input never parses into partial
// truth.
func SplitMarkerTokens(s string) ([]string, bool) {
	var toks []string
	var cur strings.Builder
	inQuote, escaped := false, false
	flush := func() {
		if cur.Len() > 0 {
			toks = append(toks, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '"':
			inQuote = !inQuote
		case (r == ' ' || r == '\t') && !inQuote:
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	if inQuote || escaped {
		return nil, false
	}
	flush()
	return toks, true
}
