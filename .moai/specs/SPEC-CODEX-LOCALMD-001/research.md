# research.md — SPEC-CODEX-LOCALMD-001

**Topic**: `moai codex` 런처가 `CLAUDE.local.md`를 `AGENTS.local.md`와 함께 `developer_instructions`로 적재하도록 확장
**Date**: 2026-09-22
**Generator**: FO-PLAN-1 3-lens fan-out run `wf_44049864-004`
**Tree**: `WT-codex-local-md@0314801c2` (`git diff develop --stat` 상 `internal/cli` delta 없음 — develop과 동일 내용)

아래 본문은 위 팬아웃 실행의 합성 결과를 원문 그대로 보존한 것이고, 문서 끝에 본 SPEC 작성자(manager-spec)의 재검증 기록을 덧붙인다.

---

## Preserved synthesis (verbatim from /tmp/t1078-research.md)

### 1. The seam already exists: one loader, one call site, six launch paths

The extension point is `codexLocalDeveloperInstructionArgs(projectRoot)` at `internal/cli/codex_launcher.go:114-142`. It currently implements the full pattern for exactly ONE file: `os.Lstat` → refuse if `!info.Mode().IsRegular()`, `os.ReadFile`, empty body → `nil` (no argv change), absent file → `nil`, then `json.Marshal` of the body into a single `-c developer_instructions=<JSON string>` argv token (a JSON string is a valid TOML basic string) (codebase-precedent).

Uniformity across every launch path is **structurally guaranteed, not per-path**: the loader is invoked exactly once, at `codex_launcher.go:495` inside `runCodexLaunch`, downstream of `codexInitOfferGate` (line 492) and before the spawn/direct fork (lines 501-504). Every launch shape funnels through it — `runCodex` → `runCodexLaunch` for bare/cli/app verbs; `--spawn` via `codexSpawnLaunch(req)` which receives the same assembled `req.Args`; `-w` only changes `req.Dir`; `-f` is applied before routing (codebase-precedent, constraints-risks). Consequence for design: the per-file guard and two-file synthesis logic must live INSIDE `codexLocalDeveloperInstructionArgs` (or its immediate call site); adding CLAUDE.local.md anywhere else forks the behavior (constraints-risks).

The new behavior therefore inherits uniformity for free by staying inside the existing funnel — but note the ordering requirement (CLAUDE first, AGENTS later) has no in-repo convention to inherit; see Section 3.

### 2. Per-file conventions the second file must replicate

- **Guard shape.** The repo's canonical per-path guard is `codexGuardInstructionPath` (`internal/cli/codex_contract.go:92-140`): per-component Lstat, closed-set `IsRegular` leaf judgement, parent-components-must-be-directories, `EvalSymlinks` containment ("judged by STRUCTURE, not by a snapshot"), with `codexModeName` (`codex_contract.go:144-163`) rendering the refused mode ("symlink", "named pipe", …). The launcher's own guard (`codex_launcher.go:119-129`) reuses `codexPathGuardError` and this diagnostic vocabulary; a second guarded file should follow the same shape, not invent new wording (codebase-precedent).
- **Fail-closed.** Any probe/read/encode error aborts the launch (`codex_launcher.go:495-498`, wrapped as `"load Codex local instructions: %w"`). `TestCodexLocalInstructions_SymlinkIsRefused` (`codex_local_instructions_test.go:121-139`) pins the exact error substring `"not a regular file (symlink)"` and requires `launches == 0` after refusal. No fail-open/skip-file path exists (codebase-precedent, constraints-risks).
- **Read-only inputs, test-pinned.** `TestCodexLocalInstructions_InjectedAsDeveloperInstructions` (`codex_local_instructions_test.go:44-50`) re-reads `AGENTS.local.md` after launch and asserts byte-identity; `codex_contract_link_test.go:173` counts renames == 0. The contract writer `secureCodexInstructionContract` guards `AGENTS.local.md` for containment (third entry of `codexInstructionRelPathsFn`, `codex_contract.go:54-56`) but only ever writes `AGENTS.md`/`CLAUDE.md`. The CLAUDE.local.md path must never be written, renamed, or linked (all three lenses).
- **SPEC-anchored read-only mandate.** SPEC-CODEX-LAUNCHER-001 REQ-CL-013 ("the launcher reads state and execs; it does not write") remains in force via SPEC-CODEX-LAUNCH-VERB-001 REQ-CLV-006 (no writes under `CODEX_HOME`), reaffirmed in the 2026-09-01 partial-supersession note (card t391) (prior-SPEC-memory).
- **Reusable guard precedents from SPEC history.** REQ-CI-011 (regular-file-in-project, symlink refusal → "out of bounds; neither read, write, nor import through such a path") and REQ-CL-008 (non-empty qualification: the JSON-string value, not raw bytes — `{ }`/`false`/`0`/`[]` must not qualify) are directly reusable patterns for per-file guards (prior-SPEC-memory).

