# RC1 릴리스 체크리스트 (08-D #2015 — **게시 전 상태**)

이 문서는 릴리스 승격 판단의 증거 목록이다. **GitHub Release 게시·tag push·서명 identity 운영은 이 체크리스트의 효과가 아니며 별도 사용자 권한·결정 단계다.** 필수 항목이 하나라도 `failed`·`blocked`·`not_run`이면 verdict는 `release blocked`다(REQ-08D-04 — 평균 점수로 상쇄하지 않는다). 현재 verdict: **`release blocked`**(아래 not_run/HOLD 항목 존재).

## (a) version · SemVer prerelease

- [x] version `2026.11.0-rc.1`은 SemVer 2.0.0 prerelease(`-rc.1` 하이픈 식별자)이며 `2026.11.0`보다 precedence가 낮다(semver.org, 2026-11-08 검증). 문서·tag·Release는 prerelease로 취급한다.
- [x] tag는 `moai-cli-v2026.11.0-rc.1`(04-A 결정 B: 같은 repo, `moai-cli-v<version>` 접두사). Go RC 채널의 `v`/`go-v` 접두사 필터에는 걸리지 않는다. rc.N은 prerelease라 GitHub가 latest로 지정하지 않으므로 Go stable updater의 `/releases/latest`에도 나오지 않는다. **정식판은 다르다**: `/releases/latest`는 접두사로 거르지 않으므로, 접미사 없는 `moai-cli-v<version>`을 `make_latest` 기본값(true)으로 게시하면 그 릴리스가 latest가 되어 Go stable updater가 실패한다. 정식판은 `--latest=false`(또는 `docs/release-attestation.md` 운영자 승인 항목 6의 해결안)로만 게시한다. 근거: `docs/release-attestation.md` "prerelease와 정식판의 latest"·승인 항목 6, `tests/release_suite.rs` `go_feed_contamination_table_reproduces_plain_v_risk`·`rust_final_release_hides_newer_go_release_from_stable_clients`(E6 #2198).
- [ ] **not_run — tag ruleset 적용 대기**: 결정됨(01:1x(11-09) 생성 카드, 17:52Z(11-10) 표 37–41 A: PR #2375 병합 뒤 리더 적용·재확인). 기존 `Release tag immutability (v*)`(19583648, `refs/tags/v*`)는 `moai-cli-v*`를 덮지 않는다(2026-11-12T03:59Z `gh api …/rulesets/19583648 --jq .conditions.ref_name.include` → `["refs/tags/v*"]`). 리더 적용과 readback(#2197 AC-2) 전까지 REQ-08D-09는 절차상 규약으로만 성립한다.

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
- [x] 결과: **OK**(exit 0) — `312ccf8c14af41de648c5a7af4821296ea677f43d11a76415948f3ea84b16b2b  /tmp/moai-rc-target-agent-7/release/moai-cli`. 참고: 이것은 **로컬 개발 빌드 산출물의 자체 검증**이며, 실제 release artifact(tarball·installer)의 checksum 목록은 게시 단계에서 만든다(당시 cargo-dist 미도입 — 이후 E1 #2193에서 도입, 아래 줄 참조).
- [x] **E1 #2193 갱신 — cargo-dist 도입**: `dist-workspace.toml` 설정 + `dist build`(cargo-dist 0.34.0)로 archive별 `.sha256` + 통합 `sha256.sum` 네이티브 산출 실측(target/distrib — `shasum -a 256 -c` OK). 게시 단계 checksum 목록은 이제 dist가 산출한다.
- [x] **재현 명령 고정 (E1 #2193 수리, 사용자 결정 2026-11-09 07:0x — REQ-E1-01 재정의)**: 새 빈 세션 CARGO_HOME에 `cargo install --locked cargo-dist@0.34.0 cargo-cyclonedx@0.5.10 cargo-auditable@0.7.7`로 세 도구를 모두 설치한다. `dist-workspace.toml`의 `cargo-auditable = true`를 유지하므로 cargo-auditable도 필수다. 기존 두 도구만 설치한 절차는 빈 CARGO_HOME에서 `dist build --artifacts=all` exit 255로 실패했다(감사 #2193 댓글 6738998094). 아래 로컬 재현 블록이 설치부터 검증까지의 전체 명령이다. **산출 범위(실측 2026-11-09T07:1xZ, tree `e41bb62e7`)**: `--artifacts=all`은 `target/distrib/`에 archive 바이너리 tarball + archive별 `.sha256` + **통합 `sha256.sum`** + `source.tar.gz`(+`.sha256`)을 낸다. `--artifacts=local`(빈 distrib 재현)은 로컬 조합 archive + archive별 `.sha256`**만** 낸다 — `sha256.sum`과 source tarball은 제외되므로 게시 단계 checksum 목록에는 `--artifacts=all`이 필요하다. 체크섬의 상대 파일명은 `target/distrib/`을 기준으로 검증해야 한다. 두 항목 `OK` + `WARNING: 1 line is improperly formatted`는 파일 끝 빈 줄에 따른 기존 관측이다.
- [ ] **not_run — release artifact checksum 목록**: artifact 미생성(게시 금지 범위).

## (e) provenance / 서명 (AC-08D-5)

- [ ] **HOLD/not_run — attestation**: 방식은 결정됨(GitHub artifact attestation 단독, `actions/attest@v4` + `gh attestation verify <artifact> -R modu-ai/moai-adk`)으나 **생성 실측 없음**(게시·workflow 실행 미허용). `gh attestation verify`는 not_run이다.
- [ ] **E4 #2196 — 설계·dry-run 완료, 생성은 승인 대기**: `docs/release-attestation.md`. cargo-dist 생성 workflow(`tag-namespace = "moai-cli"`, attestation과 GitHub Release를 host 단계에 배치)와 사용자 확인 절차(`gh attestation verify -R modu-ai/moai-adk --signer-workflow …`), macOS quarantine 안내를 담았다. `gh attestation verify` 절차는 공개 저장소(cli/cli v2.92.0)의 실제 attestation으로 성공·실패 경우를 실측했다. 이 저장소의 attestation 생성은 운영자 승인 항목이다(그 문서 6절).
- [x] REQ-08D-08 준수: **checksum이 서명 검증을 대체한다고 쓰지 않았다.** checksum은 무결성 재확인이며 provenance(빌드 출처 연결)는 attestation 계열만 담당한다(in-toto/Reproducible Builds 근거, Issue 본문).

## (f) known-good restore

- [x] 08-C T1 `e2e_setup_update_interrupt_restore_preserves_user_edits` passed — receipt `docs/e2e-receipts/receipts.jsonl`(AC-08C-1, 2026-11-08T15:04:16Z, 08-C 수리 head `7cc351d1d` 병합 후 재실행): 트랜잭션 중단 → recovery가 known-good 대체, 사용자 편집 바이트 보존, 성공 host committed 유지.

## (g) 3.x updater 격리 (AC-08D-4 — Release 미게시 → not_run)

- [ ] **not_run — `/releases/latest` 확인(게시 후 재실행)**: 실측(2026-11-09): `gh api …/releases/latest` → `v3.1.2`(Go stable — RC 부재 확인), `gh release view moai-cli-v2026.11.0-rc.1` → **release not found**(미게시 확인). 게시 단계에서: RC는 `--prerelease --latest=false --verify-tag`로 올려 Go stable updater(`non-prerelease, non-draft latest`)가 접근하지 못함을 재확인한다.
- [x] 설계 정합: tag namespace `moai-cli-v*`는 Go checker의 `v`/`go-v` 접두사 필터(RC 채널)에 걸리지 않는다(04-A 특성화 시험 `go_updater_feed_contamination_table` — `tests/update_cli.rs`… 04-B 이후 `tests/install_cli.rs`가 계승, 본 repo 기록). `/releases/latest`(stable 채널)는 접두사로 거르지 않는다. 그래서 rc.N(prerelease)은 latest가 될 수 없어 노출되지 않지만, 정식판은 `--latest=false` 없이 게시하면 latest가 되어 노출된다(E6 #2198 `tests/release_suite.rs` `rust_final_release_hides_newer_go_release_from_stable_clients`: dist 기본값이면 Go stable 검사가 `checksum_unavailable` 오류, `make_latest=false`면 더 새 Go 릴리스가 그대로 보임; `docs/release-attestation.md` 운영자 승인 항목 6). **다만 실측은 같은 repo 게시 후에만 성립** — 게시 전까지는 설계·fixture 근거 상태다.
- [ ] **not_run — `moai-cli-v*` 전용 tag ruleset**: (a) 참조. 결정됨, 리더 적용 대기(#2197).

## (h) 문서 예제 재현 (AC-08D-8)

- [x] 사용자 문서(README·user-entry-draft): 3.x 명령을 차세대 필수 단계로 안내하는 문장 **0건**. `/moai`·`$moai` 자연어 진입이 기본, 수동 명령은 명시적 fallback(`docs/user-entry-draft.md`)으로 분리.
- [x] docs 전역 기계 grep(`git grep -nP 'moai (init|update)\b' <tip> -- ':(glob)docs/*.md' README.md README.ko.md | grep -v 'fallback\|3\.x'`, tip `df215316ff`): 3건 — `docs/github-inventory-rc1.md:49`(forge 제목-매칭 규칙의 예시 문장), `docs/decisions-2026-11-08.md:11`(결정 장부의 요구 기록, "설계 반영만, 지금 실행 없음"), `docs/decisions-2026-11-08.md:165`(로컬 설치 방식 결정 — Go `moai update` 의미 변경 금지). 셋 다 내부 장부·기술 문서로 사용자 안내가 아니며 REQ-08D-06 위반 아님 — 기계 grep 결과를 있는 그대로 기록한다(11-09 기록은 2건, 165행은 11-10 결정으로 추가됨).
- [x] 지원표·체크리스트에 secret·개인 절대 경로 0건 (AC-08D-6 grep — 대상 3문서 0건; 일치는 시험 파일 자체의 패턴 리터럴뿐).
- [x] `.github/workflows/rc-ci.yml` yaml 파싱 exit 0, `pull_request` trigger 존재, `contents: read`, cargo test/clippy/fmt·Python unittest step 확인 — **파일 수정 불요**(00-E 작성본이 이미 08-A 명세 항목을 충족; 근거: 위 확인 출력). ruleset에 `rc-ci` 필수 체크 등록은 저장소 설정 = 별도 결정(HOLD).

## (i) SBOM / license

- [x] **SBOM 생성 실측 (E1 #2193 수리 — REQ-E1-01은 사용자 결정 2026-11-09 07:0x로 재정의: "dist build(바이너리·sha256) + cargo cyclonedx 단계(SBOM)")**: SBOM의 실제 산출은 **`cargo cyclonedx` 단계**가 담당한다. 세션 CARGO_HOME에서 워크트리 루트로 `cargo cyclonedx`(cargo-cyclonedx 0.5.10)를 실행한 뒤 `mv moai-cli.cdx.xml target/distrib/moai-cli.cdx.xml`로 Git이 무시하는 산출물 경로로 옮긴다. 루트에 XML을 남기거나 커밋하지 않는다. CycloneDX 1.3의 XML을 파싱하면 **`<component>` 요소 46개**(2026-11-09 tree `b35b75eb3` 기준 — 의존성 43개 + metadata의 패키지·lib·bin 3개)다. 그 뒤 `Cargo.lock`이 `1e4eb3c94e`·`2917621b4b`·`7db70948cc`에서 바뀌었다(`[[package]]` 수: `f552a5e781` 57 → tip `df215316ff` 103, `git show <sha>:Cargo.lock | grep -c '^\[\[package\]\]'` — dev·플랫폼 의존 포함이라 SBOM 수와 같은 척도는 아니다). 현재 tree의 component 수는 `cargo cyclonedx`를 다시 실행해야 알 수 있어 **not_run**이며, 게시 직전 재생성 수를 다시 적는다. `<components>` 컨테이너 2개는 세지 않는다. 이전 48 수치는 정정한다(인계받은 XML과 head `f92e97a3d5` / tree `b35b75eb3`의 재생성 XML에서 같은 46개를 확인). 재현: `python3 -c 'import xml.etree.ElementTree as E; r=E.parse("target/distrib/moai-cli.cdx.xml").getroot(); print(len(r.findall(".//{http://cyclonedx.org/schema/bom/1.3}component")))'` → `46`. 이전 결정 01:1x("SBOM=cargo-dist"가 build가 산출)는 이 재정의로 대체된다.
- [x] `dist-workspace.toml`에 `cargo-cyclonedx = true`·`cargo-auditable = true` 설정(dist 0.34 스키마 수용 실측). **관측 한계(upstream)**: dist build가 산출 목록에 `target/distrib/moai-cli.cdx.xml`을 announce하지만 파일을 실체화하지 않는다(2026-11-09 agent-7 2회 + 같은 날 수리 재현 1회, 플러그인 설치 상태에서도 동일 — upstream dist 0.34 관측 한계). 재정의된 REQ에서 SBOM 산출은 cyclonedx 단계가 담당하므로 이 한계는 REQ 이행을 막지 않는다.
- [ ] **not_run — attestation 연동**: SBOM의 attestation(`actions/attest@v4`)은 게시 단계(별도 권한) — REQ-08D-08에 따라 checksum이 서명을 대체한다고 쓰지 않음.
- [x] license: `publish = false` + Cargo.lock(의존 고정) — 정식 릴리스 시 그 tag에서 SBOM을 재생성해 첨부한다(아래 "릴리스 게시 절차" 3단계).

### E1 로컬 전체 재현 — 새 빈 CARGO_HOME

macOS arm64의 깨끗한 임시 clone 루트에서 아래 블록을 그대로 실행한다. Rust `1.94.1`(저장소 toolchain), 네이티브 빌드 도구, Python 3와 `shasum`이 필요하다. 설치·빌드는 네트워크를 사용한다. CARGO_HOME과 CARGO_TARGET_DIR은 각각 새 빈 디렉터리이며 다른 세션의 도구·레지스트리·빌드 캐시를 복사하지 않는다. `target/distrib`가 이미 있는 clone은 사용하지 않는다. all 산출물을 보관한 뒤 빈 distrib에서 local 범위를 확인하고, 마지막에 all 산출물과 SBOM을 복원한다. 이 블록은 로컬 빌드만 수행한다.

<!-- e1-local-repro:begin -->
```sh
set -eu
export CARGO_HOME="$(mktemp -d /tmp/moai-e1-cargo.XXXXXX)"
export CARGO_TARGET_DIR="$(mktemp -d /tmp/moai-e1-target.XXXXXX)"
test -z "$(ls -A "$CARGO_HOME")"
test -z "$(ls -A "$CARGO_TARGET_DIR")"
test ! -e target/distrib
test ! -e target/distrib-all
test ! -e target/distrib-local
export PATH="$CARGO_HOME/bin:$PATH"
printf 'CARGO_HOME=%s\nCARGO_TARGET_DIR=%s\n' "$CARGO_HOME" "$CARGO_TARGET_DIR"
cargo install --locked cargo-dist@0.34.0 cargo-cyclonedx@0.5.10 cargo-auditable@0.7.7
cargo install --list
for tool in dist cargo-cyclonedx cargo-auditable; do
    test "$(command -v "$tool")" = "$CARGO_HOME/bin/$tool"
done
dist --version
cargo cyclonedx --version
dist build --artifacts=all
python3 -c 'import subprocess; subprocess.run(["shasum", "-a", "256", "-c", "sha256.sum"], cwd="target/distrib", check=True)'
cargo cyclonedx
mv moai-cli.cdx.xml target/distrib/moai-cli.cdx.xml
python3 -c 'import xml.etree.ElementTree as E; r=E.parse("target/distrib/moai-cli.cdx.xml").getroot(); print("SBOM components:", len(r.findall(".//{http://cyclonedx.org/schema/bom/1.3}component")))'
mv target/distrib target/distrib-all
dist build --artifacts=local
python3 -c 'from pathlib import Path; import subprocess; p=Path("target/distrib"); sums=list(p.glob("*.sha256")); assert sums; [subprocess.run(["shasum", "-a", "256", "-c", f.name], cwd=p, check=True) for f in sums]'
test ! -e target/distrib/sha256.sum
test ! -e target/distrib/source.tar.gz
mv target/distrib target/distrib-local
mv target/distrib-all target/distrib
test ! -e moai-cli.cdx.xml
git status --short
```
<!-- e1-local-repro:end -->

## (j) 미결 목록

| 미결 | 성격 | 릴리스에 끼치는 것 |
| --- | --- | --- |
| `moai-cli-v*` tag ruleset 적용 | 리더(결정 A, 17:52Z(11-10)) | 적용·readback 전까지 REQ-08D-09는 절차 규약 |
| attestation 생성 workflow(`actions/attest@v4`) | release 운영(운영자) | AC-08D-5 not_run — provenance 부재 상태로는 게시 불가 |
| 배포 산출물 OS/arch matrix·locale | 사용자 결정(CI matrix는 01:1x(11-09) 결정: macOS arm64·x86_64 + Linux x86_64·arm64, Windows 빌드만) | 지금 배포 target은 `aarch64-apple-darwin` 하나(`dist-workspace.toml` `targets`) |
| SBOM(cargo-cyclonedx) | 사용자 결정 2026-11-09 07:0x 재정의("dist build + cargo cyclonedx 단계") → **E1에서 도입·실측 완료** | (i) 참조 — attestation 연동만 게시 단계 not_run |
| macOS 서명 | 결정 01:1x(11-09): notarization 없음, attestation+sha256 | quarantine 안내(`docs/release-attestation.md` 4.2)로 대신 — 미서명 표기 필요 |
| T2 호스트 로딩 운영자 실측 | 운영자(자격증명) | certified 지원 행 0 — 지원표 머리의 한계 문단 유지 |

## (k) 독립 감사 verdict

- [ ] `docs/release-review.md` — 독립 reviewer가 reviewer identity·공유 context 범위·대조 digest·gap 표와 함께 기록한다. 필수 gap이 있으면 verdict는 `release blocked`(현재 예상 verdict: **release blocked** — (d) artifact 목록·(e) attestation·(g) 게시 실측이 not_run).

## 릴리스 게시 절차 (별도 권한 단계 — 이 체크리스트가 수행하지 않는다)

1. 운영자가 위 HOLD/not_run을 해소·승인한다. 2. `git tag moai-cli-v2026.11.0-rc.1 <candidate>` + `--verify-tag`. 3. `gh release create --prerelease --latest=false --verify-tag` + 자산 + `sha256.sum` + `moai-cli.cdx.xml`(그 tag에서 재생성한 SBOM, (i) 절차). 4. attestation 생성 후 `gh attestation verify`. 5. `/releases/latest`가 RC tag을 반환하지 않음을 재확인. 6. 이 체크리스트의 not_run 항목을 실측값으로 갱신한다. 같은 version 재배포 금지 — 교정은 `rc.2`.

## 리허설 기록 (2026-11-09, #2262 — tag·release·공개·서명 0)

tree `v2026.11.0` tip `cf4e662aff5b33ceb96c47757eafa9eb6f250f20`, 실행 2026-11-09T04:58Z~05:05Z(`date -u`), 개발 Mac(Darwin/arm64). 위 (a)~(k)의 체크 표시는 08-D 당시 기록이고, 이 표는 같은 항목을 이번에 다시 실행한 결과다. verdict는 그대로 **`release blocked`**다.

이 표는 2026-11-09 tip `cf4e662aff` 기록이다. 그 뒤 바뀐 행은 2026-11-12 대조(#2157 댓글 6812666007, #2492)에 따라 '현행' 열에 적는다(tip `df215316ff746224d3d14806737a15017891fd83`, 원격 관측 2026-11-12T03:59Z). '현행'의 not_run은 이번 대조가 실행하지 않은 항목이다. 다시 리허설하면 이 표 전체를 새 tip으로 교체한다.

| 항목 | 상태 | 명령 → 관측 | 현행 (2026-11-12, tip `df215316ff`) |
| --- | --- | --- | --- |
| (a) version | passed | `grep -m1 '^version' Cargo.toml` → `version = "2026.11.0-rc.1"` | 맞음 |
| (a) tag 미생성 | passed | `git ls-remote --tags origin 'moai-cli-v*'` → 출력 없음(tag 0) | 맞음 — 2026-11-12T03:59Z `git ls-remote --tags origin 'moai-cli-v*'` 0줄 |
| (a) `moai-cli-v*` tag ruleset | blocked | `gh api repos/modu-ai/moai-adk/rulesets` → `26791615 RC1 integration gate (v2026.11.0) target=branch`, `19583648 Release tag immutability (v*) target=tag`; `…/rulesets/19583648 --jq .conditions.ref_name` → `include: ["refs/tags/v*"]` — `moai-cli-v*`는 덮이지 않음. 추가는 저장소 설정(운영자 결정) | 관측은 같음(03:59Z ruleset 2개, 19583648 `["refs/tags/v*"]`). **결정됨**: 01:1x(11-09) 생성 카드, 17:52Z(11-10) 표 37–41 A — 리더 적용 대기(#2197 OPEN) |
| (b) candidate SHA | not_run | 후보 SHA는 게시 승인 때 정한다. 이 리허설의 tree는 위 tip이다 | 맞음 |
| (c) `cargo test --offline --locked` | passed | exit 0, 369 passed / 0 failed | **낡음(수치)** — 그 뒤 시험이 크게 늘었다(agent-8 부분 실행 6811336861만 727 passed). 현재 결과는 not_run |
| (c) `cargo clippy --offline --locked --all-targets -- -D warnings` | passed | exit 0 | not_run(이번 대조는 실행하지 않음) |
| (c) `cargo fmt --check` | passed | exit 0 | not_run(이번 대조는 실행하지 않음) |
| (c) `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py'` | passed | exit 0, Ran 80, OK | **낡음(수치)** — Ran 80 → agent-8 보고 Ran 312. 현재 결과는 not_run |
| (c) `python3 -I tools/check_support_matrix.py docs/support-matrix.md docs/e2e-receipts/` | passed | exit 0, `supported rows: 6` / `passed receipts: 6 (9 receipts total)` | receipt 수는 맞음(tip `df215316ff` `receipts.jsonl` 9줄, `result: passed` 6: AC-08C-5·1·2·7·8, HA-02). 도구 실행은 not_run |
| (d) 로컬 release 빌드 sha256 자체 검증 | passed | `cargo build --offline --locked --release` exit 0; `shasum -a 256 <target>/release/moai-cli > sha256.sum && shasum -a 256 -c sha256.sum` → `OK`, `f828b1a8df658dfce07f49d8f06d750c617689bbb2eb0787852cfcce17f8ffd1`(CARGO_TARGET_DIR=/tmp/moai-rc-target-agent-9) | **낡음(값)** — 그 뒤 소스·`Cargo.lock`이 바뀌어 같은 값일 수 없다. 현재 값은 not_run |
| (d) release artifact checksum 목록 | not_run | artifact 미생성(게시 금지 범위) | 맞음 — release 0(03:59Z `gh release view moai-cli-v2026.11.0-rc.1` → release not found) |
| (e) attestation·`gh attestation verify` | blocked | workflow 실행·게시 미허용(사용자 승인) | 맞음 |
| (e) 서명(codesign/notarization) | blocked | 서명 키 사용 금지 | **결정됨** — notarization 없음, attestation+sha256(01:1x(11-09)), quarantine 안내 문서로 대신(`docs/release-attestation.md` 4.2) |
| (f) known-good restore | passed | `docs/e2e-receipts/receipts.jsonl`에 `AC-08C-1 passed` 1건(08-C T1, 2026-11-08) — 이번에 다시 실행한 것은 위 `cargo test` 전체다 | **낡음(digest)** — 그 receipt의 `skill_digest` `bff1c53b…` ≠ tip `df215316ff` `skills/moai/SKILL.md` sha256 `25081f6c…`. 새 receipt는 #2014 가지(미병합) |
| (g) Go updater 격리(게시 전 상태) | passed | `gh api repos/modu-ai/moai-adk/releases/latest --jq '.tag_name+" prerelease="+(.prerelease\|tostring)'` → `v3.1.2 prerelease=false`; `gh release view moai-cli-v2026.11.0-rc.1` → `release not found` | 맞음 — 03:59Z `v3.1.2 prerelease=false`, RC release not found |
| (g) 게시 후 `/releases/latest` 재확인 | not_run | 게시 후에만 측정 가능 | 맞음 |
| (h) 3.x 명령 안내 grep | passed | `grep -nE 'moai (init\|update)\b' docs/*.md README.md \| grep -v 'fallback\|3\.x'` → 2건(`docs/decisions-2026-11-08.md:11`, `docs/github-inventory-rc1.md:49`, 둘 다 내부 장부) — 08-D 기록과 같음. 이 tip에는 저장소 루트 `README.md`가 없다(grep 경고 `No such file`) — README 확인 부분은 not_run | **낡음** — tip `df215316ff`에는 `README.md`·`README.ko.md`가 있다. `git grep -nP 'moai (init\|update)\b' df215316ff -- ':(glob)docs/*.md' README.md README.ko.md \| grep -v 'fallback\|3\.x'` → README 0건, docs 3건(`decisions:11`·`decisions:165`·`github-inventory:49`, 모두 내부 장부). `git grep`은 `-P`를 쓴다(`-E`는 `\b` 미지원으로 0건을 낸다) |
| (h) secret·개인 경로 | passed | `tests/support_matrix_test.py` `test_real_documents_have_no_secret_shapes`(개인 홈 경로·토큰·개인키 모양) 통과. 같은 모양을 릴리스 노트에도 grep → 0건. 패턴 문자열은 이 문서에 적지 않는다(적으면 시험이 문서 자체를 잡는다) | not_run(이번 대조는 실행하지 않음) |
| (h) `rc-ci.yml` | passed | `ruby -ryaml` 파싱 → trigger `pull_request`(branches `v2026.11.0`)·`workflow_dispatch`, `permissions: {contents: read}` | 맞음(보강: 4개 target matrix — Linux x86_64·arm64, macOS arm64·x86_64 — 와 집계 `rc-ci` job, Windows 빌드 job) |
| (h) ruleset 필수 체크 | blocked | `gh api …/rulesets/26791615` → required_status_checks `["moai/audit"]`만 — `rc-ci` 필수 등록은 저장소 설정(HOLD) | 맞음(HOLD 유지) — 03:59Z `["moai/audit"]` |
| (i) SBOM | blocked | HOLD(사용자 결정 대기) | **낡음** — 결정 07:0x(11-09)로 'dist build + cargo cyclonedx 단계'로 재정의되어 도입됨((i)절). 현재 tree의 component 수는 재생성 전까지 not_run |
| (k) 독립 감사 verdict(`docs/release-review.md`) | not_run | 이 리허설은 감사가 아니다. #2262 PR의 독립 감사는 agent-2 | 맞음 — verdict 초기값 `release blocked` 그대로 |
| T2 host 로딩 E2E | not_run | #2014 OPEN — receipt `AC-08C-3`·`AC-08C-4`·`AC-08C-6` = `not_run`(receipts.jsonl) | 맞음 — #2014 OPEN, receipt AC-08C-3·4·6 `not_run` |

이 리허설에서 하지 않은 것: `git tag`, `gh release create`, 공개 게시, 서명 키 사용, workflow 실행, 저장소 설정 변경.
