package cli

// update_prior_manifest.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M1b (card t1865),
// decision-index Q2 option (a), conditional: the evidence-PRODUCING half of the
// ownership repair.
//
//   - Each update run keeps the manifest as it was loaded at the start of the
//     run, byte for byte, in that run's backup directory as manifest.prior.json
//     (REQ-OWR-015). The copy is created atomically — a temporary file in the
//     same directory, then a rename — so a partial or failed write never reads
//     as a valid generation. It is also what lets a restore put the manifest
//     back together with the files a run removed (REQ-OWR-010).
//   - When the run backup cannot hold that copy, or the bytes of any removal
//     candidate, the run removes no file at all (REQ-OWR-016; the migration's
//     collect -> back up everything -> remove structure).
//   - The retained copies of earlier runs are the prior manifest generations
//     the ownership verdict reads as evidence (REQ-OWR-017): at most the newest
//     priorManifestGenerations of them, any damaged one counting as no
//     evidence. That number is a lookup bound ONLY — nothing here deletes,
//     prunes or shortens the retention of a run backup.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// priorManifestName is the file the start-of-run manifest is kept under.
const priorManifestName = backup.PriorManifestName

// priorManifestGenerations bounds how many retained generations the evidence
// source reads (newest first). It is a lookup bound only; the backup retention
// policy (backup.PruneBackups) is not touched by it and does not read it.
const priorManifestGenerations = 8

// priorManifestState is the outcome of preserving the start-of-run manifest,
// handed from the Backup stage to the migration.
type priorManifestState int

const (
	// The zero value means no Backup stage ran before the migration (a direct
	// call): the migration is then not blocked on a manifest copy.
	_ priorManifestState = iota
	// priorManifestWritten: the copy is in the run backup.
	priorManifestWritten
	// priorManifestAbsent: the project had no manifest file at load time, so
	// there was nothing to preserve and nothing failed.
	priorManifestAbsent
	// priorManifestFailed: the copy could not be made; the run removes nothing.
	priorManifestFailed
)

// priorManifestTempWrite writes data to a fresh temporary file in dir and
// returns its path. It and priorManifestRename are the two observation points of
// the atomic creation: a test replaces them with recording (or failing) wrappers
// of the operating functions, and t.Cleanup restores them. No production code
// path replaces them.
var priorManifestTempWrite = writePriorManifestTemp

// priorManifestRename moves the temporary file onto its final name.
var priorManifestRename = os.Rename

// priorManifestWriter is how the Backup stage preserves the manifest. Its
// default is writePriorManifest; a test wraps it to inject a failure.
var priorManifestWriter = writePriorManifest

