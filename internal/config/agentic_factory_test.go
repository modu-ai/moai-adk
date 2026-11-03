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

// collabBody builds a workflow.yaml whose agentic_factory.collaboration block is inner.
func collabBody(inner string) string {
	var b strings.Builder
	b.WriteString("workflow:\n  agentic_factory:\n    collaboration:\n")
	for _, l := range strings.Split(strings.TrimRight(inner, "\n"), "\n") {
		if l != "" {
			b.WriteString("      " + l + "\n")
		}
	}
	return b.String()
}

func loadCollab(t *testing.T, body string) AgenticFactoryConfig {
	t.Helper()
	t.Setenv(EnvConfigDir, "")
	root := t.TempDir()
	if body != "" {
		writeAgenticFactoryWorkflow(t, root, body)
	}
	cfg, err := LoadAgenticFactory(root)
	if err != nil {
		t.Fatalf("LoadAgenticFactory: %v", err)
	}
	return cfg
}

func TestAgenticFactoryCollaborationDefaults(t *testing.T) {
	cases := map[string]string{
		"no file":          "",
		"empty block":      "workflow:\n  agentic_factory:\n",
		"null collab":      "workflow:\n  agentic_factory:\n    collaboration:\n",
		"empty map":        "workflow:\n  agentic_factory:\n    collaboration: {}\n",
		"null list":        collabBody("allowed_repositories:\n"),
		"empty list":       collabBody("allowed_repositories: []\n"),
		"null permissions": collabBody("permissions:\n"),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			c := loadCollab(t, body).Collaboration()
			if c.Enabled || c.EnabledExplicit || len(c.AllowedRepositoriesRaw) != 0 || len(c.AllowedRepositories("github")) != 0 || c.EditBody || c.Close || c.Labels || c.Assignees {
				t.Fatalf("defaults wrong: %+v", c)
			}
		})
	}
}

func TestAgenticFactoryCollaborationExplicitFalse(t *testing.T) {
	c := loadCollab(t, collabBody("enabled: false\n")).Collaboration()
	if c.Enabled || !c.EnabledExplicit {
		t.Fatalf("explicit false: %+v", c)
	}
}

func TestAgenticFactoryCollaborationParsedValues(t *testing.T) {
	c := loadCollab(t, collabBody("enabled: true\nallowed_repositories: [o/n, grp/sub/proj]\npermissions:\n  edit_body: true\n  labels: true\n")).Collaboration()
	want := CollaborationConfig{Enabled: true, EnabledExplicit: true, AllowedRepositoriesRaw: []string{"o/n", "grp/sub/proj"}, EditBody: true, Labels: true}
	if !reflect.DeepEqual(c, want) {
		t.Fatalf("got %+v, want %+v", c, want)
	}
	for k, f := range map[string]func(CollaborationConfig) bool{
		"edit_body": func(c CollaborationConfig) bool { return c.EditBody },
		"close":     func(c CollaborationConfig) bool { return c.Close },
		"labels":    func(c CollaborationConfig) bool { return c.Labels },
		"assignees": func(c CollaborationConfig) bool { return c.Assignees },
	} {
		got := loadCollab(t, collabBody("permissions:\n  "+k+": true\n")).Collaboration()
		if !f(got) || got.Enabled {
			t.Errorf("%s alone: %+v", k, got)
		}
	}
}

func TestNormalizeRepositoryPath(t *testing.T) {
	for in, want := range map[string]string{"Owner/Name": "owner/name", "GRP/Sub/Proj": "grp/sub/proj", "a/b": "a/b", "": ""} {
		if got := NormalizeRepositoryPath("github", in); got != want {
			t.Errorf("github: NormalizeRepositoryPath(%q) = %q, want %q", in, got, want)
		}
	}
	// ASCII only: non-ASCII letters are untouched.
	if got := NormalizeRepositoryPath("github", "Ünï/ABC"); got != "Ünï/abc" {
		t.Errorf("non-ASCII altered: %q", got)
	}
	// gitlab, unknown and empty provider: unchanged (exact match).
	for _, p := range []string{"gitlab", "", "bitbucket", "GitHub"} {
		if got := NormalizeRepositoryPath(p, "Owner/Name"); got != "Owner/Name" {
			t.Errorf("provider %q lower-cased: %q", p, got)
		}
	}
}

