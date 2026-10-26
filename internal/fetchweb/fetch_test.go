package fetchweb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
