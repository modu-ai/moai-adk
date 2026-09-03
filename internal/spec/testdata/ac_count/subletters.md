# Fixture — trailing lowercase sub-letters are distinct identifiers

Hand-derived expectation: live=3 excluded=2.

A sub-lettered identifier is its own criterion, never folded into its numeric
prefix. The fourth line declares AC-003 and AC-003a together on purpose: under
the sub-letter-blind grammar both occurrences collapsed into one identifier,
so this line is the case the widened grammar must split.

### AC-SYN-007a — a live criterion carrying a sub-letter

### AC-SYN-008b [RETIRED] — a retired criterion carrying a sub-letter

- AC-SYN-09c [REF] points at a criterion declared elsewhere.

- AC-003 and AC-003a are two distinct live criteria in this fixture.