### 3. Synthesis and encoding: one token, provenance ordering is net-new

- **Concatenate BEFORE marshaling.** The body is JSON-marshaled into ONE `-c developer_instructions=<json-string>` token (`codex_launcher.go:137-141`); tests round-trip it via `json.Unmarshal`. Two-file synthesis must concatenate bodies before marshaling so the result stays a single TOML basic-string token. Splitting into two `-c developer_instructions=` tokens risks duplicate-key behavior with unverified precedence (constraints-risks). Grep confirms `codex_launcher.go:141` is the sole argv producer of `developer_instructions` in `internal/` (other hits are generated `.codex/agents/moai/*.toml` and `internal/template/agentemit/writer.go`, a separate surface — the latter uses `literalString` TOML multi-line literals at `writer.go:38-109`, a different encoding precedent not applicable here).
- **Provenance separation has NO precedent.** Nothing in the repo combines two instruction-file bodies into one `developer_instructions` value, and no precedent exists for ordering TEXTUAL provenance of two sources. The only "which source wins" rationale anywhere is `codexChildEnv` (`codex_launcher.go:301-316`), where appended env entries win by last-wins — but that governs environment variables, not instruction text. The required CLAUDE-first / AGENTS-later ordering is net-new design, currently enforceable only by a new test (codebase-precedent, constraints-risks).
- **Empty-file edge must be made explicit.** Today an empty body returns `nil, nil`, indistinguishable from absent. With two files: "CLAUDE.local.md empty + AGENTS.local.md present" must still inject AGENTS content; "both present but empty" yields no `-c` at all. Per-file empty handling must be an explicit rule in the synthesis (constraints-risks), consistent with the REQ-CL-008 non-empty qualification tradition (prior-SPEC-memory).
- **No truncation, already honored.** The current loader imposes no size limit on the body; nothing in the codebase truncates instruction bodies today (codebase-precedent).

### 4. The contract this topic reverses — three generations of recorded intent

This is the pivotal finding, flagged independently by all three lenses. The local-instructions contract has a documented history, and the topic supersedes its most recent generation:

- **Generation 1 (superseded, 2026-08-24).** SPEC-CODEX-LAUNCHER-001 v0.6.0 named the local instruction file `CLAUDE.local.md` (REQ-CL-016); v0.7.0 split REQ-CL-015/016 out to SPEC-CODEX-INIT-001, whose REQ-CI-008 made the local uncommitted instruction file reachable from BOTH harnesses through the `@AGENTS.md` import chain, "referenced from exactly one place so it is not loaded twice" — i.e. a Claude-local file was once intended to feed Codex (prior-SPEC-memory).
- **Generation 2 (current, 2026-09-10).** Commit `6c647bbe2` ("feat(cli)!: align LLM harness instruction contract") deliberately REVERSED the shared-import model: renamed the file `AGENTS.local.md`, made it Codex-only, injected by the launcher as `-c developer_instructions`, with the recorded intent "without cross-importing Claude-local state." Corroborated by `CHANGELOG.md:77` ("CLAUDE.local.md, .claude/settings.local.json, and Claude automatic memory remain Claude-only…"), `AGENTS.md` §8 (~line 264: "`AGENTS.local.md` is Codex-only and uncommitted… shared `AGENTS.md` and `CLAUDE.md` never import it… not forwarded to Codex"), `docs-site/content/en/advanced/codex-dual-harness.md:33` (ko/ja/zh mirrors: "CLAUDE.local.md … remain Claude-only"), and the shipped template `internal/template/templates/AGENTS.md.tmpl` §8 (constraints-risks, prior-SPEC-memory, codebase-precedent).
- **The topic is therefore not extending silence — it reverses an explicit, multi-surface, documented contract.** How the lenses frame this differs (see contradictions): codebase-precedent calls it a three-surface reversal; prior-SPEC-memory calls it an explicit supersession that, read against Generation 1, is a *return to the original intent* via a different mechanism (the `-c` channel instead of the import chain). Either way, every recorded surface must change together: commit rationale, `CHANGELOG.md`, repo `AGENTS.md` §8, the `AGENTS.md.tmpl` template (which regenerates via `moai update`), the 4-locale docs-site pages, launcher help copy, and the help-pinning test (all three lenses).
- **Related standing decision.** SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 (M4) eliminated all 17 `CLAUDE.local.md` cross-references from `internal/template/templates/` — template-published content must never reference CLAUDE.local.md. The new launcher code is dev-owned Go with no template mirror, consistent with that boundary (prior-SPEC-memory).
- **Why injection exists at all.** Per `CHANGELOG.md:77`, the `-c developer_instructions` injection was chosen to bypass Codex's 32 KiB default project_doc limit — the no-truncation requirement for the ~61 KiB file aligns with why injection was chosen over the default doc path (prior-SPEC-memory).

### 5. Hard couplings: what must change together

- **The contract path table is pinned at exactly 3.** `secureCodexInstructionContract` (`codex_contract.go:222-227`) errors with "instruction path table must name exactly three paths" unless `codexInstructionRelPathsFn()` returns 3. CLAUDE.local.md must NOT be added to `defaultCodexInstructionRelPaths` (`codex_contract.go:54-56`) — that table is for guarded/linked shared files (AGENTS.md, CLAUDE.md, AGENTS.local.md), and tests override the seam with exactly-3 variant tables (constraints-risks).
- **Name constant SSOT.** `codexLocalInstructionName = "AGENTS.local.md"` (`codex_contract.go:33`); a sibling constant for CLAUDE.local.md is required, read only via the constant (constraints-risks).
- **Help copy is test-pinned.** `codexCmd.Long` (`codex_launcher.go:334-335`) documents the injection, and `TestCodexLocalInstructions_DocumentedInLauncherHelp` (`codex_local_instructions_test.go:141-147`) greps for `"AGENTS.local.md"`, `"developer instructions"`, `"Codex-only"`. Note "Codex-only" becomes inaccurate once CLAUDE.local.md is also loaded — help text and test must move in the same change (codebase-precedent, constraints-risks).
- **No gitignore change needed.** `internal/template/templates/.gitignore:174-175` already ignores both `/AGENTS.local.md` and `/CLAUDE.local.md` (constraints-risks).
- **No SPEC exists to amend.** No `.moai/specs/` file mentions `AGENTS.local.md`; commit `6c647bbe2` contains no `.moai/specs/` paths in its diff; no SPEC directory for this change (no SPEC-CODEX-LOCAL-MD or similar) exists (all three lenses).

### 6. Risks, several with no in-repo precedent

