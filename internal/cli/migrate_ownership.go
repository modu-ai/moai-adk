package cli

// migrate_ownership.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card t1865): the one
// ownership verdict function the project migration, the dry-run preview and the
// remaining-file list share (REQ-OWR-001).
//
// A verdict is two separate steps and the second never follows from the first:
//
//	ownership class  (closed set of four)   managed | user-modified | user-created | unregistered
//	disposition      (two outcomes)         keep(<reason>) | removal-eligible
//
// `removal-eligible` is the candidate verdict the function gives from THIS run's
// inputs: the class is managed, the copy is not kept for profile reach, its user
// counterpart is confirmed, the removal switch is on, and the path stays inside
// the project root. Run-time conditions — the backup copy and OS read/delete
// errors — are outside the verdict. The function reads the file system and the
// manifest; it never writes and never removes.
//
// Safe default: when evidence is missing the verdict is a keep with a recorded
// reason. Nothing here ever promotes a file on byte equality with a template.

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// ownershipClass is the closed ownership class set.
type ownershipClass string

const (
	ownershipManaged      ownershipClass = "managed"
	ownershipUserModified ownershipClass = "user-modified"
	ownershipUserCreated  ownershipClass = "user-created"
	ownershipUnregistered ownershipClass = "unregistered"
)

// valid reports whether c is one of the four classes.
func (c ownershipClass) valid() bool {
	switch c {
	case ownershipManaged, ownershipUserModified, ownershipUserCreated, ownershipUnregistered:
		return true
	}
	return false
}

// keepReason is the closed set of reasons a file is kept. The set grows only by
// a SPEC revision: an open vocabulary would make per-reason tallies meaningless
// (design.md D-1).
type keepReason string

const (
	keepUserModified           keepReason = "user-modified"
	keepUserCreated            keepReason = "user-created"
	keepNoProvenanceRecord     keepReason = "no-provenance-record"
	keepRecordHashMismatch     keepReason = "record-hash-mismatch"
	keepProfileReach           keepReason = "profile-reach"
	keepCounterpartUnconfirmed keepReason = "counterpart-unconfirmed"
	keepNonCatalog             keepReason = "non-catalog"
	keepRemovalGateOff         keepReason = "removal-gate-off"
	keepSymlinkEscape          keepReason = "symlink-escape"
	keepReadFailed             keepReason = "read-failed"
	keepNoTemplateHash         keepReason = "no-template-hash"
	keepEmptyDirNoProvenance   keepReason = "empty-dir-no-provenance"
	keepBackupFailed           keepReason = "backup-failed"
)

// evidenceIncompletePrefix prefixes the parameterised reason that names which
// evidence item an unregistered file lacks.
const evidenceIncompletePrefix = "evidence-incomplete:"

// evidenceIncompleteReason is the reason for an unregistered file that has some
// evidence but lacks the named item.
func evidenceIncompleteReason(item evidenceItem) keepReason {
	return keepReason(evidenceIncompletePrefix + string(item))
}

// isKnownKeepReason reports whether r belongs to the closed reason set.
func isKnownKeepReason(r keepReason) bool {
	switch r {
	case keepUserModified, keepUserCreated, keepNoProvenanceRecord, keepRecordHashMismatch,
		keepProfileReach, keepCounterpartUnconfirmed, keepNonCatalog, keepRemovalGateOff,
		keepSymlinkEscape, keepReadFailed, keepNoTemplateHash, keepEmptyDirNoProvenance,
		keepBackupFailed:
		return true
	}
	if item, ok := strings.CutPrefix(string(r), evidenceIncompletePrefix); ok {
		for _, known := range allEvidenceItems {
			if evidenceItem(item) == known {
				return true
			}
		}
	}
	return false
}

// remainingRow is one file the migration leaves in place: the unit of the
// remaining-file list (REQ-OWR-009).
type remainingRow struct {
	Rel     string
	Class   ownershipClass
	Reason  keepReason
	Missing []evidenceItem
}

// ownershipEnv carries the inputs of a verdict. Every field is read-only
// evidence the caller already holds; the function resolves nothing on its own
// except the file's own bytes and symlink target.
type ownershipEnv struct {
	// entry looks up a path in the project manifest.
	entry func(rel string) (*manifest.FileEntry, bool)
	// evidence yields prior manifest generations for unregistered files.
	evidence provenanceSource
	// counterpartConfirmed reports whether the user-root counterpart of rel is
	// manifest-tracked, hash-equal, and the current version's content.
	counterpartConfirmed func(rel string) bool
	// notInCatalog reports whether the catalog carries no installer source for rel.
	notInCatalog func(rel string) bool
	// profileMapped is true when the launch ledger maps the project to a profile.
	profileMapped bool
	// removalEnabled is the removal switch (default false).
	removalEnabled bool
	// rootResolved is the project root resolved through symlinks.
	rootResolved string
}

