package webhook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/docstore"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Every request here goes to an httptest server on the loopback interface of this machine. The secrets
// are made up for the test; nothing reaches a network or a real host.

var errCrash = errors.New("simulated crash")

// crashDocs writes through to a real claim store and, at the chosen put, reports a crash after the write.
type crashDocs struct {
	inner   *claim.Store
	puts    int
	crashAt int
}

func (c *crashDocs) GetDoc(kind, name string) (claim.Doc, error) { return c.inner.GetDoc(kind, name) }

func (c *crashDocs) PutDoc(kind, name string, rev int64, data []byte) (claim.Doc, error) {
	c.puts++
	d, err := c.inner.PutDoc(kind, name, rev, data)
	if err == nil && c.puts == c.crashAt {
		return claim.Doc{}, errCrash
	}
	return d, err
}

type site struct {
	t    *testing.T
	srv  *httptest.Server
	st   *claim.Store
	docs *crashDocs
	q    map[string]*docstore.Queue // by repo
	cli  *http.Client
}

var secrets = map[string]string{"o/a": "secret-for-a", "o/b": "secret-for-b"}

func newSite(t *testing.T, maxDoc int) *site {
	t.Helper()
	st, err := claim.OpenWithLimit(t.TempDir(), 100)
	if err != nil {
		t.Fatal(err)
	}
	s := &site{t: t, st: st, docs: &crashDocs{inner: st}, q: map[string]*docstore.Queue{}}
	h := &Handler{Provider: "github", MaxBody: 4 << 10}
	for repo, sec := range secrets {
		q := &docstore.Queue{Docs: s.docs, Kind: docstore.CollectorKind, Key: repo + "/hooks",
			MaxDocBytes: maxDoc, SeenLimit: 4, MaxIDBytes: 36, Alive: func() bool { return true }}
		s.q[repo] = q
		h.Routes = append(h.Routes, Route{Repo: repo, Secret: func() []byte { return []byte(sec) }, Queue: q})
	}
	s.srv = httptest.NewServer(h)
	t.Cleanup(s.srv.Close)
	if ip := s.srv.Listener.Addr().(*net.TCPAddr).IP; !ip.IsLoopback() {
		t.Fatalf("the test server is not on the loopback interface: %v", ip)
	}
	s.cli = &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}}
	t.Cleanup(s.cli.CloseIdleConnections)
	return s
}

func issueBody(repo string, number int, action string) []byte {
	return []byte(fmt.Sprintf(`{"action":%q,"issue":{"number":%d},"repository":{"full_name":%q}}`, action, number, repo))
}

