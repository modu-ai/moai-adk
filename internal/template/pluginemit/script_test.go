// script_test.go — scripts/check-plugin-discoverable.sh against a recording
// stand-in for the Claude CLI (AC-006 (b) to (e), plan-audit BI-3).
//
// The script is the one place that runs the real CLI, and it does so only on
// demand. These tests drive it with a recording wrapper first on PATH, so no
// real tool starts: they prove the contract the script states — which verbs it
// starts, from where, under which environment, and what it refuses — and a
// regression in any of them turns red in every `go test` run.
package pluginemit_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

const stubClaude = `#!/bin/sh
{
  echo "argv: $*"
  echo "cwd: $(pwd)"
  awk 'BEGIN { for (n in ENVIRON) if (n ~ /^(MOAI|CLAUDE|CODEX)_/) print "env: " n "=" ENVIRON[n] }'
} >> "$HARNESS_STUB_LOG"
case "$1 $2" in
  "plugin details") cat "$HARNESS_STUB_DETAILS" ;;
  *) echo '{"outcome":"ok"}' ;;
esac
`

// scriptNames returns the names the script must find listed: the skill
// directories of the tier view of the template tree plus its command stems. It
// reads neither the generator nor the committed payload, so a payload that
// drifts from the template cannot define its own expectation.
func scriptNames(t *testing.T) []string {
	t.Helper()
	raw := os.DirFS(rawTemplateDir)
	cat, err := template.LoadCatalog(raw)
	if err != nil {
		t.Fatal(err)
	}
	view, err := template.SlimFS(raw, cat)
	if err != nil {
		t.Fatal(err)
	}
	skills, err := fs.ReadDir(view, ".claude/skills")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, s := range skills {
		names = append(names, s.Name())
	}
	cmds, err := os.ReadDir(filepath.Join(rawTemplateDir, "templates", ".claude", "commands", "moai"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cmds {
		names = append(names, strings.TrimSuffix(strings.TrimSuffix(c.Name(), ".tmpl"), ".md"))
	}
	if len(names) == 0 {
		t.Fatal("no expected names; an empty sweep asserts nothing")
	}
	sort.Strings(names)
	return names
}

// detailsText renders the component inventory in the shape of
// `claude plugin details` (observed at claude 2.1.288).
func detailsText(names []string, mcp string) string {
	return fmt.Sprintf("moai 0.0.0\n  Source: moai@moai-adk\n\nComponent inventory\n  Skills (%d)  %s\n  Agents (0)\n  Hooks (0)\n  MCP servers (1)  %s  (tool schemas resolved at runtime; not counted)\n  LSP servers (0)\n",
		len(names), strings.Join(names, ", "), mcp)
}

type scriptRun struct {
	out    string
	exit   int
	log    string
	home   string
	canary string
}

// runScript runs the discoverable script with the stand-in first on PATH and a
// poisoned caller environment. homeArg builds the argument from a scratch dir.
func runScript(t *testing.T, details string, homeArg func(dir string) string) scriptRun {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the script is a POSIX shell script")
	}
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh on PATH")
	}
	root := t.TempDir()
	stubDir := filepath.Join(root, "stub")
	canary := filepath.Join(root, "canary")
	recorder := filepath.Join(root, "recorder")
	logPath := filepath.Join(root, "stub.log")
	detailsPath := filepath.Join(root, "details.txt")
	scratch := filepath.Join(root, "scratch-tmp")
	home := filepath.Join(root, "home")
	for _, d := range []string{stubDir, canary, scratch, home} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for p, c := range map[string]string{
		filepath.Join(stubDir, "claude"): stubClaude,
		recorder:                         "#!/bin/sh\necho \"recorder $*\" >> \"" + logPath + ".recorder\"\n",
		detailsPath:                      details,
	} {
		mode := os.FileMode(0o755)
		if p == detailsPath {
			mode = 0o644
		}
		if err := os.WriteFile(p, []byte(c), mode); err != nil {
			t.Fatal(err)
		}
	}

	pid := strconv.Itoa(os.Getpid())
	env := []string{
		"PATH=" + stubDir + ":/usr/bin:/bin",
		"TMPDIR=" + scratch,
		"HARNESS_STUB_LOG=" + logPath,
		"HARNESS_STUB_DETAILS=" + detailsPath,
		"MOAI_CLAUDE_BIN=" + recorder,
		"CLAUDE_CODE_PLUGIN_CACHE_DIR=" + canary,
		"MOAI_PLANTED_" + pid + "=1",
		"CLAUDE_PLANTED_" + pid + "=1",
		"CODEX_PLANTED_" + pid + "=1",
	}
	script, err := filepath.Abs(filepath.Join(repoRoot, "scripts", "check-plugin-discoverable.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(sh, script, homeArg(home))
	cmd.Env = env
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	exit := 0
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run script: %v", err)
		}
		exit = ee.ExitCode()
	}
	logBytes, _ := os.ReadFile(logPath)
	if rec, err := os.ReadFile(logPath + ".recorder"); err == nil {
		t.Errorf("the poisoned MOAI_CLAUDE_BIN recorder ran: %s", rec)
	}
	if entries, _ := os.ReadDir(canary); len(entries) != 0 {
		t.Errorf("the canary directory was written: %d entries", len(entries))
	}
	return scriptRun{out: out.String(), exit: exit, log: string(logBytes), home: home, canary: canary}
}

func (r scriptRun) lastLine() string {
	lines := strings.Split(strings.TrimRight(r.out, "\n"), "\n")
	return lines[len(lines)-1]
}

