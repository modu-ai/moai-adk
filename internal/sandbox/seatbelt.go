package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	// @MX:WARN: [AUTO] sandbox-exec has been deprecated by Apple (operational since macOS 10.5)
	// @MX:REASON: Apple no longer officially supports sandbox-exec and it may be
	//             removed in a future macOS version. An App Sandbox entitlement-based
	//             alternative is planned for v3.1+.
)

// SeatbeltBackend implements SandboxBackend for macOS using sandbox-exec.
// It generates SBPL profiles and wraps commands with `sandbox-exec -p <profile>`.
type SeatbeltBackend struct{}

// NewSeatbeltBackend returns a new SeatbeltBackend.
func NewSeatbeltBackend() *SeatbeltBackend {
	return &SeatbeltBackend{}
}

// Available reports whether sandbox-exec is available at /usr/bin/sandbox-exec.
func (s *SeatbeltBackend) Available() bool {
	_, err := exec.LookPath("sandbox-exec")
	return err == nil
}

// Exec runs cmd inside a macOS seatbelt sandbox with the given options.
//
// @MX:WARN: [AUTO] execSandboxExec — the SBPL profile is generated just before exec and never written to a file
// @MX:REASON: The -p flag of sandbox-exec accepts an inline profile, so no
//             tmpfile is required. However, very long profiles may exceed the
//             arg list limit. The current implementation uses -p; we can switch
//             to -f (file) mode if needed.
func (s *SeatbeltBackend) Exec(opts SandboxOptions, cmd []string) ([]byte, error) {
	if !s.Available() {
		return nil, ErrSandboxBackendUnavailable
	}
	if len(cmd) == 0 {
		return nil, fmt.Errorf("sandbox exec: empty command")
	}

	maxBytes := opts.MaxOutputBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxOutputBytes
	}

	// Generate SBPL profile
	profile, err := GenerateSBPL(opts)
	if err != nil {
		return nil, fmt.Errorf("sandbox: generate SBPL: %w", err)
	}

	// Environment variable scrubbing
	env := ScrubEnv(os.Environ(), opts.EnvPassthrough)

	// sandbox-exec -p <profile> <cmd...>
	execArgs := append([]string{"-p", profile}, cmd...)

	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	sbxCmd := exec.CommandContext(ctx, "sandbox-exec", execArgs...)
	sbxCmd.Stdout = &limitedWriter{buf: &buf, limit: maxBytes}
	sbxCmd.Stderr = sbxCmd.Stdout
	sbxCmd.Env = env

	runErr := sbxCmd.Run()

	output := buf.Bytes()
	if int64(len(output)) >= maxBytes {
		return output[:maxBytes], fmt.Errorf("%w: output exceeded %d bytes",
			ErrSandboxOutputTruncated, maxBytes)
	}

	return output, runErr
}

// Profile returns the SBPL profile that would be applied for opts.
func (s *SeatbeltBackend) Profile(opts SandboxOptions) (string, error) {
	return GenerateSBPL(opts)
}
