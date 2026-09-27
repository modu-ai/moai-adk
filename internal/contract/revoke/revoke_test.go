package revoke_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/revoke"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// isolate points MOAI_HOME at a fresh temporary directory so the contract
// store never touches the real home.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv(paths.EnvHome, t.TempDir())
}

func signHuman(t *testing.T, p *signtest.Project) {
	t.Helper()
	seams, rec := signtest.Seams(true, map[string]string{}, signtest.SpecID)
	res, err := sign.Sign(p.Options(), seams)
	if err != nil || res.Refusal != "" {
		t.Fatalf("human sign: err %v refusal %s\n%s", err, res.Refusal, rec.Out.String())
	}
}

func signLLM(t *testing.T, p *signtest.Project) {
	t.Helper()
	opts := p.ReceiptOptions("llm", "llm")
	p.WriteReceipt(p.Receipt(opts, nil))
	seams, rec := signtest.Seams(false, map[string]string{})
	res, err := sign.Sign(opts, seams)
	if err != nil || res.Refusal != "" {
		t.Fatalf("receipt sign: err %v refusal %s\n%s", err, res.Refusal, rec.Out.String())
	}
}

func seal(t *testing.T, p *signtest.Project) string {
	t.Helper()
	c, err := contract.Decode(p.ReadFile(signtest.SpecRel(signtest.SpecID, contract.ContractFile)))
	if err != nil || c.Signature == nil {
		t.Fatalf("decode signed contract: %v (signature %v)", err, c)
	}
	return c.Signature.Seal
}

