package homestate

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

// AC-019 (i) — the btime record survives a preceding line longer than the
// bufio.Scanner default 64 KiB token limit. The intr line reaches that length
// on hosts with many interrupt sources.
func TestProcStatBootTimeFindsBtimeAfterLongLine(t *testing.T) {
	longIntr := "intr " + strings.Repeat("0 ", 128*1024) // 256 KiB + prefix
	src := "cpu  1 2 3 4\n" + longIntr + "\nctxt 42\nbtime 1789526916\nprocesses 7\n"

	boot, err := procStatBootTime(strings.NewReader(src))
	if err != nil {
		t.Fatalf("procStatBootTime: %v, want the btime after the %d-byte intr line", err, len(longIntr))
	}
	if got := boot.Unix(); got != 1789526916 {
		t.Fatalf("boot = %d, want 1789526916", got)
	}
}

// AC-019 (ii) — a read that fails before the btime record reports the boot time
// unavailable with that read error as the cause.
func TestProcStatBootTimeReportsReadError(t *testing.T) {
	readErr := errors.New("procfs read failed mid-stream")
	src := io.MultiReader(strings.NewReader("cpu  1 2 3 4\nctxt 42\n"), iotest.ErrReader(readErr))

	boot, err := procStatBootTime(src)
	if err == nil {
		t.Fatalf("procStatBootTime = %v, nil; want the read error", boot)
	}
	if !errors.Is(err, readErr) {
		t.Fatalf("cause = %v, want the read error %v", err, readErr)
	}
	if errors.Is(err, errProcStatNoBtime) {
		t.Fatalf("cause = %v reads as \"no btime record\"; a failed read must stay distinguishable", err)
	}
}

// AC-019 (iii) — a source read to its end without a btime record is unavailable
// with a cause distinct from a read error.
func TestProcStatBootTimeWithoutBtimeIsUnavailable(t *testing.T) {
	for name, src := range map[string]string{
		"no btime line":           "cpu  1 2 3 4\nctxt 42\nprocesses 7\n",
		"empty source":            "",
		"btime only as substring": "cpu btime 1789526916\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := procStatBootTime(strings.NewReader(src))
			if !errors.Is(err, errProcStatNoBtime) {
				t.Fatalf("cause = %v, want errProcStatNoBtime", err)
			}
		})
	}
}

// AC-019 (iv) — a malformed or non-positive btime value is unavailable.
func TestProcStatBootTimeRejectsMalformedBtime(t *testing.T) {
	for name, line := range map[string]string{
		"not a number": "btime soon",
		"zero":         "btime 0",
		"negative":     "btime -5",
		"extra field":  "btime 1789526916 extra",
		"no value":     "btime",
	} {
		t.Run(name, func(t *testing.T) {
			boot, err := procStatBootTime(strings.NewReader("cpu  1 2\n" + line + "\n"))
			if !errors.Is(err, errProcStatMalformedBtime) {
				t.Fatalf("procStatBootTime = (%v, %v), want errProcStatMalformedBtime", boot, err)
			}
		})
	}
}

// A final btime line without a trailing newline is still a complete record.
func TestProcStatBootTimeAcceptsUnterminatedFinalLine(t *testing.T) {
	boot, err := procStatBootTime(strings.NewReader("cpu  1 2\nbtime 1789526916"))
	if err != nil || boot.Unix() != 1789526916 {
		t.Fatalf("procStatBootTime = (%v, %v), want 1789526916", boot, err)
	}
}
