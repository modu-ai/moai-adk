//go:build !windows

package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

func preApprovalRunStartup(t *testing.T, preserveLedger bool) *preApprovalFixture {
	t.Helper()
	if os.Getenv(envCodexPreApprovalLive) != "1" {
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
	attempt := 1
	if preserveLedger {
		prior, readErr := preApprovalReadLedger(filepath.Join(evidenceDir, "ledger.json"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if !preApprovalM1aBaseline(evidenceDir, prior) {
			t.Fatal("M1-a startup ledger or raw evidence differs from the recorded baseline")
		}
		attempt, readErr = preApprovalNextAttempt(prior)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if attempt == 2 {
			for _, name := range []string{"evidence.json", "outcome.txt", "live.jsonl"} {
				if _, err := os.Stat(filepath.Join(evidenceDir, "discriminator", "attempt-1", name)); err != nil {
					t.Fatalf("attempt-1 archive required before retry: %s: %v", name, err)
				}
			}
		}
		priorStarts := 0
		for _, row := range prior {
			if row["kind"] == "startup" {
				priorStarts++
			}
		}
		if priorStarts+4 > preApprovalMaxStartupCalls {
			t.Fatal("startup invocation budget would be exceeded")
		}
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
	for _, args := range [][]string{{"-C", root, "add", "README"}, {"-C", root, "commit", "-q", "-m", "fixture"}} {
		if out, err := git(args...); err != nil {
			t.Fatalf("fixture git: %v: %s", err, out)
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
	var ledger []preApprovalStartupRow
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
		for _, name := range []string{"disc-control", "disc-treatment", "car010", "car011"} {
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
			if name == "car010" {
				user += "\n[mcp_servers.decoy]\ncommand = " + strconv.Quote(filepath.Join(binDir, "decoy")) + "\nargs = [\"mcp-server\"]\nenabled_tools = [\"spec_progress\"]\n"
			}
			if err := os.WriteFile(filepath.Join(codexHome, "config.toml"), []byte(user), 0o600); err != nil {
				t.Fatal(err)
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
			ledger = append(ledger, row)
			if !row.Passed {
				ledger = append(ledger, preApprovalStartupRow{Kind: "stop", Fixture: name, Reason: "startup MCP launch or non-model gate failed"})
				break
			}
			t.Logf("%s startup exit=%d moai=%d decoy=%d items=0", name, exitCode, moaiLaunches, decoyLaunches)
		}
		if pass == 0 {
			preApprovalWriteJSON(t, filepath.Join(evidenceDir, "discriminator", "export-manifest.json"), discManifest)
			for name, manifest := range carManifests {
				preApprovalWriteJSON(t, filepath.Join(evidenceDir, name, "export-manifest.json"), manifest)
			}
			if preserveLedger {
				for _, group := range []string{"discriminator", "car010", "car011"} {
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
	if len(ledger) != 4 {
		t.Fatalf("startup stopped after %d entries; see ledger.json", len(ledger))
	}
	if preserveLedger {
		for i := range ledger {
			group := ledger[i].Fixture
			if strings.HasPrefix(group, "disc-") {
				group = "discriminator"
			}
			current, err := os.ReadFile(filepath.Join(evidenceDir, group, "export-manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			var export preApprovalManifest
			if err := json.Unmarshal(current, &export); err != nil || export.ExportedNS <= 0 {
				t.Fatalf("invalid %s export manifest: %v", group, err)
			}
			ledger[i].ExportID = fmt.Sprintf("export-%d", export.ExportedNS)
			snapshot := filepath.Join(evidenceDir, group, "exports", ledger[i].ExportID, "export-manifest.json")
			body, err := os.ReadFile(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(body, current) {
				t.Fatal("attempt export differs from current manifest")
			}
			ledger[i].ExportSHA256 = preApprovalSHA(body)
		}
	}
	ledgerPath := filepath.Join(evidenceDir, "ledger.json")
	var rows []map[string]any
	if preserveLedger {
		rows, err = preApprovalReadLedger(ledgerPath)
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		priorStarts := 0
		for _, row := range rows {
			if row["kind"] == "startup" {
				priorStarts++
			}
		}
		if priorStarts+len(ledger) > preApprovalMaxStartupCalls {
			t.Fatal("startup invocation budget exceeded")
		}
	}
	for _, entry := range ledger {
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
		rows = append(rows, row)
	}
	if err := preApprovalWriteLedger(ledgerPath, rows); err != nil {
		t.Fatal(err)
	}
	return &preApprovalFixture{evidenceDir: evidenceDir, tempRoot: tempRoot, root: root,
		codexHome: filepath.Join(tempRoot, "codex-home"), codexBin: codexBin,
		moaiBin: moaiBin, moaiHash: moaiHash, launchLog: launchLog, token: token,
		worktrees: worktrees, git: git, procs: procs, attempt: attempt}
}