func store(t *testing.T, p *signtest.Project) *receipt.Store {
	t.Helper()
	s, err := receipt.Open(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func eventCount(t *testing.T, p *signtest.Project, kind string) int {
	t.Helper()
	ev, err := store(t, p).Events()
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, l := range ev {
		if kind == "" || l.Kind == kind {
			n++
		}
	}
	return n
}

func records(t *testing.T, p *signtest.Project) []string {
	t.Helper()
	m, _ := filepath.Glob(filepath.Join(escalation.RecordDir(p.Root, signtest.Card), "*.md"))
	slices.Sort(m)
	return m
}

func opts(p *signtest.Project) revoke.Options {
	return revoke.Options{Root: p.Root, SpecID: signtest.SpecID, Card: signtest.Card}
}

func seams() revoke.Seams {
	return revoke.Seams{
		Now:     func() time.Time { return time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC) },
		GitHead: func(string) (string, error) { return strings.Repeat("a", 40), nil },
	}
}

// TestRevoke covers the revoke outcomes and the A3 revoke reader (AC-GR-023).
func TestRevoke(t *testing.T) {
	t.Run("human-signed", func(t *testing.T) {
		isolate(t)
		p := signtest.New(t)
		signHuman(t, p)
		s := seal(t, p)
		res, err := revoke.Revoke(opts(p), seams())
		if err != nil || res.Status != revoke.StatusRevoked {
			t.Fatalf("revoke: status %q err %v", res.Status, err)
		}
		if n := eventCount(t, p, receipt.KindRevoke); n != 1 {
			t.Errorf("revoke events = %d, want 1", n)
		}
		fp := escalation.Fingerprint("revoke-operator", s)
		want := escalation.RecordPath(p.Root, signtest.Card, "revoke-operator", fp, 0)
		if got := records(t, p); len(got) != 1 || got[0] != want {
			t.Fatalf("records = %v, want [%s]", got, want)
		}
		data, err := os.ReadFile(want)
		if err != nil {
			t.Fatal(err)
		}
		r, err := escalation.ParseRecord(data)
		if err != nil {
			t.Fatal(err)
		}
		if r.SchemaVersion != 1 || r.Kind != escalation.KindRevoke || r.Class != "revoke-operator" ||
			r.Status != escalation.StatusResolved || r.Decider != "human" || r.Card != signtest.Card ||
			r.Spec != signtest.SpecID || r.Fingerprint != fp {
			t.Errorf("record frontmatter = %+v", r)
		}
		if nd, err := escalation.NeedsDecision(p.Root, signtest.Card); err != nil || nd {
			t.Errorf("A2 NeedsDecision on a revoke record = %v (%v), want false", nd, err)
		}
		if b, err := revoke.Blocked(p.Root, signtest.Card, signtest.SpecID, s); err != nil || !b {
			t.Errorf("reader after revoke = %v (%v), want blocked", b, err)
		}
	})
	t.Run("llm-receipt-signed", func(t *testing.T) {
		isolate(t)
		p := signtest.New(t)
		signLLM(t, p)
		res, err := revoke.Revoke(opts(p), seams())
		if err != nil || res.Status != revoke.StatusRevoked {
			t.Fatalf("revoke: status %q err %v", res.Status, err)
		}
		if len(records(t, p)) != 1 || eventCount(t, p, receipt.KindRevoke) != 1 {
			t.Error("revoke of a receipt signature did not write one event and one record")
		}
	})
	t.Run("repeat-is-idempotent", func(t *testing.T) {
		isolate(t)
		p := signtest.New(t)
		signHuman(t, p)
		if _, err := revoke.Revoke(opts(p), seams()); err != nil {
			t.Fatal(err)
		}
		before, recs := eventCount(t, p, ""), records(t, p)
		snap := p.Snapshot()
		res, err := revoke.Revoke(opts(p), seams())
		if err != nil || res.Status != revoke.StatusAlreadyRevoked {
			t.Fatalf("second revoke: status %q err %v", res.Status, err)
		}
		if eventCount(t, p, "") != before || !slices.Equal(records(t, p), recs) {
			t.Error("second revoke wrote")
		}
		p.AssertUnchanged(t, snap)
	})
	t.Run("unsigned", func(t *testing.T) {
		isolate(t)
		p := signtest.New(t)
		snap := p.Snapshot()
		res, err := revoke.Revoke(opts(p), seams())
		if err != nil || res.Status != revoke.StatusNotSigned {
			t.Fatalf("status %q err %v, want not-signed", res.Status, err)
		}
		if eventCount(t, p, "") != 0 {
			t.Error("unsigned revoke wrote an event")
		}
		p.AssertUnchanged(t, snap)
	})
	t.Run("tampered-store", func(t *testing.T) {
		isolate(t)
		p := signtest.New(t)
		signHuman(t, p)
		st := store(t, p)
		for i := 0; i < 3; i++ {
			if _, err := st.AppendEvent(receipt.KindRevoke, receipt.RevokeEvent{Spec: "x", Card: "y", Seal: "z"}); err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(st.Dir, receipt.EventsFile)
		data, _ := os.ReadFile(path)
		lines := strings.Split(string(data), "\n")
		lines[1] = strings.Replace(lines[1], `"x"`, `"X"`, 1)
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := revoke.Revoke(opts(p), seams())
		if !errors.Is(err, revoke.ErrIntegrity) {
			t.Fatalf("err = %v, want integrity", err)
		}
		if len(records(t, p)) != 0 {
			t.Error("tampered-store revoke wrote a record")
		}
	})
	t.Run("card-mismatch", func(t *testing.T) {
		isolate(t)
		p := signtest.New(t)
		signHuman(t, p)
		o := opts(p)
		o.Card = "t9999"
		_, err := revoke.Revoke(o, seams())
		if !errors.Is(err, revoke.ErrUsage) {
			t.Fatalf("err = %v, want usage", err)
		}
		if eventCount(t, p, "") != 0 || len(records(t, p)) != 0 {
			t.Error("card mismatch wrote")
		}
	})

	// Reader fixtures (r0-r6).
	card, spec := "t4242", "SPEC-READER-001"
	sealNow := strings.Repeat("b", 64)
	write := func(t *testing.T, root string, r escalation.Record) string {
		t.Helper()
		data, err := r.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		path := escalation.RecordPath(root, card, r.Class, r.Fingerprint, 0)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	revokeRec := func(sealed string) escalation.Record {
		return escalation.Record{SchemaVersion: 1, Card: card, Spec: spec, Kind: escalation.KindRevoke,
			Class: "revoke-operator", Fingerprint: escalation.Fingerprint("revoke-operator", sealed),
			Status: escalation.StatusResolved, Decider: "human", Occurrences: 1,
			Observation: "revoked", Options: []string{"re-sign", "abandon"}}
	}
	readerCases := []struct {
		name    string
		setup   func(t *testing.T, root string)
		blocked bool
		wantErr bool
		a2Needs bool
		checkA2 bool
	}{
		{"r0_no_directory", func(*testing.T, string) {}, false, false, false, true},
		{"r1_revoke_record", func(t *testing.T, root string) { write(t, root, revokeRec(sealNow)) }, true, false, false, true},
		{"r2_new_signature", func(t *testing.T, root string) { write(t, root, revokeRec(strings.Repeat("c", 64))) }, false, false, false, false},
		{"r3_other_seal", func(t *testing.T, root string) { write(t, root, revokeRec(strings.Repeat("d", 64))) }, false, false, false, false},
		{"r4_status_open", func(t *testing.T, root string) {
			r := revokeRec(sealNow)
			r.Status = escalation.StatusOpen
			write(t, root, r)
		}, true, false, false, false},
		{"r5_broken_header", func(t *testing.T, root string) {
			dir := escalation.RecordDir(root, card)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "broken.md"), []byte("no frontmatter\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, true, true, false, false},
		{"r6_open_contract_record", func(t *testing.T, root string) {
			write(t, root, escalation.Record{SchemaVersion: 1, Card: card, Spec: spec, Kind: escalation.KindContract,
				Class: escalation.ClassAcceptanceChange, Fingerprint: escalation.Fingerprint(escalation.ClassAcceptanceChange, "x"),
				Status: escalation.StatusOpen, Occurrences: 1, Observation: "o", Options: []string{"a", "b"}})
		}, false, false, true, true},
	}
	for _, rc := range readerCases {
		t.Run("reader/"+rc.name, func(t *testing.T) {
			root := t.TempDir()
			rc.setup(t, root)
			b, err := revoke.Blocked(root, card, spec, sealNow)
			if b != rc.blocked || (err != nil) != rc.wantErr {
				t.Fatalf("Blocked = %v, err %v; want %v, err %v", b, err, rc.blocked, rc.wantErr)
			}
			if rc.checkA2 {
				nd, err := escalation.NeedsDecision(root, card)
				if err != nil || nd != rc.a2Needs {
					t.Fatalf("A2 NeedsDecision = %v (%v), want %v", nd, err, rc.a2Needs)
				}
			}
		})
	}
}
