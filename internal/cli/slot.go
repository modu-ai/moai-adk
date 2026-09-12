package cli

// slot.go — `moai slot`, the lane-facing surface of the resource slot lease
// (SPEC-RESOURCE-SLOT-LEASE-001, card t607).
//
// A lane about to run a heavy command acquires the resource first; a second
// lane's acquire is refused while the first holds it. The lease record lives
// in the PRIMARY checkout's .moai/state/slot-leases, shared by every linked
// worktree, and the root is resolved with kanban.ResolveSlotLeaseRoot — the
// same function the opt-in PreToolUse guard uses, so the CLI never writes a
// tree the guard does not read (plan-audit N1).
//
// Deliberately separate from `moai integration`: its own verbs, record, lock
// and config key.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/spf13/cobra"
)

// Exit codes a lane can branch on without parsing prose. Held and busy are
// different facts — someone owns the resource vs. a peer was mid-mutation —
// and get different codes.
const (
	slotExitHeld = 3
	slotExitBusy = 4
)

// slotNotGiven is how an omitted --name or --command reads on status.
const slotNotGiven = "(not given)"

// slotExitError carries an intentional exit code while keeping the kanban
// sentinel reachable through errors.Is.
type slotExitError struct {
	code int
	err  error
}

func (e *slotExitError) Error() string { return e.err.Error() }
func (e *slotExitError) Unwrap() error { return e.err }
func (e *slotExitError) ExitCode() int { return e.code }

// slotResult maps a kanban error to its exit code; other errors pass through.
func slotResult(err error) error {
	switch {
	case err == nil:
		return nil
	case kanban.IsSlotLeaseHeld(err):
		return &slotExitError{code: slotExitHeld, err: err}
	case kanban.IsSlotLeaseBusy(err):
		return &slotExitError{code: slotExitBusy, err: fmt.Errorf("%w — transient; retry shortly (the resource may be free)", err)}
	default:
		return err
	}
}

// slotLeaseRoot resolves the shared root: CLAUDE_PROJECT_DIR (else the cwd),
// normalized to the primary checkout by the resolver the guard also uses.
// Outside a git repository there is no worktree to disagree with, so the start
// directory itself is the root.
func slotLeaseRoot() string {
	start := strings.TrimSpace(os.Getenv(config.EnvClaudeProjectDir))
	if start == "" {
		start = resolveProjectDir()
	}
	if root, err := kanban.ResolveSlotLeaseRoot(start); err == nil {
		return root
	}
	return start
}

func newSlotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slot [command]",
		Short: "Acquire, inspect, and release a named heavy-resource slot (cross-session lease)",
		Long: `Acquire, inspect, and release a lease on a named heavy resource.

A session acquires the resource before a heavy command and releases it after;
another session's acquire is refused while a live holder is inside its declared
bound. The record lives in the primary checkout's .moai/state/slot-leases,
visible from every linked worktree.

A lease past its declared bound (--max-duration), or whose owning session is
gone, is taken over without --force; --force takes over a live holder and is
recorded. The opt-in PreToolUse guard (workflow.slot_lease.enabled, default
false) refuses configured commands while another session holds the resource;
these verbs work regardless.

Exit codes: 0 success, 3 held by another session, 4 busy (retry), 1 other.`,
	}
	cmd.AddCommand(newSlotAcquireCmd(), newSlotStatusCmd(), newSlotReleaseCmd())
	return cmd
}

