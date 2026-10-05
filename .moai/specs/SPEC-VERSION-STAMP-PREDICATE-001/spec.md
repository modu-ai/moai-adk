---
id: SPEC-VERSION-STAMP-PREDICATE-001
title: "권위 토큰 술어와 정확-경로 등록부 — 두 방향을 닫고, 닫지 못한 것들을 열린 채로 이름 붙인다"
version: "0.3.0"
status: completed
created: 2026-09-01
updated: 2026-09-01
author: manager-spec
priority: Medium
phase: "v3.1.5 target"
module: internal/cli, .moai/docs
lifecycle: spec-anchored
tags: "version-management, release, documentation-rot, stamp-registry, partial-guarantee"
tier: M
---

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 |
|---|---|---|---|
| 0.1.0 | 2026-09-01 | manager-spec | 최초 작성(카드 t392, plan-phase). `.moai/reports/t392/baseline.md`와 이 세션의 재측정(트리 `9a3e2dabe`) 위에 세움. 선행 카드 t388(`SPEC-VERSION-STAMP-GUARD-001`, `status: completed`)의 §2·§4·§7을 상속한다. 요구 11 / 수락 11, Tier M |
| 0.2.0 | 2026-09-01 | manager-spec | iter-2 수리. `.moai/reports/t392/plan-audit.md`(FAIL 0.79, Tier M 문턱 0.80)의 D1-D8에 답한다. 요구 11 → **14** / 수락 11 → **14**(Tier M 상한 16/16 이내; 신설 셋 = REQ/AC-VSP-012 모집단 · 013 유령 가드 · 014 유지 계약). 변경: **D1** 스윕 모집단을 「추적 파일 집합」으로 **결정**하고 드라이버가 `git ls-files` 한 번으로 얻게 했다 — 34/28/121/155가 처음부터 그 모집단에서 측정됐으므로 재도출이 아니라 귀속이 성립한다(REQ-VSP-010 축소). **D2** 범프가 스탬프 7개만 건드린다는 것을 numstat으로 재고(`git show --numstat eba919e44` → 6파일 전부 스탬프), 그 결과 **「스윕 개수 = 28」 상수가 애초에 틀렸음**을 확인해 제거했다. 비공허성은 「등록부 28 · 스탬프 7 · 스윕 ⊇ 스탬프 집합」으로 옮겼고 셋 다 범프에 불변이다(REQ-VSP-005 재작성 + REQ-VSP-013 유지 계약). **D3** §3을 2×2에서 3축 열거로 다시 그리고 모집단을 명시했으며 `stamp`/`prose` 판별식을 REQ-VSP-002에 넣었다. **D4** `plan.md` §E의 닫힌 「셋」을 연 형태로 바꾸고 두 항목을 더했다. **D5** M4→M5 파손은 D2가 상수를 없애면서 **소멸**했다(§D에 순서-무관 근거 기록). **D6** §2.3 수치를 트리 SHA에 못박고 이 카드 자신의 산출물이 커밋될 때의 델타를 미리 선언했다. **D7** 유령 항목 가드를 REQ-VSP-013으로 승격(선택 아님 — 등록부가 스윕의 상위집합이 되면서 삭제된 서술 경로를 잡는 유일한 단언이 됐다). **D8** REQ-VSP-008을 행동 서술로 고쳤다. 반증한 감사 주장 1건: 미추적 토큰 보유 파일은 6이 아니라 **7**이며 전부 이 카드 자신의 산출물이다(`comm -23` 실측) |
| 0.3.0 | 2026-09-01 | manager-spec | iter-3 수리. **Tier M 2회 상한을 운영자가 이번 한 라운드에 한해 해제해 성립한 반복이다** — 통상 경로가 아니다. `.moai/reports/t392/plan-audit-iter2.md`(PASS-WITH-DEBT 0.84)의 N1-N6·D6에 답한다. 요구 14 → **15** / 수락 14 → **15**(상한 16/16 이내; 신설 = REQ/AC-VSP-015 스윕 도달 범위). 변경: **N1** iter-2의 뺄셈이 스윕 도달 범위의 하한을 함께 없앴다. 크기 리터럴을 넣지 않고 **양변을 실행 시점에 얻는** 단언 둘로 닫는다 — 「등록부 경로 전부가 드라이버 모집단 안에 있다」와 「코어가 판정한 경로 수 = 드라이버가 넘긴 경로 수」(REQ-VSP-015). 감사가 제시한 `git ls-files | wc -l` = 10,048 리터럴은 **채택하지 않는다**: 범프 불변이지만 개발 불변이 아니어서 D2가 제거한 결함 부류를 그대로 되살린다(§6.1, `plan.md` AP-11). **N2·N3은 한 뿌리다** — 영문 문서에 한국어 문안을 못박고 있었다(`grep -cP '[가-힣]' .moai/docs/version-management.md` → 0). `plan.md` §E를 **영문으로 재작성**하고 판정 grep 여섯을 그 영문에만 존재하는 구절로 다시 못박았으며(현재 전부 0건 = 미리 못박은 RED), 닫힌-수 정규식을 구조형으로 바꿔 **뮤턴트 8개**(영문 6 · 숫자형 1 · 한국어 구형 1)에 전부 걸리고 열린 문안에 0건임을 실측했다. **N4** §3 4행을 「파일이 실재하는 항목」으로 한정하고 항목-상태 두 가지(유령 항목 → REQ-VSP-013, 제외 그룹 안 등록 → REQ-VSP-015)를 표 아래 주로 붙였다. **N5** AC-VSP-010(b)를 거부 목록에서 **허용 목록**으로 뒤집었다(REQ-VSP-010의 형태와 일치). **N6** AC-VSP-012 양방향을 `git add` 대신 합성 모집단 둘로 바꿨다(§6 순수 코어 설계와 일치). **D6** §2.3 델타에서 정수 예측을 걷어내고 「이 카드가 커밋하는 산출물 1개당 1행, 병합 시점에 잰다」로 자기 유지형으로 바꿨다 — 예측이 세 번째로 스스로 낡았다. **선택 항목 채택**: §3.1(나)에 서술 21항목의 소비자 테스트 0건 실측을 기록했다 |

---

## §1 문제 — 목록 밖에 살아 있는 스탬프 사이트

`.moai/docs/version-management.md`의 「Files Requiring Version Sync」는 t388이 고쳐 놓았다.
유령 항목이 빠졌고 누락 4경로가 들어갔으며, `internal/cli/version_sync_list_test.go`가
「목록이 없는 경로를 가리키면」 실패한다. 그 보장은 **목록 안만** 본다.

목록 밖에 스탬프 사이트가 하나 더 있다. `internal/cli/testdata/` 의 golden 픽스처 6개다.

    internal/cli/testdata/{status,doctor}-{dark,light,nocolor}.golden

이 파일들은 `version.Version` 의 실행 시점 값을 그대로 찍는다 —
`status_golden_test.go` 와 `doctor_golden_test.go` 는 `version.Version` 을 pin 하지 않는다
(대조: `version_test.go:180-186` 은 `version.Version = "v0.0.0-test"` 로 pin 하고 defer 로 되돌린다).
따라서 범프가 이 6개를 다시 찍지 않으면 `go test ./internal/cli/...` 가 깨진다.
**스탬프 사이트다.** 그런데 t388 목록에 없다.

같은 모양이 이미 반복됐다. `b37e86b64`(`test(cli): stamp doctor/status golden fixtures at v3.1.3`)
는 범프 이후의 사후 수리였고, `61921f1ba`(v3.1.4 범프, 7파일 9줄)에는 golden 이 없다.
그 결과가 **병합되지 않은 `origin/release/v3.1.4` 에 지금 남아 있다**:

    git show origin/release/v3.1.4:pkg/version/version.go            → Version = "v3.1.4"
    git show origin/release/v3.1.4:internal/cli/testdata/status-nocolor.golden:6 → moai-adk v3.1.3

t388 §1의 `hugo.toml` 사고와 같은 모양이 다른 사이트에서 진행 중이다.

### §1.1 [HARD] 그러나 그 트리는 이 검사의 RED이 아니다

