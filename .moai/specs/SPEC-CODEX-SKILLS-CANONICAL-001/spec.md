---
id: SPEC-CODEX-SKILLS-CANONICAL-001
title: "Codex 듀얼 하네스 M1 — 배포 스킬을 .agents/skills/ 에도 노출해 Claude Code · Codex CLI 양쪽에서 보이게 한다"
version: "0.4.0"
status: draft
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/template
lifecycle: spec-anchored
tier: M
tags: "codex, skills, deployer, symlink, embed, dual-harness, cleanup"
---

# SPEC-CODEX-SKILLS-CANONICAL-001 — 스킬 카탈로그 이중 노출

## HISTORY

- 2026-08-22 (plan-phase, iter-4, v0.4.0) — 독립 감사 **두 건**(0.78 / 0.7625)이 수렴한 결함을 닫는다. 무게중심은 하나다: **실행 순서가 만드는 dangling 링크**(§A.10). `moai update` 는 clean→deploy 이고, 청소는 `ManagedCleanTargets` 를 순서대로 돌며 정본 글롭이 앞서므로, 미러를 처리할 시점에 모든 링크가 dangling 이다. 그런데 `backupThenRemove` 는 `os.Stat` 으로 판정해 `IsNotExist` 에서 조용히 `return 0, nil` 한다 — **`filepath.Glob` 은 매치하는데 제거는 일어나지 않는다.** 직접 재현했다(§A.10 출력). REQ-CSC-008 의 목적이 규정대로 구현해도 무효화되던 상태였고, 그것을 검출해야 할 MUST AC 두 개(008·009)는 fixture 가 **실 디렉터리**라 정확히 반대편만 보고 있었다. 요구사항 쪽은 `os.Lstat` 판정 + dangling 제거(본체) + 순서 배치(이중 방어)를 REQ-CSC-008 에 함께 걸고 §B.D6 에 "(b) 를 중복이라며 지우지 않는다"를 [HARD] 로 못박았으며, **`os.Lstat` 전환이 공유 코드라 기존 7개 청소 뿌리 전부의 동작을 바꾼다는 사실**을 부작용이 아니라 폭발 반경이 적힌 결정으로 기록했다. 판정 쪽은 AC-CSC-008 의 fixture 를 네 형태(살아 있는 링크 / dangling 링크 / 복사 모드 실 디렉터리 / 사용자 실 디렉터리)로 확장했다 — dangling 팔이 이 결함을 잡는 유일한 팔이다. 그 밖에 REQ-CSC-001 의 무조건 `shall` 에 예외 절을 달아 REQ-011·014 와의 모순을 해소했고(iter-2 N1), REQ-CSC-010 의 백업 금지를 "이번 실행이 다시 만들 미러"로 한정해 무백업 손실 경로를 닫았으며(iter-2 N2, §A.11 — 판별자는 `backupThenRemove` 주석에 이미 있던 기준을 그대로 쓴다), 출력 seam 을 REQ-CSC-005 에 반환값 형태로 확정했고(iter-2 N6 — `internal/template` 에 `io.Writer` 가 **없다**는 실측), `.gitignore` 범위를 `.agents/` 전체에서 `.agents/skills/moai*` 로 좁혔다(iter-2 N7, §B.D7). **접두 철자 오류를 정정한 것이 §A.9 의 가장 큰 변화다** — 코드 상수는 `moai-`(하이픈)이고 카탈로그에는 이름이 정확히 `moai` 인 스킬이 있어, iter-3 이 "불변식은 코드에 이미 있다"고 적은 것은 **틀렸다**. 코드에 있는 것은 다른 불변식이고 그쪽은 이미 깨져 있다. 본 SPEC 전체가 `moai`(하이픈 없음)로 통일되고 `EmbeddedMoaiSkillNames()` 는 어떤 집합 정의에도 쓰지 않는다. **예산: REQ 16 / AC 16 불변 — 은퇴시킨 항목 없음.** 모든 수정이 기존 번호 안의 절 추가·수정으로 처리됐다. optional 도 전부 반영했다(§A.2 조건 명시, §A.3 리스팅 예산 축, §D.4 Codex 제외 사유 정정, §F 경로 기준 명시, §F/§G 순서). 감사 두 건이 갈린 지점 — audit#1 의 N10/N11/N12(plan §F 가 spec 개정을 못 따라옴) — 은 **디스크 실물 확인 결과 iter-3 에서 이미 닫혀 있었고**, audit#2 가 같은 결론을 냈다. audit#1 이 개정 전 스냅샷을 읽은 것으로 보인다.
- 2026-08-22 (plan-phase, iter-3, v0.3.0) — plan-audit iteration 1(FAIL 0.775, Tier M 임계 0.80)의 blocking 8건을 닫는다. 감사가 지목한 것 중 **판정 계층의 구멍 세 개가 핵심**이었다. (1) AC-CSC-001 은 `fs.DirEntry.IsDir()` 가 `Lstat` 기반이라 디렉터리 링크가 **양쪽 집합에서 동시에 빠져** 등식이 유지되는 형태였다 — 자기가 막으려는 실패에서 통과했다. (2) AC-CSC-010 / AC-CSC-013 은 Go 테스트가 접근할 수 없는 "변경 전 커밋 기준선"을 참조했다. (3) REQ-CSC-010 은 기존 manifest seam 에서 **구현 불가능**했다. 이 셋 때문에 §A 에 실측 세 건(§A.6 manifest seam, §A.7 pre-clean 백업 비대칭, §A.8 `.gitignore`)을 추가하고, 요구사항을 12 → **16**(예산 상한), 판정을 15 → **16**(예산 상한)으로 늘렸다. 새 요구사항은 REQ-CSC-013·014(대상이 이미 존재하는 두 상태), REQ-CSC-015(배포 스킬 이름 `moai` 접두 불변식), REQ-CSC-016(미러는 사용자 저장소 버전 관리 대상 아님)이며, REQ-CSC-010 은 "manifest 에 기록한다"에서 "**기록하지도 백업하지도 않는다**"로 방향이 뒤집혔다(§A.6·§A.7 실측 결과). §D.7 의 미해결 항목 하나는 **틀린 방향을 보고 있었다** — 감사와 본 SPEC 이 각각 측정한 결과 링크 방향은 안전하고 위험은 복사 모드 쪽이었다. 감사가 정정한 §A.3 부분합(`optional-pack:*` 13, `harness-generated` 스킬 0)도 반영했다. 선택 항목 D12(REQ 원자성 분리)는 **의도적으로 기각**했다 — 사유는 §G.
- 2026-08-22 (plan-phase, iter-2, v0.2.0) — 청소 글롭의 **접두 한정**을 구속력 있는 형태로 고정했다. §B.D5 가 네임스페이스 분리 계약(`.moai/docs/harness-namespace-doctrine.md` §24.1)을 명시적으로 인용하고, `ManagedCleanTargets` 확장이 `moai update` 의 **동작 변경**임을 §A.5 에 기록했다. 판정 쪽에서는 AC-CSC-008 을 **양팔 단일 테스트**로 다시 썼다 — `moai-gone` 제거와 `hns-user-owned` 생존을 한 테스트 안에서 함께 단언한다. 제거만 단언하는 테스트는 사용자 소유 스킬을 조용히 지우면서도 통과하기 때문이다. AC-CSC-009 는 그 위에서 나머지 사용자 소유 네임스페이스(`harness-*` · `my-harness-*` · 임의 이름)로 범위를 넓힌다. 요구사항·판정 개수는 12/15 로 불변이며, 설계 방향(D1·D2·D3)은 리드 승인대로 유지한다.
- 2026-08-22 (plan-phase, iter-1, v0.1.0) — Tier M 최초 작성. 선행 실측 두 건(`.moai/reports/t91/README.md` = M0 Codex 전제, `.moai/reports/t81/m1-preflight-measurements.md` = M1 선행 관측)을 근거로 한다. 재설계 문서(`.moai/reports/moai-adk-dual-harness-codex-20260817.md`)가 M1 을 "**코드 변경 0, 최저 비용**"으로 분류한 것은 **거짓 전제**로 판정했다 — 근거는 §A.2 의 `//go:embed` 심볼릭 링크 무음 소실이다. 카드 본문의 "스킬 32개"와 선행 실측의 "36개"도 **둘 다 틀렸다**; 실측을 다시 수행한 결과 템플릿 배포 대상은 **34개**다(§A.1). 세 번째 정정은 본 SPEC 이 새로 발견한 것으로, 배포되는 스킬 수가 **카탈로그 tier 에 따라 달라진다**는 사실이다(§A.3) — 재설계 문서의 성공 지표 "Codex `/skills` 에 32개 노출"은 기본 슬림 init 에서 원리적으로 달성 불가능하다.

