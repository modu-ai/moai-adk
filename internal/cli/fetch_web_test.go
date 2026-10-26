// fetch_web_test.go — SPEC-FETCH-WEB-001 M3. Command-level tests for
// `moai fetch web`: the output contract and exit-code mapping per terminal
// condition (AC-FW-004), the command surface (AC-FW-005), the off-gate
// zero-network proof (AC-FW-007), the ONE end-to-end test against a fake
// page server and a fake Jev server (AC-FW-008), and the no-live-endpoint
// isolation guard (AC-FW-014). No test contacts a real endpoint.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/fetchweb"
	"github.com/modu-ai/moai-adk/internal/jev"
)

// roundTripFunc adapts a function to an http.RoundTripper for stub page
// transports (no server, no socket).
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// stubAsker scripts probabilities by question id (fetchweb's b<i>/c<i>
// convention) without any server.
type stubAsker struct {
	script map[string]float64
}

func (s stubAsker) Ask(_ context.Context, req jev.Request) jev.Result {
	answers := make([]jev.Answer, 0, len(req.Questions))
	for _, q := range req.Questions {
		answers = append(answers, jev.Answer{
			QuestionID: q.ID, Kind: jev.KindNoul,
			Probability: s.script[q.ID], Model: jev.VerifiedInterpretedModel,
		})
	}
	return jev.Result{Availability: jev.Available, Answers: answers}
}

// tripwireTransport fails the test on ANY request: the off-gate proof's
// instrumented transport.
type tripwireTransport struct {
	t     *testing.T
	calls int
}

func (tt *tripwireTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tt.calls++
	tt.t.Fatalf("network request escaped the off-gate: %s %s", r.Method, r.URL)
	return nil, fmt.Errorf("unreachable")
}

// tripwireAsker fails the test on any Jev call.
type tripwireAsker struct {
	t     *testing.T
	calls int
}

func (a *tripwireAsker) Ask(_ context.Context, req jev.Request) jev.Result {
	a.calls++
	a.t.Fatalf("Jev call escaped the off-gate: state=%q", req.State)
	return jev.Result{Availability: jev.Disabled}
}

// pageFixture builds an HTML page with n paragraph blocks of ~60 approximate
// tokens each under one heading.
func pageFixture(n int) string {
	var b strings.Builder
	b.WriteString("<html><body><h1>Doc</h1>")
	for i := 0; i < n; i++ {
		text := fmt.Sprintf("paragraph %d content words here", i)
		text += strings.Repeat("x", 240-len(text))
		fmt.Fprintf(&b, "<p>%s</p>", text)
	}
	b.WriteString("</body></html>")
	return b.String()
}

// runFetchWebCmd executes the command against injected seams and returns
// stdout, stderr, and the resolved exit code.
func runFetchWebCmd(t *testing.T, root string, pageClient *http.Client, asker fetchweb.Answerer, args ...string) (string, string, int) {
	t.Helper()
	origRoot, origPage, origAsker := findProjectRootFn, fetchWebPageClientFn, fetchWebAskerFn
	findProjectRootFn = func() (string, error) { return root, nil }
	fetchWebPageClientFn = func() *http.Client { return pageClient }
	fetchWebAskerFn = func() fetchweb.Answerer { return asker }
	t.Cleanup(func() {
		findProjectRootFn, fetchWebPageClientFn, fetchWebAskerFn = origRoot, origPage, origAsker
	})

	cmd := newFetchCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(append([]string{"web"}, args...))
	err := cmd.Execute()
	code := 0
	if err != nil {
		if c, ok := ResolveExitCode(err); ok {
			code = c
		} else {
			code = 1
		}
	}
	return out.String(), errBuf.String(), code
}

// textResponse builds a canned HTTP response for a stub transport.
func textResponse(r *http.Request, body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Request:    r,
	}, nil
}

// AC-FW-005: the command exists with the exact name `fetch web`, takes
// <url> and "<question>", and documents --json.
func TestFetchWebCommand_Surface(t *testing.T) {
	cmd := newFetchCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"web", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	help := out.String()
	if !strings.Contains(help, "web <url>") {
		t.Errorf("help must document the web <url> form:\n%s", help)
	}
	if !strings.Contains(help, "--json") {
		t.Errorf("help must document --json:\n%s", help)
	}
	if cmd.Name() != "fetch" {
		t.Errorf("parent command name = %q, want fetch", cmd.Name())
	}
	for _, c := range cmd.Commands() {
		if c.Name() == "web" {
			return
		}
	}
	t.Errorf("no `web` subcommand under fetch")
}

