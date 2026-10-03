//go:build ignore

// entry_probe.go measures the launcher entry parse of the tree it is pointed at WITHOUT adding a file to
// that tree: the probe test (entry_probe_test.go.txt, committed beside this file as plain data) is mapped
// into package internal/cli by a `go test -overlay` run, so the checkout is never written.
//
//	go run entry_probe.go -root <repo checkout> [-probe <path to entry_probe_test.go.txt>]
//
// Output: the verbatim `go test -v` output of TestZZT1399EntryProbe (CCPARSE rows for cc/glm via
// parseLauncherEntry, CODEX rows via runCodex) followed by one `exit <code>` line.
// It backs ledger rows RED-11, RED-12 and RED-13 of acceptance.md.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	root := flag.String("root", "", "repository checkout to measure (read only)")
	probe := flag.String("probe", "", "probe test source (default: entry_probe_test.go.txt beside this file)")
	flag.Parse()
	if *root == "" {
		fmt.Fprintln(os.Stderr, "need -root")
		os.Exit(2)
	}
	if *probe == "" {
		_, self, _, _ := runtime.Caller(0)
		*probe = filepath.Join(filepath.Dir(self), "entry_probe_test.go.txt")
	}
	absRoot, err := filepath.Abs(*root)
	check(err)
	tmp, err := os.MkdirTemp("", "t1399-entry-probe-")
	check(err)
	defer os.RemoveAll(tmp)
	body, err := os.ReadFile(*probe)
	check(err)
	probeGo := filepath.Join(tmp, "zz_t1399_probe_test.go")
	check(os.WriteFile(probeGo, body, 0o644))
	overlay := filepath.Join(tmp, "overlay.json")
	// the overlay key is the path the file WOULD have inside the package; the file does not exist there
	key := filepath.Join(absRoot, "internal", "cli", "zz_t1399_probe_test.go")
	check(os.WriteFile(overlay, []byte(fmt.Sprintf(`{"Replace":{%q:%q}}`, key, probeGo)), 0o644))
	cmd := exec.Command("go", "test", "-overlay", overlay, "-run", "TestZZT1399EntryProbe", "-count=1", "-v", "./internal/cli/")
	cmd.Dir = absRoot
	out, err := cmd.CombinedOutput()
	fmt.Print(string(out))
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = 127
		}
	}
	fmt.Printf("exit %d\n", code)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
