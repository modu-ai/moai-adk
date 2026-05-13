package sandbox

import (
	"context"
)

type BubblewrapBackend struct{}

func (b *BubblewrapBackend) Available() bool { return false }
func (b *BubblewrapBackend) Exec(ctx interface{}, opts SandboxOptions, cmd []string) ([]byte, error) { return nil, ErrSandboxBackendUnavailable }
func (b *BubblewrapBackend) Profile(opts SandboxOptions) (string, error) { return "", nil }
func (b *BubblewrapBackend) buildArgs(opts SandboxOptions, cmd []string) ([]string, error) { return nil, ErrSandboxBackendUnavailable }

type SeatbeltBackend struct{}

func (s *SeatbeltBackend) Available() bool { return false }
func (s *SeatbeltBackend) Exec(ctx interface{}, opts SandboxOptions, cmd []string) ([]byte, error) { return nil, ErrSandboxBackendUnavailable }
func (s *SeatbeltBackend) Profile(opts SandboxOptions) (string, error) { return "", nil }
func (s *SeatbeltBackend) generateSBPL(opts SandboxOptions) (string, error) { return "", ErrSandboxBackendUnavailable }

type DockerBackend struct{}

func (d *DockerBackend) Available() bool { return false }
func (d *DockerBackend) Exec(ctx interface{}, opts SandboxOptions, cmd []string) ([]byte, error) { return nil, ErrSandboxBackendUnavailable }
func (d *DockerBackend) Profile(opts SandboxOptions) (string, error) { return "", nil }

type Launcher struct{}

func NewLauncher() *Launcher { return &Launcher{} }
func (l *Launcher) Exec(ctx context.Context, sandbox Sandbox, cmd []string, opts SandboxOptions) ([]byte, error) { return nil, ErrSandboxBackendUnavailable }
func (l *Launcher) resolveBackend(declared Sandbox, goos string) Sandbox { return declared }

func generateSBPL(opts SandboxOptions) (string, error) { return "", ErrSandboxBackendUnavailable }
func generateBwrapArgs(opts SandboxOptions, cmd []string) ([]string, error) { return nil, ErrSandboxBackendUnavailable }
func generateDockerSnippet(opts SandboxOptions) (string, error) { return "", ErrSandboxBackendUnavailable }
func ScrubEnv(parent []string, passthrough []string) []string { return parent }

