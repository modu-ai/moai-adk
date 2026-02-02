package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// --- NewRootCommand ---

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	if cmd == nil {
		t.Fatal("NewRootCommand() returned nil")
	}

	if cmd.Use != "moai" {
		t.Errorf("root command Use = %s, want 'moai'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("root command Short is empty")
	}

	if cmd.Long == "" {
		t.Error("root command Long is empty")
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	cmd := NewRootCommand()

	expectedCommands := []string{
		"init",
		"doctor",
		"status",
		"update",
		"statusline",
		"hook",
		"version",
		"self-update",
		"migrate",
		"claude",
		"glm",
		"worktree",
		"rank",
	}

	for _, expected := range expectedCommands {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == expected {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("root command missing subcommand '%s'", expected)
		}
	}
}

func TestRootCommandSubcommandCount(t *testing.T) {
	cmd := NewRootCommand()

	// Should have 12 subcommands registered in root.go
	// plus built-in help and completion commands from cobra
	subCmds := cmd.Commands()
	if len(subCmds) < 12 {
		t.Errorf("expected at least 12 subcommands, got %d", len(subCmds))
	}
}

func TestRootCommandVersionFlag(t *testing.T) {
	cmd := NewRootCommand()

	if cmd.Version == "" {
		t.Error("root command has empty version")
	}
}

func TestSubcommandsHaveHelp(t *testing.T) {
	rootCmd := NewRootCommand()

	for _, subCmd := range rootCmd.Commands() {
		if subCmd.Short == "" {
			t.Errorf("subcommand '%s' has empty Short description", subCmd.Name())
		}

		if subCmd.Long == "" {
			t.Errorf("subcommand '%s' has empty Long description", subCmd.Name())
		}
	}
}

// --- Hook command ---

func TestHookCommandRequiresArgument(t *testing.T) {
	rootCmd := NewRootCommand()

	var hookCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "hook" {
			hookCmd = cmd
			break
		}
	}

	if hookCmd == nil {
		t.Fatal("hook command not found")
	}

	if hookCmd.Args == nil {
		t.Error("hook command has no Args validator")
	}
}

// --- Version command ---

func TestVersionCommandOutput(t *testing.T) {
	rootCmd := NewRootCommand()

	var versionCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			versionCmd = cmd
			break
		}
	}

	if versionCmd == nil {
		t.Fatal("version command not found")
	}

	if versionCmd.RunE == nil {
		t.Error("version command has no RunE function")
	}
}

func TestVersionCommandHasJSONFlag(t *testing.T) {
	rootCmd := NewRootCommand()

	var versionCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			versionCmd = cmd
			break
		}
	}

	if versionCmd == nil {
		t.Fatal("version command not found")
	}

	jsonFlag := versionCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Fatal("version command missing --json flag")
	}

	if jsonFlag.Shorthand != "j" {
		t.Errorf("json flag shorthand = %q, want %q", jsonFlag.Shorthand, "j")
	}
}

func TestVersionCommandRunE(t *testing.T) {
	rootCmd := NewRootCommand()

	var versionCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			versionCmd = cmd
			break
		}
	}

	if versionCmd == nil {
		t.Fatal("version command not found")
	}

	// Capture output
	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)

	// Execute version command (text mode)
	err := versionCmd.RunE(versionCmd, []string{})
	if err != nil {
		t.Fatalf("version RunE error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("version command produced no output")
	}
}

func TestVersionCommandJSONOutput(t *testing.T) {
	rootCmd := NewRootCommand()

	var versionCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			versionCmd = cmd
			break
		}
	}

	if versionCmd == nil {
		t.Fatal("version command not found")
	}

	// Set JSON flag
	if err := versionCmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("failed to set json flag: %v", err)
	}

	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)

	err := versionCmd.RunE(versionCmd, []string{})
	if err != nil {
		t.Fatalf("version RunE error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("version JSON command produced no output")
	}
	// JSON output should start with "{"
	if len(output) > 0 && output[0] != '{' {
		t.Errorf("JSON output should start with '{', got %q", output[:1])
	}
}

// --- Init command ---

func TestNewInitCommand(t *testing.T) {
	cmd := NewInitCommand()

	if cmd == nil {
		t.Fatal("NewInitCommand() returned nil")
	}
	if cmd.Use != "init [path]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "init [path]")
	}
	if cmd.RunE == nil {
		t.Error("init command has no RunE function")
	}
}

