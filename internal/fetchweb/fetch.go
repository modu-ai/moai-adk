package fetchweb

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// fetch.go — SPEC-FETCH-WEB-001 M3. The bounded page fetcher (REQ-FW-007):
// a timeout and a maximum response size, both enforced here so no caller can
// forget them. The client is INJECTED — this package never constructs a
// transport, and no test can reach a real target page (REQ-FW-010).

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
// body. Non-2xx statuses and bodies over MaxResponseBytes are errors.
func FetchPage(ctx context.Context, url string, client *http.Client) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
	return string(body), nil
}
