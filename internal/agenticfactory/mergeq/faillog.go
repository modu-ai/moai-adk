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
// control characters and passed through core.Redact before it is written, as a best effort and not as the guarantee:
// the guarantee is the file mode and that nothing else points at the content.
type FailureLog struct {
	Dir      string // created with mode 0700
	Cap      int    // bytes kept per file; DefaultFailureLogCap when not positive
	TotalCap int64  // bytes all the *.log files of Dir may hold together; DefaultFailureLogTotalCap when not positive
}

// DefaultFailureLogTotalCap is how many bytes all the logs of one directory may hold together: when a write takes
// the directory past it, the oldest files are removed first.
const DefaultFailureLogTotalCap = 4 * DefaultFailureLogCap

// failureLogExt ends the name of every log file; the retention pass looks at these and nothing else.
const failureLogExt = ".log"

// failureLogTmpPrefix starts the name of the file a log is written into before it is renamed onto its name.
const failureLogTmpPrefix = ".tmp-"

// Write stores out for the failure of kind (gate, merge, git) of card's candidate and returns the file's path.
func (l *FailureLog) Write(card, candidate, kind string, out []byte) (string, error) {
	if l == nil || l.Dir == "" {
		return "", fmt.Errorf("mergeq: no failure log directory is configured")
	}
	if err := os.MkdirAll(l.Dir, 0o700); err != nil {
		return "", err
	}
	_ = os.Chmod(l.Dir, 0o700) // MkdirAll honours the umask; an existing directory keeps whatever mode it had
	path := filepath.Join(l.Dir, failureLogName(card, candidate, kind))

	limit := l.Cap
	if limit <= 0 {
		limit = DefaultFailureLogCap
	}
	body := out
	dropped := 0
	if len(body) > limit {
		dropped = len(body) - limit
		body = body[dropped:]
	}
	// The raw text is masked once before the controls are stripped: a marker inside an escape sequence would be
	// removed with the sequence, and the value after it would no longer have a marker to follow.
	text, _ := core.Redact(string(body))
	text, _ = core.Redact(StripControls(text))
	text, _ = core.Redact(text)
	// The masking and the stripping change the length ("key=1" becomes "key=[REDACTED]"), so the cap is applied again
	// to what will be written; the cut moves forward to the start of a character.
	if len(text) > limit {
		cut := len(text) - limit
		for cut < len(text) && !utf8.RuneStart(text[cut]) {
			cut++
		}
		dropped += cut
		text = text[cut:]
	}
	if dropped > 0 {
		text = fmt.Sprintf("[%d earlier bytes of output are not kept]\n%s", dropped, text)
	}

	// The text goes into a file of its own (a name no other writer has, mode 0600) and is renamed onto the log's name:
	// two landers of one key never meet in the same file, and a reader sees the old log or the new one whole. Renaming
	// replaces what is at the path, a link included, and never follows it.
	tmp, err := os.CreateTemp(l.Dir, failureLogTmpPrefix+"*")
	if err != nil {
		return "", err
	}
	discard := func() { _ = tmp.Close(); _ = os.Remove(tmp.Name()) }
	if err := tmp.Chmod(0o600); err != nil {
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

// prune removes the oldest log files of the directory until what is left fits the total cap. keep, the file just
// written, is never removed, whatever its size. A file that cannot be removed or read is skipped: the retention pass
// is a bound on how long output stays, not a reason to fail a write that has already succeeded.
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
	for _, e := range entries {
		if !e.Type().IsRegular() || !strings.HasSuffix(e.Name(), failureLogExt) {
			continue
		}
		fi, ierr := e.Info()
		if ierr != nil {
			continue
		}
		files = append(files, logFile{filepath.Join(l.Dir, e.Name()), fi.Size(), fi.ModTime()})
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
