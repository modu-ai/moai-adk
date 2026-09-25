package cli

// codex_launcher.go — SPEC-CODEX-LAUNCHER-001 M3+M4
// (REQ-CL-001/002/003/012, REQ-CL-014 help copy):
// the `moai codex` command surface. Cobra registration sits in the launch
// group next to cc/glm/cg; the verb routing accepts exactly {bare, cli, app}
// for the launch forms and {status} for the readout (closed sets — unknown
// tokens are rejected, never routed to a launch); --spawn moves the launch to
// a new tmux window.
//
// Two tables, not one. ROUTING answers "what did the operator ask for" and
// stays closed. ARGV TRANSLATION, downstream of it, answers "what does codex
// accept": a routed verb reaches the child ONLY where it names a real codex
// subcommand, so a verb moai synthesized never lands in the child's argv.
//
// -w/--worktree is consumed HERE and never forwarded: it points the child's
// working directory at an existing or newly created worktree. CODEX_HOME
// reaches the child as an explicit environment entry rather than by ambient
// inheritance, on both the direct and the new-window path.
//
// The readout rows come from M2's codexReadiness VERBATIM (AC-CL-004 — the
// command surface never re-words a row), and the binary/auth values come from
// the shared probe, so no second classification path forks here (REQ-CL-007).
// The status readout never writes; -w may create a worktree. POSIX direct
// launch replaces moai with Codex to preserve the factory owner PID; Windows
// retains the child Start/wait path.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/execerr"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/spf13/cobra"
)

// Diagnostics — named constants compared with == by the tests (AC-CL-002's
// "usage constant", AC-CL-011's single-line install action).
const (
	// codexUsageDiag is the one-line usage diagnostic every unknown token
	// receives (AC-CL-002). Byte-identical for all six probe tokens so the
	// rejection cannot leak which token was seen.
	codexUsageDiag = "unknown verb - usage: moai codex [cli] [-w [worktree]] [-- codex-args...] | moai codex status | moai codex app"

	// A bare -w creates an automatically named worktree, like moai cc.
	codexWorktreeValueDiag = "-w could not generate a worktree name"

	// codexInstallHint is the single diagnostic line the launch verbs print
	// when the codex binary is unresolved (AC-CL-011: exactly one line, exact
	// match; the readout forms still succeed in the same state).
	codexInstallHint = "codex not found - install the Codex CLI first: https://developers.openai.com/codex/cli"

	// codexSpawnReadoutDiag rejects --spawn on the readout forms (AC-CL-003):
	// a readout is not something to open a new window for.
	codexSpawnReadoutDiag           = "--spawn applies to the launch verbs only (moai codex cli --spawn / moai codex app --spawn)"
	codexDuplicateLocalOverrideDiag = "duplicate developer_instructions override: local instruction inputs and operator config both set this key"
)

// codexVerb classifies a routed token: which tokens launch and which render
// the readout is a CLOSED SET (AC-CL-002) — an unknown token never falls
// through to a launch.
type codexVerb int

const (
	codexVerbReadout   codexVerb = iota // "status"
	codexVerbLaunchCli                  // "" (bare) and "cli"
	codexVerbLaunchApp
)

// codexVerbRouting is the routing table the tests read symbolically: the
// launch-token set derived from it must equal {"", cli, app} and the
// readout-token set must equal {status}. Absence from the table IS the
// rejection — no default branch exists.
//
// The bare token launches. The risk that argued for the opposite default —
// an accidental invocation carrying the session away — does not apply here:
// the calling shell is still there when the launched Codex process exits.
var codexVerbRouting = map[string]codexVerb{
	"":       codexVerbLaunchCli,
	"cli":    codexVerbLaunchCli,
	"status": codexVerbReadout,
	"app":    codexVerbLaunchApp,
}

// codexChildSubcommand is the ARGV TRANSLATION table — strictly downstream of
// codexVerbRouting and a strict SUBSET of it. A class present here forwards
// its subcommand token to the child; a class absent here forwards nothing but
// the operator's own tail.
//
// app is the only entry, because app is the only routed verb naming a real
// codex subcommand. codex's usage is `codex [OPTIONS] [PROMPT]` with no cli
// subcommand, so forwarding a synthesized cli token would hand codex a PROMPT
// reading "cli"; the bare form's token is the empty string, which is worse.
var codexChildSubcommand = map[codexVerb]string{
	codexVerbLaunchApp: "app",
}

// codexChildArgs assembles the child's argv tail: the translated subcommand
// token where one exists, then the operator's tail verbatim.
func codexChildArgs(kind codexVerb, tail []string) []string {
	args := make([]string, 0, len(tail)+1)
	if sub, ok := codexChildSubcommand[kind]; ok {
		args = append(args, sub)
	}
	return append(args, tail...)
}

