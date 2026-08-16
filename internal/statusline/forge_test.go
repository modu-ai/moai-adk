package statusline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Forge selection is the one part of the counts feature that can be verified
// without the forge: which command gets built, and from what. The fetch itself
// needs the CLI and a repository on that host, so these tests cover everything
// up to the exec boundary and stop there deliberately.

func TestRemoteHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"https", "https://github.com/modu-ai/moai-adk.git", "github.com"},
		{"https no suffix", "https://gitlab.com/group/proj", "gitlab.com"},
		{"ssh scheme", "ssh://git@github.com/o/r.git", "github.com"},
		{"ssh scheme with port", "ssh://git@git.example.com:2222/o/r.git", "git.example.com"},
		// scp-like: url.Parse reads this as a path with no host, so it is
		// split by hand. This is the default form `git clone git@...` writes.
		{"scp-like", "git@github.com:modu-ai/moai-adk.git", "github.com"},
		{"scp-like no user", "github.com:o/r.git", "github.com"},
		{"uppercase is normalised", "https://GitHub.COM/o/r", "github.com"},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"no host at all", "/local/path/repo.git", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := remoteHost(tt.raw); got != tt.want {
				t.Errorf("remoteHost(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestForgeFromRemote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw      string
		wantName string
		wantOK   bool
	}{
		{"https://github.com/o/r.git", "github", true},
		{"git@gitlab.com:g/p.git", "gitlab", true},
		// A self-hosted instance carries no signal in its name: a company
		// GitLab and a company GitHub Enterprise look identical here, so
		// neither is guessed and the config override decides.
		{"https://git.example.com/g/p.git", "", false},
		{"https://bitbucket.org/o/r.git", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()
			got, ok := forgeFromRemote(tt.raw)
			if ok != tt.wantOK {
				t.Fatalf("forgeFromRemote(%q) ok = %v, want %v", tt.raw, ok, tt.wantOK)
			}
			if ok && got.name != tt.wantName {
				t.Errorf("forgeFromRemote(%q) = %q, want %q", tt.raw, got.name, tt.wantName)
			}
		})
	}
}

func TestResolveForge(t *testing.T) {
	t.Parallel()

	const ghRemote = "https://github.com/o/r.git"

	tests := []struct {
		name     string
		remote   string
		override string
		wantName string
		wantOK   bool
	}{
		{"detection when no override", ghRemote, "", "github", true},
		{"override beats detection", ghRemote, "gitlab", "gitlab", true},
		{"override is case insensitive", ghRemote, "GitLab", "gitlab", true},
		{"override is trimmed", ghRemote, "  gitlab  ", "gitlab", true},
		{"none disables the segment", ghRemote, "none", "", false},
		{"off disables the segment", ghRemote, "off", "", false},
		// A typo must show an absent segment rather than quietly counting
		// against whatever the hostname suggested — a visible symptom the
		// operator can trace back to the value they just typed.
		{"unknown override does not fall back", ghRemote, "githbu", "", false},
		{"override rescues a self-hosted host", "https://git.example.com/g/p.git", "gitlab", "gitlab", true},
		{"no remote and no override", "", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := resolveForge(tt.remote, tt.override)
			if ok != tt.wantOK {
				t.Fatalf("resolveForge(%q, %q) ok = %v, want %v", tt.remote, tt.override, ok, tt.wantOK)
			}
			if ok && got.name != tt.wantName {
				t.Errorf("resolveForge(%q, %q) = %q, want %q", tt.remote, tt.override, got.name, tt.wantName)
			}
		})
	}
}

// The two CLIs disagree about defaults in a way that inverts the fix: `gh`
// lists closed items unless told `--state open`, while `glab` already defaults
// to open and would be narrowed by a state flag. Copying one command's shape
// onto the other silently returns the wrong number rather than failing.
func TestForgeArgs_StateHandlingDiffers(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"issue", "pr"} {
		gh := strings.Join(forgeGitHub.argsFor(kind), " ")
		if !strings.Contains(gh, "--state open") {
			t.Errorf("gh %s args must name the open state: %q", kind, gh)
		}

		gl := strings.Join(forgeGitLab.argsFor(kind), " ")
		if strings.Contains(gl, "--state") || strings.Contains(gl, "--closed") ||
			strings.Contains(gl, "--merged") || strings.Contains(gl, "--all") {
			t.Errorf("glab %s args must not pass a state flag; open is its default: %q", kind, gl)
		}
	}
}

// glab's short flags are not stable across its own subcommands: on `mr list`
// -F abbreviates --output, but on `issue list` -F abbreviates --output-format,
// which selects details/ids/urls and yields no JSON at all.
func TestForgeArgs_GitLabUsesLongFormOutput(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"issue", "pr"} {
		args := forgeGitLab.argsFor(kind)
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "--output json") {
			t.Errorf("glab %s args must request JSON by long-form flag: %q", kind, joined)
		}
		for _, a := range args {
			if a == "-F" || a == "-O" {
				t.Errorf("glab %s args must not use a short output flag: %q", kind, joined)
			}
		}
	}
}

// Every forge must reduce its listing to a bare integer, because the caller
// parses one number without knowing which CLI answered.
func TestForgeArgs_AllEndInACount(t *testing.T) {
	t.Parallel()

	for _, f := range []forgeSpec{forgeGitHub, forgeGitLab} {
		for _, kind := range []string{"issue", "pr"} {
			args := f.argsFor(kind)
			if n := len(args); n < 2 || args[n-2] != "--jq" || args[n-1] != "length" {
				t.Errorf("%s %s args must end in --jq length, got %v", f.name, kind, args)
			}
		}
	}
}

func TestForgeArgs_PRAndIssueDiffer(t *testing.T) {
	t.Parallel()

	if strings.Join(forgeGitHub.argsFor("pr"), " ") == strings.Join(forgeGitHub.argsFor("issue"), " ") {
		t.Error("gh pr and issue args are identical; one of them counts the wrong thing")
	}
	// GitLab names them merge requests, which is the whole reason the two
	// argument lists exist rather than one with a substituted noun.
	if forgeGitLab.argsFor("pr")[0] != "mr" {
		t.Errorf("glab change requests are `mr`, got %q", forgeGitLab.argsFor("pr")[0])
	}
	// An unrecognised kind must not silently produce a change-request listing.
	if strings.Join(forgeGitHub.argsFor("anything-else"), " ") != strings.Join(forgeGitHub.argsFor("issue"), " ") {
		t.Error("an unknown kind must fall back to the issue listing")
	}
}

func TestForgeOverride(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string // empty means: write no file
		want string
	}{
		{"reads the key", "statusline:\n  forge: gitlab\n", "gitlab"},
		{"trims whitespace", "statusline:\n  forge: \"  gitlab  \"\n", "gitlab"},
		{"absent key yields empty", "statusline:\n  theme: catppuccin-mocha\n", ""},
		{"malformed yaml fails open", "statusline:\n  forge: [", ""},
		{"missing file fails open", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if tt.body != "" {
				dir := filepath.Join(root, ".moai", "config", "sections")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "statusline.yaml"), []byte(tt.body), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			if got := forgeOverride(root); got != tt.want {
				t.Errorf("forgeOverride = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestForgeOverride_EmptyRootIsANoOp(t *testing.T) {
	t.Parallel()

	if got := forgeOverride(""); got != "" {
		t.Errorf("forgeOverride(\"\") = %q, want empty", got)
	}
}
