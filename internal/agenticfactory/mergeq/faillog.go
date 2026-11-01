package mergeq

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// DefaultFailureLogCap is how many bytes of one failure's output a log file keeps (the end of it).
const DefaultFailureLogCap = 256 << 10

// FailureLog keeps the output of a failed gate or git step, which must not enter the ledger or the verb's output:
// a deny-list of secret shapes cannot be made complete, so the output is not put anywhere a reader did not choose
// to look. It goes to a file in a directory only the user can read; the reason says that file's path.
//
// One file per card, candidate and kind, overwritten by the next failure of that kind. The content is stripped of
// control characters and passed through core.Redact before it is written. That is a best effort and not a guarantee:
// a secret can survive it, and any process of the same user can read the file. What this code does set is a 0600 file
// in a 0700 directory that is not a link, no extended ACL on the file or the directory (macOS only: the output is
// written after the ACL is cleared, and when clearing fails it is dropped), a cap per file and one cap for the whole
// directory with the oldest files removed first. It does not cover the directories above Dir, ACLs on other
// platforms, or a reader of the same user; the verb's help says that the logs can hold secrets.
type FailureLog struct {
	Dir      string // created with mode 0700
	Cap      int    // bytes kept per file; DefaultFailureLogCap when not positive
	TotalCap int64  // bytes the *.log files of Dir and the temp files of writes in flight may hold together; DefaultFailureLogTotalCap when not positive

	clearACL func(path string) error // clearExtendedACL when nil; a test replaces it to make clearing fail
}

// stripACL removes the extended ACL of path, by the platform's means.
func (l *FailureLog) stripACL(path string) error {
	if l.clearACL != nil {
		return l.clearACL(path)
	}
	return clearExtendedACL(path)
}

// DefaultFailureLogTotalCap is how many bytes all the logs of one directory may hold together: when a write takes
// the directory past it, the oldest files are removed first.
const DefaultFailureLogTotalCap = 4 * DefaultFailureLogCap

// failureLogExt ends the name of every log file; the retention pass looks at these and nothing else.
const failureLogExt = ".log"

// failureLogTmpPrefix starts the name of the file a log is written into before it is renamed onto its name.
const failureLogTmpPrefix = ".tmp-"

// failureLogHeaderCap is how many bytes of a header (the gate text) a log keeps: the start of it.
const failureLogHeaderCap = 2 << 10

// Write stores out for the failure of kind (gate, merge, git) of card's candidate and returns the file's path.
func (l *FailureLog) Write(card, candidate, kind string, out []byte) (string, error) {
	return l.write(card, candidate, kind, "", out)
}

// mask is what every text passes before it is written: the masking once on the raw text (a marker inside an escape
// sequence would be removed with the sequence, and the value after it would no longer have a marker to follow), the
// stripping of control characters, and the masking again on the result.
func mask(s string) string {
	s, _ = core.Redact(s)
	s, _ = core.Redact(StripControls(s))
	s, _ = core.Redact(s)
	return s
}

// maskTail is mask for output that is cut to its last limit bytes: it returns the masked tail and how many bytes were
// cut after the masking began. The first masking can make the text 2.5 times longer ("key=1" becomes "key=[REDACTED]"),
// and every pass costs time in proportion to the text it sees, so the cut comes right after the first pass (what
// remains is already masked, and a cut in it uncovers nothing) and the last cut keeps the file within the cap.
func maskTail(s string, limit int) (string, int) {
	s, _ = core.Redact(s)
	s, cut := tailOf(s, limit)
	s, _ = core.Redact(StripControls(s))
	s, _ = core.Redact(s)
	s, more := tailOf(s, limit)
	return s, cut + more
}

// tailOf is the last limit bytes of s, cut forward to the start of a character, and how many bytes were cut.
func tailOf(s string, limit int) (string, int) {
	if len(s) <= limit {
		return s, 0
	}
	cut := len(s) - limit
	for cut < len(s) && !utf8.RuneStart(s[cut]) {
		cut++
	}
	return s[cut:], cut
}

// write is Write with a header: a line of text that stands before the output and is kept whatever the output's size
// (the gate text, which the reason and the ledger carry only as a short id).
func (l *FailureLog) write(card, candidate, kind, header string, out []byte) (string, error) {
	if l == nil || l.Dir == "" {
		return "", fmt.Errorf("mergeq: no failure log directory is configured")
	}
	if err := l.prepareDir(); err != nil {
		return "", err
	}
	path := filepath.Join(l.Dir, failureLogName(card, candidate, kind))

	limit := l.Cap
	if limit <= 0 {
		limit = DefaultFailureLogCap
	}
	// The cut before the masking bounds the work the masking does; the cut after it bounds the file, because the
	// masking and the stripping change the length ("key=1" becomes "key=[REDACTED]").
	body, dropped := tailOf(string(out), limit)
	text, more := maskTail(body, limit)
	dropped += more
	if dropped > 0 {
		text = fmt.Sprintf("[%d earlier bytes of output are not kept]\n%s", dropped, text)
	}
	if header != "" {
		head := mask(header)
		if len(head) > failureLogHeaderCap {
			head = strings.ToValidUTF8(head[:failureLogHeaderCap], "") + " [the rest of the text is not kept]"
		}
		text = head + "\n" + text
	}

	// The text goes into a file of its own (a name no other writer has, mode 0600) and is renamed onto the log's name:
	// two landers of one key never meet in the same file, and a reader sees the old log or the new one whole. Renaming
	// replaces what is at the path, a link included, and never follows it.
	tmp, err := os.CreateTemp(l.Dir, failureLogTmpPrefix+"*")
	if err != nil {
		return "", err
	}
	discard := func() { _ = tmp.Close(); _ = os.Remove(tmp.Name()) }
	// The output is written only after the file is known to carry no ACL: when clearing fails, the output is dropped
	// and the caller records that logging failed.
	if err := tmp.Chmod(0o600); err != nil {
		discard()
		return "", err
	}
	if err := l.stripACL(tmp.Name()); err != nil {
		discard()
		return "", err
	}
	if _, err := tmp.WriteString(text); err != nil {
		discard()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return "", err
	}
	l.prune(path)
	return path, nil
}

