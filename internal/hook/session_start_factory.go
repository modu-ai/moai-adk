// session_start_factory.go emits the Factory Mode bootstrap announcement into
// the session (SPEC-FACTORY-WORKER-FANOUT-001).
//
// The announcement is emitted HERE rather than by the launcher because the
// launcher syscall.Exec's into claude (internal/cli/launch_exec_posix.go), so
// anything it writes to stdout is overwritten the moment the TUI takes the
// screen.
//
// The t85 leader loop is also injected here for the same reason: the launcher
// cannot run a loop (it exec's in place and is gone), so the loop is a
// SESSION-side behavior codified into the leader notice — the discipline text
// (whole-card routing, staggered activation) the leader session executes
// against.
//
// The same two-audience split applies. The orchestrator reads
// hookSpecificOutput.additionalContext, the operator reads systemMessage, and
// the operator needs the lane-start sentence because a session cannot launch
// another session — those terminals are opened by hand. Both copies render in
// the session's conversation_language (SPEC-SESSION-START-GUIDE-I18N-001:
// card t1603 amends the bootstrap guide's two-audience English rule for this
// surface only — decision-index Q1; agent_prompt_language still governs the
// other agent-facing surfaces); the commands, run id, and socket path are
// protocol tokens and are emitted verbatim in every locale so the operator's
// paste keeps working. The notice states no lane count, no per-lane line, and
// no free-slot list (SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-009).
package hook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryBootstrapNotice returns the factory announcement for this session in
// lang, or "" when the session is not part of a factory run. root is the
// project root the stale-run check reads run state under.
// sessionID keys the stale-run check: a session whose launch label or record
// carries a legacy role value gets the stale-run notice instead of a leader
// or lane notice (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-022, REQ-RNC-025).
//
// Fail-open throughout, matching the surrounding hook code: an unparseable
// lane count degrades to omitting the count-dependent copy rather than failing
// the session start, and an unknown lang degrades to English, never to an
// empty notice.
func factoryBootstrapNotice(root, sessionID, lang string) string {
	return factoryBootstrapNoticeLang(root, sessionID, lang, lang)
}

// factoryBootstrapNoticeLang is factoryBootstrapNotice with the stale-run
// recovery notice's locale pinned separately from the guidance locale. The
// operator-facing copy passes lang twice (stale-run recovery localizes like
// the rest of the operator notice); the agent-facing channel pins the
// stale-run recovery notice to English — that surface names relaunch
// commands and recovery state, and sits outside the bootstrap-guide surface
// the conversation-language rule localizes
// (SPEC-SESSION-START-GUIDE-I18N-001 § Out of Scope).
func factoryBootstrapNoticeLang(root, sessionID, lang, staleRunLang string) string {
	if label := os.Getenv(config.EnvMoaiFactoryWorker); label != "" {
		if factory.IsLegacyFactoryRoleValue(label) {
			// Run-state gated: an active run prescribes once, a dead run
			// unbinds once, an unmeasurable one degrades — never an
			// unconditional prescription (SPEC-STALE-RUN-LABEL-001).
			return staleRunPrescriptionGate(context.Background(), root, sessionID, label, os.Getenv(config.EnvFactoryRunID), staleRunLang)
		}
		return factoryLaneNotice(label, factoryLanesEnv(), lang)
	}
	if os.Getenv(config.EnvMoaiFactoryWorkers) == "" {
		return ""
	}
	if notice := staleRunNoticeFor(root, sessionID, staleRunLang); notice != "" {
		return notice
	}
	return factoryLeaderNotice(os.Getenv(config.EnvFactoryRunID), factoryLanesEnv(), lang)
}