// ownershipVerdict is the verdict for one file.
type ownershipVerdict struct {
	class ownershipClass
	// removalEligible is true only when every disposition condition holds.
	removalEligible bool
	// keep is the reason when the disposition is keep; empty when eligible.
	keep keepReason
	// missing lists the evidence items an unregistered file lacks.
	missing []evidenceItem
	// provenance is the recorded provenance (TemplateManaged for an attested file).
	provenance manifest.Provenance
	// attestation is set only for an unregistered file promoted to managed.
	attestation *attestation
	// data is the file's current bytes, when they were read.
	data []byte
	// bytesMatchTemplate is true when the bytes equal the recorded template hash.
	bytesMatchTemplate bool
}

// classifyOwnershipFn is the verdict function the migration calls. It is a
// variable only so a test can inject a deliberately wrong classifier and watch
// the preservation assertions go red; no production path replaces it.
var classifyOwnershipFn = classifyOwnership

// classifyOwnership decides the ownership class and the disposition of the file
// at abs, whose project-relative slash path is rel.
func classifyOwnership(env ownershipEnv, rel, abs string) ownershipVerdict {
	entry, registered := env.entry(rel)
	if !registered || entry == nil {
		return classifyUnregistered(env, rel, abs)
	}

	v := ownershipVerdict{provenance: entry.Provenance}
	if entry.Provenance == manifest.UserCreated {
		v.class, v.keep = ownershipUserCreated, keepUserCreated
		return v
	}

	// The deletion preconditions are mandatory: a failed read or an empty
	// recorded template hash PRESERVES the file. Both mean the bytes cannot be
	// verified against the template, so the file is treated as user-modified.
	data, err := os.ReadFile(abs)
	if err != nil {
		v.class, v.keep = ownershipUserModified, keepReadFailed
		return v
	}
	v.data = data
	if entry.TemplateHash == "" {
		v.class, v.keep = ownershipUserModified, keepNoTemplateHash
		return v
	}
	v.bytesMatchTemplate = manifest.HashBytes(data) == entry.TemplateHash
	if !v.bytesMatchTemplate || entry.Provenance == manifest.UserModified {
		v.class, v.keep = ownershipUserModified, keepUserModified
		return v
	}
	v.class = ownershipManaged
	return disposition(env, v, rel, abs)
}

// classifyUnregistered judges a file the manifest has no entry for. It is
// `unregistered` unless the prior generations prove all three evidence items.
func classifyUnregistered(env ownershipEnv, rel, abs string) ownershipVerdict {
	v := ownershipVerdict{class: ownershipUnregistered}
	data, err := os.ReadFile(abs)
	if err != nil {
		v.keep = keepReadFailed
		return v
	}
	v.data = data
	ev := provenanceEvidence(env.evidence, rel, manifest.HashBytes(data))
	if ev.attested == nil {
		v.keep, v.missing = ev.reason, ev.missing
		return v
	}
	v.class = ownershipManaged
	v.provenance = manifest.TemplateManaged
	v.attestation = ev.attested
	v.bytesMatchTemplate = true
	return disposition(env, v, rel, abs)
}

// disposition applies the removal gates to a managed file in the fixed order:
// profile reach, user counterpart, root boundary, removal switch. The first
// gate that does not hold names the keep reason.
func disposition(env ownershipEnv, v ownershipVerdict, rel, abs string) ownershipVerdict {
	// A project the launch ledger maps to a profile keeps its shipped copies
	// (template_managed); a retired (deprecated) copy is not in the kept set.
	if env.profileMapped && v.provenance == manifest.TemplateManaged {
		v.keep = keepProfileReach
		return v
	}
	if !env.counterpartConfirmed(rel) {
		if env.notInCatalog(rel) {
			v.keep = keepNonCatalog
		} else {
			v.keep = keepCounterpartUnconfirmed
		}
		return v
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil || !withinRootBoundary(env.rootResolved, resolved) {
		v.keep = keepSymlinkEscape
		return v
	}
	if !env.removalEnabled {
		v.keep = keepRemovalGateOff
		return v
	}
	v.removalEligible = true
	return v
}

// reasonTag renders the bracketed reason a report line carries.
func reasonTag(v ownershipVerdict) string {
	if v.keep == "" {
		return ""
	}
	tag := " [reason: " + string(v.keep)
	if len(v.missing) > 0 {
		names := make([]string, len(v.missing))
		for i, m := range v.missing {
			names[i] = string(m)
		}
		tag += "; missing evidence: " + strings.Join(names, ", ")
	}
	return tag + "]"
}