func TestInitCommandForceFlag(t *testing.T) {
	cmd := NewInitCommand()

	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Fatal("init command missing --force flag")
	}
	if forceFlag.Shorthand != "f" {
		t.Errorf("force flag shorthand = %q, want %q", forceFlag.Shorthand, "f")
	}
	if forceFlag.DefValue != "false" {
		t.Errorf("force flag default = %q, want %q", forceFlag.DefValue, "false")
	}
}

func TestInitCommandMaxArgs(t *testing.T) {
	cmd := NewInitCommand()

	// Should accept 0 or 1 arguments
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("init should accept 0 args, got error: %v", err)
	}
	if err := cmd.Args(cmd, []string{"./path"}); err != nil {
		t.Errorf("init should accept 1 arg, got error: %v", err)
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err == nil {
		t.Error("init should reject 2 args")
	}
}

// --- Doctor command ---

func TestNewDoctorCommand(t *testing.T) {
	cmd := NewDoctorCommand()

	if cmd == nil {
		t.Fatal("NewDoctorCommand() returned nil")
	}
	if cmd.Use != "doctor" {
		t.Errorf("Use = %q, want %q", cmd.Use, "doctor")
	}
	if cmd.RunE == nil {
		t.Error("doctor command has no RunE function")
	}
}

// --- Status command ---

func TestNewStatusCommand(t *testing.T) {
	cmd := NewStatusCommand()

	if cmd == nil {
		t.Fatal("NewStatusCommand() returned nil")
	}
	if cmd.Use != "status" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status")
	}
	if cmd.RunE == nil {
		t.Error("status command has no RunE function")
	}
}

// --- Update command ---

func TestNewUpdateCommand(t *testing.T) {
	cmd := NewUpdateCommand()

	if cmd == nil {
		t.Fatal("NewUpdateCommand() returned nil")
	}
	if cmd.Use != "update" {
		t.Errorf("Use = %q, want %q", cmd.Use, "update")
	}
	if cmd.RunE == nil {
		t.Error("update command has no RunE function")
	}
}

func TestUpdateCommandDryRunFlag(t *testing.T) {
	cmd := NewUpdateCommand()

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("update command missing --dry-run flag")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("dry-run flag default = %q, want %q", dryRunFlag.DefValue, "false")
	}
}

// --- Statusline command ---

func TestNewStatuslineCommand(t *testing.T) {
	cmd := NewStatuslineCommand()

	if cmd == nil {
		t.Fatal("NewStatuslineCommand() returned nil")
	}
	if cmd.Use != "statusline" {
		t.Errorf("Use = %q, want %q", cmd.Use, "statusline")
	}
	if cmd.RunE == nil {
		t.Error("statusline command has no RunE function")
	}
}

func TestStatuslineCommandFormatFlag(t *testing.T) {
	cmd := NewStatuslineCommand()

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("statusline command missing --format flag")
	}
	if formatFlag.DefValue != "" {
		t.Errorf("format flag default = %q, want empty", formatFlag.DefValue)
	}
}

// --- Self-update command ---

func TestNewSelfUpdateCommand(t *testing.T) {
	cmd := NewSelfUpdateCommand()

	if cmd == nil {
		t.Fatal("NewSelfUpdateCommand() returned nil")
	}
	if cmd.Use != "self-update" {
		t.Errorf("Use = %q, want %q", cmd.Use, "self-update")
	}
	if cmd.RunE == nil {
		t.Error("self-update command has no RunE function")
	}
}

func TestSelfUpdateCommandFlags(t *testing.T) {
	cmd := NewSelfUpdateCommand()

	checkOnlyFlag := cmd.Flags().Lookup("check-only")
	if checkOnlyFlag == nil {
		t.Error("self-update command missing --check-only flag")
	}

	versionFlag := cmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Error("self-update command missing --version flag")
	}
}

// --- Migrate command ---

func TestNewMigrateCommand(t *testing.T) {
	cmd := NewMigrateCommand()

	if cmd == nil {
		t.Fatal("NewMigrateCommand() returned nil")
	}
	if cmd.Use != "migrate" {
		t.Errorf("Use = %q, want %q", cmd.Use, "migrate")
	}
	if cmd.RunE == nil {
		t.Error("migrate command has no RunE function")
	}
}