- **RISK — argv overflow, no guard anywhere.** Repo-root `CLAUDE.local.md` measures 61,360 bytes; JSON escaping grows newlines to `\n` and HTML-escapes `< > &` to `<` etc. Two ~60 KiB files concatenated in one argv token can approach/exceed Linux `MAX_ARG_STRLEN` (131,072 bytes/arg) → exec fails E2BIG. `codexLocalDeveloperInstructionArgs` has NO size check today; fail-closed-on-overflow is new code. Grep for `E2BIG|ErrArgListTooLong|argument list too long` across all `*.go` returns zero hits — first-of-kind in this repo. Additionally, on the `--spawn` path the whole argv is shell-quoted into a single tmux command string (`buildCodexSpawnCommand`, `codex_launcher.go:193-204`), a second length ceiling (codebase-precedent, constraints-risks).
- **RISK — `-w` worktree provenance mismatch.** Instructions are read from `projectRoot` (resolved from process cwd, `codex_launcher.go:468-474`) while the child `Dir` may be a resolved worktree (lines 478-486). `-w` injects the PRIMARY root's local files into a worktree session; the worktree's own untracked local files are never read. Existing behavior for AGENTS.local.md, doubled in consequence with two files (constraints-risks).
- **RISK — duplicate `-c` override from the operator tail.** `childArgs := append(localArgs, codexChildArgs(kind, tail)...)` (`codex_launcher.go:499`) places the launcher's override before the operator's verbatim tail; `-- -c developer_instructions=...` produces two overrides of one key with unverified precedence (constraints-risks).
- **RISK — TOCTOU.** The guard Lstats then ReadFiles; `ReadFile` follows symlinks, so a path swapped between the two calls escapes the guard. Pre-existing for AGENTS.local.md; per-file replication doubles the window rather than closing it (constraints-risks).
- **GAP — readout invisibility.** `moai codex status` (`codex_readiness.go`, six rows) has no local-instruction row, so a refusal condition (e.g. symlinked CLAUDE.local.md) is invisible in the readout; the only signal is the launch-path error wrap (constraints-risks).

### 7. Honest gaps and NONE-found results (preserved as reported)