// @MX:ANCHOR: [AUTO] Single local-instruction producer for every launch form.
// @MX:REASON: Compose source order and framing before encoding exactly one override.
// @MX:SPEC: SPEC-CODEX-LOCALMD-001
// codexLocalDeveloperInstructionArgs reads both local inputs fresh on each
// launch. JSON string encoding preserves UTF-8 in a TOML basic string.
func codexLocalDeveloperInstructionArgs(projectRoot string) ([]string, error) {
	var payload strings.Builder
	for _, name := range []string{codexClaudeLocalName, codexLocalInstructionName} {
		body, err := readCodexLocalInstruction(projectRoot, name)
		if err != nil {
			return nil, err
		}
		if len(body) == 0 {
			continue
		}
		if payload.Len() > 0 {
			payload.WriteByte('\n')
		}
		fmt.Fprintf(&payload, "<!-- source: %s -->\n", name)
		payload.Write(body)
	}
	if payload.Len() == 0 {
		return nil, nil
	}
	encoded, _ := json.Marshal(payload.String()) // A string is always JSON-encodable.
	return []string{"-c", "developer_instructions=" + string(encoded)}, nil
}

// codexHasDeveloperOverride scans the operator's config options, including
// attached short options, but stops at Codex's own end-of-options marker.
func codexHasDeveloperOverride(tail []string) bool {
	for i := 0; i < len(tail); i++ {
		token := tail[i]
		var value string
		switch {
		case token == "--":
			return false
		case token == "-c" || token == "--config":
			if i+1 >= len(tail) {
				continue
			}
			i++
			value = tail[i]
		case strings.HasPrefix(token, "--config="):
			value = strings.TrimPrefix(token, "--config=")
		case strings.HasPrefix(token, "-c"):
			value = strings.TrimPrefix(strings.TrimPrefix(token, "-c"), "=")
		default:
			continue
		}
		key, _, ok := strings.Cut(value, "=")
		if ok && strings.TrimSpace(key) == "developer_instructions" {
			return true
		}
	}
	return false
}

// @MX:WARN: [AUTO] Measure the final representation, not the source body.
// @MX:REASON: Shell quoting can quadruple single quotes after JSON encoding.
func checkCodexInstructionSize(size int, representation string) error {
	if size > config.DefaultCodexInstructionArgBytes {
		return fmt.Errorf("codex %s is %d bytes, exceeds %d-byte ceiling", representation, size, config.DefaultCodexInstructionArgBytes)
	}
	return nil
}

// launches reports whether the verb class starts a process (AC-CL-002's
// launch set derivation).
func (v codexVerb) launches() bool { return v == codexVerbLaunchCli || v == codexVerbLaunchApp }

// codexLaunchRequest is the immutable launch intent both paths consume: the
// resolved codex binary, the argv tail handed to it, and the working
// directory (the PROJECT ROOT, not the process cwd — AC-CL-002's cwd axis).
type codexLaunchRequest struct {
	Program string
	Args    []string // argv AFTER the program token: [verb, passthrough...]
	Dir     string
}

// codexDirectLaunchFn is the direct-launch seam: it receives the fully
// assembled *exec.Cmd (Stdin/Stdout/Stderr already assigned to the parent's
// own values — AC-CL-002's stdio identity) and runs it. The capture harness
// records the cmd's seven fields here.
var codexDirectLaunchFn = defaultCodexDirectLaunch

// codexSpawnLaunchFn is the spawn seam: it receives the NEW-WINDOW TARGET —
// (dir, program, args) of the codex child itself, NOT a tmux invocation
// (AC-CL-003: the capture compares these tokens with the direct path's).
// The default implementation opens the tmux window via the shared tmuxSpawnFn
// (spawn.go owns the only exec.Command("tmux") primitive — AC-CL-016's
// closed set of executables this SPEC's files launch).
var codexSpawnLaunchFn = defaultCodexSpawnLaunch

var codexSpawnPaneIdentityFn = defaultCodexSpawnPaneIdentity
var codexSpawnCleanupPaneFn = tmuxKillPane

