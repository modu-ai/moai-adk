// Package main은 cross-file resolver와 lockset builder를 제공합니다.
// Package main provides the cross-file resolver (Layer 2) and lockset builder (Layer 3).
//
// # Cross-file Resolver (Layer 2)
//
// Builds a per-package symbol table by walking package-level const/var declarations
// and flat map[string]string composite literals (sufficient for the mockReleaseData
// pattern from PR #783). External-package identifiers are gracefully skipped.
//
// # Lockset Builder (Layer 3)
//
// Walks the corpus (excluding vendor/, node_modules/, .git/, testdata/, .moai/),
// uses Layer 1 to extract assertions, Layer 2 to resolve identifiers, and builds
// a Lockset keyed by "<file>:<line>" canonical form.
//
// # Scope limits (Wave 6)
//
// - Single Go module assumed; cross-module resolution deferred.
// - Nested map literals and struct literals deferred; flat map[string]string covered.
// - Function-local variable resolution deferred; package-level scope sufficient for AC-CIAUT-016.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"maps"
	"path/filepath"
	"strings"
	"time"
)

// @MX:NOTE: [AUTO] Layer 2 + Layer 3 implementation. Shared AST cache prevents double-parsing.

// symbolEntry はパッケージレベルシンボルの宣言情報です。
// symbolEntry records the declaration of a package-level symbol.
type symbolEntry struct {
	// File はシンボルが宣言されたファイルです。
	File string
	// Line は宣言の行番号です。
	Line int
	// Value はシンボルの文字列値です。
	Value string
	// Translatable は i18n:translatable コメントが付いているか否かです。
	Translatable bool
}

// locksetBuilder はロックセットの構築を担当します。
// locksetBuilder manages the construction of a Lockset over a corpus.
type locksetBuilder struct {
	// astCache は解析済み AST ファイルのキャッシュです。
	astCache map[string]*cachedAST
	// progressWriter はプログレス出力先です (nil = discard)。
	progressWriter io.Writer
}

// cachedAST は解析済み AST とファイルセットのペアです。
// cachedAST holds a parsed AST and its associated FileSet.
type cachedAST struct {
	fset    *token.FileSet
	astFile *ast.File
}

// Lockset はスキャンコーパスにわたるロック済みリテラルのセットです。
// Lockset is the union of locked literals across the scan corpus.
// @MX:ANCHOR: [AUTO] Core output structure consumed by both oracle modes.
// @MX:REASON: fan_in >= 3 (main.go, diff.go, lockset_test.go, main_test.go).
type Lockset struct {
	// Literals はキー "<file>:<line>" でインデックスされたロック済みリテラルマップです。
	Literals map[string]LockedLiteral
	// BuiltAt はロックセットの構築時刻です。
	BuiltAt time.Time
	// Corpus はスキャン対象のファイルパスリストです。
	Corpus []string
}

// Freeze はロックセットの不変ビューを返します。
// Freeze returns an immutable view of the lockset.
func (ls *Lockset) Freeze() *Lockset {
	frozen := &Lockset{
		Literals: make(map[string]LockedLiteral, len(ls.Literals)),
		BuiltAt:  ls.BuiltAt,
		Corpus:   append([]string(nil), ls.Corpus...),
	}
	maps.Copy(frozen.Literals, ls.Literals)
	return frozen
}

// newLocksetBuilder は新しい locksetBuilder を生成します。
// newLocksetBuilder creates a new locksetBuilder with an empty AST cache.
func newLocksetBuilder() *locksetBuilder {
	return &locksetBuilder{
		astCache: make(map[string]*cachedAST),
	}
}

// getAST は AST キャッシュから解析済み AST を返します。キャッシュミスの場合は解析します。
// getAST returns a parsed AST from cache, parsing on cache miss.
func (b *locksetBuilder) getAST(path string) (*token.FileSet, *ast.File, error) {
	if cached, ok := b.astCache[path]; ok {
		return cached.fset, cached.astFile, nil
	}
	fset, f, err := parseGoFileProduction(path)
	if err != nil {
		return nil, nil, err
	}
	b.astCache[path] = &cachedAST{fset: fset, astFile: f}
	return fset, f, nil
}

