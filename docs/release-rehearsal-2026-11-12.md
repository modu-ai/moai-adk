# 릴리스 리허설 기록 — 2026-11-12 (Issue #2512)

- 실행 tree: `v2026.11.0` tip `bafc4a7111a33768a581d98d277fdb4540b95115` (worktree `issue/2512-release-rehearsal`, 2026-11-12T06:43Z 생성).
- 기준 점검표: #2492 가지 `b448d3b2fe1755765c9bcffa43ed3f0d53d8e0c3`의 `docs/release-checklist.md`(정정본). tip의 점검표는 그 앞 판이다.
- 범위: 게시·태그·저장소 설정 변경·실제 HOME 변경이 없는 단계만 실행했다. 태그 생성, GitHub Release, attestation·artifact 업로드, ruleset 적용, 실제 HOME 설치는 0건이다(HOLD, 사용자 승인). GitHub는 GET으로만 읽었다.
- 실행 환경: macOS arm64, 2026-11-12T06:4x–07:2xZ. HOME이 필요한 명령은 worktree의 ignored `target/` 아래 새 임시 HOME에서 실행했다. E1 블록은 점검표 그대로 `/tmp`에 새 빈 `CARGO_HOME`·`CARGO_TARGET_DIR`을 만들고, crates.io에서 도구를 설치했다(네트워크 사용).
- 상태 어휘: `passed`(실행, exit 0) · `checked`(문서·설정을 읽어 확인, 실행 없음) · `not_run`(실행하지 않음 — 사유 열).

## 단계별 영수증