// post sends one delivery to path; sig == "" signs with secret.
func (s *site) post(path, event, id string, body []byte, secret, sig string) int {
	s.t.Helper()
	req, err := http.NewRequest(http.MethodPost, s.srv.URL+path, bytes.NewReader(body))
	if err != nil {
		s.t.Fatal(err)
	}
	if sig == "" {
		sig = Sign([]byte(secret), body)
	}
	req.Header.Set("X-Hub-Signature-256", sig)
	if id != "" {
		req.Header.Set("X-GitHub-Delivery", id)
	}
	req.Header.Set("X-GitHub-Event", event)
	resp, err := s.cli.Do(req)
	if err != nil {
		s.t.Fatal(err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func (s *site) pending(repo string) []string {
	s.t.Helper()
	p, err := s.q[repo].Pending()
	if err != nil {
		s.t.Fatal(err)
	}
	var out []string
	for _, e := range p {
		out = append(out, e.ID+"="+e.Key+":"+string(e.Kind))
	}
	return out
}

func eq(t *testing.T, what string, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("%s = %v, want %v", what, got, want)
	}
}

func TestASignedDeliveryIsQueuedThenAcknowledged(t *testing.T) {
	s := newSite(t, 0)
	if code := s.post("/hook/o/a", "issues", "d1", issueBody("o/a", 7, "edited"), secrets["o/a"], ""); code != http.StatusAccepted {
		t.Fatalf("status = %d", code)
	}
	eq(t, "queue a", s.pending("o/a"), "d1=7:modified")
	if code := s.post("/hook/o/a", "pull_request", "d2", []byte(`{"action":"opened","pull_request":{"number":9},"repository":{"full_name":"o/a"}}`), secrets["o/a"], ""); code != http.StatusAccepted {
		t.Fatalf("pull request status = %d", code)
	}
	eq(t, "queue a", s.pending("o/a"), "d1=7:modified", "d2=9:added")
}

// A forged or unsigned delivery is refused before its body is read, and nothing is queued or acknowledged.
func TestForgedAndUnsignedDeliveriesAreRefused(t *testing.T) {
	s := newSite(t, 0)
	body := issueBody("o/a", 7, "edited")
	cases := map[string]string{
		"wrong secret":      Sign([]byte("not-the-secret"), body),
		"missing":           "-",
		"no sha256 prefix":  strings.TrimPrefix(Sign([]byte(secrets["o/a"]), body), "sha256="),
		"signed other body": Sign([]byte(secrets["o/a"]), []byte(`{}`)),
	}
	for name, sig := range cases {
		hdr := sig
		if sig == "-" {
			hdr = ""
		}
		req, _ := http.NewRequest(http.MethodPost, s.srv.URL+"/hook/o/a", bytes.NewReader(body))
		if hdr != "" {
			req.Header.Set("X-Hub-Signature-256", hdr)
		}
		req.Header.Set("X-GitHub-Delivery", "forged")
		req.Header.Set("X-GitHub-Event", "issues")
		resp, err := s.cli.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want 401", name, resp.StatusCode)
		}
	}
	eq(t, "queue a", s.pending("o/a"))
}

// Repositories never meet: each route has its own secret and queue; a delivery signed for one, or
// claiming the other repository in its payload, never enters the wrong queue.
func TestDeliveriesForDifferentRepositoriesNeverMix(t *testing.T) {
	s := newSite(t, 0)
	s.post("/hook/o/a", "issues", "d1", issueBody("o/a", 7, "edited"), secrets["o/a"], "")
	s.post("/hook/o/b", "issues", "d1", issueBody("o/b", 7, "edited"), secrets["o/b"], "") // same id and number, other repository
	eq(t, "queue a", s.pending("o/a"), "d1=7:modified")
	eq(t, "queue b", s.pending("o/b"), "d1=7:modified")

	if code := s.post("/hook/o/a", "issues", "d2", issueBody("o/b", 8, "edited"), secrets["o/a"], ""); code != http.StatusBadRequest {
		t.Errorf("a payload for b on a's route (signed with a's secret): %d, want 400", code)
	}
	if code := s.post("/hook/o/a", "issues", "d3", issueBody("o/a", 8, "edited"), secrets["o/b"], ""); code != http.StatusUnauthorized {
		t.Errorf("a's payload signed with b's secret: %d, want 401", code)
	}
	if code := s.post("/hook/o/c", "issues", "d4", issueBody("o/c", 8, "edited"), "x", ""); code != http.StatusNotFound {
		t.Errorf("an unknown repository: %d, want 404", code)
	}
	if code := s.post("/hook/O/A", "issues", "d5", issueBody("O/A", 9, "edited"), secrets["o/a"], ""); code != http.StatusAccepted {
		t.Errorf("GitHub names compare case-insensitively: %d", code)
	}
	eq(t, "queue a", s.pending("o/a"), "d1=7:modified", "d5=9:modified")
	eq(t, "queue b", s.pending("o/b"), "d1=7:modified")
}

func TestARedeliveredIdIsAcknowledgedWithoutQueueingTwice(t *testing.T) {
	s := newSite(t, 0)
	for i := range 3 {
		if code := s.post("/hook/o/a", "issues", "d1", issueBody("o/a", 7, "edited"), secrets["o/a"], ""); code != http.StatusAccepted {
			t.Fatalf("delivery %d: %d", i, code)
		}
	}
	eq(t, "queue a", s.pending("o/a"), "d1=7:modified")
	if err := s.q["o/a"].Commit([]string{"d1"}); err != nil {
		t.Fatal(err)
	}
	s.post("/hook/o/a", "issues", "d1", issueBody("o/a", 7, "edited"), secrets["o/a"], "") // after it was consumed
	eq(t, "queue a", s.pending("o/a"))
}

// A full queue answers 503 (the host retries), keeps every event it holds, and accepts again once the
// consumer has made room.
func TestAFullQueueAnswers503AndKeepsWhatItHolds(t *testing.T) {
	s := newSite(t, 520)
	var held []string
	full := false
	for i := range 20 {
		id := fmt.Sprintf("delivery-%02d", i)
		code := s.post("/hook/o/a", "issues", id, issueBody("o/a", 100+i, "edited"), secrets["o/a"], "")
		if code == http.StatusServiceUnavailable {
			full = true
			break
		}
		if code != http.StatusAccepted {
			t.Fatalf("delivery %d: %d", i, code)
		}
		held = append(held, fmt.Sprintf("%s=%d:modified", id, 100+i))
	}
	if !full || len(held) == 0 {
		t.Fatalf("the queue never filled (held %d)", len(held))
	}
	eq(t, "queue a", s.pending("o/a"), held...)
	eq(t, "queue b is untouched by a's pressure", s.pending("o/b"))
	if err := s.q["o/a"].Commit([]string{"delivery-00"}); err != nil {
		t.Fatal(err)
	}
	if code := s.post("/hook/o/a", "issues", "delivery-retry", issueBody("o/a", 999, "edited"), secrets["o/a"], ""); code != http.StatusAccepted {
		t.Errorf("after room was made: %d", code)
	}
}

// The queue write lands but the process dies before it can answer: the host sees an error, redelivers,
// and the redelivery is absorbed — one event.
func TestACrashAfterTheQueueWriteIsAbsorbedByTheRedelivery(t *testing.T) {
	s := newSite(t, 0)
	s.docs.puts, s.docs.crashAt = 0, 1
	if code := s.post("/hook/o/a", "issues", "d1", issueBody("o/a", 7, "edited"), secrets["o/a"], ""); code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (not acknowledged)", code)
	}
	s.docs.crashAt = 0
	if code := s.post("/hook/o/a", "issues", "d1", issueBody("o/a", 7, "edited"), secrets["o/a"], ""); code != http.StatusAccepted {
		t.Fatalf("redelivery status = %d", code)
	}
	eq(t, "queue a", s.pending("o/a"), "d1=7:modified")
}

