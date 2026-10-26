package config

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// SPEC-JEV-LEDGER-001 (card t1762) M1: the workflow.jev.consumers config
// surface, its tolerant per-consumer decode, and the hook time-budget
// constants. These tests are deliberately NOT parallel: the loader tests swap
// the process-wide slog default, stdout and stderr to prove the resolver and
// the loader stay silent (REQ-JEVL-008).

// jevDecodeConsumer decodes one consumer body the way the loader does, through
// the type's own UnmarshalYAML.
func jevDecodeConsumer(t *testing.T, body string) JevConsumerConfig {
	t.Helper()
	var c JevConsumerConfig
	if err := yaml.Unmarshal([]byte(body), &c); err != nil {
		t.Fatalf("decoding %q must never fail: %v", body, err)
	}
	return c
}

// AC-JEVL-006: the mode enum is accepted, an absent or unrecognized value
// resolves to shadow, and the effective mode is shadow for every accepted value.
func TestResolveJevConsumer_Mode(t *testing.T) {
	cases := []struct {
		name           string
		body           string
		wantConfigured string
		wantWarnings   int
	}{
		{"shadow", "mode: shadow", "shadow", 0},
		{"advise", "mode: advise", "advise", 0},
		{"steer", "mode: steer", "steer", 0},
		{"confirm", "mode: confirm", "confirm", 0},
		{"absent", "threshold: 0.5", "", 0},
		{"bogus", "mode: bogus", "bogus", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveJevConsumer(jevDecodeConsumer(t, tc.body))
			if got.Effective != JevModeShadow {
				t.Errorf("effective mode = %q, want %q for every value in this SPEC", got.Effective, JevModeShadow)
			}
			if got.Configured != tc.wantConfigured {
				t.Errorf("configured mode = %q, want %q", got.Configured, tc.wantConfigured)
			}
			if len(got.Warnings) != tc.wantWarnings {
				t.Errorf("warnings = %v, want %d", got.Warnings, tc.wantWarnings)
			}
		})
	}
}

// A mode that is not a string scalar is unrecognized, never an effective mode.
func TestResolveJevConsumer_ModeShapes(t *testing.T) {
	for name, body := range map[string]string{
		"integer":  "mode: 5",
		"mapping":  "mode: {a: 1}",
		"sequence": "mode: [advise]",
		"bool":     "mode: true",
	} {
		t.Run(name, func(t *testing.T) {
			got := ResolveJevConsumer(jevDecodeConsumer(t, body))
			if got.Effective != JevModeShadow || got.Configured != "" || len(got.Warnings) != 1 {
				t.Errorf("got %+v, want shadow, empty configured value and one warning", got)
			}
		})
	}
	t.Run("null-is-absent", func(t *testing.T) {
		got := ResolveJevConsumer(jevDecodeConsumer(t, "mode: ~"))
		if got.Configured != "" || len(got.Warnings) != 0 {
			t.Errorf("a null mode must read as absent, got %+v", got)
		}
	})
}

// AC-JEVL-007 (resolver half): the threshold lives in the closed interval
// [0, 1], 0 is distinguishable from absent, and anything else resolves to
// absent with a warning returned as data.
func TestResolveJevConsumer_Threshold(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	cases := []struct {
		name         string
		body         string
		want         *float64
		wantWarnings int
	}{
		{"zero", "threshold: 0", f(0), 0},
		{"one", "threshold: 1", f(1), 0},
		{"half", "threshold: 0.5", f(0.5), 0},
		{"negative", "threshold: -0.1", nil, 1},
		{"above-one", "threshold: 1.5", nil, 1},
		{"string", `threshold: "high"`, nil, 1},
		{"numeric-string", `threshold: "0.5"`, nil, 1},
		{"mapping", "threshold: {a: 1}", nil, 1},
		{"bool", "threshold: true", nil, 1},
		{"nan", "threshold: .nan", nil, 1},
		{"absent", "mode: shadow", nil, 0},
		{"null-is-absent", "threshold: ~", nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveJevConsumer(jevDecodeConsumer(t, tc.body))
			switch {
			case tc.want == nil && got.Threshold != nil:
				t.Errorf("threshold = %v, want absent", *got.Threshold)
			case tc.want != nil && got.Threshold == nil:
				t.Errorf("threshold absent, want %v", *tc.want)
			case tc.want != nil && *got.Threshold != *tc.want:
				t.Errorf("threshold = %v, want %v", *got.Threshold, *tc.want)
			}
			if len(got.Warnings) != tc.wantWarnings {
				t.Errorf("warnings = %v, want %d", got.Warnings, tc.wantWarnings)
			}
		})
	}
}

