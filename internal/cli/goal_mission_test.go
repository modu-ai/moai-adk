package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/mission"
)

func TestNewAutoMissionCommandTreatsTextAsData(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	marker := filepath.Join(t.TempDir(), "must-not-exist")
	text := "한국어 임무; $(touch " + marker + ") && echo injected"
	cmd := newGoalCmd()
	cmd.SetContext(context.Background())
	sessionID := "018f4f4a-7b7c-7a11-8f4d-555555555555"
	cmd.SetArgs([]string{"--auto", "--session", sessionID, text})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	stored, err := mission.LoadAutoMission(root, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Text != text || stored.MissionMode != mission.ModeAuto || stored.ProgressionMode != goal.DefaultProgressionMode {
		t.Fatalf("stored mission = %+v", stored)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("mission text was interpreted as a command: %v", err)
	}
}

func governanceCLIArgs(t *testing.T, root, session string, action mission.Action, target string, revision int64, head string) []string {
	t.Helper()
	state, err := mission.LoadAutoMission(root, session)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := autoSnapshotHash(state.ContractHash, target, revision)
	dir := filepath.Join(root, ".moai", "state", "mission", "governance")
	decisionPath := filepath.Join(dir, string(action)+"-decision.json")
	auditPath := filepath.Join(dir, string(action)+"-audit.json")
	base := mission.GovernanceReceipt{Version: 1, MissionID: session, ContractHash: state.ContractHash, SnapshotHash: snapshot, Action: action, Targets: []string{target}, ExpiresAt: time.Now().Add(time.Hour), HeadSHA: head}
	decision := base
	decision.Kind, decision.Issuer, decision.Status = mission.GovernanceDecision, "mission-governor", mission.GovernanceRecommended
	audit := base
	audit.Kind, audit.Issuer, audit.Status = mission.GovernanceAudit, "sync-auditor", mission.GovernancePassed
	if err := mission.WriteGovernanceReceipt(root, decisionPath, decision); err != nil {
		t.Fatal(err)
	}
	if err := mission.WriteGovernanceReceipt(root, auditPath, audit); err != nil {
		t.Fatal(err)
	}
	return []string{"--governor-receipt", decisionPath, "--audit-receipt", auditPath}
}

func TestAutoMissionLifecycleApprovesSealsAndRunsPublishedOperation(t *testing.T) {
	root, store := todoFixture(t)
	ctx := context.Background()
	item, err := kanban.CaptureGTDItem(ctx, store, kanban.CaptureInput{Content: "approved mission card", Source: "user", SourceAllowed: true, Sensitivity: kanban.SensitivityPrivate, EventID: "mission-cli"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = kanban.ClarifyGTDItem(ctx, store, kanban.ClarifyInput{ItemID: item.ItemID, Disposition: kanban.DispositionAction, DesiredOutcome: "landed", CompletionEvidence: "CI", Authority: "queue", SourceTrusted: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = kanban.OrganizeGTDItem(ctx, store, kanban.OrganizeInput{ItemID: item.ItemID, Class: kanban.ClassAction})
	if err != nil {
		t.Fatal(err)
	}
	sessionID := "018f4f4a-7b7c-7a11-8f4d-666666666666"
	run := func(args ...string) (string, error) {
		cmd := newGoalCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(args)
		err := cmd.Execute()
		return out.String(), err
	}
	if _, err := run("--auto", "--session", sessionID, "finish the approved item; $(false)"); err != nil {
		t.Fatal(err)
	}
	if _, err := run("run", "--session", sessionID, "--action", "publish", "--target", "gtd:"+item.ItemID, "--recommend"); err == nil {
		t.Fatal("unapproved mission ran")
	}
	if rec, _ := store.Load(); len(rec.Items) != 0 {
		t.Fatal("unapproved mission changed queue")
	}
	if _, err := run("approve", "--session", sessionID, "--scope", "gtd:"+item.ItemID, "--action", "publish", "--completion-evidence", "CI"); err != nil {
		t.Fatal(err)
	}
	if _, err := run("run", "--session", sessionID, "--action", "publish", "--target", "gtd:"+item.ItemID); err == nil {
		t.Fatal("operation proceeded without governor recommendation")
	}
	if _, err := run("resume", "--session", sessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := run("run", "--session", sessionID, "--action", "publish", "--target", "repo:outside", "--recommend"); err == nil {
		t.Fatal("out-of-scope operation allowed")
	}
	if _, err := run("resume", "--session", sessionID); err != nil {
		t.Fatal(err)
	}
	current, err := kanban.LoadGTDItem(ctx, store, item.ItemID)
	if err != nil {
		t.Fatal(err)
	}
	govArgs := governanceCLIArgs(t, root, sessionID, mission.ActionPublish, "gtd:"+item.ItemID, current.SourceRevision, gitFixtureCLI(t, root, "rev-parse", "HEAD"))
	if _, err := run(append([]string{"run", "--session", sessionID, "--action", "publish", "--target", "gtd:" + item.ItemID}, govArgs...)...); err != nil {
		t.Fatal(err)
	}
	stored, err := mission.LoadAutoMission(root, sessionID)
	if err != nil || stored.State != mission.StateRunning || stored.ContractHash == "" || len(stored.OperationIDs) != 1 {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	rec, err := store.Load()
	if err != nil || len(rec.Items) != 1 {
		t.Fatalf("queue=%+v err=%v", rec, err)
	}
	out, err := run("status", "--session", sessionID, "--json")
	if err != nil {
		t.Fatal(err)
	}
	var status mission.AutoMission
	if err := json.Unmarshal([]byte(out), &status); err != nil || status.SessionID != sessionID {
		t.Fatalf("status=%q err=%v", out, err)
	}
	if _, err := run("revoke", "--session", sessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := run("resume", "--session", sessionID); err == nil {
		t.Fatal("revoked mission resumed")
	}
}

func TestAutoMissionProductionGitCommitOwner(t *testing.T) {
	root, _ := todoFixture(t)
	git := func(args ...string) string { return gitFixtureCLI(t, root, args...) }
	git("checkout", "-qb", "WT-auto")
	git("config", "user.name", "Test")
	git("config", "user.email", "test@example.invalid")
	if err := os.WriteFile(filepath.Join(root, "owned.txt"), []byte("owned"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "foreign.txt"), []byte("foreign"), 0600); err != nil {
		t.Fatal(err)
	}
	session := "018f4f4a-7b7c-7a11-8f4d-aaaaaaaaaaaa"
	run := func(args ...string) error {
		cmd := newGoalCmd()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	if err := run("--auto", "--session", session, "commit approved file"); err != nil {
		t.Fatal(err)
	}
	target := "repo:" + root
	if err := run("approve", "--session", session, "--scope", target, "--action", "commit", "--completion-evidence", "commit"); err != nil {
		t.Fatal(err)
	}
	baseArgs := []string{"run", "--session", session, "--action", "commit", "--target", target, "--repo", root, "--worktree-branch", "WT-auto", "--path", "owned.txt", "--message", "feat: auto owner"}
	if err := run(baseArgs...); err == nil {
		t.Fatal("commit accepted without authoritative test receipt")
	}
	if err := run("resume", "--session", session); err != nil {
		t.Fatal(err)
	}
	receipt := filepath.Join(root, ".moai", "reports", "tests.json")
	if err := os.MkdirAll(filepath.Dir(receipt), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(receipt, []byte(`{"head_sha":"`+git("rev-parse", "HEAD")+`","status":"passed"}`), 0600); err != nil {
		t.Fatal(err)
	}
	baseArgs = append(baseArgs, governanceCLIArgs(t, root, session, mission.ActionCommit, target, 1, git("rev-parse", "HEAD"))...)
	if err := run(append(baseArgs, "--tests-receipt", receipt)...); err != nil {
		t.Fatal(err)
	}
	if got := git("status", "--porcelain"); !strings.Contains(got, "?? foreign.txt") || strings.Contains(got, "A  foreign.txt") {
		t.Fatalf("foreign path staging state: %q", got)
	}
	if !strings.Contains(git("log", "-1", "--format=%B"), "MoAI-Operation:") {
		t.Fatal("operation receipt trailer missing")
	}
}

func TestAutoMissionProductionLocalDevelopMergeOwner(t *testing.T) {
	root, _ := todoFixture(t)
	git := func(args ...string) string { return gitFixtureCLI(t, root, args...) }
	git("config", "user.name", "Test")
	git("config", "user.email", "test@example.invalid")
	git("checkout", "-qb", "develop")
	if err := os.WriteFile(filepath.Join(root, "base.txt"), []byte("base"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "base.txt")
	git("commit", "-qm", "base")
	base := git("rev-parse", "HEAD")
	git("checkout", "-qb", "WT-card")
	if err := os.WriteFile(filepath.Join(root, "card.txt"), []byte("card"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "card.txt")
	git("commit", "-qm", "card")
	card := git("rev-parse", "HEAD")
	git("checkout", "-q", "develop")
	session := "018f4f4a-7b7c-7a11-8f4d-bbbbbbbbbbbb"
	lease := filepath.Join(root, ".git", "auto-lease.json")
	if err := mission.WriteIntegrationLease(lease, mission.IntegrationLease{SessionID: session, BaseSHA: base}); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) error {
		cmd := newGoalCmd()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	target := "repo:" + root
	if err := run("--auto", "--session", session, "merge approved card"); err != nil {
		t.Fatal(err)
	}
	if err := run("approve", "--session", session, "--scope", target, "--action", "local_develop_merge", "--completion-evidence", "merge"); err != nil {
		t.Fatal(err)
	}
	if err := run("run", "--session", session, "--action", "local_develop_merge", "--target", target, "--repo", root, "--worktree-branch", "WT-card", "--card-sha", card, "--base-sha", base, "--integration-lease", filepath.Join(root, ".git", "missing-lease")); err == nil {
		t.Fatal("local merge accepted without integration lease")
	}
	if err := run("resume", "--session", session); err != nil {
		t.Fatal(err)
	}
	mergeArgs := []string{"run", "--session", session, "--action", "local_develop_merge", "--target", target, "--repo", root, "--worktree-branch", "WT-card", "--card-sha", card, "--base-sha", base, "--integration-lease", lease}
	mergeArgs = append(mergeArgs, governanceCLIArgs(t, root, session, mission.ActionLocalMerge, target, 1, base)...)
	if err := run(mergeArgs...); err != nil {
		t.Fatal(err)
	}
	if parents := strings.Fields(git("show", "-s", "--format=%P", "HEAD")); len(parents) != 2 {
		t.Fatalf("not no-ff: %v", parents)
	}
}

// testGitBinary resolves git from PATH so the fixtures run on every CI
// platform instead of only where a host-specific install path exists.
func testGitBinary(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("git not found on PATH: %v", err)
	}
	return path
}

func gitFixtureCLI(t *testing.T, root string, args ...string) string {
	t.Helper()
	all := append([]string{"-C", root}, args...)
	out, err := exec.Command(testGitBinary(t), all...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestAutoMissionTestsReceiptTrustBoundary(t *testing.T) {
	repo := t.TempDir()
	head := strings.Repeat("a", 40)
	inside := filepath.Join(repo, "receipt.json")
	write := func(path, status, sha string, mode os.FileMode) {
		t.Helper()
		if err := os.WriteFile(path, []byte(`{"head_sha":"`+sha+`","status":"`+status+`"}`), mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
	}
	write(inside, "failed", head, 0600)
	if err := validateMissionTestsReceipt(repo, inside, head); err == nil {
		t.Fatal("failed tests receipt accepted")
	}
	write(inside, "passed", strings.Repeat("b", 40), 0600)
	if err := validateMissionTestsReceipt(repo, inside, head); err == nil {
		t.Fatal("stale tests receipt accepted")
	}
	write(inside, "passed", head, 0644)
	if err := validateMissionTestsReceipt(repo, inside, head); err == nil {
		t.Fatal("weak receipt permissions accepted")
	}
	write(inside, "passed", head, 0600)
	outside := filepath.Join(t.TempDir(), "receipt.json")
	write(outside, "passed", head, 0600)
	if err := validateMissionTestsReceipt(repo, outside, head); err == nil {
		t.Fatal("out-of-scope receipt accepted")
	}
	link := filepath.Join(repo, "receipt-link.json")
	if err := os.Symlink(inside, link); err != nil {
		t.Fatal(err)
	}
	if err := validateMissionTestsReceipt(repo, link, head); err == nil {
		t.Fatal("symlink receipt accepted")
	}
	if err := validateMissionTestsReceipt(repo, inside, head); err != nil {
		t.Fatalf("valid receipt rejected: %v", err)
	}
}

func TestAutoMissionLifecycleRefusalAndJSONBranches(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	run := func(args ...string) (string, error) {
		cmd := newGoalCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(args)
		return out.String(), cmd.Execute()
	}
	session := "018f4f4a-7b7c-7a11-8f4d-cccccccccccc"
	if _, err := run("--auto", "--session", session); err == nil {
		t.Fatal("empty mission accepted")
	}
	if _, err := run("--auto", "--session", "invalid", "mission"); err == nil {
		t.Fatal("invalid session accepted")
	}
	if _, err := run("--auto", "--json", "--session", session, "json mission"); err != nil {
		t.Fatalf("json create err=%v", err)
	}
	missing := "018f4f4a-7b7c-7a11-8f4d-dddddddddddd"
	if _, err := loadRequiredAutoMission(root, "invalid"); err == nil {
		t.Fatal("invalid load session accepted")
	}
	if _, err := loadRequiredAutoMission(root, missing); err == nil {
		t.Fatal("missing mission loaded")
	}
	if _, err := run("approve", "--session", session); err == nil {
		t.Fatal("incomplete contract sealed")
	}
	if _, err := run("revoke", "--session", missing); err == nil {
		t.Fatal("missing mission revoked")
	}
	if _, err := run("resume", "--session", session); err == nil {
		t.Fatal("draft mission resumed")
	}
}

func TestAutoMissionTamperedSealedContractBlocksBeforeQueueEffect(t *testing.T) {
	root, store := todoFixture(t)
	ctx := context.Background()
	prepare := func(event, content string) kanban.GTDItem {
		item, err := kanban.CaptureGTDItem(ctx, store, kanban.CaptureInput{Content: content, Source: "user", SourceAllowed: true, Sensitivity: kanban.SensitivityPrivate, EventID: event})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := kanban.ClarifyGTDItem(ctx, store, kanban.ClarifyInput{ItemID: item.ItemID, Disposition: kanban.DispositionAction, DesiredOutcome: "done", CompletionEvidence: "CI", Authority: "queue", SourceTrusted: true}); err != nil {
			t.Fatal(err)
		}
		if _, err := kanban.OrganizeGTDItem(ctx, store, kanban.OrganizeInput{ItemID: item.ItemID, Class: kanban.ClassAction}); err != nil {
			t.Fatal(err)
		}
		return item
	}
	a := prepare("tamper-a", "approved A")
	b := prepare("tamper-b", "unapproved B")
	session := "018f4f4a-7b7c-7a11-8f4d-eeeeeeeeeeee"
	run := func(args ...string) error {
		cmd := newGoalCmd()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	if err := run("--auto", "--session", session, "publish approved item"); err != nil {
		t.Fatal(err)
	}
	if err := run("approve", "--session", session, "--scope", "gtd:"+a.ItemID, "--action", "publish", "--completion-evidence", "CI"); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(root, ".moai", "state", "mission", session+".json")
	raw, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	var state mission.AutoMission
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatal(err)
	}
	state.Contract.Scope = []string{"gtd:" + b.ItemID}
	raw, _ = json.Marshal(state)
	if err := os.WriteFile(statePath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := run("run", "--session", session, "--action", "publish", "--target", "gtd:"+b.ItemID, "--recommend"); err == nil || !strings.Contains(err.Error(), "contract_integrity_mismatch") {
		t.Fatalf("tampered contract err=%v", err)
	}
	record, err := store.LoadPure()
	if err != nil || len(record.Items) != 0 {
		t.Fatalf("tampered contract changed queue: %+v err=%v", record.Items, err)
	}
	raw, err = os.ReadFile(statePath)
	if err != nil || json.Unmarshal(raw, &state) != nil || state.State != mission.StateBlocked || state.LastBlocker != "contract_integrity_mismatch" {
		t.Fatalf("blocked state=%+v err=%v", state, err)
	}
}

func TestGTDAutonomyEndToEndProductionCLIWithGovernanceReceipts(t *testing.T) {
	root, store := todoFixture(t)
	ctx := context.Background()
	item, err := kanban.CaptureGTDItem(ctx, store, kanban.CaptureInput{Content: "production auto flow", Source: "user", SourceAllowed: true, Sensitivity: kanban.SensitivityPrivate, EventID: "production-auto-flow"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.ClarifyGTDItem(ctx, store, kanban.ClarifyInput{ItemID: item.ItemID, Disposition: kanban.DispositionAction, DesiredOutcome: "dispatched", CompletionEvidence: "assignment", Authority: "queue,dispatch", SourceTrusted: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.OrganizeGTDItem(ctx, store, kanban.OrganizeInput{ItemID: item.ItemID, Class: kanban.ClassAction}); err != nil {
		t.Fatal(err)
	}
	session := "018f4f4a-7b7c-7a11-8f4d-f11111111111"
	target := "gtd:" + item.ItemID
	run := func(args ...string) error {
		cmd := newGoalCmd()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	if err := run("--auto", "--session", session, "publish pick dispatch"); err != nil {
		t.Fatal(err)
	}
	if err := run("approve", "--session", session, "--scope", target, "--action", "publish", "--action", "pick", "--action", "dispatch", "--completion-evidence", "assignment"); err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.AcquireSlotLease(root, kanban.SlotLeaseRequest{Resource: "worker-10", SessionID: session, MaxDuration: time.Hour}); err != nil {
		t.Fatal(err)
	}
	head := gitFixtureCLI(t, root, "rev-parse", "HEAD")
	writeReceipts := func(action mission.Action, revision int64) (string, string) {
		t.Helper()
		state, err := mission.LoadAutoMission(root, session)
		if err != nil {
			t.Fatal(err)
		}
		snapshot := autoSnapshotHash(state.ContractHash, target, revision)
		dir := filepath.Join(root, ".moai", "state", "mission", "governance")
		decisionPath := filepath.Join(dir, string(action)+"-decision.json")
		auditPath := filepath.Join(dir, string(action)+"-audit.json")
		base := mission.GovernanceReceipt{Version: 1, MissionID: session, ContractHash: state.ContractHash, SnapshotHash: snapshot, Action: action, Targets: []string{target}, ExpiresAt: time.Now().Add(time.Hour), HeadSHA: head}
		decision := base
		decision.Kind = mission.GovernanceDecision
		decision.Issuer = "mission-governor"
		decision.Status = mission.GovernanceRecommended
		audit := base
		audit.Kind = mission.GovernanceAudit
		audit.Issuer = "sync-auditor"
		audit.Status = mission.GovernancePassed
		if err := mission.WriteGovernanceReceipt(root, decisionPath, decision); err != nil {
			t.Fatal(err)
		}
		if err := mission.WriteGovernanceReceipt(root, auditPath, audit); err != nil {
			t.Fatal(err)
		}
		return decisionPath, auditPath
	}
	current, err := kanban.LoadGTDItem(ctx, store, item.ItemID)
	if err != nil {
		t.Fatal(err)
	}
	writeReceipts(mission.ActionPublish, current.SourceRevision)
	writeReceipts(mission.ActionPick, current.SourceRevision+1)
	writeReceipts(mission.ActionDispatch, current.SourceRevision+1)
	governorPattern := filepath.Join(root, ".moai", "state", "mission", "governance", "{action}-decision.json")
	auditPattern := filepath.Join(root, ".moai", "state", "mission", "governance", "{action}-audit.json")
	approvedState, _ := mission.LoadAutoMission(root, session)
	if _, err := mission.LoadGovernanceReceipts(root, strings.ReplaceAll(governorPattern, "{action}", "publish"), strings.ReplaceAll(auditPattern, "{action}", "publish"), mission.GovernanceExpectation{MissionID: session, ContractHash: approvedState.ContractHash, SnapshotHash: autoSnapshotHash(approvedState.ContractHash, target, current.SourceRevision), Action: mission.ActionPublish, Targets: []string{target}, HeadSHA: head, Now: time.Now()}); err != nil {
		t.Fatalf("preflight governance: %v", err)
	}
	completionPath := filepath.Join(root, ".moai", "state", "mission", "governance", "completion.json")
	superviseArgs := []string{"run", "--supervise", "--session", session, "--target", target, "--governor-receipt", governorPattern, "--audit-receipt", auditPattern, "--completion-receipt", completionPath, "--lane", "worker-10", "--run-id", "auto-run-1"}
	if err := run(superviseArgs...); err == nil {
		t.Fatal("supervisor completed without sealed completion receipt")
	}
	blocked, _ := mission.LoadAutoMission(root, session)
	if blocked.State != mission.StateBlocked || len(blocked.OperationIDs) != 3 || blocked.Snapshot == nil {
		t.Fatalf("completion block lost reconciled work: %+v", blocked)
	}
	completion := mission.CompletionReceipt{Version: 1, MissionID: session, ContractHash: blocked.ContractHash, SnapshotHash: blocked.Snapshot.SnapshotHash, HeadSHA: head, Issuer: "completion-auditor", Status: "passed", Evidence: map[string]bool{"assignment": true}, ExpiresAt: time.Now().Add(time.Hour)}
	if err := mission.WriteCompletionReceipt(root, completionPath, completion); err != nil {
		t.Fatal(err)
	}
	if err := run("resume", "--session", session); err != nil {
		t.Fatal(err)
	}
	if err := run(superviseArgs...); err != nil {
		failed, _ := mission.LoadAutoMission(root, session)
		t.Fatalf("%v state=%+v", err, failed)
	}
	record, err := store.LoadPure()
	if err != nil || len(record.Items) != 1 || record.Items[0].State != kanban.BacklogStatePicked || len(record.Runtime.Assignments) != 1 {
		t.Fatalf("record=%+v err=%v", record, err)
	}
	assignment := record.Runtime.Assignments[0]
	if assignment.CardID != record.Items[0].ID || assignment.OwnerLabel != "worker-10" || assignment.RunID != "auto-run-1" {
		t.Fatalf("assignment=%+v card=%+v", assignment, record.Items[0])
	}
}

func TestAutoMissionSupervisorSplitWorktreesCommitMergeAndCompletion(t *testing.T) {
	root, _ := todoFixture(t)
	git := func(dir string, args ...string) string { return gitFixtureCLI(t, dir, args...) }
	git(root, "config", "user.name", "Test")
	git(root, "config", "user.email", "test@example.invalid")
	develop := filepath.Join(root, ".claude", "worktrees", "develop")
	card := filepath.Join(root, ".claude", "worktrees", "card")
	if err := os.MkdirAll(filepath.Dir(develop), 0o700); err != nil {
		t.Fatal(err)
	}
	git(root, "worktree", "add", "-q", "-b", "develop", develop, "HEAD")
	git(root, "worktree", "add", "-q", "-b", "WT-card", card, "HEAD")
	if err := os.WriteFile(filepath.Join(card, "owned.txt"), []byte("owned"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(card, "foreign.txt"), []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(develop, "develop-foreign.txt"), []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}

	session := "018f4f4a-7b7c-7a11-8f4d-c33333333333"
	target := "repo:" + root
	run := func(args ...string) error {
		cmd := newGoalCmd()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	if err := run("--auto", "--session", session, "commit then merge"); err != nil {
		t.Fatal(err)
	}
	if err := run("approve", "--session", session, "--scope", target, "--action", "commit", "--action", "local_develop_merge", "--completion-evidence", "tests", "--completion-evidence", "landed", "--max-operations", "2"); err != nil {
		t.Fatal(err)
	}
	if err := run("run", "--supervise", "--session", session, "--target", target, "--repo", root); err == nil || !strings.Contains(err.Error(), "split_worktree_required") {
		t.Fatalf("single repo supervised topology accepted: %v", err)
	}
	state, _ := mission.LoadAutoMission(root, session)
	governorPattern := filepath.Join(root, ".moai", "state", "mission", "governance", "{action}-decision.json")
	auditPattern := filepath.Join(root, ".moai", "state", "mission", "governance", "{action}-audit.json")
	writeGovernance := func(action mission.Action, snapshot, head string) {
		base := mission.GovernanceReceipt{Version: 1, MissionID: session, ContractHash: state.ContractHash, SnapshotHash: snapshot, Action: action, Targets: []string{target}, ExpiresAt: time.Now().Add(time.Hour), HeadSHA: head}
		decision := base
		decision.Kind, decision.Issuer, decision.Status = mission.GovernanceDecision, "mission-governor", mission.GovernanceRecommended
		audit := base
		audit.Kind, audit.Issuer, audit.Status = mission.GovernanceAudit, "sync-auditor", mission.GovernancePassed
		if err := mission.WriteGovernanceReceipt(root, strings.ReplaceAll(governorPattern, "{action}", string(action)), decision); err != nil {
			t.Fatal(err)
		}
		if err := mission.WriteGovernanceReceipt(root, strings.ReplaceAll(auditPattern, "{action}", string(action)), audit); err != nil {
			t.Fatal(err)
		}
	}
	cardBase := git(card, "rev-parse", "HEAD")
	developBase := git(develop, "rev-parse", "HEAD")
	testsReceipt := filepath.Join(card, ".moai", "reports", "tests.json")
	if err := os.MkdirAll(filepath.Dir(testsReceipt), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testsReceipt, []byte(`{"head_sha":"`+cardBase+`","status":"passed"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	writeGovernance(mission.ActionCommit, autoGitSnapshotHash(state.ContractHash, target, mission.ActionCommit, card, cardBase, ""), cardBase)
	lease := filepath.Join(root, ".git", "integration-lease.json")
	if err := mission.WriteIntegrationLease(lease, mission.IntegrationLease{SessionID: session, BaseSHA: developBase}); err != nil {
		t.Fatal(err)
	}
	completionPath := filepath.Join(root, ".moai", "state", "mission", "governance", "completion-split.json")
	args := []string{"run", "--supervise", "--session", session, "--target", target, "--card-worktree", card, "--develop-worktree", develop, "--worktree-branch", "WT-card", "--path", "owned.txt", "--message", "feat: split", "--base-sha", developBase, "--integration-lease", lease, "--tests-receipt", testsReceipt, "--governor-receipt", governorPattern, "--audit-receipt", auditPattern, "--completion-receipt", completionPath}
	if err := run(args...); err == nil {
		t.Fatal("missing merge governance did not block")
	}
	cardSHA := git(card, "rev-parse", "HEAD")
	if cardSHA == cardBase {
		t.Fatal("card worktree was not committed")
	}
	if got := git(card, "status", "--porcelain"); !strings.Contains(got, "?? foreign.txt") {
		t.Fatalf("card foreign state lost: %q", got)
	}
	writeGovernance(mission.ActionLocalMerge, autoGitSnapshotHash(state.ContractHash, target, mission.ActionLocalMerge, develop, developBase, cardSHA), developBase)
	if err := run("resume", "--session", session); err != nil {
		t.Fatal(err)
	}
	args = append(args, "--card-sha", cardSHA)
	if err := run(args...); err == nil {
		t.Fatal("missing completion receipt did not block")
	}
	blocked, _ := mission.LoadAutoMission(root, session)
	developHead := git(develop, "rev-parse", "HEAD")
	if parents := strings.Fields(git(develop, "show", "-s", "--format=%P", "HEAD")); len(parents) != 2 {
		t.Fatalf("not no-ff: %v", parents)
	}
	if got := git(develop, "status", "--porcelain"); !strings.Contains(got, "?? develop-foreign.txt") {
		t.Fatalf("develop foreign state lost: %q", got)
	}
	if err := mission.WriteCompletionReceipt(root, completionPath, mission.CompletionReceipt{Version: 1, MissionID: session, ContractHash: state.ContractHash, SnapshotHash: blocked.Snapshot.SnapshotHash, HeadSHA: developHead, Issuer: "completion-auditor", Status: "passed", Evidence: map[string]bool{"tests": true, "landed": true}, LandedAncestry: true, ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := run("resume", "--session", session); err != nil {
		t.Fatal(err)
	}
	if err := run(args...); err != nil {
		t.Fatal(err)
	}
	done, _ := mission.LoadAutoMission(root, session)
	if done.State != mission.StateCompleted || len(done.OperationIDs) != 2 {
		t.Fatalf("state=%+v", done)
	}
}

func TestAutoMissionExternalDeliveryWithoutAuthoritativeProviderIsStableBlocked(t *testing.T) {
	root, store := todoFixture(t)
	session := "018f4f4a-7b7c-7a11-8f4d-f22222222222"
	target := "repo:" + root
	run := func(args ...string) error {
		cmd := newGoalCmd()
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)
		return cmd.Execute()
	}
	if err := run("--auto", "--session", session, "release approved develop"); err != nil {
		t.Fatal(err)
	}
	if err := run("approve", "--session", session, "--scope", target, "--action", "batch_push", "--completion-evidence", "origin_develop"); err != nil {
		t.Fatal(err)
	}
	err := run("run", "--session", session, "--action", "batch_push", "--target", target)
	if err == nil || !strings.Contains(err.Error(), "provider_unsupported") {
		t.Fatalf("unsupported provider err=%v", err)
	}
	state, loadErr := mission.LoadAutoMission(root, session)
	if loadErr != nil || state.State != mission.StateBlocked || !strings.Contains(state.LastBlocker, "provider_unsupported") {
		t.Fatalf("state=%+v err=%v", state, loadErr)
	}
	raw, exportErr := kanban.ExportGTD(context.Background(), store, true)
	if exportErr != nil {
		t.Fatal(exportErr)
	}
	if strings.Contains(string(raw), `"action":"batch_push"`) {
		t.Fatalf("unsupported delivery prepared an operation: %s", raw)
	}
}

func TestAuthoritativeDispatchEvidenceRefusalMatrix(t *testing.T) {
	root, store := todoFixture(t)
	ctx := context.Background()
	session := "018f4f4a-7b7c-7a11-8f4d-d44444444444"
	item, err := kanban.CaptureGTDItem(ctx, store, kanban.CaptureInput{Content: "dispatch evidence", Source: "user", SourceAllowed: true, Sensitivity: kanban.SensitivityPrivate, EventID: "dispatch-evidence"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.ClarifyGTDItem(ctx, store, kanban.ClarifyInput{ItemID: item.ItemID, Disposition: kanban.DispositionAction, DesiredOutcome: "done", CompletionEvidence: "assignment", Authority: "dispatch", SourceTrusted: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.OrganizeGTDItem(ctx, store, kanban.OrganizeInput{ItemID: item.ItemID, Class: kanban.ClassAction}); err != nil {
		t.Fatal(err)
	}
	engaged, err := kanban.EngageGTDItem(ctx, store, kanban.EngageInput{ItemID: item.ItemID, Authorized: true, EvidenceFresh: true, DependenciesReady: true, LaneAvailable: true, ResourcesAvailable: true})
	if err != nil || engaged.CardID == "" {
		t.Fatalf("engage=%+v err=%v", engaged, err)
	}
	if err := store.Mutate(func(record *kanban.BacklogRecord) error {
		for i := range record.Items {
			if record.Items[i].ID == engaged.CardID {
				record.Items[i].State = kanban.BacklogStatePicked
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	current, err := kanban.LoadGTDItem(ctx, store, item.ItemID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := authoritativeDispatchEvidence(ctx, store, root, session, item.ItemID, engaged.CardID, "worker-10", "run-1", current.SourceRevision); err == nil {
		t.Fatal("missing lease accepted")
	}
	if _, err := kanban.AcquireSlotLease(root, kanban.SlotLeaseRequest{Resource: "worker-10", SessionID: session, MaxDuration: time.Hour}); err != nil {
		t.Fatal(err)
	}
	if values, err := authoritativeDispatchEvidence(ctx, store, root, session, item.ItemID, engaged.CardID, "worker-10", "run-1", current.SourceRevision); err != nil || values["picked"] != "true" {
		t.Fatalf("values=%v err=%v", values, err)
	}
	for name, args := range map[string]struct {
		item, card, lane, run string
		rev                   int64
	}{
		"stale":        {item.ItemID, engaged.CardID, "worker-10", "run-1", current.SourceRevision + 1},
		"wrong-card":   {item.ItemID, "t999", "worker-10", "run-1", current.SourceRevision},
		"empty-card":   {item.ItemID, "", "worker-10", "run-1", current.SourceRevision},
		"empty-lane":   {item.ItemID, engaged.CardID, "", "run-1", current.SourceRevision},
		"empty-run":    {item.ItemID, engaged.CardID, "worker-10", "", current.SourceRevision},
		"missing-item": {"gtd-missing", engaged.CardID, "worker-10", "run-1", current.SourceRevision},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := authoritativeDispatchEvidence(ctx, store, root, session, args.item, args.card, args.lane, args.run, args.rev); err == nil {
				t.Fatal("invalid evidence accepted")
			}
		})
	}
	if err := kanban.RecordFactoryCardAssignment(root, "foreign-run", engaged.CardID, "worker-10", ""); err != nil {
		t.Fatal(err)
	}
	if values, err := authoritativeDispatchEvidence(ctx, store, root, session, item.ItemID, engaged.CardID, "worker-10", "run-1", current.SourceRevision); err == nil || values["lane_owner_free"] != "false" {
		t.Fatalf("conflict values=%v err=%v", values, err)
	}
}

func TestGoalMissionSupervisorArgumentRefusals(t *testing.T) {
	root, _ := todoFixture(t)
	cmd := newGoalCmd()
	cmd.SetContext(context.Background())
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := runGoalMissionSupervisor(cmd, "bad", false, "gtd:x", missionGitRunOptions{}); err == nil {
		t.Fatal("invalid session accepted")
	}
	draftID := "018f4f4a-7b7c-7a11-8f4d-e55555555555"
	if err := mission.SaveAutoMission(root, mission.AutoMission{SessionID: draftID, Text: "draft", MissionMode: mission.ModeAuto, State: mission.StateDraft}); err != nil {
		t.Fatal(err)
	}
	if err := runGoalMissionSupervisor(cmd, draftID, false, "gtd:x", missionGitRunOptions{}); err == nil || !strings.Contains(err.Error(), "sealed contract") {
		t.Fatalf("draft err=%v", err)
	}
	run := func(args ...string) error {
		c := newGoalCmd()
		c.SetOut(io.Discard)
		c.SetErr(io.Discard)
		c.SetArgs(args)
		return c.Execute()
	}
	noTarget := "018f4f4a-7b7c-7a11-8f4d-e66666666666"
	if err := run("--auto", "--session", noTarget, "no target"); err != nil {
		t.Fatal(err)
	}
	if err := run("approve", "--session", noTarget, "--scope", "repo:"+root, "--action", "publish", "--completion-evidence", "done"); err != nil {
		t.Fatal(err)
	}
	if err := runGoalMissionSupervisor(cmd, noTarget, false, "", missionGitRunOptions{}); err == nil || !strings.Contains(err.Error(), "target required") {
		t.Fatalf("target err=%v", err)
	}
	autoTarget := "018f4f4a-7b7c-7a11-8f4d-e77777777777"
	if err := run("--auto", "--session", autoTarget, "auto target"); err != nil {
		t.Fatal(err)
	}
	if err := run("approve", "--session", autoTarget, "--scope", "gtd:missing", "--action", "publish", "--completion-evidence", "done"); err != nil {
		t.Fatal(err)
	}
	if err := runGoalMissionSupervisor(cmd, autoTarget, false, "", missionGitRunOptions{}); err == nil {
		t.Fatal("missing GTD snapshot accepted")
	}
}

func TestGoalMissionOperationEarlyRefusalBranches(t *testing.T) {
	root, store := todoFixture(t)
	cmd := newGoalCmd()
	cmd.SetContext(context.Background())
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := runGoalMissionOperation(cmd, "bad", false, "publish", "gtd:x", false, missionGitRunOptions{}); err == nil {
		t.Fatal("invalid session accepted")
	}
	draftID := "018f4f4a-7b7c-7a11-8f4d-f55555555555"
	if err := mission.SaveAutoMission(root, mission.AutoMission{SessionID: draftID, Text: "draft", MissionMode: mission.ModeAuto, State: mission.StateDraft}); err != nil {
		t.Fatal(err)
	}
	if err := runGoalMissionOperation(cmd, draftID, false, "publish", "gtd:x", false, missionGitRunOptions{}); err == nil || !strings.Contains(err.Error(), "not approved") {
		t.Fatalf("draft err=%v", err)
	}
	missingContractID := "018f4f4a-7b7c-7a11-8f4d-f56565656565"
	if err := mission.SaveAutoMission(root, mission.AutoMission{SessionID: missingContractID, Text: "approved without contract", MissionMode: mission.ModeAuto, State: mission.StateApproved}); err != nil {
		t.Fatal(err)
	}
	if err := runGoalMissionOperation(cmd, missingContractID, false, "publish", "gtd:x", false, missionGitRunOptions{}); err == nil || !strings.Contains(err.Error(), "contract missing") {
		t.Fatalf("missing contract err=%v", err)
	}
	run := func(args ...string) error {
		c := newGoalCmd()
		c.SetOut(io.Discard)
		c.SetErr(io.Discard)
		c.SetArgs(args)
		return c.Execute()
	}
	approve := func(id string, action mission.Action, target string) {
		if err := run("--auto", "--session", id, "refusal"); err != nil {
			t.Fatal(err)
		}
		if err := run("approve", "--session", id, "--scope", target, "--action", string(action), "--completion-evidence", "done"); err != nil {
			t.Fatal(err)
		}
	}
	missingID := "018f4f4a-7b7c-7a11-8f4d-f66666666666"
	approve(missingID, mission.ActionPublish, "gtd:missing")
	if err := runGoalMissionOperation(cmd, missingID, false, "publish", "gtd:missing", false, missionGitRunOptions{}); err == nil {
		t.Fatal("missing item accepted")
	}
	item, err := kanban.CaptureGTDItem(context.Background(), store, kanban.CaptureInput{Content: "not published", Source: "user", SourceAllowed: true, Sensitivity: kanban.SensitivityPrivate, EventID: "op-refusal"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.ClarifyGTDItem(context.Background(), store, kanban.ClarifyInput{ItemID: item.ItemID, Disposition: kanban.DispositionAction, DesiredOutcome: "done", CompletionEvidence: "done", Authority: "queue", SourceTrusted: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := kanban.OrganizeGTDItem(context.Background(), store, kanban.OrganizeInput{ItemID: item.ItemID, Class: kanban.ClassAction}); err != nil {
		t.Fatal(err)
	}
	target := "gtd:" + item.ItemID
	pickID := "018f4f4a-7b7c-7a11-8f4d-f77777777777"
	approve(pickID, mission.ActionPick, target)
	if err := runGoalMissionOperation(cmd, pickID, false, "pick", target, false, missionGitRunOptions{}); err == nil || !strings.Contains(err.Error(), "published_card_missing") {
		t.Fatalf("pick err=%v", err)
	}
	engaged, err := kanban.EngageGTDItem(context.Background(), store, kanban.EngageInput{ItemID: item.ItemID, Authorized: true, EvidenceFresh: true, DependenciesReady: true, LaneAvailable: true, ResourcesAvailable: true})
	if err != nil {
		t.Fatal(err)
	}
	if engaged.CardID == "" {
		t.Fatal("card missing")
	}
	dispatchID := "018f4f4a-7b7c-7a11-8f4d-f88888888888"
	approve(dispatchID, mission.ActionDispatch, target)
	if err := runGoalMissionOperation(cmd, dispatchID, false, "dispatch", target, false, missionGitRunOptions{}); err == nil || !strings.Contains(err.Error(), "dispatch_input_missing") {
		t.Fatalf("dispatch err=%v", err)
	}
	commitID := "018f4f4a-7b7c-7a11-8f4d-f99999999999"
	approve(commitID, mission.ActionCommit, "repo:"+root)
	if err := runGoalMissionOperation(cmd, commitID, false, "commit", "repo:"+root, false, missionGitRunOptions{Repository: filepath.Join(root, "missing"), Branch: "WT-x", Paths: []string{"x"}}); err == nil {
		t.Fatal("missing repo accepted")
	}
	if err := runGoalMissionOperation(cmd, commitID, false, "commit", "repo:"+root, false, missionGitRunOptions{Repository: root, Branch: "WT-x", Paths: []string{"x"}}); err == nil || !strings.Contains(err.Error(), "git_scope_invalid") {
		t.Fatalf("scope err=%v", err)
	}
	mergeID := "018f4f4a-7b7c-7a11-8f4d-faaaaaaaaaaa"
	approve(mergeID, mission.ActionLocalMerge, "repo:"+root)
	if err := runGoalMissionOperation(cmd, mergeID, false, "local_develop_merge", "repo:"+root, false, missionGitRunOptions{Repository: root, CardWorktree: root, CardSHA: "wrong"}); err == nil || !strings.Contains(err.Error(), "card_sha_mismatch") {
		t.Fatalf("card err=%v", err)
	}
}