// factoryBootstrapNoticeForSource returns the announcement only for a
// genuinely new session, on a startup-only allowlist: the factory environment
// survives resume / clear / compact / fork, and re-announcing the bootstrap
// would tell the operator to open lane terminals that are already open. The
// allowlist rather than a denylist of the re-entry sources keeps a newly added
// source silent by default. An empty source is treated as startup (a caller
// that predates the field, or a test building the input by hand). The stale-run
// notice shares the same gate — a relaunch reminder repeated on every re-entry
// is the same noise problem in reverse.
func factoryBootstrapNoticeForSource(source, root, sessionID, lang string) string {
	if source != "" && source != "startup" {
		return ""
	}
	// The agent-facing channel (additionalContext) localizes the bootstrap
	// GUIDE per conversation_language (card t1603, decision-index Q1) and
	// keeps the stale-run recovery notice English — that recovery surface is
	// outside the guide surface the card localizes. The operator-facing copy
	// (factoryBootstrapNotice, called separately by the handler) localizes
	// both as before.
	return factoryBootstrapNoticeLang(root, sessionID, lang, langEnglish)
}

// factoryLaneRuleForSource returns the lane SessionStart rule
// (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-019) for this session, or "" when
// the session receives none. The rule fires on source startup — under every
// clear policy, which is why the policy value is never read here — and on
// clear, where it re-enters a freshly reset session so the lane still holds
// the in-session card transition; resume and compact receive nothing, matching the bootstrap gate's
// re-entry discipline. The rule is chosen by the backend variable: a
// Claude-harness lane (claude or glm) gets the next-card rule, a
// Codex-harness lane (gpt) the owned-card rule naming its leased card. A
// leader (no lane label), a legacy label, a keyless environment, a gpt lane
// without a card id, and an unknown backend value all receive no rule.
// lang is the session's conversation language (REQ-SD-019).
func factoryLaneRuleForSource(source, lang string) string {
	return factoryLaneRuleForSourceInDir(source, lang, "")
}

// factoryLaneRuleForSourceInDir is factoryLaneRuleForSource with the session
// directory the card id may be derived from (card t1776 audit round-12: a
// gpt lane standing in its card worktree without the card env (config.EnvFactoryCard) stamped
// still carries a card — resolving it from the path keeps such a session on
// the owned-card rule instead of misdirecting it at the parent-checkout
// queue verb). dir may be empty; an empty dir resolves env-only.
func factoryLaneRuleForSourceInDir(source, lang, dir string) string {
	if source != "" && source != "startup" && source != "clear" {
		return ""
	}
	label := os.Getenv(config.EnvMoaiFactoryWorker)
	if label == "" || factory.IsLegacyFactoryRoleValue(label) {
		return ""
	}
	if _, ok := factory.SplitFactoryLaneLabel(label); !ok {
		return ""
	}
	switch os.Getenv(config.EnvFactoryBackend) {
	case factory.BackendClaude, factory.BackendGLM:
		// REQ-TCD-011 (SPEC-TODO-CLASSIFY-DISPATCH-001): the launcher's
		// stamped dispatch selection picks the rule — manual mode receives a
		// manual-mode rule, never the self-dispatch instruction. Any other
		// value (the code default, absence included) reads as auto-dispatch:
		// the default is fail-open.
		//
		// Card t1451: both Claude-harness rules end with the stall-recovery
		// rule. Only the clear-surviving lane rule can carry it — a turn ended
		// by an API error cannot arm a recheck afterwards — and the Codex
		// owned-card rule below does not, because that harness has no session
		// cron tool.
		m := factoryMessagesFor(lang)
		recheck := fmt.Sprintf(m.laneRecheckRule, laneArmExpression())
		if os.Getenv(config.EnvFactoryAutoDispatch) == config.FactoryDispatchManual {
			return m.laneManualDispatchRule + "\n\n" + recheck
		}
		return m.laneNextCardRule + "\n\n" + recheck
	case factory.BackendGPT:
		// Env override first (the launcher's stamp), else the REGISTERED
		// card worktree this session stands in (card t1776 audit rounds
		// 12+14: the path fallback must not mint a card from a coincidental
		// `worktrees/` parent — only the MoAI tiers count).
		cardID := strings.TrimSpace(os.Getenv(config.EnvFactoryCard))
		if cardID == "" {
			cardID = cardIDFromRegisteredWorktree(dir)
		}
		if cardID == "" {
			// The card-less queue boot (moai codex -l before its first
			// lease): the role core's read-first duty conflicts with the
			// boot prompt exactly as it does for a leased card — the
			// BOOT-form precedence note names the queue verb this boot
			// owns (audit round-8: the empty return left the conflict
			// unresolved on the boot path).
			return codexLaneBootPrecedenceNote
		}
		// t1776 audit round-3: the role core's read-the-full-rule duty
		// conflicts with the lane's boot directive (run `moai todo --auto`
		// now; the operator-measured session burned 72% of its context on
		// exactly that reading). The owned-card rule carries the precedence
		// note so the two texts resolve instead of contradicting.
		return fmt.Sprintf(factoryMessagesFor(lang).laneOwnedCardRule, cardID, config.EnvFactoryCard) +
			"\n\n" + codexLanePrecedenceNote
	default:
		return ""
	}
}

