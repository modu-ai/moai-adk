package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
)

// newTestReporter builds a printerReporter over a plain-mode printer with
// captured stdout / stderr buffers.
func newTestReporter() (*printerReporter, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	p := printer.New(
		printer.WithWriters(out, errBuf),
		printer.WithMode(printer.ModePlain),
	)
	return newPrinterReporter(p), out, errBuf
}

// TestPrinterReporterRoutesToStderr verifies the ProgressReporter adapter
// routes every step event through the Printer to stderr (REQ-CTX-015): the
// former project.ConsoleReporter wrote to stdout via fmt.Printf.
func TestPrinterReporterRoutesToStderr(t *testing.T) {
	r, out, errBuf := newTestReporter()

	r.StepStart("Detection", "Analyzing project structure")
	r.StepUpdate("halfway")
	r.StepComplete("Detected project structure")

	r.StepStart("Validation", "")
	r.StepError(errors.New("bad yaml"))

	if out.Len() != 0 {
		t.Errorf("reporter must not write to stdout, got %q", out.String())
	}
	errOut := errBuf.String()
	for _, want := range []string{
		"Detection: Analyzing project structure...",
		"halfway",
		"✓ Detected project structure",
		"Validation...",
		"✗ Error: bad yaml",
	} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr missing %q; got %q", want, errOut)
		}
	}
}

// TestPrinterReporterUnpairedCalls verifies graceful degradation when
// PhaseExecutor-style pairing is violated: no panic, status still lands
// on stderr (the legacy ConsoleReporter was stateless and tolerated this).
func TestPrinterReporterUnpairedCalls(t *testing.T) {
	r, out, errBuf := newTestReporter()

	r.StepUpdate("orphan update")
	r.StepComplete("orphan complete")
	r.StepComplete("") // empty orphan completion is silently ignored
	r.StepError(errors.New("orphan error"))

	if out.Len() != 0 {
		t.Errorf("reporter must not write to stdout, got %q", out.String())
	}
	errOut := errBuf.String()
	for _, want := range []string{"orphan update", "orphan complete", "orphan error"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr missing %q; got %q", want, errOut)
		}
	}

	// Empty completion message after a started step falls back to a
	// generic label.
	r.StepStart("S", "")
	r.StepComplete("")
	if !strings.Contains(errBuf.String(), "✓ Completed") {
		t.Errorf("empty completion must render generic label, got %q", errBuf.String())
	}
}
