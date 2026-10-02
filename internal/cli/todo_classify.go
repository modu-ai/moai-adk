// todo_classify.go — SPEC-TODO-CLASSIFY-DISPATCH-001 M2: the add path's
// classification wiring.
//
// The seam (todoCardDecider) resolves every admitted card's judgment INSIDE
// the same locked write that appends the card (REQ-TCD-001, plan G1): there
// is no two-step add-then-classify window a concurrent `factory next` could
// observe. A decider that is unavailable or fails never blocks admission —
// the fail-safe default is promoted and exactly one stderr notice names the
// fallback (REQ-TCD-003; the serial default is the fail-safe direction, a
// wrongly-parallel default could run true-serial cards concurrently).
//
// The ONLY injection seam is --classification-file (plan D.3): the product
// computes no LLM call and knows no external tooling. An operator (or a local
// tool the operator runs) supplies a validated judgement file; the product
// records a decider-attributed judgment it did not compute.
package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// todoCardDecider is the classification seam the add paths resolve
// (REQ-TCD-012). The shipped value is the deterministic default; the
// --classification-file flag swaps in a static carrier for one invocation.
// Tests replace it to inject judgments and failures.
var todoCardDecider factory.CardDecider = factory.DefaultCardDecider{}

// todoClassificationFallbackNotice is the ONE stderr line the fallback
// prints (REQ-TCD-003). Exactly one line, so a supervising reader can count
// fallbacks by counting lines.
const todoClassificationFallbackNotice = "todo add: classification decider unavailable; admitted with the fail-safe default (priority=normal, blocked=false, mode=serial, decider=default)"

// todoDeciderFromClassificationFile resolves the add's decider from the
// --classification-file value ("<path>" or "-" for standard input), BEFORE
// the locked write:
//
//   - a file that cannot be read is a TRANSPORT unavailability, not a
//     refusal: the seam resolves to UnavailableCardDecider and the locked
//     write promotes the fail-safe default (REQ-TCD-003);
//   - content that is read but malformed or out-of-set is a USAGE error:
//     the add is refused with exit 2 and nothing is written (REQ-TCD-004).
func todoDeciderFromClassificationFile(path string) (factory.CardDecider, error) {
	var (
		data []byte
		err  error
	)
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return factory.UnavailableCardDecider{
			Cause: fmt.Errorf("read --classification-file %s: %w", path, err),
		}, nil
	}
	cls, parseErr := factory.ParseCardClassificationJSON(data)
	if parseErr != nil {
		return nil, &exitCodeError{
			code: 2,
			msg:  fmt.Sprintf("todo add: --classification-file %s refused: %v (nothing written)", path, parseErr),
		}
	}
	return factory.StaticCardDecider{Class: cls}, nil
}

// todoClassifyInLock resolves one card's classification inside the locked
// write. A decider error promotes the fail-safe default and prints exactly
// one stderr notice; a healthy decider's judgment is recorded verbatim. The
// classified-at stamp is taken here, at the write, for every branch.
func todoClassifyInLock(dec factory.CardDecider, text string, errOut io.Writer) factory.CardClassification {
	cls, err := dec.Classify(text)
	if err != nil {
		_, _ = fmt.Fprintln(errOut, todoClassificationFallbackNotice)
		cls = factory.DefaultCardClassification()
	}
	cls.ClassifiedAt = time.Now().UTC().Format(time.RFC3339)
	return cls
}

// todoApplyClassification attaches the resolved classification to the named
// card inside the caller's locked write.
func todoApplyClassification(rec *factory.BacklogRecord, cardID string, cls factory.CardClassification) {
	for i := range rec.Items {
		if rec.Items[i].ID == cardID {
			c := cls
			rec.Items[i].Classification = &c
			return
		}
	}
}
