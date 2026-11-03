package web

// server_branches_test.go — small server.go paths that no test reached (Issue
// #1948, coverage of the hand-written code): the address the loopback indicator
// shows before and without a listener, and the codex probe the CLI layer injects.
// Tests only; no listener is opened and no browser is started.
//
// Defect answer: these tests catch an indicator that shows a blank or a
// non-loopback address, and an injected codex probe that is ignored.

import (
	"context"
	"testing"
)

func TestDisplayBindAddrWithoutAListener(t *testing.T) {
	for _, tc := range []struct {
		port int
		want string
	}{
		{0, loopbackHost},
		{8123, loopbackHost + ":8123"},
	} {
		s, err := NewServer(Config{ProjectRoot: t.TempDir(), Port: tc.port, NoOpen: true})
		if err != nil {
			t.Fatal(err)
		}
		if got := s.displayBindAddr(); got != tc.want {
			t.Errorf("port %d: displayBindAddr = %q, want %q", tc.port, got, tc.want)
		}
	}
}

func TestNewServerUsesTheInjectedCodexProbe(t *testing.T) {
	called := false
	s, err := NewServer(Config{
		ProjectRoot: t.TempDir(), NoOpen: true,
		CodexStateProbe: func(context.Context) CodexStateView {
			called = true
			return CodexStateView{Installed: true, Version: "9.9"}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := s.app.codexStateProbe(context.Background())
	if !called || !got.Installed || got.Version != "9.9" {
		t.Errorf("the injected probe was not used: called %v view %+v", called, got)
	}
}
