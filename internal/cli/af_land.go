package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq"
	"github.com/modu-ai/moai-adk/internal/gitenv"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// afLandTimeout bounds `af land` and `af reconcile`: a landing runs the gates, which take far longer than a
// ledger write.
const afLandTimeout = 30 * time.Minute

// afGateKeep is how much of the end of a gate's output is kept in memory and, from there, in the failure log: the end
// is what explains a failure, and the volume must not fill memory.
const afGateKeep = mergeq.DefaultFailureLogCap

// The merge queue's operational entry points (AC-AFL-014): `af land` lands an enqueued card through the queue's
// compare-and-swap of the trunk ref, and `af reconcile` settles the rows a restart left in verifying or landing
// by git ancestry. Both are thin: mergeq.Merger owns every rule.

// afMergeFlags are the flags both verbs take.
type afMergeFlags struct {
	db, repo, trunk, session string
}

func (f *afMergeFlags) bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.db, "db", "", "ledger file (default: <project dir>/af/ledger.db)")
	cmd.Flags().StringVar(&f.repo, "repo", "", "the git repository the trunk lives in: the project directory (default), or any path with the same git directory, "+
		"a linked worktree of it included; any other path is refused")
	cmd.Flags().StringVar(&f.trunk, "trunk", mergeq.DefaultTrunk, "the trunk ref; only the default "+mergeq.DefaultTrunk+" is accepted and it cannot be changed (check out nothing on it)")
	cmd.Flags().StringVar(&f.session, "session", "moai-af", "who holds the landing lock (recorded as given)")
}

// afMerger builds the merge queue over the opened ledger. verify may be nil for a verb that runs no gates.
func afMerger(f afMergeFlags, st *ledger.Store, scratch string, verify mergeq.Verify) *mergeq.Merger {
	repo := f.repo
	if repo == "" {
		repo = resolveProjectDir()
	}
	pid := os.Getpid() // the lock is held while this command runs, so this process is the holder
	start, _ := homestate.ProbeProcessIdentity(pid)
	// the failure logs sit beside the ledger, in a directory of their own
	var failLog *mergeq.FailureLog
	if path, err := afLedgerPath(f.db); err == nil {
		failLog = &mergeq.FailureLog{Dir: filepath.Join(filepath.Dir(path), "gate-logs"), Cap: mergeq.DefaultFailureLogCap}
	}
	return &mergeq.Merger{
		Log:    failLog,
		Git:    mergeq.GitCLI{Repo: repo, ScratchD: scratch, Who: mergeq.Identity{Name: "moai af", Email: "af@moai.invalid"}},
		Queue:  mergeq.NewLedgerQueue(st),
		Trunk:  f.trunk,
		Verify: verify, Holder: mergeq.Holder{Session: f.session, PID: pid, PIDStart: start},
		LockTTL: mergeq.DefaultLockTTL, MaxRetries: mergeq.DefaultMaxRetries,
		Message: func(card string) string { return "merge(" + card + "): land" },
	}
}

// afGateWaitDelay is how long to wait for the gate's output pipes after its context ended: a descendant that keeps
// them open must not hold the command past the timeout.
const afGateWaitDelay = 5 * time.Second

// afGateTimeout is the default time a gate may run.
const afGateTimeout = 20 * time.Minute

