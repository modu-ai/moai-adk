package sign

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/contract"
)

// KickoffNotice is the REQ-CONTRACT-019 notice printed after every successful
// signature, on both signing paths and in both modes.
const KickoffNotice = "Notice: this signature does not yet replace Implementation Kickoff Approval; the Kickoff gate stays in force in every mode."

// Signer values accepted by Options.Signer.
const (
	SignerHuman   = contract.SignerHuman
	SignerLLM     = contract.DeciderLLM
	SignerLLMJev  = contract.DeciderLLMJev
	ModeContract  = "contract"
	deciderJevRaw = "jev"
)

var (
	// ErrUsage wraps every invocation error (the CLI maps it to exit 2).
	ErrUsage = errors.New("contract sign: usage")
	// ErrInternal wraps a produced contract that failed its own pre-write
	// verification; nothing is written.
	ErrInternal = errors.New("contract sign: internal error")
)

// Options is one signing invocation. The policy fields are the caller's
// effective workflow.autonomy values; this package never reads configuration.
type Options struct {
	// ProjectRoot is the project directory containing .moai/specs/.
	ProjectRoot string
	// SpecIDs are the SPECs to sign; repeats are ignored.
	SpecIDs []string
	// Signer is "" or "human" (human path), or "llm" / "llm+jev" (receipt
	// path; it names the receipt's effective decider).
	Signer string
	// ReceiptPath is the --receipt value (receipt path only). An absolute
	// path is taken relative to ProjectRoot.
	ReceiptPath string
	// Resign re-signs a signed contract (REQ-CONTRACT-022).
	Resign bool

	Mode         string // workflow.autonomy.mode (effective)
	BatchSign    bool   // workflow.autonomy.contract.batch_sign
	SecondReview string // workflow.autonomy.contract.second_review
	PushDevelop  bool   // workflow.autonomy.contract.push_develop
	// Decider is the effective workflow.autonomy.kickoff.decider.
	Decider string
	// DeciderJevSole is true when the configured decider is the rejected
	// "jev" (a configuration error). A Decider of "jev" is treated the same.
	DeciderJevSole bool
	// JevEnabled and JevMinConfidence are carried for the receipt issuer; the
	// A1 receipt validator does not evaluate them.
	JevEnabled       bool
	JevMinConfidence float64
	// BudgetDefault fills an absent contract budget.
	BudgetDefault contract.Budget

	// RegistryRuleIDs and RegistryFrozenFiles are the constitution registry
	// values verify needs.
	RegistryRuleIDs     []string
	RegistryFrozenFiles []string
	// AgentMarkers is the closed agent-environment marker set (variable
	// names) checked on the human path. It must be non-empty there.
	AgentMarkers []string
}

// Seams are the side-effecting operations Sign performs. A nil field takes
// its default.
type Seams struct {
	// IsTTY reports whether standard input is a terminal. Default: stdin is
	// a character device.
	IsTTY func() bool
	// Getenv reads an environment variable. Default: os.Getenv.
	Getenv func(string) string
	// ReadLine reads the one confirmation line. Default: one line of stdin.
	ReadLine func() (string, error)
	// Out receives the summary, prompt, refusals, and the notice. Default:
	// io.Discard.
	Out io.Writer
	// Now is the signing clock. Default: time.Now.
	Now func() time.Time
	// GitIdentity returns git user.name and user.email in root; an unset
	// value is "" (not an error). Default: `git config` subprocesses.
	GitIdentity func(root string) (name, email string, err error)
	// GitHead returns HEAD in root. Default: `git rev-parse HEAD`.
	GitHead func(root string) (string, error)
	// NewBatchID returns a fresh non-empty batch identifier. Default:
	// 16 random bytes, hex.
	NewBatchID func() string
	// WriteFile replaces path atomically. Default: temp file in the same
	// directory, then internal/atomicfile.Replace.
	WriteFile func(path string, data []byte, perm os.FileMode) error
}

