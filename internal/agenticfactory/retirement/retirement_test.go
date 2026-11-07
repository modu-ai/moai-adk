// Package retirement holds the release gate of AC-AFL-015 (SPEC-AGENTIC-FACTORY-LOCAL-001): the
// Issue/remote code of the Agentic Factory is gone and nothing imports it. It has no code of its
// own, only the test.
package retirement

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// rootEnv points the test at another checkout of the module (for example a tree extracted from an
// earlier commit, to observe the RED); unset, the test reads the module it is part of.
const rootEnv = "MOAI_RETIREMENT_ROOT"

// retired are the package directories (relative to internal/agenticfactory) that must not exist and
// must not be imported.
var retired = []string{
	"forge", "mirror", "gitref", "resident",
	"effect/ghintake", "effect/ctxledger", "effect/labels", "effect/ghexec", "effect/ghidentity", "effect/gitexec",
}

// kept are the packages the retirement must leave alone (positive control).
var kept = []string{"ledger", "mergeq", "claim", "core", "runner"}

func moduleRoot(t *testing.T) string {
	t.Helper()
	if r := os.Getenv(rootEnv); r != "" {
		return r
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		if dir == filepath.Dir(dir) {
			t.Fatal("no go.mod above " + wd)
		}
	}
}

// goListTimeout bounds one go list so a stuck module download cannot hang the suite.
const goListTimeout = 3 * time.Minute

// goList runs go list at root and returns stdout; stderr is part of the failure message. goos, when not
// empty, sets GOOS so files that only build on another platform are listed too.
func goList(t *testing.T, root, goos string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), goListTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", append([]string{"list"}, args...)...)
	cmd.Dir = root
	if goos != "" {
		cmd.Env = append(os.Environ(), "GOOS="+goos)
	}
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		t.Fatalf("go list %v in %s: %v\n%s", args, root, err, errb.String())
	}
	return out.String()
}

// TestRetirement_NoImportOfForgeMirrorGitref (AC-AFL-015): no non-test file under internal/ imports a
// retired package, no retired package directory remains, the `moai factory resident` command is not
// registered, and the packages outside the retirement are still there.
func TestRetirement_NoImportOfForgeMirrorGitref(t *testing.T) {
	root := moduleRoot(t)
	module := strings.TrimSpace(goList(t, root, "", "-m"))
	base := module + "/internal/agenticfactory/"

	// (a) imports: go list's Imports field is the non-test import set of each package. It is read for the
	// host platform and for the release platforms, because a file built only on windows or linux can import
	// a retired package without the host listing showing it.
	for _, goos := range []string{"", "windows", "linux", "darwin"} {
		var edgesToCore int
		listing := goList(t, root, goos, "-e", "-f", `{{.ImportPath}} {{join .Imports " "}}`, "./internal/...")
		for line := range strings.SplitSeq(strings.TrimSpace(listing), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			for _, imp := range fields[1:] {
				if imp == base+"core" {
					edgesToCore++
				}
				for _, r := range retired {
					if imp == base+r || strings.HasPrefix(imp, base+r+"/") {
						t.Errorf("GOOS=%q: %s imports the retired package %s", goos, fields[0], imp)
					}
				}
			}
		}
		// Control, per platform: the scan sees import edges at all. An empty listing (go list -e swallows load
		// errors) would pass the import check vacuously, and a sum over platforms would hide one that is empty.
		if edgesToCore == 0 {
			t.Fatalf("GOOS=%q: no package imports agenticfactory/core, the import scan sees nothing", goos)
		}
	}
	// (b) and (c): a directory that is left behind fails even when nothing imports it.
	for _, r := range retired {
		if _, err := os.Stat(filepath.Join(root, "internal", "agenticfactory", filepath.FromSlash(r))); err == nil {
			t.Errorf("the retired directory internal/agenticfactory/%s still exists", r)
		}
	}

	// The factory resident command: no source file of the cli names it.
	cliDir := filepath.Join(root, "internal", "cli")
	entries, err := os.ReadDir(cliDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if strings.HasPrefix(name, "factory_resident") {
			t.Errorf("internal/cli/%s still exists", name)
		}
		src, err := os.ReadFile(filepath.Join(cliDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(src, []byte("newFactoryResidentCommand")) {
			t.Errorf("internal/cli/%s still registers the factory resident command", name)
		}
	}

	// Positive control: the packages outside the retirement remain and are part of the listing.
	listed := goList(t, root, "", "-e", "-f", "{{.ImportPath}}", "./internal/agenticfactory/...")
	paths := strings.Fields(listed)
	for _, k := range kept {
		if !slices.Contains(paths, base+k) {
			t.Errorf("the package %s must remain after the retirement", base+k)
		}
	}
}
