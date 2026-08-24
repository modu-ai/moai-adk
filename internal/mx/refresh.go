package mx

import (
	"os"
	"path/filepath"
	"time"
)

// RefreshStats reports what one incremental refresh actually did. FilesParsed
// counts the files whose content was re-read AND re-parsed (change
// consumption); change DETECTION additionally hashes every walked file, which
// is the cost the update budget measures.
type RefreshStats struct {
	FilesParsed     int           // files re-parsed (changed or new)
	ChangedDetected int           // files whose hash moved vs the inventory (incl. new)
	RemovedDetected int           // inventoried files that vanished
	FullRescan      bool          // wrong-tree/absent provenance forced a full pass
	Duration        time.Duration // measured wall cost of the refresh
}

// RefreshIndex brings the mx sidecar up to date with the working tree,
// re-parsing ONLY files whose content hash differs from the stamped inventory
// (plus files the inventory has never seen), and dropping the tags of files
// that vanished (REQ-GF-007). No LLM, no network — mechanical paths only.
//
// Per-tree anchoring (REQ-GF-008): when the loaded provenance names a
// different tree root, nothing is trusted incrementally — the refresh becomes
// a full rescan and the sidecar is re-anchored to THIS tree. Two worktrees of
// one repository therefore never share refresh state.
//
// The sidecar is left untouched when nothing changed (zero-change refresh is
// a no-op write).
//
// @MX:NOTE: [AUTO] RefreshIndex — detection hashes every walked file but parses only changed ones; FilesParsed is the changed-files-only observable
func RefreshIndex(stateDir, scanRoot string, ignore []string) (*RefreshStats, error) {
	start := time.Now()
	stats := &RefreshStats{}

	if ignore == nil {
		ignore = DefaultScanIgnore
	}

	mgr := NewManager(stateDir)
	sidecar, err := mgr.Load()
	if err != nil {
		return nil, err
	}

	// Anchor check: no provenance, or provenance from a different tree ⇒ the
	// inventory cannot be trusted for this tree; rescan everything.
	trusted := sidecar != nil && sidecar.Provenance != nil &&
		sidecar.Provenance.TreeRoot == scanRoot && len(sidecar.Provenance.FileInventory) > 0
	stats.FullRescan = !trusted

	var oldInventory map[string]string
	if trusted {
		oldInventory = sidecar.Provenance.FileInventory
	}

	newInventory := make(map[string]string)
	parsedTagsByFile := make(map[string][]Tag)

	s := NewScanner()
	s.SetIgnorePatterns(ignore)

	err = walkScanFiles(scanRoot, ignore, func(absPath, rel string) error {
		sum, err := HashFile(absPath)
		if err != nil {
			// Unreadable file: not inventoried, not parsed (fail-open, the
			// scanner's own posture).
			return nil
		}
		newInventory[rel] = sum
		if trusted {
			if old, ok := oldInventory[rel]; ok && old == sum {
				return nil // unchanged — detection read only, no parse
			}
		}
		stats.ChangedDetected++
		tags, err := s.ScanFile(absPath)
		if err != nil {
			s.errors = append(s.errors, err.Error())
			return nil
		}
		stats.FilesParsed++
		parsedTagsByFile[absPath] = tags
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Vanished files: inventoried before, not walked now.
	var vanished []string
	if trusted {
		for rel := range oldInventory {
			if _, ok := newInventory[rel]; !ok {
				vanished = append(vanished, rel)
			}
		}
	}
	stats.RemovedDetected = len(vanished)

	// Assemble the refreshed tag set: keep tags of unchanged files verbatim,
	// take the freshly parsed tags for changed/new files, drop vanished.
	now := time.Now()
	var tags []Tag
	for _, tag := range sidecar.Tags {
		rel, relErr := filepath.Rel(scanRoot, tag.File)
		if relErr != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		newSum, stillExists := newInventory[rel]
		if !stillExists {
			continue // vanished
		}
		if _, reparsed := parsedTagsByFile[tag.File]; reparsed {
			continue // replaced below
		}
		_ = newSum
		tags = append(tags, tag)
	}
	for _, parsed := range parsedTagsByFile {
		for i := range parsed {
			parsed[i].LastSeenAt = now
		}
		tags = append(tags, parsed...)
	}

	stats.Duration = time.Since(start)

	if stats.ChangedDetected == 0 && stats.RemovedDetected == 0 && trusted {
		// Zero-change no-op: leave the sidecar (and its provenance) untouched.
		return stats, nil
	}

	refreshed := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags:          tags,
		ScannedAt:     now,
		Provenance:    StampMXScan(scanRoot, newInventory),
	}
	if err := mgr.Write(refreshed); err != nil {
		return nil, err
	}
	return stats, nil
}

// walkScanFiles walks scanRoot applying the scanner's ignore semantics
// (directory-base and relative-path pattern matching) and invokes fn for each
// non-ignored regular file with both absolute and slash-relative paths.
func walkScanFiles(scanRoot string, ignore []string, fn func(absPath, rel string) error) error {
	return filepath.Walk(scanRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			for _, pattern := range ignore {
				if matched, mErr := filepath.Match(pattern, filepath.Base(path)); mErr == nil && matched {
					return filepath.SkipDir
				}
			}
			return nil
		}
		rel, relErr := filepath.Rel(scanRoot, path)
		if relErr != nil {
			return relErr
		}
		for _, pattern := range ignore {
			if matched, mErr := filepath.Match(pattern, rel); mErr == nil && matched {
				return nil
			}
		}
		return fn(path, filepath.ToSlash(rel))
	})
}
