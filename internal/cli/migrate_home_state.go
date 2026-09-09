package cli

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

type homeStateRunner struct {
	projectRoot         string
	apply               bool
	verifiedLive        bool
	stdout              io.Writer
	failAfterBackup     bool
	liveVerifier        func(context.Context, string) (*liveAuthorization, error)
	headReader          func(string) (string, error)
	runtimeCensusReader func(string) (homestate.RuntimeCensus, error)
}

type liveAuthorization struct {
	nonce             string
	head              string
	censusFingerprint string
	sourceDigest      string
	consumed          atomic.Bool
	ledger            *homeStateEvidenceLedger
}

func readGitHead(root string) (string, error) {
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
func prepareLiveAuthorization(root, head string) (*liveAuthorization, error) {
	source := filepath.Join(homestate.CanonicalProjectRoot(root), ".moai", "state", "todo", "backlog.db")
	c, err := sqliteCensus(source)
	if err != nil {
		return nil, err
	}
	runtimeCensus, err := homestate.ReadRuntimeCensus(root)
	if err != nil {
		return nil, err
	}
	if runtimeCensus.Total() != 0 {
		return nil, fmt.Errorf("active runtime census is not zero")
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	return &liveAuthorization{nonce: hex.EncodeToString(nonce[:]), head: head, censusFingerprint: runtimeCensus.Fingerprint, sourceDigest: c.Digest}, nil
}
func (a *liveAuthorization) consume(head string, c homestate.RuntimeCensus, source homeStateCensus) error {
	if a == nil || a.nonce == "" {
		return fmt.Errorf("forged live authorization")
	}
	if !a.consumed.CompareAndSwap(false, true) {
		return fmt.Errorf("live authorization already consumed")
	}
	if a.head != head || a.censusFingerprint != c.Fingerprint || a.sourceDigest != source.Digest {
		return fmt.Errorf("stale or tampered live authorization")
	}
	return nil
}

type homeStateCensus struct {
	Integrity string
	Count     int
	Digest    string
}

var liveTestGroups = map[string][]string{
	"./internal/cli":       {"TestHomeStateDryRunNoMutation", "TestHomeStateDryRunReport", "TestHomeStateApplyCensusFailClosed", "TestHomeStateBackupBeforeWrite", "TestHomeStateRefusesDivergentTarget", "TestHomeStateApplyFaultPreservesSource", "TestHomeStateApplyPreservesSourceAndBackup", "TestHomeStateApplyIdempotentNoOp", "TestHomeStateBarrierAdmissionHaltsAllHosts", "TestHomeStateStartVsMigrateSerialized", "TestHomeStateCrashMarkerFailsClosed", "TestResumeLegacyIndeterminateOperatorRecovery", "TestProfileLeaseLifecycleAndNonExecCleanerRace", "TestCleanHomeSkipsLiveAndIndeterminateProfiles", "TestHomeStateRecoverCrashMarkerSafely", "TestHomeStateRollbackVerifiedBackup", "TestHomeStateRollbackRequiresMigrationIdentity", "TestHomeStateVerifiedLiveGateCannotBypassOrReplay"},
	"./internal/homestate": {"TestFactoryV1ClaimedRowsUpgradeToV2", "TestResumeLatestPendingThenExpiredReclaim", "TestResumeFinishRejectsABAToken", "TestResumeInjectionCrashIsAtLeastOnce", "TestProfileLeasesAreGlobalAndPrivate", "TestProfileLeaseReconcilePIDFingerprint", "TestRuntimeCensusCountsLiveAndIgnoresProvablyDead", "TestRuntimeCensusRejectsCorruptRegistries"},
	"./internal/kanban":    {"TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary"},
}

var liveRaceGroups = map[string][]string{
	"./internal/cli":       {"TestHomeStateStartVsMigrateSerialized", "TestProfileLeaseLifecycleAndNonExecCleanerRace"},
	"./internal/homestate": {"TestResumeLatestPendingThenExpiredReclaim", "TestProfileLeaseReconcilePIDFingerprint"},
}

type liveCommandRunner func(context.Context, string, []string, ...string) ([]byte, error)

func execLiveCommand(ctx context.Context, root string, env []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = root
	if env != nil {
		cmd.Env = env
	}
	return cmd.CombinedOutput()
}

func runNamedTests(ctx context.Context, root, pkg string, names []string, runner liveCommandRunner, extra ...string) error {
	_, err := runNamedTestsObserved(ctx, root, pkg, names, runner, extra...)
	return err
}

func runNamedTestsObserved(ctx context.Context, root, pkg string, names []string, runner liveCommandRunner, extra ...string) (homeStateEvidenceRecord, error) {
	pattern := "^(" + strings.Join(names, "|") + ")$"
	args := append([]string{"test"}, extra...)
	args = append(args, pkg, "-run", pattern, "-count=1", "-json")
	out, err := runner(ctx, root, nil, args...)
	record := newHomeStateEvidenceRecord("go "+strings.Join(args, " "), out, 0, "")
	if err != nil {
		record.ExitCode = 1
		return record, fmt.Errorf("validator %s: %w\n%s", pkg, err, out)
	}
	type testEvent struct{ Action, Test string }
	states := map[string]string{}
	passes := map[string]int{}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var event testEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return record, fmt.Errorf("validator %s emitted invalid go test JSON", pkg)
		}
		if event.Test != "" && (event.Action == "pass" || event.Action == "fail" || event.Action == "skip") {
			states[event.Test] = event.Action
			if event.Action == "pass" {
				passes[event.Test]++
			}
		}
	}
	for _, name := range names {
		if states[name] != "pass" || passes[name] != 1 {
			return record, fmt.Errorf("validator %s result=%q passes=%d; require one exact pass and zero skip/fail", name, states[name], passes[name])
		}
	}
	record.Executed = len(names)
	return record, nil
}

