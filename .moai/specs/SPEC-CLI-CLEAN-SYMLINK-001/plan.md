---
id: SPEC-CLI-CLEAN-SYMLINK-001
title: "Plan — moai update 청소 경로의 심볼릭 링크 인식"
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

# plan.md — SPEC-CLI-CLEAN-SYMLINK-001

> 수치 기준(3면 일치 — spec.md §C/§F, acceptance.md): **요구사항 12건(REQ-CSL-001…012),
> AC 11건(AC-CSL-001…011), 픽스처 형태 5종(FX-1 라이브 디렉터리 링크 / FX-2 라이브 파일
> 링크 / FX-3 dangling[비-글롭 뿌리·글로브 매치·config 뿌리 3배치] / FX-4 실디렉터리 대조 /
> FX-5 hns 사용자 소유 대조)**. 본 문서의 마일스톤 닫힘 조건은 이 수치를 그대로 인용한다.

## §A. Context

- 카드 t173("moai update 청소 경로 링크 인식"). 워크트리 `.claude/worktrees/t173`
  (branch `WT-clean-links`, HEAD `18f7cfc19` = origin/main + 도시에 커밋, tree clean).
- **연구는 사전 수집 완료**: `.moai/reports/t173/measurements.md`(508줄, 커밋됨)가 본
  SPEC의 research.md를 대체한다 — 분기 추적(§1, file:line 전부 이 트리에서 직독 확인),
  청소 뿌리 인벤토리(§2), 4회 재현 매트릭스(§3), t81 D2-D4 원문(§4), 미측정 gap(§5).
  run-phase에서 새로 측정해야 할 항목은 §C에 열거했다.
- 수정 표면: `internal/cli/update/deploy/deploy.go`(분류 + 진행줄)와 그 테스트 파일.
  `internal/template/deployer.go`는 건드리지 않는다(spec.md §E).
- 착지: t81(가)와 같은 release/v3.1.3. 교차 계약(REQ-CSL-009)은 두 카드 모두 착지한
  뒤 통합 시점에 재확인한다(acceptance.md §D.7).

## §B. Known Issues (이 계획이 닫는 것)

1. **Run D 결함**: dangling 링크 at 비-글롭 뿌리 → clean "Skipped (not found)"(deploy.go:139-146,
   링크 잔존) → deploy `MkdirAll` EEXIST(deployer.go:189) → 부분 파괴 중단 + 재실행 영구 루프.
   본 계획의 M1이 이것을 Go 테스트로 재현하고(RED) 링크 분기로 닫는다(GREEN).
2. **무인식**: 라이브 디렉터리 링크(백업 0건·링크만 제거·대상 무사·재배포, Run B/D)와 라이브
   파일 링크(대상 바이트 백업·복원, Run B)가 진행 출력에 전혀 나타나지 않는다 → M1/M2가
   형태를 이름붙인 진행줄로 가시화한다.
3. **글로브 dangling 영구 잔존**: `moai-dangling-custom` 무소식 no-op(deploy.go:372-375) →
   M1이 제거+진행줄로 바꾼다(동작 변경 — §D 비준 항목).
4. `.moai/config` 뿌리의 링크 의미론이 코드 추적만으로 남아 있다(도시에 gap 4) → M2가
   Go 테스트로 실측 전환.

## §C. Pre-flight

- 개발 방식: **TDD**(quality.yaml `tdd_settings.red_green_refactor: true`,
  `test_first_required: true`, 커밋당 최소 80%, 패키지 목표 85%).
- 기존 재현: Run A~D는 바이너리 E2E로 이미 실측됨(도시에 §3). M1의 RED는 이것을 Go 테스트로
  옮기는 것이다 — 새로운 조사가 아니라 전환.
- run-phase에서 **신규 실측 필요** 항목(도시에 §5 gap — 가정 금지):
  - gap 2: dangling 재실행 루프의 직접 확인(두 번째 실행도 성공) — AC-CSL-001 단언 (5).
  - gap 3: 파일 뿌리 dangling(settings.json)의 `atomicWriteFile` rename 대체 — §D.7.
  - gap 4: `.moai/config` dangling — AC-CSL-003이 Go 테스트로 전환.
  - gap 5: 심볼릭 링크 테스트의 비-darwin 동작 — REQ-CSL-012의 t.Skip 경로(CI 매트릭스).
- 테스트 격리: `t.TempDir()` 전용. 프로젝트 루트·`~/.claude` 미접촉(CLAUDE.local.md §6, §13).
- 바이너리 재현 금지: 본 카드의 검증은 Go 테스트 경계에서 수행한다(도시에 gap 6의 반).
  E2E 재현은 이미 4회 예산으로 완료됨.

## §D. Constraints (구속 + 비준 대기)

1. **[HARD] 링크 전용 분기 선순위** — 분류는 `os.Lstat`로, `fs.ModeSymlink` 검사가 IsDir
   판정 **이전**에. 링크는 어떤 형태든 파일/디렉터리 분기에 흡수되지 않는다(카드 구속).
