package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGatewayChildRejectsArgumentsAndUnverifiedFactory(t *testing.T) {
	for _, args := range [][]string{{"secret"}, {}} {
		cmd := newGatewayChildCommand(nil)
		cmd.SetArgs(args)
		cmd.SetIn(strings.NewReader("private fixture"))
		var output bytes.Buffer
		cmd.SetOut(&output)
		cmd.SetErr(&output)
		if err := cmd.Execute(); err == nil {
			t.Fatal("unverified child started")
		}
		if strings.Contains(output.String(), "private fixture") {
			t.Fatal("config leaked")
		}
		if !cmd.Hidden {
			t.Fatal("internal command exposed")
		}
	}
}
func TestGatewayChildFactoryFailureHasNoHandoff(t *testing.T) {
	called := false
	cmd := newGatewayChildCommand(func(json.RawMessage) (http.Handler, error) {
		called = true
		return nil, errors.New("sensitive fixture")
	})
	cmd.SetIn(strings.NewReader(`{"parent_pid":1,"parent_fingerprint":"fixture","lifetime":1000000000,"poll_interval":1000000,"payload":{}}` + "\n"))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(nil)
	err := cmd.Execute()
	if !called {
		t.Fatal("factory not reached")
	}
	if err == nil {
		t.Fatal("failed factory started child")
	}
	if strings.Contains(output.String(), "sensitive fixture") || strings.Contains(output.String(), "Address") {
		t.Fatalf("invalid handoff or secret: %s", output.String())
	}
}

func TestGatewayChildInvalidParentDoesNotPublishHandoff(t *testing.T) {
	for _, closer := range []bool{false, true} {
		cmd := newGatewayChildCommand(func(json.RawMessage) (http.Handler, error) {
			return http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected request") }), nil
		})
		reader := strings.NewReader(`{"parent_pid":1,"parent_fingerprint":"unmatched-fixture","lifetime":1000000000,"poll_interval":1000000,"payload":{}}` + "\n")
		if closer {
			cmd.SetIn(io.NopCloser(reader))
		} else {
			cmd.SetIn(reader)
		}
		cmd.SetArgs(nil)
		var output bytes.Buffer
		cmd.SetOut(&output)
		if err := cmd.Execute(); err == nil || output.Len() != 0 {
			t.Fatalf("published invalid parent: %s %v", output.String(), err)
		}
	}
	cmd := newGatewayChildCommand(func(json.RawMessage) (http.Handler, error) {
		t.Fatal("malformed input reached factory")
		return nil, nil
	})
	cmd.SetIn(strings.NewReader("bad"))
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err == nil {
		t.Fatal("invalid config accepted")
	}
}
