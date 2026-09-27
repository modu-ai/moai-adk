# M2 attempt — AGENTS.md | 8. Harness-local instructions

surface = live
pre_chars = 875
post_chars = 861
chars = 14
post_hash = d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-live.

## before

````
## 8. Harness-local instructions

`CLAUDE.local.md` is a common local input shared with Claude workflows; `AGENTS.local.md` is the
Codex-specific input. For every local launch shape (bare, `cli`, `app`, `--spawn`, and `-w`),
`moai codex` reads the non-empty regular files from the project root in that order,
prefixes each body with its own provenance header, and passes the combined text as one session
`developer_instructions` override. A `-w` child still reads the original project root.

Shared `AGENTS.md` and `CLAUDE.md` never import or link either local file. The launcher refuses
links and non-regular inputs, reads through the descriptor it inspected, and fails before launch on
an operator-supplied `developer_instructions` collision or an oversized direct/spawn argument.
Codex Web sessions do not run the local MoAI launcher, so this injection is local-CLI-only.
````

## after

````
## 8. Harness-local instructions

`CLAUDE.local.md` (common, shared with Claude workflows) and `AGENTS.local.md` (Codex-specific) are
the local inputs. For every local launch shape (bare, `cli`, `app`, `--spawn`, `-w`), `moai codex`
reads the non-empty regular files from the project root in that order, prefixes each body with its
provenance header, and passes the combined text as one session `developer_instructions` override;
a `-w` child still reads the original project root.

Shared `AGENTS.md` and `CLAUDE.md` never import or link either local file. The launcher refuses
links and non-regular inputs, reads through the descriptor it inspected, and fails before launch on
an operator-supplied `developer_instructions` collision or an oversized direct/spawn argument.
Codex Web sessions do not run the local launcher, so this injection is local-CLI-only.
````
