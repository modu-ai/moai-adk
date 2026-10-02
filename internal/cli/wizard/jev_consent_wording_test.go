package wizard

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The Jev consent sentence is user-facing text repeated across the init wizard,
// the moai web consent note, the docs-site pages and the workflow.yaml comment.
// Once Jev is switched on it orders `moai todo --auto` candidates and can give an
// optional Kickoff cross-check a second signal, so the old absolute "decides
// nothing" wording is no longer true. This guard keeps every surface on the
// "signal only" wording, in all four locales (card t1429).

// jevOldWording are the absolute phrasings that must not come back.
var jevOldWording = []string{
	"decides nothing",
	"스스로 결정하지는 않습니다",
	"判断そのものは行いません",
	"它本身不做任何决定。",
}

// jevMCPOldWording are the absolute "never a gate input" phrasings; they are
// matched on the MCP guide pages only, where the sentence they replaced lived.
var jevMCPOldWording = []string{
	"gate input",
	"게이트 입력으로 쓰지 않습니다",
	"ゲート入力には使いません",
	"门控输入",
}

// jevSignalMarker is the per-locale phrase every consent surface now carries.
var jevSignalMarker = map[string]string{
	"en": "only as a signal",
	"ko": "신호로만 쓰입니다",
	"ja": "信号としてだけです",
	"zh": "也只作为信号",
}

func jevWordingRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the working directory")
		}
		dir = parent
	}
}

func jevWordingRead(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

// TestJevConsentWording_NoAbsoluteWordingRemains sweeps every user-facing tree
// that carries the consent sentence for the old absolute phrasings.
func TestJevConsentWording_NoAbsoluteWordingRemains(t *testing.T) {
	root := jevWordingRepoRoot(t)
	scanned := 0
	for _, tree := range []string{
		"internal/cli/wizard",
		"internal/web",
		"docs-site/content",
		".moai/config/sections",
		"internal/template/templates/.moai/config/sections",
	} {
		err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(tree)), func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || strings.HasSuffix(path, "_test.go") {
				return err
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			scanned++
			rel, _ := filepath.Rel(root, path)
			text := string(b)
			for _, old := range jevOldWording {
				if strings.Contains(text, old) {
					t.Errorf("%s still carries the absolute wording %q — state that Jev's answer is read by a person and is used automatically only as a signal", rel, old)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", tree, err)
		}
	}
	if scanned == 0 {
		t.Fatal("scanned no files — the guard is blind")
	}
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		rel := "docs-site/content/" + lang + "/guides/mcp-server.md"
		text := jevWordingRead(t, root, rel)
		for _, old := range jevMCPOldWording {
			if strings.Contains(text, old) {
				t.Errorf("%s still carries the absolute wording %q", rel, old)
			}
		}
	}
}

// TestJevConsentWording_SignalWordingPresent pins the new wording on every
// surface in every locale, so reverting a single locale or file is caught.
func TestJevConsentWording_SignalWordingPresent(t *testing.T) {
	root := jevWordingRepoRoot(t)
	cases := []struct {
		file  string
		langs []string
	}{
		{"internal/cli/wizard/questions.go", []string{"en"}},
		{"internal/cli/wizard/translations.go", []string{"ko", "ja", "zh"}},
		{"internal/web/jevkey.go", []string{"en"}},
		{"internal/web/assets/i18n.js", []string{"en", "ko", "ja", "zh"}},
		{"docs-site/content/en/getting-started/init-wizard.md", []string{"en"}},
		{"docs-site/content/ko/getting-started/init-wizard.md", []string{"ko"}},
		{"docs-site/content/ja/getting-started/init-wizard.md", []string{"ja"}},
		{"docs-site/content/zh/getting-started/init-wizard.md", []string{"zh"}},
		{"docs-site/content/en/guides/mcp-server.md", []string{"en"}},
		{"docs-site/content/ko/guides/mcp-server.md", []string{"ko"}},
		{"docs-site/content/ja/guides/mcp-server.md", []string{"ja"}},
		{"docs-site/content/zh/guides/mcp-server.md", []string{"zh"}},
	}
	for _, c := range cases {
		text := jevWordingRead(t, root, c.file)
		for _, lang := range c.langs {
			if !strings.Contains(text, jevSignalMarker[lang]) {
				t.Errorf("%s (%s) is missing the signal-only wording %q", c.file, lang, jevSignalMarker[lang])
			}
		}
	}
	for _, rel := range []string{
		".moai/config/sections/workflow.yaml",
		"internal/template/templates/.moai/config/sections/workflow.yaml",
	} {
		if !strings.Contains(jevWordingRead(t, root, rel), "makes no decision itself") {
			t.Errorf("%s comment is missing the revised wording", rel)
		}
	}
}
