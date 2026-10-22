package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This native formatter accepts real files and deliberately has no portable
// /dev/stdin interface. Parsing a named file keeps gofmt's full-file contract:
// unlike its stdin mode, package-less fragments are not accepted.
const nativeFileFormatter = `package main
import ("encoding/json"; "go/parser"; "go/token"; "os"; "path/filepath")
func main() {
 var body []byte
 var path string
 regular := false
 if len(os.Args) == 3 {
  path, _ = filepath.Abs(os.Args[2])
  body, _ = os.ReadFile(os.Args[2])
  info, err := os.Stat(os.Args[2])
  regular = err == nil && info.Mode().IsRegular()
 }
 record, _ := json.Marshal(struct{ Args []string; Input, Path string; Regular bool }{os.Args[1:], string(body), path, regular})
 if os.WriteFile(os.Getenv("MOAI_FORMATTER_LOG"), record, 0600) != nil { os.Exit(22) }
 if len(os.Args) != 3 || os.Args[1] != "-l" || !regular { os.Exit(23) }
 if _, err := parser.ParseFile(token.NewFileSet(), path, body, parser.AllErrors); err != nil { os.Exit(25) }
 switch os.Getenv("MOAI_FORMATTER_RESULT") {
 case "unformatted": os.Stdout.WriteString(path+"\n")
 case "failure": os.Exit(24)
 }
}
`

func TestPreCommitHookNativeFormatterReadsStagedFile(t *testing.T) {
	binary := writeNativeToolFixture(t, "gofmt", nativeFileFormatter)
	for _, tc := range []struct {
		name, staged, result, hint string
		code                       int
	}{
		{"clean staged bytes", stagedReadFormatted, "clean", "", 0},
		{"unformatted staged bytes", stagedReadFormatted, "unformatted", "need formatting", 1},
		{"formatter failure", stagedReadFormatted, "failure", "could not check the staged bytes", 1},
		{"package-less fragment", "func f() {}\n", "clean", "could not check the staged bytes", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitInitRepo(t)
			stageFile(t, repo, "main.go", tc.staged)
			if err := os.WriteFile(filepath.Join(repo, "main.go"), []byte(stagedReadUnformatted), 0o600); err != nil {
				t.Fatal(err)
			}
			log := filepath.Join(t.TempDir(), "formatter.json")
			code, stderr := runPreCommitHook(t, repo, []string{
				"PATH=" + filepath.Dir(binary) + string(os.PathListSeparator) + os.Getenv("PATH"),
				"MOAI_FORMATTER_LOG=" + log, "MOAI_FORMATTER_RESULT=" + tc.result, "SKIP_MOAI_PRECOMMIT=",
			})
			raw, err := os.ReadFile(log)
			if err != nil {
				t.Fatalf("formatter was not run: %v\n%s", err, stderr)
			}
			var observed struct {
				Args        []string
				Input, Path string
				Regular     bool
			}
			if err := json.Unmarshal(raw, &observed); err != nil {
				t.Fatal(err)
			}
			if len(observed.Args) != 2 || observed.Args[0] != "-l" || !observed.Regular {
				t.Errorf("formatter args=%q regular=%v, want native regular staged file", observed.Args, observed.Regular)
			}
			if observed.Input != tc.staged {
				t.Errorf("formatter input=%q, want staged bytes %q", observed.Input, tc.staged)
			}
			parent, parentErr := os.Stat(filepath.Dir(observed.Path))
			gitdir, gitErr := os.Stat(filepath.Join(repo, ".git"))
			if parentErr != nil || gitErr != nil || !os.SameFile(parent, gitdir) {
				t.Errorf("formatter file must stay in fixture Git metadata: path=%q parent=%v gitdir=%v", observed.Path, parentErr, gitErr)
			}
			if _, err := os.Stat(observed.Path); !os.IsNotExist(err) {
				t.Errorf("staged temporary file remains: %q err=%v", observed.Path, err)
			}
			if code != tc.code || tc.hint != "" && !strings.Contains(stderr, tc.hint) {
				t.Errorf("hook exit=%d want=%d hint=%q\n%s", code, tc.code, tc.hint, stderr)
			}
		})
	}
}

func TestPreCommitHookStagedFragmentFailsFullFileParsing(t *testing.T) {
	repo := gitInitRepo(t)
	stageFile(t, repo, "fragment.go", "func f() {}\n")
	code, stderr := runPreCommitHook(t, repo, []string{"SKIP_MOAI_PRECOMMIT="})
	if code != 1 || !strings.Contains(stderr, "could not check the staged bytes") {
		t.Fatalf("package-less staged file must fail full-file parsing: code=%d\n%s", code, stderr)
	}
}

func TestPreCommitHookStagedTemporaryFileFailureBlocks(t *testing.T) {
	for _, tool := range []string{"mktemp", "rm"} {
		t.Run(tool, func(t *testing.T) {
			binary := writeNativeToolFixture(t, tool, "package main\nimport \"os\"\nfunc main() { os.Exit(37) }\n")
			repo := gitInitRepo(t)
			stageFile(t, repo, "main.go", stagedReadFormatted)
			code, stderr := runPreCommitHook(t, repo, []string{
				"PATH=" + filepath.Dir(binary) + string(os.PathListSeparator) + os.Getenv("PATH"), "SKIP_MOAI_PRECOMMIT=",
			})
			if code != 1 || !strings.Contains(stderr, "could not check the staged bytes") {
				t.Fatalf("%s failure must block: code=%d\n%s", tool, code, stderr)
			}
		})
	}
}
