# t386 — 감사 산출물 보관 규약 명문화 (verdict)

- **브랜치**: WT-audit-evidence-store (기준: 로컬 develop b7462203a 흡수, fast-forward f7cabfc29→b7462203a)
- **카드**: t386 lane-10 (G2b 배차, 순차 1장째)
- **클래스**: 규약 문서 + 다중 지침 갱신 (문서 축)

## Claim

감사 판정서(plan-audit / sync-audit)의 보관 규약을 명문화해, 판정 근거가 세션 종료와 함께 사라지는 결함을 기전으로 닫는다. 산출물 4파일(규약 문서 2표 + kanban-dispatch 쌍둥리 2표), 중립성 위반 0건, 쌍둥리 byte-identical 2쌍.

## Evidence

### 실증 3건 (리드 지시에 따라 근거로 채용)

| # | 실증 | 귀속 |
|---|------|------|
| 1 | t336 verdict.md가 primary 체크아웃에만 존재(8,682B), develop 트리에는 부재 — 저장소 이력 미기입 | lane-10 직접 관측: `ls /Users/goos/MoAI/moai-adk-go/.moai/reports/t336/` vs `ls .moai/reports/t336/`(워크트리, b7462203a 기준) — verdict.md가 양측에서 갈림 |
| 2 | codex SPEC 2건(SIDECAR-GUARD·WIRING)의 sync-audit 판정서가 develop에 부재 — spec/plan/acceptance/progress 4파일만 존재 | lane-10 직접 관측: `ls .moai/specs/SPEC-CODEX-SIDECAR-GUARD-001/` 및 `ls .moai/specs/SPEC-CODEX-WIRING-001/`. 점수(SIDECAR-GUARD PASS 94.5·WIRING PASS 94.7)는 리드 배차 보고 인용 — lane-10은 재측정 안 함 |
| 3 | t304 증거가 통합 창 진입 직전까지 미커밋 — 규율을 아는 레인도 놓친 실증 | 리드 배차 보고 서술. 정합 관측: b7462203a 흡수 diff에 `.moai/reports/t304/` 5파일 신규 포함(착지는 나중에 커밋돼서 가능) |

→ 3번째 실증이 규약의 서사를 결정: **근본 원인은 규율 부재가 아니라 기전 부재** — 규약 문서 Why 절 26-28행에 반영.

### 산출물 (4파일)

| 파일 | 상태 |
|------|------|
| `internal/template/templates/.moai/docs/audit-artifact-convention.md` | 신규 119행(why 서사 보강 후, `wc -l` 실측), 중립성 0매치 |
| `.moai/docs/audit-artifact-convention.md` | 미러, byte-identical (`diff -q` OK) |
| `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | § Completion is read, never trusted에 판정서 파일 읽기 문단 추가 |
| `.claude/rules/moai/workflow/kanban-dispatch.md` | 미러, byte-identical (`diff -q` OK) |

### 검증 명령 (이 트리, 이번 실행)

- 중립성 재스캔: `grep -cE 'SPEC-[A-Z]+-[A-Z0-9-]+|t[0-9]{3}|[0-9a-f]{40}|20[0-9]{2}-[0-9]{2}-[0-9]{2}'` → 규약 문서 0, kanban-dispatch 0
- 미러 패리티: `diff -q` ×2쌍 → 모두 무출력(동일)
- 빌드: `make build` → catalog.yaml 갱신(12,899B) + `go build` 성공(ExitCode 0, v3.1.2-1396-gb7462203a)
- 관리 뿌리 판독: `internal/cli/update/deploy/deploy.go` `ManagedCleanTargets`(56행) — 대상은 .claude/settings.json·commands/moai·agents/moai·skills/moai*·rules/moai·output-styles/moai·hooks/moai (+.moai/config 주석 92행). **`.moai/docs`는 clean 대상 아님** — update 시 템플릿 배포로 덮어씀(byte-identical 미러라 결과 동일, SSOT은 템플릿)

## Baseline-attribution

- 측정 트리: WT-audit-evidence-store (develop b7462203a 흡수 후, 산출물 작성 전 단계 기준)
- 관리 뿌리 판독 대상 커밋: b7462203a 트리의 deploy.go

## Gaps (미검증)

- **plan-auditor.md·sync-auditor.md 산출물 반출 조항 — 미착지**. 리드 지시(2026-09-02 배차 교신): lane-9의 t367(plan-auditor 루브릭 개정)·t302(sync-audit 판정 소유권)가 그 두 파일의 의미를 바꾸므로, **두 카드가 닫힌 뒤** 리드 재개 신호에 따라 착지. 규약 문서는 이 조항을 참조만 하고(§ What makes the convention stick — Auditor side), 조항 자체는 미구현 상태.
- 회귀 테스트(빌드 외): 규약 문서는 정적 문서라 테스트 대상 없음. kanban-dispatch 문단도 텍스트 조항.

## Residual-risk

- 감사자가 조항 없이도 산출물을 내는가에 대한 재발 방지는 현재 2계층(규약 문서 + 리드 읽기 절차)뿐 — 3계층(에이전트 HARD 완료조건) 착지까지는 "안 낸 감사"가 다시 발생할 수 있음. 리드 지시에 따른 의도적 보류.
- `.moai/reports/plan-audit/` 금지 경로 서술은 현재 .gitignore 상태에 의존 — gitignore 정책이 바뀌면 규약 문서 § Where의 서술도 갱신 필요.

## 다음

- 리드 재개 신호(t367·t302 종결) → plan-auditor.md·sync-auditor.md 반출 조항 착지 → 이 판정서 갱신
- 이 카드의 진행 완료 판정은 리드가 본 판정서를 읽고 수행