// writePriorManifestTemp writes data to a temporary file in dir (a name other
// than the final one, so a half-written file never answers to it) and syncs it.
func writePriorManifestTemp(dir string, data []byte) (string, error) {
	f, err := os.CreateTemp(dir, priorManifestName+".tmp-*")
	if err != nil {
		return "", fmt.Errorf("create temporary prior manifest: %w", err)
	}
	name := f.Name()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(name)
		return "", fmt.Errorf("write temporary prior manifest: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(name)
		return "", fmt.Errorf("sync temporary prior manifest: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return "", fmt.Errorf("close temporary prior manifest: %w", err)
	}
	return name, nil
}

// writePriorManifest creates <runDir>/manifest.prior.json holding data. The
// final name is reached only by renaming a complete temporary file onto it, so
// it either holds every byte or does not exist. A temporary file left by a
// failed rename is complete-looking but carries another name and is ignored by
// the reader.
func writePriorManifest(runDir string, data []byte) error {
	tmp, err := priorManifestTempWrite(runDir, data)
	if err != nil {
		return fmt.Errorf("preserve prior manifest: %w", err)
	}
	final := filepath.Join(runDir, priorManifestName)
	if err := priorManifestRename(tmp, final); err != nil {
		return fmt.Errorf("preserve prior manifest: rename into place: %w", err)
	}
	return nil
}

// readStartManifest returns the manifest file's bytes exactly as they are now,
// or nil when the project has none. The caller takes the copy right after the
// run's manifest load, before anything in the run can save the manifest.
func readStartManifest(projectRoot string) []byte {
	data, err := os.ReadFile(filepath.Join(projectRoot, defs.MoAIDir, defs.ManifestJSON))
	if err != nil {
		return nil
	}
	return data
}

// preservePriorManifest keeps data (the start-of-run manifest bytes) in the
// run's backup directory. It returns the directory used (the run's own when the
// config backup made one, a fresh run-scoped one otherwise), the outcome, and
// the failure when there is one. A nil data means the project had no manifest to
// preserve.
func preservePriorManifest(projectRoot, configBackupPath string, data []byte) (string, priorManifestState, error) {
	if data == nil {
		return configBackupPath, priorManifestAbsent, nil
	}
	runDir, err := ensureRunBackupDir(projectRoot, configBackupPath)
	if err != nil {
		return configBackupPath, priorManifestFailed, err
	}
	if err := priorManifestWriter(runDir, data); err != nil {
		return runDir, priorManifestFailed, err
	}
	return runDir, priorManifestWritten, nil
}

// recordManifestDeployedVersion refreshes the manifest-level version to the
// binary that deployed. The deployment timestamp moves only when the version
// does, so a repeat run at the same version leaves the manifest bytes unchanged.
func recordManifestDeployedVersion(mgr manifest.Manager, ver string) {
	mf := mgr.Manifest()
	if mf == nil || ver == "" || mf.Version == ver {
		return
	}
	mf.Version = ver
	mf.DeployedAt = time.Now().UTC().Format(time.RFC3339)
}

// priorGenerationSource is the operating provenanceSource: the manifest.prior.json
// copies retained in earlier run backups, newest first.
type priorGenerationSource struct {
	projectRoot string
	// exclude is a run directory whose copy is not read (the current run's own
	// copy is the start-of-run manifest, not a prior generation).
	exclude string
	loaded  bool
	gens    []priorGeneration
	// reads counts the generation files the reader opened or tried to open.
	reads int
}

// newPriorGenerationSource returns a lazily-loading source over projectRoot's
// run backups.
func newPriorGenerationSource(projectRoot, excludeRunDir string) *priorGenerationSource {
	return &priorGenerationSource{projectRoot: projectRoot, exclude: excludeRunDir}
}

func (s *priorGenerationSource) generations() []priorGeneration {
	if !s.loaded {
		s.gens, s.reads = readPriorGenerations(s.projectRoot, s.exclude)
		s.loaded = true
	}
	return s.gens
}

// priorFileJSON is the part of a manifest entry the evidence resolver reads.
type priorFileJSON struct {
	Provenance      manifest.Provenance `json:"provenance"`
	TemplateHash    string              `json:"template_hash"`
	DeployedHash    string              `json:"deployed_hash"`
	DeployedVersion string              `json:"deployed_version"`
}

// readPriorGenerations reads the newest priorManifestGenerations retained
// manifest.prior.json copies under projectRoot's backup directory. A run backup
// is a candidate only if it holds an entry of that exact name; a generation that
// is not a regular file, cannot be read, is truncated, is not JSON, or has no
// files object is no evidence and is skipped without failing the run. The second
// result is the number of generation files tried. This is the only place the
// priorManifestGenerations bound is consulted.
func readPriorGenerations(projectRoot, excludeRunDir string) ([]priorGeneration, int) {
	backupsDir := filepath.Join(projectRoot, defs.BackupsDir)
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return nil, 0
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	// Run backup directory names are timestamps, so a descending sort is newest first.
	sort.Sort(sort.Reverse(sort.StringSlice(names)))

	var gens []priorGeneration
	reads := 0
	for _, name := range names {
		dir := filepath.Join(backupsDir, name)
		if excludeRunDir != "" && filepath.Clean(dir) == filepath.Clean(excludeRunDir) {
			continue
		}
		path := filepath.Join(dir, priorManifestName)
		info, err := os.Lstat(path)
		if err != nil {
			continue // no generation here
		}
		if reads >= priorManifestGenerations {
			break
		}
		reads++
		if !info.Mode().IsRegular() {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var parsed struct {
			Files *map[string]priorFileJSON `json:"files"`
		}
		if err := json.Unmarshal(data, &parsed); err != nil || parsed.Files == nil {
			continue
		}
		g := priorGeneration{name: name, files: make(map[string]priorRecord, len(*parsed.Files))}
		for rel, f := range *parsed.Files {
			g.files[rel] = priorRecord{
				provenance:      f.Provenance,
				templateHash:    f.TemplateHash,
				deployedHash:    f.DeployedHash,
				deployedVersion: f.DeployedVersion,
			}
		}
		gens = append(gens, g)
	}
	return gens, reads
}
