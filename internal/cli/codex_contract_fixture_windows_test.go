//go:build windows

package cli

// codex_contract_fixture_windows_test.go — windows twin of the unix fixture
// helpers. Named pipes and unix sockets are not creatable here, so both
// return errCodexFixtureUnsupported: the containment table SKIPS those
// cells and LISTS the skipped kinds (AC-CI-011 — a quiet skip reads as a
// fake pass). The windows floor (108 cells) runs on the mode-injection,
// directory, and `..`-escape axes.

func makeCodexFIFOFixture(path string) error {
	return errCodexFixtureUnsupported
}

func makeCodexSocketFixtureAt(path string) (func(), error) {
	return nil, errCodexFixtureUnsupported
}
