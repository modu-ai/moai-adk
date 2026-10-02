package homestate

// card_pr_states.go — the github-flow card delivery states
// (SPEC-GITHUB-FLOW-DEFAULT-001 M2-B, design D-4 option S-a).
//
// RED skeleton: the names only. The transition rows, guards and consumers land
// in the implementation commit.

const (
	// CardPROpen is a card whose branch is pushed and whose pull request is
	// open against the integration target with auto-merge requested.
	CardPROpen = "pr-open"
	// CardMergedPR is a card whose pull request was observed merged.
	CardMergedPR = "merged-pr"
)
