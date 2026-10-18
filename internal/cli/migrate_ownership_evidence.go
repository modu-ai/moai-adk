package cli

// migrate_ownership_evidence.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865), design.md D-2: the evidence an UNREGISTERED file must carry before it
// may be classified `managed`.
//
// Byte equality with a template — the embedded template tree included — is not
// ownership: this binary's embedded tree need not be the tree that deployed the
// file, and a user can copy a template folder by hand. Ownership of a file the
// manifest does not know is established only by three items that do not derive
// from the file's own bytes:
//
//  1. the deployed version  — the moai binary version a prior manifest
//     generation recorded as the writer of that path;
//  2. the original deployed path — the prior generation's key for the file is
//     this very path (not the same hash under some other key);
//  3. a prior manifest record — that record's provenance is exactly
//     manifest.TemplateManaged. Any other provenance value (user_created,
//     user_modified, deprecated, generated_managed, empty, unknown) is not
//     evidence: the deployer records an existing file it did not write as
//     user_created and skips it, so a byte-identical user file can sit in a
//     prior generation with a record that proves the opposite.
//
// Even with all three present, the disk bytes must still equal the recorded
// template hash — a necessary check, never a sufficient one.
//
// The generations come from a provenanceSource (an interface, so a test can
// supply them directly). The operating source reads the retained
// manifest.prior.json copies of earlier run backups (update_prior_manifest.go).

import "github.com/modu-ai/moai-adk/internal/manifest"

// evidenceItem names one of the three evidence items.
type evidenceItem string

const (
	evidenceDeployedVersion evidenceItem = "deployed-version"
	evidenceOriginalPath    evidenceItem = "original-deployed-path"
	evidencePriorRecord     evidenceItem = "prior-manifest-record"
)

// allEvidenceItems is the canonical evidence order every list is written in.
var allEvidenceItems = []evidenceItem{evidenceDeployedVersion, evidenceOriginalPath, evidencePriorRecord}

// priorRecord is one path's record in a prior manifest generation.
type priorRecord struct {
	provenance      manifest.Provenance
	templateHash    string
	deployedHash    string
	deployedVersion string
}

// priorGeneration is one prior manifest generation, named for the run backup
// it came from.
type priorGeneration struct {
	name  string
	files map[string]priorRecord
}

// provenanceSource yields prior manifest generations, newest first. A
// generation that is absent, unreadable, partial, or not a valid manifest is
// simply not yielded: it is no evidence.
type provenanceSource interface {
	generations() []priorGeneration
}

// noEvidenceSource yields nothing; every unregistered file is then kept.
type noEvidenceSource struct{}

func (noEvidenceSource) generations() []priorGeneration { return nil }

// newProvenanceSource builds the evidence source the migration uses for one
// project. It is a package-level variable, like removalEnabled: a test swaps it
// to supply generations directly, and t.Cleanup restores it. No production code
// path replaces it with a test value.
var newProvenanceSource = func(projectRoot string) provenanceSource {
	return newPriorGenerationSource(projectRoot, "")
}

// attestation records the evidence an attested managed verdict rests on.
type attestation struct {
	DeployedVersion string
	OriginalPath    string
	Generation      string
}

// evidenceResult is what provenanceEvidence found for one unregistered path.
type evidenceResult struct {
	// attested is non-nil only when all three items are present and the disk
	// bytes match the recorded template hash.
	attested *attestation
	// reason and missing explain an unattested result.
	reason  keepReason
	missing []evidenceItem
}

// provenanceEvidence resolves the three evidence items for an unregistered path
// from the prior generations (newest first). diskHash is the hash of the file's
// current bytes.
//
// Where generations disagree about a path, the NEWEST generation holding a
// record for it decides (design.md D-2, debt N14): an older complete record
// never outvotes a newer incomplete one, so disagreement resolves toward
// keeping the file. The recorded template hash alone is compared with the disk
// bytes (debt N15): the registered path compares the template hash only, and a
// deployed-hash match alone describes a merged or retracked file, not a
// template copy.
func provenanceEvidence(src provenanceSource, rel, diskHash string) evidenceResult {
	gens := src.generations()
	var rec priorRecord
	var genName string
	found := false
	for _, g := range gens {
		if r, ok := g.files[rel]; ok {
			rec, genName, found = r, g.name, true
			break
		}
	}
	if !found {
		// A record with the same template hash under another key means the file
		// was deployed somewhere else: the original deployed path is missing.
		for _, g := range gens {
			for key, r := range g.files {
				if key != rel && r.templateHash != "" && r.templateHash == diskHash {
					return evidenceResult{
						reason:  evidenceIncompleteReason(evidenceOriginalPath),
						missing: []evidenceItem{evidenceOriginalPath},
					}
				}
			}
		}
		return evidenceResult{reason: keepNoProvenanceRecord, missing: append([]evidenceItem(nil), allEvidenceItems...)}
	}

	var missing []evidenceItem
	if rec.deployedVersion == "" {
		missing = append(missing, evidenceDeployedVersion)
	}
	if rec.provenance != manifest.TemplateManaged {
		missing = append(missing, evidencePriorRecord)
	}
	if len(missing) > 0 {
		return evidenceResult{reason: evidenceIncompleteReason(missing[0]), missing: missing}
	}
	if rec.templateHash == "" || rec.templateHash != diskHash {
		return evidenceResult{reason: keepRecordHashMismatch}
	}
	return evidenceResult{attested: &attestation{
		DeployedVersion: rec.deployedVersion,
		OriginalPath:    rel,
		Generation:      genName,
	}}
}
