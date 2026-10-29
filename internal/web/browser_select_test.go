package web

// browser_select_test.go — the read-only browser selection surface
// (SPEC-AGENTIC-FACTORY-001 REQ-AF-016, Issue #1948).
//
// Defect answer: these tests catch the monitor showing a browser the operator
// never chose (chrome painted as a saved selection, or a fallback shown after a
// block), a blocked selection that loses its resume condition, a page render
// that runs the egolite connection check, and a surface that offers anything
// other than the three choices or accepts a write.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// webFakeProbe answers LookPath from a set and records every Run, so a test can
// prove the page never executes a browser command.
type webFakeProbe struct {
	installed map[string]bool
	ran       []string
}

func (p *webFakeProbe) LookPath(name string) error {
	if p.installed[name] {
		return nil
	}
	return errors.New("not found")
}

func (p *webFakeProbe) Run(_ context.Context, name string, args ...string) (int, string, error) {
	p.ran = append(p.ran, name+" "+strings.Join(args, " "))
	return 0, "aside 1", nil
}

func browserApp(t *testing.T, selected string, readErr error, installed ...string) (*app, *webFakeProbe) {
	t.Helper()
	probe := &webFakeProbe{installed: map[string]bool{}}
	for _, n := range installed {
		probe.installed[n] = true
	}
	a := newApp(Config{ProjectRoot: t.TempDir()})
	a.browserSelected = func(string) (string, error) { return selected, readErr }
	a.browserProbe = probe
	return a, probe
}

func TestBrowserSelection_StatePerCase(t *testing.T) {
	cases := []struct {
		name          string
		selected      string
		readErr       error
		installed     []string
		wantSelected  string
		wantEffective string
		wantReason    string
		wantNote      string
	}{
		{name: "no selection shows chrome as effective, not as selected", wantEffective: "chrome"},
		{name: "aside selected and installed", selected: "aside", installed: []string{"aside"}, wantSelected: "aside", wantEffective: "aside"},
		{name: "egolite selected and missing blocks without fallback", selected: "egolite", installed: []string{"aside"}, wantSelected: "egolite", wantReason: "not-installed"},
		{name: "egolite installed passes the preflight", selected: "egolite", installed: []string{"ego-browser"}, wantSelected: "egolite", wantEffective: "egolite"},
		{name: "unknown value is reported, nothing is chosen", selected: "firefox", wantNote: "firefox"},
		{name: "unreadable selection is reported, nothing is chosen", readErr: errors.New("boom"), wantNote: "boom"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, probe := browserApp(t, tc.selected, tc.readErr, tc.installed...)
			vm := a.buildBrowserSelection(context.Background())

			if len(vm.Choices) != 3 || vm.Choices[0].ID != "egolite" || vm.Choices[1].ID != "aside" || vm.Choices[2].ID != "chrome" {
				t.Fatalf("choices = %+v, want exactly egolite, aside, chrome", vm.Choices)
			}
			var marked string
			for _, c := range vm.Choices {
				if c.Selected {
					marked = c.ID
				}
			}
			if marked != tc.wantSelected {
				t.Fatalf("marked selected = %q, want %q", marked, tc.wantSelected)
			}
			if vm.Effective != tc.wantEffective {
				t.Fatalf("effective = %q, want %q", vm.Effective, tc.wantEffective)
			}
			if vm.BlockReason != tc.wantReason {
				t.Fatalf("block reason = %q, want %q", vm.BlockReason, tc.wantReason)
			}
			if tc.wantReason != "" && vm.Resume == "" {
				t.Fatalf("blocked selection lost its resume condition")
			}
			if tc.wantNote != "" && !strings.Contains(vm.Note, tc.wantNote) {
				t.Fatalf("note = %q, want it to mention %q", vm.Note, tc.wantNote)
			}
			// Only the install probe `aside --version` may run; the egolite
			// connection check (`ego-browser nodejs`) must never run on a render.
			for _, c := range probe.ran {
				if c != "aside --version" {
					t.Fatalf("page build ran %q; only the aside version probe is allowed", c)
				}
			}
		})
	}
}

func TestMonitorRendersBrowserSelection_ReadOnly(t *testing.T) {
	a, _ := browserApp(t, "egolite", nil)
	m, err := a.buildMonitor(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	html := renderTempl(t, Monitor(tg1ShellVM(), m))

	for _, want := range []string{`id="browser-selection"`, "egolite", "aside", "chrome", "not-installed"} {
		if !strings.Contains(html, want) {
			t.Errorf("monitor browser panel missing %q:\n%s", want, html)
		}
	}
	// Display only: saving the choice belongs to the settings owner, and a
	// named control here would submit with the settings form.
	panel := html[strings.Index(html, `id="browser-selection"`):]
	if end := strings.Index(panel, "</section>"); end > 0 {
		panel = panel[:end]
	}
	for _, bad := range []string{"<input", "<select", "<button", ` name="`} {
		if strings.Contains(panel, bad) {
			t.Errorf("browser panel renders a control %q; it must be display-only", bad)
		}
	}
}

func TestBrowserAPI_ReadOnlyJSON(t *testing.T) {
	a, _ := browserApp(t, "aside", nil, "aside")

	get := httptest.NewRequest(http.MethodGet, "/api/browser", nil)
	get.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, get)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/browser = %d", rec.Code)
	}
	var vm BrowserSelectionVM
	if err := json.Unmarshal(rec.Body.Bytes(), &vm); err != nil {
		t.Fatalf("response is not JSON: %v\n%s", err, rec.Body.String())
	}
	if vm.Effective != "aside" || len(vm.Choices) != 3 {
		t.Fatalf("api vm = %+v", vm)
	}

	post := httptest.NewRequest(http.MethodPost, "/api/browser", strings.NewReader("selected=chrome"))
	post.Host = "127.0.0.1:8080"
	post.Header.Set("Sec-Fetch-Site", "same-origin")
	rec = httptest.NewRecorder()
	a.routes().ServeHTTP(rec, post)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/browser = %d, want 405", rec.Code)
	}
}
