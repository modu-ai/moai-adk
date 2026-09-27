package escalation

import (
	"slices"
	"testing"
)

// goExported lists every exported top-level declaration kind and skips
// unexported ones; methods are qualified by their receiver type.
func TestGoExportedDeclarations(t *testing.T) {
	src := []byte(`package p

type Box[T any] struct{}
type inner struct{}

const Max, min = 1, 2
var Default, hidden int

func Top() {}
func low() {}
func (b *Box[T]) Get() T { var z T; return z }
func (i inner) Hidden() {}
`)
	got, ok := goExported(src)
	if !ok {
		t.Fatal("parse failed")
	}
	want := []string{"Box", "Max", "Default", "Top", "Box.Get", "inner.Hidden"}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("goExported = %v, want %v", got, want)
	}
	if _, ok := goExported([]byte("not go")); ok {
		t.Error("unparseable source reported as parsed")
	}
}

// A language with no grammar (scaffolded) is reported unsupported, which the
// caller lists as not-observed.
func TestAstxDeclsUnsupportedLanguage(t *testing.T) {
	if _, ok := astxDecls("r", "script.R", []byte("f <- function() 1\n")); ok {
		t.Error("scaffolded language reported as supported")
	}
}
