// todo_merge_backup.go — backup, restore, and ordering evidence for the queue
// merge (SPEC-TODO-QUEUE-HOME-MERGE-001 M2, plan.md §F).
//
// REQ-TQM-001..003: before any store mutation, every present queue artifact
// of BOTH stores is copied byte-for-byte and verified by byte count and
// SHA-256 equality; a failed verification aborts with nothing deleted. The
// ordering evidence (AC-TQM-001) is produced by the tooling itself: a path-set
// + mtime snapshot of each store directory before the backup and after
// hash-verification, with the comparison — not a hand-written listing —
// proving zero mutation across the window.
//
// The restore half is the rehearsed rollback (REQ-TQM-015/016): reverse copy
// plus re-hash, exercised on fixture stores only.
package kanban

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// mergeStoreArtifacts enumerates the queue artifacts a store directory may
// carry (REQ-TQM-001's enumeration): the engine database, its WAL/SHM
// siblings, the legacy backlog.json compatibility document, and the one-time
// migration quarantine. Absent entries are recorded as absent, never
// invented.
var mergeStoreArtifacts = []string{
	"backlog.db",
	"backlog.db-wal",
	"backlog.db-shm",
	backlogFileName,
	backlogFileName + backlogMigratedSuffix,
}

// BackedUpArtifact is one verified backup copy.
type BackedUpArtifact struct {
	Store        string `json:"store"`         // absolute store directory
	Name         string `json:"name"`          // artifact base name
	SourcePath   string `json:"source_path"`   // absolute source path
	BackupPath   string `json:"backup_path"`   // absolute copy path
	SourceSHA256 string `json:"source_sha256"` // SHA-256 measured on the source
	CopySHA256   string `json:"copy_sha256"`   // SHA-256 measured on the copy
	Bytes        int64  `json:"bytes"`         // byte count of the copy
}

// AbsentArtifact records an enumerated artifact that does not exist in its
// store (acceptance.md §D.5 empty/absent case).
type AbsentArtifact struct {
	Store string `json:"store"`
	Name  string `json:"name"`
}

// QueueBackup is one verified backup pass's record.
type QueueBackup struct {
	CreatedAt time.Time          `json:"created_at"`
	Artifacts []BackedUpArtifact `json:"artifacts"`
	Absent    []AbsentArtifact   `json:"absent"`
}

