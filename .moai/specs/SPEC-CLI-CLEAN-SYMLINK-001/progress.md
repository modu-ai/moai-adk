---
id: SPEC-CLI-CLEAN-SYMLINK-001
title: "Progress — moai update 청소 경로의 심볼릭 링크 인식"
version: "0.1.1"
status: in-progress
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

# progress.md — SPEC-CLI-CLEAN-SYMLINK-001

## Authoring record

Card t173("moai update 청소 경로 링크 인식")의 plan-phase 산물. 워크트리
`.claude/worktrees/t173`(branch `WT-clean-links`, HEAD `18f7cfc19` = origin/main + 도시에
커밋). 이 에이전트는 git을 건드리지 않았다 — 커밋은 오케스트레이터 소관.

## Phase 1 — 탐사 SKIP 사유 (research pre-gathered)

심층 탐사는 위임 **이전에** 오케스트레이터가 완료했고 전량이 커밋된 도시에
`.moai/reports/t173/measurements.md`(508줄)에 기록되어 있다 — 분기 추적(§1, file:line),
청소 뿌리 인벤토리(§2), 4회 재현 매트릭스 Run A~D(§3), t81 D2~D4 원문(§4), 미측정
gap 7건(§5). 이 위임에서는 도시에 인용 라인 전부를 본 트리(`18f7cfc19`)에서 직독
확인했고(전부 일치), 추가로 두 건을 이 트리에서 신규 확인했다:

1. `internal/template/deployer.go:189`의 `MkdirAll`과 에러 래핑(`template deploy mkdir %q`)
   — Run D의 verbatim 에러와 일치.
2. stdlib `os.MkdirAll`(GOROOT `os/path.go`)의 fast-path `Stat` 추적 — 라이브 디렉터리
   링크에서 nil 반환(보존 처분 기각 근거), dangling에서 slow-path `Mkdir` EEXIST(Run D
   결함의 배포 측 메커니즘).

따라서 research.md 재수집은 중복 — SKIP, 근거 인용함. Tier M 산물 3종 + progress.md.

## 리드 보충 반영 (2026-08-22, 산물 작성 후 1회)

