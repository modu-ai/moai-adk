package jev

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test seams. No test in this package contacts the real endpoint: every client
// under test either carries a fake Doer or is pointed at an httptest server.
// ---------------------------------------------------------------------------

// countingDoer records how many HTTP round trips were attempted and returns a
// canned response or error. Its Calls counter is what the "constructed no
// request" assertions read — a gate that merely suppressed the ANSWER would
// still show a non-zero count here.
type countingDoer struct {
	Calls    int
	Bodies   []string
	Response func(attempt int) (*http.Response, error)
}

func (d *countingDoer) Do(req *http.Request) (*http.Response, error) {
	d.Calls++
	if req.Body != nil {
		raw, _ := io.ReadAll(req.Body)
		d.Bodies = append(d.Bodies, string(raw))
	}
	return d.Response(d.Calls)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

const okBody = `{"model":"jev-1.13.0","usage":{"input_tokens":123},"answers":[` +
	`{"question_id":"q1","kind":"noul","noul":true,"probability":0.62}]}`

// testClient builds an enabled client with a stored credential and the supplied
// Doer — the standard fixture for the transport tests.
func testClient(d *countingDoer) *Client {
	c := New(true)
	c.HTTP = d
	c.LoadCredential = func() string { return "NOT-A-REAL-KEY-fixture" }
	return c
}

func oneQuestion() Request {
	return Request{
		State:     "the card says the premise is stale",
		Questions: []Question{{ID: "q1", Text: "Is the premise dead?", Kind: KindNoul}},
	}
}

// ---------------------------------------------------------------------------
// The pinned model id (REQ-JEVC-003)
// ---------------------------------------------------------------------------

func TestModelID_IsPinnedNotAnAlias(t *testing.T) {
	if ModelID == "" {
		t.Fatal("ModelID is empty — no model is pinned")
	}
	if strings.Contains(ModelID, "latest") {
		t.Fatalf("ModelID = %q — a moving alias makes every measurement unattributable (REQ-JEVC-003)", ModelID)
	}
}

// TestModelID_MatchesVendorVersionedShape guards the shape of the pin, not just
// its non-aliasness. The two are different defects: an alias moves under you,
// while an id of the wrong shape is simply not a model the endpoint knows, and
// the endpoint is the only party that can say so. Every test here answers a
// fake server that echoes whatever id it is handed, so a malformed pin travels
// through the whole suite green and fails only in production. The vendor
// publishes versioned ids as the family name plus three dot-separated numbers
// (`jev-1.13.0`); the bare family name (`jev-1.13`) names a release in prose
// and is not an accepted id.
func TestModelID_MatchesVendorVersionedShape(t *testing.T) {
	shape := regexp.MustCompile(`^jev-\d+\.\d+\.\d+$`)
	if !shape.MatchString(ModelID) {
		t.Fatalf("ModelID = %q does not match the vendor's versioned-id shape %s — the endpoint rejects an id it does not publish, and no test here can detect that because every transport test answers a fake server", ModelID, shape)
	}
	// Negative control: the guard must reject the bare family name, or it is
	// asserting nothing that the previous test did not already cover.
	if shape.MatchString("jev-1.13") {
		t.Fatal("negative control failed: the shape pattern accepts the bare family name, so it would not have caught the defect it exists to catch")
	}
}

// AC: the alias appears in no request this package constructs. The zero-hit is
// paired with a positive control — the pinned id MUST appear — so a clean scan
// is attributable to the payload rather than to a scan that never read it.
func TestRequestBody_CarriesPinnedIDAndNoAlias(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) { return jsonResponse(200, okBody), nil }}
	c := testClient(d)
	if got := c.Ask(context.Background(), oneQuestion()); !got.OK() {
		t.Fatalf("Ask availability = %q (%s), want available", got.Availability, got.Condition)
	}
	if len(d.Bodies) != 1 {
		t.Fatalf("recorded %d request bodies, want 1", len(d.Bodies))
	}
	body := d.Bodies[0]
	if strings.Contains(body, "jev-latest") {
		t.Errorf("request body carries the moving alias jev-latest:\n%s", body)
	}
	if !strings.Contains(body, ModelID) {
		t.Fatalf("positive control failed: request body does not carry the pinned id %q, so the alias scan above read nothing meaningful:\n%s", ModelID, body)
	}
}

