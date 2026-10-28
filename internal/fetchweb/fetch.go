package fetchweb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
)

// fetch.go — SPEC-FETCH-WEB-001 M3. The bounded page fetcher (REQ-FW-007):
// a timeout and a maximum response size, both enforced here so no caller can
// forget them. The client is INJECTED — this package never constructs a
// transport for a fetch, and no test can reach a real target page
// (REQ-FW-010); the one client it does construct, NewPageClient (ssrf.go),
// is the guarded production client, never a test transport.

const (
	// FetchTimeout bounds one page fetch. Callers set it on the injected
	// client (the CLI's default client carries it); tests shorten it.
	FetchTimeout = 30_000_000_000 // 30s in nanoseconds (time.Duration literal form kept import-free)

	// MaxResponseBytes bounds the page body. A page larger than this is a
	// fetch error rather than a silent truncation: a truncated HTML document
	// can split into blocks that misrepresent the page.
	MaxResponseBytes = 5 << 20 // 5 MiB
)

// FetchPage GETs url through the injected client and returns the response
// body decoded to UTF-8. Non-2xx statuses, non-HTTP(S) URL schemes, bodies
// over MaxResponseBytes, and bodies that cannot be decoded honestly (F5,
// t1968 sync-audit) are errors.
func FetchPage(ctx context.Context, rawURL string, client *http.Client) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	// The initial URL's scheme is validated here so the refusal is named and
	// client-independent (F1: the redirect policy covers the later hops).
	if u, err := url.Parse(rawURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("fetch web: unsupported URL scheme (only http and https are fetched): %q", rawURL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("fetch web: bad url: %w", err)
	}
	req.Header.Set("User-Agent", "moai-adk-fetchweb/1")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch web: page fetch failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch web: page returned HTTP %d", resp.StatusCode)
	}
	// Read one byte PAST the bound: a body exactly at the bound is fine, a
	// body past it is an error.
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("fetch web: page body could not be read: %w", err)
	}
	if len(body) > MaxResponseBytes {
		return "", fmt.Errorf("fetch web: page body over the %d-byte bound; refused rather than truncated", MaxResponseBytes)
	}
	return decodePageBody(body, resp.Header.Get("Content-Type"))
}

// decodePageBody decodes the fetched bytes to a UTF-8 string (F5, t1968
// sync-audit). Precedence: a UTF-8 BOM is definitive for UTF-8 (and is
// stripped, so it never surfaces as page text); otherwise the Content-Type
// charset label is honored; otherwise the body is taken as UTF-8 — but only
// if it actually is. A declared charset that is unknown, bytes that fail
// their declared decoding, or bytes the decoder can only decode by SILENT
// U+FFFD SUBSTITUTION (F5b, t1968 sync-audit iter2), is a fetch error: a
// latin-1 or EUC-KR page read as raw bytes — or truncated and handed back
// with replacement characters — would be mojibake returned as "verbatim"
// text with exit 0, which the verbatim guarantee forbids. The HTML meta
// charset declaration is deliberately NOT sniffed — with no header charset
// the fallback is UTF-8 (validation included), and a mis-declared page
// fails loudly instead.
func decodePageBody(body []byte, contentType string) (string, error) {
	// A UTF-8 BOM wins over any header label: it is byte-level evidence.
	if len(body) >= 3 && body[0] == 0xEF && body[1] == 0xBB && body[2] == 0xBF {
		body = body[3:]
		if !utf8.Valid(body) {
			return "", fmt.Errorf("fetch web: page has a UTF-8 BOM but is not valid UTF-8; refused rather than returned mojibake")
		}
		return string(body), nil
	}
	label := charsetLabel(contentType)
	if isUTF8Family(label) {
		if !utf8.Valid(body) {
			return "", fmt.Errorf("fetch web: page is not valid UTF-8 (charset %q); refused rather than returned mojibake", label)
		}
		return string(body), nil
	}
	enc, _ := charset.Lookup(label) // label is non-UTF-8 and non-empty here; the name return is unused
	if enc == nil {
		return "", fmt.Errorf("fetch web: Content-Type charset %q is not a known encoding; refused rather than guessed", label)
	}
	decoded, err := enc.NewDecoder().Bytes(body)
	if err != nil {
		return "", fmt.Errorf("fetch web: page charset %q could not be decoded: %w", label, err)
	}
	// F5b (t1968 sync-audit iter2): a SUCCESSFUL decode can still be a
	// corrupted one — the x/text decoders substitute U+FFFD for invalid
	// sequences instead of erroring, so a truncated euc-kr page or a
	// latin-1 page with an undefined byte passed as "verbatim" with exit 0.
	// Scan the decoded text (valid UTF-8 by decoder contract, so the rune
	// index finds exactly the U+FFFD encodings): any U+FFFD the source
	// bytes could not legitimately have encoded is a substitution → the
	// page is undecodable, a fetch error.
	//
	// F5c (t1968 sync-audit iter3) made the provenance COUNT-based (each
	// legitimately-encoded U+FFFD in the source accounts for exactly one
	// decoded U+FFFD); D2 (t1968 debt-closure round) upgrades it to
	// POSITIONAL provenance — see allFFFDLegit for both the corner the
	// count comparison misses and the walk that closes it.
	if i := bytes.IndexRune(decoded, utf8.RuneError); i >= 0 {
		if !allFFFDLegit(body, decoded, enc) {
			return "", fmt.Errorf("fetch web: page charset %q decoded with U+FFFD substitution at byte %d; refused rather than returned corrupted text", label, i)
		}
	}
	// A decoded page may be up to ~3x the raw bound in bytes (one source
	// byte can widen to three UTF-8 bytes); the raw-bound check above bounds
	// the page itself, and this bounded blowup needs no second cap.
	return string(decoded), nil
}

