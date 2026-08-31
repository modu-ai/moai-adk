---
id: SPEC-VERSION-STAMP-GUARD-001
title: 버전 스탬프 목록의 유령을 잡는다 — 절반의 회귀 보장과 그 절반을 밝히는 일
version: "0.4.0"
status: draft
created: 2026-08-31
updated: 2026-09-01
author: manager-spec
priority: Medium
phase: "v3.1.5 target"
module: internal/cli, .moai/docs
lifecycle: spec-anchored
tags: "version-management, release, documentation-rot, ghost-path, partial-guarantee"
tier: S
---

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 |
|---|---|---|---|
| 0.1.0 | 2026-08-31 | manager-spec | 최초 작성(카드 t388, plan-phase). `.moai/reports/t388/baseline.md`(트리 `9328a5242` 실측) 위에 세움. 요구 7 / 수락 7 |
| 0.2.0 | 2026-08-31 | manager-spec | plan-audit iter-1 **FAIL 0.77**(`.moai/reports/t388/plan-audit.md`) 대응. D1·D3·D4·D5-D10 정정, D2는 미해결 blocker로 기록 |
| 0.3.0 | 2026-08-31 | manager-spec | **운영자 판정으로 카드 분할.** 토큰 술어 가드 축 전체가 카드 **t392**(Tier M, t388 착지에 의존)로 이관됐다. 남은 것은 문서 수정 3건 + **유령 검사 하나**다. 삭제: 0.2.0의 REQ-VSG-004(목록 비의존 스윕) · REQ-VSG-005(토큰 술어) · REQ-VSG-006(토큰 가드 비공허성)과 그에 딸린 §3 거부 목록 설계 · §4 `eba919e44` RED 픽스처 · `internal/versionstamp/` 패키지 · AC 3건. **D2는 수리가 아니라 범위 이관으로 해소**됐고 측정치는 §2에 보존한다. D1·D3는 §3과 함께 소멸. **D4는 남아 이 개정판의 주 결함**이며 AC-VSG-001에서 닫았다. D7은 살아남은 검사에 맞춰 재적용(합성 입력 고정), D8은 REQ-VSG-004의 `Where` 절로 이월. 요구 6 / 수락 6, 연속 번호 재부여(Tier S 상한 8/8 이내). 축소 과정에서 §7 R-1의 「CI가 얕게 체크아웃해서」 근거를 삭제 대신 **범위 한정**했다 — 실측상 `ci.yml` checkout 7개 중 6개가 `fetch-depth: 0`이고 `go test ./...` job도 그중 하나(`ci.yml:129`)라 그 근거는 이 검사에 성립하지 않는다. 성립하는 두 근거(도달 가능성은 깊이로 해결되지 않음 · 검사의 VCS 비의존)로 교체하고, `spec-lint.yml`의 fetch-depth 부재는 실행 주체가 아니므로 관측으로만 남겼다. **측정 정정 1건**: §2 토큰 히스토그램이 옆에 인쇄된 명령으로 재현되지 않았다 — `-n` 출력에서 토큰을 뽑아 **파일 이름 속 버전 문자열이 매 매치 줄마다 중복 계수**됐고(`v2.14.0` 112→72, `v2.12.0` 90→83), `-h` 기반으로 재측정해 교체했다. 면적(2,225줄 / 592파일)은 영향 없음. 함정 자체를 §2에 [HARD]로 기록했다(t392가 그대로 물려받는다). 요구·수락 불변이므로 `version`은 0.3.0 유지 |
| 0.4.0 | 2026-09-01 | manager-spec | plan-audit iter-2 **FAIL 0.80**(`.moai/reports/t388/plan-audit-iter2.md`) 대응. **D1(major/blocking)**: 검사의 RED 증거를 유령에 귀속시키던 서술을 걷어내고, **단언별로 각자의 RED을 관측**하도록 재설계했다 — 앵커는 M2가 만들 소제목을 가리키므로 M1 트리에서 파싱은 0건이고 그때 우는 것은 **개수 단언**(AC-VSG-005)이지 존재 단언이 아니다. 존재 단언의 RED은 M2 착지 트리에서 **경로 한 줄을 치환**해 관측한다(추가가 아니라 치환이라 개수가 7로 유지돼 원인이 하나다). 기대 RED 문자열을 측정 **전에** 못박았다(`plan.md` §D 단언 메시지 계약). D5: REQ-VSG-005의 순환 비교를 「검사가 보유한 기대 개수(M2 수정 후 7)」로 교체. D4: AC-VSG-006 3항을 판단에서 **grep 가능한 계기**로 교체(양성 존재 + 리터럴 거부 목록). D2: `§3의 단위 고정 조항` 걸린 포인터 제거. **D3: 측정 정정** — 「이름에 버전 토큰이 든 파일이 둘」은 거짓이고 거부 목록 범위 안에 **여덟**이다(재측정 목록·계수 §2). D6: 존재 보고를 비치명(`t.Errorf`)으로 못박아 한 번의 실행에서 두 단언이 모두 관측되게 했다. D7: AC-VSG-006의 Given을 문서로 한정. D8: `71-78행` 인용을 `71-74·77-78행`으로 정정. D9: REQ-VSG-004의 `Where` 절을 본절로 접었다. 요구 6 / 수락 6 불변 |

