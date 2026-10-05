# 아키텍처 개요

> `/moai codemaps`로 생성된 아키텍처 지도입니다. 규모 표와 엣지·fan-in 수치는 아래 **재측정 트리**에서
> 직접 잰 것이며, 다른 트리·다른 시점에서 옮겨온 값은 없습니다. **Go** 버전도 이번에는 재측정 트리의
> `go.mod`에서 직접 읽었습니다(`go 1.26.8`).

**모듈**: `github.com/modu-ai/moai-adk` · **Go**: 1.26.8
**현재 부분 재측정**: worktree `.claude/worktrees/t1510`, 브랜치 `WT-self-improve-protected-zone`, base `d0378d37c`(develop 팁 — 재생성 전 팁), 카드 t1510. 앵커 `d0378d37c`(t1524 판 본문) 뒤 비테스트 Go 소스 변경 57개(`IsDescribedWorthy` 술어 — 검사기 값과 같은 술어로 독립 재현, 신규 17 · 수정 40 · 삭제 0)를 대조했다. 창은 일곱 카드 착지분이다 — 자기 개선 보호 구역(card t1510 — 신규 `internal/hook/protected_zone_guard.go`·`internal/hook/protected_zone_shell.go`·`internal/hook/protected_zone_path.go` 3파일과 `internal/config` 로더 2파일, 매니페스트 2벌; 아래 레이어 표 hook·config 행) · 카드 발행 품질(card t1454 — `internal/cli/todo_issuance.go`·`factory_bundle.go` 신규, cli 행 +6) · 메모리 정리(card t1502 — `internal/cli/memory_fold.go` 신규와 memo/taxonomy +2) · 직렬 슬롯 기록 운전자(card t1513) · 게이트 바이너리 노후(card t1528) · 로스터 가드(card t1525·c6ad7989c). 카드별 서술은 § `modules.md` t1510 판이 운반한다. § 규모 표의 일곱 값을 같은 명령으로 다시 쟀다 — 비테스트 1559→**1576**, 테스트 2851→**2885**, 패키지 169 불변(`internal/hook/memo/taxonomy`는 앵커에도 존재 — 이번 창 +2 파일), 최상위 디렉터리 87 불변(internal 83), 내부 import 엣지 476/301→**476/313**, 임베드 템플릿 606→**607**. go.mod는 앵커 이후에도 한 줄도 바뀌지 않았다.
**이전 부분 재측정**: worktree `.claude/worktrees/develop`, 브랜치 `develop`, base `d0378d37c`(develop 팁 — 재생성 전 팁), 카드 t1524. 앵커 `f4c483a5a`(t1485 판 스탬프) 뒤 비테스트 Go 소스 변경 50개(`IsDescribedWorthy` 술어 — 검사기 값과 같은 술어로 독립 재현, 신규 23 · 수정 27 · 삭제 0)를 대조했다. 창은 여섯 카드 착지분이다 — `.moai` 위생(card t1518 — 신규 `internal/hygiene`, 아래 레이어 표 infrastructure 행) · 실행 바이너리 신선도(card t1465 — `internal/session`의 ccversion 4파일과 cli session·doctor 표면) · 감사 상한(card t1500 — `internal/runtime/audit_ceiling.go`와 `moai spec ceiling`, 세 감사 리졸버의 fail-closed) · 훅 수리 배치(card t1499) · 하네스 보존 수리(card t1463·t1467) · 문언·라벨(card t1504·t1517). 카드별 서술은 § `modules.md` t1524 판이 운반한다. § 규모 표의 일곱 값을 같은 명령으로 다시 쟀다 — 비테스트 1536→1559, 테스트 2828→2851, 패키지 168→169(신규 `internal/hygiene`), 최상위 디렉터리 86→87(internal 82→83), 내부 import 엣지 470/295→476/301, 임베드 템플릿 604→606(t1483 카드의 rules 분할 원본 `factory-dispatch-cards.md`·`factory-dispatch-gates.md` 신규). go.mod는 앵커 이후에도 한 줄도 바뀌지 않았다. 테스트 0 패키지 6도 재확인해 멤버 변동이 없었다.
**이전 부분 재측정**: 카드 워크트리, 브랜치 `WT-codemaps-regen3`, base `83086bec5`(로컬 develop 팁), 카드 t1485. 앵커 `27aa8e282`(t1456 판 본문) 뒤 `moai graph check`가 보고한 described-source-diff 294(임계 40)와 인용 부재 12(큐 도메인 패키지 개명 — card t1399)를 이 트리에서 소진한다. 카드별 서술은 § `modules.md` t1485 판이 운반한다. § 규모 표의 일곱 값을 같은 명령으로 다시 쟀다 — 비테스트 1496→1536, 테스트 2735→2828, 패키지 165→168, 최상위 디렉터리 84→86(internal 80→82 — 개명 1:1에 `internal/decision`·`internal/auditverdict` 신규), 내부 import 엣지 461/289→470/295, 임베드 템플릿 604 불변. go.mod는 앵커 이후에도 한 줄도 바뀌지 않았다.
**이전 부분 재측정**: worktree `.moai/worktrees/t1456`, 브랜치 `WT-codemaps-regen2`, base `5501c06af`(develop 흡수 후), 카드 t1456. 앵커 `a2e03d8e0`(t1443 판 본문) 뒤 비테스트 Go 소스 변경 50개(`IsDescribedWorthy` 술어 — 검사기 값과 같은 술어로 독립 재현, 신규 8 · 수정 42)를 대조했다. 창은 아홉 Go 카드 착지분이다 — 할당량 게이트의 워크트리 기록 판독(card t1442) · 감사 모델 소비화(card t1423) · 스테일 런 자가 치유(card t1345) · codex 리뷰 소유권(card t1422) · 레인 재점검 룰(card t1451) · 파싱 실패 줄 보존(card t1433) · 다중 런 합류 선택 해제(card t1444) · 싱크 게이트 생존·hpp/hxx(card t1420·t1412 잔여) · CI 수리 배치(card t1455 — 신규 표면 없음); Aside 브라우저(card t1439)는 Go 소스 0으로 임베드 템플릿 몫이다. 카드별 서술은 § `modules.md` t1456 판이 운반한다. § 규모 표의 일곱 값을 같은 명령으로 다시 쟀다 — 비테스트 1488→1496, 테스트 2699→2735, 패키지 165 불변, 최상위 디렉터리 84 불변(internal 80), 내부 import 엣지 460/288→461/289, 임베드 템플릿 602→604(`moai-ref-aside-browser/SKILL.md`·`run/external-delegation.md` 신규 — card t1439·t1455). go.mod는 앵커 이후에도 한 줄도 바뀌지 않았다. 이 판에서 엣지 산식 문언을 실측 필터와 일치시켰다(t1443 감사 F1 — 아래 표의 패키지 단위 행). 이전 판 산문의 fold 단위 전체 경로 표기 유무는 fold 가드 재측정으로 확인했다(초록 유지).
**이전 부분 재측정**: worktree `.moai/worktrees/t1443`, 브랜치 `WT-codemaps-regen`, base `4bf547bca`, 카드 t1443. 앵커 `c2703f698`(t1297 병합 판) 뒤 비테스트 Go 소스 변경 71개(`IsDescribedWorthy` 술어 — 검사기 값과 같은 술어로 독립 재현, 신규 15 · 수정 56 · 삭제 0)를 대조했다. 창은 14카드 착지분이다 — 관리 세션 계층(card t1375) · quota-aware 임대(card t1347) · web 에이전트 설정 부활(card t1411) · 에이전트 3층 등급(card t1391) · 런치 모델 6단 선위(card t1441) · 보고서 아티팩트 전달(card t1427) · 우선순위 --auto 픽(card t1400) · Jev 근접중복 비신호화(card t1428) · CI 수리(card t1384) · acceptEdits 수리(card t1414) · 보존 단일 작성자(card t1425) · 세션 종료 착지 가드(card t1393) · 리더 합류 공지 게이트 요약(card t1344) · Jev 표시 전용 문언(card t1403) — 카드별 서술은 § `modules.md` t1443 판이 운반한다. § 규모 표의 일곱 값을 같은 명령으로 다시 쟀다 — 비테스트 1469→1488, 테스트 2653→2699, 패키지 164→165(신규 `internal/settings/agentfm` — card t1411 착지분; t1305 판이 지웠던 하위 패키지의 복귀), 최상위 디렉터리 84 불변(internal 80), 내부 import 엣지 457/286→460/288, 임베드 템플릿 601→602(`moai-domain-html-report/references/artifact-contract.md` 신규 — card t1427). 표기값 1469와 앵커 시점 실측 1473의 차이 넷은 t1297 측정 트리와 앵커 사이 흡수분이다. go.mod는 앵커 이후 처음 움직였다 — `github.com/gorilla/websocket` 신규 직접 require(관리 세션의 Codex App-Server 전송, card t1375)와 `santhosh-tekuri/jsonschema/v6`의 indirect→직접 승격(§ `dependencies.md` t1443 판). 이 판은 이전 판 산문의 fold 단위 전체 경로 표기 한 곳도 벌거벗은 파일명 형태로 고쳤다(t1297 판의 `cwd_changed_relocate.go` — 내용 불변, 표기 형태만).
**이전 부분 재측정**: worktree `.moai/worktrees/t1297`, 브랜치 `WT-codemaps-regen`, base `a9f43a6fc`, 카드 t1297. 전면 재생성 카드로, `moai graph check`가 보고한 미커버 창(stale 48 · 임계 40)을 이 트리에서 소진한다. 앵커 `0a8780201`(t1378 판 스탬프) 뒤 비테스트 Go 소스 변경 48개(`IsDescribedWorthy` 술어 — 검사기 값과 같은 술어로 독립 재현)를 대조했다. 창은 13카드 착지분이다 — gate 계열(card t1379 sync-gate 다언어 검사 · t1383 codex 리뷰 게이트 스코핑(`codex_review_scope.go`) · t1385 캐시 결함 형식화 · t1388 모노레포 go vet · t1389 멀티모듈 CHANGED_LANGS), **card t1338 레인 자율 배치 착지**(`internal/factorylane` 패키지와 cli 4파일 — 본문 서술은 t1338 판이 이미 갖춰 있고 이 판이 창으로 소진한다), card t1373 stale-run 라벨(`internal/hook/stale_run_gate.go`·`internal/factorymsg/run_state.go`), card t1381 templ 재생성(`internal/web/*_templ.go` — 파일 수 불변), card t1387 문서 위임 cwd 고정(`internal/hook/cwd_changed.go`와 같은 패키지의 `cwd_changed_relocate.go`), card t1339 세션 앵커 귀속 수리(`internal/session/anchor_trace.go`·`anchor_relocate_audit.go` 신규 — 24→26), card t1394 GLM 창 픽스처(테스트 전용 — described 0). § 규모 표의 일곱 값을 같은 명령으로 다시 쟀다 — 비테스트 1455→1469, 테스트 2628→2653, 패키지 163→164(`internal/factorylane` — t1338 착지분), 최상위 디렉터리 83→84(internal 79→80), 내부 import 엣지 453/282→457/286, 임베드 템플릿 601은 재확인 결과 변동 없었다. 본문 내용 갱신(훅 stale-run 게이트 · 세션 앵커 trace·audit · factorymsg run-state · config 앵커 설정(card t1339) · dependencies 전면 재측정)은 § `modules.md`·`entry-points.md`·`dependencies.md`의 해당 행이 나른다.
**이전 부분 재측정**: worktree `.claude/worktrees/t1378`, 브랜치 `WT-codex-lane-slots`, base `0a8780201`, 카드 t1378. 스탬프 앵커 `b8f437bae`(카드 런 베이스) 뒤 비테스트 Go 소스 변경 19개(`IsDescribedWorthy` 술어)를 대조했다. § 규모 표의 값을 같은 명령으로 다시 쟀다 — 비테스트 1452→1455(`internal/cli` +1: `factory_launch_timing.go`, `internal/homestate` +2: `run_capacity.go`·`process_identity_batch.go`), 테스트 2619→2628(이 카드 신규 8 — 클레임 용량·유한 경로 레거시 2, 런치 타이밍 1, 배치 프로브·플랫폼 핀 4, 런 용량 기록 1; 나머지 1은 흡수 창 분), 패키지 163·최상위 디렉터리 83·내부 import 엣지 453/282(`internal/factory`의 `internal/homestate` 임포트는 앵커 이전부터 있었다)·임베드 템플릿 601은 재확인 결과 변동 없었다. 본문 내용 갱신(codex 레인 슬롯 클레임의 용량 인식 성장·런 선언 용량 기록·런치 타이밍 보고·소유자 프로브 배치화)은 § `modules.md`의 해당 행이 나른다.

