// Package codexapp owns a bounded, ordered Codex App Server stdio connection.
// It never reads authentication files or contacts a model endpoint directly.
package codexapp

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrClosed   = errors.New("codex app server closed")
	ErrRead     = errors.New("codex app server read failed")
	ErrEOF      = errors.New("codex app server EOF")
	ErrProtocol = errors.New("invalid Codex App Server protocol message")
	ErrLimit    = errors.New("codex app server message or event limit exceeded")
)

// RPCError deliberately excludes server-provided message/data, which can contain secrets.
type RPCError struct {
	Code int `json:"code"`
}

func (e *RPCError) Error() string { return fmt.Sprintf("Codex App Server RPC error %d", e.Code) }

type Message struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *RPCError       `json:"error,omitempty"`
}
type Config struct {
	Binary               string
	Home                 string
	MaxMessageBytes      int
	QueueSize            int
	ExpectedConfigSHA256 string
}
type reply struct {
	m   Message
	err error
}

// Client keeps process lifetime independent of individual RPC contexts. A canceled
// blocked write closes the transport because its delivery is indeterminate.
type Client struct {
	cmd      *exec.Cmd
	lease    *os.File
	in       io.WriteCloser
	max      int
	events   chan Message
	done     chan struct{}
	exited   chan struct{}
	write    chan struct{}
	mu       sync.Mutex
	next     uint64
	pending  map[string]chan reply
	requests map[string]bool
	err      error
	once     sync.Once
}

// Start launches only the official stdio interface. Home must be a caller-owned
// private profile; process cwd is Home to avoid loading the caller's project config.
// @MX:ANCHOR: [AUTO] Managed App Server process startup.
// @MX:REASON: Authentication, discovery and conversation clients share this boundary.
func Start(ctx context.Context, cfg Config) (*Client, error) {
	args := []string{"app-server", "--stdio"}
	for _, value := range []string{`cli_auth_credentials_store="file"`, `approval_policy="never"`, `sandbox_mode="read-only"`, `web_search="disabled"`, `features.shell_tool=false`, `features.codex_hooks=false`, `features.hooks=false`, `features.plugin_hooks=false`, `features.plugins=false`, `features.apps=false`, `features.view_image=false`, `features.multi_agent=false`, `features.multi_agent_v2=false`, `tools.update_plan.enabled=false`, `tools.experimental_request_user_input.enabled=false`} {
		args = append(args, "-c", value)
	}
	return startProcess(ctx, cfg, args)
}
func startProcess(ctx context.Context, cfg Config, args []string) (*Client, error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, errors.New("invalid App Server lifetime context")
	}
	if !filepath.IsAbs(cfg.Home) || !filepath.IsAbs(cfg.Binary) {
		return nil, errors.New("app server paths must be absolute")
	}
	info, err := os.Lstat(cfg.Home)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || !privateOwner(cfg.Home, info) {
		return nil, errors.New("app server home must be a private directory")
	}
	resolved, err := filepath.EvalSymlinks(cfg.Home)
	if err != nil || resolved != filepath.Clean(cfg.Home) {
		return nil, errors.New("app server home must not traverse symlinks")
	}
	configPath := filepath.Join(cfg.Home, "config.toml")
	if info, err := os.Lstat(configPath); err == nil {
		if cfg.ExpectedConfigSHA256 == "" || !info.Mode().IsRegular() || !privateOwner(configPath, info) || info.Size() > 1<<20 {
			return nil, errors.New("app server profile configuration is not authorized")
		}
		raw, err := os.ReadFile(configPath)
		if err != nil {
			return nil, errors.New("app server profile configuration unavailable")
		}
		sum := sha256.Sum256(raw)
		if hex.EncodeToString(sum[:]) != cfg.ExpectedConfigSHA256 {
			return nil, errors.New("app server profile configuration digest mismatch")
		}
	} else if !os.IsNotExist(err) || cfg.ExpectedConfigSHA256 != "" {
		return nil, errors.New("app server profile configuration unavailable")
	}
	if cfg.MaxMessageBytes == 0 {
		cfg.MaxMessageBytes = 8 << 20
	}
	if cfg.QueueSize == 0 {
		cfg.QueueSize = 128
	}
	if cfg.MaxMessageBytes < 128 || cfg.MaxMessageBytes > 64<<20 || cfg.QueueSize < 1 || cfg.QueueSize > 4096 {
		return nil, ErrLimit
	}
	lease, err := profileLease(filepath.Join(cfg.Home, ".moai-appserver.lock"))
	if err != nil {
		return nil, errors.New("app server profile is busy or cannot be locked")
	}
	started := false
	defer func() {
		if !started {
			// Construction failed; the flock is released by process exit too.
			_ = lease.Close()
		}
	}()
	cmd := exec.Command(cfg.Binary, args...)
	cmd.Dir = cfg.Home
	for _, key := range []string{"PATH", "LANG", "LC_ALL", "SYSTEMROOT", "SystemRoot", "WINDIR", "TEMP", "TMP", "TMPDIR"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "CODEX_HOME="+cfg.Home, "HOME="+cfg.Home, "USERPROFILE="+cfg.Home)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.New("app server stdin unavailable")
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		_ = in.Close() // process never started; discarding the pipe
		return nil, errors.New("app server stdout unavailable")
	}
	cmd.Stderr = io.Discard
	if err = cmd.Start(); err != nil {
		_ = in.Close() // start failed; discarding the pipes
		_ = out.Close()
		return nil, errors.New("app server start failed")
	}
	started = true
	c := &Client{lease: lease, cmd: cmd, in: in, max: cfg.MaxMessageBytes, events: make(chan Message, cfg.QueueSize), done: make(chan struct{}), exited: make(chan struct{}), write: make(chan struct{}, 1), pending: map[string]chan reply{}, requests: map[string]bool{}}
	go c.read(out)
	go func() {
		select {
		case <-ctx.Done():
			c.fail(ctx.Err())
		case <-c.done:
		}
	}()
	return c, nil
}
func idKey(raw json.RawMessage) (string, error) {
	var s string
	if json.Unmarshal(raw, &s) == nil && s != "" {
		b, _ := json.Marshal(s)
		return string(b), nil
	}
	var n json.Number
	if json.Unmarshal(raw, &n) == nil {
		if _, e := strconv.ParseInt(string(n), 10, 64); e == nil {
			return string(n), nil
		}
	}
	return "", ErrProtocol
}

