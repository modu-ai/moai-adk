# Acceptance Criteria — SPEC-V3R6-DOCS-PREREWRITE-README-001 (revision 0.1.1)

## A. Pin and conventions

- **Document-level pin.** Tree commit `cb30e05b3e5f64774b8dabd87e80c24332f4551d` (short `cb30e05b3e`), tree object `9c08830bdec23b81334840307ee745e468500174`. Every cell that reads the tree names the commit in its command, so the pin is in the cell, not only in this header. Measured 2026-10-26 (before 2026-10-26T10:31:30Z) in this plan pass.
- **Cell form (verification-completeness §2.1).** A RED-now cell is one read-only invocation with no `;`, `&&`, `|`, subshell, redirection, or `echo $?` wrapper. Its verbatim stdout and its exit code are separate fields. Zero-match exit codes are confirmed with a measurement wrapper run **outside** the cell and are marked `wrapper-confirmed`.
- **RED-now.** A cell is red when it reports a nonzero match count, a mismatched version string, or (for the link cells) a missing required link.
- **Classes.** `release-blocking (gated)`: re-executable on the tree, counted toward the completion verdict once its gate rows carry verdicts. `regression-guard`: not counted as a pass; the reason is stated in §D.
- **Naming.** This child's evidence cells carry the `R-` prefix (README child); each maps to the umbrella cell (`E-` id at `bdb7b764`) whose command form it adapts, scoped to this child's file set and re-measured on `cb30e05b3e`. No umbrella figure is carried over; where this tree differs from the umbrella's tree (e.g. the badge shows v3.2.0, not the v3.1.3 the umbrella measured at `8816de4dae`), the fresh value governs.

## A.1 Traceability matrix (AC to REQ)

| AC | REQ | Milestone (plan.md §B) | Gate rows | Class | Ledger entries |
|----|-----|------------------------|-----------|-------|----------------|
| AC-PRC-001-r | REQ-PRC-001 (README part), REQ-README-001, -002, -003 | M2, M4 | Q13, Q17, QR-2, QR-3 | release-blocking (gated) | R-01, R-01n, R-02, R-03, R-04 |
| AC-PRC-003 | REQ-PRC-003, REQ-README-004 | M2+M3 (one commit) | Q8 (leader-lens resolution) | regression-guard | R-20a, R-20b (tree); R-20c (external); R-20d |
| AC-PRC-004-r | REQ-PRC-004, REQ-README-002 | M4 | Q1, QR-1 | release-blocking (gated) | R-08, R-08r, R-08b, R-08b2 |

## B. Acceptance criteria

### AC-PRC-001-r — the README stale citations are gone (REQ-PRC-001, README partial criterion)