**이전 부분 재측정**: worktree `.claude/worktrees/t1374`, 브랜치 `WT-t1368-ci-repair`, base `ca7191cba`, 카드 t1374. 스탬프 앵커 `3e6d78f73`(t1351 판) 뒤 비테스트 Go 소스 변경 50개(`IsDescribedWorthy` 술어)를 대조했다. § 규모 표의 값을 같은 명령으로 다시 쟀다 — 비테스트 1436→1452(`internal/cli` +9: `todo_claim`·`todo_show`·`todo_classify`·`todo_ghost_notice`·`doctor_owner_label`·`doctor_todo_ghost`와 `worktree/sweep.go`+`sweep_cwd_{posix,windows}.go`, `internal/factory` +2: `classification.go`·`todo_owner_label.go`, `internal/homestate` +1: `factory_run_resume.go`, 나머지는 창 안 138커밋의 산문·픽스처 인접분), 테스트 2590→2619, 패키지 162→163, 최상위 디렉터리 82→83(internal 78→79), 패키지 단위 내부 import 엣지 448→453·최상위 집계 277→282(`go list` 패키지 의존 쌍 기준), 임베드 템플릿 598→601. 본문 내용 갱신(`moai worktree sweep`·감사 모델 핀과 GLM 구형 모델 삭제·todo claim/show/분류 연쇄·`--leader` 개명)은 § `modules.md`·`entry-points.md`의 해당 절이 나른다.
**이전 부분 재측정**: worktree `.moai/worktrees/t1351`, 브랜치 `WT-codemaps-refresh10`, base `145c3d98c`, 카드 t1351. 스탬프 앵커 `145c3d98c`(t1333 판의 트리) 뒤 비테스트 Go 소스 변경 54개를 대조했다. § 규모 표의 값 중 셋이 움직였다 — 비테스트 1429→1436(`internal/cli` +4: `clean_reports_archive.go`·`factory_lane_relaunch.go`·`mcp_factory_card.go`·`mcp_todo.go`, `internal/homestate` +1: `card_worktree.go`, `internal/settings` +1: `projectscalars.go`), 테스트 2569→2590, 패키지 단위 내부 import 엣지 449→448. 같은 명령으로 재확인해 변동 없음: 패키지 162 · 최상위 디렉터리 82(internal 78) · 최상위 집계 엣지 277(`internal/cli` 70·`internal/hook` 39 포함) · 임베드 템플릿 598. § 규모 표의 나머지 넷은 이 값들로 갱신했고, 본문 내용 갱신(F1 자가 배차 표면·main 커밋 금지 제2 거부 클래스·reports 생명주기·무손실 설정 저장)은 § `modules.md`·`entry-points.md`·`data-flow.md`의 해당 절이 나른다.
**이전 부분 재측정**: worktree `.moai/worktrees/t1333`, 브랜치 `WT-codemaps-refresh9`, base `145c3d98c`, 카드 t1333. 스탬프 앵커 `afecf81e9`(t1257 문서층 병합판) 뒤 비테스트 Go 소스 변경 44개(`IsDescribedWorthy` 술어)를 대조했다. § 규모 표의 일곱 값을 같은 명령으로 다시 졌다 — 비테스트 1422→1429(신규 7 — `internal/cli` 3·`internal/factory` 4), 테스트 2545→2569, 패키지 162·최상위 디렉터리 82(`internal` 78)·내부 import 엣지 449/277·임베드 템플릿 598은 재확인 결과 변동 없었다. § 도식 절의 스테일 파일 수 한 곳(`internal/hook` 141→155 — 앵커 이전부트 스테일)을 이 트리 실측으로 정정했다.
**이전 재측정**: worktree `.claude/worktrees/t1305`, 브랜치 `WT-codemaps-refresh8`, base `afecf81e9e96`, 카드 t1305. 스탬프 앵커 `a3a9e653e`(t1295 판) 뒤 비테스트 Go 소스 변경 75개(`IsDescribedWorthy` 술어)를 대조했다. § 규모 표의 다섯 값(비테스트·테스트 파일 수, 패키지 총수, 엣지 둘)을 같은 명령으로 다시 쟀다 — 비테스트 1419→1422, 테스트 2544→2545, 패키지 164→162(하네스의 cellguard 가드 패키지와 settings의 agentfm 스키마 패키지 둘이 소멸 — t1246 배치가 에이전트 모델 표면을 은퇴시키며 함께 갔다), 내부 import 엣지 452→449·278→277. 임베드 템플릿 598은 같은 명령으로 재확인해 변동이 없었다. 최상위 디렉터리 82(internal 78 + cmd 2 + pkg 2)와 테스트 0 패키지 6도 재확인해 변동이 없었다. § 들어맞지 않는 패키지의 스테일 파일 수 세 곳(`internal/template` 31→36, `internal/homestate` 15→27, `internal/harness` 87→86)과 `cellguard` 은퇴 산문을 이 트리 실측으로 정정했다.
**이전 재측정**: worktree `.moai/worktrees/t1295`, 브랜치 `WT-codemaps-freshness`, base `cee197917`, 카드 t1295. 앵커 `fdc5361c3` 뒤 비테스트 Go 소스 81개가 바뀌었다. 아래 규모 표는 이 트리에서 `find`와 `go list -deps -json ./...`로 다시 셌다. 이번 내용 갱신은 MoAI 워크트리 위치·이전 경로와 Factory 역할 전환에 한정한다. 과거 판의 변경 설명은 이력으로 남긴다.
**최초 측정 트리**: worktree `.claude/worktrees/t592`, 브랜치 `WT-home-state-rollout`, HEAD `e7bd89ee3`, 2026-09-10
**재측정 트리**: worktree `.claude/worktrees/t869`, 브랜치 `WT-codemaps-refresh`, HEAD `a851b205c`, 2026-09-18 — § 규모 표 전체, § 구조 판정의 수치, 레이어 표의 대표 패키지, § 도식에 들어맞지 않는 패키지의 파일 수와 신규 항목(`internal/mission`·`internal/codextools`). 측정 명령은 각 표의 산출 명령 칸에 있습니다. 파일 크기(KB) 서술은 이번에 다시 쟀고, 그 밖의 서술형 판단은 앞 판을 이어받았습니다.
**정정 재측정**: worktree `.claude/worktrees/t872`, 브랜치 `WT-codemaps-citations`, HEAD `9a8cc4277`, 2026-09-18 — § 규모 표의 다섯 값(비테스트·테스트 파일 수, 패키지 총수, 최상위 디렉터리 수)과 테스트 전용 디렉터리 서술. 위 재측정 직후 한 패키지가 삭제돼 그만큼만 다시 쟀고, 나머지 값(엣지 365/222, 임베드 589)은 같은 명령으로 재확인해 변동이 없었습니다.
**정기 재측정**: worktree `.claude/worktrees/t999`, 브랜치 `WT-codemaps-remediation`, HEAD `56c64891a`, 2026-09-20 — § 규모 표 **일곱 값 전부**와 `internal/cli` fan-out, § `modules.md`의 패키지별 파일 수, 그리고 신규 패키지 3개(`internal/auditreceipt` · 하네스의 `rosterguard` · `cellguard`)의 서술. 각 값의 산출 명령은 표 안에 있고, 전부 이 트리에서 직접 실행했습니다. 서술형 판단 중 이번에 다시 확인한 것은 테스트 0 패키지 4개와 테스트 전용 디렉터리 1개뿐이며, 나머지 구조 판정은 앞 판을 이어받았습니다.
**정기 재측정**: worktree `.claude/worktrees/t1069`, 브랜치 `WT-graph-restamp`, HEAD `0314801c2`, 2026-09-22 — § 규모 표의 여섯 값(비테스트·테스트 파일 수, 패키지 총수, 최상위 디렉터리 수, 엣지 둘)과 § 구조 판정의 `internal/cli` import 수, § 들어맞지 않는 패키지의 파일 수 두 곳(`internal/hook` · `internal/homestate`), 신규 패키지 3개(`internal/jev` · `internal/jevcred` · `internal/jevmeasure`)의 서술. 각 값의 산출 명령은 표 안에 있고 전부 이 트리에서 직접 실행했습니다. 임베드 템플릿 파일 수(588)와 테스트 0 패키지 4개·테스트 전용 디렉터리 1개도 같은 명령으로 재확인해 변동이 없었습니다.
**부분 재측정**: worktree `.claude/worktrees/t1092`, 브랜치 `WT-codemaps-restamp`, base `08113ff0f`, 2026-09-23 — 카드 t1092. § 규모 표 일곱 값을 같은 명령으로 다시 쟀습니다 — 비테스트 1259→1272, 테스트 2155→2190, 패키지 총수 151→152(신규 `internal/factorymsg`), 최상위 디렉터리 77→78(`internal` 73→74), 내부 import 엣지 378→381(패키지 단위)·234→237(최상위 집계), 임베드 템플릿 588→589. § 구조 판정·§ 도식에 들어맞지 않는 패키지 절은 이번 변경과 무관해 손대지 않았습니다.

