//go:build unix

package cli

// jev_cost_fifo_unix_test.go — sync audit round 1 P2 of card t1971 (receipt
// rcpt-11b837f9258ea3ca20d0620e): a FIFO at the decision-ledger path must not
// hang `moai jev cost` forever. The report's ledger open is non-blocking and
// its handle is checked to be a regular file; a non-regular ledger path is
// treated as absent (the REQ-JEVL-016 absent class — zero calls, exit 0).
// syscall.Mkfifo is unix-only, so this file rides the same build constraint as
// the production open in jev_cost_open_unix.go (the read-side mirror of the
// recorder's openLedger). The 5 s guard is a hang guard for the test, not a
// timing claim about the report; go test -timeout bounds the regression from
// outside as the loud backstop.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestJevCost_FIFOAtLedgerPathDoesNotHang(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "state", "jev")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(dir, "decisions.jsonl"), 0o600); err != nil {
		t.Skipf("syscall.Mkfifo unavailable here: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)
	// The flag value is reset before the run because pflag keeps values set
	// by an earlier Execute on the same command object (runJevCostCmd).
	if err := costCmd.Flags().Set("json", "false"); err != nil {
		t.Fatal(err)
	}

	// The command runs on a spawned goroutine so a regression (a blocking
	// open) fails through the 5 s guard instead of hanging the test binary;
	// t.Fatal is never called inside the goroutine.
	done := make(chan string, 1)
	go func() {
		var out bytes.Buffer
		rootCmd.SetOut(&out)
		rootCmd.SetErr(&out)
		rootCmd.SetArgs([]string{"jev", "cost"})
		if err := rootCmd.Execute(); err != nil {
			done <- fmt.Sprintf("Execute error: %v", err)
			return
		}
		done <- out.String()
	}()
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
	})

	select {
	case out := <-done:
		if !strings.Contains(out, "calls: 0") {
			t.Errorf("a FIFO at the ledger path must report calls: 0; got:\n%s", out)
		}
		if strings.Contains(out, "skipped") {
			t.Errorf("a FIFO ledger carries no malformed lines; got:\n%s", out)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("moai jev cost blocked on a FIFO at the ledger path: the report's ledger read must be non-blocking")
	}
}
