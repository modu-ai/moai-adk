<!-- moai-ac-prefix: CR -->

# Fixture — native prefix declaration replaces the default grammar

Hand-derived expectation: live=1 excluded=1.

The declaration line REPLACES the default `AC` prefix for this file: its
criteria are declared under `CR`, so counting `AC` here would return 0 and a
0 is a RED flag, not a pass. The AC-shaped token below is a CROSS-REFERENCE —
it cites a criterion declared in another namespace, so it carries [REF] and
must stay excluded rather than being counted by a grammar that kept AC as a
second prefix.

### CR-01 — a live criterion under the declared prefix

### CR-02 [RETIRED] — a retired criterion under the declared prefix

- AC-EXT-01 [REF] cites a criterion declared elsewhere and is not counted.
