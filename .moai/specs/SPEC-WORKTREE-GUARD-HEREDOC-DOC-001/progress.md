# Progress — SPEC-WORKTREE-GUARD-HEREDOC-DOC-001

> 카드 t512 (GH #1659). 워크트리 `.claude/worktrees/t512` · 브랜치
> `WT-guard-heredoc`.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-07
plan_audit_verdict: PASS 0.91 (iter1, Major 0 / Minor 5 — 판정문 .moai/reports/t512/plan-audit-verdict.md, 읽기 전용. Minor 5건 post-audit revision 완료 — 아래 resolution map. 판정 이후 산출물이 변경됐으므로 artifact-hash는 불일치 — run 페이즈 Phase 1 Plan Audit Gate는 skip-eligible 아님, 재집행 또는 재판정이 원칙)
plan_artifacts: spec.md, plan.md, acceptance.md, research.md, progress.md
plan_pin_tree: 6a46c0edb (plan 작성 시점 HEAD, 카드 커밋 0개 — RED-now 셀 전체의 측정 트리)

### iter1 Minor resolution map (post-audit revision, 2026-09-07)

| 항목 | 수정 | 위치 (수정 후) |
|------|------|---------------|
| F2 (load-bearing) | AC-WGHD-002 `brace expansion` 앵커 문턱 ≥1 → **≥2** 승격 + 셀 안에 뮤턴트 근거 명시(quoted 문장 2회/unquoted 1회=3 기준; 1이면 quoted 탈락 뮤턴트가 기계 통과) + 소개 문단의 뮤턴트 서술을 양방향으로 정정 | acceptance.md AC-WGHD-002 (intro + GREEN 경로), plan.md T1 Verification 주석 (≥2 근거 병기 — 두 문서 정렬) |
| F1 | AC-WGHD-005 RED 셀의 grep 출력 순서를 실측 순서(live 먼저, mirror 나중 — 원장 프로브 방향과 동일)로 정정. 재측정으로 verbatim+rc=1 재확보 | acceptance.md AC-WGHD-005 RED-now 셀 |
| F3 | ls RED 셀의 스트림 귀속 정정 — 두 "No such file" 행은 **stderr**이고 stdout은 공백(별도 관측 `2>/dev/null` → 공백+rc=1로 확인). doctrine §2.1의 공백-stdout 완전 관측 형태로 기재 | acceptance.md AC-WGHD-006 RED-now 셀 |
| F4 | 매달린 "§2.0" 참조 제거 — F2의 셀 내 뮤턴트 근거 재작성에서 흡수 (존재하지 않는 절 지칭 소멸) | acceptance.md AC-WGHD-002 |
| F5 | plan.md T1 초안의 규범 문장("is the correct treatment for the whole body") → 관측 서술로 정밀화("command substitutions in the same position are already folded"). 처방은 M3 업스트림 초안의 수정 방향 1 논거로 이동. loosening-danger 경계(unquoted 총알의 "correct behavior, not a defect")는 불변 유지 | plan.md §3 T1 삽입 초안 브리지 문단 |

F2/F1/F3 재측정(파이프 없는 rc 캡처, `6a46c0edb` 재핀): `brace expansion`=0/rc=1,
`unquoted`=0/rc=1, `1659` grep live-first 출력/rc=1, ls stdout 공백+rc=1 —
문턱 승격(≥2) 후에도 RED-now 상태 불변(0 < 2).

notes: >
  5 artifacts authored (Tier M 세트 + research.md). RED-now 셀은
  acceptance.md § 문서 핀 규칙과 research.md §7 배치 표에 따라 6a46c0edb 에
  핀됐다. iter1 판정 후 Minor 5건 수정으로 acceptance.md/plan.md이 변경됐다 —
  RED-now 셀의 측정값 자체는 재측정으로 동일함을 확인했다(트리 무변경).

## §E.2 Run-phase Evidence

측정 좌표: 브랜치 `WT-guard-heredoc`, base `6a46c0edb`, run 커밋 M1 `504d6457c` · M2 `b482942fc` · M3 `76c296623`.

> **집행 주체 특기**: run 페이즈는 manager-develop 위임으로 시작했으나 GLM 백엔드
> 레이트리밋(HTTP 429)으로 4회 중도 종료됐다(마지막 2회는 각각 5시간 한도·즉시
> 재발). 위임 에이전트는 사전 점검·M1 삽입·GREEN 앵커 1차 측정까지 완료한 상태로
> 멈췄고, 남은 구간(미러 정합·뮤턴트 관찰·M2/M3·커밋·§E)은 운영자의 「계속 진행」
> 지시 아래 오케스트레이터가 직접 집행했다. 위임 에이전트의 기여(삽입 본문,
> 앵커 1차 측정)는 그대로 유지됐고 본 섹션의 수치는 전부 오케스트레이터 재측정이다.

### M1 — 독트린 삽입 + 미러 패리티 + 뮤턴트 관찰

| 항목 | 측정 | 결과 |
|------|------|------|
| 미러 정합 (위임 에이전트의 드리프트 정정 후) | `cmp` live vs mirror | **exit 0** — 드리프트는 505행 1단어(behaviour→behavior, T1 초안 원문 기준 live가 정답)뿐이었다 |
| 미러 drift 테스트 | `go test ./internal/template/ -run TestRuleTemplateMirrorDrift -count=1` | **ok** (0.470s) |
| 뮤턴트 관찰 [AC-WGHD-001] | 미러에 알려진 차이 재삽입 → `cmp` | **differ: char 44091, line 505, rc=1** |
| 뮤턴트 관찰 (계속) | drift 테스트 (뮤턴트 상태) | **FAIL** (0.505s) — 기계가 비대칭을 잡는다 |
| 복원 (재편집, git restore 미사용) | 재편집 후 `cmp` + drift 테스트 | **exit 0** + **ok** (0.258s) |

### GREEN 앵커 (사전 RED `6a46c0edb` 핀 → 삽입 후, 파이프 없는 rc)

| 앵커 | RED-now (사전, 핀) | GREEN (M1 후) | 판정 |
|------|-------------------|----------------|------|
| `grep -c "brace expansion"` | 0 / rc=1 | **3** / rc=0 (≥2 충족 — quoted 2 + unquoted 1) | PASS |
| `grep -c "unquoted"` | 0 / rc=1 | **2** / rc=0 (≥1 충족) | PASS |
| `grep -c "re-measured"` | — (≥1) | **1** / rc=0 | PASS |
| 중립성 `grep -cE "1659\|SPEC-WORKTREE-GUARD-HEREDOC"` (템플릿 사본) | — (0 요구) | **0** / rc=1 | PASS |

### M2/M3 초안

- `reporter-reply-1659.md` (M2, 한국어): 4요소 충족 — ①분석 검증+매트릭스 인용(트리 핀 병기) ②소속=바이너리(리포 독트린 :481 인용, branch_guard 아님 :485) ③우회=Write 도구·단순 명령 분할(약화 조언 없음) ④업스트림 초안 안내·게시는 원격 착지 후(초안 명시). 어조: 제보자 진단의 검증으로 시작, 우회는 도움, 업스트림은 발견의 대변(리드 [HARD] 반영).
- `upstream-draft-claude-code.md` (M3, 영문): 5요소 충족 — 매트릭스 verbatim / provable-inertness 논증(제보자 credit) / 수정 방향 1(quoted 본문 중괄호 접기) / 수정 방향 2(거부 문면이 구문 지목) / 제보자 credit. 범위 제한 준수(스크립트 파일 우회 요청 없음, unquoted 동작 변경 요청 없음). **게시는 리드 몫 — 초안 전용.**

### 범위 증명

`git diff --stat 6a46c0edb..HEAD` — 변경 파일: 독트린 2본(live+미러), SPEC 5종, 리포트 4종(원장+판정서+초안 2). **Go 소스 변경 0건**, `branch_guard.go`/훅 무접촉. E2/E3(go build·커버리지)은 문서 전용 SPEC이라 대상 없음 — 정직 표기. `go test ./internal/template/ -run TestRuleTemplateMirrorDrift` 만 유일한 Go 게이트(위 표 ok).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-07
run_commit_sha: 76c296623 (M3 — run phase spans 504d6457c / b482942fc / 76c296623)
evidence_path: .moai/reports/t512/ (재현 원장 · plan-audit 판정서 · 초안 2종)
anchor_green: brace expansion=3 · unquoted=2 · re-measured=1 · 중립성=0 (전부 rc 동반 재측정, HEAD 76c296623)
new_defects: 0 (Go 변경 없음 — 미러 패리티·drift 테스트가 유일 게이트, 양쪽 ok)

## §E.4 Sync-phase Audit-Ready Signal

(sync 페이즈에서 manager-docs 작성 — sync_commit_sha 기입)

## §F Phase 4 Mode Selection

- Input parameters: tier=M · scope(files)=4(독트린 2본 + 초안 2종) · domains=2(rules 문서, 리포트 초안) · language mix=Markdown 100% · concurrency benefit=LOW(문서 저작, 직렬 자연) · agent-team prereqs=미요청
- Mode evaluation: direct=미선정(산출물 4종 + 미러 규율 수반) · serial=**선정** · fanout=미선정(연구 병렬 이익 없음) · sweep=미선정(기계적 대량 변형 아님)
- Decision: serial
- Justification: 단일 문서 저작 스폰(독트린 편집+미러 cmp+초안 2종)으로 코딩도 연구도 아닌 직렬 자연형이다. 병렬화 이익이 전혀 없어 fanout/sweep은 부적합하고, 산출물 간 의존(독트린 삽입 → 앵커 GREEN → 초안 인용)이 있어 serial이 유일한 합리적 축이다. plan-audit iter2 PASS 0.98(현재 아티팩트 해시) 이후 첫 run 스폰이다.