// A consumer entry that is not a mapping resolves to shadow with one warning;
// a bare `name:` (null) is an empty entry, not a defect.
func TestResolveJevConsumer_EntryShape(t *testing.T) {
	got := ResolveJevConsumer(jevDecodeConsumer(t, "42"))
	if got.Effective != JevModeShadow || len(got.Warnings) != 1 {
		t.Errorf("scalar entry: got %+v, want shadow and one warning", got)
	}
	got = ResolveJevConsumer(jevDecodeConsumer(t, "[a, b]"))
	if len(got.Warnings) != 1 {
		t.Errorf("sequence entry: got %+v, want one warning", got)
	}
	got = ResolveJevConsumer(jevDecodeConsumer(t, "~"))
	if len(got.Warnings) != 0 || got.Threshold != nil {
		t.Errorf("null entry: got %+v, want an empty entry", got)
	}
}

// A key that is not a name is skipped with a section-level warning; the other
// consumers still load.
func TestJevConsumers_NonScalarKey(t *testing.T) {
	var wrapper struct {
		Consumers JevConsumers `yaml:"consumers"`
	}
	src := "consumers:\n  ? [a, b]\n  : {mode: advise}\n  ok: {mode: steer}\n"
	if err := yaml.Unmarshal([]byte(src), &wrapper); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(wrapper.Consumers.Entries) != 1 || len(wrapper.Consumers.SectionWarnings()) != 1 {
		t.Errorf("entries=%v warnings=%v, want one entry and one section warning", wrapper.Consumers.Entries, wrapper.Consumers.Warnings)
	}
	if got, ok := wrapper.Consumers.Resolve("ok"); !ok || got.Configured != "steer" {
		t.Errorf("the valid consumer did not survive its neighbour: %+v (found=%v)", got, ok)
	}
}

// The decoders follow aliases: a consumer that aliases another consumer's
// mapping resolves like the anchor does.
func TestJevConsumers_AliasEntry(t *testing.T) {
	var wrapper struct {
		Consumers JevConsumers `yaml:"consumers"`
	}
	src := "consumers:\n  a: &base {mode: advise, threshold: 0.25}\n  b: *base\n"
	if err := yaml.Unmarshal([]byte(src), &wrapper); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	b, ok := wrapper.Consumers.Resolve("b")
	if !ok || b.Configured != "advise" || b.Threshold == nil || *b.Threshold != 0.25 {
		t.Errorf("aliased consumer resolved to %+v (found=%v)", b, ok)
	}
	if _, ok := wrapper.Consumers.Resolve("missing"); ok {
		t.Error("Resolve reported a consumer that is not configured")
	}
}

// jevCapture swaps the process-wide sinks a stray write could reach, runs fn,
// and returns everything that was written to any of them.
func jevCapture(t *testing.T, fn func()) (stdout, stderr string, records int) {
	t.Helper()
	oldOut, oldErr, oldLog, oldDefault := os.Stdout, os.Stderr, log.Writer(), slog.Default()

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	var wg sync.WaitGroup
	var outBuf, errBuf bytes.Buffer
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = io.Copy(&outBuf, outR) }()
	go func() { defer wg.Done(); _, _ = io.Copy(&errBuf, errR) }()

	counter := &jevCountingHandler{}
	os.Stdout, os.Stderr = outW, errW
	log.SetOutput(errW)
	slog.SetDefault(slog.New(counter))
	restore := func() {
		os.Stdout, os.Stderr = oldOut, oldErr
		log.SetOutput(oldLog)
		slog.SetDefault(oldDefault)
	}
	defer restore()

	fn()

	restore()
	_ = outW.Close()
	_ = errW.Close()
	wg.Wait()
	return outBuf.String(), errBuf.String(), counter.count()
}

