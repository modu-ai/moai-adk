package adapt

import (
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/monitor"
)

// ContextItems turns an Issue's context ledger into the monitor's context items, read-only.
//
//   - Stage maps one to one from the ledger's link state; a restricted item shows as restricted.
//   - State: a stale item is stale (the ledger keeps it, so the count does not change and the stage and
//     hash it was judged at stay); otherwise prepared/masked evidence is inaccessible-local, a committed
//     one is inaccessible-unpushed, and a pushed or remote-verified one is current.
//   - Link is built only for a pushed or remote-verified item that is neither pending nor restricted, and
//     only from the item's full 40-hex commit SHA - never a branch name: linkBase + commit + "/" + path.
//     Everything else carries no link at all.
//   - RemoteVerified is true only for a remote-verified item; the collector turns that, together with the
//     stage, into availability. Pushing alone proves nothing.
func ContextItems(led *ctxmodel.Ledger, linkBase string) []monitor.ContextItem {
	if led == nil {
		return nil
	}
	out := make([]monitor.ContextItem, 0, len(led.Items))
	for _, it := range led.Items {
		ci := monitor.ContextItem{
			Section: string(it.Section), TargetHead: it.TargetHead, Hash: it.SHA256, Owner: it.Owner,
			Stage: stageOf(it), RemoteVerified: it.State == ctxmodel.StateRemoteVerified && !it.Stale,
		}
		switch {
		case it.Stale:
			ci.State = monitor.CtxStale
		case it.Pending || it.State == ctxmodel.StatePrepared || it.State == ctxmodel.StateMasked:
			ci.State = monitor.CtxInaccessibleLocal
		case it.State == ctxmodel.StateCommitted:
			ci.State = monitor.CtxInaccessibleUnpushed
		default:
			ci.State = monitor.CtxCurrent
		}
		if !it.Pending && !it.Restricted && (it.State == ctxmodel.StatePushed || it.State == ctxmodel.StateRemoteVerified) && fullSHA(it.Commit) {
			ci.Link = strings.TrimRight(linkBase, "/") + "/" + it.Commit + "/" + it.Path
		}
		out = append(out, ci)
	}
	return out
}

func stageOf(it ctxmodel.Item) monitor.EvidenceStage {
	if it.Restricted {
		return monitor.StageRestricted
	}
	switch it.State {
	case ctxmodel.StateMasked:
		return monitor.StageMasked
	case ctxmodel.StateCommitted:
		return monitor.StageCommitted
	case ctxmodel.StatePushed:
		return monitor.StagePushed
	case ctxmodel.StateRemoteVerified:
		return monitor.StageRemoteVerified
	}
	return monitor.StagePrepared
}

func fullSHA(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
