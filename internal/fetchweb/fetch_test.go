package fetchweb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/html/charset"
)

// fetch_test.go — SPEC-FETCH-WEB-001 M3. The bounded page fetcher: a timeout
// and a maximum response size, against a fake page server only (REQ-FW-007,
// REQ-FW-010 — no test reaches a real target page).

func TestFetchPage_ReturnsBodyOnOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body><p>hello page</p></body></html>"))
	}))
	defer srv.Close()
	page, err := FetchPage(t.Context(), srv.URL, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(page, "hello page") {
		t.Errorf("page body = %q", page)
	}
}

func TestFetchPage_Non2xxIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	if _, err := FetchPage(t.Context(), srv.URL, srv.Client()); err == nil {
		t.Fatal("404 must be a fetch error")
	}
}

func TestFetchPage_OversizeBodyIsError(t *testing.T) {
	big := strings.Repeat("x", MaxResponseBytes+1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(big))
	}))
	defer srv.Close()
	if _, err := FetchPage(t.Context(), srv.URL, srv.Client()); err == nil {
		t.Fatal("a body over the max response size must be a fetch error, not a silent truncation to valid HTML")
	}
}

func TestFetchPage_UnreachableIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // closed immediately: the port refuses
	if _, err := FetchPage(context.Background(), srv.URL, srv.Client()); err == nil {
		t.Fatal("unreachable page must be a fetch error")
	}
}

// F5 (t1968 sync-audit): the response charset is honored. A page served as
// iso-8859-1 decodes to correct UTF-8 — "café naïve" — instead of returning
// raw latin-1 bytes as "verbatim" text with exit 0.
func TestFetchPage_HonorsContentTypeCharset(t *testing.T) {
	// "café naïve" encoded in iso-8859-1: é = 0xE9, ï = 0xEF.
	latin1 := []byte{'c', 'a', 'f', 0xE9, ' ', 'n', 'a', 0xEF, 'v', 'e'}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=iso-8859-1")
		_, _ = w.Write(latin1)
	}))
	defer srv.Close()
	page, err := FetchPage(t.Context(), srv.URL, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(page, "café naïve") {
		t.Errorf("latin-1 page = %q, want the decoded UTF-8 text %q", page, "café naïve")
	}
}

// F5 (t1968 sync-audit): a UTF-8 BOM is definitive for UTF-8 and is stripped,
// so it does not surface as a mojibake character inside the page text.
func TestFetchPage_UTF8BOMStripped(t *testing.T) {
	body := append([]byte{0xEF, 0xBB, 0xBF}, []byte("<html><body><p>bom page</p></body></html>")...)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	page, err := FetchPage(t.Context(), srv.URL, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(page, "\ufeff") { // U+FEFF escape
		t.Error("UTF-8 BOM must be stripped, not returned as page text")
	}
	if !strings.Contains(page, "bom page") {
		t.Errorf("page = %q, want the body after the BOM", page)
	}
}

// F5 (t1968 sync-audit): when no charset is declared the body must at least
// BE valid UTF-8 — undecodable bytes are a fetch error, never mojibake
// returned as verbatim text with exit 0.
func TestFetchPage_InvalidBytesWithoutCharsetIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("plain \xff text")) // 0xFF: invalid in UTF-8
	}))
	defer srv.Close()
	if _, err := FetchPage(t.Context(), srv.URL, srv.Client()); err == nil {
		t.Fatal("undecodable body without a charset must be a fetch error, not silently returned")
	}
}

// F1 (t1968 sync-audit): the initial URL's scheme is validated — only http
// and https are fetched, whatever the client would otherwise do with a
// file:/// or other scheme.
func TestFetchPage_NonHTTPSchemeIsError(t *testing.T) {
	if _, err := FetchPage(t.Context(), "file:///etc/passwd", http.DefaultClient); err == nil || !strings.Contains(err.Error(), "scheme") {
		t.Fatalf("err = %v, want a named non-HTTP(S) scheme refusal", err)
	}
}

// F5b (t1968 sync-audit iter2): a DECLARED charset whose bytes are corrupted
// decodes with silent U+FFFD substitution — the x/text decoder replaces
// invalid sequences instead of erroring, so a truncated euc-kr page passed as
// "verbatim" with exit 0. A substitution means the page is undecodable: a
// fetch error, never a corrupted passthrough.
func TestFetchPage_DeclaredCharsetCorruptedBytesIsError(t *testing.T) {
	// EUC-KR lead byte 0xA1 with no valid trail byte: an incomplete sequence
	// the decoder silently replaces with U+FFFD (probe-confirmed).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=euc-kr")
		_, _ = w.Write([]byte("<html><body><p>Price \xa1</p></body></html>"))
	}))
	defer srv.Close()
	if _, err := FetchPage(t.Context(), srv.URL, srv.Client()); err == nil {
		t.Fatal("corrupted euc-kr body must be a fetch error, not a silent U+FFFD substitution returned as page text")
	}
}

