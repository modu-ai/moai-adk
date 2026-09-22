package cli

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCodexLocalInstructions_InjectedAsDeveloperInstructions(t *testing.T) {
	root := t.TempDir()
	wantBody := "# Codex 개인 지침\n\n- 따옴표: \\\"그대로\\\"\n- UTF-8을 보존한다.\n"
	localPath := filepath.Join(root, "AGENTS.local.md")
	if err := os.WriteFile(localPath, []byte(wantBody), 0o600); err != nil {
		t.Fatalf("write AGENTS.local.md: %v", err)
	}

	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	if _, _, err := runCodexCmd(t, "--", "--model", "o3"); err != nil {
		t.Fatalf("launch: %v", err)
	}

	got := codexArgvTail(t, cap)
	if len(got) < 2 || got[0] != "-c" {
		t.Fatalf("child argv = %#v, want a leading -c developer_instructions override", got)
	}
	encoded, ok := strings.CutPrefix(got[1], "developer_instructions=")
	if !ok {
		t.Fatalf("config override = %q, want developer_instructions=<TOML string>", got[1])
	}
	var body string
	if err := json.Unmarshal([]byte(encoded), &body); err != nil {
		t.Fatalf("developer_instructions value is not a JSON-compatible TOML string: %v", err)
	}
	if body != "<!-- source: AGENTS.local.md -->\n"+wantBody {
		t.Errorf("developer instructions differ:\n got %q\nwant %q", body, wantBody)
	}
	if wantTail := []string{"--model", "o3"}; !reflect.DeepEqual(got[2:], wantTail) {
		t.Errorf("operator tail = %#v, want %#v", got[2:], wantTail)
	}
	after, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("read AGENTS.local.md after launch: %v", err)
	}
	if string(after) != wantBody {
		t.Errorf("AGENTS.local.md was rewritten during launch")
	}
}

func TestCodexInstructionContract_DoesNotLinkClaudeLocalIntoAgents(t *testing.T) {
	root := t.TempDir()
	agents := []byte("# AGENTS.md\n\nshared contract\n")
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), agents, 0o600); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("# CLAUDE.md\n\n@AGENTS.md\n"), 0o600); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.local.md"), []byte("# Claude only\n"), 0o600); err != nil {
		t.Fatalf("write CLAUDE.local.md: %v", err)
	}

	if err := secureCodexInstructionContract(codexContractRequest{ProjectRoot: root}); err != nil {
		t.Fatalf("secure contract: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !reflect.DeepEqual(after, agents) {
		t.Errorf("AGENTS.md changed in response to CLAUDE.local.md:\n%s", after)
	}
}

func TestCodexLocalInstructions_DirectSpawnAndAppSharePrefix(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, codexLocalInstructionName), []byte("local\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	prevInTmux := inTmuxFn
	inTmuxFn = func() bool { return true }
	t.Cleanup(func() { inTmuxFn = prevInTmux })
	prevLookPath := spawnLookPath
	spawnLookPath = func(file string) (string, error) { return "/stub/bin/" + file, nil }
	t.Cleanup(func() { spawnLookPath = prevLookPath })
	for _, args := range [][]string{{"cli"}, {"app", "--spawn"}} {
		if _, _, err := runCodexCmd(t, args...); err != nil {
			t.Fatalf("run %v: %v", args, err)
		}
	}
	if len(cap.records) != 2 {
		t.Fatalf("launch records = %d, want 2", len(cap.records))
	}
	for i, rec := range cap.records {
		if len(rec.Argv) < 3 || rec.Argv[1] != "-c" || !strings.HasPrefix(rec.Argv[2], "developer_instructions=") {
			t.Errorf("record %d argv = %#v, want local instruction prefix", i, rec.Argv)
		}
	}
	if got := cap.records[1].Argv[3:]; !reflect.DeepEqual(got, []string{"app"}) {
		t.Errorf("app tail = %#v, want [app]", got)
	}
}

func TestCodexLocalInstructions_AbsentLeavesArgvUnchanged(t *testing.T) {
	root := t.TempDir()
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	if _, _, err := runCodexCmd(t, "--", "--model", "o3"); err != nil {
		t.Fatal(err)
	}
	if got := codexArgvTail(t, cap); !reflect.DeepEqual(got, []string{"--model", "o3"}) {
		t.Errorf("argv = %#v, want operator tail unchanged", got)
	}
}