type jevCountingHandler struct {
	mu sync.Mutex
	n  int
}

func (h *jevCountingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *jevCountingHandler) Handle(context.Context, slog.Record) error {
	h.mu.Lock()
	h.n++
	h.mu.Unlock()
	return nil
}
func (h *jevCountingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *jevCountingHandler) WithGroup(string) slog.Handler      { return h }
func (h *jevCountingHandler) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.n
}

// jevWriteWorkflow writes <root>/.moai/config/sections/workflow.yaml.
func jevWriteWorkflow(t *testing.T, root, body string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	return filepath.Join(root, ".moai")
}

// The control setting: non-default (the compiled default is false) and owned by
// a block that has no decoder tolerance of its own, so a section-level reset to
// defaults is visible on it.
const jevOtherSetting = "    commit_identity_guard:\n        enabled: true\n"

func jevAssertSettingsKept(t *testing.T, cfg *Config) {
	t.Helper()
	if !cfg.Workflow.Jev.Enabled {
		t.Error("workflow.jev.enabled reverted to its default: the section was reset")
	}
	if !cfg.Workflow.CommitIdentityGuard.Enabled {
		t.Error("another workflow setting reverted to its default: the section was reset")
	}
}

// AC-JEVL-007 (loader half): one malformed consumer is quarantined to that
// consumer. The rest of the workflow section, and every other consumer, still
// load, and nothing reaches stdout, stderr or a log record.
func TestLoadWorkflow_MalformedConsumerQuarantined(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	body := "workflow:\n" +
		jevOtherSetting +
		"    jev:\n" +
		"        enabled: true\n" +
		"        consumers:\n" +
		"            t-zero: {threshold: 0}\n" +
		"            t-one: {threshold: 1}\n" +
		"            t-half: {threshold: 0.5}\n" +
		"            t-negative: {threshold: -0.1}\n" +
		"            t-big: {threshold: 1.5}\n" +
		"            t-string: {threshold: \"high\"}\n" +
		"            t-absent: {mode: advise}\n" +
		"            t-mapping: {threshold: {a: 1}}\n"
	configDir := jevWriteWorkflow(t, t.TempDir(), body)

	var cfg *Config
	stdout, stderr, records := jevCapture(t, func() {
		var err error
		cfg, err = NewLoader().Load(configDir)
		if err != nil {
			t.Errorf("Load: %v", err)
			return
		}
		for _, name := range slices.Sorted(maps.Keys(cfg.Workflow.Jev.Consumers.Entries)) {
			ResolveJevConsumer(cfg.Workflow.Jev.Consumers.Entries[name])
		}
	})
	if cfg == nil {
		t.FailNow()
	}
	jevAssertSettingsKept(t, cfg)

	entries := cfg.Workflow.Jev.Consumers.Entries
	if len(entries) != 8 {
		t.Fatalf("loaded %d consumers, want all 8: %v", len(entries), slices.Sorted(maps.Keys(entries)))
	}
	wantThreshold := map[string]*float64{
		"t-zero": jevPtr(0.0), "t-one": jevPtr(1.0), "t-half": jevPtr(0.5),
		"t-negative": nil, "t-big": nil, "t-string": nil, "t-absent": nil, "t-mapping": nil,
	}
	wantWarned := map[string]bool{"t-negative": true, "t-big": true, "t-string": true, "t-mapping": true}
	for name, want := range wantThreshold {
		got, ok := cfg.Workflow.Jev.Consumers.Resolve(name)
		if !ok {
			t.Errorf("%s: not present in the loaded map", name)
			continue
		}
		switch {
		case want == nil && got.Threshold != nil:
			t.Errorf("%s: threshold = %v, want absent", name, *got.Threshold)
		case want != nil && (got.Threshold == nil || *got.Threshold != *want):
			t.Errorf("%s: threshold = %v, want %v", name, got.Threshold, *want)
		}
		if (len(got.Warnings) > 0) != wantWarned[name] {
			t.Errorf("%s: warnings = %v, want warned = %v", name, got.Warnings, wantWarned[name])
		}
	}
	if got, _ := cfg.Workflow.Jev.Consumers.Resolve("t-absent"); got.Configured != "advise" || got.Effective != JevModeShadow {
		t.Errorf("t-absent configured/effective = %q/%q, want advise/shadow", got.Configured, got.Effective)
	}
	// 0 is distinguishable from absent.
	zero, _ := cfg.Workflow.Jev.Consumers.Resolve("t-zero")
	absent, _ := cfg.Workflow.Jev.Consumers.Resolve("t-absent")
	if zero.Threshold == nil || absent.Threshold != nil {
		t.Error("threshold 0 must be a present value and an absent threshold must stay nil")
	}
	if stdout != "" || stderr != "" || records != 0 {
		t.Errorf("load and resolve must be silent: stdout=%q stderr=%q slog records=%d", stdout, stderr, records)
	}
}