운영자 지시는 `origin/release/v3.1.4` 를 「도달 가능한 RED 후보」로 지목했다. **이 전제는 틀렸고,
틀린 이유가 이 SPEC의 설계를 결정한다.**

이 검사의 술어는 **권위 토큰**이다. `origin/release/v3.1.4` 에서 권위 토큰은 `v3.1.4` 이고,
낡은 golden 은 `v3.1.3` 을 담고 있다. 그러므로 그 트리에서 golden 은 **스윕에 걸리지 않는다** —
검사는 아무 말도 하지 않는다. 그 트리는 드리프트를 **보여주지만** RED을 **만들지 않는다**.

이것은 운영자 지시 자신이 1(d)로 지목한 성질이다: 「낡은 스탬프는 정의상 옛 토큰을 담으므로
현재-토큰 스윕에 나타나지 않는다」. 그 성질이 등록된 스탬프에만 적용되는 것이 아니라
**미등록 사이트에도 똑같이 적용된다**는 것이 여기서 드러났다.

완결성 단언의 RED은 **지금 이 트리(`9a3e2dabe`)** 에 있다. 여기서 권위 토큰은 `v3.1.3` 이고
golden 6개가 그 토큰을 담고 있으며 등록부에 없다. 실측:

    git grep -lF v3.1.3 -- . <거부 목록> | grep testdata   → 6 파일

도달 가능성 확인도 함께 적는다. `origin/release/v3.1.4`(`26898312e`)는 이 워크트리에서
`git show` 로 읽히지만 HEAD의 조상이 **아니다**(`git merge-base --is-ancestor` rc=1).
CI의 `fetch-depth: 0` 은 체크아웃된 ref의 이력만 가져오므로 그 tip은 CI에 들어오지 않는다.
t388 §7 R-1의 도달 불가 문제는 **재발한다**. 다만 위 이유로 애초에 그 트리를 RED으로
쓰지 않으므로 이 SPEC은 그 문제에 걸리지 않는다.

---

## §2 술어 — 권위 토큰 하나, 정확 경로 등록부

운영자 판정(재론하지 않는다): 판별식은 **권위 버전 토큰**이며, 등록부는 **정확 경로**다.
글로브도 디렉터리 접두사도 쓰지 않는다.

권위 토큰의 출처는 `pkg/version/version.go:8` 의 `Version` 기본값이다(트리 값 `v3.1.3`;
최고 태그는 `v3.1.2` — 결번 상태다). 이 값이 무엇인지는 t388 §1.1이 이미 확정했다:
파생값이 아니라 **ldflags 부재 시의 폴백**이며 손으로 유지된다.

### §2.0 [HARD] 모집단 — 스윕이 읽는 것은 **추적 파일 집합**이다

술어를 정하는 것과 **무엇에 대고 술어를 적용하는가**를 정하는 것은 다른 결정이다. 초판은
후자를 정하지 않은 채 수치를 `git grep`(추적 파일)으로 재고 드라이버를 `filepath.WalkDir`
(파일시스템)로 적었다. 두 집합은 같지 않다. 여기서 **결정한다.**

**스윕의 모집단은 이 저장소가 담고 있는 추적 파일 집합이다.** 드라이버는 그 목록을
`git ls-files` 한 번으로 얻고, 각 경로의 **내용**을 읽어 술어를 적용한다.

사유는 두 층이다.

**(1) 의미상 옳다.** 이 검사가 막으려는 결함은 「범프 커밋이 저장소가 담은 파일 하나를 다시
쓰지 못했다」이다. 추적되지 않는 파일은 어떤 범프 커밋의 의무 대상도 아니고, 다른 체크아웃에
존재하지도 않으며, 남의 테스트를 깨뜨리지도 않는다. **미추적 파일은 스탬프 사이트가 아니다**
— 정의상 그렇다(§7에 범위 밖으로 명시).

**(2) 파일시스템 walk는 판정 트리를 측정 트리와 갈라놓는다.** 취하지 않은 선택지의 결과를
실측으로 적는다(트리 `9a3e2dabe`, 이 세션):

    grep -rlF --exclude-dir=.git v3.1.3 . | wc -l   → 162      (git grep: 155)
    comm -23 <파일시스템 목록> <추적 목록>            → 7 파일

그 7개는 **전부 이 카드 자신의 산출물**이다(`.moai/reports/t392/` 3 + 이 SPEC 디렉터리 4).
전부 제외 그룹 안이라 오늘은 34가 우연히 일치하지만, 그것은 운이지 성질이 아니다. 이
워크트리를 벗어나면 다르다:

    find /Users/goos/MoAI/moai-adk-go/.claude/worktrees -maxdepth 1 -mindepth 1 -type d | wc -l  → 183
    grep -rlF v3.1.3 /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t379 | wc -l                 → 144

primary 체크아웃에서 walk는 gitignore된 형제 워크트리 183개로 내려가고, **그중 한 개만으로도
토큰 보유 파일이 144개**다. `.gitignore` 는 `.claude/worktrees/`(200행) ·
`.moai/worktrees/`(216행) · `bin/`(12) · `dist/`(27) · `/docs-site/public/`(295) ·
`node_modules/`(274)를 감춘다 — `bin/` 과 `docs-site/public/` 은 primary에 **실재하며** 오늘
권위 토큰 보유는 0이지만, 그것은 지금 무엇이 빌드돼 있는가의 성질이지 설계의 성질이 아니다.

walk를 택하면 그 뿌리들을 prune 목록으로 열거해야 하고, **그 목록이 곧 썩는 두 번째 등록부**가
된다 — `.gitignore` 에 새 빌드 산출 디렉터리가 하나 추가되는 순간 조용히 스윕에 들어온다.
이 카드가 존재하는 이유가 바로 그 결함 부류이므로, 그것을 재생산하는 설계를 택하지 않는다.

**비용: REQ-VSP-010이 좁혀진다.** 초판은 「git 호출 없음」이었다. 이제는 「이력 없음 · 체크아웃된
색인과 작업 트리 외의 어떤 ref도 조회하지 않음 · 네트워크 없음」이다. `git ls-files` 는 셋 중
어느 것도 위반하지 않는다 — 현재 체크아웃의 색인만 읽는다. 이 축소가 지키려던 두 성질은
그대로다: (a) 도달 불가 브랜치 의존이 원천 차단되고, (b) 판정부가 순수 함수로 남아 합성 입력
RED이 가능하다(git 호출은 얇은 드라이버에만 있다).

색인에 있으나 작업 트리에 파일이 없는 항목(리베이스 중 등)은 스윕이 건너뛴다. 일시적 상태이며
§8 R-8에 잔여 위험으로 남긴다.

### §2.1 면적 — 카드 수치를 쓰지 않는다

카드는 592파일 / 2225줄(base `9328a5242`)을 근거로 Tier M을 주장했다. 이 세션이 트리
`9a3e2dabe` 에서 다시 쟀다:

    git grep -nE 'v[0-9]+\.[0-9]+\.[0-9]+' -- . ':!.moai/reports' ':!.moai/specs' \
      ':!.moai/release-notes' ':!CHANGELOG.md' ':!*_test.go' ':!docs-site/content/*/changelog*'
    → 2229 줄 / 595 파일

    git grep -lF v3.1.3 -- <같은 거부 목록>   → 34 파일   (전체 트리 기준 155)

**595는 이 SPEC의 면적이 아니다.** 그것은 「모든 `vX.Y.Z` 토큰」을 술어로 잡았을 때의 값이고,
t388 §2가 이미 그 술어로는 판별자가 성립하지 않음을 실측으로 보였다(`v3.0.0` 270출현이
낡은 스탬프이면서 동시에 옳은 역사 서술이다). 술어를 권위 토큰으로 좁히면 34다.

### §2.2 34 파일의 분류 (트리 `9a3e2dabe`, 실측)

| 부류 | 수 | 내용 |
|---|---|---|
| 스탬프 | 7 | README×4 · `.moai/config/sections/system.yaml` · `docs-site/hugo.toml` · `pkg/version/version.go` |
| **미등록 스탬프** | **6** | `internal/cli/testdata/{status,doctor}-{dark,light,nocolor}.golden` |
| 서술 | 20 | docs-site 5문서 × 4로케일(`faq` · `statusline` · `claude-cloud` · `codex-dual-harness` · `moai-feedback`) — 예시 출력 인용 |
| 서술 | 1 | `.moai/docs/version-management.md` — t388이 남긴 「닫지 못한다」 서술 |

