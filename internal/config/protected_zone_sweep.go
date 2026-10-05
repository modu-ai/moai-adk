package config

import (
	"fmt"
	"io/fs"
)

// ZoneSweep is the result of checking a manifest file's entries against a tree.
type ZoneSweep struct {
	// Resolved counts the paths entries that matched an existing path.
	Resolved int
	// Skipped counts the runtime_paths entries, which are exempt per entry.
	Skipped int
	// Dead lists the paths entries that matched nothing: the zone has silently shrunk.
	Dead []ZoneEntry
}

// SweepZoneEntries checks the entries that came from source against tree: a
// paths entry must match something in the tree its file ships in or is swept
// against, while a runtime_paths entry is exempt because it is created at init,
// update or run time. The sweep is how a renamed or removed file turns a manifest
// entry into a dead one without anybody noticing.
func SweepZoneEntries(z ProtectedZone, source string, tree fs.FS) (ZoneSweep, error) {
	var out ZoneSweep
	var live []string
	err := fs.WalkDir(tree, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", p, err)
		}
		if p == "." {
			return nil
		}
		folded := FoldZoneText(p)
		live = append(live, folded)
		if d.IsDir() {
			// A directory counts as present even while empty.
			live = append(live, folded+"/")
		}
		return nil
	})
	if err != nil {
		return out, err
	}
	for _, e := range z.Entries {
		if e.Source != source {
			continue
		}
		if e.Runtime {
			out.Skipped++
			continue
		}
		matched := false
		for _, p := range live {
			if e.Match(p) {
				matched = true
				break
			}
		}
		if matched {
			out.Resolved++
		} else {
			out.Dead = append(out.Dead, e)
		}
	}
	return out, nil
}