// defaultCodexSpawnLaunch opens a detached tmux window running codex
// directly. The command string is shell-quoted token-by-token so a tail
// containing spaces, quotes, or $ survives the round trip.
func defaultCodexSpawnLaunch(dir, program string, args []string) error {
	command := buildCodexSpawnCommand(program, args)
	if err := checkCodexInstructionSize(len(command), "spawn command"); err != nil {
		return err
	}
	paneID, err := tmuxSpawnFn(dir, command)
	if err != nil {
		return fmt.Errorf("spawn tmux window: %w", err)
	}
	env := os.Environ()
	factory := factoryLaunchEnabled(env)
	if factory || codexSpawnAnchorFn != nil {
		runID := launchEnvValue(env, config.EnvMoaiKanbanID)
		pid, start, identityErr := codexSpawnPaneIdentityFn(paneID)
		if identityErr == nil && codexSpawnAnchorFn != nil {
			// A pane left running without its lock would be an unanchored
			// writer; close it instead.
			if anchorErr := codexSpawnAnchorFn(pid, start); anchorErr != nil {
				cleanupErr := codexSpawnCleanupPaneFn(paneID)
				var clearErr error
				if factory {
					// REQ-002d holds on this refusal path too: no run keeps
					// the launching process's identity.
					clearErr = clearFactoryRunOwner(dir, runID)
				}
				return fmt.Errorf("anchor spawned Codex worktree: %w", errors.Join(anchorErr, cleanupErr, clearErr))
			}
		}
		if identityErr == nil && factory {
			_, identityErr = registerFactoryLaunchPending(context.Background(), dir, env, pid, start)
		}
		if identityErr == nil && factory {
			// REQ-002b — the pane shape: this launcher returns and exits
			// immediately, so the record-time stamp names a process that is
			// already gone by the time anyone reads it. Restamp with the pane
			// identity the resolver above already probed live, so the run row
			// and the run's role='lead' peer name one process.
			identityErr = stampFactoryRunOwner(dir, runID, pid, start)
		}
		if identityErr != nil {
			cleanupErr := codexSpawnCleanupPaneFn(paneID)
			if !factory {
				return fmt.Errorf("anchor spawned Codex worktree: %w", errors.Join(identityErr, cleanupErr))
			}
			// REQ-002d — refuse, and leave no run carrying the launching
			// process's identity: that identity is known in advance to die, and
			// a run holding it would be retired while its session was meant to
			// be alive.
			clearErr := clearFactoryRunOwner(dir, runID)
			return fmt.Errorf("register spawned factory launch-pending endpoint: %w", errors.Join(identityErr, cleanupErr, clearErr))
		}
	}
	_, _ = fmt.Fprintf(os.Stdout, "Spawned pane %s running `%s` in %s\n", paneID, command, dir)
	_, _ = fmt.Fprintln(os.Stdout, "Switch to it with: tmux select-window -t "+paneID)
	return nil
}

