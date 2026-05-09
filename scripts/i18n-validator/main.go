// Package main은 i18n-validator 도구를 구현합니다.
// Package main implements the i18n-validator tool — a static analyzer that scans Go
// source files for string literals participating in testify assertions and flags them
// as "translation-locked". When a translation PR modifies a locked literal, the
// validator exits non-zero so CI can block the change.
//
// # Two oracles (CLI flag --diff vs --all-files)
//
//   - --all-files (default): intra-state oracle. Builds a lockset from the current
//     working tree and checks for data/test mismatches within the same snapshot.
//     Catches partial-translation regressions (one side updated, the other not).
//     Cannot catch PR #783-class regressions where BOTH data and test are translated
//     together (no intra-state mismatch exists in that case).
//
//   - --diff <git-rev>: temporal/baseline oracle. Parses git diff to find changed
//     string literals, rebuilds the baseline lockset for those files via git show,
//     and cross-references. Catches the PR #783 regression class. Requires the
//     working directory to be a git repository; falls back to --all-files with a
//     warning when git is unavailable.
//
// # Budget
//
// The --budget flag (default 30s) enforces a wall-clock ceiling. On deadline
// the validator exits with code 4 and the canonical message from AC-CIAUT-023.
//
// # Magic comment
//
// String literals annotated with // i18n:translatable (same line or immediately
// preceding line) are exempt from the lock check.
//
// # Out of scope (Wave 6)
//
// - Non-Go languages (follow-up SPEC)
// - IDE/LSP integration
// - Auto-fix suggestions
// - Hot-cache mode for ≤5s rerun
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// @MX:NOTE: [AUTO] Entry point for i18n-validator. Two-oracle design (--all-files + --diff).
// @MX:ANCHOR: exported data structures are consumed by both test files and diff.go.
// @MX:REASON: LockedLiteral, Lockset, Violation are referenced from main_test.go, lockset_test.go, diff_test.go.

// defaultBudget は全リポジトリスキャンのデフォルトウォールクロック制限時間です。
// defaultBudget is the default wall-clock budget for a full repository scan.
const defaultBudget = 30 * time.Second

// magicCommentToken は翻訳可能なリテラルを示す magic comment トークンです。
// magicCommentToken is the exact token that exempts a literal from lock checks.
const magicCommentToken = "i18n:translatable"

// allowedMethods はロック対象の testify assertion メソッドセットです。
// allowedMethods is the set of testify assertion methods that lock literals.
var allowedMethods = map[string]struct{}{
	"Equal":          {},
	"Equalf":         {},
	"EqualValues":    {},
	"Contains":       {},
	"NotContains":    {},
	"ContainsString": {},
	"Eq":             {},
}

// defaultCallers は assertion の呼び出し元として認識されるパッケージ名です。
// defaultCallers is the set of caller names recognised as testify callers.
var defaultCallers = map[string]struct{}{
	"assert":  {},
	"require": {},
	"s":       {},
	"suite":   {},
}

// corpusExclusions はスキャン対象から除外するディレクトリパスのリストです。
// corpusExclusions is the list of directory paths excluded from corpus scanning.
var corpusExclusions = []string{
	"vendor/",
	"node_modules/",
	".git/",
	"testdata/",
	".moai/",
}

// LockedLiteral はテストファイルが依存する string リテラルを記録します。
// LockedLiteral records a string literal that a test file depends on.
// @MX:ANCHOR: [AUTO] Core data structure referenced by all layers.
// @MX:REASON: fan_in >= 3 (main.go, lockset.go, diff.go + all test files).
type LockedLiteral struct {
	// File はリテラル宣言サイトのファイルパスです。
	File string
	// Line はリテラル宣言サイトの行番号です。
	Line int
	// Column はリテラル宣言サイトの列番号です。
	Column int
	// Text はアンクォートされたリテラル値です。
	Text string
	// AssertionRef はこのリテラルをロックした assertion の参照です。
	AssertionRef AssertionRef
	// Translatable は i18n:translatable magic comment が存在する場合 true です。
	Translatable bool
	// SourceTest はリテラルを参照するテスト関数名です。
	SourceTest string
	// AssertedText はテストが assert している期待値です (--all-files mismatch 検出用)。
	AssertedText string
}