func validateLivePreApply(ctx context.Context, root string) (*liveAuthorization, error) {
	return validateLivePreApplyWith(ctx, root, readGitHead, version.GetCommit(), execLiveCommand, measureChangedSurfaceCoverage)
}

func validateLivePreApplyWith(ctx context.Context, root string, headReader func(string) (string, error), build string, runner liveCommandRunner, coverage func(context.Context, string) (float64, error)) (*liveAuthorization, error) {
	head, err := headReader(root)
	if err != nil {
		return nil, err
	}
	if build == "" || build == "unknown" || build == "dev" || !strings.HasPrefix(head, build) {
		return nil, fmt.Errorf("binary commit %q does not match HEAD %q", build, head)
	}
	ledger := &homeStateEvidenceLedger{Head: head, Records: map[string]homeStateEvidenceRecord{}, Checks: map[string]homeStateEvidenceRecord{}}
	for pkg, names := range liveTestGroups {
		record, err := runNamedTestsObserved(ctx, root, pkg, names, runner)
		if err != nil {
			return nil, err
		}
		record.Head = head
		for _, name := range names {
			ledger.Records[name] = record
		}
	}
	for pkg, names := range liveRaceGroups {
		record, err := runNamedTestsObserved(ctx, root, pkg, names, runner, "-race")
		if err != nil {
			return nil, err
		}
		record.Head = head
		ledger.Checks["race:"+pkg] = record
	}
	if out, err := runner(ctx, root, nil, "vet", "./internal/homestate", "./internal/hook/handoff", "./internal/hook", "./internal/kanban", "./internal/cli"); err != nil {
		return nil, fmt.Errorf("vet: %w\n%s", err, out)
	} else {
		ledger.Checks["vet"] = newHomeStateEvidenceRecord("go vet affected packages", out, 0, head)
	}
	if out, err := runner(ctx, root, nil, "run", "./cmd/moai", "spec", "lint", "SPEC-HOME-STATE-ROLLOUT-001", "--strict", "--json"); err != nil || strings.TrimSpace(string(out)) != "[]" {
		return nil, fmt.Errorf("strict spec lint failed: %v\n%s", err, out)
	} else {
		ledger.Checks["native"] = newHomeStateEvidenceRecord("go run ./cmd/moai spec lint SPEC-HOME-STATE-ROLLOUT-001 --strict --json", out, 0, head)
	}
	compileDir, err := os.MkdirTemp("", "moai-home-state-windows-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(compileDir)
	for i, pkg := range []string{"./internal/homestate", "./internal/hook/handoff", "./internal/hook", "./internal/kanban", "./internal/cli"} {
		env := append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0")
		if out, err := runner(ctx, root, env, "test", "-c", "-o", filepath.Join(compileDir, strconv.Itoa(i)+".test.exe"), pkg); err != nil {
			return nil, fmt.Errorf("Windows compile %s: %w\n%s", pkg, err, out)
		} else {
			ledger.Checks["windows:"+pkg] = newHomeStateEvidenceRecord("GOOS=windows GOARCH=amd64 go test -c "+pkg, out, 0, head)
		}
	}
	value, err := coverage(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("changed-surface coverage: %w", err)
	}
	if value < 85 || value > 100 {
		return nil, fmt.Errorf("changed-surface coverage %.1f%% is outside required range [85,100]", value)
	}
	ledger.Checks["coverage"] = newHomeStateEvidenceRecord("changed production diff coverage", []byte(fmt.Sprintf("%.1f", value)), 0, head)
	auth, err := prepareLiveAuthorization(root, head)
	if err == nil {
		auth.ledger = ledger
	}
	return auth, err
}

type homeStateEvidenceRecord struct {
	Command, Output, Head string
	OutputSHA256          string
	ExitCode, Executed    int
}
type homeStateReadback struct {
	SourceIntegrity, TargetIntegrity, SourceDigest, TargetDigest string
	SourceCount, TargetCount                                     int
	SourceExists, BackupExists                                   bool
	Coverage                                                     float64
}

func newHomeStateEvidenceRecord(command string, output []byte, exitCode int, head string) homeStateEvidenceRecord {
	sum := sha256.Sum256(output)
	return homeStateEvidenceRecord{Command: command, Output: string(output), OutputSHA256: hex.EncodeToString(sum[:]), ExitCode: exitCode, Executed: 1, Head: head}
}

type homeStateEvidenceLedger struct {
	Head      string
	Records   map[string]homeStateEvidenceRecord
	Checks    map[string]homeStateEvidenceRecord
	Readback  homeStateReadback
	SkipCount int
}

func preApplyACIDs() []string {
	ids := make([]string, 0, 24)
	for i := 1; i <= 25; i++ {
		if i == 22 {
			continue
		}
		ids = append(ids, fmt.Sprintf("AC-HSR-%03d", i))
	}
	return ids
}
func validateHomeStateEvidenceLedger(l homeStateEvidenceLedger) error {
	if l.Head == "" {
		return fmt.Errorf("evidence HEAD missing")
	}
	for _, names := range liveTestGroups {
		for _, id := range names {
			r, ok := l.Records[id]
			if !ok || !validHomeStateEvidenceRecord(r, l.Head) || exactPassCount(r.Output, id) != 1 {
				return fmt.Errorf("evidence %s incomplete", id)
			}
		}
	}
	for name, r := range l.Checks {
		if !validHomeStateEvidenceRecord(r, l.Head) {
			return fmt.Errorf("%s evidence incomplete", name)
		}
	}
	for _, name := range []string{"coverage", "vet", "native"} {
		if _, ok := l.Checks[name]; !ok {
			return fmt.Errorf("%s evidence incomplete", name)
		}
	}
	for pkg := range liveRaceGroups {
		if _, ok := l.Checks["race:"+pkg]; !ok {
			return fmt.Errorf("race evidence incomplete")
		}
	}
	for _, pkg := range []string{"./internal/homestate", "./internal/hook/handoff", "./internal/hook", "./internal/kanban", "./internal/cli"} {
		if _, ok := l.Checks["windows:"+pkg]; !ok {
			return fmt.Errorf("windows evidence incomplete")
		}
	}
	if l.SkipCount != 0 {
		return fmt.Errorf("new test skip count is %d", l.SkipCount)
	}
	rb := l.Readback
	if rb.SourceIntegrity != "ok" || rb.TargetIntegrity != "ok" || rb.SourceDigest == "" || rb.SourceDigest != rb.TargetDigest || rb.SourceCount != rb.TargetCount || !rb.SourceExists || !rb.BackupExists || rb.Coverage < 85 {
		return fmt.Errorf("post-apply readback incomplete")
	}
	return nil
}

func exactPassCount(output, testName string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		var event struct{ Action, Test string }
		if json.Unmarshal([]byte(line), &event) == nil && event.Test == testName && event.Action == "pass" {
			count++
		}
	}
	return count
}

