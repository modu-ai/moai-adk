package cli

// Card t1139 follow-up (sync-audit F1/F3): the update render context carries
// project.name / user.name read from the existing config. A name the render
// cannot carry verbatim — one that trips the renderer's unexpanded-token check
// ($TEAM, {{.Version}}) or breaks the double-quoted YAML scalar the templates
// wrap it in (", \) — must not halt the update or break the 3-way merge.
// loadUpdateIdentity reads such a name as "", the render carries the pre-t1139
// empty name, and — because no render could have written the name, so the
// snapshot BASE differs from it — the merge keeps the user's value as a
// customization.
//
// prepareSafeInitHome sets environment variables, the update helper chdirs, and
// captureProcessStderr swaps os.Stderr, so no test in this file may run in
// parallel.

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// captureProcessStderr runs fn with os.Stderr redirected to a file under
// t.TempDir() and returns what was written. The section-merge warning
// ("Warning: merge failed for …") is printed to os.Stderr directly, not to the
// command's error stream, so this is the only place it can be observed.
func captureProcessStderr(t *testing.T, fn func()) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stderr-*")
	if err != nil {
		t.Fatalf("create stderr capture: %v", err)
	}
	orig := os.Stderr
	os.Stderr = f
	func() {
		defer func() { os.Stderr = orig }()
		fn()
	}()
	if err := f.Close(); err != nil {
		t.Fatalf("close stderr capture: %v", err)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("read stderr capture: %v", err)
	}
	return string(data)
}

// runForcedTemplateSyncResult drives the same forced template-sync cycle as
// runForcedTemplateSyncAt but returns the error and every output stream instead
// of failing the test, so a halted update is an assertion rather than a fatal.
func runForcedTemplateSyncResult(t *testing.T, root string) (output string, syncErr error) {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if chErr := os.Chdir(root); chErr != nil {
		t.Fatalf("chdir to fixture: %v", chErr)
	}
	defer func() { _ = os.Chdir(origDir) }()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("config", false, "")
	_ = cmd.Flags().Set("force", "true")
	_ = cmd.Flags().Set("yes", "true")
	var buf, errBuf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&errBuf)
	cmd.SetContext(context.Background())

	processErr := captureProcessStderr(t, func() {
		syncErr = runTemplateSyncWithReporter(cmd, nil, true)
	})
	return buf.String() + errBuf.String() + processErr, syncErr
}

// setIdentityName rewrites the init-rendered `name: "<old>"` line of a section
// file to the given value, written as a valid YAML double-quoted scalar — the
// state a user reaches by hand-editing the file.
func setIdentityName(t *testing.T, file, old, value string) {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	edited := replaceOnce(t, string(data), `name: "`+old+`"`, "name: "+strconv.Quote(value))
	if err := os.WriteFile(file, []byte(edited), 0o644); err != nil {
		t.Fatalf("write %s: %v", file, err)
	}
	if got := sectionValue(t, file, strings.TrimSuffix(filepath.Base(file), ".yaml"), "name"); got != value {
		t.Fatalf("fixture edit did not take: %s name = %q, want %q", file, got, value)
	}
}

// TestUpdateForce_UnrenderableIdentityNamesSurvive is the F1/F3 end-to-end
// regression: after a real init, the user hand-edits both names to values the
// render cannot carry; two forced updates must both succeed, keep both names
// verbatim, and print no section-merge failure for either file.
func TestUpdateForce_UnrenderableIdentityNamesSurvive(t *testing.T) {
	cases := []struct {
		label, project, user string
	}{
		{"dollar token", "$TEAM-proj", "$TEAM"},
		{"template action", "{{.Version}}", "x {{.Version}}"},
		{"double quote", `p "q"`, `Kim "Goos"`},
		{"all three", `$TEAM {{.Version}} "q"`, `"$TEAM" {{.Version}}`},
		{"backslash", `C:\Users\p`, `DOMAIN\goos`},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			root := initIdentityProject(t)
			projectFile := sectionsFile(root, "project.yaml")
			userFile := sectionsFile(root, "user.yaml")
			setIdentityName(t, projectFile, identityProjectName, tc.project)
			setIdentityName(t, userFile, identityUserName, tc.user)

			for i := 1; i <= 2; i++ {
				output, syncErr := runForcedTemplateSyncResult(t, root)
				if syncErr != nil {
					t.Fatalf("update #%d halted: %v", i, syncErr)
				}
				for _, line := range strings.Split(output, "\n") {
					if strings.Contains(line, "merge failed") &&
						(strings.Contains(line, "user.yaml") || strings.Contains(line, "project.yaml")) {
						t.Errorf("update #%d: section merge failed: %q", i, line)
					}
				}
				if got := sectionValue(t, projectFile, "project", "name"); got != tc.project {
					t.Errorf("update #%d: project.name = %q, want %q", i, got, tc.project)
				}
				if got := sectionValue(t, userFile, "user", "name"); got != tc.user {
					t.Errorf("update #%d: user.name = %q, want %q", i, got, tc.user)
				}
			}
		})
	}
}

