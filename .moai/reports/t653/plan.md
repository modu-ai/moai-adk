# t653 — AS-4 소유 thread resume · model 변경 · fork · compaction — 카드 실행 계획

- 카드: t653 (GATEWAY-APPSERVER-20260912 / AS-4)
- 관할 SPEC: SPEC-MOAI-GATEWAY-001 v0.11.0, plan.md §0.11.0 (lines 29-42), acceptance.md §0.11.0 AS-010~AS-013 (lines 582-616)
- 작성 기준선: worktree `.claude/worktrees/t653`, branch `WT-gateway-as4-resume`, HEAD `74d872aaf` (= origin/develop 계열). 선행 t652 = `530bd7330`.
- 이 문서는 plan 단계 산출물이다. 코드 변경은 run 단계 소관이며, 여기서 측정한 모든 line 참조는 작성 시점 값이다.

## 1. Pre-flight inventory — 현재 코드 표면

### 1.1 turn ledger / public prefix / tool continuation (t652 turn-bridge)

| 소유자 | 위치 | 현재 상태 |
|---|---|---|
| 대화 상태 머신 | `internal/codexbridge/engine.go:68-80` (`conversation` struct: `phase`, `turn`, `pending`, `prefix`, `model`) | phase ∈ new/starting/active/waiting/responding/idle/failed/canceled. turn ID는 메모리 원장(`c.turn`) |
| turn 경계 스텝 | `engine.go:263` `Engine.Step` | thread/start → turn/start → item/tool/call 대기 → rpc.Respond → turn/completed. HTTP tool 경계를 넘어 같은 thread/turn을 이음 |
| 단일 이벤트 디먹스 | `engine.go:135` `Engine.read` | threads map(`engine.go:88`)으로 threadId 라우팅. 전체 큐 실패 시 transport fail-closed |
| 지연 turn 정리 | `engine.go:592` `lateStarted` | start RPC 타임아웃 뒤 도착하는 turn/started 정리 |
| 인터럽트/취소 | `engine.go:493` `interrupt`, `engine.go:509` `Cancel` | conversation 단위 canceled phase + turn/interrupt |
| 복구 배리어 영속화 | `internal/codexbridge/store.go:65` `FileStore.save`, 레코드 `store.go:27-30` | **주석(store.go:19-22)이 "Records deliberately cannot be resumed by a new Engine in AS3"로 못박음** — 프로세스 재시작 resume은 현재 구조적으로 금지 |
| App Server 어댑터 | `internal/gateway/appserver.go:33` `Send`, 응답 조립 `appserver.go:69` `appServerResponse` | 출력 정책 주석(appserver.go:40-42): 구독/API 모두 서버 출력 정책, MoAI 바이트·취소 독립 — t653 계약과 일치, 변경 불필요 |
| 권한 | `internal/gateway/appserver_authority.go:40` `Authorize`, `:58` `Check` | account kind(chatgpt/apiKey) + generation scope. generation 불일치 = ErrManagedAuthority |
| prefix 대조 입력 | `engine.go:49-53` (`Request.ExpectedPrefix`, `PrefixDigest`), `engine.go:275` 대조 | prefix 불일치는 ErrScope로 거절. prefix 원본 계산은 게이트웨이 측 receipt projection이 소유 |
| 완료 prefix 원장 | `internal/gateway/receipt/core.go:79` `Publish`, `:99` `Check`, `:127` `Fork` | hash-only candidate(완료 마크 `Candidate.Complete`, `core.go:75`). lastTurnId 개념의 MoAI 측 근사치는 `Candidate.Previous` 체인 |
| family 인덱스 | `internal/gateway/conversation/family.go:64` `Manager` — `New :109`, `Resume :152`, `Continue :184`, `Select :217`, `Complete :232`, `Fork :257`, lease `:312` | family/UUID/project/CWD 묶음. Fork는 이미 `--resume <parent> --fork-session --session-id <id>` 조립(family.go:297) |
| native 전사 검증 | `internal/gateway/conversation/native.go:92` `transcriptModel` | sessionId/cwd 불일치·synthetic 행 거절. 마지막 end_turn model만 회수 |

