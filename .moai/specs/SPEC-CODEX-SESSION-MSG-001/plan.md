# plan.md — SPEC-CODEX-SESSION-MSG-001

> Tier L 구현 계획. 연구는 plan-phase에서 완료(research.md) — run-phase는 구현부터 시작한다. 마일스톤 순서는 결정 가역성(변경 가능성 높은 결정 먼저): 데이터 모델 → 도구 표면 → 교리/문서 → e2e.

## §A. Context

- 워크트리: `.claude/worktrees/t187` (브랜치 `WT-codex-session-msg` @ 76b2c4ece, origin/main 기준 — release/v3.1.3 배치 밖, 최종 PR은 main 대상 Route B).
- SPEC 산출물: `spec.md`(REQ-CSM-001..015) / `plan.md`(본 문서) / `acceptance.md`(AC-CSM-001..013) / `design.md`(채택 설계) / `research.md`(A2A 실측·3축 비교) / `progress.md`.
- 확장 표면: `internal/sessionmsg/`(신규 패키지), `internal/cli/mcp_session_msg.go`(신규), `internal/cli/mcp_server.go`(등록 4줄 추가), `internal/mcp/catalog.go`(4항목 추가), `internal/config/defaults.go`(임계값 5개), `.claude/rules/moai/workflow/cross-session-messaging.md` + 템플릿 미러(교리 확장), `.claude/rules/moai/core/moai-mcp-tools.md` + 템플릿 미러(카탈로그 21→25).
- 재사용(신규 발명 없음): `add()` 등록 패턴(mcp_server.go:113-135), 카탈로그 가드·콘솔 파생(catalog.go), 자문적 록 패턴(`internal/session/registry_lock_unix.go`), `internal/atomicfile`, C-HRA-008 정적 가드 패턴(mcp_boundary_test.go).

## §B. Known Issues (위임 프롬프트 주입 항목)

- **B1 크로스플랫폼**: 록 파일 분리(`_unix.go`/`_windows.go` 빌드 태그 — `internal/session` 선례 준수). 검증 `GOOS=windows GOARCH=amd64 go build ./...`.
- **B3/B11 서브에이전트 경계(C-HRA-008)**: CLI/MCP 경로에 AskUserQuestion 금지. `internal/cli/mcp_session_msg.go`에 `mcp_boundary_test.go` 패턴의 정적 가드 추가(AC-CSM-009).
- **B6 spec-lint**: 본 SPEC은 `### Out of Scope —` H3 형식 준수(작성 시점 lint 통과 — progress.md §E.1).
- **B8/B10 작업 나무 위생**: `.moai/state/**` 런타임 파일 무변경. 브로커 테스트는 전부 `t.TempDir()` 격리(스토어 생성자가 경로를 받는 `NewStore(path)` 형태).
- **B14 하드코딩 방지**: 임계값 `internal/config/defaults.go` 단일 원천(`var` 선언 — 테스트 대체 가능 형태, `DefaultCodexReviewGateTimeout` 선례). 환경변수 추가 시 `internal/config/envkeys.go` 상수.
- **t184 MCP 스큐**: 새 도구는 서버(세션) 재시작 전에 인식되지 않는다 — e2e 절차 0단계(REQ-CSM-015).
- **테스트 부하 규율**: 전체 스위트 로컬 금지 — `go test ./internal/sessionmsg/... ./internal/cli/ -run <타깃>` 만. e2e는 직렬 1회.

## §C. Pre-flight (run-phase 시작 전)

```bash
git branch --show-current && git rev-parse HEAD      # WT-codex-session-msg 확인
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
grep -rn "session_msg" internal/ --include="*.go" | grep -v _test | wc -l   # 0 (신규 확인)
grep -n "frozen" internal/session/registry.go                              # 동결 경계 재확인
ls internal/mcp/catalog.go internal/cli/mcp_boundary_test.go                # 재사용 표면 존재
```

## §D. Constraints — PRESERVE 목록 (무변경)

