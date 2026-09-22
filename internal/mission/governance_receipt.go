package mission

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

type GovernanceReceiptKind string
type GovernanceReceiptStatus string

const (
	GovernanceDecision GovernanceReceiptKind = "decision"
	GovernanceAudit    GovernanceReceiptKind = "audit"

	GovernanceRecommended GovernanceReceiptStatus = "recommended"
	GovernancePassed      GovernanceReceiptStatus = "passed"
	GovernanceFailed      GovernanceReceiptStatus = "failed"
)

// AuxiliarySignal is ONE model-produced advisory signal the governor was
// shown, recorded as a separate item alongside the binding fields so a reader
// of the receipt can see what evidence accompanied the decision. It is
// deliberately plain-typed (no import of the judgment package): the receipt
// records what was displayed, and a display-only signal never becomes a
// completion-predicate element or a piece of the landed-ancestry or
// authoritative-readback evidence the binding fields carry.
//
// @MX:NOTE: [AUTO] deliberately absent from validateGovernanceBinding — the binding fields decide, the auxiliary item only records; the integrity digest still covers it so a receipt cannot gain or lose a recorded signal silently.
// @MX:SPEC: SPEC-JEV-GOAL-DIST-001
type AuxiliarySignal struct {
	QuestionID  string  `json:"question_id"`
	Kind        string  `json:"kind"`
	Noul        bool    `json:"noul,omitempty"`
	Probability float64 `json:"probability"`
}

// GovernanceReceipt is untrusted until LoadGovernanceReceipts validates its
// content digest and binds every execution-relevant field to the current
// contract and snapshot. Advisor prose is intentionally not part of it.
//
// AuxiliarySignals is deliberately absent from validateGovernanceBinding: the
// binding fields decide, the auxiliary item only records. The integrity digest
// still covers it (canonicalGovernanceReceipt keeps it), so a receipt cannot
// gain or lose a recorded signal without breaking its integrity.
type GovernanceReceipt struct {
	Version          int                     `json:"version"`
	Kind             GovernanceReceiptKind   `json:"kind"`
	MissionID        string                  `json:"mission_id"`
	ContractHash     string                  `json:"contract_hash"`
	SnapshotHash     string                  `json:"snapshot_hash"`
	Action           Action                  `json:"action"`
	Targets          []string                `json:"targets"`
	ExpiresAt        time.Time               `json:"expires_at"`
	Issuer           string                  `json:"issuer"`
	HeadSHA          string                  `json:"head_sha"`
	Status           GovernanceReceiptStatus `json:"status"`
	AuxiliarySignals []AuxiliarySignal       `json:"auxiliary_signals,omitempty"`
	Integrity        string                  `json:"integrity"`
}

type GovernanceExpectation struct {
	MissionID, ContractHash, SnapshotHash, HeadSHA string
	Action                                         Action
	Targets                                        []string
	Now                                            time.Time
}

type GovernanceEvidence struct {
	Decision GovernanceReceipt
	Audit    GovernanceReceipt
}

func canonicalGovernanceReceipt(r GovernanceReceipt) GovernanceReceipt {
	r.Integrity = ""
	r.Targets = append([]string(nil), r.Targets...)
	sort.Strings(r.Targets)
	// Deep-copy the auxiliary items too: the digest must not depend on the
	// caller mutating a slice after writing, exactly as for Targets.
	r.AuxiliarySignals = append([]AuxiliarySignal(nil), r.AuxiliarySignals...)
	return r
}

func governanceReceiptDigest(r GovernanceReceipt) (string, error) {
	raw, err := json.Marshal(canonicalGovernanceReceipt(r))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func marshalGovernanceReceipt(r GovernanceReceipt) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func ParseGovernanceReceipt(raw []byte) (GovernanceReceipt, error) {
	var r GovernanceReceipt
	if err := json.Unmarshal(raw, &r); err != nil {
		return GovernanceReceipt{}, errors.New("mission governance: receipt_parse")
	}
	return r, nil
}

func governanceDir(root string, create bool) (string, error) {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil || !filepath.IsAbs(resolvedRoot) {
		return "", errors.New("mission governance: invalid_root")
	}
	dir := filepath.Join(resolvedRoot, ".moai", "state", "mission", "governance")
	if create {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", err
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return "", err
		}
	}
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(resolvedRoot, resolvedDir)
	if err != nil || rel == ".." || strings.HasPrefix(filepath.ToSlash(rel), "../") {
		return "", errors.New("mission governance: receipt_outside")
	}
	return resolvedDir, nil
}

