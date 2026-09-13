# AS4 최종 좁은 보정 — native fork 완료 의무

2차 감사 FAIL 0.875의 B2에 따라 기존 native Agent fork 양성 의무를 REQ-MG-015와 AS-013 및 계획에 명시 유지했다.
설치 가용성 조사 → 가능한 다른 AS4 구현·검증 진행 → native fork를 포함한 최종 완료 판정 순서다.
설치본에서 기능을 찾지 못하면 NOT-RUN 전제 실패이며 기능 삭제 또는 성공이 아니다. 부모 이력 상속·분기 위치·
병렬/중첩 격리·resume 양성과 foreign/변조 귀속 음성 증거 전 AS4 및 전체 지원 완료를 보류한다.
일반 non-fork와 --fork-session 성공은 이를 대체하지 않는다. 새로운 protocol이나 파라미터는 도입하지 않았다.

검증 명령과 출력:

```text
go run ./cmd/moai spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
```

기준: legacy WT / WT-unified-gateway / HEAD 81c1d58f9cf7045594ee61d5e4ff380948ce9eba.
현재 파일 SHA-256:

```json
{
  "spec.md": "ccdbdd091cd2ee26292725abc6b1220c25f9a6385046bf27efd196188b48e086",
  "design.md": "db1c4a8c455ba061e7d8047d1021ed59b52c439e5afa9ec23ac414acc55eab03",
  "plan.md": "58b509852be7bdcaf1c0ee34d07c11e8f866d9c49d19fc2ca314e1344a45e0a3",
  "acceptance.md": "82231badae807bbbb18f1a65935ff0f6ccfb8f0f18bcae4d4b12482572b3ac1c"
}
```

이 작업에서 native fork를 실행하지 않았다. 구현·운영 Gap은 유지한다. 네 SPEC 문서는 이 hash로 동결하고,
커밋·push·HTML 변경을 수행하지 않았다.

---

# AS4 계약 보강 기록 — 재설계 2차

## Claim

AS4 감사 1차 FAIL(0.8125)의 B1/B2/MP6/MP7에 대응하여 기존 네 문서를 좁게 수정했다. 버전 0.11.0 및 번호·상태를 유지했다.
아래 1차 기록은 대체된 설계 이력이며 현재 계약이 아니다.

- B1: PreCompact 예약으로 다음 HTTP를 구분하는 설계를 제거했다. 정상 요약 turn의 실제 응답과 인증된 PostCompact의
  exact summary digest/scope/epoch를 결속하고 공개 이력을 재설정한다. 추가 명시 compact RPC는 0회다.
- B2: 일반 non-fork Agent의 native child context와 --fork-session을 구분했다. 일반 자식은 독립 thread,
  명시 session fork는 launcher 원본 family와 exact prefix 원장의 completedTurnID를 lastTurnId로 사용한다.
  native Agent(fork)/subtask는 공식적으로 별도 이력 상속 기능이므로 이 결과로 지원을 주장하지 않는다.
- MP6: 실제 launch_exec_posix.go/launch_exec_windows.go를 읽고 //go:build !windows 및 //go:build windows를
  현재 의무에 명시했다. Windows 런타임 실행 성공 주장은 하지 않았다.
- MP7: 공식 모델 표·GLM 모델 문서·Claude model-config·token-counting 문서를 직접 열어 명목 한도와 UI 가정,
  실제 계정 수용 및 계량을 구분하고 조사 marker를 근거로 해결했다. 대형 입력 실증은 운영 Gap이다.

## Evidence

```text
go run ./cmd/moai spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
```

기존 compact-history-shape-result/comparison/verdict를 읽었다. comparison의 seed_user_canonical_equal,
seed_system_text_equal, compact_assistant_text_exact, summary_hash_matches_PostCompact는 모두 true이며
http_prompt_id_present는 false다. 이 작업에서 probe를 재실행하지 않았다.