// Result reports what Sign did.
type Result struct {
	// Refusal is a sign refusal code (contract.SignRefusalCodes), or "" when
	// every contract was signed.
	Refusal string
	// SpecID is the SPEC a refusal concerns ("" when it concerns the whole
	// invocation).
	SpecID string
	// Cause is the one-line cause printed with the refusal.
	Cause string
	// Reasons are the verify reason codes when Refusal is verify_failed.
	Reasons []string
	// Written and Unwritten are the contract.yaml paths (slash,
	// project-relative) written and not written. On success Written holds
	// every contract; on a write failure they are the exact split.
	Written   []string
	Unwritten []string
}

// refusal is a sign refusal on its way out of the flow.
type refusal struct {
	code, specID, cause string
	reasons             []string
}

func refuse(code, specID, format string, a ...any) *refusal {
	return &refusal{code: code, specID: specID, cause: fmt.Sprintf(format, a...)}
}

// target is one SPEC being signed, carried from validation to the write.
type target struct {
	id       string
	rel      string // slash, project-relative contract path
	path     string // resolved file to write
	perm     os.FileMode
	in       contract.Inputs
	orig     *contract.Contract
	want     *contract.Contract // intended body (signature excluded)
	sha      string             // measured acceptance hash ("" when acceptance.md is absent)
	count    int                // measured live AC count
	body     []byte             // rendered body without signature
	digest   string             // body digest = signature.contract_sha256
	signable string             // Verify(original).SignableContractSHA256
	final    []byte             // body plus signature
}

// Sign signs the contracts named in opts, following design.md § Signing
// Flow in order: path gates, de-duplication, per-SPEC load and signature
// state, acceptance measurement, budget fill, verify on the would-be
// contract, git identity and HEAD, the confirmation (human) or the receipt
// validation (receipt path), then the signature — sealed last, verified
// before any write, written atomically per file. Every refusal leaves every
// file untouched and returns a Result naming the code; the error return is
// reserved for invocation and I/O problems (the CLI maps it to exit 2).
//
// @MX:ANCHOR: [AUTO] Single signing entry point for SPEC contracts.
// @MX:REASON: `moai contract sign` (human and receipt paths, --resign,
// batch) and the configuration tests call it; its refusal order and the
// written signature are what verify, show, and the downstream kickoff and
// escalation layers read.
func Sign(opts Options, seams Seams) (Result, error) {
	s := withDefaults(seams)
	human := opts.Signer == "" || opts.Signer == SignerHuman
	if err := checkUsage(opts, human); err != nil {
		return Result{}, err
	}
	ids := dedupe(opts.SpecIDs)

	if r := pathGates(opts, s, human, ids); r != nil {
		return report(s.Out, r), nil
	}

	targets := make([]*target, 0, len(ids))
	for _, id := range ids {
		t, r, err := prepare(opts, id)
		if err != nil {
			return Result{}, err
		}
		if r != nil {
			return report(s.Out, r), nil
		}
		targets = append(targets, t)
	}

	name, email, err := s.GitIdentity(opts.ProjectRoot)
	if err != nil {
		return Result{}, fmt.Errorf("contract sign: read git identity: %w", err)
	}
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" {
		return report(s.Out, refuse(contract.RefuseGitIdentityMissing, "",
			"git user.name and user.email must both be set (name %q, email %q)", name, email)), nil
	}
	head, err := s.GitHead(opts.ProjectRoot)
	if err != nil {
		return Result{}, fmt.Errorf("contract sign: read HEAD: %w", err)
	}

	sig := contract.Signature{
		Operator: contract.Operator{Name: name, Email: email},
		SignedAt: s.Now().UTC().Format(time.RFC3339),
		HeadSHA:  head,
	}
	if human {
		r, err := confirm(opts, s, targets)
		if err != nil {
			return Result{}, err
		}
		if r != nil {
			return report(s.Out, r), nil
		}
		sig.SignerKind, sig.Method = SignerHuman, contract.MethodInteractiveTTY
		if len(targets) > 1 {
			sig.BatchID = s.NewBatchID()
		}
	} else {
		t := targets[0]
		if r := checkReceipt(opts, t); r != nil {
			return report(s.Out, r), nil
		}
		sig.SignerKind, sig.Method = opts.Signer, contract.MethodReceipt
		sig.Receipt = &contract.Receipt{Path: contract.ReceiptFile, SHA256: sha256Hex(t.in.Receipt),
			Provenance: contract.ProvenanceFile}
	}

	for _, t := range targets {
		if err := seal(t, sig, opts.Resign); err != nil {
			return Result{}, err
		}
	}
	return write(s, targets)
}

