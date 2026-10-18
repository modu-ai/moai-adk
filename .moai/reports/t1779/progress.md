# t1779 진행 기록 (lane-18, Class B run→sync)

카드: t1779 [운영자 지시 10-15·브레인 영실이 요청·Codex 레인 기동 분석] [Class B·P2] moai codex -l 의 TTY·샌드박스 거부를 작업 전 한 줄로
run: tn1p4e · lane: lane-18 · 트리: `.moai/worktrees/t1779` · 브랜치: `WT-10-15-codex-3`

## 1. 범위(카드 정본 요지)

- `moai codex -l` 이 TERM=dumb·비TTY·샌드박스("Operation not permitted": PATH alias, ps 금지) 환경에서 코드 실행 전 한 줄 거부("터미널에서 직접 실행") + 런치 pending 행 롤백 확인.
- Codex 세션 안에서의 재귀 `moai codex -l` 기동 거부.
- 수용: TERM=dumb 픽스처에서 등록 행 0·거부 한 줄 · 재귀 거부 시험.
- 관측 원인: 로그 끝 moai codex -l 두 번 실패(TERM=dumb → Operation not permitted), 그 전 45초·19초 소요 — 사전 감지 부재로 spawn 후 세션 내 실패.

## 2. 코드 사실 (이 트리 실측)

- 롤백 경로 존재: `defaultCodexSpawnLaunch`(codex_launcher.go:294+) → tmuxSpawnFn → pane identity → `registerFactoryLaunchPending`(:307) → stamp → 실패 시 cleanup+`rollbackFactoryLaunchPending`(:317). **사전 감지 게이트 부재**가 결함.
- 재귀 마커: `config.EnvFactoryBackend == factory.BackendGPT`(Codex 세션이 운반 — factory_card.go:120 factoryLaneRefusal 의 기확립 비교 패턴).
- 삽입점: lane-join 경로에서 `defaultCodexSpawnLaunch` 호출 이전(외부 지점) — tmux 창·pending 행 등 어떤 부수효과도 선행하지 않는 최외곽.

## 3. 결정 기록

- D-t1779-1: 수용 범위를 카드가 핀한 두 픽스처로 한정 — (a) TERM=dumb → 한 줄 거부+등록 행 0, (b) EnvFactoryBackend=gpt 재귀 거부. 샌드박스 ps-프로브 감지는 수용 미핀으로 보류(구현 시 저렴하면 병행 가능, 시험 의무 아님).
- 위임: manager-develop(RED 먼저 — TestCodexLaneLaunchRefusesDumbTerm·TestCodexLaneLaunchRefusesRecursive), 스폰 완료.

## 4. 병행 구조

- t1863(merge-ready) 착지 대기: 트리거("(card t1860" + 슬롯 free) 발화 시 본 트리를 keep 로 이탈→t1863 트리에서 착지(흡수 — catalog.yaml 는 gen-catalog-hashes→재측정 b2cd 전체→complete)→복귀. 이중 틱 크론 da3de420 무장.

## 자원 입장 규칙 (보드 d-20261018T151201Z-8c1e, 10-18)

- 다음 Go 호출부터 적용: 신규 go test(-run 한정 포함)·go build·go vet·golangci-lint·make — 호출 전 1분 load 32 이하 AND 메모리 free 50% 이상 확인, 첫 충족 후 (레인 18 mod 6)×30s=0s 대기·재확인·시작. 레인당 Go 호출 1회(에이전트 포함), 다중 패키지 -p 4 이하. 착지 remeasure 가 go-test-heavy 보유 중엔 개발 우선순위 신규 호출 금지(Monitor until-loop 60s 재확인). load>32 로 인한 실패는 load 를 함께 기록·조건 충족 시 재실행(결함·PASS 미보고). 시험 제외·timeout 증가·PASS 대체·추가 슬롯 없음. 대리자에게 전달 완료(수리 검증 잔여 호출에 적용).

## Run-phase completion (delegate 1ccf2f355b)

