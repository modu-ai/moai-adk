package codexapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/net/websocket"
)

// SharedSocketPath is confined to the validated private authentication profile.
func SharedSocketPath(home string) string { return filepath.Join(home, "moai-appserver.sock") }

// ServeShared is the sole refresh-token owner. Its supervisor must outlive any
// individual gateway; cancellation shuts down the server and all connections.
func ServeShared(ctx context.Context, cfg Config) error {
	if err := validateConfig(ctx, cfg); err != nil {
		return err
	}
	lease, err := profileLease(filepath.Join(cfg.Home, ".moai-appserver.lock"))
	if err != nil {
		return ErrProfileBusy
	}
	// Release on every return without replacing the supervisor's outcome.
	defer func() { _ = lease.Close() }()
	socket := SharedSocketPath(cfg.Home)
	if info, err := os.Lstat(socket); err == nil {
		// A supervisor killed abruptly may leave the official server alive.
		// Never unlink a listening socket merely because its supervisor vanished.
		if info.Mode()&os.ModeSocket == 0 || !privateOwner(socket, info) {
			return errors.New("shared App Server socket is not private")
		}
		probe, dialErr := net.DialTimeout("unix", socket, 250*time.Millisecond)
		if dialErr == nil {
			_ = probe.Close() // Successful dial already establishes a live owner.
			return ErrProfileBusy
		}
		if !errors.Is(dialErr, syscall.ECONNREFUSED) && !errors.Is(dialErr, os.ErrNotExist) {
			return errors.New("shared App Server socket liveness unavailable")
		}
		if err := os.Remove(socket); err != nil {
			return errors.New("stale App Server socket unavailable")
		}
	} else if !os.IsNotExist(err) {
		return errors.New("shared App Server socket unavailable")
	}
	cmd := exec.CommandContext(ctx, cfg.Binary, appServerArgs("unix://"+socket)...)
	configureProcess(cmd, cfg.Home)
	inheritSharedLease(cmd, lease)
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	err = cmd.Run()
	// Cleanup is restricted to this supervisor's private socket while its
	// lifetime lease is still held. Never remove another process's profile state.
	if info, statErr := os.Lstat(socket); statErr == nil && info.Mode()&os.ModeSocket != 0 && privateOwner(socket, info) {
		_ = os.Remove(socket)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return errors.New("shared App Server exited")
	}
	return nil
}

// ConnectShared uses Codex's official WebSocket-over-Unix transport. Request IDs
// and event streams stay connection-local while authentication has one owner.
// @MX:WARN: [AUTO] Reader and cancellation watcher share connection lifetime.
// @MX:REASON: Closing a client must stop its goroutines without stopping peer sessions.
func ConnectShared(ctx context.Context, cfg Config) (*Client, error) {
	if err := validateConfig(ctx, cfg); err != nil {
		return nil, err
	}
	socket := SharedSocketPath(cfg.Home)
	info, err := os.Lstat(socket)
	if err != nil || info.Mode()&os.ModeSocket == 0 || !privateOwner(socket, info) {
		return nil, errors.New("shared App Server socket unavailable or not private")
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
	raw, err := (&net.Dialer{}).DialContext(ctx, "unix", socket)
	if err != nil {
		return nil, errors.New("shared App Server connection unavailable")
	}
	deadline := time.Now().Add(5 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = raw.SetDeadline(deadline)
	wsConfig, err := websocket.NewConfig("ws://localhost/", "http://localhost/")
	if err != nil {
		_ = raw.Close() // Preserve the configuration error.
		return nil, err
	}
	conn, err := websocket.NewClient(wsConfig, raw)
	if err != nil {
		_ = raw.Close() // Preserve the handshake error.
		return nil, errors.New("shared App Server handshake failed")
	}
	_ = raw.SetDeadline(time.Time{})
	conn.MaxPayloadBytes = cfg.MaxMessageBytes
	stream := &websocketLines{conn: conn}
	c := &Client{in: stream, max: cfg.MaxMessageBytes, events: make(chan Message, cfg.QueueSize), done: make(chan struct{}), exited: make(chan struct{}), write: make(chan struct{}, 1), pending: map[string]chan reply{}, requests: map[string]bool{}}
	go c.read(stream)
	go func() {
		select {
		case <-ctx.Done():
			c.fail(ctx.Err())
		case <-c.done:
		}
	}()
	return c, nil
}

// websocketLines adapts one JSON message per frame to the existing bounded
// JSON-line reader without buffering an unbounded conversation.
type websocketLines struct {
	conn   *websocket.Conn
	unread *bytes.Reader
}

func (s *websocketLines) Read(p []byte) (int, error) {
	if s.unread == nil || s.unread.Len() == 0 {
		var message []byte
		if err := websocket.Message.Receive(s.conn, &message); err != nil {
			return 0, err
		}
		var compact bytes.Buffer
		if json.Compact(&compact, message) != nil {
			return 0, ErrProtocol
		}
		s.unread = bytes.NewReader(append(compact.Bytes(), '\n'))
	}
	return s.unread.Read(p)
}
func (s *websocketLines) Write(p []byte) (int, error) {
	if err := websocket.Message.Send(s.conn, string(bytes.TrimSuffix(p, []byte{'\n'}))); err != nil {
		return 0, err
	}
	return len(p), nil
}
func (s *websocketLines) Close() error { return s.conn.Close() }
