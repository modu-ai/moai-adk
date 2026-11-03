package browser

import (
	"context"
	"errors"
	"testing"
)

// fakeProbe records every external command the resolver issues, so a test can
// assert that no browser other than the selected one is ever touched.
type fakeProbe struct {
	installed map[string]bool
	runs      map[string]fakeRun
	lookups   []string
	ran       []string
}

type fakeRun struct {
	code   int
	output string
	err    error
}

func (f *fakeProbe) LookPath(name string) error {
	f.lookups = append(f.lookups, name)
	if f.installed[name] {
		return nil
	}
	return errors.New("not found")
}

func (f *fakeProbe) Run(_ context.Context, name string, args ...string) (int, string, error) {
	key := name
	for _, a := range args {
		key += " " + a
	}
	f.ran = append(f.ran, key)
	r, ok := f.runs[key]
	if !ok {
		return 127, "", errors.New("unexpected command: " + key)
	}
	return r.code, r.output, r.err
}

const (
	egoCheck   = "ego-browser nodejs -e console.log('ego-browser ready')"
	asideCheck = "aside --version"
)

func TestResolve_ExplicitWinsNoFallback(t *testing.T) {
	cases := []struct {
		name       string
		selected   string
		probe      *fakeProbe
		wantUse    Choice
		wantReason Reason
		wantRan    []string
		wantLook   []string
	}{
		{
			name:     "a no selection uses chrome without probing",
			selected: "",
			probe:    &fakeProbe{},
			wantUse:  Chrome,
		},
		{
			name:     "b aside selected and installed",
			selected: "aside",
			probe: &fakeProbe{runs: map[string]fakeRun{
				asideCheck: {code: 0, output: "aside 1.26.1023.1115"},
			}},
			wantUse: Aside,
			wantRan: []string{asideCheck},
		},
		{
			name:       "c egolite selected and not installed",
			selected:   "egolite",
			probe:      &fakeProbe{installed: map[string]bool{"aside": true}},
			wantReason: ReasonNotInstalled,
			wantLook:   []string{"ego-browser"},
		},
		{
			name:     "d egolite installed but functional check fails",
			selected: "egolite",
			probe: &fakeProbe{
				installed: map[string]bool{"ego-browser": true},
				runs:      map[string]fakeRun{egoCheck: {code: 1, output: "cannot connect to ego lite"}},
			},
			wantReason: ReasonConnectionFailed,
			wantLook:   []string{"ego-browser"},
			wantRan:    []string{egoCheck},
		},
		{
			name:     "d2 egolite check exits zero without the ready line",
			selected: "egolite",
			probe: &fakeProbe{
				installed: map[string]bool{"ego-browser": true},
				runs:      map[string]fakeRun{egoCheck: {code: 0, output: "something else"}},
			},
			wantReason: ReasonConnectionFailed,
			wantLook:   []string{"ego-browser"},
			wantRan:    []string{egoCheck},
		},
		{
			name:     "d3 egolite ready",
			selected: "egolite",
			probe: &fakeProbe{
				installed: map[string]bool{"ego-browser": true},
				runs:      map[string]fakeRun{egoCheck: {code: 0, output: "ego-browser ready\n"}},
			},
			wantUse:  Egolite,
			wantLook: []string{"ego-browser"},
			wantRan:  []string{egoCheck},
		},
		{
			name:     "g aside selected and not installed never falls back",
			selected: "aside",
			probe: &fakeProbe{
				installed: map[string]bool{"ego-browser": true},
				runs:      map[string]fakeRun{asideCheck: {code: 127, err: errors.New("not found")}},
			},
			wantReason: ReasonNotInstalled,
			wantRan:    []string{asideCheck},
		},
		{
			name:     "chrome explicit needs no probe",
			selected: "chrome",
			probe:    &fakeProbe{},
			wantUse:  Chrome,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Resolve(context.Background(), tc.selected, tc.probe)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Reason != tc.wantReason {
				t.Fatalf("reason = %q, want %q", got.Reason, tc.wantReason)
			}
			if got.Blocked() != (tc.wantReason != "") {
				t.Fatalf("Blocked() = %v with reason %q", got.Blocked(), got.Reason)
			}
			if got.Use != tc.wantUse {
				t.Fatalf("use = %q, want %q", got.Use, tc.wantUse)
			}
			if got.Blocked() && got.Resume == "" {
				t.Fatalf("blocked resolution carries no resume condition")
			}
			if got.Blocked() && got.Use != "" {
				t.Fatalf("blocked resolution names a browser to use: %q", got.Use)
			}
			if got.Selected != tc.selected {
				t.Fatalf("selected = %q, want it preserved as %q", got.Selected, tc.selected)
			}
			if !equal(tc.probe.ran, tc.wantRan) {
				t.Fatalf("commands run = %v, want %v", tc.probe.ran, tc.wantRan)
			}
			if !equal(tc.probe.lookups, tc.wantLook) {
				t.Fatalf("lookups = %v, want %v", tc.probe.lookups, tc.wantLook)
			}
		})
	}
}

func TestResolve_UnknownChoiceRejected(t *testing.T) {
	p := &fakeProbe{}
	_, err := Resolve(context.Background(), "firefox", p)
	if !errors.Is(err, ErrUnknownChoice) {
		t.Fatalf("err = %v, want ErrUnknownChoice", err)
	}
	if len(p.ran)+len(p.lookups) != 0 {
		t.Fatalf("probed for an invalid choice: %v %v", p.ran, p.lookups)
	}
}

func TestChoices_ExactlyThree(t *testing.T) {
	got := Choices()
	want := []Choice{Egolite, Aside, Chrome}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("Choices() = %v, want %v", got, want)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestPreflight_NeverConnects(t *testing.T) {
	p := &fakeProbe{installed: map[string]bool{"ego-browser": true}}
	got, err := Preflight(context.Background(), "egolite", p)
	if err != nil || got.Blocked() || got.Use != Egolite {
		t.Fatalf("Preflight installed egolite = %+v, %v", got, err)
	}
	if len(p.ran) != 0 {
		t.Fatalf("Preflight ran commands: %v", p.ran)
	}
	got, _ = Preflight(context.Background(), "egolite", &fakeProbe{})
	if got.Reason != ReasonNotInstalled {
		t.Fatalf("Preflight missing egolite reason = %q", got.Reason)
	}
}
