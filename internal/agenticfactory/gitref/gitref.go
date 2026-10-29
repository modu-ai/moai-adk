// Package gitref is a claim.ControlRef kept on a git branch (Issue #1948,
// refs/heads/moai-control). Every key is one JSON file under control/, holding
// {"rev": N, "data": ...}. A write is a new commit on top of the fetched head,
// pushed without force; a non-fast-forward rejection means the branch moved, so
// the head is fetched again and the write is retried only if the key's own
// revision is still the expected one. That makes a per-key compare-and-swap out
// of git's per-branch one, and the history append-only.
//
// What this cannot do, and what it does about it, is in the field notes of
// the lane-18 report: it cannot stop a writer with push rights from rewriting
// an unprotected branch, but it notices (the last head it saw must remain an
// ancestor of the new head) and refuses to write on top of a rewritten history.
// It never creates the branch: a missing genesis is "store unavailable".
package gitref

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

const (
	defaultBranch = "moai-control"
	schemaPath    = "control/schema.json"
	schemaName    = "moai-control"
	schemaVersion = 1
	maxTries      = 12
	retryWait     = 10 * time.Millisecond
)

var keyPattern = regexp.MustCompile(`^(issue|request|agent|intent|audit|context|checkpoint|collector|worktree|session|grant|evidence)/[a-z0-9-]{1,64}$`)

// Ref reads and writes the control branch through a local working repository
// (Dir), which is only used for plumbing: it is never checked out.
type Ref struct {
	Dir    string   // a git repository (a clone or an empty init) used for fetch and object writes
	Remote string   // remote name or URL
	Branch string   // branch name, default moai-control
	Env    []string // extra environment, e.g. a hermetic git config in tests
	// Author and Committer are the "Name <email>" strings written into the
	// commits this Ref creates. Empty means DefaultIdentity (Committer empty
	// means the same as Author); the git configuration is never consulted, so a
	// personal user.email cannot end up in the history. They are commit
	// metadata only: no signature is made and nothing is authenticated by them,
	// and the credentials used to push are untouched.
	Author    string
	Committer string

	mu       sync.Mutex
	lastHead string
	runs     int
}

// DefaultIdentity is the machine identity used when none is configured. The
// address is a noreply one on a reserved (.invalid) domain, not a mailbox.
const DefaultIdentity = "MoAI Control <noreply@moai-control.invalid>"

type identity struct{ name, email string }

// parseIdentity checks a "Name <email>" string: one name, one angle-bracketed
// address, no control characters (so a value cannot inject another header).
func parseIdentity(s string) (identity, error) {
	bad := func() (identity, error) {
		return identity{}, fmt.Errorf("%w: commit identity %q must look like \"Name <noreply@host>\"", core.ErrIDInvalid, s)
	}
	for _, c := range s {
		if c < 0x20 || c == 0x7f {
			return bad()
		}
	}
	open := strings.Index(s, " <")
	if open <= 0 || !strings.HasSuffix(s, ">") || strings.Count(s, "<") != 1 || strings.Count(s, ">") != 1 {
		return bad()
	}
	name, email := strings.TrimSpace(s[:open]), strings.TrimSpace(s[open+2:len(s)-1])
	if name == "" || email == "" || !strings.Contains(email, "@") || strings.ContainsAny(email, " \t") {
		return bad()
	}
	return identity{name, email}, nil
}

// identities resolves the author and committer, validating both.
func (r *Ref) identities() (author, committer identity, err error) {
	a := r.Author
	if a == "" {
		a = DefaultIdentity
	}
	if author, err = parseIdentity(a); err != nil {
		return
	}
	c := r.Committer
	if c == "" {
		return author, author, nil
	}
	committer, err = parseIdentity(c)
	return
}

// New returns a Ref for dir and remote.
func New(dir, remote string) *Ref { return &Ref{Dir: dir, Remote: remote, Branch: defaultBranch} }

// Invocations reports how many git processes this Ref has started.
func (r *Ref) Invocations() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.runs
}

type result struct {
	out, errOut string
	exit        int
}

// git runs one git command. A non-zero exit is returned in result, not as an error.
func (r *Ref) git(ctx context.Context, stdin []byte, env []string, args ...string) (result, error) {
	r.runs++
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", r.Dir}, args...)...)
	cmd.Env = append(append(os.Environ(), r.Env...), env...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	err := cmd.Run()
	res := result{out: strings.TrimRight(out.String(), "\n"), errOut: strings.TrimSpace(errb.String())}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		res.exit = ee.ExitCode()
		return res, nil
	}
	return res, err
}

func unavailable(format string, a ...any) error {
	msg, _ := core.Redact(fmt.Sprintf(format, a...))
	return fmt.Errorf("%w: %s", core.ErrStoreUnavailable, msg)
}

