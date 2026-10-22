package toolpolicy

import (
	"os"
	"os/exec"
	"testing"
)

func TestWindowsRecordLockProbePreservesLiveAndUnknownHolders(t *testing.T) {
	for _, pid := range []int{os.Getpid(), 0, -1} {
		if pidProvablyDead(pid) {
			t.Errorf("live or unidentifiable holder pid=%d was declared provably dead", pid)
		}
	}
	if uint64(^uint(0)>>1) > uint64(^uint32(0)) {
		largePID := uint64(^uint32(0)) + 1
		tooLarge := int(largePID)
		if pidProvablyDead(tooLarge) {
			t.Error("a PID that cannot be represented by the native API was declared dead")
		}
	}
}

func TestWindowsRecordLockProbeRecognizesExitedChild(t *testing.T) {
	child := exec.Command(os.Args[0], "-test.run=^$")
	child.Env = append(os.Environ(), "MOAI_TOOLPOLICY_LOCK_CHILD=exit")
	if err := child.Run(); err != nil {
		t.Fatal(err)
	}
	if !pidProvablyDead(child.Process.Pid) {
		t.Fatalf("completed child pid=%d was not recognized as provably dead", child.Process.Pid)
	}
}