- **NONE found — multi-file synthesis precedent.** Nothing in the repo combines two instruction bodies into one `developer_instructions` value (codebase-precedent).
- **NONE found — textual provenance-ordering precedent.** The env last-wins comment is the closest analogue but governs environment, not instruction text (codebase-precedent).
- **NONE found — argv-overflow handling anywhere** (`E2BIG`, `ErrArgListTooLong`, `argument list too long`: zero hits across all `*.go`) (codebase-precedent, constraints-risks).
- **NONE found — SPEC coverage.** No SPEC mentions `AGENTS.local.md`; none covers this change; no `.moai/reports/` measurement covers injection size/overflow characteristics (all three lenses).
- **Unverified — Codex CLI external semantics.** Neither lens could verify from in-repo evidence: how Codex resolves `developer_instructions` against discovered `AGENTS.md`/`AGENTS.override.md`/`project_doc_fallback_filenames` (the topic's official single-file-precedence assertion is taken from the task statement, not confirmed in-repo); whether `developer_instructions` composes with or replaces file discovery; and the precedence of two `-c` overrides of one key (codebase-precedent, constraints-risks).
- **Unverified — spawn-path length ceiling.** Whether tmux or the intermediate `/bin/sh -c` layer imposes an additional command-length ceiling on this Darwin host (constraints-risks).
- **Current repo state:** `AGENTS.local.md` is absent at the worktree root (only the absent path is exercised today); `AGENTS.local.md` loading has no dedicated SPEC — its record lives only in commit `6c647bbe2`, `CHANGELOG.md:77`, `AGENTS.md` §8, and the 4-locale docs page.

### contradictions

- **C1 — Reversal vs. supersession vs. return: how to characterize the topic's relationship to the recorded contract.** All three lenses agree on the facts (commit `6c647bbe2`, `CHANGELOG.md:77`, `AGENTS.md` §8, `docs-site/content/*/advanced/codex-dual-harness.md:33`, `AGENTS.md.tmpl` §8 all record CLAUDE-local-stays-Claude-only), but they characterize the change differently: **codebase-precedent** issues a "CONTRADICTION WARNING — the topic reverses a documented contract… This is not extending silence." **prior-SPEC-memory** frames it as "an explicit supersession" that the plan must mark, and adds that Generation 1 (SPEC-CODEX-LAUNCHER-001 v0.6.0 REQ-CL-016 + SPEC-CODEX-INIT-001 REQ-CI-008) originally intended exactly this reachability (CLAUDE.local.md feeding Codex via the import chain) — making the topic arguably a *return to Generation-1 intent* through the `-c` mechanism rather than a pure reversal. **constraints-risks** treats it operationally (the "Codex-only" help phrasing becomes inaccurate). The reader must decide the framing; all three concur that every recorded surface must change together.
- **C2 — Injection's relationship to Codex file discovery: "bypass/guard" (recorded) vs. "unverified composition" (examined).** **prior-SPEC-memory**, citing `CHANGELOG.md:77`, records that `-c developer_instructions` injection "exists precisely to bypass Codex's 32 KiB default project_doc limit" and "guards the root-plus-nested template discovery chain" — a framing in which injection coexists with (guards) file-based discovery. **codebase-precedent** (gap 3) states it could NOT verify how Codex resolves conflicting `developer_instructions` against root `AGENTS.md`/`AGENTS.override.md`, and that the task's official single-file-precedence premise is unconfirmed from in-repo evidence. **constraints-risks** states whether `developer_instructions` composes with or replaces discovered `AGENTS.md` is unverified, and duplicate `-c`-key precedence is unverified. The recorded rationale and the examined-code gap point in different directions; external Codex behavior remains unresolved.
- **C3 — Attribution of `codex_launcher.go` line 334.** **codebase-precedent** and **constraints-risks** both identify lines 334-335 as `codexCmd.Long` help copy; **prior-SPEC-memory**'s evidence list labels "codex_launcher.go line 334 (status readout text)." Minor factual discrepancy in line attribution — the reader should re-check the file to confirm which surface line 334 belongs to.
- **C4 (tension) — Robustness of the fail-closed guarantee.** **codebase-precedent** presents fail-closed as solid established precedent ("No fail-open/skip-file path exists"), pinned by `TestCodexLocalInstructions_SymlinkIsRefused`. **constraints-risks** shows the guarantee has a hole: Lstat→ReadFile TOCTOU means `ReadFile` can follow a symlink swapped in between the calls, escaping the guard — and per-file replication doubles the window rather than closing it. Both statements are individually true; they conflict on whether the existing fail-closed precedent is a guarantee to inherit or a weakness to at least not worsen.

### confidence

HIGH across all three lenses on the code-side mechanics (loader seam, single-call-site funnel, guard/fail-closed/read-only conventions, 3-path table coupling, test pins — every file read directly in this worktree) and on the two-generation contract history (verified verbatim in SPEC bodies, commit message, CHANGELOG, AGENTS.md §8, docs-site). HIGH that the topic supersedes the recorded "without cross-importing Claude-local state" decision and requires coordinated updates across code, help copy + its test, template, repo AGENTS.md, and 4-locale docs. The dominant unresolved items are external to the repo: Codex CLI's `developer_instructions` composition/precedence semantics (C2), spawn-path shell-length ceilings on this host, and the absence of any SPEC, multi-file-synthesis precedent, provenance-ordering precedent, or argv-overflow precedent for the net-new parts of the design.

---

## Author verification record (manager-spec, this run, tree WT-codex-local-md@0314801c2)

아래 항목은 본 SPEC 작성자가 이번 실행에서 직접 읽어 재확인한 근거다 (verification-claim-integrity §2 baseline 귀속).

### 공식 URL 검증 결과 (카드 인용 URL)

- **판정: VERIFIED.** `https://learn.chatgpt.com/docs/agent-configuration/agents-md` (WebFetch, 2026-09-22 이번 실행) 확인 결과:
  - "In each directory along the path, it checks for `AGENTS.override.md`, then `AGENTS.md`," 그 후에야 "any fallback names in `project_doc_fallback_filenames`" — precedence 서열 확인.
  - **"Codex includes at most one file per directory."** — 디렉터리당 최대 1개 파일 선택 명시. 빈 파일은 skip. 글로벌 레벨에서도 "uses only the first non-empty file at this level".
  - → **결론**: 루트에 `AGENTS.md`가 존재하면 fallback 이름으로 `CLAUDE.local.md`를 추가하는 접근은 동작하지 않는다 — 카드의 공식 제약 확인. `-c developer_instructions` 주입 경로가 유일한 수단이다.

### 코드 baseline 재확인 (file:line, 직접 판독)

- `codexLocalDeveloperInstructionArgs` — `internal/cli/codex_launcher.go:114-142`. Lstat→IsRegular→ReadFile→empty→nil, `json.Marshal` 후 단일 `-c developer_instructions=` 토큰 1쌍 반환 확인. truncation 없음 확인.
- 유일 호출 지점 — `internal/cli/codex_launcher.go:495` (`runCodexLaunch`), `codexInitOfferGate`(:492) 이후, spawn/direct 분기(:501-504) 이전. 오류 wrap `"load Codex local instructions: %w"`(:497). `childArgs := append(localArgs, codexChildArgs(kind, tail)...)`(:499) 확인.
- `buildCodexSpawnCommand` — `internal/cli/codex_launcher.go:193-204`. argv 전체를 토큰별 shell-quote해 단일 문자열화 확인 (spawn 경로 제2 길이 상한의 근거).
- `-w`는 child `Dir`만 변경 — `codex_launcher.go:478-486`; `projectRoot`는 `findProjectRootFn`/cwd에서 해석 — `:468-474`.
- `codexCmd.Long` 헬프 카피 — `internal/cli/codex_launcher.go:334-335` ("If AGENTS.local.md exists at the project root, its content is injected as Codex-only developer instructions…"). C3의 쟁점 해소: **334-335는 헬프 카피가 맞다** (status readout 아님).
- `codex_contract.go` — `codexLocalInstructionName = "AGENTS.local.md"`(:33), `defaultCodexInstructionRelPaths` 3-path 테이블(:54-56), `codexGuardInstructionPath`(:92-140), `codexModeName`(:144-163), exactly-3 핀(`:225-227`, `"instruction path table must name exactly three paths"`) 전부 직접 확인.
- `codex_local_instructions_test.go` — `TestCodexLocalInstructions_InjectedAsDeveloperInstructions`(:12), `_DirectSpawnAndAppSharePrefix`(:78), `_AbsentLeavesArgvUnchanged`(:109), `_SymlinkIsRefused`(:121-139, 오류 substring `"not a regular file (symlink)"` + `launches == 0` 핀), `_DocumentedInLauncherHelp`(:141-147, `{"AGENTS.local.md", "developer instructions", "Codex-only"}` grep 핀) 확인.
- 리포 루트 `CLAUDE.local.md` — **61,360 bytes** 측정 (`wc -c`, 이번 실행). `AGENTS.local.md`는 워크트리 루트에 부재 (`ls` 확인 — "No such file or directory").
- `CHANGELOG.md:77` — "…`CLAUDE.local.md`, `.claude/settings.local.json`, and Claude automatic memory remain Claude-only…" 및 32 KiB bypass 근거 문언 확인 (grep 직접 판독).
- 리포 `AGENTS.md:264` — "`AGENTS.local.md` is Codex-only and uncommitted. `moai codex` reads it from the project root and …" 확인. 템플릿 미러 `internal/template/templates/AGENTS.md.tmpl:270` 동일 문단 확인 ("…are not forwarded to Codex." 포함).
- `docs-site/content/en/advanced/codex-dual-harness.md:33` — "`AGENTS.local.md` is Codex-only … `CLAUDE.local.md` … remain Claude-only." 확인.
- `internal/template/templates/.gitignore:174-175` — `/AGENTS.local.md`, `/CLAUDE.local.md` 양쪽 기존 ignore 확인 (gitignore 변경 불요).

### phase 라벨 귀속

- git tag 목록에서 최신 semver 태그는 `v3.1.2` (`moai_cp/*` 항목은 빌드 식별 태그로 제외), `CHANGELOG.md` 최상단은 `[Unreleased]` → 다음 미출시 버전 **v3.1.3** 을 release target으로 기재.

### 남는 미검증 (Gaps)

- Codex CLI의 `developer_instructions`와 파일 discovery(루트 `AGENTS.md` 등) 간 compose/replace 관계 — 공식 문서는 파일 선택 규칙만 답했고 `-c` override와의 상호작용은 이번 fetch로도 확인되지 않았다. LIVE acceptance(AC-LMD-012)가 이 갭의 실측 수단이다.
- 본 호스트(Darwin)의 tmux/`/bin/sh -c` 명령 길이 상한 — 미측정.