| id | 점검표 위치 | 상태 | 명령 | exit | 관측 | 점검표와 |
| --- | --- | --- | --- | --- | --- | --- |
| a1 | (a) version SemVer prerelease | passed | `grep -m1 '^version' Cargo.toml`; `target/release/moai-cli --version` | 0 | `version = "2026.11.0-rc.1"`; `moai-cli 2026.11.0-rc.1` | 일치 |
| a2 | (a) tag 이름 `moai-cli-v<version>` | passed | `git ls-remote --tags origin 'moai-cli-v*' \| wc -l` | 0 | `0` — tag 없음 | 일치 |
| a3 | (a) tag ruleset 적용 | not_run | `gh api …/rulesets`, `…/rulesets/19583648 --jq .conditions.ref_name.include`(읽기만) | 0 | ruleset 2개(`26791615` branch, `19583648` tag), `["refs/tags/v*"]` | 일치 — 적용은 리더 몫(#2197, HOLD) |
| b1 | (b) exact candidate SHA | not_run | — | — | 후보 SHA는 게시 승인 때 정한다. 이 리허설의 tree는 위 tip | 일치 |
| b2 | (b) base SHA | passed | `git ls-remote origin refs/heads/v2026.11.0` | 0 | `bafc4a7111…`(worktree 생성 시) | 일치 |
| b3 | (b) 대조 절차(PR head == audited SHA) | not_run | — | — | 감사·PR 단계 절차. 리허설은 PR을 만들지 않는다 | 일치 |
| c1 | (c) `cargo test --offline --locked` | not_run | — | — | 리더 규칙: 필터 없는 전체 cargo test는 로컬에서 돌리지 않고 rc-ci를 인용한다. 참고: 이 tip을 포함한 PR #2510 head `8dcae619c3`(tip + 문서·시험)의 rc-ci run 43168523252가 4개 OS target에서 success | 수치 낡음(11-09 369 passed) |
| c2 | (c) `cargo clippy … -D warnings` | not_run | — | — | c1과 같은 사유(rc-ci `check` job 인용) | — |
| c3 | (c) `cargo fmt --check` | passed | `cargo fmt --check` | 0 | 출력 없음 | 일치 |
| c4 | (c) Python discover | passed | `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py'` | 0 | `Ran 394 tests in 769.421s`, `OK` | 수치 다름(11-09 Ran 80) |
| c5 | (c) `git diff --check`·`--cached --check` | passed | `git diff --check && git diff --cached --check` | 0 | 출력 없음 | 일치 |
| c6 | (c) `check_support_matrix` | passed | `python3 -I tools/check_support_matrix.py docs/support-matrix.md docs/e2e-receipts/` | 0 | `supported rows: 6`, `passed receipts: 6 (9 receipts total)` | 일치 |
| d1 | (d) 로컬 release 빌드 sha256 자체 검증(AC-08D-3) | passed | `cargo build --offline --locked --release`; `shasum -a 256 target/release/moai-cli > sha256.sum && shasum -a 256 -c sha256.sum` | 0 | 빌드 3m 28s; `OK`, `5736806815b2d20babe3abfc2ef77c699b9361bd91734c59f633321ab67764d3` | 값 다름(11-09 `312ccf8c…`) — tree가 바뀌었으므로 예상된 차이 |
| d2 | (d) cargo-dist `dist build`·archive별 `.sha256`·`sha256.sum` | passed | E1 블록(아래 e1) 안의 `dist build --artifacts=all`, `shasum -a 256 -c sha256.sum` | 0 | `moai-cli-aarch64-apple-darwin.tar.xz: OK`, `source.tar.gz: OK` | 일치 |
| d3 | (d) 재현 명령 고정(E1 #2193) | passed | E1 블록 전체(아래 e1) | 0 | 622s, 도구 3개 정확한 버전 설치 확인 | 일치(실행 위치 차이는 아래 차이 목록 5) |
| d4 | (d) release artifact checksum 목록 | not_run | — | — | 게시 단계 산출물. 로컬 `target/distrib/sha256.sum`은 생겼지만 게시 목록은 만들지 않았다 | 일치 |
| e1x | E1 로컬 전체 재현 블록 | passed | 점검표 `<!-- e1-local-repro -->` 블록을 그대로 `sh`로 실행 | 0 | `cargo-dist 0.34.0`, `cargo-cyclonedx 0.5.10`, `cargo-auditable v0.7.7`; `SBOM components: 76`; local 범위에서 `sha256.sum`·`source.tar.gz` 없음 확인; 마지막 `git status --short` 출력 없음 | 수치 다름(46 → 76, 차이 목록 2) |
| e1 | (e) attestation(`actions/attest@v4`, `gh attestation verify`) | not_run | — | — | 게시·workflow 실행 권한 필요(HOLD, 사용자 승인) | 일치 |
| e2 | (e) E4 #2196 생성 workflow | not_run | — | — | 설계·dry-run만 있고 생성은 승인 대기(HOLD) | 일치 |
| e3 | (e) REQ-08D-08 문구(checksum이 서명을 대체한다고 쓰지 않음) | checked | 점검표 (e)절과 `docs/release-attestation.md` 읽기 | — | 대체한다는 문장 없음 | 일치 |
| f1 | (f) known-good restore(08-C T1) | passed | `cargo test --offline --locked --test e2e_host_fixture e2e_setup_update_interrupt_restore_preserves_user_edits` | 0 | `1 passed` | 일치(receipt digest 낡음은 #2492 표 그대로) |
| g1 | (g) 게시 후 `/releases/latest` 재확인 | not_run | 게시 전 상태만 읽음: `gh api …/releases/latest`, `gh release view moai-cli-v2026.11.0-rc.1` | 0 | `v3.1.2 prerelease=false`, `release not found` | 일치 — 게시 후 측정은 not_run |
| g2 | (g) tag namespace 설계 정합(특성화 시험) | passed | `cargo test --offline --locked --test install_cli go_updater_feed_contamination_table`; `--test release_suite rust_final_relea` | 0 | 각 `1 passed`(`rust_final_release_hides_newer_go_release_from_stable_clients`) | 일치 |
| g3 | (g) `moai-cli-v*` 전용 tag ruleset | not_run | a3과 같음 | — | 리더 적용 대기(#2197, HOLD) | 일치 |
| h1 | (h) 사용자 문서의 3.x 필수 안내 0건 | passed | `git grep -nP 'moai (init\|update)\b' HEAD -- ':(glob)docs/*.md' README.md README.ko.md \| grep -v 'fallback\|3\.x'` | 0 | README·README.ko 0건 | 일치 |
| h2 | (h) docs 전역 기계 grep | passed | h1과 같은 명령 | 0 | docs 3건: `decisions-2026-11-08.md:11`·`:165`, `github-inventory-rc1.md:49`(모두 내부 장부) | 본문 (h) 둘째 줄의 "2건"과 다름(차이 목록 3) |
| h3 | (h) secret·개인 경로 0건 | passed | `python3 -m unittest tests.support_matrix_test tests.oss_docs_test` | 0 | `Ran 18 tests`, `OK` | 일치 |
| h4 | (h) `rc-ci.yml` | passed | YAML 파싱(`ruby -ryaml`) | 0 | trigger `pull_request`·`workflow_dispatch`, `permissions: {contents: read}`, jobs `check`·`rc-ci`·`windows-build` | 일치 |
| h5 | (h) ruleset 필수 체크 | passed | `gh api …/rulesets/26791615`(읽기만) | 0 | `["moai/audit"]` | 일치(HOLD 유지) |
| i1 | (i) SBOM 생성(cargo cyclonedx) | passed | E1 블록의 `cargo cyclonedx` + component 수 계산 | 0 | `target/distrib/moai-cli.cdx.xml` 86,181 bytes, components 76 | 수치 다름(46, 차이 목록 2) |
| i2 | (i) `dist-workspace.toml` 설정 | checked | `grep -n 'cargo-cyclonedx\|cargo-auditable\|targets' dist-workspace.toml` | 0 | `cargo-cyclonedx = true`, `cargo-auditable = true`, `targets = ["aarch64-apple-darwin"]` | 일치(주석의 결정 시각은 차이 목록 6) |
| i3 | (i) SBOM attestation 연동 | not_run | — | — | 게시 단계(HOLD) | 일치 |
| i4 | (i) license | checked | `grep -nE '^(license\|repository\|homepage\|description\|publish)' Cargo.toml` | 0 | `publish = false`만 있음 | 일치(메타데이터는 #2500 진행 중) |
| j1 | (j) `moai-cli-v*` tag ruleset 적용 | not_run | — | — | 리더(결정 A) — HOLD | 일치 |
| j2 | (j) attestation 생성 workflow | not_run | — | — | 운영자 — HOLD | 일치 |
| j3 | (j) 배포 산출물 OS/arch matrix·locale | not_run | — | — | 사용자 결정 대기. 이번 산출물은 `aarch64-apple-darwin` 하나 | 일치 |
| j4 | (j) SBOM | passed | i1과 같음 | 0 | components 76 | 일치(도입·실측됨) |
| j5 | (j) macOS 서명 | not_run | — | — | 결정: notarization 없음, quarantine 안내 문서. 서명 단계 자체가 없다 | 일치 |
| j6 | (j) T2 host 로딩 운영자 실측 | not_run | — | — | 운영자 자격증명 필요(HOLD) | 일치 |
| k1 | (k) 독립 감사 verdict(`docs/rc-release-review.md`) | not_run | — | — | 리허설은 감사가 아니다. verdict는 `release blocked` 그대로 | 일치 |
| p1 | 게시 절차 1 운영자 승인 | not_run | — | — | 사용자 승인(HOLD) | 일치 |
| p2 | 게시 절차 2 `git tag … --verify-tag` | not_run | — | — | 태그 생성 금지(HOLD) | 일치 |
| p3 | 게시 절차 3 `gh release create --prerelease --latest=false` + 자산 + `sha256.sum` + SBOM | not_run | — | — | 게시 금지(HOLD) | 일치 |
| p4 | 게시 절차 4 attestation 생성·검증 | not_run | — | — | e1과 같음 | 일치 |
| p5 | 게시 절차 5 `/releases/latest` 재확인 | not_run | — | — | g1과 같음 | 일치 |
| p6 | 게시 절차 6 not_run 항목 갱신 | not_run | — | — | 게시 뒤 단계 | 일치 |
| x1 | Issue #2512 범위: operations 문서 ↔ 코드 | passed | `cargo test --offline --locked --bin moai-cli operations_doc` | 0 | `3 passed`(`operations_doc_matches_code` 포함) | — |
| x2 | Issue #2512 범위: `install-inspect`(임시 HOME) | passed | `target/release/moai-cli install-inspect --json` | 0 | `writes: false`, `installations` 0, `self.binary_sha256`이 d1 값과 같음. PATH 조회는 읽기만 한다(PATH에 Go `moai`가 있어 `found: true`) | — |

## 점검표와 다른 곳(AC-3)

문서 수정은 #2492 병합 뒤 후속으로 한다. 이 목록은 보고만 한다.

1. d1 로컬 release 빌드 sha256: 점검표 `312ccf8c…`(11-09) → 지금 `5736806815…`. tree가 바뀌었으므로 예상된 차이다. 점검표가 값을 "그 날 기록"으로 표시하는지 확인이 필요하다.
2. SBOM component 수: 점검표 46(11-09 tree `b35b75eb3`) → 지금 76. #2492가 "재생성 전까지 not_run"으로 둔 값을 이번에 다시 쟀다.
3. (h) 본문 둘째 줄은 docs grep "2건"이다. #2492 리허설 표의 '현행' 열은 3건으로 고쳤지만, 본문 줄은 그대로다.
4. (c) 수치: Python `Ran 80`(11-09) → `Ran 394`. cargo test 수치는 로컬에서 재지 않았다.
5. E1 블록은 "깨끗한 임시 clone 루트"에서 실행하라고 하지만, 리더 규칙(새 clone 금지)에 따라 깨끗한 worktree(`target/distrib` 없음, 추적 변경 0)에서 실행했다. `CARGO_HOME`·`CARGO_TARGET_DIR`은 블록 그대로 새 빈 `/tmp` 폴더였다.
6. `dist-workspace.toml` 주석은 아직 "decision 01:1x"를 가리킨다(#2492 S2, 코드 소유자 몫).
7. E1 블록은 `/tmp/moai-e1-cargo.*`(651M)·`/tmp/moai-e1-target.*`(1.3G)를 남긴다. 블록에 정리 단계가 없으며, 이 리허설은 지우지 않았다(삭제 금지).

## 재현

worktree의 ignored `target/agent26/rehearse.py`가 위 명령을 단계 id 순서로 실행하고, 각 단계의 명령·exit·소요 시간·출력 끝을 `target/agent26/receipts.jsonl`에 남겼다(로컬 증거, 추적 안 함).
