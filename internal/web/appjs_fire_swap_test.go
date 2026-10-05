package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// SPEC-APPJS-FIRE-GUARD-001 amendment 2 (card t1108) — the post-swap
// indicator is measured only after a REAL hx-boost swap and its
// htmx:afterSettle event, and the swap proves itself through the four
// REQ-AFG-016 premise legs.
//
// Every fixture below is a one-shot copy of the committed probe written under
// t.TempDir() (the TestAppJsHandlersFireSelectorMiss pattern): the committed
// probe is never modified, and each copy's lifetime is the framework's. A
// copy is built from exact anchors; an anchor that is absent or ambiguous
// fails the test before anything runs, so a fixture that silently stopped
// mutating anything can never read as a result (REQ-AFG-008 premise shape).
//
// Throttling (AC-AFG-014) is Emulation.setCPUThrottlingRate on the probe's
// own tab: it ends when the probe closes that tab, so there is no background
// load and nothing to reap beyond what launchFireGuardChrome already hangs on
// t.Cleanup (REQ-AFG-006).

// fireProbeEdit is one exact-anchor replacement applied to a probe copy.
type fireProbeEdit struct {
	anchor, replacement string
}

// writeFireProbeFixture writes a copy of the committed probe with the given
// edits applied. Each anchor must occur exactly once in the committed source.
func writeFireProbeFixture(t *testing.T, name string, edits ...fireProbeEdit) string {
	t.Helper()
	src, err := os.ReadFile(fireGuardProbePath(t))
	if err != nil {
		t.Fatalf("read committed probe: %v", err)
	}
	out := string(src)
	for _, e := range edits {
		if n := strings.Count(out, e.anchor); n != 1 {
			t.Fatalf("fixture %s: anchor occurs %d times (want exactly 1) — the probe changed shape; update the fixture:\n%s", name, n, e.anchor)
		}
		out = strings.Replace(out, e.anchor, e.replacement, 1)
	}
	path := filepath.Join(t.TempDir(), name+".py")
	if err := os.WriteFile(path, []byte(out), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
	return path
}

// The anchors the fixtures edit, verbatim from the committed probe.
const (
	fireAnchorSettleWaitCall = `rep["p5_settle_wait"], rep["p5_settle_wait_detail"] = await wait_after_settle(cdp)`
	fireAnchorListenerOpen   = `  m.settled=new Promise(function(res){`
	fireAnchorListenerClose  = "  });\n  window.__fireSwap=m;"
	fireAnchorOldTriggerTag  = "  if(old){old.setAttribute('data-fire-old-trigger', token);}\n"
	fireAnchorPremiseRead    = `legs = await ev(cdp, SWAP_PREMISE_JS % json.dumps(token)) or {}`
	fireAnchorSwapSelector   = `"selector": '#settings-form a[href="/settings?tab=audit"]',`
)

// fireMutantPathPoll is AC-AFG-014 mutant M1: the afterSettle wait reverted to
// the pre-amendment shape, polling location.pathname for the swap target's
// path. That condition is already true before the click, so the "wait" does
// not wait — and the mutant then claims the wait observed the event, exactly
// the belief a wait that is not a wait induces. Whatever turns it red must be
// a real consequence of not waiting.
func fireMutantPathPoll(t *testing.T) string {
	return writeFireProbeFixture(t, "mutant_m1_path_poll", fireProbeEdit{
		anchor: fireAnchorSettleWaitCall,
		replacement: `await poll(cdp, "location.pathname", "/settings")
        rep["p5_settle_wait"], rep["p5_settle_wait_detail"] = "observed", "M1 mutant: location.pathname polled instead of htmx:afterSettle"`,
	})
}

// fireMutantNoWait is AC-AFG-014 mutant M2: the afterSettle wait removed. The
// run goes straight to phase 6 and, like M1, claims the event was observed.
func fireMutantNoWait(t *testing.T) string {
	return writeFireProbeFixture(t, "mutant_m2_no_wait", fireProbeEdit{
		anchor:      fireAnchorSettleWaitCall,
		replacement: `rep["p5_settle_wait"], rep["p5_settle_wait_detail"] = "observed", "M2 mutant: afterSettle wait removed"`,
	})
}

// fireFixtureLateListener is AC-AFG-015 (iii) and AC-AFG-014 (d): the
// htmx:afterSwap / htmx:afterSettle listeners are attached only AFTER the swap
// has settled. The copy first confirms the swap by a DOM condition (the
// profile trigger is a new node without the old-trigger tag), then waits a
// fixed 2s far beyond the settle delay, and only then attaches — so both
// events are missed deterministically. "Attach right after the click" is
// deliberately NOT the form: the swap happens in the asynchronous response
// handling, so a listener attached right after the click can still see it.
func fireFixtureLateListener(t *testing.T) string {
	return writeFireProbeFixture(t, "fixture_late_listener",
		fireProbeEdit{
			anchor:      fireAnchorListenerOpen,
			replacement: `  m.attach=function(){m.settled=new Promise(function(res){`,
		},
		fireProbeEdit{
			anchor:      fireAnchorListenerClose,
			replacement: "  });};\n  window.__fireSwap=m;",
		},
		fireProbeEdit{
			anchor: fireAnchorSettleWaitCall,
			replacement: `rep["mutant_swap_confirmed_by_dom"] = await poll(cdp, """(function(){var t=document.querySelector('[data-pop="profile"]');return !!t && !t.hasAttribute('data-fire-old-trigger');})()""", True, timeout=30.0)
        await asyncio.sleep(2.0)
        await ev(cdp, "window.__fireSwap.attach()")
        rep["p5_settle_wait"], rep["p5_settle_wait_detail"] = await wait_after_settle(cdp)`,
		},
	)
}

// fireFixtureTagAfterSwap is AC-AFG-015 (iv): the old-trigger tag is put on
// the post-swap node instead of the pre-click node. Swap, events and document
// are untouched, so only leg (d) can go false.
func fireFixtureTagAfterSwap(t *testing.T) string {
	return writeFireProbeFixture(t, "fixture_tag_after_swap",
		fireProbeEdit{anchor: fireAnchorOldTriggerTag, replacement: ""},
		fireProbeEdit{
			anchor: fireAnchorPremiseRead,
			replacement: `await ev(cdp, """(function(){var t=document.querySelector('[data-pop="profile"]');if(t){t.setAttribute('data-fire-old-trigger', %s);}return !!t;})()""" % json.dumps(token))
    legs = await ev(cdp, SWAP_PREMISE_JS % json.dumps(token)) or {}`,
		},
	)
}

// fireFixtureFullNavigation is AC-AFG-015 (ii): the swap selector reverted to
// the pre-amendment a[href="/todo"], which has no boost ancestor — the click
// is a main-frame navigation.
func fireFixtureFullNavigation(t *testing.T) string {
	return writeFireProbeFixture(t, "fixture_full_navigation", fireProbeEdit{
		anchor:      fireAnchorSwapSelector,
		replacement: `"selector": 'a[href="/todo"]',`,
	})
}

// fireFailureReasons flattens a report's failures into "entry: reason" rows,
// sorted, so a red run's cause is one comparable string.
func fireFailureReasons(r fireProbeReport) []string {
	rows := make([]string, 0, len(r.Failures))
	for _, f := range r.Failures {
		entry := f.Entry
		if entry == "" {
			entry = "(window)"
		}
		rows = append(rows, entry+": "+f.Reason)
	}
	sort.Strings(rows)
	return rows
}

func fireHasFailure(r fireProbeReport, entry, reasonSubstr string) bool {
	for _, f := range r.Failures {
		if f.Entry == entry && strings.Contains(f.Reason, reasonSubstr) {
			return true
		}
	}
	return false
}

// fireVariantTally is one variant's outcome over N runs under one condition.
type fireVariantTally struct {
	name    string
	fired   int            // exit 0 + post-swap fired + premise proven
	red     int            // exit 1 + post-swap NOT judged fired
	other   []string       // runs that were neither, with why
	reasons map[string]int // red-reason distribution
	elapsed []float64      // click -> htmx:afterSettle, ms, when recorded
}

func fireMedian(xs []float64) float64 {
	if len(xs) == 0 {
		return -1
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	if len(s)%2 == 1 {
		return s[len(s)/2]
	}
	return (s[len(s)/2-1] + s[len(s)/2]) / 2
}

// runFireSettleVariant runs one probe variant n times against the real-root
// family under the given extra flags and tallies the AC-AFG-014 outcome.
func runFireSettleVariant(t *testing.T, name, script, cdpPort, baseURL string, n int, extra ...string) fireVariantTally {
	t.Helper()
	tally := fireVariantTally{name: name, reasons: map[string]int{}}
	args := append([]string{"--primary-entries-only"}, extra...)
	for i := 1; i <= n; i++ {
		label := fmt.Sprintf("t1108-%s-%02d", name, i)
		code, r := runFireGuardProbe(t, script, cdpPort, baseURL, label, args...)
		if code == 2 {
			t.Fatalf("%s: machine fault (exit 2) — not a product signal; report=%s", label, mustJSON(t, r))
		}
		if r.P5SettleElapsedMs != nil {
			tally.elapsed = append(tally.elapsed, *r.P5SettleElapsedMs)
		}
		switch {
		case code == 0 && r.P6PopoverAfterSwapFired && len(fireSwapPremiseProblems(r)) == 0:
			tally.fired++
		case code == 1 && !r.P6PopoverAfterSwapFired:
			tally.red++
			tally.reasons[strings.Join(fireFailureReasons(r), " | ")]++
		default:
			tally.other = append(tally.other, fmt.Sprintf("%s exit=%d post_swap_fired=%v premise=%v failures=%v",
				label, code, r.P6PopoverAfterSwapFired, fireSwapPremiseProblems(r), fireFailureReasons(r)))
		}
	}
	return tally
}

func logFireTally(t *testing.T, stage string, tl fireVariantTally, n int) {
	t.Helper()
	t.Logf("AC-AFG-014 %s | %-6s | fired %d/%d | red %d/%d | other %d | settle median %.1f ms (n=%d) %v",
		stage, tl.name, tl.fired, n, tl.red, n, len(tl.other), fireMedian(tl.elapsed), len(tl.elapsed), tl.elapsed)
	keys := make([]string, 0, len(tl.reasons))
	for k := range tl.reasons {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t.Logf("AC-AFG-014 %s | %-6s | red reason x%d: %s", stage, tl.name, tl.reasons[k], k)
	}
	for _, o := range tl.other {
		t.Logf("AC-AFG-014 %s | %-6s | neither fired nor red: %s", stage, tl.name, o)
	}
}

// Amendment 3 (plan.md §A000): the CI green step sets this switch so
// TestAppJsFirePostSwapSettleWait starts at stage 2. Only the exact value "1"
// selects that path; unset, empty, and anything else keep the B1 order.
const (
	fireSettleStage2Env    = "MOAI_BROWSER_GUARD_SETTLE_STAGE2"
	fireSettlePathLocalB1  = "local-b1"
	fireSettlePathCIStage2 = "ci-stage2"
)

// fireSettleStartPath maps the switch's raw value to the measurement path.
func fireSettleStartPath(raw string) string {
	if raw == "1" {
		return fireSettlePathCIStage2
	}
	return fireSettlePathLocalB1
}

// TestAppJsFirePostSwapSettleWait is AC-AFG-014: under a tab-scoped CPU
// throttle of 12x the post-swap indicator fires 10/10 on the committed probe
// with all four premise legs true every time, while M1 (path poll) and M2 (no
// wait) are each red 10/10 under the same condition.
//
// The measurement order is the lead's decision (B1) and is fixed:
//  1. throttle 12x alone; ONLY if M1 and M2 are each 10/10 red does this
//     stage decide the AC (the committed probe must then fire 10/10);
//  2. any other outcome applies the same settle-delay amplification to all
//     three variants, and the amplification must be SHOWN to take effect: the
//     committed probe's click -> htmx:afterSettle times are recorded 10x
//     without and 10x with it, and the amplified median must be larger.
//
// With MOAI_BROWSER_GUARD_SETTLE_STAGE2=1 (the CI path, amendment 3) the
// unamplified M1/M2 runs are skipped: the committed probe still runs 10x
// unamplified as the effect baseline, then stage 2 decides with the same
// conditions.
//
// The same test observes AC-AFG-014 (d): the late-listener copy misses
// htmx:afterSettle deterministically, and the probe must exit 1 naming
// popover_after_swap with "afterSettle wait expired" — expiry is never a pass.
func TestAppJsFirePostSwapSettleWait(t *testing.T) {
	chromePath := requireFireGuardPrereqs(t)
	baseURL := startFireGuardServer(t)
	cdpPort := launchFireGuardChrome(t, chromePath)

	const (
		runs            = 10
		throttle        = "12"
		amplifiedSettle = "1000" // ms; htmx 2.0.4 default is 20
	)
	variants := []struct{ name, script string }{
		{"normal", fireGuardProbePath(t)},
		{"M1", fireMutantPathPoll(t)},
		{"M2", fireMutantNoWait(t)},
	}

	measure := func(stage string, only int, extra ...string) map[string]fireVariantTally {
		out := map[string]fireVariantTally{}
		for _, v := range variants[:only] {
			tl := runFireSettleVariant(t, strings.ToLower(stage)+"-"+v.name, v.script, cdpPort, baseURL, runs, extra...)
			tl.name = v.name
			logFireTally(t, stage, tl, runs)
			out[v.name] = tl
		}
		return out
	}

	raw, set := os.LookupEnv(fireSettleStage2Env)
	path := fireSettleStartPath(raw)
	t.Logf("AC-AFG-014 path=%s %s=%q (set=%v)", path, fireSettleStage2Env, raw, set)

	throttled := []string{"--cpu-throttle", throttle}
	var stage1 map[string]fireVariantTally
	if path == fireSettlePathCIStage2 {
		// Only the committed probe: the amplification-effect baseline.
		stage1 = measure("stage1[throttle=12x,amplification=none,baseline-only]", 1, throttled...)
	} else {
		stage1 = measure("stage1[throttle=12x,amplification=none]", len(variants), throttled...)
	}
	decided := stage1
	decidingStage := "stage1"
	if path == fireSettlePathCIStage2 || stage1["M1"].red != runs || stage1["M2"].red != runs {
		if path == fireSettlePathCIStage2 {
			t.Logf("AC-AFG-014: %s — stage 2 from the start: settle-delay amplification %s ms on all three variants", path, amplifiedSettle)
		} else {
			t.Logf("AC-AFG-014: stage 1 did not make M1 and M2 each %d/%d red (M1 %d, M2 %d) — stage 2: settle-delay amplification %s ms on all three variants",
				runs, runs, stage1["M1"].red, stage1["M2"].red, amplifiedSettle)
		}
		stage2 := measure("stage2[throttle=12x,amplification="+amplifiedSettle+"ms]", len(variants),
			append(throttled, "--settle-delay-ms", amplifiedSettle)...)
		without, with := stage1["normal"].elapsed, stage2["normal"].elapsed
		t.Logf("AC-AFG-014 amplification effect: click->afterSettle ms WITHOUT %v (median %.1f), WITH %v (median %.1f)",
			without, fireMedian(without), with, fireMedian(with))
		if len(without) != runs || len(with) != runs {
			t.Fatalf("amplification effect not observed: %d/%d timings without and %d/%d with — the effect observation is mandatory, and without it stage 2 cannot decide the AC",
				len(without), runs, len(with), runs)
		}
		if fireMedian(with) <= fireMedian(without) {
			t.Fatalf("amplification did not take effect (median with %.1f ms <= without %.1f ms) — stage 2 cannot decide the AC",
				fireMedian(with), fireMedian(without))
		}
		decided, decidingStage = stage2, "stage2"
	}

	if got := decided["normal"].fired; got != runs {
		t.Errorf("(a) %s: the committed probe fired the post-swap indicator with a proven swap %d/%d times (want %d/%d)", decidingStage, got, runs, runs, runs)
	}
	for _, m := range []string{"M1", "M2"} {
		if got := decided[m].red; got != runs {
			t.Errorf("(b) %s: mutant %s was red %d/%d times (want %d/%d) — a wait that does not wait must not pass", decidingStage, m, got, runs, runs, runs)
		}
	}

	// (d) the expired wait is red, and it is named.
	code, r := runFireGuardProbe(t, fireFixtureLateListener(t), cdpPort, baseURL, "t1108-late-listener-expiry",
		append([]string{"--primary-entries-only"}, throttled...)...)
	if code != 1 {
		t.Fatalf("(d) late-listener copy exited %d (want 1) — an expired afterSettle wait must be red; failures=%v", code, fireFailureReasons(r))
	}
	if r.MutantSwapConfirmedByDOM == nil || !*r.MutantSwapConfirmedByDOM {
		t.Fatalf("(d) the late-listener copy did not confirm the swap by its DOM condition before attaching (mutant_swap_confirmed_by_dom=%v) — the fixture is not the late-listener form", r.MutantSwapConfirmedByDOM)
	}
	if r.P6PopoverAfterSwapFired {
		t.Error("(d) the post-swap indicator was judged fired after an expired wait")
	}
	if !fireHasFailure(r, "popover_after_swap", "afterSettle wait expired") {
		t.Errorf("(d) the report does not name popover_after_swap with \"afterSettle wait expired\": %v", fireFailureReasons(r))
	}
}

// TestAppJsFireSwapPremise is AC-AFG-015 (i)-(iv): each premise leg goes red
// on a fixture that falsifies it, and the report names the leg.
func TestAppJsFireSwapPremise(t *testing.T) {
	chromePath := requireFireGuardPrereqs(t)
	baseURL := startFireGuardServer(t)
	cdpPort := launchFireGuardChrome(t, chromePath)

	run := func(label, script string) fireProbeReport {
		t.Helper()
		code, r := runFireGuardProbe(t, script, cdpPort, baseURL, label, "--primary-entries-only")
		if code == 2 {
			t.Fatalf("%s: machine fault (exit 2); report=%s", label, mustJSON(t, r))
		}
		t.Logf("AC-AFG-015 %s: exit=%d premise=%v false_legs=%v settle=%s post_swap_fired=%v failures=%v",
			label, r.Exit, r.P5SwapPremise, r.P5SwapPremiseFalseLegs, r.P5SettleWait, r.P6PopoverAfterSwapFired, fireFailureReasons(r))
		return r
	}
	wantFalseLegs := func(label string, r fireProbeReport, want ...string) {
		t.Helper()
		for _, leg := range fireSwapPremiseLegs {
			v, ok := r.P5SwapPremise[leg]
			if !ok {
				t.Errorf("%s: premise leg %s is absent from the report", label, leg)
				continue
			}
			wantFalse := false
			for _, w := range want {
				if w == leg {
					wantFalse = true
				}
			}
			if v == wantFalse {
				t.Errorf("%s: premise leg %s = %v (want %v)", label, leg, v, !wantFalse)
			}
		}
		if !slicesEqual(r.P5SwapPremiseFalseLegs, want) {
			t.Errorf("%s: p5_swap_premise_false_legs = %v (want %v)", label, r.P5SwapPremiseFalseLegs, want)
		}
		for _, leg := range want {
			if !fireHasFailure(r, "swap_boosted_tab", leg) {
				t.Errorf("%s: the report does not name the false leg %s on swap_boosted_tab: %v", label, leg, fireFailureReasons(r))
			}
		}
	}

	// (i) the committed probe: every leg true, exit 0.
	if r := run("t1108-premise-i-normal", fireGuardProbePath(t)); r.Exit != 0 {
		t.Errorf("(i) committed probe exited %d (want 0): %v", r.Exit, fireFailureReasons(r))
	} else {
		wantFalseLegs("(i)", r)
	}

	// (ii) full navigation: (a)(b)(c) false, (d) true, exit 1.
	r := run("t1108-premise-ii-full-navigation", fireFixtureFullNavigation(t))
	if r.Exit != 1 {
		t.Errorf("(ii) exited %d (want 1)", r.Exit)
	}
	wantFalseLegs("(ii)", r, "a_boost_ancestor", "b_same_document", "c_swap_events")
	if r.P6PopoverAfterSwapFired {
		t.Error("(ii) the post-swap indicator was judged fired after a full navigation")
	}

	// (iii) late listener: only (c) false, exit 1.
	r = run("t1108-premise-iii-late-listener", fireFixtureLateListener(t))
	if r.Exit != 1 {
		t.Errorf("(iii) exited %d (want 1)", r.Exit)
	}
	if r.MutantSwapConfirmedByDOM == nil || !*r.MutantSwapConfirmedByDOM {
		t.Errorf("(iii) the fixture did not confirm the swap by its DOM condition before attaching (mutant_swap_confirmed_by_dom=%v)", r.MutantSwapConfirmedByDOM)
	}
	wantFalseLegs("(iii)", r, "c_swap_events")

	// (iv) old-trigger tag on the post-swap node: only (d) false, exit 1.
	r = run("t1108-premise-iv-tag-after-swap", fireFixtureTagAfterSwap(t))
	if r.Exit != 1 {
		t.Errorf("(iv) exited %d (want 1)", r.Exit)
	}
	wantFalseLegs("(iv)", r, "d_swap_inserted_trigger")
}

// TestAppJsFireSwapPremiseLegs is AC-AFG-015 (v): the probe's premise rule,
// fed synthetic reports through --judge-report, goes red on each leg alone
// and names only that leg. Ungated — no browser, no server — so the rule's
// reverse direction stays alive on every `go test` of this package. An
// all-true control proves the synthetic report is not red for some other
// reason.
func TestAppJsFireSwapPremiseLegs(t *testing.T) {
	t.Parallel()
	probe := fireGuardProbePath(t)
	driven, _ := fireManifestFamilies(t)

	synthetic := func(falseLeg string) map[string]any {
		premise := map[string]bool{}
		for _, leg := range fireSwapPremiseLegs {
			premise[leg] = leg != falseLeg
		}
		return map[string]any{
			"driven_entries":                 driven,
			"p1_has_glm_btn":                 true,
			"p1_load_referenceerrors":        []string{},
			"p2_glm_handler_fired":           true,
			"p3_panel_hidden_before":         true,
			"p3_has_close_btn":               true,
			"p3_popover_open_fired":          true,
			"p3_popover_close_btn_fired":     true,
			"p3_popover_outside_close_fired": true,
			"p4_tab_count":                   3,
			"p4_settings_tabs_fired":         true,
			"p5_swap_clicked":                true,
			"p5_settle_wait":                 "observed",
			"p5_swap_premise":                premise,
			"p5_swap_referenceerrors":        []string{},
			"p6_panel_hidden_before":         true,
			"p6_panel_flip_observed":         true,
			"p7_has_copy_btn":                true,
			"p7_copy_handler_fired":          true,
			"p7_load_referenceerrors":        []string{},
		}
	}
	judgeSynthetic := func(name string, rep map[string]any) (int, fireProbeReport) {
		t.Helper()
		path := filepath.Join(t.TempDir(), name+".json")
		if err := os.WriteFile(path, []byte(mustJSON(t, rep)), 0o600); err != nil {
			t.Fatalf("write synthetic report: %v", err)
		}
		out, err := exec.Command("python3", probe, "--judge-report", path).Output()
		code := 0
		if err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("run --judge-report: %v", err)
			}
			code = exitErr.ExitCode()
		}
		var r fireProbeReport
		if jerr := json.Unmarshal(out, &r); jerr != nil {
			t.Fatalf("%s: parse judged report (exit %d): %v\n%s", name, code, jerr, out)
		}
		return code, r
	}

	if code, r := judgeSynthetic("control-all-true", synthetic("")); code != 0 {
		t.Fatalf("control: an all-true synthetic report exited %d (want 0) — the fixture is red for a reason other than the premise: %v", code, fireFailureReasons(r))
	}

	for _, leg := range fireSwapPremiseLegs {
		t.Run(leg, func(t *testing.T) {
			code, r := judgeSynthetic("only-"+leg, synthetic(leg))
			if code != 1 {
				t.Fatalf("a report with only %s false exited %d (want 1)", leg, code)
			}
			if !fireHasFailure(r, "swap_boosted_tab", leg) {
				t.Errorf("the failure does not name %s on swap_boosted_tab: %v", leg, fireFailureReasons(r))
			}
			if r.P6PopoverAfterSwapFired {
				t.Errorf("the post-swap indicator was judged fired with %s false", leg)
			}
			for _, f := range r.Failures {
				for _, other := range fireSwapPremiseLegs {
					if other != leg && strings.Contains(f.Reason, other) {
						t.Errorf("failure %q names %s, which is true in this report — only %s may be named", f.Reason, other, leg)
					}
				}
			}
		})
	}
}

// TestAppJsFireSettleStartPath pins the amendment 3 switch: only the exact
// value "1" selects the CI path (start at stage 2); unset, empty, and every
// other value keep the local B1 order. Ungated — no browser.
func TestAppJsFireSettleStartPath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw  string
		want string
	}{
		{"", fireSettlePathLocalB1},
		{"1", fireSettlePathCIStage2},
		{" 1", fireSettlePathLocalB1},
		{"1 ", fireSettlePathLocalB1},
		{"true", fireSettlePathLocalB1},
		{"0", fireSettlePathLocalB1},
		{"01", fireSettlePathLocalB1},
	}
	for _, c := range cases {
		if got := fireSettleStartPath(c.raw); got != c.want {
			t.Errorf("fireSettleStartPath(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}