// checkUsage rejects an invocation the CLI should not have made.
func checkUsage(opts Options, human bool) error {
	switch {
	case opts.ProjectRoot == "":
		return fmt.Errorf("%w: no project root", ErrUsage)
	case len(opts.SpecIDs) == 0:
		return fmt.Errorf("%w: no SPEC ID", ErrUsage)
	case !human && opts.Signer != SignerLLM && opts.Signer != SignerLLMJev:
		return fmt.Errorf("%w: --signer must be human, llm, or llm+jev (got %q)", ErrUsage, opts.Signer)
	case human && opts.ReceiptPath != "":
		return fmt.Errorf("%w: --receipt requires --signer llm or llm+jev", ErrUsage)
	case !human && opts.ReceiptPath == "":
		return fmt.Errorf("%w: --signer %s requires --receipt", ErrUsage, opts.Signer)
	case human && len(opts.AgentMarkers) == 0:
		return fmt.Errorf("%w: no agent-environment markers supplied for the human path", ErrUsage)
	}
	for _, id := range opts.SpecIDs {
		if !contract.ValidSpecID(id) {
			return fmt.Errorf("%w: %w: %q", ErrUsage, contract.ErrInvalidSpecID, id)
		}
	}
	return nil
}

func dedupe(ids []string) []string {
	var out []string
	for _, id := range ids {
		if !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	return out
}

// pathGates applies step 1 of the signing flow and the batch gate.
func pathGates(opts Options, s Seams, human bool, ids []string) *refusal {
	if !human {
		switch {
		case len(ids) > 1:
			return refuse(contract.RefuseBatchNonHuman, "", "the receipt path signs one SPEC per invocation (%d given)", len(ids))
		case opts.Mode != ModeContract:
			return refuse(contract.RefuseModeNotContract, "", "the receipt path requires workflow.autonomy.mode: contract (mode is %q)", opts.Mode)
		case opts.DeciderJevSole || opts.Decider == deciderJevRaw:
			return refuse(contract.RefuseKickoffDeciderJevSole, "",
				"workflow.autonomy.kickoff.decider is jev, and Jev is never a sole decider (use llm+jev)")
		}
		return nil
	}
	for _, m := range opts.AgentMarkers {
		if s.Getenv(m) != "" {
			return refuse(contract.RefuseAgentMarker, "",
				"agent-environment marker %s is set; the human signature must be given outside an agent session", m)
		}
	}
	if !s.IsTTY() {
		return refuse(contract.RefuseNotTTY, "", "signing on the human path requires an interactive terminal on standard input")
	}
	if len(ids) > 1 && !opts.BatchSign {
		return refuse(contract.RefuseBatchDisabled, "",
			"%d SPECs given but workflow.autonomy.contract.batch_sign is false", len(ids))
	}
	return nil
}

// prepare runs steps 2-5 for one SPEC: load, signature state, acceptance
// measurement, budget fill, render, and verify of the would-be contract.
func prepare(opts Options, id string) (*target, *refusal, error) {
	dir, err := contract.ResolveSpecDir(opts.ProjectRoot, id)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrUsage, err)
	}
	in, err := contract.LoadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	in.Policy = contract.Policy{SecondReview: opts.SecondReview, PushDevelop: opts.PushDevelop,
		Mode: opts.Mode, BudgetDefault: opts.BudgetDefault}
	in.RegistryRuleIDs, in.RegistryFrozenFiles = opts.RegistryRuleIDs, opts.RegistryFrozenFiles

	c, err := contract.Decode(in.Contract)
	if err != nil {
		r := refuse(contract.RefuseVerifyFailed, id, "contract.yaml does not decode: %v", err)
		r.reasons = []string{contract.ReasonSchemaInvalid}
		return nil, r, nil
	}
	signed := c.Signature != nil
	switch {
	case signed && !opts.Resign:
		return nil, refuse(contract.RefuseAlreadySigned, id, "the contract is already signed; use --resign after an acceptance change"), nil
	case !signed && opts.Resign:
		return nil, refuse(contract.RefuseNotSigned, id, "--resign needs a signed contract"), nil
	}

	t := &target{id: id, rel: ".moai/specs/" + id + "/" + contract.ContractFile, in: in, orig: c}
	if in.AcceptancePresent {
		res, err := contract.CountAC(contract.NormalizeAcceptance(in.Acceptance))
		switch {
		case err != nil || res.IsAmbiguous():
			return nil, refuse(contract.RefuseACCountAmbiguous, id, "acceptance.md has no usable AC count (ambiguous IDs %v)", res.Ambiguous), nil
		case res.Live == 0:
			return nil, refuse(contract.RefuseACCountZero, id, "acceptance.md carries no live AC ID"), nil
		}
		t.sha, t.count = contract.AcceptanceHash(in.Acceptance), res.Live
		if a := c.Acceptance; !signed && a != nil && a.SHA256 != nil && *a.SHA256 != t.sha {
			return nil, refuse(contract.RefuseDraftAcceptanceStale, id,
				"the draft records acceptance hash %s but acceptance.md measures %s", short(*a.SHA256), short(t.sha)), nil
		}
	}

	if opts.Resign {
		receiptPath := opts.Signer != "" && opts.Signer != SignerHuman
		if r := resignTamper(in, id, receiptPath); r != nil {
			return nil, r, nil
		}
	}

	want := contract.Canonical(*c)
	want.Signature = nil
	edit := bodyEdit{sha: t.sha, count: t.count, bind: in.AcceptancePresent}
	if want.Acceptance != nil && edit.bind {
		sha, count := t.sha, t.count
		want.Acceptance.SHA256, want.Acceptance.ACCount = &sha, &count
	}
	if !c.HasSection(contract.SectionBudget) {
		b := opts.BudgetDefault
		want.Budget, edit.budget = &b, &b
	}
	t.want = &want

	if t.body, err = renderBody(in.Contract, edit, &want); err != nil {
		return nil, nil, err
	}
	wb := in
	wb.Contract = t.body
	rep := contract.Verify(wb)
	var reasons []string
	for _, r := range rep.Reasons {
		if r != contract.ReasonUnsigned {
			reasons = append(reasons, r)
		}
	}
	switch {
	case slices.ContainsFunc(reasons, func(r string) bool { return r != contract.ReasonPlanAuditNotPassing }):
		r := refuse(contract.RefuseVerifyFailed, id, "the contract fails verify: %s", strings.Join(reasons, ", "))
		r.reasons = reasons
		return nil, r, nil
	case len(reasons) > 0:
		return nil, refuse(contract.RefusePlanAuditNotPassing, id,
			"plan_audit.verdict must be PASS or PASS-WITH-DEBT (got %q)", verdict(c)), nil
	}
	t.digest = rep.ContractSHA256
	t.signable = contract.Verify(in).SignableContractSHA256

	if t.path, err = filepath.EvalSymlinks(filepath.Join(dir, contract.ContractFile)); err != nil {
		return nil, nil, fmt.Errorf("contract sign: resolve %s: %w", t.rel, err)
	}
	info, err := os.Stat(t.path)
	if err != nil {
		return nil, nil, fmt.Errorf("contract sign: stat %s: %w", t.rel, err)
	}
	t.perm = info.Mode().Perm()
	return t, nil, nil
}