// AssertionRef はテスト assertion のコールサイトを識別します。
// AssertionRef identifies the test call site that established the lock.
type AssertionRef struct {
	// File はテストファイルのパスです。
	File string
	// Line は assertion コールの行番号です。
	Line int
	// TestName は呼び出し元の関数名またはメソッド名です。
	TestName string
	// Method は testify メソッド名です。
	Method string
}

// Violation は diff comparator が発行する違反情報です。
// Violation is what the diff comparator emits when a locked literal changes.
// @MX:ANCHOR: [AUTO] Used by both --all-files and --diff oracles.
// @MX:REASON: fan_in >= 3 (main.go oracle, diff.go oracle, test files).
type Violation struct {
	// File は変更されたリテラルのファイルパスです。
	File string
	// Line は変更されたリテラルの行番号です。
	Line int
	// BaselineText は変更前のリテラル値です。
	BaselineText string
	// HeadText は変更後のリテラル値です。
	HeadText string
	// LockedBy はリテラルをロックした assertion の参照です。
	LockedBy AssertionRef
	// Reason は違反の理由説明です。
	Reason string
}

// BudgetResult はバジェット付きスキャンの結果を表します。
// BudgetResult represents the result of a budget-constrained scan run.
type BudgetResult struct {
	// Violations は発見された違反のリストです。
	Violations []Violation
	// ExitCode はプロセス終了コードです。0=正常, 1=違反, 4=タイムアウト
	ExitCode int
	// Stderr はエラー出力の内容です。
	Stderr string
}

// parseGoFile は指定パスの Go ファイルを AST に解析します。
// parseGoFile parses a Go file at the given path into an AST.
func parseGoFile(t interface {
	Helper()
	Fatal(...any)
}, path string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	return fset, f
}

// parseGoFileProduction は本番コードで使用する Go ファイル解析関数です。
// parseGoFileProduction parses a Go file for production use.
func parseGoFileProduction(path string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return fset, f, nil
}