## §A. 검증된 기준선 (실측)

### A.1 스킬 인벤토리 — 카드 전제 정정

worktree `WT-skills-canonical`, HEAD 기준 `find` 실측:

| 집합 | 수 | 명령 |
|---|---|---|
| 템플릿 배포 대상 `internal/template/templates/.claude/skills/*/` | **34** | `find … -mindepth 1 -maxdepth 1 -type d \| wc -l` |
| 로컬 `.claude/skills/*/` | 44 | 동일 |
| 그중 dev-only `hns-*` | 10 | `-name 'hns-*'` |
| 템플릿 스킬 디렉터리 중 `SKILL.md` 보유 | 34 | `find … -name SKILL.md \| wc -l` |

44 = 34 + 10 으로 일관된다. 앞선 두 숫자(카드 32, 선행 실측 36)는 채택하지 않는다. 특히 36 은 `ls | wc -l` 이 셸 별칭 때문에 long-format 으로 실행되어 `.` 과 `..` 두 줄이 더해진 값으로 보인다(34+2, 그리고 44+2=46 도 같은 형태) — **줄 수는 항목 수가 아니다.**

`.agents/` 는 로컬·템플릿 양쪽에 존재하지 않는다. 신규 생성 대상이다.

### A.2 `//go:embed` 는 심볼릭 링크를 무음으로 버린다 — 설계를 구속하는 사실

선행 실측(finding B, 최소 재현 포함): `internal/template/templates/` 안에 심볼릭 링크(파일·디렉터리 모두)를 두면 `//go:embed all:templates` 결과 FS 에서 **사라지며, 빌드 오류도 경고도 발생하지 않는다**.

M0 이 확인한 "Codex 가 `.agents/skills/` 아래 심볼릭 링크를 따라간다"는 **배포된 사용자 프로젝트의 런타임 사실**이고, **빌드타임 임베드 사실이 아니다**. 두 사실을 혼동한 설계는 스킬 0개를 무음으로 배포한다.

**조건 명시**: 무음 소실은 **디렉터리 패턴 임베드**에 한정된 조건부 사실이다. 링크를 `//go:embed` 패턴에 **직접 지목**하면 무음이 아니라 `cannot embed irregular file` **빌드 오류**다. 본 프로젝트는 `//go:embed all:templates` — 전자에 해당한다. 회귀 가드(AC-CSC-001)가 필요한 이유가 정확히 이 조건에서 나온다: 빌드가 잡아주지 않는다.