golden 6개는 §6의 pin으로 술어에서 **제거**된다(등록이 아니라 제거다 — 운영자 판정 2).
그러면 등록부는 **28항목**이 된다: 스탬프 7 + 서술 21.

### §2.3 [HARD] 거부 목록은 한 조항이 아니라 열거다

「거부 목록이 감추는 것」을 통째로 「릴리스 부산물」로 뭉뚱그리는 것이 t388이 이름 붙인 첫
안티패턴(포괄 면제)이다. 그룹마다 사유가 다르고 감추는 양도 다르므로 따로 적는다.
수는 전부 권위 토큰 `v3.1.3` 기준 **파일 수**이며, 각 그룹을 개별 명령으로 쟀다
(`git grep -lF v3.1.3 -- <그룹> | wc -l`).

[HARD] **아래 수는 상수가 아니라 시점 관측이다. 트리 `9a3e2dabe` 에 못박는다.** 검사는 이
수들을 하나도 들고 있지 않다 — 검사의 상수는 §4 REQ-VSP-005의 둘(28 · 7)뿐이고, 이 표는
REQ-VSP-007이 요구하는 **문서 서술**이다. 표가 낡아도 검사는 울지 않으므로, 낡음을 알리는
것은 아래 두 장치다.

- **자기 유지형 델타.** [HARD] **정수 예측을 두지 않는다.** 규칙만 적는다: 이 카드가
  커밋하는 **산출물 파일 하나마다** 그 파일이 속한 제외 그룹의 수가 1 늘고 합과 전체 트리
  값도 그만큼 는다. `34` 는 움직이지 않는다 — 이 카드의 산출물은 전부 제외 그룹 안이다.
  실제 값은 **병합 트리에서** 잰다(§8 R-7).
  - **왜 정수를 걷어내는가**(iter-3). 0.2.0판은 「산출물 7개 → `.moai/reports` 57 → 60,
    합 121 → 128, 트리 155 → 162」로 정수를 예측했고, 그 예측은 **그것을 검사한 감사
    보고서 자신이 여덟 번째 산출물이 되면서** 그 자리에서 거짓이 됐다. 같은 일이 세 번
    반복됐다(iter-1 표 → iter-2 델타 → iter-2 델타를 무효화한 iter-2 보고서). 세 번째는
    우연이 아니라 형태의 문제다: **이 카드를 감사하는 행위 자체가 모집단을 늘리므로**,
    커밋 전에 고정한 정수는 원리적으로 낡는다. 규칙은 그 행위에 불변이고, 정수는 아니다.
- **재도출 방아쇠.** 어떤 그룹의 실측 수가 그 그룹의 **사유와 더는 맞지 않으면** — 특히
  `docs-site/content/*/changelog*` 가 0이 아니게 되면 — 그 조항은 「0을 감추는 조항」이
  아니게 되므로 표 전체를 다시 도출하고 사유를 다시 쓴다. 이 문장이 그 조항의 유일한
  재측정 계약이다.

| 제외 그룹 | 이 그룹만의 사유 | 감추는 파일 |
|---|---|---|
| `.moai/reports/` | 세션 측정 기록. 시점에 못박힌 관측이므로 범프가 갱신하면 오히려 위조가 된다 | 57 |
| `.moai/specs/` | SPEC 본문. 역사 서술이며 완료 후 불변이다 | 58 |
| `.moai/release-notes/` | 릴리스마다 **새로 집필**되는 산출물이지 다시 쓰이는 스탬프가 아니다(t388 §1.3) | 1 |
| `CHANGELOG.md` | 같은 이유. 범프 커밋 `61921f1ba` 의 numstat 에 없다 | 1 |
| `*_test.go` | 검사 자신의 머리 주석과 SPEC 프론트매터 픽스처. 갱신 의무가 없다 | 4 |
| `docs-site/content/*/changelog*` | 변경 이력 페이지 | **0** |

합 121. `121 + 34 = 155` 로 전체 **추적 파일** 값과 정확히 맞으므로 이 열거는 **완전**하다
(귀속 확인). 서로소임은 항등식에서 **추론한 것이 아니라** 따로 쟀다: 여섯 그룹의 합집합을 한
pathspec으로 재도 121이다(측정 주체는 iter-1 plan-audit,
`.moai/reports/t392/plan-audit.md` § What I re-measured and found sound — 이 세션이 다시
실행하지는 않았다). 항등식만으로는 겹침이 빈틈을 상쇄한 경우와 구별되지 않는다.

초판에 있던 `.git/` 행은 **삭제한다.** 모집단이 `git ls-files` 이므로 `.git/` 은 애초에 후보에
들어오지 않는다 — 제외 조항이 아니라 존재하지 않는 대상이다. 제외 그룹은 **여섯**이다.

두 가지를 함께 적는다.

- **`docs-site/content/*/changelog*` 은 오늘 아무것도 감추지 않는다**(0). 카드의 술어에서
  물려받은 조항이며 이 술어에는 무효다. 지우지 않고 남기되 「무엇을 감추는지 잰 적 없는
  조항」이 아니라 「0을 감추는 조항」으로 기록한다.
- **`*_test.go` 제외가 이 카드의 사고를 거의 감출 뻔했다.** 실측한 4개 중
  `internal/cli/mcp_project_root_test.go:38,259` 는 `phase: "v3.1.3"` 을 인라인 픽스처로
  들고 있다 — 갱신 의무가 없으므로 제외가 옳다. 그러나 golden 6개가 스윕에 보이는 유일한
  이유는 그것이 `testdata/*.golden` 이지 `*_test.go` 가 아니기 때문이다. 같은 픽스처가
  `_test.go` 안에 인라인으로 있었다면 이 SPEC의 술어는 그것을 보지 못한다(§8 R-2).

### §2.4 [HARD] `-h` 함정을 물려받는다

t388 §2가 [HARD]로 남긴 함정: `git grep -n` 출력에서 토큰을 뽑으면 **파일 이름 안의 버전
문자열이 매 매치 줄마다 함께 세어진다**. 거부 목록 범위 안에서 이름에 버전 토큰이 든 파일은
**여덟**이며, 여섯이 `.moai/release/` 에 있다 — 제외된 `.moai/release-notes/` 와 한 글자
차이인 **다른 디렉터리**다. 재측정(트리 `9a3e2dabe`, 8파일 확인).

이 술어에서는 함정이 **아직 물지 않는다**: 그 여덟은 전부 v2.x 토큰만 담아
`.moai/release/` 의 권위 토큰 파일 수는 실측 **0** 이다. 요구 REQ-VSP-006이 존재하는 이유는
지금 물기 때문이 아니라, `git ls-files` 가 그 여덟을 **경로 문자열로** 넘겨주므로 판정을
내용이 아닌 경로에 걸면 조용히 물기 시작하기 때문이다.

---

## §3 무엇이 닫히고 무엇이 안 닫히는가 — 축은 둘이 아니라 셋이다

초판은 이 절을 (등록됨 × 토큰 보유) 2×2로 그렸다. **그 격자는 과잉 주장이었다.** 등록/스탬프
칸을 통째로 REQ-VSP-004에 귀속시켰는데, REQ-VSP-004는 `prose` 분류를 신선도 판정에서
명시적으로 **면제**한다 — 28항목 중 21이 `prose` 이므로 그 칸은 7/28에서만 닫힌다. 분류를
등록에 접어 넣은 것이 과잉 주장의 원인이었고, 접힌 축이 격자 **밖에** 다섯째 경우를 남겼다.

축은 셋이다: **등록됨 · 분류 · 권위 토큰 보유**.

**모집단을 먼저 적는다.** 아래 표는 §2.0의 추적 파일 집합에서 §2.3의 여섯 제외 그룹을 뺀
나머지에만 적용된다. 제외 그룹 안의 파일은 어느 행에도 들지 않는다 — 그 침묵은 검사가
「괜찮다」고 말한 것이 아니라 **보지 않았다**는 뜻이다(§8 R-2).

