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
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
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
	// keepNonRegularFile: the path is a FIFO, socket, device or a link to one. It is
	// kept without ever being opened, because reading such a file can block forever.
	keepNonRegularFile keepReason = "non-regular-file"
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
		keepBackupFailed, keepNonRegularFile:
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

// counterpartStatus is the state of a file's user-root counterpart (design.md
// D-7). `n/a` is for a row whose verdict never reached the counterpart question
// (a user-modified file, an unattested unregistered file) or whose path has no
// installer source to be a counterpart of.
type counterpartStatus string

const (
	counterpartStatusConfirmed     counterpartStatus = "confirmed"
	counterpartStatusAbsent        counterpartStatus = "absent"
	counterpartStatusUnconfirmed   counterpartStatus = "unconfirmed"
	counterpartStatusNotApplicable counterpartStatus = "n/a"
)

// profileSideStatus is the state of the profile-side counterpart. This update
// never opens a profile root (REQ-OWR-005, decision index Q5), so a project the
// launch ledger maps to a profile records `not-checked` and any other project
// `n/a`; there is no third state because nothing here reads that side.
type profileSideStatus string

const (
	profileSideNotChecked    profileSideStatus = "not-checked"
	profileSideNotApplicable profileSideStatus = "n/a"
)

// remainingRow is one file the migration leaves in place: the unit of the
// remaining-file list (REQ-OWR-009).
type remainingRow struct {
	Rel     string
	Class   ownershipClass
	Reason  keepReason
	Missing []evidenceItem
	// Counterpart is the user-root counterpart state; ProfileSide the profile
	// side (always `not-checked` for a profile-mapped project).
	Counterpart counterpartStatus
	ProfileSide profileSideStatus
	// CaseVariantOf is the registered or shipped spelling this path differs from
	// only in letter case, or "" (REQ-OWR-007).
	CaseVariantOf string
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
	// counterpartStatus optionally reports the counterpart's state for the
	// remaining row. When nil the state is derived from counterpartConfirmed
	// (confirmed or unconfirmed). It never decides the disposition: that stays
	// with counterpartConfirmed.
	counterpartStatus func(rel string) counterpartStatus
	// caseAlias resolves a walk path that is not a manifest key to the single key
	// that differs from it only in letter case and names the same file
	// (REQ-OWR-007, design.md D-4). Optional.
	caseAlias func(rel string) (key string, ok bool)
	// shippedSpelling returns the shipped template path that differs from rel only
	// in letter case, or "". Optional.
	shippedSpelling func(rel string) string
}

// counterpartOf reports the user-root counterpart state of rel for a row.
func (e ownershipEnv) counterpartOf(rel string) counterpartStatus {
	if e.counterpartStatus != nil {
		return e.counterpartStatus(rel)
	}
	if e.counterpartConfirmed(rel) {
		return counterpartStatusConfirmed
	}
	return counterpartStatusUnconfirmed
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
	// counterpart is the user-root counterpart state, set when the verdict
	// reached the counterpart question; profileSide is set for every verdict.
	counterpart counterpartStatus
	profileSide profileSideStatus
	// key is the manifest key the verdict applies to when it differs from the
	// walk path (a case alias); empty otherwise.
	key string
	// caseVariantOf is the spelling this path differs from only in letter case:
	// the registered key of a case alias, else the shipped path.
	caseVariantOf string
}

// manifestKey is the manifest key to act on for a walk path rel: the registered
// spelling for a case alias, rel itself otherwise.
func (v ownershipVerdict) manifestKey(rel string) string {
	if v.key != "" {
		return v.key
	}
	return rel
}

// row renders the verdict as a remaining-list row for rel. A state the verdict
// never established is `n/a`, never an empty string.
func (v ownershipVerdict) row(rel string) remainingRow {
	r := remainingRow{
		Rel: rel, Class: v.class, Reason: v.keep, Missing: v.missing,
		Counterpart: v.counterpart, ProfileSide: v.profileSide, CaseVariantOf: v.caseVariantOf,
	}
	if r.Counterpart == "" {
		r.Counterpart = counterpartStatusNotApplicable
	}
	if r.ProfileSide == "" {
		r.ProfileSide = profileSideNotApplicable
	}
	return r
}

// rowKeptFor is the row of a removal-eligible file that stays because the run
// could not take every backup copy: same facts, the given keep reason.
func (v ownershipVerdict) rowKeptFor(rel string, reason keepReason) remainingRow {
	r := v.row(rel)
	r.Reason = reason
	return r
}

// classifyOwnershipFn is the verdict function the migration calls. It is a
// variable only so a test can inject a deliberately wrong classifier and watch
// the preservation assertions go red; no production path replaces it.
var classifyOwnershipFn = classifyOwnership

