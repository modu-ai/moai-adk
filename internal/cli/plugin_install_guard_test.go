package cli

// plugin_install_guard_test.go — SPEC-PLUGIN-MARKETPLACE-001 M3a, AC-016: the
// automated callers of `moai init` set the plugin opt-out, and a new caller
// cannot appear unnoticed. M3b, AC-018 (c): TestInstallScriptsPluginStepGuarded
// reads the install scripts (static: it proves a shape is present, not that the
// script behaves; the behavior of install.sh is the harness's job, and
// install.ps1 and install.bat are not executed locally).

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

// --- AC-018 (c): the install scripts call the verb by the installed path, guarded ---

// scriptLines splits a script into lines with carriage returns removed.
func scriptLines(src string) []string {
	return strings.Split(strings.ReplaceAll(src, "\r", ""), "\n")
}

// braceNet is the number of "{" minus the number of "}" on a line.
func braceNet(line string) int {
	return strings.Count(line, "{") - strings.Count(line, "}")
}

// shPluginStepProblems reads install.sh: every line that runs `plugin install`
// uses "$TARGET_PATH" (a bare moai is not on PATH when the install directory is
// not), and sits in an `if` or an `||` list, which `set -e` does not abort on.
func shPluginStepProblems(src string) []string {
	var problems []string
	calls := 0
	for _, raw := range scriptLines(src) {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "echo") ||
			strings.HasPrefix(line, "print_") || !strings.Contains(line, "plugin install") {
			continue
		}
		if !strings.Contains(line, `"$TARGET_PATH" plugin install`) {
			problems = append(problems, "call without the installed path: "+line)
			continue
		}
		calls++
		if !strings.HasPrefix(line, "if ") && !strings.Contains(line, "||") {
			problems = append(problems, "unguarded call under set -e: "+line)
		}
	}
	if calls == 0 {
		problems = append(problems, `no non-comment line runs "$TARGET_PATH" plugin install`)
	}
	return problems
}

// ps1PluginStepProblems reads install.ps1: the call is `& $targetPath plugin
// install` after the install step assigned $targetPath, inside a try whose
// catch neither rethrows nor exits, because $ErrorActionPreference = "Stop"
// turns a thrown error into an abort of the whole installer.
func ps1PluginStepProblems(src string) []string {
	lines := scriptLines(src)
	var problems []string
	call, assigned := -1, -1
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if assigned < 0 && strings.HasPrefix(line, "$targetPath =") {
			assigned = i
		}
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "Write-Host") ||
			strings.HasPrefix(line, "Print-") || !strings.Contains(line, "plugin install") {
			continue
		}
		if !strings.HasPrefix(line, "& $targetPath plugin install") {
			problems = append(problems, "call without the install path variable: "+line)
			continue
		}
		if call < 0 {
			call = i
		}
	}
	if call < 0 {
		return append(problems, "no non-comment line runs & $targetPath plugin install")
	}
	if assigned < 0 || assigned > call {
		problems = append(problems, "$targetPath is not assigned before the call")
	}
	// the nearest enclosing try block of the call
	tryAt, bal, level := -1, 0, 0
	for i := call - 1; i >= 0 && tryAt < 0; i-- {
		bal += braceNet(lines[i])
		for bal > level {
			level++
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "try") {
				tryAt = i
			}
		}
	}
	if tryAt < 0 {
		return append(problems, "the call is not inside a try block")
	}
	end, bal := tryAt, 0
	for ; end < len(lines); end++ {
		bal += braceNet(lines[end])
		if bal <= 0 {
			break
		}
	}
	catchAt := -1
	for j := end; j < len(lines) && j <= end+2; j++ {
		if strings.Contains(lines[j], "catch") {
			catchAt = j
			break
		}
	}
	if catchAt < 0 {
		return append(problems, "the try block has no catch")
	}
	text := lines[catchAt][strings.Index(lines[catchAt], "catch"):]
	body, depth := []string{text}, braceNet(text)
	for j := catchAt + 1; depth > 0 && j < len(lines); j++ {
		body = append(body, lines[j])
		depth += braceNet(lines[j])
	}
	for _, l := range body {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "throw") || strings.HasPrefix(t, "exit") {
			problems = append(problems, "the catch rethrows or exits: "+t)
		}
	}
	return problems
}

// batPluginStepProblems reads install.bat: the call is "%TARGET_PATH%" plugin
// install, and the first thing that ends the script after it is `exit /b 0`
// (a bare `goto :eof` would hand back the call's errorlevel), with no
// `exit /b` after it that passes a level on.
func batPluginStepProblems(src string) []string {
	lines := scriptLines(src)
	var problems []string
	call := -1
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		low := strings.ToLower(line)
		if line == "" || strings.HasPrefix(low, "rem ") || strings.HasPrefix(line, "::") ||
			strings.Contains(low, "echo") || !strings.Contains(line, "plugin install") {
			continue
		}
		if !strings.HasPrefix(line, `"%TARGET_PATH%" plugin install`) {
			problems = append(problems, "call without the install path: "+line)
			continue
		}
		if call < 0 {
			call = i
		}
	}
	if call < 0 {
		return append(problems, `no non-comment line runs "%TARGET_PATH%" plugin install`)
	}
	firstEnd := ""
	for _, raw := range lines[call+1:] {
		low := strings.ToLower(strings.TrimSpace(raw))
		isExit := strings.HasPrefix(low, "exit /b")
		if firstEnd == "" && (isExit || strings.HasPrefix(low, "goto :eof")) {
			firstEnd = low
		}
		if isExit && strings.TrimSpace(strings.TrimPrefix(low, "exit /b")) != "0" {
			problems = append(problems, "an exit after the call passes a level on: "+low)
		}
	}
	if firstEnd != "exit /b 0" {
		problems = append(problems, "the first script end after the call is "+strconv.Quote(firstEnd)+", want exit /b 0")
	}
	return problems
}