| # | 등록됨 | 분류 | 권위 토큰 | 누가 잡는가 | 오늘의 실례 |
|---|---|---|---|---|---|
| 1 | 예 | `stamp` | 담음 | (정상) | 스탬프 7개 |
| 2 | 예 | `stamp` | 안 담음 | **REQ-VSP-004** | `eba919e44` 의 `hugo.toml` 모양 |
| 3 | 예 | `prose` | 담음 | (정상) | docs-site 서술 페이지 |
| 4 | 예 | `prose` | 안 담음 | (정상 — 면제) | 옛 릴리스를 인용한 서술. **파일이 실재하는 항목에 한한다** |
| 5 | 아니오 | — | 담음 | **REQ-VSP-003** | golden 6개의 오늘 모양 |
| 6 | 아니오 | — | 안 담음 | **아무도 잡지 못한다** | `origin/release/v3.1.4` 의 golden (§1.1) |

[HARD] **위 여섯 행은 「파일의 상태」이지 「등록부 항목의 상태」가 아니다.** 세 축이 전부
파일의 성질이므로, 파일이 없는 항목이나 모집단 밖의 항목은 어느 행에도 들지 않는다 —
그런데도 등록부 안에는 있다. 그 두 **항목-상태**는 표가 아니라 아래가 받는다. 4행을 「파일이
실재하는 항목에 한한다」로 한정한 것이 이 구분을 표 안에 남긴 자리다.

| 항목-상태 | 오늘 존재하는가 | 누가 잡는가 |
|---|---|---|
| 등록됐으나 작업 트리에 **파일이 없다**(유령 항목) | 아니오 — 삭제·이동으로 생긴다 | **REQ-VSP-013** |
| 등록됐으나 그 경로가 **제외 그룹 안**이라 모집단에 없다 | 아니오(잠재). 오늘 28항목은 전부 §2.3 거부 목록을 적용한 측정에서 나왔으므로 구성상 모집단 안이다 | **REQ-VSP-015** |

둘째 상태를 이름 붙이는 이유: 그런 항목이 `stamp` 로 분류되면 스윕에 영영 들어오지 못해
REQ-VSP-005의 「스윕 ⊇ 스탬프」가 **영구히** 실패한다. REQ-VSP-015가 그 항목을 먼저,
경로 이름으로 부르며 잡는다.

### §3.1 [HARD] 닫지 못하는 것 둘 — 닫히지 않은 채로 이름을 붙인다

**(가) 6행 — 미등록 + 낡음.** 등록부는 그 파일을 모르고, 스윕은 권위 토큰만 보므로 옛 토큰만
남은 파일을 지나친다. 이 칸이 **왜 좁은가**: 새 스탬프 사이트는 태어나는 순간 보통 현재
토큰을 담아 5행에서 시작하고, `go test ./...` 가 CI 매 실행마다 도는 한 REQ-VSP-003이 그날
잡는다. 6행에 도달하려면 (a) 사이트가 처음부터 낡은 토큰을 달고 태어나거나, (b) **검사가 그
사이 돌지 않아야** 한다. (b)는 `verification-completeness.md` §1.3(계속 발화)의 소관이며
§8 R-1으로 남긴다.

**(나) 3·4행의 오분류 — 진짜 스탬프인데 `prose` 로 등록된 파일.** 격자 밖에 있던 다섯째
경우이며, **이것이 `hugo.toml` 사고의 모양 그대로다.** REQ-VSP-003은 통과한다(등록돼 있다).
REQ-VSP-004는 건너뛴다(`prose` 다). REQ-VSP-009도 통과한다(스탬프 집합에도 문서 목록에도
없으므로 두 집합이 여전히 같다). 세 단언 전부에 보이지 않는다.

이 경우의 **경계**를 적는다. 오분류된 스탬프에 그 값을 읽는 소비자 테스트가 있으면 범프
이후 **그 테스트가** 깨진다 — golden 6개가 정확히 그 모양이었고 `red-observation.md` §R.3이
그것을 실행으로 보였다. 소비자 테스트가 없으면 아무것도 깨지지 않는다 — `hugo.toml` 이
정확히 그 모양이었다. 그러므로 이 구멍의 크기는 「소비자 테스트가 없는 스탬프 사이트의 수」와
같고, 이 SPEC은 그 수를 재지 않았다.

**재지 않은 것과 별개로, 오늘의 침묵의 크기는 쟀다**(iter-3). 구멍의 **미래** 크기는 위
이유로 못 재지만, **오늘 등록부의 서술 21항목 중 소비자 테스트를 가진 것이 몇인가**는 두
명령이면 된다. 트리 `9a3e2dabe` 에서:

```
$ git grep -lF version-management.md -- '*.go'
internal/cli/version_sync_list_test.go
exit 0
$ git grep -lF 'docs-site/content' -- '*.go'
scripts/convert-nextra-to-hextra/main.go
scripts/docs-version-snapshot/main.go
exit 0
```

앞의 하나는 테스트지만 **버전 토큰이 아니라 「Files Requiring Version Sync」 목록을** 읽고,
뒤의 둘은 테스트가 아닌 릴리스 시점 도구다. 즉 **서술 21항목 중 범프 누락으로 깨질 소비자
테스트를 가진 것은 0개다.** §3.1의 틀로 읽으면 — 구멍의 크기 = 소비자 테스트가 없는 스탬프
사이트 수 — 오늘의 서술 집합 어디에서 오분류가 일어나도 **침묵은 최대**다. 이 수치는 위
「재지 않았다」를 뒤집지 않고 **날카롭게 한다**: 재지 못한 것은 미래의 크기이고, 잰 것은
오늘의 무음 정도다.

**닫지 않는 이유.** REQ-VSP-002의 판별식(「범프가 이 파일을 다시 써야 하는가」)은 사람이
읽어 적용하는 것이지 기계가 결정할 수 있는 것이 아니다. 기계적으로 닫으려면 범프를 흉내 내
전 스위트를 돌려야 하며 그것은 이 카드의 범위 밖이다. **이름만 붙이고 열어 둔다**(§7).

---

## §4 요구 (GEARS)

**REQ-VSP-001** — The check shall derive the authoritative version token from the `Version`
default in `pkg/version/version.go`, read from the working tree, and shall use exactly that
token as its sweep predicate. The check shall not sweep on the general `vX.Y.Z` shape.

**REQ-VSP-002** — An exact-path registry shall name every swept file carrying the authoritative
token, and shall classify each entry as either `stamp` or `prose`. An entry shall be classified
`stamp` when a version bump must rewrite that file, or a check that reads it breaks when the bump
does not; otherwise it shall be classified `prose`. The registry shall contain no glob, no
directory prefix, and no wildcard segment. The registry shall be permitted to name a file that
does not carry the authoritative token — a prose citation whose release has aged out remains a
correct entry — so the registry is a superset of the sweep result and never a subset of it.

**REQ-VSP-003** — When the sweep finds a file carrying the authoritative token that the
registry does not name, the check shall fail and shall name that path in its failure output.

**REQ-VSP-004** — When a registry entry classified `stamp` does not carry the authoritative
token, the check shall fail and shall name that path. Entries classified `prose` shall not be
judged for freshness, because a prose citation of an older release is correct as written.

**REQ-VSP-005** — The check shall hold, independently of both the sweep and the registry parse,
an expected total entry count and an expected `stamp`-classified entry count, and shall fail when
either differs from what it read. The sweep result shall contain every `stamp`-classified path;
when it does not, the check shall fail and shall name every missing path. A sweep that matched
nothing shall therefore be reported as a failure, never as a pass. The check shall not hold an
expected count for the sweep result itself, because the number of files carrying the
authoritative token changes at every version bump while the registry does not.

**REQ-VSP-006** — The sweep shall decide a file's membership from that file's content only. A
version token appearing in a file's name shall not admit the file to the sweep, and shall not
contribute to any count the check reports.

**REQ-VSP-007** — The sweep's exclusion set shall be a literal enumeration, and the version-sync
documentation shall state, for each excluded group, the reason that group alone is excluded and
the measured number of authoritative-token files it hides. A single clause covering the whole
exclusion set shall not stand in for the enumeration.