---

## §1 문제 — 목록이 썩었고, 썩은 것을 아무도 알려주지 않았다

`.moai/docs/version-management.md`의 「Files Requiring Version Sync」(제목 66행, 항목 71-74·77-78행
— 75행은 공백, 76행은 `**Configuration Files:**` 라벨이다)는
릴리스마다 손으로 갱신해야 할 파일을 사람이 적어 둔 목록이다. 이 목록에는 강제력이 없다 —
어긋나도 아무 신호가 나지 않는다. 그래서 어긋났다.

권위 집합의 출처: v3.1.4 범프 커밋 `61921f1ba`. 그 커밋이 바꾼 것은 **7파일 9줄**이다.

    .moai/config/sections/system.yaml   2줄
    README.ja.md                        1줄
    README.ko.md                        1줄
    README.md                           1줄
    README.zh.md                        1줄
    docs-site/hugo.toml                 2줄
    pkg/version/version.go              1줄

문서의 목록과 대조하면:

| 종류 | 항목 | 상태 |
|---|---|---|
| 유령 | `internal/template/templates/.moai/config/config.yaml` | HEAD에 없음 — `test -e <경로>` 종료 1 |
| 누락 | `README.ja.md`, `README.zh.md`, `docs-site/hugo.toml`, `pkg/version/version.go` | 매 범프마다 바뀌는데 목록에 없음 |

`docs-site/hugo.toml`은 이 결함이 실제로 릴리스를 오염시킨 증거다. v3.1.3 범프 커밋
`eba919e44`는 6파일만 건드렸고 `hugo.toml`을 빠뜨렸다 — 그 트리에서
`pkg/version/version.go:8`은 `v3.1.3`인데 `docs-site/hugo.toml:55`는 `v3.1.2`였다.
카드 t274가 별도로 뒤늦게 고쳤다(`175d63f3f`).

**이 사례는 누락이었지 유령이 아니다.** §4가 이 구분 위에 서 있다.

### §1.1 `version.go`는 누락이 아니라 모순이다

문서의 「Single Source of Truth」절은 두 곳에서 `pkg/version/version.go`가 **파생값**이라고
단언한다:

    8:  - [HARD] `pkg/version/version.go` reads from git tags at build time
    12: - Runtime Access: `pkg/version/version.go` via `git describe`

그런데 모든 범프 커밋이 `version.go:8`을 손으로 고친다. 둘 다 참일 수 없다.

이 세션이 빌드 경로에 대고 확인한 사실:

    Makefile:20        LDFLAGS := -ldflags "... -X $(MODULE)/pkg/version.Version=$(VERSION) ..."
    Makefile:36        build:   go build   $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/moai
    Makefile:72        install: go install $(LDFLAGS) ./cmd/moai
    .goreleaser.yml:22 -X github.com/modu-ai/moai-adk/pkg/version.Version={{.Version}}

