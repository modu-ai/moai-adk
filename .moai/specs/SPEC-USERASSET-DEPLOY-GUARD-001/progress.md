# progress.md — SPEC-USERASSET-DEPLOY-GUARD-001

status: draft
card: t1591 (lane-20)
tree: .moai/worktrees/t1591 @ WT-user-asset-bundle

## Card Provenance

- 리더 배차 2026-10-08 (Class C — plan→run→sync 전체). 본 SPEC은 카드의 13개 원장 항목을 그대로 범위로 한다.
- 원장 출처: t1561 게이트 P2 4건 + 배포 표면 (리더 발행 2026-10-08) + t1547 릴레이 r10 (원장 8) + t1574 3차 릴레이 (원장 9·10) + t1498 게이트 원장 계통 + lane-15 (원장 11)·lane-5 (원장 12)·lane-7 대장 (원장 13) 릴레이.
- 계통: t1509 "사용자 자산·배포 축" 보존 계약 후속 (릴레이 누계 7건). t1560 게이트 이관분은 SPEC-PROGRESS-RECORD-IO-001 원장에 상호참조 존속 — 본 카드 비소관.
- 비소관 축: t1547 도메인 정합 (update/reconcile.go:169/:590/:204/:184, update_template_sync.go:541 — 리더 보유 후속 카드), 형제 카드 t1594 (t1547 r5-r12 잔여).

## Gate-Relay Provenance (라운드 1-13)

- 누적: t1547기 턴종료 게이트에서 라운드 1-13. **오버레이 재현만 수행, 전체 CI/windows 미검증은 게이트 스스로 선언.**
- 수렴: r8에서 8좌표 수렴 선언, r9-13이 잔여 추가.
- r4 검증 라인 (역사적 verbatim, 고정 HEAD 40e6c310):
  `go test -p 2 -overlay /tmp/t1547-review-overlay.json ./internal/userassets ./internal/cli -run '^TestReview(FIFOInstall|RemovePendingBundle|InitRetryAfterInstallFailure|CodexInstalledSchemaReference)$' -count=1 -timeout 120s -v` → `--- FAIL` 4케이스.
- 스테일 경고: `/tmp/t1547-review-overlay.json` 등 이전 세대 오버레이는 의존 금지 — 본 카드의 재현은 본 트리에서 새로 작성 (M0).

## Intake Record

- 베이스 흡수: 81786284e → db0c514d3, 16커밋, FF 병합 — 인테이크 시 관측 완료. t1547 수리 라운드 일부 포함 → 릴레이 일부가 이미 수리 상태 (재고정 표 11a·5b 행).
- 브랜치 개명: WT-t1561-p2-4 → WT-user-asset-bundle (관측). 트리 디렉터리는 카드 id 유지(t1591).
- 임대 기록 (양립 기록): 인테이크 시 `factory next` 거부 **2회** (직렬 슬롯 — 진행 중 형제 카드 점유) → 이후 임대 **성공** (2026-10-09 ~00:04, factory next가 카드 행 t1591 stage=- worktree=t1591 반환). 장부(record) 정산은 리더 소관.
- 경계 준수: 본 에이전트(manager-spec)는 `.moai/specs/SPEC-USERASSET-DEPLOY-GUARD-001/` 외 수정 없음. 커밋 없음(레인이 스테이지 경계에서 커밋), 큐 변경 없음, 사용자 질의 없음.

## Re-Anchor Table (HEAD db0c514d3 — 관측 요약)

전문(인용 라인 포함)은 research.md §1. 라벨: observed / relayed / repaired / retired / indeterminate.

| # | 원장 좌표 | 판정 |
|---|---|---|
| 1, 10-P1 | lock_guard_windows.go:23 | observed — 마커 무소유권·무회수 (lock.go .lock과 비대칭) |
| 2 | doctor_user_install.go:134/:165 | observed — 적재 오류 OK 위장·공허 일치 |
| 3 | install.go:239 | observed — 첫 스테이징 전 회수 항목 병합 부재 |
| 4 | migrate_project_assets.go:89/:119 | observed — 미등록 미러 사본 영구 보존 |
| 5 | deployer.go:192 / :347 / bundle.go:145 | observed / **repaired**(ListTemplates 제외, 758→414 코멘트) / **retired**(template/bundle.go 부재) / emission compared 축은 doctor_agentemit_embed.go로 재고정 |
| 6a | install.go:297 | observed — 반입 항목 해시 미갱신 |
| 6b | install.go:743 | observed — 0o644 하드코딩 (.sh 실행권 상실) |
| 7a | install.go:392 | observed(무변환 복사 경로) — 내용 주장 relayed |
| 7b | doctor_harness.go:58 | observed — skillsDir 통째 교체 |
| 8a | agentfm.go:128/:282 | observed — 중복 행·PostFormValue 첫 행 |
| 8b | install.go:459 | observed — 클로저 전개 부재 (prune R-f-② 유예 팔은 존재 — 변별 재현 필요) |
| 8c | install.go:211 | observed — ReadFile 직접 판정 (FIFO 블록) |
| 8d | deployer_mode.go:37 | **retired** — M7 퇴역, 생 면 internal/cli/skills.go로 재표현 |
| 9a | paths.go:38 | observed(루트 정의) — 집합 불포함 relayed (config 적재면 pre_tool.go:355 관측) |
| 9b | doctor_harness.go:57 | observed — 빈 디렉터리 존재만으로 선택 |
| 10 | install.go:255 | observed — 플래그 일괄 영속화 |
| 10a | install.go:260 | **indeterminate** — SchemaVersion=1 뿐, v1/v2 좌표 부재 → REQ-JRN-004 재정식화 |
| 11a | update.go:565 | **repaired** — 확인창 뒤 이전 코멘트+호출 관측 → 회귀 가드 |
| 11b | init.go:947/:905 | observed — 재개 경로 부재 |
| 12 | install.go:750 | observed — rename 직전 재검증 없음 |
| 13a/13b | install.go:373/:584 | observed(경로) / **relayed-unverified** (런 재현 확정 대상 — 카드 명시) |

