//go:build !windows

package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

type preApprovalStartupRow struct {
	Kind            string            `json:"kind"`
	Fixture         string            `json:"fixture"`
	TempRoot        string            `json:"temp_root"`
	FixtureRoot     string            `json:"fixture_root"`
	FixtureSentinel bool              `json:"fixture_sentinel"`
	StartedNS       int64             `json:"started_ns"`
	EndedNS         int64             `json:"ended_ns"`
	Argv            []string          `json:"argv"`
	Exit            int               `json:"exit"`
	BoundBy         string            `json:"bound_by"`
	InputsSHA256    map[string]string `json:"inputs_sha256"`
	ProviderBaseURL string            `json:"provider_base_url"`
	AuthFilePresent bool              `json:"auth_file_present"`
	EnvAuthPresent  bool              `json:"env_auth_present"`
	MoaiLaunches    int               `json:"moai_launches"`
	DecoyLaunches   int               `json:"decoy_launches"`
	Passed          bool              `json:"passed"`
	StdoutSHA256    string            `json:"stdout_sha256"`
	StderrSHA256    string            `json:"stderr_sha256"`
	LaunchesSHA256  string            `json:"launches_sha256"`
	StartupID       string            `json:"startup_id,omitempty"`
	ExportID        string            `json:"export_id,omitempty"`
	ExportSHA256    string            `json:"export_sha256,omitempty"`
	Reason          string            `json:"reason,omitempty"`
}

type preApprovalManifest struct {
	ExportedNS int64             `json:"exported_ns"`
	Files      map[string]string `json:"files"`
}

type preApprovalFixture struct {
	evidenceDir, tempRoot, root, codexHome, codexBin, moaiBin, moaiHash, launchLog, token string
	worktrees                                                                             []string
	git                                                                                   func(...string) (string, error)
	procs                                                                                 *liveProcs
	attempt                                                                               int
}

const preApprovalMaxStartupCalls = 16

func preApprovalWriteJSON(t *testing.T, path string, value any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

func preApprovalExport(t *testing.T, manifest *preApprovalManifest, dir, name string, data []byte) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	manifest.Files[name] = preApprovalSHA(data)
}

func preApprovalRunSplit(procs *liveProcs, cmd *exec.Cmd) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := procs.start(cmd); err != nil {
		return nil, nil, err
	}
	err := cmd.Wait()
	procs.mu.Lock()
	procs.waited[cmd] = true
	procs.mu.Unlock()
	_ = killLiveProcGroup(cmd.Process.Pid)
	return stdout.Bytes(), stderr.Bytes(), err
}