**이전 재측정**: worktree `.claude/worktrees/t1274`, 브랜치 `WT-codemaps-refresh4`, base `cf4b82755`, 2026-09-26 — 카드 t1274. `find internal cmd pkg`로 비테스트·테스트 Go 파일을, `go list ./...`로 패키지를, `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...` 후 모듈 경로 필터로 내부 import 엣지를, `find internal/template/templates -type f`로 임베드 원본을 다시 셌습니다. 앵커 `4a05fd3d6`(card t1238) 이후 described-worthy(비테스트 Go·testdata 제외) 끝점 변경은 52개이며, 이번 판은 신규 패키지 `internal/escalation`(+`escalationtest`) 서술과 `internal/cli`·`internal/hook`·`internal/config`·`internal/auditreceipt` 행 갱신으로 반영했습니다. 아래 규모 표의 값은 이 트리의 값으로 갱신했습니다.

**이전 재측정**: worktree `.claude/worktrees/t1187`, 브랜치 `WT-codemaps-source-refresh`, HEAD `a8a9b9376`, 2026-09-25 — 카드 t1187. `find internal cmd pkg`로 비테스트·테스트 Go 파일을, `go list ./...`로 패키지를, `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...`로 내부 import 엣지를, `find internal/template/templates -type f`로 임베드 원본을 다시 셌습니다. 앵커 `bd71c59e4` 이후 끝점 변경은 전체 126개 파일 중 게이트가 세는 비테스트 Go 소스 41개입니다. 아래 규모 표의 값은 이 트리의 값으로 갱신했습니다.

