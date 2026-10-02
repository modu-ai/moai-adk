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
// retains the child Start/wait path. The kanban entry (-k) stays refused
// (SPEC-CODEX-FACTORY-RETIRE-001). The factory surface narrowed to exactly
// the lane entry `-l` / `--lane` (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-003,
// whose trigger SPEC-LAUNCHER-ENTRY-FLAGS-001 renamed from `-f lane`): a
// supervising loop that leases each card on the parent checkout and starts one
// interactive Codex session in the card's own worktree; `-f` in every shape
// keeps its refusal (REQ-SD-004).

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
	"github.com/modu-ai/moai-adk/internal/factory"
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
	codexUsageDiag = "unknown verb - usage: moai codex [cli] [-w <worktree>] [-l] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app"

	// Codex only enters an already-created worktree; -w needs its name or path.
	codexWorktreeValueDiag = "-w requires an existing worktree name or path"

	// codexInstallHint is the single diagnostic line the launch verbs print
	// when the codex binary is unresolved (AC-CL-011: exactly one line, exact
	// match; the readout forms still succeed in the same state).
	codexInstallHint = "codex not found - install the Codex CLI first: https://developers.openai.com/codex/cli"

	// codexSpawnReadoutDiag rejects --spawn on the readout forms (AC-CL-003):
	// a readout is not something to open a new window for.
	codexSpawnReadoutDiag = "--spawn applies to the launch verbs only (moai codex cli --spawn / moai codex app --spawn)"

	// codexDebugReadoutDiag rejects -d/--debug on the readout forms
	// (SPEC-CODEX-DEBUG-MODE-001 REQ-003), mirroring the --spawn discipline:
	// a readout starts nothing there is nothing to trace.
	codexDebugReadoutDiag           = "-d/--debug applies to the launch verbs only (moai codex cli -d / moai codex app -d)"
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
	// Env is the ASSEMBLED child environment (codexChildEnv, the debug
	// RUST_LOG injection when debug mode is on, then the factory identity):
	// assembled by the launcher ahead of the pre-seam reports so the
	// child-env assembly step is a traced step like the others
	// (SPEC-CODEX-DEBUG-MODE-001 REQ-005/REQ-009/REQ-010).
	Env []string
	// Debug carries the debug linkage decision the assembly applied.
	Debug bool
	// timing is the launch's pre-exec collector — nil when the launch is
	// neither a debug launch nor a lane launch.
	timing *factoryLaunchTiming
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

