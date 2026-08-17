# t92 — Agent Teams 허용으로 독트린 전환 (측정+문서 카드)

> 카드 원문(backlog t92): "[3.1.1][치명·최우선] Agent Teams 플래그 결정 — 실측: .claude/settings.json:360과 템플릿 settings.json.tmpl:386 둘 다 CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=\"1\"(리드 grep 확인) — CLAUDE.md §15 RETIRED 선언과 정면 모순, 템플릿이 전 배포 사용자에게 켠 상태로 배포. … 1단계: 이 저장소에서 teammate 전환 재현 — 결과에 따라 독트린 또는 설정 수정. 템플릿 동반 필수."
>
> 운영자 결정(2026-08-17, 방향 반전): "문서와 지침을 고쳐서 agent teams 사용을 허용" — 플래그 제거가 아니라 RETIRED 선언 철회.

## 판정 요약

| 항목 | 결과 |
|---|---|
| 1단계: teammate 전환 재현 | **재현됨** — named 스폰이 `in_process_teammate`로 전환, Agent 결과 채널로 무반환(~1h·2회 확인 무응답, TaskStop으로 종료) |
| 반증 관측(리드) | 같은 버전에서 named 워커 5개(A~E) 정상 완료·결과 반환(2026-08-17) — **불일치, 미해결** |
| 독트린 전환 | 완료 — 로컬 16파일 + 템플릿 미러 16파일 + catalog.yaml, "사용 허용(실험적·명시 요청 전용)" 상태로 재서술, 폐기 이력은 계보 보존 |
| 검증 | `make build` 후 `go test ./internal/template/` 전량 ok(leak STRICT·sentinel·parity 포함, 3회 반복) |

## 전환 원칙 (적용된 서술 규칙)

1. **실험적 재허용 + 명시 요청 전용**: `--team`/`--mode team`/`Team` 스케일 라벨이 Agent Teams 레이어 선택. Phase 4 판정 트리의 자동 선택은 유지 안 함(Tier L 자동 라우팅은 `manager-kanban` 그대로).
2. **센티널은 역사 문서화로 생존**: `MODE_TEAM_UNAVAILABLE` 문자열이 run.md에 존재해야 하는 CI 감사(`agentless_audit_test.go` REQ-WF003-011)는 센티널을 "retired-era 폴백 마커"로 문서화해 만족 — 테스트 변경 0.
3. **양면 증거의 정직한 기록**: 리드의 5-워커 정상 완복 관측과 본 세션의 무반환 관측을 나란히 기록(§C.1·CLAUDE.md §15) — "결과 반환 신뢰성 미증명, 세션마다 검증" 권고.
4. **중립성**: 배포 문서(템플릿 포함)에서 내부 날짜·CC 버전 제거 — 근거(날짜·버전·관측 세부)는 본 리포트가 단독 보존. cp가 덮어쓴 4개 템플릿의 기존 중립화(본문 날짜 줄)는 leak 테스트가 잡아 복원.
5. **제약 조건부 병기**: 리드 전달 외부 감사 제약 7건(중첩 금지·세션당 1팀·in-process teammate의 백그라운드 스폰 불가·/resume 미복원·권한 spawn 시점 고정·/model 미상속·팀 상태 수동 편집 금지 + skills/mcpServers 미적용 + GLM 상속 미실측)을 §C.1에 "변환 재시 시 적용" 조건으로 문서화.

## 변경 파일 (16 로컬 + 16 템플릿 + catalog.yaml)

로컬: `CLAUDE.md`, `.claude/rules/moai/workflow/{orchestration-mode-selection,spec-workflow,dynamic-workflows}.md`, `.claude/rules/moai/core/settings-management.md`, `.claude/rules/moai/development/{agent-authoring,orchestrator-templates}.md`, `.claude/skills/moai-foundation-quality/modules/integration-patterns.md`, `.claude/skills/moai/SKILL.md`, `.claude/skills/moai/references/reference.md`, `.claude/skills/moai/workflows/{moai,run}.md`, `.claude/skills/moai/workflows/run/{mode-orchestration,context-loading,phase-execution}.md`, `.claude/agents/moai/manager-kanban.md`
템플릿: 위 전부의 `internal/template/templates/` 미러 + `internal/template/catalog.yaml`(해시 재생성)

