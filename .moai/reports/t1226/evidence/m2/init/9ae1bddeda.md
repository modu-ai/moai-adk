# M2 attempt — AGENTS.md | 8. Harness-local instructions

surface = init
pre_chars = 568
post_chars = 540
chars = 28
post_hash = 93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477

post_hash is the frozen-multiset sha256 of the whole post-attempt surface (every attempted M2 row replaced, every ADMIT M1 row removed), measured with the AC-ALH-004 pipeline on $SCRATCH/post-init.

## before

````
## 8. Harness-local instructions

`moai codex` reads non-empty common local guidance followed by Codex-specific local guidance
from the project root. It preserves each input body and adds source headers in one session
`developer_instructions` override. Shared instruction documents never import these local inputs.
Other harness-local settings and memory remain owned by their harness.

Codex Web sessions read `AGENTS.md`, but local `.codex/hooks.json`, the status line, and the MoAI
launcher injection do not run there. Treat Web sessions as read-and-review first.

````

## after

````
## 8. Harness-local instructions

`moai codex` reads non-empty common local guidance, then Codex-specific local guidance, from the
project root, keeping each body and adding source headers in one session `developer_instructions`
override. Shared instruction documents never import these inputs; other harness-local settings and
memory stay owned by their harness.

Codex Web sessions read `AGENTS.md`, but `.codex/hooks.json`, the status line and the MoAI launcher
injection do not run there — treat Web sessions as read-and-review first.

````
