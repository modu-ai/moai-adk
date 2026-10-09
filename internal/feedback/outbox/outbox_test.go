package outbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/feedback"
)

// moaiHomes tracks each test's ESTABLISHED moai home (card-review finding,
// P1): fixtures must never fall back to the ambient MOAI_HOME — a suite run
// under a real user home used to inherit it, flip its consent, and create
// store files in it (measured on a canary home before this fix). Every
// home a test touches is one these helpers minted.
var moaiHomes sync.Map // *testing.T → established home path

// freshTestHome establishes a FRESH temporary moai home for this test —
// for arms that need a clean slate mid-test.
func freshTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	moaiHomes.Store(t, home)
	t.Cleanup(func() { moaiHomes.Delete(t) })
	return home
}

// testHome returns the test's established home, creating one on first use.
// It NEVER reads the ambient MOAI_HOME.
func testHome(t *testing.T) string {
	t.Helper()
	if v, ok := moaiHomes.Load(t); ok {
		return v.(string)
	}
	return freshTestHome(t)
}

// jsonMarshal / jsonUnmarshalObj are tiny wrappers so the fixture helpers
// keep one import block.
func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }

func jsonUnmarshalObj(line string) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// spoolFixture seeds the user-scoped spool with one entry per kind. The
// FIRST call in a test establishes the test's isolated moai home (testHome
// — never the ambient MOAI_HOME); later calls reuse it (a fresh home per
// call would wipe the consent, ledger, and queue the earlier drains wrote).
func spoolFixture(t *testing.T, entries ...bugreport.Kind) string {
	t.Helper()
	home := testHome(t)
	t.Setenv("CI", "")
	dir := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}
	var b strings.Builder
	for _, kind := range entries {
		entry := map[string]any{
			"kind":    string(kind),
			"verdict": string(bugreport.VerdictMoai),
			"frames":  []string{"internal/cli.Execute"},
		}
		if kind == bugreport.KindHookTimeout {
			entry["verdict"] = string(bugreport.VerdictAmbiguous)
		}
		line, err := jsonMarshal(entry)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	path := filepath.Join(dir, "spool.jsonl")
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		t.Fatalf("write spool: %v", err)
	}
	return path
}

func consentOn(t *testing.T) {
	t.Helper()
	home := testHome(t)
	configDir := filepath.Join(home, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	body := "participation:\n  enabled: true\n  asked: true\n"
	if err := os.WriteFile(filepath.Join(configDir, "participation.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("seed consent: %v", err)
	}
}

func outboxRows(t *testing.T) []map[string]any {
	t.Helper()
	home := testHome(t)
	raw, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir), "outbox.log"))
	if err != nil {
		return nil
	}
	var rows []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" {
			continue
		}
		m, err := jsonUnmarshalObj(line)
		if err != nil {
			t.Fatalf("outbox line: %v", err)
		}
		rows = append(rows, m)
	}
	return rows
}

func countOutcome(rows []map[string]any, outcome string) int {
	n := 0
	for _, r := range rows {
		if r["outcome"] == outcome {
			n++
		}
	}
	return n
}