**REQ-VSP-008** — The golden-fixture tests shall render a fixed version token independent of the
build-time value, restoring the original value when the test ends, and the six fixtures shall be
regenerated so that they carry no version token. The fixtures shall be removed from the predicate
by this fixing rather than added to the registry.

**REQ-VSP-009** — The registry's `stamp`-classified path set shall equal the path set the
version-sync documentation names under its version-stamp heading. When the two sets differ, the
check shall fail and shall name every differing path.

**REQ-VSP-010** — The check shall read the checked-out index and the working tree only. It shall
query no history, no ref other than the checked-out one, and no network. Its judgment core shall
invoke no external process; obtaining the file population is the one place an external invocation
is permitted, and it shall live in the driver.

**REQ-VSP-011** — The version-sync documentation and this SPEC shall state which directions the
combined checks close, and shall enumerate cases they do not close as an open list that does not
claim to be exhaustive. The enumeration shall name at least: a file both absent from the registry
and carrying a non-authoritative token; a genuine stamp site classified `prose`; a stamp site
inlined inside a file the exclusion set hides; a date-shaped stamp; and a rendered version.
Neither document shall assert that the list can no longer rot.

**REQ-VSP-012** — The sweep shall read its file population from the set of files this repository
tracks. It shall not enumerate that population by walking the filesystem, because a filesystem
walk reaches paths the repository does not carry and the reachable set differs between checkouts.

**REQ-VSP-013** — Every registry entry shall resolve to a file present in the working tree. When
an entry does not resolve, the check shall fail and shall name that entry's path.

**REQ-VSP-014** — The version-sync documentation shall state who edits the registry, the event
that obliges an edit, and what the check reports between that event and the edit. It shall state
that a version bump obliges no registry edit.

**REQ-VSP-015** — The check shall assert that every registry path is present in the population the
driver returned, and shall fail naming every registry path the population does not contain. The
check shall additionally assert that the number of paths its judgment core examined equals the
number of paths the driver handed it, and shall fail reporting both numbers when they differ. The
check shall not hold an expected size for the population, because that size changes whenever a
file is added to or removed from the repository, which is unrelated to a version bump.

### §4.1 요구 ↔ 수락 추적

| 요구 | 수락 | 방향 / 성질 |
|---|---|---|
| REQ-VSP-001 | AC-VSP-001 | 술어의 출처 |
| REQ-VSP-002 | AC-VSP-002 | 등록부 형태(정확 경로) |
| REQ-VSP-003 | AC-VSP-003 | 완결성 — 미등록 + 현재 |
| REQ-VSP-004 | AC-VSP-004 | 신선도 — 등록 + 낡음 |
| REQ-VSP-005 | AC-VSP-005 | 비공허성 |
| REQ-VSP-006 | AC-VSP-006 | 내용 대 경로 |
| REQ-VSP-007 | AC-VSP-007 | 거부 목록 열거 |
| REQ-VSP-008 | AC-VSP-008 | pin + 픽스처 재생성 |
| REQ-VSP-009 | AC-VSP-009 | 등록부 ⇄ 문서 정합 |
| REQ-VSP-010 | AC-VSP-010 | 이력·타 ref 비의존 |
| REQ-VSP-011 | AC-VSP-011 | 부분 보장 서술(열린 열거) |
| REQ-VSP-012 | AC-VSP-012 | 스윕 모집단 = 추적 파일 집합 |
| REQ-VSP-013 | AC-VSP-013 | 유령 항목 가드 |
| REQ-VSP-014 | AC-VSP-014 | 등록부 유지 계약 |
| REQ-VSP-015 | AC-VSP-015 | 스윕 도달 범위(모집단 하한) |

요구 15 / 수락 15 — Tier M 상한(각 16)을 넘지 않는다
(`spec-workflow.md` § SPEC Complexity Tier). 전문은 `acceptance.md` §D.

---

## §5 Tier 도출 — 카드의 근거는 무효, 결론은 다른 근거로 성립

카드는 Tier M을 주장하며 595파일 / 2229줄을 들었다. **그 근거는 무효다** — §2.1이 보였듯
그것은 이 SPEC이 쓰지 않는 술어의 면적이고, 권위 토큰 술어의 면적은 34다.

측정된 범위에서 다시 도출한다.

| 산출 | 수 | 성격 |
|---|---|---|
| 새 검사 파일 `internal/cli/version_stamp_registry_test.go` | 1 | 저작. 순수 코어 + 모집단 드라이버(`git ls-files`) + 28항목 등록부 리터럴 |
| pin 편집 `status_golden_test.go` · `doctor_golden_test.go` | 2 | 저작(각 수십 줄) |
| 재생성 픽스처 `testdata/*.golden` | 6 | 데이터. `UPDATE_GOLDEN=1` 산출 |
| 문서 절 `.moai/docs/version-management.md` | 1 | 저작 |
| 계 | **10** | |

- **파일 수 10** → Tier S의 「5파일 미만」을 넘고 Tier M의 5-15 대역 안이다.
- **LOC** 은 저작분 300 안팎으로 S와 M의 경계에 걸린다. 판정을 가르는 것은 LOC이 아니라
  파일 수다.
- golden 6개를 「데이터라 세지 않는다」로 읽으면 저작 파일은 4개가 되어 Tier S에 들어간다.
  **그렇게 읽지 않는다** — 여섯은 diff에 나타나는 파일이고 AC-VSP-008의 판정 대상이다.

**Tier M.** 상한 16/16 안에 요구 15 / 수락 15로 든다. 산출물은 Tier M 3종
(spec.md · plan.md · acceptance.md) + 모든 Tier 공통의 progress.md.

iter-2 수리 뒤에도 이 도출이 **그대로 성립하는지**를 따로 확인한다. 파일 수 10은 변하지
않았다 — 수리는 요구 셋(012·013·014)을 더했지만 산출 **파일**을 더하지 않았다. REQ-VSP-012는
드라이버의 모집단 획득 방식을 바꾸고(같은 파일), REQ-VSP-013은 항목마다 `os.Stat` 한 번을
더하며(같은 파일), REQ-VSP-014는 이미 저작 대상인 문서 절에 문단 하나를 더한다. LOC은
저작분 350 안팎으로 여전히 S/M 경계에 걸리고, 판정을 가르는 것은 여전히 파일 수 10이다.
요구·수락 수 14/14는 Tier M 상한 16 아래이므로 tier-up 신호가 아니다. **Tier M 유지.**

iter-3 수리 뒤에도 같은 확인을 한다. 신설 REQ/AC-VSP-015는 **같은 검사 파일 안에** 단언 둘을
더할 뿐 산출 파일을 더하지 않으므로 파일 수는 여전히 **10**이고, 나머지 수리(N2·N3·N4·N5·N6·
D6)는 전부 `plan.md`·`acceptance.md`·`spec.md` 본문 편집이라 산출물에 손대지 않는다. 저작 LOC
는 350 안팎에서 크게 움직이지 않는다. 요구·수락 15/15는 상한 16 **바로 아래**다 —
[HARD] 이 카드에서 요구를 하나라도 더 늘려야 한다면 그것은 tier-up 신호이며, 늘리기 전에
분할을 먼저 검토한다. **Tier M 유지.**

---

## §6 검사의 거처와 등록부의 거처

**검사**: 새 파일 `internal/cli/version_stamp_registry_test.go`. t388의
`version_sync_list_test.go` 와 같은 패키지이므로 `repoRootFromCLITest`
(`internal/cli/hook_flush_test.go:22`)와 `parseVersionStampEntries`(REQ-VSP-009가 쓴다)를
그대로 재사용한다. 새 패키지를 만들지 않는다. `.github/workflows/ci.yml:208` 의
`go test ./...` 가 별도 job 없이 잡는다.

**등록부**: 문서가 아니라 **Go 리터럴**로 검사 안에 둔다. 사유 셋.

1. 문서의 「Files Requiring Version Sync」는 범프하는 사람이 읽는 **작업 지시**다. 서술 21경로를
   거기 적으면 실제로 손대야 하는 7개가 묻힌다.
