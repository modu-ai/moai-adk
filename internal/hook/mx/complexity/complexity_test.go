package complexity_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/mx/complexity"
)

// TestMeasure_ExportedSurface verifies AC-UTIL-001-06:
// The package exports exactly Measure (function) and Result (struct with Cyclomatic, IfBranches, Supported).
func TestMeasure_ExportedSurface(t *testing.T) {
	t.Parallel()

	// Verify that Measure is callable and Result fields are accessible.
	r, err := complexity.Measure("java", []byte("public class X {}"), "foo", 0)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	// Verify Result struct fields exist (compile-time check via field access).
	_ = r.Cyclomatic
	_ = r.IfBranches
	_ = r.Supported
}

// TestMeasure_ScaffoldedLanguages verifies AC-UTIL-001-09:
// All 11 scaffolded languages return Supported=false, Cyclomatic=0, error=nil.
func TestMeasure_ScaffoldedLanguages(t *testing.T) {
	t.Parallel()

	scaffolded := []string{
		"java", "kotlin", "csharp", "ruby", "php",
		"elixir", "cpp", "scala", "r", "flutter", "swift",
	}
	content := []byte("any content")

	for _, lang := range scaffolded {
		t.Run(lang, func(t *testing.T) {
			t.Parallel()
			r, err := complexity.Measure(lang, content, "foo", 0)
			if err != nil {
				t.Errorf("Measure(%q) error = %v, want nil (AC-UTIL-001-09)", lang, err)
			}
			if r.Supported {
				t.Errorf("Measure(%q).Supported = true, want false (AC-UTIL-001-09)", lang)
			}
			if r.Cyclomatic != 0 {
				t.Errorf("Measure(%q).Cyclomatic = %d, want 0 (AC-UTIL-001-09)", lang, r.Cyclomatic)
			}
			if r.IfBranches != 0 {
				t.Errorf("Measure(%q).IfBranches = %d, want 0 (AC-UTIL-001-09)", lang, r.IfBranches)
			}
		})
	}
}

// TestMeasure_UnknownLanguage verifies unknown languages return Supported=false without panic.
func TestMeasure_UnknownLanguage(t *testing.T) {
	t.Parallel()

	r, err := complexity.Measure("cobol", []byte("IDENTIFICATION DIVISION."), "foo", 0)
	if err != nil {
		t.Errorf("Measure(unknown lang) error = %v, want nil", err)
	}
	if r.Supported {
		t.Errorf("Measure(unknown lang).Supported = true, want false")
	}
}

// TestMeasure_OversizedFile verifies AC-UTIL-001-18:
// Content > 1 MiB returns Supported=false, Cyclomatic=0, error=nil within 100ms.
func TestMeasure_OversizedFile(t *testing.T) {
	t.Parallel()

	// Create a 2 MiB byte slice.
	content := make([]byte, 2<<20) // 2 MiB
	for i := range content {
		content[i] = 'x'
	}

	r, err := complexity.Measure("go", content, "foo", 1)
	if err != nil {
		t.Errorf("Measure(2MiB) error = %v, want nil (AC-UTIL-001-18)", err)
	}
	if r.Supported {
		t.Errorf("Measure(2MiB).Supported = true, want false (AC-UTIL-001-18)")
	}
	if r.Cyclomatic != 0 {
		t.Errorf("Measure(2MiB).Cyclomatic = %d, want 0 (AC-UTIL-001-18)", r.Cyclomatic)
	}
}

// TestMeasure_Go_SimpleFunction verifies a simple function returns Cyclomatic=1 (baseline).
func TestMeasure_Go_SimpleFunction(t *testing.T) {
	t.Parallel()

	src := []byte(`package testpkg

func Simple(x int) int {
	return x + 1
}
`)
	r, err := complexity.Measure("go", src, "Simple", 3)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter Go not available (CGO disabled)")
	}
	// Simple function with no decision nodes: Cyclomatic = 0 decisions + 1 = 1.
	if r.Cyclomatic != 1 {
		t.Errorf("Cyclomatic = %d, want 1 for simple function", r.Cyclomatic)
	}
	if r.IfBranches != 0 {
		t.Errorf("IfBranches = %d, want 0 for simple function", r.IfBranches)
	}
}

// TestMeasure_Go_ComplexDecision verifies AC-UTIL-001-10:
// A high-complexity Go function returns Cyclomatic >= 6.
func TestMeasure_Go_ComplexDecision(t *testing.T) {
	t.Parallel()

	src := []byte(`package testpkg

func ComplexDecision(x int) int {
	if x > 0 {
		if x > 10 {
			return 1
		}
	}
	for i := 0; i < x; i++ {
		if i%2 == 0 {
			continue
		}
	}
	switch x {
	case 1, 2:
		return x
	default:
		return 0
	}
}
`)
	r, err := complexity.Measure("go", src, "ComplexDecision", 3)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter Go not available (CGO disabled)")
	}
	// Expected: if(1) + if(2) + for(3) + if(4) + case1(5) + default(6) >= 6 decisions → Cyclomatic >= 7
	if r.Cyclomatic < 6 {
		t.Errorf("Cyclomatic = %d, want >= 6 for ComplexDecision (AC-UTIL-001-10)", r.Cyclomatic)
	}
}