func preApprovalRootState(t *testing.T, root string, git func(...string) (string, error)) []byte {
	t.Helper()
	status, err := git("-C", root, "status", "--porcelain", "--ignored")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if path != root && !d.IsDir() {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			paths = append(paths, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return []byte("status:\n" + status + "files:\n" + strings.Join(paths, "\n") + "\n")
}

func preApprovalSHA(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func preApprovalDiscriminatorPrompt(root string) string {
	return fmt.Sprintf("Use the exec custom tool to call the moai MCP tool codex_role_audit exactly once with role %q, worktree_root %q, task %q, and out %q. Do not call any other tool, use shell commands, or change files. Report the tool refusal or result verbatim.",
		"sync-auditor", root, "Return a read-only audit verdict.", "AGENTS.md")
}

func preApprovalDiscriminatorArgs(root, prompt string) []string {
	return []string{"exec", "--strict-config", "-s", "workspace-write", "-c", `approval_policy="never"`, "-C", root, "--json", prompt}
}

func preApprovalArgvBytes(args []string) []byte { return []byte(strings.Join(args, "\n") + "\n") }

func preApprovalArgvMatches(root string, prompt, exported []byte) bool {
	return bytes.Equal(exported, preApprovalArgvBytes(preApprovalDiscriminatorArgs(root, string(prompt))))
}

// TestCodexPreApprovalStartup measures MCP launch only. Its CODEX_HOME has no
// login file or auth environment, and the provider points at a closed local
// port. Any model item is a stop, never evidence of preflight success.
func TestCodexPreApprovalStartup(t *testing.T) {
	preApprovalRunStartup(t, false)
}

// TestCodexPreApprovalStartupPreserve exercises M1-b's non-model startup
// preparation without proceeding to either LIVE discriminator arm.
func TestCodexPreApprovalStartupPreserve(t *testing.T) {
	preApprovalRunStartup(t, true)
}

// preApprovalRunCar011Startup prepares only the carried read-only role fixture.
// The other three fixtures have already consumed their M1-b startup allowance.
func preApprovalRunCar011Startup(t *testing.T) *preApprovalFixture {
	return preApprovalRunStartupFor(t, true, "car011")
}

func preApprovalAppendStartup(t *testing.T, path string, rows *[]map[string]any, entry preApprovalStartupRow, stopReason string) {
	t.Helper()
	body, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	var row map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&row); err != nil {
		t.Fatal(err)
	}
	next := append(append([]map[string]any{}, *rows...), row)
	if stopReason != "" {
		next = append(next, map[string]any{"kind": "stop", "fixture": entry.Fixture, "reason": stopReason, "started_ns": time.Now().UnixNano()})
	}
	if err := preApprovalWriteLedger(path, next); err != nil {
		t.Fatal(err)
	}
	*rows = next
}

func preApprovalAppendStop(t *testing.T, path string, rows *[]map[string]any, fixture, reason string) {
	t.Helper()
	next := append(append([]map[string]any{}, *rows...), map[string]any{"kind": "stop", "fixture": fixture, "reason": reason, "started_ns": time.Now().UnixNano()})
	if err := preApprovalWriteLedger(path, next); err != nil {
		t.Fatal(err)
	}
	*rows = next
}

func preApprovalExportReference(evidenceDir, fixture string) (string, string, error) {
	group := fixture
	if strings.HasPrefix(group, "disc-") {
		group = "discriminator"
	}
	current, err := os.ReadFile(filepath.Join(evidenceDir, group, "export-manifest.json"))
	if err != nil {
		return "", "", err
	}
	var export preApprovalManifest
	if err := json.Unmarshal(current, &export); err != nil || export.ExportedNS <= 0 {
		return "", "", fmt.Errorf("invalid %s export manifest: %v", group, err)
	}
	id := fmt.Sprintf("export-%d", export.ExportedNS)
	snapshot, err := os.ReadFile(filepath.Join(evidenceDir, group, "exports", id, "export-manifest.json"))
	if err != nil {
		return "", "", err
	}
	if !bytes.Equal(snapshot, current) {
		return "", "", errors.New("export snapshot differs from current manifest")
	}
	return id, preApprovalSHA(snapshot), nil
}

func preApprovalStartupCount(rows []map[string]any) int {
	count := 0
	for _, row := range rows {
		if row["kind"] == "startup" {
			count++
		}
	}
	return count
}

func TestCodexPreApprovalStartupLedgerRetention(t *testing.T) {
	dir := t.TempDir()
	ledgerPath := filepath.Join(dir, "ledger.json")
	rawPath := filepath.Join(dir, "failed.jsonl")
	if err := os.WriteFile(rawPath, []byte(`{"type":"thread.started"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rows := []map[string]any{{"kind": "startup", "fixture": "disc-control"}}
	if err := preApprovalWriteLedger(ledgerPath, rows); err != nil {
		t.Fatal(err)
	}
	preApprovalAppendStartup(t, ledgerPath, &rows, preApprovalStartupRow{Kind: "startup", Fixture: "disc-treatment", Passed: false}, "startup MCP launch or non-model gate failed")
	retained, err := preApprovalReadLedger(ledgerPath)
	if err != nil || len(retained) != 3 || preApprovalStartupCount(retained) != 2 || retained[2]["kind"] != "stop" {
		t.Fatalf("failed attempt was not recorded: rows=%+v err=%v", retained, err)
	}
	if _, err := os.Stat(rawPath); err != nil {
		t.Fatalf("failed attempt raw output was lost: %v", err)
	}
	capPath := filepath.Join(dir, "cap-ledger.json")
	var capRows []map[string]any
	for preApprovalStartupCount(capRows) < preApprovalMaxStartupCalls {
		preApprovalAppendStartup(t, capPath, &capRows, preApprovalStartupRow{Kind: "startup", Fixture: "car011"}, "")
	}
	preApprovalAppendStop(t, capPath, &capRows, "car011", "startup invocation budget exceeded")
	retained, err = preApprovalReadLedger(capPath)
	if err != nil || preApprovalStartupCount(retained) != preApprovalMaxStartupCalls || retained[len(retained)-1]["kind"] != "stop" {
		t.Fatalf("budget refusal changed startup count: rows=%d err=%v", preApprovalStartupCount(retained), err)
	}
	stopNS := preApprovalNumber(retained[len(retained)-1]["started_ns"])
	if stopNS <= 0 {
		t.Fatal("stop row has no timestamp")
	}
	earlier := append([]map[string]any{{"kind": "live", "fixture": "disc-control", "started_ns": stopNS - 1}}, retained...)
	if err := preApprovalWriteLedger(filepath.Join(dir, "earlier-ledger.json"), earlier); err != nil {
		t.Fatalf("LIVE before stop was rejected: %v", err)
	}
	untimed := append(append([]map[string]any{}, retained[:len(retained)-1]...), map[string]any{"kind": "stop", "fixture": "car011"})
	if err := preApprovalWriteLedger(filepath.Join(dir, "untimed-ledger.json"), untimed); err == nil {
		t.Fatal("untimed stop was written")
	}
	late := append(append([]map[string]any{}, retained...), map[string]any{"kind": "live", "fixture": "disc-control", "started_ns": stopNS + 1})
	if err := preApprovalWriteLedger(capPath, late); err == nil {
		t.Fatal("LIVE after stop was written")
	}
	unchanged, err := preApprovalReadLedger(capPath)
	if err != nil || len(unchanged) != len(retained) {
		t.Fatalf("refused LIVE changed ledger: rows=%d err=%v", len(unchanged), err)
	}
}

func preApprovalRunStartup(t *testing.T, preserveLedger bool) *preApprovalFixture {
	return preApprovalRunStartupFor(t, preserveLedger, "")
}

func preApprovalRunStartupFor(t *testing.T, preserveLedger bool, only string) *preApprovalFixture {
	t.Helper()
	if os.Getenv(envCodexPreApprovalLive) != "1" && !(only == "car011" && os.Getenv(envCodexRoleLive) == "1") {
		t.Skip("NOT_RUN " + envCodexPreApprovalLive + " is not 1")
	}
	evidenceDir := os.Getenv(envT1172EvidenceDir)
	if evidenceDir == "" {
		t.Skip("NOT_RUN " + envT1172EvidenceDir + " is empty")
	}
	if !filepath.IsAbs(evidenceDir) && !strings.HasPrefix(evidenceDir, "../") {
		evidenceDir = filepath.Join("../..", evidenceDir)
	}
	var err error
	evidenceDir, err = filepath.Abs(evidenceDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(evidenceDir, "ledger.json")
	var rows []map[string]any
	attempt := 1
	if preserveLedger {
		prior, readErr := preApprovalReadLedger(ledgerPath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if !preApprovalM1aBaseline(evidenceDir, prior) {
			t.Fatal("M1-a startup ledger or raw evidence differs from the recorded baseline")
		}
		if only == "" {
			attempt, readErr = preApprovalNextAttempt(prior)
			if readErr != nil {
				t.Fatal(readErr)
			}
		} else if only == "car011" {
			attempt, readErr = preApprovalNextCarAttempt(prior, only)
			if readErr != nil {
				t.Fatal(readErr)
			}
		} else {
			t.Fatalf("unsupported startup fixture %q", only)
		}
		if attempt == 2 && only == "" {
			for _, name := range []string{"evidence.json", "outcome.txt", "live.jsonl"} {
				if _, err := os.Stat(filepath.Join(evidenceDir, "discriminator", "attempt-1", name)); err != nil {
					t.Fatalf("attempt-1 archive required before retry: %s: %v", name, err)
				}
			}
		}
		if attempt == 2 && only == "car011" {
			for _, name := range []string{"ac-car-011-evidence.json", "ac-car-011-live.jsonl"} {
				if _, err := os.Stat(filepath.Join(evidenceDir, "car011", "attempt-1", name)); err != nil {
					t.Fatalf("car011 attempt-1 archive required before retry: %s: %v", name, err)
				}
			}
		}
		startupNeed := 4
		if only != "" {
			startupNeed = 1
		}
		if preApprovalStartupCount(prior)+startupNeed > preApprovalMaxStartupCalls {
			rows = prior
			fixture := "disc"
			if only != "" {
				fixture = only
			}
			preApprovalAppendStop(t, ledgerPath, &rows, fixture, "startup invocation budget would be exceeded")
			t.Fatal("startup invocation budget would be exceeded")
		}
		rows = prior
	}
	codexBin := os.Getenv("MOAI_CODEX_BIN")
	if codexBin == "" {
		codexBin = "codex"
	}
	codexBin, err = exec.LookPath(codexBin)
	if err != nil {
		t.Skipf("CODEX_NOT_INSTALLED: %v", err)
	}
	procs := newLiveProcs(t)
	moaiBin := buildLiveMoai(t, procs)
	moaiBytes, err := os.ReadFile(moaiBin)
	if err != nil {
		t.Fatal(err)
	}
	moaiHash := preApprovalSHA(moaiBytes)
	tempRoot := canonicalDir(t, t.TempDir())
	root := filepath.Join(tempRoot, "fixture")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := liveCommand(ctx, "git", args...)
		returnString, runErr := procs.run(cmd)
		return string(returnString), runErr
	}
	for _, args := range [][]string{{"-C", root, "init", "-q"}, {"-C", root, "config", "user.name", "MoAI Fixture"}, {"-C", root, "config", "user.email", "fixture@example.invalid"}} {
		if out, err := git(args...); err != nil {
			t.Fatalf("fixture git: %v: %s", err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "README"), []byte("preapproval fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if only == "car011" {
		installAuditRoles(t, root)
	}
	for _, args := range [][]string{{"-C", root, "add", "README"}, {"-C", root, "commit", "-q", "-m", "fixture"}} {
		if out, err := git(args...); err != nil {
			t.Fatalf("fixture git: %v: %s", err, out)
		}
	}
	if only == "car011" {
		if out, err := git("-C", root, "add", ".codex/agents/moai"); err != nil {
			t.Fatalf("fixture role add: %v: %s", err, out)
		}
		if out, err := git("-C", root, "commit", "-q", "-m", "audit roles"); err != nil {
			t.Fatalf("fixture role commit: %v: %s", err, out)
		}
	}
	token := preApprovalSentinel(t, root)
	var worktrees []string
	repoWorktrees, err := git("-C", "../..", "worktree", "list", "--porcelain")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(repoWorktrees, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			worktrees = append(worktrees, strings.TrimPrefix(line, "worktree "))
		}
	}
	binDir := t.TempDir()
	launchLog := filepath.Join(binDir, "mcp-launches.log")
	for _, name := range []string{"moai", "decoy"} {
		path := filepath.Join(binDir, name)
		script := fmt.Sprintf("#!/bin/sh\nprintf '%s %%s\\n' \"$$\" >> %q\nexec %q mcp-server\n", name, launchLog, moaiBin)
		if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	baseProject := string(codexwiring.EnsureMCPTable(nil))
	baseProject = strings.Replace(baseProject, `command = "moai"`, "command = "+strconv.Quote(filepath.Join(binDir, "moai")), 1)
	if !strings.Contains(baseProject, filepath.Join(binDir, "moai")) {
		t.Fatal("writer did not emit replaceable moai command")
	}
	startupCalls := 0
	discManifest := preApprovalManifest{ExportedNS: time.Now().UnixNano(), Files: map[string]string{}}
	carManifests := map[string]*preApprovalManifest{
		"car010": {Files: map[string]string{}},
		"car011": {Files: map[string]string{}},
	}
	commit, err := git("-C", "../..", "rev-parse", "--short", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	fixtureHead, err := git("-C", root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	repoText := []byte("is_repo=true\nhead=" + strings.TrimSpace(fixtureHead) + "\n")
	buildText := []byte("commit=" + strings.TrimSpace(commit) + "\nsha256=" + moaiHash + "\n")
	prompt := []byte(preApprovalDiscriminatorPrompt(root))
	liveArgv := preApprovalArgvBytes(preApprovalDiscriminatorArgs(root, string(prompt)))
	for pass := 0; pass < 2; pass++ {
		fixtures := []string{"disc-control", "disc-treatment", "car010", "car011"}
		if only != "" {
			fixtures = []string{only}
		}
		for _, name := range fixtures {
			if err := preApprovalCleanFixture(root, tempRoot, token, worktrees, git); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(root, ".codex"), 0o700); err != nil {
				t.Fatal(err)
			}
			project := baseProject
			if name == "disc-treatment" {
				project += "\n[mcp_servers.moai.tools.codex_role_audit]\napproval_mode = \"approve\"\n"
			}
			if err := os.WriteFile(filepath.Join(root, ".codex", "config.toml"), []byte(project), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".fixture-sentinel"), []byte(token), 0o600); err != nil {
				t.Fatal(err)
			}
			codexHome := filepath.Join(tempRoot, "codex-home")
			if err := os.RemoveAll(codexHome); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(codexHome, 0o700); err != nil {
				t.Fatal(err)
			}
			user := "[projects." + strconv.Quote(root) + "]\ntrust_level = \"trusted\"\n"
			if name == "car010" || only == "car011" {
				user += "\n[mcp_servers.decoy]\ncommand = " + strconv.Quote(filepath.Join(binDir, "decoy")) + "\nargs = [\"mcp-server\"]\nenabled_tools = [\"spec_progress\"]\n"
			}
			if err := os.WriteFile(filepath.Join(codexHome, "config.toml"), []byte(user), 0o600); err != nil {
				t.Fatal(err)
			}
			if only == "car011" {
				t.Setenv(codexHomeEnvVar, codexHome)
			}
			rootState := preApprovalRootState(t, root, git)
			if pass == 0 {
				inputDir := filepath.Join(evidenceDir, name, "inputs")
				manifest := &discManifest
				prefix := strings.TrimPrefix(name, "disc-") + "/"
				if name == "disc-control" || name == "disc-treatment" {
					inputDir = filepath.Join(evidenceDir, "discriminator", "inputs")
				} else {
					manifest = carManifests[name]
					prefix = ""
					manifest.ExportedNS = time.Now().UnixNano()
				}
				for file, contents := range map[string][]byte{
					"codex-home-config.toml": []byte(user),
					"project-config.toml":    []byte(project),
					"repo.txt":               repoText,
					"moai-build.txt":         buildText,
					"root-state.txt":         rootState,
				} {
					preApprovalExport(t, manifest, inputDir, prefix+file, contents)
				}
				if strings.HasPrefix(name, "disc-") {
					preApprovalExport(t, manifest, inputDir, prefix+"argv.txt", liveArgv)
					preApprovalExport(t, manifest, inputDir, prefix+"prompt.txt", prompt)
				} else {
					labels := []string{"mission-governor", "super-advisor"}
					if name == "car010" {
						labels = []string{"direct", "parent"}
					}
					for _, label := range labels {
						labelPrompt := []byte(fmt.Sprintf("Run the %s read-only audit route for this fixture. Report the result without changing files.", label))
						labelArgv := preApprovalArgvBytes(preApprovalDiscriminatorArgs(root, string(labelPrompt)))
						if only == "car011" {
							labelPrompt = []byte(preApprovalCar011Prompt(label))
							args, argErr := preApprovalCar011Args(root, label, codexBin)
							if argErr != nil {
								t.Fatalf("car011 %s argv: %v", label, argErr)
							}
							labelArgv = preApprovalArgvBytes(args)
						}
						preApprovalExport(t, manifest, inputDir, label+"/argv.txt", labelArgv)
						preApprovalExport(t, manifest, inputDir, label+"/prompt.txt", labelPrompt)
					}
				}
				continue
			}
			if err := os.Remove(launchLog); err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			args := []string{"exec", "--strict-config", "-C", root, "--json", "-c", `model_providers.dead={name="dead",base_url="http://127.0.0.1:9/v1",wire_api="responses",stream_max_retries=0,request_max_retries=0}`, "-c", `model_provider="dead"`, "startup check"}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			cmd := liveCommand(ctx, codexBin, args...)
			cmd.Dir = root
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + t.TempDir(), "CODEX_HOME=" + codexHome, "MOAI_HOME=" + t.TempDir(), "LANG=C"}
			if preserveLedger {
				if preApprovalStartupCount(rows) >= preApprovalMaxStartupCalls {
					preApprovalAppendStop(t, ledgerPath, &rows, name, "startup invocation budget exceeded")
					t.Fatal("startup invocation budget exceeded")
				}
			}
			start := time.Now().UnixNano()
			out, stderr, runErr := preApprovalRunSplit(procs, cmd)
			end := time.Now().UnixNano()
			cancel()
			log, _ := os.ReadFile(launchLog)
			moaiLaunches := strings.Count(string(log), "moai ")
			decoyLaunches := strings.Count(string(log), "decoy ")
			exitCode := 0
			if runErr != nil {
				exitCode = 1
				if e, ok := runErr.(*exec.ExitError); ok {
					exitCode = e.ExitCode()
				}
			}
			row := preApprovalStartupRow{Kind: "startup", Fixture: name, TempRoot: tempRoot, FixtureRoot: root,
				FixtureSentinel: true, StartedNS: start, EndedNS: end, Argv: append([]string{codexBin}, args...),
				Exit: exitCode, BoundBy: "test", ProviderBaseURL: "http://127.0.0.1:9/v1", MoaiLaunches: moaiLaunches,
				DecoyLaunches: decoyLaunches, StdoutSHA256: preApprovalSHA(out), StderrSHA256: preApprovalSHA(stderr),
				InputsSHA256: map[string]string{"codex_home_config": preApprovalSHA([]byte(user)),
					"project_config": preApprovalSHA([]byte(project)), "moai_binary": moaiHash,
					"argv": preApprovalSHA(liveArgv), "prompt": preApprovalSHA(prompt), "root_state": preApprovalSHA(rootState)},
			}
			startupDir := filepath.Join(evidenceDir, "startup")
			if err := os.MkdirAll(startupDir, 0o755); err != nil {
				t.Fatal(err)
			}
			baseName := name
			if preserveLedger {
				row.StartupID = fmt.Sprintf("attempt-%d/%s-%d", attempt, name, start)
				startupDir = filepath.Join(startupDir, fmt.Sprintf("attempt-%d", attempt))
				baseName = fmt.Sprintf("%s-%d", name, start)
				if err := os.MkdirAll(startupDir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(startupDir, baseName+".jsonl"), out, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(startupDir, baseName+".err"), stderr, 0o600); err != nil {
				t.Fatal(err)
			}
			launchesPath := filepath.Join(startupDir, baseName+".launches.json")
			preApprovalWriteJSON(t, launchesPath, map[string]int{"moai": moaiLaunches, "decoy": decoyLaunches})
			launchesBytes, err := os.ReadFile(launchesPath)
			if err != nil {
				t.Fatal(err)
			}
			row.LaunchesSHA256 = preApprovalSHA(launchesBytes)
			proof := preApprovalStartupProof{Fixture: name, Stdout: out, Stderr: stderr,
				Launches: map[string]int{"moai": moaiLaunches, "decoy": decoyLaunches}}
			row.Passed = preApprovalStartupValid(proof) && end-start <= int64(25*time.Second)
			if preserveLedger {
				row.ExportID, row.ExportSHA256, err = preApprovalExportReference(evidenceDir, name)
				if err != nil {
					preApprovalAppendStartup(t, ledgerPath, &rows, row, "export reference failed: "+err.Error())
					t.Fatal(err)
				}
			}
			stopReason := ""
			if !row.Passed {
				stopReason = "startup MCP launch or non-model gate failed"
			}
			preApprovalAppendStartup(t, ledgerPath, &rows, row, stopReason)
			startupCalls++
			if !row.Passed {
				t.Fatal(stopReason)
			}
			t.Logf("%s startup exit=%d moai=%d decoy=%d items=0", name, exitCode, moaiLaunches, decoyLaunches)
		}
		if pass == 0 {
			if only == "" {
				preApprovalWriteJSON(t, filepath.Join(evidenceDir, "discriminator", "export-manifest.json"), discManifest)
			}
			for name, manifest := range carManifests {
				if only != "" && name != only {
					continue
				}
				preApprovalWriteJSON(t, filepath.Join(evidenceDir, name, "export-manifest.json"), manifest)
			}
			if preserveLedger {
				groups := []string{"discriminator", "car010", "car011"}
				if only != "" {
					groups = []string{only}
				}
				for _, group := range groups {
					body, err := os.ReadFile(filepath.Join(evidenceDir, group, "export-manifest.json"))
					if err != nil {
						t.Fatal(err)
					}
					var export preApprovalManifest
					if err := json.Unmarshal(body, &export); err != nil || export.ExportedNS <= 0 {
						t.Fatalf("invalid %s export manifest: %v", group, err)
					}
					snapshot := filepath.Join(evidenceDir, group, "exports", fmt.Sprintf("export-%d", export.ExportedNS), "export-manifest.json")
					if err := os.MkdirAll(filepath.Dir(snapshot), 0o755); err != nil {
						t.Fatal(err)
					}
					if _, err := os.Stat(snapshot); err == nil {
						t.Fatal("attempt export snapshot already exists")
					} else if !os.IsNotExist(err) {
						t.Fatal(err)
					}
					if err := os.WriteFile(snapshot, body, 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			if only != "" {
				continue
			}
			discDir := filepath.Join(evidenceDir, "discriminator")
			diff := exec.Command("diff", "-r", "inputs/control", "inputs/treatment")
			diff.Dir = discDir
			output, diffErr := diff.CombinedOutput()
			if e, ok := diffErr.(*exec.ExitError); !ok || e.ExitCode() != 1 {
				t.Fatalf("arm diff command: %v: %s", diffErr, output)
			}
			if !preApprovalOnlyProjectConfigDiff(string(output)) {
				t.Fatalf("arm diff exceeds approval table: %s", output)
			}
			if err := os.WriteFile(filepath.Join(discDir, "arm-diff.txt"), output, 0o600); err != nil {
				t.Fatal(err)
			}
			preApprovalWriteJSON(t, filepath.Join(discDir, "arm-diff.meta.json"), map[string]int64{"written_ns": time.Now().UnixNano()})
		}
	}
	wantStartups := 4
	if only != "" {
		wantStartups = 1
	}
	if startupCalls != wantStartups {
		t.Fatalf("startup stopped after %d entries; see ledger.json", startupCalls)
	}
	return &preApprovalFixture{evidenceDir: evidenceDir, tempRoot: tempRoot, root: root,
		codexHome: filepath.Join(tempRoot, "codex-home"), codexBin: codexBin,
		moaiBin: moaiBin, moaiHash: moaiHash, launchLog: launchLog, token: token,
		worktrees: worktrees, git: git, procs: procs, attempt: attempt}
}