func hasOutcome(rows []map[string]any, outcome string) bool {
	return countOutcome(rows, outcome) > 0
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

// validBuildForTest seeds a build identity that passes the payload
// allowlists: the go-test binary's compiled identity carries commit "none",
// which the commit allowlist rightly refuses (a dev build must not
// publish), so every test that reaches payload build seeds this.
func validBuildForTest(t *testing.T) {
	t.Helper()
	prev := buildIdentityForTest
	buildIdentityForTest = func() (string, string) {
		return "v3.2.0", "abcdef1234567"
	}
	t.Cleanup(func() { buildIdentityForTest = prev })
}

// withBuildForTest varies the build identity the fingerprint uses so each
// drain in a test sees a fresh fingerprint (the caps test needs distinct
// fingerprints hour by hour; production reads the compiled identity).
func withBuildForTest(t *testing.T, seq int) {
	t.Helper()
	prev := buildIdentityForTest
	buildIdentityForTest = func() (string, string) {
		return "v3.2.0-test" + itoa(seq), "abcdef1234567"
	}
	t.Cleanup(func() { buildIdentityForTest = prev })
}

func TestAmbiguousRetainedLocally(t *testing.T) {
	spoolFixture(t, bugreport.KindHookTimeout)
	consentOn(t)

	if err := Drain(); err != nil {
		t.Fatalf("Drain: %v", err)
	}

	// Nothing queued: the bugreport queue stays absent.
	home := os.Getenv("MOAI_HOME")
	qPath := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir), "queue.json")
	if _, err := os.Stat(qPath); err == nil {
		t.Fatal("an ambiguous signal produced a queue file")
	}

	// One outbox row names the retention.
	rows := outboxRows(t)
	if len(rows) != 1 {
		t.Fatalf("outbox rows = %d, want 1", len(rows))
	}
	if rows[0]["outcome"] != "ambiguous" {
		t.Fatalf("outcome = %v, want ambiguous", rows[0]["outcome"])
	}
	if s, _ := rows[0]["reason"].(string); !strings.Contains(s, "retained locally") {
		t.Fatalf("reason %v does not name the retention", rows[0]["reason"])
	}

	// The spool is consumed: a second drain logs nothing new.
	if err := Drain(); err != nil {
		t.Fatalf("second Drain: %v", err)
	}
	if rows := outboxRows(t); len(rows) != 1 {
		t.Fatalf("outbox rows after re-drain = %d, want 1 (the spool was consumed)", len(rows))
	}
}

func TestAmbiguousMakesNoModelCall(t *testing.T) {
	spoolFixture(t, bugreport.KindHookTimeout)
	consentOn(t)

	// The outbox package has no model seam and cannot reach one: it imports
	// neither os/exec nor net/http (REQ-ANON-025, enforced by the publish
	// package's import guard). The observable half: the drain writes only
	// the retention row — no queue item, nothing else.
	if err := Drain(); err != nil {
		t.Fatalf("Drain: %v", err)
	}
	rows := outboxRows(t)
	if len(rows) != 1 || rows[0]["outcome"] != "ambiguous" {
		t.Fatalf("drain rows = %+v, want exactly one ambiguous retention row", rows)
	}
}

func TestOutboundPayloadScrubbedAndClassified(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	// A clean payload passes with zero masking findings: the queue carries
	// the contract title, and the outbox row records queued with no masking.
	if err := Drain(); err != nil {
		t.Fatalf("Drain: %v", err)
	}
	home := os.Getenv("MOAI_HOME")
	raw, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir), "queue.json"))
	if err != nil {
		t.Fatalf("read queue: %v", err)
	}
	if !strings.Contains(string(raw), "[auto-report] panic ") {
		t.Fatalf("queue lacks the contract title:\n%s", raw)
	}
	rows := outboxRows(t)
	if len(rows) != 1 || rows[0]["outcome"] != "queued" {
		t.Fatalf("rows = %+v, want one queued row", rows)
	}
	if m, _ := rows[0]["masked"].(bool); m {
		t.Fatal("a clean payload was masked")
	}
}

func TestBlockedOrMaskedPayloadWithheld(t *testing.T) {
	// A payload whose rendered body carries a credential-shaped string trips
	// the scrub tripwire: withheld, not queued, one withheld row. The
	// fixture is assembled from parts (the same shape the scrub tests use)
	// so the source file never carries the literal.
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	tripwireInputForTest = func() (string, string) {
		secret := "AKIA" + "IOSFODNN7EXAMPLE"
		return "[auto-report] panic abcdef1234567890",
			"<!-- moai-bugreport:v1 schema=v1 -->\n" + secret + " in the frames"
	}
	t.Cleanup(func() { tripwireInputForTest = nil })

	if err := Drain(); err != nil {
		t.Fatalf("Drain: %v", err)
	}
	home := os.Getenv("MOAI_HOME")
	if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir), "queue.json")); err == nil {
		t.Fatal("a blocked payload was queued")
	}
	rows := outboxRows(t)
	if len(rows) != 1 || rows[0]["outcome"] != "withheld" {
		t.Fatalf("rows = %+v, want one withheld row", rows)
	}
}

