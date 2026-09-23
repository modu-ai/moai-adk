//go:build darwin || linux

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

type operationalPromptWriter struct {
	terminal *operationalTerminal
	prompt   string
	writes   []string
}

func (w *operationalPromptWriter) Write(p []byte) (int, error) {
	w.writes = append(w.writes, string(p))
	if len(w.writes) == 1 {
		if string(p) != w.prompt {
			return 0, fmt.Errorf("first write=%q", p)
		}
		_, _ = w.terminal.Write([]byte("\x1b[1m›\x1b[22m " + w.prompt))
	}
	return len(p), nil
}

func TestFactoryOperationalFixtureUsesProductionInit(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(t.TempDir(), "moai")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/moai")
	cmd.Dir = "../.."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	env := append(os.Environ(), "PATH="+filepath.Dir(bin)+string(os.PathListSeparator)+os.Getenv("PATH"))
	env = operationalHookShellEnv(t, bin, env)
	if err := prepareOperationalProject(root, bin, env); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".moai/config/sections", ".codex/hooks.json", ".codex/config.toml", "AGENTS.md"} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Errorf("production init artifact %s: %v", path, err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(root, ".codex/hooks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	entries := doc.Hooks["UserPromptSubmit"]
	if len(entries) != 1 || len(entries[0].Hooks) != 1 {
		t.Fatalf("UserPromptSubmit hooks=%+v", entries)
	}
	want := "moai hook user-prompt-submit --harness codex"
	if got := entries[0].Hooks[0].Command; got != want {
		t.Fatalf("UserPromptSubmit command=%q want production command %q", got, want)
	}
	shell := launchEnvValue(env, "SHELL")
	resolve := exec.Command(shell, "-lc", "command -v moai")
	resolve.Env = env
	out, err := resolve.Output()
	if err != nil {
		t.Fatal(err)
	}
	got, err := filepath.EvalSymlinks(strings.TrimSpace(string(out)))
	if err != nil {
		t.Fatal(err)
	}
	wantBin, err := filepath.EvalSymlinks(bin)
	if err != nil {
		t.Fatal(err)
	}
	if got != wantBin {
		t.Fatalf("Codex login-shell hook resolved %q want built launcher %q", got, wantBin)
	}

	run := "run-codex-hook-process-boundary"
	if err := recordFactoryRunStart(root, run, "codex", ""); err != nil {
		t.Fatal(err)
	}
	start := homestate.CurrentProcessFingerprint()
	store, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	closeOnCleanup(t, "factory message broker", store)
	if _, err := store.RegisterLaunchPending(context.Background(), factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: run, Backend: "codex",
		Role: "lead", Slot: "lead", PID: os.Getpid(), ProcessStart: start,
	}); err != nil {
		t.Fatal(err)
	}
	hookEnv := make([]string, 0, len(env)+6)
	for _, item := range env {
		key, _, _ := strings.Cut(item, "=")
		switch key {
		case config.EnvMoaiKanbanID, config.EnvMoaiKanbanBackend, config.EnvMoaiFactoryWorker, config.EnvMoaiFactoryWorkers, config.EnvMoaiSessionPID, config.EnvClaudeProjectDir:
			continue
		}
		hookEnv = append(hookEnv, item)
	}
	hookEnv = append(hookEnv,
		config.EnvMoaiKanbanID+"="+run,
		config.EnvMoaiKanbanBackend+"=codex",
		config.EnvMoaiFactoryWorker+"=",
		config.EnvMoaiFactoryWorkers+"=1",
		config.EnvMoaiSessionPID+"="+fmt.Sprint(os.Getpid()),
		config.EnvClaudeProjectDir+"="+root,
	)
	payload, err := json.Marshal(map[string]any{
		"session_id": "01a0c977-bdb7-7013-81e2-3bc3a96269c3", "turn_id": "turn-live6",
		"cwd": root, "hook_event_name": "UserPromptSubmit", "model": "gpt-5.6-sol",
		"permission_mode": "default", "prompt": "Reply with exactly FACTORY_READY_1 and do not call any tool.",
		"transcript_path": nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	hookProcess := exec.Command(shell, "-lc", entries[0].Hooks[0].Command)
	hookProcess.Dir, hookProcess.Env, hookProcess.Stdin = root, hookEnv, bytes.NewReader(payload)
	hookOut, err := hookProcess.CombinedOutput()
	if err != nil {
		t.Fatalf("Codex UserPromptSubmit process boundary: %v: %s", err, hookOut)
	}
	if !bytes.Contains(hookOut, []byte("factory messaging bound")) {
		t.Fatalf("Codex UserPromptSubmit did not run built binding path: %s", hookOut)
	}
	bound, err := store.ResolveLane(context.Background(), "lead")
	if err != nil || bound.SessionUUID != "01a0c977-bdb7-7013-81e2-3bc3a96269c3" {
		t.Fatalf("process-boundary bind=%+v err=%v output=%s", bound, err, hookOut)
	}
}

func TestFactoryOperationalTrustBootstrap(t *testing.T) {
	if got := operationalBootstrapArgs(); !reflect.DeepEqual(got, []string{"codex"}) {
		t.Fatalf("bootstrap argv=%v", got)
	}
	for _, tc := range []struct {
		text             string
		dir, hooks, want bool
	}{
		{"Trusting hooks...\nAsk Codex to do anything", true, true, true},
		{"Ask Codex to do anything\nTrusting hooks...", true, true, false},
		{"Trusting hooks...\nAsk Codex to do anything", false, true, false},
		{"Trusting hooks...\nAsk Codex to do anything", true, false, false},
	} {
		if got := operationalBootstrapReady(tc.text, tc.dir, tc.hooks); got != tc.want {
			t.Errorf("ready=%v want=%v", got, tc.want)
		}
	}
	for _, fail := range []bool{false, true} {
		var order []string
		err := finishOperationalBootstrap(func() (bool, error) {
			order = append(order, "poll")
			if fail {
				return false, fmt.Errorf("trust failed")
			}
			return true, nil
		}, func() error { order = append(order, "stop"); return nil }, func() error { order = append(order, "no-broker"); return nil }, time.Second)
		want := []string{"poll", "stop", "no-broker"}
		if fail {
			want = []string{"poll", "stop"}
		}
		if !reflect.DeepEqual(order, want) || (err != nil) != fail {
			t.Fatalf("order=%v err=%v", order, err)
		}
	}
	t.Run("cleanup blocks measurement", func(t *testing.T) {
		checked := false
		err := finishOperationalBootstrap(func() (bool, error) { return true, nil }, func() error { return fmt.Errorf("owned child remains") }, func() error { checked = true; return nil }, time.Second)
		if err == nil || checked {
			t.Fatal("measurement allowed despite cleanup failure")
		}
	})
	t.Run("timeout cleans once", func(t *testing.T) {
		stops := 0
		err := finishOperationalBootstrap(func() (bool, error) { return false, nil }, func() error { stops++; return nil }, func() error { t.Fatal("measurement before readiness"); return nil }, 0)
		if err == nil || stops != 1 {
			t.Fatalf("err=%v stops=%d", err, stops)
		}
	})
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	if err := operationalNoBroker(root); err != nil {
		t.Fatal(err)
	}
	dir, err := homestate.FactoryDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "messages", "r"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "messages", "r", "broker.db"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if operationalNoBroker(root) == nil {
		t.Fatal("seeded broker accepted")
	}
}

func TestFactoryOperationalBootstrapCapturesOwnedMCPBeforeCleanup(t *testing.T) {
	var order []string
	pid, err := finishOperationalBootstrapCapture(
		func() (bool, error) { order = append(order, "poll"); return true, nil },
		func() (int, error) { order = append(order, "capture"); return 4242, nil },
		func() error { order = append(order, "stop"); return nil },
		func(got int) error {
			order = append(order, "dead-check")
			if got != 4242 {
				return fmt.Errorf("captured pid=%d", got)
			}
			return nil
		},
		time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}
	if pid != 4242 {
		t.Fatalf("pid=%d", pid)
	}
	if want := []string{"poll", "capture", "stop", "dead-check"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("order=%v want=%v", order, want)
	}

	t.Run("capture failure still cleans", func(t *testing.T) {
		stops := 0
		_, err := finishOperationalBootstrapCapture(
			func() (bool, error) { return true, nil },
			func() (int, error) { return 0, fmt.Errorf("MCP absent") },
			func() error { stops++; return nil },
			func(int) error { t.Fatal("post-cleanup check after capture failure"); return nil },
			time.Second,
		)
		if err == nil || stops != 1 {
			t.Fatalf("err=%v stops=%d", err, stops)
		}
	})
}

func TestFactoryOperationalBootstrapMCPBelongsToTerminalTree(t *testing.T) {
	processes := []byte(`
  10     1 script -q /dev/null
  20    10 codex
  21    20 /tmp/built/moai mcp-server
  30     1 codex
  31    30 /tmp/built/moai mcp-server
  40    10 /tmp/built/moai mcp-server
`)
	descendants := map[int]string{20: "codex-start", 21: "mcp-start", 40: "orphan-start"}
	if got := operationalSelectTerminalOwnedMCP(processes, descendants); got != 21 {
		t.Fatalf("selected pid=%d want terminal-descendant MCP with a terminal-descendant owner", got)
	}
	t.Run("non-Codex descendant owner rejected", func(t *testing.T) {
		processes := []byte("  10 1 script\n  50 10 sh\n  51 50 /tmp/built/moai mcp-server\n")
		descendants := map[int]string{50: "shell-start", 51: "mcp-start"}
		if got := operationalSelectTerminalOwnedMCP(processes, descendants); got != 0 {
			t.Fatalf("selected non-Codex-owned MCP pid=%d", got)
		}
	})
}

func TestFactoryOperationalPromptReadinessGate(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		want         bool
	}{
		{name: "loading", output: "OpenAI Codex\nmodel: loading\ndirectory: loading", want: false},
		{name: "status without composer", output: "OpenAI Codex\nmodel: gpt-5.6-sol\ndirectory: /tmp/project\nsession id: actual", want: false},
		{name: "ready after loading", output: "model: loading\ndirectory: loading\nmodel: gpt-5.6-sol\ndirectory: /tmp/project\nAsk Codex to do anything", want: true},
		{name: "old composer before hook trust", output: "Ask Codex to do anything\nTrusting hooks...", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := operationalPromptReady(tc.output); got != tc.want {
				t.Fatalf("ready=%v want=%v output=%q", got, tc.want, tc.output)
			}
		})
	}
}