// F5b, the latin-1 shape: a latin-1-declared page carrying a byte that is
// invalid in that charset must be a fetch error. iso-8859-1 resolves to
// windows-1252 in the label table (probe-confirmed), where 0x81 is undefined
// and decodes to U+FFFD by substitution — passthrough is mojibake with exit 0.
func TestFetchPage_Latin1InvalidByteIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=iso-8859-1")
		_, _ = w.Write([]byte("cost \x81 total"))
	}))
	defer srv.Close()
	if _, err := FetchPage(t.Context(), srv.URL, srv.Client()); err == nil {
		t.Fatal("latin-1-declared body with an undefined byte must be a fetch error, not a U+FFFD-substituted passthrough")
	}
}

// F5b control: the scan rejects SUBSTITUTED output, not every non-UTF-8
// decode — a cleanly-encoded euc-kr page (B0 A1 = 가) still passes.
func TestFetchPage_ValidEUCKRPasses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=euc-kr")
		_, _ = w.Write([]byte("<html><body><p>\xb0\xa1</p></body></html>"))
	}))
	defer srv.Close()
	page, err := FetchPage(t.Context(), srv.URL, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(page, "가") {
		t.Errorf("valid euc-kr page = %q, want the decoded text %q", page, "가")
	}
}

// F5b qualifier: a page whose SOURCE legitimately encodes U+FFFD in its
// declared charset is not a substitution. GB18030 can encode U+FFFD
// (4-byte form, probe-confirmed), so a page carrying that exact sequence
// passes — the scan must not reject a replacement character the document
// really contained.
func TestFetchPage_LegitimateReplacementCharPasses(t *testing.T) {
	enc, _ := charset.Lookup("gb18030")
	if enc == nil {
		t.Fatal("gb18030 must resolve for this test's premise")
	}
	src, err := enc.NewEncoder().String("�")
	if err != nil {
		t.Fatalf("gb18030 must encode U+FFFD (probe-confirmed): %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=gb18030")
		_, _ = w.Write([]byte("x" + src + "y"))
	}))
	defer srv.Close()
	page, err := FetchPage(t.Context(), srv.URL, srv.Client())
	if err != nil {
		t.Fatalf("legitimately-encoded U+FFFD must pass, got %v", err)
	}
	if !strings.ContainsRune(page, '�') {
		t.Errorf("page = %q, want the source's legitimate U+FFFD preserved", page)
	}
}

// F5c (t1968 sync-audit iter3): the iter2 scan tested only the FIRST decoded
// U+FFFD and accepted whenever the encoded form appeared ANYWHERE in the body,
// so a source that legitimately encoded U+FFFD early and was corrupted later
// passed silently (auditor probe: gb18030 legit form + orphan lead byte →
// err nil, page with TWO U+FFFD). Count-based provenance: one
// legitimately-encoded U+FFFD in the source accounts for exactly one decoded
// U+FFFD; any decoded count beyond the source's legitimate count is a
// substitution and must be a fetch error.
func TestFetchPage_LegitFFFDPlusLaterCorruptionIsError(t *testing.T) {
	enc, _ := charset.Lookup("gb18030")
	if enc == nil {
		t.Fatal("gb18030 must resolve for this test's premise")
	}
	src, err := enc.NewEncoder().String("�")
	if err != nil {
		t.Fatalf("gb18030 must encode U+FFFD (probe-confirmed): %v", err)
	}
	// "x" + gb18030(U+FFFD) + "y" is legitimate; the trailing 0x81 is an
	// orphan lead byte the decoder silently substitutes (probe-confirmed).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=gb18030")
		_, _ = w.Write([]byte("x" + src + "y\x81"))
	}))
	defer srv.Close()
	if _, err := FetchPage(t.Context(), srv.URL, srv.Client()); err == nil {
		t.Fatal("legitimate U+FFFD plus a later corrupted byte must be a fetch error, not a silent extra U+FFFD substitution")
	}
}

// F5b adjacent residual (named in the iter2 audit): a UTF-8 alias label
// outside isUTF8Family (unicode-1-1-utf-8) routes through charset.Lookup's
// decoder, which substitutes invalid bytes silently — the same substitution
// scan must catch it there too.
func TestFetchPage_ExoticUTF8AliasInvalidBytesIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=unicode-1-1-utf-8")
		_, _ = w.Write([]byte("plain \xff text"))
	}))
	defer srv.Close()
	if _, err := FetchPage(t.Context(), srv.URL, srv.Client()); err == nil {
		t.Fatal("invalid bytes under an exotic UTF-8 alias label must be a fetch error, not substituted passthrough")
	}
}
