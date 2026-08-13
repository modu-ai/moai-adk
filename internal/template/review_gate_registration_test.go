package template

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Both Stop-hook review-gate wrappers were unreachable because nothing
// registered them: neither the rendered settings.json.tmpl Stop array nor the
// repo's own .claude/settings.json listed them, so Claude Code never invoked
// either script no matter how the project was configured. Two earlier sweeps
// dropped sibling wiring the same silent way, so these tests pin the
// registration on BOTH surfaces and pin the shell self-gate that keeps the
// registration free for the users who never opt in.

// reviewGateWrappers are the two wrappers under test, with the timeout each
// must carry. 900s mirrors config.DefaultCodexReviewGateTimeout /
// config.DefaultMultiReviewGateTimeout: a real review runs far past the 5s
// moai-default hook budget.
var reviewGateWrappers = map[string]float64{
	"handle-codex-review-gate.sh": 900,
	"handle-multi-review-gate.sh": 900,
}

// stopHookEntry models one entry of a settings.json Stop array. The two
// surfaces spell the invocation differently — the template uses
// "command": "bash" + "args", the repo's settings.json uses a single quoted
// command string — which is the pre-existing renderer convention, not drift.
// Both spellings are decoded here and reduced to (script, timeout).
type stopHookEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Timeout float64  `json:"timeout"`
}

// script returns the hook script basename this entry invokes. It reads the
// LAST args element so the same extraction covers: the guarded exec form
// (args = ["-c", "<guard>", "<path>"] — path is last), the legacy exec form
// (args = ["<path>"]), and the shell form (args empty → fall back to command).
func (e stopHookEntry) script() string {
	target := e.Command
	if len(e.Args) > 0 {
		target = e.Args[len(e.Args)-1]
	}
	target = strings.Trim(target, `"`)
	return filepath.Base(target)
}

