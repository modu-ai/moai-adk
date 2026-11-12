# 지원표 (08-D #2015, RC1 — 게시 전 상태)

이 표의 유일한 입력은 08-C E2E receipt(`docs/e2e-receipts/receipts.jsonl`, `schemas/e2e-receipt.schema.json`)다. RL-10: receipt가 없는 조합은 지원으로 표기하지 않고, 지원 행 수 == passed receipt 수. 검증은 `python3 -I tools/check_support_matrix.py docs/support-matrix.md docs/e2e-receipts/`(exit 0 = 일치).

**상태 한계(숨기지 않음)**: 이 표의 passed 행은 T1 결정적 fixture(호스트 `none`) 5건과 T2 정적 검증 1건이다. 실제 호스트 로딩(certified) — Claude/Codex가 `/moai`·`$moai`로 skill을 로딩하는 E2E — 는 운영자 실측(T2, 자격증명 필요) 전까지 전부 `not_run`이며, "지원" 행이 곧 실사용 보증이 아님을 이 문서 머리에 명시한다. RC1은 게시 전 단계다.

## 지원 대상 조합 (P04)

rc.1의 지원 대상은 **Darwin/arm64(`aarch64-apple-darwin`) × Claude Code, Darwin/arm64 × Codex** 두 조합뿐이다(사용자 결정 P04, #2262). 아래 표에서 "지원 대상"은 범위를 정한 것이지 지원 보증이 아니다. 보증은 passed receipt가 있는 행만 하고(RL-10), 두 조합의 host 로딩 E2E(T2)는 receipt가 오기 전까지 `not_run`이다.

| 조합 | rc.1 범위 | 현재 결과 |
| --- | --- | --- |
| Darwin/arm64 × Claude Code | 지원 대상 | T2 not_run(AC-08C-3·6·8 절반), 정적 검증 AC-08C-5만 passed |
| Darwin/arm64 × Codex | 지원 대상 | T2 not_run(AC-08C-4) |
| Darwin/x86_64(`x86_64-apple-darwin`, Tier 2) | 미지원(후속) | host 측정 없음. CI check set만 측정(아래 "CI OS matrix") |
| Linux x86_64·arm64(ubuntu-24.04, ubuntu-24.04-arm) | 미지원(후속) | host 측정 없음. CI check set만 측정(아래 "CI OS matrix") |
| Windows(windows-2025) | 빌드 + self-replace 구현만(결정 01:1x(11-09), E3 #2195) | **not_run** — CI `windows-build` job은 `--all-targets`로 라이브러리·`moai-cli` binary·portable 시험 target을 컴파일한다. 시험 실행·실행 중 교체·host 로딩은 측정하지 않음 |
| ZCode·그 밖의 host | 미지원(후속) | 측정하지 않음 |

## 조합 범위 (P12)

RC1의 host 실측 조합은 개발 Mac 단일이다: Darwin / arm64 (Rust triple `aarch64-apple-darwin`, Tier 1). 나머지 OS·arch·mode의 host 항목은 전부 `not_run`이다(숨김 없음). rc.2 E2(#2194)부터 `rc-ci`가 결정 01:1x(11-09)의 네 target(Linux x86_64·arm64, macOS arm64·x86_64)에서 check set을 돌린다. 사용자 결정 U10(#2554)부터 PR 에서는 linux-x86_64 만 돌고, 네 target 과 `windows-build` 는 병합하는 push(`v2026.11.0`) 실행에서 같은 check set 으로 돈다. 그 결과는 아래 "CI OS matrix" 절에 따로 적는다. CI check set 통과는 host 지원 판정(receipt)이 아니라서 위 지원 행에 더하지 않는다.

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
| ubuntu-24.04 / arm 변형 / Darwin x86_64 | — | local | host 항목 전체 | not_run | 미지원(후속) — P04로 rc.1 범위 밖. host E2E receipt 없음(CI check set은 아래 "CI OS matrix" 절, host 판정 아님) |
| windows-2025 | — | local | 빌드(`cargo build --offline --locked --all-targets`, 라이브러리·binary·portable 시험 target 컴파일) | 병합 뒤 push 실행의 `windows-build` 참조(#2554 이전에는 PR 실행) | E3 #2195 — 컴파일 증거일 뿐 시험 실행이나 host 지원 보증이 아니다 |
| windows-2025 | — | local | 첫 시험 target 컴파일(`--all-targets`) | failed(run `42707532063`, 수리 전 이력) | 당시 `src/agents.rs:1127·1166·1180`, `src/claude.rs:1149`의 Unix 전용 코드가 조건부 컴파일되지 않아 실패했다. 수리에서는 권한 시험과 symlink 구간만 `cfg(unix)`로 제한하고 `--all-targets` 범위를 유지한다. 수리 후보의 결과는 PR의 최신 run으로 확인한다 |
| windows-2025 | — | local | 추가 integration 시험 컴파일(`--all-targets`) | failed(run `42716692062`, head `1a29cfb4d`, 수리 중 이력) | library 수리 뒤 `tests/apply_receipt.rs`의 Unix chmod 시험에서 실패했다. 권한 시험만 Unix로 제한하고, context·baseline·tree digest·identity의 혼합 시험은 Unix symlink 구간만 제한하여 portable 검사를 유지한다. 최종 Windows 빌드 결과는 PR의 후속 run으로 확인한다 |
| windows-2025 | — | local | 시험 실행·실행 중 binary 교체(self-replace)·host 로딩 | not_run | 결정 01:1x(11-09): Windows는 빌드 + self-replace 구현만. 교체 경로는 `src/self_update.rs`의 모의 replacer 단위 시험(macOS)으로만 확인했고, 실제 `self_replace` 호출은 시험에서 실행하지 않는다(시험 binary 자신을 교체하므로). 호출부(binary update 경로)는 04-B |
| — | — | worktree / cloud | 전체 | not_run | RC1 범위 밖 |
| Linux CI (ubuntu rc-ci) | — | local | AC-08C-7 재측정 | not_run | `sandbox-exec`는 macOS 전용 — Linux 대체 차단 수단 미결(HOLD, #2100 감사 지적) |
| — | GitLab | local | Forge adapter 계약 문서화 (REQ-B4-01) | not_run | **documented, not_run** — capability 선언만 존재(`src/forge/gitlab.rs`: merge_request·pipeline·note = documented, fixture 스키마 `schemas/gitlab-fixture.schema.json`), 어댑터·실호출 경로 없음(REQ-B4-01; 계정 없음 — decisions 2026-11-09 01:1x, #2171) |

## CI OS matrix (rc.2 E2 #2194) — check set 실측, host 지원 판정 아님

입력은 GitHub Actions `rc-ci` run의 job 결과와 job log뿐이다. 이 절의 행은 host E2E receipt가 아니므로 지원 행 수(RL-10)에 들어가지 않는다. 결론 칸은 GitHub job 결론(`success`/`failure`)을 그대로 적는다.

run `42619829090`(PR #2309, `pull_request`, head `15d8178bc2d9a8dd1ba16bfaae54246bd1ecb099`, 2026-11-09). 모든 runner는 공개 저장소용 표준 GitHub-hosted runner(무료)다.

| target | runner label | Rust host(`rustup show`) | cargo test(`--offline --locked`) | python unittest | macOS 전용 Rust 시험 2개 | darwin 한정 Python 시험 7개 | job 결론 | job id | 대기 / 실행 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Linux x86_64 | `ubuntu-24.04` | `x86_64-unknown-linux-gnu` | 518 / 0 실패 / 0 ignored(59 binary) | Ran 118, OK(skipped=7) | 컴파일 안 됨(`cfg(target_os = "macos")`) | 0 실행 · 7 skip | success | 129168786780 | 3s / 4m05s |
| Linux arm64 | `ubuntu-24.04-arm` | `aarch64-unknown-linux-gnu` | 518 / 0 실패 / 0 ignored(59 binary) | Ran 118, OK(skipped=7) | 컴파일 안 됨 | 0 실행 · 7 skip | success | 129168786544 | 5s / 7m42s |
| macOS arm64 | `macos-15` | `aarch64-apple-darwin` | 517 / 0 실패 / 0 ignored(59 binary) | Ran 118, OK(skipped=1) | 2 실행 ok(`e2e_never_writes_outside_temp_home`, `privacy_06_operations_need_no_network`) | 6 실행 · 1 skip | success | 129168786250 | 2s / 3m51s |
| macOS x86_64 | `macos-15-intel` | `x86_64-apple-darwin` | 517 / 0 실패 / 0 ignored(59 binary) | Ran 118, OK(skipped=1) | 2 실행 ok | 6 실행 · 1 skip | success | 129168786658 | 4s / 7m12s |

- cargo 시험 수 차이: 공통 515개에 Linux는 `tests/script_write_race.rs`(`#![cfg(target_os = "linux")]`, 3개)를 더하고, macOS는 위 macOS 전용 2개를 더한다.
- macOS에서 skip된 Python 1개는 `claude` 바이너리가 필요한 시험이다(hosted runner에 없음). Claude·Codex 실세션은 어느 target에서도 돌리지 않았다.
- AC-08C-7(사용자 HOME 쓰기 차단)은 macOS 두 target의 `e2e_never_writes_outside_temp_home`이 hosted runner에서도 ok였다. Linux에는 대체 차단 수단이 없어 위 not_run 행이 그대로 유효하다.
- 첫 run `42618040618`(head `f557b035a`)은 macOS 두 target만 success였다. Linux 두 target은 check set은 초록이었는데, workflow의 macOS 전용 요약 블록 결함(grep 일치 없음 + `pipefail`)으로 job이 failure가 됐다. 그 결함을 고친 뒤 위 run에서 측정했다.
- windows-2025는 위 matrix에 없다. 별도 job `windows-build`(E3 #2195)가 `--all-targets`로 라이브러리·binary·portable 시험 target을 컴파일하되 시험은 실행하지 않는다. Unix 권한 시험과 Unix symlink 구간만 조건부 컴파일한다. `rc-ci` 집계(`needs: [check]`)에 들어가지 않으므로 `rc-ci` 결론에 영향이 없다. 결과는 위 not_run 절의 windows-2025 행에 적는다.

## RELEASE 시험군 (rc.2 E6 #2198) — fixture 결과, host 지원 판정 아님

입력은 `tests/release_suite.rs`의 시험 4개뿐이다. 가짜 release feed, 문서화된 GitHub 규칙의 모델, 임시 HOME만 쓴다. 네트워크 호출, 실제 릴리스·tag, 실제 HOME은 없다. 이 절의 행은 host E2E receipt가 아니므로 지원 행 수(RL-10)에 들어가지 않는다.

- Go 3.x 업데이터 규칙은 `origin/main` `4fa2aa9db`의 `internal/update/checker.go`·`internal/cli/deps.go`를 다시 읽어 옮겼다(시험 쪽 모델이며 `src/`에는 들어가지 않는다).
- `/releases/latest`와 `make_latest` 규칙은 GitHub REST "Releases" 문서, tag 트리거 규칙은 workflow 문법 "Filter pattern cheat sheet"를 따른다(둘 다 2026-11-09에 읽음).

### feed 오염 표 (REQ-E6-01)

기록된 실제 feed(`tests/fixtures/go-releases-sample.json`) 위에 후보 하나를 게시한 결과다. Go 안정 클라이언트는 `v3.1.2`(`/releases/latest`), Go RC 클라이언트는 `v3.1.2-rc.1`(`/releases` 목록, 익명)이다. 후보 자산은 cargo-dist 산출물(`sha256.sum`·`.sha256`, `checksums.txt` 없음)이고, 표시한 행만 GoReleaser 형식(`checksums.txt` 있음)이다.

| 후보 tag | prerelease | `checksums.txt` | make_latest | Go 안정 결과 | Go RC 결과 |
| --- | --- | --- | --- | --- | --- |
| `moai-cli-v2026.11.0-rc.1` | 예 | 없음 | 기본(true) | 업데이트 없음 | 업데이트 없음 |
| `v2026.11.0-rc.1` | 예 | 없음 | 기본(true) | 업데이트 없음 | **검사 오류**(checksum_unavailable) |
| `v2026.11.0-rc.1` | 예 | 있음 | 기본(true) | 업데이트 없음 | **2026.11.0 업데이트로 읽음** |
| `v2026.11.0` | 아니오 | 없음 | 기본(true) | **검사 오류** | **검사 오류** |
| `moai-cli-v2026.11.0` | 아니오 | 없음 | 기본(true) | **검사 오류**(latest를 가져감) | 업데이트 없음 |
| `moai-cli-v2026.11.0` | 아니오 | 없음 | false | 업데이트 없음 | 업데이트 없음 |
| `go-v3.1.3-rc.1`(대조) | 예 | 있음 | 기본(true) | 업데이트 없음 | `go-v3.1.3-rc.1` 업데이트 |

- 평문 `v*` tag는 Go RC 채널 필터(`go-v`·`v` 접두사, 첫 항목)를 통과한다. Go 응답 구조체에는 draft·prerelease 필드가 없어서, 토큰이 있는 RC 클라이언트에는 기록된 feed 맨 위의 draft `v3.1.3`도 업데이트로 보인다.
- `moai-cli-v*` namespace는 RC 목록에서는 걸러지지만 `/releases/latest`는 막지 못한다. 더 새 Go 릴리스 `v3.2.0`이 있어도 dist 기본값으로 게시한 정식판이 latest를 가져가면 안정 클라이언트는 오류로 끝나고, `checksums.txt`가 있어도 tag가 `0.11.0`으로 읽혀 `v3.2.0`을 보지 못한다. `make_latest=false`이면 `v3.2.0`이 그대로 보인다(`docs/release-attestation.md` 운영자 승인 항목 6).

### release workflow 트리거 (모델, workflow 실행 아님)

| tag | dist 기본 `'**[0-9]+.[0-9]+.[0-9]+*'` | `tag-namespace = "moai-cli"` `'moai-cli**[0-9]+.[0-9]+.[0-9]+*'` |
| --- | --- | --- |
| `v3.1.3` | 걸림 | 안 걸림 |
| `go-v3.2.1-rc.2` | 걸림 | 안 걸림 |
| `v2026.11.0-rc.1` | 걸림 | 안 걸림 |
| `moai-cli-v2026.11.0-rc.1` | 걸림 | 걸림 |
| `moai-cli-v2026.11.0` | 걸림 | 걸림 |
| `x-moai-cli-v2026.11.0` | 걸림 | 안 걸림 |
| `moai-cli-nightly` | 안 걸림 | 안 걸림 |

### update 중단 복구 (known-good)

실제 v1→v2 update(파일 3개)를 끝낸 뒤 codex host를 중단 지점 k로 되돌려 `recovery-plan`·`recovery-apply`를 돌린 결과다. 복구에는 실제 v2 apply가 저장한 known-good 사본을 쓴다.

| 중단 지점 k | 게시된 entry | 분류(SKILL.md · new.md · update.md) | 복구된 파일 | codex 트리 = v1 digest | claude host · 가짜 Go 상태 |
| --- | --- | --- | --- | --- | --- |
| 0 | 없음 | before · before · before | 없음 | 예 | 그대로 |
| 1 | SKILL.md | after · before · before | SKILL.md | 예 | 그대로 |
| 2 | + workflows/new.md | after · after · before | SKILL.md, workflows/new.md | 예 | 그대로 |
| 3(receipt commit 전) | + workflows/update.md | after · after · after | 세 파일 모두 | 예 | 그대로 |

- 마지막으로 게시된 entry는 `published: false`로 남긴다(rename과 receipt 저장 사이의 중단). 분류는 이 표시가 아니라 현재 바이트로 정해진다.
- 설계 §17 RELEASE의 attestation 실측과 Issues end-to-end 실제 완주는 이 절에서 돌리지 않았다(not_run). attestation은 `docs/release-attestation.md`의 운영자 승인 항목을 따른다.

## 집계

- 지원 행: 6 == passed receipt 6 (AC-08C-1·2·7·8, HA-02, AC-08C-5) — `tools/check_support_matrix.py`가 강제한다.
- not_run: 위 표 전부. failed/blocked: 0.
- 실측 시각: 2026-11-08T15:04:16Z(T1 suite — 08-C 수리 head `7cc351d1d` 병합 후 재실행), 2026-11-08T14:55:59Z(T2 스크립트). 실행 명령·환경은 각 receipt의 detail과 `docs/e2e-receipts/T1-cargo-test-output.txt`에 남긴다. T1 실행은 레인 워크트리 세션 가드상 HOME 래퍼 없이 실행됐다(fixture 내부 임시 HOME이 격리 경계이며 `e2e_never_writes_outside_temp_home`이 사용자 HOME 비의존을 증명) — 08-C 문서의 `HOME=$(mktemp -d)` 래퍼와의 차이는 receipt detail에 기록. receipt의 moai-cli SHA는 재실행 시마다 달라진다(빌드 바이너리) — 지원표는 최신 receipt 값을 따른다.

secret·개인 절대 경로·비공개 운영 정보: 없음(AC-08D-6 grep으로 검증).