1. **codex_task 패밀리**: `internal/cli/{codex_task,mcp_codex,codex_jobs,codex_job_control}.go` — 무변경(경계만 문서화, spec.md §F.1).
2. **glm 패밀리**: `internal/cli/mcp_glm.go` 등 — 무변경.
3. **세션 레지스트리**: `internal/session/registry.go` Entry 스키마 + `.moai/state/active-sessions.json` — 무변경(REQ-COORD-024 동결). `internal/session`의 록 패턴은 복제해서 사용, 원본은 건드리지 않는다.
4. **Claude 네이티브 런타임**: SendMessage/ListAgents 무관여.
5. **기존 21 도구**: 이름·동작 무변경 — 카탈로그는 추가 전용.
6. **교리 기존 조항**: `cross-session-messaging.md` 기존 절은 문자 그대로 유지 — 확장 절만 추가.
7. **`.moai/state/**` 런타임 관리 파일**: usage-log 등 무변경 (신규 루트 `session-msg/` 추가와 무관).
8. **AGENTS.md**: 무수정(SPEC-AGENTS-MD-CANON-001 소관).

금지: `--no-verify`, main 직접 push(repo-local 전 Tier PR 정책), `git add -A`(명시 경로만), 로컬 전체 스위트.

## §E. Self-Verification (run-phase 보고 의무)

acceptance.md §F의 AC 이항 검증 매트릭스 + manager-develop §E 표준(E1-E8)을 5절 형식(Claim/Evidence/Baseline-attribution/Gaps/Residual-risk)으로 보고. 각 항목은 (a) 명령 (b) 축어 출력 (c) HEAD SHA 귀속을 명시. TDD 사이클이므로 E8(RED 실패 출력) 포함.

## §F. Milestones (우선순위 순, 시간 추정 없음)

### M1 — 데이터 모델 + 파일 스토어 (사이클: tdd)

`internal/sessionmsg/` 패키지. envelope.go(Message/Part/Delivery/Envelope + 검증) → agent.go(등록 멱등·하트비트·조회) → store.go(send/poll/ack + 지연 스윕: 클레임 TTL 환원·메시지 TTL 삭제) → lock.go(에이전트별 자문적 록, 플랫폼 분리). defaults.go에 임계값 5개(`DefaultSessionMsgMessageTTL` 24h, `DefaultSessionMsgClaimTTL` 10m, `DefaultSessionMsgAgentOfflineMinutes` 30, `DefaultSessionMsgPollBatch` 16, `DefaultSessionMsgMaxTextBytes` 65536).
커밋: `feat(SPEC-CODEX-SESSION-MSG-001): M1 session message envelope + file store`
AC 연결: AC-CSM-001..006.

### M2 — MCP 도구 표면 4종 (사이클: tdd)

`internal/cli/mcp_session_msg.go`(얇은 핸들러 — 인자 파싱→코어 호출→구조화 결과) + `mcp_server.go` 등록 4줄(`add()` 패턴, `session_msg_list`만 read-only hint) + `internal/mcp/catalog.go` 4항목. 정적 가드: `TestSessionMsg_NoAskUserQuestion`·`TestSessionMsg_NoInlineGetenv`(mcp_boundary_test.go 패턴). 경계 단언: 핸들러/코어에 exec.Command·codex-jobs 토큰 0건(AC-CSM-008).
커밋: `feat(SPEC-CODEX-SESSION-MSG-001): M2 session_msg MCP tool surface (register/list/send/poll)`
AC 연결: AC-CSM-007..010.

### M3 — 교리 전문화 + Template-First (문서·템플릿)

`cross-session-messaging.md`에 "Codex broker path" 확장 절(design.md §8 매핑표) + `moai-mcp-tools.md` 카탈로그 21→25 갱신 + 재시작 전제 문구(REQ-CSM-015). 둘 다 본품과 `internal/template/templates/` 미러 동시 편집 + `make build`. 도구 설명 문자열에 규율 짧은 형태 반영(REQ-CSM-014). 템플릿 중립성(§25) 준수 — SPEC-ID·내부 날짜 미기입.
커밋: `docs(SPEC-CODEX-SESSION-MSG-001): M3 doctrine codification + template mirrors`
AC 연결: AC-CSM-011..012.