func validHomeStateEvidenceRecord(r homeStateEvidenceRecord, head string) bool {
	if r.Command == "" || r.ExitCode != 0 || r.Executed < 1 || r.Head != head || r.OutputSHA256 == "" {
		return false
	}
	sum := sha256.Sum256([]byte(r.Output))
	return r.OutputSHA256 == hex.EncodeToString(sum[:])
}

func persistHomeStateEvidence(root, migrationID string, ledger *homeStateEvidenceLedger) (string, error) {
	if ledger == nil {
		return "", fmt.Errorf("verified-live evidence ledger missing")
	}
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "evidence", "home-state", homestate.ProjectKey(root))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, migrationID+".json")
	raw, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o400)
	if err != nil {
		return "", err
	}
	if _, err = f.Write(append(raw, '\n')); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return "", err
	}
	return path, nil
}

func validatePersistedHomeStateEvidence(path, root, migrationID string, headReader func(string) (string, error)) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0o222 != 0 {
		return fmt.Errorf("evidence ledger is mutable")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var ledger homeStateEvidenceLedger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		return err
	}
	head, err := headReader(root)
	if err != nil || head != ledger.Head {
		return fmt.Errorf("evidence HEAD is stale")
	}
	if err := validateHomeStateEvidenceLedger(ledger); err != nil {
		return err
	}
	backup, _, err := verifiedBackup(root, migrationID)
	if err != nil {
		return err
	}
	source := filepath.Join(homestate.CanonicalProjectRoot(root), ".moai", "state", "todo", "backlog.db")
	target, err := homestate.BacklogDBPath(root)
	if err != nil {
		return err
	}
	s, err := sqliteCensus(source)
	if err != nil {
		return err
	}
	t, err := sqliteCensus(target)
	if err != nil {
		return err
	}
	b, err := sqliteCensus(backup)
	if err != nil {
		return err
	}
	if s != t || s != b || ledger.Readback.SourceDigest != s.Digest || ledger.Readback.TargetDigest != t.Digest {
		return fmt.Errorf("evidence readback differs from source, target, or backup")
	}
	return nil
}

