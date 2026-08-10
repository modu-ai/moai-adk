package quality

// gate_python_test.go — Python toolchain hardening (issue #1265 adjacent).
//
// The pytest step invoked a bare `pytest` off $PATH. A project using a .venv,
// uv, or poetry has no global pytest, so the optional step silently no-opped
// (or resolved an interpreter belonging to a different project). The Python
// lint steps also lacked changedExts scoping, unlike the C#/.NET entry, so
// ruff and mypy ran on commits touching no Python at all.

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// pythonToolchain returns the Python entry from the package-level table.
func pythonToolchain(t *testing.T) *langToolchain {
	t.Helper()
	for i := range toolchains {
		if len(toolchains[i].markerFiles) > 0 && toolchains[i].markerFiles[0] == "pyproject.toml" {
			return &toolchains[i]
		}
	}
	t.Fatal("Python toolchain not found in the toolchains table")
	return nil
}

// venvPytestRelPath is the platform-appropriate in-venv pytest path.
func venvPytestRelPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(".venv", "Scripts", "pytest.exe")
	}
	return filepath.Join(".venv", "bin", "pytest")
}

// writeFile writes content at dir/rel, creating parents.
func writeFile(t *testing.T, dir, rel, content string) string {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestResolvePytestRunner_PrefersProjectLocalRunners exercises the resolution
// order: .venv → uv → poetry → bare pytest.
func TestResolvePytestRunner_PrefersProjectLocalRunners(t *testing.T) {
	base := gateStep{name: pytestStepName, binary: "pytest", args: nil, optional: true}

	t.Run(".venv pytest wins over everything", func(t *testing.T) {
		dir := t.TempDir()
		venv := writeFile(t, dir, venvPytestRelPath(), "#!/bin/sh\n")
		writeFile(t, dir, "uv.lock", "")
		writeFile(t, dir, "poetry.lock", "")

		got := resolvePytestRunner(base, dir)
		if got.binary != venv {
			t.Errorf("binary = %q, want the in-venv pytest %q", got.binary, venv)
		}
		if len(got.args) != 0 {
			t.Errorf("args = %v, want none", got.args)
		}
		// A resolved project-local runner must not be skipped by the optional
		// $PATH LookPath check — the binary is an absolute path, not a $PATH name.
		if got.optional {
			t.Errorf("resolved in-venv step is still optional; the LookPath skip would " +
				"discard a runner we already proved exists")
		}
	})

	t.Run("uv.lock selects `uv run pytest`", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "uv.lock", "")
		writeFile(t, dir, "poetry.lock", "")

		got := resolvePytestRunner(base, dir)
		if got.binary != "uv" {
			t.Errorf("binary = %q, want \"uv\"", got.binary)
		}
		if len(got.args) != 2 || got.args[0] != "run" || got.args[1] != "pytest" {
			t.Errorf("args = %v, want [run pytest]", got.args)
		}
	})

	t.Run("[tool.uv] in pyproject.toml selects uv", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "pyproject.toml", "[project]\nname = \"x\"\n\n[tool.uv]\ndev-dependencies = []\n")

		got := resolvePytestRunner(base, dir)
		if got.binary != "uv" {
			t.Errorf("binary = %q, want \"uv\"", got.binary)
		}
	})

	t.Run("poetry.lock selects `poetry run pytest`", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "poetry.lock", "")

		got := resolvePytestRunner(base, dir)
		if got.binary != "poetry" {
			t.Errorf("binary = %q, want \"poetry\"", got.binary)
		}
		if len(got.args) != 2 || got.args[0] != "run" || got.args[1] != "pytest" {
			t.Errorf("args = %v, want [run pytest]", got.args)
		}
	})

	t.Run("no project-local marker falls back to bare pytest", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "pyproject.toml", "[project]\nname = \"x\"\n")

		got := resolvePytestRunner(base, dir)
		if got.binary != "pytest" {
			t.Errorf("binary = %q, want the bare \"pytest\" fallback", got.binary)
		}
		if !got.optional {
			t.Errorf("bare pytest fallback must stay optional so a missing global " +
				"pytest skips rather than fails the commit")
		}
	})

	t.Run("step name is preserved so the exit-5 handling still applies", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "uv.lock", "")
		if got := resolvePytestRunner(base, dir); got.name != pytestStepName {
			t.Errorf("name = %q, want %q — runStep keys the exit-5 "+
				"no-tests-collected handling off the step name", got.name, pytestStepName)
		}
	})
}

// TestResolvePytestRunner_NonPytestStepUnchanged guards the transform's scope.
func TestResolvePytestRunner_NonPytestStepUnchanged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "uv.lock", "")
	base := gateStep{name: "go test", binary: "go", args: []string{"test", "./..."}}

	got := resolvePytestRunner(base, dir)
	if got.binary != "go" || len(got.args) != 2 {
		t.Errorf("non-pytest step was rewritten: %+v", got)
	}
}

// TestPythonLintSteps_ScopedToPyExtension asserts the ruff and mypy steps carry
// `.py` changedExts scoping, matching the `.cs` precedent on the .NET entry.
func TestPythonLintSteps_ScopedToPyExtension(t *testing.T) {
	tc := pythonToolchain(t)
	if len(tc.lintSteps) == 0 {
		t.Fatal("Python toolchain has no lint steps")
	}
	for _, step := range tc.lintSteps {
		if len(step.changedExts) == 0 {
			t.Errorf("lint step %q has no changedExts; it runs on commits touching "+
				"no Python at all (the .cs precedent on the .NET entry scopes this)", step.name)
			continue
		}
		found := false
		for _, ext := range step.changedExts {
			if ext == ".py" {
				found = true
			}
		}
		if !found {
			t.Errorf("lint step %q changedExts = %v, want to include \".py\"", step.name, step.changedExts)
		}
	}
}

// TestPythonToolchain_ResolvedByDetect asserts detectToolchain returns a Python
// toolchain whose test step has been resolved to the project-local runner.
func TestPythonToolchain_ResolvedByDetect(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[project]\nname = \"x\"\n")
	writeFile(t, dir, "uv.lock", "")

	g := &QualityGate{config: &GateConfig{Enabled: true, ProjectDir: dir}}
	tc := g.detectToolchain()
	if tc == nil {
		t.Fatal("detectToolchain returned nil for a pyproject.toml project")
	}
	if tc.testStep == nil {
		t.Fatal("Python toolchain has no test step")
	}
	if tc.testStep.binary != "uv" {
		t.Errorf("detectToolchain did not resolve the pytest runner: binary = %q, want \"uv\"",
			tc.testStep.binary)
	}

	// The package-level table must not have been mutated.
	if pythonToolchain(t).testStep.binary != "pytest" {
		t.Error("resolution mutated the package-level toolchains table")
	}
}