---

## 규모

| 값 | 수치 | 산출 명령 |
|---|---|---|
| 비테스트 Go 파일 | 1559 | `find internal cmd pkg -name '*.go' -not -name '*_test.go' \| wc -l` |
| 테스트 Go 파일 | 2851 | `find internal cmd pkg -name '*_test.go' \| wc -l` |
| Go 패키지 총수 | 169 | `go list ./... \| wc -l` |
| 최상위 디렉터리 | 87 | `internal` 83(`ls -d internal/*/`) + `cmd` 2 + `pkg` 2 |
| 내부 import 엣지 (패키지 단위) | 476 | `go list -deps -json ./...`의 프로젝트 패키지 `Imports` 중 **모듈 내부 경로**(`github.com/modu-ai/moai-adk/` 접두 — `internal`·`pkg`·`cmd` 전부 포함; `internal` 전용 필터로는 이 값보다 작게 나온다 — t1456 판에서 t1443 감사 F1대로 문언을 실측 필터와 일치시켰다) |
| 내부 import 엣지 (최상위 집계) | 301 | 위를 `internal/<X>` · `pkg/<X>` · `cmd/<X>` 수준으로 접고 self-edge 제거 |
| 임베드 템플릿 파일 | 606 | `find internal/template/templates -type f \| wc -l` |