### M4 — e2e 실주고받기 검증 (직렬 1회)

acceptance.md §E 절차: 0) 바이너리 rm+cp 재설치 + 세션 재시작(재스폰) 1) CODEX_HOME 격리 + 수동 `codex mcp add` 2) 승인 정책 처리 3-7) 등록→왕복→ack 8) 증거를 progress.md §E.2에 `session-msg-e2e` 마커로 기록 9) 격리 위생 확인(실제 ~/.codex 무변경).
커밋: `test(SPEC-CODEX-SESSION-MSG-001): M4 cross-session e2e evidence`
AC 연결: AC-CSM-013.

### 3-phase close 계획 (sync)

- Route B(전 Tier PR — repo-local 정책): run 완료 후 PR 생성은 lane 오케스트레이터/manager-git 소관.
- sync: 단일 sync 커밋이 `implemented → completed` 전이를 실운반(별도 Mx 커밋 없음 — 3-phase close). `progress.md` §E.4에 `sync_commit_sha` 기입은 pending-backfill 패턴(자기 커밋 SHA는 물리적으로 못 알므로 sync 커밋에 플레이스홀더 → 후속 커밋 백필, D3 면세 조항). **§E.5는 만들지 않는다**(은퇴 — MX 태그는 §E.4에 접힘).
- close 커밋 제목은 전체 SPEC-ID 개별 명시(`chore(SPEC-CODEX-SESSION-MSG-001): ... 3-phase close`) — 결합/축약 scope 금지.
- frontmatter `era: V3R6` 명시로 분류 고정(H-override) — §E 마커 과도기 오분류 방지.

## §G. Anti-Patterns

- **AP-1 codex_task 별칭**: session_msg가 서브프로세스를 스폰하거나 잡 레코드를 쓰는 순간 경계 붕괴(REQ-CSM-010) — "잠깐 codex를 불러 대답하게 하자"는 별도 SPEC이다.
- **AP-2 레지스트리 직접 기록**: codx 엔트리를 `active-sessions.json`에 넣는 설계로의 회귀 — 동결 위반(research.md §5).
- **AP-3 전역 스윕 데몬**: TTL 정리를 위해 상시 프로세스/타이머를 두는 것 — 지연 스윕(호출 시점 정리)으로 충분하다(부하 규율).
- **AP-4 확인 = 클레임 혼동**: poll을 읽음 확인으로 치고 ack를 생략하면 at-least-once가 깨진다 — 클레임은 소유권, ack(삭제)가 확인이다.
- **AP-5 MCP 재시작 생략**: 새 도구 등록 후 세션을 재시작하지 않고 "도구가 안 보인다"로 디버깅하는 낭비(t184 스큐) — e2e 0단계.
- **AP-6 승인 정책 무시 e2e**: 무정책 `codex exec`로 MCP를 부르면 "user cancelled MCP tool call"(P3) — 대화형 세션 또는 정책 구성.
- **AP-7 과잉 Part**: raw/url 부분, Task 상태기계를 미리 구현하는 것 — 참조 형상 이상은 이식 경계를 흐린다(design.md §3.1).

## §H. Cross-References

- spec.md §A(측정 전제)·§F(경계) / design.md(구조) / research.md(근거) / acceptance.md(AC·e2e)
- 감사: `.moai/reports/t187/codex-support-audit.md`
- 인접 SPEC: SPEC-CODEX-PHASE2-001(codex_task 경계), SPEC-V3R6-MULTI-SESSION-COORD-001(레지스트리 동결), SPEC-MCP-CONSOLE-001(카탈로그 가드)
- 규칙: `.claude/rules/moai/workflow/cross-session-messaging.md`(M3 확장), `.claude/rules/moai/core/moai-mcp-tools.md`(M3 갱신), CLAUDE.local.md §2(Template-First)·§4(로컬 검증 규율)·§11(바이너리 재설치)