**귀결**: 링크를 만드는 주체는 템플릿 트리가 아니라 **배포기**여야 한다. 따라서 M1 은 "코드 변경 0"이 아니다.

### A.3 배포되는 스킬 수는 tier 에 따라 달라진다 (신규 발견)

`internal/template/catalog.yaml` 실측: 스킬 카탈로그 항목 34개 중 **core 21개 / non-core 13개**이며, non-core 13개는 **전부 `optional-pack:*`** 이다(`grep -c 'tier: optional-pack'` → 13). `harness_generated` tier 의 **스킬은 0개**다 — 해당 블록은 `skills: []` 이고 그 tier 를 가진 유일한 항목은 에이전트 `builder-harness` 다. (iter-2 까지 이 부분합을 "12 + 1" 로 적었던 것은 오류다. 합계 13 과 §A.3 의 결론은 영향받지 않지만, 남의 계수를 정정하는 절 안에서 부분합이 틀린 것은 그 자체로 흠이므로 정정한다 — AP-7 은 이 문서에도 적용된다.)

`harness_generated.skills` 가 **비어 있는 슬롯**이라는 사실은 그냥 넘길 대목이 아니다. 여기 스킬이 하나라도 들어오는 순간 §B.D5 의 접두 한정 청소와 미러 집합이 어긋날 수 있다 — 그래서 REQ-CSC-015 가 필요하다(§A.9).

`slimFS`(`internal/template/slim_fs.go`)는 non-core 항목을 임베드 FS 수준에서 숨기고, `moai init` 은 **기본이 슬림**이다(`shouldDistributeAll`: `--all` 또는 `MOAI_DISTRIBUTE_ALL=1` 일 때만 전량). `moai update` 는 슬림을 거치지 않고 전량 FS 를 쓴다.

즉 사용자 프로젝트에 실제로 존재하는 스킬 수는 **21 또는 34**이며 실행 경로에 따라 다르다. 미러 집합을 상수로 박으면 슬림 프로젝트에서 **존재하지 않는 스킬을 가리키는 깨진 링크 13개**를 만든다.

한 축을 더 적어 둔다. 노출 개수는 tier 외에 **Codex 의 스킬 리스팅 예산**(재설계 문서 기준 컨텍스트 2% 또는 8,000자)에도 걸린다 — 링크를 34개 만든다는 것이 34개가 노출된다는 뜻은 아니다. REQ 는 "`.agents/skills/<name>/SKILL.md` 로 읽을 수 있는 접근 경로"까지만 요구하므로 본 SPEC 의 요구사항에는 영향이 없고, 예산 축의 실측은 범위 밖이다.

### A.4 청소 경로에 `.agents/` 가 없다

`ManagedCleanTargets`(`internal/cli/update/deploy/deploy.go`)가 반환하는 관리 대상은 7개이며 전부 `.claude/` 하위이다. `.agents/` 는 **없다**. 등록하지 않으면 은퇴·개명된 스킬이 사용자 프로젝트에 영구 잔존한다. 이는 가설이 아니라 이미 관측된 실패 형태다 — M0 실측에서 사용자 실환경 `~/.codex/skills/` 에 2026-06-07 자 구 moai 스킬이 다수 잔존했고, 그중 `moai-lang-*` · `moai-platform-*` 는 현재 카탈로그에 없는 이름이다.

### A.5 청소 목록 확장은 `moai update` 의 동작 변경이다

`ManagedCleanTargets` 에 항목을 더하는 것은 `moai update` 가 **지우는 대상을 늘리는** 일이다. 배포 로직 추가와 달리 이것은 되돌리기 어려운 쪽의 변경이므로, 근거를 여기 남긴다.

정당화는 이 변경이 막는 실패가 **이미 두 번 실측된 같은 형태**라는 점이다.

1. `.moai/config` 통째 삭제 사고 — 관리 대상 뿌리의 청소 규칙이 실제 파일 수명과 어긋났을 때 무엇이 사라지는지가 관측됐다(CLAUDE.local.md §2.3).
2. 사용자 실환경 `~/.codex/skills/` 의 구 moai 스킬 잔존 — 청소 경로가 **없을 때** 은퇴·개명된 스킬이 어떻게 누적되는지가 관측됐다(M0 실측, §A.4).

두 사고는 방향이 반대지만 뿌리가 같다: **청소 규칙과 배포 규칙이 같은 집합을 가리키지 않으면 조용히 어긋난다.** `.agents/skills/` 에 배포하면서 청소 목록에 넣지 않으면 2번이 재현되고, 넓게 잡으면 1번이 재현된다. 그래서 등록하되 **접두를 `moai*` 로 한정**한다(§B.D5).

### A.6 `manifest.Track` 은 디렉터리 링크에서 실패한다 — 실측

`manifest.Manager.Track(path, provenance, hash)` 는 내부에서 `HashFile(absPath)` 를 호출하고 실패하면 **error 를 반환한다**(`internal/manifest/manifest.go` `Track`). `HashFile` 은 `os.Open` + `io.Copy` 다(`internal/manifest/hasher.go`).

디렉터리를 가리키는 심볼릭 링크에 대해 이 조합이 어떻게 동작하는지 직접 실행했다(go1.26.x, darwin/arm64):

```
open err: <nil>
copy err: read lnk: is a directory
```

`os.Open` 은 링크를 따라가 디렉터리를 열어 **성공**하고, `io.Copy` 가 EISDIR 로 **실패**한다. 즉 `.agents/skills/moai-x` 를 링크로 만든 뒤 `Track` 을 부르면 error 가 올라오고, 호출부가 그대로 전파하면 `Deploy` 가 실패한다 — REQ-CSC-011(fail-open)과 정면으로 충돌한다.

