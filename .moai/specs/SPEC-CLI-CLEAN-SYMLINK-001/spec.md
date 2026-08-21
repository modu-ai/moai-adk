---
id: SPEC-CLI-CLEAN-SYMLINK-001
title: "moai update 청소 경로의 심볼릭 링크 인식 (symlink-dedicated classification)"
version: "0.1.0"
status: draft
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: internal/cli/update/deploy
lifecycle: spec-anchored
tier: M
era: V3R6
tags: "update, clean, symlink, deploy, backup, t173"
---

# SPEC-CLI-CLEAN-SYMLINK-001

## HISTORY

- 2026-08-22: 최초 draft (card t173 — "moai update 청소 경로 링크 인식"). 모든 측정 근거는
  plan-phase §A 측정 기록 `.moai/reports/t173/measurements.md`(이하 "도시에")가 소유하며,
  본 SPEC은 그 위에 처분 의미론만을 정의한다. 코드 인용 앵커는 본 워크트리 HEAD
  `18f7cfc19`(= `4b2f203fe` + 도시에 커밋)이며, 이 트리에서 아래 인용 라인 전부를 직독 확인했다.
- t81(가) 미러 배포 카드와 같은 release(v3.1.3)에 착지한다. 두 카드의 교차 계약은
  REQ-CSL-009가 소유한다.
- 2026-08-22 (보충 반영): 리드 보충 항목 4 — 복사 모드 미러 갱신 판별자 질문을 §A에
  맥락 항목으로 추가하고 t81(가)(SPEC-CODEX-SKILLS-CANONICAL-001 v0.7.0) 배포 측으로
  라우팅했다(§E). REQ/AC/형태 수치 불변(12/11/5).
- 2026-08-22 (리드 결정 반영): 두 처분이 비준됐다(FX-1 제거+가시화, FX-3b 제거 — plan.md
  §D-5 DECIDED). 판별자 라우팅 정정 — t81(가) 소관이 아니라 **리드 큐의 후속 카드 후보**로
  (§A 보충 항목 4, §E). REQ/AC/형태 수치 불변(12/11/5).

## §A. 문제 서술 (Problem)

`moai update`의 청소 단계(`CleanMoaiManagedPaths`, deploy.go:101)는 배포 전에 관리 뿌리
7곳(`ManagedCleanTargets`, deploy.go:50-82)과 별도 8번째 뿌리 `.moai/config`(deploy.go:168-182)을
지운다. 이 분류 체계는 심볼릭 링크를 인식하지 못한다 — 모든 판정이 `os.Stat`(링크 추적)이고,
링크는 우연히 파일 분기나 디렉터리 분기에 흡수된다. 도시에의 4회 재현(Run A~D)이 실측한 결과:

1. **출하 결함 (Run D)** — 비-글롭 관리 뿌리에 dangling 링크가 있으면: 루트 사전검사의
   `os.Stat`(deploy.go:139)이 링크를 추적해 ENOENT → "Skipped (not found)"(:140-146)로
   **링크가 존재하는데도 남는다** → 배포의 `os.MkdirAll`(deployer.go:189)이 dangling 링크에서
   EEXIST → **update가 트리 부분 파괴 상태로 중단**하고, 재실행해도 같은 Skip이 재현되어
   사용자가 링크를 수동 제거하기 전까지 update는 영구 불능이다.
2. **무인식 (Run B/D)** — 살아 있는 디렉터리 링크: WalkDir이 링크 루트를 스킵(deploy.go:441)해
   백업 0건, `os.RemoveAll`은 링크만 제거, 대상 무사, 실 디렉터리로 재배포 — 중단은 없으나
   **진행 출력 어디에도 링크였다는 사실이 나타나지 않는다**. 살아 있는 파일 링크: 대상 바이트가
   백업·복원되고 링크 소멸 — 역시 무인식. 글로브 매치 dangling 링크(`moai-dangling-custom`)는
   무소식 no-op(deploy.go:372-375)로 **영구 잔존**한다.
3. t81 iter4 감사의 "Lstat 치환" 제안은 어느 트리에도 존재하지 않는 기각안이다(도시에 §1.2,
   4증거). 단순 치환은 라이브 디렉터리 링크에서 EISDIR, dangling에서 ENOENT 하드 실패를 낸다.

### 복사 모드 미러 갱신 판별자 (리드 보충 항목 4)

