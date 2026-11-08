# P02: interview result and package discovery

User selection: use a versioned interview result schema; prefer an explicit repository package digest pin; when no pin exists, use the user-level package. The already-decided first full deep interview remains unchanged.

Alternatives presented: always use the user-level package unless separately requested; or require an independent selection for every project/cloud environment. The chosen approach enables reproducible shared projects without forcing every project to pin.

A minimal schema v1, JSON-subset YAML creation, full support-tree digest and explicit pin selection are now connected to the root Rust `moai-cli`; see `rust-product-path.md`. Lossless YAML reconfiguration and host-specific discovery/loading still require implementation and verification. A missing or mismatched explicitly pinned package must not be silently treated as the requested pinned package. This record does not assert actual Claude or Codex discovery, nor does it decide P05.

Evidence: [RC1 plan](https://chatgpt.com/space/page_43faf222414c81919de71f7d7830e26c), [decision checklist](https://chatgpt.com/space/page_cfa4c2d5bea481918690cc7e131bd86f), [home and project contract](https://chatgpt.com/space/page_83f162ef14d08191917913738184f167).
