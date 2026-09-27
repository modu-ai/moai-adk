package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/runtime"
)

// prepareDecide makes the fixture's decide preconditions hold.
func prepareDecide(t *testing.T, p *signtest.Project) {
	t.Helper()
	spec := ".moai/specs/" + signtest.SpecID
	p.WriteFile(spec+"/research.md", "# research\n")
	p.Git("add", "-A")
	p.Git("commit", "-q", "-m", "docs: plan\n\nAuthored-By-Agent: manager-spec\n\n🗿 MoAI\n")
	h, err := runtime.NewInMemoryCache().ComputeHash(p.Path(spec))
	if err != nil {
		t.Fatal(err)
	}
	p.WriteFile(".moai/reports/"+signtest.Card+"/plan-audit-1.md",
		fmt.Sprintf("Verdict: PASS\nOverall Score: 0.90\nplan_artifact_hash: %s\n", h))
	p.Git("add", "-A")
	p.Git("commit", "-q", "-m", "report")
}

func writeJudgement(t *testing.T, raw string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "judgement.json")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

const goodJudgement = `{"agent":"lead-session","model":"m","answer":"approve","confidence":0.8,"reason":"Scope matches.","reason_refs":["contract.yaml:2"]}`

func storeCounts(t *testing.T, root string) (int, int) {
	t.Helper()
	s, err := receipt.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := s.Receipts()
	e, _ := s.Events()
	return len(r), len(e)
}

// repoState captures what decide must never change.
func repoState(t *testing.T, p *signtest.Project) string {
	t.Helper()
	db, _ := homestate.BacklogDBPath(p.Root)
	b, _ := os.ReadFile(db)
	return strings.Join([]string{
		p.Git("for-each-ref"), p.Git("worktree", "list", "--porcelain"), string(b),
		string(p.ReadFile(signtest.SpecRel(signtest.SpecID, contract.ContractFile))),
		string(p.ReadFile(signtest.SpecRel(signtest.SpecID, "spec.md"))),
		string(p.ReadFile(signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile))),
	}, "\n---\n")
}

