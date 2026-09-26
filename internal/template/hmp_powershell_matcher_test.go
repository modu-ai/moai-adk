package template

// SPEC-HOOK-MATCHER-POWERSHELL-001 (card t1224) — template-side guards.
//
// The PowerShell tool is a shell tool with its own name. A hook registered for
// the matcher "Bash" never sees a PowerShell call, so every (event, hook
// script) pair that names Bash must also be registered for PowerShell, or sit
// on the exclusion list below with a reason. The guard is closed-world: a new
// Bash registration with neither fails, and so does an exclusion that no
// longer names a Bash registration.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// hmpMatcherExclusions lists (event, hook script) pairs registered for Bash
// that intentionally have no PowerShell registration, keyed "<event> <script>",
// with the reason. Empty today: every Bash registration has a counterpart.
var hmpMatcherExclusions = map[string]string{}

// hmpRegistration is one (event, matcher, hook script) triple.
type hmpRegistration struct {
	Event   string
	Matcher string
	Script  string
}

// hmpProjectDirPrefixes are the spellings of the project-dir prefix a hook
// script path may carry; the script identity is the path with it removed.
var hmpProjectDirPrefixes = []string{
	"${CLAUDE_PROJECT_DIR}/",
	"$CLAUDE_PROJECT_DIR/",
	`"$CLAUDE_PROJECT_DIR"/`,
}

// hmpScriptOf returns a hook entry's script identity: the final args element
// without the project-dir prefix, or the command when args is absent.
func hmpScriptOf(command string, args []string) string {
	s := command
	if len(args) > 0 {
		s = args[len(args)-1]
	}
	for _, p := range hmpProjectDirPrefixes {
		s = strings.TrimPrefix(s, p)
	}
	return s
}

