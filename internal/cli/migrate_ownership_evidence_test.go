package cli

// migrate_ownership_evidence_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M1a
// (card t1865), AC-OWR-05 at FUNCTION level: an unregistered file is promoted
// to managed only when all three evidence items (deployed version, original
// deployed path, a prior manifest record whose provenance is exactly the
// template-managed value) are present AND the disk bytes match the recorded
// template hash. The prior generation is built from the REAL provenance
// constants of internal/manifest, never from a hand-written string. The
// operating-source arms (h1)(h2) live in update_prior_manifest_test.go (M1b).

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestOwnershipAttestedManagedOnlyWithFullEvidence is AC-OWR-05 (a)-(g).
func TestOwnershipAttestedManagedOnlyWithFullEvidence(t *testing.T) {
	const p = ".agents/skills/moai-plan/SKILL.md"
	pBytes := ownEmbeddedBytes(t, p)
	h := manifest.HashBytes(pBytes)
	const otherKey = ".agents/skills/moai-plan-moved/SKILL.md"

	complete := priorRecord{provenance: manifest.TemplateManaged, templateHash: h, deployedHash: h, deployedVersion: "v-prior"}
	withProv := func(prov manifest.Provenance) priorRecord { r := complete; r.provenance = prov; return r }
	noVersion := complete
	noVersion.deployedVersion = ""

	allThree := []evidenceItem{evidenceDeployedVersion, evidenceOriginalPath, evidencePriorRecord}

	judge := func(t *testing.T, root string, gens []priorGeneration, gateOn bool) ownershipVerdict {
		t.Helper()
		rootResolved, err := filepath.EvalSymlinks(root)
		if err != nil {
			t.Fatal(err)
		}
		env := ownershipEnv{
			entry:                func(string) (*manifest.FileEntry, bool) { return nil, false }, // P is NOT in the current manifest
			evidence:             fakeProvenance{gens: gens},
			counterpartConfirmed: func(string) bool { return true },
			notInCatalog:         func(string) bool { return false },
			removalEnabled:       gateOn,
			rootResolved:         rootResolved,
		}
		return classifyOwnership(env, p, filepath.Join(root, filepath.FromSlash(p)))
	}
	newRoot := func(t *testing.T, content []byte) string {
		t.Helper()
		root := t.TempDir()
		ownWriteFile(t, root, p, content)
		return root
	}

	t.Run("Positive", func(t *testing.T) {
		root := newRoot(t, pBytes)
		gens := []priorGeneration{{name: "gen-newest", files: map[string]priorRecord{p: complete}}}

		v := judge(t, root, gens, false)
		if v.class != ownershipManaged {
			t.Fatalf("complete evidence: class %q, want managed (verdict %+v)", v.class, v)
		}
		if v.attestation == nil {
			t.Fatalf("an attested managed verdict must carry its attestation record")
		}
		if v.attestation.DeployedVersion != "v-prior" || v.attestation.OriginalPath != p || v.attestation.Generation != "gen-newest" {
			t.Errorf("attestation %+v does not record version/path/generation", *v.attestation)
		}
		if v.removalEligible || v.keep != keepRemovalGateOff {
			t.Errorf("switch off: eligible=%v keep=%q, want keep(%s)", v.removalEligible, v.keep, keepRemovalGateOff)
		}
		if on := judge(t, root, gens, true); !on.removalEligible || on.keep != "" {
			t.Errorf("switch on + confirmed counterpart: %+v, want removal-eligible", on)
		}
	})

	mutants := []struct {
		name    string
		content []byte // nil: the template bytes
		gens    []priorGeneration
		class   ownershipClass
		keep    keepReason
		missing []evidenceItem
	}{
		{"a-no-deployed-version", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{p: noVersion}}},
			ownershipUnregistered, evidenceIncompleteReason(evidenceDeployedVersion), []evidenceItem{evidenceDeployedVersion}},
		{"b-record-only-under-another-path", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{otherKey: complete}}},
			ownershipUnregistered, evidenceIncompleteReason(evidenceOriginalPath), []evidenceItem{evidenceOriginalPath}},
		{"c-no-record-anywhere", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{".agents/skills/unrelated/SKILL.md": complete}}},
			ownershipUnregistered, keepNoProvenanceRecord, allThree},
		{"c-no-generation-at-all", nil, nil, ownershipUnregistered, keepNoProvenanceRecord, allThree},
		{"d-disk-bytes-differ-from-record", []byte("# edited after the recorded deploy\n"),
			[]priorGeneration{{name: "g", files: map[string]priorRecord{p: complete}}},
			ownershipUnregistered, keepRecordHashMismatch, nil},
		{"e-prior-provenance-user-created", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{p: withProv(manifest.UserCreated)}}},
			ownershipUnregistered, evidenceIncompleteReason(evidencePriorRecord), []evidenceItem{evidencePriorRecord}},
		{"f-prior-provenance-user-modified", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{p: withProv(manifest.UserModified)}}},
			ownershipUnregistered, evidenceIncompleteReason(evidencePriorRecord), []evidenceItem{evidencePriorRecord}},
		{"g-prior-provenance-deprecated", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{p: withProv(manifest.Deprecated)}}},
			ownershipUnregistered, evidenceIncompleteReason(evidencePriorRecord), []evidenceItem{evidencePriorRecord}},
		{"g-prior-provenance-generated-managed", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{p: withProv(manifest.GeneratedManaged)}}},
			ownershipUnregistered, evidenceIncompleteReason(evidencePriorRecord), []evidenceItem{evidencePriorRecord}},
		{"g-prior-provenance-empty", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{p: withProv("")}}},
			ownershipUnregistered, evidenceIncompleteReason(evidencePriorRecord), []evidenceItem{evidencePriorRecord}},
		{"g-prior-provenance-unknown-string", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{p: withProv("zzz")}}},
			ownershipUnregistered, evidenceIncompleteReason(evidencePriorRecord), []evidenceItem{evidencePriorRecord}},
		// N14 disposal: generations that disagree about the path. The NEWEST
		// generation holding a record for the path decides, so an older
		// complete record never outvotes a newer incomplete one.
		{"n14-newest-incomplete-older-complete", nil,
			[]priorGeneration{
				{name: "newest", files: map[string]priorRecord{p: withProv(manifest.UserCreated)}},
				{name: "older", files: map[string]priorRecord{p: complete}},
			},
			ownershipUnregistered, evidenceIncompleteReason(evidencePriorRecord), []evidenceItem{evidencePriorRecord}},
		// N15 disposal: the attested path compares the recorded TemplateHash
		// only, exactly like the registered path; a DeployedHash match alone
		// (a merged/retracked entry) is not evidence.
		{"n15-deployed-hash-match-only", nil,
			[]priorGeneration{{name: "g", files: map[string]priorRecord{p: {
				provenance: manifest.TemplateManaged, templateHash: manifest.HashBytes([]byte("some other template")),
				deployedHash: h, deployedVersion: "v-prior"}}}},
			ownershipUnregistered, keepRecordHashMismatch, nil},
	}
	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			content := m.content
			if content == nil {
				content = pBytes
			}
			root := newRoot(t, content)
			v := judge(t, root, m.gens, true)
			if v.class != m.class {
				t.Errorf("class %q, want %q", v.class, m.class)
			}
			if v.keep != m.keep {
				t.Errorf("keep %q, want %q", v.keep, m.keep)
			}
			if v.removalEligible {
				t.Errorf("an unregistered file without full evidence was removal-eligible")
			}
			if v.attestation != nil {
				t.Errorf("an unattested verdict carries an attestation: %+v", *v.attestation)
			}
			if !reflect.DeepEqual(v.missing, m.missing) {
				t.Errorf("missing evidence %v, want %v", v.missing, m.missing)
			}
			if got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p))); err != nil || string(got) != string(content) {
				t.Errorf("classification must not touch the file: %v", err)
			}
		})
	}
}