func TestOperationalExecutableIdentityCanonicalizesPathAliases(t *testing.T) {
	realDir := t.TempDir()
	realBin := filepath.Join(realDir, "moai")
	if err := os.WriteFile(realBin, []byte("fixture"), 0700); err != nil {
		t.Fatal(err)
	}
	aliasRoot := t.TempDir()
	aliasDir := filepath.Join(aliasRoot, "bin-link")
	if err := os.Symlink(realDir, aliasDir); err != nil {
		t.Fatal(err)
	}
	aliasBin := filepath.Join(aliasDir, "moai")
	if !operationalExecutableIdentityEqual(aliasBin, realBin) {
		t.Fatalf("alias %q did not match canonical %q", aliasBin, realBin)
	}
	if !operationalLoadedExecutable([]byte("p123\nfcwd\nn"+realBin+"\n"), aliasBin) {
		t.Fatal("lsof record did not canonicalize executable alias")
	}
	other := filepath.Join(realDir, "other")
	if err := os.WriteFile(other, []byte("fixture"), 0700); err != nil {
		t.Fatal(err)
	}
	if operationalExecutableIdentityEqual(aliasBin, other) || operationalLoadedExecutable([]byte("n"+other+"\n"), aliasBin) {
		t.Fatal("unrelated executable accepted")
	}
}