// managedFactoryCodexLaunchFunc is the managed-divert seam (SPEC-FACTORY-
// MANAGED-SESSION-001 M3): it receives the binary, the argv with the
// program name at args[0], the launcher env, and the directory the launch
// resolved (project root or the -w worktree — the one the doors launch in).
// Tests override it to observe the divert without starting a real App Server.
var managedFactoryCodexLaunchFunc = func(bin string, args, env []string, dir string) error {
	return runManagedFactoryCodex(bin, args, env, dir, os.Stdin)
}

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
			identityErr = stampFactoryRunOwner(dir, launchEnvValue(factoryEnv, config.EnvFactoryRunID), pid, start)
		}
		if identityErr != nil {
			cleanupErr := codexSpawnCleanupPaneFn(paneID)
			cleanupErr = errors.Join(cleanupErr, rollbackFactoryLaunchPending(context.Background(), dir, pending))
			if codexExplicitFactoryEnv(factoryEnv) && launchEnvValue(factoryEnv, config.EnvMoaiFactoryWorker) == "" {
				cleanupErr = errors.Join(cleanupErr, clearFactoryRunOwner(dir, launchEnvValue(factoryEnv, config.EnvFactoryRunID)))
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
	config.EnvFactoryRunID,
	config.EnvFactoryLeadAddr,
	config.EnvFactoryLeadName,
	config.EnvFactoryBackend,
	config.EnvFactoryCard,
	config.EnvFactorySettingsInjected,
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
		"  -l, --lane            join the active factory as the next lane (CLI only)\n" +
		"  Factory sessions show launch_pending until their first prompt binds a session UUID.\n" +
		"  --spawn               open the launch in a new tmux window\n" +
		"  -d, --debug           launcher debug: trace the pre-exec path to stderr\n" +
		"                        (launcher-owned; the token never reaches codex)\n" +
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
	// Same registration rationale as --spawn (SPEC-CODEX-DEBUG-MODE-001):
	// DisableFlagParsing means the launcher itself strips the debug tokens
	// via stripCodexDebugFlag before the verb lookup.
	codexCmd.Flags().Bool("debug", false, "launcher debug trace of the pre-exec path (cli/app verbs only)")
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
	// The debug tokens are launcher-owned on this launcher (SPEC-CODEX-DEBUG-
	// MODE-001 REQ-001): the codex CLI rejects them, so the head's tokens are
	// stripped here — never forwarded — and activate the launcher debug trace
	// downstream. Tokens after -- are codex's own and were never inspected
	// (REQ-002).
	head, debugRequested := stripCodexDebugFlag(head)
	// The factory entries classify before anything else is read or written
	// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-003/-004, narrowing
	// REQ-CFR-001..005, SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-001/002/006): `-l` /
	// `--lane` routes to the per-card relaunch, every other factory shape
	// prints its one refusal line. Tokens after -- are codex's own.
	entry, diag := codexFactoryEntryClassify(head)
	if diag != "" {
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
	if entry == codexFactoryEntryLane {
		// The relaunch stays the parent and selects each card's worktree
		// itself: no --spawn, no -w, and no verb or passthrough composes
		// with it — the closed-set discipline refuses the combination
		// rather than adapting it. The debug token (already stripped above)
		// composes: the relaunch loop traces its own pre-exec steps under
		// debug mode.
		if spawn || worktree.present || hasTail || len(stripCodexFactoryTokens(head)) > 0 {
			return codexUsageFailure(cmd)
		}
		return runCodexFactoryLane(cmd, factoryEntry, debugRequested)
	}
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
		return runCodexLaunch(cmd, kind, tail, spawn, worktree, factoryEntry, debugRequested)
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
	if debugRequested {
		// REQ-003: the debug token on a readout refuses with the named
		// diagnostic — a readout starts nothing there is nothing to trace.
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexDebugReadoutDiag)
		return &exitCodeError{code: 1}
	}
	if hasTail {
		return codexUsageFailure(cmd)
	}
	return runCodexReadout(cmd)
}

// The factory refusal line. It carries the same sentinel as the `moai cg`
// refusal (D2), so one grep finds both. It is the ONE wording source REQ-SD-004
// names: every `-f` shape prints it byte-identically
// (AC-SD-004 compares stderr against this constant). Since
// SPEC-LAUNCHER-ENTRY-FLAGS-001 the lane entry is `-l`, so no `-f` shape is an
// entry on moai codex, `-f lane` included. The retired `-k` entry prints
// retiredEntryRefusal instead (launcher_retired_entries.go).
const codexFactoryRefusalDiag = factoryUnsupportedBackendSentinel +
	": moai codex has no -f entry; the lane entry is 'moai codex -l'; use 'moai cc -f' or 'moai glm -f' for the factory leader"

// codexFactoryLegacyEntryCanonical is the canonical-form clause shared by the
// lane-shape legacy refusals: on moai codex the only factory entry is the
// `-l` relaunch.
const codexFactoryLegacyEntryCanonical = "'moai codex -l' is the only Codex factory entry"

// codexFactoryLegacyRefusalDiag builds the REQ-RNC-003/-005/-007 refusal for
// a legacy role spelling at the codex -f value position (AC-SD-021). The
// message mirrors the REQ-RNC producer shapes (%q is the legacy …; <the
// canonical form>) and names the canonical form for this surface: `moai
// codex -l` for the lane shapes, `moai cc -f` / `moai glm -f` for the
// leader (codex launches no leader). ok is false for any non-legacy value —
// the caller falls through to the REQ-SD-004 line.
func codexFactoryLegacyRefusalDiag(value string) (diag string, ok bool) {
	lowered := strings.ToLower(value)
	if n, isLabel := factory.SplitFactoryLegacyLabel(lowered); isLabel {
		return fmt.Sprintf("%q is the legacy lane label; use %q — %s", value, factory.FactoryLaneLabel(n), codexFactoryLegacyEntryCanonical), true
	}
	if factory.IsLegacyFactoryRoleValue(lowered) {
		return fmt.Sprintf("%q is the legacy role token; %s", value, codexFactoryLegacyEntryCanonical), true
	}
	if factory.IsLegacyLeaderSpelling(lowered) {
		return fmt.Sprintf("%q is the legacy leader spelling; the factory leader launches with 'moai cc -f' or 'moai glm -f'", value), true
	}
	return "", false
}

// codexFactoryEntry classifies the head's factory-entry tokens
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-003/-004).
type codexFactoryEntry int

