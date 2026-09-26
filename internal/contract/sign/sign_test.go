package sign_test

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

var fixedNow = time.Date(2026, 9, 26, 18, 30, 5, 0, time.FixedZone("KST", 9*3600))

// --- AC-014 signer half: signing-path gates ---------------------------------

func TestSign_HumanPathRefusesWithoutTerminal(t *testing.T) {
	p := signtest.New(t)
	seams, rec := signtest.Seams(false, noMarkers, signtest.SpecID)
	assertRefuses(t, p, p.Options(), seams, rec, contract.RefuseNotTTY)
	if !strings.Contains(rec.Out.String(), "interactive terminal") {
		t.Errorf("output does not state that signing requires an interactive terminal:\n%s", rec.Out.String())
	}
	if rec.ReadLines != 0 {
		t.Errorf("read %d lines, want 0", rec.ReadLines)
	}
}

func TestSign_ReceiptPathRequiresContractMode(t *testing.T) {
	p := signtest.New(t)
	opts := p.ReceiptOptions("llm", "llm")
	opts.Mode = "guided"
	p.WriteReceipt(p.Receipt(opts, nil))
	seams, rec := signtest.Seams(true, noMarkers)
	assertRefuses(t, p, opts, seams, rec, contract.RefuseModeNotContract)
}

func TestSign_ReceiptPathNeedsNoTerminal(t *testing.T) {
	p := signtest.New(t)
	opts := p.ReceiptOptions("llm", "llm")
	p.WriteReceipt(p.Receipt(opts, nil))
	seams, rec := signtest.Seams(false, noMarkers)
	mustSign(t, opts, seams, rec)
	if rec.ReadLines != 0 || rec.TTYCalls != 0 {
		t.Errorf("receipt path used the terminal: %d reads, %d tty checks", rec.ReadLines, rec.TTYCalls)
	}
	assertSignedValid(t, p, signtest.SpecID, opts)
}

// --- AC-015 signer half: signature record on both paths ---------------------

func TestSign_HumanPathRecord(t *testing.T) {
	p := signtest.New(t)
	opts := p.Options()

	seams, rec := humanSeams("SPEC-WRONG-001")
	assertRefuses(t, p, opts, seams, rec, contract.RefuseConfirmationMismatch)
	if rec.ReadLines != 1 {
		t.Errorf("confirmation read %d times, want 1", rec.ReadLines)
	}

	seams, rec = humanSeams(signtest.SpecID)
	seams.Now = func() time.Time { return fixedNow }
	res := mustSign(t, opts, seams, rec)
	if !strings.Contains(rec.Out.String(), signtest.SpecID) {
		t.Errorf("prompt does not display the token %s:\n%s", signtest.SpecID, rec.Out.String())
	}
	if want := []string{signtest.SpecRel(signtest.SpecID, contract.ContractFile)}; strings.Join(res.Written, ",") != strings.Join(want, ",") {
		t.Errorf("Written = %v, want %v", res.Written, want)
	}

	c := decodeSigned(t, p, signtest.SpecID)
	measuredSHA := contract.AcceptanceHash([]byte(signtest.Acceptance))
	if c.Acceptance.SHA256 == nil || *c.Acceptance.SHA256 != measuredSHA {
		t.Errorf("acceptance.sha256 = %v, want %s", c.Acceptance.SHA256, measuredSHA)
	}
	if c.Acceptance.ACCount == nil || *c.Acceptance.ACCount != 2 {
		t.Errorf("acceptance.ac_count = %v, want 2", c.Acceptance.ACCount)
	}
	if c.Budget == nil || *c.Budget != opts.BudgetDefault {
		t.Errorf("budget = %v, want %+v", c.Budget, opts.BudgetDefault)
	}
	s := c.Signature
	if s == nil {
		t.Fatal("no signature block written")
	}
	digest, err := contract.Digest(c)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	checks := []struct{ name, got, want string }{
		{"signer_kind", s.SignerKind, "human"},
		{"operator.name", s.Operator.Name, signtest.OperatorName},
		{"operator.email", s.Operator.Email, signtest.OperatorEmail},
		{"signed_at", s.SignedAt, "2026-09-26T09:30:05Z"},
		{"head_sha", s.HeadSHA, p.Head},
		{"method", s.Method, contract.MethodInteractiveTTY},
		{"contract_sha256", s.ContractSHA256, digest},
		{"acceptance_sha256", s.AcceptanceSHA256, measuredSHA},
	}
	for _, ck := range checks {
		if ck.got != ck.want {
			t.Errorf("signature.%s = %q, want %q", ck.name, ck.got, ck.want)
		}
	}
	if s.Receipt != nil || s.BatchID != "" || s.Supersedes != "" {
		t.Errorf("unexpected optional fields: receipt %v batch %q supersedes %q", s.Receipt, s.BatchID, s.Supersedes)
	}
	if seal, _ := contract.ComputeSeal(*s); seal == "" || seal != s.Seal {
		t.Errorf("seal = %q, want recomputation %q", s.Seal, seal)
	}
	if !strings.Contains(rec.Out.String(), sign.KickoffNotice) {
		t.Errorf("notice missing:\n%s", rec.Out.String())
	}
	assertSignedValid(t, p, signtest.SpecID, opts)
}