보충 항목 4(복사 모드 미러 갱신 판별자)를 spec.md §A 맥락 항목 + §E Out-of-Scope
라우팅으로 반영했다. 인용 원본(읽기 전용, 미커밋 브랜치 내용 — 본 트리에 없음):
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t81/.moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/spec.md`
(v0.7.0). 확인 요점: REQ-CSC-004(복사 폴백)·REQ-CSC-005(모드 경고 보고)는 있으나
**판별자는 정의 없음** — §D가 폴백 플랫폼 미러 고착(2회차 배포부터 REQ-CSC-014 "실
항목" 분기에 걸려 사용자 항목으로 보존)을 잔존 고지로 적는다.

**라우팅 충돌 기록(리드 인지 필요)**: t81(가) v0.7.0 §D 마지막 문장은 "판별자 도입과
갱신 경로는 승계 SPEC(`t173`)이 닫는다"라고 적고 있으나, 리드 보충 지시는 이를
t81(가) 배포 측 요구사항으로 라우팅하라고 한다. 본 반영은 리드 지시를 따랐다(REQ/AC
추가 없음, 수치 12/11/5 불변) — t81(가) 측 §D 서술과의 정합은 리드·t81(가)가 조율할
사항이다.

**→ 해결(2026-08-22 리드 결정)**: 어느 형제 SPEC도 판별자를 소유하지 않는다 — 후속 카드
후보로 **리드 큐가 관리**한다. t81(가) §D의 t173 지목 문장은 이 리드 결정으로 대체되며,
t81(가) 측 §D 서술 정리는 리드 몫이다.

## 리드 비준 대기 항목 (plan.md §D-5로 이관됨)

1. **FX-1 라이브 디렉터리 링크 처분 = 제거+가시화**(spec.md §B.1 — 보존 기각 근거 포함).
   회귀면: 현행과 동일한 결과(새 파손 없음)이나 "링크 유지 의도 사용자에게 결과 노출"이
   새로 가시화됨. 보존으로 뒤집으면 배포 측 변경이 필수(범위 밖) + 순서 제약 활성화.
2. **FX-3b 글로브 dangling 링크의 제거로 변경**(종전 영구 잔존 → 제거+진행줄) — 관리
   네임스페이스 위생이나 동작 변경은 동작 변경.

Implementation Kickoff Approval 이전에 리드가 위 2건을 비준/기각해야 한다
(plan.md `[NEEDS CLARIFICATION]` 마커).

**→ 해결(2026-08-22): 두 항목 모두 리드 비준 완료** — 아래 "리드 결정 확정" 참조.
plan.md `[NEEDS CLARIFICATION]` 마커는 폐지되어 DECIDED 항목으로 대체됐다.

## 리드 결정 확정 (2026-08-22)

1. **FX-1 라이브 디렉터리 링크 처분 = 제거 + 진행줄 가시화 — 비준됨.** 결정적 근거:
   실측된 `MkdirAll` fast-path 링크 추적(보존 시 배포 기록이 사용자 외부 트리로 유입,
   차단하려면 범위 밖 배포 측 게이팅 필요).
2. **FX-3b 글로브 매치 dangling = 제거 — 비준됨.** dangling에는 데이터 손실 표면이
   없다(대상 부재) — 관리 네임스페이스 위생. 새 [HARD] 회귀면(사용자 dangling 링크
   제거)은 양극 픽스처로 고정(AC-CSL-002 ↔ AC-CSL-007).
3. **판별자 라우팅 정정**: "t81(가) 소관"은 오류 — t81(가)는 자기 범위에서 제외하며
   (v0.7.0 D1 수정 (b)), 소유는 어느 형제 SPEC도 아닌 **리드 큐의 후속 카드 후보**.
   위 보충 반영의 라우팅 충돌 기록도 이 결정으로 해결된다.
4. **Implementation Kickoff Approval — 조건부 허가(리드 전달, 2026-08-22)**: 리드가
   plan-audit를 선승인했고 런 진입은 감사 판정을 따른다(plan-audit PASS 조건).
   plan_status는 audit-ready 유지.

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase 산물 작성 완료: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
(Tier M 3종 + progress.md — 매 티어 발행).

- 요구사항: **12건**(`REQ-CSL-001` … `REQ-CSL-012`) — Tier M 상한 16 이내.
- 수용기준: **11건**(`AC-CSL-001` … `AC-CSL-011`) — Tier M 상한 16 이내. 12 REQ 전부
  §D.2 추적표에서 커버.
- 픽스처 형태: **5종**(FX-1 라이브 디렉터리 링크 / FX-2 라이브 파일 링크 / FX-3 dangling
  3배치 / FX-4 실디렉터리 대조 / FX-5 hns 사용자 소유 대조) — spec·plan·acceptance
  3면 동일 수치(t81 D3 이관).
- 마일스톤: 4개(M1 링크 분기+처분 → M2 5형태 양극 테스트 → M3 계약 테스트 → M4 회귀
  스윕+플랫폼) — 결정 가역성 순(처분 의미론 최우선).
- t81 D2/D3/D4 이관: REQ-CSL-010(형태 일치)/§F 수치 3면 일치/REQ-CSL-011(비공허) +
  acceptance.md 전 단언의 ≥2축 결합.
- 개발 방식: TDD(quality.yaml) — M1의 RED는 Run D 재현의 Go 테스트 전환.
- 미해명 마커: **0건** — 유일 마커(처분 비준)는 2026-08-22 리드 결정으로 DECIDED 항목에
  대체됐다(위 "리드 결정 확정").

plan_status: audit-ready
plan_complete_at: 2026-08-22

### Plan-audit 진행 (iter-1 FAIL → 수정 적용, iter-2 대기)

- Iteration 1 (2026-08-22): **FAIL 0.92** —
  `.moai/reports/plan-audit/SPEC-CLI-CLEAN-SYMLINK-001-review-1.md`
  (must-pass MP-1~MP-7 전부 PASS; blocking 2건 = D1 하위-카운트 드리프트 · D2 계약 용어
  미고정, optional D3·D4. 4차원: Clarity 0.75 / Testability 1.0 등, 조화평균 0.92 —
  Tier M 임계 0.80은 충족하나 blocking-class가 판정을 지배).
- 수정 적용 (v0.1.1, 2026-08-22): D1 — spec.md §B 서두 "배치 2곳/두 진입 팔"을 3배치로
  정정 + HISTORY 기록. D2 — "보유" = 렌더링 후 기록 경로로 고정(AC-CSL-009 Then-1 ·
  plan M3; settings.json/.tmpl 사실을 반증 인용으로 명기). D3 — AC-CSL-003에 Then-4
  진행줄 단언 추가. D4 — AC-CSL-001 RED 라벨 귀속 정정(재실행 = 코드 추적, M1 RED에서
  관측). 수치 **12 REQ / 11 AC / 5 형태 불변**(AC-CSL-003 단언 3→4는 AC 개수가 아닌
  단언 수 — 카운트 규율 밖).
- 다음: iter-2 delta 재감사(D1·D2 중심) → PASS 시 런 진입(리드 조건부 Kickoff 승인 —
  plan-audit PASS 조건).

## §E.2 Run-phase Evidence

Run-phase 수행: manager-develop (cycle_type=tdd), 워크트리 `.claude/worktrees/t173`
(branch `WT-clean-links`). 모든 증거는 이 트리에서 이 에이전트가 직접 실행해 관측한
verbatim 출력이다 (baseline-attribution: run 시작 HEAD `d19f849be`, run 중 로컬 커밋
순차 적립).

### M1 — 링크 전용 분기 + 형태별 처분

- **RED (AC-CSL-001, E8 — 구현 이전 verbatim)**: 명령
  `go test ./internal/cli/update/deploy/ -run 'TestCleanMoaiManagedPaths_DanglingSymlinkAtNonGlobRoot' -count=1 -v`.
  관측된 실패 (전체 출력은 run 트랜스크립트에 verbatim 보존):
  ```
  deploy_symlink_test.go:90: dangling symlink still present after clean (Lstat err = <nil>, want IsNotExist)
  deploy_symlink_test.go:96: progress output has no line naming .claude/agents/moai as a symlink:
        ✓ Skipped .claude/agents/moai (not found)          ← Run D의 Skip 재현
  deploy_symlink_test.go:101: deploy-side MkdirAll(...) failed: mkdir .../.claude/agents/moai: file exists   ← EEXIST 재현
  deploy_symlink_test.go:110: (재실행 MkdirAll 동일 EEXIST — 재실행 루프 폐쇄 실패)
  --- FAIL: TestCleanMoaiManagedPaths_DanglingSymlinkAtNonGlobRoot (0.00s)
  ```
  AC-CSL-001 RED 기준(단언 1·3·5 실패) 그대로 + 단언 2(진행줄) 실패.
- **RED (M2 형태 일괄, 구현 이전)**: 같은 경계의 `-run 'Symlink|...|UserOwned...' -v`
  실행에서 링크 8종 전부 FAIL (glob dangling 잔존·config dangling EEXIST·file-root
  dangling 잔존·live dir/file 링크 진행줄 부재·target-stat 귀속 상이), 대조군 2종
  (FX-4 실디렉터리·FX-5 hns 미접촉) PASS — 양극 확인.
- **GREEN**: `go test ./internal/cli/update/deploy/ -count=1` →
  `ok github.com/modu-ai/moai-adk/internal/cli/update/deploy 0.559s` (기존 테스트
  전량 포함 녹색 — REQ-CSL-006 실디렉터리 비회귀). 구현: `backupThenRemove` 판정
  `os.Lstat` + `fs.ModeSymlink` 검사를 IsDir 이전에 배치 (REQ-CSL-001), 링크 분기
  `removeSymlink`가 형태별 처분 실행 (dangling/디렉터리/파일), 비-글롭 사전검사
  Lstat 전환 (dangling이 "Skipped (not found)"로 넘어가지 않게 — 의도된 출력 변경),
  진행줄 토큰 `dangling symlink` / `directory symlink, target untouched` /
  `file symlink(, target bytes backed up)`.
- **정식 검증 (M1 닫힘)**: `go test ./internal/cli/update/... -count=1` → 6패키지
  전부 `ok` · `go vet ./internal/cli/...` → exit 0 · `golangci-lint run
  ./internal/cli/update/...` → `0 issues.` · `GOOS=windows GOARCH=amd64 go build ./...`
  exit 0 · `GOOS=linux GOARCH=amd64 go build ./...` exit 0 · `gofmt -l` 신규 파일
  0건 (`deploy_test.go` 1건은 본 run 미수정 파일의 선행 드리프트 — baseline).
- 커밋: M1 커밋 = 이 커밋 (deploy.go + deploy_symlink_test.go + 4 산물 frontmatter
  `draft → in-progress`). deploy_symlink_forms_test.go는 M2 커밋 소관으로 미커밋
  유지 (RED 관측은 위에 기록됨 — 테스트와 구현이 함께 녹색 도달).

### M2 — 5형태 픽스처 양극 테스트 (deploy_symlink_forms_test.go)

- **RED**: M2 형태 일괄 RED는 M1 GREEN 이전에 관측했다 (위 M1 RED 항 — 링크 8종
  FAIL / 대조군 2종 PASS). `TestBackupThenRemove_LiveFileSymlinkTemplateCarried`
  1건는 신규 3-값 시그니처에 의존해 GREEN과 함께 컴파일되므로 독립 RED가 불가능
  (시그니처 변경 자체가 실패 테스트들에 이끌린 GREEN 리팩터링) — 해당 하위 처분
  (carried 파일 링크)의 행위 클래스는 AC-CSL-001/005 RED가 커버하며, 본 단위
  테스트는 그 표면화로 기록한다 (tdd 스킬 red-flag에 대한 정직한 고지).
- **GREEN**: `go test ./internal/cli/update/deploy/ -count=1` →
  `ok github.com/modu-ai/moai-adk/internal/cli/update/deploy 0.452s`.
  AC-CSL-002…007 전부 통과 (must-flag 5종 + must-not-flag 대조 2종).
- **AC-CSL-010 자점검 — 형태 일치 (REQ-CSL-010)**:

  | 테스트 (AC) | 제품 형태 | 픽스처 형태 | 일치 |
  |---|---|---|---|
  | DanglingSymlinkAtNonGlobRoot (001) | dangling, 비-글롭 뿌리 | absent 경로 지목 링크 at `.claude/agents/moai` | ✓ |
  | DanglingSymlinkAtGlobMatchName (002) | dangling, 글로브 매치 | absent 지목, 이름 `moai-dangling-custom` | ✓ |
  | DanglingSymlinkAtConfigRoot (003) | dangling, config 뿌리 | absent 지목 at `.moai/config` | ✓ |
  | DanglingSymlinkAtFileRoot (§D.3) | dangling, 파일 루트 | absent 지목 at settings.json | ✓ |
  | LiveDirectorySymlinkRoots (004) | 라이브 **디렉터리** 링크 | 실재 dir(센티널 포함) 지목 디렉터리 링크, 비-글롭+글로브 양배치 | ✓ |
  | LiveFileSymlinkSettings (005) | 라이브 **파일** 링크 | 실재 파일 지목 파일 링크 (디렉터리 링크 아님) | ✓ |
  | BackupThenRemove_TemplateCarried | 파일 링크, carried 경로 | 실재 파일 지목 + 템플릿 carried | ✓ |
  | RealDirectoryControl (006) | 링크 없음 (must-not-flag) | 실 디렉터리 + 비관리 파일 | ✓ |
  | UserOwnedNamespaceUntouched (007) | 사용자 소유(내부 링크 포함) | hns-mine + 내부 dangling badlink | ✓ |

- **AC-CSL-010 자점검 — 축 결합 (REQ-CSL-011)**: 모든 링크 단언 ≥2축 —
  001: 링크 부재+진행줄+MkdirAll+재배포+재실행 (5축) · 002: 링크 부재+진행줄+
  형제 스킬 제거/재배포 (3축) · 003: 링크 부재+재배포+진행줄 (3축) · 004:
  실디렉터리 전환+외부 센티널 무결+진행줄 (+백업 0건은 보조축) · 005: 백업
  바이트=센티널+최종 실재 파일+복원 흐름+진행줄 (4축) · 006: 백업 존재+진행줄
  부재+재배포 (3축) · 007: 내용 무결+badlink 잔존+백업 0건 스캔 (3축).
  bare "백업 수 == 0" 단독 단언: **0건** (백업 카운트는 004 보조축·007 3축
  결합으로만 등장).
- 파일 뿌리 dangling(도시에 gap 3)은 `DanglingSymlinkAtFileRoot`로 실측 전환 —
  clean이 링크를 제거+진행줄, deploy 쓰기 성공으로 폐쇄.

### M3 — 계약 테스트 (독립 검증 축)

- **AC-CSL-009 교차 계약** (`deploy_contract_test.go`, 외부 `deploy_test` 패키지 —
  deploy의 leaf 성질 보존): `template.EmbeddedTemplates()`를 deployer와 동일하게
  걷는다(디렉터리 스킵·`.tmpl` 접미 제거 — deployer.go 렌더링 규칙) → 렌더링 후
  기록 경로 집합. 비-글롭 루트 7개 + config 8번째 뿌리 전부 "정확한 파일 또는
  하위 파일 ≥1 보유"로 커버, 글로브 `.claude/skills/moai*` 매치 ≥1. 루트 1
  `.claude/settings.json`은 `settings.json.tmpl`→렌더링 경로 판정으로 통과(원시
  경로 판독 반증 — 도시에 §1.1(b) 준거). **본 트리에서 계약 성립 — 이 테스트는
  t81(가) `.agents/` 추가 후 release/v3.1.3 통합 시점 재실행 대상의 드리프트
  가드** (가드 특성상 현행 녹색이 정상; §D.7-1).
- **AC-CSL-008 순서 독립** (`deploy_symlink_order_test.go`, `backupThenRemove`
  단위 경계 — 새 seam 없음): 실디렉터리 A(비관리 파일 포함, 템플릿 보유 이름) +
  A 지목 라이브 디렉터리 링크 B(글로브 매치 이름)를 (a) B→A, (b) A→B 양순서
  처리 — (b)에서는 A 제거 후 B가 dangling으로 강등되어 dangling 처분으로 제거된다
  (spec §D.3 둘째 경계 사례의 직접 입증). 양순서 최종 상태 동일: A 실디렉터리
  재배포+템플릿 파일, B 부재, A 비관리 파일 백업 존재. 4축 전부 양순서 일치.
  RED 소급: (b)의 B-제거 다리는 M1이 도입한 dangling 처분에 의존 — 구형 코드의
  같은 경로(Stat ENOENT 무소식 no-op, 도시에 §2.4)에서는 B가 잔존해 bGone=false
  로 실패한다. 구형 시그니처(2-값)와의 컴파일 결합 때문에 관측 RED 대신 도시에
  실측 no-op 경로 + M1 RED로부터의 구성적 귀속으로 기록한다(미관측 고지).
- **GREEN**: `go test ./internal/cli/update/deploy/ -count=1` → ok 0.552s ·
  `go test ./internal/cli/update/... -count=1` → 6패키지 ok · lint `0 issues.`.

### M4 — 회귀 스윕 + 플랫폼

- **AC-CSL-011 스킵 경로** (`TestMakeSymlink_SkipsWhenCreationFails`): os.Symlink의
  결정적 실패(EEXIST 충돌)에서 헬퍼가 t.Skip으로 건너뛰는가 — reached-플래그가
  스킵 시 미도달임으로 검증 (서브테스트 내 할당 도달 안 함). 전체 링크 테스트가
  이 헬퍴를 경유하므로 비-darwin 매트릭스에서 생성 불가 시 전원 skip-not-fail.
- **루프/자기지목 스팟** (`TestCleanMoaiManagedPaths_SelfReferentialLinkSpotCheck`,
  §D.3 셋째 경계): 자기지목 링크는 형태 판정 Stat이 ELOOP으로 실패 → "stat symlink
  target" 귀속 에러로 유종 종결(무한 순회 없음, 제거 전 중단 — 미분류 형태의
  loud-failure 처분). 잔여위험으로 §E.3에 고지.
- **커버리지**: 최초 측정 82.7% (E3 목표 85% 미달) — per-function 분석으로
  `InventoryManagedPaths` **0%**(선행 공백, 본 run 미수정 분야)과 미커버 분기
  식별. `deploy_inventory_test.go`(인벤토리 계약 + carried 실화일 제거 + config
  루트 Lstat 에러 + 파일링크 백업실패 중단 순서) 추가 후 **91.6%**
  (`ok ... coverage: 91.6% of statements`, coverprofile 직독).
- **파괴적 사이트 레지스트리 연쇄 (범위 공개)**: 지정 스코프 스위트
  `go test ./internal/cli/ -count=1`가 `TestDestructiveTargetRegistry_CoversAllSites`
  실패로 신규 `removeSymlink`(4개 os.RemoveAll 사이트)의 미등록을 포착 —
  update-트리 한정 실행으로는 도달하지 않는 가드다.
  `internal/cli/update_destructive_registry.go`에 removeSymlink 행 1건(보호 근거:
  링크 엔트리만 제거·대상 구조적 미도달 + 파일링크 분기 백업-선행) 추가, 헤더
  12행/22사이트 갱신. deploy.go의 파괴적 사이트 색인이라는 기계적 결합 산물로
  본 SPEC 봉투 내 연쇄로 판정해 수행(카탈로그-해시 재생성 패턴과 동형) — 오케스트레이터
  검토용 명시 공개. 수정 후 전체 스위트 재실행:
  `go test ./internal/cli/ -count=1` → `ok ... 323.244s` (FAIL 0건).
- **문서 영향 조사**: 배포 템플릿 전체에서 update 진행 출력 예시 그레프
  (`Skipped.*not found` / `backed up.*unmanaged`) — **양쪽 exit 1, 예시 없음**.
  sync-phase 문서 갱신 후보 없음.
- **§E 전량 (plan.md §E 대로)**: deploy 패키지 `ok` · update 트리 6패키지 `ok` ·
  `-run 'Symlink|Clean' -count=1 -v` 29 PASS · `go vet ./internal/cli/...` exit 0 ·
  `golangci-lint run ./internal/cli/update/...` `0 issues.` ·
  `golangci-lint run ./internal/cli/` `0 issues.` · coverage 91.6% ·
  `go build ./...` / `GOOS=windows GOARCH=amd64` / `GOOS=linux GOARCH=amd64` 전부
  exit 0 · E4 서브에이전트 경계 그레프 exit 1(매치 0건 — 해당 없음).

### E1 — AC 이원 행렬 (11/11)

| AC | 판정 | 검증 | 근거 테스트 |
|---|---|---|---|
| AC-CSL-001 (P0) | PASS | 5단언 전부 | DanglingSymlinkAtNonGlobRoot |
| AC-CSL-002 (P0) | PASS | 3단언 | DanglingSymlinkAtGlobMatchName |
| AC-CSL-003 (P1) | PASS | 4단언(진행줄 포함) | DanglingSymlinkAtConfigRoot |
| AC-CSL-004 (P1) | PASS | 양배치 4단언 | LiveDirectorySymlinkRoots |
| AC-CSL-005 (P1) | PASS | 4단언 | LiveFileSymlinkSettings |
| AC-CSL-006 (P2) | PASS | 3단언 | RealDirectoryControlNoSymlinkLines |
| AC-CSL-007 (P2) | PASS | 3단언 | UserOwnedNamespaceUntouched |
| AC-CSL-008 (P1) | PASS | 양순서 4축 | LinkedRootsOrderIndependence |
| AC-CSL-009 (P1) | PASS | 루트 8+글로브 1 | CleanSetCoveredByDeployment |
| AC-CSL-010 (P1) | PASS | 형태 표·축 표기 | §E.2 M2 자점검 표 |
| AC-CSL-011 (P2) | PASS | reached-플래그 | MakeSymlink_SkipsWhenCreationFails |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-22
run_commit_sha: 73e451ef7             # M4(런 최종 구현 커밋) — backfill 커밋에서 확정
run_status: complete                  # M1→M4 전 마일스톤 완료, AC 11/11
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0       # SPEC 디렉터리 밖 변경은 internal/cli/update/deploy/* (봉투 내) + update_destructive_registry.go(기계적 연쇄, M4 항목에 공개) — PRESERVE 위반 0건
l44_pre_commit_fetch: n-a-no-push     # factory lane — 푸시 없음, 리드가 release/v3.1.3에 통합
l44_post_push_fetch: n-a-no-push
new_warnings_or_lints_introduced: 0   # 패키지 lint 0 issues (baseline 0에서 신규 0), gofmt 신규 파일 전부 정상 (deploy_test.go 선행 드리프트 1건은 미수정 baseline)
cross_platform_build:
  darwin_arm64: pass                  # go build ./... exit 0
  windows_amd64: pass                 # GOOS=windows GOARCH=amd64 go build ./... exit 0
  linux_amd64: pass                   # GOOS=linux GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 11             # 코드 7 (deploy.go + 테스트 5 + 레지스트리 1) + SPEC 산물 4
m1_to_mN_commit_strategy: per-milestone feat commits M1..M4 on WT-clean-links, no push, explicit pathspec staging
residual_risk:
  - 자기지목/ELOOP 링크는 "stat symlink target" 에러로 update 중단(제거 아님) — SPEC이 분류하지 않은 형태의 보수적 처분. dangling-등가 제거로 전환하려면 syscall 결합(ELOOP 구분)이 필요해 본 카드 범위 밖; 후속 카드 후보.
  - 배포 경로(deployer)의 링크 인식은 여전히 없음(spec §E 명시적 제외) — 관리 뿌리 밖에서 사용자가 배포 목적지에 링크를 심는 형태는 불변.
  - non-darwin 실측 없음(darwin/arm64 단일 호스트) — REQ-CSL-012의 skip 경로가 CI 매트릭스에서 관측될 것.
```

