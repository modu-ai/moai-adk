// dotgithub_template_test.go: behavior + parity guards for the .github assets
// that the template tree ships and this repository also deploys at its own
// .github/ root (dogfood copies).
//
// The detect-language composite action is executed for real: the run script
// is extracted from the template action.yml and run under the same shell
// options GitHub Actions uses for `shell: bash` (`-e -o pipefail`), against
// t.TempDir() fixtures, with GITHUB_OUTPUT captured. The cases pin the three
// behaviors the template was fixed for: `.exs` mapping to elixir (it was
// detected but unmapped, so Elixir-script projects resolved to unknown),
// manifest preference over first-file-found, and multi-language output for
// mixed repositories.
//
// The script is also verified to parse and run under macOS's bash 3.2 — the
// oldest bash a composite action can meet on a macos runner — which is why
// the extension->language case lives in a lang_of() function instead of
// inline in the $( ) command substitution (bash 3.2's parser rejects a case
// statement inside command substitution).
//
// Sentinel on failure: DETECT_LANGUAGE_BEHAVIOR / LABEL_SYNC_PUSH_BRANCHES /
// DOTGITHUB_PARITY_DRIFT.
//
// Design notes:
//   - Like release_workflow_pipefail_test.go and branch_protection_parity_
//     test.go, this lives under internal/template/ (NOT internal/template/
//     templates/), so it is not user-distributed template content and does
//     not trigger the template-neutrality CI guard.
//   - The parity table covers the three .github file pairs that exist in
//     both trees; branch-protection already has its own dedicated guard
//     (TestBranchProtectionParity).
package template_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// detectActionRelPath is the template source of the composite action; the
// deployed copy at .github/actions/detect-language/action.yml is its mirror.
const detectActionRelPath = "internal/template/templates/.github/actions/detect-language/action.yml"

// labelSyncRelPath is the template source of the label-sync workflow.
const labelSyncRelPath = "internal/template/templates/.github/workflows/label-sync.yml"

