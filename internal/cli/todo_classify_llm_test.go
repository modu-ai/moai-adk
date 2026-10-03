// todo_classify_llm_test.go — SPEC-TCD-LLM-DECIDER-001 M2: the LLM-backed
// decider's transport tests (AC-TLD-001 happy path, AC-TLD-003 failure
// taxonomy, AC-TLD-004 identity overwrite, AC-TLD-006 transport discipline).
// Every scenario runs against an httptest server through the endpoint seam —
// zero real network in the suite (plan §D.5). The seams are package vars, so
// these tests stay NON-parallel.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// todoLLMFakeKey is the credential every seam-swapped test flows. The secret
// test asserts this marker never reaches process output.
const todoLLMFakeKey = "moai-fake-glm-key-DO-NOT-PRINT-7f3a"

// withTodoLLMSeams swaps the decider's endpoint, HTTP client, and key-loader
// seams for the test's duration (restored on cleanup).
func withTodoLLMSeams(t *testing.T, endpoint string, client todoLLMHTTPDoer, keyLoader func() string) {
	t.Helper()
	oldEndpoint, oldClient, oldKey := todoLLMEndpoint, todoLLMHTTPClient, todoLLMKeyLoader
	todoLLMEndpoint = func() string { return endpoint }
	todoLLMHTTPClient = client
	todoLLMKeyLoader = keyLoader
	t.Cleanup(func() {
		todoLLMEndpoint, todoLLMHTTPClient, todoLLMKeyLoader = oldEndpoint, oldClient, oldKey
	})
}

// fakeKeyLoader returns a credential seam serving the fake test key.
func fakeKeyLoader() func() string {
	return func() string { return todoLLMFakeKey }
}

// llmEndpointRecorder records every request an httptest endpoint received —
// the positive-control surface for the transport-discipline assertions.
type llmEndpointRecorder struct {
	mu       sync.Mutex
	requests []recordedLLMRequest
}

type recordedLLMRequest struct {
	header    http.Header
	body      []byte
	arrivedAt time.Time
}

func (r *llmEndpointRecorder) add(h http.Header, body []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, recordedLLMRequest{header: h, body: body, arrivedAt: time.Now()})
}

func (r *llmEndpointRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.requests)
}

// llmServer starts an httptest endpoint that records requests and answers
// with status + body (empty body means a bare status response).
func llmServer(t *testing.T, status int, body string) (*httptest.Server, *llmEndpointRecorder) {
	t.Helper()
	rec := &llmEndpointRecorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b, _ := io.ReadAll(req.Body)
		rec.add(req.Header.Clone(), b)
		w.WriteHeader(status)
		if body != "" {
			_, _ = io.WriteString(w, body)
		}
	}))
	t.Cleanup(server.Close)
	return server, rec
}