// parseStopHooks extracts the flattened Stop-array entries from a settings JSON
// document.
func parseStopHooks(t *testing.T, raw string) []stopHookEntry {
	t.Helper()
	var doc struct {
		Hooks struct {
			Stop []struct {
				Hooks []stopHookEntry `json:"hooks"`
			} `json:"Stop"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("settings JSON did not parse: %v", err)
	}
	var out []stopHookEntry
	for _, group := range doc.Hooks.Stop {
		out = append(out, group.Hooks...)
	}
	if len(out) == 0 {
		t.Fatal("Stop array is empty; the settings document is not what this test expects")
	}
	return out
}

// assertRegistered asserts every review-gate wrapper appears exactly once in
// entries with its required timeout.
func assertRegistered(t *testing.T, surface string, entries []stopHookEntry) {
	t.Helper()
	seen := map[string]int{}
	for _, e := range entries {
		if want, ok := reviewGateWrappers[e.script()]; ok {
			seen[e.script()]++
			if e.Timeout != want {
				t.Errorf("%s: %s timeout = %v, want %v", surface, e.script(), e.Timeout, want)
			}
		}
	}
	for script := range reviewGateWrappers {
		switch seen[script] {
		case 1: // registered exactly once
		case 0:
			t.Errorf("%s: %s is NOT registered in the Stop array — the gate can never fire", surface, script)
		default:
			t.Errorf("%s: %s registered %d times, want exactly 1", surface, script, seen[script])
		}
	}
}

// TestReviewGatesRegisteredInTemplateSettings pins the registration in the
// deployed surface: the rendered settings.json.tmpl Stop array.
func TestReviewGatesRegisteredInTemplateSettings(t *testing.T) {
	for _, platform := range []string{"darwin", "linux", "windows"} {
		t.Run(platform, func(t *testing.T) {
			rendered := renderTemplate(t, ".claude/settings.json.tmpl", testContext(platform))
			assertRegistered(t, "settings.json.tmpl ("+platform+")", parseStopHooks(t, rendered))
		})
	}
}

// TestReviewGatesRegisteredInRepoSettings pins the registration in this repo's
// own .claude/settings.json, so the gates are reachable here too.
func TestReviewGatesRegisteredInRepoSettings(t *testing.T) {
	path := filepath.Join(repoRootFromTemplatePkg(t), ".claude", "settings.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	assertRegistered(t, ".claude/settings.json", parseStopHooks(t, string(raw)))
}

// TestReviewGateWrappersSelfGateBeforeBinaryResolution pins the cost
// constraint that makes registration acceptable. Both gates ship OFF, so a
// wrapper that resolved and exec'd the moai binary unconditionally would add
// two cold starts to every turn-end for every user. Each wrapper must read the
// opt-in and exit early BEFORE it resolves the binary — the ordering is what is
// pinned here; the runtime behaviour is proven by the executable test below.
func TestReviewGateWrappersSelfGateBeforeBinaryResolution(t *testing.T) {
	root := repoRootFromTemplatePkg(t)
	for script, gate := range map[string]string{
		"handle-codex-review-gate.sh": "codex",
		"handle-multi-review-gate.sh": "multi",
	} {
		for _, dir := range []string{
			filepath.Join(root, ".claude", "hooks", "moai"),
			filepath.Join(root, "internal", "template", "templates", ".claude", "hooks", "moai"),
		} {
			path := filepath.Join(dir, script)
			t.Run(filepath.Base(dir)+"/"+script+"/"+gate, func(t *testing.T) {
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read %s: %v", path, err)
				}
				body := string(raw)

				optIn := strings.Index(body, "workflow.yaml")
				earlyExit := strings.Index(body, `!= "true" ] || exit 0`)
				if earlyExit < 0 {
					earlyExit = strings.Index(body, `= "true" ] || exit 0`)
				}
				resolve := strings.Index(body, "Resolve the moai binary")

				if optIn < 0 {
					t.Fatalf("%s: no workflow.yaml opt-in read; the wrapper cannot self-gate", path)
				}
				if earlyExit < 0 {
					t.Fatalf("%s: no early `exit 0` on the disabled path", path)
				}
				if resolve < 0 {
					t.Fatalf("%s: binary-resolution block not found; test needs updating", path)
				}
				if optIn >= earlyExit || earlyExit >= resolve {
					t.Errorf("%s: self-gate must precede binary resolution (opt-in read @%d, early exit @%d, resolve @%d)",
						path, optIn, earlyExit, resolve)
				}
				// The gate must read its OWN key, not its sibling's.
				if !strings.Contains(body, `gate="`+gate+`"`) {
					t.Errorf("%s: self-gate does not read the %q key", path, gate)
				}
				// No new shared failure mode: the parse stays dependency-free.
				// Comment lines are excluded — they discuss the constraint.
				var code []string
				for _, line := range strings.Split(body, "\n") {
					if !strings.HasPrefix(strings.TrimSpace(line), "#") {
						code = append(code, line)
					}
				}
				executable := strings.Join(code, "\n")
				for _, banned := range []string{"yq", "jq"} {
					if strings.Contains(executable, banned+" ") || strings.Contains(executable, "v "+banned) {
						t.Errorf("%s: wrapper depends on %q; the self-gate must be pure shell", path, banned)
					}
				}
			})
		}
	}
}

// TestReviewGateOptInDocumentedInTemplateWorkflowYAML pins discoverability: a
// user cannot opt into a key that appears nowhere in the config they were
// given. Both blocks must ship, both must ship false.
func TestReviewGateOptInDocumentedInTemplateWorkflowYAML(t *testing.T) {
	path := filepath.Join(repoRootFromTemplatePkg(t),
		"internal", "template", "templates", ".moai", "config", "sections", "workflow.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	body := string(raw)
	for _, want := range []string{
		"    codex:\n        review_gate:\n            enabled: false\n",
		"    multi:\n        review_gate:\n            enabled: false\n",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("template workflow.yaml missing the opt-in block:\n%s", want)
		}
	}
	// The distributed default is OFF for both gates.
	if strings.Contains(body, "review_gate:\n            enabled: true") {
		t.Error("template workflow.yaml must never ship a review gate enabled")
	}
}

// repoRootFromTemplatePkg resolves the repository root from this package's
// directory (internal/template).
func repoRootFromTemplatePkg(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Join(wd, "..", "..")
}