// ---------------------------------------------------------------------------
// The gate (REQ-JEVC-017)
// ---------------------------------------------------------------------------

func TestDisabled_ConstructsNoRequest(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) {
		t.Fatal("transport called while the capability is disabled")
		return nil, nil
	}}
	c := New(false)
	c.HTTP = d
	c.LoadCredential = func() string { return "NOT-A-REAL-KEY-fixture" }

	got := c.Ask(context.Background(), oneQuestion())
	if got.Availability != Disabled {
		t.Errorf("Availability = %q, want %q", got.Availability, Disabled)
	}
	if got.OK() {
		t.Error("OK() = true while disabled")
	}
	if d.Calls != 0 {
		t.Errorf("transport calls = %d, want 0 (REQ-JEVC-017)", d.Calls)
	}
	if len(got.Answers) != 0 {
		t.Errorf("Answers = %v, want none", got.Answers)
	}
}

// The gate must sit BELOW the credential check: a disabled capability with a
// credential present still constructs nothing, and a disabled capability
// without one reports disabled rather than no-credential — the doctor check
// needs exactly that distinction.
func TestDisabled_ReportsDisabledNotNoCredential(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) { return nil, errors.New("unreachable") }}
	c := New(false)
	c.HTTP = d
	c.LoadCredential = func() string { return "" }
	if got := c.Ask(context.Background(), oneQuestion()); got.Availability != Disabled {
		t.Errorf("Availability = %q, want %q", got.Availability, Disabled)
	}
	if d.Calls != 0 {
		t.Errorf("transport calls = %d, want 0", d.Calls)
	}
}

// ---------------------------------------------------------------------------
// Fail-open (REQ-JEVC-005, REQ-JEVC-006)
// ---------------------------------------------------------------------------

func TestNoCredential_TypedUnavailableNotAnError(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) {
		t.Fatal("transport called with no credential")
		return nil, nil
	}}
	c := New(true)
	c.HTTP = d
	c.LoadCredential = func() string { return "" }

	got := c.Ask(context.Background(), oneQuestion())
	if got.Availability != NoCredential {
		t.Errorf("Availability = %q, want %q", got.Availability, NoCredential)
	}
	if d.Calls != 0 {
		t.Errorf("transport calls = %d, want 0", d.Calls)
	}
	if got.Condition == "" {
		t.Error("Condition is empty — the observed condition must be named")
	}
}

func TestTransportConditions_MapToTypedUnavailable(t *testing.T) {
	cases := []struct {
		name string
		resp func(int) (*http.Response, error)
		want Availability
	}{
		{"401 unauthorized", func(int) (*http.Response, error) { return jsonResponse(401, `{"error":"bad key"}`), nil }, Unauthorized},
		{"429 rate limited", func(int) (*http.Response, error) { return jsonResponse(429, `{"error":"slow down"}`), nil }, RateLimited},
		{"529 overloaded", func(int) (*http.Response, error) { return jsonResponse(529, `{"error":"overloaded"}`), nil }, Overloaded},
		{"network unreachable", func(int) (*http.Response, error) {
			return nil, &net.OpError{Op: "dial", Err: errors.New("connect: no route to host")}
		}, Unreachable},
		{"unexpected 500", func(int) (*http.Response, error) { return jsonResponse(500, `{"error":"boom"}`), nil }, Unreachable},
		{"undecodable body", func(int) (*http.Response, error) { return jsonResponse(200, `not json`), nil }, Unreachable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := testClient(&countingDoer{Response: tc.resp})
			c.MaxRetries = 0
			got := c.Ask(context.Background(), oneQuestion())
			if got.Availability != tc.want {
				t.Errorf("Availability = %q, want %q", got.Availability, tc.want)
			}
			if got.OK() {
				t.Error("OK() = true on an unavailable result")
			}
			if got.Condition == "" {
				t.Error("Condition is empty — REQ-JEVC-006 requires the observed condition to be carried")
			}
			if line := got.NoticeLine(); strings.Count(line, "\n") != 0 {
				t.Errorf("NoticeLine() = %q, want a single line (REQ-JEVC-007: at most one notice line)", line)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Retry policy (REQ-JEVC-008 / AC-JEVC-009)
// ---------------------------------------------------------------------------

func TestRetry_NotIssuedForANonRepeatableRequest(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) {
		return nil, &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	}}
	c := testClient(d)
	c.MaxRetries = 3
	req := oneQuestion() // Repeatable is false by default — the fail-safe direction.

	c.Ask(context.Background(), req)
	if d.Calls != 1 {
		t.Errorf("transport calls = %d, want 1 — a request not declared repeatable MUST NOT be retried (REQ-JEVC-008)", d.Calls)
	}
}

func TestRetry_BoundedForARepeatableRequestThatFailedBeforeDelivery(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) {
		return nil, &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	}}
	c := testClient(d)
	c.MaxRetries = 2
	req := oneQuestion()
	req.Repeatable = true

	c.Ask(context.Background(), req)
	if d.Calls != 3 {
		t.Errorf("transport calls = %d, want 3 (1 attempt + MaxRetries=2)", d.Calls)
	}
}