t81(가)(SPEC-CODEX-SKILLS-CANONICAL-001 v0.7.0)의 미러는 심볼릭 링크 생성이 불가한
플랫폼에서 실 디렉터리 복사로 폴백하며(REQ-CSC-004) 어느 모드가 쓰였는지 경고로
보고한다(REQ-CSC-005). 그 SPEC은 이 시점에 **판별자를 정의하지 않는다** — "우리가 지난
실행에 만든 복사본"과 "사용자가 만든 디렉터리"를 가르는 신호(마커 파일·소유 표시, 그
무엇도)가 요구사항에 없다. v0.7.0 §D의 고지: 폴백 플랫폼에서 2회차 배포부터 복사본이
REQ-CSC-014의 "실 항목" 분기에 걸려 사용자 항목으로 취급되어 보존되므로 정본은 갱신되고
미러는 1회차 내용에 고착한다(경고조차 우리 산출물을 사용자 항목으로 오귀속해 보고한다).

**청소 경로와의 관계 — 판별자는 clean에 하중을 받지 않는다.** 본 SPEC의 제거 처분
아래에서 관리 뿌리의 실 디렉터리는 누가 만들었는지와 무관하게 FX-4로 균일 처리된다
(비관리 파일 백업 → 제거 → 재배포). 청소는 관리 뿌리의 실 항목을 보존하는 분기를
두지 않으므로 "우리 복사본 vs 사용자 디렉터리"를 구분할 이유가 없다. 판별자가 하중을
받는 곳은 배포 측 — t81(가)의 REQ-CSC-014 스킵 분기가 보존과 갱신 중 무엇을 할지
결정하는 지점이다. `.agents/` 미러 배치는 현 청소 집합 밖이라 본 SPEC이 일절 모르며,
그 등록 여부·시점은 t81(가)와 교차 계약(REQ-CSL-009, acceptance §D.7-1)의 소관이다.
따라서 판별자 도입·갱신 경로는 **범위외 — t81(가) §D에 기록된 한계, 후속 카드 후보(리드
큐 관리)**다(§E).

## §B. 형태 분류와 분기별 처분 (branch disposition)

청소 분류기는 판정 시점에서 링크를 추적하지 않는다(`os.Lstat` 의미론). `fs.ModeSymlink`
비트가 있으면 **IsDir 판정 이전에** 링크 전용 분기로 진입하며, 그 분기 안에서 형태별 처분이
정해진다. 어떤 링크 형태도 파일 분기나 디렉터리 분기에 흡수되지 않는다.

픽스처 형태는 5종(도시에 §1.3 매트릭스가 형태×분기 근거 데이터; dangling은 배치 2곳이
같은 형태의 두 진입 팔이다):

| 형태 | 진입 팔 | 현행 (os.Stat) 실측 | 이 SPEC의 처분 (링크 전용 분기) |
|---|---|---|---|
| **FX-1** 라이브 디렉터리 링크 (관리 뿌리 → 실재 디렉터리) | 비-글롭 루트 / 글로브 매치 | 중단 없음. 백업 0건, 링크만 제거, 대상 무사, 실 디렉터리 재배포, **무인식** (Run B/D) | **링크만 제거 + 형태를 이름붙인 진행줄.** 대상 일절 미접촉, 백업 0건 연속(WalkDir-스킵 의미론 유지), 배포가 실 디렉터리 재생성 |
| **FX-2** 라이브 파일 링크 (settings.json → 실재 파일) | 파일 루트 | 대상 바이트 백업 + 머지 복원, 링크 소멸, **무인식** (Run B) | **대상 바이트 독해 백업 + 링크 제거 + 진행줄.** 3-way merge 복원 흐름 유지 |
| **FX-3a** dangling 링크, 비-글롭 뿌리 | 루트 사전검사 | "Skipped (not found)" → 링크 잔존 → deploy EEXIST → **부분 파괴 중단 + 영구 루프** (Run D) | **링크 제거 + dangling임을 이름붙인 진행줄** → 배포 정상 진행, 실 디렉터리 재배포 |
| **FX-3b** dangling 링크, 글로브 매치 이름 | 글로브 팔 | 무소식 no-op → **영구 잔존** (Run D) | 동일 링크 분기: **링크 제거 + 진행줄** |
| **FX-3c** dangling 링크, `.moai/config` 뿌리 | 8번째 제거 뿌리(사전검사 없음) | 코드 추적 상 FX-3a와 동형(도시에 §2.1, gap 4 — 미실측) | 동일 링크 분기 적용 — run-phase Go 테스트로 실측 전환 |
| **FX-4** 실 디렉터리 / 실 파일 (대조) | 정상 분기 | 정상 청소·백업·재배포 (Run A) | **변경 없음** — 현행 분기 의미론 그대로 |
| **FX-5** 사용자 소유 네임스페이스 (hns-*, harness-*, 비-moai 스킬 — 내부 링크 포함) | 청소 대상 아님 | 일절 미접촉 (Run C) | **변경 없음 — 일절 미접촉** (must-not-flag 극) |