func TestMalformedRequestsAreRefused(t *testing.T) {
	s := newSite(t, 0)
	sec := secrets["o/a"]
	good := issueBody("o/a", 7, "edited")
	if code := s.post("/hook/o/a", "issues", "", good, sec, ""); code != http.StatusBadRequest {
		t.Errorf("no delivery id: %d", code)
	}
	if code := s.post("/hook/o/a", "issues", strings.Repeat("x", 37), good, sec, ""); code != http.StatusBadRequest {
		t.Errorf("overlong delivery id: %d", code)
	}
	if code := s.post("/hook/o/a", "issues", "d1", []byte(`not json`), sec, ""); code != http.StatusBadRequest {
		t.Errorf("bad JSON: %d", code)
	}
	if code := s.post("/hook/o/a", "issues", "d2", bytes.Repeat([]byte("a"), 5<<10), sec, ""); code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversize body: %d", code)
	}
	if code := s.post("/hook/o/a", "ping", "d3", good, sec, ""); code != http.StatusOK {
		t.Errorf("ping: %d", code)
	}
	if code := s.post("/hook/o/a", "star", "d4", good, sec, ""); code != http.StatusAccepted {
		t.Errorf("an event this channel does not carry is acknowledged and dropped: %d", code)
	}
	if code := s.post("/hook/o/a/extra", "issues", "d5", good, sec, ""); code != http.StatusNotFound {
		t.Errorf("extra path: %d", code)
	}
	resp, err := s.cli.Get(s.srv.URL + "/hook/o/a")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET: %d", resp.StatusCode)
	}
	eq(t, "queue a", s.pending("o/a"))
}

