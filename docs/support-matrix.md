# 지원표 (08-D #2015, RC1 — 게시 전 상태)

이 표의 유일한 입력은 08-C E2E receipt(`docs/e2e-receipts/receipts.jsonl`, `schemas/e2e-receipt.schema.json`)다. RL-10: receipt가 없는 조합은 지원으로 표기하지 않고, 지원 행 수 == passed receipt 수. 검증은 `python3 -I tools/check_support_matrix.py docs/support-matrix.md docs/e2e-receipts/`(exit 0 = 일치).

**상태 한계(숨기지 않음)**: 이 표의 passed 행은 T1 결정적 fixture(호스트 `none`) 5건과 T2 정적 검증 1건이다. 실제 호스트 로딩(certified) — Claude/Codex가 `/moai`·`$moai`로 skill을 로딩하는 E2E — 는 운영자 실측(T2, 자격증명 필요) 전까지 전부 `not_run`이며, "지원" 행이 곧 실사용 보증이 아님을 이 문서 머리에 명시한다. RC1은 게시 전 단계다.

## 지원 대상 조합 (P04)

rc.1의 지원 대상은 **Darwin/arm64(`aarch64-apple-darwin`) × Claude Code, Darwin/arm64 × Codex** 두 조합뿐이다(사용자 결정 P04, #2262). 아래 표에서 "지원 대상"은 범위를 정한 것이지 지원 보증이 아니다. 보증은 passed receipt가 있는 행만 하고(RL-10), 두 조합의 host 로딩 E2E(T2)는 receipt가 오기 전까지 `not_run`이다.

| 조합 | rc.1 범위 | 현재 결과 |
| --- | --- | --- |
| Darwin/arm64 × Claude Code | 지원 대상 | T2 not_run(AC-08C-3·6·8 절반), 정적 검증 AC-08C-5만 passed |
| Darwin/arm64 × Codex | 지원 대상 | T2 not_run(AC-08C-4) |
| Darwin/x86_64(`x86_64-apple-darwin`, Tier 2) | 미지원(후속) | 측정하지 않음 |
| Linux(ubuntu-24.04, arm 변형 포함) | 미지원(후속) | 측정하지 않음 |
| Windows(windows-2025) | 빌드 + self-replace 구현만(결정 01:1x(11-09), E3 #2195) | **not_run** — CI `windows-build` job이 모든 target을 컴파일만 한다. 시험 실행·실행 중 교체·host 로딩은 측정하지 않음 |
| ZCode·그 밖의 host | 미지원(후속) | 측정하지 않음 |

## 조합 범위 (P12)

RC1의 실측 조합은 개발 Mac 단일이다: Darwin / arm64 (Rust triple `aarch64-apple-darwin`, Tier 1). 나머지 OS·arch·mode는 전부 `not_run`이다(숨김 없음, OS matrix workflow는 RC1에서 만들지 않는다).

## 지원 행 (passed receipt 기반 — 6행)

| OS/arch (Rust tier) | host | host exact version | mode | skill digest (sha256) | moai-cli SHA (sha256) | AC | capability | discovery | 명시 호출 | SETUP | UPDATE | 권한 거부 | restore | 결과 | receipt |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Darwin/arm64 (Tier 1, aarch64-apple-darwin) | none (T1 fixture) | null | local | `bff1c53b61ae749391f93e74f554d97a118b5984a970cc6a3d198f700f9d9f58` | `178d87719ebc58a8e8e44aa80be75b261a88fd28bcb63b5af12a59a9db6ceb38` | AC-08C-1 | setup→update→중단→복구 전 경로 | — | — | plan/apply(간접) | ✓ (v1→v2) | — | ✓ known-good 복구 + 사용자 편집 바이트 보존 | passed | receipts.jsonl |
| Darwin/arm64 (Tier 1) | none (T1 fixture) | null | local | `bff1c53b…` | `178d8771…` | AC-08C-2 | UPDATE 무결성 fail-closed | — | — | — | ✓ (변조 stage 거절, HOME 트리 바이트 동일) | — | — | passed | receipts.jsonl |
| Darwin/arm64 (Tier 1) | none (T1 fixture) | null | local | `bff1c53b…` | `178d8771…` | HA-02 | UPDATE 재실행 | — | — | — | ✓ (전 행 unchanged/preserved, diff 0) | — | — | passed | receipts.jsonl |
| Darwin/arm64 (Tier 1) | none (T1 fixture) | null | local | `bff1c53b…` | `178d8771…` | AC-08C-7 | 사용자 HOME 쓰기 차단 — **macOS 전용 측정**(`sandbox-exec` deny-home; 시험은 `#[cfg(target_os = "macos")]` 한정 — Linux에서는 미컴파일, 숨은 skip 아님. **Linux CI 통과는 AC-08C-7 측정이 아니다**, #2100 감사 지적) | — | — | — | ✓ (실제 HOME 리다이렉션 없이 성공 — 사용자 상태 비의존) | ✓ (deny 프로필에서 쓰기 불가 관측) | — | passed | receipts.jsonl |
| Darwin/arm64 (Tier 1) | none (T1 fixture) | null | local | `bff1c53b…` | `178d8771…` | AC-08C-8 | clean clone 포터블 READ/PLAN 자산 | ✓ (SKILL.md·routes.json·참조 workflow 전부 존재) | — | — | — | — | — | passed | receipts.jsonl |
| Darwin/arm64 (Tier 1) | Claude Code | 2.1.319 (Claude Code) | local | `bff1c53b…` | absent (설치본 없음, 04-A 미발행) | AC-08C-5 | skill 정적 검증 (`claude plugin validate`) | ✓ (frontmatter·구조 유효) | — | — | — | — | — | passed | receipts.jsonl |

## not_run 행 (receipt 없음 — 숨김 없음)

| OS/arch | host | mode | 항목 | 결과 | 사유 |
| --- | --- | --- | --- | --- | --- |
| Darwin/arm64 | Claude Code | local | 명시 호출이 SETUP route 도달 (AC-08C-3) | not_run | 운영자 실측 대기 — `ANTHROPIC_API_KEY` 필요(체크아웃 밖 1회) |
| Darwin/arm64 | Claude Code | local | 권한 거부 시 apply 차단·트리 불변 (AC-08C-6) | not_run | 동일(운영자 실측 대기) |
| Darwin/arm64 | Codex | local | 명시 호출이 SETUP route 도달 (AC-08C-4) | not_run | `CODEX_API_KEY` 필요(운영자 실측 대기) |
| Darwin/arm64 | Claude Code | local | clean clone READ (AC-08C-8 T2 절반) | not_run | 운영자 실측 대기 |
| ubuntu-24.04 / arm 변형 / Darwin x86_64 | — | local | 전체 | not_run | 미지원(후속) — P04로 rc.1 범위 밖. OS matrix workflow는 RC1에서 만들지 않음(08-D 소유) |
| windows-2025 | — | local | 빌드(`cargo build --offline --locked --all-targets`) | CI 실측(PR의 `windows-build` run) | E3 #2195 — 컴파일 증거일 뿐 지원 보증이 아니다 |
| windows-2025 | — | local | 시험 실행·실행 중 binary 교체(self-replace)·host 로딩 | not_run | 결정 01:1x(11-09): Windows는 빌드 + self-replace 구현만. 교체 경로는 `src/self_update.rs`의 모의 replacer 단위 시험(macOS)으로만 확인했고, 실제 `self_replace` 호출은 시험에서 실행하지 않는다(시험 binary 자신을 교체하므로). 호출부(binary update 경로)는 04-B |
| — | — | worktree / cloud | 전체 | not_run | RC1 범위 밖 |
| Linux CI (ubuntu rc-ci) | — | local | AC-08C-7 재측정 | not_run | `sandbox-exec`는 macOS 전용 — Linux 대체 차단 수단 미결(HOLD, #2100 감사 지적) |
| — | GitLab | local | Forge adapter 계약 문서화 (REQ-B4-01) | not_run | **documented, not_run** — capability 선언만 존재(`src/forge/gitlab.rs`: merge_request·pipeline·note = documented, fixture 스키마 `schemas/gitlab-fixture.schema.json`), 어댑터·실호출 경로 없음(REQ-B4-01; 계정 없음 — decisions 2026-11-09 01:1x, #2171) |

## 집계

- 지원 행: 6 == passed receipt 6 (AC-08C-1·2·7·8, HA-02, AC-08C-5) — `tools/check_support_matrix.py`가 강제한다.
- not_run: 위 표 전부. failed/blocked: 0.
- 실측 시각: 2026-11-08T15:04:16Z(T1 suite — 08-C 수리 head `7cc351d1d` 병합 후 재실행), 2026-11-08T14:55:59Z(T2 스크립트). 실행 명령·환경은 각 receipt의 detail과 `docs/e2e-receipts/T1-cargo-test-output.txt`에 남긴다. T1 실행은 레인 워크트리 세션 가드상 HOME 래퍼 없이 실행됐다(fixture 내부 임시 HOME이 격리 경계이며 `e2e_never_writes_outside_temp_home`이 사용자 HOME 비의존을 증명) — 08-C 문서의 `HOME=$(mktemp -d)` 래퍼와의 차이는 receipt detail에 기록. receipt의 moai-cli SHA는 재실행 시마다 달라진다(빌드 바이너리) — 지원표는 최신 receipt 값을 따른다.

secret·개인 절대 경로·비공개 운영 정보: 없음(AC-08D-6 grep으로 검증).