// AC-FW-007 (the card's named zero-network proof): with
// workflow.jev.enabled=false, the command makes ZERO network requests — no
// page fetch, no Jev call — emits the WebFetch fallback notice, exits 2.
func TestFetchWeb_OffGateZeroNetwork(t *testing.T) {
	gateOff := jevProject(t, false)
	trip := &tripwireTransport{t: t}
	client := &http.Client{Transport: trip}
	askerTrips := &tripwireAsker{t: t}
	out, errBuf, code := runFetchWebCmd(t, gateOff, client, askerTrips, "https://example.com/page", "what is this?")
	if code != 2 {
		t.Errorf("exit code = %d, want 2 while gated off", code)
	}
	if !strings.Contains(errBuf, "WebFetch") {
		t.Errorf("stderr must name the WebFetch fallback, got %q", errBuf)
	}
	if trip.calls != 0 {
		t.Errorf("%d network requests escaped the off-gate", trip.calls)
	}
	if askerTrips.calls != 0 {
		t.Errorf("%d Jev calls escaped the off-gate", askerTrips.calls)
	}
	if out != "" {
		t.Errorf("gated-off run must print no result payload, got %q", out)
	}
}

// AC-FW-004: statuses map to exit codes 0/0/1/1/2, the first output line
// begins `#`, blocks carry `##` heading-path lines and verbatim text, and
// --json emits the same information.
func TestFetchWeb_StatusExitContract(t *testing.T) {
	gateOn := jevProject(t, true)
	pageClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return textResponse(r, pageFixture(40)) // ~2,400 tokens: scored path
	})}

	t.Run("filtered exits 0", func(t *testing.T) {
		script := map[string]float64{}
		for i := 0; i < 40; i++ {
			script[fmt.Sprintf("b%d", i)] = 0.05
		}
		script["b7"] = 0.95
		script["c7"] = 0.9
		out, _, code := runFetchWebCmd(t, gateOn, pageClient, stubAsker{script}, "https://x.test/p", "q?")
		if code != 0 {
			t.Fatalf("exit = %d, want 0", code)
		}
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if !strings.HasPrefix(lines[0], "#") || !strings.Contains(lines[0], "filtered") {
			t.Errorf("first line = %q", lines[0])
		}
		if !strings.Contains(out, "## Doc") {
			t.Errorf("blocks must carry heading-path lines:\n%s", out)
		}
	})

	t.Run("no_match branch a: zero candidates exits 1 without first-block keep", func(t *testing.T) {
		script := map[string]float64{}
		for i := 0; i < 40; i++ {
			script[fmt.Sprintf("b%d", i)] = 0.3
		}
		out, _, code := runFetchWebCmd(t, gateOn, pageClient, stubAsker{script}, "https://x.test/p", "q?")
		if code != 1 {
			t.Fatalf("exit = %d, want 1 (no_match)", code)
		}
		if !strings.HasPrefix(out, "# no_match") {
			t.Errorf("status line = %q", strings.Split(out, "\n")[0])
		}
		if strings.Contains(out, "## ") {
			t.Errorf("no_match must emit no blocks (first-block keep does not apply):\n%s", out)
		}
	})

	t.Run("no_match branch b: candidate below backfill floor still filtered", func(t *testing.T) {
		// b0 candidate kept at stage 2; kept set ~60 tokens < 300 -> backfill
		// adds more; result is filtered with the first block present.
		script := map[string]float64{}
		for i := 0; i < 40; i++ {
			script[fmt.Sprintf("b%d", i)] = 0.05
		}
		script["b0"] = 0.9
		script["c0"] = 0.9
		out, _, code := runFetchWebCmd(t, gateOn, pageClient, stubAsker{script}, "https://x.test/p", "q?")
		if code != 0 {
			t.Fatalf("exit = %d, want 0 (branch b: candidate exists -> filtered)", code)
		}
		if !strings.Contains(out, "paragraph 0 ") {
			t.Errorf("first block must be present in branch b output:\n%.200s", out)
		}
	})

	t.Run("passthrough exits 0 with whole document", func(t *testing.T) {
		small := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return textResponse(r, pageFixture(3)) // ~180 tokens
		})}
		trip := &tripwireAsker{t: t}
		out, _, code := runFetchWebCmd(t, gateOn, small, trip, "https://x.test/p", "q?")
		if code != 0 {
			t.Fatalf("exit = %d, want 0 (passthrough)", code)
		}
		if !strings.HasPrefix(out, "# passthrough") {
			t.Errorf("status line = %q", strings.Split(out, "\n")[0])
		}
		if strings.Count(out, "## Doc") != 3 {
			t.Errorf("passthrough must keep every block, got %d", strings.Count(out, "## Doc"))
		}
	})

	t.Run("empty exits 1", func(t *testing.T) {
		empty := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return textResponse(r, "<html><body><nav>menu</nav></body></html>")
		})}
		out, _, code := runFetchWebCmd(t, gateOn, empty, stubAsker{}, "https://x.test/p", "q?")
		if code != 1 {
			t.Fatalf("exit = %d, want 1 (empty)", code)
		}
		if !strings.HasPrefix(out, "# empty") {
			t.Errorf("status line = %q", strings.Split(out, "\n")[0])
		}
	})

	t.Run("fetch_error exits 2", func(t *testing.T) {
		broken := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("connection refused")
		})}
		out, errBuf, code := runFetchWebCmd(t, gateOn, broken, stubAsker{}, "https://x.test/p", "q?")
		if code != 2 {
			t.Fatalf("exit = %d, want 2 (fetch_error)", code)
		}
		if !strings.HasPrefix(out, "# fetch_error") {
			t.Errorf("status line = %q", strings.Split(out, "\n")[0])
		}
		if errBuf == "" {
			t.Error("fetch_error should surface the condition on stderr")
		}
	})

	t.Run("json emits the same information", func(t *testing.T) {
		script := map[string]float64{}
		for i := 0; i < 40; i++ {
			script[fmt.Sprintf("b%d", i)] = 0.05
		}
		script["b7"] = 0.95
		script["c7"] = 0.9
		out, _, code := runFetchWebCmd(t, gateOn, pageClient, stubAsker{script}, "https://x.test/p", "q?", "--json")
		if code != 0 {
			t.Fatalf("exit = %d, want 0", code)
		}
		var doc map[string]any
		if err := json.Unmarshal([]byte(out), &doc); err != nil {
			t.Fatalf("--json output not parseable: %v\n%s", err, out)
		}
		if doc["status"] != "filtered" || doc["exit_code"] != float64(0) {
			t.Errorf("json status/exit_code = %v/%v", doc["status"], doc["exit_code"])
		}
	})
}