func TestARouteWithoutASecretOrALimitRefusesEverything(t *testing.T) {
	q := &docstore.Queue{Docs: &crashDocs{}, Kind: docstore.CollectorKind, Key: "k"}
	for name, h := range map[string]*Handler{
		"no secret": {MaxBody: 1024, Routes: []Route{{Repo: "o/a", Secret: func() []byte { return nil }, Queue: q}}},
		"no limit":  {Routes: []Route{{Repo: "o/a", Secret: func() []byte { return []byte("s") }, Queue: q}}},
		"no queue":  {MaxBody: 1024, Routes: []Route{{Repo: "o/a", Secret: func() []byte { return []byte("s") }}}},
	} {
		srv := httptest.NewServer(h)
		body := issueBody("o/a", 1, "edited")
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/hook/o/a", bytes.NewReader(body))
		req.Header.Set("X-Hub-Signature-256", Sign([]byte("s"), body))
		req.Header.Set("X-GitHub-Delivery", "d")
		req.Header.Set("X-GitHub-Event", "issues")
		resp, err := (&http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}}).Do(req)
		srv.Close()
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode < 500 {
			t.Errorf("%s: status %d, want a refusal", name, resp.StatusCode)
		}
	}
}

// From the socket to the consumer: a delivery over loopback is queued, the collector reports it once,
// and it leaves the queue only when the consumer acknowledges.
func TestADeliveryReachesTheCollectorAndLeavesOnlyOnAck(t *testing.T) {
	s := newSite(t, 0)
	s.post("/hook/o/a", "issues", "d1", issueBody("o/a", 7, "edited"), secrets["o/a"], "")
	now := time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)
	col := &forge.Collector{
		Client: &forge.Client{T: noCalls{}, B: &forge.Budget{}, Now: func() time.Time { return now },
			Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3}},
		Path: "/p", Interval: forge.Adaptive{Min: time.Second, Max: time.Minute, Safety: time.Hour},
		Differ: forge.DiffByKey("number"), Durable: s.q["o/a"],
	}
	rep := col.Tick(t.Context(), now)
	if len(rep.Changes) != 1 || rep.Changes[0].Key != "7" || rep.Changes[0].Kind != forge.ChangeModified {
		t.Fatalf("report = %+v", rep.Changes)
	}
	if again := col.Tick(t.Context(), now.Add(time.Second)); len(again.Changes) != 1 {
		t.Fatalf("unacknowledged event must repeat: %+v", again.Changes)
	}
	if err := col.Ack(rep); err != nil {
		t.Fatal(err)
	}
	eq(t, "queue a", s.pending("o/a"))
	if done := col.Tick(t.Context(), now.Add(2*time.Second)); len(done.Changes) != 0 {
		t.Errorf("acknowledged event repeated: %+v", done.Changes)
	}
}

// noCalls is a forge transport that fails the test if the collector reaches for the network: with a
// healthy push channel the first contact must rely on the feed.
type noCalls struct{}

func (noCalls) Do(context.Context, forge.Request) (forge.Response, error) {
	panic("the collector polled the host")
}

type busyDocs struct{}

func (busyDocs) GetDoc(string, string) (claim.Doc, error) { return claim.Doc{}, core.ErrNotHeld }
func (busyDocs) PutDoc(string, string, int64, []byte) (claim.Doc, error) {
	return claim.Doc{}, fmt.Errorf("%w: documents", core.ErrBusy)
}

// A busy document lock is not a server fault: the delivery is answered 503 (try again), not acknowledged.
func TestABusyQueueAnswers503NotAFault(t *testing.T) {
	q := &docstore.Queue{Docs: busyDocs{}, Kind: docstore.CollectorKind, Key: "o/a/hooks"}
	h := &Handler{Provider: "github", MaxBody: 4 << 10, Routes: []Route{{Repo: "o/a", Secret: func() []byte { return []byte("s") }, Queue: q}}}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	body := issueBody("o/a", 7, "edited")
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/hook/o/a", bytes.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", Sign([]byte("s"), body))
	req.Header.Set("X-GitHub-Delivery", "d1")
	req.Header.Set("X-GitHub-Event", "issues")
	resp, err := (&http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil}}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", resp.StatusCode)
	}
}