func (r *Ref) branch() string {
	if r.Branch == "" {
		return defaultBranch
	}
	return r.Branch
}

func pathOf(key string) (string, error) {
	if !keyPattern.MatchString(key) {
		return "", fmt.Errorf("%w: control key %q", core.ErrIDInvalid, key)
	}
	return "control/" + key + ".json", nil
}

// sync fetches the branch and returns its head after checking that the history
// was not rewritten and that the schema is one this code understands.
func (r *Ref) sync(ctx context.Context) (string, error) {
	res, err := r.git(ctx, nil, nil, "fetch", "--no-tags", r.Remote, "refs/heads/"+r.branch())
	if err != nil {
		return "", unavailable("git fetch: %v", err)
	}
	if res.exit != 0 {
		return "", unavailable("git fetch exited %d: %s", res.exit, res.errOut)
	}
	res, err = r.git(ctx, nil, nil, "rev-parse", "FETCH_HEAD")
	if err != nil || res.exit != 0 {
		return "", unavailable("rev-parse FETCH_HEAD: %s %v", res.errOut, err)
	}
	head := res.out
	if r.lastHead != "" && r.lastHead != head {
		anc, err := r.git(ctx, nil, nil, "merge-base", "--is-ancestor", r.lastHead, head)
		if err != nil {
			return "", unavailable("merge-base: %v", err)
		}
		switch anc.exit {
		case 0:
		case 1:
			return "", fmt.Errorf("%w: %.12s is no longer an ancestor of %.12s", core.ErrHistoryRewritten, r.lastHead, head)
		default:
			return "", unavailable("merge-base exited %d: %s", anc.exit, anc.errOut)
		}
	}
	if head != r.lastHead { // an unchanged head was already checked
		if err := r.checkSchema(ctx, head); err != nil {
			return "", err
		}
	}
	r.lastHead = head
	return head, nil
}

func (r *Ref) checkSchema(ctx context.Context, head string) error {
	res, err := r.git(ctx, nil, nil, "cat-file", "-p", head+":"+schemaPath)
	if err != nil {
		return unavailable("read %s: %v", schemaPath, err)
	}
	if res.exit != 0 {
		return unavailable("%s is missing on the control branch", schemaPath)
	}
	var s struct {
		Schema  string `json:"schema"`
		Version int    `json:"schema_version"`
	}
	if json.Unmarshal([]byte(res.out), &s) != nil || s.Schema != schemaName || s.Version != schemaVersion {
		return unavailable("unknown control schema %q version %d", s.Schema, s.Version)
	}
	return nil
}

type entry struct {
	Rev  int64           `json:"rev"`
	Data json.RawMessage `json:"data"`
}

// read returns the stored entry at head, or found=false.
func (r *Ref) read(ctx context.Context, head, path string) (entry, bool, error) {
	ls, err := r.git(ctx, nil, nil, "ls-tree", head, "--", path)
	if err != nil || ls.exit != 0 {
		return entry{}, false, unavailable("ls-tree: %s %v", ls.errOut, err)
	}
	if ls.out == "" {
		return entry{}, false, nil
	}
	res, err := r.git(ctx, nil, nil, "cat-file", "-p", head+":"+path)
	if err != nil || res.exit != 0 {
		return entry{}, false, unavailable("cat-file %s: %s %v", path, res.errOut, err)
	}
	var e entry
	if err := json.Unmarshal([]byte(res.out), &e); err != nil {
		return entry{}, false, unavailable("corrupt control entry %s: %v", path, err)
	}
	return e, true, nil
}