2. t388의 검사가 이미 그 문서 목록을 읽는다. 문서는 역할을 유지한다.
3. Go 리터럴은 서식 변경으로 파싱 0건이 되지 않는다 — t388 §7 R-4의 구멍이 이 검사에는
   생기지 않는다.

두 곳에 스탬프 목록이 생기는 값을 REQ-VSP-009가 치른다: `stamp` 분류 집합과 문서 목록이
같아야 하며, 다르면 검사가 양쪽 차분을 이름으로 부른다.

[HARD] 검사의 판정부는 **순수 함수**로 짠다 — 입력은 (권위 토큰, 스윕 결과, 등록부,
경로별 내용, 문서 목록), 출력은 finding 목록. **모집단 드라이버**는 그 위의 얇은 층이며,
`git ls-files` 호출은 오직 거기에만 있다(REQ-VSP-010·012). 이 분리가 없으면
AC-VSP-004·005·006·009·013의 RED을 합성 입력으로 관측할 수 없다.

### §6.1 [HARD] 등록부의 유지 계약 — 범프는 등록부를 건드리지 않는다

초판은 이 계약을 적지 않았고, §8 R-4는 **틀린 방향의** 위험을 적으면서 미측정이라고 했다.
쟀다:

    git show --numstat --format='%h %s' eba919e44   → 6파일: system.yaml · README×4 · version.go
    git show --numstat --format='%h %s' 61921f1ba   → 7파일: 위 6 + docs-site/hugo.toml

**범프는 스탬프만 건드린다. 서술 21경로는 한 줄도 바뀌지 않는다.** 이 사실이 초판 설계에서
치명적이었던 이유를 그대로 적는다: 초판 REQ-VSP-005는 「스윕 개수 = 28」을 단언했는데, 범프
직후 서술 21은 옛 토큰을 그대로 들고 있으므로 스윕은 **7** 을 잡는다. 검사는 릴리스마다,
`sweep matched=7 expected=28` 이라는 **경로 하나 없는 개수**로 무너졌을 것이다.

그리고 그 상수는 깨지기 쉬웠던 것이 아니라 **애초에 틀렸다.** 서술이 옛 릴리스를 인용한
상태는 REQ-VSP-004가 면제로 인정하는 **정상**이다(§3 4행). 스윕 개수를 상수와 대조하는 것은
이 SPEC이 다른 자리에서 옳다고 말한 상태를 결함으로 판정하는 일이었다. 그래서 수리는 상수를
키우는 것이 아니라 **없애는 것**이다(REQ-VSP-005 재작성).

남는 상수 둘은 범프에 불변이다.

| 상수 | 값 | 왜 범프에 불변인가 |
|---|---|---|
| 등록부 항목 수 | 28 | 등록부는 파일 목록이지 토큰 목록이 아니다. 범프는 파일을 더하거나 지우지 않는다 |
| `stamp` 분류 수 | 7 | 범프는 이 7개를 **전부** 새 토큰으로 다시 쓴다. 세는 대상이 바뀌지 않는다 |

**계약.**

- **누가**: 등록부는 파일을 더하거나 지우는 **그 커밋의 작성자**가 함께 고친다. 별도 담당자를
  두지 않는다.
- **언제**: 범프 때가 **아니다**. 권위 토큰을 담은 파일이 새로 생기거나, 등록된 파일이
  삭제·이동될 때다.
- **그 사이 검사는 무엇을 말하는가**: 새 파일이면 REQ-VSP-003이 **그 경로를 이름으로** 부르며
  실패한다. 삭제·이동이면 REQ-VSP-013이 **그 항목을 이름으로** 부르며 실패한다. 둘 다 개수가
  아니라 경로를 낸다 — 사람이 상수를 깎아 조용히 넘어가는 값싼 수리 경로가 없다.
- **§6의 사유 1은 그대로 산다**: 범프가 등록부를 건드리지 않으므로 `version-management.md` 의
  「Files Requiring Version Sync」(범프하는 사람이 읽는 작업 지시)에 서술 21경로를 넣을 이유가
  없다. 문서는 **등록부 파일의 이름 하나**와 위 계약만 적는다(REQ-VSP-014).

이 계약이 감추지 않는 비용: 서술 페이지를 새로 쓰면서 현재 토큰을 인용하면 REQ-VSP-003이
그날 운다. 그 빈도는 릴리스 주기가 아니라 **문서 집필 빈도**에 비례하며, 이 SPEC은 그 값을
재지 않았다(§8 R-4).

Template-First 규칙은 걸리지 않는다 — 산출물이 `.claude/` 아래에도
`internal/template/templates/` 아래에도 없고, `.moai/` 아래 산출물은 `.moai/specs/` 와
`.moai/docs/` 뿐이다.

### §6.2 [HARD] 뺄셈이 함께 없앤 것 — 스윕의 도달 범위 (REQ-VSP-015)

「스윕 개수 = 28」을 없앤 것은 옳았지만(§6.1), 그것이 **스윕이 얼마나 넓게 닿았는가**에 대한
유일한 하한이기도 했다. 없앤 뒤 남은 단언들이 무엇을 묶는지 적으면: 28과 7은 **등록부의**
성질이라 스윕이 돌았는지 아무 말도 하지 않고, 「스윕 ⊇ 스탬프」는 7경로만 묶으며,
REQ-VSP-003은 위에서만 묶는다. **7과 34 사이가 비어 있었다.**

그 구간은 부수적이지 않다. 범프 직후에는 정상 스윕 결과가 스탬프 집합과 **정확히 같아지므로**
(측정: 범프 커밋 `61921f1ba` 트리에서 거부 목록 적용 스윕 = 7), 그 순간 검사는 「저장소 전체에
닿아 7을 찾았다」와 「거의 아무 데도 못 닿고 7을 찾았다」를 구별하지 못한다. 도달 가능한 버그도
있다: REQ-VSP-007이 제외 집합을 사람이 고치는 리터럴 열거로 두므로, 누가 `docs-site/` 한 줄을
넣으면 서술 20여 경로가 조용히 모집단에서 빠지고 네 단언이 전부 초록으로 남는다.

**채택하지 않은 수리 — 모집단 크기 리터럴.** iter-2 감사는 `git ls-files | wc -l` = 10,048 을
하한으로 제시했다. 범프에 불변인 것은 맞다(범프 커밋 둘 다 전부 `M`). **그러나 개발에는
불변이 아니다** — 파일을 하나 더하는 평범한 커밋마다 움직인다. 그 수를 리터럴로 박으면
검사는 무관한 커밋에서 **맨 숫자 하나로** 실패하고, 사람은 숫자를 새 값으로 깎아 넘어간다.
그것이 정확히 §6.1이 제거한 결함 부류다. **더 나쁜 형태로 되살리지 않는다**(`plan.md` AP-11).

**채택한 형태 — 양변을 실행 시점에 얻는 단언 둘.** 리터럴 크기 대신, 모집단이 **자기가 받은
것을 실제로 훑었는지**를 묻는다.

| 단언 | 좌변 | 우변 | 무엇을 잡는가 | 왜 범프·개발 양쪽에 불변인가 |
|---|---|---|---|---|
| **등록부 ⊆ 모집단** | 등록부 28경로 | 드라이버가 넘긴 경로 집합 | 잘린 드라이버 · cwd 로 좁혀진 `git ls-files` · 과잉 제외 그룹 | 양쪽 다 파일 추가/삭제 커밋에서 **함께** 움직인다(§6.1 계약이 같은 커밋에서 등록부를 고치게 한다). 범프는 둘 다 안 건드린다 |
| **판정 수 = 넘긴 수** | 코어가 훑은 경로 수 | 드라이버가 넘긴 경로 수 | 조용히 건너뛰는 코어(조기 반환 · 필터 버그) | 양변 모두 그 실행에서 나온다. 상수가 없다 |

첫째는 위에 적은 `docs-site/` 시나리오를 **경로 이름으로** 잡는다 — 빠진 서술 20여 경로가
등록부 항목이기 때문이다. 둘째는 드라이버가 옳게 넘겼는데 코어가 안 본 경우를 잡는다.
첫째는 §3 표 아래 「제외 그룹 안 등록」 항목-상태도 함께 닫는다.