// The ambiguous case is the one the requirement is written for: the request may
// have been delivered, so repetition is NOT provably side-effect-free even on a
// request the caller declared repeatable.
func TestRetry_NotIssuedForAnAmbiguousFailure(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) {
		// Not a dial error: the request may have reached the server and the
		// failure happened while reading the response.
		return nil, errors.New("unexpected EOF reading response")
	}}
	c := testClient(d)
	c.MaxRetries = 3
	req := oneQuestion()
	req.Repeatable = true

	c.Ask(context.Background(), req)
	if d.Calls != 1 {
		t.Errorf("transport calls = %d, want 1 — an ambiguous failure MUST NOT be retried (REQ-JEVC-008)", d.Calls)
	}
}

// ---------------------------------------------------------------------------
// Size bounds (REQ-JEVC-009)
// ---------------------------------------------------------------------------

func TestOversize_RefusesRatherThanTruncates(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) {
		t.Fatal("transport called with an oversize request")
		return nil, nil
	}}
	c := testClient(d)

	huge := strings.Repeat("x", (MaxStateAndQuestionTokens+1_000)*bytesPerTokenEstimate)
	req := Request{State: huge, Questions: []Question{{ID: "q1", Text: "?", Kind: KindNoul}}}

	got := c.Ask(context.Background(), req)
	if got.Availability != Oversize {
		t.Errorf("Availability = %q, want %q", got.Availability, Oversize)
	}
	if d.Calls != 0 {
		t.Errorf("transport calls = %d, want 0 — an oversize request is refused, never truncated", d.Calls)
	}
	// Negative control: the same client accepts a request inside the bound, so
	// the refusal above is attributable to the size rather than to a client
	// that refuses everything.
	d2 := &countingDoer{Response: func(int) (*http.Response, error) { return jsonResponse(200, okBody), nil }}
	c2 := testClient(d2)
	if got := c2.Ask(context.Background(), oneQuestion()); !got.OK() {
		t.Fatalf("negative control failed: an in-bounds request was refused with %q (%s)", got.Availability, got.Condition)
	}
}

func TestOversize_WholeRequestBoundSeparateFromStateBound(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) {
		t.Fatal("transport called with an oversize request")
		return nil, nil
	}}
	c := testClient(d)

	// Each question is individually small, and the state plus the LONGEST
	// question stays under the 32k bound; only their sum crosses 64k.
	state := strings.Repeat("s", 20_000*bytesPerTokenEstimate)
	var qs []Question
	for i := 0; i < 20; i++ {
		qs = append(qs, Question{ID: "q", Text: strings.Repeat("q", 3_000*bytesPerTokenEstimate), Kind: KindNoul})
	}
	got := c.Ask(context.Background(), Request{State: state, Questions: qs})
	if got.Availability != Oversize {
		t.Errorf("Availability = %q, want %q — the whole-request bound must be checked separately", got.Availability, Oversize)
	}
}

func TestEmptyQuestions_IsRefusedNotSent(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) {
		t.Fatal("transport called with no questions")
		return nil, nil
	}}
	c := testClient(d)
	got := c.Ask(context.Background(), Request{State: "anything"})
	if got.OK() {
		t.Error("OK() = true for a request carrying no questions")
	}
	if d.Calls != 0 {
		t.Errorf("transport calls = %d, want 0", d.Calls)
	}
}