테스트 대 비테스트 비율은 **1.83 : 1**입니다. `go list`의 패키지 가운데 테스트 Go 파일이
0개인 곳은 6개입니다. 그중 `cmd/moai`·`cmd/t657-merge`·`internal/template/scripts`·
`scripts/convert-nextra-to-hextra`는 실행 파일이고, `internal/closure/closuretest`·
`internal/escalation/escalationtest`는 다른 패키지의 테스트가 쓰는 픽스처입니다.
`internal/skills`는 비테스트 Go 파일이 없는 테스트 전용 디렉터리입니다.

> **엣지 수 정정 이력.** 앵커 `25a3212a9` 판은 패키지 단위 엣지를 1638로 적었지만 당시
> 명령 문자열이 생략형이라 재현할 수 없었습니다. 이후 `52f863f36` 판은 완전한 명령으로
> 345 / 208을, `e7bd89ee3` 판은 351 / 214를, `9a8cc4277` 판은 365 / 222를, `56c64891a` 판은
> 371 / 227을 측정했고, t1069 판은 같은 명령으로 378 / 234를 측정했습니다. 판마다 늘기만 한
> 것은 추세가 아니라 이 기간에 패키지가 삭제되지 않았다는 사실의 반영입니다.
>
> **이전 판 이후 트리에서 사라진 것.** `internal/gateway`(하위 `auth`·`conversation`·`opaque`·
> `receipt`·`translate` 포함), `internal/codexapp`, `internal/codexbridge`와 CLI 쪽
> `internal/cli/gateway_*.go`·`internal/cli/gpt.go`·`internal/cli/gpt_auth.go`는 지금 트리에
> 존재하지 않습니다. 이전 판은 이 패키지들을 서술하지 않았으므로 지울 서술도 없었습니다.
> `internal/harness/cellguard`도 같은 자리다 — docs-site profile-matrix 셀 감시 가드였으나
> t1246 배치가 profile-matrix 표면째 지웠고(t1305 판 경유 삭제), 본문의 은퇴 산문은
> 벌거벗은 이름으로 남는다. fold-judgments 의 omission 기록은 이 블록이 소극 인용으로 덮는다(t1443 판).

