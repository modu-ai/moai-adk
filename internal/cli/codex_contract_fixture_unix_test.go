//go:build !windows

package cli

// codex_contract_fixture_unix_test.go — real containment fixtures only
// non-Windows platforms can build: named pipes (mkfifo) and unix sockets.
// The windows twin (codex_contract_fixture_windows_test.go) returns
// errCodexFixtureUnsupported so those fixture kinds are SKIPPED and LISTED
// (AC-CI-011's windows floor runs on the mode-injection, directory, and
// `..`-escape axes).

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func makeCodexFIFOFixture(path string) error {
	return syscall.Mkfifo(path, 0o600)
}

// makeCodexSocketFixtureAt creates a unix socket FILE at path (listener
// closed — the file must remain for Lstat to see ModeSocket). When the path
// exceeds the kernel's sun_path limit (long t.TempDir names), the socket is
// created at a short root and moved into place — a rename of the socket
// entry keeps its type.
func makeCodexSocketFixtureAt(path string) (func(), error) {
	listener, err := net.Listen("unix", path)
	if err == nil {
		if cerr := closeUnixListenerKeepFile(listener); cerr != nil {
			return nil, cerr
		}
		return func() { _ = os.Remove(path) }, nil
	}
	short := filepath.Join(os.TempDir(), fmt.Sprintf("codex-sock-%d-%d", os.Getpid(), time.Now().UnixNano()))
	listener2, err2 := net.Listen("unix", short)
	if err2 != nil {
		return nil, err // surface the ORIGINAL, path-specific failure
	}
	if cerr := closeUnixListenerKeepFile(listener2); cerr != nil {
		_ = os.Remove(short)
		return nil, cerr
	}
	if rerr := os.Rename(short, path); rerr != nil {
		_ = os.Remove(short)
		return nil, rerr
	}
	return func() { _ = os.Remove(path) }, nil
}

// closeUnixListenerKeepFile closes the listener WITHOUT unlinking the socket
// file — Go removes a socket it created on Close by default, and the fixture
// needs the entry to remain so Lstat sees ModeSocket.
func closeUnixListenerKeepFile(listener net.Listener) error {
	if ul, ok := listener.(*net.UnixListener); ok {
		ul.SetUnlinkOnClose(false)
	}
	return listener.Close()
}
