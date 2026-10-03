// factory_helper_test.go — the subprocess re-entry point for the cross-process
// criteria (the state-lock exclusion and the integration-window race).
//
// Sessions are distinct OS processes, and an in-process mutex passes a
// same-process test perfectly while protecting nothing in production
// (acceptance.md §A.4). The criteria therefore re-execute this test binary as
// a child process with MOAI_KANBAN_HELPER naming the operation; the child
// performs exactly that operation and exits.
package factory

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// substrateLockFileName names the lock artifact the lock-hold helper and the
// cross-process exclusion test share, beneath the test's temporary root.
const substrateLockFileName = "state.lock"

// integrationHelperStallCeiling bounds the child's own wait inside the
// interleaving hook, so a parent that dies mid-round cannot leave a child
// spinning. It is a backstop BEHIND the parent's exec.CommandContext deadline
// and its t.Cleanup-registered kill, never a substitute for either.
const integrationHelperStallCeiling = 30 * time.Second

// TestFactoryHelperProcess is not a real test: when MOAI_KANBAN_HELPER is unset
// it returns immediately so the normal `go test` run skips over it. When set,
// the child performs the named operation against the HELPER_* environment and
// exits — its stdout/stderr/exit code ARE the observation.
func TestFactoryHelperProcess(t *testing.T) {
	switch os.Getenv("MOAI_KANBAN_HELPER") {
	case "":
		return
	case "lock-hold":
		root := os.Getenv("HELPER_ROOT")
		release := os.Getenv("HELPER_RELEASE_FILE")
		lock, err := acquireStateLockImpl(filepath.Join(root, substrateLockFileName))
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
				_ = lock.release()
				os.Exit(5)
			}
			time.Sleep(10 * time.Millisecond)
		}
		if relErr := lock.release(); relErr != nil {
			fmt.Fprintf(os.Stderr, "lock-hold: release: %v\n", relErr)
			os.Exit(6)
		}
		os.Exit(0)
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