const (
	codexFactoryEntryAbsent codexFactoryEntry = iota // no factory token in the head
	codexFactoryEntryLane                            // -l / --lane: the per-card relaunch
	codexFactoryEntryOther                           // every other factory shape: refused
)

// codexFactoryEntryClassify scans the head (the tokens before --) for a
// factory entry and returns its classification plus the refusal line for the
// refused shapes ("" when the head carries none or the lane entry). It is a
// scan, not a parser: the value tokens it consumes mirror parseFactoryFlag's
// spellings — a following token that looks like a flag is never a value, and
// the `=` forms are read in place. A refused shape fires before anything else
// is read or written (REQ-SD-004).
//
// The retired `-k` spelling is refused first, in any position and beside any
// other token, with retiredEntryRefusal (SPEC-LAUNCHER-ENTRY-FLAGS-001
// REQ-010) — the same priority the cc and glm entry parse gives it.
//
// The lane entry is `-l` / `--lane` (SPEC-LAUNCHER-ENTRY-FLAGS-001): it takes
// no argument, composes with no other entry token (`-f`) and no operator
// --name, and the leader selector composes with it only. `-f` is no Codex
// entry in any shape — `-f lane` included. Legacy role tokens (`worker` /
// `agent`, their numbered labels, and `lead`) refuse with the
// REQ-RNC-003/-005/-007 message naming the canonical form (AC-SD-021) — the
// REQ-SD-004 line is reserved for non-legacy shapes.
func codexFactoryEntryClassify(head []string) (codexFactoryEntry, string) {
	if refuseRetiredEntry(head) != nil {
		return codexFactoryEntryOther, retiredEntryRefusal
	}
	var (
		lane, leadSeen bool
		other          string // the first non-lane entry token, which a lane token collides with
		refusal        string // the first refusal line a factory token earns, in token order
	)
	for i := 0; i < len(head); i++ {
		token := head[i]
		var value string
		hasValue := false
		switch {
		case token == laneFlagShort || token == laneFlagLong:
			if i+1 < len(head) && !strings.HasPrefix(head[i+1], "-") {
				return codexFactoryEntryOther, laneFlagArgumentError
			}
			lane = true
			continue
		case strings.HasPrefix(token, laneFlagShort+"="), strings.HasPrefix(token, laneFlagLong+"="):
			return codexFactoryEntryOther, laneFlagArgumentError
		case token == leadFlagLong || strings.HasPrefix(token, leadFlagLong+"="):
			leadSeen = true
			continue
		// --factory-run is NOT classified here (card t1444 ②): the token
		// travels to parseCodexFactoryEntry, which owns its validation and
		// its precise refusals (requires -l/--lane, selector conflict). The
		// usage line has always advertised the flag; the classifier was
		// refusing what the help promised.
		case token == factoryFlagLong || token == factoryFlagShort:
			if i+1 < len(head) && !strings.HasPrefix(head[i+1], "-") {
				value, hasValue = head[i+1], true
				i++
			}
		case strings.HasPrefix(token, factoryFlagLong+"="), strings.HasPrefix(token, factoryFlagShort+"="):
			value = strings.TrimPrefix(strings.TrimPrefix(token, factoryFlagShort+"="), factoryFlagLong+"=")
			hasValue = true
		default:
			continue
		}
		if other == "" {
			other = "-f/--factory"
		}
		if refusal != "" {
			continue
		}
		refusal = codexFactoryRefusalDiag
		if hasValue {
			// AC-SD-021: legacy role spellings refuse with the
			// REQ-RNC-003/-005/-007 message naming the canonical form — on
			// moai codex too, not the REQ-SD-004 line. This is the check the
			// M5 classification left one branch away.
			if diag, isLegacy := codexFactoryLegacyRefusalDiag(value); isLegacy {
				refusal = diag
			}
		}
	}
	switch {
	case lane && other != "":
		return codexFactoryEntryOther, fmt.Sprintf(entryTokenConflict, other, "-l/--lane")
	case leadSeen && !lane:
		return codexFactoryEntryOther, leaderNeedsLaneEntry
	case refusal != "":
		return codexFactoryEntryOther, refusal
	case lane && operatorSuppliedName(head):
		return codexFactoryEntryOther, laneFlagNameError
	case lane:
		return codexFactoryEntryLane, ""
	}
	return codexFactoryEntryAbsent, ""
}