관측 방법 선언: 본 표의 모든 "observed"는 plan-phase(2026-10-09)의 Read/grep 관측이다. 어떤 테스트도 실행하지 않았다 — 결함의 런타임 재현은 전부 런 M0 소관이다.

## Open Questions for Plan-Audit

1. **원장 11a 이미 수리** — RED-first가 재현 불가(GREEN)로 끝나면: REQ-SRF-006 회귀 가드 확정 + 원장 항목 소관 종료 표기로 처리할지 (리더 처분 동반).
2. **원장 5 부분 수리** — ListTemplates 제외 분항 재수리 금지 확인. 생존 분항(SRF-002/003)만 본 SPEC이 소관.
3. **원장 8d 좌표 퇴역** — skills.go 재표현이 원장 의도와 일치하는지 (deployer_mode.go:37의 원래 메커니즘 소멸 확인됨).
4. **원장 10a 재정식화** — v1/v2 오판 주장 → REQ-JRN-004(버전 게이트) 전환이 타당한지, 아니면 원장 좌표 소관 종료인지.
5. **원장 13a/13b 미검증 릴레이** — M0 재현이 반박 시 REQ-CNV 적용 범위 축소 절차 (7a만 잔존) 사전 승인 요청.
6. **template/bundle.go:145 부재** — 원장 좌표 소관 종료 또는 cli/bundle.go 재지정 판정.
7. **원장 8b prune 절반** — remove.go R-f-② 유예 팔과의 관계 변별을 M0에 포함할지.
8. **windows 커버리지** — 원장 1의 검증 구성(표 테스트 + GOOS 빌드 게이트, REQ-LOCK-002)이 plan-audit 기준에 충분한지.
9. **web 폼 계약 변경(8a)** — 폼 키 스코프화의 호환 기간 필요성.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-09
- 아티팩트 셋: Tier L 5종 (spec.md, plan.md, acceptance.md, design.md, research.md) + progress.md — 동시 발행 완료.
- 사전 검증: SPEC ID 정규식 `PASS` (Bash verbatim), frontmatter 12 필수 필드 적합, ID 유일성 확인(`SPEC-USERASSET-*` 0건).
- 독립 감사 종결 — plan-audit iter3+Addendum 5 **PASS-WITH-DEBT 0.94**(기준 0.85 상회·blocking 0·audited_sha 7d2e2a36e·receipts rcpt-cf62fd3b4adf15332e74b11e, rcpt-9941027c067ed9cab2bcf821)·부채 2건(R4→run M6, R5→sync).

## §E.2 Run-phase Evidence

_(pending run-phase — manager-develop 소관. M0 RED 배터리 verbatim 관측이 최초 기록된다.)_

## §E.3 Run-phase Audit-Ready Signal

_(pending run-phase — manager-develop 소관.)_

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase — manager-docs 소관. sync_commit_sha: )_

## §F Phase 4 Mode Selection

- 입력 파라미터: tier L · 범위 약 15-20 파일(구현+신규 재현 테스트) · 도메인 4(internal/userassets, internal/cli, internal/template, internal/web) · 언어 혼합 Go 100% · 병렬 이득 LOW(coding-heavy) · agent-team 전제 N/A(미요청).
- 모드 평가: direct 미해당(다중 파일·의미 변경) / fanout 미선택(coding-heavy — Anthropic coding-task 병렬성 유의) / sweep 미선택(기계 균일 변환 아님·파일 30 미만·상호 의존 마일스톤) / **serial 선택** / agent-team 미요청(실험적 명시 전용).
- **Decision: serial** — 마일스톤 M0→M7 순서에 1회 1개 manager-develop 스폰(트리 내 단일 작성자 — 팩토리 레인 one-writer 규율과 합치).
- 근거: Anthropic coding-task caveat(코딩 과업의 순차 기본) + 결정 가역성 순서(plan §F)가 마일스톤 간 의존을 만드는 구조. 팩토리 레인 크론+태스크 규율이 연속성 담당 — ac_converge goal 미무장(중복 감시자 방지).
- Boundary Case 해당 없음(기준 여유 있음).
