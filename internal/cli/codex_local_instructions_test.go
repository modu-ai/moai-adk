package cli

import (
	"encoding/json"
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
	if body != wantBody {
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
	for _, want := range []string{"AGENTS.local.md", "developer instructions", "Codex-only"} {
		if !strings.Contains(codexCmd.Long, want) {
			t.Errorf("launcher help does not mention %q", want)
		}
	}
}
