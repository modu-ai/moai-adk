package web

// factory_screen_m9_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M9 (card t1399),
// AC-019: the web console drops the chain session board, serves the factory
// screen at /factory, and keeps the Todo screen live under the renamed live
// area key. These tests were written before the rename and observed red on the
// pre-M9 tree (the route did not exist, the board panel existed, the live key
// was still the retired word); the redirect half is in legacy_routes_test.go.

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/session"
)

// m9ScreenBody renders GET path for a console served from projectRoot and
// requires a 200.
func m9ScreenBody(t *testing.T, projectRoot, path string) string {
	t.Helper()
	a := newApp(Config{ProjectRoot: projectRoot, ProfileName: "default"})
	a.recordLastProfile = func(string) error { return nil }
	rec := serveGet(t, a.routes(), path)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want 200\nbody:\n%s", path, rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// TestFactoryScreenOmitsChainBoard — AC-019: the chain session board is gone
// from the screen's source and from what it renders. The rendered half writes
// the four chain-role records first: before the removal those records filled
// the board's role cards (their session ids were drawn), so a screen that
// still reads them would show the ids below.
func TestFactoryScreenOmitsChainBoard(t *testing.T) {
	if strings.Contains(readSource(t, "screens.templ"), "Chain session board") {
		t.Error("screens.templ still carries the chain session board panel")
	}

	root := t.TempDir()
	for _, role := range []string{"leader", "plan", "run", "sync"} {
		writeKanbanRecord(t, root, factory.Record{
			SessionID: "sess-chain-" + role, SpecID: "SPEC-CHAIN-001", Role: role, Backend: factory.BackendClaude,
		})
	}
	body := m9ScreenBody(t, root, "/factory")
	for _, gone := range []string{
		"Chain session board", "roles--chain", ".viewA\"", ".noSession\"", ".noStart\"", "chain.stopped",
		"sess-chain-leader", "sess-chain-plan", "sess-chain-run", "sess-chain-sync", "SPEC-CHAIN-001",
	} {
		if strings.Contains(body, gone) {
			t.Errorf("the factory screen still renders %q", gone)
		}
	}
}

// TestFactoryScreenStillShowsFactoryLanes — AC-019: the lanes panel and the
// SPEC pipeline panel stay on the screen now served at /factory.
func TestFactoryScreenStillShowsFactoryLanes(t *testing.T) {
	root := t.TempDir()
	pid := os.Getpid()
	writeFactoryRegistry(t, root, map[string]int{"lane-2": pid})
	writeActiveSessions(t, root, []session.Entry{liveEntry("sess-lane-2", pid)})
	writeKanbanRecord(t, root, factory.Record{
		SessionID: "sess-lane-2", SpecID: "SPEC-EXAMPLE-001", Role: "lane",
		Backend: factory.BackendGLM, Lane: 2, CardID: "t207",
	})

	body := m9ScreenBody(t, root, "/factory")
	for _, want := range []string{
		`data-lane="2"`, "t207", "SPEC-EXAMPLE-001", "state--live",
		`.lanes"`,       // the "Factory lanes" panel head
		`.viewB"`,       // the "SPEC pipeline" panel head
		`class="board"`, // the four-column pipeline board
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the factory screen is missing %q", want)
		}
	}
}

// m9AppJSEvents reads the EVENTS list of the embedded client script.
func m9AppJSEvents(t *testing.T) []string {
	t.Helper()
	m := regexp.MustCompile(`var EVENTS = \[([^\]]*)\]`).FindStringSubmatch(readSource(t, "assets/app.js"))
	if m == nil {
		t.Fatal("assets/app.js declares no EVENTS list")
	}
	var out []string
	for _, q := range regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(m[1], -1) {
		out = append(out, q[1])
	}
	sort.Strings(out)
	return out
}

// TestWebLiveKeyContract — AC-019: events.go and the app.js list agree on the
// key set, and every live area a rendered screen declares is one of those
// keys, so no area is left waiting for an event the server never sends.
func TestWebLiveKeyContract(t *testing.T) {
	var serverKeys []string
	for k := range watchMap {
		serverKeys = append(serverKeys, k)
	}
	sort.Strings(serverKeys)
	clientKeys := m9AppJSEvents(t)
	if strings.Join(serverKeys, ",") != strings.Join(clientKeys, ",") {
		t.Fatalf("events.go watch map keys %v and the app.js EVENTS list %v disagree", serverKeys, clientKeys)
	}
	known := map[string]bool{}
	for _, k := range clientKeys {
		known[k] = true
	}
	if !known["factory"] {
		t.Errorf("the live key set %v carries no \"factory\" area", clientKeys)
	}

	liveRe := regexp.MustCompile(`data-live="([^"]+)"`)
	root := t.TempDir()
	for _, path := range []string{"/", "/factory", "/specs", "/monitor", "/todo"} {
		for _, m := range liveRe.FindAllStringSubmatch(m9ScreenBody(t, root, path), -1) {
			if !known[m[1]] {
				t.Errorf("%s declares data-live=%q, which is not in the live key set %v", path, m[1], clientKeys)
			}
		}
	}
}

// TestTodoScreenStillLive — AC-019: the todo queue keeps its live updates. The
// Todo screen and the overview todo card carry the renamed live area, and the
// watched queue directory still resolves to that same event, so a queue write
// still refreshes both.
func TestTodoScreenStillLive(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"/todo", "/"} {
		body := m9ScreenBody(t, root, path)
		if !strings.Contains(body, `data-live="factory"`) {
			t.Errorf("GET %s carries no data-live=\"factory\" area for the todo queue", path)
		}
	}
	todo := m9ScreenBody(t, root, "/todo")
	marker := strings.Index(todo, `data-live="factory"`)
	if marker < 0 || !strings.Contains(todo[marker:], "todo.title") {
		t.Error("the todo section does not sit inside the renamed live area")
	}
	if got := resolvedWatchPaths(root)[filepath.Join(root, ".moai", "state", "todo")]; got != "factory" {
		t.Errorf("the todo queue directory resolves to event %q, want \"factory\"", got)
	}
}