// TestDetectLanguageActionBehavior runs the template action's script against
// fixture repositories and asserts the emitted GITHUB_OUTPUT lines. It fails
// on the pre-fix script for the `.exs`-only case (language=unknown).
func TestDetectLanguageActionBehavior(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("the composite action runs bash; the parity guard below is platform-neutral")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}

	script := detectActionRunScript(t)

	tests := []struct {
		name          string
		files         map[string]string
		wantLanguage  string
		wantLanguages string
		why           string
	}{
		{
			name:          "exs-only project maps to elixir",
			files:         map[string]string{"lib/app.exs": "x", "run.exs": "x"},
			wantLanguage:  "elixir",
			wantLanguages: "elixir",
			why:           "the reproduced defect: .exs was detected but unmapped, resolving an Elixir-scripts project to unknown",
		},
		{
			name:          "ex sources also map to elixir",
			files:         map[string]string{"lib/app.ex": "x"},
			wantLanguage:  "elixir",
			wantLanguages: "elixir",
			why:           "the pre-existing .ex mapping must survive the fix",
		},
		{
			name:          "mix.exs manifest detects elixir",
			files:         map[string]string{"mix.exs": "x"},
			wantLanguage:  "elixir",
			wantLanguages: "elixir",
			why:           "an Elixir project with no sources yet is still declared by its manifest",
		},
		{
			name:          "manifest beats first file found",
			files:         map[string]string{"go.mod": "x", "aaa.py": "x"},
			wantLanguage:  "go",
			wantLanguages: "go,python",
			why:           "a root manifest declares intent; the first file the scan happens to return must not win the primary pick",
		},
		{
			name:          "mixed project lists every language",
			files:         map[string]string{"a.go": "x", "b.go": "x", "c/main.go": "x", "app.ts": "x"},
			wantLanguage:  "go",
			wantLanguages: "go,typescript",
			why:           "a mixed repository reports all of its languages instead of collapsing to the first file found",
		},
		{
			name:          "no-manifest primary follows file count",
			files:         map[string]string{"src/Main.kt": "x", "src/Util.kt": "x", "one.swift": "x"},
			wantLanguage:  "kotlin",
			wantLanguages: "kotlin,swift",
			why:           "with no manifest the primary is the most-matched language, not whichever file sorted first",
		},
		{
			name:          "tsconfig makes the package a typescript project",
			files:         map[string]string{"tsconfig.json": "x", "package.json": "x", "src/i.js": "x"},
			wantLanguage:  "typescript",
			wantLanguages: "typescript,javascript",
			why:           "tsconfig.json subsumes the package.json hit for the primary pick; the .js sources still surface javascript",
		},
		{
			name:          "Groovy Gradle Kotlin project",
			files:         map[string]string{"build.gradle": "plugins { id 'org.jetbrains.kotlin.jvm' }", "src/Main.kt": "x"},
			wantLanguage:  "kotlin",
			wantLanguages: "kotlin",
			why:           "Gradle DSL files are shared by Java and Kotlin; sources determine their language",
		},
		{
			name:          "Kotlin DSL Java project",
			files:         map[string]string{"build.gradle.kts": "plugins { java }", "src/Main.java": "x"},
			wantLanguage:  "java",
			wantLanguages: "java",
			why:           "Gradle DSL files are shared by Java and Kotlin; sources determine their language",
		},
		{
			name:          "Kotlin DSL Kotlin project",
			files:         map[string]string{"build.gradle.kts": "plugins { kotlin(\"jvm\") }", "src/Main.kt": "x"},
			wantLanguage:  "kotlin",
			wantLanguages: "kotlin",
			why:           "Gradle DSL files are shared by Java and Kotlin; sources determine their language",
		},
		{
			name:          "Groovy Gradle Java project",
			files:         map[string]string{"build.gradle": "plugins { id \"java\" }", "src/Main.java": "x"},
			wantLanguage:  "java",
			wantLanguages: "java",
			why:           "Gradle DSL files are shared by Java and Kotlin; sources determine their language",
		},
		{
			name:          "mixed Gradle follows JVM source counts",
			files:         map[string]string{"build.gradle": "x", "src/Main.kt": "x", "src/Util.kt": "x", "src/Interop.java": "x"},
			wantLanguage:  "kotlin",
			wantLanguages: "kotlin,java",
			why:           "Gradle DSL files are shared by Java and Kotlin; sources determine their language",
		},
		{
			name:          "source-less Gradle retains Java fallback",
			files:         map[string]string{"build.gradle": "x"},
			wantLanguage:  "java",
			wantLanguages: "java",
			why:           "Gradle DSL files are shared by Java and Kotlin; sources determine their language",
		},
		{
			name:          "source-less Kotlin DSL retains unknown fallback",
			files:         map[string]string{"build.gradle.kts": "x"},
			wantLanguage:  "unknown",
			wantLanguages: "",
			why:           "Gradle DSL files are shared by Java and Kotlin; sources determine their language",
		},
		{
			name:          "Gradle does not displace earlier manifest",
			files:         map[string]string{"go.mod": "x", "build.gradle": "x", "src/Main.kt": "x"},
			wantLanguage:  "go",
			wantLanguages: "go,kotlin",
			why:           "Gradle DSL files are shared by Java and Kotlin; sources determine their language",
		},
		{
			name:          "dependency Kotlin does not affect Gradle",
			files:         map[string]string{"build.gradle": "x", "vendor/Main.kt": "x", "src/Main.java": "x"},
			wantLanguage:  "java",
			wantLanguages: "java",
			why:           "Gradle DSL files are shared by Java and Kotlin; sources determine their language",
		},
		{
			name:          "nothing matches reports unknown",
			files:         map[string]string{"README.md": "x"},
			wantLanguage:  "unknown",
			wantLanguages: "",
			why:           "the pre-existing unknown fallback and empty languages list",
		},
		{
			name:          "vendor and node_modules are pruned",
			files:         map[string]string{"main.go": "x", "node_modules/x/main.js": "x", "vendor/y/lib.rs": "x"},
			wantLanguage:  "go",
			wantLanguages: "go",
			why:           "dependency trees must not contribute languages",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			for name, content := range tc.files {
				path := filepath.Join(dir, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", name, err)
				}
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatalf("write %s: %v", name, err)
				}
			}

			outputs := runDetectActionScript(t, script, dir)
			if got := outputs["language"]; got != tc.wantLanguage {
				t.Errorf("DETECT_LANGUAGE_BEHAVIOR: language = %q, want %q — %s", got, tc.wantLanguage, tc.why)
			}
			if got := outputs["languages"]; got != tc.wantLanguages {
				t.Errorf("DETECT_LANGUAGE_BEHAVIOR: languages = %q, want %q — %s", got, tc.wantLanguages, tc.why)
			}
		})
	}
}