// prepareDir makes sure Dir is a real directory of mode 0700 before anything is written into it or changed about it.
// A link, or anything that is not a directory, is refused: following it would change the mode of whatever it points at
// and put the output there. Only the last component of Dir is looked at; the components above it are the caller's
// path to the ledger and are out of scope, and a link swapped in between this look and the write needs the same
// user's write access to the parent, which already reads the ledger.
func (l *FailureLog) prepareDir() error {
	fi, err := os.Lstat(l.Dir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(l.Dir, 0o700); err != nil {
			return err
		}
		fi, err = os.Lstat(l.Dir)
	}
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return fmt.Errorf("mergeq: the failure log directory %s is a link or not a directory", l.Dir)
	}
	// Before anything is made in the directory: a file made in a directory with an inheritable ACL starts with it.
	if err := l.stripACL(l.Dir); err != nil {
		return err
	}
	return os.Chmod(l.Dir, 0o700) // MkdirAll honours the umask; an existing directory keeps whatever mode it had
}

// failureLogTmpMaxAge is how old a temp file must be before the retention pass takes it for the leftover of a write
// that died. A write creates its temp file, fills it and renames it within a moment, so a minute is far past any write
// in flight: a younger temp file is left alone and counted in the total.
const failureLogTmpMaxAge = time.Minute

// prune removes the temp files that a dead write left behind, then the oldest log files of the directory until what is
// left fits the total cap. The temp files that are not old enough to be leftovers belong to a write in flight: they are
// not touched, and their bytes count toward the cap. keep, the file just written, is never removed, whatever its size.
// A file that cannot be removed or read is skipped: the retention pass is a bound on how long output stays, not a
// reason to fail a write that has already succeeded.
func (l *FailureLog) prune(keep string) {
	limit := l.TotalCap
	if limit <= 0 {
		limit = DefaultFailureLogTotalCap
	}
	entries, err := os.ReadDir(l.Dir)
	if err != nil {
		return
	}
	type logFile struct {
		path string
		size int64
		mod  time.Time
	}
	var files []logFile
	var sum int64
	now := time.Now()
	for _, e := range entries {
		temp := strings.HasPrefix(e.Name(), failureLogTmpPrefix)
		if !e.Type().IsRegular() || (!temp && !strings.HasSuffix(e.Name(), failureLogExt)) {
			continue
		}
		fi, ierr := e.Info()
		if ierr != nil {
			continue
		}
		path := filepath.Join(l.Dir, e.Name())
		if temp {
			if now.Sub(fi.ModTime()) > failureLogTmpMaxAge {
				if rerr := os.Remove(path); rerr != nil && !os.IsNotExist(rerr) {
					sum += fi.Size() // a leftover that stays still holds its bytes
				}
				continue
			}
			sum += fi.Size() // a write in flight: counted, never removed
			continue
		}
		files = append(files, logFile{path, fi.Size(), fi.ModTime()})
		sum += fi.Size()
	}
	if sum <= limit {
		return
	}
	sort.Slice(files, func(i, j int) bool {
		if !files[i].mod.Equal(files[j].mod) {
			return files[i].mod.Before(files[j].mod)
		}
		return files[i].path < files[j].path
	})
	for _, f := range files {
		if sum <= limit {
			return
		}
		if f.path == keep {
			continue
		}
		if rerr := os.Remove(f.path); rerr == nil || os.IsNotExist(rerr) {
			sum -= f.size
		}
	}
}

// controlSequences finds the sequences a terminal acts on: CSI (ESC [ ... final byte, or the single-byte C1 form),
// OSC (ESC ] ... BEL or ESC \) and the remaining two-byte escapes. An OSC that is never closed ends at its line, so it
// cannot swallow the lines after it (and with them the reason the gate failed).
var controlSequences = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]|\x{9b}[0-?]*[ -/]*[@-~]|\x1b\][^\x07\x1b\n]*(?:\x07|\x1b\\)?|\x1b[@-_]`)

// StripControls removes escape sequences whole, then the control characters other than newline and tab (carriage
// return, bell, backspace, form feed, NUL, delete and the C1 range): a log is read in a terminal, where an escape
// sequence would be executed by whoever reads it.
func StripControls(s string) string {
	return strings.Map(func(r rune) rune {
		if (r < 0x20 && r != '\n' && r != '\t') || (r >= 0x7f && r <= 0x9f) {
			return -1
		}
		return r
	}, controlSequences.ReplaceAllString(s, ""))
}

// failureLogName is a file name that holds only a-z, A-Z, 0-9, '.', '_' and '-', never starts with a dot and never
// holds a separator, however the card id and the candidate are spelled. A short digest of the three parts keeps two
// names that clean to the same text apart.
func failureLogName(card, candidate, kind string) string {
	sum := sha256.Sum256([]byte(card + "\x00" + candidate + "\x00" + kind))
	clean := func(s string, n int) string {
		var b strings.Builder
		for _, r := range s {
			switch {
			case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
				b.WriteRune(r)
			default:
				b.WriteByte('_')
			}
			if b.Len() >= n {
				break
			}
		}
		return b.String()
	}
	return clean(card, 40) + "-" + hex.EncodeToString(sum[:6]) + "-" + clean(kind, 12) + ".log"
}
