package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// writeAgenticFactoryWorkflow writes body to <root>/.moai/config/sections/workflow.yaml
// and returns the file path.
func writeAgenticFactoryWorkflow(t *testing.T, root, body string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	p := filepath.Join(dir, "workflow.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func strPtr(s string) *string { return &s }

func TestAgenticFactoryBrowserUnsetIsChromeNotExplicit(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	cases := map[string]string{
		"no file":       "",
		"no block":      "workflow:\n  factory_lease:\n    duration: 20m\n",
		"null":          "workflow:\n  agentic_factory:\n    browser:\n      selected:\n",
		"empty block":   "workflow:\n  agentic_factory:\n",
		"only codex":    "workflow:\n  agentic_factory:\n    codex:\n      enabled: true\n",
		"empty browser": "workflow:\n  agentic_factory:\n    browser:\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			if body != "" {
				writeAgenticFactoryWorkflow(t, root, body)
			}
			cfg, err := LoadAgenticFactory(root)
			if err != nil {
				t.Fatalf("LoadAgenticFactory: %v", err)
			}
			if cfg.BrowserSelected != nil {
				t.Fatalf("BrowserSelected = %q, want nil (unset)", *cfg.BrowserSelected)
			}
			got, err := cfg.EffectiveBrowser()
			if err != nil {
				t.Fatalf("EffectiveBrowser: %v", err)
			}
			want := BrowserResolution{Browser: "chrome", Explicit: false}
			if got != want {
				t.Fatalf("EffectiveBrowser = %+v, want %+v", got, want)
			}
		})
	}
}

func TestAgenticFactoryBrowserExplicitValueWins(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	for _, v := range []string{"egolite", "aside", "chrome"} {
		t.Run(v, func(t *testing.T) {
			root := t.TempDir()
			writeAgenticFactoryWorkflow(t, root, "workflow:\n  agentic_factory:\n    browser:\n      selected: "+v+"\n")
			cfg, err := LoadAgenticFactory(root)
			if err != nil {
				t.Fatalf("LoadAgenticFactory: %v", err)
			}
			got, err := cfg.EffectiveBrowser()
			if err != nil {
				t.Fatalf("EffectiveBrowser: %v", err)
			}
			if want := (BrowserResolution{Browser: v, Explicit: true}); got != want {
				t.Fatalf("EffectiveBrowser = %+v, want %+v", got, want)
			}
		})
	}
}

// An explicit selection is returned as written, never swapped for another
// browser: the resolver has no availability input, so it cannot fall back.
func TestResolveAgenticFactoryBrowserNeverReturnsOtherThanSelection(t *testing.T) {
	for _, v := range []string{"egolite", "aside", "chrome"} {
		got, err := ResolveAgenticFactoryBrowser(strPtr(v))
		if err != nil {
			t.Fatalf("Resolve(%q): %v", v, err)
		}
		if got.Browser != v || !got.Explicit {
			t.Fatalf("Resolve(%q) = %+v, want selection kept with Explicit=true", v, got)
		}
	}
	got, err := ResolveAgenticFactoryBrowser(nil)
	if err != nil || got != (BrowserResolution{Browser: "chrome", Explicit: false}) {
		t.Fatalf("Resolve(nil) = %+v, %v", got, err)
	}
}

func TestAgenticFactoryBrowserInvalidIsRejected(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	bad := map[string]string{
		"unknown name": "firefox",
		"wrong case":   "Chrome",
		"empty string": `""`,
		"padded":       `" aside"`,
		"number":       "3",
		"bool":         "true",
		"list":         "[aside]",
		"mapping":      "{a: b}",
	}
	for name, v := range bad {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			p := writeAgenticFactoryWorkflow(t, root, "workflow:\n  agentic_factory:\n    browser:\n      selected: "+v+"\n")
			cfg, err := LoadAgenticFactory(root)
			if err == nil {
				t.Fatalf("LoadAgenticFactory accepted %s; cfg=%+v", v, cfg)
			}
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("error does not wrap ErrInvalidConfig: %v", err)
			}
			msg := err.Error()
			for _, part := range []string{"workflow.agentic_factory.browser.selected", "egolite", "aside", "chrome", p} {
				if !strings.Contains(msg, part) {
					t.Errorf("error %q lacks %q", msg, part)
				}
			}
			if cfg.BrowserSelected != nil {
				t.Errorf("rejected value leaked into config: %q", *cfg.BrowserSelected)
			}
		})
	}
}

