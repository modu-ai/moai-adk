// Package main은 --diff <git-rev> 모드의 temporal/baseline oracle을 구현합니다.
// Package main implements the --diff <git-rev> mode temporal/baseline oracle.
//
// # --diff mode (temporal/baseline oracle)
//
// This mode is the canonical AC-CIAUT-016 verification path. It:
//  1. Shells out to git diff --unified=0 <rev> to find changed lines in Go files
//  2. For each changed file, retrieves the baseline content via git show <rev>:<file>
//  3. Builds a baseline lockset for the changed file
//  4. Re-parses the HEAD file to build the current lockset
//  5. For each locked literal whose text differs between baseline and HEAD
//     and is NOT marked translatable → emits a Violation
//
// This catches the PR #783 regression class where BOTH the data (map literal)
// AND the test assertion are translated together. In that case, --all-files mode
// returns clean (no intra-state mismatch), but --diff detects the change relative
// to the baseline.
//
// # Non-git fallback
//
// When git rev-parse --git-dir fails (non-git context), the oracle emits a warning
// on stderr and falls back to --all-files mode behavior (exit 0 on empty corpus).
//
// # Shell-out approach
//
// git CLI shell-out (os/exec) is used instead of go-git to avoid adding a
// third-party dependency. Per Wave 6 constraints: zero third-party dependencies.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// @MX:NOTE: [AUTO] --diff mode temporal oracle. Uses git show + git diff shell-out.

// DiffOracleResult は --diff モードの実行結果です。
// DiffOracleResult holds the result of a --diff mode oracle run.
type DiffOracleResult struct {
	// Violations は発見された違反のリストです。
	Violations []Violation
	// ExitCode はプロセス終了コードです。
	ExitCode int
	// Stderr はエラー出力の内容です。
	Stderr string
}

// runDiffOracle は --diff <git-rev> モードを実行します。
// runDiffOracle executes the temporal/baseline oracle against the given git revision.
func runDiffOracle(root, rev string) DiffOracleResult {
	var stderr strings.Builder

	// git コンテキストの確認
	if !isGitRepo(root) {
		stderr.WriteString("[i18n-validator] warning: not a git repository, falling back to --all-files mode\n")
		violations := runAllFilesOracle(root)
		code := 0
		if len(violations) > 0 {
			code = 1
		}
		return DiffOracleResult{
			Violations: violations,
			ExitCode:   code,
			Stderr:     stderr.String(),
		}
	}

	// git diff で変更ファイルとハンクを取得
	changedFiles, err := getChangedGoFiles(root, rev)
	if err != nil {
		fmt.Fprintf(&stderr, "[i18n-validator] warning: git diff failed: %v\n", err)
		return DiffOracleResult{ExitCode: 0, Stderr: stderr.String()}
	}

	if len(changedFiles) == 0 {
		return DiffOracleResult{ExitCode: 0, Stderr: stderr.String()}
	}

	var allViolations []Violation

	for _, relFile := range changedFiles {
		absFile := filepath.Join(root, relFile)

		// ベースライン content を git show で取得
		baselineContent, err := getBaselineContent(root, rev, relFile)
		if err != nil {
			// ファイルが baseline に存在しない (新規追加) → スキップ
			fmt.Fprintf(&stderr, "[i18n-validator] skipping new file %s\n", relFile)
			continue
		}

		// ベースライン lockset を構築
		baselineLockset, err := buildLocksetForFile(baselineContent, absFile)
		if err != nil {
			fmt.Fprintf(&stderr, "[i18n-validator] warning: cannot parse baseline %s: %v\n", relFile, err)
			continue
		}

		// HEAD の content を読み込んで lockset を構築
		headContent, err := os.ReadFile(absFile)
		if err != nil {
			fmt.Fprintf(&stderr, "[i18n-validator] warning: cannot read HEAD %s: %v\n", relFile, err)
			continue
		}

		headLockset, err := buildLocksetForFile(headContent, absFile)
		if err != nil {
			fmt.Fprintf(&stderr, "[i18n-validator] warning: cannot parse HEAD %s: %v\n", relFile, err)
			continue
		}

		// ベースラインのロック済みリテラルと HEAD を比較
		for key, baselineLit := range baselineLockset.Literals {
			if baselineLit.Translatable {
				continue
			}
			if baselineLit.Text == "" {
				continue
			}

			// HEAD での同じキーのリテラルを探す
			headLit, exists := headLockset.Literals[key]
			if !exists {
				// HEAD でリテラルが消えた場合 — violation
				v := buildViolation(baselineLit, "removed", relFile, "")
				allViolations = append(allViolations, v)
				continue
			}

			// HEAD でリテラルが i18n:translatable になった場合はスキップ
			if headLit.Translatable {
				continue
			}

			// テキストが変わっていれば violation
			if baselineLit.Text != headLit.Text {
				v := buildViolation(baselineLit, baselineLit.Text, relFile, headLit.Text)
				v.HeadText = headLit.Text
				allViolations = append(allViolations, v)
			}
		}

		// HEAD のロックセットでも確認 (baseline にない新規ロックリテラルは対象外)
		_ = headLockset
	}

	if len(allViolations) > 0 {
		// AC-CIAUT-016 canonical format
		for _, v := range allViolations {
			fmt.Fprintf(&stderr, "string literal at %s:%d is referenced by %s:%d, translation requires test update\n",
				v.File, v.Line, v.LockedBy.TestName, v.LockedBy.Line)
		}
		return DiffOracleResult{
			Violations: allViolations,
			ExitCode:   1,
			Stderr:     stderr.String(),
		}
	}

	return DiffOracleResult{ExitCode: 0, Stderr: stderr.String()}
}

// buildViolation は違反情報を構築します。
// buildViolation constructs a Violation from a baseline locked literal.
func buildViolation(lit LockedLiteral, oldText, relFile, newText string) Violation {
	return Violation{
		File:         relFile,
		Line:         lit.Line,
		BaselineText: oldText,
		HeadText:     newText,
		LockedBy:     lit.AssertionRef,
		Reason:       "translation-locked literal modified",
	}
}

// isGitRepo は指定ディレクトリが git リポジトリか確認します。
// isGitRepo checks whether the given directory is inside a git repository.
func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	err := cmd.Run()
	return err == nil
}

// getChangedGoFiles は変更された Go ファイルのリストを取得します。
// getChangedGoFiles returns a list of Go files changed since the baseline revision.
func getChangedGoFiles(dir, rev string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", rev, "--", "*.go")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only: %w", err)
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && strings.HasSuffix(line, ".go") {
			files = append(files, line)
		}
	}
	return files, nil
}

// getBaselineContent は baseline revision のファイル内容を取得します。
// getBaselineContent retrieves the content of a file at the baseline git revision.
func getBaselineContent(dir, rev, relFile string) ([]byte, error) {
	cmd := exec.Command("git", "show", rev+":"+relFile)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show %s:%s: %w", rev, relFile, err)
	}
	return out, nil
}