func defaultCodexSpawnPaneIdentity(paneID string) (int, string, error) {
	deadline := time.Now().Add(2 * time.Second)
	for {
		pid, err := tmuxPanePID(paneID)
		if err == nil {
			start, state := homestate.ProbeProcessIdentity(pid)
			if state == homestate.ProcessIdentityLive && start != "" {
				return pid, start, nil
			}
		}
		if !time.Now().Before(deadline) {
			return 0, "", errors.New("spawned Codex pane process identity unavailable")
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// buildCodexSpawnCommand renders the shell command string for the new tmux
// window: the resolved CODEX_HOME as a command-scoped assignment, then the
// codex binary and its argv tail, every token quoted.
//
// The assignment is what makes the two launch paths agree. A tmux window
// inherits the tmux SERVER's environment, not this process's, so without it
// the new window could resolve a different CODEX_HOME than the direct path
// put on its child.
func buildCodexSpawnCommand(program string, args []string) string {
	parts := make([]string, 0, len(args)+10)
	// resolveCodexHomeDir's second result is the source label, not an error.
	if home, _ := resolveCodexHomeDir(); home != "" {
		parts = append(parts, codexHomeEnvVar+"="+shellQuote(home))
	}
	// A tmux server can carry attribution from the session that created it.
	// Codex must bind through its own process identity, never a foreign Claude
	// UUID or an outer launcher's PID.
	parts = append(parts, config.EnvClaudeCodeSessionID+"=", config.EnvMoaiSessionPID+"=")
	for _, key := range []string{
		config.EnvHome,
		config.EnvMoaiKanbanID,
		config.EnvMoaiKanbanBackend,
		// The rest of the kanban launch facts (moai codex -k): a tmux window
		// inherits the server's environment, not this process's.
		config.EnvMoaiKanban,
		config.EnvMoaiKanbanSpec,
		config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanLeadAddr,
		config.EnvMoaiKanbanLeadName,
		config.EnvMoaiFactoryWorker,
		config.EnvMoaiFactoryWorkers,
		config.EnvClaudeProjectDir,
	} {
		if value := os.Getenv(key); value != "" {
			parts = append(parts, key+"="+shellQuote(value))
		}
	}
	parts = append(parts, "exec", shellQuote(program))
	for _, a := range args {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

// splitCodexDashDash splits head (verb position, before --) from tail
// (verbatim passthrough, from -- on). Tokens after -- are never inspected.
func splitCodexDashDash(args []string) (head, tail []string, hasTail bool) {
	for i, a := range args {
		if a == "--" {
			return args[:i], args[i+1:], true
		}
	}
	return args, nil, false
}

// codexWorktreeArg is the -w/--worktree value as the operator supplied it,
// already removed from the token stream. present distinguishes "no flag" from
// "flag with an empty value" — the second is an error, the first is not.
type codexWorktreeArg struct {
	present bool
	value   string
}

// stripCodexWorktreeFlag removes -w/--worktree (and their =value forms) from
// the verb-position tokens and returns the value. Only the head is scanned:
// tokens after -- belong to codex and are never inspected. The four accepted
// token shapes mirror normalizeWorktreeFlag's, so the two surfaces accept the
// same spellings.
func stripCodexWorktreeFlag(head []string) ([]string, codexWorktreeArg) {
	rest := make([]string, 0, len(head))
	arg := codexWorktreeArg{}
	for i := 0; i < len(head); i++ {
		token := head[i]
		switch {
		case token == "-w" || token == "--worktree":
			arg.present = true
			// The value is the next token unless that token is itself a flag
			// (a bare -w triggers an automatically named worktree).
			if i+1 < len(head) && !strings.HasPrefix(head[i+1], "-") {
				arg.value = head[i+1]
				i++
			}
		case strings.HasPrefix(token, "--worktree="):
			arg.present = true
			arg.value = strings.TrimPrefix(token, "--worktree=")
		case strings.HasPrefix(token, "-w="):
			arg.present = true
			arg.value = strings.TrimPrefix(token, "-w=")
		default:
			rest = append(rest, token)
		}
	}
	return rest, arg
}

// resolveCodexWorktreeDir turns the operator's -w value into the directory the
// child starts in. Existing paths are re-entered; a new short name is created
// by resolveOrCreateCodexWorktreeDir before the child starts.
//
// Claude creates its own -w worktree. Codex has no corresponding flag, so
// MoAI consumes it and uses the existing worktree materializer.
//
// Absolute values are validated by the SAME rule cc applies
// (resolveWorktreeL2Path), so an out-of-prefix path fails with cc's own
// diagnostic rather than a second, divergent one.
func resolveCodexWorktreeDir(projectRoot, value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s", codexWorktreeValueDiag)
	}

	path := value
	if filepath.IsAbs(path) {
		if err := resolveWorktreeL2Path([]string{"--worktree", value}); err != nil {
			return "", err
		}
	} else {
		if projectRoot == "" {
			return "", fmt.Errorf("cannot resolve worktree %q: the project root is unresolved", value)
		}
		path = filepath.Join(projectRoot, ".claude", "worktrees", value)
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("worktree %q does not exist at %s", value, path)
	}
	return path, nil
}

// codexWorktreeAdd is shared with the session worktree materializer. The
// function variable also lets tests prove creation without mutating a real
// repository's worktree registry.
var codexWorktreeAdd = gitWorktreeAddReal

// codexWorktreeBase resolves a configured base first and otherwise uses the
// remote default branch. Never branch from whichever shared checkout happens
// to be active: it may be an unrelated card branch.
var codexWorktreeBase = func(projectRoot string) (string, error) {
	if base := config.LoadWorktreeBaseBranch(projectRoot); base != "" && hook.WorktreeBaseBranchResolvable(base) {
		return base, nil
	}
	if out, err := runGitCommand(projectRoot, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"); err == nil {
		return strings.TrimSpace(out), nil
	}
	return "", errors.New("remote default branch is unresolved; configure git_strategy.worktree_base_branch before creating a Codex worktree")
}

// resolveOrCreateCodexWorktreeDir returns the tree the child starts in and
// whether this call created it. An existing tree may already have a writer;
// a created one cannot, so only existing trees go through the writer check.
func resolveOrCreateCodexWorktreeDir(projectRoot, value string) (string, bool, error) {
	if projectRoot == "" {
		return "", false, errors.New("cannot create worktree: project root is unresolved")
	}
	if value == "" {
		value = "codex-" + sessionWorktreeResolveSessionShort()
	}
	if filepath.IsAbs(value) {
		path, err := resolveCodexWorktreeDir(projectRoot, value)
		return path, false, err
	}
	if value == "." || value == ".." || filepath.Base(value) != value || strings.ContainsAny(value, `/\\`) {
		return "", false, fmt.Errorf("invalid worktree name %q", value)
	}
	path := filepath.Join(projectRoot, sessionWorktreeSubdir, value)
	if info, err := os.Lstat(path); err == nil {
		if !info.IsDir() {
			return "", false, fmt.Errorf("worktree path %s is not a directory", path)
		}
		return path, false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, fmt.Errorf("inspect worktree path: %w", err)
	}
	base, err := codexWorktreeBase(projectRoot)
	if err != nil {
		return "", false, err
	}
	// Resolve the base to a commit BEFORE creating the tree, so the check
	// below compares the new HEAD with what the base meant at decision time.
	baseCommit, err := codexResolveBaseCommit(projectRoot, base)
	if err != nil {
		return "", false, err
	}
	branch := SessionWorktreeBranchPrefix + value
	if _, err := codexWorktreeAdd(path, branch, base); err != nil {
		return "", false, fmt.Errorf("create worktree %q: %w", value, err)
	}
	if err := codexWorktreeBaseCheck(path, baseCommit, base); err != nil {
		return "", true, err
	}
	return path, true, nil
}

// Worktree anchor seams. The capture harness pins them open for the
// launch-mechanics tests, which use plain directories; the anchor tests run
// the real bodies.
var (
	// codexWorktreeWriterCheck refuses a tree another live session is anchored in.
	codexWorktreeWriterCheck = worktreeWriterRefusal
	// codexWorktreeAnchorLock places this launch's lock on the tree.
	codexWorktreeAnchorLock = placeCodexAnchorLock
	// codexResolveBaseCommit resolves the creation base to a commit.
	codexResolveBaseCommit = resolveBaseCommitReal
	// codexWorktreeBaseCheck verifies a created tree starts at that commit.
	codexWorktreeBaseCheck = verifyCodexWorktreeBase
	// codexAnchorReplaceHook runs between a launcher's first read of a dead
	// lock and its replacement guard; tests use it to line two launchers up.
	codexAnchorReplaceHook func()
	// codexSpawnAnchorFn, when set, anchors the tree to the spawned pane's
	// process on the new-window path.
	codexSpawnAnchorFn func(pid int, start string) error
)

// codexAnchorReplaceGuardName is the per-tree replacement guard, created
// exclusively inside the tree's private git directory.
const codexAnchorReplaceGuardName = "moai-anchor-replace"

// resolveBaseCommitReal resolves base to the commit it names now.
func resolveBaseCommitReal(projectRoot, base string) (string, error) {
	out, err := runGitCommand(projectRoot, "rev-parse", "--verify", "--quiet", base+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("resolve worktree base %q: %w", base, err)
	}
	return strings.TrimSpace(out), nil
}

// verifyCodexWorktreeBase refuses a created tree whose HEAD is not the
// resolved base commit. The tree is left in place: it is evidence of what the
// materializer actually did, and deleting it would destroy that evidence.
func verifyCodexWorktreeBase(tree, baseCommit, base string) error {
	out, err := runGitCommand(tree, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("verify worktree base of %s: %w", tree, err)
	}
	if head := strings.TrimSpace(out); head != baseCommit {
		return fmt.Errorf("worktree %s HEAD %s does not match the resolved base %s (%s); refusing to launch - the worktree is left in place for inspection",
			tree, head, base, baseCommit)
	}
	return nil
}

// @MX:WARN: [AUTO] lock replacement is a read-modify-write on shared git state
// @MX:REASON: two launchers can see the same dead lock at once; without the
// exclusive guard and the re-read under it, the later unlock erases the
// earlier launcher's fresh lock and both believe they own the tree.
//
// placeCodexAnchorLock locks tree for pid so the shared anchor decision sees
// the session for its lifetime. An unlocked tree is locked; a lock already
// naming pid is kept; a lock whose holder is confirmed dead is replaced, but
// only under an exclusively created guard and only if the lock is unchanged
// when re-read under it. A live or undetermined holder is never displaced and
// nothing is ever forced.
func placeCodexAnchorLock(tree string, pid int, start string) error {
	reason := session.CodexAnchorLockReason(filepath.Base(tree), pid, start)
	lock, err := readWorktreeLock(tree)
	if err != nil {
		return fmt.Errorf("anchor worktree %s: %w", tree, err)
	}
	if !lock.Locked {
		return lockCodexTree(tree, reason, pid)
	}
	if held, ok := session.LockReasonPID(lock.Reason); ok && held == pid {
		return nil
	}
	if !session.LockHolderConfirmedDead(lock) {
		return fmt.Errorf("anchor worktree %s: the worktree lock is held (%s); a lock whose holder is live or undetermined is never replaced", tree, lock.Reason)
	}
	if codexAnchorReplaceHook != nil {
		codexAnchorReplaceHook()
	}
	gitDir, err := runGitCommand(tree, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return fmt.Errorf("anchor worktree %s: locate its git directory: %w", tree, err)
	}
	guard := filepath.Join(strings.TrimSpace(gitDir), codexAnchorReplaceGuardName)
	f, err := os.OpenFile(guard, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("anchor worktree %s: another launcher is replacing the worktree lock (%v); if no launcher is running, remove %s and retry", tree, err, guard)
	}
	_ = f.Close()
	defer func() { _ = os.Remove(guard) }()
	again, err := readWorktreeLock(tree)
	if err != nil || again != lock {
		return fmt.Errorf("anchor worktree %s: the worktree lock changed while it was being replaced", tree)
	}
	if _, err := runGitCommand(tree, "worktree", "unlock", tree); err != nil {
		return fmt.Errorf("anchor worktree %s: release the dead holder's lock: %w", tree, err)
	}
	return lockCodexTree(tree, reason, pid)
}

// lockCodexTree writes the lock and reads it back: a lock that does not name
// pid afterwards belongs to someone else, whatever `git worktree lock` said.
func lockCodexTree(tree, reason string, pid int) error {
	if _, err := runGitCommand(tree, "worktree", "lock", "--reason", reason, tree); err != nil {
		return fmt.Errorf("anchor worktree %s: git worktree lock: %w", tree, err)
	}
	after, err := readWorktreeLock(tree)
	if err != nil {
		return fmt.Errorf("anchor worktree %s: read back the worktree lock: %w", tree, err)
	}
	if held, ok := session.LockReasonPID(after.Reason); !ok || held != pid {
		return fmt.Errorf("anchor worktree %s: the worktree lock names %q, not this launcher", tree, after.Reason)
	}
	return nil
}

// codexChildEnv is the environment handed to the codex child: the parent's own
// environment with the RESOLVED CODEX_HOME appended. Appending (rather than
// replacing) keeps the rest of the parent environment intact, and last-wins
// makes the appended entry the value the child reads.
//
// Explicit rather than ambient: where the parent has no CODEX_HOME at all,
// there is nothing to inherit, and the child would otherwise resolve its own
// default independently of the value the readout reports.
func codexChildEnv() []string {
	env := make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key == config.EnvClaudeCodeSessionID || key == config.EnvMoaiSessionPID {
			continue
		}
		env = append(env, entry)
	}
	// resolveCodexHomeDir's second result is the source label, not an error.
	if home, _ := resolveCodexHomeDir(); home != "" {
		env = append(env, codexHomeEnvVar+"="+home)
	}
	return env
}

// codexCmd — the launcher-family sibling of cc/glm/cg (same group, same
// DisableFlagParsing discipline: --spawn and -- passthrough are handled by
// the launcher itself). SilenceErrors/SilenceUsage keep every diagnostic
// byte-identical to the named constants above (cobra would otherwise prefix
// "Error: " and append the usage block, breaking the exact-match cells).
var codexCmd = &cobra.Command{
	Use:   "codex [cli | status | app]",
	Short: "Codex launcher: launch the Codex CLI, or print the readiness readout",
	Long: "Launch Codex, or ask what is installed and wired.\n" +
		"\n" +
		"Called with no verb, this launches the Codex CLI at the project root.\n" +
		"The readiness readout moved to an explicit alias: moai codex status\n" +
		"reports six rows and starts nothing - the codex binary, CODEX_HOME,\n" +
		"the auth provider, the project wiring, the generated agent TOMLs, and\n" +
		"the harness entry. An incomplete wiring row is informational, not an\n" +
		"error: moai init --llm gpt generates the .codex wiring files.\n" +
		"Common local guidance shared with Claude, then Codex-specific local\n" +
		"guidance: both non-empty project-root files are injected as\n" +
		"developer instructions, in that order.\n" +
		"\n" +
		"  moai codex            launch the Codex CLI at the project root\n" +
		"  moai codex cli        the same launch, named explicitly\n" +
		"  moai codex status     print the readiness readout (starts nothing)\n" +
		"  moai codex app        launch the Codex desktop app (codex app)\n" +
		"  moai codex -f        start a Factory lead\n" +
		"  moai codex -f agent  join the active Factory run as the next agent-n\n" +
		"  moai codex -f agent-2 --factory-run <id>  join a named run\n" +
		"  -w [worktree]         re-enter or create a worktree from the remote\n" +
		"                        default branch; omit name to generate one\n" +
		"  --spawn               open the launch in a new tmux window\n" +
		"  -- <codex-args...>    arguments after -- pass to codex verbatim",
	Example: "  # Launch the Codex CLI here\n" +
		"  moai codex\n" +
		"\n" +
		"  # Show what is installed, resolved, and wired (starts nothing)\n" +
		"  moai codex status\n" +
		"\n" +
		"  # Launch, passing arguments through verbatim\n" +
		"  moai codex -- --model o3\n" +
		"\n" +
		"  # Re-enter or create a worktree\n" +
		"  moai codex -w feat-login\n" +
		"\n" +
		"  # Launch the desktop app in a new tmux window\n" +
		"  moai codex app --spawn\n" +
		"\n" +
		"  # Join a Factory run with an automatically numbered agent\n" +
		"  moai codex -f agent --factory-run <id>",
	GroupID:            "launch",
	DisableFlagParsing: true,
	SilenceErrors:      true,
	SilenceUsage:       true,
	RunE:               runCodex,
}

func init() {
	rootCmd.AddCommand(codexCmd)
	// Registered for --help visibility and the neutrality flag-usage scan
	// (AC-CL-013); DisableFlagParsing means the launcher itself strips the
	// flag via stripSpawnFlag before the verb lookup.
	codexCmd.Flags().Bool("spawn", false, "open the launch in a new tmux window (cli/app verbs only)")
}

// runCodex routes the invocation: --help first (DisableFlagParsing means
// cobra never intercepts it), then --spawn stripping, then the closed-set
// verb lookup, then the readout or launch path.
func runCodex(cmd *cobra.Command, args []string) error {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return cmd.Help()
		}
		if a == "--" {
			break
		}
	}

	args, spawn := stripSpawnFlag(args)
	head, tail, hasTail := splitCodexDashDash(args)
	var factoryRun string
	var runErr error
	head, factoryRun, runErr = stripFactoryRunFlag(head)
	if runErr != nil {
		return runErr
	}
	// -f is consumed before the verb lookup (same precedence as -w): the
	// factory token selects this session's factory role and is never a codex
	// verb. The env is applied only on a launch path — a readout with -f is
	// a usage error, exactly as -w with a readout is.
	head, factoryLead, factoryRole, factoryLane, ferr := stripCodexFactoryFlag(head)
	if ferr != nil {
		return ferr
	}
	// -k is consumed next, before any factory state is applied, so a -k/-f
	// mix is refused before either mode touches the environment.
	head, kanbanEntry, kerr := stripCodexKanbanFlag(head)
	if kerr == nil && kanbanEntry.enabled && (factoryLead || factoryRole != "" || factoryLane != "") {
		kerr = errors.New(codexKanbanUsageDiag)
	}
	if kerr != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), kerr.Error())
		return &exitCodeError{code: 1}
	}
	if spawn && (factoryLead || factoryRole != "" || factoryLane != "") {
		// Let the pane's MoAI process own the broker endpoint and App Server.
		// Claiming a slot here would race the pane and leave a dead owner.
		if err := checkSpawnPrereqs(); err != nil {
			return err
		}
		program, err := os.Executable()
		if err != nil {
			return err
		}
		if err := codexSpawnLaunchFn(launchProjectRoot(), program, append([]string{"codex"}, args...)); err != nil {
			return fmt.Errorf("spawn managed Factory Codex: %w", err)
		}
		return nil
	}
	var factoryRestore func()
	if factoryLead || factoryRole != "" || factoryLane != "" {
		var applyErr error
		factoryRestore, applyErr = applyCodexFactoryEntry(cmd, factoryRole, factoryLane)
		if applyErr != nil {
			return applyErr
		}
		defer factoryRestore()
		restoreRun, selectErr := enterSelectedFactoryRun(launchProjectRoot(), factoryRun, factoryRole != "" || factoryLane != "")
		if selectErr != nil {
			return selectErr
		}
		defer restoreRun()
		if factoryLead {
			if err := recordFactoryRunStart(launchProjectRoot(), os.Getenv(config.EnvMoaiKanbanID), codexFactoryBackend, ""); err != nil {
				return fmt.Errorf("record Codex factory run: %w", err)
			}
		}
	} else if factoryRun != "" {
		return fmt.Errorf("--factory-run requires -f/--factory")
	}
	// -w is consumed before the verb lookup so its tokens can never be
	// mistaken for a verb, and so the verb position keeps its one-token shape.
	head, worktree := stripCodexWorktreeFlag(head)
	if len(head) > 1 {
		return codexUsageFailure(cmd)
	}
	verb := ""
	if len(head) == 1 {
		verb = head[0]
	}
	kind, ok := codexVerbRouting[verb]
	if !ok {
		return codexUsageFailure(cmd)
	}

	if kind.launches() {
		if kanbanEntry.enabled {
			defer applyCodexKanbanEntry(cmd, kanbanEntry)()
		}
		return runCodexLaunch(cmd, kind, tail, spawn, worktree)
	}
	if worktree.present || kanbanEntry.enabled || factoryLead || factoryRole != "" || factoryLane != "" {
		// A readout starts no process, so it has no working directory to
		// point anywhere — and no factory role to enter either.
		return codexUsageFailure(cmd)
	}
	if spawn {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexSpawnReadoutDiag)
		return &exitCodeError{code: 1}
	}
	if hasTail {
		return codexUsageFailure(cmd)
	}
	return runCodexReadout(cmd)
}

