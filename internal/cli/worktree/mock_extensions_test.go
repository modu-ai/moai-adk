package worktree

// mock_extensions_test.go adds interface methods to mockWorktreeManager
// that were introduced after the original mock definition.
// This file supplements subcommands_test.go without modifying it.

// Extended mock fields added as package-level vars to avoid changing the original mock struct.
var (
	mockSyncFunc           func(string, string, string) error
	mockDeleteBranchFunc   func(string) error
	mockIsBranchMergedFunc func(string, string) (bool, error)
)

func (m *mockWorktreeManager) Sync(wtPath, baseBranch, strategy string) error {
	if mockSyncFunc != nil {
		return mockSyncFunc(wtPath, baseBranch, strategy)
	}
	return nil
}

func (m *mockWorktreeManager) DeleteBranch(name string) error {
	if mockDeleteBranchFunc != nil {
		return mockDeleteBranchFunc(name)
	}
	return nil
}

func (m *mockWorktreeManager) IsBranchMerged(branch, base string) (bool, error) {
	if mockIsBranchMergedFunc != nil {
		return mockIsBranchMergedFunc(branch, base)
	}
	return false, nil
}
