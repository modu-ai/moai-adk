package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeParticipationFile writes a participation.yaml under a temporary
// MOAI_HOME and points the env at it. Every test here isolates the real home
// with t.Setenv("MOAI_HOME", ...) per the acceptance preface.
func writeParticipationFile(t *testing.T, body string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	if body != "" {
		dir := filepath.Join(home, "config")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("mkdir config: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "participation.yaml"), []byte(body), 0o600); err != nil {
			t.Fatalf("write participation.yaml: %v", err)
		}
	}
	return home
}

// hostileProjectTree builds a project tree whose TRACKED section file carries
// participation keys and an attacker repository — the mutant fixture: a reader
// that falls back to the project tier dies here.
func hostileProjectTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	body := "feedback:\n  repository: attacker/x\n  auto_submit: false\n  participation: true\n  participation_asked: true\n"
	if err := os.WriteFile(filepath.Join(dir, "feedback.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write hostile feedback.yaml: %v", err)
	}
	return root
}

func TestUserParticipationDefaultsOff(t *testing.T) {
	// No file at all.
	writeParticipationFile(t, "")
	up := ReadUserParticipation()
	if up.Enabled || up.Asked || up.Repository != "" {
		t.Fatalf("absent file read %+v, want all-false/empty", up)
	}
	if UserParticipationRepository() != DefaultFeedbackRepository {
		t.Fatalf("repository with no user file = %q, want the compiled default", UserParticipationRepository())
	}

	// A file with the keys present but false/empty.
	writeParticipationFile(t, "participation:\n  enabled: false\n  asked: false\n")
	up = ReadUserParticipation()
	if up.Enabled || up.Asked {
		t.Fatalf("false keys read %+v, want both false", up)
	}

	// A directory in place of the file.
	home := writeParticipationFile(t, "")
	if err := os.MkdirAll(filepath.Join(home, "config", "participation.yaml"), 0o700); err != nil {
		t.Fatalf("mkdir over the file path: %v", err)
	}
	up = ReadUserParticipation()
	if up.Enabled || up.Asked {
		t.Fatalf("directory-at-path read %+v, want both false", up)
	}

	// A malformed file.
	writeParticipationFile(t, "participation: [broken\n  yaml::\n")
	up = ReadUserParticipation()
	if up.Enabled || up.Asked {
		t.Fatalf("malformed file read %+v, want both false", up)
	}

	// A wrongly typed value.
	writeParticipationFile(t, "participation:\n  enabled: \"yes-please\"\n")
	up = ReadUserParticipation()
	if up.Enabled {
		t.Fatalf("wrongly typed enabled read %+v, want false", up)
	}

	// An unreadable home directory.
	home = t.TempDir()
	t.Setenv("MOAI_HOME", filepath.Join(home, "config", "participation.yaml", "deeper", "file.yaml"))
	up = ReadUserParticipation()
	if up.Enabled || up.Asked {
		t.Fatalf("unreadable home read %+v, want both false", up)
	}
}

func TestUserParticipationIgnoresTrackedProjectFile(t *testing.T) {
	root := hostileProjectTree(t)
	writeParticipationFile(t, "")
	t.Chdir(root)

	// The hostile project file says participation: true; the reader never
	// opens a project file, so nothing is enabled.
	up := ReadUserParticipation()
	if up.Enabled || up.Asked {
		t.Fatalf("tracked project file enabled participation: %+v", up)
	}
	if UserParticipationRepository() != DefaultFeedbackRepository {
		t.Fatalf("project-tier repository leaked: %q", UserParticipationRepository())
	}

	// A user file saying false under the same tree also reads false.
	writeParticipationFile(t, "participation:\n  enabled: false\n  asked: true\n")
	up = ReadUserParticipation()
	if up.Enabled {
		t.Fatalf("user file false read enabled: %+v", up)
	}
}

func TestUserParticipationRepositoryDefault(t *testing.T) {
	// (a) hostile project repository + no user file: the compiled default.
	root := hostileProjectTree(t)
	writeParticipationFile(t, "")
	t.Chdir(root)
	if got := UserParticipationRepository(); got != DefaultFeedbackRepository {
		t.Fatalf("hostile project repository won: %q", got)
	}

	// (b) user file enabled true, no repository key: still the default.
	writeParticipationFile(t, "participation:\n  enabled: true\n  asked: true\n")
	up := ReadUserParticipation()
	if !up.Enabled {
		t.Fatal("user file enabled true read false")
	}
	if got := UserParticipationRepository(); got != DefaultFeedbackRepository {
		t.Fatalf("repository = %q, want the default with no user repository key", got)
	}

	// (c) a well-formed user repository is honoured.
	writeParticipationFile(t, "participation:\n  enabled: true\n  repository: someone/mirror\n")
	if got := UserParticipationRepository(); got != "someone/mirror" {
		t.Fatalf("repository = %q, want someone/mirror", got)
	}

	// (d) a malformed user repository falls back to the default.
	writeParticipationFile(t, "participation:\n  enabled: true\n  repository: \"not a repo shape\"\n")
	if got := UserParticipationRepository(); got != DefaultFeedbackRepository {
		t.Fatalf("repository = %q, want the default for a malformed value", got)
	}
}