func TestResolveAgenticFactoryBrowserInvalidNoFallback(t *testing.T) {
	got, err := ResolveAgenticFactoryBrowser(strPtr("firefox"))
	if err == nil {
		t.Fatal("expected error for unknown browser")
	}
	if got != (BrowserResolution{}) {
		t.Fatalf("rejected value produced a resolution %+v, want zero value (no fallback)", got)
	}
}

func TestAgenticFactoryUnparseableFileIsError(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	root := t.TempDir()
	writeAgenticFactoryWorkflow(t, root, "workflow: [unclosed\n")
	if _, err := LoadAgenticFactory(root); err == nil || !errors.Is(err, ErrInvalidYAML) {
		t.Fatalf("err = %v, want ErrInvalidYAML (an unreadable file must not silently become the chrome default)", err)
	}
}

// The reader never writes: loading leaves the user's file byte-identical, and
// the selection reads back unchanged across repeated loads and unrelated edits.
func TestAgenticFactorySelectionSurvivesRoundTripByteForByte(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	root := t.TempDir()
	body := "# user comment\nworkflow:\n  factory_lease:\n    duration: 20m\n  agentic_factory:\n    browser:\n      selected: aside   # my choice\n    codex:\n      enabled: false\n  unrelated_user_key: keep-me\n"
	p := writeAgenticFactoryWorkflow(t, root, body)

	for i := 0; i < 2; i++ {
		cfg, err := LoadAgenticFactory(root)
		if err != nil {
			t.Fatalf("load %d: %v", i, err)
		}
		if res, _ := cfg.EffectiveBrowser(); res != (BrowserResolution{Browser: "aside", Explicit: true}) {
			t.Fatalf("load %d: %+v", i, res)
		}
	}
	after, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != body {
		t.Fatalf("file changed by load:\n%s", after)
	}

	// An unrelated edit by the user leaves the selection intact.
	edited := strings.Replace(body, "duration: 20m", "duration: 30m", 1)
	if err := os.WriteFile(p, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadAgenticFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if res, _ := cfg.EffectiveBrowser(); res.Browser != "aside" || !res.Explicit {
		t.Fatalf("selection lost after unrelated edit: %+v", res)
	}
}

func TestAgenticFactoryCodexOptIn(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	cases := []struct {
		name         string
		body         string
		wantEnabled  bool
		wantExplicit bool
	}{
		{"no file", "", false, false},
		{"unset", "workflow:\n  agentic_factory:\n    browser:\n      selected: chrome\n", false, false},
		{"null", "workflow:\n  agentic_factory:\n    codex:\n      enabled:\n", false, false},
		{"true", "workflow:\n  agentic_factory:\n    codex:\n      enabled: true\n", true, true},
		{"false", "workflow:\n  agentic_factory:\n    codex:\n      enabled: false\n", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.body != "" {
				writeAgenticFactoryWorkflow(t, root, tc.body)
			}
			cfg, err := LoadAgenticFactory(root)
			if err != nil {
				t.Fatal(err)
			}
			en, ex := cfg.CodexOptIn()
			if en != tc.wantEnabled || ex != tc.wantExplicit {
				t.Fatalf("CodexOptIn = (%v,%v), want (%v,%v)", en, ex, tc.wantEnabled, tc.wantExplicit)
			}
		})
	}
}

func TestAgenticFactoryCodexNonBoolRejected(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	for name, v := range map[string]string{
		"string true": `"true"`, "int": "1", "word": "yes-please", "empty string": `""`, "list": "[true]",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			p := writeAgenticFactoryWorkflow(t, root, "workflow:\n  agentic_factory:\n    codex:\n      enabled: "+v+"\n")
			cfg, err := LoadAgenticFactory(root)
			if err == nil {
				t.Fatalf("accepted %s", v)
			}
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("not ErrInvalidConfig: %v", err)
			}
			for _, part := range []string{"workflow.agentic_factory.codex.enabled", "bool", p} {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("error %q lacks %q", err.Error(), part)
				}
			}
			if en, ex := cfg.CodexOptIn(); en || ex {
				t.Errorf("rejected value produced opt-in (%v,%v)", en, ex)
			}
		})
	}
}

