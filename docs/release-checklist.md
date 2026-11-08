# RC1 릴리스 체크리스트 (08-D #2015 — **게시 전 상태**)

이 문서는 릴리스 승격 판단의 증거 목록이다. **GitHub Release 게시·tag push·서명 identity 운영은 이 체크리스트의 효과가 아니며 별도 사용자 권한·결정 단계다.** 필수 항목이 하나라도 `failed`·`blocked`·`not_run`이면 verdict는 `release blocked`다(REQ-08D-04 — 평균 점수로 상쇄하지 않는다). 현재 verdict: **`release blocked`**(아래 not_run/HOLD 항목 존재).

## (a) version · SemVer prerelease

- [x] version `2026.11.0-rc.1`은 SemVer 2.0.0 prerelease(`-rc.1` 하이픈 식별자)이며 `2026.11.0`보다 precedence가 낮다(semver.org, 2026-11-08 검증). 문서·tag·Release는 prerelease로 취급한다.
- [x] tag는 `moai-cli-v2026.11.0-rc.1`(04-A 결정 B: 같은 repo, `moai-cli-v<version>` 접두사 — Go stable updater의 `/releases/latest`와 Go RC 채널의 `v`/`go-v` 접두사 필터 어느 쪽에도 걸리지 않는다).
- [ ] **HOLD/not_run — tag ruleset**: 기존 ruleset `Release tag immutability (v*)`(id 19583648)은 `moai-cli-v*`를 덮지 않는다 → 같은 version 재배포 금지(REQ-08D-09)의 ruleset 근거가 없다. `moai-cli-v*` 전용 ruleset 추가는 저장소 설정 변경 = **운영자 결정(HOLD)**. 결정 전까지 REQ-08D-09는 절차상 규약으로만 성립한다.

## (b) exact candidate SHA · base SHA

- [x] 내용 커밋(이 체크리스트가 기록하는 exact candidate content SHA): **`c29451fc08f6c5976bdb589d75567c576b9eb76f`** (8-D 산출물 전체를 담은 커밋). PR head는 이 위의 최종 커밋이며 감사는 PR head == audited SHA로 대조한다(PR 본문 head 필드 참조).
- [x] base: `v2026.11.0` tip (PR 개설 직전 재조회 — PR 본문 base 필드) · 08-C 수리 head `7cc351d1d` 병합 포함(merge `4f9631483`)
- [x] 대조 절차: 감사는 PR head == audited SHA, merge-base == base tip을 확인한다(통합 게이트). 본 체크리스트의 candidate SHA와 감사 verdict SHA가 다르면 재감사다.

## (c) 로컬 build/test (명령 · exit code)

- [x] `CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-target-agent-7 cargo test --offline --locked` — exit 0
- [x] `CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-target-agent-7 cargo clippy --offline --locked --all-targets -- -D warnings` — exit 0
- [x] `cargo fmt --check` — exit 0
- [x] `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py'` — exit 0
- [x] `git diff --check` · `git diff --cached --check` — exit 0
- [x] `python3 -I tools/check_support_matrix.py docs/support-matrix.md docs/e2e-receipts/` — exit 0 (지원 행 6 == passed receipt 6, AC-08D-1)

## (d) artifact sha256 (AC-08D-3 — 로컬 빌드 산출물로 실제 실행)

- [x] 명령(본문 AC-08D-3 그대로, 세션 규칙상 target dir만 유지): `CARGO_HOME=/tmp/moai-rc-product-cargo CARGO_TARGET_DIR=/tmp/moai-rc-target-agent-7 cargo build --offline --locked --release && shasum -a 256 /tmp/moai-rc-target-agent-7/release/moai-cli > /tmp/sha256.sum && shasum -a 256 -c /tmp/sha256.sum`
- [x] 결과: **OK**(exit 0) — `312ccf8c14af41de648c5a7af4821296ea677f43d11a76415948f3ea84b16b2b  /tmp/moai-rc-target-agent-7/release/moai-cli`. 참고: 이것은 **로컬 개발 빌드 산출물의 자체 검증**이며, 실제 release artifact(tarball·installer)의 checksum 목록은 게시 단계에서 만든다(cargo-dist 미도입 — 수동 `sha256.sum` 절차를 이 항목이 선행한다).
- [ ] **not_run — release artifact checksum 목록**: artifact 미생성(게시 금지 범위).

## (e) provenance / 서명 (AC-08D-5)

- [ ] **HOLD/not_run — attestation**: 방식은 결정됨(GitHub artifact attestation 단독, `actions/attest@v4` + `gh attestation verify <artifact> -R modu-ai/moai-adk`)으나 **생성 실측 없음**(게시·workflow 실행 미허용). `gh attestation verify`는 not_run이다.
- [x] REQ-08D-08 준수: **checksum이 서명 검증을 대체한다고 쓰지 않았다.** checksum은 무결성 재확인이며 provenance(빌드 출처 연결)는 attestation 계열만 담당한다(in-toto/Reproducible Builds 근거, Issue 본문).

## (f) known-good restore

- [x] 08-C T1 `e2e_setup_update_interrupt_restore_preserves_user_edits` passed — receipt `docs/e2e-receipts/receipts.jsonl`(AC-08C-1, 2026-11-08T15:04:16Z, 08-C 수리 head `7cc351d1d` 병합 후 재실행): 트랜잭션 중단 → recovery가 known-good 대체, 사용자 편집 바이트 보존, 성공 host committed 유지.