// AC-FW-008 — the ONE end-to-end test: a fake page server and a fake Jev
// server, both bound, one full command invocation, no real endpoint
// contacted.
func TestFetchWeb_E2EFakeServers(t *testing.T) {
	gateOn := jevProject(t, true)

	page := pageFixture(40)
	pageSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(page))
	}))
	defer pageSrv.Close()

	// Fake Jev server: answers noul probabilities from the request's own
	// question ids — block 7 relevant at both stages.
	jevSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			State     string `json:"state"`
			Questions map[string]struct {
				Instructions string `json:"instructions"`
			} `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("fake jev: bad request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		answers := map[string]map[string]any{}
		for id := range req.Questions {
			p := 0.05
			switch id {
			case "b7":
				p = 0.95
			case "c7":
				p = 0.9
			}
			answers[id] = map[string]any{"type": "noul", "noul": p}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model":   jev.VerifiedInterpretedModel,
			"answers": answers,
			"usage":   map[string]any{"input_tokens": 10},
		})
	}))
	defer jevSrv.Close()

	asker := jev.New(true)
	asker.Endpoint = jevSrv.URL
	asker.LoadCredential = func() string { return "test-key" }

	out, _, code := runFetchWebCmd(t, gateOn, pageSrv.Client(), asker, pageSrv.URL, "which paragraph seven?")
	if code != 0 {
		t.Fatalf("exit = %d, want 0 (filtered)", code)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if !strings.HasPrefix(lines[0], "#") || !strings.Contains(lines[0], "filtered") {
		t.Errorf("first line = %q", lines[0])
	}
	if !strings.Contains(out, "paragraph 7 ") {
		t.Errorf("the scored block's text must appear verbatim:\n%.400s", out)
	}
	if !strings.Contains(out, "## Doc") {
		t.Errorf("heading-path lines missing:\n%.200s", out)
	}
}

// AC-FW-014 isolation guard: no test in this file or the fetchweb package
// references the production Jev endpoint as a dial target.
func TestFetchWeb_NoLiveEndpointInTests(t *testing.T) {
	files := []string{
		"fetch_web_test.go",
		filepath.Join("..", "fetchweb", "blocks_test.go"),
		filepath.Join("..", "fetchweb", "scoring_test.go"),
		filepath.Join("..", "fetchweb", "fetch_test.go"),
		filepath.Join("..", "fetchweb", "output_test.go"),
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "api.typesafe.ai") && !strings.Contains(line, "EndpointURL") {
				t.Errorf("%s: test references the production endpoint: %s", f, strings.TrimSpace(line))
			}
		}
	}
}
