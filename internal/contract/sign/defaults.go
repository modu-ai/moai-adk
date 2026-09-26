package sign

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// Default seam implementations. They are the only process, terminal, and
// file-system side effects of the package.

// withDefaults returns s with every nil field set to its default.
func withDefaults(s Seams) Seams {
	if s.IsTTY == nil {
		s.IsTTY = stdinIsTerminal
	}
	if s.Getenv == nil {
		s.Getenv = os.Getenv
	}
	if s.ReadLine == nil {
		s.ReadLine = stdinLineReader()
	}
	if s.Out == nil {
		s.Out = io.Discard
	}
	if s.Now == nil {
		s.Now = time.Now
	}
	if s.GitIdentity == nil {
		s.GitIdentity = gitIdentity
	}
	if s.GitHead == nil {
		s.GitHead = gitHead
	}
	if s.NewBatchID == nil {
		s.NewBatchID = randomBatchID
	}
	if s.WriteFile == nil {
		s.WriteFile = writeAtomic
	}
	return s
}

// stdinIsTerminal reports whether standard input is a character device.
func stdinIsTerminal() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// stdinLineReader reads one line of standard input per call, without the
// line terminator. A final line without a newline is returned with a nil
// error; an empty stream returns io.EOF.
func stdinLineReader() func() (string, error) {
	r := bufio.NewReader(os.Stdin)
	return func() (string, error) {
		line, err := r.ReadString('\n')
		if err != nil && (!errors.Is(err, io.EOF) || line == "") {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
}

// gitEnv is the child environment for git: the parent environment without
// GIT_DIR, GIT_WORK_TREE, and GIT_INDEX_FILE, so git resolves the repository
// from its working directory rather than from an inherited override.
func gitEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		switch name {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE":
			continue
		}
		env = append(env, kv)
	}
	return env
}

// runGit runs git with args in root and returns trimmed standard output.
//
// @MX:WARN: [AUTO] Spawns a git subprocess.
// @MX:REASON: The only process creation in the signing path; it must stay out
// of internal/contract (verify is subprocess-free for hooks) and must scrub
// GIT_DIR/GIT_WORK_TREE/GIT_INDEX_FILE so a hook-inherited override cannot
// point it at another repository.
func runGit(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Env = gitEnv()
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

// gitIdentity returns git user.name and user.email in root. An unset key
// makes `git config` exit 1 with no output; that reads as "" rather than an
// error, so the caller refuses with git_identity_missing.
func gitIdentity(root string) (string, string, error) {
	get := func(key string) (string, error) {
		v, err := runGit(root, "config", key)
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 && v == "" {
			return "", nil
		}
		if err != nil {
			return "", fmt.Errorf("git config %s: %w", key, err)
		}
		return v, nil
	}
	name, err := get("user.name")
	if err != nil {
		return "", "", err
	}
	email, err := get("user.email")
	if err != nil {
		return "", "", err
	}
	return name, email, nil
}

// gitHead returns HEAD in root.
func gitHead(root string) (string, error) {
	head, err := runGit(root, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return head, nil
}

// randomBatchID returns 16 random bytes as lowercase hex.
func randomBatchID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand does not fail on supported platforms; fall back to the
		// clock so the id stays non-empty.
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// writeAtomic writes data to a temp file in path's directory and renames it
// onto path.
func writeAtomic(path string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".contract-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	cleanup := func(e error) error {
		_ = tmp.Close()
		_ = os.Remove(name)
		return e
	}
	if _, err := tmp.Write(data); err != nil {
		return cleanup(err)
	}
	if err := tmp.Sync(); err != nil {
		return cleanup(err)
	}
	if err := tmp.Chmod(perm); err != nil {
		return cleanup(err)
	}
	if err := tmp.Close(); err != nil {
		return cleanup(err)
	}
	if err := atomicfile.Replace(name, path); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}
