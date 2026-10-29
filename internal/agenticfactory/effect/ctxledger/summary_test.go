package ctxledger

import (
	"errors"
	"strings"
	"testing"
)

const (
	sha1 = "0123456789abcdef0123456789abcdef01234567"
	sha2 = "89abcdef0123456789abcdef0123456789abcdef"
	tmpl = "https://example.org/o/r/blob/{sha}/{path}"
)

func durable() Checkpoint {
	return Checkpoint{
		Issue:        7,
		Commit:       Durable{SHA: sha1, RemoteVerified: true},
		ACK:          Durable{SHA: sha1, RemoteVerified: true},
		Manifest:     Manifest{Path: ".moai/evidence/issue-7/manifest.json", Hash: "sha256:abc123", Durable: Durable{SHA: sha1, RemoteVerified: true}},
		LinkTemplate: tmpl,
	}
}

func counted() (*Ledger, *CountingRemote, *fakeRemote) {
	r := newRemote()
	c := &CountingRemote{Inner: r}
	l, _ := newLedger(c, 4096)
	return l, c, r
}

func TestNothingIsPublishedUntilEveryEarlierStepIsRemoteDurable(t *testing.T) {
	cases := map[string]func(*Checkpoint){
		"commit not remote-verified":   func(c *Checkpoint) { c.Commit.RemoteVerified = false },
		"ack not remote-verified":      func(c *Checkpoint) { c.ACK.RemoteVerified = false },
		"manifest not remote-verified": func(c *Checkpoint) { c.Manifest.Durable.RemoteVerified = false },
		"commit sha missing":           func(c *Checkpoint) { c.Commit.SHA = "" },
		"manifest at another sha":      func(c *Checkpoint) { c.Manifest.Durable.SHA = sha2 },
		"ack at another sha":           func(c *Checkpoint) { c.ACK.SHA = sha2 },
		"short sha":                    func(c *Checkpoint) { c.Commit.SHA, c.ACK.SHA, c.Manifest.Durable.SHA = "0123abc", "0123abc", "0123abc" },
		"upper-case sha": func(c *Checkpoint) {
			u := strings.ToUpper(sha1)
			c.Commit.SHA, c.ACK.SHA, c.Manifest.Durable.SHA = u, u, u
		},
		"no manifest hash": func(c *Checkpoint) { c.Manifest.Hash = "" },
	}
	for name, mut := range cases {
		l, c, r := counted()
		cp := durable()
		mut(&cp)
		res := l.PublishSummaryLink(t.Context(), cp)
		if res.Status != StepSkipped || res.Reason == "" {
			t.Errorf("%s: %+v, want skipped with a reason", name, res)
		}
		if n := c.Snapshot(); n.Total() != 0 || len(r.comments[7]) != 0 {
			t.Errorf("%s: remote touched (%+v)", name, n)
		}
	}
}

func TestADurableCheckpointPublishesOneCommentPinnedToTheRemoteVerifiedSHA(t *testing.T) {
	l, c, r := counted()
	res := l.PublishSummaryLink(t.Context(), durable())
	if res.Status != StepPublished || res.Err != nil {
		t.Fatalf("res = %+v", res)
	}
	want := "https://example.org/o/r/blob/" + sha1 + "/.moai/evidence/issue-7/manifest.json"
	if res.Link != want {
		t.Errorf("link = %q, want %q", res.Link, want)
	}
	if len(r.comments[7]) != 1 || !strings.Contains(r.comments[7][0].Body, want) || !strings.Contains(r.comments[7][0].Body, sha1) {
		t.Fatalf("comments = %+v", r.comments[7])
	}
	if n := c.Snapshot(); n.Create != 1 || n.Update != 0 {
		t.Errorf("calls = %+v", n)
	}
}

func TestTheSameCheckpointAgainIsIdempotent(t *testing.T) {
	l, c, r := counted()
	l.PublishSummaryLink(t.Context(), durable())
	before := c.Snapshot()
	res := l.PublishSummaryLink(t.Context(), durable())
	after := c.Snapshot()
	if res.Status != StepPublished || len(r.comments[7]) != 1 {
		t.Fatalf("res=%+v comments=%d", res, len(r.comments[7]))
	}
	if after.Create != before.Create || after.Update != before.Update {
		t.Errorf("a repeat wrote again: %+v -> %+v", before, after)
	}
}

