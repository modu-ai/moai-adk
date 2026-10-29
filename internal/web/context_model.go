// context_model.go — the per-Issue context-ledger model on the monitor
// (Issue #1948). It reads the ledger lane 18's ctxmodel stores and lane 20's
// ctxledger.PublishModel posts, through the contextModel seam, and only
// reads. A field the model does not carry renders as a gap; nothing is
// derived or invented, and no link is built: the model holds a path and a
// pinned commit, not a URL.
package web

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
)

// ContextModelItemVM is one ledger record. Summary is not shown: the
// published comment carries it only after ctxmodel.Render's mask check.
type ContextModelItemVM struct {
	// Redacted is set when refining masked something on this item.
	Redacted   bool
	Section    string
	Title      string
	Owner      string
	Path       string
	Commit     string // 12-char short sha
	TargetHead string // 12-char short sha
	Hash       string // 12-char short sha256
	Revision   string
	Verified   string
	VerifiedAt string
	State      string // PREPARED..REMOTE_VERIFIED, or a gap
	Status     string // available | pending | restricted | stale | not-linkable
}

// ContextModelVM is one Issue's ledger model.
type ContextModelVM struct {
	Known bool
	Note  string
	Items []ContextModelItemVM
}

func (a *app) buildContextModel(ctx context.Context, issue int) ContextModelVM {
	if a.contextModel == nil {
		return ContextModelVM{Note: "not provided"}
	}
	raw, err := a.contextModel(ctx, issue)
	if err != nil {
		return ContextModelVM{Note: "model unreadable: " + refineText(err.Error())}
	}
	led, err := ctxmodel.Unmarshal(raw)
	if err != nil {
		return ContextModelVM{Note: "model malformed: " + refineText(err.Error())}
	}
	if led.Issue != issue {
		return ContextModelVM{Note: "model belongs to another Issue"}
	}
	vm := ContextModelVM{Known: true}
	for _, it := range led.Items {
		vm.Items = append(vm.Items, ContextModelItemVM{
			Redacted: anyMasked(string(it.Section), it.Title, it.Owner, it.Path, it.Revision, it.Verified, it.VerifiedAt, string(it.State)),
			Section:  orGap(string(it.Section)), Title: masked(it.Title), Owner: masked(it.Owner),
			Path: orGap(it.Path), Commit: shortOrGap(it.Commit, 12), TargetHead: shortOrGap(it.TargetHead, 12),
			Hash: shortOrGap(it.SHA256, 12), Revision: orGap(it.Revision), Verified: orGap(it.Verified),
			VerifiedAt: orGap(it.VerifiedAt), State: orGap(string(it.State)), Status: modelStatus(it),
		})
	}
	return vm
}

// modelStatus is the model's own rule: only Available() is linkable.
func modelStatus(it ctxmodel.Item) string {
	switch {
	case it.Pending:
		return "pending"
	case it.Restricted:
		return "restricted"
	case it.Stale:
		return "stale"
	case it.Available():
		return "available"
	}
	return "not-linkable"
}

func masked(s string) string { return orGap(s) }

func shortOrGap(s string, n int) string {
	if s == "" {
		return prGap
	}
	return s[:min(n, len(s))]
}