### 1.2 어댑터 배선 상태 (전제 관측)

`NewAppServerAdapter`(`appserver.go:27`)와 `codexbridge.New`(`engine.go:94`)의 생산 호출 지점이 없다 — `internal/`, `cmd/` 전역 grep에서 _test 파일 밖에 정의와 테스트뿐이다. HTTP turn/pending RPC의 생산 배선은 t654(제품 검증) 또는 별도 카드의 표면으로 읽힌다. t653은 이 배선을 만들지 않는다(카드 범위 밖) — 다만 AS-010의 "새 MoAI process로 시작" 실증이 생산 배선을 전제하는지는 M2 착수 시 재확인이 필요하다.

### 1.3 gap 목록 — t653이 추가해야 하는 것

1. **Resume 없음**: Engine은 `thread/start`만 호출(`engine.go:334`). `thread/resume` 경로 부재. FileStore 레코드가 재개 불가로 설계돼 있어(store.go:19-22) 프로세스 재시작 뒤 소유 thread 재개(AS-010)는 배리어 스키마 확장 + 신규 RPC 경로가 필요하다. 계정·family·agent·thread 귀속 고정은 `codextools.Binding`(registry.go:30: ConversationID/ThreadID/AccountScope)이 이미 운반한다.
2. **idle model 변경 없음**: `engine.go:275`가 `c.model != q.Model`을 무조건 ErrScope로 거절하고, model은 생성 시 고정(`engine.go:228`) + turn/start마다 재전송(`engine.go:406`)이다. idle phase 한정 모델 전환 경로와 active/waiting 전환 차단이 전부 신규다.
3. **compaction 처리 없음**: gateway/codexbridge 전체에서 PostCompact/PreCompact/compact 처리 0건. "정상 요약 turn → 반환 summary와 인증된 PostCompact 대조 → public history 재설정" 전체가 신규다.
4. **경계 fork 없음**: `conversation.Manager.Fork`(family.go:257)는 부모의 **현재** 완료 시점에서 분기한다. AS-013의 "exact prefix의 completedTurnID로 공식 lastTurnId 분기"는 부모가 이후 turn을 더 완료했을 때 경계 이후 사실이 자식에 유입되지 않아야 하므로, 경계 매개변수화(prefix/candidate 지정) fork가 신규다. receipt.Manifest.Fork(core.go:127)는 candidate 전수 복사라 경계 제한이 없다.
5. **native Agent(fork)/subtask 가용성 probe 없음**: 어떤 조사·기록 산출물도 없다. M1 전체가 신규다.

## 2. Milestone

순서는 plan.md §0.11.0 지침을 따른다 — native fork 가용성 확인과 독립적으로 진행 가능한 resume·모델·압축·명시 분기를 먼저 구현하고, 마지막 판정에서 native fork 증거를 포함해 AS-013 전체를 확인한다. 각 milestone 종료 시 5구획 보고(Claim/Evidence/Baseline-attribution/Gaps/Residual-risk)를 `.moai/reports/t653/` 아래 남긴다.

### M1 — native Agent(fork)/subtask 가용성 probe (실증, 우선순위 High)

- **내용**: 설치 Claude의 실제 호출 표면으로 native Agent(fork)/subtask 가용성을 확인하고 결과를 기록한다. 찾지 못하면 **전제 실패 NOT-RUN**으로 기록 — 기존 native fork 양성 의무(AS-013 3번째 Given)는 삭제하지 않고, AS4/전체 지원 완료는 이 증거 전까지 보류다. 일반 자식·`--fork-session` 성공으로 대체하지 않는다.
- **파일**: 코드 변경 없음 (probe는 실제 Claude 세션 + 기록 산출물).
- **증거**: `.moai/reports/t653/m1-native-fork-probe.md` — probe 방법, 호출 표면, 관측 결과, NOT-RUN 여부 판정.
- **RED 조건**: 자동화 RED 없음(실증 항목). NOT-RUN인 경우에도 "probe를 수행했고 못 찾았다"는 기록 자체가 산출물이다.

