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
	Reason          string            `json:"reason,omitempty"`
}

type preApprovalManifest struct {
	ExportedNS int64             `json:"exported_ns"`
	Files      map[string]string `json:"files"`
}

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

// TestCodexPreApprovalStartup measures MCP launch only. Its CODEX_HOME has no
// login file or auth environment, and the provider points at a closed local
// port. Any model item is a stop, never evidence of preflight success.
func TestCodexPreApprovalStartup(t *testing.T) {
	if os.Getenv(envCodexPreApprovalLive) != "1" {
		t.Skip("NOT_RUN " + envCodexPreApprovalLive + " is not 1")
	}
	evidenceDir := os.Getenv(envT1172EvidenceDir)
	if evidenceDir == "" {
		t.Skip("NOT_RUN " + envT1172EvidenceDir + " is empty")
	}
	if !filepath.IsAbs(evidenceDir) {
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
	prompt := []byte("Call codex_role_audit exactly once through the exec custom tool. Do not use shell commands or change files.\n")
	liveArgv := []byte("exec\n-C\n" + root + "\n--json\n--ask-for-approval\nnever\n")
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
						preApprovalExport(t, manifest, inputDir, label+"/argv.txt", liveArgv)
						preApprovalExport(t, manifest, inputDir, label+"/prompt.txt", prompt)
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
				DecoyLaunches: decoyLaunches, InputsSHA256: map[string]string{"codex_home_config": preApprovalSHA([]byte(user)),
					"project_config": preApprovalSHA([]byte(project)), "moai_binary": moaiHash,
					"argv": preApprovalSHA(liveArgv), "prompt": preApprovalSHA(prompt), "root_state": preApprovalSHA(rootState)},
			}
			ledger = append(ledger, row)
			startupDir := filepath.Join(evidenceDir, "startup")
			if err := os.MkdirAll(startupDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(startupDir, name+".jsonl"), out, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(startupDir, name+".err"), stderr, 0o600); err != nil {
				t.Fatal(err)
			}
			preApprovalWriteJSON(t, filepath.Join(startupDir, name+".launches.json"), map[string]int{"moai": moaiLaunches, "decoy": decoyLaunches})
			if strings.Contains(string(out), `"type":"item.`) || moaiLaunches == 0 ||
				(name == "car010" && decoyLaunches == 0) || end-start > int64(25*time.Second) ||
				strings.Contains(string(stderr), "Not inside a trusted directory") ||
				strings.Contains(string(stderr), "unknown configuration field") {
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
	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evidenceDir, "ledger.json"), append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if len(ledger) != 4 {
		t.Fatalf("startup stopped after %d entries; see ledger.json", len(ledger))
	}
}