**커밋 목록 (run-phase)**: M1 `be0959428` · M2 `dda35151b` · M3 `e78fd8f5d` ·
M4 `<this-commit>` — 전부 `WT-clean-links`, 미푸시.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-22
sync_commit_sha: pending-backfill-SPEC-CLI-CLEAN-SYMLINK-001  # backfill 커밋에서 확정 (spec-frontmatter-schema.md D3)
sync_status: complete
changelog_entry_position: CHANGELOG.md [Unreleased] > Fixed (최상단)
frontmatter_status_transitions:
  spec_md: in-progress -> completed   # sync 커밋에서 단일 전이 (3-phase close)
  updated_field: unchanged            # 이미 2026-08-22
docs_surfaces_unchanged: README 4-locale + docs-site — 배포 템플릿에 moai update 출력 예시 없음 (run-phase grep 2건 모두 exit 1)
route: factory-lane                   # WT-clean-links 미푸시 — 리드가 release/v3.1.3에 통합
```

## §F Phase 4 Mode Selection

Logged by the orchestrator (lane-9) before the first run-phase Agent() spawn.

**Input parameters**: tier M; scope ~3-4 files (deploy.go clean path + tests); domains 1 (Go CLI code); language mix Go; concurrency benefit LOW (single coupled code path — classifier → disposition → tests all in one seam); Agent Teams N/A.

**Mode evaluation**: direct — not selected (>trivial: new branch semantics + 11-AC test suite). serial — **selected** (coding-heavy Tier M, single-agent dependency chain M1→M4). fanout — not selected (1 domain). sweep — not selected (3-4 files, new-code work).

**Decision: serial** (manager-develop, cycle_type=tdd)

**Justification**: the change is one classifier seam in backupThenRemove plus its disposition semantics and test fixtures — a tightly coupled single-author surface. Kickoff: lead conditional grant 2026-08-22 + plan-audit iter-2 PASS 1.00 (condition satisfied).

**Plan Audit Gate skip record**: most recent verdict PASS (iter-2, 1.00 ≥ Tier M 0.80); plan-artifact hash unchanged since the verdict (this §F addition and the audit-report append are not hash subjects); skip-eligible per the three-condition contract.
