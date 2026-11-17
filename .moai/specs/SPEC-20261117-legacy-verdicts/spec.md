---
id: SPEC-20261117-legacy-verdicts
title: "레거시 자산 1,107행 판정 — 상태 어휘와 고정 시험"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

`.moai-cli/manifests/legacy-asset-map.json`의 Go 세대 자산 1,107행이 전부 `unchecked`였고,
`tests/product_cli.rs`의 고정 시험이 그 상태를 강제했어요. 2026-11-17 사용자 결정 "개념은 새 설계로, 파일은
안 옮김"(D8 = A 유지)은 행마다 판정을 주고 `unchecked` 행을 0으로 만들라고 해요. 이 SPEC은 상태 어휘
`adopted(<path>)`·`rejected(<reason>)`·`deferred(<design or SPEC id>)`를 넣고, 모든 행을 판정하고, 시험이 어휘와
판정마다 경로·사유·대상을 요구하게 바꿔요. 판정은 커밋한 스크립트가 경로에서 다시 만들어요.

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`). `When` 요구가 없어 실패 경로 요구도 없어요.

- REQ-1: Each row of the legacy asset map shall carry one status of the form `adopted(<path>)`, `rejected(<reason>)` or `deferred(<design or SPEC id>)`.
- REQ-2: The legacy asset map shall hold no row whose status is `unchecked`.
- REQ-3: Each `adopted` verdict shall belong to one of the 16 shipped rows and name the file that carries it.
- REQ-4: Each `rejected` verdict shall state a reason of at least three words that claims a replacement only where one exists.
- REQ-5: Each `deferred` verdict shall name one of the four designs of the 2026-11-17 decision or an existing SPEC folder.
- REQ-6: The source manifest of the skill package shall pin the map's sha256 and its verdict counts.
- REQ-7: The change shall create no `.moai-cli/manifests/host-capabilities.json`, `resources/` or `adapters/codex/` path.
- REQ-8: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-legacy-verdicts/spec.md`.
- REQ-9: The script `tools/legacy_verdicts.py` shall regenerate the committed map byte for byte.

## Acceptance

- AC-1 (REQ-1…REQ-6): 고정 시험이 어휘·채택 목록·경로·사유·대상·개수를 확인해요.
  verify: `cargo test --locked --offline --test product_cli source_manifest_covers_every_package_file`
- AC-2 (REQ-2): `unchecked` 행이 없어요.
  verify: `grep -c '"status": "unchecked"' .moai-cli/manifests/legacy-asset-map.json` → `0`
- AC-3 (REQ-6, REQ-9): 스크립트가 지도를 그대로 다시 만들고, 개수가 manifest와 같아요.
  verify: `python3 -I tools/legacy_verdicts.py` → exit 0, `{"verdicts": {"adopted": 16, "deferred": 300, "rejected": 791}, "rows": 1107}`;
  `python3 -m unittest tests/legacy_verdicts_test.py`
- AC-4 (REQ-7): 목적지 충돌 경로를 만들지 않았어요.
  verify: `ls -d .moai-cli/manifests/host-capabilities.json resources adapters/codex` → 셋 다 없음(exit 1)