[HARD] **남는 위험(닫지 않는다).** 등록부 경로만 담고 나머지를 버리는 드라이버는 두 단언을
모두 통과한다. 그런 드라이버는 손으로 그렇게 써야만 나오고 §6의 「모집단 획득은 한 자리」
설계가 그 자리를 한 줄로 좁혀 두었으므로 리뷰에 드러난다. **크기 하한 없이 남는 유일한
구멍이며, 여기에 이름을 붙여 열어 둔다**(§8 R-9). 제외 집합 리터럴은 이제 검사의 유일한
무가드 폭발 반경이다 — REQ-VSP-015가 그 반경을 등록부가 덮는 만큼 줄이지만 전부는 아니다.

---

## §7 범위 밖

### Out of Scope — 미등록 + 낡음 (§3 표 6행)

- 등록부에 없으면서 권위 토큰도 담지 않은 파일은 탐지하지 않는다(§3 표 6행). 후계자를 지정하지
  않는다 — 닫으려면 계보 전체(`v3.*` 이력 집합)를 술어로 잡아야 하고, t388 §2가 그 술어로는
  이 저장소의 버전과 남의 버전을 가를 수 없음을 실측으로 보였다. 열린 채로 둔다.
- 대신 §3.1(가)가 그 칸의 도달 경로를 좁혔고 §8 R-1이 계속 발화 의무로 받는다.

### Out of Scope — `prose` 로 오분류된 진짜 스탬프

- 등록은 돼 있으나 분류가 틀린 스탬프 사이트는 REQ-VSP-003·004·009 어느 것에도 걸리지
  않는다(§3.1 나). 이 SPEC은 **닫지 않는다.**
- 닫지 않는 이유: REQ-VSP-002의 판별식은 사람이 읽어 적용하는 것이고, 기계로 결정하려면
  범프를 흉내 내 전 스위트를 돌려야 한다. 그 비용은 이 카드의 범위 밖이다.
- 후계자: 미발행. §8 R-2에 열린 항목으로 기록한다.

### Out of Scope — 모집단 크기의 하한 [iter-3 신설]

- 검사는 모집단의 **크기**에 대한 기대값을 들지 않는다(REQ-VSP-015 후단). 「추적 파일이
  10,048개여야 한다」류의 리터럴은 넣지 않는다.
- 닫지 않는 이유: 그 수는 범프에는 불변이지만 **평범한 개발 커밋마다 움직인다.** 리터럴로
  두면 무관한 커밋에서 맨 숫자로 실패하고, 값을 깎는 값싼 수리 경로를 만든다 — D2가 제거한
  결함 부류를 되살리는 것이다(§6.2).
- 대신 두는 것: 양변을 실행 시점에 얻는 단언 둘(등록부 ⊆ 모집단 · 판정 수 = 넘긴 수).
  이 둘이 덮지 못하는 한 칸은 §8 R-9로 이름 붙여 열어 둔다.

### Out of Scope — 추적되지 않는 파일

- `git ls-files` 밖의 파일은 스윕의 모집단이 아니다(§2.0, REQ-VSP-012). 미추적 파일에 권위
  토큰이 있어도 검사는 아무 말도 하지 않는다.
- **그것이 옳다고 판단한 근거**: 저장소가 담지 않은 파일은 어떤 범프 커밋의 의무 대상이 될 수
  없고, 다른 체크아웃에 존재하지 않으며, 남의 테스트를 깨뜨리지 않는다. 스탬프 사이트라는
  개념 자체가 「범프 커밋이 다시 써야 하는 저장소의 파일」이므로 미추적 파일은 정의상 그
  대상이 아니다.
- 남는 위험: 새 스탬프 사이트를 만들고 `git add` 하지 않은 상태에서는 검사가 침묵한다. 커밋
  시점에 추적으로 바뀌고 그때부터 REQ-VSP-003이 본다 — 즉 **결함이 저장소에 들어오는 순간**
  부터 검사가 발화한다. 이 창을 닫지 않는다.

### Out of Scope — 날짜형 스탬프

- `docs-site/hugo.toml:56` 의 `releaseDate = "2026-08-24"` 는 범프 커밋이 매번 바꾸지만
  **버전 토큰을 담지 않는다**. 어떤 토큰 술어도 이것을 보지 못한다.
- 별개의 날짜형 술어가 필요하며 이 SPEC은 만들지 않는다. 문서에 「날짜도 함께 바뀐다」는
  t388의 표기를 유지하는 것으로 그친다.
- 후계자: 미발행. 열린 항목으로 §8 R-3에 기록한다.

### Out of Scope — 렌더되는 버전

- `internal/template/templates/.moai/config/sections/system.yaml.tmpl:6,9` 는
  `version: "{{.Version}}"` 로 **렌더**한다. 리터럴 버전 문자열이 없으므로 스윕에 걸리지
  않고, 걸려서도 안 된다.
- t388 §1.2가 이미 「스탬프가 아니다」를 사유와 함께 문서에 남겼다. 그 서술을 건드리지 않는다.

### Out of Scope — 범프의 기계적 자동화

- 버전 범프를 수행하는 명령·스크립트는 만들지 않는다. 운영자가 t388에서 기각한
  「범프 대상의 기계적 유도」를 다시 열지 않는다.
- 검사는 **어긋남을 보고**할 뿐 고치지 않는다.

### Out of Scope — 릴리스 절차

- `scripts/release.sh`, 태깅, GoReleaser, `hns-release-specialist` 는 건드리지 않는다.
- `origin/release/v3.1.4` 의 낡은 golden 을 이 SPEC이 고치지 않는다. 그 브랜치는 병합 시점에
  `develop` 을 흡수하면서 pin 을 함께 받는다.

### Out of Scope — t388이 세운 검사

- `internal/cli/version_sync_list_test.go` 의 유령 검사와 그 상수
  `expectedVersionStampEntries = 7` 은 그대로 둔다. 이 SPEC은 그 파일을 읽지 않고, 같은
  패키지의 헬퍼만 재사용한다.

---

## §8 잔여 위험

각 항목은 **관측한 것**과 **관측하지 않은 것**을 나눠 적는다.

- **R-1 — 검사가 조용히 멈추면 §3 표 6행이 넓어진다.**
  관측: `.github/workflows/ci.yml:208` 이 `go test ./...` 를 돈다(파일 판독). §3이 보였듯
  「미등록 + 현재」 칸에서 「미등록 + 낡음」 칸으로 넘어가는 유일한 통로가 **범프 사이에 검사가
  돌지 않는 것**이므로, 이 검사의 가치는 계속 발화에 전적으로 의존한다.
  관측하지 않음: 이 검사가 실제로 CI에서 실행되어 실패를 낸 기록은 아직 없다(run-phase의
  몫). 검사가 발화를 멈춘 것을 **무엇이 알려주는가**에 대한 답도 이 SPEC에는 없다 —
  `verification-completeness.md` §1.3이 요구하는 계속-발화 답을 만들지 못했고, 그 부재를
  부채로 기록한다.

- **R-2 — 제외 그룹이 미래의 스탬프 사이트를 감출 수 있다.**
  관측: `*_test.go` 제외가 감추는 권위 토큰 파일은 4개이며 전부 갱신 의무가 없다(실측,
  경로 열거 §2.3). golden 6개가 스윕에 보이는 것은 그것이 `testdata/*.golden` 이기 때문이다.
  관측하지 않음: 같은 픽스처가 `_test.go` 안에 인라인으로 있었다면 이 술어는 보지 못한다 —
  **그 반사실을 실행으로 확인하지 않았다.** 코드 판독으로만 판단했다. 이 SPEC은 이 구멍을
  닫지 않고 이름만 붙인다.
  같은 항목에 **오분류 구멍**을 함께 둔다(§3.1 나): 진짜 스탬프를 `prose` 로 등록하면 세 단언
  전부에 보이지 않는다. 관측: golden 6개는 소비자 테스트가 있어 범프 누락 시 그 테스트가
  깨졌다(`red-observation.md` §R.3, 실행). 관측하지 않음: 소비자 테스트가 **없는** 스탬프
  사이트가 이 저장소에 몇 개인지 — 즉 이 구멍의 실제 크기 — 는 재지 않았다. `hugo.toml` 이 그
  부류였다는 것만 안다.

