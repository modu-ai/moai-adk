package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestSupervisorChildHelper(t *testing.T) {
	mode := os.Getenv("MOAI_TEST_SUPERVISOR_HELPER")
	if mode == "" {
		return
	}
	if mode == "no-read" {
		time.Sleep(3 * time.Second)
		os.Exit(36)
	}
	cfg, err := ReadChildConfig(os.Stdin)
	if err != nil {
		os.Exit(31)
	}
	switch mode {
	case "eof":
		os.Exit(32)
	case "bad":
		_, _ = io.WriteString(os.Stdout, "not-json\n")
		os.Exit(33)
	case "stall":
		time.Sleep(30 * time.Second)
		os.Exit(34)
	}
	err = RunChildWithControl(context.Background(), cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "local-child") }), os.Stdout, os.Stdin)
	if err != nil {
		os.Exit(35)
	}
	os.Exit(0)
}

func testChildOptions(t *testing.T, mode string) StartOptions {
	t.Helper()
	return StartOptions{Executable: os.Args[0], Args: []string{"-test.run=^TestSupervisorChildHelper$"}, Env: append(os.Environ(), "MOAI_TEST_SUPERVISOR_HELPER="+mode), StartupTimeout: 5 * time.Second, Config: ChildConfig{Lifetime: 10 * time.Second, PollInterval: 20 * time.Millisecond, Overlay: json.RawMessage(`{"modelPicker":[]}`)}}
}
func stopChild(t *testing.T, child *ChildProcess) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := child.Stop(ctx); err != nil {
		t.Errorf("stop child: %v", err)
	}
}
func assertClosed(t *testing.T, address string) {
	t.Helper()
	conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		t.Fatalf("listener still open: %s", address)
	}
}

func TestSupervisorChildStartsAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	child, err := StartChild(ctx, testChildOptions(t, "serve"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stopChild(t, child) })
	host, port, err := net.SplitHostPort(child.Address)
	if err != nil || host != "127.0.0.1" || port == "0" {
		t.Fatalf("bad handoff %s %v", child.Address, err)
	}
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get("http://" + child.Address + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if string(body) != "local-child" {
		t.Fatal(string(body))
	}
	if _, err := os.Stat(child.OverlayPath); err != nil {
		t.Fatal(err)
	}
	// Cancellation uses the private control pipe and waits for graceful cleanup.
	cancel()
	stopChild(t, child)
	assertClosed(t, child.Address)
	if _, err := os.Stat(child.OverlayPath); !os.IsNotExist(err) {
		t.Fatalf("overlay remains after cancel: %v", err)
	}
}

func TestSupervisorStartupFailureAndDeadline(t *testing.T) {
	for _, mode := range []string{"eof", "bad", "stall"} {
		t.Run(mode, func(t *testing.T) {
			opts := testChildOptions(t, mode)
			opts.StartupTimeout = 200 * time.Millisecond
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			child, err := StartChild(ctx, opts)
			if err == nil || child != nil {
				if child != nil {
					stopChild(t, child)
				}
				t.Fatalf("startup unexpectedly succeeded: %v", err)
			}
		})
	}
}

func TestSupervisorRejectsInvalidConfigWithoutSecrets(t *testing.T) {
	for _, raw := range []string{`{`, strings.Repeat("x", MaxChildConfigBytes+1), `{"unknown":"secret"}`} {
		if _, err := ReadChildConfig(strings.NewReader(raw)); err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatalf("invalid config error: %v", err)
		}
	}
	opts := testChildOptions(t, "serve")
	opts.Executable = "/missing/binary"
	if _, err := StartChild(context.Background(), opts); err == nil {
		t.Fatal("missing binary accepted")
	}
	opts = testChildOptions(t, "serve")
	opts.Config.Payload = json.RawMessage(strings.Repeat("x", MaxChildConfigBytes+1))
	if _, err := StartChild(context.Background(), opts); err == nil {
		t.Fatal("huge config accepted")
	}
}