// ---------------------------------------------------------------------------
// Batching (REQ-JEVC-010)
// ---------------------------------------------------------------------------

func TestBatching_SeveralQuestionsOverOneStateSendOneRequest(t *testing.T) {
	body := `{"model":"jev-1.13.0","usage":{"input_tokens":200},"answers":[` +
		`{"question_id":"a","kind":"noul","noul":true,"probability":0.7},` +
		`{"question_id":"b","kind":"noul","noul":false,"probability":0.3},` +
		`{"question_id":"c","kind":"score","score":0.5,"probability":0.5}]}`
	d := &countingDoer{Response: func(int) (*http.Response, error) { return jsonResponse(200, body), nil }}
	c := testClient(d)

	got := c.Ask(context.Background(), Request{
		State: "one state",
		Questions: []Question{
			{ID: "a", Text: "first?", Kind: KindNoul},
			{ID: "b", Text: "second?", Kind: KindNoul},
			{ID: "c", Text: "third?", Kind: KindScore},
		},
	})
	if !got.OK() {
		t.Fatalf("Availability = %q (%s)", got.Availability, got.Condition)
	}
	if d.Calls != 1 {
		t.Errorf("transport calls = %d, want 1 — N questions over one state is ONE request (REQ-JEVC-010)", d.Calls)
	}
	if len(got.Answers) != 3 {
		t.Errorf("len(Answers) = %d, want 3", len(got.Answers))
	}
}

// ---------------------------------------------------------------------------
// Usage accounting (REQ-JEVC-004)
// ---------------------------------------------------------------------------

func TestUsage_RecordsInputTokensAndPinnedModelPerCall(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) { return jsonResponse(200, okBody), nil }}
	c := testClient(d)

	got := c.Ask(context.Background(), oneQuestion())
	if got.Usage.InputTokens <= 0 {
		t.Errorf("Usage.InputTokens = %d, want > 0", got.Usage.InputTokens)
	}
	if got.Usage.Model != ModelID {
		t.Errorf("Usage.Model = %q, want the pinned %q", got.Usage.Model, ModelID)
	}

	c.Ask(context.Background(), oneQuestion())
	log := c.UsageLog()
	if len(log) != 2 {
		t.Fatalf("len(UsageLog()) = %d, want 2 — the record is per CALL", len(log))
	}
	for i, u := range log {
		if u.Model != ModelID {
			t.Errorf("UsageLog()[%d].Model = %q, want %q", i, u.Model, ModelID)
		}
	}
}

// An unavailable call is still a call: its record must exist so a later cost or
// provenance question is answerable from the record rather than recollection.
func TestUsage_RecordedEvenWhenTheCallDidNotSucceed(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) { return jsonResponse(429, `{}`), nil }}
	c := testClient(d)
	c.MaxRetries = 0
	c.Ask(context.Background(), oneQuestion())
	if len(c.UsageLog()) != 1 {
		t.Errorf("len(UsageLog()) = %d, want 1 — a sent-but-refused call is still a call", len(c.UsageLog()))
	}
}

// A call the gate stopped never reached the network and is NOT charged.
func TestUsage_NotRecordedWhenNoRequestWasConstructed(t *testing.T) {
	c := New(false)
	c.HTTP = &countingDoer{Response: func(int) (*http.Response, error) { return nil, nil }}
	c.Ask(context.Background(), oneQuestion())
	if len(c.UsageLog()) != 0 {
		t.Errorf("len(UsageLog()) = %d, want 0 — a disabled call constructs no request and is charged nothing", len(c.UsageLog()))
	}
}

// ---------------------------------------------------------------------------
// Secret screening (REQ-JEVC-021 / AC-JEVC-015)
// ---------------------------------------------------------------------------