## (g) 3.x updater 격리 (AC-08D-4 — Release 미게시 → not_run)

- [ ] **not_run — `/releases/latest` 확인(게시 후 재실행)**: 실측(2026-11-09): `gh api …/releases/latest` → `v3.1.2`(Go stable — RC 부재 확인), `gh release view moai-cli-v2026.11.0-rc.1` → **release not found**(미게시 확인). 게시 단계에서: RC는 `--prerelease --latest=false --verify-tag`로 올려 Go stable updater(`non-prerelease, non-draft latest`)가 접근하지 못함을 재확인한다.
- [x] 설계 정합: tag namespace `moai-cli-v*`는 Go checker의 `v`/`go-v` 접두사 필터와 `/releases/latest` 어느 쪽에도 노출되지 않는다(04-A 특성화 시험 `go_updater_feed_contamination_table` — `tests/update_cli.rs`… 04-B 이후 `tests/install_cli.rs`가 계승, 본 repo 기록). **다만 실측은 같은 repo 게시 후에만 성립** — 게시 전까지는 설계 근거 상태다.
- [ ] **HOLD — `moai-cli-v*` 전용 tag ruleset**: (a) 참조. 사용자 결정 대상.

## (h) 문서 예제 재현 (AC-08D-8)

- [x] 사용자 문서(README·user-entry-draft): 3.x 명령을 차세대 필수 단계로 안내하는 문장 **0건**. `/moai`·`$moai` 자연어 진입이 기본, 수동 명령은 명시적 fallback(`docs/user-entry-draft.md`)으로 분리.
- [x] docs 전역 기계 grep(`grep -nE 'moai (init|update)\b' docs/*.md README.md | grep -v 'fallback\|3\.x'`): 2건 — `docs/github-inventory-rc1.md:49`(forge 제목-매칭 규칙의 예시 문장)과 `docs/decisions-2026-11-08.md:11`(결정 장부의 요구 기록, "설계 반영만, 지금 실행 없음"). 둘 다 내부 장부·기술 문서로 사용자 안내가 아니며 REQ-08D-06 위반 아님 — 기계 grep 결과를 있는 그대로 기록한다.
- [x] 지원표·체크리스트에 secret·개인 절대 경로 0건 (AC-08D-6 grep — 대상 3문서 0건; 일치는 시험 파일 자체의 패턴 리터럴뿐).
- [x] `.github/workflows/rc-ci.yml` yaml 파싱 exit 0, `pull_request` trigger 존재, `contents: read`, cargo test/clippy/fmt·Python unittest step 확인 — **파일 수정 불요**(00-E 작성본이 이미 08-A 명세 항목을 충족; 근거: 위 확인 출력). ruleset에 `rc-ci` 필수 체크 등록은 저장소 설정 = 별도 결정(HOLD).

## (i) SBOM / license

- [ ] **HOLD/not_run — CycloneDX `bom.xml`·cargo-auditable**: 도구 설치·실행이 필요해 RC1 보류(선택지 B: Cargo.lock — 의존 4개 — 만 게시 단계에서 첨부하는 안, 사용자 결정 대기). 정식 릴리스 재검토.

## (j) 미결 목록

| 미결 | 성격 | 릴리스에 끼치는 것 |
| --- | --- | --- |
| `moai-cli-v*` tag ruleset 추가 여부 | 저장소 설정(운영자) | REQ-08D-09 근거 — 미결 시 재배포 금지가 절차 규약으로만 성립 |
| attestation 생성 workflow(`actions/attest@v4`) | release 운영(운영자) | AC-08D-5 not_run — provenance 부재 상태로는 게시 불가 |
| P12 OS/arch matrix 확대 | 사용자 결정 | 지원표가 실측 단일 조합으로 좁게 유지됨 |
| SBOM(cargo-cyclonedx) | 사용자 결정 | (i) not_run |
| macOS codesign/notarization | 운영자(Developer ID) | 미서명 표기 또는 수동 절차(HOLD) |
| T2 호스트 로딩 운영자 실측 | 운영자(자격증명) | certified 지원 행 0 — 지원표 머리의 한계 문단 유지 |

## (k) 독립 감사 verdict

- [ ] `docs/rc-release-review.md` — 독립 reviewer가 reviewer identity·공유 context 범위·대조 digest·gap 표와 함께 기록한다. 필수 gap이 있으면 verdict는 `release blocked`(현재 예상 verdict: **release blocked** — (d) artifact 목록·(e) attestation·(g) 게시 실측이 not_run).

## 릴리스 게시 절차 (별도 권한 단계 — 이 체크리스트가 수행하지 않는다)

1. 운영자가 위 HOLD/not_run을 해소·승인한다. 2. `git tag moai-cli-v2026.11.0-rc.1 <candidate>` + `--verify-tag`. 3. `gh release create --prerelease --latest=false --verify-tag` + 자산 + `sha256.sum`. 4. attestation 생성 후 `gh attestation verify`. 5. `/releases/latest`가 RC tag을 반환하지 않음을 재확인. 6. 이 체크리스트의 not_run 항목을 실측값으로 갱신한다. 같은 version 재배포 금지 — 교정은 `rc.2`.
