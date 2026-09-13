package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

const MaxChildConfigBytes = 1 << 20

var errChildConfig = errors.New("gateway child configuration invalid")

type ChildConfig struct {
	ParentPID         int           `json:"parent_pid"`
	ParentFingerprint string        `json:"parent_fingerprint"`
	Lifetime          time.Duration `json:"lifetime"`
	PollInterval      time.Duration `json:"poll_interval"`
	// Overlay must contain only nonsecret Claude settings. It is created by the child.
	Overlay json.RawMessage `json:"overlay,omitempty"`
	// Payload is private handler configuration, transported exclusively over stdin.
	Payload json.RawMessage `json:"payload,omitempty"`
}
type ChildHandoff struct {
	Address     string `json:"address"`
	OverlayPath string `json:"overlay_path,omitempty"`
}
type StartOptions struct {
	Executable     string
	Args           []string
	Env            []string
	StartupTimeout time.Duration
	Config         ChildConfig
}
type ChildProcess struct {
	ChildHandoff
	command  *exec.Cmd
	input    io.WriteCloser
	done     chan struct{}
	waitErr  error
	stopOnce sync.Once
}

func ReadChildConfig(r io.Reader) (ChildConfig, error) {
	// Read exactly the configuration line: a buffered reader discarded here could
	// consume a coalesced private stop byte belonging to RunChildWithControl.
	var raw []byte
	var one [1]byte
	for {
		n, err := io.ReadFull(r, one[:])
		if n == 1 {
			raw = append(raw, one[0])
		}
		if len(raw) > MaxChildConfigBytes || err != nil && err != io.EOF {
			return ChildConfig{}, errChildConfig
		}
		if err == io.EOF || one[0] == '\n' {
			break
		}
	}
	var cfg ChildConfig
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return ChildConfig{}, errChildConfig
	}
	if _, err := dec.Token(); err != io.EOF {
		return ChildConfig{}, errChildConfig
	}
	if err := validateChildConfig(cfg); err != nil {
		return ChildConfig{}, err
	}
	return cfg, nil
}
func validateChildConfig(c ChildConfig) error {
	if c.ParentPID <= 0 || c.ParentFingerprint == "" || c.Lifetime <= 0 || c.PollInterval <= 0 {
		return errChildConfig
	}
	return nil
}

// StartChild starts a separate process without replacing the launcher. The CLI
// can still syscall.Exec Claude after the positive bound-port handoff. Args must
// be fixed internal-command arguments, never credentials or handler configuration.
func StartChild(ctx context.Context, o StartOptions) (*ChildProcess, error) {
	if o.Executable == "" || o.StartupTimeout <= 0 {
		return nil, errChildConfig
	}
	o.Config.ParentPID = os.Getpid()
	o.Config.ParentFingerprint = homestate.CurrentProcessFingerprint()
	if err := validateChildConfig(o.Config); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(o.Config)
	if err != nil || len(raw)+1 > MaxChildConfigBytes {
		return nil, errChildConfig
	}
	command := exec.Command(o.Executable, o.Args...)
	command.Env = o.Env
	command.WaitDelay = time.Second
	configureSupervisorProcess(command)
	input, err := command.StdinPipe()
	if err != nil {
		return nil, errors.New("gateway child input pipe failed")
	}
	output, err := command.StdoutPipe()
	if err != nil {
		_ = input.Close()
		return nil, errors.New("gateway child handoff pipe failed")
	}
	// Child stderr is intentionally discarded: arbitrary error strings must not
	// expose private configuration or credentials to the terminal or a public log.
	if err := command.Start(); err != nil {
		_ = input.Close()
		_ = output.Close()
		return nil, errors.New("gateway child start failed")
	}
	child := &ChildProcess{command: command, input: input, done: make(chan struct{})}
	// @MX:WARN: [AUTO] Wait is the sole process reaper and synchronizes waitErr.
	// @MX:REASON: Every startup failure and Stop path joins this goroutine.
	go func() { child.waitErr = command.Wait(); close(child.done) }()
	result := make(chan error, 1)
	ioDone := make(chan struct{})
	// @MX:WARN: [AUTO] Pipe I/O is bounded by startup timeout and process termination.
	// @MX:REASON: Killing and waiting closes the child pipe on every failure path.
	go func() {
		defer close(ioDone)
		if _, err := input.Write(append(raw, '\n')); err != nil {
			result <- errors.New("gateway child input failed")
			return
		}
		var h ChildHandoff
		err := json.NewDecoder(io.LimitReader(output, 4096)).Decode(&h)
		if err == nil {
			host, port, e := net.SplitHostPort(h.Address)
			n, e2 := strconv.Atoi(port)
			if e != nil || e2 != nil || host != "127.0.0.1" || n <= 0 || n > 65535 {
				err = errors.New("invalid bound address")
			}
		}
		if err == nil {
			child.ChildHandoff = h
		}
		result <- err
	}()
	timer := time.NewTimer(o.StartupTimeout)
	defer timer.Stop()
	select {
	case err = <-result:
		if err == nil {
			select {
			case <-child.done:
				err = errors.New("gateway child exited before handoff")
			default:
			}
		}
	case <-timer.C:
		err = errors.New("gateway child startup timed out")
	case <-ctx.Done():
		err = errors.New("gateway child startup cancelled")
	}
	if err != nil {
		stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = child.Stop(stopCtx)
		_ = output.Close()
		<-ioDone
		return nil, errors.New("gateway child startup failed")
	}
	// @MX:WARN: [AUTO] A live launcher cancellation requests graceful child stop.
	// @MX:REASON: Parent exec discards this watcher; child PID/fingerprint monitoring remains authoritative.
	go func() {
		select {
		case <-ctx.Done():
			stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = child.Stop(stopCtx)
		case <-child.done:
		}
	}()
	return child, nil
}

// Stop requests private-pipe shutdown, then kills and reaps an unresponsive child.
// It is safe to call more than once. EOF alone is not a shutdown request, because
// POSIX exec closes the launcher's pipe while preserving the supervised PID.
func (c *ChildProcess) Stop(ctx context.Context) error {
	c.stopOnce.Do(func() {
		wrote := make(chan struct{})
		// @MX:WARN: [AUTO] A child that never reads stdin must not block Stop.
		// @MX:REASON: The kill deadline breaks pipe writes on every platform; both writer and process are joined.
		go func() { _, _ = c.input.Write([]byte{1}); _ = c.input.Close(); close(wrote) }()
		select {
		case <-c.done:
		case <-time.After(time.Second):
			_ = c.command.Process.Kill()
		}
		_ = c.input.Close()
		<-wrote
	})
	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		_ = c.command.Process.Kill()
		<-c.done
		return ctx.Err()
	}
}
func (c *ChildProcess) Wait(ctx context.Context) error {
	select {
	case <-c.done:
		return c.waitErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
