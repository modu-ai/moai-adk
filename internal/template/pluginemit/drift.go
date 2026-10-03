// drift.go — compare a committed tree with an emission, and regenerate it
// (REQ-009).
//
// Drift is the read-only check behind `make plugin-emit-check`: it reports a
// byte difference, a mode difference, a missing file and an extra file, and
// never writes. Write is the one regeneration path, reached only through
// `make plugin-emit`.
package pluginemit

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
)

// Difference kinds reported by Drift.
const (
	DriftBytes   = "bytes"
	DriftMode    = "mode"
	DriftMissing = "missing"
	DriftExtra   = "extra"
)

// Difference is one way a committed tree departs from an emission.
type Difference struct {
	Kind string
	Path string
}

// generatedRoots are the directories the generator owns outright: any file
// found inside one that the emission does not carry is an extra file. The
// repository root itself holds much else and is never scanned.
var generatedRoots = []string{
	PluginRoot,
	path.Dir(ClaudeMarketplacePath),
	path.Dir(CodexMarketplacePath),
}

// Drift compares the files committed under root with pub. It is read-only.
//
// Modes are compared as "executable or not": git records only 0644 and 0755,
// and a checkout applies the user's umask to the rest, so the execute bit is
// the one part of the mode a committed file reliably carries. Windows has no
// execute bit, so the mode comparison is not made there.
//
// @MX:NOTE: the read-only drift gate behind `make plugin-emit-check`; it must never write, because a write here would regenerate silently and erase the evidence CI needs
func Drift(pub *Publication, root string) ([]Difference, error) {
	var diffs []Difference
	for _, p := range sortedPaths(pub.Files) {
		full := filepath.Join(root, filepath.FromSlash(p))
		data, err := os.ReadFile(full)
		if errors.Is(err, fs.ErrNotExist) {
			diffs = append(diffs, Difference{DriftMissing, p})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("pluginemit: read committed %s: %w", p, err)
		}
		if !bytes.Equal(data, pub.Files[p]) {
			diffs = append(diffs, Difference{DriftBytes, p})
		}
		if runtime.GOOS == "windows" {
			continue
		}
		info, err := os.Stat(full)
		if err != nil {
			return nil, fmt.Errorf("pluginemit: stat committed %s: %w", p, err)
		}
		if (info.Mode().Perm()&0o111 != 0) != (pub.Modes[p]&0o111 != 0) {
			diffs = append(diffs, Difference{DriftMode, p})
		}
	}

	extras, err := extraFiles(pub, root)
	if err != nil {
		return nil, err
	}
	diffs = append(diffs, extras...)
	sort.SliceStable(diffs, func(i, j int) bool { return diffs[i].Path < diffs[j].Path })
	return diffs, nil
}

// extraFiles lists the files inside the generated roots that pub does not carry.
func extraFiles(pub *Publication, root string) ([]Difference, error) {
	var extras []Difference
	for _, dir := range generatedRoots {
		top := filepath.Join(root, filepath.FromSlash(dir))
		err := filepath.WalkDir(top, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) && p == top {
					return nil
				}
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if _, ok := pub.Files[rel]; !ok {
				extras = append(extras, Difference{DriftExtra, rel})
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("pluginemit: scan %s: %w", dir, err)
		}
	}
	return extras, nil
}

// Write materialises pub under root: every file is written and set to its mode
// (an explicit Chmod, because WriteFile leaves the mode of an existing file and
// applies the umask to a new one), and a file inside the generated roots that
// pub does not carry is removed together with any directory it leaves empty.
func Write(pub *Publication, root string) error {
	for _, p := range sortedPaths(pub.Files) {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return fmt.Errorf("pluginemit: mkdir for %s: %w", p, err)
		}
		if err := os.WriteFile(full, pub.Files[p], pub.Modes[p]); err != nil {
			return fmt.Errorf("pluginemit: write %s: %w", p, err)
		}
		if err := os.Chmod(full, pub.Modes[p]); err != nil {
			return fmt.Errorf("pluginemit: chmod %s: %w", p, err)
		}
	}
	extras, err := extraFiles(pub, root)
	if err != nil {
		return err
	}
	for _, x := range extras {
		full := filepath.Join(root, filepath.FromSlash(x.Path))
		if err := os.Remove(full); err != nil {
			return fmt.Errorf("pluginemit: remove %s: %w", x.Path, err)
		}
		// os.Remove refuses a non-empty directory, which ends the climb.
		for dir := filepath.Dir(full); dir != root && os.Remove(dir) == nil; dir = filepath.Dir(dir) {
		}
	}
	return nil
}

func sortedPaths(files map[string][]byte) []string {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}
