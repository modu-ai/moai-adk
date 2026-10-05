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
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

type CompletionReceipt struct {
	Version        int             `json:"version"`
	MissionID      string          `json:"mission_id"`
	ContractHash   string          `json:"contract_hash"`
	SnapshotHash   string          `json:"snapshot_hash"`
	HeadSHA        string          `json:"head_sha"`
	Issuer         string          `json:"issuer"`
	Status         string          `json:"status"`
	Evidence       map[string]bool `json:"evidence"`
	LandedAncestry bool            `json:"landed_ancestry"`
	ExpiresAt      time.Time       `json:"expires_at"`
	Integrity      string          `json:"integrity"`
}

type CompletionExpectation struct {
	MissionID, ContractHash, SnapshotHash, HeadSHA string
	RequiredEvidence                               []string
	RequireLandedAncestry                          bool
	Now                                            time.Time
}

func canonicalCompletionReceipt(r CompletionReceipt) CompletionReceipt {
	r.Integrity = ""
	if r.Evidence != nil {
		keys := make([]string, 0, len(r.Evidence))
		for key := range r.Evidence {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		ordered := make(map[string]bool, len(keys))
		for _, key := range keys {
			ordered[key] = r.Evidence[key]
		}
		r.Evidence = ordered
	}
	return r
}

func completionReceiptDigest(r CompletionReceipt) (string, error) {
	raw, err := json.Marshal(canonicalCompletionReceipt(r))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func WriteCompletionReceipt(root, path string, receipt CompletionReceipt) error {
	path, err := containedGovernancePath(root, path, true)
	if err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("mission completion: receipt_unsafe")
	}
	digest, err := completionReceiptDigest(receipt)
	if err != nil {
		return err
	}
	receipt.Integrity = digest
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	return writeCompletionReceiptAtomic(path, raw)
}

func writeCompletionReceiptAtomic(path string, raw []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".completion-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
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

func LoadCompletionReceipt(root, path string, exp CompletionExpectation) (CompletionReceipt, error) {
	path, err := containedGovernancePath(root, path, false)
	if err != nil {
		return CompletionReceipt{}, err
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return CompletionReceipt{}, errors.New("mission completion: receipt_missing")
	}
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 {
		return CompletionReceipt{}, errors.New("mission completion: receipt_unsafe")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil || resolved != path {
		return CompletionReceipt{}, errors.New("mission completion: receipt_unsafe")
	}
	raw, err := os.ReadFile(resolved)
	if err != nil {
		return CompletionReceipt{}, err
	}
	var r CompletionReceipt
	if json.Unmarshal(raw, &r) != nil {
		return CompletionReceipt{}, errors.New("mission completion: receipt_parse")
	}
	digest, err := completionReceiptDigest(r)
	if err != nil || len(digest) != len(r.Integrity) || subtle.ConstantTimeCompare([]byte(digest), []byte(r.Integrity)) != 1 {
		return CompletionReceipt{}, errors.New("mission completion: receipt_integrity")
	}
	if r.Version != 1 || r.Issuer != "completion-auditor" || r.Status != "passed" {
		return CompletionReceipt{}, errors.New("mission completion: receipt_status")
	}
	if !r.ExpiresAt.After(exp.Now) {
		return CompletionReceipt{}, errors.New("mission completion: receipt_stale")
	}
	if r.MissionID != exp.MissionID || r.ContractHash != exp.ContractHash || r.SnapshotHash != exp.SnapshotHash || r.HeadSHA != exp.HeadSHA {
		return CompletionReceipt{}, errors.New("mission completion: receipt_lineage")
	}
	if len(r.Evidence) != len(exp.RequiredEvidence) {
		return CompletionReceipt{}, errors.New("mission completion: evidence_invalid")
	}
	for _, key := range exp.RequiredEvidence {
		if !r.Evidence[key] {
			return CompletionReceipt{}, errors.New("mission completion: evidence_invalid")
		}
	}
	if exp.RequireLandedAncestry && !r.LandedAncestry {
		return CompletionReceipt{}, errors.New("mission completion: ancestry_missing")
	}
	return r, nil
}