### §B.1 FX-1 처분 선택 — 제거+가시화 vs 보존

라이브 디렉터리 링크의 처분 후보는 (1) **제거+가시화**(현행 결과 + 진행줄)와 (2) **보존**이다.
본 SPEC은 (1)을 채택하며, 근거는 둘 다 이 트리에서 확인했다:

- **보존은 배포 경로와 결합된다**: `os.MkdirAll`은 fast-path에서 `Stat`(링크 추적)으로
  디렉터리 여부를 판정한다(GOROOT `os/path.go` `MkdirAll` 직독) — 라이브 디렉터리 링크에서
  **nil을 반환**하므로, 보존된 링크를 통해 배포 기록(`atomicWriteFile`, deployer.go:201)이
  사용자의 외부 디렉터리 내부로 유입된다. 보존을 안전하게 만드려면 배포 측 링크 판정이
  필수인데 그것은 이 카드의 범위 밖이다(§E).
- **대상 무사는 링크 제거가 선행하기 때문에 성립한다**(Run B 실측 — 링크가 사라진 뒤
  재배포되므로 대상은 한 번도 건드려지지 않았다).
- (1)은 현행 결과를 바꾸지 않고 관측성만 더한다 — 회귀면이 가장 작은 방향이다.

선택의 대가: 관리 뿌리를 심볼릭 링크로 관리하던 사용자는 매 update마다 링크를 잃는다
(현행부터 그랬고, 이제는 적어도 진행줄로 보인다). 이 처분이 [HARD] 회귀면에 미치는 지점은
plan.md §D에 명시했고 **리드 비준 완료(2026-08-22, plan.md §D-5 DECIDED)**다.

### §B.2 순서 독립성과 "미러 먼저" 제약

t81(가)가 추가할 미러 링크(`.agents/skills/<name>` → `../../.claude/skills/<name>`)의 대상은
관리 뿌리다. 제거 처분 아래에서는 순서가 무관하다: 어느 쪽을 먼저 처리하든 살아 있는 링크와
dangling 링크 모두 "제거+진행줄"로 수렴하고, 배포가 양 끝을 다시 만들기 때문에 최종 상태가
같다(REQ-CSL-008이 이를 소유). "미러 뿌리를 대상 뿌리보다 먼저 처리하라"는 순서 제약은
**보존 처분이 활성화되는 경우에만 하중을 받는 대비 요건**이다 — 그 경우 대상-먼저 처리가
라이브 미러를 dangling으로 강등시켜 파괴하기 때문이다. 이 대비 제약의 활성화는 이 SPEC의
수정(amendment)을 요구한다(plan.md §D 참조).

## §C. 요구사항 (GEARS)

본 SPEC은 요구사항 12건(REQ-CSL-001 … REQ-CSL-012)을 가진다. 인용 라인은 앵커
`18f7cfc19` 기준(이후 드리프트 가능).

- **REQ-CSL-001** (분류) — **While** 청소 대상 경로의 `os.Lstat` 모드에 `fs.ModeSymlink`
  비트가 있는 경우, 청소 분류기는 해당 항목을 IsDir 판정 이전에 링크 전용 분기로 진입시켜
  형태별로 정의된 처분을 실행해야 한다(shall). 어떤 심볼릭 링크도 — 살아 있든 dangling이든 —
  파일 분기나 디렉터리 분기에 도달해서는 안 된다(shall not).

- **REQ-CSL-002** (dangling 결함 수정) — **When** 청소 집합의 어느 항목이 dangling 심볼릭
  링크인 경우(링크 자체는 존재, `os.Stat` 추적 시 대상 부재) — 비-글롭 뿌리, 글로브 매치
  이름, `.moai/config` 뿌리 어디든 — 청소 단계는 링크 자체를 제거하고(shall), 해당 경로가
  dangling 심볼릭 링크였음을 이름붙은 진행줄을 출력하며(shall), 배포 단계가 EEXIST 없이
  진행되도록 해야 한다(shall). update는 이 형태로 인해 중단되어서는 안 된다(shall not).