func TestSign_ReceiptPathRecord(t *testing.T) {
	cases := []struct {
		name    string
		decider string
		jev     bool
		mutate  func(*contract.KickoffReceipt)
	}{
		{"r1_configured_llm", "llm", true, nil},
		{"r2_derived_llm", "llm", true, nil},
		{"r3_recorded_fallback", "llm+jev", false, func(r *contract.KickoffReceipt) {
			reason := "jev_disabled"
			r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm"
			r.Fallback = &contract.ReceiptFallback{Applied: true, Reason: &reason}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := signtest.New(t)
			opts := p.ReceiptOptions(tc.decider, "llm")
			opts.JevEnabled = tc.jev
			receipt := p.Receipt(opts, tc.mutate)
			p.WriteReceipt(receipt)
			seams, rec := signtest.Seams(false, noMarkers)
			mustSign(t, opts, seams, rec)
			if rec.ReadLines != 0 {
				t.Errorf("confirmation reader called %d times", rec.ReadLines)
			}
			s := decodeSigned(t, p, signtest.SpecID).Signature
			if s == nil {
				t.Fatal("no signature")
			}
			if s.Method != contract.MethodReceipt || s.SignerKind != "llm" || s.Receipt == nil ||
				s.Receipt.Path != contract.ReceiptFile || s.Receipt.Provenance != contract.ProvenanceFile ||
				s.Receipt.SHA256 != signtest.SHA256Hex(receipt) {
				t.Errorf("receipt signature = %+v (receipt %+v)", s, s.Receipt)
			}
			if !strings.Contains(rec.Out.String(), sign.KickoffNotice) {
				t.Errorf("notice missing:\n%s", rec.Out.String())
			}
			assertSignedValid(t, p, signtest.SpecID, opts)

			if tc.name == "r1_configured_llm" {
				changed := append([]byte(nil), receipt...)
				changed[len(changed)-2] ^= 0x01
				p.WriteReceipt(changed)
				rep := p.Verify(signtest.SpecID, opts)
				if !strings.Contains(strings.Join(rep.Reasons, ","), contract.ReasonReceiptMismatch) {
					t.Errorf("after a receipt byte change reasons = %v, want receipt_mismatch", rep.Reasons)
				}
			}
		})
	}
}

// --- AC-017 signer half: batch signing ----------------------------------------

func batchProject(t *testing.T) *signtest.Project {
	t.Helper()
	p := signtest.New(t)
	p.AddSpec("SPEC-FIXTURE-002", signtest.Draft{}, signtest.Acceptance)
	p.AddSpec("SPEC-FIXTURE-003", signtest.Draft{}, signtest.Acceptance)
	return p
}