func TestInstallScriptsPluginStepGuarded(t *testing.T) {
	// Mutants: each checker must report a problem for a script that satisfies a
	// naive grep for the verb call and violates the requirement, and none for
	// the well-formed shape (the positive control).
	goodSh := "set -e\nmain() {\n    if ! \"$TARGET_PATH\" plugin install; then\n        print_warning \"x\"\n    fi\n}\n"
	goodPs1 := "function Main {\n    $targetPath = Install-Binary -BinaryPath $b\n    try {\n        & $targetPath plugin install\n" +
		"        if ($LASTEXITCODE -ne 0) { Print-Warning \"x\" }\n    }\n    catch {\n        Print-Warning \"x\"\n    }\n}\n"
	goodBat := "set \"TARGET_PATH=%INSTALL_DIR%\\moai.exe\"\n\"%TARGET_PATH%\" version\n\"%TARGET_PATH%\" plugin install\n" +
		"if errorlevel 1 echo [WARNING] x\nexit /b 0\n:show_help\nexit /b 0\n"
	mutate := func(src, from, to string) string {
		if !strings.Contains(src, from) {
			t.Fatalf("mutant fixture does not contain %q", from)
		}
		return strings.Replace(src, from, to, 1)
	}
	t.Run("checkers-detect-mutants", func(t *testing.T) {
		for _, c := range []struct {
			name   string
			check  func(string) []string
			src    string
			wantOK bool
		}{
			{"sh-good", shPluginStepProblems, goodSh, true},
			{"sh-bare-moai", shPluginStepProblems, mutate(goodSh, `"$TARGET_PATH" plugin install`, "moai plugin install"), false},
			{"sh-unguarded", shPluginStepProblems, mutate(goodSh, "if ! \"$TARGET_PATH\" plugin install; then", "\"$TARGET_PATH\" plugin install\n    if false; then"), false},
			{"sh-comment-only", shPluginStepProblems, "set -e\n# \"$TARGET_PATH\" plugin install || true\nprint_info \"run moai plugin install\"\n", false},
			{"ps1-good", ps1PluginStepProblems, goodPs1, true},
			{"ps1-bare-moai", ps1PluginStepProblems, mutate(goodPs1, "& $targetPath plugin install", "& moai plugin install"), false},
			{"ps1-no-try", ps1PluginStepProblems, mutate(mutate(goodPs1, "    try {\n", ""), "    }\n    catch {\n        Print-Warning \"x\"\n    }\n", ""), false},
			{"ps1-catch-rethrows", ps1PluginStepProblems, mutate(goodPs1, "catch {\n        Print-Warning \"x\"", "catch {\n        throw"), false},
			{"ps1-catch-exits", ps1PluginStepProblems, mutate(goodPs1, "catch {\n        Print-Warning \"x\"", "catch {\n        exit 1"), false},
			{"bat-good", batPluginStepProblems, goodBat, true},
			{"bat-bare-moai", batPluginStepProblems, mutate(goodBat, "\"%TARGET_PATH%\" plugin install", "moai plugin install"), false},
			{"bat-exit-propagates", batPluginStepProblems, mutate(goodBat, "exit /b 0\n:show_help", "exit /b %errorlevel%\n:show_help"), false},
			{"bat-goto-eof-ends", batPluginStepProblems, mutate(goodBat, "exit /b 0\n:show_help", "goto :eof\n:show_help"), false},
		} {
			got := c.check(c.src)
			if c.wantOK && len(got) != 0 {
				t.Errorf("%s: well-formed shape rejected: %v", c.name, got)
			}
			if !c.wantOK && len(got) == 0 {
				t.Errorf("%s: mutant accepted, the checker cannot fail on it", c.name)
			}
		}
	})

	t.Run("real-scripts", func(t *testing.T) {
		read := func(rel string) string {
			b, err := os.ReadFile(filepath.Join("..", "..", rel))
			if err != nil {
				t.Fatalf("read %s: %v", rel, err)
			}
			if len(b) < 1000 {
				t.Fatalf("empty sweep: %s has %d bytes", rel, len(b))
			}
			return string(b)
		}
		for rel, check := range map[string]func(string) []string{
			"install.sh":  shPluginStepProblems,
			"install.ps1": ps1PluginStepProblems,
			"install.bat": batPluginStepProblems,
		} {
			for _, p := range check(read(rel)) {
				t.Errorf("%s: %s", rel, p)
			}
		}
		for _, name := range []string{"install.sh", "install.ps1"} {
			if read(name) != read(filepath.Join("docs-site", "static", name)) {
				t.Errorf("docs-site/static/%s differs from the root %s", name, name)
			}
		}
	})
}
