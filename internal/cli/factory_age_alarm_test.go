package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// SPEC-CARD-AGE-ALARM-001 (card t1838), milestone M3: the status wiring —
// byte-identical output while disabled (REQ-CAA-001), the banded oldest-first
// alarm section on both surfaces (REQ-CAA-006), the additive age_alarms JSON
// block with the quota key last, the refused-key degrade with exactly one
// warning line (REQ-CAA-002), the no-write invariant (REQ-CAA-007), and the
// future-commit validity rule on the real status surface (REQ-CAA-004, iter1
// D2).

// caaEnableAlarm writes a workflow.yaml enabling the alarm; body may override
// the bands.
func caaEnableAlarm(t *testing.T, root, body string) {
	t.Helper()
	if body == "" {
		body = "workflow:\n  card_age_alarm:\n    enabled: true\n"
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// caaSeedAlarmCards places one card per band position (all without events, so
// every anchor rides the updated_at fallback): gate-warn at 12h, gate-alert
// at 57h, stall-warn at 4h, and a below-every-threshold card at 30m.
func caaSeedAlarmCards(t *testing.T, root string) {
	t.Helper()
	stamp := func(d time.Duration) string { return fcNow.Add(-d).Format(time.RFC3339Nano) }
	fcPlace(t, root,
		homestate.Card{CardID: "gate-warn", State: homestate.CardKickoff, UpdatedAt: stamp(12 * time.Hour)},
		homestate.Card{CardID: "gate-alert", State: homestate.CardNeedsDecision, UpdatedAt: stamp(57 * time.Hour)},
		homestate.Card{CardID: "stall-warn", State: homestate.CardRun, UpdatedAt: stamp(4 * time.Hour)},
		homestate.Card{CardID: "below", State: homestate.CardPlan, UpdatedAt: stamp(30 * time.Minute)},
	)
}

// TestFactoryStatusDisabledUnchanged is the REQ-CAA-001 regression guard
// (AC-CAA-007): with the flag unset the output is byte-identical to the
// pre-SPEC golden over the same seeded card set — the four card rows the
// pre-change renderer prints, nothing else.
func TestFactoryStatusDisabledUnchanged(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaSeedAlarmCards(t, root)
	out, _, err := runFactory(t, "status")
	if err != nil {
		t.Fatalf("factory status: %v", err)
	}
	golden := "below run=run-cli state=plan stage=- version=1 owner=- lease=none gate=- prefer=- after=- contract=none mode=serial priority=normal\n" +
		"gate-alert run=run-cli state=needs-decision stage=- version=1 owner=- lease=none gate=- prefer=- after=- contract=none mode=serial priority=normal\n" +
		"gate-warn run=run-cli state=kickoff stage=- version=1 owner=- lease=none gate=- prefer=- after=- contract=none mode=serial priority=normal\n" +
		"stall-warn run=run-cli state=run stage=- version=1 owner=- lease=none gate=- prefer=- after=- contract=none mode=serial priority=normal\n"
	if out != golden {
		t.Fatalf("disabled status output is not the pre-SPEC golden:\n--- got ---\n%s--- want ---\n%s", out, golden)
	}
	// An explicitly disabled flag is byte-identical to the unset flag.
	caaEnableAlarm(t, root, "workflow:\n  card_age_alarm:\n    enabled: false\n")
	out2, _, err := runFactory(t, "status")
	if err != nil {
		t.Fatalf("factory status (enabled: false): %v", err)
	}
	if out2 != golden {
		t.Fatalf("explicitly disabled output differs from the unset-flag golden:\n%s", out2)
	}
}

// TestFactoryStatusAgeAlarmText covers REQ-CAA-006's text half: enabled, the
// alarm section renders one age-alarm line per over-threshold card, oldest
// first, banded, carrying the class age and the three signal ages, and the
// card rows keep their shape.
func TestFactoryStatusAgeAlarmText(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaSeedAlarmCards(t, root)
	caaEnableAlarm(t, root, "")
	out, _, err := runFactory(t, "status")
	if err != nil {
		t.Fatalf("factory status: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	var alarms []string
	for _, l := range lines {
		if strings.HasPrefix(l, "age-alarm ") {
			alarms = append(alarms, l)
		}
	}
	if len(alarms) != 3 {
		t.Fatalf("got %d age-alarm lines, want 3 (the below-threshold card renders none):\n%s", len(alarms), out)
	}
	wantOrder := []struct {
		card, state, class, band, age string
	}{
		{"gate-alert", "needs-decision", "gate-wait", "alert", "57h0m0s"},
		{"gate-warn", "kickoff", "gate-wait", "warn", "12h0m0s"},
		{"stall-warn", "run", "no-progress", "warn", "4h0m0s"},
	}
	for i, w := range wantOrder {
		want := "age-alarm card=" + w.card + " run=run-cli state=" + w.state +
			" class=" + w.class + " band=" + w.band + " age=" + w.age +
			" stage=unknown commit=unknown activity=unknown"
		if alarms[i] != want {
			t.Fatalf("alarm line %d =\n  %s\nwant\n  %s", i, alarms[i], want)
		}
	}
	// The card rows render unchanged beside the alarm section.
	for _, card := range []string{"gate-alert", "gate-warn", "stall-warn", "below"} {
		if !strings.Contains(out, card+" run=run-cli state=") {
			t.Fatalf("card row of %s missing from the enabled output:\n%s", card, out)
		}
	}
}

// TestFactoryStatusAgeAlarmJSON covers REQ-CAA-006's JSON half: the additive
// age_alarms array appears when enabled with an over-threshold card, stays
// absent when disabled, and every pre-existing key keeps its place with the
// quota block last.
func TestFactoryStatusAgeAlarmJSON(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaSeedAlarmCards(t, root)

	out, _, err := runFactory(t, "status", "--json")
	if err != nil {
		t.Fatalf("factory status --json (disabled): %v", err)
	}
	var disabled map[string]any
	if err := json.Unmarshal([]byte(out), &disabled); err != nil {
		t.Fatalf("disabled output is not JSON: %v", err)
	}
	if _, has := disabled["age_alarms"]; has {
		t.Fatal("age_alarms key present while the alarm is disabled")
	}

	caaEnableAlarm(t, root, "")
	out, _, err = runFactory(t, "status", "--json")
	if err != nil {
		t.Fatalf("factory status --json (enabled): %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("enabled output is not JSON: %v", err)
	}
	alarms, ok := got["age_alarms"].([]any)
	if !ok || len(alarms) != 3 {
		t.Fatalf("age_alarms = %#v, want 3 rows", got["age_alarms"])
	}
	first, _ := alarms[0].(map[string]any)
	if first["card"] != "gate-alert" || first["class"] != "gate-wait" || first["band"] != "alert" {
		t.Fatalf("first age_alarms row = %#v, want the oldest gate-alert card", first)
	}
	if first["age"] != "57h0m0s" {
		t.Fatalf("age_alarms row age = %#v, want 57h0m0s", first["age"])
	}
	if _, has := first["stage_age"]; !has {
		t.Fatalf("age_alarms row carries no stage_age signal: %#v", first)
	}
	// Pre-existing keys all present, and age_alarms sits before the quota
	// block so every pre-existing key keeps its place.
	for _, key := range []string{"run", "cards", "unavailable", "lease_duration", "renew_interval"} {
		if _, has := got[key]; !has {
			t.Fatalf("pre-existing key %q missing from the enabled report", key)
		}
	}
	if i, j := strings.Index(out, "\"renew_interval\""), strings.Index(out, "\"age_alarms\""); i < 0 || j < 0 || i > j {
		t.Fatalf("age_alarms does not follow renew_interval in the serialized report:\n%s", out)
	}
	if i, j := strings.Index(out, "\"age_alarms\""), strings.Index(out, "\"quota\""); j >= 0 && i > j {
		t.Fatal("age_alarms appears after the quota block; quota must stay last")
	}
}

// TestFactoryStatusAgeAlarmRefusedKey covers REQ-CAA-002's degrade: a key
// outside its bound omits the alarm section, prints exactly one warning line
// naming the refused key, and never refuses the rest of the status output.
func TestFactoryStatusAgeAlarmRefusedKey(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaSeedAlarmCards(t, root)
	caaEnableAlarm(t, root, "workflow:\n  card_age_alarm:\n    enabled: true\n    gate_warn_after: 0s\n")
	out, _, err := runFactory(t, "status")
	if err != nil {
		t.Fatalf("factory status with a refused key: %v", err)
	}
	if strings.Contains(out, "age-alarm card=") {
		t.Fatalf("alarm rows rendered despite the refused key:\n%s", out)
	}
	warnings := 0
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.Contains(l, "age-alarm warning:") {
			warnings++
		}
	}
	if warnings != 1 {
		t.Fatalf("got %d age-alarm warning lines, want exactly 1:\n%s", warnings, out)
	}
	if !strings.Contains(out, "workflow.card_age_alarm.gate_warn_after") {
		t.Fatalf("the warning does not name the refused key:\n%s", out)
	}
	if !strings.Contains(out, "gate-warn run=run-cli state=kickoff") {
		t.Fatalf("the card rows did not survive the refused key:\n%s", out)
	}
	// Sync-audit iter1 F4: the JSON surface carries the same degrade — the
	// age_alarm_warning key present naming the refused key, no age_alarms
	// block.
	outJSON, _, err := runFactory(t, "status", "--json")
	if err != nil {
		t.Fatalf("factory status --json with a refused key: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(outJSON), &got); err != nil {
		t.Fatalf("refused-key output is not JSON: %v", err)
	}
	if _, has := got["age_alarms"]; has {
		t.Fatalf("age_alarms rendered in JSON despite the refused key:\n%s", outJSON)
	}
	warning, _ := got["age_alarm_warning"].(string)
	if warning == "" || !strings.Contains(warning, "workflow.card_age_alarm.gate_warn_after") {
		t.Fatalf("age_alarm_warning missing or not naming the refused key in JSON:\n%s", outJSON)
	}
}

// TestFactoryStatusAgeAlarmWarnBoundary pins the warn-band admission edge of
// sync-audit iter1 F5: a card whose class age equals its class's warn band
// exactly crosses the threshold (the SPEC's "crosses" read as >=) and alarms
// at band=warn, while a card one second below it renders no row.
func TestFactoryStatusAgeAlarmWarnBoundary(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	fcPlace(t, root,
		homestate.Card{CardID: "at-boundary", State: homestate.CardKickoff,
			UpdatedAt: fcNow.Add(-6 * time.Hour).Format(time.RFC3339Nano)},
		homestate.Card{CardID: "below-boundary", State: homestate.CardKickoff,
			UpdatedAt: fcNow.Add(-(5*time.Hour + 59*time.Minute + 59*time.Second)).Format(time.RFC3339Nano)},
	)
	caaEnableAlarm(t, root, "") // compiled defaults: gate warn band 6h
	out, _, err := runFactory(t, "status")
	if err != nil {
		t.Fatalf("factory status: %v", err)
	}
	var atLine string
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.HasPrefix(l, "age-alarm card=at-boundary ") {
			atLine = l
		}
	}
	if atLine == "" {
		t.Fatalf("a card exactly at the warn band did not alarm (F5: age == warn_after must alarm):\n%s", out)
	}
	want := "age-alarm card=at-boundary run=run-cli state=kickoff class=gate-wait band=warn age=6h0m0s" +
		" stage=unknown commit=unknown activity=unknown"
	if atLine != want {
		t.Fatalf("boundary alarm line =\n  %s\nwant\n  %s", atLine, want)
	}
	if strings.Contains(out, "age-alarm card=below-boundary ") {
		t.Fatalf("a card below the warn band alarmed:\n%s", out)
	}
}

// TestAgeAlarmNoWrites is the REQ-CAA-007 guard (AC-CAA-008): one full alarm
// computation over a seeded DB leaves the events count and every card row's
// version and updated_at unchanged. A mutant that makes the alarm path append
// an event or bump a row fails this test.
func TestAgeAlarmNoWrites(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaSeedAlarmCards(t, root)
	caaEnableAlarm(t, root, "")

	before := caaCardRowsAndEvents(t, root)
	if _, _, err := runFactory(t, "status"); err != nil {
		t.Fatalf("factory status: %v", err)
	}
	if _, _, err := runFactory(t, "status", "--json"); err != nil {
		t.Fatalf("factory status --json: %v", err)
	}
	after := caaCardRowsAndEvents(t, root)
	if before != after {
		t.Fatalf("the alarm computation wrote:\nbefore %s\nafter  %s", before, after)
	}
}

func caaCardRowsAndEvents(t *testing.T, root string) string {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT card_id, version, updated_at FROM cards ORDER BY card_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var b strings.Builder
	for rows.Next() {
		var id, updated string
		var version int64
		if err := rows.Scan(&id, &version, &updated); err != nil {
			t.Fatal(err)
		}
		b.WriteString(id + ":" + strconv.FormatInt(version, 10) + ":" + updated + " ")
	}
	var events int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events`).Scan(&events); err != nil {
		t.Fatal(err)
	}
	b.WriteString("events=" + strconv.Itoa(events))
	return b.String()
}

// caaFutureCommitRepo builds a git repository whose single head commit is
// committer-dated ahead of the reading clock — the skewed committing machine
// of REQ-CAA-004.
func caaFutureCommitRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_COMMITTER_DATE="+fcNow.Add(24*time.Hour).Format(time.RFC3339),
			"GIT_AUTHOR_DATE="+fcNow.Add(-48*time.Hour).Format(time.RFC3339),
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "a.txt")
	git("commit", "-q", "-m", "future committer stamp")
	return dir
}

// TestFactoryStatusAgeAlarmFutureCommit covers REQ-CAA-004's validity rule on
// the live surface (iter1 D2): a head commit committer-dated in the future
// renders the commit signal unknown — excluded from the newest-of, never a
// negative age — and suppresses no alarm band: the stage checkpoint still
// drives the class age.
func TestFactoryStatusAgeAlarmFutureCommit(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaEnableAlarm(t, root, "")
	repo := caaFutureCommitRepo(t)

	// Walk one card into `run` at stamps 10h old, so the stage checkpoint
	// (the T4b event) is the only valid basis member — and it sits past the
	// stall alert band, proving the future commit suppresses no band.
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES('op',1,'2026-09-26T00:00:00Z','2026-09-26T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	stage := fcNow.Add(-10 * time.Hour)
	if _, err := db.RecordPicked(ctx, fcRun, "futcommit", homestate.CardFields{WorktreePath: &repo}, "op", stage); err != nil {
		t.Fatal(err)
	}
	step := func(to string, version int64, at time.Time) homestate.Card {
		c, err := db.Transition(ctx, homestate.TransitionRequest{
			RunID: fcRun, CardID: "futcommit", To: to, ExpectedVersion: version,
			Actor: "op", Owner: "op", Now: at,
		})
		if err != nil {
			t.Fatalf("transition -> %s: %v", to, err)
		}
		return c
	}
	c := step(homestate.CardAssigned, 1, stage.Add(time.Minute))
	c = step(homestate.CardLeased, c.Version, stage.Add(2*time.Minute))
	c = step(homestate.CardRun, c.Version, stage.Add(3*time.Minute))
	_ = db.Close()

	out, _, err := runFactory(t, "status")
	if err != nil {
		t.Fatalf("factory status: %v", err)
	}
	var alarmLine string
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.HasPrefix(l, "age-alarm card=futcommit ") {
			alarmLine = l
		}
	}
	if alarmLine == "" {
		t.Fatalf("no alarm row for the future-commit card — the band was suppressed:\n%s", out)
	}
	want := "age-alarm card=futcommit run=run-cli state=run class=no-progress band=alert age=9h57m0s" +
		" stage=9h57m0s commit=unknown activity=9h58m0s"
	if alarmLine != want {
		t.Fatalf("alarm line =\n  %s\nwant\n  %s", alarmLine, want)
	}
}

// caaDatedRepo builds a git repository whose single head commit is
// committer-dated at — the controlled clock of the env-isolation and
// signature-trap probes of sync-audit iter1. When signatureTrap is set the
// repo also arms the F2 trap: log.showSignature=true plus a gpg.program whose
// fake signer records each invocation into the marker file and "signs" the
// head commit, so any signature verification during a later read (the helper
// runs on both the sign and the verify path) leaves the marker behind.
func caaDatedRepo(t *testing.T, at time.Time, signatureTrap bool) (dir, marker string) {
	t.Helper()
	root := t.TempDir()
	dir = filepath.Join(root, "repo")
	marker = filepath.Join(root, "gpg-ran")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_COMMITTER_DATE="+at.Format(time.RFC3339),
			"GIT_AUTHOR_DATE="+at.Format(time.RFC3339),
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "a.txt")
	if signatureTrap {
		gpg := filepath.Join(root, "fake-gpg.sh")
		script := fmt.Sprintf("#!/bin/sh\necho run >> %q\ncat > /dev/null\n"+
			"printf -- '-----BEGIN PGP SIGNATURE-----\\n\\niHUEABYIAB0WIQTdummy\\n=abcd\\n-----END PGP SIGNATURE-----\\n'\n"+
			"echo '[GNUPG:] SIG_CREATED 0 0 17 0 0 0' >&2\nexit 0\n", marker)
		if err := os.WriteFile(gpg, []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
		git("config", "gpg.program", gpg)
		git("config", "user.signingkey", "DEADBEEF")
		git("config", "log.showSignature", "true")
		git("commit", "-q", "-S", "-m", "dated commit")
	} else {
		git("commit", "-q", "-m", "dated commit")
	}
	return dir, marker
}

// caaLazyFetchTrapRepo builds the sync-audit iter2 N1 trap: a repository whose
// head commit object is missing and whose config arms git's lazy fetch at a
// helper program. extensions.partialClone turns a missing-object read into a
// lazy fetch against remote.origin.url; that URL is an ext:: transport the
// repo config allows (protocol.ext.allow=always), so the helper executes
// whenever the read lazily fetches. The missing object is made the cheap
// honest way — the object store is moved aside after the commit (kept in the
// temp root, removed with it at cleanup) and an empty one recreated; no
// network, no pruning, deterministic. The helper records each invocation in
// the marker file and exits 1, so a lazy fetch both executes repo-designated
// code and fails the read. It deliberately does NOT read stdin: the ext
// transport keeps the pipe open and a reader would block the read.
func caaLazyFetchTrapRepo(t *testing.T, at time.Time) (dir, marker string) {
	t.Helper()
	root := t.TempDir()
	dir = filepath.Join(root, "repo")
	marker = filepath.Join(root, "lazy-fetch-ran")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_COMMITTER_DATE="+at.Format(time.RFC3339),
			"GIT_AUTHOR_DATE="+at.Format(time.RFC3339),
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "a.txt")
	git("commit", "-q", "-m", "head soon missing")
	if err := os.Rename(filepath.Join(dir, ".git", "objects"), filepath.Join(root, "objects-setaside")); err != nil {
		t.Fatal(err)
	}
	for _, sub := range []string{"info", "pack"} {
		if err := os.MkdirAll(filepath.Join(dir, ".git", "objects", sub), 0o750); err != nil {
			t.Fatal(err)
		}
	}
	helper := filepath.Join(root, "fake-ext-helper.sh")
	script := fmt.Sprintf("#!/bin/sh\necho run >> %q\nexit 1\n", marker)
	if err := os.WriteFile(helper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	git("config", "remote.origin.url", "ext::"+helper)
	git("config", "protocol.ext.allow", "always")
	git("config", "extensions.partialClone", "origin")
	// Premise control: the helper script itself runs and leaves the marker,
	// so the marker mechanism is proven before the read's absence is read as
	// isolation (the F2 pattern). The script exits 1 by design — a lazy fetch
	// must fail fast — so the marker, not the exit code, is the premise. The
	// lazy-fetch invocation of the helper is pinned by the leg failing on the
	// pre-fix code — the RED run observes it.
	_ = exec.Command("/bin/sh", helper).Run()
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("premise control: the ext helper ran but left no marker (%s): the marker mechanism is broken", marker)
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	return dir, marker
}

// TestFactoryAgeAlarmGitEnvIsolation pins sync-audit iter1 F1 and F2 and
// sync-audit iter2 N1: the commit-signal reader of REQ-CAA-004 is
// repo-isolated and helper-free. An inherited GIT_DIR / GIT_WORK_TREE (git
// exports them into every hook process) outranks `git -C` and would redirect
// the read to a foreign repository — the commit signal then reports the
// foreign head and a stale-card alarm is suppressed (F1). A repo config with
// log.showSignature=true and a gpg.program executes that helper program
// during the read (F2). The trap is armed in BOTH repositories so the marker
// assertion fails whichever repository a non-isolated read lands in: the
// foreign one (env leak) or the target's own config (env scrub without the
// signature flag — the two fixes must land together). A promisor worktree
// whose head object is missing arms git's lazy fetch: without
// GIT_NO_LAZY_FETCH=1 the read fetches through remote.origin.url — an ext::
// transport the repo config allows — so the repo-designated helper program
// executes (N1); with the pin the read fails without external access and the
// commit signal renders unknown. The N1 leg's mutant discriminator is the
// lazy-fetch marker (both mutant and fixed read render unknown on this
// deterministic miss); the commit=unknown assertion documents the miss, the
// marker assertion kills the mutant.
func TestFactoryAgeAlarmGitEnvIsolation(t *testing.T) {
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaEnableAlarm(t, root, "")
	// Repos first, env after: the setup git calls must run against the repos
	// being created, not the foreign context armed below.
	target, targetMarker := caaDatedRepo(t, fcNow.Add(-5*time.Hour), true)
	foreign, foreignMarker := caaDatedRepo(t, fcNow.Add(-5*time.Hour).Add(-365*24*time.Hour), true)
	lazyRepo, lazyMarker := caaLazyFetchTrapRepo(t, fcNow.Add(-5*time.Hour))
	// Premise control: the fake signer ran during setup, so the marker
	// mechanism itself is proven before its absence is read as isolation.
	for _, m := range []string{targetMarker, foreignMarker} {
		if _, err := os.Stat(m); err != nil {
			t.Fatalf("signature trap never fired during setup (%s): the marker mechanism is broken", m)
		}
		if err := os.Remove(m); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GIT_DIR", filepath.Join(foreign, ".git"))
	t.Setenv("GIT_WORK_TREE", foreign)

	fcPlace(t, root,
		homestate.Card{
			CardID: "enviso", State: homestate.CardRun,
			UpdatedAt:    fcNow.Add(-10 * time.Hour).Format(time.RFC3339Nano),
			WorktreePath: target,
		},
		homestate.Card{
			CardID: "lazyfetch", State: homestate.CardRun,
			UpdatedAt:    fcNow.Add(-10 * time.Hour).Format(time.RFC3339Nano),
			WorktreePath: lazyRepo,
		},
	)
	out, _, err := runFactory(t, "status")
	if err != nil {
		t.Fatalf("factory status: %v", err)
	}
	var alarmLine, lazyLine string
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.HasPrefix(l, "age-alarm card=enviso ") {
			alarmLine = l
		}
		if strings.HasPrefix(l, "age-alarm card=lazyfetch ") {
			lazyLine = l
		}
	}
	if alarmLine == "" {
		t.Fatalf("no alarm row for the env-isolation card:\n%s", out)
	}
	if lazyLine == "" {
		t.Fatalf("no alarm row for the lazy-fetch card:\n%s", out)
	}
	// F1: the commit signal reads the card's worktree head (committer-dated
	// fcNow-5h), never the foreign repository the inherited GIT_DIR names
	// (fcNow-5h-1y). Non-fatal so the F2 leg below reports in the same run.
	if !strings.Contains(alarmLine, " commit=5h0m0s") {
		t.Errorf("commit signal did not read the card's worktree (F1: inherited GIT_DIR outranked -C):\n  %s\nwant commit=5h0m0s (head of %s)", alarmLine, target)
	}
	// F2: no configured helper program executed during the read.
	for repo, m := range map[string]string{"target": targetMarker, "foreign": foreignMarker} {
		if _, err := os.Stat(m); err == nil {
			t.Errorf("the gpg helper of the %s repository executed during the commit read (F2): %s exists", repo, m)
		}
	}
	// N1: the missing head object renders the commit signal unknown (the
	// deterministic miss — the value the fixed code must produce), and no
	// lazy fetch ran the repo-designated ext helper during the read.
	if !strings.Contains(lazyLine, " commit=unknown") {
		t.Errorf("commit signal on the missing-object promisor worktree =\n  %s\nwant commit=unknown", lazyLine)
	}
	if _, err := os.Stat(lazyMarker); err == nil {
		t.Errorf("the ext helper of the promisor worktree executed during the commit read (N1: git lazily fetched through remote.origin.url): %s exists", lazyMarker)
	}
}

// TestFactoryAgeAlarmChildEnvProtocolPin pins sync-audit iter3 N3: the
// version-independent transport block on the age-alarm commit read. The
// lazy-fetch pin (GIT_NO_LAZY_FETCH=1, iter2 N1) is only honored from git
// 2.46.0 — the guard does not exist in older releases (source-verified:
// promisor-remote.c @ v2.43.0..v2.45.0 carries no env check) — so on git
// < 2.46.0 the ext:: transport armed by remote.origin.url +
// protocol.ext.allow=always stays reachable. GIT_ALLOW_PROTOCOL is
// different: is_transport_allowed consults its env allow-list BEFORE any
// repo config (transport.c @ v2.43.0:1070-1076, master:1095-1109), so an
// EMPTY value — an empty allow-list — blocks every transport including
// ext:: regardless of the attacker's repo config, on every supported git
// version.
//
// Seam: the env-map seam, not the promisor marker leg. On git >= 2.46.0 the
// lazy-fetch pin already blocks the marker leg's vector, so the new pin is
// unobservable there (the leg cannot discriminate it); what the marker leg
// cannot see, the constructed env itself can. The helper the production
// code calls is invoked directly and the constructed environment is
// asserted as a map.
func TestFactoryAgeAlarmChildEnvProtocolPin(t *testing.T) {
	env := ageAlarmChildEnv()
	values := make(map[string]string, len(env))
	for _, kv := range env {
		if k, v, ok := strings.Cut(kv, "="); ok {
			values[k] = v
		}
	}
	// N3: present, with an EMPTY value (an empty env allow-list).
	v, ok := values["GIT_ALLOW_PROTOCOL"]
	if !ok {
		// Name only the git-prefixed keys — the full env carries unrelated
		// and sometimes sensitive parent-process values a failure log must
		// not echo.
		gitKeys := make([]string, 0, len(values))
		for k := range values {
			if strings.HasPrefix(k, "GIT_") {
				gitKeys = append(gitKeys, k)
			}
		}
		sort.Strings(gitKeys)
		t.Fatalf("GIT_ALLOW_PROTOCOL absent from the child env's git keys %v: the version-independent transport block (N3) is missing — on git < 2.46.0 the ext:: transport of remote.origin.url stays reachable", gitKeys)
	}
	if v != "" {
		t.Errorf("GIT_ALLOW_PROTOCOL=%q, want the EMPTY allow-list: a non-empty value whitelists transports and lets repo config extend it; an empty one short-circuits is_transport_allowed before any repo config", v)
	}
	// Regression: the earlier pins survive alongside the new one.
	if v, ok := values["GIT_NO_LAZY_FETCH"]; !ok || v != "1" {
		t.Errorf("GIT_NO_LAZY_FETCH = present=%v value=%q, want present with \"1\" (iter2 N1 pin regressed)", ok, v)
	}
	if v, ok := values["GIT_TERMINAL_PROMPT"]; !ok || v != "0" {
		t.Errorf("GIT_TERMINAL_PROMPT = present=%v value=%q, want present with \"0\" (iter1 F1 pin regressed)", ok, v)
	}
}
