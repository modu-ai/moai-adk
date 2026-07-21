package cli

// Coverage for the init-path git auto-detection that replaced the git_mode /
// git_provider wizard questions: mode comes from whether any remote is
// configured, provider from the origin remote's host, and an explicit
// --git-mode / --git-provider flag always wins over detection.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// gitDetectInitRepo creates a git repository in a fresh temp dir and returns its path.
// Remotes are added by the caller.
func gitDetectInitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := exec.Command("git", "-C", dir, "init").Run(); err != nil {
		t.Skipf("git init unavailable: %v", err)
	}
	return dir
}

// gitAddRemote registers a remote on the repository at dir.
func gitAddRemote(t *testing.T, dir, name, url string) {
	t.Helper()
	if err := exec.Command("git", "-C", dir, "remote", "add", name, url).Run(); err != nil {
		t.Fatalf("git remote add %s %s: %v", name, url, err)
	}
}

// TestDetectGitConfig drives detection against real repositories.
func TestDetectGitConfig(t *testing.T) {
	tests := []struct {
		name         string
		remotes      map[string]string // remote name -> URL
		gitRepo      bool
		wantMode     string
		wantProvider string
	}{
		{
			name:         "non-git directory falls back to manual",
			gitRepo:      false,
			wantMode:     "manual",
			wantProvider: "github",
		},
		{
			name:         "git repo with no remote is manual",
			gitRepo:      true,
			wantMode:     "manual",
			wantProvider: "github",
		},
		{
			name:         "github https origin is personal + github",
			gitRepo:      true,
			remotes:      map[string]string{"origin": "https://github.com/modu-ai/moai-adk.git"},
			wantMode:     "personal",
			wantProvider: "github",
		},
		{
			name:         "github ssh origin is personal + github",
			gitRepo:      true,
			remotes:      map[string]string{"origin": "git@github.com:modu-ai/moai-adk.git"},
			wantMode:     "personal",
			wantProvider: "github",
		},
		{
			name:         "gitlab.com origin is personal + gitlab",
			gitRepo:      true,
			remotes:      map[string]string{"origin": "https://gitlab.com/group/proj.git"},
			wantMode:     "personal",
			wantProvider: "gitlab",
		},
		{
			name:         "self-hosted non-github origin is personal + gitlab",
			gitRepo:      true,
			remotes:      map[string]string{"origin": "git@git.example.com:team/proj.git"},
			wantMode:     "personal",
			wantProvider: "gitlab",
		},
		{
			name:         "remote present but not named origin keeps github default",
			gitRepo:      true,
			remotes:      map[string]string{"upstream": "https://gitlab.com/group/proj.git"},
			wantMode:     "personal",
			wantProvider: "github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dir string
			if tt.gitRepo {
				dir = gitDetectInitRepo(t)
			} else {
				dir = t.TempDir()
			}
			for name, url := range tt.remotes {
				gitAddRemote(t, dir, name, url)
			}

			mode, provider := detectGitConfig(dir)
			if mode != tt.wantMode {
				t.Errorf("mode = %q, want %q", mode, tt.wantMode)
			}
			if provider != tt.wantProvider {
				t.Errorf("provider = %q, want %q", provider, tt.wantProvider)
			}
		})
	}
}