func containedGovernancePath(root, path string, create bool) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("mission governance: receipt_outside")
	}
	dir, err := governanceDir(root, create)
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(path)
	parent, err := filepath.EvalSymlinks(filepath.Dir(clean))
	if err != nil {
		return "", err
	}
	clean = filepath.Join(parent, filepath.Base(clean))
	rel, err := filepath.Rel(dir, clean)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(filepath.ToSlash(rel), "../") {
		return "", errors.New("mission governance: receipt_outside")
	}
	return clean, nil
}

func WriteGovernanceReceipt(root, path string, receipt GovernanceReceipt) error {
	path, err := containedGovernancePath(root, path, true)
	if err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("mission governance: receipt_unsafe")
	}
	digest, err := governanceReceiptDigest(receipt)
	if err != nil {
		return err
	}
	receipt.Integrity = digest
	// The digest call just serialized the same fixed-shape receipt. Adding a
	// hexadecimal string cannot introduce a JSON encoding failure, so a second
	// defensive branch would be unreachable and would obscure the invariant.
	raw, _ := marshalGovernanceReceipt(receipt)
	tmp, err := os.CreateTemp(filepath.Dir(path), ".governance-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	// os.CreateTemp creates the file with mode 0600; avoid reopening it so the
	// write remains portable to Windows sharing semantics.
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := atomicfile.Replace(name, path); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func readGovernanceReceipt(root, path string) (GovernanceReceipt, error) {
	path, err := containedGovernancePath(root, path, false)
	if err != nil {
		return GovernanceReceipt{}, err
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return GovernanceReceipt{}, errors.New("mission governance: receipt_missing")
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return GovernanceReceipt{}, errors.New("mission governance: receipt_unsafe")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil || resolved != path {
		return GovernanceReceipt{}, errors.New("mission governance: receipt_unsafe")
	}
	raw, err := os.ReadFile(resolved)
	if err != nil {
		return GovernanceReceipt{}, err
	}
	r, err := ParseGovernanceReceipt(raw)
	if err != nil {
		return GovernanceReceipt{}, err
	}
	digest, err := governanceReceiptDigest(r)
	if err != nil || len(digest) != len(r.Integrity) || subtle.ConstantTimeCompare([]byte(digest), []byte(r.Integrity)) != 1 {
		return GovernanceReceipt{}, errors.New("mission governance: receipt_integrity")
	}
	return r, nil
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aa, bb := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(aa)
	sort.Strings(bb)
	for i := range aa {
		if aa[i] != bb[i] {
			return false
		}
	}
	return true
}

func validateGovernanceBinding(r GovernanceReceipt, kind GovernanceReceiptKind, status GovernanceReceiptStatus, issuer string, exp GovernanceExpectation) error {
	if r.Version != 1 || r.Kind != kind || r.Status != status || r.Issuer != issuer {
		if kind == GovernanceAudit && r.Status == GovernanceFailed {
			return errors.New("mission governance: audit_failed")
		}
		return errors.New("mission governance: receipt_status")
	}
	if !r.ExpiresAt.After(exp.Now) {
		return errors.New("mission governance: receipt_stale")
	}
	if r.MissionID != exp.MissionID || r.ContractHash != exp.ContractHash || r.SnapshotHash != exp.SnapshotHash || r.Action != exp.Action || !sameStringSet(r.Targets, exp.Targets) || r.HeadSHA != exp.HeadSHA {
		return errors.New("mission governance: receipt_lineage")
	}
	return nil
}

func LoadGovernanceReceipts(root, decisionPath, auditPath string, exp GovernanceExpectation) (GovernanceEvidence, error) {
	decision, err := readGovernanceReceipt(root, decisionPath)
	if err != nil {
		return GovernanceEvidence{}, err
	}
	audit, err := readGovernanceReceipt(root, auditPath)
	if err != nil {
		return GovernanceEvidence{}, err
	}
	if err := validateGovernanceBinding(decision, GovernanceDecision, GovernanceRecommended, "mission-governor", exp); err != nil {
		return GovernanceEvidence{}, err
	}
	if err := validateGovernanceBinding(audit, GovernanceAudit, GovernancePassed, "sync-auditor", exp); err != nil {
		return GovernanceEvidence{}, err
	}
	if decision.Issuer == audit.Issuer {
		return GovernanceEvidence{}, errors.New("mission governance: audit_not_independent")
	}
	return GovernanceEvidence{Decision: decision, Audit: audit}, nil
}