func TestAgenticFactoryWIPLimitUnsetIsNotUnlimited(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	cases := map[string]string{
		"no file":     "",
		"no block":    "workflow:\n  factory_lease:\n    duration: 20m\n",
		"empty block": "workflow:\n  agentic_factory:\n",
		"empty queue": "workflow:\n  agentic_factory:\n    pr_queue:\n",
		"null":        "workflow:\n  agentic_factory:\n    pr_queue:\n      wip_limit:\n",
		"explicit nl": "workflow:\n  agentic_factory:\n    pr_queue:\n      wip_limit: null\n",
		"only codex":  "workflow:\n  agentic_factory:\n    codex:\n      enabled: true\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			if body != "" {
				writeAgenticFactoryWorkflow(t, root, body)
			}
			cfg, err := LoadAgenticFactory(root)
			if err != nil {
				t.Fatalf("LoadAgenticFactory: %v", err)
			}
			if cfg.PRQueueWIPLimit != nil {
				t.Fatalf("PRQueueWIPLimit = %d, want nil (unset)", *cfg.PRQueueWIPLimit)
			}
			if limit, set := cfg.WIPLimit(); limit != 0 || set {
				t.Fatalf("WIPLimit = (%d,%v), want (0,false): unset must not invent a number", limit, set)
			}
		})
	}
}

func TestAgenticFactoryWIPLimitValid(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	for _, want := range []int{1, 7, 1000000} {
		t.Run(strconv.Itoa(want), func(t *testing.T) {
			root := t.TempDir()
			writeAgenticFactoryWorkflow(t, root, "workflow:\n  agentic_factory:\n    pr_queue:\n      wip_limit: "+strconv.Itoa(want)+"\n")
			cfg, err := LoadAgenticFactory(root)
			if err != nil {
				t.Fatalf("LoadAgenticFactory: %v", err)
			}
			if limit, set := cfg.WIPLimit(); limit != want || !set {
				t.Fatalf("WIPLimit = (%d,%v), want (%d,true)", limit, set, want)
			}
		})
	}
}

func TestAgenticFactoryWIPLimitInvalidIsRejected(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	bad := map[string]string{
		"zero":         "0",
		"negative":     "-1",
		"float":        "2.5",
		"string":       `"3"`,
		"bool":         "true",
		"list":         "[3]",
		"mapping":      "{a: 3}",
		"overflow":     "99999999999999999999999999",
		"empty string": `""`,
	}
	for name, v := range bad {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			p := writeAgenticFactoryWorkflow(t, root, "workflow:\n  agentic_factory:\n    pr_queue:\n      wip_limit: "+v+"\n")
			cfg, err := LoadAgenticFactory(root)
			if err == nil {
				t.Fatalf("LoadAgenticFactory accepted %s; cfg=%+v", v, cfg)
			}
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("error does not wrap ErrInvalidConfig: %v", err)
			}
			for _, part := range []string{"workflow.agentic_factory.pr_queue.wip_limit", ">= 1", p} {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("error %q lacks %q", err.Error(), part)
				}
			}
			if limit, set := cfg.WIPLimit(); limit != 0 || set {
				t.Errorf("rejected value leaked: WIPLimit = (%d,%v)", limit, set)
			}
		})
	}
}

// Existing defaults are untouched: the Loader's view of the workflow section is
// identical with and without the new block, and two canary defaults hold.
func TestAgenticFactoryLeavesExistingDefaultsUnchanged(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	base := "workflow:\n  factory_lease:\n    duration: 15m\n"
	with := base + "  agentic_factory:\n    browser:\n      selected: egolite\n    codex:\n      enabled: true\n"

	load := func(body string) *Config {
		root := t.TempDir()
		writeAgenticFactoryWorkflow(t, root, body)
		cfg, err := NewLoader().Load(filepath.Join(root, ".moai"))
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		return cfg
	}
	a, b := load(base), load(with)
	if !reflect.DeepEqual(a.Workflow, b.Workflow) {
		t.Fatal("adding workflow.agentic_factory changed the loaded Workflow section")
	}

	def := NewDefaultWorkflowConfig()
	if def.FactoryLease.Duration != "15m0s" || def.FactoryLease.RenewInterval != "5m0s" {
		t.Fatalf("canary defaults changed: %+v", def.FactoryLease)
	}

	// A config without the section yields the unset behaviour.
	cfg, err := LoadAgenticFactory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if en, ex := cfg.CodexOptIn(); en || ex || cfg.BrowserSelected != nil {
		t.Fatalf("missing section must read as unset: %+v", cfg)
	}
}
