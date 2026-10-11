package outbox

// send_claim.go — the per-item cross-process send-ownership claim
// (review-gate finding, P2): two concurrent flushes each loaded the same
// queue snapshot and EACH sent the item — searches=2, creates=2, a public
// duplicate filed twice from the user's account. From the duplicate lookup
// through the outcome record, one item belongs to exactly one flush: the
// claim is taken BEFORE the search and released only after the outcome is
// recorded. A second flush whose snapshot still carries the item sees a
// live claim and skips it; a flush that died holding its claim leaves an
// owner-labelled file that the next flush reclaims through the same
// verified-dead rule as every other section (a crash between the create
// call and the outcome record can still re-send — gh create is not
// idempotent — and the retry's duplicate lookup then finds the issue and
// turns the repeat into an occurrence comment; the same accepted edge as
// two concurrent first filers).

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// The claim fails fast: one short retry, which is what lets a
// verified-dead owner's claim be broken and taken on the second attempt. A
// LIVE claim fails within roughly a millisecond — the skipping flush must
// not wait behind the owning one, it must move to its next item.
const (
	claimSendRetries    = 1
	claimSendRetryDelay = time.Millisecond
)

// queueItemIDPattern is the only shape an item ID may take: the drain mints
// IDs as f<sequence digits> and nothing else reaches a path (card-review
// finding, P2): a tampered queue ID carrying a traversal used to resolve
// the claim path OUTSIDE the user-scoped store, where a lookalike file was
// treated as a stale lock and deleted.
var queueItemIDPattern = regexp.MustCompile(`^f[0-9]+$`)

// ItemSendClaimPath returns the per-item send-ownership claim file's path
// under the user-scoped store (diagnostics and tests). The ID must pass the
// f<digits> allowlist BEFORE any path is constructed — a traversal-shaped
// ID is an error, never a path.
func ItemSendClaimPath(itemID string) (string, error) {
	if !queueItemIDPattern.MatchString(itemID) {
		return "", fmt.Errorf("outbox: queue item id %q fails the f<digits> allowlist", itemID)
	}
	return StorePath("send-" + itemID + ".claim")
}

// ClaimItemSend takes the item's send-ownership claim. An error means
// another flush owns the item right now (a live claim) — the caller skips
// the item; a verified-dead owner's claim was broken and re-taken on the
// way to success.
func ClaimItemSend(ctx context.Context, itemID string) (func() error, error) {
	path, err := ItemSendClaimPath(itemID)
	if err != nil {
		return nil, err
	}
	release, err := atomicfile.ClaimSection(ctx, path, 0o600, claimSendRetries, claimSendRetryDelay)
	if err != nil {
		return nil, fmt.Errorf("outbox: claim item %s for send: %w", itemID, err)
	}
	return release, nil
}