// build はコーパスルートからロックセットを構築します。
// build constructs a Lockset from the given corpus root directory.
func (b *locksetBuilder) build(root string) *Lockset {
	return b.buildWithProgress(root)
}

// buildWithProgress はプログレス出力付きでロックセットを構築します。
// buildWithProgress builds a Lockset with optional progress reporting.
func (b *locksetBuilder) buildWithProgress(root string) *Lockset {
	ls := &Lockset{
		Literals: make(map[string]LockedLiteral),
		BuiltAt:  time.Now(),
	}

	// シンボルテーブル: パッケージディレクトリ → 識別子名 → symbolEntry
	symTable := make(map[string]map[string]symbolEntry)

	// パス 1: 全ファイルをスキャンしてシンボルテーブルを構築
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if isExcluded(path) {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		if b.progressWriter != nil {
			_, _ = fmt.Fprintf(b.progressWriter, "[i18n-validator] scanning %s\n", rel)
		}

		fset, f, err := b.getAST(path)
		if err != nil {
			return nil
		}
		ls.Corpus = append(ls.Corpus, path)

		// シンボルテーブルへのエントリ追加
		pkgDir := filepath.Dir(path)
		if symTable[pkgDir] == nil {
			symTable[pkgDir] = make(map[string]symbolEntry)
		}
		buildSymbolTable(fset, f, path, symTable[pkgDir])
		return nil
	})

	// パス 2: テストファイルから assertion を抽出してロックセットを構築
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || isExcluded(path) {
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset, f, err := b.getAST(path)
		if err != nil {
			return nil
		}

		pkgDir := filepath.Dir(path)
		lits := extractLockedLiterals(fset, f, path)

		// 同一コールの literal/identifier ペアを追跡
		// (inline literal が assert.Equal の 期待値, identifier が変数参照の場合)
		callAssertions := make(map[int][]LockedLiteral) // key: call line
		for _, lit := range lits {
			callAssertions[lit.AssertionRef.Line] = append(callAssertions[lit.AssertionRef.Line], lit)
		}

		for _, lit := range lits {
			if lit.Text != "" {
				// インラインリテラル — テストが直接アサートしている期待値
				key := fmt.Sprintf("%s:%d", lit.File, lit.Line)
				if _, exists := ls.Literals[key]; !exists {
					ls.Literals[key] = lit
				}
			} else {
				// 識別子参照 — Layer 2 で解決
				// AssertionRef.TestName に "identName|funcName" 形式で格納
				parts := strings.SplitN(lit.AssertionRef.TestName, "|", 2)
				if len(parts) != 2 {
					continue
				}
				identName := parts[0]
				funcName := parts[1]

				// シンボルテーブルから解決
				syms, ok := symTable[pkgDir]
				if !ok {
					continue
				}
				sym, ok := syms[identName]
				if !ok {
					// 未解決の識別子 — graceful skip
					continue
				}

				// 同一コールの inline literal を探す (mismatch 検出のため)
				var assertedText string
				callLits := callAssertions[lit.AssertionRef.Line]
				for _, cl := range callLits {
					if cl.Text != "" {
						assertedText = cl.Text
						break
					}
				}

				key := fmt.Sprintf("%s:%d", sym.File, sym.Line)
				if _, exists := ls.Literals[key]; !exists {
					ref := AssertionRef{
						File:     lit.AssertionRef.File,
						Line:     lit.AssertionRef.Line,
						TestName: funcName,
						Method:   lit.AssertionRef.Method,
					}
					ls.Literals[key] = LockedLiteral{
						File:         sym.File,
						Line:         sym.Line,
						Text:         sym.Value,
						AssertionRef: ref,
						Translatable: sym.Translatable,
						SourceTest:   funcName,
						AssertedText: assertedText,
					}
				}
			}
		}
		return nil
	})

	return ls.Freeze()
}