func TestSign_BatchOneConfirmation(t *testing.T) {
	p := batchProject(t)
	ids := []string{"SPEC-FIXTURE-001", "SPEC-FIXTURE-002", "SPEC-FIXTURE-001", "SPEC-FIXTURE-003"}
	opts := p.Options(ids...)
	opts.BatchSign = true
	seams, rec := humanSeams("sign 3 contracts")
	res := mustSign(t, opts, seams, rec)
	if rec.ReadLines != 1 {
		t.Errorf("confirmation read %d times, want exactly 1", rec.ReadLines)
	}
	if len(res.Written) != 3 {
		t.Errorf("written = %v, want 3 files", res.Written)
	}
	batch := ""
	for _, id := range []string{"SPEC-FIXTURE-001", "SPEC-FIXTURE-002", "SPEC-FIXTURE-003"} {
		s := decodeSigned(t, p, id).Signature
		if s == nil || s.BatchID == "" {
			t.Fatalf("%s: missing signature or batch_id", id)
		}
		if batch == "" {
			batch = s.BatchID
		} else if s.BatchID != batch {
			t.Errorf("%s: batch_id %q, want shared %q", id, s.BatchID, batch)
		}
		assertSignedValid(t, p, id, opts)
	}
}

func TestSign_BatchOneInvalidWritesNothing(t *testing.T) {
	p := batchProject(t)
	p.AddSpec("SPEC-FIXTURE-002", signtest.Draft{Verdict: "FAIL"}, signtest.Acceptance)
	opts := p.Options("SPEC-FIXTURE-001", "SPEC-FIXTURE-002", "SPEC-FIXTURE-003")
	opts.BatchSign = true
	seams, rec := humanSeams("sign 3 contracts")
	res := assertRefuses(t, p, opts, seams, rec, contract.RefusePlanAuditNotPassing)
	if res.SpecID != "SPEC-FIXTURE-002" {
		t.Errorf("refusal SpecID = %q, want SPEC-FIXTURE-002", res.SpecID)
	}
	if rec.ReadLines != 0 {
		t.Errorf("prompted before validating every SPEC (%d reads)", rec.ReadLines)
	}
}

func TestSign_BatchDisabled(t *testing.T) {
	p := batchProject(t)
	seams, rec := humanSeams("sign 2 contracts")
	assertRefuses(t, p, p.Options("SPEC-FIXTURE-001", "SPEC-FIXTURE-002"), seams, rec, contract.RefuseBatchDisabled)
	if rec.ReadLines != 0 {
		t.Errorf("batch_disabled prompted (%d reads)", rec.ReadLines)
	}
}

func TestSign_BatchNonHuman(t *testing.T) {
	p := batchProject(t)
	opts := p.ReceiptOptions("llm", "llm")
	opts.SpecIDs = []string{"SPEC-FIXTURE-001", "SPEC-FIXTURE-002"}
	opts.BatchSign = true
	seams, rec := signtest.Seams(false, noMarkers)
	assertRefuses(t, p, opts, seams, rec, contract.RefuseBatchNonHuman)
}

func TestSign_BatchWriteFailureReportsSplit(t *testing.T) {
	p := batchProject(t)
	opts := p.Options("SPEC-FIXTURE-001", "SPEC-FIXTURE-002", "SPEC-FIXTURE-003")
	opts.BatchSign = true
	seams, rec := humanSeams("sign 3 contracts")
	before := p.Snapshot()
	calls := 0
	seams.WriteFile = func(path string, data []byte, perm os.FileMode) error {
		calls++
		if calls == 2 {
			return errors.New("disk full")
		}
		return os.WriteFile(path, data, perm)
	}
	res, err := sign.Sign(opts, seams)
	if err == nil {
		t.Fatalf("want a write error, got result %+v\n%s", res, rec.Out.String())
	}
	if len(res.Written) != 1 || len(res.Unwritten) != 2 {
		t.Fatalf("split = written %v / unwritten %v, want 1 / 2", res.Written, res.Unwritten)
	}
	for _, rel := range res.Unwritten {
		if string(p.ReadFile(rel)) != before[rel] {
			t.Errorf("%s changed although reported unwritten", rel)
		}
	}
	if string(p.ReadFile(res.Written[0])) == before[res.Written[0]] {
		t.Errorf("%s reported written but unchanged", res.Written[0])
	}
	out := rec.Out.String()
	if !strings.Contains(out, res.Unwritten[0]) || !strings.Contains(out, res.Written[0]) {
		t.Errorf("output does not report the split:\n%s", out)
	}
}

// --- AC-022 signer half: notice on both paths and modes ------------------------

