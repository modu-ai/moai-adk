# Personal supplement (AGENTS.local.md)

Machine- and operator-specific notes for this checkout. Tracked by the operator's decision; keep secrets, tokens and private paths out — write `~` or the variable name instead.

- Response language: Korean (colloquial in chat, clean written register in reports). Code, identifiers, commands and paths stay verbatim.
- Operator: single developer on a Mac; GitHub auth through `gh` as the operator. Peer Claude sessions are labelled `leader`, `agent-1` … `agent-N`; `agent-2` audits read-only. Extra sessions are opened only by the operator.
- Build cache for this machine: `CARGO_HOME=/tmp/moai-rc-product-cargo`, `CARGO_TARGET_DIR=/tmp/moai-rc-product-target` (debug binary at `/tmp/moai-rc-product-target/debug/moai-cli`). Rebuild with `cargo build --offline --locked` under the same variables before using `moai-cli` here.
- Live Go MoAI (v3.2.1-rc.6) is installed on this machine and in this folder (`.moai/`, `.claude/`, `~/.moai/`, Go `moai` on PATH). Preserve all of it; dogfood the Rust generator in temporary projects or on this repository's own `AGENTS.*` only.
- Development sessions on this machine (operator decision 2026-11-08): start with `moai-cli session --role leader --name leader --advisor fable` or `moai-cli session --role agent --name agent-<n> --advisor fable` (main model `opus`, effort `medium`; `--dry-run` shows the composed `claude` command). Fable is the advisor for hard reasoning only; moving the choice into a setting is later work. Local Claude Code is 2.1.319, above the AGENTS.md (v2.1.277) and Fable (v2.1.257) thresholds.
- Local tools available to sessions on this machine: Hide (operator workspace, `hide` CLI), ego-browser (logged-in browsing for Page reads), `gh`. None of them grants publication or deletion permission.
- Memory: session lessons live in the Claude auto-memory store for this project (`feedback_*.md`, `project_*.md`, indexed by `MEMORY.md`); re-read the index at session start.