func TestFactoryOperationalPromptReadinessProcessBoundary(t *testing.T) {
	terminal := &operationalTerminal{}
	cmd := exec.Command("/bin/sh", "-c", "printf 'model: loading\\ndirectory: loading\\n'; sleep 0.10; printf 'model: gpt-5.6-sol\\ndirectory: /tmp/project\\nAsk Codex to do anything\\n'")
	cmd.Stdout, cmd.Stderr = terminal, terminal
	started := time.Now()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if err := waitOperationalPromptReady([]*operationalTerminal{terminal}, 2*time.Second); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed < 75*time.Millisecond {
		t.Fatalf("loading frame was accepted as ready after %s", elapsed)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestFactoryOperationalPromptSubmitWaitsForActivityAndSettle(t *testing.T) {
	prompt := "Reply with exactly FACTORY_READY_1 and do not call any tool."
	terminal := &operationalTerminal{pid: os.Getpid()}
	writer := &operationalPromptWriter{terminal: terminal, prompt: prompt}
	terminal.stdin = writer
	started := time.Now()
	if err := submitOperationalPrompt(terminal, prompt, time.Second); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(writer.writes, []string{prompt, "\r"}) {
		t.Fatalf("writes=%q", writer.writes)
	}
	if elapsed := time.Since(started); elapsed < 450*time.Millisecond {
		t.Fatalf("CR sent before bounded settle interval: %s", elapsed)
	}
}

func TestFactoryOperationalHookTrustOnce(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".codex"), 0700); err != nil {
		t.Fatal(err)
	}
	hooks, err := codexwiring.RenderHooks(nil)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".codex/hooks.json")
	if err := os.WriteFile(path, hooks, 0600); err != nil {
		t.Fatal(err)
	}
	screen := "Hooks need review\n8 hooks are new or changed.\nHooks can run outside the sandbox after you trust them.\n› 1. Review hooks\n2. Trust all and continue\n3. Continue without trusting (hooks won't run)\nPress enter to confirm or esc to go back"
	for _, tc := range []struct {
		name, text, approved string
		want                 bool
	}{
		{"exact", screen, root, true},
		{"permission", "Allow command? Press enter to confirm", root, false},
		{"review only", "Review hooks", root, false},
		{"without trusting", strings.Replace(screen, "› 1. Review hooks", "1. Review hooks\n› 3. Continue without trusting", 1), root, false},
		{"foreign", screen, root + "-foreign", false},
		{"stale", screen + "\x1b[Jordinary model prompt", root, false},
		{"changed selection", screen + "\x1b[8;1H› 3. Continue without trusting", root, false},
		{"fragmented ANSI", strings.ReplaceAll(screen, " ", "\x1b[7;6H"), root, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := operationalHookTrust{operationalDirectoryTrust: operationalDirectoryTrust{root: root}}
			for _, piece := range []string{tc.text[:len(tc.text)/2], tc.text[len(tc.text)/2:]} {
				s.observe(piece)
			}
			seq := s.takeSequence(tc.approved)
			if (seq != "") != tc.want {
				t.Fatalf("sequence=%q want=%v", seq, tc.want)
			}
			if tc.want && seq != "\x1b[B\r" {
				t.Fatalf("must move selected option 1 down exactly once, then confirm: %q", seq)
			}
			if s.takeSequence(tc.approved) != "" {
				t.Fatal("second sequence allowed")
			}
		})
	}
	t.Run("single terminal write", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		closeOnCleanup(t, "pipe reader", r)
		closeOnCleanup(t, "pipe writer", w)
		term := operationalTerminal{stdin: w, trust: operationalDirectoryTrust{root: root, sent: true}, hookTrust: operationalHookTrust{operationalDirectoryTrust: operationalDirectoryTrust{root: root}}}
		_, _ = term.Write([]byte(screen))
		for i := 0; i < 2; i++ {
			if err := term.acceptFixtureHookTrust(); err != nil {
				t.Fatal(err)
			}
		}
		_, _ = w.Write([]byte("x"))
		_ = r.SetReadDeadline(time.Now().Add(time.Second))
		buf := make([]byte, 16)
		n, err := r.Read(buf)
		if err != nil || string(buf[:n]) != "\x1b[B\rx" {
			t.Fatalf("input=%q err=%v", buf[:n], err)
		}
		// Model the observed menu's initial selected index: exactly one Down
		// before CR selects Trust all (2), never Review (1) or skip hooks (3).
		selected := 1
		for input := string(buf[:n-1]); input != ""; {
			switch {
			case strings.HasPrefix(input, "\x1b[B"):
				selected++
				input = input[3:]
			case strings.HasPrefix(input, "\r"):
				if selected != 2 {
					t.Fatalf("confirmed option %d", selected)
				}
				input = input[1:]
			default:
				t.Fatalf("unexpected key bytes %q", input)
			}
		}
	})
	if err := os.WriteFile(path, append(hooks, ' '), 0600); err != nil {
		t.Fatal(err)
	}
	s := operationalHookTrust{operationalDirectoryTrust: operationalDirectoryTrust{root: root}}
	s.observe(screen)
	if s.takeSequence(root) != "" {
		t.Fatal("mismatched hooks accepted")
	}
}

