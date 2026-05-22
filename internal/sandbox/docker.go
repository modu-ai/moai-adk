package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// DockerBackend implements SandboxBackend for CI environments using Docker.
// It wraps commands with `docker run --rm ...` using ephemeral containers.
// Only activated when CI=1 is set or explicitly declared sandbox: docker.
type DockerBackend struct {
	// defaultImage is the Docker image used when SandboxOptions.DockerImage is empty.
	defaultImage string
}

// NewDockerBackend returns a new DockerBackend with the default image.
func NewDockerBackend() *DockerBackend {
	return &DockerBackend{
		defaultImage: "alpine:latest",
	}
}

// Available reports whether docker is installed and the daemon is reachable.
func (d *DockerBackend) Available() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}

	// daemon ping: docker info (quick check)
	ctx, cancel := context.WithTimeout(context.Background(), dockerProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{.ID}}")
	cmd.Env = os.Environ()
	_, err := cmd.Output()
	return err == nil
}

// Exec runs cmd inside an ephemeral Docker container.
// Network policy is determined by opts.NetworkAllowlist:
//   - empty list → --network=none
//   - non-empty → --network=bridge (with manual firewall outside Docker scope)
func (d *DockerBackend) Exec(opts SandboxOptions, cmd []string) ([]byte, error) {
	if !d.Available() {
		return nil, ErrSandboxBackendUnavailable
	}
	if len(cmd) == 0 {
		return nil, fmt.Errorf("sandbox docker exec: empty command")
	}

	maxBytes := opts.MaxOutputBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxOutputBytes
	}

	image := opts.DockerImage
	if image == "" {
		image = d.defaultImage
	}

	// Assemble docker run args
	dockerArgs := []string{"run", "--rm"}

	// Network policy
	allHosts := append(DefaultNetworkAllowlist, opts.NetworkAllowlist...)
	if len(allHosts) == 0 {
		dockerArgs = append(dockerArgs, "--network=none")
	} else {
		dockerArgs = append(dockerArgs, "--network=bridge")
	}

	// writable scope — -v mounts
	for _, p := range opts.WritableScope {
		dockerArgs = append(dockerArgs, "-v", p+":"+p)
	}

	// Working directory (the first writable scope)
	if len(opts.WritableScope) > 0 {
		dockerArgs = append(dockerArgs, "-w", opts.WritableScope[0])
	}

	// Environment variable scrubbing
	env := ScrubEnv(os.Environ(), opts.EnvPassthrough)
	for _, e := range env {
		dockerArgs = append(dockerArgs, "-e", e)
	}

	dockerArgs = append(dockerArgs, image)
	dockerArgs = append(dockerArgs, cmd...)

	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	dockerCmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	dockerCmd.Stdout = &limitedWriter{buf: &buf, limit: maxBytes}
	dockerCmd.Stderr = dockerCmd.Stdout

	runErr := dockerCmd.Run()

	output := buf.Bytes()
	if int64(len(output)) >= maxBytes {
		return output[:maxBytes], fmt.Errorf("%w: output exceeded %d bytes",
			ErrSandboxOutputTruncated, maxBytes)
	}

	return output, runErr
}

// Profile returns a Dockerfile snippet showing the container configuration.
func (d *DockerBackend) Profile(opts SandboxOptions) (string, error) {
	return GenerateDockerSnippet(opts)
}
