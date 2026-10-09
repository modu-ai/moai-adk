package publish

// lookup.go — the duplicate lookup (design.md section 7 step 3): one
// fingerprint-shaped search, then an EXACT title-key match. The query
// carries only the fingerprint, which leaks nothing; near-titles are
// ignored (a different issue is created and the consumer merges by
// fingerprint family — the accepted approximation). A closed issue still
// counts: the search passes --state all and no state filter exists here.

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/feedback"
)

// findIssue searches the repo for an issue whose title is EXACTLY the
// queued item's title key, returning it with its occurrence-marker count.
// A nil issue means "no match": the caller takes the create path.
func findIssue(ctx context.Context, r Runner, repo string, item feedback.QueueItem) (*RemoteIssue, int, error) {
	if item.Fingerprint == "" {
		// A manual-flow item has no fingerprint and never belongs on this
		// path; treat it as unmatched rather than searching on a title
		// fragment.
		return nil, 0, nil
	}
	issues, err := r.SearchIssues(ctx, repo, item.Fingerprint)
	if err != nil {
		return nil, 0, err
	}
	for i := range issues {
		if issues[i].Title == item.Title {
			return &issues[i], OccurrenceCount(issues[i].Comments), nil
		}
	}
	return nil, 0, nil
}