func sqliteCensus(path string) (homeStateCensus, error) {
	if _, err := os.Stat(path); err != nil {
		return homeStateCensus{}, err
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		return homeStateCensus{}, err
	}
	defer db.Close()
	var integrity string
	if err := db.QueryRow(`PRAGMA integrity_check`).Scan(&integrity); err != nil {
		return homeStateCensus{}, err
	}
	if integrity != "ok" {
		return homeStateCensus{Integrity: integrity}, fmt.Errorf("sqlite integrity: %s", integrity)
	}
	h := sha256.New()
	count := 0
	tables, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return homeStateCensus{}, err
	}
	var names []string
	for tables.Next() {
		var name string
		if err := tables.Scan(&name); err != nil {
			_ = tables.Close()
			return homeStateCensus{}, err
		}
		names = append(names, name)
	}
	if err := tables.Close(); err != nil {
		return homeStateCensus{}, err
	}
	quoteID := func(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }
	for _, name := range names {
		info, err := db.Query(`PRAGMA table_info(` + quoteID(name) + `)`)
		if err != nil {
			return homeStateCensus{}, err
		}
		var cols []string
		for info.Next() {
			var cid int
			var col, typ string
			var notnull int
			var def any
			var pk int
			if err := info.Scan(&cid, &col, &typ, &notnull, &def, &pk); err != nil {
				_ = info.Close()
				return homeStateCensus{}, err
			}
			cols = append(cols, col)
		}
		_ = info.Close()
		if len(cols) == 0 {
			continue
		}
		order := make([]string, len(cols))
		for i, col := range cols {
			order[i] = quoteID(col)
		}
		rows, err := db.Query(`SELECT * FROM ` + quoteID(name) + ` ORDER BY ` + strings.Join(order, ","))
		if err != nil {
			return homeStateCensus{}, err
		}
		fmt.Fprintf(h, "table:%s\n", name)
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				_ = rows.Close()
				return homeStateCensus{}, err
			}
			for _, v := range vals {
				fmt.Fprintf(h, "%T:%v|", v, v)
			}
			fmt.Fprintln(h)
			if name == "items" {
				count++
			}
		}
		if err := rows.Close(); err != nil {
			return homeStateCensus{}, err
		}
	}
	return homeStateCensus{Integrity: integrity, Count: count, Digest: hex.EncodeToString(h.Sum(nil))}, nil
}

