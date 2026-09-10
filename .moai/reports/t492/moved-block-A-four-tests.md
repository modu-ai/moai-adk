#### The four tests

Applied in order. Every test is answerable by reading the sentence the ref appears in.

**Test 1 — Substitution.** Replace the ref token with the SHA it resolves to *right now*, then re-read the sentence **as a later reader will act on it — not as you read it at the moment of substitution.** The evaluation time is load-bearing: a Test 1 applied at the substitution instant returns ANCHOR for every live-state claim there is, because substituting today's tip into "the base you will start from" reads correctly today and is wrong for every reader after.

- Still says what it meant, *and still will when acted on later* → **ANCHOR** (an address at which a measurement was taken).
- Now says something different, narrower, or weaker — **including a meaning correct at the instant of substitution that decays afterwards** → **SUBJECT** (the claim is *about* the moving thing).

**Test 2 — Falsification source.** Conditional: it runs only when the claim currently reads false, and it returns an attribution rather than a class. Were the commits that flipped it authored by this work?

- Authored by this work → **true signal**. The claim is genuinely broken; fix the work, do not touch the ref. Pinning here hides a real defect, which is the worse of the two errors.
- Not authored by this work → **spurious red** from upstream drift; remediate per the branches below.

**Test 3 — Re-measurement expectation.** Re-run this claim next week with no work done in between. Is the same answer expected? Yes → ANCHOR. No, *and that variance is the point of the claim* → SUBJECT.

**Tie-break.** Tests 1 and 3 normally agree. Where they disagree, **run Test 4 — do not resolve to ANCHOR.** A disagreement is not noise to be settled by precedence; it is the signature of a live-state claim, in which Test 1 says "the value fits" while Test 3 says "the value must not be fixed". Treat a Test 1 / Test 3 disagreement as evidence *against* ANCHOR rather than for it.

**Test 4 — Read-time action.** Runs when Tests 1 and 3 both return SUBJECT, and whenever they disagree. Test 2 is deliberately outside this gate — it is conditional and returns no class, so a gate naming "Tests 1-3" would be unsatisfiable for any claim that reads true, which is most of them. Ask: must a later reader *act* on this claim by measuring something?

- **No — the claim is narrative.** It describes what mainline carries, quotes a command as text, or records a coordinate as the subject of a correction. Nothing is measured at read time. → **S1**.
- **Yes — the claim asserts the current state of a moving thing and a reader will act on it.** → **S2**.

