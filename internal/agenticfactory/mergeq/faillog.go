package mergeq

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	Dir string // created with mode 0700
	Cap int    // bytes kept per file; DefaultFailureLogCap when not positive
}

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
	if dropped > 0 {
		text = fmt.Sprintf("[%d earlier bytes of output are not kept]\n%s", dropped, text)
	}

	// A file already at the path is removed, never followed: a link left there must not redirect the write.
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", err
	}
	_ = f.Chmod(0o600)
	if _, err := f.WriteString(text); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return path, nil
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