2. **[HARD] 픽스처 형태 일치** — 테스트 픽스처의 링크 형태(파일/디렉터리/dangling)는 시험
   대상 제품 형태와 일치(REQ-CSL-010, t81 D2).
3. **[HARD] 비공허 단언** — 링크 단언은 ≥2 관측 축 결합, bare "백업 수 == 0" 금지
   (REQ-CSL-011, t81 D4).
4. **[HARD] 수치 3면 일치** — 12 REQ / 11 AC / 5 형태가 spec·plan·acceptance에서 동일.
   개정 시 마일스톤 닫힘 조건이 따라온다(t81 D3).
5. **처분 — 리드 비준 완료 (DECIDED 2026-08-22)** — 근거와 대안 기각은 spec.md §B.1.
   - **DECIDED · FX-1(라이브 디렉터리 링크) = 제거 + 진행줄 가시화.** 결정적 근거: 이
     트리에서 실측한 `MkdirAll` fast-path의 링크 추적 — 보존은 배포 기록이 사용자 외부
     트리로 스며드는 경로를 열고, 그것을 막는 배포 측 게이팅은 범위 밖. 회귀면 고지(기록으로
     남긴다): 링크를 유지하려던 사용자에게 결과는 여전히 파괴적이나 종전과 동일(새 파손
     없음)이고 이제 진행줄로 보인다. 뒤집으려면 본 SPEC 수정(amendment) + 배포 측 변경이
     필요하다.
   - **DECIDED · FX-3b(글로브 매치 dangling) = 제거.** 근거: dangling에는 데이터 손실
     표면이 없다(대상 부재) — 관리 네임스페이스(`moai*`) 위생. **[HARD] 새 회귀면: 사용자가
     심어 둔 글로브 매치 dangling 링크가 제거된다(종전 영구 잔존).** 양극 픽스처로 고정 —
     must-flag AC-CSL-002(제거+진행줄) ↔ must-not-flag AC-CSL-007(hns 사용자 소유 dangling
     `badlink` 잔존: 제거가 관리 네임스페이스에만 적용됨을 입증).
   - 결정에 따른 런 확인 항목(probe/verify)은 §C의 기존 목록이 그대로 담당한다(gap 2
     재실행 루프, gap 3 파일 뿌리 dangling, gap 5 비-darwin) — 별도 신설 없음.
6. **대기 순서 제약(휴면)** — 보존 처분이 활성화되는 경우에만 "미러 뿌리를 대상 뿌리보다
   먼저" 가 하중을 받는다(spec.md §B.2). 제거 처분 아래서는 REQ-CSL-008의 순서 독립이
   그 위험을 해소한다.
7. 진행줄 문구의 안정 토큰: 각 링크 진행줄은 **링크 경로 + "symlink" 계열 토큰**을 포함해야
   한다(단언이 grep 가능해야 한다). 정확한 문구는 run-phase 자유.

## §E. Self-Verification (run-phase §E 증거 명령)

```bash
# M1~M3 단언 ( affected packages only — CLAUDE.local.md §4 )
go test ./internal/cli/update/deploy/... -count=1
go test ./internal/cli/update/... -count=1
# 정밀 재실행 (캐시 무시)
go test ./internal/cli/update/deploy/... -run 'Symlink|Clean' -count=1 -v
# 정적 검사
go vet ./internal/cli/...
golangci-lint run ./internal/cli/update/...
# 커버리지 (목표 85%, 커밋당 하한 80%)
go test ./internal/cli/update/deploy/... -cover -count=1
```

전체 스위트는 CI가 판정한다(로컬 전량 실행 금지 — CLAUDE.local.md §4).

## §F. Milestones (역순정 결정 가능성 순 — 의미론 결정 → 테스트 표면 → 계약 → 기계적)

### M1 — 링크 전용 분기 + 형태별 처분 (의사결정 밀도 최고 → 최우선)

- RED: Run D를 Go 테스트로 전환 — `t.TempDir()` 프로젝트에 `.claude/agents/moai` dangling
  링크를 심고 청소+배포를 실행하면 현행 코드에서 EEXIST로 실패하는지 확인(AC-CSL-001의
  Given-When-Then 그대로).
- GREEN: `backupThenRemove`(deploy.go:371-399)의 분류를 `os.Lstat`으로, `fs.ModeSymlink`
  검사를 IsDir 이전에 배치. 링크 분기 안에서 형태 판정(대상 Stat: 부재=dangling / 디렉터리 /
  파일)과 처분 실행(제거 + 표준 토큰 진행줄; FX-2는 독해 백업 선행). 비-글롭 사전검사
  (:139-150)도 Lstat으로 — dangling이 "Skipped (not found)"로 넘어가지 않게.
