// Package symbol is the graph builder's code-extraction seam (REQ-GF-013):
// it consumes the astx extractor WITHOUT pulling any navigator-tier
// dependency into this package's transitive set (AC-GF-016, verified by
// go list -deps on this package). The mapping into graph edges happens one
// package up (internal/graph); this package stays domain-typed.
package symbol

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/mx"
	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// CallEdge is one extracted call relationship: the caller (file + enclosing
// function, joined by line containment) and the callee name as written.
type CallEdge struct {
	// File is the repo-relative source file.
	File string
	// Caller is the enclosing function/method name ("" at top level).
	Caller string
	// Callee is the callee identifier (last segment for selectors).
	Callee string
	// Line is the 1-indexed call line.
	Line int
	// Grade is the resolution grade used to derive this edge.
	Grade string
}

// ImportEdge is one extracted import: file → module.
type ImportEdge struct {
	// File is the repo-relative source file.
	File string
	// Module is the imported module/package, normalized to a repository-local
	// path when it belongs to this project (Module prefix stripped);
	// external imports keep their full path.
	Module string
	// Local reports whether Module is a repository-local package (the doc
	// import layer's domain — only local imports are comparable with it).
	Local bool
	// Line is the 1-indexed import line.
	Line int
	// Grade is the resolution grade used to derive this edge.
	Grade string
}

// GradeMatrix returns the per-language call-resolution grade for every
// registered language (astx.SupportedLanguages universe), published as data.
func GradeMatrix() map[string]string {
	out := map[string]string{}
	for _, lang := range astx.SupportedLanguages() {
		out[lang] = astx.GradeFor(lang)
	}
	return out
}

// ValidateGradeMatrix reports gradeless cells as defect verdicts naming the
// language (REQ-GF-016 / AC-GF-019): an omitted cell may not pass silently.
func ValidateGradeMatrix(matrix map[string]string) []string {
	var defects []string
	for _, lang := range astx.SupportedLanguages() {
		grade, ok := matrix[lang]
		if !ok || grade == "" {
			defects = append(defects, fmt.Sprintf("grade-matrix defect: language %s carries no grade (empty cell)", lang))
			continue
		}
		switch grade {
		case astx.GradeFull, astx.GradeNameBased, astx.GradeNone:
		default:
			defects = append(defects, fmt.Sprintf("grade-matrix defect: language %s carries invalid grade %q", lang, grade))
		}
	}
	sort.Strings(defects)
	return defects
}

// Extract walks the described source trees (the codemaps described roots —
// the same universe the freshness gate judges) and returns code-derived
// call/import edges plus the grade matrix. Languages without call captures
// (grade none) contribute nothing; per-file failures fail open, never fatal.
//
// Import modules are normalized to repository-local paths when the project's
// module path prefixes them (go.mod `module` line), so code-imports and the
// doc import layer speak the same package domain.
//
// Resolution is name-based: callee names are matched without scope; the
// matrix publishes exactly this.
//
// @MX:NOTE: [AUTO] symbol.Extract — REQ-GF-013 seam: astx consumed outside the navigator with no navigator-tier dep (tiers stays a graph-level concern)
// @MX:SPEC:SPEC-V3R6-GRAPH-FRESHNESS-001
func Extract(projectRoot string) (calls []CallEdge, imports []ImportEdge, matrix map[string]string, err error) {
	matrix = GradeMatrix()
	modulePrefix := modulePath(projectRoot)

	for _, root := range mx.DefaultDescribedRoots {
		absRoot := filepath.Join(projectRoot, filepath.FromSlash(root))
		walkErr := filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			if info.IsDir() {
				if info.Name() == "testdata" || strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			lang := astx.DetectLanguage(path)
			if lang == "" || astx.GradeFor(lang) == astx.GradeNone {
				return nil
			}
			// Regular-file guard (CR round-2 3855001937): extraction reads
			// content; a FIFO/socket under a walked root is skipped, not opened.
			if !info.Mode().IsRegular() {
				return nil
			}
			set, xErr := astx.ExtractCalls(lang, path)
			if xErr != nil || !set.Supported {
				return nil
			}
			rel, relErr := filepath.Rel(projectRoot, path)
			if relErr != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			grade := astx.GradeFor(lang)

			for _, call := range set.Calls {
				calls = append(calls, CallEdge{
					File:   rel,
					Caller: enclosingFunction(set.Functions, call.Line),
					Callee: call.Callee,
					Line:   call.Line,
					Grade:  grade,
				})
			}
			for _, imp := range set.Imports {
				local, isLocal := localizeModule(imp.Module, modulePrefix)
				imports = append(imports, ImportEdge{
					File:   rel,
					Module: local,
					Local:  isLocal,
					Line:   imp.Line,
					Grade:  grade,
				})
			}
			return nil
		})
		if walkErr != nil {
			return nil, nil, nil, fmt.Errorf("symbol: walk %s: %w", root, walkErr)
		}
	}

	sort.Slice(calls, func(i, j int) bool {
		if calls[i].File != calls[j].File {
			return calls[i].File < calls[j].File
		}
		return calls[i].Line < calls[j].Line
	})
	sort.Slice(imports, func(i, j int) bool {
		if imports[i].File != imports[j].File {
			return imports[i].File < imports[j].File
		}
		return imports[i].Line < imports[j].Line
	})
	return calls, imports, matrix, nil
}

// enclosingFunction returns the name of the innermost range containing line,
// or "" when the call sits at no captured declaration.
func enclosingFunction(ranges []astx.FuncRange, line int) string {
	best := ""
	bestStart := -1
	for _, r := range ranges {
		if line >= r.StartLine && line <= r.EndLine {
			if r.StartLine >= bestStart { // innermost = latest start
				bestStart = r.StartLine
				best = r.Name
			}
		}
	}
	return best
}

// modulePath reads the `module` directive from <root>/go.mod ("" when absent
// or unreadable). Not hardcoded — the module path is a property of the
// scanned project.
func modulePath(projectRoot string) string {
	data, err := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(after)
		}
	}
	return ""
}

// localizeModule strips the project's module prefix from an import path so
// internal dependencies appear as repository-local paths (the same domain
// the doc import layer speaks). External imports pass through unchanged;
// isLocal reports whether the strip happened.
func localizeModule(module, modulePrefix string) (local string, isLocal bool) {
	if modulePrefix == "" {
		return module, false
	}
	if l, ok := strings.CutPrefix(module, modulePrefix+"/"); ok {
		return l, true
	}
	return module, false
}