주입 경로는 둘 다 있다 — 로컬 `make build`/`make install`과 릴리스 빌드(GoReleaser). 그 둘로
만든 바이너리에서는 기본값이 보이지 않는다. 기본값이 사용자에게 보이는 경로는 **손빌드
하나**다: `go install ./cmd/moai`를 맨손으로 부르면 LDFLAGS가 없어 `version.go`의 값이 그대로
박힌다. 이 경로는 실재하며 `CLAUDE.local.md` §11이 독립적으로 기록하고 금지한다.

따라서 `version.go`의 `Version`은 파생값이 아니라 **ldflags 부재 시의 폴백**이며, 손으로
유지되는 스탬프가 맞다. 문서 8·12행이 틀렸다. 다만 노출 면적은 릴리스 바이너리가 아니라
손빌드 경로에 한정된다 — 「모든 사용자에게 보인다」로 넓혀 읽지 않는다.

(`Version`은 `const`가 아니라 패키지 수준 `var`다 — `-X`로 주입하려면 그래야 한다.
이 SPEC은 「상수」로 부르지 않고 「기본값」 또는 「폴백」으로 부른다.)

### §1.2 유령 항목의 후계자는 스탬프가 아니다

유령 경로가 가리키려던 자리는 지금
`internal/template/templates/.moai/config/sections/system.yaml.tmpl`이다. 그 파일은 버전을
**렌더**한다:

    6: version: "{{.Version}}"
    9: template_version: "{{.Version}}"

리터럴 버전 문자열이 없으므로 범프 대상이 **아니다**. 목록을 그냥 지우면 다음 사람이 같은
자리를 다시 적어 넣을 것이므로, 「스탬프가 아니다」를 사유와 함께 남긴다.

### §1.3 목록은 두 종류를 섞어 놓았다

현재 목록은 성격이 다른 둘을 한 덩어리로 적어 두었다.

- **버전 스탬프** — 범프 커밋이 기존 문자열을 새 문자열로 **다시 쓴다**. 위 7파일.
- **릴리스 산출물** — `CHANGELOG.md`, `.moai/release-notes/vX.Y.Z.ko.md`. 릴리스마다
  **새로 집필**되며 범프 커밋은 건드리지 않는다(`61921f1ba` numstat에 없음).

둘을 섞으면 「목록을 다 처리했다」가 두 가지 다른 행위를 뜻하게 된다. 그리고 §5가 보이듯
**유령 검사의 대상 집합도 이 분리 없이는 정의되지 않는다** — 릴리스 산출물 쪽 경로는
플레이스홀더라 존재 검사에 걸리면 안 되기 때문이다.

---

## §2 전제 — 왜 토큰 술어 가드가 이 카드에 없는가

0.1.0과 0.2.0은 「권위 버전과 다른 모든 `vX.Y.Z` 토큰은 설명되어야 한다」는 술어 위에 회귀
가드를 세우려 했다. 그 설계는 **카드 t392**(Tier M, 이 카드 착지에 의존)로 이관됐다. 운영자
판정이며 이 SPEC은 재론하지 않는다.

이관의 근거가 된 측정을 여기 보존한다 — t392를 설계하는 사람이 같은 벽을 다시 발견하지
않도록.

[HARD] **계수 단위.** 아래에는 단위가 다른 두 수가 있다. **줄(line)** 수는 술어에 걸리는
면적이고, **출현(occurrence)** 수는 토큰 히스토그램의 단위다. 한 줄에 토큰이 둘 이상 있을 수
있어 둘은 일치하지 않는다(실측 2,225줄 / 2,494출현). 단위를 붙이지 않은 수는 두 배 가까이
어긋난 두 값 중 어느 쪽인지 판별할 수 없으므로, 어느 수든 단위를 붙여 적는다.