- `.moai/config` 뿌리(:168-182)는 `backupThenRemove`를 공유하므로 자동 승계(단위 테스트로 확인).
- 닫힘 조건: **AC-CSL-001 전 단언(5개) 통과 + REQ-CSL-001/002/005 구현 + 기존
  `deploy` 패키지 테스트 전량 녹색(실 디렉터리 비회귀, REQ-CSL-006)**.

### M2 — 5형태 픽스처 양극 테스트 (M1 결정을 표면화)

- FX-1 라이브 디렉터리 링크(AC-CSL-004): 글로브 배치 + 비-글롭 배치 양쪽.
- FX-2 라이브 파일 링크(AC-CSL-005): settings.json 형태 — 파일 링크로 심는다(디렉터리 링크
  금지, REQ-CSL-010).
- FX-3 dangling: 비-글롭(AC-CSL-001, M1에서 이미 RED-GREEN) + 글로브(AC-CSL-002) +
  config 뿌리(AC-CSL-003).
- FX-4 실디렉터리 대조(AC-CSL-006, must-not-flag 극) + FX-5 hns 사용자 소유 대조
  (AC-CSL-007, must-not-flag 극) — 내부에 dangling 링크를 심어 미접촉을 입증.
- 각 단언 ≥2 관측 축 결합(REQ-CSL-011). 닫힘 조건: **5형태 픽스처 전부 + AC-CSL-002…007
  통과 + AC-CSL-010 자점검(형태 일치 표·축 표기) 완료**.

### M3 — 계약 테스트 (독립적 검증 축)

- 교차 계약(AC-CSL-009): 임베디드 템플릿 FS 대 `ManagedCleanTargets` — 모든 비-글롭 루트는
  템플릿이 보유, 모든 글로브 패턴은 ≥1 매치. 발산 시 실패. t81(가)의 `.agents/` 추가 후
  release/v3.1.3 통합 시점에 재실행(§D.7).
- 순서 독립(AC-CSL-008): 링크로 연결된 두 청소 대상(예: 글로브 루트 A의 실디렉터리 + A를
  가리키는 라이브 디렉터리 링크 B)을 `backupThenRemove` 수준에서 양쪽 순서로 처리하고
  최종 상태 동일성 비교(새 seam 불필요 — 기존 단위 경계).
- 닫힘 조건: **AC-CSL-008, AC-CSL-009 통과**.

### M4 — 회귀 스윕 + 플랫폼 (기계적 — 최후)

- 심볼릭 링크 생성 헬퍼가 `os.Symlink` 실패 시 `t.Skip` 반환(AC-CSL-011, REQ-CSL-012).
  비-darwin CI 매트릭스(linux/windows)에서 skip-not-fail 확인.
- §E 명령 전량 + 커버리지 ≥85%(deploy 패키지). 진행 출력 문구가 사용자 문서(update 출력
  예시)에 등장한다면 sync-phase에서 갱신 후보로 보고(본 카드에서 문서 수정은 강제 아님).
- 닫힘 조건: **AC-CSL-011 + §E 전 명령 통과 + 커버리지 기준 충족**.

## §G. Anti-Patterns (t81 D2-D4 + 도시에 교훈의 행동화)

- **AP-1 (D2) 형태 어긋난 픽스처**: 파일 링크로 디렉터리 링크 제품 형태를 시험하지 않는다.
  라이브 파일 링크와 라이브 디렉터리 링크는 다른 분기를 탄다.
- **AP-2 (D3) 수치 드리프트**: "5형태"를 "4형태"로 줄여 적거나 닫힘 조건이 개정을 못
  따라오게 하지 않는다. 개정 시 3면 동시 갱신.
- **AP-3 (D4) 공허 단언**: "디렉터리 링크 백업 수 == 0"을 단독 단언으로 쓰지 않는다 —
  WalkDir-스킵 덕에 구현과 무관하게 항상 참이다. 링크 제거·대상 무사·메시지와 결합.
- **AP-4 배포 쪽 변경 유혹**: 보존 처분을 지지하려고 `deployer.go`를 함께 고치려는 것 —
  범위 밖(spec.md §E). 
- **AP-5 바이너리 E2E 재탕**: 이미 4회 예산으로 실측 완료. run-phase 검증은 Go 테스트로.
- **AP-6 t.Setenv/HOME 오염**: 테스트는 `t.TempDir()` 프로젝트 루트에서만(CLAUDE.local.md §6, §13).

## §H. Cross-References

- spec.md(처분 표 §B, 요구사항 §C) / acceptance.md(AC 매트릭스 §D) / progress.md(§E.1).
- 도시에: `.moai/reports/t173/measurements.md`. t81 감사: `.claude/worktrees/t81/.moai/reports/t81/plan-audit-iter4.md`.
- `internal/cli/CLAUDE.md`(서브에이전트 경계·경로 규율), `internal/template/CLAUDE.md`(템플릿 계약).
- quality.yaml(TDD 설정), CLAUDE.local.md §4(검증 규율)·§6(테스트 격리).