// llmEnvelope wraps a judgment JSON in the Anthropic-compatible response
// envelope whose single text block carries it.
func llmEnvelope(judgment string) string {
	inner := map[string]any{
		"model": "glm-test",
		"content": []map[string]any{
			{"type": "text", "text": judgment},
		},
	}
	data, err := json.Marshal(inner)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// llmJudgment formats one model judgment as the JSON text block payload.
func llmJudgment(priority, mode string, blocked bool, decider, reason string) string {
	return fmt.Sprintf(`{"priority":%q,"blocked":%t,"mode":%q,"decider":%q,"reason":%q}`,
		priority, blocked, mode, decider, reason)
}

// wantHappyJudgment is the closed-set judgment the happy-path fixtures
// return.
func wantHappyJudgment() kanban.CardClassification {
	return kanban.CardClassification{
		Priority: kanban.ClassPriorityHigh,
		Blocked:  false,
		Mode:     kanban.ClassModeParallelizable,
		Decider:  kanban.DeciderIdentityLLM,
		Reason:   "spans several systems",
	}
}

// TestLLMDeciderClassify_HappyPathRecordsForcedLLMIdentity — AC-TLD-001's
// decider-level half: a valid closed-set response yields exactly the
// returned judgment, with the recorded identity forced to llm REGARDLESS of
// what the response claims (REQ-TLD-004 — the claim here is "human").
func TestLLMDeciderClassify_HappyPathRecordsForcedLLMIdentity(t *testing.T) {
	server, rec := llmServer(t, http.StatusOK, llmEnvelope(llmJudgment("high", "parallelizable", false, "human", "spans several systems")))
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	cls, err := newLLMCardDecider().Classify("refactor the parser")
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if rec.count() != 1 {
		t.Fatalf("endpoint received %d requests, want 1", rec.count())
	}
	want := wantHappyJudgment()
	if cls.Priority != want.Priority || cls.Blocked != want.Blocked || cls.Mode != want.Mode ||
		cls.Decider != kanban.DeciderIdentityLLM || cls.Reason != want.Reason {
		t.Errorf("classification = %+v, want the model's judgment with identity forced to llm", cls)
	}
	if cls.ClassifiedAt != "" {
		t.Errorf("decider stamped classified_at %q; the stamp belongs to the write, not the judgment", cls.ClassifiedAt)
	}
}

// TestLLMDeciderClassify_ClaimedJevIdentityIsAcceptedNotFailed — AC-TLD-004:
// a response claiming the refused jev identity is NOT a failure — the
// identity is overwritten to llm and the judgment accepted (OD-C: claimed
// identity sits outside the failure taxonomy).
func TestLLMDeciderClassify_ClaimedJevIdentityIsAcceptedNotFailed(t *testing.T) {
	server, _ := llmServer(t, http.StatusOK, llmEnvelope(llmJudgment("normal", "serial", false, "jev", "routine cleanup")))
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	cls, err := newLLMCardDecider().Classify("clean tmp files")
	if err != nil {
		t.Fatalf("classify errored on a claimed identity, want acceptance with overwrite: %v", err)
	}
	if cls.Decider != kanban.DeciderIdentityLLM {
		t.Errorf("recorded decider = %q, want the forced %q", cls.Decider, kanban.DeciderIdentityLLM)
	}
}

// TestLLMDeciderClassify_OutOfSetPriorityFails — AC-TLD-003(v): a priority
// outside the closed set is an error return (the seam's fallback branch
// owns the degradation), never an out-of-set record.
func TestLLMDeciderClassify_OutOfSetPriorityFails(t *testing.T) {
	server, _ := llmServer(t, http.StatusOK, llmEnvelope(llmJudgment("urgent", "serial", false, "llm", "seems urgent")))
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	_, err := newLLMCardDecider().Classify("on fire")
	if err == nil {
		t.Fatal("out-of-set priority accepted, want an error return")
	}
	if !strings.Contains(err.Error(), "urgent") {
		t.Errorf("error %q does not name the offending value", err)
	}
}

// TestLLMDeciderClassify_NonJSONTextFails — AC-TLD-003(iii): a non-JSON body
// is an error return.
func TestLLMDeciderClassify_NonJSONTextFails(t *testing.T) {
	server, _ := llmServer(t, http.StatusOK, llmEnvelope("I cannot classify that"))
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	if _, err := newLLMCardDecider().Classify("x"); err == nil {
		t.Fatal("non-JSON text accepted, want an error return")
	}
}

// TestLLMDeciderClassify_MalformedEnvelopeFails — AC-TLD-003(iii) arm two:
// a body that is not the response envelope at all is an error return.
func TestLLMDeciderClassify_MalformedEnvelopeFails(t *testing.T) {
	server, _ := llmServer(t, http.StatusOK, "<html>gateway error</html>")
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	if _, err := newLLMCardDecider().Classify("x"); err == nil {
		t.Fatal("malformed envelope accepted, want an error return")
	}
}

// TestLLMDeciderClassify_HTTPServerErrorFails — AC-TLD-003(ii): a non-2xx
// response is an error return.
func TestLLMDeciderClassify_HTTPServerErrorFails(t *testing.T) {
	server, _ := llmServer(t, http.StatusInternalServerError, "boom")
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	if _, err := newLLMCardDecider().Classify("x"); err == nil {
		t.Fatal("HTTP 500 accepted, want an error return")
	}
}

// TestLLMDeciderClassify_ConnectionRefusedFails — AC-TLD-003(i): an
// unreachable endpoint is an error return.
func TestLLMDeciderClassify_ConnectionRefusedFails(t *testing.T) {
	server, _ := llmServer(t, http.StatusOK, llmEnvelope(llmJudgment("high", "serial", false, "llm", "unused")))
	url := server.URL
	server.Close() // the port is closed; every connection to it is refused
	withTodoLLMSeams(t, url+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	if _, err := newLLMCardDecider().Classify("x"); err == nil {
		t.Fatal("connection refused accepted, want an error return")
	}
}

// TestLLMDeciderClassify_MissingCredentialFailsWithoutRequest — REQ-TLD-006:
// absent credentials fail BEFORE any request, so no unauthenticated call
// ever leaves the process.
func TestLLMDeciderClassify_MissingCredentialFailsWithoutRequest(t *testing.T) {
	server, rec := llmServer(t, http.StatusOK, llmEnvelope(llmJudgment("high", "serial", false, "llm", "unused")))
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, func() string { return "" })

	if _, err := newLLMCardDecider().Classify("x"); err == nil {
		t.Fatal("missing credential accepted, want an error return")
	}
	if n := rec.count(); n != 0 {
		t.Errorf("endpoint received %d requests without credentials, want 0", n)
	}
}

// TestLLMDeciderClassify_SlowEndpointTimesOutFails — AC-TLD-003(iv): a
// response stalling past the configured timeout is an error return (the
// test shortens the seam client's ceiling; the production ceiling itself is
// asserted against the config constant below).
func TestLLMDeciderClassify_SlowEndpointTimesOutFails(t *testing.T) {
	stall := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // far past the 100ms test ceiling; the client aborts first
	}))
	t.Cleanup(stall.Close)
	withTodoLLMSeams(t, stall.URL+glmMessagesPath, &http.Client{Timeout: 100 * time.Millisecond}, fakeKeyLoader())

	if _, err := newLLMCardDecider().Classify("x"); err == nil {
		t.Fatal("timeout accepted, want an error return")
	}
}