// hmpRegistrations parses a settings document and returns every hook
// registration it carries.
func hmpRegistrations(t *testing.T, body string) []hmpRegistration {
	t.Helper()
	var doc struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Command string   `json:"command"`
				Args    []string `json:"args"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatalf("settings do not parse as JSON: %v", err)
	}
	var out []hmpRegistration
	for event, blocks := range doc.Hooks {
		for _, b := range blocks {
			for _, h := range b.Hooks {
				out = append(out, hmpRegistration{Event: event, Matcher: b.Matcher, Script: hmpScriptOf(h.Command, h.Args)})
			}
		}
	}
	return out
}

// hmpMatcherNames splits a matcher into the tool names it lists.
func hmpMatcherNames(matcher string) []string {
	return strings.FieldsFunc(matcher, func(r rune) bool { return r == '|' || r == ',' })
}

func hmpNames(matcher, tool string) bool {
	for _, n := range hmpMatcherNames(matcher) {
		if strings.TrimSpace(n) == tool {
			return true
		}
	}
	return false
}

// hmpParityViolations returns one message per (event, script) pair that is
// registered for Bash with no PowerShell registration and no exclusion, plus
// one per stale exclusion.
func hmpParityViolations(regs []hmpRegistration, exclusions map[string]string) []string {
	bash := map[string]bool{}
	ps := map[string]bool{}
	for _, r := range regs {
		key := r.Event + " " + r.Script
		if hmpNames(r.Matcher, "Bash") {
			bash[key] = true
		}
		if hmpNames(r.Matcher, "PowerShell") {
			ps[key] = true
		}
	}
	var out []string
	for key := range bash {
		_, excluded := exclusions[key]
		switch {
		case !ps[key] && !excluded:
			out = append(out, "Bash registration has no PowerShell counterpart: "+key)
		case ps[key] && excluded:
			out = append(out, "excluded Bash registration also has a PowerShell counterpart: "+key)
		}
	}
	for key := range exclusions {
		if !bash[key] {
			out = append(out, "exclusion names no Bash registration: "+key)
		}
	}
	sort.Strings(out)
	return out
}

// hmpRepoRoot walks up to the directory containing go.mod.
func hmpRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found")
		}
		dir = parent
	}
}

// hmpLocalFile reads a repository-local file, skipping when it is absent (a
// consumer checkout carries no local dogfood copy).
func hmpLocalFile(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(hmpRepoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Skipf("%s not present: %v", rel, err)
	}
	return string(data)
}

const hmpPreToolScript = ".claude/hooks/moai/handle-pre-tool.sh"

// hmpIsolateHome points the MoAI home at a per-test directory (REQ-HMP-014).
func hmpIsolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
}

// TestHMPMatcherParityTemplate pins AC-HMP-003 / AC-HMP-001 on the rendered
// distributed template, for every platform the template renders.
func TestHMPMatcherParityTemplate(t *testing.T) {
	hmpIsolateHome(t)
	for _, platform := range []string{"darwin", "linux", "windows"} {
		t.Run(platform, func(t *testing.T) {
			regs := hmpRegistrations(t, renderTemplate(t, ".claude/settings.json.tmpl", testContext(platform)))
			for _, v := range hmpParityViolations(regs, hmpMatcherExclusions) {
				t.Error(v)
			}
		})
	}
}

// TestHMPMatcherParityLocal pins the same guard on the local dogfood settings,
// and that the local and template PreToolUse routes to the pre-tool wrapper
// name the same matchers (AC-HMP-001 "the two files agree").
func TestHMPMatcherParityLocal(t *testing.T) {
	hmpIsolateHome(t)
	local := hmpRegistrations(t, hmpLocalFile(t, ".claude/settings.json"))
	for _, v := range hmpParityViolations(local, hmpMatcherExclusions) {
		t.Error(v)
	}
	tmpl := hmpRegistrations(t, renderTemplate(t, ".claude/settings.json.tmpl", testContext("darwin")))
	matchers := func(regs []hmpRegistration) []string {
		var out []string
		for _, r := range regs {
			if r.Event == "PreToolUse" && r.Script == hmpPreToolScript {
				out = append(out, r.Matcher)
			}
		}
		sort.Strings(out)
		return out
	}
	gotLocal, gotTmpl := matchers(local), matchers(tmpl)
	if strings.Join(gotLocal, "\n") != strings.Join(gotTmpl, "\n") {
		t.Errorf("PreToolUse matchers for %s differ: local %q, template %q", hmpPreToolScript, gotLocal, gotTmpl)
	}
	ps := false
	for _, m := range gotLocal {
		if hmpNames(m, "PowerShell") {
			ps = true
		}
	}
	if !ps {
		t.Errorf("no PreToolUse registration of %s names PowerShell: %q", hmpPreToolScript, gotLocal)
	}
}

// TestHMPMatcherParityMutants proves the guard can fail (AC-HMP-003).
func TestHMPMatcherParityMutants(t *testing.T) {
	hmpIsolateHome(t)
	base := hmpRegistrations(t, renderTemplate(t, ".claude/settings.json.tmpl", testContext("linux")))

	t.Run("m1 PowerShell registration removed", func(t *testing.T) {
		var mutated []hmpRegistration
		for _, r := range base {
			if r.Event == "PreToolUse" && r.Script == hmpPreToolScript && hmpNames(r.Matcher, "PowerShell") {
				continue
			}
			mutated = append(mutated, r)
		}
		got := strings.Join(hmpParityViolations(mutated, hmpMatcherExclusions), "\n")
		want := "no PowerShell counterpart: PreToolUse " + hmpPreToolScript
		if !strings.Contains(got, want) {
			t.Errorf("guard did not name %q; got %q", want, got)
		}
	})

	t.Run("m2 PowerShell on a different script", func(t *testing.T) {
		fixture := `{"hooks": {"PreToolUse": [
			{"matcher": "Write|Edit|Bash", "hooks": [{"command": "bash", "args": ["-c", "x", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-pre-tool.sh"]}]},
			{"matcher": "Bash", "hooks": [{"command": "bash", "args": ["-c", "x", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/second.sh"]}]},
			{"matcher": "PowerShell", "hooks": [{"command": "bash", "args": ["-c", "x", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/second.sh"]}]}
		]}}`
		got := hmpParityViolations(hmpRegistrations(t, fixture), nil)
		if len(got) != 1 || !strings.Contains(got[0], "no PowerShell counterpart: PreToolUse "+hmpPreToolScript) {
			t.Errorf("want exactly the handle-pre-tool.sh violation, got %q", got)
		}
	})

	t.Run("stale exclusion", func(t *testing.T) {
		got := strings.Join(hmpParityViolations(base, map[string]string{"PreToolUse .claude/hooks/moai/gone.sh": "x"}), "\n")
		if !strings.Contains(got, "exclusion names no Bash registration: PreToolUse .claude/hooks/moai/gone.sh") {
			t.Errorf("stale exclusion not reported: %q", got)
		}
	})

	t.Run("command without args is the identity", func(t *testing.T) {
		if got := hmpScriptOf("${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/x.sh", nil); got != ".claude/hooks/moai/x.sh" {
			t.Errorf("hmpScriptOf(command only) = %q", got)
		}
	})
}

// hmpRiskWarning is the tag the pre-tool wrapper prints on its Bash
// Risk-Amplifier warning.
const hmpRiskWarning = "[moai:bash-risk]"

// hmpRunWrapper runs a pre-tool wrapper body with stdin payload, a stub moai
// first on PATH, and the stderr log inside a per-test project, and returns
// the wrapper's stderr plus the log body.
func hmpRunWrapper(t *testing.T, body, payload string) (stderr, logBody string) {
	t.Helper()
	proj := t.TempDir()
	bin := t.TempDir()
	home := t.TempDir()
	script := filepath.Join(proj, ".claude", "hooks", "moai", "handle-pre-tool.sh")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}
	stub := "#!/bin/sh\ncat >/dev/null\nexit 0\n"
	if err := os.WriteFile(filepath.Join(bin, "moai"), []byte(stub), 0o755); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	logPath := filepath.Join(proj, ".moai", "logs", "hook-stderr.log")
	cmd := exec.Command("bash", script)
	cmd.Dir = proj
	cmd.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"HOME="+home,
		"CLAUDE_PROJECT_DIR="+proj,
		"MOAI_HOOK_STDERR_LOG="+logPath,
	)
	cmd.Stdin = strings.NewReader(payload)
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		t.Fatalf("wrapper exited with %v; stderr %q", err, errBuf.String())
	}
	data, _ := os.ReadFile(logPath)
	return errBuf.String(), string(data)
}

func hmpWrapperPayload(tool string) string {
	data, err := json.Marshal(map[string]any{
		"hook_event_name": "PreToolUse",
		"tool_name":       tool,
		"tool_input":      map[string]any{"command": "a | b | c | d | e | f | g"},
	})
	if err != nil {
		panic(err)
	}
	return string(data)
}

// TestHMPWrapperRiskWarningBashOnly pins AC-HMP-013 (D6 = keep Bash-only) on
// the rendered template wrapper and on the local copy.
func TestHMPWrapperRiskWarningBashOnly(t *testing.T) {
	hmpIsolateHome(t)
	if runtime.GOOS == "windows" {
		t.Skip("the pre-tool wrapper is a POSIX shell script")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}
	bodies := map[string]string{
		"template": renderTemplate(t, ".claude/hooks/moai/handle-pre-tool.sh.tmpl", testContext("darwin")),
	}
	if data, err := os.ReadFile(filepath.Join(hmpRepoRoot(t), filepath.FromSlash(hmpPreToolScript))); err == nil {
		bodies["local"] = string(data)
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			bashErr, bashLog := hmpRunWrapper(t, body, hmpWrapperPayload("Bash"))
			if !strings.Contains(bashErr, hmpRiskWarning) || !strings.Contains(bashLog, hmpRiskWarning) {
				t.Fatalf("premise: Bash control produced no warning; stderr %q log %q", bashErr, bashLog)
			}
			psErr, psLog := hmpRunWrapper(t, body, hmpWrapperPayload("PowerShell"))
			if strings.Contains(psErr, hmpRiskWarning) || strings.Contains(psLog, hmpRiskWarning) {
				t.Errorf("PowerShell payload produced the Bash Risk-Amplifier warning (D6 = Bash-only); stderr %q log %q", psErr, psLog)
			}
		})
	}
}

// hmpScopeComment is the phrase the wrapper's matcher-scope comment must carry:
// it names the PowerShell matcher delivered alongside Write|Edit|Bash.
const hmpScopeComment = `matchers "Write|Edit|Bash" and "PowerShell"`

// TestHMPWrapperScopeComment pins AC-HMP-013's comment clause in both copies.
func TestHMPWrapperScopeComment(t *testing.T) {
	hmpIsolateHome(t)
	tmpl := renderTemplate(t, ".claude/hooks/moai/handle-pre-tool.sh.tmpl", testContext("darwin"))
	if !strings.Contains(tmpl, hmpScopeComment) {
		t.Errorf("template wrapper comment does not name the delivered matcher (%s)", hmpScopeComment)
	}
	if local := hmpLocalFile(t, hmpPreToolScript); !strings.Contains(local, hmpScopeComment) {
		t.Errorf("local wrapper comment does not name the delivered matcher (%s)", hmpScopeComment)
	}
}
