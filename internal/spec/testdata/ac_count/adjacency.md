# Fixture — the six adjacency cases

Hand-derived expectation: live=4 excluded=2.

1. Adjacent marking: AC-SYN-010 [REF] — excluded.
2. Same line, not adjacent: AC-SYN-011 is live and AC-SYN-012 [RETIRED] is not.
3. Token placed before: [RETIRED] AC-SYN-013 — live, a leading token marks nothing.
4. Floating token, adjacent to no identifier at all: [REF]
5. Closing backtick intervenes: `AC-SYN-014` [REF] — live, the backtick breaks adjacency.
6. A newline intervenes below: the identifier ends its line, the token opens the next.

AC-SYN-015
[REF]