- **R-3 — 날짜형 스탬프는 어떤 검사에도 걸리지 않는다.**
  관측: `hugo.toml:55-56` 에서 `version = "v3.1.3"` 바로 아래 `releaseDate = "2026-08-24"`
  가 있다(실측). 버전 줄은 REQ-VSP-004가 보고 날짜 줄은 아무도 보지 않는다.
  관측하지 않음: 날짜가 실제로 어긋난 채 릴리스된 전례가 있는지 커밋 이력으로 확인하지
  않았다. t388 §7 R-3의 서술을 그대로 물려받았을 뿐이다.

- **R-4 — 등록부 유지비는 릴리스가 아니라 문서 집필에 비례한다. [iter-2 재작성]**
  초판 R-4는 「서술 21경로가 범프에서 새 토큰으로 갱신되어 새 미등록 항목을 만드는지 재지
  않았다」고 적었다. **그 서술은 방향이 틀렸고, 측정은 `git show --numstat` 한 번 거리에
  있었다.** 쟀다: `eba919e44` 는 6파일, `61921f1ba` 는 7파일이며 **전부 스탬프**다 — 범프는
  서술을 **한 줄도** 갱신하지 않는다(§6.1). 따라서 초판이 걱정한 「갱신이 만드는 churn」은
  일어나지 않고, 대신 초판 REQ-VSP-005의 「스윕 = 28」 상수가 범프마다 무너졌을 것이다. 그
  상수를 제거한 것이 이 위험의 수리다.
  관측: 범프 커밋 둘의 numstat(위). 직전 토큰 `v3.1.2` 보유 파일 8개가 전부 v3.1.3 집합에도
  들어 있다(교집합 8, `v3.1.2` 전용 0 — 실측).
  관측하지 않음: 서술 페이지가 **새로 집필되면서** 현재 토큰을 인용하는 빈도. 그 값이 크면
  REQ-VSP-003이 자주 발화하고 사람이 등록부에 경로를 더하는 비용이 반복된다. 그 비용은
  릴리스 주기와 무관하며, 이 SPEC은 재지 않고 기록만 한다. 이것이 정확-경로 등록부
  (운영자 판정 1)의 값이다.

- **R-5 — 서술 항목은 조용히 노화하며, 그것을 결함으로 판정하지 않는다. [iter-2 재작성]**
  관측: 상수는 둘이고(등록부 28 · `stamp` 7) 둘 다 등록부의 성질이므로 범프에 불변이다
  (§6.1 표). 초판이 들고 있던 「스윕 = 28」 상수는 제거됐다.
  서술 항목이 낡아 스윕에서 빠지는 것은 이제 **아무 단언도 어기지 않는다** — REQ-VSP-002가
  등록부를 스윕의 상위집합으로 정의하고 REQ-VSP-004가 `prose` 를 면제하기 때문이다.
  관측하지 않음: 그렇게 노화한 서술 항목이 언젠가 전부 옛 토큰만 담게 되어도 등록부에
  그대로 남는다. REQ-VSP-013은 **파일이 사라진 것**만 잡지 **인용이 낡은 것**은 잡지 않는다.
  등록부가 죽은 무게를 쌓는 이 경로를 이 SPEC은 닫지 않고 받아들인다 — 닫으려면 「이 서술이
  아직 이 파일을 인용할 이유가 있는가」를 기계가 판정해야 하는데 그것은 문서 편집 판단이다.

- **R-8 — 색인과 작업 트리가 어긋난 순간이 있다. [iter-2 신설]**
  관측: 모집단은 `git ls-files`(색인)이고 술어는 파일 **내용**(작업 트리)에 적용된다
  (§2.0). 두 면이 같지 않은 순간이 존재한다 — 리베이스 중, 병합 충돌 중, 혹은 파일을 지우고
  `git rm` 을 아직 안 한 상태.
  관측하지 않음: 그 상태에서 검사를 실제로 돌려 보지 않았다. 설계상 색인에만 있는 항목은
  스윕이 건너뛰며, 그것이 완결성 단언을 조용히 약화시키는 크기는 재지 않았다. 일시적
  상태이고 CI는 깨끗한 체크아웃에서 도므로 CI 판정에는 닿지 않는다고 **판단**했을 뿐이다.

- **R-9 — 모집단에 크기 하한이 없고, 등록부가 덮지 못하는 손실은 남는다. [iter-3 신설]**
  관측: REQ-VSP-015의 두 단언은 **등록부 28경로가 모집단에 있는가**와 **코어가 받은 만큼
  훑었는가**를 묻는다. 등록부 경로만 남기고 나머지를 버린 모집단은 둘 다 통과한다
  (§6.2의 「남는 위험」).
  관측하지 않음: 그 통과 상태를 실제 뮤턴트로 만들어 확인하지 않았다 — 그런 드라이버를
  쓰려면 등록부를 드라이버 안으로 들여와야 하고, 그것은 §6이 금지한 자기 참조라 다른 단언에
  먼저 걸린다고 **판단**했을 뿐 실행으로 보이지 않았다. 크기 리터럴을 놓지 않기로 한 대가가
  이 한 칸이며, 그 대가를 치르는 이유는 §6.2에 적었다.

- **R-6 — 권위 토큰의 출처 자신이 스탬프다.**
  관측: `pkg/version/version.go:8` 은 스윕의 술어를 공급하면서 동시에 등록부의 `stamp`
  항목이다(실측, 34 목록에 포함). 그 줄이 잘못 고쳐지면 술어 전체가 조용히 이동하고,
  이동한 술어에 대해 트리는 대체로 깨끗해 보인다 — 스윕이 0건이 되고 REQ-VSP-005가 그때
  운다.
  관측하지 않음: 「잘못된 토큰으로 이동했을 때 스윕이 정확히 0이 된다」를 실행으로 확인하지
  않았다. 합성 입력 RED(AC-VSP-005)이 이 경로를 모델링하지만 실제 트리에서 재현하지는
  않았다.

- **R-7 — 이 SPEC의 측정은 병합 전 트리의 것이다.**
  관측: 모든 수치는 워크트리 `.claude/worktrees/t392`, HEAD `9a3e2dabe` 에서 쟀다.
  같은 세션에서 `origin/develop` 은 이미 `64bba61aa` 로 나아갔다(실측).
  관측하지 않음: 그 사이 커밋들이 권위 토큰 파일 집합을 바꿨는지 재지 않았다. run-phase는
  병합 트리에서 34·28·121 세 수를 **다시 재고**, 다르면 상수를 고쳐야 한다.

---

## §9 참조

- `.moai/reports/t392/baseline.md` — 트리 `9a3e2dabe` 기준선(§B.2 34파일 분류 · §B.4 두 방향)
- `.moai/specs/SPEC-VERSION-STAMP-GUARD-001/spec.md` — 선행 카드 t388. §2(술어 이관 근거 +
  `-h` 함정) · §4(절반의 보장) · §5.1(단언별 RED 규율) · §7(잔여 위험)
- `internal/cli/version_sync_list_test.go` — t388의 유령 검사. 헬퍼 재사용 대상
- `internal/cli/status_golden_test.go` · `internal/cli/doctor_golden_test.go` — pin 부재 지점
- `internal/cli/version_test.go:180-186` — pin 선례(`v0.0.0-test` + defer 복원)
- `internal/cli/hook_flush_test.go:22` — `repoRootFromCLITest`
- `.moai/docs/version-management.md` — 「Files Requiring Version Sync」(제목 66행) + 수정 대상
  부분 보장 서술
- `pkg/version/version.go:8` — 권위 토큰의 출처
- `docs-site/hugo.toml:55-56` — 버전 줄과 날짜 줄이 나란히 있는 자리
- `.github/workflows/ci.yml:208` — `go test ./...` 실행 주체
- `.claude/rules/moai/development/verification-completeness.md` §1.1 · §1.3 · §2 — 빈 스윕 ·
  계속 발화 · 두 셀 채택
- 커밋: `61921f1ba`(v3.1.4 범프, golden 없음) · `b37e86b64`(golden 사후 스탬프) ·
  `eba919e44`(v3.1.3 범프, hugo.toml 누락) · `26898312e`(`origin/release/v3.1.4` tip)
