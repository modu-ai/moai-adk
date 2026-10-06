package homestate

import (
	"embed"
	"strings"
)

// The shipped hub-file list (SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-020): the
// paths the factory hub chain orders by. The list is embedded DATA compiled
// into the product — the loader reads nothing but its own embedded file, so
// no user project ever reaches for a developer-machine tree, a SPEC
// directory, or a reports path to obtain it (design §7.4). The tracked
// baseline copy in this repository's SPEC directory is the measurement
// source the tests answer the embedded copy to; this file names no path to
// it and needs none.
//
//go:embed hub_files.txt
var hubFilesFS embed.FS

// HubFiles returns the embedded hub paths: one per line, '#' comments and
// blank lines ignored, order preserved. Deterministic — every call returns
// the same slice contents.
func HubFiles() []string {
	raw, err := hubFilesFS.ReadFile("hub_files.txt")
	if err != nil {
		// The data file is compiled in; a read failure is unreachable. An
		// empty list is the honest degradation — the chain simply forms no
		// ordering.
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}