// codexUsageFailure prints the usage constant to stderr and fails with rc 1.
func codexUsageFailure(cmd *cobra.Command) error {
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexUsageDiag)
	return &exitCodeError{code: 1}
}

// runCodexReadout renders M2's six rows VERBATIM to stdout and succeeds —
// whatever the probes found (AC-CL-004/006: informational, rc 0, stdout only).
func runCodexReadout(cmd *cobra.Command) error {
	r := probeCodexReadiness(cmd.Context())
	for _, row := range r.rows() {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), row)
	}
	return nil
}

// runCodexLaunch resolves the binary and hands off to the direct or spawn
// path. A missing binary is a single-line install hint, launch count 0
// (AC-CL-011).
func runCodexLaunch(cmd *cobra.Command, kind codexVerb, tail []string, spawn bool, worktree codexWorktreeArg) error {
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexInstallHint)
		return &exitCodeError{code: 1}
	}

	// The launch cwd is the PROJECT ROOT, not the process cwd (AC-CL-002):
	// a call from a subdirectory still launches at the root. An unresolvable
	// root degrades to the process cwd rather than refusing to launch.
	projectRoot := ""
	if root, rerr := findProjectRootFn(); rerr == nil && root != "" {
		projectRoot = root
	} else if cwd, gerr := os.Getwd(); gerr == nil {
		projectRoot = cwd
	}

	// SPEC-CODEX-INIT-001: the init-offer gate — the ONE call site every
	// launch form passes through right before launching, the bare form
	// included. The gate takes no spawn argument: both launch paths cross the
	// same function (REQ-CI-002).
	if err := codexInitOfferGate(cmd, projectRoot); err != nil {
		return err
	}
	localArgs, err := codexLocalDeveloperInstructionArgs(projectRoot)
	if err != nil {
		return fmt.Errorf("load Codex local instructions: %w", err)
	}
	if len(localArgs) > 0 && codexHasDeveloperOverride(tail) {
		return errors.New(codexDuplicateLocalOverrideDiag)
	}
	// Prepare the worktree only after read-only launch validation. A failed
	// init offer or instruction check must not leave an orphan worktree.
	dir := projectRoot
	if worktree.present {
		resolved, created, werr := resolveOrCreateCodexWorktreeDir(projectRoot, worktree.value)
		if werr == nil && !created {
			// An existing tree may already have a writer; refuse to be the second.
			werr = codexWorktreeWriterCheck(resolved)
		}
		if werr != nil {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), werr.Error())
			return &exitCodeError{code: 1}
		}
		dir = resolved
	}
	childArgs := append(localArgs, codexChildArgs(kind, tail)...)
	req := codexLaunchRequest{Program: binaryPath, Args: childArgs, Dir: dir}
	if factoryLaunchEnabled(os.Environ()) && !spawn {
		if worktree.present {
			if err := codexWorktreeAnchorLock(dir, codexDirectAnchorPID(), homestate.CurrentProcessFingerprint()); err != nil {
				return err
			}
		}
		return runManagedFactoryCodex(req)
	}
	if spawn {
		if worktree.present {
			// The process that becomes Codex is the pane's; it exists only
			// after the spawn, so the lock is placed from inside it.
			codexSpawnAnchorFn = func(pid int, start string) error { return codexWorktreeAnchorLock(dir, pid, start) }
			defer func() { codexSpawnAnchorFn = nil }()
		}
		return codexSpawnLaunch(req)
	}
	if worktree.present {
		// The lock names the process that becomes Codex, before it starts.
		if err := codexWorktreeAnchorLock(dir, codexDirectAnchorPID(), homestate.CurrentProcessFingerprint()); err != nil {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			return &exitCodeError{code: 1}
		}
	}
	return codexDirectLaunch(req)
}