직접 열어 확인한 공식 근거:

- https://code.claude.com/docs/en/sub-agents#what-loads-at-startup — 일반 자식의 독립 context 및 fork 예외
- https://platform.claude.com/docs/en/models/overview — Opus 5/Sonnet 5 1M, Haiku 4.5 200K
- https://docs.z.ai/guides/llm/glm-5.1 및 https://docs.z.ai/guides/llm/glm-4.7 — 200K context, 128K output, text
- https://code.claude.com/docs/en/model-config#correct-the-window-for-a-gateway-or-custom-model-id — ID 인식별 env 적용 차이
- https://platform.claude.com/docs/en/build-with-claude/token-counting — 추정과 실제 usage의 차이

설치 schema ThreadForkParams.lastTurnId의 직접 판독:
`Optional last turn id to fork through, inclusive.` 및 `The referenced turn cannot be in progress.`

## Baseline-attribution

동일 legacy WT, WT-unified-gateway, HEAD 81c1d58f9cf7045594ee61d5e4ff380948ce9eba.
현재 네 파일 SHA-256:

```json
{
  "spec.md": "4069787de0dbc4f69d9cccfd21964a25589084215c0f65a18750f26784d6fbba",
  "design.md": "3ab6665a7467dd79c1b5aad9c7e991ed0bc4b1a59f7cfd4f9bddf09e74981047",
  "plan.md": "6d46a7cbc320c947bd3a3dd3238880a59adcd2a83d2195c42dd60e1ff1d06395",
  "acceptance.md": "c1ca7267fef3acfd7fa0edad2c0fc9e61bd09c52b57431c2b8854bdfb175c5d5"
}
```

## Gaps

새 제품 rebase·회상·fork를 실행하지 않았다. 각 실제 압축 모드, nested 일반 자식, 명시 세션 분기,
native Agent(fork)/subtask, 계정별 대형 입력·Windows 실행은 제품 실증을 별도로 요구한다.

## Residual-risk

반환 요약 digest 일치는 인증이 아니다. 세션 전용 IPC 권한과 정확한 epoch/완료 응답 원장 검증이 필요하다.
후속 wrapper의 substring만으로 입력을 생략하면 공격자가 일반 입력을 누락시킬 수 있으므로 구조 파싱과
exact summary 대조를 요구한다. 인식할 수 없는 구조는 명시 거절하지만 양성 검증 없이 완료 판정하지 않는다.

---

## 대체된 1차 기록

# AS4 계약 보강 기록

## Claim

기존 SPEC 0.11.0의 REQ-MG-015/017, AS-010~013과 출력 정책 판정을 좁게 보강했다. 새 REQ/AC 번호 및 상태 전이는 없다.
AS3 ManagedSessionAuthority 및 승인된 초기 native/후발 dispatcher 계약을 유지한다. 구현 완료 보고가 아니다.

- 압축은 인증된 예약, 결속된 HTTP, 단일 compact 실행, 실제 완료, 비사실적 안내와 불투명한 이어쓰기 기준점 순서다.
- 압축 후 새 입력만 반영하며 이전 사실·도구 결과 회상과 새 프로세스 resume가 완료 조건이다.
- 부모 결속을 모르는 자식 ID로 fork를 추정하지 않는다. 독립 자식 성공으로 부모 이력 상속을 통과 처리하지 않는다.
- 작성 중 운영자가 “API도 App Server 출력 정책 사용”을 승인했다는 오케스트레이터 전달을 반영했다.
  구독/API 모두 서버 출력 정책을 쓰며 바이트·취소 제한은 동일 생성 토큰 상한이 아니다. API 별도 과금과 선택 인증 방식,
  자동 과금 전환 금지를 유지한다.

## Evidence

이 작업에서 실행한 명령과 출력:

```text
moai session current
01a08e7b-6aa0-7361-ab7e-ea8da1f02228

git fetch origin main
From https://github.com/modu-ai/moai-adk
 * branch                main       -> FETCH_HEAD

git rev-list --count --left-right origin/main...HEAD
0 2879

go run ./cmd/moai spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
```

이 작업에서는 기존 probe 결과를 읽고 계약 근거를 확인했으며 probe를 재실행하지 않았다.
`probe/compact-hook-result.json`은 Claude 2.1.269, success=true, seed/compact exit_code=0,
PreCompact → compact HTTP → SessionStart(compact) → PostCompact 순서를 기록한다.
`probe/agent-identity-parallel-summary.json`의 자식 요청 2/5와 3/4는 각각 같은 agent-id이며 두 자식은 서로 다르다.
함께 읽은 verdict는 parent_tool_use_id와 HTTP agent-id의 직접 결속을 확보하지 못했다고 명시한다.

설치 schema `probe/schema/v2/ThreadCompactStartParams.json`은 required=[threadId],
ThreadCompactStartResponse는 properties=[]다. ThreadResumeParams.history 설명은
`[UNSTABLE] FOR CODEX CLOUD - DO NOT USE.`로 시작한다.
공식 소스 `/tmp/openai-codex-audit.zbvNXB/codex-rs/app-server-protocol/src/protocol/v2/item.rs:414`의
ContextCompaction은 id만 가진다. thread.rs:1119/1126은 compact params 및 빈 response다.
사실 요약을 반환하는 공식 필드로 해석하지 않았다.

## Baseline-attribution

트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
브랜치: `WT-unified-gateway`
HEAD: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`
기존 다수 dirty/untracked 항목을 보존했다. 커밋·push는 수행하지 않았다.

| 파일 | 변경 전 SHA-256 | 변경 후 SHA-256 |
|---|---|---|
| spec.md | 6d54a5e9415c588f320283883a5b766f0cd35dea685bde78e7da4f7de3c71265 | dc5b33a83e080b12173fc8a9c12a045c7c3a4a71bcae1873752c6480bc31401a |
| design.md | cdb78efa528caa63dd4005bedb308a185035b8e258393df1561a509609c585ab | 19596742159e80eff605230e887635cd9a9d0e94e24cb3ed3948dc3cf05d1689 |
| plan.md | 7521f77145f35d8e482c4d53224c184536d216258fb2d4f8c2b6c591a8d2783e | 07c03c8fef3d5b07cfe54f39f38e9832a0e9370cd34c5cb55c6f8847574ce97b |
| acceptance.md | 35386461fb3a7b271384c8f4e6976b5adb296996fdc4d070cf2094a6bb6b2e8b | 7670b8ab80b22ed97cbcd65282a3ffb62ce02919136fe7d12fa0d2f200d0fd08 |

## Gaps

- 인증된 hook 예약과 압축 HTTP의 유일한 결속은 제품 구현·실행으로 아직 입증하지 않았다.
- 수동 대화형 UI, 자동·자식 압축, 완료 응답 유실 복구, 압축 후 실제 모델의 사실 회상은 이 작업에서 실행하지 않았다.
- HTTP만으로 정확한 부모 도구 호출 및 중첩 부모 귀속을 확정할 근거가 없다. 해당 실제 fork 경로는 완료 Gap이다.
- API 키 생성 호출은 수행하지 않았다. 정책 승인은 모델의 실측 출력 한도 보장이 아니다.

## Residual-risk

예약과 일반 요청이 같은 범위에서 경쟁하면 단순히 “다음 HTTP”를 잡는 구현은 잘못된 입력을 압축할 수 있다.
따라서 계약은 유일한 결속을 요구하고 불명확하면 실행 전 거절한다. Claude의 압축 후 history에 기준점이 실제로
보존되는지와 모델 이력 회상은 제품 E2E로 검증해야 한다. 설치 protocol 변경 시 완료 사건 대응도 재확인해야 한다.
