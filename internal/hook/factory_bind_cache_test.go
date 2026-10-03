package hook

// factory_bind_cache_test.go — SPEC-FACTORY-DECISION-AUTO-001 M7: the bind
// cache behind the run-state probe (REQ-FDA-020/021) and the degraded inbox
// surface (REQ-FDA-022).

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
)

// countOpens wraps the broker-open seam and counts calls; failNext makes every
// open fail, which would surface as a degraded notice if the path were taken.
func countOpens(t *testing.T, fail bool) *int {
	t.Helper()
	n := 0
	prev := factoryHookOpenStore
	factoryHookOpenStore = func(root, run string) (*factorymsg.Store, error) {
		n++
		if fail {
			return nil, errors.New("open past budget")
		}
		return prev(root, run)
	}
	t.Cleanup(func() { factoryHookOpenStore = prev })
	return &n
}

func countProbes(t *testing.T) *int {
	t.Helper()
	n := 0
	prev := factoryHookProbeRun
	factoryHookProbeRun = func(ctx context.Context, dbPath, run string) (factorymsg.RunState, string, error) {
		n++
		return prev(ctx, dbPath, run)
	}
	t.Cleanup(func() { factoryHookProbeRun = prev })
	return &n
}

// AC-FDA-020 — a matching cache on a live run still probes the run but opens
// no broker and queries no peer, and emits no degraded notice even when an
// open would fail.
func TestFDA_BindCacheHitSkipsTheBrokerButNotTheProbe(t *testing.T) {
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=active")
	first := rebindPrompt(t, root, "s1")
	if !strings.Contains(first, "factory messaging bound: run=runX slot=lane-3") {
		t.Fatalf("prompt 1 did not bind: %q", first)
	}
	probes := countProbes(t)
	opens := countOpens(t, true)
	second := rebindPrompt(t, root, "s1")
	if *opens != 0 {
		t.Fatalf("cache hit opened the broker %d time(s) for the bind", *opens)
	}
	if *probes < 1 {
		t.Fatalf("cache hit skipped the run-state probe")
	}
	if strings.Contains(second, "degraded") {
		t.Fatalf("already-bound session on a live run got a degraded notice: %q", second)
	}
}

// A mismatching cache (another session) takes the full bind.
func TestFDA_BindCacheMissTakesTheFullBind(t *testing.T) {
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=active")
	_ = rebindPrompt(t, root, "s1")
	opens := countOpens(t, false)
	_ = rebindPrompt(t, root, "s2")
	if *opens == 0 {
		t.Fatalf("a different session reused another session's bind cache")
	}
}

// AC-FDA-021 — a cached session whose run retires rebinds on the next prompt;
// a probe failure on a hit is surfaced, not suppressed.
func TestFDA_RetiredRunInvalidatesTheBindCache(t *testing.T) {
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=active")
	if first := rebindPrompt(t, root, "s1"); !strings.Contains(first, "factory messaging bound") {
		t.Fatalf("prompt 1 did not bind: %q", first)
	}
	recordFactoryRunWithStatus(t, root, "runX", "retired")
	recordActiveFactoryRun(t, root, "runY")
	next := rebindPrompt(t, root, "s1")
	if !strings.Contains(next, "factory lane rebound:") || !strings.Contains(next, "runY") {
		t.Fatalf("cached session on a retired run did not rebind: %q", next)
	}
	if got := laneRows(t, root, "runY"); got != "lane-3|s1|1" {
		t.Fatalf("runY peers = %q, want lane-3|s1|1", got)
	}
}

func TestFDA_ProbeFailureOnACacheHitIsSurfaced(t *testing.T) {
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=active")
	_ = rebindPrompt(t, root, "s1")
	prev := factoryHookProbeRun
	factoryHookProbeRun = func(context.Context, string, string) (factorymsg.RunState, string, error) {
		return factorymsg.RunStateUnavailable, "", errors.New("probe broke")
	}
	t.Cleanup(func() { factoryHookProbeRun = prev })
	if got := rebindPrompt(t, root, "s1"); !strings.Contains(got, "degraded") {
		t.Fatalf("probe failure on a cache hit was suppressed: %q", got)
	}
}

// AC-FDA-022 — a degraded inbox state logs at warn every time and surfaces a
// notice at most once per interval per session.
func TestFDA_DegradedInboxStateWarnsAndNotifiesRateLimited(t *testing.T) {
	rebindEnv(t, "runX", "lane-3", "claude")
	root := rebindRoot(t, "runX=active")
	_ = rebindPrompt(t, root, "s1")
	var logs bytes.Buffer
	prevLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(prevLogger) })
	prevOpen := factoryHookOpenInbox
	factoryHookOpenInbox = func(string, string, time.Duration) (*factorymsg.Store, error) {
		return nil, errors.New("database is locked")
	}
	t.Cleanup(func() { factoryHookOpenInbox = prevOpen })
	prevInterval := factoryDegradedNoticeInterval
	factoryDegradedNoticeInterval = time.Hour
	t.Cleanup(func() { factoryDegradedNoticeInterval = prevInterval })

	first := rebindPrompt(t, root, "s1")
	second := rebindPrompt(t, root, "s1")
	if !strings.Contains(first, "factory messaging degraded: inbox") {
		t.Fatalf("first degraded inbox claim was not surfaced: %q", first)
	}
	if strings.Contains(second, "factory messaging degraded: inbox") {
		t.Fatalf("degraded inbox notice repeated inside the interval: %q", second)
	}
	if n := strings.Count(logs.String(), "factory inbox degraded"); n < 2 {
		t.Fatalf("warn log lines = %d, want one per occurrence (>=2):\n%s", n, logs.String())
	}
}