// codexDirectLaunch assembles the child (stdio = the parent's OWN os.Stdin /
// os.Stdout / os.Stderr values — the interactive-tty precondition,
// AC-CL-002), runs it through the seam, and propagates the child's exit code
// verbatim (AC-CL-002 rc axis, AC-CL-016).
func codexDirectLaunch(req codexLaunchRequest) error {
	for _, arg := range req.Args {
		if strings.HasPrefix(arg, "developer_instructions=") {
			if err := checkCodexInstructionSize(len(arg), "direct instruction token"); err != nil {
				return err
			}
		}
	}
	c := exec.Command(req.Program, req.Args...)
	c.Dir = req.Dir
	c.Env = codexChildEnv()
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := codexDirectLaunchFn(c)
	if err == nil {
		return nil
	}
	return codexPropagateLaunchError(err)
}

// codexPropagateLaunchError maps a launch failure onto the exit-code
// discipline: a genuine subprocess exit becomes a DELIBERATE exitCodeError
// carrying the child's code (this is the launcher's contract — unlike every
// other wrap site, the child's code IS the answer), an ExitCoder the seam
// handed us propagates as-is, and anything else (start failure) is described
// via StatusDetail — never a raw %w chain (ResolveExitCode refuses those by
// design; the conversion here is what makes propagation deliberate).
func codexPropagateLaunchError(err error) error {
	var raw *exec.ExitError
	if errors.As(err, &raw) {
		return &exitCodeError{code: raw.ExitCode()}
	}
	if code, ok := ResolveExitCode(err); ok {
		return &exitCodeError{code: code}
	}
	return fmt.Errorf("codex: %s", execerr.StatusDetail(err))
}

// codexSpawnLaunch checks the shared --spawn preconditions (the SAME
// diagnostics moai cc --spawn emits, byte-identical — AC-CL-003) and opens
// the new window through the seam.
func codexSpawnLaunch(req codexLaunchRequest) error {
	if err := checkCodexInstructionSize(len(buildCodexSpawnCommand(req.Program, req.Args)), "spawn command"); err != nil {
		return err
	}
	if err := checkSpawnPrereqs(); err != nil {
		return err
	}
	if err := codexSpawnLaunchFn(req.Dir, req.Program, req.Args); err != nil {
		return fmt.Errorf("spawn tmux window: %w", err)
	}
	return nil
}