**면적 (줄 / 파일)** — 트리 HEAD `9328a5242`:

    git grep -nE 'v[0-9]+\.[0-9]+\.[0-9]+' -- . \
      ':!.moai/reports' ':!.moai/specs' ':!.moai/release-notes' \
      ':!CHANGELOG.md' ':!*_test.go' ':!docs-site/content/*/changelog*'

    -> 2225줄 / 592파일

**토큰 히스토그램 (출현)** — 같은 거부 목록, 같은 트리. 명령과 출력은 **같은 실행**이다:

    git grep -hoE 'v[0-9]+\.[0-9]+\.[0-9]+' -- . \
      ':!.moai/reports' ':!.moai/specs' ':!.moai/release-notes' \
      ':!CHANGELOG.md' ':!*_test.go' ':!docs-site/content/*/changelog*' \
      | sort | uniq -c | sort -rn | head -6

    270 v3.0.0
     83 v2.12.0
     80 v3.1.1
     80 v2.1.219
     72 v2.14.0
     68 v2.1.198

    (전체 출현: 2494)

[HARD] **`-h`가 load-bearing이다.** `-n`(경로:행번호:내용) 출력에서 토큰을 뽑으면 **파일 이름
안의 버전 문자열이 매 매치 줄마다 함께 세어진다**. 거부 목록 범위 안에서 이름에 버전 토큰이
든 파일은 **여덟**이며, 각자 자기 매치 줄 수만큼 부풀린다(재측정, 트리 `9328a5242`):

    40  docs/design/v2.14.0-release-plan.md
    24  .moai/release/RELEASE-NOTES-v2.17.0.md
    16  .moai/release/MIGRATION-v2.17.0.md
    12  .moai/release/RELEASE-NOTES-v2.16.0.md
     7  .moai/marketing/awesome-lists/github-release-v2.12.0-enhanced.md
     6  .moai/release/RELEASE-NOTES-v2.15.0.md
     4  .moai/release/v2.15.0-draft.md
     4  .moai/release/RELEASE-NOTES-v2.20.0.md
    ---
   113  합계

    측정: awk -F: '{print $1}' <위 -n 출력> | grep -E 'v[0-9]+\.[0-9]+\.[0-9]+' \
            | sort | uniq -c | sort -rn

이 113이 그대로 부풀림의 총량이다 — `-n` 출력에서 뽑은 전체 출현은 **2607**, `-h` 기반
**2494**, 차 **113**으로 정확히 맞는다. 토큰별로는 `v2.14.0` 72→112(+40),
`v2.12.0` 83→90(+7), `v2.17.0` 25→65(+40 = 24+16, 두 파일이 합산됨)이다. 0.3.0 이전 판이
실제로 `-n` 기반 값을 실었고, 옆에 인쇄된 명령으로는 재현되지 않았다. **버전 토큰을 세는
도구는 경로가 아니라 내용만 보아야 한다** — t392가 그대로 물려받을 함정이다.

[HARD] **여덟 중 여섯이 `.moai/release/`에 있다는 점을 함께 적는다.** 이 거부 목록은
`.moai/release-notes/`를 제외하지만 `.moai/release/`는 **제외하지 않는다**. 이름이 한 글자
차이라 t392가 「릴리스 노트는 이미 뺐다」로 읽고 넘길 자리이며, 실제로는 부풀림 113줄 중
**66줄**이 그 디렉터리에서 나온다(실측:
`awk -F: '{print $1}' <-n 출력> | grep -E 'v[0-9]+\.[0-9]+\.[0-9]+' | grep -c '^\.moai/release/'`
→ 66). 두 경로는 서로 다른 디렉터리다.

파일 상위(출현): `go.sum` 243 · `go.mod` 95. 디렉터리 상위(줄): `docs-site/content` 535 ·
`internal/template` 233 · `.claude/rules` 145.

**값으로 가르는 판별자는 성립하지 않는다.** `v3.0.0`(270회)과 `v3.1.1`(80회)은 이 제품의 과거
버전이면서 **동시에 정당한 역사 서술**이다. 같은 문자열이 한쪽에서는 낡은 스탬프이고 다른
쪽에서는 옳은 기록이므로, 토큰만 보고 가를 수 없다. `v3.*`로 좁혀도 남의 소프트웨어 버전
(`v3.13.1`, `v3.5.0`)이 섞이고, `go.sum` 243회·`go.mod` 95회는 전부 의존성 핀이다. t392가
설계해야 할 것이 바로 이 판별자다.