// laneArmExpression renders the arm expression of the standing recheck, `*/N * * * *`,
// for the effective watchdog renewal period, workflow.factory_lease.renew_interval
// (SPEC-FACTORY-LEASE-RENEW-001 REQ-FLR-012). The configuration is read from the
// project root the factory opens: the launch base, CLAUDE_PROJECT_DIR or else the
// working directory, resolved through factory.ResolveTodoQueueRoot. That is the
// resolution the factory's own root uses (internal/cli factoryCardRoot), so the
// rendered period is the one the factory enforces. A configuration the factory
// refuses renders the default period, because a session-start hook must not block;
// the factory's own open names the refused key.
func laneArmExpression() string {
	renew := config.DefaultFactoryLeaseRenewInterval
	base := os.Getenv(config.EnvClaudeProjectDir)
	if base == "" {
		base, _ = os.Getwd()
	}
	if base != "" {
		if _, configured, err := homestate.EffectiveFactoryLease(factory.ResolveTodoQueueRoot(base)); err == nil {
			renew = configured
		}
	}
	return fmt.Sprintf("*/%d * * * *", int(renew/time.Minute))
}

// factoryLanesEnv reads the run's fan-out size from the environment the
// launcher published. A missing or malformed value reads as 0, which the
// notice builders treat as "count unknown" rather than as an error.
func factoryLanesEnv() int {
	n, err := strconv.Atoi(os.Getenv(config.EnvMoaiFactoryWorkers))
	if err != nil || n < 1 {
		return 0
	}
	return n
}

// cardIDFromRegisteredWorktree resolves the card id from dir ONLY when dir
// stands inside a MoAI-MANAGED card worktree — a directory whose ancestry
// carries a `worktrees` segment whose parent is `.moai` or `.claude` (the
// two MoAI worktree tiers). A checkout that merely happens to live under a
// plain `worktrees/` parent (e.g. ~/worktrees/myrepo) resolves no card:
// minting an owned-card rule from a coincidental directory name would hand
// a card-less boot both the queue-start directive and its contradiction
// (card t1776 audit round-14). Empty when dir is not a registered-tier
// card worktree.
func cardIDFromRegisteredWorktree(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return ""
	}
	d := filepath.Clean(dir)
	for {
		parent := filepath.Dir(d)
		if parent == d {
			return ""
		}
		// SPEC-WORKTREE-HOME-ROOT-001 M3 (REQ-WHR-008): the home-root tier
		// <MoaiHome>/worktrees/projects/<project-key>/<name> resolves the
		// card id to the tree name — the walk's parent is the project key,
		// never the card.
		if filepath.Base(filepath.Dir(parent)) == "projects" && filepath.Base(filepath.Dir(filepath.Dir(parent))) == "worktrees" {
			return filepath.Base(d)
		}
		if filepath.Base(parent) == "worktrees" {
			gp := filepath.Base(filepath.Dir(parent))
			if gp == ".moai" || gp == ".claude" {
				return filepath.Base(d)
			}
			// A worktrees parent outside the MoAI tiers: not a card tree AT
			// THIS LEVEL — the walk CONTINUES (card t1776 audit round-16: a
			// session deep inside a card worktree whose own subtree carries
			// an inner `worktrees` directory still resolves the card above;
			// an early return minted the boot directive for a card session).
		}
		d = parent
	}
}