- AC-5 (REQ-1…REQ-9): 바뀐 곳에 닿는 검사가 통과해요.
  verify: `moai spec check --all --project .`(이 가지에서 빌드한 바이너리); `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --bin moai`;
  `cargo test --locked --offline --test go_state_classes_cli`; `cargo test --locked --offline --test launcher_three_hosts`;
  `cargo test --locked --offline --test ac_command_lint`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests/source_manifest_test.py tests/legacy_verdicts_test.py`

## Scope

- write: `.moai-cli/manifests/legacy-asset-map.json`, `skills/moai/references/source-manifest.json`, `tests/product_cli.rs`,
  `tools/legacy_verdicts.py`, `tests/legacy_verdicts_test.py`, `docs/rust-product-path.md`,
  `.moai/specs/SPEC-20261117-legacy-verdicts/spec.md`
- read-only: `docs/decisions-2026-11-08.md`, `skills/moai/**`, `templates/claude/output-styles/**`, `src/agents.rs`, `agents/**`,
  Go 원본 `4fa2aa9db2`(git 객체)
- forbidden: `.moai-cli/manifests/host-capabilities.json`, `resources/`, `adapters/codex/` 생성, 실제 HOME, 네트워크 쓰기

## Plan

1. RED: 고정 시험을 새 어휘로 바꾸고, 옛 지도(전부 `unchecked`)에서 실패하는지 확인.
2. 판정: `tools/legacy_verdicts.py`가 경로 규칙으로 1,107행을 판정해요. 지도 모양(머리 2칸 들여쓰기, 한 줄에 한 행)은
   그대로 두고 `status`와 `purpose`만 바꿔요.
3. manifest의 sha256·`status`·`verdicts`, `docs/rust-product-path.md`의 "Every row stays `unchecked`" 문장.
4. 검사, 커밋, mutation, `## Result`.
5. 감사 FAIL(B1, N1–N5) 수리: 진단 hook 행, navigator 행, 가이드가 아닌 참고 skill 3개, 채택 목록 고정, 사유 정직성,
   스크립트 커밋과 재현 시험.

## Decisions

- D1 채택 16행(사용자 요청 목록, 카드 노트): `.claude/skills/moai/SKILL.md`·`workflows/{plan,run,sync}.md`(각 2벌) →
  `skills/moai/…` 같은 이름 파일, `AGENTS.md`·`AGENTS.md.tmpl` → `src/agents.rs`, 출력 스타일 3개(각 2벌) →
  `templates/claude/output-styles/`(카드 노트의 `adapters/claude/output-styles/`는 #2709에서 옮겨졌어요). 시험의
  `LEGACY_ADOPTED`와 스크립트의 `ADOPTED`가 같은 목록이고, 다른 행의 `adopted`는 시험이 거부해요.
- D2 보류 대상(행 수):
  - `design:code-atlas-read-plan` 22 — codemaps(작업 흐름·입구·명령), 프로젝트 codebase-analysis, plan
    context-discovery, `moai-workflow-project`의 navigator 참고·스크립트(codemaps와 이어진 AST 층, 리더 결정 N1).
  - `design:language-domain-guides` 232 — 언어 규칙, quality 규칙, coding-standards, `moai-domain-*`, `moai-ref-*`
    (아래 셋 제외), `moai-foundation-quality`.
  - `design:edit-time-diagnostics-hook` 10 — `moai-workflow-loop`(SKILL.md:4 "Automated feedback loop with LSP
    diagnostics and AST-grep", :40 Write/Edit 뒤 PostToolUse LSP 진단, `lsp.yaml`)과 router `workflows/fix.md`·`loop.md`
    (LSP 진단 sweep). 감사 B1이 고친 판정이에요.
  - `design:project-agent-skill-generation` 36 — harness 작업 흐름·입구·명령, builder-harness 역할, agent/skill 작성
    규칙 3개, `moai-meta-harness`(빈 `seeds/.gitkeep` 2행 제외), `moai-harness-learner`, `moai-jev-skill-suggestion`.
- D3 hook 규칙 `core/hooks-system.md`·`core/agent-hooks.md`는 거절이에요: Go hook 스크립트의 일반 규칙이고, PostToolUse
  LSP 검증은 시간 제한 표의 한 줄(hooks-system.md:365)로만 나와요. 진단 경로를 설명하지 않아요.
- D4 거절 사유는 대체물이 있을 때만 그 경로를 적고, 없으면 "not ported (D8 = A)"예요(감사 N4). 대체물이 있다고 적은
  것: 입구(`skills/moai/SKILL.md`), router `moai.md`·`project*`·`review`(→ `audit.md`)·`gate`(→ `verify.md`)·`plan/*`·
  `run/*`·`sync/*`, `moai-workflow-tdd`(→ `references/internal/test-design.md`·`tdd-results.md`), `-spec`(→ `moai spec`),
  `-worktree`(→ `moai worktree-plan`), `-docs-claim-check`(→ `references/internal/doc-claims.md`), `moai-ref-aside-browser`
  (→ `references/adapters/aside.md`), `@MX`(→ ACC `@cc:` 설계). 규칙·역할·hook 규칙은 "not ported"에 Rust 쪽 위치만
  덧붙였어요(대체한다고 하지 않아요). `moai-ref-cross-model-audit`·`moai-ref-jev-question-design`은 가이드가 아니라
  거절이에요(리더 결정 N2).
- D5 규칙 행의 목적지 충돌(host-capabilities.json, `resources/`, `adapters/codex/`)은 거절 사유에 적어 정리했고 파일은
  만들지 않았어요.
- D6 `@MX` 행은 보류가 아니라 거절이에요: 네 설계에 들지 않고, ACC 설계가 옛 표식을 읽기만 한다고 이미 정했어요.
- D7 빈 `.gitkeep` 10행은 "empty placeholder file … nothing to port"로 거절해요.
- D8 판정은 `tools/legacy_verdicts.py`(Python `-I`, 결정적)가 만들어요. 인자 없이 돌리면 다시 만든 지도와 커밋된 지도를
  비교해 다르면 exit 1, `--write`는 지도를 써요. 규칙에 없는 경로는 추측하지 않고 `ValueError`로 멈춰요.

## Result

(검사 뒤에 채워요.)