func TestANewCheckpointUpdatesTheSameCommentInPlace(t *testing.T) {
	l, _, r := counted()
	l.PublishSummaryLink(t.Context(), durable())
	next := durable()
	next.Commit.SHA, next.ACK.SHA, next.Manifest.Durable.SHA = sha2, sha2, sha2
	if res := l.PublishSummaryLink(t.Context(), next); res.Status != StepPublished {
		t.Fatalf("res = %+v", res)
	}
	if len(r.comments[7]) != 1 || !strings.Contains(r.comments[7][0].Body, sha2) || strings.Contains(r.comments[7][0].Body, sha1) {
		t.Errorf("comments = %+v", r.comments[7])
	}
}

// A failed comment is a failed step, reported by value. It is not an error the caller would mistake for
// a failed recovery, and a retry with the same checkpoint finishes the job without a second comment.
func TestAFailedCommentOnlyFailsItsOwnStepAndARetryCompletesIt(t *testing.T) {
	l, _, r := counted()
	r.createErr = status(502)
	res := l.PublishSummaryLink(t.Context(), durable())
	if res.Status != StepFailed || res.Err == nil || res.Link != "" {
		t.Fatalf("res = %+v", res)
	}
	r.createErr = nil
	res = l.PublishSummaryLink(t.Context(), durable())
	if res.Status != StepPublished || len(r.comments[7]) != 1 {
		t.Errorf("retry: %+v comments=%d", res, len(r.comments[7]))
	}
}

func TestADeniedCommentIsParkedAndNotRetried(t *testing.T) {
	l, c, r := counted()
	r.createErr = status(403, "x-ratelimit-remaining", "4000")
	if res := l.PublishSummaryLink(t.Context(), durable()); res.Status != StepFailed {
		t.Fatalf("res = %+v", res)
	}
	calls := c.Snapshot().Total()
	for i := 0; i < 3; i++ {
		l.PublishSummaryLink(t.Context(), durable())
	}
	if after := c.Snapshot().Total(); after != calls {
		t.Errorf("a denied step kept calling the remote: %d -> %d", calls, after)
	}
}

func TestUnsafeLinksAndTextAreRefusedBeforeAnyCall(t *testing.T) {
	cases := map[string]func(*Checkpoint){
		"http link":           func(c *Checkpoint) { c.LinkTemplate = "http://example.org/{sha}/{path}" },
		"credentials in link": func(c *Checkpoint) { c.LinkTemplate = "https://user:secret@example.org/{sha}/{path}" },
		"token in query":      func(c *Checkpoint) { c.LinkTemplate = "https://example.org/{sha}/{path}?token=abc123def456ghi789" },
		"no sha placeholder":  func(c *Checkpoint) { c.LinkTemplate = "https://example.org/main/{path}" },
		"home path in path":   func(c *Checkpoint) { c.Manifest.Path = "/Us" + "ers/someone/m.json" },
		"email in path":       func(c *Checkpoint) { c.Manifest.Path = "person" + "@" + "example.org/m.json" },
		"dot segments":        func(c *Checkpoint) { c.Manifest.Path = "../../m.json" },
		"empty path":          func(c *Checkpoint) { c.Manifest.Path = "" },
		"secret in hash":      func(c *Checkpoint) { c.Manifest.Hash = "ghp_abcdefghijklmnopqrstuvwxyz0123456789" },
	}
	for name, mut := range cases {
		l, c, r := counted()
		cp := durable()
		mut(&cp)
		res := l.PublishSummaryLink(t.Context(), cp)
		if res.Status == StepPublished || res.Err == nil && res.Status != StepSkipped {
			t.Errorf("%s: %+v", name, res)
		}
		if c.Snapshot().Total() != 0 || len(r.comments[7]) != 0 {
			t.Errorf("%s: remote touched", name)
		}
	}
}

func TestTheCountingRemoteCountsEveryKindOfCall(t *testing.T) {
	r := newRemote()
	c := &CountingRemote{Inner: r}
	ctx := t.Context()
	_, _, _ = c.Find(ctx, 7, "m")
	cm, _ := c.Create(ctx, 7, "b")
	_ = c.Update(ctx, 7, cm.ID, "b2")
	n := c.Snapshot()
	if n.Find != 1 || n.Create != 1 || n.Update != 1 || n.Total() != 3 {
		t.Errorf("counts = %+v", n)
	}
	c.Reset()
	if c.Snapshot().Total() != 0 {
		t.Error("reset did not clear the counts")
	}
	var _ error = errors.New("keep the import honest")
}