// TestIdentityLoader_AcceptedNamesRenderVerbatim pins loadUpdateIdentity's
// accept/reject rule to the renderer in both directions: every name it keeps
// renders through the real embedded project.yaml / user.yaml templates without
// error and parses back byte-identical, and every name it drops does NOT. The
// second direction is what the over-rejecting filter violated — it dropped
// names ("cost$5", "a{{b", ZWJ emoji) init renders verbatim, and a dropped
// init-origin name is erased by the next merge. The corpus is every ASCII byte
// plus the multi-byte sequences the renderer's token pattern or the YAML
// double-quoted scalar treat specially.
func TestIdentityLoader_AcceptedNamesRenderVerbatim(t *testing.T) {
	t.Parallel()
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("embedded templates: %v", err)
	}
	renderer := template.NewRenderer(fsys)

	corpus := []string{"$TEAM", "${TEAM}", "{{.Version}}", "{{X}}", "a{{b", "a}}b", "\u0085", "\u2028", "\u2029", "\ufeff", "é", "구스", "$HOME", "${HOME}", "cost$5", "my$app", "👩\u200d💻 x", " padded ", "a#b: c"}
	for b := 0; b < 0x80; b++ {
		corpus = append(corpus, "a"+string(rune(b))+"b", string(rune(b)))
	}

	accepted := map[string]bool{}
	for _, value := range corpus {
		root := t.TempDir()
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		quoted := strconv.Quote(value)
		if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte("project:\n  name: "+quoted+"\n"), 0o644); err != nil {
			t.Fatalf("write project.yaml: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "user.yaml"), []byte("user:\n  name: "+quoted+"\n"), 0o644); err != nil {
			t.Fatalf("write user.yaml: %v", err)
		}
		if raw := config.LoadProjectName(root); raw != value {
			t.Fatalf("fixture %q: stored name reads back as %q", value, raw)
		}
		projectName, userName := loadUpdateIdentity(root)
		if projectName != userName {
			t.Errorf("%q: project and user rules disagree: project %q, user %q", value, projectName, userName)
		}
		if projectName != "" && projectName != value {
			t.Errorf("%q: kept as %q, want the value verbatim or empty", value, projectName)
			continue
		}
		kept := projectName == value
		if kept {
			accepted[value] = true
		}
		ctx := template.NewTemplateContext(template.WithProject(value, root), template.WithUser(value))
		for _, tc := range []struct{ tmpl, key string }{
			{".moai/config/sections/project.yaml.tmpl", "project"},
			{".moai/config/sections/user.yaml.tmpl", "user"},
		} {
			roundTrips := false
			if out, renderErr := renderer.Render(tc.tmpl, ctx); renderErr == nil {
				var doc map[string]map[string]any
				if yaml.Unmarshal(out, &doc) == nil {
					roundTrips = doc[tc.key]["name"] == value
				}
			}
			if kept != roundTrips {
				t.Errorf("%q via %s: kept = %v but renders verbatim = %v", value, tc.tmpl, kept, roundTrips)
			}
		}
	}

	// Positive controls: ordinary printable names and the init-origin names of
	// the N1 regression are kept, so a vacuous "drop everything" rule fails.
	if len(accepted) < 80 {
		t.Errorf("kept only %d of %d corpus names; ordinary printable names must pass", len(accepted), len(corpus))
	}
	for _, want := range []string{"cost$5", "my$app", "a{{b", "👩\u200d💻 x", "구스", "a\tb"} {
		if !accepted[want] {
			t.Errorf("%q was dropped; init renders it verbatim, so the update must keep it", want)
		}
	}
	// Negative controls: the hand-edit names the render cannot carry are
	// dropped, so a vacuous "keep everything" rule fails.
	for _, want := range []string{"$TEAM", "{{.Version}}", "a\"b", "a\\b", "a\nb"} {
		if accepted[want] {
			t.Errorf("%q was kept; the render cannot carry it verbatim", want)
		}
	}
	for _, raw := range []string{"$TEAM", "{{.Version}}"} {
		ctx := template.NewTemplateContext(template.WithUser(raw))
		if _, renderErr := renderer.Render(".moai/config/sections/user.yaml.tmpl", ctx); !errors.Is(renderErr, template.ErrUnexpandedToken) {
			t.Errorf("raw %q: render error = %v, want ErrUnexpandedToken (control lost its teeth)", raw, renderErr)
		}
	}
}
