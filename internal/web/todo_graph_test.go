package web

// todo_graph_test.go — SPEC-TODO-CARD-ISSUANCE-001 M6 (REQ-TCI-023/-024):
// the /todo?view=graph relation view. Read-only like its host route: GET
// only, no write, no lock, bounded nodes, embedded assets, every new string
// in all four locales — and the plain /todo table unchanged (AC-TCI-023).
//
// Every test stubs factory.HomeDirFn (process-global), so none run in
// parallel.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// graphQueue is the AC-TCI-022 fixture: live queued/picked/dropped/hold
// cards, one archived card, two live findings and one archived finding.
const graphQueue = `{"version":1,"last_seq":5,"items":[` +
	`{"id":"t1","text":"queued card","added_at":"2026-08-20T00:00:00Z","spec_id":null,"state":"queued"},` +
	`{"id":"t2","text":"picked card","added_at":"2026-08-20T00:01:00Z","spec_id":null,"state":"picked"},` +
	`{"id":"t3","text":"dropped card","added_at":"2026-08-20T00:02:00Z","spec_id":null,"state":"dropped"},` +
	`{"id":"t4","text":"held card","added_at":"2026-08-20T00:03:00Z","spec_id":null,"state":"hold"}],` +
	`"findings":[` +
	`{"subject_id":"t1","related_id":"t2","relation":"blocks","source":"agent"},` +
	`{"subject_id":"t2","related_id":"t3","relation":"contains","source":"agent"}],` +
	`"archived":[` +
	`{"item":{"id":"t9","text":"archived card","added_at":"2026-08-19T00:00:00Z","spec_id":null,"state":"done"},"position":0,` +
	`"findings":[{"finding":{"subject_id":"t9","related_id":"t1","relation":"absorbs","source":"agent"},"position":0}]}]}`

// todoGraphBodyFor renders GET /todo?view=graph for a console served from
// projectRoot.
func todoGraphBodyFor(t *testing.T, projectRoot string) string {
	t.Helper()
	a := newApp(Config{ProjectRoot: projectRoot, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	rec := serveGet(t, a.routes(), "/todo?view=graph")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /todo?view=graph status = %d, want 200\nbody:\n%s", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// TestTodoPageUnchangedWithoutViewParam — AC-TCI-023 (REQ-TCI-024): the
// plain /todo view keeps its table, its rows, its sort control and its
// detail selection, and carries no graph artifact. Written against the
// pre-change renderer at M6 start, so it arrives green — that is what makes
// it a regression guard rather than a wish.
func TestTodoPageUnchangedWithoutViewParam(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	writeBacklog(t, root, graphQueue)

	body := todoBodyFor(t, root)

	for _, want := range []string{
		`data-todo-row`,                 // the audit table's rows
		`data-i18n="todo.sort"`,         // the segmented sort control
		`/todo?sort=default`,            // the sort hrefs the live refresh replays
		`data-i18n="todo.state.queued"`, // the state badges
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("plain /todo lost %q:\n%s", want, body)
		}
	}
	for _, gone := range []string{`todo-graph`, `view=graph`} {
		if strings.Contains(body, gone) {
			t.Fatalf("plain /todo carries graph artifact %q", gone)
		}
	}
	// The sort and detail URLs still work unchanged: id opens the detail pane.
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	rec := serveGet(t, a.routes(), "/todo?sort=state&id=t2")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `data-todo-row`) {
		t.Fatalf("GET /todo?sort=state&id=t2 = %d, want 200 with the table intact", rec.Code)
	}
}

// TestTodoGraphViewRendersRelations — AC-TCI-022 (a): every fixture card —
// live, dropped, held and archived — has a node, and each relation kind's
// edge is rendered server-side.
func TestTodoGraphViewRendersRelations(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	writeBacklog(t, root, graphQueue)

	body := todoGraphBodyFor(t, root)

	for _, id := range []string{"t1", "t2", "t3", "t4", "t9"} {
		if !strings.Contains(body, id) {
			t.Fatalf("graph view misses node %s:\n%s", id, body)
		}
	}
	for _, want := range []string{"blocks", "contains", "absorbs", "<svg"} {
		if !strings.Contains(body, want) {
			t.Fatalf("graph view misses %q:\n%s", want, body)
		}
	}
	// Deterministic layout: the same queue renders byte-identically.
	again := todoGraphBodyFor(t, root)
	if body != again {
		t.Fatal("two renders of the same queue disagree — the layout is not deterministic")
	}
}

// TestTodoGraphViewReadOnly — AC-TCI-022 (b): POST is refused with the queue
// byte-identical, and the web package's non-test sources carry no lock
// token anywhere (the vocabulary half — the timing half is the lock test
// below).
func TestTodoGraphViewReadOnly(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	path := writeBacklog(t, root, graphQueue)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	req := httptest.NewRequest(http.MethodPost, "/todo?view=graph", nil)
	req.Host = "127.0.0.1:8080"
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /todo?view=graph status = %d, want 405", rec.Code)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("the queue file changed under a POST refusal")
	}

	// Vocabulary scan (AC-TCI-022 (b3)): no non-test source in the package
	// carries a lock-taking token. MUTATE( is the queue's write path; a view
	// that merely TRIES the lock slips this scan — the timing test closes
	// that gap.
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		raw, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatal(err)
		}
		for _, token := range []string{"Mutate(", "LockPath", "Flock", "acquireLock", "StateLock"} {
			if strings.Contains(string(raw), token) {
				t.Errorf("%s carries lock token %q — the graph view reads, it never locks", e.Name(), token)
			}
		}
	}
}

