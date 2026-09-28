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
// working directory at an existing worktree. CODEX_HOME
// reaches the child as an explicit environment entry rather than by ambient
// inheritance, on both the direct and the new-window path.
//
// The readout rows come from M2's codexReadiness VERBATIM (AC-CL-004 — the
// command surface never re-words a row), and the binary/auth values come from
// the shared probe, so no second classification path forks here (REQ-CL-007).
// The status readout never writes; -w requires an existing worktree. POSIX direct
// launch replaces moai with Codex (the -w lock names that one pid); Windows
// retains the child Start/wait path. Kanban (-k) remains unsupported.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/execerr"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/spf13/cobra"
)

// Diagnostics — named constants compared with == by the tests (AC-CL-002's
// "usage constant", AC-CL-011's single-line install action).
const (
	// codexUsageDiag is the one-line usage diagnostic every unknown token
	// receives (AC-CL-002). Byte-identical for all six probe tokens so the
	// rejection cannot leak which token was seen.
	codexUsageDiag = "unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app"

	// Codex only enters an already-created worktree; -w needs its name or path.
	codexWorktreeValueDiag = "-w requires an existing worktree name or path"

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

// codexLocalDeveloperInstructionArgs is codexLocalDeveloperInstructions without
// the list of files read, for callers that only need the argv pair.
func codexLocalDeveloperInstructionArgs(projectRoot string) ([]string, error) {
	args, _, err := codexLocalDeveloperInstructions(projectRoot)
	return args, err
}

// @MX:ANCHOR: [AUTO] Single local-instruction producer for every launch form.
// @MX:REASON: Compose source order and framing before encoding exactly one override.
// @MX:SPEC: SPEC-CODEX-LOCALMD-001
// codexLocalDeveloperInstructions reads both local inputs fresh on each
// launch, AGENTS.local.md ahead of CLAUDE.local.md (SPEC-INSTRUCTION-FILES-UNIFY-001
// REQ-IFU-006). JSON string encoding preserves UTF-8 in a TOML basic string.
// It stays pure: it returns the names it read so the caller, which holds the
// diagnostic stream, decides on any advisory (REQ-IFU-007) — nothing addressed
// to the operator enters the payload.
func codexLocalDeveloperInstructions(projectRoot string) (args, read []string, err error) {
	var payload strings.Builder
	for _, name := range []string{codexLocalInstructionName, codexClaudeLocalName} {
		body, err := readCodexLocalInstruction(projectRoot, name)
		if err != nil {
			return nil, nil, err
		}
		if len(body) == 0 {
			continue
		}
		if payload.Len() > 0 {
			payload.WriteByte('\n')
		}
		// REQ-IFU-008: the literal name of the file actually read.
		fmt.Fprintf(&payload, "<!-- source: %s -->\n", name)
		payload.Write(body)
		read = append(read, name)
	}
	if payload.Len() == 0 {
		return nil, nil, nil
	}
	encoded, _ := json.Marshal(payload.String()) // A string is always JSON-encodable.
	return []string{"-c", "developer_instructions=" + string(encoded)}, read, nil
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
	Program    string
	Args       []string // argv AFTER the program token: [verb, passthrough...]
	Dir        string
	FactoryEnv []string
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
func defaultCodexSpawnLaunch(dir, program string, args, factoryEnv []string) error {
	command := buildCodexSpawnCommandWithEnv(program, args, factoryEnv)
	if err := checkCodexInstructionSize(len(command), "spawn command"); err != nil {
		return err
	}
	paneID, err := tmuxSpawnFn(dir, command)
	if err != nil {
		return fmt.Errorf("spawn tmux window: %w", err)
	}
	if codexSpawnAnchorFn != nil || codexExplicitFactoryEnv(factoryEnv) {
		var pending factorymsg.Peer
		// -w: the pane's process is the one that becomes Codex, so the worktree
		// lock is placed on its identity. A pane left running without its lock
		// would be an unanchored writer; close it instead.
		pid, start, identityErr := codexSpawnPaneIdentityFn(paneID)
		if identityErr == nil && codexSpawnAnchorFn != nil {
			identityErr = codexSpawnAnchorFn(pid, start)
		}
		if identityErr == nil && codexExplicitFactoryEnv(factoryEnv) {
			pending, identityErr = registerFactoryLaunchPending(context.Background(), dir, factoryEnv, pid, start)
		}
		if identityErr == nil && codexExplicitFactoryEnv(factoryEnv) && launchEnvValue(factoryEnv, config.EnvMoaiFactoryWorker) != "" {
			identityErr = stampCodexLaneClaim(dir, factoryEnv, pid)
		}
		if identityErr == nil && launchEnvValue(factoryEnv, config.EnvMoaiFactoryWorker) == "" && codexExplicitFactoryEnv(factoryEnv) {
			identityErr = stampFactoryRunOwner(dir, launchEnvValue(factoryEnv, config.EnvMoaiKanbanID), pid, start)
		}
		if identityErr != nil {
			cleanupErr := codexSpawnCleanupPaneFn(paneID)
			cleanupErr = errors.Join(cleanupErr, rollbackFactoryLaunchPending(context.Background(), dir, pending))
			if codexExplicitFactoryEnv(factoryEnv) && launchEnvValue(factoryEnv, config.EnvMoaiFactoryWorker) == "" {
				cleanupErr = errors.Join(cleanupErr, clearFactoryRunOwner(dir, launchEnvValue(factoryEnv, config.EnvMoaiKanbanID)))
			}
			return fmt.Errorf("prepare spawned Codex: %w", errors.Join(identityErr, cleanupErr))
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

// codexLaneLaunchEnvKeys are the lane launch keys an outer session may
// export. A plain Codex launch scrubs them. An explicit -f launch supplies a
// freshly selected factory identity after the scrub.
var codexLaneLaunchEnvKeys = []string{
	config.EnvMoaiKanban,
	config.EnvMoaiKanbanID,
	config.EnvMoaiKanbanSpec,
	config.EnvMoaiKanbanLabel,
	config.EnvMoaiKanbanLeadAddr,
	config.EnvMoaiKanbanLeadName,
	config.EnvMoaiKanbanBackend,
	config.EnvMoaiKanbanCard,
	config.EnvMoaiKanbanSettingsInjected,
	config.EnvMoaiFactoryWorker,
	config.EnvMoaiFactoryWorkers,
	config.EnvFactoryRole,
}

// codexSpawnForwardedEnv lists the variables buildCodexSpawnCommand copies
// from this process onto the tmux command line when they are set. Tests that
// assert the exact command pin each of them, so a lane session's exports
// cannot change the expected string.
//
// The plain spawn command blanks every lane key; the explicit factory path
// adds its selected values afterward.
var codexSpawnForwardedEnv = []string{
	config.EnvHome,
	config.EnvClaudeProjectDir,
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
	return buildCodexSpawnCommandWithEnv(program, args, nil)
}

func buildCodexSpawnCommandWithEnv(program string, args, factoryEnv []string) string {
	parts := make([]string, 0, len(args)+len(codexLaneLaunchEnvKeys)+10)
	// resolveCodexHomeDir's second result is the source label, not an error.
	if home, _ := resolveCodexHomeDir(); home != "" {
		parts = append(parts, codexHomeEnvVar+"="+shellQuote(home))
	}
	// A tmux server can carry attribution from the session that created it.
	// Codex must bind through its own process identity, never a foreign Claude
	// UUID or an outer launcher's PID.
	parts = append(parts, config.EnvClaudeCodeSessionID+"=", config.EnvMoaiSessionPID+"=")
	// The same holds for a Claude lane's kanban/factory identity: the pane
	// would otherwise inherit it from this process or the tmux server and
	// present itself as a peer of the lane's run.
	for _, key := range codexLaneLaunchEnvKeys {
		parts = append(parts, key+"=")
	}
	for _, entry := range factoryEnv {
		key, value, _ := strings.Cut(entry, "=")
		parts = append(parts, key+"="+shellQuote(value))
	}
	for _, key := range codexSpawnForwardedEnv {
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
			// The value is the next token unless that token is itself a flag.
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
// child starts in. The tree must already exist; Codex has no CLI worktree
// creation flag, so MoAI consumes -w without forwarding it to Codex.
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
		if err := resolveWorktreeL2Path([]string{"--worktree", value}, os.Stderr); err != nil {
			return "", err
		}
	} else {
		if projectRoot == "" {
			return "", fmt.Errorf("cannot resolve worktree %q: the project root is unresolved", value)
		}
		if value == "." || value == ".." || filepath.Base(value) != value || strings.ContainsAny(value, `/\\`) {
			return "", fmt.Errorf("invalid worktree name %q", value)
		}
		modern := filepath.Join(projectRoot, sessionWorktreeSubdir, value)
		legacy := filepath.Join(projectRoot, claudeNativeWorktreeSubdir, value)
		_, modernErr := os.Lstat(modern)
		_, legacyErr := os.Lstat(legacy)
		if modernErr == nil && legacyErr == nil {
			return "", fmt.Errorf("worktree %q exists in both %s and %s; use an absolute path", value, modern, legacy)
		}
		if modernErr == nil {
			path = modern
		} else if legacyErr == nil {
			path = legacy
		} else if errors.Is(modernErr, os.ErrNotExist) && errors.Is(legacyErr, os.ErrNotExist) {
			return "", fmt.Errorf("worktree %q does not exist at %s or %s; create it before launching Codex", value, modern, legacy)
		} else {
			return "", fmt.Errorf("inspect worktree %q: %w", value, errors.Join(modernErr, legacyErr))
		}
	}

	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("worktree %q does not exist at %s; create it before launching Codex", value, path)
		}
		return "", fmt.Errorf("inspect worktree %q: %w", value, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("worktree path %s is not a directory", path)
	}
	return path, nil
}

// Worktree anchor seams. The capture harness pins them open for the
// launch-mechanics tests, which use plain directories; the anchor tests run
// the real bodies.
var (
	// codexWorktreeWriterCheck refuses a tree another live session is anchored in.
	codexWorktreeWriterCheck = worktreeWriterRefusal
	// codexWorktreeAnchorLock places this launch's lock on the tree.
	codexWorktreeAnchorLock = placeCodexAnchorLock
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
//
// Removed from the inherited environment: the foreign attribution pair
// (CLAUDE_CODE_SESSION_ID, MOAI_SESSION_PID) and the lane launch keys,
// so a codex run inside a Claude lane never presents that lane's identity
// Posture keys such as MOAI_AUTONOMY_TIER carry no identity and pass through.
func codexChildEnv() []string {
	drop := map[string]bool{config.EnvClaudeCodeSessionID: true, config.EnvMoaiSessionPID: true}
	for _, key := range codexLaneLaunchEnvKeys {
		drop[key] = true
	}
	env := make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if drop[key] {
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
		"  -w <worktree>         launch in an existing worktree (never creates one)\n" +
		"  -f                   start a factory run as leader (CLI only)\n" +
		"  -f lane              join the active factory as the next lane (CLI only)\n" +
		"  -f lane-<n>          join as a numbered lane (CLI only)\n" +
		"  Factory sessions show launch_pending until their first prompt binds a session UUID.\n" +
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
		"  # Enter an existing worktree\n" +
		"  moai codex -w feat-login\n" +
		"\n" +
		"  # Launch the desktop app in a new tmux window\n" +
		"  moai codex app --spawn",
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
	// Kanban remains unsupported. Tokens after -- are Codex's own.
	if diag := codexEntryRefusal(head); diag != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diag)
		return &exitCodeError{code: 1}
	}
	head, factoryEntry, err := parseCodexFactoryEntry(head)
	if err != nil {
		return err
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
	if factoryEntry.Enabled && kind == codexVerbLaunchApp {
		return fmt.Errorf("-f/--factory requires the Codex CLI launch")
	}

	if kind.launches() {
		return runCodexLaunch(cmd, kind, tail, spawn, worktree, factoryEntry)
	}
	if worktree.present || factoryEntry.Enabled {
		// A readout starts no process, so it has no working directory to
		// point anywhere.
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

// The Kanban entry remains unsupported by the Codex launcher.
const (
	codexKanbanRefusalDiag = kanbanUnsupportedBackendSentinel +
		": moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead"
)

// codexEntryRefusal scans the head (the tokens before --) for Kanban entry.
func codexEntryRefusal(head []string) string {
	for _, a := range head {
		switch {
		case a == kanbanFlagShort || a == kanbanFlagLong ||
			strings.HasPrefix(a, kanbanFlagShort+"=") || strings.HasPrefix(a, kanbanFlagLong+"="):
			return codexKanbanRefusalDiag
		}
	}
	return ""
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
func runCodexLaunch(cmd *cobra.Command, kind codexVerb, tail []string, spawn bool, worktree codexWorktreeArg, factoryEntry factoryFlagParse) error {
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
	localArgs, localRead, err := codexLocalDeveloperInstructions(projectRoot)
	if err != nil {
		return fmt.Errorf("load Codex local instructions: %w", err)
	}
	// REQ-IFU-007: the fallback advisory goes to the diagnostic stream, never
	// into developer_instructions.
	if msg := localInstructionsAdvisory(slices.Contains(localRead, codexLocalInstructionName),
		slices.Contains(localRead, codexClaudeLocalName)); msg != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), msg)
	}
	if len(localArgs) > 0 && codexHasDeveloperOverride(tail) {
		return errors.New(codexDuplicateLocalOverrideDiag)
	}
	// Resolve the existing worktree only after read-only launch validation.
	dir := projectRoot
	if worktree.present {
		resolved, werr := resolveCodexWorktreeDir(projectRoot, worktree.value)
		if werr == nil {
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
	if spawn {
		if err := checkSpawnPrereqs(); err != nil {
			return err
		}
	}
	if factoryEntry.Enabled {
		// The run is project-scoped even when Codex enters an existing worktree.
		restore, ferr := enterCodexFactory(projectRoot, factoryEntry)
		if ferr != nil {
			return ferr
		}
		defer restore()
		req.FactoryEnv = codexFactoryEnv(factoryEntry)
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
	c.Env = append(c.Env, req.FactoryEnv...)
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
	if err := checkCodexInstructionSize(len(buildCodexSpawnCommandWithEnv(req.Program, req.Args, req.FactoryEnv)), "spawn command"); err != nil {
		return err
	}
	if err := checkSpawnPrereqs(); err != nil {
		return err
	}
	if err := codexSpawnLaunchFn(req.Dir, req.Program, req.Args, req.FactoryEnv); err != nil {
		return fmt.Errorf("spawn tmux window: %w", err)
	}
	return nil
}