// stripCodexFactoryTokens removes the `-f`/`--factory` token and its lane
// value from the verb-position tokens. codexFactoryEntryClassify has already
// validated the shape, so the strip is mechanical.
func stripCodexFactoryTokens(head []string) []string {
	rest := make([]string, 0, len(head))
	for i := 0; i < len(head); i++ {
		token := head[i]
		switch {
		case token == factoryFlagLong || token == factoryFlagShort:
			if i+1 < len(head) && !strings.HasPrefix(head[i+1], "-") {
				i++
			}
		case strings.HasPrefix(token, factoryFlagLong+"="), strings.HasPrefix(token, factoryFlagShort+"="):
		default:
			rest = append(rest, token)
		}
	}
	return rest
}

// enterCodexRelaunchJoin is the relaunch loop's lane-join step (card t1444
// ②): the parsed entry's run id is the explicit selector — named or empty,
// and the loop joins through the SHARED gate enterFactoryLaneRun (REQ-010's
// one implementation): a launcher-private join copy is the AC-013 defect
// shape, and the guided ambiguity refusal and the discovery fallback are
// doors cc/glm lanes already pass.
func enterCodexRelaunchJoin(root string, entry factoryFlagParse, timing *factoryLaunchTiming) (func(), error) {
	return enterFactoryLaneRun(root, entry.RunID, entry.Lead, timing)
}

// @MX:NOTE: the supervising loop stays the parent (design.md §6): lease the
// next card through the F1 machinery on the parent checkout, ensure its
// worktree, start ONE interactive Codex child there, wait, repeat. The stop
// condition is `next`'s no-card answer — which the REQ-SD-025 merge-ready
// skip rule feeds, so a card the Codex harness cannot advance never
// livelocks the loop. The child learns its card through MOAI_KANBAN_CARD
// (REQ-SD-019); no process replacement happens on this path (design.md §6) —
// the launcher stays the parent across every card.
// @MX:SPEC: SPEC-FACTORY-SELF-DISPATCH-001
func runCodexFactoryLane(cmd *cobra.Command, entry factoryFlagParse, debug bool) error {
	// The loop drives the F1 lease machinery itself, so it inherits the
	// `next` verb's own precondition: the parent checkout (REQ-SD-010).
	if err := factoryAssertParentCheckout(resolveProjectDir()); err != nil {
		return err
	}
	// Debug mode (SPEC-CODEX-DEBUG-MODE-001): the relaunch loop traces its
	// own pre-exec steps — binary resolution, run resolution, the lane claim
	// with its claimed label — through the same collector, then dumps once
	// before the first session handoff. Without debug the collector stays
	// nil and nothing changes.
	var launchTiming *factoryLaunchTiming
	if debug {
		launchTiming = &factoryLaunchTiming{debug: true}
	}
	endBinary := launchTiming.beginDebug(launchStepBinaryResolve, "")
	binaryPath, err := codexLookPath(codexBinaryName)
	endBinary()
	if err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexInstallHint)
		return &exitCodeError{code: 1}
	}
	root := factoryCardRoot()
	// The run id and the git requirement arrive together: a lane join
	// resolves the single active run and refuses outside a git working tree
	// before any write (REQ-SD-005) — the same door the cc/glm lane join
	// uses. The parsed entry's run id is the explicit selector (card t1444
	// ②): named or empty, the shared gate resolves it.
	restoreRun, err := enterCodexRelaunchJoin(root, entry, launchTiming)
	if err != nil {
		return err
	}
	defer restoreRun()
	runID := strings.TrimSpace(os.Getenv(config.EnvFactoryRunID))
	// The label is claimed atomically — the next free lane-<n>, bumped past a
	// live hold — so two codex lanes cannot start under one label.
	endClaim := launchTiming.beginDebug(factoryStepLaneClaim, "")
	label, err := resolveFactoryLaneName(root, "", kanban.BackendGPT, true, cmd.ErrOrStderr())
	endClaim()
	if launchTiming != nil && err == nil {
		// REQ-005/§D.1 edge: the lane-claim step carries the claimed label
		// and the run id — session labels, not secrets; environment VALUES
		// stay out of the trace (REQ-006).
		launchTiming.annotateDetail("label=" + label + " run=" + runID)
	}
	if err != nil {
		return err
	}
	if debug {
		launchTiming.debugDump(cmd.ErrOrStderr())
	}
	// The stamps arm the loop's own next calls (lane admission, the
	// merge-ready skip) and identify the lane; the backend value rides the
	// same export the cc/glm launches use (REQ-SD-002's stamp set, gpt). The
	// factory card verbs (next/stage/complete) and the factory notices read the
	// lane label from the lane-worker marker alone (REQ-RNC-011's kept name), and
	// the child environment carries the same marker.
	restoreLane := enterFactoryLaneMode(label, 0, "", config.FactoryDispatchAuto)
	defer restoreLane()
	restoreBackend := exportFactoryLaunchFacts("", factory.BackendGPT)
	defer restoreBackend()

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		card, leased, err := factoryNextLeaseOnce(ctx, root, runID, label)
		if err != nil {
			return fmt.Errorf("codex lane: %w", err)
		}
		if !leased {
			return nil // no card available: the loop's stop condition (REQ-SD-003)
		}
		wt, _, err := factoryEnsureCardWorktree(ctx, root, runID, card, label, cmd.ErrOrStderr())
		if err != nil {
			return fmt.Errorf("codex lane: %w", err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s stage=%s worktree=%s\n", card.CardID, dash(card.Stage), filepath.Base(wt))
		if err := launchCodexCardSession(binaryPath, wt, label, card.CardID, debug); err != nil {
			// On that session's exit, continue with the next card
			// (REQ-SD-003): a child that failed to start or exited non-zero
			// does not stop the loop; its card stays leased until expiry.
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex lane: card %s session: %v\n", card.CardID, err)
		}
	}
}