// codexLanePrecedenceNote resolves the role core's read-first duty against
// the lane boot directive for a Codex auto lane (card t1776 audit round-3).
// The precedence lives HERE, not in the role core: the dispatch core sits
// within 7 UTF-16 units of TestRoleInjectionAssemblyBudget's cap (audit
// round-6 P1 — any core addition breaches it), so the hook context carries
// the resolution. No queue-start verb rides it either: this branch is an
// already-card-assigned session working its card worktree, and `moai todo
// --auto` belongs to the card-less queue boot only (audit round-4).
const codexLanePrecedenceNote = "Codex auto-lane precedence: the boot directive governs your first action — " +
	"proceed with the leased card's work; " +
	"open the rule files when that work needs them, not before."

// codexLaneBootPrecedenceNote is the card-less queue boot's precedence
// note (card t1776 audit round-8): `moai codex -l` boots WITHOUT a card id,
// and this boot owns the queue verb — the note names it so the role core's
// read-first duty resolves against the boot directive instead of
// contradicting it.
const codexLaneBootPrecedenceNote = "Codex auto-lane precedence: the boot directive governs your first action — " +
	"run `moai todo --auto-lane` before any rule-file reading; " +
	"open the rule files when the cards' work needs them, not before."

// codexLaneEnvFailFast is the t1776 one-line stop for a broken lane launch:
// a session stamped with the Codex factory backend but missing its lane
// stamps — the launcher writes the worker-label stamp and the run-id stamp
// together with the backend stamp (codex_launcher.go), so a
// missing one means the launch did not go through the lane door — gets ONE
// agent-facing stop line instead of silence. The operator-measured failure:
// an unstamped session inferred lane behavior by reading todo_auto*.go and
// factory-dispatch*.md in full, burning the dispatch context before any
// autonomous progress; the line forbids exactly that. Fully stamped lanes and
// non-codex backends receive nothing.
func codexLaneEnvFailFast() string {
	if os.Getenv(config.EnvFactoryBackend) != factory.BackendGPT {
		return ""
	}
	if os.Getenv(config.EnvMoaiFactoryWorker) != "" && os.Getenv(config.EnvFactoryRunID) != "" {
		return ""
	}
	missing := config.EnvMoaiFactoryWorker
	if os.Getenv(config.EnvMoaiFactoryWorker) != "" {
		missing = config.EnvFactoryRunID
	}
	// REQ-SRL-009: the env names ride the config constants, never literals.
	return fmt.Sprintf(
		"MOAI factory lane launch is incomplete: %s=gpt but the lane stamps are not both set (%s, %s; empty: %s) — stop. "+
			"Do not read the engine or the rule sources and do not infer lane behavior; "+
			"report this line and relaunch through `moai codex -l`.",
		config.EnvFactoryBackend, config.EnvMoaiFactoryWorker, config.EnvFactoryRunID, missing)
}