- 구현 커밋: `1ccf2f355b` fix(cli): refuse moai codex -l before execution in unusable terminals — 6 파일 +315행, 트리 클린, 미푸시.
- 3 RED-GREEN: RED-1 DumbTerm(런치 진행 1회 — 거부 부재 실측) / RED-2 Recursive / RED-3 샌드박스 시그니처(guard 수준) → 전부 GREEN. 롤백 양성 대조 시험 추가(spawn 문 등록→스탬프 실패→정확한 행 롤백).
- 삽입점 판단(브리프 정정): -l 경로는 defaultCodexSpawnLaunch 를 통과하지 않음 — runCodex→runCodexFactoryLane 진입(codexLaneParentCheckout 이전)이 사이드이펙트 전 외곽점. ps 프로브는 homestate 소유(exec-primitive 가드 준수), TestMain 핀으로 기존 시험 무관성 유지.
- 검증: 패밀리 8종 + 새 시험 5종 전부 ok, gofmt 클린, native/windows 빌드 OK, 자원 규칙 하 호출 전 입장 확인.
- 레인 독립 스팟검증: DumbTerm `ok 3.660s` · Recursive `ok 3.077s` · gofmt 클린.
- Gaps: TERM-unset+비단말 stdout 분기 전용 시험 없음(같은 거부문 공유, 코드 검토만). CI 는 통합 소관.
- 20:2xZ — **패턴 kill 금지 규칙 접수(보드 d-20261018T202324Z-2fe7)**: pkill·killall·kill $(pgrep …) 금지 — 직접 시작하고 기록한 PID만 멈춤, 에이전트는 TaskStop, 가드 거부 시 재시도 금지. 감사자·전문가 브리프 필수 문장. 활성 sync-auditor 에게 전달 완료.
- 20:5xZ — **콜드 감사자 최종 보고(3사실)**: (1) **종결체인 영수증 게이트 — 리드 결정 에스컬레이션**: 시작 마커가 감사 종료 후 기록돼 기존 영수증 전부 "시작 이전" 판독; MCP audit_multi 는 pkill 사고로 세션 표면 철수, CLI `moai verify codex-review` 는 fail 시 무기록 실측 → 자력 재발행 불능. 수리 경로: ① MCP 부활 후 감사자 audit_multi 재호출, ② 아래 수리 착지 후 라운드-4 델타 재심사 — 리더 결정. (2) 신규 codex 리뷰(0.163.0, HEAD 3d66b477d9) — **P2 2건, 판정 PASS-WITH-DEBT 불변**(선택 등급): P2-A shell_environment_policy.inherit=core/none 시 마커 미도달 → codexShellEnvSetArgs 미러 추가가 채무 D-t1779-c 의 구체 수리 처방; P2-B netScrubLaneEnv·clearFactoryAmbientEnv 마커 세정 누락 → 신규 채무 D-t1779-d. 레인 수리 항목으로 판정 파일 기재 완료. (3) 감사자 메모리 영속화 완료(agent-memory/sync-auditor). 착지 시 라운드-4 델타 재심사 전환 가능.
- 21:0xZ — **리더 t1779 처분 접수**: (1) 종결 체인 영수증은 ① 경로 — MCP 재연결 후 정상 경로 audit_multi 재호출로 새 영수증 수령, CLI fail-무기록·우회 금지. (2) P2-A(codexShellEnvSetArgs 마커 미러·D-t1779-c 처방)·P2-B(netScrubLaneEnv·clearFactoryAmbientEnv 마커 세정)를 MCP 대기 중 지금 수리 — 각 실패 경로 RED 먼저, 시험은 8c1e 입장 조건. 완료 후 MCP 복귀 시 라운드-4 델타 감사(codex 필수 pass)가 (1)의 새 영수증 겸임. 착지 순번 맨 뒤 유지. 대리자 재위임 완료.
- 21:2xZ — **P2-A/B 수리 완료(리더 round 3)**: 커밋 `04989cba68`(4 파일 +61/−2, 트리 클린). P2-A — 셸-env 미러 입력에 마커 추가(codexShellEnvSetArgs 경로, inherit=core/none 커버; child-env 스탬프는 inherit=all 용 유지). P2-B — netScrubLaneEnv·factoryAmbientEnvKeys(clearFactoryTestEnv 무료 위생) 마커 세정. RED verbatim(미러 부재·양 경로 누출) → GREEN. 레인 독립 스팟검증: MirrorsSessionMarker ok 1.514s · AmbientScrub ok 0.687s · gofmt 클린. 잔여: plain-launch argv 미러 없음(round 2 기록 유지). **라운드-4 델타 감사는 MCP 복귀 후 예정(새 영수증 겸임).**