// detectActionSpec mirrors the subset of the composite-action schema the
// tests consume: the runs.steps[].run script of the detect step.
type detectActionSpec struct {
	Runs struct {
		Steps []struct {
			ID  string `yaml:"id"`
			Run string `yaml:"run"`
		} `yaml:"steps"`
	} `yaml:"runs"`
}

// detectActionRunScript loads the template action.yml and returns the detect
// step's script. YAML round-trip (rather than regex extraction) so the script
// is exactly what a runner would receive after block-scalar dedent.
func detectActionRunScript(t *testing.T) string {
	t.Helper()

	root := findProjectRootForMirrorTest(t)
	data, err := os.ReadFile(filepath.Join(root, detectActionRelPath))
	if err != nil {
		t.Fatalf("read %s: %v", detectActionRelPath, err)
	}

	var spec detectActionSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatalf("parse %s: %v", detectActionRelPath, err)
	}
	for _, step := range spec.Runs.Steps {
		if step.ID == "detect" {
			if strings.TrimSpace(step.Run) == "" {
				t.Fatalf("%s: detect step has an empty run script", detectActionRelPath)
			}
			return step.Run
		}
	}
	t.Fatalf("%s: no step with id \"detect\" found", detectActionRelPath)
	return ""
}

// runDetectActionScript writes script to a scratch file and executes it in
// dir with GitHub's `shell: bash` options (`-e -o pipefail`), returning the
// parsed GITHUB_OUTPUT key=value pairs.
func runDetectActionScript(t *testing.T, script, dir string) map[string]string {
	t.Helper()

	outDir := t.TempDir()
	scriptPath := filepath.Join(outDir, "detect.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	outputPath := filepath.Join(outDir, "github-output.txt")

	cmd := exec.Command("bash", "--noprofile", "--norc", "-e", "-o", "pipefail", scriptPath)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GITHUB_OUTPUT="+outputPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("DETECT_LANGUAGE_BEHAVIOR: script failed in %s: %v\noutput:\n%s", dir, err, out)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read GITHUB_OUTPUT: %v", err)
	}
	outputs := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		if key, value, ok := strings.Cut(line, "="); ok {
			outputs[key] = value
		}
	}
	return outputs
}