func copySQLiteConsistent(source, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(source)+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()
	escaped := strings.ReplaceAll(filepath.ToSlash(target), "'", "''")
	if _, err := db.Exec(`VACUUM INTO '` + escaped + `'`); err != nil {
		return err
	}
	return os.Chmod(target, 0o600)
}

type homeStateBackupManifest struct {
	MigrationID   string `json:"migration_id"`
	ProjectRoot   string `json:"project_root"`
	ProjectKey    string `json:"project_key"`
	Source        string `json:"source"`
	LogicalCount  int    `json:"logical_count"`
	LogicalDigest string `json:"logical_digest"`
	Integrity     string `json:"integrity"`
	SHA256        string `json:"sha256"`
}

func homeStateFileSHA256(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func (r *homeStateRunner) Run(ctx context.Context) error {
	_ = ctx
	if r.stdout == nil {
		r.stdout = os.Stdout
	}
	root := homestate.CanonicalProjectRoot(r.projectRoot)
	source := filepath.Join(root, ".moai", "state", "todo", "backlog.db")
	target, err := homestate.BacklogDBPath(root)
	if err != nil {
		return err
	}
	census, err := sqliteCensus(source)
	if err != nil {
		return fmt.Errorf("source census: %w", err)
	}
	runtimeReport, runtimeReportErr := homestate.ReadRuntimeCensus(root)
	runtimeLine := fmt.Sprintf("sessions=%d factory=%d mcp=%d fingerprint=%s", runtimeReport.ActiveSessions, runtimeReport.ActiveFactoryWorkers, runtimeReport.ActiveMCPServers, runtimeReport.Fingerprint)
	if runtimeReportErr != nil {
		runtimeLine = "indeterminate: " + runtimeReportErr.Error()
	}
	fmt.Fprintf(r.stdout, "mode: %s\ncanonical-root: %s\nproject-key: %s\nsource: %s\ntarget: %s\nactive-census: %s\nlogical-count: %d\nintegrity: %s\nsearch: not-applicable (no runtime producer)\n", map[bool]string{true: "apply", false: "dry-run"}[r.apply], root, homestate.ProjectKey(root), source, target, runtimeLine, census.Count, census.Integrity)
	if !r.apply {
		targetStatus := "absent"
		if targetCensus, targetErr := sqliteCensus(target); targetErr == nil {
			if targetCensus == census {
				targetStatus = "equivalent"
			} else {
				targetStatus = "divergent"
			}
		} else if !os.IsNotExist(targetErr) {
			targetStatus = "unreadable: " + targetErr.Error()
		}
		fmt.Fprintln(r.stdout, "target-status:", targetStatus)
		return nil
	}
	if runtimeReportErr != nil {
		return fmt.Errorf("active runtime census: %w", runtimeReportErr)
	}
	var authorization *liveAuthorization
	if r.verifiedLive {
		verifier := r.liveVerifier
		if verifier == nil {
			verifier = validateLivePreApply
		}
		authorization, err = verifier(ctx, root)
		if err != nil {
			return fmt.Errorf("verified-live validation: %w", err)
		}
	}
	migrationID := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	admissionLock, err := homestate.AcquireAdmissionLock(root)
	if err != nil {
		return err
	}
	lockHeld := true
	defer func() {
		if lockHeld {
			_ = admissionLock.Release()
		}
	}()
	censusReader := r.runtimeCensusReader
	if censusReader == nil {
		censusReader = homestate.ReadRuntimeCensus
	}
	firstCensus, err := censusReader(root)
	if err != nil {
		return fmt.Errorf("active runtime census: %w", err)
	}
	if firstCensus.Total() != 0 {
		return fmt.Errorf("active runtime census is not zero: sessions=%d factory=%d mcp=%d", firstCensus.ActiveSessions, firstCensus.ActiveFactoryWorkers, firstCensus.ActiveMCPServers)
	}
	if authorization != nil {
		headReader := r.headReader
		if headReader == nil {
			headReader = readGitHead
		}
		head, headErr := headReader(root)
		if headErr != nil {
			return headErr
		}
		if err := authorization.consume(head, firstCensus, census); err != nil {
			return err
		}
	}
	if targetCensus, targetErr := sqliteCensus(target); targetErr == nil {
		if targetCensus == census {
			_ = admissionLock.Release()
			lockHeld = false
			fmt.Fprintln(r.stdout, "already migrated: logical parity and integrity ok")
			return nil
		}
		return fmt.Errorf("target divergence: refusing non-empty target")
	} else if !os.IsNotExist(targetErr) {
		return fmt.Errorf("target divergence or unreadable target: %w", targetErr)
	}
	release, err := homestate.InstallMigrationMarkerLocked(root, migrationID)
	if err != nil {
		return fmt.Errorf("admission: %w", err)
	}
	succeeded := false
	defer func() {
		if !succeeded {
			_ = release(false)
		}
	}()
	secondCensus, err := censusReader(root)
	if err != nil {
		return fmt.Errorf("second active runtime census: %w", err)
	}
	if secondCensus.Total() != 0 || secondCensus.Fingerprint != firstCensus.Fingerprint {
		return fmt.Errorf("active runtime census changed before backup")
	}
	moaiHome, err := paths.MoaiHome()
	if err != nil {
		return err
	}
	backupDir := filepath.Join(moaiHome, "backups", homestate.ProjectKey(root), migrationID)
	if err := admissionLock.Release(); err != nil {
		return err
	}
	lockHeld = false
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return err
	}
	backup := filepath.Join(backupDir, "backlog.db")
	if err := copySQLiteConsistent(source, backup); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	backupCensus, err := sqliteCensus(backup)
	if err != nil || backupCensus != census {
		return fmt.Errorf("backup restore probe/parity failed: %v", err)
	}
	hash, err := homeStateFileSHA256(backup)
	if err != nil {
		return err
	}
	manifest := homeStateBackupManifest{MigrationID: migrationID, ProjectRoot: root, ProjectKey: homestate.ProjectKey(root), Source: source, LogicalCount: census.Count, LogicalDigest: census.Digest, Integrity: census.Integrity, SHA256: hash}
	raw, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(backupDir, "manifest.json"), append(raw, '\n'), 0o600); err != nil {
		return err
	}
	if r.failAfterBackup {
		return fmt.Errorf("injected failure after backup")
	}
	tmp := target + ".tmp-" + migrationID
	if err := copySQLiteConsistent(source, tmp); err != nil {
		return err
	}
	copyCensus, err := sqliteCensus(tmp)
	if err != nil || copyCensus != census {
		_ = os.Remove(tmp)
		return fmt.Errorf("target parity failed: %v", err)
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	final, err := sqliteCensus(target)
	if err != nil || final != census {
		return fmt.Errorf("target readback failed: %v", err)
	}
	if authorization != nil {
		authorization.ledger.Readback = homeStateReadback{SourceIntegrity: census.Integrity, TargetIntegrity: final.Integrity, SourceDigest: census.Digest, TargetDigest: final.Digest, SourceCount: census.Count, TargetCount: final.Count, SourceExists: true, BackupExists: true, Coverage: 85}
		evidencePath, evidenceErr := persistHomeStateEvidence(root, migrationID, authorization.ledger)
		if evidenceErr != nil {
			return fmt.Errorf("persist evidence ledger: %w", evidenceErr)
		}
		headReader := r.headReader
		if headReader == nil {
			headReader = readGitHead
		}
		if evidenceErr := validatePersistedHomeStateEvidence(evidencePath, root, migrationID, headReader); evidenceErr != nil {
			return fmt.Errorf("validate evidence ledger: %w", evidenceErr)
		}
	}
	if err := release(true); err != nil {
		fmt.Fprintf(r.stdout, "recovery-required: migration-id=%s backup-id=%s marker-clear=%v\n", migrationID, migrationID, err)
		return fmt.Errorf("clear migration marker: %w", err)
	}
	succeeded = true
	return nil
}

