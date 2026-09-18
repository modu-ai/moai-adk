package cli

// M5 absorption tests (SPEC-INIT-TUX-I18N-001): AC-ITI-005's storage half,
// AC-ITI-006's nine-case preservation table, and AC-ITI-007's command-level
// contract over the wizard-run and session-worktree seams.
//
// Every behavioral case drives the REAL save path (WritePreferences,
// SyncToProjectConfig, persistProjectConfig) against a temp profile store and
// a temp project; only the wizard form itself is injected, through the
// profileWizardRunner seam (design.md §2.2). Nothing here opens a huh form.
//
// requireAbsorbedRouting gates every case: before the M5 absorption lands,
// runProfileSetup carries no seam calls, and invoking RunE would open the v1
// huh form on the tester's terminal — so the tests fail on the gate and never
// reach the command. The gate is the same source-scan judgement the M4 RED
// ledger made with git grep (progress.md M4 section).
//
// These tests swap package variables and chdir, so none of them may call
// t.Parallel.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
)

// sentinelUserName is the user name the AC-ITI-007 enter seam writes into the
// sentinel preferences file. When the captured initial values carry it, the
// profile read observed the file the enter seam wrote — i.e. the enter seam
// ran before the profile read (AC-ITI-007's ordering clause).
const sentinelUserName = "entered-by-enter-seam"

// cliPackageDir is the package directory captured at process start, before
// any test chdirs into a temp project; the source-scan gate reads the product
// file relative to it.
var cliPackageDir = func() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}()

// requireAbsorbedRouting fails fast — BEFORE any RunE invocation — when
// runProfileSetup has not been absorbed onto the seams yet. Without the
// routing, invoking the command here would open the v1 huh form on the
// tester's terminal, so the gate refuses to run the command at all.
func requireAbsorbedRouting(t *testing.T) {
	t.Helper()
	if cliPackageDir == "" {
		t.Fatal("cannot resolve the package directory; refusing to judge the routing gate")
	}
	data, err := os.ReadFile(filepath.Join(cliPackageDir, "profile_setup.go"))
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	code := nonCommentLines(string(data))
	for _, needle := range []string{"profileWizardRunner(", "enterSessionWorktreeFn(", "cleanupSessionWorktreeFn("} {
		if !strings.Contains(code, needle) {
			t.Fatalf("RED: runProfileSetup does not route through %s yet — the M5 absorption has not landed; refusing to invoke the command (it would open the v1 huh form)", needle)
		}
	}
}

// isolateProfileStore redirects the profile store under t.TempDir() and
// returns its base. Swaps a package variable — callers must not run parallel.
func isolateProfileStore(t *testing.T) string {
	t.Helper()
	base := filepath.Join(t.TempDir(), "claude-profiles")
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = base
	t.Cleanup(func() { profile.BaseDirOverride = orig })
	return base
}

// absorbSeedFiles are the project config files a seeded absorb project starts
// with. Every value differs from the answers the tests inject, so an
// unwritten file would keep its seed and fail the assertion.
var absorbSeedFiles = map[string]string{
	"user.yaml": "user:\n  name: seed-user\n",
	"language.yaml": "language:\n" +
		"  conversation_language: ja\n" +
		"  conversation_language_name: ja\n" +
		"  git_commit_messages: ja\n" +
		"  code_comments: ja\n" +
		"  documentation: ja\n",
	// quality.yaml carries the development mode under its constitution key —
	// the on-disk shape the config manager reads (see
	// seedRemovedQuestionProject for the same shape).
	"quality.yaml": "constitution:\n" +
		"  development_mode: tdd\n",
	"statusline.yaml": "statusline:\n" +
		"  theme: catppuccin-mocha\n" +
		"  segments:\n" +
		"    model: true\n" +
		"    context: false\n",
	"git-convention.yaml": "git_convention:\n" +
		"  convention: angular\n",
}

