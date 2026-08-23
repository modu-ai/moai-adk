package feedback

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// queueRoot returns a project root with a .moai marker and a store over its
// canonical queue path.
func queueRoot(t *testing.T) (string, *QueueStore) {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("creating .moai marker: %v", err)
	}
	return root, NewQueueStore(QueuePathForRoot(root))
}

// envSecretValue is a masked-by-name environment value: long enough to be
// masked, and not credential-shaped, so the classifier lets the report through
// and the queue sees the case it exists for — a submission that was allowed,
// was attempted, and failed.
const envSecretValue = "sup3rs3cr3tv4lue"

// scrubbed returns the masked Result for a report whose body carries one
// maskable environment value.
func scrubbed(t *testing.T, title, body string) Result {
	t.Helper()

	opt := testOptions()
	opt.Environ = func() []string { return []string{"GITHUB_TOKEN=" + envSecretValue} }
	res, err := Scrub(Input{Title: title, Body: body}, opt)
	if err != nil {
		t.Fatalf("Scrub returned error: %v", err)
	}
	return res
}

// AC-F-017 — a failed `gh issue create` leaves the report in the queue, and
// what lands there is the MASKED body.
func TestQueueEnqueuesOnSendFailure(t *testing.T) {
	t.Parallel()

	root, store := queueRoot(t)
	token := envSecretValue
	res := scrubbed(t, "init wizard drops the answer", "the token is "+token+" ok")

	item, err := store.EnqueueMasked(res)
	if err != nil {
		t.Fatalf("EnqueueMasked returned error: %v", err)
	}
	if item.ID == "" {
		t.Errorf("enqueued item has no id")
	}

	path := QueuePathForRoot(root)
	if got, want := store.Path(), path; got != want {
		t.Errorf("store path = %q, want %q", got, want)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("queue file missing: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("queue holds %d items, want 1", len(rec.Items))
	}
	got := rec.Items[0]
	if got.Body != res.Body {
		t.Errorf("queued body is not the masked body:\n got %q\nwant %q", got.Body, res.Body)
	}
	if got.Title != res.Title {
		t.Errorf("queued title is not the masked title:\n got %q\nwant %q", got.Title, res.Title)
	}
	if strings.Contains(got.Body, token) {
		t.Fatalf("queued body carries the raw credential")
	}
	if got.QueuedAt == "" {
		t.Errorf("queued item has no timestamp")
	}

	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not POSIX on windows")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat queue: %v", err)
	}
	if perm := info.Mode().Perm(); perm != queueFilePerm {
		t.Errorf("queue perm = %O, want %O", perm, queueFilePerm)
	}
}

// AC-F-018 — a later success removes the item, and the file stays valid JSON.
//
// Removal is why the queue is a single JSON document rather than an
// append-only log (AP-7): an append-only shape cannot express "sent".
func TestQueueResolvesOnSuccess(t *testing.T) {
	t.Parallel()

	root, store := queueRoot(t)
	item, err := store.EnqueueMasked(scrubbed(t, "title", "the token is "+envSecretValue))
	if err != nil {
		t.Fatalf("EnqueueMasked returned error: %v", err)
	}

	removed, err := store.Resolve(item.ID)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if !removed {
		t.Fatalf("Resolve reported no removal for id %q", item.ID)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load after Resolve returned error: %v", err)
	}
	if len(rec.Items) != 0 {
		t.Fatalf("queue holds %d items after resolve, want 0", len(rec.Items))
	}

	raw, err := os.ReadFile(QueuePathForRoot(root)) //nolint:gosec // test-owned temp path
	if err != nil {
		t.Fatalf("reading queue after resolve: %v", err)
	}
	var probe map[string]any
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatalf("queue is not valid JSON after resolve: %v", err)
	}

	removed, err = store.Resolve(item.ID)
	if err != nil {
		t.Fatalf("second Resolve returned error: %v", err)
	}
	if removed {
		t.Errorf("second Resolve reported a removal for an absent id")
	}
}

// A blocked report must never enter the retry queue: the queue exists to
// re-send, and re-sending a report the classifier refused would submit it
// through the back door.
func TestQueueRefusesBlockedResult(t *testing.T) {
	t.Parallel()

	root, store := queueRoot(t)
	res := scrubbed(t, "", "The template renderer is exploitable: a crafted project name gives remote code execution.")
	if res.Verdict != VerdictBlocked {
		t.Fatalf("fixture is not blocked, verdict = %q", res.Verdict)
	}

	if _, err := store.EnqueueMasked(res); err == nil {
		t.Fatalf("EnqueueMasked accepted a blocked result")
	}
	if _, err := os.Stat(QueuePathForRoot(root)); !os.IsNotExist(err) {
		t.Errorf("a refused enqueue created the queue file, stat err = %v", err)
	}
}

// [HARD] (D4) The queue is for `gh issue create` failure only. The pre-submit
// draft (.moai/state/feedback-draft-<ts>.md) holds PRE-SCRUB raw text and is
// owned by a different failure; reading one as a queue entry would put raw
// text into a public issue.
//
// The store's read scope is asserted behaviourally: a draft sitting in the
// same .moai/state tree is invisible to Load, and its raw content never
// reaches an item.
func TestQueueNeverReadsPreScrubDraft(t *testing.T) {
	t.Parallel()

	root, store := queueRoot(t)
	rawSecret := "the token is " + envSecretValue
	draft := filepath.Join(root, ".moai", "state", "feedback-draft-20260823T000000Z.md")
	if err := os.MkdirAll(filepath.Dir(draft), 0o755); err != nil {
		t.Fatalf("creating state dir: %v", err)
	}
	if err := os.WriteFile(draft, []byte("# raw title\n\n"+rawSecret+"\n"), 0o600); err != nil {
		t.Fatalf("writing draft: %v", err)
	}

	if _, err := store.EnqueueMasked(scrubbed(t, "masked title", "an ordinary report")); err != nil {
		t.Fatalf("EnqueueMasked returned error: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(rec.Items) != 1 {
		t.Fatalf("queue holds %d items, want exactly the one enqueued", len(rec.Items))
	}
	for _, it := range rec.Items {
		if strings.Contains(it.Body, "raw title") || strings.Contains(it.Body, rawSecret) {
			t.Fatalf("a draft file was read as a queue entry")
		}
	}
}

// The lock has to span the whole read-modify-write, or concurrent enqueues
// lose updates.
func TestQueueMutateSerializesConcurrentEnqueues(t *testing.T) {
	t.Parallel()

	_, store := queueRoot(t)
	res := scrubbed(t, "title", "an ordinary report")

	const writers, each = 4, 3
	var wg sync.WaitGroup
	errs := make(chan error, writers*each)
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				if _, err := store.EnqueueMasked(res); err != nil {
					errs <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent EnqueueMasked returned error: %v", err)
	}

	rec, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(rec.Items) != writers*each {
		t.Fatalf("queue holds %d items, want %d — a concurrent update was lost", len(rec.Items), writers*each)
	}
	seen := map[string]bool{}
	for _, it := range rec.Items {
		if seen[it.ID] {
			t.Fatalf("duplicate id issued: %q", it.ID)
		}
		seen[it.ID] = true
	}
}