func TestSign_KickoffNoticeEveryPathAndMode(t *testing.T) {
	for _, mode := range []string{"guided", "contract"} {
		t.Run("human_"+mode, func(t *testing.T) {
			p := signtest.New(t)
			opts := p.Options()
			opts.Mode = mode
			seams, rec := humanSeams(signtest.SpecID)
			mustSign(t, opts, seams, rec)
			if !strings.Contains(rec.Out.String(), sign.KickoffNotice) {
				t.Errorf("notice missing:\n%s", rec.Out.String())
			}
		})
	}
	t.Run("receipt_contract", func(t *testing.T) {
		p := signtest.New(t)
		opts := p.ReceiptOptions("llm", "llm")
		p.WriteReceipt(p.Receipt(opts, nil))
		seams, rec := signtest.Seams(false, noMarkers)
		mustSign(t, opts, seams, rec)
		if !strings.Contains(rec.Out.String(), sign.KickoffNotice) {
			t.Errorf("notice missing:\n%s", rec.Out.String())
		}
	})
	if !strings.Contains(sign.KickoffNotice, "does not yet replace Implementation Kickoff Approval") {
		t.Errorf("notice text %q does not state the REQ-CONTRACT-019 claim", sign.KickoffNotice)
	}
}

// --- AC-024 signer half: re-sign after an acceptance change --------------------

func TestSign_Resign(t *testing.T) {
	p := signtest.New(t)
	opts := p.Options()
	seams, rec := humanSeams(signtest.SpecID)
	mustSign(t, opts, seams, rec)
	d0 := decodeSigned(t, p, signtest.SpecID).Signature.ContractSHA256
	oldSHA := contract.AcceptanceHash([]byte(signtest.Acceptance))
	newSHA := contract.AcceptanceHash([]byte(signtest.AcceptanceThree))
	p.WriteFile(signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile), signtest.AcceptanceThree)

	seams, rec = humanSeams(signtest.SpecID)
	assertRefuses(t, p, opts, seams, rec, contract.RefuseAlreadySigned)

	resign := opts
	resign.Resign = true
	seams, rec = signtest.Seams(false, noMarkers, signtest.SpecID)
	assertRefuses(t, p, resign, seams, rec, contract.RefuseNotTTY)

	seams, rec = humanSeams(signtest.SpecID)
	mustSign(t, resign, seams, rec)
	out := rec.Out.String()
	for _, want := range []string{oldSHA[:12] + " → " + newSHA[:12], "2 → 3", "PASS (carried)"} {
		if !strings.Contains(out, want) {
			t.Errorf("re-sign summary lacks %q:\n%s", want, out)
		}
	}
	c := decodeSigned(t, p, signtest.SpecID)
	if *c.Acceptance.SHA256 != newSHA || *c.Acceptance.ACCount != 3 {
		t.Errorf("acceptance binding = %s / %d, want %s / 3", *c.Acceptance.SHA256, *c.Acceptance.ACCount, newSHA)
	}
	if c.Signature.Supersedes != d0 {
		t.Errorf("supersedes = %q, want D0 %q", c.Signature.Supersedes, d0)
	}
	assertSignedValid(t, p, signtest.SpecID, opts)
}

func TestSign_ResignUnsignedRefuses(t *testing.T) {
	p := signtest.New(t)
	opts := p.Options()
	opts.Resign = true
	seams, rec := humanSeams(signtest.SpecID)
	assertRefuses(t, p, opts, seams, rec, contract.RefuseNotSigned)
}

// --- AC-025 signer half: agent-environment markers -------------------------------

func TestSign_AgentMarkers(t *testing.T) {
	for _, env := range []map[string]string{{"CLAUDECODE": "1"}, {"CLAUDE_CODE_SESSION_ID": "abc"}} {
		for name := range env {
			t.Run(name, func(t *testing.T) {
				p := signtest.New(t)
				seams, rec := signtest.Seams(true, env, signtest.SpecID)
				assertRefuses(t, p, p.Options(), seams, rec, contract.RefuseAgentMarker)
				if !strings.Contains(rec.Out.String(), name) {
					t.Errorf("output does not name %s:\n%s", name, rec.Out.String())
				}
				if rec.ReadLines != 0 {
					t.Errorf("prompted despite a marker (%d reads)", rec.ReadLines)
				}
			})
		}
	}
	t.Run("empty_markers_proceed", func(t *testing.T) {
		p := signtest.New(t)
		seams, rec := signtest.Seams(true, map[string]string{"CLAUDECODE": "", "CLAUDE_CODE_SESSION_ID": ""}, signtest.SpecID)
		mustSign(t, p.Options(), seams, rec)
	})
	t.Run("receipt_path_ignores_markers", func(t *testing.T) {
		p := signtest.New(t)
		opts := p.ReceiptOptions("llm", "llm")
		p.WriteReceipt(p.Receipt(opts, nil))
		seams, rec := signtest.Seams(false, map[string]string{"CLAUDECODE": "1"})
		mustSign(t, opts, seams, rec)
	})
}