// extractLockedLiterals は Go AST ファイルから testify assertion 内のリテラルを抽出します。
// extractLockedLiterals extracts string literals from testify assertions in a Go AST file.
func extractLockedLiterals(fset *token.FileSet, f *ast.File, filePath string) []LockedLiteral {
	var results []LockedLiteral
	comments := f.Comments

	// currentFunc は現在のスコープの関数名を追跡します。
	var currentFunc string

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				if node.Recv != nil && len(node.Recv.List) > 0 {
					// メソッドレシーバーの処理
					recvType := ""
					if star, ok := node.Recv.List[0].Type.(*ast.StarExpr); ok {
						if ident, ok := star.X.(*ast.Ident); ok {
							recvType = "(*" + ident.Name + ")"
						}
					} else if ident, ok := node.Recv.List[0].Type.(*ast.Ident); ok {
						recvType = ident.Name
					}
					currentFunc = recvType + "." + node.Name.Name
				} else {
					currentFunc = node.Name.Name
				}
			}

		case *ast.CallExpr:
			sel, ok := node.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			// メソッド名が allowedMethods にあるか確認
			if _, methodOK := allowedMethods[sel.Sel.Name]; !methodOK {
				return true
			}
			// 呼び出し元が allowedCallers にあるか確認
			callerIdent, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			callerName := callerIdent.Name
			if _, callerOK := defaultCallers[callerName]; !callerOK {
				return true
			}

			// 引数内の string リテラルを抽出
			// skip the first argument if it's a *testing.T parameter (i.e. an Ident named "t", "T", etc.)
			callPos := fset.Position(node.Pos())
			args := node.Args
			// For free-function callers (assert/require), skip the first *testing.T arg
			if callerName == "assert" || callerName == "require" {
				if len(args) > 0 {
					if ident, ok := args[0].(*ast.Ident); ok && (ident.Name == "t" || ident.Name == "T") {
						args = args[1:]
					}
				}
			}
			// isEqualMethod は Equal 系メソッド (Contains 以外) かどうかを返します。
			isEqualMethod := sel.Sel.Name != "Contains" && sel.Sel.Name != "NotContains" && sel.Sel.Name != "ContainsString"

			for idx, arg := range args {
				// For Equal-type methods:
				//   idx=0: "expected" literal — capture BasicLit only
				//   idx=1: "actual" value — capture Ident/IndexExpr only (for identifier resolution)
				//   idx>=2: skip
				// For Contains-type methods: capture all args (haystack + needle)
				if isEqualMethod && idx >= 2 {
					break
				}
				switch a := arg.(type) {
				case *ast.BasicLit:
					if a.Kind == token.STRING {
						// For Equal-type: only lock the expected value at idx=0.
						// idx=1 is the "actual" — if it's an inline literal, no lock needed
						// (the Ident/IndexExpr branch below handles identifier resolution at idx=1).
						if isEqualMethod && idx != 0 {
							break
						}
						litPos := fset.Position(a.Pos())
						text := unquoteString(a.Value)
						translatable := isTranslatable(fset, comments, a.Pos())
						ref := AssertionRef{
							File:     filePath,
							Line:     callPos.Line,
							TestName: currentFunc,
							Method:   sel.Sel.Name,
						}
						results = append(results, LockedLiteral{
							File:         filePath,
							Line:         litPos.Line,
							Column:       litPos.Column,
							Text:         text,
							AssertionRef: ref,
							Translatable: translatable,
							SourceTest:   currentFunc,
						})
					}
				case *ast.Ident:
					// 識別子参照 — Layer 2 で解決
					litPos := fset.Position(a.Pos())
					ref := AssertionRef{
						File:     filePath,
						Line:     callPos.Line,
						TestName: currentFunc,
						Method:   sel.Sel.Name,
					}
					results = append(results, LockedLiteral{
						File:         filePath,
						Line:         litPos.Line,
						Column:       litPos.Column,
						Text:         "", // Layer 2 で解決される
						AssertionRef: ref,
						SourceTest:   currentFunc,
					})
					// 識別子名を AssertionRef.TestName に一時格納 ("identName|funcName" 形式)
					results[len(results)-1].AssertionRef.TestName = a.Name + "|" + currentFunc
				case *ast.IndexExpr:
					// map[key] 参照 — e.g. mockData["title"]
					// 変数名とキーを記録して Layer 2 で解決
					if mapIdent, ok := a.X.(*ast.Ident); ok {
						if keyLit, ok := a.Index.(*ast.BasicLit); ok && keyLit.Kind == token.STRING {
							litPos := fset.Position(a.Pos())
							keyStr := unquoteString(keyLit.Value)
							ref := AssertionRef{
								File:     filePath,
								Line:     callPos.Line,
								TestName: currentFunc,
								Method:   sel.Sel.Name,
							}
							results = append(results, LockedLiteral{
								File:         filePath,
								Line:         litPos.Line,
								Column:       litPos.Column,
								Text:         "", // Layer 2 で解決される
								AssertionRef: ref,
								SourceTest:   currentFunc,
							})
							// "mapName[key]|funcName" 形式で一時格納
							results[len(results)-1].AssertionRef.TestName = mapIdent.Name + "[" + keyStr + "]|" + currentFunc
						}
					}
				}
			}
		}
		return true
	})

	return results
}

// isTranslatable は AST ノードの位置に対して i18n:translatable コメントが存在するか確認します。
// isTranslatable checks whether an i18n:translatable magic comment applies to a literal.
func isTranslatable(fset *token.FileSet, comments []*ast.CommentGroup, pos token.Pos) bool {
	litLine := fset.Position(pos).Line
	for _, cg := range comments {
		for _, c := range cg.List {
			commentLine := fset.Position(c.Pos()).Line
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			if strings.Contains(text, magicCommentToken) {
				// 同一行または直前行のコメントのみ有効
				if commentLine == litLine || commentLine == litLine-1 {
					return true
				}
			}
		}
	}
	return false
}

// unquoteString は Go の引用文字列から引用符を除去します。
// unquoteString removes Go string quoting from a literal value.
func unquoteString(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		// 基本的なアンクォート: エスケープシーケンスは保持
		inner := s[1 : len(s)-1]
		return inner
	}
	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		return s[1 : len(s)-1]
	}
	return s
}

// runAllFilesOracle は --all-files モードを実行して違反リストを返します。
// runAllFilesOracle runs the intra-state oracle over the given corpus root.
func runAllFilesOracle(root string) []Violation {
	builder := newLocksetBuilder()
	ls := builder.build(root)
	return checkAllFiles(ls)
}

