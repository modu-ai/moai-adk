# Fixture — three markup shapes, two countable states

Hand-derived expectation: live=3 excluded=3.

Every identifier below is uniformly marked or uniformly unmarked across all of
its occurrences, so this fixture counts rather than halts. The retired heading
criterion carries two marked occurrences and is the partial-marking mutation
target; this paragraph deliberately names no identifier, because an unmarked
mention here would itself make that identifier ambiguous.

## Shape A — heading

### AC-SYN-001 — a live criterion declared by this fixture

### AC-SYN-002 [RETIRED] — a retired criterion declared by this fixture

## Shape B — table cell

| ID | Requirement | Note |
|---|---|---|
| AC-SYN-003 | REQ-001 | live |
| AC-SYN-004 [RETIRED] | REQ-002 | retired |

## Shape C — two-digit inline

- AC-SYN-05 is a live criterion written inline with a two-digit identifier.
- AC-SYN-06 [REF] points at a criterion declared elsewhere.

## Second occurrence of the retired heading criterion

The retirement of AC-SYN-002 [RETIRED] is restated here so the fixture carries a
multi-occurrence identifier whose occurrences are uniformly marked.