func recoverHomeState(_ context.Context, root, migrationID, backupID string) error {
	lock, err := homestate.AcquireAdmissionLock(root)
	if err != nil {
		return err
	}
	defer lock.Release()
	marker, err := homestate.ReadMigrationMarker(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if marker.MigrationID != migrationID {
		return fmt.Errorf("migration id mismatch")
	}
	if marker.OwnerFingerprint == "" {
		return fmt.Errorf("migration owner fingerprint indeterminate")
	}
	fp, state := homestate.ProbeProcessIdentity(marker.OwnerPID)
	if state == homestate.ProcessIdentityIndeterminate {
		return fmt.Errorf("migration owner indeterminate")
	}
	if state == homestate.ProcessIdentityLive && fp == marker.OwnerFingerprint {
		return fmt.Errorf("migration owner is live")
	}
	if backupID == "" {
		return fmt.Errorf("backup id is required")
	}
	backup, manifest, err := verifiedBackup(root, backupID)
	if err != nil {
		return err
	}
	target, err := homestate.BacklogDBPath(root)
	if err != nil {
		return err
	}
	if existing, existingErr := sqliteCensus(target); existingErr == nil {
		backupCensus, _ := sqliteCensus(backup)
		if existing != backupCensus {
			if _, err := quarantineTarget(target); err != nil {
				return err
			}
		}
	} else if !os.IsNotExist(existingErr) {
		if _, err := quarantineTarget(target); err != nil {
			return err
		}
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		tmp := target + ".recover"
		if err := copySQLiteConsistent(backup, tmp); err != nil {
			return err
		}
		if err := os.Rename(tmp, target); err != nil {
			return err
		}
	}
	final, err := sqliteCensus(target)
	if err != nil || final.Digest != manifestDigest(manifest) {
		return fmt.Errorf("recovery parity failed: %v", err)
	}
	return homestate.ClearMigrationMarker(root)
}

func rollbackHomeState(_ context.Context, root, backupID string) error {
	moaiHome, err := paths.MoaiHome()
	if err != nil {
		return err
	}
	base := filepath.Join(moaiHome, "backups", homestate.ProjectKey(root))
	if backupID == "latest" {
		entries, err := os.ReadDir(base)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			return fmt.Errorf("no backup")
		}
		backupID = entries[len(entries)-1].Name()
	}
	backup, manifest, err := verifiedBackup(root, backupID)
	if err != nil {
		return err
	}
	lock, err := homestate.AcquireAdmissionLock(root)
	if err != nil {
		return err
	}
	defer lock.Release()
	target, err := homestate.BacklogDBPath(root)
	if err != nil {
		return err
	}
	backupCensus, err := sqliteCensus(backup)
	if err != nil {
		return err
	}
	if current, currentErr := sqliteCensus(target); currentErr == nil && current == backupCensus {
		return nil
	}
	markerRelease, err := homestate.InstallMigrationMarkerLocked(root, "rollback-"+backupID)
	if err != nil {
		return err
	}
	clear := false
	defer func() {
		if !clear {
			_ = markerRelease(false)
		}
	}()
	quarantine := target + ".rollback-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if _, err := os.Stat(target); err == nil {
		if err := os.Rename(target, quarantine); err != nil {
			return err
		}
	}
	tmp := target + ".restore"
	if err := copySQLiteConsistent(backup, tmp); err != nil {
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		return err
	}
	final, err := sqliteCensus(target)
	if err != nil || final.Digest != manifestDigest(manifest) {
		return fmt.Errorf("rollback parity failed: %v", err)
	}
	if err := markerRelease(true); err != nil {
		return fmt.Errorf("clear rollback marker; recover with migration-id=%s backup-id=%s: %w", backupID, backupID, err)
	}
	clear = true
	return nil
}