// @MX:WARN: [AUTO] Single reader owns ordered event delivery and process reaping.
// @MX:REASON: Never block on full event queues; terminate before another response can deadlock.
func (c *Client) read(out io.ReadCloser) {
	defer close(c.events)
	defer close(c.exited)
	// The failure surface already fired from the read side (fail is once-only),
	// so the Wait status cannot change the reported error; the flock is also
	// released by process exit.
	defer func() { _ = c.cmd.Wait(); _ = c.lease.Close() }()
	scan := bufio.NewScanner(out)
	scan.Buffer(make([]byte, 0, 4096), c.max+1)
	for scan.Scan() {
		line := scan.Bytes()
		if len(line) > c.max {
			c.fail(ErrLimit)
			return
		}
		var m Message
		if json.Unmarshal(line, &m) != nil {
			c.fail(ErrProtocol)
			return
		}
		if m.Method != "" {
			if m.Result != nil || m.Error != nil {
				c.fail(ErrProtocol)
				return
			}
			if m.ID != nil {
				key, e := idKey(m.ID)
				if e != nil {
					c.fail(e)
					return
				}
				c.mu.Lock()
				duplicate := c.requests[key] || len(c.requests) >= cap(c.events)
				c.requests[key] = true
				c.mu.Unlock()
				if duplicate {
					c.fail(ErrProtocol)
					return
				}
			}
			select {
			case c.events <- m:
			case <-c.done:
				return
			default:
				c.fail(ErrLimit)
				return
			}
			continue
		}
		key, e := idKey(m.ID)
		if e != nil || (m.Result == nil) == (m.Error == nil) {
			c.fail(ErrProtocol)
			return
		}
		c.mu.Lock()
		ch := c.pending[key]
		delete(c.pending, key)
		c.mu.Unlock()
		if ch != nil {
			ch <- reply{m: m}
		}
	}
	if scan.Err() != nil {
		if errors.Is(scan.Err(), bufio.ErrTooLong) {
			c.fail(ErrLimit)
		} else {
			c.fail(ErrRead)
		}
	} else {
		c.fail(ErrEOF)
	}
}
func (c *Client) fail(err error) {
	c.once.Do(func() {
		c.mu.Lock()
		c.err = err
		for key, ch := range c.pending {
			ch <- reply{err: err}
			delete(c.pending, key)
		}
		c.mu.Unlock()
		close(c.done)
		_ = c.in.Close()         // shutdown path; pipe discards carry no payload
		_ = c.cmd.Process.Kill() // an already-exited process is the expected case
	})
}
func (c *Client) Err() error             { c.mu.Lock(); defer c.mu.Unlock(); return c.err }
func (c *Client) Events() <-chan Message { return c.events }
func (c *Client) Close() error {
	c.fail(ErrClosed)
	select {
	case <-c.exited:
		return nil
	case <-time.After(3 * time.Second):
		return errors.New("app server cleanup did not complete")
	}
}
func (c *Client) send(ctx context.Context, m Message) error {
	if ctx == nil {
		return errors.New("rpc context required")
	}
	b, e := json.Marshal(m)
	if e != nil {
		return ErrProtocol
	}
	if len(b) > c.max {
		return ErrLimit
	}
	select {
	case c.write <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.Err()
	}
	defer func() { <-c.write }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.Err()
	default:
	}
	result := make(chan error, 1)
	go func() { _, err := c.in.Write(append(b, '\n')); result <- err }()
	select {
	case err := <-result:
		if err != nil {
			c.fail(ErrClosed)
			return ErrClosed
		}
		return nil
	case <-ctx.Done():
		c.fail(ctx.Err())
		return ctx.Err()
	case <-c.done:
		return c.Err()
	}
}