// resignReplaceable is the set of verify reasons a re-sign may replace: the
// acceptance binding it re-measures (REQ-CONTRACT-012, REQ-CONTRACT-022).
var resignReplaceable = []string{
	contract.ReasonUnsigned,
	contract.ReasonAcceptanceHashMismatch,
	contract.ReasonACCountMismatch,
	contract.ReasonSignatureAcceptanceMismatch,
}

// resignTamper verifies the signed contract as it stands on disk. Any reason
// outside resignReplaceable — a body edited after signing, a broken seal, an
// inconsistent signature — refuses verify_failed, so --resign cannot launder
// a detected change into a fresh signature. On the receipt path the new
// receipt already sits at the fixed receipt path, so receipt_mismatch against
// the old signature is expected there; the new receipt is validated against
// the signable digest before anything is written (REQ-CONTRACT-023).
func resignTamper(in contract.Inputs, id string, receiptPath bool) *refusal {
	var reasons []string
	for _, r := range contract.Verify(in).Reasons {
		if receiptPath && r == contract.ReasonReceiptMismatch {
			continue
		}
		if !slices.Contains(resignReplaceable, r) {
			reasons = append(reasons, r)
		}
	}
	if len(reasons) == 0 {
		return nil
	}
	r := refuse(contract.RefuseVerifyFailed, id,
		"the signed contract fails verify beyond the acceptance binding: %s", strings.Join(reasons, ", "))
	r.reasons = reasons
	return r
}