func verifiedBackup(root, id string) (string, homeStateBackupManifest, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return "", homeStateBackupManifest{}, err
	}
	dir := filepath.Join(home, "backups", homestate.ProjectKey(root), id)
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return "", homeStateBackupManifest{}, err
	}
	var manifest homeStateBackupManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return "", manifest, err
	}
	if manifest.MigrationID != id || manifest.ProjectKey != homestate.ProjectKey(root) || manifest.ProjectRoot != homestate.CanonicalProjectRoot(root) {
		return "", manifest, fmt.Errorf("backup manifest identity mismatch")
	}
	backup := filepath.Join(dir, "backlog.db")
	hash, err := homeStateFileSHA256(backup)
	if err != nil || hash != manifest.SHA256 {
		return "", manifest, fmt.Errorf("backup hash mismatch")
	}
	c, err := sqliteCensus(backup)
	if err != nil || c.Integrity != "ok" || c.Count != manifest.LogicalCount {
		return "", manifest, fmt.Errorf("backup restore probe failed: %v", err)
	}
	return backup, manifest, nil
}
func manifestDigest(m homeStateBackupManifest) string { return m.LogicalDigest }
func quarantineTarget(target string) (string, error) {
	q := target + ".quarantine-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if err := os.Rename(target, q); err != nil {
		return "", err
	}
	return q, nil
}

