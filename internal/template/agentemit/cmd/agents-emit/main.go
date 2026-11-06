// Command agents-emit renders the agent definitions from the agent source
// directory (internal/template/agentemit/source). Run from the repository root:
//
//	go run ./internal/template/agentemit/cmd/agents-emit          # regenerate every copy
//	go run ./internal/template/agentemit/cmd/agents-emit -check   # compare only; exit 1 on drift
//
// `make agents-emit` and `make agents-emit-check` wrap the two forms.
package main

import (
	"fmt"
	"os"

	"github.com/modu-ai/moai-adk/internal/template/agentemit"
)

func main() {
	man, err := agentemit.LoadManifest()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(agentemit.RunCLI(".", agentemit.DefaultLayout(), man, os.Args[1:], os.Stdout, os.Stderr))
}