**귀결**: 미러 항목을 manifest 에 기록하려는 설계는 이 seam 에서 성립하지 않는다. 방향을 뒤집어, 미러는 **기록하지 않는 것**을 요구사항으로 세운다(REQ-CSC-010). 미러는 정본의 파생물이고 정본은 이미 기록되므로 잃는 정보가 없으며, 청소는 manifest 가 아니라 글롭이 담당한다(REQ-CSC-008).

### A.7 pre-clean 백업은 링크 모드와 복사 모드에서 정반대로 동작한다 — 실측

`backupThenRemove`(`internal/cli/update/deploy/deploy.go`)는 대상이 디렉터리면 `templateManagedPaths(tmplFS, relTarget)` 를 구해 **템플릿이 안 가진 파일 전부를** pre-clean 백업 트리로 복사한 뒤 지운다.

- **링크 모드 — 안전.** `filepath.WalkDir` 는 루트를 `Lstat` 하므로 링크 루트에서 정규 파일을 0개 걷고, 복사는 비정규 항목을 건너뛰며, `os.RemoveAll` 은 링크만 지운다. 정본은 무사하다. (iter-2 의 §D.7 은 "링크를 따라가 정본을 중복 저장할 수 있다"를 미해결로 남겼는데, **그 방향은 문제가 아니다**. 감사와 본 SPEC 이 각각 측정해 같은 결론에 도달했다.)
- **복사 모드 — 매 업데이트마다 전량 복제.** Windows 폴백에서 `.agents/skills/moai-*` 는 실 디렉터리다. 템플릿 트리에는 `.agents/` 가 **존재하지 않으므로**(§A.1) `templateManagedPaths` 는 **항상 공집합**이고, 그 아래 모든 파일이 unmanaged 로 분류된다. 결과: `moai update` 를 돌릴 때마다 배포된 스킬 전량이 `.moai-backups/<timestamp>/pre-clean/.agents/` 아래로 복사된다 — 매번, 영원히.

**귀결**: 미러는 정본의 파생물이므로 백업 대상이 아니어야 한다. §A.6 과 같은 원칙이며 같은 요구사항(REQ-CSC-010)이 둘 다 규정한다.

### A.8 템플릿 `.gitignore` 는 `.agents/` 를 무시하지 않는다 — 실측

`internal/template/templates/.gitignore` 에는 `.claude/` 계열 항목만 있고 `.agents/` 항목은 **없다**. 이대로 착지하면 모든 사용자 프로젝트가 커밋 후보 항목을 새로 얻는다.

- **링크 모드**: git 은 심볼릭 링크를 링크로 저장하지만, 심볼릭 링크를 지원하지 않는 Windows 체크아웃에서는 **경로 문자열을 담은 일반 텍스트 파일로 실체화된다**. 커밋되는 순간 크로스 플랫폼 결함이 된다.
- **복사 모드**: 배포된 스킬 전량이 사용자 저장소에 중복 커밋된다.

미러는 배포 산출물이지 소스가 아니다. REQ-CSC-016 이 이를 규정한다.

### A.9 미러 집합과 청소 집합은 지금 **우연히** 같다

- 미러 집합(REQ-CSC-006) = 이번 실행에서 배포된 스킬 전부 — 이름 제약 없음.
- 청소 집합(REQ-CSC-008) = `.agents/skills/moai*` — `moai` 접두만.

실측: 템플릿 스킬 34개 중 `moai` 로 시작하지 않는 이름은 **0개**(`find … -exec basename {} \; | grep -cv '^moai'` → 0). 두 집합은 현재 일치한다.

문제는 **그 일치가 어디에도 요구사항으로 적혀 있지 않다**는 것이다. §A.3 이 지목한 빈 `harness_generated.skills` 슬롯에 비-`moai` 이름이 하나 들어오면, 미러는 `.agents/skills/<비-moai>` 를 만들고 청소 글롭은 그것을 **영원히 지우지 못한다** — §A.4 가 `~/.codex/skills/` 에서 관측한 오염이 `.agents/` 에서 재현된다. §A.5 가 스스로 지목한 "청소 규칙과 배포 규칙이 같은 집합을 가리키지 않으면 조용히 어긋난다"가 바로 이 형태다.

**이 불변식은 코드에 없다. 코드에 있는 것은 다른 불변식이고, 그쪽은 이미 깨져 있다.** iter-3 까지 이 자리에는 "불변식은 `internal/template/skills_manifest.go` 의 `moaiSkillPrefix` 필터로 이미 암묵적으로 있다"고 적혀 있었는데, 실측하니 사실이 아니다.

```
const moaiSkillPrefix = "moai-"        # 하이픈 포함
grep -cv '^moai'  names → 0            # REQ-CSC-015 의 불변식(접두 moai)은 성립
grep -cv '^moai-' names → 1            # 코드 상수의 불변식(접두 moai-)은 성립하지 않는다
grep -v  '^moai-' names → moai
```

카탈로그에는 이름이 **정확히 `moai`** 인 통합 스킬이 있고, `EmbeddedMoaiSkillNames()` 는 그것을 뺀 **33개**를 돌려준다. 함수 주석이 그 제외를 의도로 명시한다.

