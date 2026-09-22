# t236 판정서 — [#1640] MOAI_PROJECT_DIR 워크트리 전환 후 스테일 (잔여 결함)

브랜치 `WT-project-dir-stale` @ `7291be34a` (워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t236`)
기준선 `65196a5a7` (= origin/develop, pre-spawn `0 0` 실측) · Class B · 측정일 2026-09-02
측정 세션 source_session_id=018df390-f2ab-46f9-9a20-0ecb2c44a0ae

## Claim

이슈 #1640의 잔여 2축(트리 무브 시 재스탬프·레지스트리 재배치 부재 / verify 도구 project_root
부재 + 폴백 무표시)을 재현→수리→재측정으로 닫았다. 이슈 미검증 2항 중 (ii) 라이브 파라미터
부착은 해소 확인, (i) PostToolUse 실발화는 착지 후 신규 세션 관측으로 남는다(문서 근거 확보).

## Evidence

### 배차 전제 재검증(리드 지시 — "잔여를 다시 긋는다")

| 항목 | 결과 | 근거 |
|---|---|---|
| 파라미터 축 착지 | 확인 — spec_progress/audit/drift/codex/glm/audit_multi/graph 3종 = 10종에 project_root 등록 | `mcp_project_root.go` 판독 + moai-mcp-tools.md 사전 카운트 "Ten" |
| 잔여 A(무브 핸들러) | 미착지 — post_tool_worktree.go 부재, matcher 커버 없음, t74 relocate가 CwdChanged에만 의존 | `ls internal/hook/` + 로컬 settings.json PostToolUse 판독 |
| 잔여 B(verify 2종) | 미착지 — 등록·핸들러 모두 resolveProjectDir() 직결 | mcp_server.go :192-206, :689-720 판독 |
| t332 판정 흡수 | "checkable half holds / runtime half unverified / needs-operator-decision" — 본 카드가 그 결정의 실행 | `.moai/reports/t332/cards/batch-2.md` t236 절 |

### 라이브 재현 (본 세션에서 직접 관측 — 상세는 reproduction-evidence.md)

- **L1**: 본 세션이 EnterWorktree를 수행한 뒤에도 세션 레지스트리 엔트리 `cwd`가 launch-time
  primary 값으로 고정(`018df390…` 엔트리 직독). CwdChanged 배선 무결 통제(래퍼 존재 +
  hook-missing.log 부재) → "이벤트 미발화"가 지배 설명.
- **L2**: probe SPEC(워크트리에만 존재)을 러닝 서버에 조회 — no-param `spec_audit`:
  `total_specs: 0`, 오류·경고·트리 표시 전무(조용한 오독 종단간 재현). param 부착 호출:
  `total_specs: 1` → **이슈 미검증 (ii) 라이브 해소**. 응답에 `_root` 부재 → 가시화 갭 확인.
- **L6(서술 정정)**: MOAI_PROJECT_DIR의 Go 소비자는 0 — 이슈 원문의 "mcp-server 폴백이 이
  변수를 읽는다"는 현재 코드 불일치. 폴백은 CLAUDE_PROJECT_DIR/server cwd. 결함 실체는
  spawn-frozen 폴백이며 스탬프는 시설 정합성용. systemMessage 문구에 반영.

### 수리 (manager-develop 위임, RED-first — 상세는 run-evidence.md)

- 커밋 `f1b379434` feat(hook): stampProjectDirEnv 추출 + post_tool 분기 + handleWorktreeMove
  (레지스트리 재배치·env 재스탬프·systemMessage) + 테스트 4종 + settings.json.tmpl matcher 확장.
- 커밋 `21734f9e9` feat(mcp): resolveToolProjectRootWithSource(프리비넌스 4단) + verify 2종
  project_root 등록·해석 + spec_progress/drift/verify 2종 `_root` {source, dir[, warning]} +
  신규 테스트 4종 + 문서 패리티 카운트 갱신(Ten→Twelve, 로컬·템플릿 바이트 동일).
- 커밋 `7291be34a` docs(changelog): [Unreleased]### Fixed 엔트리(추적되지 않는 경로 인용 회피 —
  t381 교훈 준수).
- RED: 수리 전 트리에서 신규 테스트가 명시된 사유로 실패(파라미터 무시/분기 부재) — run-evidence.md §RED.
- GREEN(레인 독립 재측정, 커밋 후 본 세션 재실행): `go test ./internal/hook/...` ok,
  `./internal/cli/...` ok, `./internal/template/...` ok (전부 env 스크럽 복합호출, -count=1),
  `go vet ./internal/hook/... ./internal/cli/...` exit 0, `make build` exit 0.

## Baseline-attribution

모든 측정은 워크트리 HEAD `65196a5a7`→`7291be34a` 구간, 본 세션에서 실행. L2의 서버 관측은
러닝 빌드 `64bba61aa`(2026-08-31) — project_root 착지 이후 빌드로 파라미터 경로 유효.

## Gaps

1. **PostToolUse 실발화(이슈 미검증 (i))** — 훅 설정은 세션 시작 스냅샷이라 본 세션에서 신규
   matcher 실험 불가. CC 문서 근거는 확재(PostToolUse는 "도구 성공 완료 직후" 실행, matcher는
   도구명 정규식, `cwd`는 호출 시점 값, `systemMessage`는 표준 출력 필드 — docs.claude.com
   hooks reference, 2026-09-02 판독). **닫는 방법: 병합 후 신규 세션에서 EnterWorktree 수행 →
   systemMessage 관측** ("re-stamped: true/false"가 CLAUDE_ENV_FILE 전달 여부까지 함께 판독).
2. **러닝 mcp-server 라이브 재확인** — 러닝 서버가 구빌드. 바이너리 재빌드 + 서버 재접속 후
   L2 재실행이 종료 관측.
3. CwdChanged 미발화의 CC 런타임 내부 사유 — 증상 수준까지만 관측됨(설계에 영향 없음).

## Residual-risk

- 신규 세션에서도 matcher가 발화하지 않으면 M1의 시스템 메시지 경로는 no-op — Gap 1 관측이 닫을
  항목이며, 발화 실패 시의 대안(CwdChanged 측 폴링 등)은 별도 카드.
- `spec_audit`은 `_root`를 실어 보내지 않음(구조체 응답 유지 — 참조 패치 선례, 커밋 메시지 문서화).
  폴백 사용 시 표시는 없으나 파라미터는 존중됨.

## 처분 메모 (리드/폐기 담당용)

- `.moai/specs/SPEC-T236-PROBE-001/` — 라이브 재현용 미커밋 스크래치. **원격 착지 확인 후
  워크트리 폐기 시 함께 소멸** (별도 삭제 불요 — 트리와 운명 공유).
- `.moai/reports/t236/{reproduction-evidence,run-evidence,verdict}.md` — 커밋된 증거 3종(본 파일 포함).
  리포 관례상 증거는 브랜치에 적재(리드 실측: origin/develop `.moai/reports/` 1,348파일 중 카드
  디렉터리 1,225) — 라이브 재현 기록이 워크트리와 함께 사라지면 판정 재구성 불가(t336 전례).
- 훅 matcher 변경은 다음 세션부터 유효 — 병합 후 운영자가 `moai update`로 settings.json을
  재배포받아야 이 저장소 도그푸드에도 적용됨.
- 병합은 리드 창 경로로 — 레인은 push하지 않음(규율 준수). 병합 커밋에서 squash여도 커밋 3건
  모두 메시지에 `t236` 명시되어 추적 가능.
