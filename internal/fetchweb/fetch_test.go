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