// afGate runs the gate command in the candidate worktree, under a timeout, in its own process group (killed whole
// when the timeout ends), with the repository-scoping variables removed from its environment and its output bounded
// by the merge queue's writer. A non-zero exit fails the card, and what the command printed is NOT part of the
// error: it goes to a failure log file (mergeq.FailureLog), and the ledger and the verb's output carry only the
// failure kind, the gate's short id, the exit status, the byte count and that file's path.
func afGate(gate string, timeout time.Duration) mergeq.Verify {
	return func(ctx context.Context, dir string) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "sh", "-c", gate)
		cmd.Dir = dir
		cmd.Env = gitenv.Env()
		cmd.WaitDelay = afGateWaitDelay
		afGateProcessGroup(cmd)
		// the end of the output is what explains a failure, so keep the end and never fail on volume
		out := mergeq.NewTailWriter(afGateKeep)
		cmd.Stdout, cmd.Stderr = out, out
		err := cmd.Run()
		// Whatever the gate returned, nothing it started may outlive it: a server it left running holds the output
		// pipes, and the group is the only thing that reaches it.
		afGateKillGroup(cmd)
		if errors.Is(err, exec.ErrWaitDelay) && cmd.ProcessState != nil && cmd.ProcessState.Success() {
			err = nil // the gate exited 0; only a child kept the pipes open
		}
		if err != nil {
			if ctx.Err() != nil {
				err = fmt.Errorf("%w (the gate ran past %s)", err, timeout)
			}
			// What the gate printed does not go into the error: the merge queue puts it in the failure log.
			return &mergeq.GateError{Gate: afGateID(gate), Cause: err, Output: append([]byte(nil), out.Bytes()...), Total: out.Total()}
		}
		return nil
	}
}

// afGateID names a gate in the reason, the verb's output, the notes and the landed note: "gate-" and the first 12 hex
// digits of the SHA-256 of the gate text. The text itself may hold a credential (a deny-list of shapes cannot be made
// complete), so it is recorded nowhere but the failure log; the same text always has the same id, and two gates
// differ in theirs.
func afGateID(gate string) string {
	sum := sha256.Sum256([]byte(gate))
	return "gate-" + hex.EncodeToString(sum[:6])
}

// afCardMayLand says whether the card is where a landing needs it: MERGING at the epoch the queue row was made
// at. A holder who moved the card back (to REVIEWING, say) withdrew the work, and the row it left behind is not
// permission to land it. A row already landed answers landed on a retry and needs no position.
func afCardMayLand(ctx context.Context, st *ledger.Store, item mergeq.Item) error {
	if item.State == mergeq.StateLanded {
		return nil
	}
	cards, err := st.Cards(ctx)
	if err != nil {
		return err
	}
	for _, c := range cards {
		if c.ID != item.CardID {
			continue
		}
		if c.Position.Park != nil {
			return fmt.Errorf("af land: %s is parked in %s (%s): a parked card does not land until its park is resolved",
				item.CardID, c.Position.Stage, c.Position.Park.Reason)
		}
		if c.Position.Stage != core.StageMerging || c.Position.RemoteIntegrated || c.Epoch != item.CardEpoch {
			return fmt.Errorf("af land: %s is not in MERGING at the epoch its queue row was made at (stage %s, epoch %d; row epoch %d): "+
				"the holder moved it out of the merge queue, so it is not landed; if its row is left in landing, "+
				"`moai af reconcile` settles it", item.CardID, c.Position.Stage, c.Epoch, item.CardEpoch)
		}
		// The rule Enqueue and finishing a card apply: a commit whose newest green, gate or review run failed is not
		// landed. A run that failed while the row waited would otherwise put the commit in the trunk, where nothing
		// takes it out again (N54). A passing run about another commit does not count (D5).
		kind, err := st.FailingNewestRun(ctx, item.CardID, item.CandidateSHA)
		if err != nil {
			return fmt.Errorf("af land: %w", err)
		}
		if kind != "" {
			return fmt.Errorf("af land: %s is not landed: the newest %s run of commit %s failed; measure the commit again "+
				"(`moai af evidence %s --kind %s --subject-sha %s --result pass`, and the other kinds the card needs) and land again; "+
				"nothing was changed", item.CardID, kind, item.CandidateSHA, item.CardID, kind, item.CandidateSHA)
		}
		return nil
	}
	return fmt.Errorf("af land: card %s does not exist", item.CardID)
}

// afReapDeadLocks expires the locks held by a process that is gone, so a landing lock whose holder died does not
// keep the queue for its whole ttl. ReapLocks frees locks and nothing else: no card lease is reclaimed and no note
// is written; taking cards back is `moai af reap`.
func afReapDeadLocks(ctx context.Context, st *ledger.Store) error {
	return st.ReapLocks(ctx, afAlive)
}

