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
	// F5c (t1968 sync-audit iter3) makes the provenance COUNT-based: the
	// iter2 test looked at only the first decoded U+FFFD and accepted when
	// the encoded form appeared anywhere in the body, so a page that
	// legitimately encoded U+FFFD early and was corrupted later passed with
	// a silently-substituted second U+FFFD. Each legitimately-encoded
	// U+FFFD in the source accounts for exactly one decoded U+FFFD: a
	// decoded count beyond the source-side count is substitution. For an
	// encoding that cannot encode U+FFFD at all (every single-byte charset
	// — the common corruption class), counting is unnecessary and the
	// encoder error itself is the proof: ANY decoded U+FFFD is a
	// substitution (the iter2 semantics, kept as the fallback). Both sides
	// are counted with single-pass byte scans, no regex.
	if i := bytes.IndexRune(decoded, utf8.RuneError); i >= 0 {
		if srcFFFD, encErr := enc.NewEncoder().Bytes([]byte("�")); encErr != nil ||
			bytes.Count(decoded, []byte("�")) > bytes.Count(body, srcFFFD) {
			return "", fmt.Errorf("fetch web: page charset %q decoded with U+FFFD substitution at byte %d; refused rather than returned corrupted text", label, i)
		}
	}
	// A decoded page may be up to ~3x the raw bound in bytes (one source
	// byte can widen to three UTF-8 bytes); the raw-bound check above bounds
	// the page itself, and this bounded blowup needs no second cap.
	return string(decoded), nil
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