// confirm prints the human-path summary and reads the one confirmation line.
func confirm(opts Options, s Seams, targets []*target) (*refusal, error) {
	out := s.Out
	say(out, "Contract signing summary\n")
	for _, t := range targets {
		printSummary(out, t, opts.Resign)
	}
	token := targets[0].id
	if len(targets) > 1 {
		token = fmt.Sprintf("sign %d contracts", len(targets))
	}
	say(out, "Type %q to sign: ", token)
	line, err := s.ReadLine()
	say(out, "\n")
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("contract sign: read confirmation: %w", err)
	}
	if strings.TrimSpace(line) != token {
		return refuse(contract.RefuseConfirmationMismatch, "", "the typed confirmation did not match %q", token), nil
	}
	return nil, nil
}

func printSummary(out io.Writer, t *target, resign bool) {
	say(out, "%s\n", t.id)
	if resign {
		oldSHA, oldCount := "none", "none"
		if a := t.orig.Acceptance; a != nil {
			if a.SHA256 != nil {
				oldSHA = short(*a.SHA256)
			}
			if a.ACCount != nil {
				oldCount = strconv.Itoa(*a.ACCount)
			}
		}
		say(out, "  acceptance: %s → %s (%s → %d ACs)\n", oldSHA, short(t.sha), oldCount, t.count)
	} else {
		say(out, "  acceptance: %s (%d ACs)\n", short(t.sha), t.count)
	}
	say(out, "  actions: %s\n", strings.Join(t.want.Actions, ", "))
	if b := t.want.Budget; b != nil {
		say(out, "  budget: turns %d, operations %d, audit_retries %d\n", b.Turns, b.Operations, b.AuditRetries)
	}
	if rv := t.want.Review; rv != nil {
		say(out, "  second model: %s\n", rv.SecondModel)
	}
	carried := ""
	if resign {
		carried = " (carried)"
	}
	say(out, "  plan-audit verdict: %s%s\n", verdict(t.orig), carried)
}