따라서 두 접두를 구분해 쓴다. **본 SPEC 전체가 쓰는 접두는 `moai`(하이픈 없음)** 이며 — 청소 글롭 `moai*`, REQ-CSC-015, AC-CSC-016, plan §H 판별 기준이 모두 이 철자다 — `moai-`(하이픈 포함)를 쓰는 `EmbeddedMoaiSkillNames()` 는 **이 SPEC 의 어떤 집합 정의에도 사용하지 않는다**. 그 함수를 재사용하면 `moai` 스킬 하나가 미러에서 조용히 빠지고(REQ-CSC-006 위반), `~/.codex/skills/moai` 가 은퇴 스킬로 오분류된다.

코드에 없는 불변식이므로 REQ-CSC-015 는 더욱 필요하다 — 세우는 것이지 인용하는 것이 아니다.

### A.10 실제 실행 순서에서 청소는 미러 링크를 한 개도 지우지 못한다 — 실측

세 사실이 겹친다. 전부 직접 측정했다.

1. **clean 이 deploy 보다 먼저다.** `internal/cli/update_template_sync.go` 의 스텝 배열에서 `deploy.CleanMoaiManagedPaths`(`:297`)가 `deployer.Deploy`(`:323`)보다 앞선다.
2. **clean 안에서 정본이 먼저 지워진다.** `CleanMoaiManagedPaths` 는 `ManagedCleanTargets` 슬라이스를 **순서대로** 돈다. 그 4번째 항목이 `.claude/skills/moai*` 글롭이므로, 신규 `.agents/skills/moai*` 항목을 뒤에 붙이면 미러를 처리할 시점에 정본은 **같은 실행에서 이미 삭제**돼 모든 미러 링크가 dangling 이다.
3. **dangling 링크는 무음으로 건너뛰어진다.** `backupThenRemove` 의 첫 동작이 `os.Stat`(링크를 따라감)이고 `os.IsNotExist` 면 `return 0, nil` — 제거 없이 성공 반환한다.

직접 실행한 재현:

```
Stat err: stat a/.agents/skills/moai-x: no such file or directory  IsNotExist: true
Lstat err: <nil>
glob: [a/.agents/skills/moai-x]
```

`filepath.Glob` 은 dangling 링크를 **매치한다**. 청소는 대상을 찾아내고도 지우지 않는다.

**귀결**: REQ-CSC-008 이 막으려는 바로 그 실패(은퇴 스킬의 영구 잔존)가 규정대로 구현해도 발생한다. §A.7 의 "링크 모드 — 안전" 서술은 링크가 **살아 있을 때**만 옳고, 실행 순서가 정확히 살아 있지 않은 순간을 만든다. 대응은 §B.D6.

### A.11 백업 금지가 잃는 것 — 무백업 손실 경로

§A.7 이 요구한 백업 금지를 절대 형태로 쓰면 새 손실 경로가 열린다. 청소 글롭은 `moai*` 이고 **`moai` 접두는 사용자도 쓸 수 있다**(그 사실이 REQ-CSC-014 의 존재 이유다). 배포 단계는 그런 항목을 지키는데 청소 단계는 글롭으로 매치해 제거하므로, 백업까지 금지하면 사용자 항목이 **경고도 백업도 없이** 사라진다.

판별자는 이미 코드에 있다. `backupThenRemove` 는 템플릿이 같은 경로를 가진 파일을 백업하지 않는데, 그 사유가 주석에 적혀 있다 — *배포가 곧바로 다시 쓰므로 유일본이 위태로운 적이 없다*. 같은 기준을 미러에 적용한다: **이번 실행이 다시 만들 미러는 백업하지 않고, 다시 만들지 않을 항목은 백업한다.** REQ-CSC-010 이 이 형태다.

## §B. 설계 결정

### D1 — 정본은 `.claude/skills/`, 미러는 `.agents/skills/`

`.claude/skills/` 를 정본으로 유지하고 배포기가 `.agents/skills/<name>` → `../../.claude/skills/<name>` 상대 심볼릭 링크를 만든다.

이 방향을 고르는 이유는 회귀 위험의 위치다. `.claude/skills/` 배포 산출물이 **바이트 단위로 무변경**이므로 Claude Code 무회귀가 **구성상(by construction)** 성립한다 — 테스트로 방어하는 것이 아니라 애초에 움직이지 않는다.

### D2 — 기각한 대안: 정본을 `.agents/skills/` 로 옮기고 `.claude/skills/` 를 링크로

M0 실측상 Codex 는 `.claude/skills/` 를 스캔하지 않으므로 역방향도 Codex 쪽에서는 동작한다. 그럼에도 **기각**한다. 이 방향은 회귀 위험을 **움직이면 안 되는 경로**(Claude Code 가 매 세션 읽는 경로)로 옮긴다. 더구나 Claude Code 의 스킬 탐색이 심볼릭 링크를 어떻게 다루는지는 본 SPEC 시점에 미실측이며, 미실측 전제 위에 정본을 올리는 것은 §A.2 가 보여준 실패 형태의 반복이다.

### D3 — 폴백은 복사, 그리고 관측 가능해야 한다

`os.Symlink` 는 Windows 에서 권한 또는 개발자 모드를 요구한다. 실패 시 **실 디렉터리 복사**로 폴백하며, 어느 모드가 쓰였는지 사용자에게 보고한다. 폴백이 조용하면 "링크인 줄 알았는데 복사본이라 정본 갱신이 반영되지 않는" 상태가 무음으로 생긴다.

### D4 — 미러 집합은 이번 실행이 실제로 배포한 집합에서 파생한다

상수(34)를 쓰지 않는다. §A.3 때문이다.

### D5 — 청소 경로 등록, 접두는 `moai*` 로만 한정

`ManagedCleanTargets` 에 `.agents/skills/moai*` 글롭을 **단 하나** 추가한다. 기존 `.claude/skills/moai*` 항목과 같은 형태(`IsGlob: true`)를 따른다.