// TestContractDecide checks decide's name, inputs, outputs, exit codes, and
// side effects (AC-GR-020).
func TestContractDecide(t *testing.T) {
	cfgHuman := autoCfg{mode: "contract", decider: "human", pushDevelop: true}
	setup := func(t *testing.T, cfg autoCfg) *signtest.Project {
		p := newContractProject(t, cfg)
		p.Git("add", "-A")
		p.Git("commit", "-q", "-m", "config")
		prepareDecide(t, p)
		p.Git("branch", "feature-x")
		p.Git("update-ref", "refs/remotes/origin/main", "HEAD")
		p.Git("worktree", "add", "-q", filepath.Join(t.TempDir(), "wt"), "feature-x")
		db, _ := homestate.BacklogDBPath(p.Root)
		_ = os.MkdirAll(filepath.Dir(db), 0o755)
		if err := os.WriteFile(db, []byte("backlog bytes"), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	t.Run("help", func(t *testing.T) {
		p := setup(t, cfgContractLLM)
		r := runContract(t, p, contractRun{}, "decide", "--help")
		for _, tok := range []string{"decide", "<card>", "--spec", "--judgement", "--json"} {
			if !strings.Contains(r.stdout, tok) {
				t.Errorf("help lacks %q:\n%s", tok, r.stdout)
			}
		}
	})
	t.Run("normal", func(t *testing.T) {
		p := setup(t, cfgContractLLM)
		before, r0, e0 := repoState(t, p), 0, 0
		r0, e0 = storeCounts(t, p.Root)
		status0 := p.Git("status", "--porcelain")
		res := runContract(t, p, contractRun{}, "decide", signtest.Card, "--spec", signtest.SpecID,
			"--judgement", writeJudgement(t, goodJudgement), "--json")
		if res.code != 0 {
			t.Fatalf("normal: %s", res)
		}
		var out struct{ Outcome string }
		if err := json.Unmarshal([]byte(res.stdout), &out); err != nil || out.Outcome != "approve" {
			t.Errorf("json outcome %q (%v)", out.Outcome, err)
		}
		r1, e1 := storeCounts(t, p.Root)
		if r1 != r0+1 || e1 != e0+1 {
			t.Errorf("store receipts %d→%d events %d→%d, want +1/+1", r0, r1, e0, e1)
		}
		body := p.ReadFile(signtest.ReceiptRel())
		s, _ := receipt.Open(p.Root)
		lines, _ := s.Receipts()
		var rd receipt.ReceiptData
		_ = lines[len(lines)-1].Decode(&rd)
		if rd.Body != string(body) {
			t.Error("kickoff-receipt.json differs from the issued receipt body")
		}
		if _, err := contract.DecodeKickoffReceipt(body); err != nil {
			t.Errorf("receipt carries a field outside the A1 schema: %v", err)
		}
		if after := repoState(t, p); after != before {
			t.Error("decide changed refs, worktrees, the queue database, the contract, or SPEC documents")
		}
		newFiles := strings.TrimSpace(strings.TrimPrefix(p.Git("status", "--porcelain"), status0))
		if newFiles != "?? "+signtest.ReceiptRel() {
			t.Errorf("new working-tree files: %q, want only the receipt", newFiles)
		}
	})
	t.Run("decider-human", func(t *testing.T) {
		p := setup(t, cfgHuman)
		r0, e0 := storeCounts(t, p.Root)
		res := runContract(t, p, contractRun{}, "decide", signtest.Card, "--spec", signtest.SpecID,
			"--judgement", writeJudgement(t, goodJudgement), "--json")
		if res.code != 0 || !strings.Contains(res.stdout, `"outcome": "human"`) {
			t.Fatalf("human: %s", res)
		}
		if r1, e1 := storeCounts(t, p.Root); r1 != r0 || e1 != e0+1 {
			t.Errorf("human path receipts %d→%d events %d→%d", r0, r1, e0, e1)
		}
		if _, err := os.Stat(p.Path(signtest.ReceiptRel())); err == nil {
			t.Error("human path wrote kickoff-receipt.json")
		}
	})
	for _, c := range []struct {
		name      string
		card      string
		judgement string
		tamper    bool
		code      int
	}{
		{"card-mismatch", "t9999", goodJudgement, false, 2},
		{"malformed-judgement", signtest.Card, `{"agent":`, false, 2},
		{"tampered-store", signtest.Card, goodJudgement, true, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := setup(t, cfgContractLLM)
			if c.tamper {
				s, _ := receipt.Open(p.Root)
				for i := 0; i < 3; i++ {
					_, _ = s.AppendEvent(receipt.KindRevoke, receipt.RevokeEvent{Spec: "x"})
				}
				path := filepath.Join(s.Dir, receipt.EventsFile)
				data, _ := os.ReadFile(path)
				if err := os.WriteFile(path, []byte(strings.Replace(string(data), `"spec":"x"`, `"spec":"y"`, 1)), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := s.Verify(); !errors.Is(err, receipt.ErrIntegrity) {
					t.Fatalf("premise: the tampered store must fail verification, got %v", err)
				}
			}
			before := repoState(t, p)
			s, _ := receipt.Open(p.Root)
			raw0, _ := os.ReadFile(filepath.Join(s.Dir, receipt.EventsFile))
			rr0, _ := os.ReadFile(filepath.Join(s.Dir, receipt.ReceiptsFile))
			res := runContract(t, p, contractRun{}, "decide", c.card, "--spec", signtest.SpecID,
				"--judgement", writeJudgement(t, c.judgement))
			if res.code != c.code {
				t.Fatalf("%s: %s, want exit %d", c.name, res, c.code)
			}
			raw1, _ := os.ReadFile(filepath.Join(s.Dir, receipt.EventsFile))
			rr1, _ := os.ReadFile(filepath.Join(s.Dir, receipt.ReceiptsFile))
			if string(raw0) != string(raw1) || string(rr0) != string(rr1) {
				t.Error("store changed on an error exit")
			}
			if repoState(t, p) != before {
				t.Error("repository state changed on an error exit")
			}
			if c.tamper && !strings.Contains(res.stderr, "integrity") {
				t.Errorf("stderr does not name the integrity failure: %s", res.stderr)
			}
			if c.name == "card-mismatch" && !strings.Contains(res.stderr, "card-mismatch") {
				t.Errorf("stderr does not name card-mismatch: %s", res.stderr)
			}
		})
	}
}

// TestContractKickoffCheck checks the gate command's exit codes and JSON, and
// that the sign command wires the compiled doctrine constant: with the
// constant false, an llm+jev approve receipt signed through the CLI is refused
// exactly as the false-injected signer refuses it (D52).
func TestContractKickoffCheck(t *testing.T) {
	t.Run("human-signed", func(t *testing.T) {
		p := newContractProject(t, autoCfg{mode: "contract", decider: "human", pushDevelop: true})
		signFixtureHuman(t, p)
		r := runContract(t, p, contractRun{}, "kickoff-check", signtest.SpecID, "--card", signtest.Card, "--json")
		if r.code != 0 {
			t.Fatalf("human signed: %s", r)
		}
		var out map[string]any
		if err := json.Unmarshal([]byte(r.stdout), &out); err != nil {
			t.Fatalf("json: %v\n%s", err, r.stdout)
		}
		for _, k := range []string{"pass", "reason", "reasons", "mode", "decider", "autonomous_kickoff_enabled", "jev_doctrine_amended", "verify_state"} {
			if _, ok := out[k]; !ok {
				t.Errorf("--json lacks %q", k)
			}
		}
		if out["mode"] != "contract" || out["decider"] != "human" {
			t.Errorf("mode/decider = %v/%v", out["mode"], out["decider"])
		}
	})
	t.Run("unsigned", func(t *testing.T) {
		p := newContractProject(t, autoCfg{mode: "contract", decider: "human", pushDevelop: true})
		r := runContract(t, p, contractRun{}, "kickoff-check", signtest.SpecID, "--card", signtest.Card, "--json")
		if r.code != 1 || !strings.Contains(r.stdout, `"reason": "not-signed-valid"`) {
			t.Fatalf("unsigned: %s", r)
		}
	})
	t.Run("sign-wires-doctrine-constant", func(t *testing.T) {
		p := newContractProject(t, autoCfg{mode: "contract", decider: "llm+jev", pushDevelop: true})
		opts := p.ReceiptOptions("llm+jev", "llm+jev")
		p.WriteReceipt(p.Receipt(opts, func(r *contract.KickoffReceipt) {
			r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm+jev"
			r.JevAnswer = signtest.JevAnswer("approve", 0.71)
		}))
		r := runContract(t, p, contractRun{}, "sign", signtest.SpecID, "--signer", "llm+jev", "--receipt", signtest.ReceiptRel())
		if r.code != 1 || !strings.Contains(r.stdout+r.stderr, contract.RefuseReceiptRequiresHuman) {
			t.Fatalf("CLI sign of an llm+jev approve receipt with the doctrine constant false: %s, want refused %s", r, contract.RefuseReceiptRequiresHuman)
		}
	})
}
