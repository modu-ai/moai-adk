package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type SendOptions struct {
	WriteTimeout time.Duration
	PollInterval time.Duration
	MaxBodyBytes int64
}

var ErrSend = errors.New("gateway authorized request failed")

// SendAuthorized serializes logout against observed plaintext HTTP header writes.
// The TLS connection is wrapped AFTER its handshake: successful net.Conn.Write
// byte counts, not httptrace callbacks (which precede buffered flushes), establish
// the header boundary. Redirects and HTTP/2 are deliberately not followed.
// @MX:WARN: [AUTO] A generation watcher cancels in-flight requests after logout.
// @MX:REASON: Other processes publish tombstones; body Close/EOF joins the watcher.
func (s *Store) SendAuthorized(ctx context.Context, expected uint64, request *http.Request, transport *http.Transport, options SendOptions) (*http.Response, error) {
	if options.WriteTimeout <= 0 || options.PollInterval <= 0 || options.MaxBodyBytes <= 0 || transport == nil || request == nil || request.Method != "POST" || request.ContentLength < 0 || request.ContentLength > options.MaxBodyBytes {
		return nil, ErrSend
	}
	var payload []byte
	if request.Body != nil {
		if request.GetBody == nil {
			return nil, ErrSend
		}
		body, e := request.GetBody()
		if e != nil {
			return nil, ErrSend
		}
		payload, e = io.ReadAll(io.LimitReader(body, options.MaxBodyBytes+1))
		_ = body.Close() // in-memory GetBody reader; no flush to lose
		if e != nil || int64(len(payload)) != request.ContentLength || int64(len(payload)) > options.MaxBodyBytes {
			return nil, ErrSend
		}
	}
	unlock, e := s.lock(ctx, "state.lock")
	if e != nil {
		return nil, e
	}
	var release sync.Once
	releaseLock := func() { release.Do(unlock) }
	defer releaseLock()
	v, e := s.read()
	if e != nil {
		return nil, e
	}
	if v.Tombstone {
		return nil, ErrCredentialAbsent
	}
	if v.Generation != expected {
		return nil, ErrCredentialChanged
	}
	a, expiry, e := parseAuth(v.Auth)
	if e != nil {
		return nil, e
	}
	sendCtx, cancel := context.WithCancel(ctx)
	req := request.Clone(sendCtx)
	req.Header = request.Header.Clone()
	req.Body = io.NopCloser(bytes.NewReader(payload))
	req.GetBody = nil
	if e = applySubscription(req, a, expiry); e != nil {
		cancel()
		return nil, e
	}
	tr := transport.Clone()
	tr.DisableKeepAlives = true
	tr.ForceAttemptHTTP2 = false
	tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	if tr.Proxy != nil {
		cancel()
		return nil, ErrSend
	}
	watcherDone := make(chan struct{})
	var watchOnce sync.Once
	startWatcher := func() {
		watchOnce.Do(func() {
			go func() {
				defer close(watcherDone)
				ticker := time.NewTicker(options.PollInterval)
				defer ticker.Stop()
				for {
					select {
					case <-sendCtx.Done():
						return
					case <-ticker.C:
						status, err := s.Status(sendCtx)
						if err != nil || !status.LoggedIn || status.Generation != expected {
							cancel()
							return
						}
					}
				}
			}()
		})
	}
	deadline := time.Now().Add(options.WriteTimeout)
	timer := time.AfterFunc(options.WriteTimeout, cancel)
	var connectionMu sync.Mutex
	var connection *headerConn
	originalDial := tr.DialTLSContext
	tr.DialTLSContext = func(dialCtx context.Context, network, address string) (net.Conn, error) {
		bounded, stop := context.WithDeadline(dialCtx, deadline)
		defer stop()
		var conn net.Conn
		var err error
		if originalDial != nil {
			conn, err = originalDial(bounded, network, address)
		} else {
			cfg := &tls.Config{MinVersion: tls.VersionTLS12}
			if tr.TLSClientConfig != nil {
				cfg = tr.TLSClientConfig.Clone()
			}
			cfg.NextProtos = []string{"http/1.1"}
			dialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: options.WriteTimeout}, Config: cfg}
			conn, err = dialer.DialContext(bounded, network, address)
		}
		if err != nil {
			return nil, err
		}
		if tc, ok := conn.(*tls.Conn); ok {
			protocol := tc.ConnectionState().NegotiatedProtocol
			if protocol != "" && protocol != "http/1.1" {
				_ = conn.Close() // protocol mismatch; the connection is being discarded
				return nil, ErrSend
			}
		}
		if sendCtx.Err() != nil {
			_ = conn.Close() // canceled dial; the connection is being discarded
			return nil, sendCtx.Err()
		}
		wrapped := &headerConn{Conn: conn, done: func() { timer.Stop(); releaseLock(); startWatcher() }}
		if err = conn.SetWriteDeadline(deadline); err != nil {
			_ = conn.Close() // setup failed; the connection is being discarded
			return nil, ErrSend
		}
		connectionMu.Lock()
		connection = wrapped
		connectionMu.Unlock()
		return wrapped, nil
	}
	response, e := tr.RoundTrip(req)
	timer.Stop()
	if e != nil {
		cancel()
		connectionMu.Lock()
		if connection != nil {
			_ = connection.Close() // transport error path; teardown discards the connection
			connection.writes.Lock()
			connection.tail = nil
			connection.writes.Unlock()
		}
		connectionMu.Unlock()
		releaseLock()
		startWatcher()
		<-watcherDone
		tr.CloseIdleConnections()
		return nil, ErrSend
	}
	// A response may arrive before the write loop ends. Its request headers must
	// still be observed before returning a stream or releasing the state lock.
	connectionMu.Lock()
	observed := connection != nil && connection.headerWritten.Load()
	connectionMu.Unlock()
	if !observed {
		_ = response.Body.Close() // unobserved response; the caller receives ErrSend regardless
		cancel()
		return nil, ErrSend
	}
	cleanup := func() { cancel(); startWatcher(); <-watcherDone; tr.CloseIdleConnections() }
	response.Body = &authorizedBody{ReadCloser: response.Body, cleanup: cleanup}
	return response, nil
}

type headerConn struct {
	net.Conn
	writes        sync.Mutex
	tail          []byte
	headerWritten atomic.Bool
	done          func()
}

func (c *headerConn) Write(p []byte) (int, error) {
	c.writes.Lock()
	defer c.writes.Unlock()
	n, e := c.Conn.Write(p) //nolint:staticcheck // QF1008: c.Write would recurse into headerConn.Write
	if n > 0 && !c.headerWritten.Load() {
		combined := append(c.tail, p[:n]...)
		if bytes.Contains(combined, []byte("\r\n\r\n")) {
			c.tail = nil
			c.headerWritten.Store(true)
			_ = c.SetWriteDeadline(time.Time{})
			c.done()
		} else {
			if len(combined) > 3 {
				combined = combined[len(combined)-3:]
			}
			c.tail = append([]byte(nil), combined...)
		}
	}
	return n, e
}

type authorizedBody struct {
	io.ReadCloser
	cleanup func()
	once    sync.Once
}

func (b *authorizedBody) Read(p []byte) (int, error) {
	n, e := b.ReadCloser.Read(p)
	if e != nil {
		b.once.Do(b.cleanup)
	}
	return n, e
}
func (b *authorizedBody) Close() error { e := b.ReadCloser.Close(); b.once.Do(b.cleanup); return e }
