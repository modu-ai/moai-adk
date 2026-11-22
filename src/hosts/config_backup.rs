//! SPEC-20261121-retired-hosts-read-safely (REQ-9): the size bound that the Go-era state preview
//! applies to a `.codex` config file. The backup writer that lived here beside it served the
//! Codex config writer and the Codex hook files, and went with them
//! (SPEC-20261121-remove-codex-project-layer).

/// Size bound for the config file, its receipt and the rendered result.
pub const CONFIG_LIMIT: u64 = 262_144;
