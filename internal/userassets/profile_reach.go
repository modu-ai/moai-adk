package userassets

// Card t1702, round 7: the read-only profile-reach judgment.
//
// A profile-launched session reads user skills and agents from its own profile
// directory, not from ~/.claude. Before a project copy of a user asset is
// removed, the removal path asks whether every real profile root already holds
// the same copy. The operator's ruling (relayed 11-06): the answer comes from
// READING the profile roots; nothing is written to one, and a doubt keeps the
// project copy.

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ReachState is the answer for one profile root.
type ReachState int

const (
	// ReachSame: the profile root holds a regular file whose content hash equals
	// the project copy's.
	ReachSame ReachState = iota + 1
	// ReachAbsent: the profile (or its skills/agents root) exists and the copy
	// is not there.
	ReachAbsent
	// ReachDiffers: a regular file is there and its content differs.
	ReachDiffers
	// ReachUnknown: the answer could not be established - an unreadable path, a
	// link, a non-regular file, a link that would leave the root, an oversized
	// file. Treated like Differs by the removal decision.
	ReachUnknown
)

func (s ReachState) String() string {
	switch s {
	case ReachSame:
		return "same"
	case ReachAbsent:
		return "absent"
	case ReachDiffers:
		return "differs"
	case ReachUnknown:
		return "unknown"
	}
	return fmt.Sprintf("ReachState(%d)", int(s))
}

// maxReachFileBytes bounds the read of a profile copy; a larger file is
// Unknown rather than hashed (a skill or agent file is small text).
const maxReachFileBytes = 8 << 20

// ProfileReach is the read-only answer for ONE profile and ONE project copy.
type ProfileReach struct {
	Profile string     // profile directory name
	Root    string     // absolute profile skills|agents root that was read
	State   ReachState // what was found
	Reason  string     // one line for the report
}

// CheckProfileReach reads, for every real profile (the shared junk-filtered
// enumeration), the profile root of slug and compares the file rel under it
// with want, the content hash of the project copy. It writes nothing and
// creates nothing. Only the two profile-derivable bases (claude-skills,
// claude-agents) have a profile root; any other slug yields no rows, and an
// empty result never licenses a removal (SafeToRemoveProjectCopy). An
// unreadable profile store is an error, never an empty answer. rel is a
// root-relative path and may not escape.
func CheckProfileReach(home string, slug RootSlug, rel string, want [32]byte) ([]ProfileReach, error) {
	clean, err := ValidateRelPath(rel)
	if err != nil {
		return nil, fmt.Errorf("profile reach: %q: %w", rel, err)
	}
	if !profileDerivable(slug) {
		return nil, nil
	}
	names, err := RealProfileDirNames(home)
	if err != nil {
		return nil, err
	}
	var out []ProfileReach
	for _, name := range names {
		rootRel, ok := config.UserRootSlugRel(string(slug) + "@" + name)
		if !ok {
			continue // unreachable: name passed the same filter upstream
		}
		dir := filepath.Join(home, filepath.FromSlash(rootRel))
		state, reason := readProfileCopy(dir, clean, want)
		out = append(out, ProfileReach{Profile: name, Root: dir, State: state, Reason: reason})
	}
	return out, nil
}

func profileDerivable(slug RootSlug) bool {
	for _, b := range config.UserRootProfileBases() {
		if string(slug) == b.Slug {
			return true
		}
	}
	return false
}

// readProfileCopy answers for one profile root. Every doubt is Unknown.
func readProfileCopy(dir, rel string, want [32]byte) (ReachState, string) {
	info, err := os.Lstat(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return ReachAbsent, "profile root does not exist"
	case err != nil:
		return ReachUnknown, "profile root not readable: " + err.Error()
	case !info.IsDir():
		return ReachUnknown, "profile root is not a plain directory (a link or a file)"
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return ReachUnknown, "profile root cannot be opened: " + err.Error()
	}
	defer root.Close()
	fi, err := root.Lstat(rel)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return ReachAbsent, "no such file under the profile root"
	case err != nil:
		return ReachUnknown, "path not readable inside the root: " + err.Error()
	case !fi.Mode().IsRegular():
		return ReachUnknown, "path is not a regular file (" + fi.Mode().Type().String() + ")"
	case fi.Size() > maxReachFileBytes:
		return ReachUnknown, "file larger than the read bound"
	}
	f, err := root.Open(rel)
	if err != nil {
		return ReachUnknown, "file cannot be opened: " + err.Error()
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxReachFileBytes+1))
	if err != nil {
		return ReachUnknown, "file read failed: " + err.Error()
	}
	if len(data) > maxReachFileBytes {
		return ReachUnknown, "file larger than the read bound"
	}
	if sha256.Sum256(data) == want {
		return ReachSame, "content hash equals the project copy"
	}
	return ReachDiffers, "content hash differs from the project copy"
}

// SafeToRemoveProjectCopy is true only when there is at least one profile row
// and every row is ReachSame. Zero rows, Absent, Differs, Unknown, and any
// state this code does not know return false: the project copy stays.
func SafeToRemoveProjectCopy(reaches []ProfileReach) bool {
	if len(reaches) == 0 {
		return false
	}
	for _, r := range reaches {
		if r.State != ReachSame {
			return false
		}
	}
	return true
}