[HARD] 이 글롭은 `moai*` 이외의 어떤 이름과도 **매치되어서는 안 된다**. 특히 `hns-*` 는 사용자 소유이며 `moai update` 가 지우지 않는다는 것이 네임스페이스 분리 계약이다(`.moai/docs/harness-namespace-doctrine.md` §24.1 — `hns-*` canonical, `harness-*` · `my-harness-*` legacy 인식 대상, 셋 다 **NOT synced (보호)**). 같은 계약이 `.claude/agents/harness/` · `.claude/commands/harness/` · `.moai/harness/` 에도 걸린다.

`.agents/` 전체나 `.agents/skills/` 전체를 대상으로 잡는 형태는 **금지**다(AP-6). 그 형태는 사용자 소유 스킬을 지우면서도 "은퇴 스킬이 사라졌다"는 관측과 구분되지 않는다 — §A.5 의 1번 사고와 같은 실패다.

[HARD] **`.agents/` 아래에서 이 계약을 지키는 것은 글롭이 좁다는 사실 하나뿐이다.** 위에 인용한 doctrine 의 기계적 집행자인 `IsUserOwnedNamespace`(`internal/cli/update/plan/plan.go`)는 판정을 전부 **`.claude/` 경로 접두**로 한다(`.claude/skills/hns-`, `.claude/skills/harness-`, `.claude/skills/my-harness-`, `.claude/agents/harness`). `.agents/` 아래 항목은 그 술어의 **시야에 없다** — 백업·preserve·doctor 어느 계층도 보지 않는다. 따라서 doctrine 인용은 *무엇을 지켜야 하는지*의 근거일 뿐, 두꺼운 방어층이 이미 있다는 뜻이 아니다. 훗날 글롭을 넓히자는 제안이 오면, 그 제안이 제거하는 것이 **유일한 방어**임을 이 문장에서 읽어야 한다.

### D6 — dangling 링크 제거: `Lstat` 판정이 본체, 순서 배치는 이중 방어

§A.10 의 결함에 두 가지를 **함께** 건다(REQ-CSC-008).

- **(a) 본체 — `os.Lstat` 판정 + dangling 제거.** 청소가 대상의 존재를 링크 자체 기준으로 판정하고, 정본이 사라진 링크도 지운다. 이것이 결함을 실제로 닫는 절이다.
- **(b) 이중 방어 — 순서 배치.** `.agents/skills/moai*` 항목을 `.claude/skills/moai*` 앞에 둬, 정상 실행에서는 미러가 살아 있는 동안 처리되게 한다.

[HARD] **(b) 를 중복이라며 지우지 않는다.** 두 감사자 모두 (b) 단독은 취약하다고 지적했고 — 슬라이스 순서는 나중에 조용히 바뀐다 — 그 지적은 옳다. 그래서 (b) 를 본체로 삼지 않았다. 그러나 (a) 단독도 하나의 회귀로 무너진다. 둘을 함께 두는 목적은 **어느 한쪽의 회귀가 단독으로 결함을 되살리지 못하게** 하는 것이다.

**받아들인 결과 — 폭발 반경.** `backupThenRemove` 의 판정을 `os.Stat` 에서 `os.Lstat` 로 바꾸는 것은 `.agents/` 만의 변경이 아니다. 그 함수는 **모든 관리 대상 청소 항목**이 지나가는 공유 코드이므로, `.claude/skills/moai*` 를 비롯한 기존 7개 뿌리에서도 dangling 링크가 이제 제거된다. 영향 범위는 "관리 대상 뿌리 아래에 있으면서 대상이 사라진 심볼릭 링크" 전부다.

이것을 **의도한 수정으로 받아들인다.** dangling 링크를 남기는 현재 동작은 어느 뿌리에서도 옳지 않고, `.agents/` 에서만 예외를 만드는 형태는 같은 결함을 다른 곳에 남긴다. 다만 이 변경이 M1 의 명목 범위(미러 배포) 밖의 공유 코드 동작 변경이라는 점은 **부작용이 아니라 기록된 결정**이며, run-phase 는 기존 뿌리에 대한 회귀를 함께 확인해야 한다.

### D7 — `.gitignore` 는 `.agents/` 전체가 아니라 `.agents/skills/moai*` 만

이 SPEC 은 청소 쪽에서 "`.agents/` 전체를 잡는 형태는 금지"라는 좁은-범위 원칙을 세운다(AP-6, §B.D5). `.gitignore` 에도 같은 원칙을 적용한다.

생성물은 `.agents/skills/moai*` 뿐이다. `.agents/` 전체를 무시하면 사용자가 만든 `.agents/skills/hns-*` 와, 후속 마일스톤(M2 AGENTS.md 정본화 등)이 `.agents/` 아래 둘 수 있는 **소스 파일까지** 조용히 추적에서 빠진다. 무시 대상은 생성물에만 건다.

## §C. 요구사항 (GEARS)