func TestCodexLocalInstructions_SymlinkIsRefused(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(target, []byte("outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, codexLocalInstructionName)); err != nil {
		t.Fatal(err)
	}
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	_, _, err := runCodexCmd(t)
	if err == nil || !strings.Contains(err.Error(), "not a regular file (symlink)") {
		t.Fatalf("error = %v, want symlink refusal", err)
	}
	if cap.count() != 0 {
		t.Errorf("launches = %d, want 0 after refusal", cap.count())
	}
}

func TestCodexLocalInstructions_DocumentedInLauncherHelp(t *testing.T) {
	for _, want := range []string{"Common local guidance", "shared with Claude", "Codex-specific local", "developer instructions", "non-empty"} {
		if !strings.Contains(codexCmd.Long, want) {
			t.Errorf("launcher help does not mention %q", want)
		}
	}
	// The help is a shipped user-facing surface: describe the inputs without
	// enumerating local filenames, as the distributed template does.
	for _, name := range []string{codexClaudeLocalName, codexLocalInstructionName} {
		if strings.Contains(codexCmd.Long, name) {
			t.Errorf("launcher help enumerates local filename %q", name)
		}
	}
}

func localInstructionPayload(t *testing.T, args []string) string {
	t.Helper()
	if len(args) != 2 || args[0] != "-c" {
		t.Fatalf("want exactly one override pair, got %d args", len(args))
	}
	encoded, ok := strings.CutPrefix(args[1], "developer_instructions=")
	if !ok {
		t.Fatal("missing developer_instructions key")
	}
	var body string
	if err := json.Unmarshal([]byte(encoded), &body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestCodexLocalInstructions_DualFileMatrix(t *testing.T) {
	states := []string{"absent", "empty", "body"}
	for _, claude := range states {
		for _, agents := range states {
			t.Run(claude+"/"+agents, func(t *testing.T) {
				root := t.TempDir()
				want := ""
				for i, name := range []string{"CLAUDE.local.md", codexLocalInstructionName} {
					state := []string{claude, agents}[i]
					if state == "absent" {
						continue
					}
					body := ""
					if state == "body" {
						body = name + " instruction without final newline"
						if want != "" {
							want += "\n"
						}
						want += "<!-- source: " + name + " -->\n" + body
					}
					if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				args, err := codexLocalDeveloperInstructionArgs(root)
				if err != nil {
					t.Fatal(err)
				}
				if want == "" {
					if len(args) != 0 {
						t.Fatal("empty/absent inputs changed argv")
					}
					return
				}
				if got := localInstructionPayload(t, args); got != want {
					t.Fatalf("payload = %q, want %q", got, want)
				}
			})
		}
	}
}

func TestCodexLocalInstructions_LargeBodySlicesAndFreshRead(t *testing.T) {
	root := t.TempDir()
	prefix := "한글🙂<&>\\\"\n"
	first := prefix + strings.Repeat("x", 61360-len(prefix))
	second := "AGENTS_MARKER\n"
	for launch := range 2 {
		if launch == 1 {
			first = "updated Claude body"
			second = "updated agents body"
		}
		for i, name := range []string{"CLAUDE.local.md", codexLocalInstructionName} {
			if err := os.WriteFile(filepath.Join(root, name), []byte([]string{first, second}[i]), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		args, err := codexLocalDeveloperInstructionArgs(root)
		if err != nil {
			t.Fatal(err)
		}
		payload := localInstructionPayload(t, args)
		body, ok := strings.CutPrefix(payload, "<!-- source: CLAUDE.local.md -->\n")
		if !ok {
			t.Fatal("missing first provenance")
		}
		left, right, ok := strings.Cut(body, "\n<!-- source: AGENTS.local.md -->\n")
		if !ok || sha256.Sum256([]byte(left)) != sha256.Sum256([]byte(first)) || sha256.Sum256([]byte(right)) != sha256.Sum256([]byte(second)) {
			t.Fatal("body slice hash differs from source")
		}
	}
}

func TestCodexLocalInstructions_ExactlyThreeContractPaths(t *testing.T) {
	previous := codexInstructionRelPathsFn
	t.Cleanup(func() { codexInstructionRelPathsFn = previous })
	for _, count := range []int{2, 4} {
		codexInstructionRelPathsFn = func() []string { return make([]string, count) }
		err := secureCodexInstructionContract(codexContractRequest{ProjectRoot: t.TempDir()})
		if err == nil || err.Error() != fmt.Sprintf("instruction path table must name exactly three paths, got %d", count) {
			t.Fatalf("count %d: %v", count, err)
		}
	}
}

func TestCodexLocalInstructions_OperatorCollision(t *testing.T) {
	for _, form := range [][]string{
		{"--config", "developer_instructions=operator"}, {"--config=developer_instructions=operator"},
		{"-c", "developer_instructions=operator"}, {"-c=developer_instructions=operator"}, {"-cdeveloper_instructions=operator"},
	} {
		for _, present := range []bool{false, true} {
			t.Run(fmt.Sprint(form, present), func(t *testing.T) {
				root := t.TempDir()
				if present {
					if err := os.WriteFile(filepath.Join(root, codexClaudeLocalName), []byte("local"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				cap := withCodexLaunchCapture(t)
				withCodexProjectRoot(t, root)
				_, _, err := runCodexCmd(t, append([]string{"--"}, form...)...)
				if present {
					if err == nil || !strings.Contains(err.Error(), "duplicate developer_instructions override") || cap.count() != 0 {
						t.Fatalf("collision: err=%v launches=%d", err, cap.count())
					}
				} else if err != nil || cap.count() != 1 || !reflect.DeepEqual(codexArgvTail(t, cap), form) {
					t.Fatalf("operator-only value must survive: err=%v launches=%d", err, cap.count())
				}
			})
		}
	}
	for _, tail := range [][]string{{}, {"--config", "model=other"}, {"-cmodel=other"}, {"--config"}, {"--config", "developer_instructions_extra=x"}, {"--", "-cdeveloper_instructions=prompt"}} {
		t.Run(fmt.Sprint(tail), func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, codexClaudeLocalName), []byte("local"), 0o600); err != nil {
				t.Fatal(err)
			}
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			if _, _, err := runCodexCmd(t, append([]string{"--"}, tail...)...); err != nil || cap.count() != 1 {
				t.Fatalf("non-collision: %v", err)
			}
		})
	}
}

func TestCodexLocalInstructions_SizeBoundaries(t *testing.T) {
	const ceiling = 126976 // Independent contract pin, not an implementation default.
	for _, spawn := range []bool{false, true} {
		for _, delta := range []int{-1, 0, 1} {
			t.Run(fmt.Sprintf("spawn=%v/delta=%d", spawn, delta), func(t *testing.T) {
				root := t.TempDir()
				cap := withCodexLaunchCapture(t)
				withCodexProjectRoot(t, root)
				prevTmux, prevLook := inTmuxFn, spawnLookPath
				inTmuxFn = func() bool { return true }
				spawnLookPath = func(s string) (string, error) { return s, nil }
				t.Cleanup(func() { inTmuxFn, spawnLookPath = prevTmux, prevLook })
				path := filepath.Join(root, codexClaudeLocalName)
				if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
					t.Fatal(err)
				}
				seed, err := codexLocalDeveloperInstructionArgs(root)
				if err != nil {
					t.Fatal(err)
				}
				length := len(seed[1])
				if spawn {
					length = len(buildCodexSpawnCommand(sentinelCodexBinaryPath, seed))
				}
				if err := os.WriteFile(path, []byte(strings.Repeat("x", ceiling+delta-length+1)), 0o600); err != nil {
					t.Fatal(err)
				}
				args := []string{}
				if spawn {
					args = append(args, "--spawn")
				}
				_, _, err = runCodexCmd(t, args...)
				if delta > 0 {
					if err == nil || cap.count() != 0 || !strings.Contains(err.Error(), "126977") || !strings.Contains(err.Error(), "126976") {
						t.Fatalf("overflow: %v, launches=%d", err, cap.count())
					}
				} else if err != nil || cap.count() != 1 {
					t.Fatalf("within limit: %v, launches=%d", err, cap.count())
				}
			})
		}
	}
}

func TestCodexLocalInstructions_SpawnQuoteExpansion(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, codexClaudeLocalName), []byte(strings.Repeat("'", 40000)), 0o600); err != nil {
		t.Fatal(err)
	}
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	args, err := codexLocalDeveloperInstructionArgs(root)
	if err != nil {
		t.Fatal(err)
	}
	quoted := buildCodexSpawnCommand(sentinelCodexBinaryPath, args)
	t.Logf("direct token=%d bytes; final spawn=%d bytes", len(args[1]), len(quoted))
	if len(args[1]) >= 126976 || len(quoted) <= 126976 {
		t.Fatal("fixture does not straddle independent limits")
	}
	if _, _, err := runCodexCmd(t); err != nil {
		t.Fatalf("direct: %v", err)
	}
	prevTmux, prevLook := inTmuxFn, spawnLookPath
	inTmuxFn = func() bool { return true }
	spawnLookPath = func(s string) (string, error) { return s, nil }
	t.Cleanup(func() { inTmuxFn, spawnLookPath = prevTmux, prevLook })
	if _, _, err := runCodexCmd(t, "--spawn"); err == nil || cap.count() != 1 {
		t.Fatalf("spawn overflow: %v launches=%d", err, cap.count())
	}
}

func TestCodexLocalInstructions_BothFilesRejectNonRegular(t *testing.T) {
	for _, name := range []string{codexClaudeLocalName, codexLocalInstructionName} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, name), 0o700); err != nil {
				t.Fatal(err)
			}
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, root)
			_, _, err := runCodexCmd(t)
			var guard *codexPathGuardError
			if !errors.As(err, &guard) || guard.Rel != name || cap.count() != 0 {
				t.Fatalf("directory: %v", err)
			}
		})
	}
}