핵심 서술 지점: CLAUDE.md §4(카탈로그 문장)·§15(재작성), orchestration-mode-selection §A 3행·§B 트리·§C.1(전면 재작성: 선택 규칙·계보·양면 증거·제약 목록)·§C.2·안티패턴·crosswalk 2행·비회귀 노트, spec-workflow §Mode Dispatch·Methodology delegation·§Agent Teams Variant(재작성), run.md 모드 표·센티널 문단, SKILL.md 플래그 5곳, moai.md 모드 선택+실행 요약 5곳, mode-orchestration 전면, context-loading 모드 값·리조버 유사코드·센티널 키·harness 표, settings-management·agent-authoring·orchestrator-templates·dynamic-workflows·integration-patterns·manager-kanban 각 폐기 서술.

## 실측 (E1 — named 스폰 경로, 본 세션 2026-08-17)

- 스폰: `Agent(subagent_type: general-purpose, name: "t92-probe-named")` + 진단 프롬프트. 스폰 응답: "will receive instructions via mailbox"(무명 스폰의 "async agent"와 다른 경로).
- 관측: 약 1시간 동안 결과 채널 반환이 없고, 상태 확인 메시지 2회에도 무응답.
- 종료: `TaskStop` 반환 메시지가 유형을 공개 — `task_type: "in_process_teammate"` (raw/e1-taskstop.txt).
- **해석**: named→teammate 변환은 이 버전(CC 2.1.233)에서 실재하며 재현됨(카드 1단계 완료). 다만 리드 세션의 5-워커 정상 완료와 같은 버전에서 상반됨 — 결과 반환 신뢰성은 세션 조건(미확인 변수)에 따라 갈리는 것으로 기록. 원인 규명은 후속 카드.

## Gaps (미검증)

- GLM 상속(치명): teammate가 lead의 `ANTHROPIC_BASE_URL`/`ANTHROPIC_AUTH_TOKEN`을 상속하는지 — teammate 경로 결과 반환 불가로 측정 못 함. §C.1에 "측정 전 신뢰 금지"로 명시.
- E1 불일치 원인(어느 조건이 반환 가능/불능을 가르는지) 미규명.
- t78 감사 제약 7건 중 실험적 검증(중첩·/resume 등)은 "변환 재현 시" 조건이라 미실행 — 문서화만.
- pre-retirement 동작(팀 자동 선택 등)의 부활 없음 — 확인 사항이지 결함 아님.

## 잔여 위험

- 배포 문서가 "허용"을 말하지만 결과 반환 신뢰성은 미증명(양면 관측) — 사용자가 `--team`을 쓸 경우 세션마다 반환 동작 확인 권고가 §C.1에 실려 있으나 기계 게이트는 아님.
- CC 버전업 시 플래그·변환 동작 변화 가능 — "re-verify on Claude Code upgrades" 문구 유지.

## 재현

```bash
# 실측 (E1): named 스폰 → 무반환 관측 → TaskStop으로 유형 공개
#   Agent(subagent_type: general-purpose, name: "t92-probe-named", prompt: <진단>)
#   SendMessage(to: "t92-probe-named", ...) ×2 → 무응답
#   TaskStop(task_id: "t92-probe-named") → task_type: in_process_teammate
# 검증: make build && go test ./internal/template/ -count=1  → ok
#       MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'Leak|Sentinel|Parity' → ok
# 잔여 스윕: grep -rn "Agent Teams.*RETIRED" <.claude + templates> → 통제된 계보 서술만
```