func TestScreening_RefusesAPayloadCarryingACredentialShapedToken(t *testing.T) {
	// Built at runtime from parts so the fixture is not itself a literal that
	// a repository secret scanner would flag.
	tokens := []string{
		"sk" + "-" + strings.Repeat("a", 32),
		"ghp" + "_" + strings.Repeat("b", 36),
		"AKIA" + strings.Repeat("C", 16),
		"AWS_SECRET_ACCESS_KEY=" + strings.Repeat("d", 40),
	}
	for _, tok := range tokens {
		t.Run(tok[:6], func(t *testing.T) {
			d := &countingDoer{Response: func(int) (*http.Response, error) {
				t.Fatal("transport called with a payload carrying a credential-shaped token")
				return nil, nil
			}}
			c := testClient(d)
			got := c.Ask(context.Background(), Request{
				State:     "log excerpt: " + tok,
				Questions: []Question{{ID: "q1", Text: "anything?", Kind: KindNoul}},
			})
			if got.Availability != SecretDetected {
				t.Errorf("Availability = %q, want %q", got.Availability, SecretDetected)
			}
			if d.Calls != 0 {
				t.Errorf("transport calls = %d, want 0 — the refusal happens BEFORE the send", d.Calls)
			}
			if got.Condition == "" {
				t.Error("Condition is empty — the refusal must report itself")
			}
		})
	}
}

// The stored credential appearing verbatim in the payload is the case the
// screener exists for, and the one a generic pattern set would miss.
func TestScreening_RefusesTheStoredCredentialAppearingVerbatim(t *testing.T) {
	const stored = "NOT-A-REAL-KEY-0123456789abcdefghij"
	d := &countingDoer{Response: func(int) (*http.Response, error) {
		t.Fatal("transport called with the stored credential in the payload")
		return nil, nil
	}}
	c := New(true)
	c.HTTP = d
	c.LoadCredential = func() string { return stored }

	got := c.Ask(context.Background(), Request{
		State:     "pasted config: " + stored,
		Questions: []Question{{ID: "q1", Text: "anything?", Kind: KindNoul}},
	})
	if got.Availability != SecretDetected {
		t.Errorf("Availability = %q, want %q", got.Availability, SecretDetected)
	}
}

// The non-detecting direction. A screener that never fires passes every test
// that only checks clean payloads; a screener that always fires passes every
// test that only checks dirty ones. Both directions are asserted.
func TestScreening_OrdinaryPayloadSends(t *testing.T) {
	d := &countingDoer{Response: func(int) (*http.Response, error) { return jsonResponse(200, okBody), nil }}
	c := testClient(d)
	got := c.Ask(context.Background(), Request{
		State: "card t1020 says the premise may be stale; the referenced file was renamed. " +
			"Paths: internal/jev/jev.go, internal/jevcred. Commit abc1234.",
		Questions: []Question{{ID: "q1", Text: "Is the premise dead?", Kind: KindNoul}},
	})
	if got.Availability == SecretDetected {
		t.Fatalf("an ordinary payload was refused as secret-bearing: %s", got.Condition)
	}
	if !got.OK() {
		t.Fatalf("Availability = %q (%s), want available", got.Availability, got.Condition)
	}
	if d.Calls != 1 {
		t.Errorf("transport calls = %d, want 1", d.Calls)
	}
}

func TestScreenPayload_BothDirectionsDirectly(t *testing.T) {
	clean := Request{State: "nothing sensitive here", Questions: []Question{{ID: "q", Text: "ok?"}}}
	if reason, hit := ScreenPayload(clean, ""); hit {
		t.Errorf("ScreenPayload(clean) reported a hit: %s", reason)
	}
	dirty := Request{State: "key=" + "sk" + "-" + strings.Repeat("z", 32), Questions: []Question{{ID: "q", Text: "ok?"}}}
	reason, hit := ScreenPayload(dirty, "")
	if !hit {
		t.Fatal("ScreenPayload(dirty) reported no hit — the screener never fires")
	}
	if reason == "" {
		t.Error("ScreenPayload reported a hit with an empty reason")
	}
	// A short stored credential must not turn every payload into a hit.
	if _, hit := ScreenPayload(clean, "ab"); hit {
		t.Error("a two-character stored credential made a clean payload read as secret-bearing")
	}
}

// ---------------------------------------------------------------------------
// Labelling (REQ-JEVC-013 / AC-JEVC-004)
// ---------------------------------------------------------------------------

