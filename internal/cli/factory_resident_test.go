package cli

// factory_resident_test.go — `moai factory resident` (Issue #1948).
//
// Defect answer: these tests catch a loop that starts without an explicit tick period, a webhook
// channel that opens without a signing-secret source, an unsigned or badly signed delivery that is
// accepted, a claim store the command creates, a server that outlives its cancellation, and an
// exec runner built without the bounds or the git transport limit. Everything is a fake: a temp
// project, a fake gh runner, 127.0.0.1:0. No network, no real gh or git.

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge/webhook"
)

// residentProject is a temp project whose settings name an existing claim store.
func residentProject(t *testing.T) (root, store string) {
	t.Helper()
	root = t.TempDir()
	store = filepath.Join(root, "store")
	if err := os.MkdirAll(store, 0o755); err != nil {
		t.Fatal(err)
	}
	afwWrite(t, root, afwDoc(store))
	return root, store
}

func residentGH(calls *int) func(context.Context, ...string) ([]byte, error) {
	return func(context.Context, ...string) ([]byte, error) {
		*calls++
		return []byte("[]"), nil
	}
}

func fixedNow() time.Time { return time.Date(2026, 10, 30, 4, 0, 0, 0, time.UTC) }

func TestResident_NotConfiguredIsANoticeNotAnError(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	err := runFactoryResident(context.Background(), residentOptions{ProjectRoot: root, Once: true, Out: &out})
	if err != nil {
		t.Fatalf("an unconfigured project: %v, want nil", err)
	}
	if !strings.Contains(out.String(), "not configured") {
		t.Errorf("no not-configured notice: %q", out.String())
	}
}