// factoryLeaderNotice is the factory leader branch. It carries, in order:
// (a) the run id and the session name that must accompany it; (b) the one
// lane-start sentence — to start a lane, enter `moai cc -l` (or `moai glm -l`,
// `moai codex -l`) in a new terminal; (c) the entry-point guide — the -f
// leader entries — the per-lane fan-out line, and the leader socket path;
// (d) the dispatch discipline — whole-card routing (each card to ONE lane,
// which runs the serial plan -> run -> sync path in-session) and the
// fan-out-only stagger rule; (e) the operational-status query plus the
// inbound-automation notice. There is no SPEC line, and — by
// SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-009 — no lane count, no per-lane launch
// line, and no free-slot list: the lane guidance is the same bytes whatever
// the run's declared lane count. lanes only gates the notice (a leader whose
// count cannot be read emits nothing, as before).
//
// QUEUE POLLING IS DELIBERATELY NOT TAUGHT HERE. The watch-dispatch-collect
// loop over the backlog queue is the foreman's (the bare /loop driver, t96):
// re-teaching it here would hand the factory leader a second, conflicting
// polling protocol. This notice carries only what is factory-specific — how a
// PICKED card is routed to lanes, how lanes are activated, and what a dispatch
// must never carry.
func factoryLeaderNotice(runID string, lanes int, lang string) string {
	if runID == "" || lanes < 1 {
		return ""
	}
	m := factoryMessagesFor(lang)

	// (a) the run id and the session name that must accompany it — printed
	// together so a disagreement is visible at the moment the notice is
	// emitted, not later when a dispatched card reaches the wrong run.
	identity := []string{
		fmt.Sprintf(m.leaderHeader, runID),
		fmt.Sprintf(m.leaderIdentity, factory.LeaderLabel()),
	}
	// (b) the lane-start sentence rides in the same block as the dispatch
	// statement above it.
	blocks := []string{
		strings.Join(identity, "\n"),
		m.leaderManual,
	}

	// (c) the entry-point guide and the per-lane fan-out, plus the leader
	// socket path when the launcher captured one.
	backend := []string{m.entryGuide, m.agentFanout}
	if addr := os.Getenv(config.EnvFactoryLeadAddr); addr != "" {
		backend = append(backend, fmt.Sprintf(m.leaderSocket, addr))
	}
	blocks = append(blocks, strings.Join(backend, "\n"))

	// (d) the dispatch discipline — localized prose with verbatim protocol
	// tokens; see factoryMessages for why the tokens are not translated.
	blocks = append(blocks, strings.Join([]string{m.leaderClasses, m.leaderStagger, m.gateSummary}, "\n"))

	// (e) the operational-status query and the inbound-automation notice, on
	// the injected-settings discriminator the launcher publishes.
	var context []string
	context = append(context, fmt.Sprintf(m.operationalStatus, runID))
	if os.Getenv(config.EnvFactorySettingsInjected) == "1" {
		context = append(context, m.settingsAuto)
	} else {
		context = append(context, m.settingsVerify)
	}
	blocks = append(blocks, strings.Join(context, "\n"))

	// (f) the two-mode auto guidance (leader variant, carrying the
	// design-surface marker REQ-005 pins) and the online docs pointer
	// (card t1603).
	blocks = append(blocks, strings.Join([]string{m.autoModeGuideLeader, m.docsPointer}, "\n"))

	return strings.Join(blocks, "\n\n") + "\n"
}

// factoryLaneNotice is the factory lane branch: a single line
// acknowledging the join, naming the label this session launched under (which
// may be a bumped number — the registry note on stderr is gone by the time
// the TUI takes the screen, so this line is where the operator reads the
// final name). It does NOT print a launch block: a lane is already running,
// and the operator has nothing left to paste.
//
// The incremental `-l` lane entry carries no run count (the launcher
// publishes 0), and the count-less sentence names the label alone rather
// than fabricating a fan-out size. Both sentences take (label, lanes);
// the per-locale word order differs (en/ja/ko say the count first, zh the
// label first), so the formats pin the argument order with explicit %[n]
// indices rather than positional verbs.
func factoryLaneNotice(label string, lanes int, lang string) string {
	if _, ok := factory.SplitFactoryLaneLabel(label); !ok {
		return ""
	}
	m := factoryMessagesFor(lang)
	var join string
	if lanes < 1 {
		join = fmt.Sprintf(m.laneJoinNoCount, label)
	} else {
		join = fmt.Sprintf(m.laneJoin, label, lanes)
	}
	// Card t224: the standing spawn authority rides the join notice — it is
	// the one message the lane is guaranteed to read at startup, and the
	// tk8hce incident showed a lane without it refusing to spawn the
	// phase-required specialist. Card t1603 moved it into the message table
	// (the laneSpawnAuthority field), so it renders in the session's
	// conversation language; en carries the canonical sentence. The lane
	// variant of the two-mode auto guidance and the docs pointer ride the
	// same notice (SPEC-SESSION-START-GUIDE-I18N-001).
	return strings.Join([]string{
		join,
		m.laneSpawnAuthority,
		strings.Join([]string{m.autoModeGuideLane, m.docsPointer}, "\n"),
	}, "\n\n")
}