func TestAnswerLabel_IsDistinguishableFromMeasurementAndJudgement(t *testing.T) {
	a := Answer{QuestionID: "q1", Kind: KindNoul, Noul: true, Probability: 0.62}
	label := a.Label()
	if !strings.Contains(label, SignalLabel) {
		t.Errorf("Label() = %q, want it to carry the %q marker", label, SignalLabel)
	}
	if !strings.Contains(label, ModelID) {
		t.Errorf("Label() = %q, want it to name the pinned model", label)
	}
	lower := strings.ToLower(label)
	for _, forbidden := range []string{"measured", "verified", "confirmed"} {
		if strings.Contains(lower, forbidden) {
			t.Errorf("Label() = %q contains %q — a model signal must not read as a mechanical measurement", label, forbidden)
		}
	}
}

func TestNoticeLine_NamesTheConditionAndStaysOneLine(t *testing.T) {
	for _, av := range []Availability{Disabled, NoCredential, Unauthorized, RateLimited, Overloaded, Unreachable, Oversize, SecretDetected} {
		r := Result{Availability: av, Condition: "observed condition"}
		line := r.NoticeLine()
		if line == "" {
			t.Errorf("NoticeLine() for %q is empty", av)
		}
		if strings.Contains(line, "\n") {
			t.Errorf("NoticeLine() for %q spans lines: %q", av, line)
		}
		if !strings.Contains(line, string(av)) {
			t.Errorf("NoticeLine() for %q = %q, want the condition named", av, line)
		}
	}
	if got := (Result{Availability: Available}).NoticeLine(); got != "" {
		t.Errorf("NoticeLine() on an available result = %q, want empty (nothing to notice)", got)
	}
}

// ---------------------------------------------------------------------------
// Endpoint and transport shape, exercised against a local httptest server.
// ---------------------------------------------------------------------------

func TestAgainstLocalServer_AuthHeaderAndPathAndNoRealNetwork(t *testing.T) {
	var gotAuth, gotPath, gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(okBody))
	}))
	defer srv.Close()

	c := New(true)
	c.Endpoint = srv.URL + "/v1/systemone"
	c.LoadCredential = func() string { return "NOT-A-REAL-KEY-local" }

	got := c.Ask(context.Background(), oneQuestion())
	if !got.OK() {
		t.Fatalf("Availability = %q (%s)", got.Availability, got.Condition)
	}
	if gotPath != "/v1/systemone" {
		t.Errorf("path = %q, want /v1/systemone", gotPath)
	}
	if gotAuth == "" || !strings.Contains(gotAuth, "NOT-A-REAL-KEY-local") {
		t.Errorf("Authorization header = %q, want it to carry the credential", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
}

func TestDefaultEndpoint_IsTheDocumentedOne(t *testing.T) {
	if EndpointURL != "https://api.typesafe.ai/v1/systemone" {
		t.Errorf("EndpointURL = %q, want https://api.typesafe.ai/v1/systemone", EndpointURL)
	}
	if New(true).Endpoint != EndpointURL {
		t.Error("a fresh client does not default to EndpointURL")
	}
}

func TestContextCancellation_IsUnavailableNotAnError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := testClient(&countingDoer{Response: func(int) (*http.Response, error) {
		return nil, context.Canceled
	}})
	c.MaxRetries = 0
	got := c.Ask(ctx, oneQuestion())
	if got.OK() {
		t.Error("OK() = true on a cancelled context")
	}
	if got.Availability != Unreachable {
		t.Errorf("Availability = %q, want %q", got.Availability, Unreachable)
	}
}

// ---------------------------------------------------------------------------
// Token estimate
// ---------------------------------------------------------------------------

func TestEstimateTokens_MonotonicAndConservative(t *testing.T) {
	if EstimateTokens("") != 0 {
		t.Error("EstimateTokens(\"\") != 0")
	}
	short := EstimateTokens(strings.Repeat("a", 100))
	long := EstimateTokens(strings.Repeat("a", 1000))
	if long <= short {
		t.Errorf("estimate is not monotonic in length: %d for 1000 bytes vs %d for 100", long, short)
	}
	// Multi-byte text must not estimate LOWER than the same rune count in
	// ASCII — under-estimating is the direction that lets an oversize payload
	// through, and the bound exists to refuse those.
	ascii := EstimateTokens(strings.Repeat("a", 30))
	cjk := EstimateTokens(strings.Repeat("가", 30))
	if cjk < ascii {
		t.Errorf("EstimateTokens under-estimates multi-byte text: %d (CJK) < %d (ASCII) for the same rune count", cjk, ascii)
	}
}
