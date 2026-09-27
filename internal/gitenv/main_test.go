package gitenv_test

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// TestMain drops the variables that name a repository before any test runs,
// so this package's own git fixtures act on their temp directories even when
// the test binary inherits a caller's git context. It lives in the external
// test package because the internal one cannot import itself; both share one
// test binary, so this TestMain covers every test here.
func TestMain(m *testing.M) {
	if err := gitenv.ScrubProcess(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