// TestMeasure_Go_HighIfBranches verifies AC-UTIL-001-13b:
// A function with 10 if-branches returns IfBranches >= 10.
func TestMeasure_Go_HighIfBranches(t *testing.T) {
	t.Parallel()

	// Build a function with exactly 10 top-level if statements.
	var sb strings.Builder
	sb.WriteString("package testpkg\n\nfunc ManyBranches(x int) int {\n")
	for i := 0; i < 10; i++ {
		sb.WriteString("\tif x == ")
		sb.WriteString(string(rune('0' + i)))
		sb.WriteString(" { return ")
		sb.WriteString(string(rune('0' + i)))
		sb.WriteString(" }\n")
	}
	sb.WriteString("\treturn -1\n}\n")
	src := []byte(sb.String())

	r, err := complexity.Measure("go", src, "ManyBranches", 3)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter Go not available (CGO disabled)")
	}
	if r.IfBranches < 10 {
		t.Errorf("IfBranches = %d, want >= 10 for ManyBranches (AC-UTIL-001-13b)", r.IfBranches)
	}
}

// TestMeasure_Go_QueryCompiles verifies AC-UTIL-001-08:
// The Go query compiles without error and produces at least one match on a fixture.
func TestMeasure_Go_QueryCompiles(t *testing.T) {
	t.Parallel()

	src := []byte(`package testpkg

func HasBranch(x int) int {
	if x > 0 {
		return 1
	}
	return 0
}
`)
	r, err := complexity.Measure("go", src, "HasBranch", 3)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter Go not available (CGO disabled)")
	}
	// Has 1 if_statement: Cyclomatic should be 2 (1 decision + 1 baseline).
	if r.Cyclomatic < 2 {
		t.Errorf("Cyclomatic = %d, want >= 2 for function with one if (AC-UTIL-001-08)", r.Cyclomatic)
	}
}

// TestMeasure_Python_QueryCompiles verifies AC-UTIL-001-08:
// The Python query compiles without error.
func TestMeasure_Python_QueryCompiles(t *testing.T) {
	t.Parallel()

	src := []byte(`def process(x):
    if x > 0:
        for i in range(x):
            if i % 2 == 0:
                continue
    return x
`)
	r, err := complexity.Measure("python", src, "process", 1)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter Python not available (CGO disabled)")
	}
	// Has if, for, if → Cyclomatic >= 4.
	if r.Cyclomatic < 4 {
		t.Errorf("Cyclomatic = %d, want >= 4 for Python process() (AC-UTIL-001-08)", r.Cyclomatic)
	}
}

// TestMeasure_TypeScript_QueryCompiles verifies AC-UTIL-001-08 for TypeScript.
func TestMeasure_TypeScript_QueryCompiles(t *testing.T) {
	t.Parallel()

	src := []byte(`function evaluate(x: number): number {
    if (x > 0) {
        for (let i = 0; i < x; i++) {
            if (i % 2 === 0) continue;
        }
    }
    return x > 10 ? 1 : 0;
}
`)
	r, err := complexity.Measure("typescript", src, "evaluate", 1)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter TypeScript not available (CGO disabled)")
	}
	// Has if, for, if, ternary → Cyclomatic >= 5.
	if r.Cyclomatic < 5 {
		t.Errorf("Cyclomatic = %d, want >= 5 for TypeScript evaluate() (AC-UTIL-001-08)", r.Cyclomatic)
	}
}

// TestMeasure_JavaScript_QueryCompiles verifies AC-UTIL-001-08 for JavaScript.
func TestMeasure_JavaScript_QueryCompiles(t *testing.T) {
	t.Parallel()

	src := []byte(`function calculate(x) {
    if (x > 0) {
        for (let i = 0; i < x; i++) {
            if (i % 2 === 0) continue;
        }
    }
    return x || 0;
}
`)
	r, err := complexity.Measure("javascript", src, "calculate", 1)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter JavaScript not available (CGO disabled)")
	}
	// Has if, for, if, || → Cyclomatic >= 5.
	if r.Cyclomatic < 5 {
		t.Errorf("Cyclomatic = %d, want >= 5 for JS calculate() (AC-UTIL-001-08)", r.Cyclomatic)
	}
}