// TestLLMDeciderClassify_ReasonIsBoundedToOneLine — plan §E: the recorded
// reason is the model's one-line rationale; an embedded newline never
// reaches the record.
func TestLLMDeciderClassify_ReasonIsBoundedToOneLine(t *testing.T) {
	server, _ := llmServer(t, http.StatusOK, llmEnvelope(llmJudgment("low", "serial", false, "llm", "cosmetic\nand deferrable\nreally")))
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	cls, err := newLLMCardDecider().Classify("x")
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if strings.ContainsAny(cls.Reason, "\n\r") {
		t.Errorf("recorded reason %q carries a line break, want a single line", cls.Reason)
	}
	if cls.Reason != "cosmetic and deferrable really" {
		t.Errorf("recorded reason = %q, want the collapsed single line", cls.Reason)
	}
}

// TestLLMDeciderClassify_RequestCarriesTaskSlotTransport — AC-TLD-006's
// request-side discipline: the request names the TASK-slot model (never the
// audit pin slot), the configured token cap, no reasoning_effort field, and
// authenticates through the credential header.
func TestLLMDeciderClassify_RequestCarriesTaskSlotTransport(t *testing.T) {
	server, rec := llmServer(t, http.StatusOK, llmEnvelope(llmJudgment("high", "serial", false, "llm", "unused")))
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	if _, err := newLLMCardDecider().Classify("judge me"); err != nil {
		t.Fatalf("classify: %v", err)
	}
	if rec.count() != 1 {
		t.Fatalf("endpoint received %d requests, want 1", rec.count())
	}
	req := rec.requests[0]
	if got := req.header.Get("x-api-key"); got != todoLLMFakeKey {
		t.Errorf("credential header carried %q, want the flowed key (positive control: the endpoint DID receive the credential)", got)
	}
	var body struct {
		Model           string         `json:"model"`
		ReasoningEffort string         `json:"reasoning_effort"`
		MaxTokens       int            `json:"max_tokens"`
		System          string         `json:"system"`
		Messages        []struct{ Text string `json:"content"` } `json:"messages"`
	}
	if err := json.Unmarshal(req.body, &body); err != nil {
		t.Fatalf("request body %s: %v", req.body, err)
	}
	if body.Model != glmTaskDefaultModel {
		t.Errorf("request model = %q, want the task-slot %q (never the audit pin slot)", body.Model, glmTaskDefaultModel)
	}
	if body.MaxTokens != config.DefaultTodoClassifyLLMMaxTokens {
		t.Errorf("request max_tokens = %d, want the configured cap %d", body.MaxTokens, config.DefaultTodoClassifyLLMMaxTokens)
	}
	if body.ReasoningEffort != "" {
		t.Errorf("request carried reasoning_effort %q, want the field omitted", body.ReasoningEffort)
	}
	if body.System == "" || len(body.Messages) != 1 || body.Messages[0].Text != "judge me" {
		t.Errorf("request prompt shape wrong: system=%d bytes, messages=%+v", len(body.System), body.Messages)
	}
}

