# Label Decision Tree — Full Criteria

Level-3 reference for `moai-workflow-docs-claim-check`. Load when a label is
contested, when two adjacent gates both look plausible, or when a composite
claim resists splitting.

## 1. The ordered tree

Walk the gates top to bottom. The **first** gate whose condition holds wins and
labeling stops. A later gate never overrides an earlier one, and an earlier gate
is never re-opened once passed.

```
 1. needs-human      -- can this be settled here at all?
        | no gate hit
 2. stale-suspected  -- was it true once, but the evidence has moved on?
        | no gate hit
 3. verified         -- does supplied evidence directly carry it?
        | no gate hit
 4. unsupported      -- default; attach exactly one reason
```

Why this order:

- `needs-human` is first because an unanswerable claim must not be scored on
  evidence that was never capable of settling it.
- `stale-suspected` precedes `verified` so a claim matching an *old* evidence
  item is not passed on that stale match.
- `stale-suspected` also precedes `unsupported` so an outdated-but-once-true
  claim is not reported as if it had never been true. The distinction matters:
  staleness is a maintenance task, unsupported is a correctness defect.
- `unsupported` is the default, so a claim can never fall through unlabeled.

## 2. Gate criteria in full

### Gate 1 — `needs-human`

Fires when settling the claim requires any of:

- Executing a command (build, test, install, package-manager, CLI invocation).
- Reaching a system this assessment cannot read: a private registry, a
  dashboard, a paid service, a production host.
- Exercising interactive behavior: a UI flow, a prompt sequence, a rendered page.
- A judgment reserved to a maintainer: whether a limitation is intentional,
  whether a roadmap statement is still endorsed, whether an omission is policy.
- Hardware or platform access that the evidence set does not include.

Every `needs-human` row must name **what** would settle it — the command plus
its target file, or the system and the specific reading needed. A bare
"needs a human" with no named next step is an incomplete row.

### Gate 2 — `stale-suspected`

Fires on a **temporal mismatch**: the evidence shows the claim's shape was once
correct, but a version, date, count, or name has since moved.

Typical shapes:

- A version number in the document is lower than the one in the evidence.
- A supported-version floor no longer matches the manifest.
- A count ("supports 12 providers") disagrees with a current enumeration.
- A renamed command, flag, package, or endpoint still appears under its old name.
- A "latest" or "current" statement pinned to a superseded value.

Distinguish from `contradicted` (a reason under `unsupported`): staleness means
the claim *was* accurate and drifted; contradiction means the evidence asserts
the opposite outright, with no plausible earlier moment at which the claim held.
When the evidence cannot distinguish the two, prefer `stale-suspected` — the
label is explicitly a suspicion and says so.

### Gate 3 — `verified`

Fires only when **all** of the following hold:

- A supplied evidence item speaks directly to the claim.
- That item can be named as an anchor (file, log, manifest, page, plus the
  locating detail: line, field, or section).
- The evidence covers the claim's full scope, not a subset of it.
- The evidence is current with respect to the document's own version, or the
  claim is version-independent.

If coverage is partial, do not weaken `verified` with a caveat — fall through
to `unsupported` / `insufficient-coverage`. A hedged pass is not a pass.

### Gate 4 — `unsupported`

The default. Attach exactly one reason:

| Reason | Fires when | Distinguishing question |
|--------|-----------|-------------------------|
| `missing-evidence` | Nothing supplied speaks to the claim | "Is there any item at all on this topic?" — no |
| `contradicted` | A supplied item asserts the opposite | "Does an item say the reverse?" — yes |
| `insufficient-coverage` | An item is on-topic but narrower than the claim | "Does the item cover the whole claim?" — no |

Reason tie-breaks:

- On-topic but silent about the specific assertion → `missing-evidence`, not
  `insufficient-coverage`. Coverage means partial support, not mere adjacency.
- Contradicts one part of a claim that was not split → the claim was not atomic;
  return to triage and split it.
- Contradicts and is also outdated → prefer `stale-suspected` (gate 2 fires
  first and gate order is binding).

## 3. Adjacent-gate tie-breaks

| Tension | Resolution |
|---------|-----------|
| `needs-human` vs `unsupported` | If evidence *could* exist but was not supplied, the label is `unsupported` / `missing-evidence`. `needs-human` is for claims no supplied evidence could settle. |
| `needs-human` vs `verified` | If the supplied evidence already settles it, gate 1 does not fire. Do not escalate a settled claim. |
| `stale-suspected` vs `verified` | Evidence matches but is older than the document's stated version → `stale-suspected`. |
| `stale-suspected` vs `contradicted` | Is there a plausible earlier moment when the claim held? Yes → stale. No → contradicted. |
| `insufficient-coverage` vs `missing-evidence` | Does any item support *part* of the claim? Yes → insufficient-coverage. No → missing-evidence. |

## 4. Claim-type table

Reference for what evidence a claim type normally needs, and where each type
tends to land.

| Claim type | Example shape | Evidence that settles it | Frequent label without it |
|-----------|---------------|--------------------------|---------------------------|
| Platform support | "Runs on Linux, macOS, Windows" | Per-platform build or test output | `insufficient-coverage` |
| Version floor | "Requires runtime 3.11 or newer" | Dependency manifest or lockfile | `missing-evidence` |
| Install step | "Install with a single command" | Recorded install transcript | `needs-human` |
| Default value | "Timeout defaults to 30 seconds" | Config schema or defaults source | `missing-evidence` |
| Capability | "Exports to CSV and Parquet" | Feature source, tests, or CLI help text | `insufficient-coverage` |
| Count | "Ships 40 built-in rules" | Enumerated listing | `stale-suspected` |
| Freshness | "Latest release is 2.9.0" | Tag or release listing | `stale-suspected` |
| Performance | "Processes 10k records per second" | Benchmark output with method | `needs-human` |
| Compatibility | "Works with any S3-compatible store" | Integration matrix | `insufficient-coverage` |
| Guarantee | "No data leaves the machine" | Design plus a network-behavior record | `needs-human` |
| Subjective | "Blazing fast", "intuitive" | None possible | excluded from labeling |

## 5. Composite-claim splitting guidance

Split until each atomic claim carries exactly one assertion. Four signals that
a sentence is still composite:

1. **Conjunction across subjects** — "and" / "or" joining platforms, versions,
   formats, or environments. Split per member of the list.
2. **A qualifier with its own truth value** — "installs in one command *without
   root*" holds two assertions: the one-command install and the no-root
   condition.
3. **A quantifier over a set** — "all adapters support retries" is one claim per
   adapter when the adapters are enumerable, and one coverage claim when they
   are not.
4. **A mixed subjective and checkable pair** — "fast and reliable on Linux"
   yields one checkable claim (Linux support) and two exclusions.

Splitting rules:

- Split, never merge. Two weak claims are more informative than one blurred row.
- Preserve the document's own wording in each atomic claim so a maintainer can
  find the sentence.
- If splitting produces a claim no evidence could ever address, that claim is
  still a row — usually `needs-human` or `missing-evidence`.
- Record the parent sentence once and number its atomic children so the mapping
  back to the document stays visible.

Stop condition: a claim is atomic when flipping any single fact in the evidence
could change its label. If two independent facts could each change the verdict,
split again.