**방향이 뒤집혔다는 점을 밝힌다.** 0.2.0의 가드는 「목록을 읽지 않는다」가 [HARD] 요구였다 —
목록의 결함을 물려받지 않기 위해서였다. 이 카드에 남은 유령 검사는 반대로 **목록을 읽는
것이 본업**이다. 검사 대상이 목록 자체이기 때문이며, 두 요구는 서로 다른 가드에 붙은 서로
다른 조건이다. 이 SPEC 안에 「목록을 읽지 않는다」는 요구는 더 이상 없다.

---

## §3 요구 (GEARS)

**REQ-VSG-001** — The version-sync documentation shall name exactly the set of files a version bump rewrites, and shall name no path that does not exist in the repository. The ghost entry shall be removed and the four omitted paths shall be added.

**REQ-VSG-002** — The documentation shall separate version stamps from release artifacts under distinct headings. Files a bump commit rewrites and files a release authors fresh shall not appear under one undifferentiated list.

**REQ-VSG-003** — The documentation shall state that `pkg/version/version.go`'s `Version` default is the no-ldflags fallback and is hand-maintained at every bump. The assertions that it is derived from git tags at build time shall be corrected rather than left standing beside the corrected list. `Version` is a package-level `var`, not a constant, and the documentation shall not call it one.

**REQ-VSG-004** — A regression check shall read the paths named under the documentation's version-stamp heading, shall fail when any of them does not exist in the working tree, and shall name the offending path in its failure output. The check shall judge only that heading's paths; paths named under the release-artifact heading shall not be judged for existence, because that heading carries a placeholder rather than a literal path.

**REQ-VSG-005** — The check shall report how many paths it parsed and shall fail when that count differs from an expected entry count the check itself holds, independent of the parse (7 after the REQ-VSG-001 correction). A run that parsed nothing shall therefore be reported as a failure, never as a pass.

**REQ-VSG-006** — The documentation and this SPEC shall state that the regression guarantee established here is partial: the check detects a named path that does not exist, and does not detect a stamp site absent from the list. Neither shall assert that the list can no longer rot.

### §3.1 요구 ↔ 수락 추적

| 요구 | 수락 |
|---|---|
| REQ-VSG-001 | AC-VSG-001 |
| REQ-VSG-002 | AC-VSG-002 |
| REQ-VSG-003 | AC-VSG-003 |
| REQ-VSG-004 | AC-VSG-004 |
| REQ-VSG-005 | AC-VSG-005 |
| REQ-VSG-006 | AC-VSG-006 |

요구 6 / 수락 6 — Tier S 상한(각 8)을 넘지 않는다(`spec-workflow.md` § SPEC Complexity Tier).
전문은 `acceptance.md` §D.

---

## §4 이 카드가 세우는 보장은 절반이다

[HARD] **이 카드가 착지해도 목록은 여전히 썩을 수 있다. 썩는 방향이 하나 줄어들 뿐이다.**

목록이 어긋나는 방향은 둘이고, 이 카드는 하나만 닫는다.

| 방향 | 예 | 이 카드 | 왜 |
|---|---|---|---|
| 목록이 **없는 경로**를 가리킨다(유령) | `internal/template/…/config.yaml` | **닫는다** — REQ-VSG-004 | 경로 존재는 목록만 읽고 판정된다 |
| **실제 스탬프 사이트**가 목록에 없다(누락) | `docs-site/hugo.toml`(v3.1.3에서 실제로 발생) | **닫지 못한다** | 누락을 찾으려면 목록 밖을 봐야 하고, 그러려면 이 저장소의 버전과 남의 버전을 가를 판별자가 필요하다(§2) |

