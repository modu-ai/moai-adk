package procrun

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

// Every case in this package drives the runner through a POSIX shell script
// (/bin/sh plus head, tr, sleep), which a Windows host does not have. The
// runner itself is platform neutral; what cannot run there is the fixture, so
// the package's tests are skipped on Windows with that reason (card t2123,
// #1969) instead of failing with "exec: /bin/sh: executable file not found".
func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		fmt.Println("SKIP: procrun tests need /bin/sh and POSIX coreutils, absent on Windows")
		os.Exit(0)
	}
	os.Exit(m.Run())
}