// TestLLMDeciderHTTPClientCarriesConfiguredTimeout — AC-TLD-006: the
// production client's timeout equals the internal/config/defaults.go
// constant (asserted through the seam).
func TestLLMDeciderHTTPClientCarriesConfiguredTimeout(t *testing.T) {
	client, ok := todoLLMHTTPClient.(*http.Client)
	if !ok {
		t.Fatalf("production HTTP seam is %T, want *http.Client", todoLLMHTTPClient)
	}
	if client.Timeout != config.DefaultTodoClassifyLLMTimeout {
		t.Errorf("production client timeout = %v, want the config default %v", client.Timeout, config.DefaultTodoClassifyLLMTimeout)
	}
}

// TestTodoAdd_LLMSelectionRecordsModelJudgment — AC-TLD-001 end to end:
// with the llm selection and a fake endpoint, the add exits 0 printing
// "<id> <position>" and the queue record carries exactly the returned
// judgment with decider llm, a classified_at stamp, and the reason.
func TestTodoAdd_LLMSelectionRecordsModelJudgment(t *testing.T) {
	_, store := todoFixture(t)
	t.Setenv(config.EnvTodoDecider, kanban.DeciderIdentityLLM)
	server, rec := llmServer(t, http.StatusOK, llmEnvelope(llmJudgment("high", "parallelizable", false, "human", "spans several systems")))
	withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

	out, errOut, err := runTodo(t, "add", "llm judged card")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if n := strings.Count(errOut, "classification decider unavailable"); n != 0 {
		t.Errorf("happy-path add printed %d fallback notices, want 0 (stderr: %q)", n, errOut)
	}
	if rec.count() != 1 {
		t.Errorf("endpoint received %d requests, want 1", rec.count())
	}
	id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
	rec2, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec2.Items {
		if it.ID != id {
			continue
		}
		c := it.Classification
		want := wantHappyJudgment()
		if c == nil || c.Priority != want.Priority || c.Blocked != want.Blocked || c.Mode != want.Mode ||
			c.Decider != kanban.DeciderIdentityLLM {
			t.Errorf("card %s classification = %+v, want the model judgment with decider llm", id, c)
		}
		if c != nil && c.ClassifiedAt == "" {
			t.Errorf("card %s classification carries no classified_at stamp", id)
		}
	}
}

// TestTodoAdd_LLMSecretNeverPrinted — AC-TLD-006's secret discipline: on the
// happy path AND a forced-failure path the fake credential appears in ZERO
// bytes of process stdout/stderr — with the positive control that the fake
// endpoint DID receive it (so the 0-hit is a flowed-and-unprinted proof,
// not an unflowed one).
func TestTodoAdd_LLMSecretNeverPrinted(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		_, _ = todoFixture(t)
		t.Setenv(config.EnvTodoDecider, kanban.DeciderIdentityLLM)
		server, rec := llmServer(t, http.StatusOK, llmEnvelope(llmJudgment("high", "serial", false, "llm", "unused")))
		withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

		out, errOut, err := runTodo(t, "add", "secret check card")
		if err != nil {
			t.Fatalf("add: %v", err)
		}
		if rec.count() == 0 || !strings.Contains(rec.requests[0].header.Get("x-api-key"), todoLLMFakeKey) {
			t.Fatal("positive control failed: the endpoint did NOT receive the credential, so the 0-hit proves nothing")
		}
		if strings.Contains(out, todoLLMFakeKey) || strings.Contains(errOut, todoLLMFakeKey) {
			t.Error("the fake credential leaked into process output")
		}
	})
	t.Run("failure path", func(t *testing.T) {
		_, _ = todoFixture(t)
		t.Setenv(config.EnvTodoDecider, kanban.DeciderIdentityLLM)
		server, rec := llmServer(t, http.StatusInternalServerError, "boom")
		withTodoLLMSeams(t, server.URL+glmMessagesPath, todoLLMHTTPClient, fakeKeyLoader())

		out, errOut, err := runTodo(t, "add", "secret check failure card")
		if err != nil {
			t.Fatalf("add: %v (an LLM failure must never refuse the add)", err)
		}
		if rec.count() == 0 || !strings.Contains(rec.requests[0].header.Get("x-api-key"), todoLLMFakeKey) {
			t.Fatal("positive control failed: the endpoint did NOT receive the credential, so the 0-hit proves nothing")
		}
		if strings.Contains(out, todoLLMFakeKey) || strings.Contains(errOut, todoLLMFakeKey) {
			t.Error("the fake credential leaked into process output on the failure path")
		}
	})
}