func TestPathTraversalKindWithheld(t *testing.T) {
	spoolFixture(t)
	consentOn(t)
	validBuildForTest(t)

	// The detail token path_traversal is always withheld (REQ-ANON-012).
	home := os.Getenv("MOAI_HOME")
	dir := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir))
	spool := filepath.Join(dir, "spool.jsonl")
	body := "{\"kind\":\"template_deploy_failure\",\"verdict\":\"moai\",\"reason\":\"path_traversal\",\"frames\":[\"internal/cli.Execute\"]}\n"
	if err := os.WriteFile(spool, []byte(body), 0o600); err != nil {
		t.Fatalf("write spool: %v", err)
	}

	if err := Drain(); err != nil {
		t.Fatalf("Drain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "queue.json")); err == nil {
		t.Fatal("a path-traversal report was queued")
	}
	rows := outboxRows(t)
	if len(rows) != 1 || rows[0]["outcome"] != "withheld" {
		t.Fatalf("rows = %+v, want one withheld row", rows)
	}
}

func TestDedupeWindow(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	now := time.Now()
	SetClockForTest(func() time.Time { return now })
	t.Cleanup(func() { SetClockForTest(nil) })

	if err := Drain(); err != nil {
		t.Fatalf("first drain: %v", err)
	}
	if rows := outboxRows(t); len(rows) != 1 || rows[0]["outcome"] != "queued" {
		t.Fatalf("first drain rows = %+v", outboxRows(t))
	}

	// The same fingerprint again, inside the 7-day window: deduplicated.
	spoolFixture(t, bugreport.KindPanic)
	if err := Drain(); err != nil {
		t.Fatalf("second drain: %v", err)
	}
	rows := outboxRows(t)
	if len(rows) != 2 || rows[1]["outcome"] != "deduped" {
		t.Fatalf("rows = %+v, want a deduped second row", rows)
	}

	// Outside the window: queued again.
	now = now.Add(8 * 24 * time.Hour)
	spoolFixture(t, bugreport.KindPanic)
	if err := Drain(); err != nil {
		t.Fatalf("third drain: %v", err)
	}
	rows = outboxRows(t)
	if len(rows) != 3 || rows[2]["outcome"] != "queued" {
		t.Fatalf("rows = %+v, want a queued third row outside the window", rows)
	}
}

func TestGlobalCaps(t *testing.T) {
	// Fresh fingerprints, hour by hour: the daily cap (3) stops the fourth
	// within the rolling 24 hours. The fixture establishes MOAI_HOME first —
	// consentOn writes to the home that is current at its call time.
	spoolFixture(t)
	consentOn(t)

	base := time.Now()
	day := 0
	SetClockForTest(func() time.Time {
		return base.Add(time.Duration(day) * time.Hour)
	})
	t.Cleanup(func() { SetClockForTest(nil) })

	for day = 0; day < 3; day++ {
		spoolFixture(t, bugreport.KindPanic)
		// A distinct build identity per signal, so the fingerprint differs.
		withBuildForTest(t, day)
		if err := Drain(); err != nil {
			t.Fatalf("drain %d: %v", day, err)
		}
	}
	// The fourth inside the same rolling day: capped.
	spoolFixture(t, bugreport.KindPanic)
	withBuildForTest(t, 99)
	if err := Drain(); err != nil {
		t.Fatalf("fourth drain: %v", err)
	}
	rows := outboxRows(t)
	if rows[len(rows)-1]["outcome"] != "capped" {
		t.Fatalf("fourth drain outcome = %v, want capped", rows[len(rows)-1]["outcome"])
	}
	if got := countOutcome(rows, "queued"); got != 3 {
		t.Fatalf("queued count = %d, want the daily cap 3", got)
	}
}