- **REQ-CSL-003** (라이브 디렉터리 링크 처분) — **When** 디렉터리 형태의 관리 뿌리가 살아
  있는 심볼릭 링크(대상 디렉터리 실재)인 경우, 청소 단계는 링크만 제거하고(shall), 링크의
  대상을 읽거나 백업하거나 수정해서는 안 되며(shall not), 해당 루트의 사전청소 백업 파일
  수를 0으로 유지하며(shall — WalkDir-스킵 의미론의 연속), 링크였음을 이름붙은 진행줄을
  출력해야 한다(shall). 배포는 해당 경로에 실제 디렉터리를 재생성해야 한다(shall).

- **REQ-CSL-004** (라이브 파일 링크 처분) — **When** 파일 루트(`.claude/settings.json`)가
  살아 있는 심볼릭 링크(대상 파일 실재)인 경우, 청소 단계는 링크를 통해 대상의 바이트를
  사전청소 백업에 복사하고(shall), 링크를 제거하며(shall), 링크였음을 이름붙은 진행줄을
  출력해야 한다(shall). 백업된 바이트가 사용자 내용으로 복원되는 하류 흐름(3-way merge)은
  보존되어야 한다(shall).

- **REQ-CSL-005** (관측성) — 청소 단계는 심볼릭 링크와 조용히 상호작용해서는 안 된다(shall
  not). 청소 집합 아래에서 이루어지는 모든 링크 제거는 링크 경로와 형태(dangling / 라이브
  디렉터리 / 라이브 파일)를 이름붙인 진행줄을 동반해야 한다(shall).

- **REQ-CSL-006** (실재 항목 비회귀) — **While** 청소 집합의 항목이 심볼릭 링크가 아닌 실
  디렉터리 또는 실 파일인 경우, 청소 단계는 현행 분기 의미론을 변경 없이 유지해야
  한다(shall) — 비관리 파일의 사전청소 백업, 제거, 재배포의 현재 동작이 그대로 적용된다.

- **REQ-CSL-007** (사용자 소유 네임스페이스) — 청소 단계는 사용자 소유 네임스페이스
  경로를 일절 건드리지 않는다(shall not) — `IsUserOwnedNamespace` 패턴족(`hns-*`,
  `harness-*`, `my-harness-*`, `.claude/skills/`의 비-moai 첫 세그먼트, `.moai/harness/`,
  `.claude/commands/harness/`, `.claude/workflows/hns-*`·`harness-*`)에 속한 경로와 그
  내부의 모든 심볼릭 링크 포함.

- **REQ-CSL-008** (순서 독립) — **When** 청소 뿌리들 가운데 둘 이상이 링크로 연결되어 있는
  경우(한 루트를 가리키는 링크가 다른 청소 대상에 있는 경우), 청소 단계는 루트 처리 순서와
  무관하게 동일한 최종 트리 상태를 만들어야 한다(shall).

- **REQ-CSL-009** (청소 집합 ↔ 배포 집합 교차 계약) — **When** 템플릿 파일시스템과 청소
  집합을 비교하는 경우, 모든 비-글롭 청소 루트는 템플릿이 보유하는 경로여야 하고(shall),
  모든 글로브 청소 패턴은 템플릿 경로 1개 이상과 매치되어야 한다(shall). 청소 집합과 배포
  집합의 발산은 결함이다. t81(가)가 `.agents/` 미러 배포를 추가한 뒤에도 이 계약은
  release/v3.1.3 통합 시점에 성립해야 한다.

- **REQ-CSL-010** (픽스처 형태 일치) — **Where** 테스트가 심볼릭 링크 요구사항을 검증하는
  경우, 테스트 픽스처는 검증 대상 제품 경로와 동일한 링크 형태(파일 링크 / 디렉터리 링크 /
  dangling)로 구성되어야 한다(shall). 파일 링크와 디렉터리 링크는 서로 다른 분기를 타므로
  형태가 어긋난 픽스처는 다른 코드를 시험한 것이 된다(t81 감사 D2 이관).

- **REQ-CSL-011** (공허 단언 금지) — **Where** 단언이 링크 처분을 관측하는 경우, 단언은
  관측 축 — 링크 존재, 대상 무사, 백업 내용, 출력 메시지 — 가운데 2개 이상의 독립 축을
  결합해야 하며(shall), "백업 파일 수 == 0" 따위의 단일 축으로만 세워서는 안 된다(shall not).
  WalkDir-스킵 때문에 구현과 무관하게 항상 참이 되는 숫자는 단언이 아니다(t81 감사 D4 이관).

- **REQ-CSL-012** (플랫폼) — **Where** 호스트 플랫폼이 심볼릭 링크 생성을 허용하지 않는
  경우(`os.Symlink` 실패), 심볼릭 링크 테스트는 실패 대신 `t.Skip`으로 건너뛰어야 한다(shall).
  링크 생성 성공을 가정해서는 안 된다(shall not).