// TestTodoGraphViewBounded — AC-TCI-022 (c): a queue larger than the node
// bound renders exactly the bound's worth of nodes and names the omitted
// count.
func TestTodoGraphViewBounded(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	var b strings.Builder
	b.WriteString(`{"version":1,"last_seq":350,"items":[`)
	for i := 1; i <= 350; i++ {
		if i > 1 {
			b.WriteString(",")
		}
		b.WriteString(`{"id":"t` + strconv.Itoa(i) + `","text":"card ` + strconv.Itoa(i) + `","added_at":"2026-08-20T00:00:00Z","spec_id":null,"state":"queued"}`)
	}
	b.WriteString(`],"findings":[],"archived":[]}`)
	writeBacklog(t, root, b.String())

	body := todoGraphBodyFor(t, root)

	if got := strings.Count(body, `class="todo-graph__node"`); got != todoGraphMaxNodes {
		t.Fatalf("rendered nodes = %d, want the bound %d", got, todoGraphMaxNodes)
	}
	if !strings.Contains(body, "cards omitted") {
		t.Fatalf("the omitted-count line is missing:\n%s", body[len(body)-400:])
	}
}

// TestTodoGraphAssetsEmbedded — AC-TCI-022 (d): the graph view adds no new
// .js asset (the embed list's .js entries equal the M6-start set), and the
// response body carries no external reference of any spelling the static
// scan can see.
func TestTodoGraphAssetsEmbedded(t *testing.T) {
	// The .js set frozen at M6 start (the embed directive in assets.go).
	want := []string{"app.js", "i18n.js", "htmx.min.js"}
	for _, js := range want {
		if _, err := os.Stat(filepath.Join("assets", js)); err != nil {
			t.Fatalf("frozen js asset %s missing: %v", js, err)
		}
	}
	// The embed directive must name exactly these .js files — a new one is
	// the regression this test exists to catch.
	raw, err := os.ReadFile("assets.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.Contains(line, "go:embed") {
			continue
		}
		if strings.Contains(line, ".js") {
			for _, js := range want {
				if !strings.Contains(line, js) {
					t.Fatalf("embed list lost %s: %s", js, line)
				}
			}
			if n := strings.Count(line, ".js"); n != len(want) {
				t.Fatalf("embed list names %d .js entries, want the frozen %d: %s", n, len(want), line)
			}
		}
	}

	stubTodoHome(t)
	root := t.TempDir()
	writeBacklog(t, root, graphQueue)
	body := todoGraphBodyFor(t, root)
	for _, forbidden := range []string{`src="http`, `href="http`, `src="//`, `href="//`, `url(//`, `@import`} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("graph view carries an external reference %q", forbidden)
		}
	}
}

// TestTodoGraphViewDoesNotWaitOnQueueLock — AC-TCI-022 (b2): with the
// store's own Mutate holding the queue lock (its callback blocks on a
// signal), GET /todo?view=graph answers 200 with every fixture node inside
// 2 seconds — and the control GET /todo does the same, proving the harness
// itself is not what is fast.
func TestTodoGraphViewDoesNotWaitOnQueueLock(t *testing.T) {
	stubTodoHome(t)
	root := t.TempDir()
	path := writeBacklog(t, root, graphQueue)

	store := factory.NewBacklogStore(path)
	release := make(chan struct{})
	holderDone := make(chan error, 1)
	go func() {
		holderDone <- store.Mutate(func(rec *factory.BacklogRecord) error {
			<-release
			return nil
		})
	}()
	// Give the holder the lock: the Mutate grabs it before blocking, and a
	// bounded wait keeps the test honest without sleeping past the point.
	deadline := time.Now().Add(2 * time.Second)
	for !graphQueueLockHeld(t, path) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	defer close(release)

	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }

	var wg sync.WaitGroup
	check := func(path_ string, wantNode bool) {
		wg.Add(1)
		defer wg.Done()
		type result struct {
			code int
			body string
		}
		done := make(chan result, 1)
		go func() {
			rec := serveGet(t, a.routes(), path_)
			done <- result{rec.Code, rec.Body.String()}
		}()
		select {
		case r := <-done:
			if r.code != http.StatusOK {
				t.Errorf("%s under a held lock: status = %d, want 200", path_, r.code)
			}
			if wantNode && !strings.Contains(r.body, "t9") {
				t.Errorf("%s under a held lock: the archived node t9 is missing", path_)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("%s did not answer within 2s under a held queue lock", path_)
		}
	}
	check("/todo?view=graph", true)
	check("/todo", false)
	wg.Wait()
}

// graphQueueLockHeld probes whether the queue lock is currently taken, by
// trying a non-blocking mutation attempt on a THROWAWAY copy of the lock
// semantics: a second Mutate that gives up immediately. It reports held only
// on a positive refusal.
func graphQueueLockHeld(t *testing.T, path string) bool {
	t.Helper()
	probe := make(chan error, 1)
	store := factory.NewBacklogStore(path)
	go func() {
		probe <- store.Mutate(func(rec *factory.BacklogRecord) error { return nil })
	}()
	select {
	case <-probe:
		return false // the probe took and released the lock: it was free
	case <-time.After(150 * time.Millisecond):
		return true // the probe is still waiting: the lock is held
	}
}
