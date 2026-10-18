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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// priorManifestMarkerCopy gives a new sibling run directory the run marker of the
// directory it stands beside. Like the two observation points above it is a
// variable only so a test can inject a failure; no production path replaces it.
var priorManifestMarkerCopy = copyRunMarker

// priorManifestWriter is how the Backup stage preserves the manifest. Its
// default is writePriorManifest; a test wraps it to inject a failure.
var priorManifestWriter = writePriorManifest

// writePriorManifestTemp writes data to a temporary file in dir (a name other
// than the final one, so a half-written file never answers to it) and syncs it.
func writePriorManifestTemp(dir string, data []byte) (string, error) {
	f, err := os.CreateTemp(dir, backup.PriorManifestTempPrefix+"*")
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
// it either holds every byte or does not exist. On any failure after the
// temporary file exists it is removed, best effort: the reader ignores its name,
// but a legacy restore copies every file at the backup root, so a leftover would
// land in the project. The failure itself is reported unchanged.
func writePriorManifest(runDir string, data []byte) error {
	tmp, err := priorManifestTempWrite(runDir, data)
	if err != nil {
		if tmp != "" {
			_ = os.Remove(tmp)
		}
		return fmt.Errorf("preserve prior manifest: %w", err)
	}
	final := filepath.Join(runDir, priorManifestName)
	if err := priorManifestRename(tmp, final); err != nil {
		_ = os.Remove(tmp)
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
	// Two runs that start in one clock second are handed the same directory name.
	// The copy of an earlier run is never replaced (REQ-OWR-017): this run's copy
	// goes to a directory no other run writes to.
	if runDir, err = claimPriorManifestDir(runDir); err != nil {
		return configBackupPath, priorManifestFailed, err
	}
	if err := priorManifestWriter(runDir, data); err != nil {
		return runDir, priorManifestFailed, err
	}
	return runDir, priorManifestWritten, nil
}

// priorManifestSiblingMax is the highest collision suffix claimPriorManifestDir
// tries. The suffix is two digits wide so that directory-name order stays age
// order (the reader sorts by name).
const priorManifestSiblingMax = 99

// claimPriorManifestDir returns the directory this run's prior-manifest copy goes
// to. A run writes its copy once, so a manifest.prior.json already in runDir is an
// earlier run's: the copy then goes to a new sibling `<runDir>-NN` (NN = 02..99),
// created with Mkdir so a directory that already exists, whoever made it, is
// never written into. The sibling carries runDir's run marker, so retention
// counts both directories as the one run it already counted. A sibling that
// cannot be given its marker would be a directory retention never counts or
// prunes, so it is removed again (it is still empty) and the claim fails. Failing
// to find or make a free sibling is an error too: the caller then blocks every
// removal.
func claimPriorManifestDir(runDir string) (string, error) {
	runDir = filepath.Clean(runDir) // a trailing separator must not turn `<dir>/-02` into a child
	_, err := os.Lstat(filepath.Join(runDir, priorManifestName))
	if errors.Is(err, fs.ErrNotExist) {
		return runDir, nil
	}
	if err != nil {
		return "", fmt.Errorf("look for an earlier prior manifest in %s: %w", runDir, err)
	}
	for n := 2; n <= priorManifestSiblingMax; n++ {
		sibling := fmt.Sprintf("%s-%02d", runDir, n)
		if err := os.Mkdir(sibling, defs.DirPerm); err != nil {
			if errors.Is(err, fs.ErrExist) {
				continue
			}
			return "", fmt.Errorf("create run backup directory %s: %w", sibling, err)
		}
		if err := priorManifestMarkerCopy(runDir, sibling); err != nil {
			_ = os.Remove(sibling) // best effort; os.Remove refuses a directory that is not empty
			return "", fmt.Errorf("give %s the run marker of %s: %w", sibling, runDir, err)
		}
		return sibling, nil
	}
	return "", fmt.Errorf("no free sibling name for %s up to -%02d", runDir, priorManifestSiblingMax)
}

// copyRunMarker gives dst the run marker src carries. A src that has no marker
// (absent, or not a regular file) has nothing to copy: that is not an error, and
// the sibling then stays without one, as it always did. A marker that exists but
// cannot be read or written is an error — retention would otherwise never count
// the sibling.
func copyRunMarker(src, dst string) error {
	path := filepath.Join(src, backup.RunMarkerName)
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	line, _, _ := strings.Cut(string(data), "\n")
	return backup.WriteRunMarker(dst, strings.TrimSpace(line))
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