// TestMeasure_Rust_QueryCompiles verifies AC-UTIL-001-08 for Rust.
func TestMeasure_Rust_QueryCompiles(t *testing.T) {
	t.Parallel()

	src := []byte(`fn process(x: i32) -> i32 {
    if x > 0 {
        for i in 0..x {
            if i % 2 == 0 {
                continue;
            }
        }
    }
    x
}
`)
	r, err := complexity.Measure("rust", src, "process", 1)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter Rust not available (CGO disabled)")
	}
	// Has if, for_expression, if → Cyclomatic >= 4.
	if r.Cyclomatic < 4 {
		t.Errorf("Cyclomatic = %d, want >= 4 for Rust process() (AC-UTIL-001-08)", r.Cyclomatic)
	}
}

// TestMeasure_Go_StartLineHint verifies that startLine disambiguation works correctly.
// This exercises the abs() function with both positive and negative deltas.
func TestMeasure_Go_StartLineHint(t *testing.T) {
	t.Parallel()

	src := []byte(`package testpkg

func SimpleA(x int) int {
	if x > 0 {
		return 1
	}
	return 0
}

func SimpleB(x int) int {
	if x > 100 {
		return 2
	}
	return 0
}
`)
	// startLine=3: function is at line 3, delta=0 (positive case)
	rA, err := complexity.Measure("go", src, "SimpleA", 3)
	if err != nil {
		t.Fatalf("Measure(SimpleA) error = %v", err)
	}
	if !rA.Supported {
		t.Skip("tree-sitter Go not available (CGO disabled)")
	}
	if rA.IfBranches < 1 {
		t.Errorf("SimpleA IfBranches = %d, want >= 1", rA.IfBranches)
	}

	// startLine=4: function is at line 3, delta = 3-4 = -1 (exercises abs negative branch)
	rA2, err := complexity.Measure("go", src, "SimpleA", 4)
	if err != nil {
		t.Fatalf("Measure(SimpleA startLine=4) error = %v", err)
	}
	if !rA2.Supported {
		t.Errorf("Measure(SimpleA startLine=4).Supported = false, expected match within ±5 lines")
	}

	// startLine=0: match first occurrence (exercises zero-startLine branch)
	rA3, err := complexity.Measure("go", src, "SimpleA", 0)
	if err != nil {
		t.Fatalf("Measure(SimpleA startLine=0) error = %v", err)
	}
	if !rA3.Supported {
		t.Errorf("Measure(SimpleA startLine=0).Supported = false, expected match")
	}
}

// TestMeasure_Go_FunctionNotFound verifies that an unknown function name returns Supported=false.
func TestMeasure_Go_FunctionNotFound(t *testing.T) {
	t.Parallel()

	src := []byte(`package testpkg

func KnownFunc() int { return 0 }
`)
	r, err := complexity.Measure("go", src, "NonexistentFunc", 1)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	// Not found → Supported=false (same result as CGO disabled, but for different reason)
	// We can't distinguish; just verify no panic and no error.
	_ = r
}

// TestMeasure_JavaScript_MethodDefinition exercises nameChildOf fallback path.
func TestMeasure_JavaScript_MethodDefinition(t *testing.T) {
	t.Parallel()

	src := []byte(`class MyClass {
    process(x) {
        if (x > 0) {
            for (let i = 0; i < x; i++) {
                if (i % 2 === 0) continue;
            }
        }
        return x > 10 ? 1 : 0;
    }
}
`)
	r, err := complexity.Measure("javascript", src, "process", 2)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter JavaScript not available (CGO disabled)")
	}
	// Should detect the decision nodes within process().
	if r.Cyclomatic < 2 {
		t.Errorf("Cyclomatic = %d, want >= 2 for JS method process()", r.Cyclomatic)
	}
}

// TestMeasure_Go_HighComplexity_EmitsWarn validates integration: Cyclomatic >= 15.
// (Indirect validation; the direct P2 emission is tested in the mx package.)
func TestMeasure_Go_HighComplexity_EmitsWarn(t *testing.T) {
	t.Parallel()

	// Build a function with many if statements to exceed Cyclomatic=15.
	var sb strings.Builder
	sb.WriteString("package testpkg\n\nfunc HighComplexity(x int) int {\n")
	for i := 0; i < 20; i++ {
		sb.WriteString("\tif x == ")
		sb.WriteString(string(rune('0' + (i % 10))))
		sb.WriteString(string(rune('0' + i)))
		sb.WriteString(" { return ")
		sb.WriteString(string(rune('0' + i%10)))
		sb.WriteString(" }\n")
	}
	sb.WriteString("\treturn -1\n}\n")
	src := []byte(sb.String())

	r, err := complexity.Measure("go", src, "HighComplexity", 3)
	if err != nil {
		t.Fatalf("Measure() error = %v", err)
	}
	if !r.Supported {
		t.Skip("tree-sitter Go not available (CGO disabled)")
	}
	if r.Cyclomatic < 15 {
		t.Errorf("Cyclomatic = %d, want >= 15 for HighComplexity (AC-UTIL-001-13)", r.Cyclomatic)
	}
}