// Call routes responses by ID, independently of interleaved notifications and server requests.
// @MX:ANCHOR: [AUTO] Correlated App Server RPC boundary.
// @MX:REASON: Managed login, model discovery and turn operations share ID routing.
func (c *Client) Call(ctx context.Context, method string, params any, result any) error {
	if strings.TrimSpace(method) == "" {
		return ErrProtocol
	}
	p, e := json.Marshal(params)
	if e != nil {
		return ErrProtocol
	}
	ch := make(chan reply, 1)
	c.mu.Lock()
	if c.err != nil {
		e = c.err
		c.mu.Unlock()
		return e
	}
	if len(c.pending) >= cap(c.events) {
		c.mu.Unlock()
		return ErrLimit
	}
	c.next++
	key := strconv.FormatUint(c.next, 10)
	c.pending[key] = ch
	c.mu.Unlock()
	defer func() { c.mu.Lock(); delete(c.pending, key); c.mu.Unlock() }()
	if e = c.send(ctx, Message{ID: json.RawMessage(key), Method: method, Params: p}); e != nil {
		return e
	}
	select {
	case r := <-ch:
		if r.err != nil {
			return r.err
		}
		if r.m.Error != nil {
			return r.m.Error
		}
		if result != nil && json.Unmarshal(r.m.Result, result) != nil {
			return ErrProtocol
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.Err()
	}
}
func (c *Client) Notify(ctx context.Context, method string, params any) error {
	p, e := json.Marshal(params)
	if e != nil || method == "" {
		return ErrProtocol
	}
	return c.send(ctx, Message{Method: method, Params: p})
}

// Respond consumes a server request ID once. An uncertain delivery cannot be retried.
func (c *Client) Respond(ctx context.Context, id json.RawMessage, result any) error {
	key, e := idKey(id)
	if e != nil {
		return e
	}
	b, e := json.Marshal(result)
	if e != nil {
		return ErrProtocol
	}
	c.mu.Lock()
	exists := c.requests[key]
	if exists {
		delete(c.requests, key)
	}
	c.mu.Unlock()
	if !exists {
		return ErrProtocol
	}
	return c.send(ctx, Message{ID: id, Result: b})
}

// DiscardRequest retires an already-delivered server request without sending a
// result. Only the trusted owner of a canceled/failed turn may call it. It does
// not cancel any client RPC or remove another server request's registration.
func (c *Client) DiscardRequest(id json.RawMessage) error {
	key, err := idKey(id)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.requests[key] {
		return ErrProtocol
	}
	delete(c.requests, key)
	return nil
}
