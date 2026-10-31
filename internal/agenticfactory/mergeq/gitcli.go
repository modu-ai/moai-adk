package mergeq

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// gitTimeout bounds every git call of the merge queue; a hung git must not hold the landing lock.
const gitTimeout = 2 * time.Minute

// gitWaitDelay is how long to wait for git's output pipes after its context ended.
const gitWaitDelay = 5 * time.Second

// maxGitOutput bounds what one git call may return, so a runaway command cannot fill memory.
const maxGitOutput = 8 << 20

// ErrMerge marks a merge of the candidate into the trunk that git refused (a conflict); the candidate is
// not mergeable as it stands and the card goes back to be fixed.
var ErrMerge = errors.New("mergeq: the candidate does not merge into the trunk")

// ErrRefMoved marks a conditional ref update that found the ref no longer at the value it was read at.
var ErrRefMoved = errors.New("mergeq: the trunk moved since it was read")

// Identity is the author the merge commit is made with. It is explicit: the queue never reads the
// operator's global git configuration for it.
type Identity struct{ Name, Email string }

// Candidate is a merge result built in a throwaway worktree: Dir is where the gates run, Result is the
// merge commit they ran against.
type Candidate struct {
	Dir    string
	Result string
}

// Git is what the merge queue needs from git. GitCLI is the production implementation.
type Git interface {
	RevParse(ctx context.Context, ref string) (string, error)
	IsAncestor(ctx context.Context, ancestor, descendant string) (bool, error)
	// PrepareCandidate checks out trunkSHA detached in a fresh worktree and merges candidateSHA into it
	// (no fast-forward), so the gates see the real merge result. Nothing it does touches the trunk ref.
	PrepareCandidate(ctx context.Context, trunkSHA, candidateSHA, message string) (Candidate, error)
	Discard(ctx context.Context, c Candidate) error
	// OriginHeadBranch is the local branch name origin/HEAD points at, or "" when there is no origin/HEAD.
	OriginHeadBranch(ctx context.Context) (string, error)
	// CheckedOutAt returns the worktree that has ref checked out, or "" when none does.
	CheckedOutAt(ctx context.Context, ref string) (string, error)
	// IsSymbolicRef says whether ref is a symbolic ref (it names another ref instead of a commit).
	IsSymbolicRef(ctx context.Context, ref string) (bool, error)
	// UpdateRef moves ref from oldSHA to newSHA only if it still points at oldSHA (`git update-ref --no-deref ref new old`).
	UpdateRef(ctx context.Context, ref, newSHA, oldSHA string) error
}

// GitCLI runs the git command line against one repository. Its environment is scrubbed of the variables
// that choose a repository, and it never prompts.
type GitCLI struct {
	Repo     string // the repository (any worktree of it)
	ScratchD string // where candidate worktrees are made; must be outside the repository's worktrees
	Who      Identity
}

