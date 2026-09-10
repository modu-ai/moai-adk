// kanban_helper_test.go — the subprocess re-entry point for the cross-process
// criteria (AC-KB-017 carrier half, AC-KB-019, AC-KB-023).
//
// Sessions are distinct OS processes, and an in-process mutex passes a
// same-process test perfectly while protecting nothing in production
// (acceptance.md §A.4). The criteria therefore re-execute this test binary as
// a child process with MOAI_KANBAN_HELPER naming the operation; the child
// performs exactly that operation and exits.
package kanban

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

// integrationHelperStallCeiling bounds the child's own wait inside the
// interleaving hook, so a parent that dies mid-round cannot leave a child
// spinning. It is a backstop BEHIND the parent's exec.CommandContext deadline
// and its t.Cleanup-registered kill, never a substitute for either.
const integrationHelperStallCeiling = 30 * time.Second

// TestKanbanHelperProcess is not a real test: when MOAI_KANBAN_HELPER is unset
// it returns immediately so the normal `go test` run skips over it. When set,
// the child performs the named operation against the HELPER_* environment and
// exits — its stdout/stderr/exit code ARE the observation.
func TestKanbanHelperProcess(t *testing.T) {
	switch os.Getenv("MOAI_KANBAN_HELPER") {
	case "":
		return
	case "resolve-role":
		role, err := ResolveDeclaredRole(os.Getenv("HELPER_ROOT"), os.Getenv("HELPER_SESSION"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve-role: %v\n", err)
			os.Exit(2)
		}
		_, _ = fmt.Fprintln(os.Stdout, role)
		os.Exit(0)
	case "lock-hold":
		root := os.Getenv("HELPER_ROOT")
		release := os.Getenv("HELPER_RELEASE_FILE")
		lock, err := AcquireBoardLock(root)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stdout, "HELD")
			fmt.Fprintf(os.Stderr, "lock-hold: %v\n", err)
			os.Exit(3)
		}
		_, _ = fmt.Fprintln(os.Stdout, "ACQUIRED")
		// Wait for the parent's release flag (bounded).
		deadline := time.Now().Add(60 * time.Second)
		for {
			if _, statErr := os.Stat(release); statErr == nil {
				break
			}
			if time.Now().After(deadline) {
				fmt.Fprintln(os.Stderr, "lock-hold: release flag never appeared")
				_ = lock.Release()
				os.Exit(5)
			}
			time.Sleep(10 * time.Millisecond)
		}
		if relErr := lock.Release(); relErr != nil {
			fmt.Fprintf(os.Stderr, "lock-hold: release: %v\n", relErr)
			os.Exit(6)
		}
		os.Exit(0)
	case "reacquire-lock":
		// Takes the lock and records ITS identity in the artifact, then
		// exits: on Unix flock releases on exit while the RECORDED identity
		// stays — exactly the changed-hands artifact the clear's re-read
		// must observe.
		lock, err := AcquireBoardLock(os.Getenv("HELPER_ROOT"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "reacquire-lock: %v\n", err)
			os.Exit(4)
		}
		_ = lock.Release()
		os.Exit(0)
	case "reader-loop":
		// A separate-process board reader (AC-KB-018's §A.4 form): read the
		// board repeatedly for HELPER_DURATION seconds; every read must be a
		// whole board. Prints "READS=<n> FAILURES=<m>" and exits 8 on any
		// torn observation.
		duration, _ := time.ParseDuration(os.Getenv("HELPER_DURATION") + "s")
		root := os.Getenv("HELPER_ROOT")
		reads, failures := 0, 0
		deadline := time.Now().Add(duration)
		for time.Now().Before(deadline) {
			_, err := LoadBoard(root)
			reads++
			if err != nil {
				failures++
				fmt.Fprintf(os.Stderr, "reader-loop: torn read: %v\n", err)
			}
		}
		_, _ = fmt.Fprintf(os.Stdout, "READS=%d FAILURES=%d\n", reads, failures)
		if failures > 0 {
			os.Exit(8)
		}
		os.Exit(0)
	case "transition-run":
		// TransitionIntoRun as the declared lead; exit 3 on the WIP refusal
		// so the parent can distinguish outcomes.
		err := TransitionIntoRun(os.Getenv("HELPER_ROOT"), os.Getenv("HELPER_SESSION"), os.Getenv("HELPER_SPEC"))
		if err == nil {
			_, _ = fmt.Fprintln(os.Stdout, "OK")
			os.Exit(0)
		}
		if IsWipLimitExceeded(err) {
			_, _ = fmt.Fprintln(os.Stdout, "WIP")
			os.Exit(3)
		}
		fmt.Fprintf(os.Stderr, "transition-run: %v\n", err)
		os.Exit(7)
	case "integration-acquire":
		// AcquireIntegrationLock as one lane's session. The recorded PID is the
		// PARENT test process's, passed in HELPER_OWNER_PID, mirroring the
		// production verb (internal/cli/integration.go records
		// session.ResolveOwnerPID(), never its own pid). Recording os.Getpid()
		// here would be a defect, not a shortcut: this child exits the instant
		// it writes, its record then reads STALE, and the second child takes it
		// over LEGITIMATELY — producing successes=2 even under a correct
		// mutation lock, which is a stale reclaim wearing the race's clothes.
		root := os.Getenv("HELPER_ROOT")
		session := os.Getenv("HELPER_SESSION")
		ownerPID, pidErr := strconv.Atoi(os.Getenv("HELPER_OWNER_PID"))
		if pidErr != nil {
			fmt.Fprintf(os.Stderr, "integration-acquire: HELPER_OWNER_PID %q: %v\n", os.Getenv("HELPER_OWNER_PID"), pidErr)
			os.Exit(10)
		}
		if marker := os.Getenv("HELPER_STALL_MARKER"); marker != "" {
			proceed := os.Getenv("HELPER_PROCEED_FLAG")
			integrationLockMutationTestHook = func() {
				if werr := os.WriteFile(marker, []byte("STALLED\n"), 0o644); werr != nil {
					fmt.Fprintf(os.Stderr, "integration-acquire: stall marker: %v\n", werr)
					return
				}
				deadline := time.Now().Add(integrationHelperStallCeiling)
				for {
					if _, statErr := os.Stat(proceed); statErr == nil {
						return
					}
					if time.Now().After(deadline) {
						fmt.Fprintln(os.Stderr, "integration-acquire: proceed flag never appeared within the child ceiling")
						return
					}
					time.Sleep(5 * time.Millisecond)
				}
			}
		}
		replaced, acqErr := AcquireIntegrationLock(root, IntegrationLock{
			SessionID: session,
			PID:       ownerPID,
			PIDSource: PIDSourceSessionOwner,
			Branch:    "release/helper",
			Worktree:  root,
		}, false)
		result := integrationAcquireOutcome(acqErr)
		replacedField := "none"
		if replaced != nil && replaced.SessionID != "" {
			replacedField = replaced.SessionID
		}
		// SESSION is the second discriminator: the parent ASSERTS the two
		// children's ids differ rather than assuming its own setup passed two.
		// The same-session path is a LEGAL re-acquire, so an undetected id
		// collision would yield successes=2 with REPLACED=none on both and a
		// live winner — every other check satisfied, the race never exercised.
		_, _ = fmt.Fprintf(os.Stdout, "RESULT=%s REPLACED=%s SESSION=%s\n", result, replacedField, session)
		if result == "error" {
			fmt.Fprintf(os.Stderr, "integration-acquire: %v\n", acqErr)
			os.Exit(9)
		}
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown helper op %q\n", os.Getenv("MOAI_KANBAN_HELPER"))
		os.Exit(64)
	}
}
