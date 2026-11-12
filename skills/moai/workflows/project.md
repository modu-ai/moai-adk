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
2. **Interview by kind.** Ask in at most three rounds, one axis per question, with the host question tool when there is one, else in the conversation.
   - New project: purpose, users, tools and languages to start with, development method, hosts, and the collaboration rules.
   - Existing project: confirm the detected tools, languages and databases instead of asking for them; then ask for purpose, development method, hosts, and what must not change.
   - An axis still unanswered after the last round is recorded in `decisions` as `{"id": "<axis>", "topic": "<axis>", "value": "HOLD", "source": "default", "scope": "project"}` and is asked again on the next run. It is never guessed.
3. **Plan and apply.** Put the answers into `result.profile` (purpose, tools, development methods, hosts, decisions) and pass it to `moai-cli project-plan <project>`. Show the plan. After the user approves it, run `moai-cli project-apply <project>` with the plan digest and a permission reference. A re-run with the same facts and answers plans `unchanged`.
4. **Refresh the managed section.** When AGENTS.md exists, regenerate its managed section through `agents-plan` and `agents-apply`, so it reflects the applied profile. Text outside the markers is never touched.
5. **Offer the code map.** `result.next_actions` carries `moai-cli index-build <project>` (stdin `{"protocol_version":1}`). Offer it and run it only after the user agrees. Detection never builds an index.

## Exceptions and Permissions

Detection reads only files inside the project, never follows symbolic links, and stops at a fixed depth and file count. It never reads HOME, never probes PATH and never starts a program, because `project.yaml` is tracked and holds project facts only. MCP configuration files are read for server names only; commands, environment values and tokens are dropped and never reach the output. A profile written by hand in general YAML is refused by `project-plan` until it is reviewed.

## Completion Check

The applied profile's digest matches the approved plan, `detected` matches a fresh `project-detect`, no MCP configuration was created or changed, no index was built without approval, and every unanswered axis is recorded as HOLD. Host use of `/moai project` is verified only after the Rust skill replaces the installed one (not_run until then).