func newAFLandCmd() *cobra.Command {
	var f afMergeFlags
	var gate string
	var gateTimeout time.Duration
	cmd := &cobra.Command{
		Use:   "land <card>",
		Short: "Land an enqueued card: verify the merge outside the lock, then move the trunk under it",
		Long: fmt.Sprintf(`Land an enqueued card.

The card's queued candidate commit is merged into the trunk in a throwaway worktree, the gate command runs there
(outside the landing lock), and only then is the landing lock taken and the trunk ref moved with a compare-and-swap.
A gate that fails fails the card; a trunk that moved meanwhile is rebuilt and verified again. The result is one JSON
line (outcome: landed, failed or waiting).

What a failure records. The reason, this verb's output, the notes and the ledger carry the failure kind, the gate's
short id ("gate-" and 12 hex digits of the SHA-256 of the gate text), the exit status or timeout, the number of bytes
the gate printed and the path of a log file. They never carry the gate text or what the gate or git printed. The gate
text and the end of the output (at most %d KiB) are written to that log file, in the gate-logs directory beside the
ledger: mode 0600 in a 0700 directory, one file per card, candidate and kind, and at most %d MiB for the whole
directory (the oldest files go first). The log is masked as well as a list of shapes can mask it.
It can still hold secrets, and any process of your user can read it: read it yourself, and do not paste it into a
conversation or a report. When the log cannot be written the reason says "logging failed" and the failure is
recorded all the same.

Which trunk and which repository. The trunk is %s and cannot be changed: --trunk takes that value or nothing, and
any other value is refused before anything is landed, with the ledger unchanged. --repo takes the project directory or
any path with the same git directory (a linked worktree of the project's repository qualifies) and refuses every
other path the same way: a landing row says "merged" without asking where. Without --repo the repository is the project
directory (CLAUDE_PROJECT_DIR, else the working directory), and it too must be the repository the ledger is bound to: a
second clone of the project, or another repository, is refused with the ledger unchanged.

%s

What is recorded as given. Nothing masks the text of the free-input flags: --session here, and --command, --reason,
--path and --actor of the other af verbs are recorded as given, so put no secret in them.

The card must be in MERGING at the epoch its queue row was made at; a card its holder moved out of the merge queue
is refused. The gate runs with sh -c (native Windows has no sh: run this verb from WSL (or Git Bash, without process-group kill) there), in its
own process group that is killed when the gate returns or --gate-timeout ends: a background process the gate left
running is killed, not waited for, and does not fail a gate that exited 0. Under Git Bash a native Windows binary
has no process group, so this verb does not kill what the gate started there; run it from WSL when the gate starts
long-lived processes. A gate that prints without end is not an error: it runs until --gate-timeout (default 20m) ends it.
The gate receives the caller's full environment except GIT_DIR and the other repository-scoping variables, so a
credential in the caller's environment is visible to it.
A landing lock whose holder process is gone is taken back first.`, mergeq.DefaultFailureLogCap>>10, mergeq.DefaultFailureLogTotalCap>>20, mergeq.DefaultTrunk, afBindHelp),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(gate) == "" {
				return fmt.Errorf("af land: --gate is required: a landing without gates would merge unverified work")
			}
			return afMergeRun(cmd, f, func(ctx context.Context, st *ledger.Store, scratch string) error {
				q := mergeq.NewLedgerQueue(st)
				item, err := q.Item(ctx, args[0], time.Now())
				if err != nil {
					return fmt.Errorf("af land: %w", err)
				}
				if item.State == mergeq.StateFailed {
					return fmt.Errorf("af land: %s failed in the queue (queue it again with enqueue)", args[0])
				}
				if err := afCardMayLand(ctx, st, item); err != nil {
					return err
				}
				if err := afReapDeadLocks(ctx, st); err != nil {
					return fmt.Errorf("af land: take back a dead holder's lock: %w", err)
				}
				m := afMerger(f, st, scratch, afGate(gate, gateTimeout))
				m.Gate, m.GateText = afGateID(gate), gate
				res, err := m.Land(ctx, args[0], item.CandidateSHA)
				if err != nil {
					return err
				}
				return afPrint(cmd, map[string]any{
					"card_id": args[0], "outcome": res.Outcome, "reason": res.Reason,
					"trunk_before": res.TrunkBefore, "trunk_after": res.TrunkAfter, "verified": res.Verified,
				})
			})
		},
	}
	f.bind(cmd)
	cmd.Flags().StringVar(&gate, "gate", "", "the command that verifies the merge result, run with sh -c in the candidate worktree")
	cmd.Flags().DurationVar(&gateTimeout, "gate-timeout", afGateTimeout, "how long the gate may run before its process group is killed")
	return cmd
}