닫지 못하는 쪽이 **이 카드를 만들게 한 실제 사고**라는 점을 분명히 한다. §1의 `hugo.toml`
사례는 누락이었다. 유령 검사는 그 사고를 재발 시점에 잡지 못한다.

그 절반은 **카드 t392**의 몫이다(Tier M, SPEC 미발행, 이 카드 착지에 의존).

「목록이 더는 썩지 않는다」에 해당하는 서술은 문서에도 이 SPEC에도 쓰지 않는다. 이 카드가
존재하는 이유가 **문서가 아는 것보다 더 많이 주장했기 때문**이므로, 같은 실수를 수리
과정에서 되풀이하지 않는다.

---

## §5 검사의 거처

기존 관례를 읽고 정했다. 새 패키지를 만들지 않는다 — 0.2.0의 `internal/versionstamp/`는
t392와 함께 나갔고, 남은 것은 파일 하나 분량이다.

- `internal/cli`에 문서 텍스트를 훑는 가드 테스트 선례가 이미 있다:
  `internal/cli/deprecated_paths_text_reference_test.go`. 그 파일의 머리 주석은 자기가
  **무엇을 잡지 못하는지**를 먼저 밝히는데, §4가 요구하는 정직성과 같은 모양이다.
- 저장소 루트 해석 헬퍼도 같은 패키지에 있다: `repoRootFromCLITest`
  (`internal/cli/hook_flush_test.go:22`, `filepath.Abs(filepath.Join("..", ".."))`).
- `.github/workflows/ci.yml:208`이 `go test ./...`를 돈다. 새 파일은 별도 CI job 없이 잡힌다.

산출은 **파일 하나**: `internal/cli/version_sync_list_test.go`.

[HARD] 검사는 저장소 객체를 조회하지 않는다. 작업 트리의 파일 존재만 본다 — `os.Stat` 수준.
이력도 다른 브랜치도 보지 않으며, 그래서 §7 R-1의 도달 불가능 문제에 걸리지 않는다.

[HARD] 검사 대상은 **버전 스탬프 소제목 아래 항목으로 한정**한다. 릴리스 산출물 소제목의
`.moai/release-notes/vX.Y.Z.ko.md`는 **플레이스홀더**이며 그 이름의 파일은 존재하지 않는다
(실측: 해당 경로 부재. 실재 파일은 `v3.1.0.ko.md`·`v3.1.3.ko.md`). 이 한정이 없으면 검사가
정상 항목을 유령으로 지목한다. §1.3의 분리가 문서 정리를 넘어 **검사의 전제**인 이유다.

### §5.1 앵커는 아직 없는 소제목을 가리킨다 — 그래서 첫 RED은 유령 RED이 아니다

[HARD] 위 한정의 직접적 귀결을 적는다. 앵커로 삼을 **버전 스탬프 소제목은 오늘 트리에
없다** — 실측하면 `.moai/docs/version-management.md`에는 `### Files Requiring Version Sync`
(66행) 아래 `**Documentation Files:**`(70행)와 `**Configuration Files:**`(76행)뿐이고, 그 축은
문서 대 설정이지 스탬프 대 산출물이 아니다. 스탬프 소제목은 문서 수정이 **만든다**.

따라서 검사가 먼저 착지한 트리에서 파싱은 **0건**이고, 그때 우는 것은 개수 단언
(REQ-VSG-005)이지 존재 단언(REQ-VSG-004)이 아니다. 유령
(`internal/template/templates/.moai/config/config.yaml`, 78행)은 `**Configuration Files:**`
아래에 있어 위 [HARD] 한정이 **읽지 않는다**.

[HARD] 그러므로 두 단언은 **각자의 RED을 따로 관측**한다. 한쪽이 낸 빨강을 다른 쪽의 증거로
적지 않는다. 근거는 `verification-completeness.md` §1.1 — 완료의 단위는 검사 파일이 아니라
「실패가 관측된 단언」이며, 아무것도 매치하지 않아서 나온 신호는 존재 단언에 대해 아무것도
단언하지 않는다.

