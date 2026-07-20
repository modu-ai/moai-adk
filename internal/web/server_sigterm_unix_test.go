//go:build !windows

package web

import (
	"context"
	"sync"
	"syscall"
	"testing"
	"time"
)

// TestServer_GracefulShutdownOnSIGTERM verifies that delivering a real SIGTERM to
// the process triggers the same clean shutdown path. The signal is caught by
// signal.NotifyContext inside ListenAndServe.
//
// Lives in a !windows file: syscall.Kill does not exist on Windows, so keeping
// this test in the shared file breaks `go vet`/compilation on windows runners.
func TestServer_GracefulShutdownOnSIGTERM(t *testing.T) {
	srv, err := NewServer(newTestConfig(t))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	var wg sync.WaitGroup
	var serveErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		serveErr = srv.ListenAndServe(context.Background())
	}()

	waitForAddr(t, srv)

	// Deliver SIGTERM to our own process; signal.NotifyContext cancels.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Skipf("cannot deliver SIGTERM on this platform: %v", err)
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		if serveErr != nil {
			t.Errorf("ListenAndServe returned %v on SIGTERM, want nil", serveErr)
		}
	case <-time.After(shutdownDrain + 2*time.Second):
		t.Fatal("ListenAndServe did not return after SIGTERM")
	}
}