func TestSupervisorRunnerLifetimeClosesPortAndOwnedOverlay(t *testing.T) {
	cfg := ChildConfig{ParentPID: os.Getpid(), ParentFingerprint: homestate.CurrentProcessFingerprint(), Lifetime: 180 * time.Millisecond, PollInterval: 20 * time.Millisecond, Overlay: json.RawMessage(`{"model":"fixture"}`)}
	if cfg.ParentFingerprint == "" {
		t.Fatal("parent fingerprint unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	done := make(chan error, 1)
	var calls atomic.Int64
	go func() {
		defer writer.Close()
		done <- RunChild(ctx, cfg, http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }), writer)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("runner not reaped")
		}
	})
	var handoff ChildHandoff
	if err := json.NewDecoder(reader).Decode(&handoff); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(handoff.OverlayPath); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
		done <- nil
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	assertClosed(t, handoff.Address)
	if _, err := os.Stat(handoff.OverlayPath); !os.IsNotExist(err) {
		t.Fatalf("owned overlay remains: %v", err)
	}
	if calls.Load() != 0 {
		t.Fatal("shutdown sent handler requests")
	}
}

func TestSupervisorRunnerParentIdentityAndCancellation(t *testing.T) {
	for _, mode := range []string{"changed fingerprint", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			cfg := ChildConfig{ParentPID: os.Getpid(), ParentFingerprint: homestate.CurrentProcessFingerprint(), Lifetime: 2 * time.Second, PollInterval: 20 * time.Millisecond}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()
			var checks atomic.Int64
			probe := func(pid int) (string, homestate.ProcessIdentityState) {
				if checks.Add(1) > 1 && mode == "changed fingerprint" {
					return "different-birth", homestate.ProcessIdentityLive
				}
				return cfg.ParentFingerprint, homestate.ProcessIdentityLive
			}
			done := make(chan error, 1)
			go func() {
				defer writer.Close()
				done <- runChild(ctx, cfg, http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("unexpected request") }), writer, probe)
			}()
			var h ChildHandoff
			if err := json.NewDecoder(reader).Decode(&h); err != nil {
				cancel()
				t.Fatal(err)
			}
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("runner did not finish")
			}
			assertClosed(t, h.Address)
		})
	}
}

func TestSupervisorParentProcessDeath(t *testing.T) {
	// The supervised identity is a real short-lived process, not the test process.
	command := exec.Command(os.Args[0], "-test.run=^TestSupervisorChildHelper$")
	cfg := ChildConfig{Lifetime: 30 * time.Second, PollInterval: 20 * time.Millisecond}
	command.Env = append(os.Environ(), "MOAI_TEST_SUPERVISOR_HELPER=stall")
	raw, _ := json.Marshal(ChildConfig{ParentPID: os.Getpid(), ParentFingerprint: homestate.CurrentProcessFingerprint(), Lifetime: time.Second, PollInterval: time.Second})
	command.Stdin = bytes.NewReader(raw)
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = command.Process.Kill(); _ = command.Wait() })
	cfg.ParentPID = command.Process.Pid
	cfg.ParentFingerprint, _ = homestate.ProbeProcessIdentity(cfg.ParentPID)
	if cfg.ParentFingerprint == "" {
		t.Fatal("spawned fingerprint unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	done := make(chan error, 1)
	go func() {
		defer writer.Close()
		done <- RunChild(ctx, cfg, http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("shutdown request") }), writer)
	}()
	var h ChildHandoff
	if err := json.NewDecoder(reader).Decode(&h); err != nil {
		t.Fatal(err)
	}
	_ = command.Process.Kill()
	_ = command.Wait()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("parent death not observed")
	}
	assertClosed(t, h.Address)
}

