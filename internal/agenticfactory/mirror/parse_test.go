package mirror

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// real404 reproduces what `gh api -i` printed for a 404 on 2026-10-29: the
// status line ends with a bare newline, the header lines with CRLF, and a blank
// CRLF line precedes the body. Only headers that matter here are kept.
const real404 = "HTTP/2.0 404 Not Found\n" +
	"Access-Control-Allow-Origin: *\r\n" +
	"Content-Type: application/json; charset=utf-8\r\n" +
	"X-Ratelimit-Limit: 5000\r\n" +
	"X-Ratelimit-Remaining: 4705\r\n" +
	"X-Ratelimit-Reset: 1793258419\r\n" +
	"X-Xss-Protection: 0\r\n" +
	"\r\n" +
	`{"message":"Not Found","status":"404"}`

func TestParseTheOutputGhReallyPrinted(t *testing.T) {
	status, h, body, ok := parseHTTP([]byte(real404))
	if !ok || status != 404 {
		t.Fatalf("status = %d ok=%v", status, ok)
	}
	if h["x-ratelimit-remaining"] != "4705" || h["x-ratelimit-reset"] != "1793258419" || h["content-type"] != "application/json; charset=utf-8" {
		t.Errorf("headers = %v", h)
	}
	if string(body) != `{"message":"Not Found","status":"404"}` {
		t.Errorf("body = %q", body)
	}
}

func TestBodyWithBlankLinesIsKeptWhole(t *testing.T) {
	pretty := "{\n  \"a\": 1,\n\n  \"b\": 2\n}"
	for _, sep := range []string{"\r\n\r\n", "\n\n"} {
		out := "HTTP/2.0 200 OK" + sep[:len(sep)/2] + "X-A: 1" + sep + pretty
		status, h, body, ok := parseHTTP([]byte(out))
		if !ok || status != 200 || h["x-a"] != "1" || string(body) != pretty {
			t.Errorf("sep %q: %d %v %q", sep, status, h, body)
		}
	}
}

func TestOutputWithoutAStatusLineIsAllBody(t *testing.T) {
	for _, in := range []string{"", `{"id":1}`, "HTTP/2.0 abc\r\n\r\nx", "HTTP/2.0\r\n\r\nx"} {
		if status, _, body, ok := parseHTTP([]byte(in)); ok || status != 0 || string(body) != in {
			t.Errorf("%q parsed as status=%d ok=%v body=%q", in, status, ok, body)
		}
	}
}

// The observed 404 reaches the caller as a typed status error, and an
// observed header set with an exhausted limit becomes a block with a reset time.
func TestRealShapesFlowThroughTheClassifier(t *testing.T) {
	f := &fakeGH{results: []result{{out: real404, errOut: "gh: Not Found (HTTP 404)", exit: 1}}}
	err := timed(f).WriteLedger(context.Background(), ledger())
	var se *forge.StatusError
	var be *BlockedError
	if !errors.As(err, &se) || se.Status != 404 || errors.As(err, &be) {
		t.Fatalf("a real 404 must be a plain typed status error: %v", err)
	}

	reset := t0.Add(9 * time.Minute)
	limited := fmt.Sprintf("HTTP/2.0 403 Forbidden\nX-Ratelimit-Remaining: 0\r\nX-Ratelimit-Reset: %d\r\n\r\n{\"message\":\"API rate limit exceeded\"}", reset.Unix())
	f = &fakeGH{results: []result{{out: limited, errOut: "gh: API rate limit exceeded (HTTP 403)", exit: 1}}}
	be = blocked(t, timed(f).WriteLedger(context.Background(), ledger()))
	if be.Reason != core.ReasonRateLimited || !be.RetryAt.Equal(reset) {
		t.Errorf("exhausted limit: %+v", be)
	}
}

func TestParsedHeadersNeverContainTheBodyText(t *testing.T) {
	_, h, _, _ := parseHTTP([]byte(real404))
	for k, v := range h {
		if strings.Contains(k, "message") || strings.Contains(v, "Not Found") {
			t.Errorf("body leaked into the headers: %q: %q", k, v)
		}
	}
}