// buildSymbolTable は AST ファイルからパッケージレベルシンボルテーブルを構築します。
// buildSymbolTable builds a package-level symbol table from an AST file.
func buildSymbolTable(fset *token.FileSet, f *ast.File, filePath string, syms map[string]symbolEntry) {
	comments := f.Comments

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		switch genDecl.Tok {
		case token.CONST, token.VAR:
			// const/var 宣言のスキャン
			for _, spec := range genDecl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range vs.Names {
					if i >= len(vs.Values) {
						continue
					}
					lit, ok := vs.Values[i].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						continue
					}
					pos := fset.Position(lit.Pos())
					translatable := isTranslatable(fset, comments, lit.Pos())
					syms[name.Name] = symbolEntry{
						File:         filePath,
						Line:         pos.Line,
						Value:        unquoteString(lit.Value),
						Translatable: translatable,
					}
				}
			}

		case token.TYPE:
			// 型宣言はスキップ
		}
	}

	// var のトップレベル宣言でマップリテラルをスキャン
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if i >= len(vs.Values) {
					continue
				}
				compLit, ok := vs.Values[i].(*ast.CompositeLit)
				if !ok {
					continue
				}
				// map[string]string リテラルの処理
				extractMapLiteralEntries(fset, f, filePath, name.Name, compLit, syms)
			}
		}
	}
}

// extractMapLiteralEntries は map リテラルのエントリをシンボルテーブルに登録します。
// extractMapLiteralEntries registers map literal entries into the symbol table.
func extractMapLiteralEntries(fset *token.FileSet, f *ast.File, filePath, varName string, compLit *ast.CompositeLit, syms map[string]symbolEntry) {
	comments := f.Comments
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		keyLit, ok := kv.Key.(*ast.BasicLit)
		if !ok || keyLit.Kind != token.STRING {
			continue
		}
		valLit, ok := kv.Value.(*ast.BasicLit)
		if !ok || valLit.Kind != token.STRING {
			continue
		}

		keyStr := unquoteString(keyLit.Value)
		valStr := unquoteString(valLit.Value)
		pos := fset.Position(valLit.Pos())
		translatable := isTranslatable(fset, comments, valLit.Pos())

		// "varName[key]" 形式でシンボルを登録
		symKey := varName + "[" + keyStr + "]"
		syms[symKey] = symbolEntry{
			File:         filePath,
			Line:         pos.Line,
			Value:        valStr,
			Translatable: translatable,
		}
	}
}

// buildLocksetForFile は単一ファイルのロックセットを構築します (--diff モードで使用)。
// buildLocksetForFile builds a lockset for a single file (used by --diff mode).
func buildLocksetForFile(content []byte, filePath string) (*Lockset, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ls := &Lockset{
		Literals: make(map[string]LockedLiteral),
		BuiltAt:  time.Now(),
		Corpus:   []string{filePath},
	}

	// シンボルテーブルを構築
	syms := make(map[string]symbolEntry)
	buildSymbolTable(fset, f, filePath, syms)

	// リテラルを抽出
	lits := extractLockedLiterals(fset, f, filePath)
	for _, lit := range lits {
		if lit.Text != "" {
			key := fmt.Sprintf("%s:%d", lit.File, lit.Line)
			if _, exists := ls.Literals[key]; !exists {
				ls.Literals[key] = lit
			}
		} else {
			// 識別子参照の解決
			parts := strings.SplitN(lit.AssertionRef.TestName, "|", 2)
			if len(parts) != 2 {
				continue
			}
			identName := parts[0]
			funcName := parts[1]

			sym, ok := syms[identName]
			if !ok {
				continue
			}

			key := fmt.Sprintf("%s:%d", sym.File, sym.Line)
			if _, exists := ls.Literals[key]; !exists {
				ref := AssertionRef{
					File:     lit.AssertionRef.File,
					Line:     lit.AssertionRef.Line,
					TestName: funcName,
					Method:   lit.AssertionRef.Method,
				}
				ls.Literals[key] = LockedLiteral{
					File:         sym.File,
					Line:         sym.Line,
					Text:         sym.Value,
					AssertionRef: ref,
					Translatable: sym.Translatable,
					SourceTest:   funcName,
				}
			}
		}
	}

	// マップリテラル値もロック対象として登録
	for key, sym := range syms {
		if strings.Contains(key, "[") {
			lsKey := fmt.Sprintf("%s:%d", sym.File, sym.Line)
			if _, exists := ls.Literals[lsKey]; !exists {
				ls.Literals[lsKey] = LockedLiteral{
					File:         sym.File,
					Line:         sym.Line,
					Text:         sym.Value,
					Translatable: sym.Translatable,
				}
			}
		}
	}

	return ls.Freeze(), nil
}