func (r scriptRun) lines(prefix string) []string {
	var out []string
	for _, l := range strings.Split(r.log, "\n") {
		if rest, ok := strings.CutPrefix(l, prefix); ok {
			out = append(out, rest)
		}
	}
	return out
}

func TestCheckPluginDiscoverable(t *testing.T) {
	names := scriptNames(t)
	identity := func(d string) string { return d }

	t.Run("refuses-a-non-empty-home", func(t *testing.T) {
		r := runScript(t, detailsText(names, "moai"), func(d string) string {
			_ = os.WriteFile(filepath.Join(d, "keep"), []byte("x"), 0o644)
			return d
		})
		if r.exit != 2 || !strings.Contains(r.out, "refused:") {
			t.Errorf("exit=%d out=%q, want exit 2 and a refused: line", r.exit, r.out)
		}
		if r.log != "" {
			t.Errorf("a tool started before the refusal: %q", r.log)
		}
	})

	t.Run("refuses-a-path-that-is-not-a-directory", func(t *testing.T) {
		r := runScript(t, detailsText(names, "moai"), func(d string) string { return filepath.Join(d, "absent") })
		if r.exit != 2 || !strings.Contains(r.out, "refused:") || r.log != "" {
			t.Errorf("exit=%d out=%q log=%q, want exit 2, a refused: line and no tool start", r.exit, r.out, r.log)
		}
	})

	t.Run("three-verbs-under-a-scrubbed-environment", func(t *testing.T) {
		r := runScript(t, detailsText(names, "moai"), identity)
		if r.exit != 0 {
			t.Fatalf("exit=%d out=%s", r.exit, r.out)
		}
		if want := fmt.Sprintf("ok: %d names listed, 0 missing", len(names)); r.lastLine() != want {
			t.Errorf("last line = %q, want %q", r.lastLine(), want)
		}

		// Exactly three recorded argv: the three named verbs, in order.
		argv := r.lines("argv: ")
		if len(argv) != 3 {
			t.Fatalf("the script started %d tool commands, want exactly 3: %v", len(argv), argv)
		}
		if !strings.HasPrefix(argv[0], "plugin marketplace add ") || !strings.HasSuffix(argv[0], " --json") ||
			argv[1] != "plugin install moai@moai-adk --json" || argv[2] != "plugin details moai@moai-adk" {
			t.Errorf("recorded verbs = %v", argv)
		}

		// The marketplace source is a local path: an existing directory that is the repository root.
		src := strings.TrimSuffix(strings.TrimPrefix(argv[0], "plugin marketplace add "), " --json")
		if info, err := os.Stat(src); err != nil || !info.IsDir() {
			t.Errorf("marketplace source %q is not an existing local directory", src)
		}
		repoAbs, _ := filepath.Abs(repoRoot)
		wantRoot, _ := filepath.EvalSymlinks(repoAbs)
		if gotRoot, _ := filepath.EvalSymlinks(src); gotRoot != wantRoot {
			t.Errorf("marketplace source resolves to %q, want the repository root %q", gotRoot, wantRoot)
		}

		// The scrub removed every planted name, including the pid-named ones no typed
		// list could contain; the only MOAI_/CLAUDE_/CODEX_ name a call sees is the home.
		m := regexp.MustCompile(`scrub: enumerated and unset (\d+) names`).FindStringSubmatch(r.out)
		if m == nil {
			t.Fatalf("no scrub count line in %q", r.out)
		}
		if n, _ := strconv.Atoi(m[1]); n < 5 {
			t.Errorf("scrub count = %d, want at least the 5 planted names", n)
		}
		envLines := r.lines("env: ")
		if len(envLines) != 3 {
			t.Errorf("recorded env names = %v, want exactly CLAUDE_CONFIG_DIR once per call", envLines)
		}
		wantHome, _ := filepath.EvalSymlinks(r.home)
		for _, l := range envLines {
			name, val, _ := strings.Cut(l, "=")
			gotHome, _ := filepath.EvalSymlinks(val)
			if name != "CLAUDE_CONFIG_DIR" || gotHome != wantHome {
				t.Errorf("a call saw %q, want only CLAUDE_CONFIG_DIR=%s", l, wantHome)
			}
		}

		// Every call ran from a scratch working directory outside the repository.
		for _, cwd := range r.lines("cwd: ") {
			resolved, _ := filepath.EvalSymlinks(cwd)
			if resolved == wantRoot || strings.HasPrefix(resolved, wantRoot+string(os.PathSeparator)) {
				t.Errorf("a call ran from %q, inside the repository", cwd)
			}
		}
	})

	t.Run("a-missing-name-fails", func(t *testing.T) {
		r := runScript(t, detailsText(names[1:], "moai"), identity)
		if r.exit != 1 || !strings.Contains(r.out, "missing: "+names[0]) {
			t.Errorf("exit=%d out=%q, want exit 1 and a missing: %s line", r.exit, r.out, names[0])
		}
	})

	t.Run("a-wrong-mcp-server-fails", func(t *testing.T) {
		r := runScript(t, detailsText(names, "other"), identity)
		if r.exit != 1 || !strings.Contains(r.out, "MCP servers") || len(r.lines("argv: ")) != 3 {
			t.Errorf("exit=%d out=%q verbs=%d, want exit 1 naming the MCP servers line after all three verbs ran", r.exit, r.out, len(r.lines("argv: ")))
		}
	})

	t.Run("an-unparsable-inventory-fails", func(t *testing.T) {
		r := runScript(t, "no inventory here\n", identity)
		if r.exit != 1 || !strings.Contains(r.out, "cannot parse") || strings.Contains(r.out, "ok:") {
			t.Errorf("exit=%d out=%q, want exit 1 saying it cannot parse the inventory", r.exit, r.out)
		}
	})
}
