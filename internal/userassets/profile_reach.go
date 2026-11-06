package userassets

// Card t1702, round 7: the read-only profile-reach judgment.
//
// A profile-launched session reads user skills and agents from its own profile
// directory, not from ~/.claude. Before a project copy of a user asset is
// removed, the removal path asks whether every real profile root already holds
// the same copy. Operator decision d-20261106T055331Z-c28f ("install writes,
// the judgment only reads"): the install and update paths keep writing to
// profile roots; THIS judgment only reads them (os.Root reads, no writes, no
// directory creation), and a doubt keeps the project copy.

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

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

// afterReachLstat is a test seam between the Lstat of the copy and its open:
// the window a concurrent actor with write access to the user's own profile
// directory could use to swap the name (audit iter7 N4). It does nothing in
// production.
var afterReachLstat = func(dir, rel string) {}

// maxReachFileBytes bounds the read of a profile copy; a larger file is
// Unknown rather than hashed (a skill or agent file is small text).
const maxReachFileBytes = 8 << 20

// ProfileReach is the read-only answer for ONE profile and ONE project copy.
type ProfileReach struct {
	Profile string     // profile directory name
	Root    string     // absolute profile skills|agents root that was read; the profile directory itself for an entry that was not read
	State   ReachState // what was found
	Reason  string     // one line for the report
}

// CheckProfileReach lists the profile store and, for every real profile (the
// shared junk-filtered enumeration), reads the profile root of slug and
// compares the file rel under it with want, the content hash of the project
// copy. A store entry the launcher could start but that enumeration drops (a
// linked directory, a name with a space, the reserved sentinel) is not read and
// is reported as an Unknown row; plain files, hidden names and dangling links
// give no row. It writes nothing and creates nothing. Only the two profile-derivable bases (claude-skills,
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
	base := ProfileRootsBase(home)
	entries, err := os.ReadDir(base)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil // no profile store: no profile to read
	}
	if err != nil {
		return nil, fmt.Errorf("read profile store %s: %w", base, err)
	}
	var out []ProfileReach
	for _, e := range entries { // ReadDir order is sorted by name
		name := e.Name()
		profileDir := filepath.Join(base, name)
		if e.IsDir() && config.ValidProfileDirName(name) {
			rootRel, ok := config.UserRootSlugRel(string(slug) + "@" + name)
			if !ok {
				continue // unreachable: name passed the same filter
			}
			dir := filepath.Join(home, filepath.FromSlash(rootRel))
			// The entry was a directory when listed; if it has become a link
			// since, reading through it could leave the store.
			if pi, lerr := os.Lstat(profileDir); lerr != nil || !pi.IsDir() {
				out = append(out, ProfileReach{Profile: name, Root: dir, State: ReachUnknown,
					Reason: "profile directory changed since it was listed"})
				continue
			}
			state, reason := readProfileCopy(dir, clean, want)
			out = append(out, ProfileReach{Profile: name, Root: dir, State: state, Reason: reason})
			continue
		}
		// Not listed by the shared install enumeration. If the launcher could
		// still start it (a linked directory, a name with a space, the reserved
		// sentinel), its roots were never read: Unknown, never a silent skip.
		if !config.LauncherProfileDirName(name) {
			continue
		}
		si, serr := os.Stat(profileDir)
		switch {
		case serr == nil && si.IsDir():
			out = append(out, ProfileReach{Profile: name, Root: profileDir, State: ReachUnknown,
				Reason: "startable profile directory outside the install enumeration (link, space or reserved name): not read"})
		case serr == nil || errors.Is(serr, fs.ErrNotExist):
			// a plain file, or a dangling link: not a profile the launcher can use
		default:
			out = append(out, ProfileReach{Profile: name, Root: profileDir, State: ReachUnknown,
				Reason: "store entry not inspectable: " + serr.Error()})
		}
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
	afterReachLstat(dir, rel)
	f, err := openRegularForRead(root, rel)
	if err != nil {
		return ReachUnknown, "file cannot be opened: " + err.Error()
	}
	defer f.Close()
	// The open follows a link that stays inside the root: the handle must be the
	// very file the Lstat inspected (audit iter7 N4).
	if hs, herr := f.Stat(); herr != nil || !os.SameFile(fi, hs) {
		return ReachUnknown, "the opened file is not the one that was inspected"
	}
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

// openRegularForRead opens rel for reading without blocking and confirms on
// the OPEN HANDLE that it is a regular file (audit iter6 M2): an earlier Lstat
// says nothing about what the name is by the time it is opened, and a blocking
// open of a FIFO waits for a writer that never comes.
func openRegularForRead(root *os.Root, rel string) (*os.File, error) {
	f, err := root.OpenFile(rel, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if !fi.Mode().IsRegular() {
		_ = f.Close()
		return nil, fmt.Errorf("%s is not a regular file (%s)", rel, fi.Mode().Type())
	}
	return f, nil
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