func TestMigrateCommandFlags(t *testing.T) {
	cmd := NewMigrateCommand()

	expectedFlags := []string{
		"dry-run",
		"force-go",
		"force-python",
		"go-binary-path",
		"rollback",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("migrate command missing --%s flag", flagName)
		}
	}
}

// --- Claude/GLM commands ---

func TestNewClaudeCommand(t *testing.T) {
	cmd := NewClaudeCommand()

	if cmd == nil {
		t.Fatal("NewClaudeCommand() returned nil")
	}
	if cmd.Use != "claude" {
		t.Errorf("Use = %q, want %q", cmd.Use, "claude")
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "cc" {
		t.Errorf("Aliases = %v, want [cc]", cmd.Aliases)
	}
}

func TestNewGLMCommand(t *testing.T) {
	cmd := NewGLMCommand()

	if cmd == nil {
		t.Fatal("NewGLMCommand() returned nil")
	}
	if cmd.Use != "glm [api-key]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "glm [api-key]")
	}
}

func TestGLMCommandMaxArgs(t *testing.T) {
	cmd := NewGLMCommand()

	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("glm should accept 0 args, got error: %v", err)
	}
	if err := cmd.Args(cmd, []string{"key123"}); err != nil {
		t.Errorf("glm should accept 1 arg, got error: %v", err)
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err == nil {
		t.Error("glm should reject 2 args")
	}
}

// --- Worktree command ---

func TestNewWorktreeCommand(t *testing.T) {
	cmd := NewWorktreeCommand()

	if cmd == nil {
		t.Fatal("NewWorktreeCommand() returned nil")
	}
	if cmd.Use != "worktree" {
		t.Errorf("Use = %q, want %q", cmd.Use, "worktree")
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "wt" {
		t.Errorf("Aliases = %v, want [wt]", cmd.Aliases)
	}
}

func TestWorktreeSubcommands(t *testing.T) {
	cmd := NewWorktreeCommand()

	expectedSubcmds := []string{
		"new",
		"list",
		"go",
		"done",
		"remove",
		"sync",
		"clean",
		"recover",
		"status",
	}

	for _, expected := range expectedSubcmds {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("worktree command missing subcommand '%s'", expected)
		}
	}
}

func TestWorktreeSubcommandCount(t *testing.T) {
	cmd := NewWorktreeCommand()

	subCmds := cmd.Commands()
	if len(subCmds) != 9 {
		t.Errorf("expected 9 worktree subcommands, got %d", len(subCmds))
	}
}

func TestWorktreeNewCmdFlags(t *testing.T) {
	cmd := NewWorktreeCommand()

	var newCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "new" {
			newCmd = sub
			break
		}
	}

	if newCmd == nil {
		t.Fatal("worktree new subcommand not found")
	}

	// Check flags
	branchFlag := newCmd.Flags().Lookup("branch")
	if branchFlag == nil {
		t.Error("worktree new missing --branch flag")
	}
	if branchFlag != nil && branchFlag.Shorthand != "b" {
		t.Errorf("branch flag shorthand = %q, want %q", branchFlag.Shorthand, "b")
	}

	baseFlag := newCmd.Flags().Lookup("base")
	if baseFlag == nil {
		t.Error("worktree new missing --base flag")
	}
	if baseFlag != nil && baseFlag.DefValue != "main" {
		t.Errorf("base flag default = %q, want %q", baseFlag.DefValue, "main")
	}

	forceFlag := newCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("worktree new missing --force flag")
	}
}

func TestWorktreeListCmdJSONFlag(t *testing.T) {
	cmd := NewWorktreeCommand()

	var listCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "list" {
			listCmd = sub
			break
		}
	}

	if listCmd == nil {
		t.Fatal("worktree list subcommand not found")
	}

	jsonFlag := listCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("worktree list missing --json flag")
	}
}

// --- Rank command ---

func TestNewRankCommand(t *testing.T) {
	cmd := NewRankCommand()

	if cmd == nil {
		t.Fatal("NewRankCommand() returned nil")
	}
	if cmd.Use != "rank" {
		t.Errorf("Use = %q, want %q", cmd.Use, "rank")
	}
}

