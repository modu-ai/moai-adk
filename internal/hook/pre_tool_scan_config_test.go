package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/security"
)

// SPEC-SEC-SCAN-SURFACE-001 M1 — the write-path cost instrument.
//
// Counts are taken on a fake scanner injected through SecurityScanFacade, and
// by snapshotting the process temp directory. ScanFile is the only route from
// this path to an `sg` spawn, so a ScanFile count of 0 proves a spawn count of
// 0 — exact for every skip case asserted here.

// countingScanFacade records every ScanFile call reaching the scanner, along
// with the config path it was handed and the set of security-scan temp files
// visible AT the moment of the call (the deferred cleanup removes them before
// the caller returns, so an after-the-fact snapshot cannot see them).
type countingScanFacade struct {
	calls       int
	configPaths []string
	duringTemp  [][]string
}

func (f *countingScanFacade) IsAvailable() bool { return true }

func (f *countingScanFacade) ScanFile(_ context.Context, _ string, configPath string) (*security.ScanResult, error) {
	f.calls++
	f.configPaths = append(f.configPaths, configPath)
	f.duringTemp = append(f.duringTemp, securityScanTempFiles())
	return &security.ScanResult{Scanned: true}, nil
}

func (f *countingScanFacade) ShouldAlert(*security.ScanResult) bool { return false }

func (f *countingScanFacade) GetReport(*security.ScanResult, string) string { return "" }

// securityScanTempFiles lists the security-scan temp files currently present in
// the process temp directory.
func securityScanTempFiles() []string {
	matches, err := filepath.Glob(filepath.Join(os.TempDir(), "moai-security-scan-*"))
	if err != nil {
		return nil
	}
	return matches
}

// newTempFilesSince returns the entries of after that are absent from before,
// so a concurrently-running test's temp file cannot be mistaken for this one's.
func newTempFilesSince(before, after []string) []string {
	seen := make(map[string]struct{}, len(before))
	for _, p := range before {
		seen[p] = struct{}{}
	}
	var added []string
	for _, p := range after {
		if _, ok := seen[p]; !ok {
			added = append(added, p)
		}
	}
	return added
}

// goPayloadContent is the clean .go payload the coverage criteria above and
// below are measured with.
//
// It carries `const`, which is a go pre-filter token (M2 derives it from
// sec-hardcoded-api-key's `const $NAME = "sk-$$$REST"`). That is load-bearing
// for the CONTROL arms only: a control asserting "a covered language still
// scans" must use a payload that survives every skip the gate applies, and once
// M2 landed a token-free payload is skipped by the pre-filter rather than by
// the coverage check the control is observing. The payload remains clean — it
// trips no rule — so no recorded decision changes.
const goPayloadContent = "package main\n\nconst greeting = \"hello\"\n\nfunc main() {\n\tprintln(greeting)\n}\n"

// TestScanWriteContentNoConfigNoScan closes AC-SSS-002: a project root with no
// resolvable ast-grep configuration dispatches no scan at all.
// Pre-implementation measurement: 1 ScanFile call.
func TestScanWriteContentNoConfigNoScan(t *testing.T) {
	fake := &countingScanFacade{}
	h := &preToolHandler{scanner: fake, projectDir: t.TempDir()}

	decision, reason := h.scanWriteContent(context.Background(), writePayload(t, "sample.go", goPayloadContent))

	if fake.calls != 0 {
		t.Errorf("expected 0 ScanFile calls with no resolvable config, got %d", fake.calls)
	}
	if decision != "" {
		t.Errorf("expected allow, got decision=%q reason=%q", decision, reason)
	}
}

// TestScanWriteContentNoConfigNoTempFile closes AC-SSS-003: the skip happens
// before the temp file is created.
// Pre-implementation measurement: exactly 1 such file is created.
func TestScanWriteContentNoConfigNoTempFile(t *testing.T) {
	t.Run("no config creates no temp file", func(t *testing.T) {
		fake := &countingScanFacade{}
		h := &preToolHandler{scanner: fake, projectDir: t.TempDir()}

		before := securityScanTempFiles()
		h.scanWriteContent(context.Background(), writePayload(t, "sample.go", goPayloadContent))
		after := securityScanTempFiles()

		if added := newTempFilesSince(before, after); len(added) != 0 {
			t.Errorf("expected no security-scan temp file to survive the call, got %v", added)
		}
		if fake.calls != 0 {
			t.Fatalf("expected 0 ScanFile calls, got %d — the during-call snapshot below would be meaningless", fake.calls)
		}
	})

	// Control: with a resolvable config the temp file IS created, so the
	// assertion above is observing the skip rather than a broken instrument.
	t.Run("control: resolvable config creates exactly one temp file", func(t *testing.T) {
		fake := &countingScanFacade{}
		h := &preToolHandler{scanner: fake, projectDir: projectRootWithShippedRuleset(t)}

		before := securityScanTempFiles()
		h.scanWriteContent(context.Background(), writePayload(t, "sample.go", goPayloadContent))

		if fake.calls != 1 {
			t.Fatalf("expected 1 ScanFile call for the control, got %d", fake.calls)
		}
		added := newTempFilesSince(before, fake.duringTemp[0])
		if len(added) != 1 {
			t.Errorf("expected exactly 1 security-scan temp file during the call, got %v", added)
		}
	})
}