func TestCodexLocalInstructions_TemplateDescribesCommonAndSpecificInputs(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "template", "templates", "AGENTS.md.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	_, section, ok := strings.Cut(string(body), "## 8. Harness-local instructions")
	if !ok {
		t.Fatal("missing harness-local contract")
	}
	section, _, _ = strings.Cut(section, "## 9.")
	if strings.Contains(section, codexClaudeLocalName) || strings.Contains(section, codexLocalInstructionName) || !strings.Contains(section, "common local guidance") || !strings.Contains(section, "Codex-specific") {
		t.Fatal("template must describe common and Codex-specific inputs without enumerating local filenames")
	}
}

func TestCodexLocalInstructions_AllFunnelsPreserveInputs(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv("MOAI_HOME", t.TempDir())
	if err := os.MkdirAll(filepath.Join(root, ".claude", "worktrees", "fixture"), 0o700); err != nil {
		t.Fatal(err)
	}
	names := []string{codexClaudeLocalName, codexLocalInstructionName, codexAgentsRelPath, codexClaudeRelPath}
	before := make(map[string]os.FileInfo)
	bodies := make(map[string][]byte)
	for _, name := range names {
		path := filepath.Join(root, name)
		bodies[name] = []byte("unique body of " + name)
		if err := os.WriteFile(path, bodies[name], 0o600); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		before[name] = info
	}
	cap := withCodexLaunchCapture(t)
	withCodexProjectRoot(t, root)
	prevTmux, prevLook, prevRename := inTmuxFn, spawnLookPath, codexRenameFn
	inTmuxFn = func() bool { return true }
	spawnLookPath = func(s string) (string, error) { return s, nil }
	renames := 0
	codexRenameFn = func(a, b string) error { renames++; return os.Rename(a, b) }
	t.Cleanup(func() { inTmuxFn, spawnLookPath, codexRenameFn = prevTmux, prevLook, prevRename })
	forms := [][]string{{}, {"cli"}, {"app"}, {"--spawn"}, {"-w", "fixture"}, {"-f"}, {"-f", "agent"}, {"-f", "agent-2"}, {"-f", "lane-3"}}
	var want []string
	for i, args := range forms {
		if _, _, err := runCodexCmd(t, args...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		record := cap.records[i]
		if len(record.Argv) < 3 {
			t.Fatalf("%v: missing override", args)
		}
		prefix := record.Argv[1:3]
		if i == 0 {
			want = append([]string(nil), prefix...)
		} else if !reflect.DeepEqual(prefix, want) {
			t.Fatalf("%v: payload differs", args)
		}
		count := 0
		for _, arg := range record.Argv[1:] {
			if strings.HasPrefix(arg, "developer_instructions=") {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("%v: %d overrides", args, count)
		}
		for _, name := range names {
			path := filepath.Join(root, name)
			info, err := os.Lstat(path)
			if err != nil {
				t.Fatal(err)
			}
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !os.SameFile(before[name], info) || info.Mode()&os.ModeSymlink != 0 || sha256.Sum256(body) != sha256.Sum256(bodies[name]) {
				t.Fatalf("%v mutated %s", args, name)
			}
		}
	}
	if renames != 0 || cap.count() != len(forms) {
		t.Fatalf("renames=%d launches=%d", renames, cap.count())
	}
}

func TestCodexLocalInstructions_DefaultSpawnChecksActualCommand(t *testing.T) {
	previous := tmuxSpawnFn
	calls := 0
	tmuxSpawnFn = func(string, string) (string, error) { calls++; return "%1", nil }
	t.Cleanup(func() { tmuxSpawnFn = previous })
	err := defaultCodexSpawnLaunch(t.TempDir(), "codex", []string{strings.Repeat("'", 40000)})
	if err == nil || calls != 0 {
		t.Fatalf("overflow reached tmux: %v calls=%d", err, calls)
	}
}

func TestCodexLocalInstructions_PrelaunchFailuresStartNothing(t *testing.T) {
	for _, failure := range []string{"missing-binary", "missing-worktree"} {
		t.Run(failure, func(t *testing.T) {
			cap := withCodexLaunchCapture(t)
			withCodexProjectRoot(t, t.TempDir())
			args := []string{"-w", "nonexistent"}
			if failure == "missing-binary" {
				codexLookPath = func(string) (string, error) { return "", os.ErrNotExist }
				args = nil
			}
			if _, _, err := runCodexCmd(t, args...); err == nil || cap.count() != 0 {
				t.Fatalf("%s: %v launches=%d", failure, err, cap.count())
			}
		})
	}
}