[HARD] 존재 보고는 **비치명**이다(Go로는 `t.Errorf`, `t.Fatalf`가 아니다). 두 단언이 한
실행에서 모두 실패하는 트리가 실제로 존재하므로(문서 수정 중간 상태), 첫 단언에서 멈추면
존재 단언의 「경로를 이름으로 지목한다」가 관측되지 않은 채 검사만 빨갛게 된다.

Template-First 규칙은 걸리지 않는다 — 산출물이 `.claude/`,
`internal/template/templates/` 아래에 없고, `.moai/` 아래 산출물은 `.moai/specs/`와
`.moai/docs/`뿐이다.

---

## §6 범위 밖

### Out of Scope — 토큰 술어 가드 일체 (t392로 이관)

- 권위 버전과 다른 버전 토큰을 훑는 가드, 그 거부 목록, 면제 표, `internal/versionstamp/`
  패키지, `eba919e44` RED 픽스처 — 전부 이 카드에서 만들지 않는다.
- 「이 저장소의 스탬프와 남의 버전을 가를 판별자」를 정의하지 않는다. 그것이 t392의 첫
  결정이다.
- 0.2.0에 있던 요구·수락 중 위에 해당하는 것은 **삭제**했지 옮겨 쓰지 않았다.

### Out of Scope — 누락 방향의 탐지

- 목록에 없는 실제 스탬프 사이트를 찾는 일은 하지 않는다(§4). 이 카드의 검사는 목록 안만
  본다.

### Out of Scope — 범프의 기계적 자동화

- 버전 범프를 수행하는 명령·스크립트는 만들지 않는다. 범프는 사람과 릴리스 하네스의 몫이다.
- 운영자가 기각한 「범프 대상의 기계적 유도」를 다시 열지 않는다.

### Out of Scope — `scripts/release.sh` 및 릴리스 하네스

- 릴리스 절차, 태깅, GoReleaser, `hns-release-specialist`는 건드리지 않는다.
- `CHANGELOG.md`·`.moai/release-notes/`의 집필 규율도 그대로 둔다. §1.3의 분리는 문서의
  **분류**이지 절차 변경이 아니다.

### Out of Scope — 버전이 아닌 스탬프

- `docs-site/hugo.toml:56`의 `releaseDate`처럼 버전 토큰이 아닌 값은 다루지 않는다(§7 R-3).

---

## §7 잔여 위험

- **R-1 — 권위 집합의 출처가 이 브랜치에서 도달 불가능하다.** §1의 7파일은 v3.1.4 범프 커밋
  `61921f1ba`에서 유도했는데, 실측 결과 그 커밋은 HEAD의 조상이 **아니다**(조상 검사 rc=1).
  병합되지 않은 `release/v3.1.4`에만 있다.

  그래서 AC-VSG-001은 그 커밋을 판정자로 쓰지 않는다. `acceptance.md`의 **리터럴 7경로**가
  판정자이고 `61921f1ba`는 **출처 인용**으로 강등됐다. 남는 사실은 이것이다: 이 SPEC이 권위로
  삼은 집합은 이 브랜치에서 재유도할 수 없으며, 그 커밋이 병합되기 전까지 「목록이 옳다」는
  주장은 리터럴 집합의 정확성에 의존한다.

  **사유를 정확히 적는다 — 「CI가 얕게 체크아웃해서」가 아니다.** 이 저장소에서 그 근거는
  대체로 성립하지 않는다. 실측: `.github/workflows/ci.yml`의 checkout 7개 중 **6개가
  `fetch-depth: 0`**(전체 이력)이며, `go test ./...`를 도는 test job의 checkout도 그중
  하나다(`ci.yml:129` — drift walker 테스트가 전체 이력을 요구한다는 주석이 붙어 있다).
  성립하는 근거는 둘이다.

  1. **도달 가능성은 깊이로 해결되지 않는다.** `fetch-depth: 0`은 체크아웃된 ref의 이력을
     전부 가져올 뿐 다른 브랜치의 팁을 가져오지 않는다. 도달 불가능한 객체는 어떤 깊이
     설정으로도 들어오지 않으며, `61921f1ba`가 정확히 그 경우다.
  2. **검사는 VCS에 의존하지 않아야 한다.** 작업 트리의 파일 존재만 보는 순수 판정이어야
     단위 테스트가 성립한다(§5).

  `.github/workflows/spec-lint.yml`에 `fetch-depth` 설정이 없는 것은 사실이지만(grep rc=1),
  그 워크플로는 이 검사의 실행 주체가 아니다 — 관측으로만 남기고 이 검사의 근거로 쓰지
  않는다.