func TestRankSubcommands(t *testing.T) {
	cmd := NewRankCommand()

	expectedSubcmds := []string{
		"login",
		"status",
		"logout",
		"exclude",
		"include",
		"sync",
	}

	for _, expected := range expectedSubcmds {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("rank command missing subcommand '%s'", expected)
		}
	}
}

func TestRankSubcommandCount(t *testing.T) {
	cmd := NewRankCommand()

	subCmds := cmd.Commands()
	if len(subCmds) != 6 {
		t.Errorf("expected 6 rank subcommands, got %d", len(subCmds))
	}
}

func TestRankLoginHasAliases(t *testing.T) {
	cmd := NewRankCommand()

	var loginCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "login" {
			loginCmd = sub
			break
		}
	}

	if loginCmd == nil {
		t.Fatal("rank login subcommand not found")
	}

	found := false
	for _, alias := range loginCmd.Aliases {
		if alias == "register" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("rank login aliases = %v, should contain 'register'", loginCmd.Aliases)
	}
}

func TestRankLoginAPIKeyFlag(t *testing.T) {
	cmd := NewRankCommand()

	var loginCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "login" {
			loginCmd = sub
			break
		}
	}

	if loginCmd == nil {
		t.Fatal("rank login subcommand not found")
	}

	apiKeyFlag := loginCmd.Flags().Lookup("api-key")
	if apiKeyFlag == nil {
		t.Error("rank login missing --api-key flag")
	}
}

func TestRankExcludeListFlag(t *testing.T) {
	cmd := NewRankCommand()

	var excludeCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "exclude" {
			excludeCmd = sub
			break
		}
	}

	if excludeCmd == nil {
		t.Fatal("rank exclude subcommand not found")
	}

	listFlag := excludeCmd.Flags().Lookup("list")
	if listFlag == nil {
		t.Error("rank exclude missing --list flag")
	}
}

