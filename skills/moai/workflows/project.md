# PROJECT

`/moai project` (`$moai project` in Codex CLI) describes the current project for MoAI: it tells a new project from an existing one, runs the matching interview, records what it detected, and offers the code map. It extends SETUP and keeps SETUP's approval rules. The project facts live in `.moai-cli/project.yaml` and the AGENTS.md managed section; no product, structure or tech documents are written (user decision 16 = B, Issue #2496).

## Input

The current folder, the user's request, and the existing `.moai-cli/project.yaml` when there is one. A folder that a first launch already initialized (`git init`, `.moai-cli/project.yaml`, AGENTS.md — Issue #2481) is still a new project when it holds no code.

## Output

A complete profile for `project-plan` (the detected facts and the interview decisions), the applied profile after the user approves the plan, and one suggested `index-build` command. Nothing is written before the plan is approved.

## Procedure

1. **Detect.** Run `moai-cli project-detect <project>`. It reads the folder and writes nothing. Read `result.detected`:
   - `project_kind` is `new` when no manifest and no code file is found, else `existing`.
   - `manifests`, `languages`, `lsp`, `mcp_servers` and `databases` are facts found in the project's own files. `lsp` names the language servers the languages call for; it does not say a server is installed. `mcp_servers` names servers already configured in the project; MoAI never creates or edits an MCP configuration.
   - Detection reads the project root only for manifests, the dependencies that name `databases`, `prisma/schema.prisma` and the MCP configuration files. A subfolder's `package.json` and a `docker-compose.yml` database image are not read; `languages` alone also counts code files down to three folder levels. The interview fills in what root detection does not see.
   - `project_kind` speaks of code only. A folder with nothing but `prisma/schema.prisma` or `.codex/config.toml` is `new`, and its `databases` and `mcp_servers` are still recorded: they are configuration facts the new-project interview starts from.
   - When `result.profile` is `null`, the folder has no `.moai-cli/project.yaml` yet and `next_actions` says to create the profile first. Run SETUP (or let a first launch create it), then run `project-detect` again. Never write a profile by hand.
   - `result.go_state` reports Go-generation project state, read only and without following links: `moai_dir` is `absent`, `folder`, `link` or `other`, and `go_manifest` says whether the folder holds the Go `manifest.json`. It sits beside `detected` and never enters the profile.
2. **Go-generation state (only when `go_state.moai_dir` is `folder`).** Ask the user whether to move the earlier MoAI generation's state aside now. Without a yes, skip this step; nothing moves.
   - After a yes, run the read-only preview `moai-cli go-state-plan <project>` and show its `items` (`move` and `leave`, each with its reason), the `restores` (`.moai/specs` copied into `.moai-cli/specs`, secrets excluded), and every `warnings` line. Running sessions that read `.claude/` or write `.moai/` cannot be detected, so say so and ask the user to close them first.
   - Apply only after a second yes to that preview: `moai-cli go-state-apply <project> --plan-digest <result.plan_digest> --permission-reference <ref>`. Everything that moves goes to `.moai-cli/backups/<UTC>/go-state/`, and nothing is deleted. A changed project refuses the digest; preview again instead of retrying.
   - Tell the user the receipt path and the way back: `moai-cli go-state-revert-plan <project> --backup <UTC>` shows it, and `moai-cli go-state-revert <project> --backup <UTC> --plan-digest <that plan_digest> --permission-reference <ref>` runs it after they agree.
   - While `.claude/settings.json` or `settings.local.json` still names a path under `.moai/`, the preview keeps the whole folder (`leave`). Report that, and do not edit the settings to get past it.
3. **Interview by kind.** Ask in at most three rounds, one axis per question, with the host question tool when there is one, else in the conversation.
   - New project: purpose, users, tools and languages to start with, development method, hosts, and the collaboration rules.
   - Existing project: confirm the detected tools, languages and databases instead of asking for them; then ask for purpose, development method, hosts, and what must not change.
   - An axis still unanswered after the last round is recorded in `decisions` as `{"id": "<axis>", "topic": "<axis>", "value": "HOLD", "source": "default", "scope": "project"}` and is asked again on the next run. It is never guessed.
4. **Plan and apply.** Put the answers into `result.profile` (purpose, tools, development methods, hosts, decisions) and pass it on stdin to `moai-cli project-plan <project>`. Show the plan. After the user approves it, run `moai-cli project-apply <project> --plan-digest <result.plan_digest> --permission-reference <ref>` with the plan output's `result` object on stdin, not the whole response: the envelope is refused as `invalid typed JSON input`. A re-run with the same facts and answers plans `unchanged`.
   - A plan of `action: conflict` on a profile that has no baseline yet (no `.moai-cli/manifests/profile-baseline.json`, for example a `schema_version` 1 file) means the tool cannot tell an interview answer from a hand edit, so every changed key is held. Apply in two steps: first plan and apply `result.profile` with every existing value kept and only the detected facts added — that writes the baseline — then plan the interview answers again; they now apply as `update`. Never edit `project.yaml` by hand to get past the conflict.
5. **Refresh the managed section.** When AGENTS.md exists, regenerate its managed section through `agents-plan` and `agents-apply`, so it reflects the applied profile. Text outside the markers is never touched.
6. **Offer the code map.** `result.next_actions` carries `moai-cli index-build <project>` (stdin `{"protocol_version":1}`). Offer it and run it only after the user agrees. Detection never builds an index.

## Exceptions and Permissions

Detection reads only files inside the project, never follows symbolic links, and stops at a fixed depth and file count. It never reads HOME, never probes PATH and never starts a program, because `project.yaml` is tracked and holds project facts only. MCP configuration files are read for server names only; commands, environment values and tokens are dropped and never reach the output. A profile written by hand in general YAML is refused by `project-plan` until it is reviewed.

## Completion Check

The applied profile's digest matches the approved plan, `detected` matches a fresh `project-detect`, no MCP configuration was created or changed, no index was built without approval, and every unanswered axis is recorded as HOLD, and Go-generation state moved only after the user agreed to the preview it was shown (with its receipt and revert command reported). Host use of `/moai project` is verified only after the Rust skill replaces the installed one (not_run until then).
