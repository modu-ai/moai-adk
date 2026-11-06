package cli

// factory_runs_rebind_test.go — AC-FCR-022 (REQ-FCR-021/020/014), decision Q1
// (a): `moai factory runs --retire <dead-run> --rebind-to <live-run>` moves the
// dead run's in-progress cards under the live run and retires it. The command
// without --rebind-to keeps SPEC-FACTORY-RETIRE-GUARD-001 semantics (its own
// tests are unchanged). Every fixture is an isolated sandbox.

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// rebindCLIFixture seeds a dead run holding cardA (needs-decision) and cardB
// (picked), a live target that already holds cardB, and cardA's approval under
// the dead run. It chdirs into the sandbox root.
func rebindCLIFixture(t *testing.T) string {
	t.Helper()
	root := runOwnerSandbox(t)
	t.Chdir(root)
	deadPID, deadStart := deadOwnerIdentity(t)
	seedRun(t, root, "run-src", deadPID, deadStart)
	seedRun(t, root, "run-tgt", os.Getpid(), homestate.CurrentProcessFingerprint())
	seedCardRows(t, root, "run-src", [3]string{"cardA", "needs-decision", "lane-1"}, [3]string{"cardB", "picked", ""})
	seedCardRows(t, root, "run-tgt", [3]string{"cardB", "assigned", "lane-2"})
	db := openAdoptFactory(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO card_dispatch(card_id,run_id,recorded_at) VALUES('cardA','run-src','2026-10-14T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`INSERT INTO leader_approvals(card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at) VALUES('uuid-a','run-src','cardA',1,'hash','leader','leader','2026-10-14T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestFactoryRunsRetireRebindTo(t *testing.T) {
	t.Run("a operator moves the cards and retires the source", func(t *testing.T) {
		root := rebindCLIFixture(t)
		approvals := lcQuery(t, root, `SELECT group_concat(card_uuid||'|'||run_id||'|'||card_id||'|'||factory_version||'|'||evidence_hash||'|'||issued_at) FROM leader_approvals`)
		out, err := runFactoryCommand(t, "runs", "--retire", "run-src", "--rebind-to", "run-tgt")
		if err != nil {
			t.Fatalf("err = %v\n%s", err, out)
		}
		for _, want := range []string{"run-src", "run-tgt", "cardA", "cardB"} {
			if !strings.Contains(out, want) {
				t.Errorf("output lacks %q:\n%s", want, out)
			}
		}
		if got := runStatusOf(t, root, "run-src"); got != "retired" {
			t.Errorf("source status = %q, want retired", got)
		}
		if got := lcQuery(t, root, `SELECT run_id FROM cards WHERE card_id='cardA'`); got != "run-tgt" {
			t.Errorf("cardA lives under %q, want run-tgt", got)
		}
		if got := lcQuery(t, root, `SELECT state FROM cards WHERE run_id='run-src' AND card_id='cardB'`); got != "abandoned" {
			t.Errorf("source cardB = %q, want abandoned (the target already holds cardB)", got)
		}
		if got := lcQuery(t, root, `SELECT group_concat(card_uuid||'|'||run_id||'|'||card_id||'|'||factory_version||'|'||evidence_hash||'|'||issued_at) FROM leader_approvals`); got != approvals {
			t.Errorf("leader_approvals changed on the operator path:\n got %s\nwant %s", got, approvals)
		}
	})

	t.Run("b a lane session is refused", func(t *testing.T) {
		root := rebindCLIFixture(t)
		t.Setenv("MOAI_FACTORY_WORKER", "lane-1")
		before := lcDump(t, root)
		out, err := runFactoryCommand(t, "runs", "--retire", "run-src", "--rebind-to", "run-tgt")
		if err == nil {
			t.Fatalf("a lane session consolidated runs:\n%s", out)
		}
		if !strings.Contains(err.Error(), factoryLaneBoundarySentinel) {
			t.Errorf("lane refusal %q does not name the lane boundary", err.Error())
		}
		if after := lcDump(t, root); after != before {
			t.Errorf("the refusal changed state")
		}
	})

	t.Run("c a target that is not live is refused", func(t *testing.T) {
		root := rebindCLIFixture(t)
		deadPID, deadStart := deadOwnerIdentity(t)
		seedRun(t, root, "run-dead2", deadPID, deadStart)
		before := lcDump(t, root)
		for _, target := range []string{"run-dead2", "run-missing"} {
			out, err := runFactoryCommand(t, "runs", "--retire", "run-src", "--rebind-to", target)
			if err == nil {
				t.Fatalf("target %s accepted:\n%s", target, out)
			}
			if !strings.Contains(err.Error(), target) {
				t.Errorf("refusal %q does not name the rejected target %s", err.Error(), target)
			}
			if after := lcDump(t, root); after != before {
				t.Errorf("the refusal for target %s changed state", target)
			}
		}
	})

	t.Run("d without --rebind-to the retire guard is unchanged", func(t *testing.T) {
		root := rebindCLIFixture(t)
		before := lcDump(t, root)
		_, err := runFactoryCommand(t, "runs", "--retire", "run-src")
		if !errors.Is(err, homestate.ErrRunHasInProgressCards) {
			t.Fatalf("err = %v, want ErrRunHasInProgressCards", err)
		}
		if after := lcDump(t, root); after != before {
			t.Errorf("the refusal changed state")
		}
	})

	t.Run("g --rebind-to alone is refused", func(t *testing.T) {
		root := rebindCLIFixture(t)
		before := lcDump(t, root)
		out, err := runFactoryCommand(t, "runs", "--rebind-to", "run-tgt")
		if err == nil {
			t.Fatalf("--rebind-to without --retire accepted:\n%s", out)
		}
		if after := lcDump(t, root); after != before {
			t.Errorf("the refusal changed state")
		}
	})

	t.Run("h a source whose owner is not dead is refused", func(t *testing.T) {
		for _, tc := range []struct {
			name, class string
			pid         int
			start       string
		}{
			{"live", "live", os.Getpid(), homestate.CurrentProcessFingerprint()},
			{"indeterminate", "indeterminate", 0, ""},
		} {
			t.Run(tc.name, func(t *testing.T) {
				root := rebindCLIFixture(t)
				seedRun(t, root, "run-notdead", tc.pid, tc.start)
				seedCardRows(t, root, "run-notdead", [3]string{"cardZ", "run", "lane-9"})
				before := lcDump(t, root)
				out, err := runFactoryCommand(t, "runs", "--retire", "run-notdead", "--rebind-to", "run-tgt")
				if err == nil {
					t.Fatalf("a %s source was consolidated:\n%s", tc.class, out)
				}
				if !strings.Contains(err.Error()+out, "run-notdead") || !strings.Contains(err.Error()+out, tc.class) {
					t.Errorf("refusal %q does not name the run and its classification %q", err.Error()+out, tc.class)
				}
				if after := lcDump(t, root); after != before {
					t.Errorf("the refusal changed state")
				}
			})
		}
	})
}