- **REQ-CSC-001** — 배포기는 이번 실행에서 `.claude/skills/<name>/` 에 배포한 모든 스킬에 대해 `.agents/skills/<name>/SKILL.md` 로 읽을 수 있는 접근 경로를 제공해야 한다(shall). 단 REQ-CSC-014 가 규정하는 **대상 선점 상태**와 REQ-CSC-011 이 규정하는 **미러 생성 실패**는 예외이며, 그 두 경우에는 경고가 접근 경로의 자리를 대신한다.
- **REQ-CSC-002** — 템플릿 소스 트리 `internal/template/templates/` 는 심볼릭 링크를 포함해서는 안 된다(shall not).
- **REQ-CSC-003** — 배포가 스킬 하나를 `.claude/skills/<name>/` 에 쓸 때(When), 배포기는 `.agents/skills/<name>` 을 `../../.claude/skills/<name>` 상대 심볼릭 링크로 만들어야 한다.
- **REQ-CSC-004** — 심볼릭 링크 생성이 지원되지 않거나 권한이 없는 환경에서(Where), 배포기는 대신 스킬 디렉터리 내용을 복사한 실 디렉터리를 만들어야 한다.
- **REQ-CSC-005** — 복사 폴백이 발동하거나 미러 생성이 실패할 때(When), 배포기는 사용된 모드와 경고를 **자신의 반환 결과에 담아 호출부로 올려야 하며**(shall), 사용자 표시는 호출부가 수행한다. 배포기 내부에서 직접 출력해서는 안 된다(shall not) — 근거는 §B.D6.
- **REQ-CSC-006** — 미러 대상 집합은 이번 실행에서 실제로 배포된 스킬 집합과 정확히 일치해야 하며(shall), 어떤 고정 상수에서도 파생해서는 안 된다(shall not).
- **REQ-CSC-007** — **미러 기능의 활성 여부가** `.claude/skills/` 아래 배포 산출물의 경로·내용·권한·manifest 항목을 변화시켜서는 안 된다(shall not). 착지 시점의 "본 변경 전과 동일" 확인은 이 요구가 아니라 §D.4 의 1회 대조가 담당한다 — 그 분리 사유는 AC-CSC-010 에 적혀 있다.
- **REQ-CSC-008** — `moai update` 의 청소 단계가 실행될 때(When), 청소 대상은 `.agents/skills/moai*` 접두 글롭을 포함해야 하며, 그 글롭은 `.agents/skills/` 아래 **정확히 `moai` 로 시작하는 이름에만** 매치되어야 한다(shall). 또한 청소는 대상의 존재 여부를 **`os.Lstat` 기준으로 판정해야 하고**(shall), 정본이 이미 사라진 **dangling 링크도 제거해야 한다**(shall) — 근거는 §A.10. 이에 더해, `.agents/skills/moai*` 항목은 `ManagedCleanTargets` 안에서 `.claude/skills/moai*` 항목보다 **앞에 배치되어야 한다**(shall).
- **REQ-CSC-009** — `.agents/skills/` 아래에 사용자 소유 항목(`hns-*` · `harness-*` · `my-harness-*` 및 그 밖의 비-`moai` 이름)이 존재하는 동안(While), 청소는 그 항목을 제거해서는 안 된다(shall not).
- **REQ-CSC-010** — 시스템은 미러 항목을 manifest 에 기록해서는 안 된다(shall not). 또한 **이번 실행이 곧바로 다시 만들 미러**(템플릿이 같은 이름의 스킬을 가진 항목)는 pre-clean 백업 트리에 보존해서는 안 된다(shall not). 그 판별에 걸리지 않는 `.agents/skills/moai*` 실 항목 — 템플릿에 대응 스킬이 없는 은퇴 이름, 그리고 사용자가 만든 `moai` 접두 항목 — 은 **기존 백업 규칙을 그대로 따라야 한다**(shall). 근거는 §A.6·§A.7·§A.11.
- **REQ-CSC-011** — 미러 생성이 링크·복사 양쪽 모두 실패할 때(When), 배포기는 경고를 남기고 계속 진행해야 하며, `.claude/skills/` 배포 결과를 취소하거나 배포 전체를 실패시켜서는 안 된다(shall not).
- **REQ-CSC-012** — 미러 대상이 이미 올바른 정본을 가리키는 링크일 때(When), 재배포는 그 대상의 상태를 변경하지 않아야 한다(멱등).
- **REQ-CSC-013** — 미러 대상이 링크이지만 **다른 곳을 가리킬 때**(When), 배포기는 이를 올바른 정본을 가리키도록 교체해야 한다(shall).
- **REQ-CSC-014** — 미러 대상 경로에 링크가 아닌 **실 항목**(사용자가 만든 디렉터리·파일)이 이미 있을 때(When), 배포기는 그것을 제거하거나 덮어써서는 안 되며(shall not), 건너뛰고 경고해야 한다(shall).
- **REQ-CSC-015** — 배포되는 모든 스킬의 디렉터리 이름은 **`moai` 접두(하이픈 없음)** 를 가져야 한다(shall). 이 불변식이 미러 집합(REQ-CSC-006)과 청소 집합(REQ-CSC-008)의 일치를 보증한다 — 근거는 §A.9.
- **REQ-CSC-016** — 미러 산출물은 사용자 저장소의 버전 관리 대상이 되어서는 안 된다(shall not). 배포되는 `.gitignore` 는 **`.agents/skills/moai*`** 를 무시해야 한다(shall). `.agents/` 전체를 무시해서는 안 된다(shall not) — 근거는 §A.8·§B.D7.

## §D. 범위 밖 (Exclusions)

이 절은 **무엇을 만들지 않는지**를 고정한다. 아래 항목은 본 SPEC 의 run-phase 산출물에 포함되지 않는다 — 즉 out of scope 이다.

### Out of Scope — 사용자 홈 `~/.codex/skills/` 정리 실행

- 잔존 구 스킬의 **삭제 실행**은 범위 밖이다. `plan.md` 는 판별 기준과 명령만 문서화하고, 실행은 별도 운영자 승인을 요구한다.
- `~/.codex/skills/.system` 은 Codex 소유이므로 어떤 경우에도 건드리지 않는다.

