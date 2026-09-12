package auth

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSendRejectsLogoutBeforeWriteAndForeignEndpoint(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	req, _ := http.NewRequest("POST", SubscriptionEndpoint, strings.NewReader(`{}`))
	ref, _ := s.Resolve()
	foreign, _ := http.NewRequest("POST", APIEndpoint, strings.NewReader(`{}`))
	if e = ref.Apply(foreign); !errors.Is(e, ErrWrongProvider) || foreign.Header.Get("Authorization") != "" {
		t.Fatal("foreign endpoint accepted")
	}
	if _, e = s.Logout(context.Background()); e != nil {
		t.Fatal(e)
	}
	calls := 0
	tr := &http.Transport{DialTLSContext: func(context.Context, string, string) (net.Conn, error) { calls++; return nil, errors.New("unexpected") }}
	if _, e = s.SendAuthorized(context.Background(), gen, req, tr, SendOptions{WriteTimeout: time.Second, PollInterval: 10 * time.Millisecond, MaxBodyBytes: 100}); !errors.Is(e, ErrCredentialAbsent) {
		t.Fatal(e)
	}
	if calls != 0 {
		t.Fatal("stale external send")
	}
}
func TestSendWriteBarrierBlocksLogoutButSSEDoesNot(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	client, server := net.Pipe()
	t.Cleanup(func() { client.Close(); server.Close() })
	dialed := make(chan struct{})
	tr := &http.Transport{DialTLSContext: func(ctx context.Context, _, _ string) (net.Conn, error) { close(dialed); return client, nil }}
	req, _ := http.NewRequest("POST", SubscriptionEndpoint, strings.NewReader(`{}`))
	type result struct {
		r *http.Response
		e error
	}
	done := make(chan result, 1)
	go func() {
		r, e := s.SendAuthorized(context.Background(), gen, req, tr, SendOptions{WriteTimeout: time.Second, PollInterval: 10 * time.Millisecond, MaxBodyBytes: 100})
		done <- result{r, e}
	}()
	<-dialed
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, e = s.Logout(ctx); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatalf("logout crossed blocked socket write: %v", e)
	}
	received := make(chan string, 1)
	go func() {
		r, e := http.ReadRequest(bufio.NewReader(server))
		if e != nil {
			received <- ""
			return
		}
		io.Copy(io.Discard, r.Body)
		r.Body.Close()
		received <- r.Header.Get("Authorization")
		io.WriteString(server, "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nTransfer-Encoding: chunked\r\n\r\n")
	}()
	if <-received == "" {
		t.Fatal("missing authentication")
	}
	resultGot := <-done
	if resultGot.e != nil {
		t.Fatal(resultGot.e)
	}
	defer resultGot.r.Body.Close()
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	if _, e = s.Logout(ctx2); e != nil {
		t.Fatalf("SSE held logout: %v", e)
	}
	read := make(chan error, 1)
	go func() { _, e := resultGot.r.Body.Read(make([]byte, 1)); read <- e }()
	select {
	case e := <-read:
		if e == nil {
			t.Fatal("in-flight stream not canceled")
		}
	case <-time.After(time.Second):
		t.Fatal("in-flight cancellation missing")
	}
}
func TestSendTimeoutCancelsBlockedSocketAndReleasesLock(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	client, server := net.Pipe()
	t.Cleanup(func() { client.Close(); server.Close() })
	tr := &http.Transport{DialTLSContext: func(context.Context, string, string) (net.Conn, error) { return client, nil }}
	req, _ := http.NewRequest("POST", SubscriptionEndpoint, strings.NewReader(`{}`))
	_, e = s.SendAuthorized(context.Background(), gen, req, tr, SendOptions{WriteTimeout: 30 * time.Millisecond, PollInterval: 10 * time.Millisecond, MaxBodyBytes: 100})
	if e == nil {
		t.Fatal("blocked write succeeded")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, e = s.Logout(ctx); e != nil {
		t.Fatal(e)
	}
}

func TestEarlyResponseCannotReleaseUnwrittenHeader(t *testing.T) {
	s, e := OpenStore(privateStorePath(t))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	gen := loginFixture(t, s)
	client, server := net.Pipe()
	t.Cleanup(func() { client.Close(); server.Close() })
	tr := &http.Transport{DialTLSContext: func(context.Context, string, string) (net.Conn, error) { return client, nil }}
	req, _ := http.NewRequest("POST", SubscriptionEndpoint, strings.NewReader(`{}`))
	go func() { io.WriteString(server, "HTTP/1.1 403 Forbidden\r\nContent-Length: 0\r\n\r\n") }()
	r, e := s.SendAuthorized(context.Background(), gen, req, tr, SendOptions{WriteTimeout: 40 * time.Millisecond, PollInterval: 10 * time.Millisecond, MaxBodyBytes: 100})
	if r != nil {
		r.Body.Close()
	}
	if e == nil {
		t.Fatal("response accepted before header write")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, e = s.Logout(ctx); e != nil {
		t.Fatal(e)
	}
}

func TestHeaderBoundaryUsesSuccessfulBytesAcrossSplitWrites(t *testing.T) {
	client, server := net.Pipe()
	t.Cleanup(func() { client.Close(); server.Close() })
	observed := 0
	c := &headerConn{Conn: client, done: func() { observed++ }}
	reads := make(chan struct{})
	go func() { io.Copy(io.Discard, server); close(reads) }()
	for _, part := range []string{"POST / HTTP/1.1\r\nAuthorization: Bearer marker\r", "\n\r", "\nbody"} {
		if _, e := c.Write([]byte(part)); e != nil {
			t.Fatal(e)
		}
	}
	if observed != 1 || !c.headerWritten.Load() || len(c.tail) != 0 {
		t.Fatal("header boundary or secret retention")
	}
	if _, e := c.Write([]byte("more")); e != nil {
		t.Fatal(e)
	}
	if observed != 1 {
		t.Fatal("duplicate barrier")
	}
	client.Close()
	server.Close()
	<-reads
}