func TestSupervisorOverlayRejectsSymlinkReplacement(t *testing.T) {
	owned, err := newOwnedOverlay([]byte(`{"model":"fixture"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer owned.cleanup()
	target := filepath.Join(t.TempDir(), "unrelated")
	if err := os.WriteFile(target, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(owned.path); err != nil {
		t.Fatal(err)
	}
	// Windows may deny symlink creation without developer privileges. A regular replacement
	// exercises the same identity-rejection contract without skipping the test.
	if err := os.Symlink(target, owned.path); err != nil {
		if err := os.WriteFile(owned.path, []byte("replacement"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := owned.cleanup(); err == nil {
		t.Fatal("replacement was treated as owned")
	}
	body, err := os.ReadFile(target)
	if err != nil || string(body) != "preserve" {
		t.Fatal("unrelated target changed")
	}
	_ = os.Remove(owned.path)
	_ = os.Remove(filepath.Dir(owned.path))
}

func TestSupervisorBlockedInputStartupDeadline(t *testing.T) {
	opts := testChildOptions(t, "no-read")
	opts.StartupTimeout = 100 * time.Millisecond
	opts.Config.Payload = json.RawMessage(`"` + strings.Repeat("x", 512<<10) + `"`)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if child, err := StartChild(ctx, opts); err == nil || child != nil {
		if child != nil {
			stopChild(t, child)
		}
		t.Fatal("blocked input succeeded")
	}
}

func TestSupervisorOverlayRejectsDirectoryReplacement(t *testing.T) {
	owned, err := newOwnedOverlay([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	originalDir := filepath.Dir(owned.path)
	moved := filepath.Join(t.TempDir(), "moved")
	if err := os.Rename(originalDir, moved); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = owned.cleanup()
		_ = os.Remove(filepath.Join(moved, "settings.json"))
		_ = os.Remove(moved)
		_ = os.Remove(originalDir)
	})
	if err := os.Mkdir(originalDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := owned.cleanup(); err == nil {
		t.Fatal("replacement directory removed as owned")
	}
	if _, err := os.Stat(originalDir); err != nil {
		t.Fatalf("unrelated replacement directory removed: %v", err)
	}
}

type rejectedHandoff struct{ handoff ChildHandoff }

func (w *rejectedHandoff) Write(b []byte) (int, error) {
	_ = json.Unmarshal(b, &w.handoff)
	return 0, io.ErrClosedPipe
}
func TestSupervisorRunnerFailsClosed(t *testing.T) {
	cfg := ChildConfig{ParentPID: os.Getpid(), ParentFingerprint: homestate.CurrentProcessFingerprint(), Lifetime: time.Second, PollInterval: 20 * time.Millisecond}
	if cfg.ParentFingerprint == "" {
		t.Fatal("fingerprint unavailable")
	}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("failure path served request") })
	if err := RunChild(context.Background(), ChildConfig{}, handler, io.Discard); err == nil {
		t.Fatal("empty config accepted")
	}
	if err := RunChild(context.Background(), cfg, nil, io.Discard); err == nil {
		t.Fatal("nil handler accepted")
	}
	invalid := cfg
	invalid.ParentFingerprint = "not-parent"
	if err := RunChild(context.Background(), invalid, handler, io.Discard); err == nil {
		t.Fatal("mismatched parent accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := RunChild(ctx, cfg, handler, io.Discard); err == nil {
		t.Fatal("cancelled startup accepted")
	}
	invalid = cfg
	invalid.Overlay = json.RawMessage(`{`)
	if err := RunChild(context.Background(), invalid, handler, io.Discard); err == nil {
		t.Fatal("invalid overlay accepted")
	}
	failed := &rejectedHandoff{}
	cfg.Overlay = json.RawMessage(`{}`)
	if err := RunChild(context.Background(), cfg, handler, failed); err == nil {
		t.Fatal("failed handoff accepted")
	}
	assertClosed(t, failed.handoff.Address)
	if _, err := os.Stat(failed.handoff.OverlayPath); !os.IsNotExist(err) {
		t.Fatalf("overlay survives failed handoff: %v", err)
	}
}

func TestSupervisorConfigAndWaitContracts(t *testing.T) {
	raw := `{"parent_pid":1,"parent_fingerprint":"fixture","lifetime":1000000,"poll_interval":1000000}`
	if _, err := ReadChildConfig(strings.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{raw + ` {}`, `{}`, strings.ReplaceAll(raw, `"lifetime":1000000`, `"lifetime":0`)} {
		if _, err := ReadChildConfig(strings.NewReader(input)); err == nil {
			t.Fatal("invalid config accepted")
		}
	}
	if _, err := StartChild(context.Background(), StartOptions{}); err == nil {
		t.Fatal("unspecified process accepted")
	}
	child := &ChildProcess{done: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if child.Wait(ctx) != context.Canceled {
		t.Fatal("wait ignored cancellation")
	}
	close(child.done)
	if err := child.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSupervisorControlEOFDoesNotStopExecSession(t *testing.T) {
	cfg := ChildConfig{ParentPID: os.Getpid(), ParentFingerprint: homestate.CurrentProcessFingerprint(), Lifetime: 120 * time.Millisecond, PollInterval: 20 * time.Millisecond}
	var handoff bytes.Buffer
	started := time.Now()
	if err := RunChildWithControl(context.Background(), cfg, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), &handoff, io.NopCloser(strings.NewReader(""))); err != nil {
		t.Fatal(err)
	}
	if time.Since(started) < 100*time.Millisecond {
		t.Fatal("EOF stopped live exec session")
	}
	var h ChildHandoff
	if err := json.Unmarshal(handoff.Bytes(), &h); err != nil {
		t.Fatal(err)
	}
	assertClosed(t, h.Address)
}

func TestSupervisorConfigPreservesCoalescedStopByte(t *testing.T) {
	cfg := ChildConfig{ParentPID: os.Getpid(), ParentFingerprint: homestate.CurrentProcessFingerprint(), Lifetime: time.Second, PollInterval: 20 * time.Millisecond}
	raw, _ := json.Marshal(cfg)
	input := bytes.NewBuffer(append(append(raw, '\n'), byte(1)))
	parsed, err := ReadChildConfig(input)
	if err != nil {
		t.Fatal(err)
	}
	if input.Len() != 1 {
		t.Fatal("configuration reader consumed stop byte")
	}
	if err := RunChildWithControl(context.Background(), parsed, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), io.Discard, io.NopCloser(input)); err != nil && err != context.Canceled {
		t.Fatal(err)
	}
}

func TestSupervisorCancelledStopStillJoinsReaper(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=^TestSupervisorChildHelper$")
	command.Env = append(os.Environ(), "MOAI_TEST_SUPERVISOR_HELPER=no-read")
	input, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	child := &ChildProcess{command: command, input: input, done: make(chan struct{})}
	go func() { child.waitErr = command.Wait(); time.Sleep(100 * time.Millisecond); close(child.done) }()
	t.Cleanup(func() { _ = command.Process.Kill(); _ = input.Close(); <-child.done })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = child.Stop(ctx)
	select {
	case <-child.done:
	default:
		t.Fatal("Stop returned before process reaper completed")
	}
}

type failedAcceptListener struct{ net.Listener }

func (l failedAcceptListener) Accept() (net.Conn, error) { return nil, io.ErrClosedPipe }
func TestSupervisorBindAndServeFailuresNeverHandoffSuccess(t *testing.T) {
	cfg := ChildConfig{ParentPID: os.Getpid(), ParentFingerprint: homestate.CurrentProcessFingerprint(), Lifetime: time.Second, PollInterval: 20 * time.Millisecond, Overlay: json.RawMessage(`{}`)}
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected upstream attempt") })
	var output bytes.Buffer
	failedBind := func(network, address string) (net.Listener, error) {
		if network != "tcp4" || address != "127.0.0.1:0" {
			t.Fatalf("unsafe bind %s %s", network, address)
		}
		return nil, io.ErrClosedPipe
	}
	if err := runChildWithListener(context.Background(), cfg, handler, &output, homestate.ProbeProcessIdentity, failedBind); err == nil || output.Len() != 0 {
		t.Fatalf("failed bind handoff: %v %s", err, &output)
	}
	failedServe := func(network, address string) (net.Listener, error) {
		listener, err := net.Listen(network, address)
		if err != nil {
			return nil, err
		}
		return failedAcceptListener{listener}, nil
	}
	if err := runChildWithListener(context.Background(), cfg, handler, &output, homestate.ProbeProcessIdentity, failedServe); err == nil {
		t.Fatal("server failure hidden")
	}
	var h ChildHandoff
	if err := json.Unmarshal(output.Bytes(), &h); err != nil {
		t.Fatal(err)
	}
	assertClosed(t, h.Address)
	if _, err := os.Stat(h.OverlayPath); !os.IsNotExist(err) {
		t.Fatalf("serve failure leaked overlay: %v", err)
	}
}
