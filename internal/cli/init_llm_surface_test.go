package cli

import "testing"

func TestInitLLMFlagReplacesAgentFlag(t *testing.T) {
	if initCmd.Flags().Lookup("llm") == nil {
		t.Fatal("--llm flag is not registered on init")
	}
	if initCmd.Flags().Lookup("agent") != nil {
		t.Fatal("legacy --agent flag is still registered on init")
	}
}

func TestUpdateAddCodexFlagRemoved(t *testing.T) {
	if updateCmd.Flags().Lookup("add-codex") != nil {
		t.Fatal("obsolete --add-codex flag is still registered on update")
	}
}
