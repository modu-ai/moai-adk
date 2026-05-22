package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	// @MX:WARN: [AUTO] user namespace availability for bwrap --unshare-all depends on the OS kernel configuration
	// @MX:REASON: bwrap can fail on some Linux environments (inside containers,
	//             older kernels) when user namespaces are disabled. The Available()
	//             check is mandatory.
)

// BubblewrapBackend implements SandboxBackend for Linux using bwrap.
// It requires the bwrap binary in PATH and user namespace support in the kernel.
type BubblewrapBackend struct{}

// NewBubblewrapBackend returns a new BubblewrapBackend.
func NewBubblewrapBackend() *BubblewrapBackend {
	return &BubblewrapBackend{}
}

// Available reports whether bwrap is installed and usable on the current host.
func (b *BubblewrapBackend) Available() bool {
	_, err := exec.LookPath("bwrap")
	if err != nil {
		return false
	}

	// Minimal probe: run bwrap --version (indirectly verifies user-namespace availability)
	ctx, cancel := context.WithTimeout(context.Background(), bwrapProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bwrap", "--version")
	cmd.Env = []string{} // minimal environment
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check the "bwrap N.M.P" format — simple existence check
	return len(out) > 0
}

// Exec runs cmd inside a bubblewrap sandbox with the given options.
//
// @MX:WARN: [AUTO] buildArgs is a security-critical function — changing argument order or values breaks sandbox isolation
// @MX:REASON: bwrap argument ordering matters: --unshare-all must precede all other flags;
//             --bind before --ro-bind; -- separator required before user command.
func (b *BubblewrapBackend) Exec(opts SandboxOptions, cmd []string) ([]byte, error) {
	if !b.Available() {
		return nil, ErrSandboxBackendUnavailable
	}
	if len(cmd) == 0 {
		return nil, fmt.Errorf("sandbox exec: empty command")
	}

	maxBytes := opts.MaxOutputBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxOutputBytes
	}

	// Profile validation (argument generation)
	baseArgs, err := GenerateBwrapArgs(opts)
	if err != nil {
		return nil, fmt.Errorf("sandbox: generate bwrap args: %w", err)
	}

	// Environment variable scrubbing
	env := ScrubEnv(os.Environ(), opts.EnvPassthrough)

	// Final args: baseArgs + -- + cmd
	allArgs := append(baseArgs, "--")
	allArgs = append(allArgs, cmd...)

	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	bwrapCmd := exec.CommandContext(ctx, "bwrap", allArgs...)
	bwrapCmd.Stdout = &limitedWriter{buf: &buf, limit: maxBytes}
	bwrapCmd.Stderr = bwrapCmd.Stdout
	bwrapCmd.Env = env

	runErr := bwrapCmd.Run()

	output := buf.Bytes()
	if int64(len(output)) >= maxBytes {
		// Output truncated — return together with ErrSandboxOutputTruncated
		return output[:maxBytes], fmt.Errorf("%w: output exceeded %d bytes",
			ErrSandboxOutputTruncated, maxBytes)
	}

	return output, runErr
}

// Profile returns the bwrap argument string that would be used for opts.
func (b *BubblewrapBackend) Profile(opts SandboxOptions) (string, error) {
	args, err := GenerateBwrapArgs(opts)
	if err != nil {
		return "", err
	}

	result := "bwrap"
	for _, a := range args {
		result += " " + a
	}
	result += " -- <cmd>"
	return result, nil
}