// countingRuleManager decorates a RuleManager, counting the resolutions each
// side of the boundary performs.
type countingRuleManager struct {
	inner            security.RuleManager
	findCalls        int
	resolveCovCalls  int
}

func (c *countingRuleManager) FindRulesConfig(projectDir string) string {
	c.findCalls++
	return c.inner.FindRulesConfig(projectDir)
}

func (c *countingRuleManager) ResolveCoverage(projectDir string) security.LanguageCoverage {
	c.resolveCovCalls++
	return c.inner.ResolveCoverage(projectDir)
}

func (c *countingRuleManager) LoadRules(configPath string) ([]string, error) {
	return c.inner.LoadRules(configPath)
}

func (c *countingRuleManager) GetDefaultRules() []string { return c.inner.GetDefaultRules() }

func (c *countingRuleManager) GetEffectiveRules(projectDir string) []string {
	return c.inner.GetEffectiveRules(projectDir)
}

// TestConfigResolvedByCallerNotScanner closes AC-SSS-004: the caller resolves
// the configuration exactly once and the scanner performs no second resolution.
// Pre-implementation measurement: scanner-side 1, caller-side 0 — the two
// counters invert, so the criterion cannot pass on the untouched tree.
func TestConfigResolvedByCallerNotScanner(t *testing.T) {
	root := projectRootWithShippedRuleset(t)

	scannerSide := &countingRuleManager{inner: security.NewRuleManager()}
	callerSide := &countingRuleManager{inner: security.NewRuleManager()}

	scanner := security.NewSecurityScannerWithConfig(&security.ScannerConfig{Rules: scannerSide})
	h := &preToolHandler{scanner: scanner, rules: callerSide, projectDir: root}

	h.scanWriteContent(context.Background(), writePayload(t, "sample.go", goPayloadContent))

	if scannerSide.findCalls != 0 {
		t.Errorf("scanner-side config resolutions: got %d, want 0", scannerSide.findCalls)
	}
	if callerSide.resolveCovCalls != 1 {
		t.Errorf("caller-side config resolutions: got %d, want 1", callerSide.resolveCovCalls)
	}
}

// TestScanWriteContentUncoveredLanguage closes AC-SSS-005: a payload in an
// ast-grep-supported but rule-uncovered language dispatches no scan.
// Pre-implementation measurement: 11 calls.
func TestScanWriteContentUncoveredLanguage(t *testing.T) {
	root := projectRootWithShippedRuleset(t)

	uncovered := []struct {
		path    string
		content string
	}{
		{"main.rs", "fn main() { println!(\"hi\"); }\n"},
		{"Sample.java", "class Sample { void run() {} }\n"},
		{"Sample.kt", "fun main() { println(\"hi\") }\n"},
		{"sample.c", "int main(void) { return 0; }\n"},
		{"sample.cpp", "int main() { return 0; }\n"},
		{"sample.rb", "def run\n  puts 'hi'\nend\n"},
		{"sample.php", "<?php function run() { echo 'hi'; }\n"},
		{"Sample.swift", "func run() { print(\"hi\") }\n"},
		{"Sample.cs", "class Sample { void Run() {} }\n"},
		{"sample.ex", "defmodule Sample do\nend\n"},
		{"Sample.scala", "object Sample { def run(): Unit = () }\n"},
	}

	for _, u := range uncovered {
		t.Run(u.path, func(t *testing.T) {
			fake := &countingScanFacade{}
			h := &preToolHandler{scanner: fake, projectDir: root}
			decision, _ := h.scanWriteContent(context.Background(), writePayload(t, u.path, u.content))
			if fake.calls != 0 {
				t.Errorf("expected 0 ScanFile calls for %s, got %d", u.path, fake.calls)
			}
			if decision != "" {
				t.Errorf("expected allow for %s, got %q", u.path, decision)
			}
		})
	}

	// Control: a covered language in the same ruleset still scans, so the
	// zeroes above are not the test scanning nothing at all.
	t.Run("control sample.go", func(t *testing.T) {
		fake := &countingScanFacade{}
		h := &preToolHandler{scanner: fake, projectDir: root}
		h.scanWriteContent(context.Background(), writePayload(t, "sample.go", goPayloadContent))
		if fake.calls != 1 {
			t.Errorf("expected 1 ScanFile call for the .go control, got %d", fake.calls)
		}
	})
}

