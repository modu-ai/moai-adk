package worktree

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectName_GoMod(t *testing.T) {
	tests := []struct {
		name       string
		moduleLine string
		want       string
	}{
		{
			name:       "multi-segment module path",
			moduleLine: "module github.com/modu-ai/moai-adk\n",
			want:       "moai-adk",
		},
		{
			name:       "single-segment module name",
			moduleLine: "module myproject\n",
			want:       "myproject",
		},
		{
			name:       "three-segment module path",
			moduleLine: "module github.com/org/sub/project\n",
			want:       "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			content := tt.moduleLine + "\ngo 1.21\n"
			if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			got := detectProjectName(dir)
			if got != tt.want {
				t.Errorf("detectProjectName = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectProjectName_GitRemote(t *testing.T) {
	tests := []struct {
		name   string
		remote string
		want   string
	}{
		{
			name:   "SSH remote with .git suffix",
			remote: "git@github.com:modu-ai/some-repo.git",
			want:   "some-repo",
		},
		{
			name:   "HTTPS remote with .git suffix",
			remote: "https://github.com/modu-ai/another-repo.git",
			want:   "another-repo",
		},
		{
			name:   "HTTPS remote without .git suffix",
			remote: "https://github.com/modu-ai/clean-repo",
			want:   "clean-repo",
		},
	}

	origGitRemote := gitRemoteFunc
	defer func() { gitRemoteFunc = origGitRemote }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			// no go.mod, so fallback to git remote
			gitRemoteFunc = func(_ string) (string, error) {
				return tt.remote, nil
			}
			got := detectProjectName(dir)
			if got != tt.want {
				t.Errorf("detectProjectName = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectProjectName_Fallback(t *testing.T) {
	dir := t.TempDir()

	origGitRemote := gitRemoteFunc
	defer func() { gitRemoteFunc = origGitRemote }()
	gitRemoteFunc = func(_ string) (string, error) {
		return "", errors.New("no remote configured")
	}

	got := detectProjectName(dir)
	want := filepath.Base(dir)
	if got != want {
		t.Errorf("detectProjectName = %q, want %q", got, want)
	}
}