// AC-JEVL-007 (non-mapping half): a `consumers` value that is not a mapping is
// quarantined to the consumers block.
func TestLoadWorkflow_NonMappingConsumersQuarantined(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	body := "workflow:\n" +
		jevOtherSetting +
		"    jev:\n" +
		"        enabled: true\n" +
		"        consumers: \"x\"\n"
	configDir := jevWriteWorkflow(t, t.TempDir(), body)

	var cfg *Config
	stdout, stderr, records := jevCapture(t, func() {
		var err error
		cfg, err = NewLoader().Load(configDir)
		if err != nil {
			t.Errorf("Load: %v", err)
		}
	})
	if cfg == nil {
		t.FailNow()
	}
	jevAssertSettingsKept(t, cfg)
	if n := len(cfg.Workflow.Jev.Consumers.Entries); n != 0 {
		t.Errorf("consumers map has %d entries, want it empty", n)
	}
	if w := cfg.Workflow.Jev.Consumers.SectionWarnings(); len(w) != 1 {
		t.Errorf("section warnings = %v, want exactly one returned as data", w)
	}
	if stdout != "" || stderr != "" || records != 0 {
		t.Errorf("load must be silent: stdout=%q stderr=%q slog records=%d", stdout, stderr, records)
	}
}

// AC-JEVL-001 (config half, partial until M2): loading and resolving the
// configuration, in either gate setting, creates nothing under
// .moai/state/jev.
func TestLoadWorkflow_JevCreatesNothing(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	for _, enabled := range []string{"false", "true"} {
		t.Run("enabled-"+enabled, func(t *testing.T) {
			root := t.TempDir()
			body := "workflow:\n    jev:\n        enabled: " + enabled + "\n" +
				"        consumers:\n            alpha: {mode: advise}\n            beta: {threshold: 0.4}\n"
			cfg, err := NewLoader().Load(jevWriteWorkflow(t, root, body))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got := len(cfg.Workflow.Jev.Consumers.Entries); got != 2 {
				t.Fatalf("loaded %d consumers, want 2 (the zero below would otherwise be vacuous)", got)
			}
			for name := range cfg.Workflow.Jev.Consumers.Entries {
				if _, ok := cfg.Workflow.Jev.Consumers.Resolve(name); !ok {
					t.Fatalf("consumer %s not resolvable", name)
				}
			}
			if _, err := os.Lstat(filepath.Join(root, ".moai", "state", "jev")); !os.IsNotExist(err) {
				t.Errorf(".moai/state/jev must not exist after load+resolve, Lstat err = %v", err)
			}
		})
	}
}

