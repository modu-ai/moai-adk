//go:build !windows

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var preApprovalCar011Roles = []string{"mission-governor", "super-advisor"}

// An error after the non-model startup must be terminal in the shared ledger.
// Write the row atomically before returning the error to the test driver.
func preApprovalCar011Preflight(ledgerPath string, rows *[]map[string]any, check func() error) error {
	if err := check(); err != nil {
		entry := map[string]any{"kind": "stop", "fixture": "car011", "reason": err.Error(), "started_ns": time.Now().UnixNano()}
		next := append(append([]map[string]any{}, *rows...), entry)
		if writeErr := preApprovalWriteLedger(ledgerPath, next); writeErr != nil {
			return fmt.Errorf("car011 preflight %v; stop ledger write failed: %w", err, writeErr)
		}
		*rows = next
		return err
	}
	return nil
}

func preApprovalCar011Version(procs *liveProcs, bin, home string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := liveCommand(ctx, bin, "--version")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "CODEX_HOME=" + home}
	out, err := procs.run(cmd)
	if err != nil {
		return "", fmt.Errorf("codex --version: %w: %s", err, out)
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "codex-cli "), nil
}

func preApprovalCar011RequireVersion(version string) error {
	if version != "0.157.0" {
		return fmt.Errorf("Codex version %q differs from 0.157.0", version)
	}
	return nil
}

// The preparation-only entry never links operator auth or starts a LIVE turn.
// Use a copied baseline directory when validating the fixture implementation.
func TestCodexPreApprovalCar011StartupOnly(t *testing.T) {
	if os.Getenv("MOAI_CAR011_PREPARE_ONLY") != "1" {
		t.Skip("NOT_RUN MOAI_CAR011_PREPARE_ONLY is not 1")
	}
	preApprovalRunCar011Startup(t)
}

type preApprovalCar011Inputs struct {
	Manifest                              preApprovalManifest
	ExportID, ExportSHA256                string
	Paths                                 []string
	Home, Project, RootState, Build, Repo []byte
	Argv, Prompt                          map[string][]byte
}

func preApprovalReadCar011Exports(dir string) (preApprovalCar011Inputs, error) {
	var out preApprovalCar011Inputs
	body, err := os.ReadFile(filepath.Join(dir, "export-manifest.json"))
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(body, &out.Manifest); err != nil || out.Manifest.ExportedNS <= 0 || len(out.Manifest.Files) != 9 {
		return out, errors.New("car011 manifest must contain nine exported inputs")
	}
	out.ExportID = fmt.Sprintf("export-%d", out.Manifest.ExportedNS)
	snapshot, err := os.ReadFile(filepath.Join(dir, "exports", out.ExportID, "export-manifest.json"))
	if err != nil || !bytes.Equal(body, snapshot) {
		return out, errors.New("car011 export snapshot differs")
	}
	out.ExportSHA256 = preApprovalSHA(snapshot)
	out.Argv, out.Prompt = map[string][]byte{}, map[string][]byte{}
	want := []string{"codex-home-config.toml", "project-config.toml", "repo.txt", "moai-build.txt", "root-state.txt"}
	for _, role := range preApprovalCar011Roles {
		want = append(want, role+"/argv.txt", role+"/prompt.txt")
	}
	for _, key := range want {
		data, readErr := os.ReadFile(filepath.Join(dir, "inputs", filepath.FromSlash(key)))
		if readErr != nil || len(data) == 0 || preApprovalSHA(data) != out.Manifest.Files[key] {
			return out, fmt.Errorf("car011 input mismatch: %s", key)
		}
		out.Paths = append(out.Paths, filepath.Join(dir, "inputs", filepath.FromSlash(key)))
		switch key {
		case "codex-home-config.toml":
			out.Home = data
		case "project-config.toml":
			out.Project = data
		case "repo.txt":
			out.Repo = data
		case "moai-build.txt":
			out.Build = data
		case "root-state.txt":
			out.RootState = data
		default:
			role, suffix, _ := strings.Cut(key, "/")
			if suffix == "argv.txt" {
				out.Argv[role] = data
			} else {
				out.Prompt[role] = data
			}
		}
	}
	actual := 0
	if err := filepath.WalkDir(filepath.Join(dir, "inputs"), func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			actual++
		}
		return nil
	}); err != nil || actual != len(want) {
		return out, fmt.Errorf("car011 unexpected input count: %d", actual)
	}
	return out, nil
}