### M2 — 소유 thread resume (우선순위 High)

- **내용**: Engine에 `thread/resume` 경로 추가. FileStore 배리어 레코드 확장(재개 가능 필드 — store.go:19-22의 AS3 한정 주석을 AS4 계약으로 갱신). 계정(scope)·family·thread 귀속 검증: 다른 계정·다른 대화·변조 mapping 거절. 합성 사실·도구 결과 유지 회상.
- **파일**: `internal/codexbridge/engine.go` (get/Step resume 분기), `internal/codexbridge/store.go` (레코드 스키마), 테스트 `engine_test.go`/`store_test.go`.
- **RED**: resume 경로 부재 상태에서 `thread/resume` 호출 시나리오 테스트가 실패하는 것을 먼저 관측. 거절군(foreign account/변조 prefix/미완료 배리어)도 각 RED.
- **증거**: `.moai/reports/t653/m2-resume-*.log` (RED/GREEN), 5구획 보고.

### M3 — idle model 변경 (우선순위 High)

- **내용**: `engine.go:275`의 무조건 model 동일성 검사를 phase별로 분해 — idle에서만 허용 모델로 전환, active/waiting/responding에서는 거절(active tool 대기 중 전환이 잘못된 turn에 적용되지 않음). turn/start params가 전환된 model을 운반. bare gpt-6 등 미허용 ID와 다른 제공자는 게이트웨이 측에서 거절.
- **파일**: `internal/codexbridge/engine.go`, `internal/gateway` 모델 선택 대조 테스트.
- **RED**: idle 전환 테스트 RED + waiting 중 전환 거절 테스트 RED.
- **증거**: `.moai/reports/t653/m3-model-*.log`. 실제 turn model 일치(AS-011 양성)는 실증.

### M4 — compaction = 정상 요약 turn (우선순위 High)

- **내용**: Claude 압축을 별도 RPC가 아니라 정상 App Server turn의 요약으로 처리. 반환 summary와 인증된 PostCompact의 exact summary digest·scope/epoch 대조 → 통과 시 public history 재설정. PreCompact-following-HTTP 추측 금지, MoAI 측 추가 thread/compact/start 호출 0회. 중복·지연·유실 PostCompact·재시작 주입 시 한 번만 rebase, 모호하면 명시 오류. foreign/변조/오래된 epoch·substring-only 거절.
- **파일**: `internal/codexbridge/engine.go` (이벤트 분류), `internal/gateway/receipt/` (history 재설정과 digest 원장), 테스트.
- **RED**: PostCompact 대조기 부재 상태의 RED + 거절군 각 RED. "MoAI compact 호출 0회"는 카운터 단정 테스트로.
- **증거**: `.moai/reports/t653/m4-compact-*.log`. 실제 Claude 압축 응답·후행 hook 수집(AS-012 전제)은 실증.

### M5 — 명시 세션 fork = lastTurnId 분기 (우선순위 Medium)

- **내용**: 원본 family + exact prefix의 completedTurnID로 공식 lastTurnId 분기. 새 family/thread에 분기 경계 이전 사실만 보존 — 부모가 이후 turn을 완료해도 경계 이후 사실 유입 차단. 새 프로세스 resume 유지. 다른 계정·미완료 경계·변조 prefix/원장·알 수 없는 원본은 외부 분기 전 거절. 일반 자식(독립 child context, AS-013 1번째 Given)은 별도 연결 실증 대상으로 유지하고 부모 귀속 추정으로 지원하지 않는다.
- **파일**: `internal/gateway/conversation/family.go` (Fork 경계 매개변수화), `internal/gateway/receipt/core.go` (경계 한정 candidate 복사), `internal/codexbridge/engine.go` (fork thread 부착), 테스트.
- **RED**: 경계 초과 유입 방지 테스트 RED + 거절군 RED + fork-뒤-새-프로세스-resume RED.
- **증거**: `.moai/reports/t653/m5-fork-*.log`. `--fork-session` 실제 분기 양성과 병렬/중첩 자식 격리는 실증.

### M6 — 증거 취합 (우선순위 Medium)

