package cli

// mcp_audit_multi_record.go — the A4 second-review record writer
// (SPEC-AUTONOMY-CLOSURE-001 REQ-CLOSURE-012, design.md §A.1). Called from
// runMultiAudit only when card_id was supplied: one append-only JSON line in
// the card evidence directory's second-review.jsonl binding the review to
// the card, the audited commit, the reviewed scope, and the signed contract
// digest. The audit result itself is never altered by this file; a failure
// rides the result's second_review_record_error field.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// appendSecondReviewRecord builds and appends the record (design.md §A.1).
// Unknown card or missing SPEC → spec_id "" (the reader classifies it
// unbound); absent contract → contract_card/digest ""; git scope failure →
// empty scope fields; append failure → the returned error.
func appendSecondReviewRecord(cfg MultiAuditConfig, result ConvergenceResult) error {
	root := cfg.ProjectRoot
	if root == "" {
		root = resolveProjectDir()
	}
	home, err := closure.ResolveEvidenceHome(root, cfg.CardID)
	if err != nil {
		return fmt.Errorf("card evidence home: %w", err)
	}
	autonomy := autonomyConfigForRoot(root)

	rec := closure.SecondReviewRecord{
		SchemaVersion: closure.SchemaVersion,
		Card:          cfg.CardID,
		Target:        "baseBranch",
		Backends:      []closure.SecondReviewBackend{},
		RecordedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	// SPEC ID from the queue store (empty when the card is unknown or
	// SPEC-less); contract facts through A1's verify core.
	specID, _ := cardToSpecID(cfg.CardID)
	rec.SpecID = specID
	if specID != "" {
		dir := filepath.Join(home, ".moai", "specs", specID)
		if inputs, lerr := contract.LoadDir(dir); lerr == nil {
			inputs.Policy = contract.Policy{
				Mode:         autonomy.Mode,
				SecondReview: autonomy.SecondReview,
				PushDevelop:  autonomy.PushDevelop,
			}
			if status, serr := spec.ParseStatus(dir); serr == nil {
				inputs.SpecStatus = status
			}
			rep := contract.Verify(inputs)
			rec.ContractCard = rep.Card
			rec.ContractSHA256 = rep.RecordedContractSHA256
		}
	}

	// The audited commit and the reviewed scope, from the same base
	// resolution the baseBranch review backends use (design.md §A.1).
	if head, herr := gitio.Head(root); herr == nil {
		rec.HeadSHA = head
		rec.Scope.HeadSHA = head
	}
	rec.Scope.BaseBranch, _ = resolveReviewBaseBranchName(root)
	rec.Scope.BaseSHA, _ = resolveReviewMergeBase(root)
	if rec.Scope.BaseSHA != "" && rec.HeadSHA != "" {
		if n, cerr := gitio.ChangedFileCount(root, rec.Scope.BaseSHA, rec.HeadSHA); cerr == nil {
			rec.Scope.ChangedFiles = n
		}
		if diff, derr := gitio.Diff(root, rec.Scope.BaseSHA, rec.HeadSHA); derr == nil {
			sum := sha256.Sum256([]byte(diff))
			rec.Scope.DiffSHA256 = hex.EncodeToString(sum[:])
		}
	}

	for _, v := range result.PerBackendVerdicts {
		verdict := v.Verdict
		if verdict == "" {
			verdict = "inconclusive"
		}
		rec.Backends = append(rec.Backends, closure.SecondReviewBackend{
			Backend: v.Backend, Gate: v.Gate, Verdict: verdict,
		})
	}
	rec.ParticipantCount = result.ParticipantCount
	rec.DisagreementFlag = result.DisagreementFlag
	rec.AuditReceipt = result.AuditReceipt
	rec.BuildCommit = result.BuildCommit

	line, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("encode record: %w", err)
	}
	dir := closure.EvidenceFor(home, cfg.CardID).Dir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("evidence directory: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(dir, closure.SecondReviewFile),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", closure.SecondReviewFile, err)
	}
	_, writeErr := f.Write(append(line, '\n'))
	closeErr := f.Close()
	if writeErr != nil {
		return fmt.Errorf("append %s: %w", closure.SecondReviewFile, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", closure.SecondReviewFile, closeErr)
	}
	return nil
}

// autonomyConfigForRoot is the effective workflow.autonomy configuration of
// the audited tree (kept as a named function so the contract facts above
// carry the tree's own policy, not this process's cwd).
func autonomyConfigForRoot(root string) config.AutonomySettings {
	wf := config.NewDefaultWorkflowConfig()
	if c, err := config.NewLoader().Load(filepath.Join(root, ".moai")); err == nil && c != nil {
		wf = c.Workflow
	}
	return config.ResolveAutonomy(wf)
}