// The consumers block survives the two serializations a Config goes through:
// the manager's YAML section save and the config cache's JSON. Both would lose
// a configured consumer silently if the raw nodes did not round-trip.
func TestJevConsumers_SurviveYAMLAndJSONRoundTrips(t *testing.T) {
	src := "consumers:\n  alpha: {mode: advise, threshold: 0.65}\n  beta: {threshold: 0}\n"
	var in struct {
		Consumers JevConsumers `yaml:"consumers"`
	}
	if err := yaml.Unmarshal([]byte(src), &in); err != nil {
		t.Fatal(err)
	}

	out, err := yaml.Marshal(in)
	if err != nil {
		t.Fatalf("yaml marshal: %v", err)
	}
	var viaYAML struct {
		Consumers JevConsumers `yaml:"consumers"`
	}
	if err := yaml.Unmarshal(out, &viaYAML); err != nil {
		t.Fatalf("yaml re-unmarshal of %q: %v", out, err)
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	var viaJSON struct {
		Consumers JevConsumers `yaml:"consumers"`
	}
	if err := json.Unmarshal(raw, &viaJSON); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	for label, got := range map[string]JevConsumers{"yaml": viaYAML.Consumers, "json": viaJSON.Consumers} {
		alpha, ok := got.Resolve("alpha")
		if !ok || alpha.Configured != "advise" || alpha.Threshold == nil || *alpha.Threshold != 0.65 {
			t.Errorf("%s round trip lost alpha: %+v (found=%v)", label, alpha, ok)
		}
		beta, ok := got.Resolve("beta")
		if !ok || beta.Threshold == nil || *beta.Threshold != 0 {
			t.Errorf("%s round trip lost beta's zero threshold: %+v (found=%v)", label, beta, ok)
		}
	}

	empty, err := yaml.Marshal(struct {
		Consumers JevConsumers `yaml:"consumers"`
	}{})
	if err != nil || !strings.Contains(string(empty), "consumers: {}") {
		t.Errorf("an empty consumers block must marshal as an empty mapping, got %q (err %v)", empty, err)
	}
}

// A field added to Workflow needs a cache schema bump (see
// TestConfigCacheSchemaVersion_BumpedForJev for the hazard).
func TestConfigCacheSchemaVersion_BumpedForJevConsumers(t *testing.T) {
	if configCacheSchemaVersion < 14 {
		t.Errorf("configCacheSchemaVersion = %d, want >= 14 — WorkflowJevConfig gained Consumers without a cache schema bump", configCacheSchemaVersion)
	}
}

// AC-JEVL-013: the budget and margin are decisions, the offset is derived.
func TestJevHookBudgetConstants(t *testing.T) {
	if DefaultJevHookBudget != 5*time.Second {
		t.Errorf("DefaultJevHookBudget = %v, want 5s", DefaultJevHookBudget)
	}
	if DefaultJevHookMargin != 1*time.Second {
		t.Errorf("DefaultJevHookMargin = %v, want 1s", DefaultJevHookMargin)
	}
	if got := JevHookDeadlineOffset(DefaultJevHookBudget, DefaultJevHookMargin); got != 4*time.Second {
		t.Errorf("offset at the defaults = %v, want 4s", got)
	}
	// The derivation must compute, not return a stored value: a test-scale
	// budget and margin give a different offset.
	if got := JevHookDeadlineOffset(250*time.Millisecond, 50*time.Millisecond); got != 200*time.Millisecond {
		t.Errorf("offset for 250ms/50ms = %v, want 200ms", got)
	}
	// A margin that swallows the budget leaves no deadline, never a negative one.
	if got := JevHookDeadlineOffset(time.Second, 2*time.Second); got != 0 {
		t.Errorf("offset with margin > budget = %v, want 0", got)
	}
}

// AC-JEVL-017: the distributed template ships the gate off and an EMPTY
// consumers mapping (present, not absent).
func TestTemplateJevConsumersEmpty(t *testing.T) {
	path := filepath.Join("..", "template", "templates", ".moai", "config", "sections", "workflow.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read template workflow.yaml: %v", err)
	}
	var doc struct {
		Workflow struct {
			Jev struct {
				Enabled   bool      `yaml:"enabled"`
				Consumers yaml.Node `yaml:"consumers"`
			} `yaml:"jev"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal template: %v", err)
	}
	if doc.Workflow.Jev.Enabled {
		t.Error("template ships workflow.jev.enabled: true; it must ship false")
	}
	node := doc.Workflow.Jev.Consumers
	if node.Kind != yaml.MappingNode {
		t.Fatalf("workflow.jev.consumers kind = %v, want an (empty) mapping; zero means the key is absent", node.Kind)
	}
	if len(node.Content) != 0 {
		t.Errorf("workflow.jev.consumers has %d nodes, want an empty mapping", len(node.Content))
	}
}

func jevPtr[T any](v T) *T { return &v }