// Read returns the data and revision of key, or core.ErrNotHeld.
func (r *Ref) Read(ctx context.Context, key string) ([]byte, int64, error) {
	path, err := pathOf(key)
	if err != nil {
		return nil, 0, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	head, err := r.sync(ctx)
	if err != nil {
		return nil, 0, err
	}
	e, found, err := r.read(ctx, head, path)
	if err != nil {
		return nil, 0, err
	}
	if !found {
		return nil, 0, core.ErrNotHeld
	}
	return []byte(e.Data), e.Rev, nil
}

// Keys lists the keys that start with prefix, sorted.
func (r *Ref) Keys(ctx context.Context, prefix string) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	head, err := r.sync(ctx)
	if err != nil {
		return nil, err
	}
	res, err := r.git(ctx, nil, nil, "ls-tree", "-r", "--name-only", head, "--", "control/")
	if err != nil || res.exit != 0 {
		return nil, unavailable("ls-tree: %s %v", res.errOut, err)
	}
	var keys []string
	for _, line := range strings.Split(res.out, "\n") {
		k := strings.TrimSuffix(strings.TrimPrefix(line, "control/"), ".json")
		if line == "" || line == schemaPath || !keyPattern.MatchString(k) || !strings.HasPrefix(k, prefix) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// CompareAndSwap writes data when the key's revision is expectRev (0 = absent).
func (r *Ref) CompareAndSwap(ctx context.Context, key string, expectRev int64, data []byte) (int64, error) {
	path, err := pathOf(key)
	if err != nil {
		return 0, err
	}
	if !json.Valid(data) {
		return 0, fmt.Errorf("%w: control data for %s is not JSON", core.ErrIDInvalid, key)
	}
	var newRev int64
	err = r.mutate(ctx, key, path, expectRev, func(cur entry, found bool) ([]byte, error) {
		newRev = cur.Rev + 1
		return json.Marshal(entry{Rev: newRev, Data: data})
	})
	if err != nil {
		return 0, err
	}
	return newRev, nil
}

// Delete removes key when its revision is expectRev.
func (r *Ref) Delete(ctx context.Context, key string, expectRev int64) error {
	path, err := pathOf(key)
	if err != nil {
		return err
	}
	return r.mutate(ctx, key, path, expectRev, nil)
}

// mutate is the shared write loop. build returns the new file content; a nil
// build removes the file.
func (r *Ref) mutate(ctx context.Context, key, path string, expectRev int64, build func(entry, bool) ([]byte, error)) error {
	if _, _, err := r.identities(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for try := 0; try < maxTries; try++ {
		head, err := r.sync(ctx)
		if err != nil {
			return err
		}
		cur, found, err := r.read(ctx, head, path)
		if err != nil {
			return err
		}
		if build == nil && !found {
			return core.ErrNotHeld
		}
		if cur.Rev != expectRev {
			return fmt.Errorf("%w: %s is at revision %d, not %d", core.ErrRevisionConflict, key, cur.Rev, expectRev)
		}
		var content []byte
		if build != nil {
			if content, err = build(cur, found); err != nil {
				return err
			}
		}
		commit, err := r.commit(ctx, head, path, content, key)
		if err != nil {
			return err
		}
		push, err := r.git(ctx, nil, nil, "push", r.Remote, commit+":refs/heads/"+r.branch())
		if err != nil {
			return unavailable("git push: %v", err)
		}
		if push.exit == 0 {
			r.lastHead = commit
			return nil
		}
		// A failed push is "the branch moved" only if the remote head really
		// changed since we based the commit on it. Git's wording is not used: a
		// hook's "remote rejected" reads just like a non-fast-forward.
		if now, serr := r.sync(ctx); serr == nil && now != head {
			// Pause with jitter so racing writers stop colliding in step, then
			// retry only if this key is untouched.
			select {
			case <-ctx.Done():
				return unavailable("waiting to retry: %v", ctx.Err())
			case <-time.After(retryWait + time.Duration(rand.Int64N(int64(2*retryWait)))):
			}
			continue
		} else if serr != nil {
			return serr
		}
		return unavailable("git push exited %d: %s", push.exit, push.errOut)
	}
	return fmt.Errorf("%w: the control branch kept moving", core.ErrBusy)
}

// commit builds a child of head that adds, replaces or (content nil) removes one
// file, using a temporary index so the working tree is never touched.
func (r *Ref) commit(ctx context.Context, head, path string, content []byte, key string) (string, error) {
	idx, err := os.CreateTemp("", "moai-control-index-*")
	if err != nil {
		return "", unavailable("temp index: %v", err)
	}
	idx.Close()
	defer os.Remove(idx.Name())
	env := []string{"GIT_INDEX_FILE=" + idx.Name()}
	os.Remove(idx.Name()) // git creates it

	step := func(stdin []byte, args ...string) (string, error) {
		res, err := r.git(ctx, stdin, env, args...)
		if err != nil || res.exit != 0 {
			return "", unavailable("git %s: %s %v", args[0], res.errOut, err)
		}
		return res.out, nil
	}
	if _, err := step(nil, "read-tree", head); err != nil {
		return "", err
	}
	if content == nil {
		if _, err := step(nil, "update-index", "--force-remove", path); err != nil {
			return "", err
		}
	} else {
		blob, err := step(append(content, '\n'), "hash-object", "-w", "--stdin")
		if err != nil {
			return "", err
		}
		if _, err := step(nil, "update-index", "--add", "--cacheinfo", "100644,"+blob+","+path); err != nil {
			return "", err
		}
	}
	tree, err := step(nil, "write-tree")
	if err != nil {
		return "", err
	}
	author, committer, err := r.identities()
	if err != nil {
		return "", err
	}
	env = append(env,
		"GIT_AUTHOR_NAME="+author.name, "GIT_AUTHOR_EMAIL="+author.email,
		"GIT_COMMITTER_NAME="+committer.name, "GIT_COMMITTER_EMAIL="+committer.email)
	msg := "control: " + key + "\n\nIssue #1948\n\n🗿 MoAI"
	return step(nil, "commit-tree", tree, "-p", head, "-m", msg)
}