func TestFactoryOperationalTerminalHasGeometry(t *testing.T) {
	cmd := exec.Command("script", operationalScriptArgs("/bin/stty", []string{"size"})...)
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "40 160") {
		t.Fatalf("terminal geometry=%q err=%v", out, err)
	}
}

func TestFactoryOperationalCleanupWaitsForExit(t *testing.T) {
	calls := 0
	probe := func(int) (string, homestate.ProcessIdentityState) {
		calls++
		if calls == 1 {
			return "owner", homestate.ProcessIdentityLive
		}
		return "", homestate.ProcessIdentityDead
	}
	if err := waitOperationalOwnedExit(map[int]string{42: "owner"}, probe, 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if calls < 2 {
		t.Fatal("cleanup did not wait for death")
	}
	live := func(int) (string, homestate.ProcessIdentityState) { return "owner", homestate.ProcessIdentityLive }
	if err := waitOperationalOwnedExit(map[int]string{42: "owner"}, live, 20*time.Millisecond); err == nil {
		t.Fatal("surviving child accepted")
	}
	unknown := func(int) (string, homestate.ProcessIdentityState) { return "", homestate.ProcessIdentityIndeterminate }
	if err := waitOperationalOwnedExit(map[int]string{42: "owner"}, unknown, 20*time.Millisecond); err == nil {
		t.Fatal("unknown child accepted")
	}
}

func TestFactoryOperationalDirectoryTrustOnce(t *testing.T) {
	root := t.TempDir()
	screen := "You are in " + root + "\nDo you trust the contents of this directory?\n› 1. Yes, continue\n2. No, quit\nPress enter to continue"
	for _, tc := range []struct {
		name, screen string
		want         bool
	}{
		{"exact", screen, true},
		{"ordinary", "What would you like to work on?", false},
		{"permission", "Would you like to run this command?\n› 1. Yes, continue\nPress enter to continue", false},
		{"foreign", strings.ReplaceAll(screen, root, root+"-other"), false},
		{"cleared permission", screen + "\x1b[JWould you like to run this command?\nPress enter to continue", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := operationalDirectoryTrust{root: root}
			state.observe(tc.screen)
			if got := state.takeEnter(); got != tc.want {
				t.Fatalf("enter=%v want=%v", got, tc.want)
			}
			if state.takeEnter() {
				t.Fatal("second Enter permitted")
			}
		})
	}
	t.Run("fragmented ANSI", func(t *testing.T) {
		state := operationalDirectoryTrust{root: root}
		ansi := strings.ReplaceAll(screen, " ", "\x1b[3;6H")
		for _, part := range []string{ansi[:21], ansi[21:47], ansi[47:]} {
			state.observe(part)
		}
		if !state.takeEnter() || state.takeEnter() {
			t.Fatal("fragmented signature must authorize exactly once")
		}
	})
	t.Run("terminal sends Enter only once", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		closeOnCleanup(t, "pipe reader", r)
		closeOnCleanup(t, "pipe writer", w)
		term := operationalTerminal{stdin: w, trust: operationalDirectoryTrust{root: root}}
		_, _ = term.Write([]byte("\x1b[6n" + screen))
		for i := 0; i < 2; i++ {
			if err := term.acceptFixtureDirectoryTrust(); err != nil {
				t.Fatal(err)
			}
		}
		_, _ = w.Write([]byte("x"))
		_ = r.SetReadDeadline(time.Now().Add(time.Second))
		buf := make([]byte, 8)
		n, err := r.Read(buf)
		if err != nil || string(buf[:n]) != "\rx" {
			t.Fatalf("terminal input=%q err=%v", buf[:n], err)
		}
	})
}
