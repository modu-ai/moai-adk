package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// This file carries the tier-generic core of the hook-preservation machinery
// introduced by SPEC-PRECOMMIT-PRESERVE-001 (t230) for the pre-commit hook and
// extended to the pre-push hook by card t257: the three-way attribution
// classifier, the provenance sidecar, and the pre-replacement backup. Both
// installers share this single implementation — a per-tier second copy is
// forbidden (t257 card scope (2)); per-hook differences live in the hookTier
// value each installer passes in.

// hookTier names the per-hook constants of one install tier. Everything the
// shared machinery needs to specialize its behaviour — messages, sidecar name,
// backup prefix — is carried here so the algorithm itself is written once.
type hookTier struct {
	// label names the hook in human-facing error messages ("pre-commit",
	// "pre-push").
	label string

	// provenanceName is the sidecar basename recording the SHA-256 of the hook
	// content this tool last wrote. It lives beside the hook in .git/hooks/,
	// so it shares the hook's lifetime exactly: wiping .git/hooks/ takes both.
	//
	// The record is the third operand of the attribution in REQ-PCP-014.
	// Without it only "installed" and "incoming" exist, and those two cannot
	// separate "the user changed it" from "we changed it" — a routine version
	// bump then reads as a user patch for every user on every release.
	provenanceName string

	// backupPrefix is the basename prefix of the backup copies taken before a
	// user-modified hook is replaced: `<hook>.bak.<stamp>`, the pattern
	// REQ-PCP-003 names. The stamp is a colon-free UTC form (20060102T150405Z)
	// — RFC3339 proper contains `:`, which is illegal in a Windows filename,
	// and a backup the user cannot create/read on their own platform is not a
	// backup.
	backupPrefix string
}

// preCommitTier is the pre-commit specialization of the shared machinery.
var preCommitTier = hookTier{
	label:          "pre-commit",
	provenanceName: moaiPreCommitProvenanceName,
	backupPrefix:   preCommitBackupPrefix,
}

// prePushTier is the pre-push specialization of the shared machinery (t257).
var prePushTier = hookTier{
	label:          "pre-push",
	provenanceName: moaiPrePushProvenanceName,
	backupPrefix:   prePushBackupPrefix,
}

// hookClass is the attribution verdict for an existing marker-bearing hook.
type hookClass int

const (
	// hookUnmodified: the hook is as MoAI last wrote it. Any difference
	// against the incoming content is an upstream version bump, so the
	// overwrite is quiet (REQ-PCP-002).
	hookUnmodified hookClass = iota
	// hookUserModified: the hook was edited after MoAI wrote it. This is the
	// loss-bearing case; the backup and the notice hang off it.
	hookUserModified
)

func (c hookClass) String() string {
	if c == hookUserModified {
		return "user-modified"
	}
	return "unmodified"
}

// hookBasis names which operands produced the verdict — the third label of the
// three-way classification, kept separate from the verdict because
// "undecidable-legacy" describes how the answer was reached, not what it was.
type hookBasis int

const (
	// hookBasisRecord: a usable provenance record existed, so the verdict
	// comes from installed-vs-recorded — the three-way comparison.
	hookBasisRecord hookBasis = iota
	// hookBasisUndecidableLegacy: no usable record existed, so attribution was
	// impossible and the verdict falls back to installed-vs-incoming, with any
	// difference read as a user edit (REQ-PCP-005). Deliberately the noisy
	// direction: a hand-patched legacy hook is the most likely thing to be
	// found without a record.
	hookBasisUndecidableLegacy
)

func (b hookBasis) String() string {
	if b == hookBasisUndecidableLegacy {
		return "undecidable-legacy"
	}
	return "record"
}

// hookAttribution is one classification of an existing marker-bearing hook.
type hookAttribution struct {
	Class hookClass
	Basis hookBasis
}

// classifyHook decides whether an existing marker-bearing hook was edited by
// the user, from three operands: the installed bytes, the digest this tool
// last recorded writing (empty when absent or unusable), and the incoming
// bytes.
//
// A two-way comparison of installed against incoming cannot make this call: an
// upstream version bump produces the same signal as a user patch, so a two-way
// design warns every user on every release that touches the hook (REQ-PCP-014).
func classifyHook(installed []byte, recordedDigest string, incoming []byte) hookAttribution {
	if recordedDigest == "" {
		// No usable record: attribution is impossible, so fall back to
		// installed-vs-incoming and read any difference as a user edit.
		if bytes.Equal(installed, incoming) {
			return hookAttribution{Class: hookUnmodified, Basis: hookBasisUndecidableLegacy}
		}
		return hookAttribution{Class: hookUserModified, Basis: hookBasisUndecidableLegacy}
	}

	if digestOfBytes(installed) == recordedDigest {
		return hookAttribution{Class: hookUnmodified, Basis: hookBasisRecord}
	}
	return hookAttribution{Class: hookUserModified, Basis: hookBasisRecord}
}

// readHookProvenance returns the recorded digest, or "" when no usable record
// exists. A missing, unreadable or malformed record is treated as absent,
// which routes the caller to the deliberately noisy legacy path (REQ-PCP-005).
func readHookProvenance(hookDir string, tier hookTier) string {
	raw, err := os.ReadFile(filepath.Join(hookDir, tier.provenanceName))
	if err != nil {
		return ""
	}
	digest := strings.TrimSpace(string(raw))
	if len(digest) != hex.EncodedLen(sha256.Size) {
		return ""
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return ""
	}
	return digest
}

// writeHookProvenance records the digest of the content just written.
func writeHookProvenance(hookDir, content string, tier hookTier) error {
	path := filepath.Join(hookDir, tier.provenanceName)
	if err := os.WriteFile(path, []byte(digestOfBytes([]byte(content))+"\n"), 0o644); err != nil {
		return fmt.Errorf("write %s provenance record: %w", tier.label, err)
	}
	return nil
}

func digestOfBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// backupHook copies the pre-run hook bytes to a timestamped backup in hookDir
// and returns the path chosen. The name is `<hook>.bak.<UTC colon-free
// stamp>`; when a candidate path is occupied a distinct sibling suffix is
// chosen instead — an existing backup is never overwritten (REQ-PCP-009).
// O_EXCL makes the never-clobber guarantee hold even if a file appears between
// the choice and the write.
func backupHook(hookDir string, preRun []byte, now time.Time, tier hookTier) (string, error) {
	stamp := now.UTC().Format("20060102T150405Z")
	for i := 0; ; i++ {
		name := tier.backupPrefix + stamp
		if i > 0 {
			name = fmt.Sprintf("%s.%d", name, i)
		}
		path := filepath.Join(hookDir, name)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
		if err != nil {
			if os.IsExist(err) {
				continue // occupied: pick a distinct name, never clobber
			}
			return "", fmt.Errorf("back up user-modified %s hook to %s: %w", tier.label, path, err)
		}
		if _, err := f.Write(preRun); err != nil {
			_ = f.Close()
			return "", fmt.Errorf("back up user-modified %s hook to %s: %w", tier.label, path, err)
		}
		if err := f.Close(); err != nil {
			return "", fmt.Errorf("back up user-modified %s hook to %s: %w", tier.label, path, err)
		}
		return path, nil
	}
}