---

## 이 코드베이스가 실제로 따르는 구조

**깔끔하게 들어맞는 이름은 없습니다.** 가장 가까운 것은 Go 표준 프로젝트 레이아웃에 얇은
헥사고날 시도를 얹은 형태지만, 실제로 지배적인 것은 **명령별 수직 슬라이스를 가진 모듈러
모놀리스**입니다.

정직하게 한 줄로 적으면 — **SPEC 단위로 증식한 수평 패키지 위에 `internal/cli`라는 단일
거대 어댑터가 얹힌 구조**입니다.

### 헥사고날에 부합하는 근거

- `cmd/` · `internal/` · `pkg/` 3분할과 `internal/`의 78개 디렉터리 분해는 표준 레이아웃 그대로입니다.
- 합성 루트가 명시적으로 하나 있습니다 — `internal/cli/deps.go`의 `Dependencies` 구조체와
  `InitDependencies()`. `git.Repository`, `hook.Registry`, `hook.Protocol`, `update.Checker`,
  `update.Orchestrator` 같은 인터페이스 타입으로 조립하므로 포트/어댑터 의도가 보입니다.
- `internal/hook/registry.go`의 `Register` / `Dispatch`는 교과서적인 핸들러 레지스트리 + 체인입니다.
- 안정 의존성 원칙을 만족합니다 — fan-in 상위를 `defs` · `paths` · `execerr` · `atomicfile` ·
  `models` 같은 cross-cutting leaf가 차지하고, 그것들의 fan-out은 0에 가깝습니다.

### 부합하지 않는 근거 — 이쪽이 더 결정적입니다

- **`internal/cli`가 다른 최상위 패키지 66개를 import 합니다**(최상위 집계 엣지 기준).
  헥사고날이라면 어댑터 하나가 전 도메인에 닿을 이유가 없습니다. 실제 모양은 "명령 하나 =
  파일 하나 = 그 명령이 필요한 것 전부 import"에 가깝습니다.
- **도메인 로직이 어댑터 안에 삽니다.** `internal/hook/quality/gate.go`가 67KB,
  `internal/hook/session_start.go`가 67KB, `internal/cli/hook.go`가 61KB입니다(`ls -l` 바이트 ÷ 1024).
  이들은 프로토콜 변환이 아니라 정책입니다.
- **자율 미션의 정책 계층도 어댑터에만 연결돼 있습니다.** 새로 생긴 `internal/mission`(15 파일)의
  비테스트 소비자는 `internal/cli/goal.go` 하나뿐이며, 그 파일이 `mission.*` 심볼을 직접 엮어
  `moai goal --auto` 표면을 만듭니다.
- **레이어 방향이 국소적으로 뒤집힙니다.** presentation인 `internal/hook`이 다른 최상위 패키지
  6개에게, `internal/statusline`이 5개에게 import 당합니다.
- **DI가 전역 변수 하나로 전달됩니다.** `var deps *Dependencies`는 컨테이너가 아니라 전역
  상태이고, 그래서 `if deps == nil` 형태의 nil 방어가 곳곳에 필요해졌습니다.
