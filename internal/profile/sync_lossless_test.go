package profile

// SPEC-WEB-SAVE-LOSSLESS-001 — AC-WSL-005: a user-name edit must change ONLY
// the `name:` row of user.yaml (seam line-splice), never a struct re-marshal.
// models.UserConfig models only `name` (pkg/models/config.go), so any
// struct-based repair loses unmodeled keys at the re-marshal point — the
// assertion here is on the resulting BYTES, which is unreachable through a
// struct round-trip (plan.md §A.1, plan-audit F1 redesign).
//
// RED on the pre-implementation tree: SyncToProjectConfig replaced the whole
// user section with models.UserConfig{Name: ...} and Save() re-marshaled it,
// wiping unmodeled keys and comments.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedUserFixture writes a user.yaml carrying an unmodeled key and comments —
// the exact loss shape GitHub #1731 reports.
func seedUserFixture(t *testing.T, root, content string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "user.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readUserFixture(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "user.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// changedLineIndexes counts differing lines (same contract as the settings
// lossless helper — duplicated here because the profile package cannot import
// internal/settings, which itself imports internal/profile).
func changedLineIndexes(t *testing.T, before, after string) []int {
	t.Helper()
	b := strings.Split(before, "\n")
	a := strings.Split(after, "\n")
	if len(a) != len(b) {
		t.Fatalf("line count changed: before=%d after=%d\n--- before ---\n%s\n--- after ---\n%s", len(b), len(a), before, after)
	}
	var changed []int
	for i := range b {
		if a[i] != b[i] {
			changed = append(changed, i)
		}
	}
	return changed
}

const richUserYAML = `# user identity — hand-maintained
user:
  name: original
  github_username: example-user
  timezone: Asia/Seoul

# trailing note comment
`

// TestSyncToProjectConfig_NameEditSplicesOneRow carries AC-WSL-005: changing
// the user name must change exactly the `name:` row; every other byte —
// unmodeled keys, comments, blank lines — survives verbatim.
func TestSyncToProjectConfig_NameEditSplicesOneRow(t *testing.T) {
	projectRoot := t.TempDir()
	seedUserFixture(t, projectRoot, richUserYAML)

	if err := SyncToProjectConfig(projectRoot, ProfilePreferences{UserName: "newuser"}); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	after := readUserFixture(t, projectRoot)
	changed := changedLineIndexes(t, richUserYAML, after)
	if len(changed) != 1 {
		t.Fatalf("name edit changed %d lines, want 1\n--- before ---\n%s\n--- after ---\n%s", len(changed), richUserYAML, after)
	}
	if line := strings.Split(after, "\n")[changed[0]]; !strings.Contains(line, "name: newuser") {
		t.Fatalf("changed line %q does not carry the new name", line)
	}
	for _, want := range []string{"github_username: example-user", "timezone: Asia/Seoul", "# user identity — hand-maintained", "# trailing note comment"} {
		if !strings.Contains(after, want) {
			t.Errorf("unmodeled content %q lost by the name edit:\n%s", want, after)
		}
	}
}

// TestSyncToProjectConfig_NameUnchangedNoWrite carries REQ-WSL-001: a
// submission equal to the persisted name writes nothing (mtime included).
func TestSyncToProjectConfig_NameUnchangedNoWrite(t *testing.T) {
	projectRoot := t.TempDir()
	seedUserFixture(t, projectRoot, richUserYAML)
	path := filepath.Join(projectRoot, ".moai", "config", "sections", "user.yaml")
	beforeStat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := SyncToProjectConfig(projectRoot, ProfilePreferences{UserName: "original"}); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	afterStat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !afterStat.ModTime().Equal(beforeStat.ModTime()) {
		t.Error("unchanged name still rewrote user.yaml (mtime moved — REQ-WSL-001)")
	}
	if got := readUserFixture(t, projectRoot); got != richUserYAML {
		t.Errorf("unchanged name rewrote user.yaml\n--- before ---\n%s\n--- after ---\n%s", richUserYAML, got)
	}
}

// SPEC-WEB-SAVE-LOSSLESS-001 — sync-audit F-1: language.yaml was the FIFTH
// residual re-marshal path — SyncToProjectConfig applied language preferences
// via SetSection("language") + Save(), a full struct re-marshal that wipes
// comments and unmodeled keys. The web console's four language selects reach
// this path, so the writer-universal clause of REQ-WSL-002/003 was violated.
// The auditor probe-reproduced the loss; this test is that probe as a
// regression guard, on the same byte-level assertion shape as the user.yaml
// splice tests above.

const richLanguageYAML = `# language — hand-maintained
language:
  # conversation locale (user comment)
  conversation_language: ko
  conversation_language_name: Korean (한국어)
  agent_prompt_language: en
  git_commit_messages: en
  code_comments: ko
  documentation: ko
  # user-added unmodeled key — must survive a language edit
  custom_locale_note: keep-me

# trailing note
`

// seedLanguageFixture writes a language.yaml carrying comments and an
// unmodeled key — the loss shape the sync-audit probe observed on the
// re-marshal path.
func seedLanguageFixture(t *testing.T, root, content string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "language.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readLanguageFixture(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "language.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// TestSyncToProjectConfig_LanguageEditSplicesRows carries the sync-audit F-1
// regression: a conversation-language edit must change ONLY the
// conversation_language + conversation_language_name rows (seam line-splice);
// comments, unmodeled keys, and untouched language scalars survive verbatim.
func TestSyncToProjectConfig_LanguageEditSplicesRows(t *testing.T) {
	projectRoot := t.TempDir()
	seedLanguageFixture(t, projectRoot, richLanguageYAML)

	if err := SyncToProjectConfig(projectRoot, ProfilePreferences{ConversationLang: "en"}); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	after := readLanguageFixture(t, projectRoot)
	changed := changedLineIndexes(t, richLanguageYAML, after)
	if len(changed) != 2 {
		t.Fatalf("language edit changed %d lines, want 2 (conversation_language + conversation_language_name)\n--- before ---\n%s\n--- after ---\n%s", len(changed), richLanguageYAML, after)
	}
	for _, want := range []string{
		"# language — hand-maintained",
		"# conversation locale (user comment)",
		"# user-added unmodeled key — must survive a language edit",
		"custom_locale_note: keep-me",
		"git_commit_messages: en",
		"code_comments: ko",
		"# trailing note",
	} {
		if !strings.Contains(after, want) {
			t.Errorf("language edit lost %q:\n%s", want, after)
		}
	}
}

// TestSyncToProjectConfig_LanguageUnchangedNoWrite carries REQ-WSL-001 on the
// language path: a submission equal to the persisted values writes nothing.
func TestSyncToProjectConfig_LanguageUnchangedNoWrite(t *testing.T) {
	projectRoot := t.TempDir()
	seedLanguageFixture(t, projectRoot, richLanguageYAML)
	path := filepath.Join(projectRoot, ".moai", "config", "sections", "language.yaml")
	beforeStat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := SyncToProjectConfig(projectRoot, ProfilePreferences{ConversationLang: "ko"}); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	afterStat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !afterStat.ModTime().Equal(beforeStat.ModTime()) {
		t.Error("unchanged language value still rewrote language.yaml (mtime moved — REQ-WSL-001)")
	}
	if got := readLanguageFixture(t, projectRoot); got != richLanguageYAML {
		t.Errorf("unchanged language value rewrote language.yaml\n--- before ---\n%s\n--- after ---\n%s", richLanguageYAML, got)
	}
}

// TestSyncToProjectConfig_NameAbsentUpsert carries the AC-WSL-005 F6 variant:
// a user.yaml without a `name:` key gets one via the upsert fallback. The Then
// is relaxed to the data level (C3): existing keys, values, and comments must
// survive; blank-line/indent normalization is tolerated.
func TestSyncToProjectConfig_NameAbsentUpsert(t *testing.T) {
	projectRoot := t.TempDir()
	seedUserFixture(t, projectRoot, "# comment survives\nuser:\n  github_username: example-user\n")

	if err := SyncToProjectConfig(projectRoot, ProfilePreferences{UserName: "first"}); err != nil {
		t.Fatalf("SyncToProjectConfig: %v", err)
	}

	after := readUserFixture(t, projectRoot)
	if !strings.Contains(after, "github_username: example-user") {
		t.Errorf("existing unmodeled key lost by the name upsert:\n%s", after)
	}
	if !strings.Contains(after, "name: first") {
		t.Errorf("name not persisted by the upsert:\n%s", after)
	}
	if !strings.Contains(after, "# comment survives") {
		t.Errorf("comment lost by the name upsert:\n%s", after)
	}
}