// TestScanWriteContentCoveredLanguageFollowsConfig is the end-to-end half of
// AC-SSS-006: the 0/1 scan-count split follows the configuration. The
// derivation itself is asserted in internal/hook/security (coverage_test.go).
func TestScanWriteContentCoveredLanguageFollowsConfig(t *testing.T) {
	t.Run("shipped ruleset scans go", func(t *testing.T) {
		fake := &countingScanFacade{}
		h := &preToolHandler{scanner: fake, projectDir: projectRootWithShippedRuleset(t)}
		h.scanWriteContent(context.Background(), writePayload(t, "sample.go", goPayloadContent))
		if fake.calls != 1 {
			t.Errorf("expected 1 ScanFile call, got %d", fake.calls)
		}
	})

	t.Run("ruleDirs without go rules skips go", func(t *testing.T) {
		root := projectRootWithShippedRuleset(t)
		rulesDir := filepath.Join(root, ".moai", "config", "astgrep-rules")
		pyDir := filepath.Join(rulesDir, "pyonly")
		if err := os.MkdirAll(pyDir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		rule := "id: py-only-probe\nlanguage: python\nseverity: error\nmessage: probe\nrule:\n  pattern: os.system($CMD)\n"
		if err := os.WriteFile(filepath.Join(pyDir, "probe.yml"), []byte(rule), 0o644); err != nil {
			t.Fatalf("write rule: %v", err)
		}
		if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs:\n  - pyonly\n"), 0o644); err != nil {
			t.Fatalf("write sgconfig: %v", err)
		}

		fake := &countingScanFacade{}
		h := &preToolHandler{scanner: fake, projectDir: root}
		h.scanWriteContent(context.Background(), writePayload(t, "sample.go", goPayloadContent))
		if fake.calls != 0 {
			t.Errorf("expected 0 ScanFile calls once the config no longer covers go, got %d", fake.calls)
		}
	})
}

// TestScanWriteContentUnreadableConfigEscalates is the end-to-end half of
// AC-SSS-007: an unreadable, unwalkable, or empty derivation escalates to a
// scan rather than skipping. This is a behaviour-preservation criterion — its
// PASS value equals today's value, and M1's skip logic is what could break it.
func TestScanWriteContentUnreadableConfigEscalates(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(t *testing.T, rulesDir string)
	}{
		{
			name: "malformed yaml",
			prepare: func(t *testing.T, rulesDir string) {
				if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs: [go\n  : : not yaml\n"), 0o644); err != nil {
					t.Fatalf("write sgconfig: %v", err)
				}
			},
		},
		{
			name: "ruleDirs names a missing directory",
			prepare: func(t *testing.T, rulesDir string) {
				if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs:\n  - does-not-exist\n"), 0o644); err != nil {
					t.Fatalf("write sgconfig: %v", err)
				}
			},
		},
		{
			name: "rule files declare no language",
			prepare: func(t *testing.T, rulesDir string) {
				dir := filepath.Join(rulesDir, "nolang")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				rule := "id: no-language-probe\nseverity: error\nmessage: probe\nrule:\n  pattern: probe($X)\n"
				if err := os.WriteFile(filepath.Join(dir, "probe.yml"), []byte(rule), 0o644); err != nil {
					t.Fatalf("write rule: %v", err)
				}
				if err := os.WriteFile(filepath.Join(rulesDir, "sgconfig.yml"), []byte("ruleDirs:\n  - nolang\n"), 0o644); err != nil {
					t.Fatalf("write sgconfig: %v", err)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := projectRootWithShippedRuleset(t)
			tc.prepare(t, filepath.Join(root, ".moai", "config", "astgrep-rules"))

			fake := &countingScanFacade{}
			h := &preToolHandler{scanner: fake, projectDir: root}
			h.scanWriteContent(context.Background(), writePayload(t, "sample.go", goPayloadContent))

			if fake.calls != 1 {
				t.Errorf("expected 1 ScanFile call (fail-open escalation), got %d", fake.calls)
			}
		})
	}
}