// launchCodexCardSession starts ONE interactive Codex session whose working
// directory is the card worktree (design.md D1): `codex -C <worktree>`, the
// card worktree's local instruction files attached. The launcher stays the
// parent and waits; the existing child-process launch form serves every
// platform (design.md §6 — Windows needs no new syscall). Under debug mode
// the child environment carries the RUST_LOG linkage (REQ-010/REQ-011).
func launchCodexCardSession(binaryPath, wt, label, cardID string, debug bool) error {
	localArgs, err := codexLocalDeveloperInstructionArgs(wt)
	if err != nil {
		return fmt.Errorf("load Codex local instructions: %w", err)
	}
	args := append([]string{"-C", wt}, localArgs...)
	for _, arg := range args {
		if strings.HasPrefix(arg, "developer_instructions=") {
			if err := checkCodexInstructionSize(len(arg), "lane instruction token"); err != nil {
				return err
			}
		}
	}
	req := codexLaunchRequest{Program: binaryPath, Args: args, Dir: wt, Debug: debug}
	c := exec.Command(req.Program, req.Args...)
	c.Dir = req.Dir
	c.Env = codexCardLaunchEnv(label, cardID)
	if debug {
		c.Env = codexApplyDebugEnv(c.Env)
	}
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return codexDirectLaunchFn(c)
}