// seedAbsorbProject writes a temp MoAI project whose persisted values differ
// from the injected answers, moves the process into it, and returns the root.
func seedAbsorbProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range absorbSeedFiles {
		if err := os.WriteFile(filepath.Join(sections, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	t.Chdir(root)
	return root
}

// seedBareDir moves the process into a directory with NO .moai, the AC-ITI-005
// (5) outside-a-project cwd.
func seedBareDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Chdir(root)
	return root
}

// sha256File returns the hex digest of the file at path, or the error text.
func sha256File(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// runnerCapture records what one profileWizardRunner call received.
type runnerCapture struct {
	calls   int
	initial wizard.ProfileResult
	opts    wizard.ProfileOptions
	locale  string
}

// absorbRun carries one command-level execution's observations.
type absorbRun struct {
	stdout   bytes.Buffer
	stderr   bytes.Buffer
	err      error
	wtLog    []string // "enter", "cleanup:true", "cleanup:false" in call order
	captured runnerCapture
}

// runAbsorbedSetup executes one explicit profile entry with the wizard form
// injected: answersFn decides what the wizard "answers". entry is "setup"
// (profileSetupCmd.RunE) or "profile" (runProfileCmd --setup); name is the
// profile-name argument ("" = none). withSentinel makes the worktree-enter
// seam write a sentinel preferences file first, so the read-after-enter
// ordering is observable through the captured initial values. Callers
// isolate the profile store BEFORE seeding stored preferences.
func runAbsorbedSetup(t *testing.T, entry, name string, withSentinel bool, answersFn func(cap *runnerCapture) (*wizard.ProfileResult, error)) *absorbRun {
	t.Helper()
	requireAbsorbedRouting(t)

	run := &absorbRun{}
	profileName := name
	if profileName == "" {
		profileName = "default"
	}

	origRunner := profileWizardRunner
	profileWizardRunner = func(initial wizard.ProfileResult, opts wizard.ProfileOptions, locale string) (*wizard.ProfileResult, error) {
		run.captured.calls++
		run.captured.initial = initial
		run.captured.opts = opts
		run.captured.locale = locale
		return answersFn(&run.captured)
	}
	t.Cleanup(func() { profileWizardRunner = origRunner })

	origEnter := enterSessionWorktreeFn
	enterSessionWorktreeFn = func(_ *config.Config, _ string, _ io.Writer) string {
		run.wtLog = append(run.wtLog, "enter")
		if withSentinel {
			if err := profile.WritePreferences(profileName, profile.ProfilePreferences{
				UserName:         sentinelUserName,
				ConversationLang: "ko",
			}); err != nil {
				t.Fatalf("enter seam sentinel write: %v", err)
			}
		}
		return ""
	}
	t.Cleanup(func() { enterSessionWorktreeFn = origEnter })

	origCleanup := cleanupSessionWorktreeFn
	cleanupSessionWorktreeFn = func(_ *config.Config, _ string, cleanExit bool, _ io.Writer) {
		run.wtLog = append(run.wtLog, fmt.Sprintf("cleanup:%v", cleanExit))
	}
	t.Cleanup(func() { cleanupSessionWorktreeFn = origCleanup })

	cmd := &cobra.Command{Use: entry}
	cmd.SetOut(&run.stdout)
	cmd.SetErr(&run.stderr)
	var args []string
	if name != "" {
		args = []string{name}
	}
	if entry == "profile" {
		cmd.Flags().BoolP("setup", "s", true, "")
		run.err = runProfileCmd(cmd, args)
		return run
	}
	run.err = profileSetupCmd.RunE(cmd, args)
	return run
}

// absorbAnswers is the injected wizard answer set used by the preservation
// and save tests: every value differs from the seedAbsorbProject seeds.
func absorbAnswers() wizard.ProfileResult {
	return wizard.ProfileResult{
		ConversationLang: "ko",
		UserName:         "absorbed-user",
		GitCommitLang:    "en",
		CodeCommentLang:  "zh",
		DocLang:          "ko",
		Model:            "opus[1m]",
		ModelPolicy:      "high",
		EffortLevel:      "xhigh",
		PermissionMode:   "plan",
		DevelopmentMode:  "ddd",
	}
}

// TestProfileSetupAbsorbed_PreservationTable is AC-ITI-006: the nine
// REQ-ITI-005 preservation behaviors, one table case each. Every case
// snapshots the pre-save file bytes; the cases that promise byte identity
// assert against the snapshot.
func TestProfileSetupAbsorbed_PreservationTable(t *testing.T) {
	// A superseded canonical id (SPEC-INIT-TUX-I18N-001: the ids in
	// ModelDeprecatedCanonicalIDs) is a stored value the picker no longer
	// offers — the deprecated-id shape case (1) must normalize.
	var deprecatedModel string
	for id := range template.ModelDeprecatedCanonicalIDs {
		deprecatedModel = id
		break
	}
	if deprecatedModel == "" {
		t.Fatal("template.ModelDeprecatedCanonicalIDs carries no entries; case (1) cannot seed a deprecated model")
	}

	for _, tc := range []struct {
		name string

		// stored seeds preferences.yaml before the run (nil = fresh store).
		stored *profile.ProfilePreferences
		// withProject seeds a MoAI project cwd (case (3) needs quality.yaml).
		withProject bool
		// answers overrides the injected answer set (nil = absorbAnswers).
		answers func(a wizard.ProfileResult) wizard.ProfileResult

		// wantCapture asserts the binding the wizard received.
		wantCapture func(t *testing.T, cap *runnerCapture)
		// wantAfter asserts the post-run state; snapshot holds the pre-save
		// file digests.
		wantAfter func(t *testing.T, run *absorbRun, snapshot map[string]string)
	}{
		{
			name:        "deprecated_model_id_normalized_before_binding",
			withProject: true,
			stored:      &profile.ProfilePreferences{Model: deprecatedModel},
			wantCapture: func(t *testing.T, cap *runnerCapture) {
				t.Helper()
				if cap.initial.Model == "" {
					t.Fatal("initial model is empty; the deprecated stored id must pre-select a normalized alias")
				}
				if cap.initial.Model == deprecatedModel {
					t.Fatalf("initial model %q is still the deprecated stored id; it must be normalized before binding", deprecatedModel)
				}
				if want := normalizeModel(deprecatedModel); cap.initial.Model != want {
					t.Fatalf("initial model = %q, want the normalized alias %q", cap.initial.Model, want)
				}
			},
		},
		{
			name:        "acceptEdits_stores_empty_and_prints_confirmation_once",
			withProject: true,
			answers: func(a wizard.ProfileResult) wizard.ProfileResult {
				a.PermissionMode = "acceptEdits"
				return a
			},
			wantAfter: func(t *testing.T, run *absorbRun, _ map[string]string) {
				t.Helper()
				prefs, err := profile.ReadPreferences("default")
				if err != nil {
					t.Fatalf("ReadPreferences: %v", err)
				}
				if prefs.PermissionMode != "" {
					t.Errorf("saved permission mode = %q, want the empty string", prefs.PermissionMode)
				}
				// Counted via the locale-stable anchor tokens: the notice is
				// localized per the wizard's ending locale (REQ-TRI-006), so
				// the English sentence is not guaranteed to be the one that
				// rendered. The "acceptEdits"+"settings.local.json" token pair
				// survives verbatim in every locale.
				notices := 0
				for _, line := range strings.Split(run.stdout.String(), "\n") {
					if strings.Contains(line, "acceptEdits") && strings.Contains(line, "settings.local.json") {
						notices++
					}
				}
				if notices != 1 {
					t.Errorf("acceptEdits confirmation line printed %d times, want exactly 1; output:\n%s", notices, run.stdout.String())
				}
			},
		},
		{
			name:        "development_mode_initial_from_project_quality",
			withProject: true,
			wantCapture: func(t *testing.T, cap *runnerCapture) {
				t.Helper()
				if cap.initial.DevelopmentMode != "tdd" {
					t.Errorf("development mode initial = %q, want the seeded quality.yaml value %q", cap.initial.DevelopmentMode, "tdd")
				}
			},
		},
		{
			name:        "stored_statusline_segments_survive_save",
			withProject: true,
			stored: &profile.ProfilePreferences{
				StatuslineSegments: map[string]bool{"model": true, "context": false, "git_branch": true},
			},
			wantAfter: func(t *testing.T, _ *absorbRun, _ map[string]string) {
				t.Helper()
				prefs, err := profile.ReadPreferences("default")
				if err != nil {
					t.Fatalf("ReadPreferences: %v", err)
				}
				if prefs.StatuslineSegments == nil {
					t.Fatal("statusline_segments was blanked by the save; the stored map must be carried through")
				}
				for _, key := range []string{"model", "context", "git_branch"} {
					if _, ok := prefs.StatuslineSegments[key]; !ok {
						t.Errorf("statusline_segments lost key %q; got %v", key, prefs.StatuslineSegments)
					}
				}
			},
		},
		{
			name:        "stored_selects_preselected",
			withProject: true,
			stored: &profile.ProfilePreferences{
				GitCommitLang:   "en",
				CodeCommentLang: "zh",
				DocLang:         "ko",
				ModelPolicy:     "low",
				EffortLevel:     "max",
			},
			wantCapture: func(t *testing.T, cap *runnerCapture) {
				t.Helper()
				want := map[string]string{
					"GitCommitLang":   "en",
					"CodeCommentLang": "zh",
					"DocLang":         "ko",
					"ModelPolicy":     "low",
					"EffortLevel":     "max",
				}
				got := map[string]string{
					"GitCommitLang":   cap.initial.GitCommitLang,
					"CodeCommentLang": cap.initial.CodeCommentLang,
					"DocLang":         cap.initial.DocLang,
					"ModelPolicy":     cap.initial.ModelPolicy,
					"EffortLevel":     cap.initial.EffortLevel,
				}
				for field, wantVal := range want {
					if got[field] != wantVal {
						t.Errorf("%s initial = %q, want the stored value %q", field, got[field], wantVal)
					}
				}
			},
		},
		{
			name:        "statusline_yaml_byte_identical",
			withProject: true,
			wantAfter: func(t *testing.T, _ *absorbRun, snapshot map[string]string) {
				t.Helper()
				before, ok := snapshot["statusline.yaml"]
				if !ok {
					t.Fatal("no pre-save snapshot for statusline.yaml")
				}
				if after := sha256File(t, filepath.Join(".moai", "config", "sections", "statusline.yaml")); after != before {
					t.Error("statusline.yaml changed across the save; it must stay byte-identical")
				}
			},
		},
		{
			name:        "git_convention_yaml_byte_identical",
			withProject: true,
			wantAfter: func(t *testing.T, _ *absorbRun, snapshot map[string]string) {
				t.Helper()
				before, ok := snapshot["git-convention.yaml"]
				if !ok {
					t.Fatal("no pre-save snapshot for git-convention.yaml")
				}
				if after := sha256File(t, filepath.Join(".moai", "config", "sections", "git-convention.yaml")); after != before {
					t.Error("git-convention.yaml changed across the save; it must stay byte-identical")
				}
			},
		},
		{
			name:        "no_statusline_theme_written",
			withProject: true,
			stored:      &profile.ProfilePreferences{StatuslineTheme: "catppuccin-mocha"},
			wantAfter: func(t *testing.T, _ *absorbRun, _ map[string]string) {
				t.Helper()
				data, err := os.ReadFile(profile.GetPreferencesPath("default"))
				if err != nil {
					t.Fatalf("read preferences.yaml: %v", err)
				}
				if strings.Contains(string(data), "statusline_theme") {
					t.Errorf("preferences.yaml carries a statusline_theme key; the wizard must never write one:\n%s", data)
				}
			},
		},
		{
			name:        "empty_stored_permission_preselects_acceptEdits",
			withProject: true,
			stored:      &profile.ProfilePreferences{PermissionMode: ""},
			wantCapture: func(t *testing.T, cap *runnerCapture) {
				t.Helper()
				if cap.initial.PermissionMode != defaultPermissionMode {
					t.Errorf("permission mode initial = %q, want %q (the empty stored value pre-selects the default)", cap.initial.PermissionMode, defaultPermissionMode)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.withProject {
				seedAbsorbProject(t)
			} else {
				seedBareDir(t)
			}
			isolateProfileStore(t)
			if tc.stored != nil {
				if err := profile.WritePreferences("default", *tc.stored); err != nil {
					t.Fatalf("seed stored prefs: %v", err)
				}
			}

			// Pre-save byte snapshot of the three project files the cases
			// reason about.
			snapshot := map[string]string{}
			for _, name := range []string{"statusline.yaml", "git-convention.yaml", "quality.yaml"} {
				path := filepath.Join(".moai", "config", "sections", name)
				if _, err := os.Stat(path); err == nil {
					snapshot[name] = sha256File(t, path)
				}
			}

			answers := absorbAnswers()
			if tc.answers != nil {
				answers = tc.answers(answers)
			}

			run := runAbsorbedSetup(t, "setup", "", false, func(cap *runnerCapture) (*wizard.ProfileResult, error) {
				a := answers
				return &a, nil
			})
			if run.err != nil {
				t.Fatalf("profile setup: %v (stderr: %s)", run.err, run.stderr.String())
			}
			if run.captured.calls != 1 {
				t.Fatalf("profileWizardRunner calls = %d, want 1", run.captured.calls)
			}
			if tc.wantCapture != nil {
				tc.wantCapture(t, &run.captured)
			}
			if tc.wantAfter != nil {
				tc.wantAfter(t, run, snapshot)
			}
		})
	}
}

// TestProfileSetupAbsorbed_SavePersistsAcrossSurfaces is AC-ITI-005's M5 half
// (the storage binding; the option-set clause (4) is M1's and lives in
// profile_options_test.go). Injected answers that differ from every seed must
// land in preferences.yaml, the project user/language sections, quality.yaml
// — and outside a project only preferences.yaml may appear.
func TestProfileSetupAbsorbed_SavePersistsAcrossSurfaces(t *testing.T) {
	t.Run("project_cwd_persists_all_surfaces", func(t *testing.T) {
		root := seedAbsorbProject(t)
		isolateProfileStore(t)
		run := runAbsorbedSetup(t, "setup", "", false, func(_ *runnerCapture) (*wizard.ProfileResult, error) {
			a := absorbAnswers()
			return &a, nil
		})
		if run.err != nil {
			t.Fatalf("profile setup: %v (stderr: %s)", run.err, run.stderr.String())
		}

		// (1) preferences.yaml carries the injected answers.
		prefs, err := profile.ReadPreferences("default")
		if err != nil {
			t.Fatalf("ReadPreferences: %v", err)
		}
		want := absorbAnswers()
		if prefs.UserName != want.UserName ||
			prefs.ConversationLang != want.ConversationLang ||
			prefs.GitCommitLang != want.GitCommitLang ||
			prefs.CodeCommentLang != want.CodeCommentLang ||
			prefs.DocLang != want.DocLang ||
			prefs.Model != want.Model ||
			prefs.ModelPolicy != want.ModelPolicy ||
			prefs.EffortLevel != want.EffortLevel ||
			prefs.PermissionMode != want.PermissionMode {
			t.Errorf("preferences.yaml mismatch:\n got %+v\nwant %+v", prefs, want)
		}

		// (2) user.yaml + language.yaml carry the injected values.
		mgr := config.NewConfigManager()
		cfg, err := mgr.LoadRaw(root)
		if err != nil {
			t.Fatalf("LoadRaw: %v", err)
		}
		if cfg.User.Name != want.UserName {
			t.Errorf("user.yaml name = %q, want %q", cfg.User.Name, want.UserName)
		}
		if cfg.Language.ConversationLanguage != want.ConversationLang ||
			cfg.Language.ConversationLanguageName != want.ConversationLang ||
			cfg.Language.GitCommitMessages != want.GitCommitLang ||
			cfg.Language.CodeComments != want.CodeCommentLang ||
			cfg.Language.Documentation != want.DocLang {
			t.Errorf("language.yaml mismatch: %+v", cfg.Language)
		}

		// (3) quality.yaml development_mode carries the injected answer.
		if string(cfg.Quality.DevelopmentMode) != want.DevelopmentMode {
			t.Errorf("quality.yaml development_mode = %q, want %q", cfg.Quality.DevelopmentMode, want.DevelopmentMode)
		}
	})

	t.Run("outside_project_only_preferences_yaml", func(t *testing.T) {
		root := seedBareDir(t)
		isolateProfileStore(t)
		run := runAbsorbedSetup(t, "setup", "", false, func(_ *runnerCapture) (*wizard.ProfileResult, error) {
			a := absorbAnswers()
			return &a, nil
		})
		if run.err != nil {
			t.Fatalf("profile setup: %v (stderr: %s)", run.err, run.stderr.String())
		}
		if _, err := os.Stat(profile.GetPreferencesPath("default")); err != nil {
			t.Errorf("preferences.yaml must be written even outside a project: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, ".moai")); !os.IsNotExist(err) {
			t.Errorf("an outside-project run must not create .moai; stat err = %v", err)
		}
	})
}

// TestProfileSetup_CommandContractOverSeams is AC-ITI-007: both explicit
// entries, each driven with an injected wizard outcome — (a) user cancel,
// (b) error, (c) success with no name and with "work". The cleanup seam must
// fire exactly once per run with the right clean-exit argument, and the enter
// seam must precede the profile read (the sentinel clause).
func TestProfileSetup_CommandContractOverSeams(t *testing.T) {
	for _, entry := range []string{"setup", "profile"} {
		for _, tc := range []struct {
			name         string
			profileName  string // "" = no name argument
			withSentinel bool
			answersFn    func(*runnerCapture) (*wizard.ProfileResult, error)

			wantErr       bool
			wantOut       func(t *testing.T, run *absorbRun)
			wantFile      bool // preferences.yaml exists after the run
			wantCleanExit bool
		}{
			{
				name:         "cancel",
				withSentinel: false,
				answersFn: func(_ *runnerCapture) (*wizard.ProfileResult, error) {
					return &wizard.ProfileResult{ConversationLang: "ko"}, wizard.ErrCancelled
				},
				wantErr:       false,
				wantFile:      false,
				wantCleanExit: true,
				wantOut: func(t *testing.T, run *absorbRun) {
					t.Helper()
					want := getProfileText("ko").SetupCancelled
					if !strings.Contains(run.stdout.String(), want) {
						t.Errorf("cancellation output lacks the ko %q; output:\n%s", want, run.stdout.String())
					}
				},
			},
			{
				name:         "error",
				withSentinel: false,
				answersFn: func(_ *runnerCapture) (*wizard.ProfileResult, error) {
					return nil, errors.New("injected wizard boom")
				},
				wantErr:       true,
				wantFile:      false,
				wantCleanExit: false,
			},
			{
				name:         "success-default",
				profileName:  "",
				withSentinel: true,
				answersFn: func(_ *runnerCapture) (*wizard.ProfileResult, error) {
					a := absorbAnswers()
					return &a, nil
				},
				wantErr:       false,
				wantFile:      true,
				wantCleanExit: true,
				wantOut: func(t *testing.T, run *absorbRun) {
					t.Helper()
					a := absorbAnswers()
					want := fmt.Sprintf(getProfileText(a.ConversationLang).SavedProfile, "default", profile.GetPreferencesPath("default"))
					if !strings.Contains(run.stdout.String(), want) {
						t.Errorf("success output lacks the saved line %q; output:\n%s", want, run.stdout.String())
					}
					if !strings.Contains(run.stdout.String(), getProfileText(a.ConversationLang).SummaryHeader) {
						t.Errorf("success output lacks the summary header; output:\n%s", run.stdout.String())
					}
				},
			},
			{
				name:         "success-work",
				profileName:  "work",
				withSentinel: true,
				answersFn: func(_ *runnerCapture) (*wizard.ProfileResult, error) {
					a := absorbAnswers()
					return &a, nil
				},
				wantErr:       false,
				wantFile:      true,
				wantCleanExit: true,
				wantOut: func(t *testing.T, run *absorbRun) {
					t.Helper()
					a := absorbAnswers()
					want := fmt.Sprintf(getProfileText(a.ConversationLang).SavedProfile, "work", profile.GetPreferencesPath("work"))
					if !strings.Contains(run.stdout.String(), want) {
						t.Errorf("success output lacks the saved line %q; output:\n%s", want, run.stdout.String())
					}
				},
			},
		} {
			t.Run(entry+"/"+tc.name, func(t *testing.T) {
				seedBareDir(t)
				isolateProfileStore(t)
				run := runAbsorbedSetup(t, entry, tc.profileName, tc.withSentinel, tc.answersFn)

				if tc.wantErr && run.err == nil {
					t.Error("command must return a non-nil error on the error outcome")
				}
				if !tc.wantErr && run.err != nil {
					t.Errorf("command error = %v, want nil (stderr: %s)", run.err, run.stderr.String())
				}

				// The wizard seam ran exactly once.
				if run.captured.calls != 1 {
					t.Errorf("profileWizardRunner calls = %d, want 1", run.captured.calls)
				}

				// Profile file state.
				name := tc.profileName
				if name == "" {
					name = "default"
				}
				_, statErr := os.Stat(profile.GetPreferencesPath(name))
				if tc.wantFile && statErr != nil {
					t.Errorf("preferences.yaml must exist after a successful save: %v", statErr)
				}
				if !tc.wantFile && statErr == nil {
					t.Error("preferences.yaml must not be created on a cancelled or errored run")
				}

				// The cleanup seam: exactly once per run, with the clean-exit
				// argument true on success and cancellation, false on error.
				wantLog := []string{"enter", fmt.Sprintf("cleanup:%v", tc.wantCleanExit)}
				if len(run.wtLog) != len(wantLog) {
					t.Errorf("worktree seam log = %v, want %v", run.wtLog, wantLog)
				} else {
					for i := range wantLog {
						if run.wtLog[i] != wantLog[i] {
							t.Errorf("worktree seam log[%d] = %q, want %q (full log %v)", i, run.wtLog[i], wantLog[i], run.wtLog)
						}
					}
				}

				// The enter seam preceded the profile read: the sentinel the
				// enter seam wrote is what the read handed the wizard.
				if tc.withSentinel && run.captured.initial.UserName != sentinelUserName {
					t.Errorf("captured initial user name = %q, want the sentinel %q — the profile read did not observe the enter seam's write", run.captured.initial.UserName, sentinelUserName)
				}

				if tc.wantOut != nil {
					tc.wantOut(t, run)
				}
			})
		}
	}
}
