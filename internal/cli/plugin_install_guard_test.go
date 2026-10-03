package cli

// plugin_install_guard_test.go — SPEC-PLUGIN-MARKETPLACE-001 M3a, AC-016: the
// automated callers of `moai init` set the plugin opt-out, and a new caller
// cannot appear unnoticed. (TestInstallScriptsPluginStepGuarded, AC-018 c,
// reads the install scripts and lands with them in M3b.)

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// pluginOptOutInitCallers lists the non-test Go files that start the moai
// binary's `init` as a child process. Each must set the opt-out.
var pluginOptOutInitCallers = map[string]bool{
	"doctor_agentemit_embed.go": true,
}

// initExecCallers returns the base names of the files in src (name -> source)
// holding an exec.Command / exec.CommandContext call whose argument list
// carries the literal "init" and whose command is not the literal "git".
func initExecCallers(t *testing.T, sources map[string]string) []string {
	t.Helper()
	var found []string
	for name, src := range sources {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		hit := false
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "exec" || (sel.Sel.Name != "Command" && sel.Sel.Name != "CommandContext") {
				return true
			}
			args := call.Args
			if sel.Sel.Name == "CommandContext" && len(args) > 0 {
				args = args[1:]
			}
			if len(args) == 0 {
				return true
			}
			if lit, ok := args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if v, _ := strconv.Unquote(lit.Value); v == "git" {
					return true
				}
			}
			for _, a := range args[1:] {
				if lit, ok := a.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if v, _ := strconv.Unquote(lit.Value); v == "init" {
						hit = true
					}
				}
			}
			return true
		})
		if hit {
			found = append(found, filepath.Base(name))
		}
	}
	sort.Strings(found)
	return found
}

func TestPluginOptOutCallersEnumerated(t *testing.T) {
	// Positive control: the scan flags a moai init caller and passes git init.
	control := map[string]string{
		"caller.go": "package x\nimport \"os/exec\"\nfunc f(bin string){ _ = exec.Command(bin, \"init\", \"p\") }\n",
		"git.go":    "package x\nimport \"os/exec\"\nfunc g(){ _ = exec.Command(\"git\", \"init\") }\n",
		"ctx.go":    "package x\nimport (\"context\";\"os/exec\")\nfunc h(bin string){ _ = exec.CommandContext(context.Background(), bin, \"init\") }\n",
		"other.go":  "package x\nimport \"os/exec\"\nfunc o(bin string){ _ = exec.Command(bin, \"version\") }\n",
	}
	if got := strings.Join(initExecCallers(t, control), ","); got != "caller.go,ctx.go" {
		t.Fatalf("scan positive control = %q, want caller.go,ctx.go", got)
	}

	// Real sweep over the non-test Go sources under internal/.
	sources := map[string]string{}
	walkErr := filepath.WalkDir(filepath.Join("..", "..", "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		sources[path] = string(b)
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk internal/: %v", walkErr)
	}
	if len(sources) < 100 {
		t.Fatalf("empty sweep: only %d non-test sources scanned", len(sources))
	}
	got := initExecCallers(t, sources)
	seen := map[string]bool{}
	for _, name := range got {
		seen[name] = true
		if !pluginOptOutInitCallers[name] {
			t.Errorf("%s starts `init` as a child process but is not in the enumerated callers; add it to the list and set %s",
				name, config.EnvSkipPluginInstall)
		}
	}
	for name := range pluginOptOutInitCallers {
		if !seen[name] {
			t.Errorf("enumerated caller %s no longer starts `init` (stale list entry)", name)
		}
		for path, src := range sources {
			if filepath.Base(path) == name && !strings.Contains(src, "EnvSkipPluginInstall") {
				t.Errorf("%s starts `init` but does not reference config.EnvSkipPluginInstall", path)
			}
		}
	}

	// The e2e journeys: every non-comment line that runs '$BIN' init carries
	// the opt-out on that same line.
	script, err := os.ReadFile(filepath.Join("..", "..", "e2e", "cli", "tux3_journeys.sh"))
	if err != nil {
		t.Fatalf("read tux3_journeys.sh: %v", err)
	}
	runs := 0
	for i, line := range strings.Split(string(script), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || !strings.Contains(line, "'$BIN' init") {
			continue
		}
		runs++
		if !strings.Contains(line, config.EnvSkipPluginInstall) {
			t.Errorf("tux3_journeys.sh:%d runs `'$BIN' init` without %s: %s", i+1, config.EnvSkipPluginInstall, trimmed)
		}
	}
	if runs < 2 {
		t.Fatalf("empty sweep: found %d `'$BIN' init` runs in tux3_journeys.sh, want at least 2", runs)
	}
}

// TestDoctorAgentEmitEmbed_SetsPluginOptOut: the `init` child the embed check
// starts inherits the opt-out (AC-016 a). A recording script stands in for the
// binary and dumps its environment.
func TestDoctorAgentEmitEmbed_SetsPluginOptOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("recording shell script is POSIX-only")
	}
	t.Setenv(config.EnvSkipPluginInstall, "") // the caller's ambient value must not matter
	dump := filepath.Join(t.TempDir(), "env.dump")
	script := filepath.Join(t.TempDir(), "fake-moai")
	body := "#!/bin/sh\nenv > '" + dump + "'\nexit 0\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	_, cleanup, err := extractEmissionViaInit(script)
	defer cleanup()
	if err != nil {
		t.Fatalf("extractEmissionViaInit: %v", err)
	}
	env, rerr := os.ReadFile(dump)
	if rerr != nil {
		t.Fatalf("the child never ran (no env dump): %v", rerr)
	}
	if !bytes.Contains(env, []byte("\n"+config.EnvSkipPluginInstall+"=1\n")) && !bytes.HasPrefix(env, []byte(config.EnvSkipPluginInstall+"=1\n")) {
		t.Fatalf("the init child did not receive %s=1:\n%s", config.EnvSkipPluginInstall, firstLinesMatching(string(env), "MOAI_"))
	}
}

func firstLinesMatching(s, prefix string) string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.HasPrefix(l, prefix) {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}