- **패키지 경계가 응집도가 아니라 SPEC 단위로 그어졌습니다.** 대부분의 패키지 doc 코멘트가
  `SPEC-XXX-NNN` 형태로 시작합니다. `goal` / `loop` / `ralph`가 셋으로,
  `guardliveness` / `guardstate`가 둘로 쪼개진 것이 그 결과입니다. `internal/stateanchor`(1 파일) ·
  `internal/chain`(4 파일) · `internal/gitenv`(1 파일)도 같은 증식의 사례이고, 이 판에서는
  `internal/mission`(15 파일)이 더해졌습니다.

---

## 레이어

패키지의 **주된 대화 상대**로 판정했습니다.

| 레이어 | 판정 규칙 | 대표 패키지 |
|---|---|---|
| presentation | 프로세스 경계 바깥의 표면(터미널·HTTP·훅 프로토콜)과 직접 말한다 | `cmd/moai`, `internal/cli`, `internal/hook`, `internal/tui`, `internal/web`, `internal/statusline`, `internal/mcp` |
| business/domain | MoAI 고유 규칙·정책만 담고 자체 I/O 프리미티브를 소유하지 않는다 | `internal/spec`, `internal/harness`, `internal/navigator`, `internal/factory`, `internal/graph`, `internal/mx`, `internal/mission`, `internal/contract` … |
| data/persistence | 디스크상 named artifact 하나의 스키마와 읽기·쓰기 계약을 소유한다 | `internal/config`, `internal/session`, `internal/settings`, `internal/manifest`, `internal/chain`, `internal/homestate`, `internal/auditreceipt` … |
| infrastructure/platform | 외부 프로세스·OS·네트워크 설비를 감싼다 | `internal/lsp`, `internal/git`, `internal/github`, `internal/astgrep`, `internal/hygiene`, `internal/tmux` … |
| cross-cutting | 정책이 없고 무관한 다수 패키지가 쓰는 leaf (fan-in ≥ 5, 도메인 지식 없음) | `internal/defs`, `internal/paths`, `internal/atomicfile`, `pkg/models`, `internal/stateanchor` … |

전체 배치는 `modules.md`에 있습니다.

### 도식에 들어맞지 않는 패키지

분류가 어긋나는 자리는 반올림이 아니라 **발견**이므로 그대로 적습니다.

- **`internal/hook` (155 파일)** — 가장 큰 불일치입니다. 겉으로는 Claude Code 훅 JSON을
  stdin에서 읽어 stdout으로 내보내는 인바운드 어댑터지만, 안에 브랜치 가드·세션 시작
  오케스트레이션·증거 기록기 같은 순수 정책이 함께 삽니다. presentation으로 부르면 정책이
  감춰지고 domain으로 부르면 stdin/stdout 계약이 감춰집니다. 어느 쪽이든 손실이 있습니다.
- **`internal/factory` (61 파일)** — 카드 도메인 규칙(`role.go` — t1399 판에서 보드·컬럼·정합성 조정 파일이 은퇴하고 역할 모델만 남았다)과
  SQLite 스토리지 엔진(`backlog_sqlite.go` — WAL · busy_timeout · IMMEDIATE 트랜잭션을 직접
  소유)이 한 패키지에 있습니다. 여기에 워킹 트리 설정 파일을 검사하는
  `settings_drift.go`까지 들어와 domain · data · 워킹 트리 검사 셋이 한 자리에 있었고, 이 판에서
  GTD 계층 10개 파일(`gtd_*.go` · `backlog_gtd_schema.go` — capture/clarify/organize/reflect/engage,
  관계, prepared operation, export/import, 스키마 마이그레이션)이 같은 `backlog.db` 위에 더해졌습니다.
- **`internal/codextools` (2 파일)** — 비테스트 import가 0인 패키지입니다. 네이티브·지연 디스패처
  도구 레지스트리를 인증된 대화 하나에 묶는 역할을 패키지 주석이 밝히지만, 트리 안의 소비자는
  자기 테스트뿐입니다(§ `modules.md` 네거티브 스페이스).
- **`internal/template` (36 파일)** — 도메인(카탈로그·모델 정책), 데이터(598개 파일의
  `//go:embed all:templates` 트리), 인프라(배포기)를 동시에 수행하고, 하위에 두 개의
  **기계 방출기**(`agentemit` · `commandemit`)를 품습니다.
- **`internal/core`** — 이름과 달리 응집된 core가 아닙니다. `core/git`은 인프라,
  `core/project` · `core/quality`는 도메인이며, `core/integration`과 `core/migration`은
  `.gitkeep` 하나뿐인 빈 디렉터리입니다.
- **`internal/mcp` (1 파일)** — `catalog.go`의 도구 이름 + write 여부 선언 리스트뿐입니다.
  프로토콜 계약 선언이라 presentation에 두었으나 실질은 `internal/defs`와 같은 상수 leaf입니다.
- **`internal/skills`** — 비테스트 Go 파일이 0개이고 `workflow_split_test.go` 하나만 있습니다.
  어떤 레이어에도 속하지 않습니다.