// TestLabelSyncPushBranches pins the label-sync workflow's automatic push
// trigger to the repository's branch set: main (default branch) and develop
// (integration branch, where label edits can land first via merged card
// branches). The workflow previously fired on main only, so a labels.yml
// edit merged to develop did not sync until a release merge carried it to
// main.
func TestLabelSyncPushBranches(t *testing.T) {
	t.Parallel()

	root := findProjectRootForMirrorTest(t)
	data, err := os.ReadFile(filepath.Join(root, labelSyncRelPath))
	if err != nil {
		t.Fatalf("read %s: %v", labelSyncRelPath, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse %s: %v", labelSyncRelPath, err)
	}

	// The `on:` key parses as a boolean scalar in YAML 1.1 flow, but the
	// yaml.Node form keeps the raw key text, so compare node values.
	branches := mappingSequenceStrings(t, mappingValue(t, mappingValue(t, doc.Content[0], "on"), "push"), "branches")

	have := map[string]bool{}
	for _, b := range branches {
		have[b] = true
	}
	for _, want := range []string{"main", "develop"} {
		if !have[want] {
			t.Errorf(
				"LABEL_SYNC_PUSH_BRANCHES: %s on.push.branches = %v, missing %q — label edits landing on the integration branch would not sync until a release merge carries them to main",
				labelSyncRelPath, branches, want,
			)
		}
	}
}

// mappingValue returns the value node for key in a YAML mapping node.
func mappingValue(t *testing.T, mapping *yaml.Node, key string) *yaml.Node {
	t.Helper()

	if mapping == nil || mapping.Kind != yaml.MappingNode {
		t.Fatalf("expected a mapping node to look up %q, got kind %v", key, mapping.Kind)
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	t.Fatalf("key %q not found in mapping", key)
	return nil
}

// mappingSequenceStrings returns the scalar strings of a sequence node found
// at key in a YAML mapping node.
func mappingSequenceStrings(t *testing.T, mapping *yaml.Node, key string) []string {
	t.Helper()

	seq := mappingValue(t, mapping, key)
	if seq.Kind != yaml.SequenceNode {
		t.Fatalf("key %q is not a sequence (kind %v)", key, seq.Kind)
	}
	values := make([]string, 0, len(seq.Content))
	for _, item := range seq.Content {
		values = append(values, item.Value)
	}
	return values
}

// dotgithubParityPairs are the .github files that exist as a deployed copy
// at the repository root and as a user-distributed template mirror. They
// must stay byte-identical: an edit to one tree without the other ships a
// stale copy to user projects via moai init / moai update (or leaves this
// repository running a template it no longer ships).
var dotgithubParityPairs = []struct{ deployed, mirror string }{
	{".github/labels.yml", "internal/template/templates/.github/labels.yml"},
	{".github/workflows/label-sync.yml", "internal/template/templates/.github/workflows/label-sync.yml"},
	{".github/actions/detect-language/action.yml", "internal/template/templates/.github/actions/detect-language/action.yml"},
}

// TestDotgithubDeployedParity asserts byte equality for each deployed/mirror
// pair, mirroring TestBranchProtectionParity's pattern.
func TestDotgithubDeployedParity(t *testing.T) {
	t.Parallel()

	root := findProjectRootForMirrorTest(t)

	for _, pair := range dotgithubParityPairs {
		pair := pair
		t.Run(pair.deployed, func(t *testing.T) {
			t.Parallel()

			deployedPath := filepath.Join(root, pair.deployed)
			mirrorPath := filepath.Join(root, pair.mirror)

			deployed, err := os.ReadFile(deployedPath)
			if err != nil {
				t.Fatalf("DOTGITHUB_PARITY_DRIFT: deployed copy unreadable %s: %v", deployedPath, err)
			}
			mirror, err := os.ReadFile(mirrorPath)
			if err != nil {
				t.Errorf("DOTGITHUB_PARITY_DRIFT: %s has no template mirror at %s; run 'cp %s %s' and stage both files",
					pair.deployed, mirrorPath, deployedPath, mirrorPath)
				return
			}
			if !bytes.Equal(deployed, mirror) {
				t.Errorf("DOTGITHUB_PARITY_DRIFT: %s differs from its mirror %s (%d vs %d bytes); run 'cp %s %s' and stage both files",
					pair.deployed, pair.mirror, len(deployed), len(mirror), deployedPath, mirrorPath)
			}
		})
	}
}