func (g GitCLI) run(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	cmd.WaitDelay = gitWaitDelay // a descendant that keeps the pipes open must not outlive the timeout
	// The C locale, always: errors are classified by git's wording ("but expected"), and git translates its
	// messages when the operator's locale (Korean here) has a catalogue. LANGUAGE is cleared because gettext
	// consults it before the other variables.
	cmd.Env = append(gitenv.Env(),
		"LC_ALL=C", "LANG=C", "LANGUAGE=",
		"GIT_TERMINAL_PROMPT=0",
		"GIT_AUTHOR_NAME="+g.Who.Name, "GIT_AUTHOR_EMAIL="+g.Who.Email,
		"GIT_COMMITTER_NAME="+g.Who.Name, "GIT_COMMITTER_EMAIL="+g.Who.Email,
	)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &limited{b: &out}, &limited{b: &errb}
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(out.String()), fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(errb.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// NewTailWriter is a writer that keeps the LAST max bytes written to it and never fails: for the output of a child
// whose end matters more than its start (a gate command: the failure is the last thing it prints) and whose
// volume must not fill memory. It is not the limit git calls have (that one refuses runaway output with an error).
func NewTailWriter(max int) *TailWriter { return &TailWriter{max: max, span: -1} }

// SecretMarkerPattern is a secret-named field that introduces a value (`password=`, `API_KEY:`); it carries no
// case flag and no capture group, so a caller wraps it as it needs.
const SecretMarkerPattern = `\b[a-z0-9_]*(?:token|secret|pass(?:phrase|w(?:or)?d)|api[_-]?key|credential|cookie|key)[a-z0-9_]*[ \t]*[:=]`

// tailCarry is how many raw bytes of the line being written the watch keeps between writes, so a marker (or the
// escape sequence or control character inside it) that two writes split is still seen whole. It is a ceiling: a
// marker whose raw form is longer than this is not recognised.
const tailCarry = 512

var (
	// tailMarker finds what introduces a secret value: a secret-named field, an Authorization header, a PEM
	// private-key header, the user part of a URL credential. A bare `Bearer <token>` with no header is not watched.
	tailMarker = regexp.MustCompile(`(?i)` + SecretMarkerPattern + `|authorization[ \t]*:|-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----|https?://[^/\s:@]+:`)
	// tailMarkerEnd is a marker that ends its line, so the value is on the next one.
	tailMarkerEnd = regexp.MustCompile(`(?i)(?:` + SecretMarkerPattern + `|authorization[ \t]*:(?:[ \t]*(?:bearer|basic|token))?)[ \t]*$`)
	// tailEscapes finds the sequences a terminal acts on: CSI (ESC [ ... final byte, or the single-byte C1 form),
	// OSC (ESC ] ... BEL or ESC \) and the remaining two-byte escapes.
	tailEscapes = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]|\x{9b}[0-?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)?|\x1b[@-_]`)
)

// StripControls removes escape sequences whole, then the control characters other than newline and tab (carriage
// return, bell, backspace, form feed, NUL, delete and the C1 range).
func StripControls(s string) string {
	plain := true
	for i := 0; i < len(s) && plain; i++ {
		plain = s[i] >= 0x20 && s[i] < 0x7f || s[i] == '\t' || s[i] == '\n' // printable ASCII: nothing to strip
	}
	if plain {
		return s
	}
	return strings.Map(func(r rune) rune {
		if (r < 0x20 && r != '\n' && r != '\t') || (r >= 0x7f && r <= 0x9f) {
			return -1
		}
		return r
	}, tailEscapes.ReplaceAllString(s, ""))
}

// markSpan is a region of the stream (absolute offsets) whose text a secret marker introduces: the line that holds
// the marker, extended over the next line when the marker ends its line. end < 0 while the span is still open.
type markSpan struct{ start, end int64 }

// TailWriter holds the last max bytes written to it; Bytes returns them. It also watches the whole stream, in
// bounded state, for lines introduced by a secret marker, so that after the start of a line has been dropped it can
// still say that the kept part of that line is a secret value whose marker is gone (MaskedPrefix).
type TailWriter struct {
	max     int
	buf     []byte
	dropped bool

	total     int64      // bytes written so far
	lineStart int64      // absolute offset of the line being written
	carry     []byte     // the last tailCarry raw bytes of that line
	curMarked bool       // that line holds a marker
	spans     []markSpan // marker spans not yet wholly behind the kept window
	span      int        // index of the open span, or -1
	chain     bool       // the open span continues over the next line (a marker ended this one)
}

// watch feeds the writer's secret-marker bookkeeping with p, which has not been stored yet.
func (w *TailWriter) watch(p []byte) {
	base := w.total
	for off := 0; off < len(p); {
		end, nl := len(p), false
		if i := bytes.IndexByte(p[off:], '\n'); i >= 0 {
			end, nl = off+i, true
		}
		w.line(p[off:end], nl, base+int64(end))
		off = end
		if nl {
			off++
		}
	}
	w.total += int64(len(p))
}

// line takes one piece of the current line; nl says that a newline follows it, at absolute offset nlAt.
func (w *TailWriter) line(seg []byte, nl bool, nlAt int64) {
	raw := append(append(make([]byte, 0, len(w.carry)+len(seg)), w.carry...), seg...)
	// Every marker holds ':', '=' or the dashes of a PEM header (or an ESC/control split of them, which the carry and
	// StripControls bring back together), so text with none of those needs no regular expression: the common
	// case of a long run of output is not slowed down.
	if !w.curMarked && bytes.ContainsAny(raw, ":=-") && tailMarker.MatchString(StripControls(string(raw))) {
		w.curMarked = true
		if w.span < 0 {
			w.spans = append(w.spans, markSpan{start: w.lineStart, end: -1})
			w.span = len(w.spans) - 1
		}
	}
	if !nl {
		w.carry = append(w.carry[:0], raw[max(0, len(raw)-tailCarry):]...)
		return
	}
	endsMarker := tailMarkerEnd.MatchString(StripControls(string(raw[max(0, len(raw)-2*tailCarry):])))
	switch {
	case w.chain && endsMarker: // the value line is itself introducing another value: stay open
	case w.chain || (w.curMarked && !endsMarker):
		w.spans[w.span].end, w.span, w.chain = nlAt+1, -1, false
	case w.curMarked: // the marker ends the line: the next line is its value
		w.chain = true
	}
	w.lineStart, w.curMarked, w.carry = nlAt+1, false, w.carry[:0]
}

// MaskedPrefix is how many leading bytes of Bytes() belong to a line that a secret marker introduced and whose
// start has been dropped, so the kept bytes are (part of) a value and the marker that would have masked them is
// outside the window. It errs toward masking: a line whose marker is in the dropped part or may straddle the cut
// counts, and so does the value line after a line that ends in a marker.
func (w *TailWriter) MaskedPrefix() int {
	front := w.total - int64(len(w.buf))
	if front == 0 {
		return 0
	}
	// prune keeps every closed span's end ahead of front, so a span that starts before the window covers its front.
	for _, s := range w.spans {
		if s.start < front {
			end := w.total
			if s.end >= 0 && s.end < end {
				end = s.end
			}
			return int(end - front)
		}
	}
	return 0
}

// prune forgets the spans that lie wholly behind the kept window.
func (w *TailWriter) prune() {
	front, k := w.total-int64(len(w.buf)), 0
	for k < len(w.spans) && w.spans[k].end >= 0 && w.spans[k].end <= front {
		k++
	}
	if k > 0 {
		w.spans = append(w.spans[:0], w.spans[k:]...)
		if w.span >= 0 {
			w.span -= k
		}
	}
}

func (w *TailWriter) Write(p []byte) (int, error) {
	w.watch(p)
	defer w.prune()
	n := len(p)
	if len(p) >= w.max {
		w.dropped = w.dropped || len(p) > w.max || len(w.buf) > 0
		w.buf = append(w.buf[:0], p[len(p)-w.max:]...)
		return n, nil
	}
	if over := len(w.buf) + len(p) - w.max; over > 0 {
		w.dropped = true
		w.buf = append(w.buf[:0], w.buf[over:]...)
	}
	w.buf = append(w.buf, p...)
	return n, nil
}

// Bytes is what the writer kept.
func (w *TailWriter) Bytes() []byte { return w.buf }

// Truncated reports that the start of the output was dropped, so the first kept line may be a fragment.
func (w *TailWriter) Truncated() bool { return w.dropped }

var _ io.Writer = (*TailWriter)(nil)

type limited struct{ b *bytes.Buffer }

func (l *limited) Write(p []byte) (int, error) {
	if l.b.Len()+len(p) > maxGitOutput {
		return 0, errors.New("mergeq: git output exceeded its limit")
	}
	return l.b.Write(p)
}

// RevParse resolves ref to the commit it names.
func (g GitCLI) RevParse(ctx context.Context, ref string) (string, error) {
	return g.run(ctx, g.Repo, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
}

// IsAncestor reports whether ancestor is reachable from descendant (equal commits count).
func (g GitCLI) IsAncestor(ctx context.Context, ancestor, descendant string) (bool, error) {
	_, err := g.run(ctx, g.Repo, "merge-base", "--is-ancestor", ancestor, descendant)
	if err == nil {
		return true, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// PrepareCandidate builds the merge of candidateSHA into trunkSHA in a throwaway worktree.
func (g GitCLI) PrepareCandidate(ctx context.Context, trunkSHA, candidateSHA, message string) (Candidate, error) {
	if g.ScratchD == "" {
		return Candidate{}, errors.New("mergeq: no scratch directory is configured for candidate worktrees")
	}
	if err := os.MkdirAll(g.ScratchD, 0o755); err != nil {
		return Candidate{}, err
	}
	dir, err := os.MkdirTemp(g.ScratchD, "cand-")
	if err != nil {
		return Candidate{}, err
	}
	if _, err := g.run(ctx, g.Repo, "worktree", "add", "--detach", dir, trunkSHA); err != nil {
		_ = os.RemoveAll(dir)
		return Candidate{}, err
	}
	c := Candidate{Dir: dir}
	if _, err := g.run(ctx, dir, "merge", "--no-ff", "--no-edit", "-m", message, candidateSHA); err != nil {
		_ = g.Discard(ctx, c)
		return Candidate{}, fmt.Errorf("%w: %v", ErrMerge, err)
	}
	res, err := g.run(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		_ = g.Discard(ctx, c)
		return Candidate{}, err
	}
	c.Result = res
	return c, nil
}

// Discard removes a candidate worktree. A tree that is already gone is not an error.
func (g GitCLI) Discard(ctx context.Context, c Candidate) error {
	if c.Dir == "" {
		return nil
	}
	// The tree is throwaway and may hold build output, so it is removed as a directory and git is told to forget
	// it; no force flag is involved (a repository-wide scan keeps force flags out of production literals).
	if err := os.RemoveAll(c.Dir); err != nil {
		return err
	}
	_, err := g.run(ctx, g.Repo, "worktree", "prune")
	return err
}

// OriginHeadBranch returns the local branch name that refs/remotes/origin/HEAD points at (refs/heads/<name>),
// or "" when the repository has no origin/HEAD.
func (g GitCLI) OriginHeadBranch(ctx context.Context) (string, error) {
	out, err := g.run(ctx, g.Repo, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD")
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}
	name, ok := strings.CutPrefix(out, "refs/remotes/origin/")
	if !ok || name == "" {
		return "", nil
	}
	return "refs/heads/" + name, nil
}

// CheckedOutAt returns the directory of a worktree that has ref checked out, or "" when none does. It reads
// `git worktree list --porcelain` for the repository, so a bare repository and every linked tree are covered.
func (g GitCLI) CheckedOutAt(ctx context.Context, ref string) (string, error) {
	out, err := g.run(ctx, g.Repo, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	var dir string
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			dir = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch "):
			if strings.TrimPrefix(line, "branch ") == ref {
				return dir, nil
			}
		}
	}
	return "", nil
}

// IsSymbolicRef reports whether ref is a symbolic ref. `for-each-ref` prints the target of a symbolic ref and
// nothing for a plain one (or a missing one), and exits 0 either way, so the answer never rests on an exit code.
func (g GitCLI) IsSymbolicRef(ctx context.Context, ref string) (bool, error) {
	out, err := g.run(ctx, g.Repo, "for-each-ref", "--format=%(symref)", ref)
	if err != nil {
		return false, err
	}
	return out != "", nil
}

// UpdateRef is the compare-and-swap that lands a card: it succeeds only if ref still names oldSHA. It never
// follows a symbolic ref (--no-deref): the trunk is a plain branch, and a link that points at another branch must
// not move that branch.
func (g GitCLI) UpdateRef(ctx context.Context, ref, newSHA, oldSHA string) error {
	if _, err := g.run(ctx, g.Repo, "update-ref", "--no-deref", ref, newSHA, oldSHA); err != nil {
		// Only a value mismatch means the trunk moved; an invalid name, a missing object or a lock failure is a
		// different problem and must not be retried as if another lander had won.
		if strings.Contains(err.Error(), "but expected") {
			return fmt.Errorf("%w: %v", ErrRefMoved, err)
		}
		return err
	}
	return nil
}