// allFFFDLegit reports whether every U+FFFD in decoded — the whole-buffer
// decoding of body under enc — originates from source bytes that legitimately
// encode U+FFFD under enc (F5b/F5c/D2, t1968 sync-audit + debt-closure
// round). The caller guarantees decoded contains at least one U+FFFD.
//
// An encoding whose encoder cannot encode U+FFFD (every single-byte charset —
// the common corruption class) can never produce one legitimately: any
// decoded U+FFFD is a substitution (the F5b semantics).
//
// Otherwise the check is POSITIONAL (D2): the F5c count comparison balances
// one legitimately-encoded U+FFFD against ANY occurrence of its encoded byte
// form in the source — including a SPURIOUS occurrence inside corrupted
// bytes (probe-confirmed: gb18030 0x81 followed by the 4-byte FFFD form
// decodes the 0x81 and the form's first byte as one valid pair, swallowing
// that byte, and the form's remainder substitutes; decoded count 2 equals
// source form count 2, so the count check balanced and the corrupted page
// passed). The walk below drives the decoder one decision at a time —
// feeding source prefixes of growing length and re-feeding unconsumed bytes,
// the same re-send transform.Reader performs — so each decoded U+FFFD is
// observed together with the exact source span that produced it; legitimate
// means that span IS enc's encoding of U+FFFD.
//
// Reliability fallback: a decoder the walk cannot mirror — output diverging
// from the whole-buffer decode, output without consumption (stateful
// buffering), or a position no prefix length advances — cannot be
// position-mapped, and the check degrades to the F5c count comparison for
// that page rather than guess. Single-byte and CJK charsets (the corruption
// classes this guard exists for) walk cleanly.
func allFFFDLegit(body, decoded []byte, enc encoding.Encoding) bool {
	srcFFFD, encErr := enc.NewEncoder().Bytes([]byte("�"))
	if encErr != nil {
		return false // the charset cannot encode U+FFFD: none is legitimate
	}
	countOK := bytes.Count(decoded, []byte("�")) <= bytes.Count(body, srcFFFD)
	dec := enc.NewDecoder()
	// One decoder decision at minimal prefix length emits at most a few
	// runes; 64 bytes is far past that, and a pathological larger emission
	// surfaces as walk/whole-buffer divergence → the count fallback below.
	buf := make([]byte, 64)
	var out []byte
	pos := 0
	for pos < len(body) {
		advanced := false
		for k := pos + 1; k <= len(body); k++ {
			nDst, nSrc, _ := dec.Transform(buf, body[pos:k], k == len(body))
			if nSrc == 0 {
				if nDst > 0 {
					// Output without consumption: stateful buffered state
					// the walk cannot attribute positionally.
					return countOK
				}
				continue // incomplete at this prefix length: feed one more byte
			}
			advanced = true
			if nDst > 0 {
				out = append(out, buf[:nDst]...)
				if bytes.ContainsRune(buf[:nDst], utf8.RuneError) &&
					!bytes.Equal(body[pos:pos+nSrc], srcFFFD) {
					return false // a U+FFFD whose source span is not the encoded form
				}
			}
			pos += nSrc
			break
		}
		if !advanced {
			// No prefix length up to the whole remaining source made the
			// decoder advance: not walkable positionally.
			return countOK
		}
	}
	if !bytes.Equal(out, decoded) {
		// Stepwise output diverged from the whole-buffer decode: the
		// decoder's behavior depends on call framing this walk cannot
		// reproduce — degrade to the count check rather than guess.
		return countOK
	}
	return true
}

// charsetLabel extracts the charset parameter from a Content-Type header
// value; empty when the header is absent or carries no charset.
func charsetLabel(contentType string) string {
	if contentType == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	return params["charset"]
}

// isUTF8Family covers the labels that mean "the bytes are UTF-8 already":
// the empty label (the UTF-8 fallback), UTF-8's spellings, and ASCII — a
// strict subset of UTF-8, so the UTF-8 validation below is the right check
// for it too.
func isUTF8Family(label string) bool {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "", "utf-8", "utf8", "us-ascii", "ascii", "ansi_x3.4-1968":
		return true
	}
	return false
}
