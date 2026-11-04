package cli

// factory_retire_guard_text.go — the operator-facing catalogue of the
// `moai factory runs --retire` guard (SPEC-FACTORY-RETIRE-GUARD-001, card
// t1852). Prose is Korean; identifiers, states and commands stay verbatim.

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	retireRefusalHeaderFmt = "run %s 은퇴를 거절했습니다. 진행 중 카드 %d개가 남아 있습니다."
	retireOwnerLineFmt     = "리더 상태: %s"
	retireCardsHeader      = "진행 중 카드 (카드 id, 상태, 소유자):"
	retireCardLineFmt      = "  %s  %s  %s"
	retireExitsHeader      = "은퇴 대신 쓸 수 있는 방법 (run이 활성인 동안):"
	retireExitJoinFmt      = "  레인으로 합류: moai cc -l --factory-run %s"
	retireExitAdoptFmt     = "  리더로 이어받기: moai cc -f --factory-run %s (codex가 이끄는 run은 거절됩니다)"
	retireExitAbandonFmt   = "  카드 종료: moai factory decide %s --choice abandon --run %s"
	retireAbandonNote      = "카드 종료는 되돌릴 수 없고, 레인 세션에서는 거절되며, 임대가 만료된 카드는 한 번 더 호출해야 종료됩니다."
	retireStrandedHdrFmt   = "이미 은퇴한 run %s에 진행 중 카드 %d개가 남아 있습니다."
	retireStrandedNote     = "--factory-run 플래그는 활성 run만 고르므로 이 run에는 합류하거나 이어받을 수 없습니다 (NO_ACTIVE_FACTORY)."
)

func writeCardLines(w io.Writer, cards []homestate.InProgressCard) {
	_, _ = fmt.Fprintln(w, retireCardsHeader)
	for _, c := range cards {
		owner := c.OwnerLabel
		if owner == "" {
			owner = "-"
		}
		_, _ = fmt.Fprintf(w, retireCardLineFmt+"\n", c.CardID, c.State, owner)
	}
}

func writeAbandonExits(w io.Writer, runID string, cards []homestate.InProgressCard) {
	for _, c := range cards {
		_, _ = fmt.Fprintf(w, retireExitAbandonFmt+"\n", c.CardID, runID)
	}
}

// writeRetireRefusal renders the refusal of an active dead-owner run that
// still holds non-terminal cards: the cards, the exits, and the abandon caveats.
func writeRetireRefusal(w io.Writer, runID string, class homestate.OwnerClassification, cards []homestate.InProgressCard) {
	_, _ = fmt.Fprintf(w, retireRefusalHeaderFmt+"\n", runID, len(cards))
	_, _ = fmt.Fprintf(w, retireOwnerLineFmt+"\n", class)
	writeCardLines(w, cards)
	_, _ = fmt.Fprintln(w, retireExitsHeader)
	_, _ = fmt.Fprintf(w, retireExitJoinFmt+"\n", runID)
	_, _ = fmt.Fprintf(w, retireExitAdoptFmt+"\n", runID)
	writeAbandonExits(w, runID, cards)
	_, _ = fmt.Fprintln(w, retireAbandonNote)
}

// writeStrandedCards renders the informational block for a retired run that
// still holds non-terminal cards. It prints nothing for an empty list.
func writeStrandedCards(w io.Writer, runID string, cards []homestate.InProgressCard) {
	if len(cards) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, retireStrandedHdrFmt+"\n", runID, len(cards))
	writeCardLines(w, cards)
	writeAbandonExits(w, runID, cards)
	_, _ = fmt.Fprintln(w, retireStrandedNote)
	_, _ = fmt.Fprintln(w, retireAbandonNote)
}