// classifyOwnership decides the ownership class and the disposition of the file
// at abs, whose project-relative slash path is rel. It also records the profile
// side of the row: `not-checked` for a profile-mapped project, because no code
// here reads a profile root (REQ-OWR-005).
func classifyOwnership(env ownershipEnv, rel, abs string) ownershipVerdict {
	v := classifyOwnershipCore(env, rel, abs)
	if v.caseVariantOf == "" && env.shippedSpelling != nil {
		v.caseVariantOf = env.shippedSpelling(rel)
	}
	if env.profileMapped {
		v.profileSide = profileSideNotChecked
	} else {
		v.profileSide = profileSideNotApplicable
	}
	return v
}

// classifyOwnershipCore is the class and disposition decision proper.
func classifyOwnershipCore(env ownershipEnv, rel, abs string) ownershipVerdict {
	entry, registered := env.entry(rel)
	// The exact spelling is looked up first. Only a path the manifest does not
	// know may be a case alias of a registered key, and only when the file system
	// says both spellings are the same file; the file is never renamed.
	key := rel
	if (!registered || entry == nil) && env.caseAlias != nil {
		if alias, ok := env.caseAlias(rel); ok {
			if aliased, found := env.entry(alias); found && aliased != nil {
				entry, registered, key = aliased, true, alias
			}
		}
	}
	if !registered || entry == nil {
		// The kind of the file is decided before anything reads it: a FIFO would
		// block the read until a writer appears.
		if unreadableKind(abs) {
			return ownershipVerdict{class: ownershipUnregistered, keep: keepNonRegularFile}
		}
		return classifyUnregistered(env, rel, abs)
	}

	v := ownershipVerdict{provenance: entry.Provenance}
	if key != rel {
		v.key, v.caseVariantOf = key, key
	}
	if entry.Provenance == manifest.UserCreated {
		v.class, v.keep = ownershipUserCreated, keepUserCreated
		return v
	}
	if unreadableKind(abs) {
		v.class, v.keep = ownershipUserModified, keepNonRegularFile
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
	return disposition(env, v, key, abs)
}

// unreadableKind reports whether abs is a FIFO, socket, device or irregular file,
// or a symlink whose target is one. It asks the file system for the kind without
// opening the path (os.Stat follows links, as the later os.ReadFile does), so it
// cannot block. A directory, a dangling link and a path that cannot be examined
// are NOT reported here: they keep failing the read the way they always did
// (`read-failed`), so a symlink to a regular file or to nothing behaves as before.
func unreadableKind(abs string) bool {
	info, err := os.Stat(abs)
	if err != nil {
		return false
	}
	return info.Mode()&(os.ModeNamedPipe|os.ModeSocket|os.ModeDevice|os.ModeCharDevice|os.ModeIrregular) != 0
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
	v.counterpart = env.counterpartOf(rel)
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
	if v.caseVariantOf != "" {
		tag += "; case-variant of: " + v.caseVariantOf
	}
	return tag + "]"
}

// userCounterpartLocation maps a project common-asset path to the user-root slug
// and the root-relative path of its counterpart (the same mapping
// userCounterpartConfirmed applies), and reports false for a path outside the
// four common-asset roots.
func userCounterpartLocation(projectRel string) (slug userassets.RootSlug, rest string, ok bool) {
	switch {
	case strings.HasPrefix(projectRel, ".claude/skills/"):
		return userassets.RootClaudeSkills, strings.TrimPrefix(projectRel, ".claude/skills/"), true
	case strings.HasPrefix(projectRel, ".agents/skills/"):
		return userassets.RootAgentsSkills, strings.TrimPrefix(projectRel, ".agents/skills/"), true
	case strings.HasPrefix(projectRel, ".claude/agents/moai/"):
		return userassets.RootClaudeAgents, strings.TrimPrefix(projectRel, ".claude/agents/moai/"), true
	case strings.HasPrefix(projectRel, ".codex/agents/moai/"):
		return userassets.RootCodexAgents, strings.TrimPrefix(projectRel, ".codex/agents/moai/"), true
	}
	return "", "", false
}

// userCounterpartStatus reports the state of projectRel's user-root counterpart
// for a remaining row. It agrees with counterpartConfirmedCaseAware on `confirmed`;
// for anything else it separates a counterpart that is not there (`absent`)
// from one that is there but cannot be confirmed (`unconfirmed`: untracked,
// differing bytes, or not this version's content), and answers `n/a` for a path
// that has no installer source at all.
func userCounterpartStatus(userManifest *userassets.Manifest, cat *template.Catalog, embedded fs.FS, homeDir, projectRel string) counterpartStatus {
	if counterpartConfirmedCaseAware(userManifest, cat, embedded, homeDir, projectRel) {
		return counterpartStatusConfirmed
	}
	slug, rest, ok := userCounterpartLocation(projectRel)
	if !ok || skillNotInCatalog(cat, projectRel) {
		return counterpartStatusNotApplicable
	}
	dir := userassets.RootBySlugDir(homeDir, slug)
	if dir == "" {
		return counterpartStatusNotApplicable
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rest))); err != nil && errors.Is(err, fs.ErrNotExist) {
		return counterpartStatusAbsent
	}
	return counterpartStatusUnconfirmed
}