// --- layout preservation ---------------------------------------------------------

func TestSign_PreservesAuthorLayout(t *testing.T) {
	p := signtest.New(t)
	rel := signtest.SpecRel(signtest.SpecID, contract.ContractFile)
	before := strings.Split(strings.TrimSuffix(string(p.ReadFile(rel)), "\n"), "\n")
	seams, rec := humanSeams(signtest.SpecID)
	mustSign(t, p.Options(), seams, rec)
	after := strings.Split(string(p.ReadFile(rel)), "\n")
	// Every original line — comments and blank lines included — survives, in order.
	i := 0
	for _, line := range after {
		if i < len(before) && line == before[i] {
			i++
		}
	}
	if i != len(before) {
		t.Errorf("original line %d %q not preserved in order; signed file:\n%s", i+1, before[i], strings.Join(after, "\n"))
	}
}

func TestSign_FlowStyleAcceptanceStillSigns(t *testing.T) {
	p := signtest.New(t)
	rel := signtest.SpecRel(signtest.SpecID, contract.ContractFile)
	text := strings.Replace(string(p.ReadFile(rel)),
		"acceptance:\n  file: acceptance.md   # the SPEC's acceptance criteria\n",
		"acceptance: {file: acceptance.md}\n", 1)
	p.WriteFile(rel, text)
	seams, rec := humanSeams(signtest.SpecID)
	mustSign(t, p.Options(), seams, rec)
	assertSignedValid(t, p, signtest.SpecID, p.Options())
	if !strings.Contains(string(p.ReadFile(rel)), "# Invariants the run must keep.") {
		t.Error("comment lost on the fallback path")
	}
}

// --- invocation errors -------------------------------------------------------------

func TestSign_UsageAndIOErrors(t *testing.T) {
	p := signtest.New(t)
	cases := map[string]func(o *sign.Options){
		"no_ids":            func(o *sign.Options) { o.SpecIDs = nil },
		"bad_signer":        func(o *sign.Options) { o.Signer = "robot" },
		"invalid_spec_id":   func(o *sign.Options) { o.SpecIDs = []string{"../etc"} },
		"missing_contract":  func(o *sign.Options) { o.SpecIDs = []string{"SPEC-ABSENT-001"} },
		"no_project_root":   func(o *sign.Options) { o.ProjectRoot = "" },
		"no_markers_human":  func(o *sign.Options) { o.AgentMarkers = nil },
		"receipt_no_path":   func(o *sign.Options) { o.Signer, o.Mode, o.Decider = "llm", "contract", "llm" },
		"human_with_receipt": func(o *sign.Options) { o.ReceiptPath = signtest.ReceiptRel() },
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			opts := p.Options()
			mut(&opts)
			seams, rec := humanSeams(signtest.SpecID)
			snap := p.Snapshot()
			res, err := sign.Sign(opts, seams)
			if err == nil {
				t.Fatalf("want an error, got %+v\n%s", res, rec.Out.String())
			}
			if res.Refusal != "" {
				t.Errorf("error path also returned refusal %q", res.Refusal)
			}
			p.AssertUnchanged(t, snap)
		})
	}
}

func TestSign_DefaultGitSeamsReadFixtureRepository(t *testing.T) {
	p := signtest.New(t)
	seams, rec := humanSeams(signtest.SpecID)
	seams.GitIdentity, seams.GitHead = nil, nil
	mustSign(t, p.Options(), seams, rec)
	s := decodeSigned(t, p, signtest.SpecID).Signature
	if s.HeadSHA != p.Head || s.Operator.Email != signtest.OperatorEmail {
		t.Errorf("default git seams recorded head %q email %q, want %q %q", s.HeadSHA, s.Operator.Email, p.Head, signtest.OperatorEmail)
	}
}