// --- compareVersions ---

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"v1 greater major", "2.0.0", "1.0.0", 1},
		{"v1 lesser major", "1.0.0", "2.0.0", -1},
		{"v1 greater minor", "1.2.0", "1.1.0", 1},
		{"v1 lesser minor", "1.1.0", "1.2.0", -1},
		{"v1 greater patch", "1.0.2", "1.0.1", 1},
		{"v1 lesser patch", "1.0.1", "1.0.2", -1},
		{"different length v1 shorter", "1.0", "1.0.1", -1},
		{"different length v2 shorter", "1.0.1", "1.0", 1},
		{"zeros equal", "0.0.0", "0.0.0", 0},
		{"large versions", "10.20.30", "10.20.29", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

// --- toKebabCase ---

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PreToolUse", "pre-tool-use"},
		{"PostToolUse", "post-tool-use"},
		{"SessionStart", "session-start"},
		{"SessionEnd", "session-end"},
		{"PreCompact", "pre-compact"},
		{"Notification", "notification"},
		{"simple", "simple"},
		{"A", "a"},
		{"AB", "a-b"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toKebabCase(tt.input)
			if got != tt.want {
				t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- getCurrentVersion ---

func TestGetCurrentVersion_Default(t *testing.T) {
	// Unset MOAI_VERSION to get default
	orig := os.Getenv("MOAI_VERSION")
	if err := os.Unsetenv("MOAI_VERSION"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if orig != "" {
			if err := os.Setenv("MOAI_VERSION", orig); err != nil {
				t.Fatal(err)
			}
		}
	}()

	v := getCurrentVersion()
	if v != "dev" {
		t.Errorf("getCurrentVersion() = %q, want %q", v, "dev")
	}
}

func TestGetCurrentVersion_FromEnv(t *testing.T) {
	orig := os.Getenv("MOAI_VERSION")
	if err := os.Setenv("MOAI_VERSION", "1.5.0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if orig != "" {
			if err := os.Setenv("MOAI_VERSION", orig); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := os.Unsetenv("MOAI_VERSION"); err != nil {
				t.Fatal(err)
			}
		}
	}()

	v := getCurrentVersion()
	if v != "1.5.0" {
		t.Errorf("getCurrentVersion() = %q, want %q", v, "1.5.0")
	}
}

// --- GitHubRelease struct ---

func TestGitHubReleaseStruct(t *testing.T) {
	release := GitHubRelease{
		TagName:     "v1.0.0",
		Name:        "Release 1.0.0",
		Body:        "Release notes",
		Draft:       false,
		Prerelease:  false,
		PublishedAt: "2025-01-01T00:00:00Z",
	}

	if release.TagName != "v1.0.0" {
		t.Errorf("TagName = %q", release.TagName)
	}
	if release.Name != "Release 1.0.0" {
		t.Errorf("Name = %q", release.Name)
	}
	if release.Draft {
		t.Error("Draft should be false")
	}
	if release.Prerelease {
		t.Error("Prerelease should be false")
	}
}

// --- HookChange struct ---

func TestHookChangeStruct(t *testing.T) {
	change := HookChange{
		Hook: "SessionStart",
		Old:  "old-command",
		New:  "new-command",
	}

	if change.Hook != "SessionStart" {
		t.Errorf("Hook = %q", change.Hook)
	}
	if change.Old != "old-command" {
		t.Errorf("Old = %q", change.Old)
	}
	if change.New != "new-command" {
		t.Errorf("New = %q", change.New)
	}
}

// --- SettingsJSON struct ---

func TestSettingsJSONStruct(t *testing.T) {
	settings := SettingsJSON{
		HookCommands: map[string]string{
			"SessionStart": "moai hook session-start",
		},
	}

	if settings.HookCommands["SessionStart"] != "moai hook session-start" {
		t.Errorf("HookCommands[SessionStart] = %q", settings.HookCommands["SessionStart"])
	}
}

// --- HookEntry struct ---

func TestHookEntryStruct(t *testing.T) {
	entry := HookEntry{
		Type:    "command",
		Command: "echo hello",
	}

	if entry.Type != "command" {
		t.Errorf("Type = %q", entry.Type)
	}
	if entry.Command != "echo hello" {
		t.Errorf("Command = %q", entry.Command)
	}
}

// --- All subcommands have RunE ---

func TestAllSubcommandsHaveRunE(t *testing.T) {
	rootCmd := NewRootCommand()

	for _, subCmd := range rootCmd.Commands() {
		// Parent commands with subcommands may not have RunE
		if len(subCmd.Commands()) > 0 {
			continue
		}
		if subCmd.RunE == nil && subCmd.Run == nil {
			t.Errorf("subcommand '%s' has neither RunE nor Run function", subCmd.Name())
		}
	}
}

// --- Worktree done command flags ---

func TestWorktreeDoneCmdFlags(t *testing.T) {
	cmd := NewWorktreeCommand()

	var doneCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "done" {
			doneCmd = sub
			break
		}
	}

	if doneCmd == nil {
		t.Fatal("worktree done subcommand not found")
	}

	baseFlag := doneCmd.Flags().Lookup("base")
	if baseFlag == nil {
		t.Error("worktree done missing --base flag")
	}
	if baseFlag != nil && baseFlag.DefValue != "main" {
		t.Errorf("base flag default = %q, want %q", baseFlag.DefValue, "main")
	}

	pushFlag := doneCmd.Flags().Lookup("push")
	if pushFlag == nil {
		t.Error("worktree done missing --push flag")
	}

	forceFlag := doneCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("worktree done missing --force flag")
	}
}

// --- Worktree sync command flags ---

func TestWorktreeSyncCmdFlags(t *testing.T) {
	cmd := NewWorktreeCommand()

	var syncCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "sync" {
			syncCmd = sub
			break
		}
	}

	if syncCmd == nil {
		t.Fatal("worktree sync subcommand not found")
	}

	baseFlag := syncCmd.Flags().Lookup("base")
	if baseFlag == nil {
		t.Error("worktree sync missing --base flag")
	}

	rebaseFlag := syncCmd.Flags().Lookup("rebase")
	if rebaseFlag == nil {
		t.Error("worktree sync missing --rebase flag")
	}
}

// --- Worktree clean command flags ---

func TestWorktreeCleanCmdFlags(t *testing.T) {
	cmd := NewWorktreeCommand()

	var cleanCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == "clean" {
			cleanCmd = sub
			break
		}
	}

	if cleanCmd == nil {
		t.Fatal("worktree clean subcommand not found")
	}

	mergedFlag := cleanCmd.Flags().Lookup("merged-only")
	if mergedFlag == nil {
		t.Error("worktree clean missing --merged-only flag")
	}
}