## §D. 비기능 제약 (Constraints)

- 개발 방식: TDD(quality.yaml `constitution.development_mode: tdd` — RED-GREEN-REFACTOR,
  커밋당 최소 커버리지 80%, 패키지 목표 85%).
- 테스트 격리: 모든 임시 디렉터리는 `t.TempDir()`(/tmp 자동 정리). 프로젝트 루트에 테스트
  부산물 금지(CLAUDE.local.md §6 [HARD]).
- 검증 규율: 수정 패키지 테스트만 로컬 실행(`go test ./internal/cli/update/...`), 전체
  스위트 판정은 CI(CLAUDE.local.md §4).
- 코드·주석 영어; 커밋 메시지 영어(Conventional Commits).
- 모든 처분은 위 배포 경로(forceUpdate 모드 — 존재 검사 우회, deployer.go:169-185)와의
  결합 아래에서 정의된다. 배포 쪽 동작을 바꾸는 요구사항은 없다(§E).

## §E. 제외 범위 (Out of Scope)

### Out of Scope — t81(가) 미러 배포 자체
- `.agents/` 미러 링크의 **배포 추가**는 t81(가) 소관이다. 본 SPEC은 청소 쪽 교차
  계약(REQ-CSL-009)과 순서 독립성(REQ-CSL-008)만 소유한다.
- **복사 모드 미러 갱신 판별자** — "우리 복사본 vs 사용자 디렉터리"를 가르는 신호와
  그 위의 미러 갱신 경로. **범위외 — t81(가) §D에 기록된 한계, 후속 카드 후보(리드 큐
  관리)**: t81(가)는 자기 범위에서 제외하며(v0.7.0 D1 수정 (b)), 어느 형제 SPEC도 소유하지
  않는다 — 리드 큐가 관리한다. 본 SPEC의 청소는 이 구분이 없어도 처분 완결이다(§A 보충
  항목 4).

### Out of Scope — 배포 경로의 링크 인식
- deployer의 `MkdirAll`/`atomicWriteFile` 링크 처리 변경. §B.1의 보존 처분을 택하더라도
  필수가 되는 배포 측 변경은 별도 SPEC 소관이다.

### Out of Scope — 기각된 "Lstat 전면 치환" 제안
- t81 iter4 감사 D1이 제안했다가 기각된 `:372` 단순 `os.Lstat` 치환(라이브 디렉터리 링크
  EISDIR / dangling ENOENT 하드 실패). 어느 트리에도 존재하지 않는 코드다(도시에 §1.2).
  본 SPEC의 링크 전용 분기와 다르다.

### Out of Scope — Windows 심볼릭 링크 생성 권한 개선
- 플랫폼 차이는 `t.Skip`으로 회피한다(REQ-CSL-012). 권한 우회·개선 없음.

### Out of Scope — 3-way merge 심층 상호작용
- 사용자가 수정한 settings.json과 파일 링크가 만나는 정밀 내용 흐름(Run B는 init 직후
  픽스처라 단순했음 — 도시에 잔여위험). REQ-CSL-004는 현행 복원 흐름의 보존만 요구한다.

## §F. 참조 (Cross-References)

- 측정 원본(본 SPEC의 모든 실측 근거): `.moai/reports/t173/measurements.md` — 분기 추적
  §1, 청소 뿌리 인벤토리 §2, 재현 매트릭스 §3, t81 D2-D4 원문 §4, 미측정 gap §5.
- t81 감사: `.claude/worktrees/t81/.moai/reports/t81/plan-audit-iter4.md`(D1 :95-135,
  D2 :137-158, D3 :160-171, D4 :173-184).
- t81(가) 미러 배포 SPEC(§A 보충 항목 4의 인용 원본, v0.7.0 — 미커밋 브랜치 내용이라
  본 트리에 없음, 읽기 전용 참조):
  `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t81/.moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/spec.md`.
- 코드: `internal/cli/update/deploy/deploy.go`(ManagedCleanTargets :50-82, 청소 루프
  :101-163, 글로브 팔 :115-137, 사전검사 :139-150, config 뿌리 :168-182,
  backupThenRemove :371-399, WalkDir 스킵 :441); `internal/template/deployer.go`
  (MkdirAll :189, atomicWriteFile :201).
- 스키마/티어: `.claude/rules/moai/development/spec-frontmatter-schema.md`;
  `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.
- 픽스처 형태 5종 · AC 11건(`acceptance.md`) — 본 문서, `plan.md` §F, `acceptance.md`
  3면에서 동일 수치를 유지한다(t81 감사 D3 이관).