- **AC-PRC-001-r** Given the tree at `cb30e05b3e`, the scoped E-01 family matches 4 lines in each of the four README files (R-01: 16 matching lines, exit 0), the scoped E-02, E-03, and E-04 families match nothing (R-02, R-03, R-04: no output, exit 1, wrapper-confirmed), and none of the 16 matching lines carries an example marker (R-01n lists them verbatim).
- When M2 and M4 apply the Q13/Q17/Q1 edits to the four README files, then each of the four scoped family cells re-run on the closing tree exits 1 with empty stdout (R-01 turns from 16 lines to no output; R-02 to R-04 stay empty).
- Then the README partial criterion of the umbrella's AC-PRC-001 is green when, additionally, the diff touches no file outside the edit surface of spec.md § 2 (REQ-README-004).
- RED-now cells: R-01 (red: 16 lines). R-02, R-03, R-04 are green-at-arrival baseline cells (the umbrella's README part is red through family F1 only); they are recorded so the closing run proves no new stale form appeared.
- Gate rows: Q13, Q17, QR-2, QR-3.
- Done when: all four scoped family cells exit 1 with empty stdout on the closing tree, and `git diff --name-only` between the pre-run and closing trees lists only the six surface files plus `.moai/specs/SPEC-V3R6-DOCS-PREREWRITE-README-001/` and `.moai/reports/t1720/`.

### AC-PRC-003 — the badge and hugo.toml name the Q8 resolution (REQ-PRC-003) — regression-guard

- **AC-PRC-003** Given the four README Release badges (line 24 of each) and hugo.toml lines 55–56 name v3.2.0 / 2026-10-23 (R-20a, R-20b), and the public release list marks v3.1.2 Latest (R-20c, external) with no v3.2.0 release object, while local tags include v3.2.0 but no v3.1.3 (R-20d).
- When the Q8 leader-lens resolution (tag `v3.1.2`, `releaseDate = "2026-08-21"`) is applied, then the four badges carry `img.shields.io/badge/Release-v3.1.2-blue.svg` and hugo.toml lines 55–56 carry `version = "v3.1.2"` and `releaseDate = "2026-08-21"`, and the badge lines and both hugo.toml lines change in the SAME commit (REQ-PRC-003; plan.md M2+M3 land as one commit).
- RED-now cells: R-20a, R-20b (tree, pinned; red because they name v3.2.0, which the external observation shows has no release object). R-20c is external, recorded with its UTC time, not tree-attributed, never counted as a pass (verification-completeness §2.1 undecidable disposition).
- Class: regression-guard. Reason: the closing comparison needs the external Latest-release state, which the repository cannot observe; the guard holds because a regression (badge drifting from the public Latest again) is re-detectable by re-observation, but the verdict is recorded as an observation, not a pass.
- Gate row: Q8 (leader-lens resolution for this card; the umbrella row stays open).
- Done when (guard): on the closing tree, R-20a shows `Release-v3.1.2-blue.svg` on all four lines, R-20b shows `v3.1.2` / `2026-08-21`, the same-commit property is shown by the **parent-diff assertion** (D3 repair): for the single closing commit C (the M2+M3 commit), `git diff C^ C -- README.md README.ko.md README.ja.md README.zh.md docs-site/hugo.toml` shows, in that ONE diff, all six line changes together — the four badge-line changes (line 24 of each README, each hunk changing `Release-v3.2.0-blue.svg` to `Release-v3.1.2-blue.svg`) and the two hugo.toml hunks (lines 55–56, to `version = "v3.1.2"` and `releaseDate = "2026-08-21"`); a two-commit split (badges in one commit, hugo values in another) FAILS this check even when both final values hold and a file listing names all five files, and a fresh external observation (`gh release list`) is recorded with its UTC time.

### AC-PRC-004-r — the CG retirement passage: README.ko copies and the missing link (REQ-PRC-004, this child's part)

- **AC-PRC-004-r** Given nine Korean docs-site pages carry the exact sentence (R-08), README.ko.md carries it twice (R-08r: 2), seven of the eight non-canonical ko pages already link to the canonical page, and `core-concepts/what-is-moai-adk.md` has no such link (R-08b: 7 lines, what-is absent).
- When M4 applies the Q1 edits, then README.ko.md carries zero copies of the sentence (R-08r re-run: exit 1, no output) and instead carries a link in the destination form `](/ko/multi-llm/cg-mode/)` (count ≥ 1, cell R-08b-r1), the en/ja/zh README counterpart paragraphs are link sentences to their locale's cg-mode page (QR-2 default; their lines are E-01-flagged and close under AC-PRC-001-r), and `docs-site/content/ko/core-concepts/what-is-moai-adk.md` carries the link in the same destination form (cell R-08b-r2: count ≥ 1; under the QR-1 default its own sentence copy is replaced by the link sentence, so R-08 on that path turns to no output).
- Scope boundary (recorded, not hidden): the seven other ko pages keep their sentence copies after this child lands — they are outside this child's surface (leader-lens scope, spec.md § 2) — so the umbrella's full AC-PRC-004 done check (E-08 listing exactly one file) is NOT achieved by this child alone and stays with the umbrella.
- RED-now cells: R-08r (red: 2 copies), R-08b (red for what-is: absent from the link baseline), R-08 (context: 9 pages, of which this child's surface covers what-is only).
- Gate rows: Q1, QR-1.
- Done when: R-08r exits 1 with no output; the destination-form link cell on README.ko.md (`git grep -c -F "](/ko/multi-llm/cg-mode/)" <closing> -- README.ko.md`; cell R-08b-r1) returns ≥ 1; the destination-form link cell on what-is (`git grep -c -F "](/ko/multi-llm/cg-mode/)" <closing> -- docs-site/content/ko/core-concepts/what-is-moai-adk.md`; cell R-08b-r2) returns ≥ 1; and (QR-1 default) R-08 scoped to what-is returns no output. The link cells grep the link-destination form `](/ko/multi-llm/cg-mode/)` — the on-convention form measured 20/20 on the tree's existing ko pages — so a comment-only mutant (a `<!-- multi-llm/cg-mode -->` remark with zero actual links) fails the check (D2 repair).

## C. Evidence ledger (verbatim; all tree cells pinned to `cb30e05b3e`)

### R-00 — pins

- `git rev-parse HEAD` — exit 0 — stdout: `cb30e05b3e5f64774b8dabd87e80c24332f4551d`
- `git rev-parse 'HEAD^{tree}'` — exit 0 — stdout: `9c08830bdec23b81334840307ee745e468500174`

### R-01 — scoped E-01 family (RED-now; umbrella E-01 adapted to the four README files)

- Command: `git grep -c -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" cb30e05b3e -- README.md README.ko.md README.ja.md README.zh.md`
- Exit code: 0.
- Stdout (verbatim; 4 lines):

```
cb30e05b3e:README.ja.md:4
cb30e05b3e:README.ko.md:4
cb30e05b3e:README.md:4
cb30e05b3e:README.zh.md:4
```

- Reading: 16 matching lines in 4 files. The README part of the umbrella's E-01 is red on this tree. The umbrella's tree `8816de4dae` showed the same per-file count for the READMEs (4 each); the badge family (R-20a) is what changed between the trees (v3.1.3 → v3.2.0).

### R-01n — the 16 matching lines verbatim (supporting; edit-site pin)

- Command: `git grep -n -E "(^|[^/])moai (cg|gpt|audit|codemaps|model)([^[:alnum:]-]|$)" cb30e05b3e -- README.md README.ko.md README.ja.md README.zh.md`
- Exit code: 0.
- Stdout (verbatim; 16 lines):

```
cb30e05b3e:README.ja.md:58:1 回の起動に進入トークンは 1 つだけなので、`-f` と `-l` の併用はエラーになる。値を付けた形（`moai cc -f <値>`、`moai cc -l lane-2`）も 1 行のエラーで拒否され、エラー文が正しい形を教えてくれる。かつてのマルチセッション・ボードモードの `-k` は削除済みで、`moai cg` は移行案内を表示して終了する（`moai migrate cg` で先に確認できる）。`moai gpt` は存在せず、GPT モデルは `moai codex` で動かす。
cb30e05b3e:README.ja.md:290:> Claude Code に載せていた旧 `moai gpt` ランチャーは削除されました。GPT モデルはネイティブハーネスの
cb30e05b3e:README.ja.md:385:`moai cg` は廃止されました。Claude や GLM を起動せず、移行案内を表示して終了します。`moai cc` の別名ではありません。`llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。
cb30e05b3e:README.ja.md:590:`moai cg` は廃止されました。Claude や GLM を起動せず、移行案内を表示して終了します。`moai cc` の別名ではありません。`llm.team_mode: cg` が残るプロジェクトでは、セッションを起動する前に移行先を明示的に選ぶ必要があります。
cb30e05b3e:README.ko.md:58:한 번의 실행에는 진입 토큰이 하나뿐이라 `-f`와 `-l`을 함께 쓰면 에러다. 값이 붙은 형태(`moai cc -f <값>`, `moai cc -l lane-2`)도 한 줄 에러로 거부되고, 에러 문구가 올바른 형태를 알려 준다. 예전 다중 세션 보드 모드의 `-k`는 제거되었고, `moai cg`는 이전 안내를 띄우고 종료한다(`moai migrate cg`로 먼저 확인). `moai gpt`는 없으며 GPT 모델은 `moai codex`로 돌린다.
cb30e05b3e:README.ko.md:291:> **GPT 게이트웨이 철회(2026-09-16).** 내장 번역 게이트웨이로 GPT 모델을 Claude Code에 얹던 옛 `moai gpt`
cb30e05b3e:README.ko.md:386:`moai cg`는 폐기되었습니다. Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다.
cb30e05b3e:README.ko.md:591:`moai cg`는 폐기되었습니다. Claude나 GLM을 실행하지 않고 설정 이전 안내와 함께 종료합니다. `moai cc`의 별칭이 아닙니다. `llm.team_mode: cg`가 남은 프로젝트는 세션을 실행하기 전에 이전할 구성을 명시적으로 선택해야 합니다.
cb30e05b3e:README.md:58:A launch carries one entry token, so `-f` together with `-l` is an error. A value after either token (`moai cc -f <value>`, `moai cc -l lane-2`) is refused with a one-line error that names the correct form. The `-k` entry of the former multi-session board mode is removed, and `moai cg` exits with a migration notice (preview it with `moai migrate cg`). `moai gpt` does not exist: GPT models run through `moai codex`.
cb30e05b3e:README.md:296:> **GPT gateway withdrawn (2026-09-16).** The former `moai gpt` launcher — Claude Code driven by GPT
cb30e05b3e:README.md:387:`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session.
cb30e05b3e:README.md:592:`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session.
cb30e05b3e:README.zh.md:58:一次启动只能带一个进入令牌，所以 `-f` 与 `-l` 同时给出会报错。带了值的形式（`moai cc -f <值>`、`moai cc -l lane-2`）会被一行错误拒绝，错误信息会指出正确的形式。早先那种多会话分列运行的模式所用的 `-k` 已被移除；`moai cg` 会显示迁移提示后退出（可先用 `moai migrate cg` 预览）。`moai gpt` 不存在，GPT 模型通过 `moai codex` 运行。
cb30e05b3e:README.zh.md:289:> **GPT 网关已撤回（2026-09-16）。** 通过内置翻译网关把 GPT 模型接入 Claude Code 的旧 `moai gpt`
cb30e05b3e:README.zh.md:584:`moai cg` 已停用。它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。
cb30e05b3e:README.zh.md:589:`moai cg` 已停用。它会显示迁移提示并退出，不会启动 Claude 或 GLM，也不是 `moai cc` 的别名。项目中若仍有 `llm.team_mode: cg`，必须先明确选择迁移方案，才能启动会话。
```

- Reading: per-locale edit sites — ko 58/291/386/591, en 58/296/387/592, ja 58/290/385/590, zh 58/289/584/589. No line carries an example marker. Lines 384–387 and 589–592 per locale are the two CG retirement paragraphs (ko 386/591 carry the exact sentence of R-08r).

### R-02 — scoped E-02 family (baseline; green at arrival)

- Command: `git grep -c -E "(^|[^/])moai (codex (job|task|setup)|todo (rename|pick)|spec list)" cb30e05b3e -- README.md README.ko.md README.ja.md README.zh.md`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.

### R-03 — scoped E-03 family (baseline; green at arrival)

- Command: `git grep -n -E "(^|[^/])moai (run|sync) SPEC-ID" cb30e05b3e -- README.md README.ko.md README.ja.md README.zh.md`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.

### R-04 — scoped E-04 family (baseline; green at arrival)

- Command: `git grep -c -E "/moai[: ](design|issue|bug|fb|init|brain|relationship)([^[:alnum:]-]|$)" cb30e05b3e -- README.md README.ko.md README.ja.md README.zh.md`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.

### R-20a — README Release badges (RED-now; umbrella E-20a re-measured)

- Command: `git grep -n -F "img.shields.io/badge/Release-" cb30e05b3e -- README.md README.ko.md README.ja.md README.zh.md`
- Exit code: 0.
- Stdout (verbatim; 4 lines):

```
cb30e05b3e:README.ja.md:24:  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.2.0-blue.svg" alt="Release"></a>
cb30e05b3e:README.ko.md:24:  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.2.0-blue.svg" alt="Release"></a>
cb30e05b3e:README.md:24:  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.2.0-blue.svg" alt="Release"></a>
cb30e05b3e:README.zh.md:24:  <a href="https://github.com/modu-ai/moai-adk/releases"><img src="https://img.shields.io/badge/Release-v3.2.0-blue.svg" alt="Release"></a>
```

### R-20b — hugo.toml version lines (RED-now; umbrella E-20b re-measured)

- Command: `git grep -n -E "version|releaseDate" cb30e05b3e -- docs-site/hugo.toml`
- Exit code: 0.
- Stdout (verbatim):

```
cb30e05b3e:docs-site/hugo.toml:55:  version = "v3.2.0"
cb30e05b3e:docs-site/hugo.toml:56:  releaseDate = "2026-10-23"
```

### R-20c — public release list (external; not tree-attributed; observed 2026-10-26, before 2026-10-26T10:31:30Z)

- Command: `gh release list --repo modu-ai/moai-adk --limit 6` — exit 0 — stdout (verbatim; tab-separated):

```
v3.1.3	Draft	v3.1.3	2026-10-23T21:05:18Z
v3.2.1	Draft	v3.2.1	2026-08-21T19:28:08Z
v3.1.2	Latest	v3.1.2	2026-08-21T06:50:41Z
v3.1.1		v3.1.1	2026-08-20T09:15:39Z
v3.1.0		v3.1.0	2026-08-15T14:55:53Z
v3.0.1		v3.0.1	2026-07-24T03:57:57Z
```

- Reading: Latest = v3.1.2, published 2026-08-21T06:50:41Z (the Q8 `releaseDate` value). v3.1.3 and v3.2.1 are Drafts. No v3.2.0 release object exists — the tree's badge names a version the release list does not carry (the "3.2 rollback" state).

### R-20d — local tags (supporting; Q8 evidence)

- Command: `git tag -l 'v3.*'` — exit 0 — stdout (verbatim; trimmed to the non-rc v3 tags): `v3.0.0`, `v3.0.1`, `v3.1.0`, `v3.1.1`, `v3.1.2`, `v3.2.0` (full output also lists the `v3.0.0-rc*`, `v3.1.0-rc.*` tags).
- Reading: no `v3.1.3` tag exists locally; `v3.2.0` is tagged locally but has no release object (R-20c). No new tag is created by this SPEC (REQ-README-004).

### R-08 — CG retirement sentence copies, Korean docs-site pages (RED-now context; umbrella E-08 re-measured)

- Command: `git grep -c -F "설정 이전 안내와 함께 종료합니다" cb30e05b3e -- docs-site/content/ko`
- Exit code: 0.
- Stdout (verbatim; 9 lines):

```
cb30e05b3e:docs-site/content/ko/advanced/tokenomics-overview.md:1
cb30e05b3e:docs-site/content/ko/cli-reference/launchers.md:1
cb30e05b3e:docs-site/content/ko/core-concepts/what-is-moai-adk.md:1
cb30e05b3e:docs-site/content/ko/cost-optimization/prompt-caching.md:1
cb30e05b3e:docs-site/content/ko/getting-started/cli.md:1
cb30e05b3e:docs-site/content/ko/getting-started/introduction.md:1
cb30e05b3e:docs-site/content/ko/getting-started/windows-guide.md:1
cb30e05b3e:docs-site/content/ko/multi-llm/_index.md:1
cb30e05b3e:docs-site/content/ko/multi-llm/cg-mode.md:1
```

- Reading: of these nine, this child's surface covers `core-concepts/what-is-moai-adk.md` only (QR-1 default); the other seven are outside this child's surface.

### R-08r — CG retirement sentence copies, README.ko (RED-now)

- Command: `git grep -c -F "설정 이전 안내와 함께 종료합니다" cb30e05b3e -- README.ko.md`
- Exit code: 0.
- Stdout (verbatim; 1 line):

```
cb30e05b3e:README.ko.md:2
```

- Reading: lines 386 and 591 (R-01n). This is the Q1 edit site.

### R-08b — cg-mode link baseline on the eight non-canonical ko pages (RED-now; umbrella E-08b re-measured)

- Command: `git grep -c -F "multi-llm/cg-mode" cb30e05b3e -- docs-site/content/ko/advanced/tokenomics-overview.md docs-site/content/ko/cli-reference/launchers.md docs-site/content/ko/core-concepts/what-is-moai-adk.md docs-site/content/ko/cost-optimization/prompt-caching.md docs-site/content/ko/getting-started/cli.md docs-site/content/ko/getting-started/introduction.md docs-site/content/ko/getting-started/windows-guide.md docs-site/content/ko/multi-llm/_index.md`
- Exit code: 0.
- Stdout (verbatim; 7 lines; `what-is-moai-adk.md` has no match):

```
cb30e05b3e:docs-site/content/ko/advanced/tokenomics-overview.md:1
cb30e05b3e:docs-site/content/ko/cli-reference/launchers.md:2
cb30e05b3e:docs-site/content/ko/cost-optimization/prompt-caching.md:2
cb30e05b3e:docs-site/content/ko/getting-started/cli.md:1
cb30e05b3e:docs-site/content/ko/getting-started/introduction.md:1
cb30e05b3e:docs-site/content/ko/getting-started/windows-guide.md:2
cb30e05b3e:docs-site/content/ko/multi-llm/_index.md:2
```

### R-08b2 — cg-mode links in the en/ja/zh what-is counterparts (supporting observation)

- Command: `git grep -c -F "multi-llm/cg-mode" cb30e05b3e -- docs-site/content/en/core-concepts/what-is-moai-adk.md docs-site/content/ja/core-concepts/what-is-moai-adk.md docs-site/content/zh/core-concepts/what-is-moai-adk.md`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.
- Reading: all four locales' what-is pages lack the link. This child's surface names the ko page (leader-lens scope); whether the en/ja/zh link additions ride in the same change under Q17 is run-phase scope discipline — they are NOT in this child's named surface and are recorded here so the omission is a decision, not an oversight.

### R-08b-r1 — destination-form cg-mode link on README.ko (RED-now; D2 repair cell)

- Command: `git grep -c -F "](/ko/multi-llm/cg-mode/)" 5fe7d20f46 -- README.ko.md`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.
- Reading: README.ko.md carries zero links in the destination form at the audited tree `5fe7d20f46` (the two retirement paragraphs at 386/591 are sentence copies, not links) — RED for the right stated reason. This cell replaces the bare-substring form (`git grep -c -F "multi-llm/cg-mode"`) in the AC-PRC-004-r done-when: that form counted a comment mentioning the path, so a comment-only mutant passed it (plan-audit D2). Positive control (measured this pass, same tree): `git grep -c -F "](/ko/multi-llm/cg-mode/)" 5fe7d20f46 -- docs-site/content/ko/cli-reference/launchers.md docs-site/content/ko/getting-started/windows-guide.md` → 2 and 2, exit 0 — the destination form is the live convention, so the zero above is an absence of the link, not a dead pattern.

### R-08b-r2 — destination-form cg-mode link on ko what-is (RED-now; D2 repair cell)

- Command: `git grep -c -F "](/ko/multi-llm/cg-mode/)" 5fe7d20f46 -- docs-site/content/ko/core-concepts/what-is-moai-adk.md`
- Exit code: 1 (wrapper-confirmed).
- Stdout: no output.
- Reading: same destination form as R-08b-r1, scoped to this child's page; RED at the audited tree `5fe7d20f46` for the same reason (R-08b already showed the page absent from the link baseline). Replaces the R-08b bare-substring re-run in the done-what link check (D2).

### R-27 — SPEC ID pre-write self-check (protocol; not a RED cell)

- Command: `ID="SPEC-V3R6-DOCS-PREREWRITE-README-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` — exit 0 — stdout: `PASS`
- Uniqueness: `ls .moai/specs | grep -c "SPEC-V3R6-DOCS-PREREWRITE-README-001"` — exit 0 — stdout: `0` (no directory with this ID existed before this authoring pass).

## D. Definition of done — closing table

| AC | REQ | Class | Gate rows | Closing check (re-run on the closing tree, single-invocation cells) |
|----|-----|-------|-----------|------------------------------|
| AC-PRC-001-r | REQ-PRC-001 (README part) | release-blocking (gated) | Q13, Q17, QR-2, QR-3 | R-01, R-02, R-03, R-04 each exit 1 with empty stdout; diff confined to the § 2 surface |
| AC-PRC-003 | REQ-PRC-003 | regression-guard | Q8 (leader-lens) | R-20a shows `Release-v3.1.2-blue.svg` ×4; R-20b shows `v3.1.2` / `2026-08-21`; same-commit property shown via the parent-diff assertion `git diff C^ C -- README.md README.ko.md README.ja.md README.zh.md docs-site/hugo.toml` exhibiting all six line changes (4 badge + 2 hugo) in ONE diff (D3 repair; a two-commit split fails); fresh R-20c observation recorded with its UTC time, not counted as a pass |
| AC-PRC-004-r | REQ-PRC-004 (this child's part) | release-blocking (gated) | Q1, QR-1 | R-08r exits 1 with no output; destination-form link cells R-08b-r1 (README.ko) and R-08b-r2 (what-is) each ≥ 1; (QR-1 default) R-08 scoped to what-is: no output |

## E. Gaps (not observed)

- GR-1 — The verb-resolution side (which `moai` verbs the tree build registers) was not re-measured in this pass; this child inherits the umbrella's E-06/E-07 measurements at `f35103480f` as the umbrella's verification cells and does not re-cite them as its own. The scoped closing check (pattern match) does not depend on them.
- GR-2 — The external release observation (R-20c) is one observation at 2026-10-26T10:31:30Z; it is not the closing-time state. The closing guard re-observes and records the time.
- GR-3 — The en/ja/zh what-is cg-mode links (R-08b2) are outside this child's named surface; whether they are added by a SWEEP child or a later change is not decided here.
- GR-4 — The seven already-linking ko pages keep their sentence copies after this child lands; the umbrella's full AC-PRC-004 closure is not achieved by this child alone.
- GR-5 — The strict docs i18n check was not run in this pass (this child's page edit is a one-line link addition inside ko; the run-phase verification sweep runs it — plan.md §B M5).
- GR-6 — Q1, Q13, Q17, QR-1, QR-2, QR-3 carry planner defaults, not operator verdicts. Q8's resolution is a leader-lens card-scoped ruling, and the umbrella Q8 row stays open.

## F. Revision history

| Date | Revision | Author | Change |
|------|----------|--------|--------|
| 2026-10-26 | 0.1.0 | manager-spec | Initial plan-phase authoring; all tree cells pinned to `cb30e05b3e` (see spec.md HISTORY). |
| 2026-10-26 | 0.1.1 | manager-spec | Repair per plan-audit verdict FAIL 0.88 at commit `5fe7d20f46` (card t1720; verdict at `.moai/reports/t1720/plan-audit-verdict.md`, gitignored). D2: the AC-PRC-004-r link cells (README.ko done-when, what-is re-run, §D row) now grep the destination form `](/ko/multi-llm/cg-mode/)` instead of the bare substring; fresh RED-now cells R-08b-r1 / R-08b-r2 added, measured on `5fe7d20f46` (both: no output, exit 1, positive control live). D3: the AC-PRC-003 same-commit check replaced — `git log -1 --name-only` file-listing (passed a two-commit mutant) superseded by the parent-diff assertion over `C^ C` for the single closing commit. D1 is repaired in spec.md (REQ-README-004 badge-line exception). |