func TestQueueBound(t *testing.T) {
	spoolFixture(t)
	consentOn(t)

	// Seed bound+1 items directly into the queue store; the oldest is
	// dropped at the bound.
	store := BugreportQueueStore()
	err := store.Mutate(func(rec *feedback.QueueRecord) error {
		rec.Items = nil
		for i := 0; i < config.DefaultBugreportQueueBound+1; i++ {
			rec.Items = append(rec.Items, feedback.QueueItem{
				ID:    "f" + itoa(i+1),
				Title: "[auto-report] panic " + itoa(i+100) + "aaaa",
				Body:  "<!-- moai-bugreport:v1 -->",
			})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed queue: %v", err)
	}

	dropped := EnforceQueueBound()
	if dropped != 1 {
		t.Fatalf("EnforceQueueBound dropped %d, want 1 (the oldest beyond the bound)", dropped)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Items) != config.DefaultBugreportQueueBound {
		t.Fatalf("queue length = %d, want the bound %d", len(rec.Items), config.DefaultBugreportQueueBound)
	}
	if rec.Items[0].ID != "f2" {
		t.Fatalf("oldest item = %q, want f2 (f1 was dropped)", rec.Items[0].ID)
	}
}

func TestAttemptsIncrementedAndCapped(t *testing.T) {
	// The sender increments Attempts per failed attempt; at the attempt
	// limit the item is dropped with a log row.
	item := feedback.QueueItem{ID: "f1", Attempts: 0}
	for i := 1; i <= config.DefaultBugreportAttemptLimit; i++ {
		item.Attempts = i
		if i < config.DefaultBugreportAttemptLimit && AttemptLimitReached(item) {
			t.Fatalf("attempt %d reported the limit reached", i)
		}
	}
	if !AttemptLimitReached(item) {
		t.Fatalf("attempt %d did not report the limit", config.DefaultBugreportAttemptLimit)
	}
}

func TestWithdrawalDiscardsQueueOnFlush(t *testing.T) {
	spoolFixture(t, bugreport.KindPanic)
	consentOn(t)
	validBuildForTest(t)

	// Drain once with consent on: one queued item.
	if err := Drain(); err != nil {
		t.Fatalf("drain: %v", err)
	}

	// Consent withdrawn: the next flush discards the unsent item and the
	// spool, keeps the sent history, and makes no network request (there is
	// no network code in this package — REQ-ANON-025).
	home := os.Getenv("MOAI_HOME")
	configDir := filepath.Join(home, "config")
	if err := os.WriteFile(filepath.Join(configDir, "participation.yaml"), []byte("participation:\n  enabled: false\n  asked: true\n"), 0o600); err != nil {
		t.Fatalf("withdraw consent: %v", err)
	}

	if err := Drain(); err != nil {
		t.Fatalf("drain after withdrawal: %v", err)
	}

	dir := filepath.Join(home, filepath.FromSlash(bugreport.BugreportStoreDir))
	if _, err := os.Stat(filepath.Join(dir, "queue.json")); err == nil {
		qraw, _ := os.ReadFile(filepath.Join(dir, "queue.json"))
		if !strings.Contains(string(qraw), "[]") {
			t.Fatalf("unsent items survived withdrawal:\n%s", qraw)
		}
	}
	rows := outboxRows(t)
	last := rows[len(rows)-1]
	if last["outcome"] != "discarded" {
		t.Fatalf("last row outcome = %v, want discarded", last["outcome"])
	}
	// The queued row from the first drain survives (the history stays).
	if !hasOutcome(rows, "queued") {
		t.Fatal("the earlier queued row was lost; the history must be kept")
	}
	// The spool is discarded.
	if _, err := os.Stat(filepath.Join(dir, "spool.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("spool survived withdrawal: %v", err)
	}
}