func TestResident_BrokenSettingsIsAnError(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agentic-factory.yaml"), []byte("workflow: [not a map\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runFactoryResident(context.Background(), residentOptions{ProjectRoot: root, Once: true}); err == nil {
		t.Fatal("broken settings were accepted")
	}
}

func TestResident_TickHasNoDefault(t *testing.T) {
	root, _ := residentProject(t)
	calls := 0
	err := runFactoryResident(context.Background(), residentOptions{
		ProjectRoot: root, GHRead: residentGH(&calls), Now: fixedNow,
		Listen: "127.0.0.1:0", Secret: func(string) []byte { return []byte("s") },
	})
	if err == nil || !strings.Contains(err.Error(), "--tick") || !strings.Contains(err.Error(), "OD-17") {
		t.Fatalf("no tick: %v, want a refusal naming --tick and OD-17", err)
	}
	// positive control: the same options with --once run.
	if err := runFactoryResident(context.Background(), residentOptions{
		ProjectRoot: root, GHRead: residentGH(&calls), Now: fixedNow, Once: true,
	}); err != nil {
		t.Fatalf("control: --once without a tick: %v", err)
	}
}

func TestResident_WebhookChannelNeedsAddressAndSecretSource(t *testing.T) {
	root, _ := residentProject(t)
	calls := 0
	base := residentOptions{ProjectRoot: root, GHRead: residentGH(&calls), Now: fixedNow, Tick: time.Hour}
	if err := runFactoryResident(context.Background(), base); err == nil || !strings.Contains(err.Error(), "--listen") {
		t.Errorf("routes without --listen: %v", err)
	}
	noSecret := base
	noSecret.Listen = "127.0.0.1:0"
	if err := runFactoryResident(context.Background(), noSecret); err == nil || !strings.Contains(err.Error(), "secret") {
		t.Errorf("routes without a secret source: %v, want a refusal naming the secret", err)
	}
}

func TestResident_ListenWithoutRoutesIsRefused(t *testing.T) {
	root := t.TempDir()
	store := filepath.Join(root, "store")
	if err := os.MkdirAll(store, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := afwDoc(store)
	af := doc["workflow"].(map[string]any)["agentic_factory"].(map[string]any)
	af["webhook"].(map[string]any)["routes"] = []any{}
	afwWrite(t, root, doc)
	calls := 0
	err := runFactoryResident(context.Background(), residentOptions{
		ProjectRoot: root, GHRead: residentGH(&calls), Now: fixedNow, Tick: time.Hour, Listen: "127.0.0.1:0",
	})
	if err == nil || !strings.Contains(err.Error(), "no webhook route") {
		t.Fatalf("--listen with no routes: %v", err)
	}
}

func TestResident_MissingClaimStoreIsNotCreated(t *testing.T) {
	root := t.TempDir()
	store := filepath.Join(root, "absent-store")
	afwWrite(t, root, afwDoc(store))
	calls := 0
	err := runFactoryResident(context.Background(), residentOptions{
		ProjectRoot: root, GHRead: residentGH(&calls), Now: fixedNow, Once: true,
	})
	if err == nil || !strings.Contains(err.Error(), "does not create") {
		t.Fatalf("missing store: %v", err)
	}
	if _, statErr := os.Stat(store); statErr == nil {
		t.Error("the command created the claim store it was asked to read")
	}
}

func TestResident_OnceRunsOnePassAndReportsIt(t *testing.T) {
	root, _ := residentProject(t)
	calls := 0
	var out bytes.Buffer
	err := runFactoryResident(context.Background(), residentOptions{
		ProjectRoot: root, GHRead: residentGH(&calls), Now: fixedNow, Once: true, Out: &out,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(out.String(), "pass 2026-10-30T04:00:00Z"); n != 1 {
		t.Errorf("passes reported = %d, want 1:\n%s", n, out.String())
	}
	if !strings.Contains(out.String(), "writes: none wired") {
		t.Errorf("the unwired write paths are not stated at start:\n%s", out.String())
	}
}

// The acceptance scenario: one validly signed and one badly signed delivery over a real loopback
// listener, then cancellation, and the server is gone within the grace period.
func TestResident_WebhookAcceptsSignedAndRefusesBadlySigned(t *testing.T) {
	root, _ := residentProject(t)
	secret := []byte("test-secret-not-real")
	calls := 0
	ready := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	var out bytes.Buffer
	go func() {
		done <- runFactoryResident(ctx, residentOptions{
			ProjectRoot: root, GHRead: residentGH(&calls), Now: fixedNow, Tick: time.Hour,
			Listen: "127.0.0.1:0", Secret: func(string) []byte { return secret }, Out: &out,
			Ready: func(a string) { ready <- a },
		})
	}()
	var addr string
	select {
	case addr = <-ready:
	case err := <-done:
		t.Fatalf("returned before listening: %v", err)
	case <-time.After(20 * time.Second):
		t.Fatal("the server never reported its address")
	}

	body := []byte(`{"action":"opened","issue":{"number":7},"repository":{"full_name":"Owner/Name"}}`)
	post := func(sig string) int {
		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/hook/Owner/Name", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Hub-Signature-256", sig)
		req.Header.Set("X-GitHub-Delivery", "d-"+sig[len(sig)-6:])
		req.Header.Set("X-GitHub-Event", "issues")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return resp.StatusCode
	}
	if got := post(webhook.Sign(secret, body)); got/100 != 2 {
		t.Errorf("a validly signed delivery: status %d, want 2xx", got)
	}
	if got := post(webhook.Sign([]byte("another-secret"), body)); got != http.StatusUnauthorized {
		t.Errorf("a badly signed delivery: status %d, want 401", got)
	}

	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("shutdown returned %v, want nil", err)
		}
	case <-time.After(residentShutdownGrace + 3*time.Second):
		t.Fatal("the command did not stop after cancellation")
	}
	if d := time.Since(start); d > residentShutdownGrace {
		t.Errorf("shutdown took %s, want within %s", d, residentShutdownGrace)
	}
	// Control: the listener is really closed, not merely reported so.
	if _, err := http.Get("http://" + addr + "/hook/Owner/Name"); err == nil {
		t.Error("the listener still answers after shutdown")
	}
}

func TestResident_ExecRunnersCarryTheirBounds(t *testing.T) {
	e := newResidentExecs("/p")
	if e.GH.MaxOutput != 8<<20 || e.GH.WaitDelay != time.Second || e.GH.AllowWrite {
		t.Errorf("gh runner = %+v, want 8MiB, 1s, no writes", e.GH)
	}
	if e.Git.MaxOutput != 8<<20 || e.Git.WaitDelay != time.Second {
		t.Errorf("git runner = %+v, want 8MiB, 1s", e.Git)
	}
	found := false
	for _, kv := range e.Git.Env {
		if kv == "GIT_ALLOW_PROTOCOL=https:ssh" {
			found = true
		}
	}
	if !found {
		t.Errorf("git runner env %v lacks GIT_ALLOW_PROTOCOL=https:ssh", e.Git.Env)
	}
}

func TestResident_IsRegisteredBesideStatus(t *testing.T) {
	names := map[string]bool{}
	for _, c := range newFactoryCommand().Commands() {
		names[c.Name()] = true
	}
	if !names["resident"] || !names["status"] {
		t.Fatalf("factory subcommands = %v, want resident registered and status kept", names)
	}
	cmd := newFactoryResidentCommand()
	if f := cmd.Flags().Lookup("tick"); f == nil || f.DefValue != "0s" {
		t.Errorf("--tick must exist with no default, got %+v", f)
	}
}
