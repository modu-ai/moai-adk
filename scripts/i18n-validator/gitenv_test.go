package main

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

func TestMain(m *testing.M) {
	if err := gitenv.ScrubProcess(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