func TestAgenticFactoryCollaborationRepositoriesDedupedByNormalizedForm(t *testing.T) {
	c := loadCollab(t, collabBody("allowed_repositories: [B/X, a/y, b/x, A/Y, c/z, B/x]\n")).Collaboration()
	want := []string{"B/X", "a/y", "c/z"}
	if got := c.AllowedRepositories("github"); !reflect.DeepEqual(got, want) {
		t.Fatalf("github: got %v, want %v", got, want)
	}
	wantExact := []string{"B/X", "a/y", "b/x", "A/Y", "c/z", "B/x"}
	for _, p := range []string{"gitlab", "", "other"} {
		if got := c.AllowedRepositories(p); !reflect.DeepEqual(got, wantExact) {
			t.Errorf("provider %q: got %v, want %v", p, got, wantExact)
		}
	}
	two := loadCollab(t, collabBody("allowed_repositories: [Group/Proj, group/proj]\n")).Collaboration()
	if got := two.AllowedRepositories("gitlab"); !reflect.DeepEqual(got, []string{"Group/Proj", "group/proj"}) {
		t.Fatalf("gitlab merged case variants: %v", got)
	}
	if got := two.AllowedRepositories("github"); !reflect.DeepEqual(got, []string{"Group/Proj"}) {
		t.Fatalf("github kept case variants: %v", got)
	}
}

func TestAgenticFactoryRepositoryCaseNoteKeepsProvenance(t *testing.T) {
	for _, part := range []string{
		"The name is not case sensitive.",
		"https://docs.github.com/en/rest/repos/repos",
		"2026-10-29",
		"uncollated (from a WebFetch summary, not compared with the original page)",
		"no official statement found on 2026-10-29; exact match only",
	} {
		if !strings.Contains(AgenticFactoryRepositoryCaseNote, part) {
			t.Errorf("case note lacks %q", part)
		}
	}
}

func TestAgenticFactoryCollaborationInvalidRepositoryEntries(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	bad := map[string]string{
		"wildcard star":     `"owner/*"`,
		"wildcard question": `"owner/na?e"`,
		"url scheme":        `"https://example.com/owner/name"`,
		"whitespace":        `"owner/na me"`,
		"leading space":     `" owner/name"`,
		"tab":               `"owner/\tname"`,
		"empty":             `""`,
		"leading slash":     `"/owner/name"`,
		"trailing slash":    `"owner/name/"`,
		"single segment":    `"owner"`,
		"empty segment":     `"owner//name"`,
		"upper wildcard":    `"OWNER/*"`,
		"number entry":      `5`,
		"bool entry":        `true`,
		"list entry":        `[a/b]`,
		"map entry":         `{a: b}`,
		"null entry":        `null`,
	}
	for name, v := range bad {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			p := writeAgenticFactoryWorkflow(t, root, collabBody("allowed_repositories:\n  - good/one\n  - "+v+"\n"))
			cfg, err := LoadAgenticFactory(root)
			if err == nil {
				t.Fatalf("accepted %s: %+v", v, cfg)
			}
			if !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("not ErrInvalidConfig: %v", err)
			}
			for _, part := range []string{"workflow.agentic_factory.collaboration.allowed_repositories", p} {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("error %q lacks %q", err.Error(), part)
				}
			}
			if !reflect.DeepEqual(cfg, AgenticFactoryConfig{}) {
				t.Errorf("non-zero config on error: %+v", cfg)
			}
		})
	}
}