- **내용**: 변경 패키지 테스트(`internal/codexbridge`, `internal/gateway/...` 영향 범위), 커버리지, `GOOS=windows go build`, lint NEW-vs-baseline 판정, 실증 목록과 NOT-RUN gap 정리, verdict 재료 취합.
- **증거**: `.moai/reports/t653/` 하의 로그 묶음 + `m6-evidence-summary.md`.

## 3. Acceptance mapping (acceptance.md lines 582-616)

| AC | milestone | 자동화 검증 가능 | 실증(실제 세션) 필요 |
|---|---|---|---|
| **AS-010** (AC-MG-009) resume | M2 | resume 경로 자동화 테스트, foreign 계정/변조 mapping/미완료 배리어 거절군, resume.history·raw reasoning 직접 재구축 금지 단정 | 소유 thread ID로 새 MoAI process resume 후 정상 답변·합성 사실 회상 — 실증. 실패 시 NOT-RUN gap + 증거 |
| **AS-011** (AC-MG-003) idle model | M3 | waiting 중 전환 거절, 다른 제공자·bare gpt-6 거절, 경고·실패 명시 | 실제 turn model이 GPT-6 Astra/5.6 선택 ID와 일치 — 실증. family 호환 성공의 계정별 별도 증거 — 실증 |
| **AS-012** (AC-MG-009) compaction | M4 | PostCompact 대조기, 중복/지연/유실/재시작 주입 1회 rebase, foreign/변조/오래된 epoch·substring-only 거절, MoAI compact 호출 0회 카운터 | 실제 Claude 압축(print·대화형·자동·자식) 요청·응답·후행 hook 수집과 digest 일치, 압축 뒤 ToolSearch 후발 도구 — 실증 |
| **AS-013** (AC-MG-009) fork | M1+M5 | 명시 fork 경계 테스트 전부(경계 초과 차단·거절군·fork-뒤-resume), native fork **가용성 기록** 자체 | non-fork Agent 둘+중첩 자식 자별 context / `--fork-session` 실분기 / native Agent(fork) 부모 이력 분기·병렬·중첩·resume 양성 — 실증. native fork 미가용 확정 시 그 Given만 전제 실패 NOT-RUN, **AS4/전체 지원 완료는 보류** (일반 자식·--fork-session 성공으로 대체 불가) |

## 4. Risks and known constraints

1. **ANTHROPIC_* 리터럴 가드**: `TestNoBareAnthropicEnvVarLiteralsInProduction`은 빌드가 아니라 **변경 패키지 테스트에서 실패**한다(`plan.md` §D lines 95-97). 새 env 키는 `internal/config/envkeys.go`를 거친다.
2. **Windows 교차 컴파일**: `GOOS=windows GOARCH=amd64 go build ./...` 필수. conversation 패키지는 이미 `platform_posix.go`/`platform_windows.go` 분리 precedent가 있다.
3. **lint baseline**: gateway 표면 파일에 t671 측정 기준 31건의 기존 golangci-lint 이슈가 있다. 관련 없는 기존 이슈를 고치지 않고, 변경 line에서 새 이슈를 추가하지 않는다.
4. **전체 스위트 금지**: 변경 패키지만 로컬 실행, 전 패키지 판정은 CI(`CLAUDE.local.md` §4.1).
5. **AS3 계약 변경**: M2가 FileStore "재개 불가" 주석(store.go:19-22)을 뒤집는 것은 t652가 세운 의도적 한계의 해제다 — run 단계에서 레코드 스키마 호환성(옛 레코드 거절 또는 마이그레이션)을 명시적으로 판정해야 한다. 묵시적 스키마 확장 금지.
6. **배선 전제**: `NewAppServerAdapter`/`codexbridge.New`의 생산 호출 지점 부재(§1.2). AS-010 실증 설계 시 생산 경로 재확인 필요 — 배선 자체는 이 카드 범위 밖.
7. **출력 정책 고정**: appserver.go:40-42의 구독/API 공통 서버 출력 정책 주석은 t653 계약과 이미 일치 — 재작성 금지, API 별도 과금 자동 전환 금지 유지.