// checkAllFiles はロックセットの各エントリを検証します。
// checkAllFiles validates each lockset entry for intra-state mismatches.
// In --all-files mode, a violation occurs when a locked literal (the const/var value)
// differs from what the test asserts. This catches partial-translation regressions.
func checkAllFiles(ls *Lockset) []Violation {
	var violations []Violation

	for _, lit := range ls.Literals {
		if lit.Translatable {
			continue
		}
		// AssertedText はテストが assert している期待値です。
		// LockedLiteral.Text はシンボルテーブルから解決された宣言値です。
		// intra-state mismatch: テストが別のテキストを期待している場合
		if lit.AssertedText != "" && lit.Text != "" && lit.AssertedText != lit.Text {
			violations = append(violations, Violation{
				File:         lit.File,
				Line:         lit.Line,
				BaselineText: lit.Text,
				HeadText:     lit.AssertedText,
				LockedBy:     lit.AssertionRef,
				Reason:       "translation-locked literal modified",
			})
		}
	}

	return violations
}

// runWithBudget はバジェット制限付きでスキャンを実行します。
// runWithBudget runs the scan with a wall-clock budget constraint.
// @MX:WARN: [AUTO] Uses goroutine + channel for cancellation.
// @MX:REASON: Budget enforcement requires concurrent timeout; cancel channel prevents goroutine leak.
func runWithBudget(root string, budget time.Duration) BudgetResult {
	type scanResult struct {
		violations []Violation
		// stderr はゴルーチン内でのみ書き込まれ、channel 経由で返される。
		// stderr is written only inside the goroutine and returned via channel
		// to avoid a data race with the timeout path.
		stderr string
	}

	done := make(chan scanResult, 1)

	go func() {
		var sb strings.Builder
		builder := newLocksetBuilder()
		builder.progressWriter = &sb
		ls := builder.buildWithProgress(root)
		done <- scanResult{
			violations: checkAllFiles(ls),
			stderr:     sb.String(),
		}
	}()

	select {
	case result := <-done:
		if len(result.violations) > 0 {
			return BudgetResult{
				Violations: result.violations,
				ExitCode:   1,
				Stderr:     result.stderr,
			}
		}
		return BudgetResult{ExitCode: 0, Stderr: result.stderr}
	case <-time.After(budget):
		msg := "validator exceeded 30s budget, consider scoping to changed files only"
		return BudgetResult{
			ExitCode: 4,
			// stderr from the goroutine is not safe to read here — omit partial output.
			Stderr: msg,
		}
	}
}

// findRepoRoot は git リポジトリのルートディレクトリを検索します。
// findRepoRoot locates the root of the current git repository.
func findRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// isExcluded はパスが除外リストに含まれるか確認します。
// isExcluded checks whether a path matches any corpus exclusion pattern.
// Path is normalized to forward-slash separators so Windows backslash paths match
// the canonical "vendor/", "node_modules/", ".git/", "testdata/", ".moai/" patterns.
func isExcluded(path string) bool {
	normalized := filepath.ToSlash(path)
	for _, excl := range corpusExclusions {
		if strings.Contains(normalized, excl) {
			return true
		}
	}
	return false
}

// main は i18n-validator CLI のエントリポイントです。
func main() {
	diffFlag := flag.String("diff", "", "git revision to compare against (baseline oracle)")
	budgetFlag := flag.Duration("budget", defaultBudget, "wall-clock budget for full repo scan")
	_ = flag.String("callers", "assert,require,s,suite", "comma-separated list of testify caller names")
	flag.Parse()

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	var result BudgetResult
	if *diffFlag != "" {
		// --diff モード: temporal/baseline oracle
		result = BudgetResult(runDiffOracle(root, *diffFlag))
	} else {
		// --all-files モード: intra-state oracle
		result = runWithBudget(root, *budgetFlag)
	}

	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	for _, v := range result.Violations {
		// AC-CIAUT-016 canonical format
		fmt.Fprintf(os.Stderr, "string literal at %s:%d is referenced by %s:%d, translation requires test update\n",
			v.File, v.Line, v.LockedBy.TestName, v.LockedBy.Line)
	}

	os.Exit(result.ExitCode)
}