// preApprovalRunCar011Live keeps the carried AC's observations but binds every
// launch to a fresh export and successful non-model startup in the same root.
func preApprovalRunCar011Live(t *testing.T) {
	if os.Getenv(envCodexRoleLive) != "1" {
		t.Skip("NOT_RUN " + envCodexRoleLive + " is not 1")
	}
	_, authPath := linkedCodexHome(t)
	f := preApprovalRunCar011Startup(t)
	ledgerPath := filepath.Join(f.evidenceDir, "ledger.json")
	rows, err := preApprovalReadLedger(ledgerPath)
	if err != nil {
		t.Fatalf("ABORTED: car011 ledger unreadable after startup: %v", err)
	}
	dir := filepath.Join(f.evidenceDir, "car011")
	var inputs preApprovalCar011Inputs
	version := ""
	err = preApprovalCar011Preflight(ledgerPath, &rows, func() error {
		var readErr error
		inputs, readErr = preApprovalReadCar011Exports(dir)
		if readErr != nil {
			return fmt.Errorf("car011 export: %w", readErr)
		}
		version, readErr = preApprovalCar011Version(f.procs, f.codexBin, f.codexHome)
		if readErr != nil {
			return readErr
		}
		if readErr = preApprovalCar011RequireVersion(version); readErr != nil {
			return readErr
		}
		fixtureHead, gitErr := f.git("-C", f.root, "rev-parse", "HEAD")
		if gitErr != nil || string(inputs.Repo) != "is_repo=true\nhead="+strings.TrimSpace(fixtureHead)+"\n" {
			return errors.New("car011 exported repository differs from fresh fixture")
		}
		if !bytes.Contains(inputs.Build, []byte("sha256="+f.moaiHash+"\n")) {
			return errors.New("car011 exported moai binary differs")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("ABORTED: %v", err)
	}
	stop := func(reason string) {
		preApprovalAppendStop(t, ledgerPath, &rows, "car011", reason)
		t.Fatalf("ABORTED: %s", reason)
	}
	budget := newLiveBudget(codexAuditLive011Budget, codexAuditLive011Window)
	var roles []liveAuditItem
	audit := &codexAuditLiveFixture{evDir: f.evidenceDir, root: f.root, codexHome: f.codexHome,
		realAuth: authPath, codexBin: f.codexBin, launchLog: f.launchLog, procs: f.procs}
	for _, role := range preApprovalCar011Roles {
		if ok, reason := budget.take(); !ok {
			stop(reason)
		}
		if fileSHA256OrEmpty(authPath) == "" {
			stop("operator auth unreadable")
		}
		args, err := preApprovalCar011Args(f.root, role, f.codexBin)
		if err != nil || !preApprovalCar011ExportsMatch(f.root, role, inputs.Argv[role], inputs.Prompt[role], args) {
			stop("car011 per-role argv or prompt differs from export")
		}
		proofRow := preApprovalLatestStartupRow(rows, "car011")
		if proofRow == nil {
			stop("car011 startup row missing")
		}
		startupID, _ := proofRow["startup_id"].(string)
		proof, launches, err := preApprovalStartupFromFiles(f.evidenceDir, "car011", startupID)
		if err != nil {
			stop(err.Error())
		}
		_, exportHash, err := preApprovalExportReference(f.evidenceDir, "car011")
		if err != nil || exportHash != inputs.ExportSHA256 {
			stop("car011 export reference differs")
		}
		hashes := map[string]string{"codex_home_config": preApprovalSHA(inputs.Home),
			"project_config": preApprovalSHA(inputs.Project), "root_state": preApprovalSHA(inputs.RootState)}
		if _, ok := preApprovalLatestStartup(rows, "car011", inputs.Manifest.ExportedNS, f.tempRoot,
			f.root, inputs.ExportID, inputs.ExportSHA256, hashes, proof, launches); !ok {
			stop("car011 startup ledger, root, proof, or export differs")
		}
		if preApprovalLiveCount(rows) >= preApprovalMaxLiveCalls {
			stop("absolute LIVE budget exceeded")
		}
		if err := preApprovalPrepareArm(f, authPath, role, inputs.Home, inputs.Project, inputs.RootState); err != nil {
			stop(err.Error())
		}
		beforeAuth := fileSHA256OrEmpty(authPath)
		if beforeAuth == "" {
			stop("operator auth changed before LIVE")
		}
		bound := codexAuditLiveCallBound
		if remaining := budget.remaining(); remaining < bound {
			bound = remaining
		}
		if bound <= 0 {
			stop("car011 wall window exhausted")
		}
		var out, diag bytes.Buffer
		probe := "audit-probe-" + role + ".txt"
		req := codexAuditRequest{Role: role, ProjectRoot: f.root, CallerDir: f.root, Root: f.root,
			Route: codexAuditRouteDirect, Program: f.codexBin,
			Task: strings.NewReader(string(inputs.Prompt[role])), Timeout: bound, Stdout: &out, Stderr: &diag}
		plan := prepareCodexAudit(context.Background(), req)
		if plan == nil {
			stop("car011 launcher refused before process start: " + diag.String())
		}
		if !bytes.Equal(inputs.Argv[role], preApprovalArgvBytes(plan.argv)) {
			_ = plan.recFile.Close()
			_ = plan.recDir.Close()
			stop("car011 launcher argv differs from exported argv")
		}
		before := listCodexRollouts(f.codexHome)
		start := time.Now().UnixNano()
		res := plan.run(context.Background())
		end := time.Now().UnixNano()
		row := map[string]any{"kind": "live", "fixture": "car011", "label": role,
			"attempt": f.attempt, "startup_id": startupID, "export_id": inputs.ExportID,
			"export_sha256": inputs.ExportSHA256, "temp_root": f.tempRoot,
			"fixture_root": f.root, "fixture_sentinel": true, "started_ns": start,
			"ended_ns": end, "argv": plan.argv, "exit": res.ExitCode, "bound_by": "test",
			"auth_sha256_before": beforeAuth, "auth_sha256_after": fileSHA256OrEmpty(authPath),
			"inputs_sha256": map[string]string{"codex_home_config": preApprovalSHA(inputs.Home),
				"project_config": preApprovalSHA(inputs.Project), "root_state": preApprovalSHA(inputs.RootState),
				"argv": preApprovalSHA(inputs.Argv[role]), "prompt": preApprovalSHA(inputs.Prompt[role]),
				"moai_binary": f.moaiHash}}
		rows = append(rows, row)
		if err := preApprovalWriteLedger(ledgerPath, rows); err != nil {
			t.Fatal(err)
		}
		records, top := audit.newRecords(t, before, "")
		rec, _ := audit.copyLaunchRecord(t, res.RecordPath)
		_, statErr := os.Stat(filepath.Join(f.root, probe))
		item := liveAuditItem{Role: role, ProbeExists: statErr == nil, ExitCode: res.ExitCode,
			LaunchRecord: res.RecordPath, ReturnedText: out.String()}
		d := deriveCodexAuditEvidence(codexAuditEvidenceInput{Role: role, Rollouts: records,
			LaunchRecord: rec, ProbeCommand: fmt.Sprintf("printf '%%s' audit > %s", probe),
			ProbeExists: item.ProbeExists})
		item.SessionSandbox, item.UsedSpawnAgent, item.Route = d.SessionSandbox, d.UsedSpawnAgent, d.Route
		item.ProbeCommandExecuted, item.ProbeExitCode, item.WriteDenied = d.ProbeCommandExecuted, d.ProbeExitCode, d.WriteDenied
		item.AttemptOutput = boundText(top.outputsMentioning(probe), 4000)
		roles = append(roles, item)
		if end-start > int64(codexAuditLiveCallBound) || fileSHA256OrEmpty(authPath) != beforeAuth {
			stop("car011 call bound or operator auth changed")
		}
	}
	ev := map[string]any{"invocations": budget.used, "aborted": budget.aborted(), "roles": roles,
		"ledger": rows, "elapsed_seconds": budget.elapsedSeconds(), "cleaned_pids": f.procs.reap()}
	emitLiveEvidence(t, f.evidenceDir, "ac-car-011-evidence.json", "ACCAR011", "", ev)
	for _, item := range roles {
		if item.SessionSandbox != "read-only" || !item.ProbeCommandExecuted || item.ProbeExitCode == nil ||
			*item.ProbeExitCode == 0 || item.ProbeExists || !item.WriteDenied || item.AttemptOutput == "" {
			t.Errorf("%s: sandbox=%q executed=%v exit=%v exists=%v denied=%v output=%q", item.Role,
				item.SessionSandbox, item.ProbeCommandExecuted, item.ProbeExitCode, item.ProbeExists,
				item.WriteDenied, item.AttemptOutput)
		}
	}
	if len(roles) != codexAuditLive011Budget {
		t.Errorf("roles run = %d, want %d", len(roles), codexAuditLive011Budget)
	}
}

func preApprovalLiveCount(rows []map[string]any) int {
	n := 0
	for _, row := range rows {
		if row["kind"] == "live" {
			n++
		}
	}
	return n
}

func TestCodexPreApprovalCar011PreparationBoundary(t *testing.T) {
	baseline := []map[string]any{{"kind": "startup", "fixture": "car011"}, {"kind": "startup", "fixture": "car011"}}
	if attempt, err := preApprovalNextCarAttempt(baseline, "car011"); err != nil || attempt != 1 {
		t.Fatalf("fresh preparation: attempt=%d err=%v", attempt, err)
	}
	if _, err := preApprovalNextCarAttempt(append(baseline, map[string]any{"kind": "startup", "fixture": "car011"}), "car011"); err == nil {
		t.Fatal("repeated preparation without completed LIVE accepted")
	}
	prior := append(append([]map[string]any{}, baseline...),
		map[string]any{"kind": "startup", "fixture": "car011"},
		map[string]any{"kind": "live", "fixture": "car011", "attempt": 1, "ended_ns": int64(10)},
		map[string]any{"kind": "live", "fixture": "car011", "attempt": 1, "ended_ns": int64(11)})
	if _, err := preApprovalNextCarAttempt(prior, "car011"); err == nil {
		t.Fatal("retry without lead invalidation accepted")
	}
	prior = append(prior, map[string]any{"kind": "invalidate", "fixture": "car011", "attempt": 1,
		"recorded_by": "lead", "recorded_ns": int64(12)})
	if attempt, err := preApprovalNextCarAttempt(prior, "car011"); err != nil || attempt != 2 {
		t.Fatalf("authorized retry: attempt=%d err=%v", attempt, err)
	}
	role := "mission-governor"
	args := []string{"exec", "-s", "read-only", "-"}
	prompt := []byte(preApprovalCar011Prompt(role))
	if !preApprovalCar011ExportsMatch("/tmp/fixture", role, preApprovalArgvBytes(args), prompt, args) {
		t.Fatal("matching role export refused")
	}
	if preApprovalCar011ExportsMatch("/tmp/fixture", "super-advisor", preApprovalArgvBytes(args), prompt, args) ||
		preApprovalCar011ExportsMatch("/tmp/fixture", role, preApprovalArgvBytes(args), append(prompt, '!'), args) ||
		preApprovalCar011ExportsMatch("/tmp/fixture", role, preApprovalArgvBytes(append(args, "extra")), prompt, args) {
		t.Fatal("wrong label, prompt, or argv accepted")
	}
}

func TestCodexPreApprovalCar011ExportIntegrity(t *testing.T) {
	dir := t.TempDir()
	manifest := preApprovalManifest{ExportedNS: time.Now().UnixNano(), Files: map[string]string{}}
	keys := []string{"codex-home-config.toml", "project-config.toml", "repo.txt", "moai-build.txt", "root-state.txt"}
	for _, role := range preApprovalCar011Roles {
		keys = append(keys, role+"/argv.txt", role+"/prompt.txt")
	}
	for _, key := range keys {
		preApprovalExport(t, &manifest, filepath.Join(dir, "inputs"), key, []byte("test "+key))
	}
	preApprovalWriteJSON(t, filepath.Join(dir, "export-manifest.json"), manifest)
	copy, err := os.ReadFile(filepath.Join(dir, "export-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	exportID := fmt.Sprintf("export-%d", manifest.ExportedNS)
	path := filepath.Join(dir, "exports", exportID, "export-manifest.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, copy, 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := preApprovalReadCar011Exports(dir); err != nil || len(got.Paths) != 9 {
		t.Fatalf("valid export refused: paths=%d err=%v", len(got.Paths), err)
	}
	roleFile := filepath.Join(dir, "inputs", "mission-governor", "prompt.txt")
	if err := os.WriteFile(roleFile, []byte("mutated"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := preApprovalReadCar011Exports(dir); err == nil {
		t.Fatal("mutated role prompt accepted")
	}
	if err := os.WriteFile(roleFile, []byte("test mission-governor/prompt.txt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := preApprovalReadCar011Exports(dir); err == nil {
		t.Fatal("mutated export snapshot accepted")
	}
}

func TestCodexPreApprovalCar011PostStartupFailuresStop(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check func(string) error
	}{
		{"fake_version", func(string) error { return preApprovalCar011RequireVersion("0.158.0") }},
		{"missing_export", func(dir string) error {
			_, err := preApprovalReadCar011Exports(dir)
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			ledgerPath := filepath.Join(dir, "ledger.json")
			rows := []map[string]any{
				{"kind": "startup", "fixture": "car011", "started_ns": int64(1), "ended_ns": int64(2)},
				{"kind": "live", "fixture": "disc-control", "started_ns": int64(3), "ended_ns": int64(4)},
			}
			if err := preApprovalWriteLedger(ledgerPath, rows); err != nil {
				t.Fatal(err)
			}
			if err := preApprovalCar011Preflight(ledgerPath, &rows, func() error { return tc.check(dir) }); err == nil {
				t.Fatal("post-startup failure accepted")
			}
			persisted, err := preApprovalReadLedger(ledgerPath)
			if err != nil {
				t.Fatal(err)
			}
			if len(persisted) != 3 || persisted[0]["kind"] != "startup" || persisted[1]["kind"] != "live" ||
				persisted[2]["kind"] != "stop" || persisted[2]["fixture"] != "car011" ||
				preApprovalNumber(persisted[2]["started_ns"]) <= 0 || preApprovalLiveCount(persisted) != 1 {
				t.Fatalf("stop was not persisted after prior calls: %+v", persisted)
			}
			if err := preApprovalCar011Preflight(ledgerPath, &persisted, func() error { return nil }); err != nil {
				t.Fatal(err)
			}
			if len(persisted) != 3 {
				t.Fatalf("successful check changed ledger: %d", len(persisted))
			}
		})
	}
}

func TestCodexPreApprovalCar011GitStatusFailureStops(t *testing.T) {
	tempRoot := canonicalDir(t, t.TempDir())
	root := filepath.Join(tempRoot, "fixture")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	token := preApprovalSentinel(t, root)
	statusCalls, cleanCalls := 0, 0
	git := func(args ...string) (string, error) {
		if len(args) < 4 || args[0] != "-C" || args[1] != root {
			return "", fmt.Errorf("unexpected git args: %v", args)
		}
		switch args[2] {
		case "rev-parse":
			return root + "\n", nil
		case "clean":
			cleanCalls++
			return "", nil
		case "status":
			statusCalls++
			return "", errors.New("injected status exit 42")
		default:
			return "", fmt.Errorf("unexpected git args: %v", args)
		}
	}
	f := &preApprovalFixture{tempRoot: tempRoot, root: root, token: token,
		codexHome: filepath.Join(tempRoot, "codex-home"), git: git}
	ledgerPath := filepath.Join(tempRoot, "ledger.json")
	rows := []map[string]any{{"kind": "startup", "fixture": "car011", "started_ns": int64(1), "ended_ns": int64(2)},
		{"kind": "live", "fixture": "disc-control", "started_ns": int64(3), "ended_ns": int64(4)}}
	if err := preApprovalWriteLedger(ledgerPath, rows); err != nil {
		t.Fatal(err)
	}
	err := preApprovalCar011Preflight(ledgerPath, &rows, func() error {
		return preApprovalPrepareArm(f, "unused-auth", "mission-governor",
			[]byte("user-config"), []byte("project-config"), []byte("root-state"))
	})
	if err == nil || !strings.Contains(err.Error(), "injected status exit 42") {
		t.Fatalf("git status failure not returned: %v", err)
	}
	stored, err := preApprovalReadLedger(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	if statusCalls != 1 || cleanCalls != 1 || len(stored) != 3 || stored[2]["kind"] != "stop" ||
		stored[2]["fixture"] != "car011" || preApprovalLiveCount(stored) != 1 {
		t.Fatalf("pre-LIVE failure escaped stop: status=%d clean=%d rows=%v", statusCalls, cleanCalls, stored)
	}
}

func preApprovalCar011Prompt(role string) string {
	return codexAuditProbeTask(role, "audit-probe-"+role+".txt", "")
}

// This builds the exact argv shape from the same role and MCP discovery data
// used by the launcher. The LIVE path also compares it with plan.argv before
// plan.run, so any future launcher drift stops before a model call.
func preApprovalCar011Args(root, role, codexBin string) ([]string, error) {
	entry, err := codexAuditLoadRole(root, role)
	if err != nil {
		return nil, err
	}
	names, err := codexAuditMCPServerNames(context.Background(), codexBin, root)
	if err != nil {
		return nil, err
	}
	instructions, err := json.Marshal(entry.instructions)
	if err != nil {
		return nil, err
	}
	args := []string{"exec", "-s", codexAuditSandbox, "-c", `approval_policy="never"`,
		"-c", `model_reasoning_effort="` + entry.effort + `"`,
		"-c", "developer_instructions=" + string(instructions)}
	for _, name := range names {
		args = append(args, "-c", "mcp_servers."+name+".enabled=false")
	}
	args = append(args, "-C", root, "--json", "-")
	return args, nil
}

func preApprovalCar011ExportsMatch(root string, role string, argv, prompt []byte, expected []string) bool {
	return (role == "mission-governor" || role == "super-advisor") && root != "" &&
		bytes.Equal(prompt, []byte(preApprovalCar011Prompt(role))) &&
		bytes.Equal(argv, preApprovalArgvBytes(expected))
}

// preApprovalNextCarAttempt does not infer a retry from a failed test process.
// The lead must first invalidate the completed attempt in the shared ledger.
func preApprovalNextCarAttempt(rows []map[string]any, fixture string) (int, error) {
	if fixture != "car011" {
		return 0, errors.New("unsupported carried fixture")
	}
	starts, calls := 0, 0
	var lastEnd, invalidatedAt int64
	for _, row := range rows {
		switch row["kind"] {
		case "stop":
			return 0, errors.New("stopped evidence directory cannot be reused")
		case "startup":
			if row["fixture"] == fixture {
				starts++
			}
		case "live":
			if row["fixture"] == fixture {
				calls++
				if preApprovalNumber(row["attempt"]) != 1 {
					return 0, errors.New("car011 retry limit reached")
				}
				if end := preApprovalNumber(row["ended_ns"]); end > lastEnd {
					lastEnd = end
				}
			}
		case "invalidate":
			if row["fixture"] == fixture && preApprovalNumber(row["attempt"]) == 1 && row["recorded_by"] == "lead" {
				invalidatedAt = preApprovalNumber(row["recorded_ns"])
			}
		}
	}
	if starts < 2 {
		return 0, fmt.Errorf("car011 M1-a/M1-b startup baseline incomplete: %d", starts)
	}
	if calls == 0 {
		if starts != 2 || invalidatedAt != 0 {
			return 0, errors.New("car011 preparation already attempted or orphan invalidation")
		}
		return 1, nil
	}
	if calls != 2 || starts != 3 || invalidatedAt <= lastEnd {
		return 0, errors.New("car011 retry requires two completed LIVE calls and later lead invalidation")
	}
	return 2, nil
}