### Out of Scope — dev-only 스킬 노출

- 로컬 전용 `hns-*` 10개는 템플릿 미러 대상이 아니므로 `.agents/skills/` 에도 노출하지 않는다.
- 사용자가 스스로 만든 `.agents/skills/` 항목의 생성·관리도 범위 밖이다. 본 SPEC 이 제공하는 것은 **파괴에 대한 보장 세 가지**뿐이다 — 청소가 비-`moai` 이름을 지우지 않고(REQ-CSC-009), 배포가 선점된 대상을 덮어쓰지 않으며(REQ-CSC-014), `moai` 접두를 쓴 사용자 실 항목은 청소에 걸리더라도 **백업을 거쳐** 제거된다(REQ-CSC-010 의 판별자 한정, §A.11).
- 세 번째 보장이 필요한 이유는 `moai` 접두를 사용자도 쓸 수 있기 때문이다. 그 경우 청소 글롭이 매치하므로 항목 자체는 제거되지만, 백업 없이 사라지지는 않는다. 제거 자체를 막는 것은 본 SPEC 의 범위가 아니다 — 배포되는 스킬과 이름 공간을 공유한 결과다.

### Out of Scope — Codex 하네스의 나머지 마일스톤

- AGENTS.md 정본화(M2), 훅 어댑터(M3), 배선 생성기(M4), 에이전트 TOML 이중 발행(M5), 플러그인 패키징(M6) 은 각각 별도 SPEC 소관이다.
- `codex mcp add` 등록, 훅 신뢰 재승인 안내, Codex 슬래시 커맨드 네임스페이스 충돌 해소도 본 SPEC 이 다루지 않는다.

### Out of Scope — 구현 세부

- 미러 로직을 담을 파일 이름, 함수 시그니처, 상수 이름은 run-phase 판단이다. 본 SPEC 은 관측 가능한 결과만 규정한다.
- 카탈로그 tier 체계 자체의 변경(어떤 스킬이 core 인지)은 다루지 않는다. 본 SPEC 은 tier 결과를 **읽을 뿐** 바꾸지 않는다.

## §E. 비기능 제약

- **OS 중립성**: darwin / linux / windows 에서 모두 REQ-CSC-001 이 성립해야 한다. windows 는 폴백 경로로 성립해도 된다.
- **Template-First**: `.claude/` · `.moai/` 아래 신규·변경 파일은 `internal/template/templates/` 에 미러하고 `make build` 를 실행한다.
- **템플릿 중립성**: 템플릿에 들어가는 내용은 SPEC ID · REQ 토큰 · 내부 날짜 · 커밋 SHA · macOS 편향 경로를 담지 않는다(`.moai/docs/template-internal-isolation-doctrine.md` §25.1).

## §F. 교차 참조

- `.moai/reports/t91/README.md` — M0 Codex 전제 실측 (스킬 스캔 경로, 심볼릭 링크 런타임 동작)
- `.moai/reports/t81/m1-preflight-measurements.md` — M1 선행 실측 (embed 심볼릭 링크 소실 최소 재현)
- `.moai/reports/moai-adk-dual-harness-codex-20260817.md` — 재설계 문서 (§A 가 정정하는 전제의 출처). **primary 체크아웃 기준 경로다** — worktree `WT-skills-canonical` 에는 이 파일이 없으므로, 이 트리 안에서 열리지 않는다고 부재로 판단하지 않는다.

## §G. 기각한 감사 지적

감사에서 지적됐으나 **의도적으로 반영하지 않은 것**을 근거와 함께 남긴다. 조용히 넘어가면 다음 감사가 같은 지적을 다시 하게 된다.

- **D12 / iter-2 D12 (REQ 원자성 분리, optional/minor) — 기각 유지.** REQ-CSC-006 과 REQ-CSC-011 이 한 항목에 `shall` 과 `shall not` 을 함께 담는다는 지적은 사실이다. 그럼에도 분리하지 않는 이유는 **예산**이다: 이번 개정으로 요구사항이 12 → 16 에 도달해 Tier M 상한(16)에 정확히 닿았고, 분리에는 번호 2개가 더 필요하다. 상한을 넘기는 것은 감사가 경계하는 과잉 형식화 그 자체이고, 기존 12개를 재번호하면 이번 개정에서 다시 쓰는 세 판정 계층 전부의 추적 매핑이 함께 흔들린다 — 원자성이라는 minor 이득을 위해 매핑 오류라는 major 위험을 사는 교환이다. 두 항목 모두 형식 GEARS 이며 MP-2 위반이 아니고, 각 절이 대응 AC 에서 별도 단언으로 갈라져 있으므로(REQ-CSC-006 → AC-CSC-003 + AC-CSC-014, REQ-CSC-011 → AC-CSC-013 의 두 팔) "어느 절이 실패했는가"는 실행 시점에 구분된다. 원자성 손실이 실제 판정 능력의 손실로 이어지지 않는다. iter-2 감사도 이 기각을 정당하다고 재확인했다. 다만 그 감사가 덧붙인 단서를 그대로 받아들인다 — **예산 논거는 blocking 수정에는 적용되지 않는다.** v0.4.0 의 수정은 전부 기존 번호 안의 절 추가·수정이며, "상한에 닿아서 못 고친다"를 사유로 쓴 항목은 하나도 없다.

- **iter-2 N9 (§G 가 §F 앞에 온다, optional/minor) — 수용.** 절 순서를 바로잡았다. 기각 목록이 아니라 반영 목록에 속하지만, 이전 판본을 읽은 사람이 순서 변경을 이상하게 여기지 않도록 여기 적어 둔다.