// checkReceipt validates the receipt (REQ-CONTRACT-023) and applies the
// signer steps that follow an accepted receipt (REQ-CONTRACT-024).
func checkReceipt(opts Options, t *target) *refusal {
	rel := opts.ReceiptPath
	if filepath.IsAbs(rel) {
		if r, err := filepath.Rel(opts.ProjectRoot, rel); err == nil {
			rel = r
		}
	}
	chk := contract.ReceiptCheck{
		Raw:                    t.in.Receipt,
		Path:                   filepath.ToSlash(rel),
		SpecID:                 t.id,
		ContractLines:          contract.ContractLineCount(t.in.Contract),
		EffectiveDecider:       opts.Decider,
		Signer:                 opts.Signer,
		SignableContractSHA256: t.signable,
		AcceptanceSHA256:       t.sha,
		ReadRepoFile: func(p string) ([]byte, error) {
			return os.ReadFile(filepath.Join(opts.ProjectRoot, filepath.FromSlash(p)))
		},
	}
	r, code := contract.ValidateKickoffReceipt(chk)
	if code != "" {
		return refuse(code, t.id, "the kickoff receipt %s fails validation", chk.Path)
	}
	if code, ok := contract.ReceiptOutcome(r); !ok {
		return refuse(code, t.id, "the kickoff receipt records outcome %q from effective decider %s", r.Outcome, r.EffectiveDecider)
	}
	return nil
}

// seal completes one target's signature, seals it last, and verifies the
// produced file before anything is written.
func seal(t *target, base contract.Signature, resign bool) error {
	sig := base
	sig.ContractSHA256, sig.AcceptanceSHA256 = t.digest, t.sha
	if resign {
		sig.Supersedes = t.orig.Signature.ContractSHA256
	}
	var err error
	if sig.Seal, err = contract.ComputeSeal(sig); err != nil {
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}
	if t.final, err = appendSignature(t.body, sig); err != nil {
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}
	fin := t.in
	fin.Contract = t.final
	if rep := contract.Verify(fin); rep.State != contract.StateSignedValid {
		return fmt.Errorf("%w: the produced contract for %s verifies %s %v; nothing was written",
			ErrInternal, t.id, rep.State, rep.Reasons)
	}
	return nil
}

// write writes every target atomically, reporting the exact split on failure.
func write(s Seams, targets []*target) (Result, error) {
	var res Result
	for i, t := range targets {
		if err := s.WriteFile(t.path, t.final, t.perm); err != nil {
			for _, rest := range targets[i:] {
				res.Unwritten = append(res.Unwritten, rest.rel)
			}
			say(s.Out, "write failed: %s: %v\n", t.rel, err)
			say(s.Out, "written: %s\n", listOrNone(res.Written))
			say(s.Out, "not written: %s\n", listOrNone(res.Unwritten))
			return res, fmt.Errorf("contract sign: write %s: %w", t.rel, err)
		}
		res.Written = append(res.Written, t.rel)
		say(s.Out, "signed %s (contract_sha256 %s)\n", t.id, short(t.digest))
	}
	say(s.Out, "%s\n", KickoffNotice)
	return res, nil
}

// report prints a refusal and converts it to a Result.
func report(out io.Writer, r *refusal) Result {
	where := ""
	if r.specID != "" {
		where = " (" + r.specID + ")"
	}
	say(out, "refused %s%s: %s\n", r.code, where, r.cause)
	if len(r.reasons) > 0 {
		say(out, "reasons: %s\n", strings.Join(r.reasons, ", "))
	}
	return Result{Refusal: r.code, SpecID: r.specID, Cause: r.cause, Reasons: r.reasons}
}

func verdict(c *contract.Contract) string {
	if c.PlanAudit == nil {
		return ""
	}
	return c.PlanAudit.Verdict
}

// short is the 12-character display prefix of a hash.
func short(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}

// say writes one formatted piece of operator output. A failed write to the
// operator's terminal cannot be reported anywhere better, so it is dropped.
func say(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}

func listOrNone(l []string) string {
	if len(l) == 0 {
		return "none"
	}
	return strings.Join(l, ", ")
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
