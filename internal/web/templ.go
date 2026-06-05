package web

// Templ component codegen for the Console rendering layer (SPEC-WEB-CONSOLE-006).
//
// The *.templ sources in this package are compiled to *_templ.go by the templ
// codegen tool — a pure-Go tool (no Node / npm / Vite). `go generate ./...`
// regenerates them; the generated *_templ.go files are committed as source
// artifacts (a fresh `go build ./...` works without invoking the codegen tool),
// and CI runs a `templ generate` + `git diff --exit-code` drift guard.
//
//go:generate go run github.com/a-h/templ/cmd/templ generate
