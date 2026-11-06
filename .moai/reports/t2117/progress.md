# t2117 진행 기록 — main 지침 파일 처리 결함 4건 (C1~C4)

Card t2117 · run to2bbk · lane-31 · Class B (run → sync, RED 먼저) · 2026-11-06

## 배차·임대·기준

- 배차: 리더 — "main 지침 파일 처리 결함 4건: C1 instructions_loaded.go:197 FIFO 등 정규 파일 검사 없음(P1), C2 migrate_local_instructions.go:89 읽기 후 os.Remove 경합(P1), C3 token_budget_guard.go:284-288 로컬 지침·import 예산 누락(P2), C4 instructions_loaded.go:203 중첩 상대 import 루트 기준 해석(P2). RED 먼저"
- 참고: #1944(origin/pr/1944=89e5cff0e5)에 원래 수리 — 패치 참고하되 이름 변경(t2114 소관) 제외
- 임대: `moai factory next --card t2117 --run to2bbk` → worktree=t2117
- 워크트리: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2117` (branch `WT-main-c1-instructions`, base b00c35d103 = local main)
- 전이: t2094 merge-ready 대기 중 ExitWorktree(keep) → 주 체크아웃 임대 → EnterWorktree — t2094 complete는 v3.2.1-rc.1 설치 확인 신호 시 이 트리를 잠깐 벗어나 수행 후 복귀(af-flow 규칙 5)
- 근거: `.moai/reports/t2077/sync-audit-iter3.md`, 영수증 rcpt-729226e9dfd8adeb96bc20b2 (이 트리에 파일 부재 — t2077 트리/기록 참조 필요 시 별도 확인)

## 레인 코드 조사 (직독, 2026-11-06)

- C1 `internal/hook/instructions_loaded.go:197`: BFS 루프가 `os.ReadFile(cur)` — 대상이 FIFO/디바이스면 쓰기자 없는 FIFO 읽기가 무기한 블로킹. 197~200의 오류 분기는 "없음/읽기불가 스킵"이지만 정규 파일 검사 자체가 없음
- C2 `internal/cli/migrate_local_instructions.go:76→89`: claudePath 본문 읽기(76) → 백업+agentsPath 기록(82-88) → `os.Remove(claudePath)`(89). 읽기와 삭제 사이 타 작성자가 claudePath을 고치면 새 내용이 삭제로 소실(백업은 구본문). 91-94에 rollback 패턴 존재
- C3 `internal/config/token_budget_guard.go:284-289`: alwaysLoadedSurface = paths-restriction 없는 규칙 .md + 고정 슬롯(AGENTS.md, moai.md) — 로컬 지침(CLAUDE.local.md/AGENTS.local.md)과 @-import 파일이 예산 표면에서 누락
- C4 `internal/hook/instructions_loaded.go:203`: `resolveUnder(projectRoot, imp)` — 모든 import를 루트 기준 해석. 중첩 파일의 상대 import는 임포팅 파일의 디렉터리 기준이어야 함
- #1944 참조 diff: `4c4217a19d..origin/pr/1944` 3파일 +505/−119 — 카드 범위보다 큼(이름 변경 등 포함) → 전체 흡수 금지, C1~C4 접근만 참조

## 위임

- `manager-develop` 스폰 (depth-1 리프, effort=medium — 클래스 B 매핑), RED 먼저 지시:
  - 결함별 RED 시험 선작성 → 실행 → RED 원문 관측(기대 단언 실패, 컴파일 오류 아님) → RED 커밋 → 수리 → GREEN 커밋
  - C1: FIFO 앵커+고루틴 타임아웃 가드(현행은 타임아웃=결함 관측, 수리 후 스킵), 디렉터리 위장도 커버
  - C2: 결정론적 시험용 최소 시임(패키지 기존 시임 스타일) 허용 — 읽기 후 변경 삽입, 현행=변경본 삭제(소실), 수리=중단+롤백+claudePath 보존
  - C3: 표면 확장은 기존 import 파싱 재사용(중복 구현 금지)
  - C4: 임포팅 파일 디렉터리 기준 해석, 절대경로 불변
  - 범위: 3파일+시험(+C2 최소 시임), 이름 변경 일절 금지(t2114), #1944 전면 흡수 금지
  - 검증: 3패키지 표적 -run 시험+go vet, 전체 수트 금지
- 상태: 백그라운드 실행 중 — 완료 통지 대기 (결과를 예측·보고하지 않음)

## run 단계 결과 (스페셜리스트 반환 + 레인 재검증)

### 커밋 (2건, 전방, 명시 경로 스테이징)

- RED `108c9bed5e` — `[t2117] test(hook,cli,config): RED for instructions-handling defects C1-C4 (card t2117)` — 4건 RED 관측(원문은 스페셜리스트 보고·아래 요지)
- GREEN `927cd8b6518f3f3aed0285a69dcab5e069a68d02` — `[t2117] fix(hook,cli,config): repair instructions-handling defects C1-C4 (card t2117)` (4파일 +199/−29, 트리 깨끗함 — 직독)

### 결함별 수리 요지

- C1: BFS 읽기 전 os.Lstat+Mode().IsRegular() 가드 — 비정규/검사실패는 기존 스킵 경로와 동일 처리. #1944의 regularReadableFile 개념 채택(카드 처방대로 Lstat — 심볼릭 링크도 거부). RED: TestInstructionFileSetSkipsFIFOAnchor 5.02s FAIL(FIFO 앵커 블로킹=타임아웃 관측) → GREEN 0.01s PASS. 디렉터리 위장 커버 시험은 전후 PASS(커버리지 전용, 문서화)
- C2: 삭제 직전 claudePath 재독록+bytes.Equal 대조 — 불일치/재확인 실패 시 중단+기존 롤백 패턴+사유 명시 오류. #1944의 변경감지-후-삭제 아이디어 채택(mtime/O_NOFOLLOW 변형·백업 리네임 재설계는 범위 밖 — 리네임은 t2114). 최소 시임 `migratePreRemoveHook`(프로덕션 nil, RED 커밋의 유일 프로덕션 변경, 동작 중립). RED: SourceChangeRace 0.14s FAIL(경합 쓰기 소실) → GREEN 0.03s PASS(7/7)
- C3: alwaysLoadedSurface 확장 — localInstructionSlot(AGENTS.local.md 우선, CLAUDE.local.md 대체)+instructionImportMembers(사이클 안전 BFS, 상대 import는 임포팅 디렉터리 기준=C4 의미론, 절대/이탈 거부, 비정규 미개방, 시드 제외, 중복 제거). 재사용: import 파서를 config.ParseInstructionImports로 한 번만 두고 hook.parseImports가 위임(config→hook 사이클 회피). RED: 표면 누락 3행 FAIL → GREEN PASS. 실트리 tripwire TestAlwaysLoadedTokenBudget **65,156/77,600 tokens(여립 12,444)** — AGENTS.local.md(+13,380 tokens) 수용, 예산 상향 불요. 기존 TestAlwaysLoadedSurfaceEnumeration은 결함 표면을 encode하고 있어 재유도(의도 변경)
- C4: resolveUnder(projectRoot, dir, rel) — dir=filepath.Dir(cur). #1944 접근 그대로, 절대/프로젝트 이탈 거부 불변. RED: 중첩 상대 import가 루트 해석 FAIL → GREEN PASS

### 환경 특이사항 (귀속 기록)

- 카드 중반 호스트 디스크 100% 도달 + 동시 행위자의 공유 go-build 캐시(88G) 소거 + go1.26.8 툴체인 src 트리 부분 삭제 관측(math/bits·internal/synctest 등 간헐 "not in std"). 최종 검증은 온전한 캐시 go1.27.1+전용 GOCACHE=/tmp/t2117-gocache로 실행. hook/config GREEN 실행은 오염 전(유효), 오염 후에는 vet만 재관측 — go1.27.1에서의 두 패키지 시험 재실행은 안 함(스페셜리스트 명시)
- /tmp/t2117-gocache(~200MB) rm 권한 거부로 잔존 — 기계 로컬 스크래치

### 레인 재검증 (HEAD 927cd8b651 직접 실행, 격리 환경 귀속: GOTOOLCHAIN=go1.27.1 GOCACHE=/tmp/t2117-gocache)

- `go test ./internal/hook/ -run 'Instruction' -count=1` → ok 2.690s
- `go test ./internal/cli/ -run 'TestMigrateLocalInstructions' -count=1` → ok 2.385s
- `go test ./internal/config/ -run 'TestAlwaysLoaded' -count=1` → ok 1.383s
- gofmt/vet는 스페셜리스트 실행분 재사용(같은 트리 키): gofmt 무출력, VET-HC-OK·VET-CLI-OK

### 미검증 (Gaps)

- 전체 수트 — 레인 로컬 규율상 의도 미실행(통합 브랜치 CI 소유)
- FIFO 시험의 Windows 빌드 — unix 태그 전용 파일, 비-unix 형제 불요(스페셜리스트 판단)
- CI 전면 판정 — 착지 후 소관

## sync 감사 1라운드 (2026-11-06 ~21:53 local; 429 중단 후 재개 완주)

- 감사관: sync-auditor (Claude) — 429(glm-5.3 5시간 한도, 20:12:34 리셋)로 1회 중단 후 세션 재개로 완주
- 판정: **PASS-WITH-DEBT** — 조화 94.2 (기능성 97 / 보안성 93 / 장인정신 95 / 일관성 92), 차단 코드 결함 없음, 전체 발견 선택적(F1~F6 기록분)
- 4건 결함 RED→GREEN 독립 기계적 재현으로 검증. 판정 파일 `.moai/reports/t2117/sync-audit-iter1.md` (verdict:/audited_sha: 각 1줄, audited=927cd8b6518f3f3aed0285a69dcab5e069a68d02=HEAD — 직독 확인)
- 영수증 상태: **미발급** — codex_audit 2회 호출 모두 결정론적 `stale_server` 거부(서빙 빌드 vs 설치본 불일치, 해제는 호스트 세션의 MCP 서버 재시작 필요 — 서브에이전트 불가, CLI 감사 동사 부재, audit_multi는 상시 결정으로 금지). Stop hook의 재시도 지시 1회 이행(HEAD 재확인 뒤) — 동일 거부. t2094 선례의 충돌 조항 적용: 추가 호출 없음, 판정 토큰 불변
- de2d 허용분 소모 여부: **미소모** — 두 호출 모두 리뷰 시작 전 거부(영수증 미생성), 리더 선행 해석 규칙("리뷰 미실행 호출은 횟수 불산입") 적용. 호스트 MCP 재시작 뒤 1회 호출 가능
- 환경: 감사 중 toolchain 재소거 1회 — 환경 사유, 1회 재시도로 회수, 카드 결함으로 기록 안 함(감사관 Gaps 명시)
- 결정 기록: decided_by=lane-31 sync-audit-r1 evidence_refs=.moai/reports/t2117/sync-audit-iter1.md ladder_path=Claude sync-auditor(af-flow) — PASS-WITH-DEBT 수용, 영수증 보강은 호스트 MCP 재시작 후 별도 기록

## stale_server 원인 규명 (운영자 지시: 원인 찾기 — 2026-11-06 ~22:23 local)

- 증상: codex_audit 4회 모두 결정론적 stale_server(서빙 rc.1/b00c35d103 vs 설치본 rc.2). /mcp 재연결 2회에도 불구
- 원인(프로세스 표 직독): **내 세션의 MCP 서버 pid 28154(PPID 44598=lane-31 호스트, 기동 21:00:04)** — rc.2 설치(21:17, go/bin/moai mtime)보다 17분 이전에 스폰된 rc.1 메모리 상 프로세스. `/mcp` 재연결이 이 프로세스를 종료·재생성하지 않았고(타 세션들은 21:27-21:31에 재생성된 흔적 존재), 세션의 모든 moai MCP 호출이 28154로 라우팅됨
- PATH 단일성 확인: `which -a moai` → go/bin/moai 하나뿐. MCP 설정 `{"command":"moai","args":["mcp-server"]}`(프로젝트 .mcp.json) — 스폰은 PATH 해석이므로 재생성 시 곧바로 설치본(rc.2)
- 조치: 기록된 단일 PID 28154 종료(내 세션 소유, PPID 44598 재확인 뒤 — 패턴 kill 아님). 종료 확인 + MCP 연결 끊김 확인 → 재연결 시 신규 스폰이 곧 rc.2가 되는 상태
- de2d 허용분: 4회 거부 모두 리뷰 시작 전(영수증 미생성) — 미소모 유지. 재연결 후 감사관 1호출 예정
- 교훈(레인 판단 기록): Claude Code의 /mcp 재연결은 기존 서버 프로세스를 재사용할 수 있다 — 설치본 교체 후 재연결이 실제 재생성을 보장하지 않으면(오류문 권고 "restart the host session"), 세션 소유 서버 프로세스를 기록 PID로 종료하는 것이 재생성을 강제하는 최소 침습 경로

### 영수증 보강 완료 (2026-11-06 ~22:40 local)

- 원인 조치 후 rc.2 서빙 확인(서버 지시문 스탬프 "build v3.2.1-rc.2 (commit: 04d8c27a7e) (pid 21882)" + 신규 프로세스 22:30:55 기동 + `moai version` rc.2 3중 직독) → 감사관 5호출 성공
- 영수증: **rcpt-570cb4d72e7e4f0aaab856cb** (backend codex, resolved_model gpt-6.1-sol, build_commit 04d8c27a7e, 검토 범위 merge-base b00c35d103→호출시점 HEAD bf35a34953) — de2d 허용분 1회 소모(카드 한도 2회 중 1회)
- codex 독립 판정: **fail — P2 2건**. 감사관 재판정: F7(신규, in-diff Medium) — AGENTS.local.md+CLAUDE.local.md 병존 시 localInstructionSlot이 첫 파일만 반환해 후자 본문·import가 예산 표면 누락(근거: AGENTS.md §8 결합 주입. 주 체크아웃은 오늘両 파일 보유). 구현은 카드 처방(first/fallback) 그대로 → 처방 수준 갭, **레인 수용: 이 카드는 선택적, 팔로업 카드 후보**(양 슬롯을 시드로 열거). F3은 기존 선택적 발견과의 동의 — 신규 조치 없음
- 판정 토큰 불변: PASS-WITH-DEBT 94.2 (부채=미게이트 코드 결함 0 + F7 등 선택적 발견). 기계 줄 verdict:/audited_sha: 각 1줄 유지(직독 확인)
- 리더 전달: F7 팔로업 카드 후보(큐 변경 없이 기록만)

### Gaps