func newAFReconcileCmd() *cobra.Command {
	var f afMergeFlags
	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Settle the queue rows a restart left in verifying or landing, by git ancestry",
		Long: `Settle the queue rows a restart left behind.

A row whose candidate is already in the trunk is settled as landed; one whose candidate is not goes back to
waiting (an older landing row is settled as failed), and only while this command holds the landing lock. A verifying
row whose card still holds a live lease is left alone. Every row is judged on its own: one failure does not stop the
others. The result is one JSON line listing the actions taken; a failed row is reported as an error after it.

The trunk is ` + mergeq.DefaultTrunk + ` and cannot be changed: --trunk takes that value or nothing, and any other value is
refused with the ledger unchanged. --repo takes the project directory or any path with the same git directory (a
linked worktree of the project's repository qualifies) and refuses every other path the same way. Without --repo the
repository is the project directory (CLAUDE_PROJECT_DIR, else the working directory), and it must be the repository the
ledger is bound to: a reconcile from another repository would settle the rows from that repository's ancestry, so it is
refused with the ledger unchanged.

` + afBindHelp,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return afMergeRun(cmd, f, func(ctx context.Context, st *ledger.Store, scratch string) error {
				if err := afReapDeadLocks(ctx, st); err != nil {
					return fmt.Errorf("af reconcile: take back a dead holder's lock: %w", err)
				}
				items, err := mergeq.NewLedgerQueue(st).Items(ctx, time.Now())
				if err != nil {
					return err
				}
				acts, rerr := afMerger(f, st, scratch, nil).Reconcile(ctx, items)
				out := make([]map[string]string, 0, len(acts))
				for _, a := range acts {
					out = append(out, map[string]string{"card_id": a.CardID, "from": string(a.From), "to": string(a.To), "why": a.Why})
				}
				if perr := afPrint(cmd, map[string]any{"actions": out}); perr != nil {
					return perr
				}
				return rerr
			})
		},
	}
	f.bind(cmd)
	return cmd
}

// afMergeRun opens the ledger and a scratch directory for the candidate worktrees, runs fn under the landing
// timeout and cleans up.
func afMergeRun(cmd *cobra.Command, f afMergeFlags, fn func(ctx context.Context, st *ledger.Store, scratch string) error) error {
	// The trust boundary: both checks run before the ledger is opened, so a refusal changes nothing.
	if err := afCheckTrunk(f.trunk); err != nil {
		return err
	}
	if err := afCheckRepoFlag(cmd.Context(), f.repo); err != nil {
		return err
	}
	path, err := afLedgerPath(f.db)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), afLandTimeout)
	defer cancel()
	st, err := ledger.Open(ctx, path)
	if err != nil {
		return fmt.Errorf("af: open the ledger %s: %w", path, err)
	}
	defer func() { _ = st.Close() }()
	// the repository afMerger will use (--repo, else the project directory) must be the one the ledger is bound to
	repo := f.repo
	if repo == "" {
		repo = resolveProjectDir()
	}
	if err := afBindRepository(ctx, st, f.db, repo, false); err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "moai-af-land-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	return fn(ctx, st, scratch)
}
