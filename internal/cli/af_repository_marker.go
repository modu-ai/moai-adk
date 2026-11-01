package cli

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// afMarkerFile is the name of the repository identity marker inside the common git directory of the repository.
//
// A ledger is bound to the path of a git directory, and a path is not a repository: a clone, a mirror, an unrelated
// repository or a repository made afterwards can stand at it, or at the path a move names (audit of a36ea8c368, N61 N65
// N67). The marker is what the repository itself can say about it. It is a random 128-bit token written once, with
// exclusive create, into the common git directory (a linked worktree shares it), mode 0600. A rename or a move of the
// directory carries it; so does a plain copy of the directory, which is a continuation only once the original is gone
// (the old git directory still existing already stops a live copy); `git clone`, `--mirror` and a new repository never
// have it. The ledger records the token (the ledger runs no git, REQ-AFL-008) and every verb compares it.
const afMarkerFile = "moai-af-repository-id"

var afMarkerShape = regexp.MustCompile(`^[0-9a-f]{32}$`)

// afMarkerPrefix is the part of a marker a message may print: enough to tell two apart, never the value.
func afMarkerPrefix(marker string) string {
	if len(marker) > 8 {
		return marker[:8]
	}
	return marker
}

// afReadMarker is the marker of the git directory gitDir, "" when it has none. A file that does not hold a marker is an
// error and not "none": the repository was marked by someone, and a damaged mark cannot be told from a forged absence.
func afReadMarker(gitDir string) (string, error) {
	path := filepath.Join(gitDir, afMarkerFile)
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	marker := strings.TrimSpace(string(b))
	if !afMarkerShape.MatchString(marker) {
		return "", fmt.Errorf("%s does not hold a repository marker", path)
	}
	return marker, nil
}

// afEnsureMarker is the marker of the git directory gitDir, created when it has none. The content is written to a file
// of its own and then linked to the final name, so the name appears complete or not at all and a creator that loses the
// race reads the winner's value: concurrent creators converge on one marker.
func afEnsureMarker(gitDir string) (string, error) {
	if marker, err := afReadMarker(gitDir); err != nil || marker != "" {
		return marker, err
	}
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	marker := hex.EncodeToString(raw[:])
	tmp, err := os.CreateTemp(gitDir, afMarkerFile+".tmp-*") // mode 0600
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.WriteString(marker + "\n"); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Link(tmp.Name(), filepath.Join(gitDir, afMarkerFile)); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return afReadMarker(gitDir) // another creator won; its value is complete
		}
		return "", err
	}
	return marker, nil
}