- **R-2 — 권위 집합은 범프 커밋 한 건에서 쟀다.** `61921f1ba` 하나이며, 그 커밋도 함께
  빠뜨린 사이트가 있다면 이 7파일에 나타나지 않는다. `docs-site/hugo.toml`이 바로 그 모양의
  선례다. **이 카드는 이 위험을 닫지 않는다** — 누락 방향이 §4에서 명시적으로 범위 밖이다.
  t392가 닫을 후보다.

- **R-3 — 버전 토큰이 아닌 스탬프는 목록에만 남는다.** `hugo.toml:56`의 `releaseDate`는 범프
  커밋이 매번 바꾸지만 버전 토큰이 아니다. 이 카드는 문서의 스탬프 목록에 「날짜 필드도 같이
  바뀐다」를 명기하는 것으로 그치며, 어떤 검사도 이를 보지 않는다. t392의 토큰 술어도 보지
  못한다 — 별개의 날짜형 술어가 필요하다.

- **R-4 — 검사는 파싱에 의존한다.** 검사는 마크다운 목록에서 경로를 뽑아낸다. 문서의 서식이
  바뀌면(불릿 기호 변경, 경로를 백틱으로 감싸기 등) 파싱이 0건을 반환할 수 있고, 그때
  「위반 0건」은 통과가 아니라 **비실행**이다. REQ-VSG-005가 이 구멍을 막으려 세운 요구이며,
  그것이 이 카드에서 비공허성 요구가 살아남은 이유다.
  근거: `.claude/rules/moai/development/verification-completeness.md` §1.1 — 훑은 집합이 빈
  통과는 아무것도 단언하지 않는다.

- **R-5 — 검사가 조용히 멈출 수 있다.** 목록의 소제목 이름이 바뀌면 파싱 대상이 사라진다.
  REQ-VSG-005의 개수 단언이 0건 경우를 실패로 잡지만, 소제목만 바뀌고 항목 수가 그대로면
  기대값도 함께 고쳐야 하며 그 동기화는 사람의 몫으로 남는다.
  근거: 같은 규칙 §1.3 — 실행되지 않은 검사의 침묵은 성공과 구별되지 않는다.

---

## §8 참조

- `.moai/reports/t388/baseline.md` — 트리 `9328a5242`에서 잰 기준선
- `.moai/reports/t388/plan-audit.md` — plan-audit iter-1(FAIL 0.77). D2가 카드 분할의 계기
- `.moai/reports/t388/plan-audit-iter2.md` — plan-audit iter-2(FAIL 0.80). D1이 0.4.0의 주 수리
- `.moai/docs/version-management.md` — 수정 대상 문서(제목 66행, 항목 71-74·77-78행)
- `internal/cli/deprecated_paths_text_reference_test.go` — 문서 텍스트 가드 선례(§5)
- `internal/cli/hook_flush_test.go:22` — `repoRootFromCLITest` 헬퍼
- `.github/workflows/ci.yml:208` — `go test ./...` 실행 주체
- `Makefile:20,36,72` · `.goreleaser.yml:22` — Version 주입 경로(§1.1)
- `.claude/rules/moai/development/verification-completeness.md` §1.1 · §1.3 · §2 — 빈 스윕 ·
  조용한 비실행 · 두 셀 채택 규율
- 범프 커밋: `61921f1ba`(v3.1.4, 7파일 9줄) · `eba919e44`(v3.1.3, 6파일, hugo.toml 누락)
- 카드 t392 — 토큰 술어 가드(Tier M, SPEC 미발행, 이 카드 착지에 의존)
