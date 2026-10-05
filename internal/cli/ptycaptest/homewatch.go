package ptycaptest

// Real HOME watch list for the pty capture tests (acceptance.md §B P8).
//
// A pty case runs product code in a child whose HOME is a temp dir, so the
// real HOME must come out untouched. Hashing the whole real ~/.moai tree is not
// a binary verdict — other live sessions write under it constantly — so the
// comparison is restricted to the files the flows under test (init, update,
// profile wizard, downgrade confirm) CAN write on the real HOME. Each row names
// the product function that writes it.

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	// watchBaseHome roots a row at the real HOME (H).
	watchBaseHome = "H"
	// watchBaseMoai roots a row at H/.moai (M).
	watchBaseMoai = "M"
	// watchKeyToken is replaced by the homestate.ProjectKey of each case
	// working directory.
	watchKeyToken = "{key}"
	// watchGlobPrefix marks the snapshot entry that records a glob's matches.
	watchGlobPrefix = "glob:"
	// watchAbsentMark is the fingerprint of a path that does not exist.
	watchAbsentMark = "absent"
)

type homeWatchKind int

const (
	// watchContent compares existence, size, and sha256.
	watchContent homeWatchKind = iota
	// watchDir compares existence and the immediate child names.
	watchDir
	// watchAbsent compares existence only.
	watchAbsent
)

type homeWatchItem struct {
	ID       string
	Base     string
	Rel      string // slash-separated, relative to Base
	Kind     homeWatchKind
	Glob     bool   // Rel is a one-level glob; compared as the union of matches
	Producer string // the product function that writes this path
}

// homeWatchList is the single watch-list constant (W1-W6).
var homeWatchList = []homeWatchItem{
	{ID: "W1", Base: watchBaseMoai, Rel: "claude-profiles/preferences.yaml", Kind: watchContent,
		Producer: "profile.WritePreferences -> GetPreferencesPath (default profile)"},
	{ID: "W1", Base: watchBaseMoai, Rel: "claude-profiles/.preferences.yaml", Kind: watchContent,
		Producer: "profile migrateOldFile (legacy file name, default profile)"},
	{ID: "W2", Base: watchBaseMoai, Rel: "claude-profiles/*/preferences.yaml", Kind: watchContent, Glob: true,
		Producer: "profile.WritePreferences -> GetPreferencesPath (named profile)"},
	{ID: "W2", Base: watchBaseMoai, Rel: "claude-profiles/*/.preferences.yaml", Kind: watchContent, Glob: true,
		Producer: "profile migrateOldFile (legacy file name, named profile)"},
	{ID: "W3", Base: watchBaseHome, Rel: ".claude/settings.json", Kind: watchContent,
		Producer: "project.ApplyAutonomyTierBundle (init), ensureGlobalSettingsEnv (init, update)"},
	{ID: "W4", Base: watchBaseHome, Rel: ".claude/hooks/moai", Kind: watchDir,
		Producer: "ensureGlobalSettingsEnv (removes the directory)"},
	{ID: "W5", Base: watchBaseHome, Rel: ".zshenv", Kind: watchContent, Producer: "shell selectConfigFile (init shell step, update shell env)"},
	{ID: "W5", Base: watchBaseHome, Rel: ".zshrc", Kind: watchContent, Producer: "shell selectConfigFile (init shell step, update shell env)"},
	{ID: "W5", Base: watchBaseHome, Rel: ".profile", Kind: watchContent, Producer: "shell selectConfigFile (init shell step, update shell env)"},
	{ID: "W5", Base: watchBaseHome, Rel: ".bash_profile", Kind: watchContent, Producer: "shell selectConfigFile (init shell step, update shell env)"},
	{ID: "W5", Base: watchBaseHome, Rel: ".bashrc", Kind: watchContent, Producer: "shell selectConfigFile (init shell step, update shell env)"},
	{ID: "W5", Base: watchBaseHome, Rel: ".config/fish/config.fish", Kind: watchContent, Producer: "shell selectConfigFile (init shell step, update shell env)"},
	{ID: "W6", Base: watchBaseMoai, Rel: "db/" + watchKeyToken, Kind: watchAbsent, Producer: "homestate.EnsureProjectLayout (init)"},
	{ID: "W6", Base: watchBaseMoai, Rel: "run/" + watchKeyToken, Kind: watchAbsent, Producer: "homestate.EnsureProjectLayout (init)"},
	{ID: "W6", Base: watchBaseMoai, Rel: "cache/search/" + watchKeyToken, Kind: watchAbsent, Producer: "homestate.EnsureProjectLayout (init)"},
}

// HomeSnapshot maps a HOME-relative path (or a glob: entry) to its fingerprint.
type HomeSnapshot map[string]string