// codexCardLaunchEnv is the per-card child environment (design.md §4): the
// eleven-key scrub of codexChildEnv, then the entries the owned-card session
// reads — the role marker, the lane label (the lane-worker marker, the one carrier
// the factory card verbs and the factory notices read), the Codex backend
// value, and the leased card's id in the card-identifier variable
// (REQ-SD-003, -019). Appending after the scrub
// is what keeps them authoritative: the inherited environment lost every
// lane key before the lane values land. The factory fan-out signal
// (MOAI_FACTORY_WORKERS) is deliberately not carried — it feeds the Stop-hook
// block cap, and a Codex lane has no moai hook peer (REQ-CFR-022).
// @MX:SPEC: SPEC-FACTORY-SELF-DISPATCH-001
func codexCardLaunchEnv(label, cardID string) []string {
	env := codexChildEnv()
	return append(env,
		config.EnvFactoryRole+"="+config.FactoryRoleLane,
		config.EnvMoaiFactoryWorker+"="+label,
		config.EnvFactoryBackend+"="+factory.BackendGPT,
		config.EnvFactoryCard+"="+cardID,
	)
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
//
// Under debug mode (SPEC-CODEX-DEBUG-MODE-001) the collector is instantiated
// on EVERY traced launch — lane or not (REQ-014's composition posture) — and
// the debug-vocabulary steps are recorded through beginDebug, which measures
// nothing without debug, so the debug-off lane wiring below records exactly
// its t1378 step set (REQ-015's freeze). The dump prints before the platform
// exec seam (REQ-009).
func runCodexLaunch(cmd *cobra.Command, kind codexVerb, tail []string, spawn bool, worktree codexWorktreeArg, factoryEntry factoryFlagParse, debug bool) error {
	// The slow-launch timing covers a codex LANE launch (REQ-012,
	// SPEC-CODEX-LANE-SLOTS-001) and — under debug mode — every launch; the
	// phase runs from the first recorded step through the exec handoff, and
	// the reports print before the platform exec seam — the direct door
	// replaces this process and prints nothing afterwards.
	laneTiming := factoryEntry.Enabled && (factoryEntry.LaneRole || factoryEntry.LaneNumber > 0)
	var launchTiming *factoryLaunchTiming
	if debug || laneTiming {
		launchTiming = &factoryLaunchTiming{debug: debug}
		defer launchTiming.reportSlow(cmd.ErrOrStderr()) // error-path guard; no-op after the pre-seam report
	}
	endBinary := launchTiming.beginDebug(launchStepBinaryResolve, "")
	binaryPath, err := codexLookPath(codexBinaryName)
	endBinary()
	if err != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), codexInstallHint)
		return &exitCodeError{code: 1}
	}

	endRoot := launchTiming.beginDebug(launchStepProjectRoot, "")
	// The launch cwd is the PROJECT ROOT, not the process cwd (AC-CL-002):
	// a call from a subdirectory still launches at the root. An unresolvable
	// root degrades to the process cwd rather than refusing to launch.
	projectRoot := ""
	if root, rerr := findProjectRootFn(); rerr == nil && root != "" {
		projectRoot = root
	} else if cwd, gerr := os.Getwd(); gerr == nil {
		projectRoot = cwd
	}
	endRoot()

	// SPEC-CODEX-INIT-001: the init-offer gate — the ONE call site every
	// launch form passes through right before launching, the bare form
	// included. The gate takes no spawn argument: both launch paths cross the
	// same function (REQ-CI-002). Under debug the phase records under its
	// debug-vocabulary name (REQ-005); without debug the t1378 name stands
	// and the local-instruction load records no separate step (the freeze).
	initStepName := factoryStepCodexPreInit
	if debug {
		initStepName = launchStepInitGate
	}
	endInit := launchTiming.begin(initStepName)
	if err := codexInitOfferGate(cmd, projectRoot); err != nil {
		endInit()
		return err
	}
	endInit()
	endLocal := launchTiming.beginDebug(launchStepLocalInstr, "")
	localArgs, localRead, err := codexLocalDeveloperInstructions(projectRoot)
	endLocal()
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
		endWt := launchTiming.beginDebug(launchStepWorktree, "")
		resolved, werr := resolveCodexWorktreeDir(projectRoot, worktree.value)
		if werr == nil {
			werr = codexWorktreeWriterCheck(resolved)
		}
		endWt()
		if werr != nil {
			launchTiming.annotateDetail("resolved " + worktree.value + " failed: " + werr.Error())
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), werr.Error())
			return &exitCodeError{code: 1}
		}
		// REQ-007: the resolved directory and the writer-check outcome ride
		// the worktree step; the anchor-lock outcome rides the handoff step,
		// where the lock is actually placed.
		launchTiming.annotateDetail("resolved " + resolved + "; writer-check ok")
		dir = resolved
	}
	childArgs := append(localArgs, codexChildArgs(kind, tail)...)
	req := codexLaunchRequest{Program: binaryPath, Args: childArgs, Dir: dir, Debug: debug, timing: launchTiming}
	// SPEC-FACTORY-MANAGED-SESSION-001 M3 (design.md D-7): the divert engages
	// only when the explicit opt-in MOAI_FACTORY_MANAGED (1/true) AND the
	// factory stamps are both in the process env. Such a launch enters the
	// managed Codex owner — the launcher keeps its PID and owns the App Server
	// child (REQ-MS-012) — instead of the doors below. Stamps alone or the
	// switch alone stay on the ordinary doors, and the tmux --spawn door never
	// diverts. `moai codex -l` card children never reach this code:
	// runCodexFactoryLane launches them through the direct door.
	if !spawn && factoryManagedRequested(os.Environ()) && factoryLaunchEnabled(os.Environ()) {
		if worktree.present {
			// Same anchor discipline as the direct door: the lock names the
			// owning process before the session starts.
			if err := codexWorktreeAnchorLock(dir, codexDirectAnchorPID(), homestate.CurrentProcessFingerprint()); err != nil {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
				return &exitCodeError{code: 1}
			}
		}
		return managedFactoryCodexLaunchFunc(binaryPath, append([]string{binaryPath}, childArgs...), os.Environ(), req.Dir)
	}
	if spawn {
		if err := checkSpawnPrereqs(); err != nil {
			return err
		}
	}
	if factoryEntry.Enabled {
		// The run is project-scoped even when Codex enters an existing worktree.
		restore, ferr := enterCodexFactory(projectRoot, factoryEntry, launchTiming)
		if ferr != nil {
			return ferr
		}
		defer restore()
		req.FactoryEnv = codexFactoryEnv(factoryEntry)
	}
	// The child-env assembly is a traced pre-exec step (REQ-005), performed
	// HERE — ahead of the pre-seam reports — so the assembly line precedes
	// the seam with the rest of the trace (REQ-009). The injection follows
	// REQ-010/REQ-011: appended last-wins only when the operator set no
	// RUST_LOG.
	envDetail := ""
	if debug {
		envDetail = codexDebugEnvDetail(true)
	}
	endEnv := launchTiming.beginDebug(launchStepChildEnv, envDetail)
	childEnv := codexChildEnv()
	if debug {
		childEnv = codexApplyDebugEnv(childEnv)
	}
	req.Env = append(childEnv, req.FactoryEnv...)
	endEnv()
	// The exec handoff is the launcher's last measurable pre-exec work — the
	// final assembly and the anchor lock — and the reports are the launcher's
	// last output before the platform exec seam (the direct door replaces the
	// process; the spawn door's child is already launched).
	endHandoff := launchTiming.begin(factoryStepExecHandoff)
	if spawn {
		if worktree.present {
			// The process that becomes Codex is the pane's; it exists only
			// after the spawn, so the lock is placed from inside it.
			codexSpawnAnchorFn = func(pid int, start string) error { return codexWorktreeAnchorLock(dir, pid, start) }
			defer func() { codexSpawnAnchorFn = nil }()
			endHandoff()
			launchTiming.annotateDetail("anchor-lock deferred to the spawned pane")
		} else {
			endHandoff()
		}
		if debug {
			// REQ-010 on the spawn door: the pane's codex reads the
			// command-scoped assignments, so the same absence guard carries
			// the injection there; an operator-supplied RUST_LOG stays
			// authoritative (REQ-011).
			if _, ok := os.LookupEnv(config.EnvRustLog); !ok {
				req.FactoryEnv = append(req.FactoryEnv, config.EnvRustLog+"="+codexDebugRustLogValue)
			}
			launchTiming.debugDump(cmd.ErrOrStderr())
		}
		launchTiming.reportSlow(cmd.ErrOrStderr())
		return codexSpawnLaunch(req)
	}
	if worktree.present {
		// The lock names the process that becomes Codex, before it starts.
		if err := codexWorktreeAnchorLock(dir, codexDirectAnchorPID(), homestate.CurrentProcessFingerprint()); err != nil {
			endHandoff()
			launchTiming.annotateDetail("anchor-lock failed: " + err.Error())
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			return &exitCodeError{code: 1}
		}
	}
	endHandoff()
	if debug {
		if worktree.present {
			launchTiming.annotateDetail("anchor-lock ok")
		}
		launchTiming.debugDump(cmd.ErrOrStderr())
	}
	launchTiming.reportSlow(cmd.ErrOrStderr())
	return codexDirectLaunch(req)
}

// codexDirectLaunch assembles the child (stdio = the parent's OWN os.Stdin /
// os.Stdout / os.Stderr values — the interactive-tty precondition,
// AC-CL-002), runs it through the seam, and propagates the child's exit code
// verbatim (AC-CL-002 rc axis, AC-CL-016). The child environment arrives
// ASSEMBLED on the request (the launcher traced its assembly pre-seam,
// SPEC-CODEX-DEBUG-MODE-001 REQ-005/REQ-009); this function is the seam
// call only.
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
	c.Env = req.Env
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