- **`internal/stateanchor` (1 파일)** — 정책이 없는 leaf처럼 보이지만 담는 것은 **결정 규칙**
  입니다(어느 프로젝트 루트가 상태의 앵커인가). cross-cutting에 두되, 우선순위 사슬 자체가
  요건으로 고정돼 있다는 점에서 순수 leaf와 다릅니다.
- **`internal/harness` (86 파일)** — **레이어가 아니라 네임스페이스입니다.** 하위의
  `rosterguard`(4 파일)는 런타임 경로가 없는 **테스트 시점 문서 드리프트 가드**입니다 —
  `go test`가 유일한 발화 경로이고 CLI·훅·MCP 어디에도 배선돼 있지 않습니다. 형제였던
  `cellguard`는 docs-site profile-matrix 셀 감시 가드였으나 t1246 배치가 profile-matrix 표면을
  은퇴시키면서 함께 사라졌습니다(§ `modules.md` 네거티브 스페이스). 같은 디렉터리의 `delegationmap`은 반대로 라우팅 원장을 읽어
  제안을 내는 프로덕션 분석기입니다. 공유하는 인터페이스도, 도달 경로도, 데이터 흐름도
  없습니다. `internal/harness/` 전체를 묶는 계약을 선언한 패키지 주석은 트리에서 찾지
  못했으므로, 이것은 인용이 아니라 **증거로부터의 추론**으로 적습니다 — 「에이전트·하네스
  메타데이터에 관한 것들」이라는 주제어가 디렉터리를 만들었을 뿐 층을 만들지는 않았습니다.
- **`internal/homestate` (27 파일)** — `~/.moai` 아래 프로젝트별 SQLite 경로와 스키마를
  소유하는 data/persistence 패키지이면서, Unix `flock`·Windows `LockFileEx`, PID 지문,
  런타임 진입 차단까지 함께 다룹니다. 이 판에서 비정준 트리 변이 게이트(`noncanonical_tree_guard.go`
  — 호출자 트리가 canonical 프로젝트 루트가 아니면 mutation 진입을 거절한다. `MOAI_HOME`은
  대상만 옮길 뿐 그 사실을 바꾸지 않는다)이 admission lock과 marker 설치 두 지점에 배선됐습니다.
  저장 계약과 플랫폼 동시성 경계가 한 패키지에 만나는
  의도적인 seam이며 패키지 단위 비테스트 소비자는 `cli`, `cli/ptycaptest`, `hook`, `hook/handoff`,
  `kanban`, `web` 여섯 곳입니다(최상위로 접으면 `cli`·`hook`·`kanban`·`web` 넷).

## HOME 상태 전환 경계

이번 판은 프로젝트 로컬 상태를 즉시 지우는 전환이 아닙니다. `internal/homestate`가
`~/.moai/db/<project-key>/{todo,factory}/`와 전역 `~/.moai/run/profile-leases.db`의 계약을
제공하고, `moai migrate home-state`가 기본 dry-run으로 이전 가능성을 검사합니다. 실제 적용은
`--apply --verified-live`와 두 차례의 zero-active census, admission lock, 백업·해시·논리 동등성
검증을 모두 통과해야 합니다. 현재 코드와 문서의 존재는 **운영 데이터 이전 완료를 뜻하지
않습니다**. 프로젝트의 `.moai/state/todo/backlog.db`는 검증된 apply 전까지 보존됩니다.

---

## 이 문서 자체가 게이트를 가진다

`.moai/project/codemaps/`의 다섯 문서는 설명이면서 동시에 **측정 대상**입니다. `internal/graph`의
freshness 게이트가 `provenance.json`의 스탬프를 기준으로 described roots(`internal`, `cmd`,
`pkg`)의 변경 파일 수를 세고, 임계 40을 넘으면 stale로 판정합니다.

그 판정에는 선행 조건이 있습니다 — 저장된 스탬프가 현재 checkout `HEAD`의 조상이어야
비교 창이 성립합니다. 객체가 해석되는 것과 조상인 것은 다른 조건이고, squash와 rebase는
내용을 남기면서 원래 커밋을 이력 밖으로 밀어냅니다. 비조상 상태는 숫자 없는 미측정(exit 2)이지
큰 숫자의 stale이 아닙니다. 순서와 복구는 § `data-flow.md` I가 소유합니다.

실질적 함의는 하나입니다: **본문을 갱신하지 않은 재스탬프는 이 문서들을 초록으로 만들지
못합니다.** 값은 스탬프가 아니라 본문이 마지막으로 실제 바뀐 지점에서 측정되기 때문입니다.

---

## 관련 문서

- `modules.md` — 패키지별 책임과 파일 수
- `dependencies.md` — fan-in / fan-out 상위와 상호 참조 3쌍
- `entry-points.md` — `main()`, Cobra 트리, 훅, MCP 표면, CI가 읽는 종료 코드 표면
- `data-flow.md` — 계층을 관통하는 경로와 codemaps freshness 게이트