// copyFilePlain copies src to dst by bytes. It is a package variable so the
// tamper test can interpose a corrupted copy and prove the hash gate fires.
var copyFilePlain = func(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// BackupQueueArtifacts backs up every present queue artifact of the given
// store directories into per-store subdirectories under backupDir, verifying
// each copy by byte count and SHA-256 equality with its source (REQ-TQM-001).
// Any verification failure aborts the whole pass with an error (REQ-TQM-002)
// — the caller treats a non-nil error as "no backup exists", never as a
// partial one.
func BackupQueueArtifacts(storeDirs []string, backupDir string) (*QueueBackup, error) {
	backup := &QueueBackup{CreatedAt: time.Now().UTC()}
	for i, storeDir := range storeDirs {
		abs, err := filepath.Abs(storeDir)
		if err != nil {
			return nil, fmt.Errorf("backup store %s: %w", storeDir, err)
		}
		label := fmt.Sprintf("store-%d", i+1)
		targetDir := filepath.Join(backupDir, label)
		if err := os.MkdirAll(targetDir, 0o700); err != nil {
			return nil, fmt.Errorf("backup store %s: %w", abs, err)
		}
		for _, name := range mergeStoreArtifacts {
			source := filepath.Join(abs, name)
			info, err := os.Stat(source)
			if err != nil {
				if !os.IsNotExist(err) {
					return nil, fmt.Errorf("backup store %s: stat %s: %w", abs, name, err)
				}
				backup.Absent = append(backup.Absent, AbsentArtifact{Store: abs, Name: name})
				continue
			}
			if info.IsDir() {
				return nil, fmt.Errorf("backup store %s: %s is a directory, refusing", abs, name)
			}
			backupPath := filepath.Join(targetDir, name)
			if err := copyFilePlain(source, backupPath); err != nil {
				return nil, fmt.Errorf("backup store %s: copy %s: %w", abs, name, err)
			}
			sourceSum, err := fileSHA256(source)
			if err != nil {
				return nil, fmt.Errorf("backup store %s: hash source %s: %w", abs, name, err)
			}
			copySum, err := fileSHA256(backupPath)
			if err != nil {
				return nil, fmt.Errorf("backup store %s: hash copy %s: %w", abs, name, err)
			}
			copyInfo, err := os.Stat(backupPath)
			if err != nil {
				return nil, fmt.Errorf("backup store %s: stat copy %s: %w", abs, name, err)
			}
			if sourceSum != copySum {
				return nil, fmt.Errorf("backup store %s: artifact %s FAILED hash verification (source %s, copy %s); aborting before any store mutation (REQ-TQM-002)", abs, name, sourceSum, copySum)
			}
			backup.Artifacts = append(backup.Artifacts, BackedUpArtifact{
				Store:        abs,
				Name:         name,
				SourcePath:   source,
				BackupPath:   backupPath,
				SourceSHA256: sourceSum,
				CopySHA256:   copySum,
				Bytes:        copyInfo.Size(),
			})
		}
	}
	return backup, nil
}

// RestoreQueueArtifacts copies every backed-up artifact back to its source
// path and re-verifies each restore by SHA-256 equality with the recorded
// backup hash (REQ-TQM-016). It restores exactly the backed-up set — it does
// not sweep the store directories.
func RestoreQueueArtifacts(backup *QueueBackup) error {
	if backup == nil {
		return fmt.Errorf("restore queue artifacts: no backup given")
	}
	for _, a := range backup.Artifacts {
		if err := copyFilePlain(a.BackupPath, a.SourcePath); err != nil {
			return fmt.Errorf("restore %s: copy: %w", a.SourcePath, err)
		}
		restored, err := fileSHA256(a.SourcePath)
		if err != nil {
			return fmt.Errorf("restore %s: hash: %w", a.SourcePath, err)
		}
		if restored != a.SourceSHA256 {
			return fmt.Errorf("restore %s: hash mismatch (backup %s, restored %s)", a.SourcePath, a.SourceSHA256, restored)
		}
	}
	return nil
}

// StoreArtifactState is one file observed in a store directory snapshot.
type StoreArtifactState struct {
	Size    int64
	ModTime time.Time
}

// StoreDirSnapshot maps a relative path to its observed state.
type StoreDirSnapshot map[string]StoreArtifactState

// StoreMutationKind classifies one observed difference between snapshots.
type StoreMutationKind string

const (
	StoreMutationAdded    StoreMutationKind = "added"
	StoreMutationRemoved  StoreMutationKind = "removed"
	StoreMutationModified StoreMutationKind = "modified"
)

// StoreDirMutation is one path-set/mtime difference between two snapshots of
// the same store directory.
type StoreDirMutation struct {
	Path string
	Kind StoreMutationKind
}

// SnapshotStoreDirState enumerates every regular file under dir with its size
// and mtime. Read-only: it moves nothing.
func SnapshotStoreDirState(dir string) (StoreDirSnapshot, error) {
	snap := StoreDirSnapshot{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		snap[rel] = StoreArtifactState{Size: info.Size(), ModTime: info.ModTime()}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("snapshot store dir %s: %w", dir, err)
	}
	return snap, nil
}

// StoreDirMutations compares two snapshots and reports every difference in
// deterministic path order. An empty result asserts zero mutation between the
// two observation points — the AC-TQM-001 ordering evidence, produced by this
// comparison rather than by hand.
func StoreDirMutations(before, after StoreDirSnapshot) []StoreDirMutation {
	var muts []StoreDirMutation
	seen := map[string]bool{}
	for path, b := range before {
		seen[path] = true
		a, ok := after[path]
		if !ok {
			muts = append(muts, StoreDirMutation{Path: path, Kind: StoreMutationRemoved})
			continue
		}
		if b.Size != a.Size || !b.ModTime.Equal(a.ModTime) {
			muts = append(muts, StoreDirMutation{Path: path, Kind: StoreMutationModified})
		}
	}
	for path := range after {
		if !seen[path] {
			muts = append(muts, StoreDirMutation{Path: path, Kind: StoreMutationAdded})
		}
	}
	sort.Slice(muts, func(i, j int) bool { return muts[i].Path < muts[j].Path })
	return muts
}

// fileSHA256 digests a file's bytes.
func fileSHA256(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return sha256Hex(raw), nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
