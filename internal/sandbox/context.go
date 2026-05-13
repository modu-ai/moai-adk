package sandbox

import (
	"fmt"
)

type Sandbox string

const (
	SandboxNone      Sandbox = "none"
	SandboxBubblewrap Sandbox = "bubblewrap"
	SandboxSeatbelt  Sandbox = "seatbelt"
	SandboxDocker    Sandbox = "docker"
)

func (s Sandbox) String() string {
	return string(s)
}

type SandboxOptions struct {
	WritableScope     []string
	ReadOnlyScope     []string
	NetworkAllowlist  []string
	EnvPassthrough    []string
	MaxOutputBytes    int64
}

func (o *SandboxOptions) Validate() error {
	for _, path := range o.WritableScope {
		if containsNullByte(path) {
			return fmt.Errorf("%w: WritableScope contains null byte", ErrSandboxProfileInvalid)
		}
	}
	for _, path := range o.ReadOnlyScope {
		if containsNullByte(path) {
			return fmt.Errorf("%w: ReadOnlyScope contains null byte", ErrSandboxProfileInvalid)
		}
	}
	if o.MaxOutputBytes < 0 {
		return fmt.Errorf("%w: MaxOutputBytes cannot be negative", ErrSandboxProfileInvalid)
	}
	return nil
}

func containsNullByte(s string) bool {
	for _, c := range s {
		if c == 0 {
			return true
		}
	}
	return false
}

type SandboxBackend interface {
	Available() bool
	Exec(ctx interface{}, opts SandboxOptions, cmd []string) ([]byte, error)
	Profile(opts SandboxOptions) (string, error)
}
