package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestProfileSandbox_HelperSubprocessLeavesNoBaseDir pins the profile sandbox
// against the one exit path that bypasses TestMain's cleanup.
//
// Several tests in this package re-execute the test binary to get a real
// subprocess (exitcode_guard_test.go, todo_test.go, launch_session_pid_exec_posix_test.go).
// Each child runs TestMain, so each child used to mint its own
// "moai-cli-profiles-*" directory — and every one of those helper bodies ends
// in os.Exit, which returns through neither m.Run() nor the
// restoreProfileBaseDir() call that follows it. The directory was therefore
// created by design and removed by nobody: tens of thousands accumulated in
// TMPDIR, one per helper invocation, until a plain readdir of the temp
// directory took seconds.
//
// The child is given its own TMPDIR so this assertion measures only what this
// test caused — no global count, nothing another lane's concurrent run can
// perturb.
func TestProfileSandbox_HelperSubprocessLeavesNoBaseDir(t *testing.T) {
	childTmp := t.TempDir()

	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "TMPDIR="+childTmp)
	// The helper exits 3 by design; a non-nil error here is the expected shape.
	_ = cmd.Run()

	leaked, err := filepath.Glob(filepath.Join(childTmp, "moai-cli-profiles-*"))
	if err != nil {
		t.Fatalf("glob child TMPDIR: %v", err)
	}
	if len(leaked) != 0 {
		t.Fatalf("helper subprocess left %d profile sandbox dir(s) behind in its TMPDIR: %v\n"+
			"the child minted its own sandbox and exited through os.Exit, so TestMain's "+
			"restoreProfileBaseDir() never ran — children must inherit the parent's sandbox instead",
			len(leaked), leaked)
	}
}