func newSlotAcquireCmd() *cobra.Command {
	var resource, sessionFlag, nameFlag, commandFlag, boundFlag string
	var force, jsonOut bool
	cmd := &cobra.Command{
		Use:   "acquire",
		Short: "Record this session as the holder of a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := kanban.ValidateSlotResourceName(resource); err != nil {
				return err
			}
			sessionID := integrationSessionID(sessionFlag)
			if sessionID == "" {
				return fmt.Errorf("cannot resolve this session's id; pass --session <id> (a lease with an invented holder can be neither released by its holder nor recognized by the guard)")
			}
			root := slotLeaseRoot()
			boundText := strings.TrimSpace(boundFlag)
			if boundText == "" {
				boundText = config.LoadSlotLeaseDefaultMaxDuration(root)
			}
			bound, err := kanban.ParseSlotLeaseMaxDuration(boundText)
			if err != nil {
				return err
			}
			// The OWNING SESSION's pid, never this process's: this command
			// exits the moment it returns, so its own pid would read stale at
			// once. Unresolvable is recorded as 0, which reads live.
			ownerPID, _ := session.ResolveOwnerPID()
			lease, err := kanban.AcquireSlotLease(root, kanban.SlotLeaseRequest{
				Resource:    resource,
				SessionID:   sessionID,
				SessionName: nameFlag,
				PID:         ownerPID,
				Command:     commandFlag,
				MaxDuration: bound,
				Force:       force,
			})
			if err != nil {
				return slotResult(err)
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return json.NewEncoder(out).Encode(map[string]any{
					"acquired":  true,
					"lease":     lease,
					"displaced": lease.Displaced,
					"root":      root,
				})
			}
			_, _ = fmt.Fprintf(out, "slot %s acquired by %s until %s\n", resource, sessionID, lease.ExpiresAt)
			if d := lease.Displaced; d != nil {
				// Never silent: the displaced session may still be running.
				_, _ = fmt.Fprintf(out, "  displaced: %s (pid %d), reason %s, held since %s\n", d.SessionID, d.PID, d.Reason, d.AcquiredAt)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&resource, "resource", "", "Resource name (a-z, 0-9, '-', 1-64 characters)")
	cmd.Flags().StringVar(&sessionFlag, "session", "", "Session id to record as holder (default: this session)")
	cmd.Flags().StringVar(&nameFlag, "name", "", "Human-facing lane name recorded alongside the id")
	cmd.Flags().StringVar(&commandFlag, "command", "", "The command this lease is for, recorded for whoever reads status")
	cmd.Flags().StringVar(&boundFlag, "max-duration", "", "Declared maximum duration, e.g. 20m (default: workflow.slot_lease.default_max_duration)")
	cmd.Flags().BoolVar(&force, "force", false, "Take the resource over from a live holder (recorded, never silent)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	_ = cmd.MarkFlagRequired("resource")
	return cmd
}

// slotStatusView is one resource's status as reported by `moai slot status`.
type slotStatusView struct {
	Resource string            `json:"resource"`
	Held     bool              `json:"held"`
	Stale    bool              `json:"stale"`
	Expired  bool              `json:"expired"`
	Lease    *kanban.SlotLease `json:"lease,omitempty"`
}

func slotView(resource string, lease *kanban.SlotLease, now time.Time) slotStatusView {
	v := slotStatusView{Resource: resource, Held: lease.Held()}
	if v.Held {
		v.Stale, v.Expired, v.Lease = lease.Stale(), lease.Expired(now), lease
	}
	return v
}

func orNotGiven(s string) string {
	if strings.TrimSpace(s) == "" {
		return slotNotGiven
	}
	return s
}

func writeSlotView(w io.Writer, v slotStatusView) {
	if !v.Held {
		_, _ = fmt.Fprintf(w, "slot %s: free\n", v.Resource)
		return
	}
	state := "held"
	switch {
	case v.Stale:
		state = "held by a session that is gone (reclaimable)"
	case v.Expired:
		state = "past its declared bound (reclaimable)"
	}
	l := v.Lease
	_, _ = fmt.Fprintf(w, "slot %s: %s\n  session: %s (pid %d)\n  name:    %s\n  command: %s\n  since:   %s\n  bound:   %s, ends %s\n",
		v.Resource, state, l.SessionID, l.PID, orNotGiven(l.SessionName), orNotGiven(l.Command), l.AcquiredAt, l.MaxDuration, l.ExpiresAt)
	if d := l.Displaced; d != nil {
		_, _ = fmt.Fprintf(w, "  displaced: %s (pid %d), reason %s at %s\n", d.SessionID, d.PID, d.Reason, d.At)
	}
}

// listSlotResources returns the resource names that have a record under root.
func listSlotResources(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, ".moai", "state", kanban.SlotLeaseDirName))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".json")
		if !ok || e.IsDir() || kanban.ValidateSlotResourceName(name) != nil {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func newSlotStatusCmd() *cobra.Command {
	var resource string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report who holds a resource (all recorded resources when --resource is omitted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := slotLeaseRoot()
			now := time.Now()
			out := cmd.OutOrStdout()
			if resource != "" {
				lease, err := kanban.ReadSlotLease(root, resource)
				if err != nil {
					return err
				}
				v := slotView(resource, lease, now)
				if jsonOut {
					return json.NewEncoder(out).Encode(map[string]any{
						"resource": v.Resource, "held": v.Held, "stale": v.Stale, "expired": v.Expired,
						"lease": v.Lease, "root": root,
					})
				}
				writeSlotView(out, v)
				return nil
			}
			names, err := listSlotResources(root)
			if err != nil {
				return err
			}
			views := make([]slotStatusView, 0, len(names))
			for _, name := range names {
				lease, readErr := kanban.ReadSlotLease(root, name)
				if readErr != nil {
					return readErr
				}
				views = append(views, slotView(name, lease, now))
			}
			if jsonOut {
				return json.NewEncoder(out).Encode(map[string]any{"leases": views, "root": root})
			}
			if len(views) == 0 {
				_, _ = fmt.Fprintln(out, "no slot leases recorded")
				return nil
			}
			for _, v := range views {
				writeSlotView(out, v)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&resource, "resource", "", "Resource name (default: every recorded resource)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	return cmd
}

func newSlotReleaseCmd() *cobra.Command {
	var resource, sessionFlag string
	var force, jsonOut bool
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release a resource this session holds",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := kanban.ValidateSlotResourceName(resource); err != nil {
				return err
			}
			sessionID := integrationSessionID(sessionFlag)
			if sessionID == "" && !force {
				return fmt.Errorf("cannot resolve this session's id; pass --session <id> or --force")
			}
			released, err := kanban.ReleaseSlotLease(slotLeaseRoot(), resource, sessionID, force)
			if err != nil {
				return slotResult(err)
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"released": true, "lease": released})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "slot %s released (was %s)\n", resource, released.SessionID)
			return nil
		},
	}
	cmd.Flags().StringVar(&resource, "resource", "", "Resource name")
	cmd.Flags().StringVar(&sessionFlag, "session", "", "Session id whose lease to release (default: this session)")
	cmd.Flags().BoolVar(&force, "force", false, "Release a lease held by a different session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit machine-readable JSON")
	_ = cmd.MarkFlagRequired("resource")
	return cmd
}

func init() {
	rootCmd.AddCommand(newSlotCmd())
}