func TestAgenticFactoryCollaborationNonListRepositories(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	for name, v := range map[string]string{"string": "owner/name", "map": "{a: b}", "number": "3", "bool": "true"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			p := writeAgenticFactoryWorkflow(t, root, collabBody("allowed_repositories: "+v+"\n"))
			cfg, err := LoadAgenticFactory(root)
			if err == nil {
				t.Fatalf("accepted %s", v)
			}
			if !errors.Is(err, ErrInvalidConfig) || !strings.Contains(err.Error(), "workflow.agentic_factory.collaboration.allowed_repositories") || !strings.Contains(err.Error(), p) {
				t.Errorf("bad error: %v", err)
			}
			if !reflect.DeepEqual(cfg, AgenticFactoryConfig{}) {
				t.Errorf("non-zero config on error: %+v", cfg)
			}
		})
	}
}

func TestAgenticFactoryCollaborationNonBoolRejected(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	keys := map[string]string{
		"workflow.agentic_factory.collaboration.enabled":               "enabled: %s\n",
		"workflow.agentic_factory.collaboration.permissions.edit_body": "permissions:\n  edit_body: %s\n",
		"workflow.agentic_factory.collaboration.permissions.close":     "permissions:\n  close: %s\n",
		"workflow.agentic_factory.collaboration.permissions.labels":    "permissions:\n  labels: %s\n",
		"workflow.agentic_factory.collaboration.permissions.assignees": "permissions:\n  assignees: %s\n",
	}
	vals := map[string]string{"string true": `"true"`, "int": "1", "word": "yes-please", "empty string": `""`, "list": "[true]", "map": "{a: b}"}
	for key, tmpl := range keys {
		for vn, v := range vals {
			t.Run(key+"/"+vn, func(t *testing.T) {
				root := t.TempDir()
				p := writeAgenticFactoryWorkflow(t, root, collabBody(strings.Replace(tmpl, "%s", v, 1)))
				cfg, err := LoadAgenticFactory(root)
				if err == nil {
					t.Fatalf("accepted %s", v)
				}
				if !errors.Is(err, ErrInvalidConfig) {
					t.Errorf("not ErrInvalidConfig: %v", err)
				}
				for _, part := range []string{key, "bool", p} {
					if !strings.Contains(err.Error(), part) {
						t.Errorf("error %q lacks %q", err.Error(), part)
					}
				}
				if !reflect.DeepEqual(cfg, AgenticFactoryConfig{}) {
					t.Errorf("non-zero config on error: %+v", cfg)
				}
			})
		}
	}
}

// The error names the value kind, never the value itself.
func TestAgenticFactoryCollaborationErrorOmitsValue(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	root := t.TempDir()
	writeAgenticFactoryWorkflow(t, root, collabBody("allowed_repositories: [\"s3cr3t-token/*\"]\n"))
	_, err := LoadAgenticFactory(root)
	if err == nil {
		t.Fatal("accepted wildcard")
	}
	if strings.Contains(err.Error(), "s3cr3t-token") {
		t.Fatalf("error leaks value: %v", err)
	}
}

func TestAgenticFactoryCollaborationReaderNeverWrites(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	root := t.TempDir()
	body := "# c\n" + collabBody("enabled: true   # mine\nallowed_repositories: [a/b, A/B]\npermissions:\n  close: true\n")
	p := writeAgenticFactoryWorkflow(t, root, body)
	if _, err := LoadAgenticFactory(root); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != body {
		t.Fatalf("file changed:\n%s", after)
	}
}

func TestAgenticFactoryCollaborationCoexistsWithOtherKeys(t *testing.T) {
	cfg := loadCollab(t, "workflow:\n  agentic_factory:\n    browser:\n      selected: aside\n    codex:\n      enabled: true\n    pr_queue:\n      wip_limit: 4\n    collaboration:\n      enabled: true\n")
	if res, _ := cfg.EffectiveBrowser(); res.Browser != "aside" {
		t.Fatalf("browser %+v", res)
	}
	if en, ex := cfg.CodexOptIn(); !en || !ex {
		t.Fatal("codex lost")
	}
	if l, ok := cfg.WIPLimit(); l != 4 || !ok {
		t.Fatal("wip lost")
	}
	if !cfg.Collaboration().Enabled {
		t.Fatal("collab lost")
	}
}