// TestHostFromRemoteURL covers the remote-URL host parser directly, including
// the forms that are awkward to materialize as real remotes.
func TestHostFromRemoteURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"https", "https://github.com/u/r.git", "github.com"},
		{"https with userinfo", "https://user@github.com/u/r.git", "github.com"},
		{"scp-like ssh", "git@github.com:u/r.git", "github.com"},
		{"ssh scheme", "ssh://git@gitlab.example.com/u/r.git", "gitlab.example.com"},
		{"ssh scheme with port", "ssh://git@gitlab.example.com:2222/u/r.git", "gitlab.example.com"},
		{"git scheme", "git://github.com/u/r.git", "github.com"},
		{"local absolute path", "/srv/git/repo.git", ""},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := hostFromRemoteURL(tt.url); got != tt.want {
				t.Errorf("hostFromRemoteURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

// TestDetectGitConfig_RemoteListError falls back to manual when the git
// invocation itself fails (git absent, permission error, ...).
func TestDetectGitConfig_RemoteListError(t *testing.T) {
	orig := gitRemoteListFunc
	gitRemoteListFunc = func(string) (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { gitRemoteListFunc = orig })

	mode, provider := detectGitConfig(t.TempDir())
	if mode != "manual" {
		t.Errorf("mode = %q, want %q on git failure", mode, "manual")
	}
	if provider != "github" {
		t.Errorf("provider = %q, want %q on git failure", provider, "github")
	}
}

// TestDetectGitConfig_OriginURLError keeps the detected mode and falls back to
// the github provider default when origin's URL cannot be read.
func TestDetectGitConfig_OriginURLError(t *testing.T) {
	origList, origURL := gitRemoteListFunc, gitOriginURLFunc
	gitRemoteListFunc = func(string) (string, error) { return "origin\n", nil }
	gitOriginURLFunc = func(string) (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() {
		gitRemoteListFunc, gitOriginURLFunc = origList, origURL
	})

	mode, provider := detectGitConfig(t.TempDir())
	if mode != "personal" {
		t.Errorf("mode = %q, want %q", mode, "personal")
	}
	if provider != "github" {
		t.Errorf("provider = %q, want %q", provider, "github")
	}
}

// TestInitGitFlagOverridesDetection asserts the explicit --git-mode /
// --git-provider flags win over detection: the project dir has no remote (so
// detection yields manual+github) yet the persisted config carries the flags.
func TestInitGitFlagOverridesDetection(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	// The wizard no longer answers any git question on the init path.
	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string, _, _ bool) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{
			ProjectName:  "flag-proj",
			ModelPolicy:  "high",
			ReportFormat: "html+md",
		}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	projectDir := filepath.Join(t.TempDir(), "flag-proj")
	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	if err := cmd.Flags().Set("git-mode", "team"); err != nil {
		t.Fatalf("set --git-mode: %v", err)
	}
	if err := cmd.Flags().Set("git-provider", "gitlab"); err != nil {
		t.Fatalf("set --git-provider: %v", err)
	}

	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "git-strategy.yaml"))
	if err != nil {
		t.Fatalf("read git-strategy.yaml: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, `mode: "team"`) {
		t.Errorf("--git-mode=team must win over detected 'manual', got:\n%s", got)
	}
	if !strings.Contains(got, `provider: "gitlab"`) {
		t.Errorf("--git-provider=gitlab must win over detected 'github', got:\n%s", got)
	}
}

// TestInitGitDetectionFillsConfig asserts detection fills the persisted config
// when no git flag is supplied: a repo with a github origin yields
// personal + github without the wizard asking anything.
func TestInitGitDetectionFillsConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string, _, _ bool) (*wizard.WizardResult, error) {
		return &wizard.WizardResult{
			ProjectName:  "detect-proj",
			ModelPolicy:  "high",
			ReportFormat: "html+md",
		}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	// Initialize the project dir as a git repo with a github origin BEFORE
	// running init, so detection sees the remote.
	projectDir := filepath.Join(t.TempDir(), "detect-proj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := exec.Command("git", "-C", projectDir, "init").Run(); err != nil {
		t.Skipf("git init unavailable: %v", err)
	}
	gitAddRemote(t, projectDir, "origin", "https://github.com/modu-ai/moai-adk.git")

	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, ".moai", "config", "sections", "git-strategy.yaml"))
	if err != nil {
		t.Fatalf("read git-strategy.yaml: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, `mode: "personal"`) {
		t.Errorf("configured remote must detect mode 'personal', got:\n%s", got)
	}
	if !strings.Contains(got, `provider: "github"`) {
		t.Errorf("github origin must detect provider 'github', got:\n%s", got)
	}
}