// Existing counts the watched paths that exist (glob bookkeeping excluded).
func (s HomeSnapshot) Existing() int {
	n := 0
	for k, v := range s {
		if !strings.HasPrefix(k, watchGlobPrefix) && v != watchAbsentMark {
			n++
		}
	}
	return n
}

// Entries counts the watched paths (glob bookkeeping excluded).
func (s HomeSnapshot) Entries() int {
	n := 0
	for k := range s {
		if !strings.HasPrefix(k, watchGlobPrefix) {
			n++
		}
	}
	return n
}

// SnapshotHome fingerprints every watch-list path under the root home (M is
// home/.moai). keys are the project keys substituted for {key}. The root is an
// argument so the same function runs on the real HOME and on a fake one.
func SnapshotHome(home string, keys []string) (HomeSnapshot, error) {
	snap := HomeSnapshot{}
	for _, it := range homeWatchList {
		base := home
		if it.Base == watchBaseMoai {
			base = filepath.Join(home, ".moai")
		}
		rels := []string{it.Rel}
		if strings.Contains(it.Rel, watchKeyToken) {
			rels = rels[:0]
			for _, k := range keys {
				rels = append(rels, strings.ReplaceAll(it.Rel, watchKeyToken, k))
			}
		}
		for _, rel := range rels {
			abs := filepath.Join(base, filepath.FromSlash(rel))
			if it.Glob {
				matches, err := filepath.Glob(abs)
				if err != nil {
					return nil, fmt.Errorf("watch glob %s: %w", abs, err)
				}
				sort.Strings(matches)
				names := make([]string, len(matches))
				for i, m := range matches {
					names[i] = homeRel(home, m)
					fp, err := fingerprint(m, it.Kind)
					if err != nil {
						return nil, err
					}
					snap[names[i]] = fp
				}
				snap[watchGlobPrefix+homeRel(home, abs)] = strings.Join(names, ",")
				continue
			}
			fp, err := fingerprint(abs, it.Kind)
			if err != nil {
				return nil, err
			}
			snap[homeRel(home, abs)] = fp
		}
	}
	return snap, nil
}

func homeRel(home, abs string) string {
	if rel, err := filepath.Rel(home, abs); err == nil {
		return rel
	}
	return abs
}

func fingerprint(path string, kind homeWatchKind) (string, error) {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return watchAbsentMark, nil
	}
	if err != nil {
		return "", fmt.Errorf("watch %s: %w", path, err)
	}
	switch kind {
	case watchAbsent:
		return "present", nil
	case watchDir:
		if !info.IsDir() {
			return fmt.Sprintf("not-a-dir size=%d", info.Size()), nil
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", fmt.Errorf("watch %s: %w", path, err)
		}
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		sort.Strings(names)
		return "dir names=" + strings.Join(names, ","), nil
	default:
		if info.IsDir() {
			return "dir", nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("watch %s: %w", path, err)
		}
		return fmt.Sprintf("file size=%d sha256=%x", info.Size(), sha256.Sum256(b)), nil
	}
}

// DiffSnapshots lists every entry whose fingerprint differs between the two
// snapshots, over the union of their keys, as "path (before -> after)".
func DiffSnapshots(before, after HomeSnapshot) []string {
	keys := map[string]bool{}
	for k := range before {
		keys[k] = true
	}
	for k := range after {
		keys[k] = true
	}
	var changed []string
	for k := range keys {
		b, bok := before[k]
		a, aok := after[k]
		if !bok {
			b = watchAbsentMark
		}
		if !aok {
			a = watchAbsentMark
		}
		if b != a || bok != aok {
			changed = append(changed, fmt.Sprintf("%s (%s -> %s)", k, b, a))
		}
	}
	sort.Strings(changed)
	return changed
}

// WatchRealHome snapshots the watch list on the real HOME now and returns the
// check to call once the case is over. Both points print the entry count and
// how many entries exist on the real HOME. workDirs are the case working
// directories whose project keys fill the W6 rows.
func WatchRealHome(tb testing.TB, workDirs ...string) func() {
	tb.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		tb.Fatalf("real HOME: %v", err)
	}
	keys := make([]string, len(workDirs))
	for i, d := range workDirs {
		keys[i] = homestate.ProjectKey(d)
	}
	before, err := SnapshotHome(home, keys)
	if err != nil {
		tb.Fatalf("real HOME snapshot: %v", err)
	}
	tb.Logf("real HOME watch list: %d entries, %d present on the real HOME (before)", before.Entries(), before.Existing())
	return func() {
		tb.Helper()
		after, err := SnapshotHome(home, keys)
		if err != nil {
			tb.Fatalf("real HOME snapshot: %v", err)
		}
		tb.Logf("real HOME watch list: %d entries, %d present on the real HOME (after)", after.Entries(), after.Existing())
		if changed := DiffSnapshots(before, after); len(changed) > 0 {
			tb.Errorf("real HOME watch list changed during the case: %v", changed)
		}
	}
}
