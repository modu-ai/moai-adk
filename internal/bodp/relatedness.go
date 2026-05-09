// Package bodp implements the Branch Origin Decision Protocol (BODP).
//
// @MX:NOTE BODP는 새 슬래시 명령어/CLI 서브명령어 ZERO 원칙. 3개 entry point
// 공유 라이브러리 (`/moai plan --branch`, `/moai plan --worktree`, `moai worktree new`).
// SPEC: SPEC-V3R3-CI-AUTONOMY-001 W7-T01 (BODP relatedness check).
package bodp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Choice represents the BODP recommendation outcome.
type Choice string

// Choice values.
const (
	ChoiceMain     Choice = "main"
	ChoiceStacked  Choice = "stacked"
	ChoiceContinue Choice = "continue"
)

// EntryPoint identifies which invocation path triggered the BODP check.
type EntryPoint string

// EntryPoint values.
const (
	EntryPlanBranch   EntryPoint = "plan-branch"
	EntryPlanWorktree EntryPoint = "plan-worktree"
	EntryWorktreeCLI  EntryPoint = "worktree-cli"
)

// CheckInput is the input passed to Check().
type CheckInput struct {
	CurrentBranch string
	NewSpecID     string
	RepoRoot      string
	EntryPoint    EntryPoint
}

// BODPDecision is the structured output from Check().
type BODPDecision struct {
	SignalA     bool
	SignalB     bool
	SignalC     bool
	Recommended Choice
	Rationale   string
	BaseBranch  string
}

// DefaultBase is the canonical base branch for fresh BODP recommendations.
const DefaultBase = "origin/main"

// rationaleByChoice maps each Choice to its Korean rationale message.
// Verbatim wording per acceptance.md AC-CIAUT-018/019.
var rationaleByChoice = map[Choice]string{
	ChoiceMain:     "현재 브랜치와 무관한 새 작업이므로 main 분기를 권장합니다.",
	ChoiceStacked:  "현재 브랜치와의 의존성이 감지되어 stacked PR을 권장합니다.",
	ChoiceContinue: "현재 브랜치에 untracked SPEC plan 또는 진행 중 작업이 있어 계속 작업을 권장합니다.",
}

// gotchaWarning is the parent-merge gotcha suffix appended when SignalC fires.
// REQ-CIAUT-047b verbatim mention of CLAUDE.local.md §18.11 Case Study.
const gotchaWarning = " (parent-merge gotcha 주의: CLAUDE.local.md §18.11 Case Study 참조)"

// gitCommand executes a git subcommand. Overridable in tests.
var gitCommand = func(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}

// ghCommand executes a gh subcommand. Overridable in tests.
var ghCommand = func(args ...string) (string, error) {
	out, err := exec.Command("gh", args...).Output()
	return string(out), err
}

// Check evaluates the 3 BODP signals and returns a recommendation.
//
// @MX:ANCHOR Check is the single entry point shared by 3 invocation paths
// (skill plan-branch, skill plan-worktree, worktree-cli). Decision matrix is
// total (no undefined cases); SignalB priority dominates A/C.
func Check(input CheckInput) (BODPDecision, error) {
	d := BODPDecision{
		SignalA: checkSignalA(input),
		SignalB: checkSignalB(input),
		SignalC: checkSignalC(input),
	}
	d.Recommended, d.BaseBranch = applyMatrix(d, input.CurrentBranch)
	d.Rationale = rationaleByChoice[d.Recommended]
	if d.SignalC {
		d.Rationale += gotchaWarning
	}
	return d, nil
}

// applyMatrix decides Recommended Choice + BaseBranch from 3 signals.
//
// 8-row truth table (a=SignalA, b=SignalB, c=SignalC):
//
//	¬a ¬b ¬c → main      @ origin/main
//	 a ¬b ¬c → stacked   @ currentBranch
//	¬a  b ¬c → continue  @ ""
//	¬a ¬b  c → stacked   @ currentBranch
//	 a  b ¬c → continue  @ "" (b dominates)
//	 a ¬b  c → stacked   @ currentBranch
//	¬a  b  c → continue  @ "" (b dominates)
//	 a  b  c → continue  @ "" (b dominates)
//
// SignalB priority: when present, the user's working tree is already engaged
// with the new SPEC; switching contexts would cause confusion.
func applyMatrix(d BODPDecision, currentBranch string) (Choice, string) {
	if d.SignalB {
		return ChoiceContinue, ""
	}
	if d.SignalA || d.SignalC {
		return ChoiceStacked, currentBranch
	}
	return ChoiceMain, DefaultBase
}

// checkSignalA detects code-level dependency: SPEC depends_on heuristic OR
// diff-path overlap with the new SPEC ID.
func checkSignalA(input CheckInput) bool {
	if dependsOn, err := parseDependsOn(input.RepoRoot, input.NewSpecID); err == nil {
		for _, dep := range dependsOn {
			if dep != "" && strings.Contains(input.CurrentBranch, dep) {
				return true
			}
		}
	}
	out, err := gitCommand("diff", "--name-only", DefaultBase+"..HEAD")
	if err != nil {
		return false
	}
	return strings.Contains(out, input.NewSpecID)
}

// checkSignalB detects working-tree co-location: untracked .moai/specs/<NewSpecID>/
// or modified files under that path.
func checkSignalB(input CheckInput) bool {
	out, err := gitCommand("status", "--porcelain")
	if err != nil {
		return false
	}
	target := ".moai/specs/" + input.NewSpecID
	for line := range strings.SplitSeq(out, "\n") {
		if strings.Contains(line, target) {
			return true
		}
	}
	return false
}

// checkSignalC detects an open PR with currentBranch as head. gh missing or
// auth failure → graceful skip (returns false, no error propagation).
func checkSignalC(input CheckInput) bool {
	if input.CurrentBranch == "" {
		return false
	}
	out, err := ghCommand("pr", "list", "--head", input.CurrentBranch, "--state", "open", "--json", "number")
	if err != nil {
		return false
	}
	trimmed := strings.TrimSpace(out)
	return trimmed != "" && trimmed != "[]"
}

// specFrontmatter is the minimal SPEC frontmatter shape this package needs.
type specFrontmatter struct {
	DependsOn []string `yaml:"depends_on"`
}

// parseDependsOn reads .moai/specs/<specID>/spec.md and returns the depends_on
// list from its YAML frontmatter, or an error if the file or frontmatter is
// missing.
func parseDependsOn(repoRoot, specID string) ([]string, error) {
	if repoRoot == "" || specID == "" {
		return nil, fmt.Errorf("parseDependsOn: empty repoRoot or specID")
	}
	specPath := filepath.Join(repoRoot, ".moai", "specs", specID, "spec.md")
	raw, err := os.ReadFile(specPath)
	if err != nil {
		return nil, err
	}
	frontmatter, ok := extractFrontmatter(raw)
	if !ok {
		return nil, fmt.Errorf("no frontmatter in %s", specPath)
	}
	var fm specFrontmatter
	if err := yaml.Unmarshal(frontmatter, &fm); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	return fm.DependsOn, nil
}

// extractFrontmatter returns the YAML body between the leading `---\n` and the
// next `---` delimiter line. Returns (nil, false) when frontmatter is missing.
func extractFrontmatter(raw []byte) ([]byte, bool) {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return nil, false
	}
	rest := text[4:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return nil, false
	}
	// After "\n---" we expect either '\n' or EOF for a valid closing line.
	after := idx + 4
	if after < len(rest) && rest[after] != '\n' {
		return nil, false
	}
	return []byte(rest[:idx]), true
}