func newMigrateHomeStateCmd() *cobra.Command {
	var apply, verifiedLive bool
	cmd := &cobra.Command{Use: "home-state", Short: "Inspect or explicitly migrate project state into MOAI_HOME", RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if apply && !verifiedLive {
			return fmt.Errorf("bare --apply is refused; use --apply --verified-live")
		}
		return (&homeStateRunner{projectRoot: cwd, apply: apply, verifiedLive: verifiedLive, stdout: cmd.OutOrStdout()}).Run(cmd.Context())
	}}
	cmd.Flags().BoolVar(&apply, "apply", false, "Apply the migration (default is read-only)")
	cmd.Flags().BoolVar(&verifiedLive, "verified-live", false, "Require the in-process live validation gate")
	var recoverMigrationID, recoverBackupID string
	recoverCmd := &cobra.Command{Use: "recover", Short: "Recover a dead-owner migration marker", RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if recoverMigrationID == "" || recoverBackupID == "" {
			return fmt.Errorf("--migration-id and --backup-id are required")
		}
		return recoverHomeState(cmd.Context(), cwd, recoverMigrationID, recoverBackupID)
	}}
	recoverCmd.Flags().StringVar(&recoverMigrationID, "migration-id", "", "Migration identifier from the marker")
	recoverCmd.Flags().StringVar(&recoverBackupID, "backup-id", "", "Verified backup identifier")
	var rollbackMigrationID, rollbackBackupID string
	rollbackCmd := &cobra.Command{Use: "rollback", Short: "Restore a verified home-state backup", RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if rollbackMigrationID == "" || rollbackBackupID == "" {
			return fmt.Errorf("--migration-id and --backup-id are required")
		}
		if rollbackMigrationID != rollbackBackupID {
			return fmt.Errorf("migration and backup identities must match")
		}
		return rollbackHomeState(cmd.Context(), cwd, rollbackBackupID)
	}}
	rollbackCmd.Flags().StringVar(&rollbackMigrationID, "migration-id", "", "Completed migration identifier")
	rollbackCmd.Flags().StringVar(&rollbackBackupID, "backup-id", "", "Verified backup identifier")
	cmd.AddCommand(recoverCmd, rollbackCmd)
	return cmd
}

func init() { migrateCmd.AddCommand(newMigrateHomeStateCmd()) }
