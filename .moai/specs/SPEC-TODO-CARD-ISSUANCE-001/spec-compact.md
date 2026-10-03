# SPEC-TODO-CARD-ISSUANCE-001 — compact (run-phase load)

카드 발행 품질 — 발행 시점의 겹침·중복 제시, 카드 관계 그래프, 묶음 직렬 경로, 과분할 억제 규칙. Tier L, 카드 t1454, 계획 시작 HEAD `2de0a2cb6`, version 0.4.0. 전체 문서: `spec.md`; 근거: `research.md`; 구조: `design.md`; 계획: `plan.md`; 기준과 RED-now 원장: `acceptance.md`; 결정: `decision-index.md`. 실행 순서는 M0 → M1 → M2 → M3 → M4 → M6 → (게이트가 열리면) M5.

## 요구 (GEARS) — 24

**A — 기준선과 발행 시점 제시 (M0, M1)**
- REQ-TCI-001 The baseline record under `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` — a tracked path present in every clone and in CI — shall hold, for every figure a threshold or rule value rests on, the producing command, its verbatim output and the tree SHA, every later threshold shall be read from it, and the commit that adds it shall carry nothing outside it and precede every commit that changes shipped behavior.
- REQ-TCI-002 When a card is admitted through `moai todo add`, `moai gtd add` or the bare fallthrough, the add path shall print on its error stream up to three most-similar cards (live, dropped with the reason prefix stripped, archived), same-component open cards, in-flight lane expected-file overlap and covering completed SPECs, each labelled with source and measure, and its output stream shall remain exactly `<id> <position>`. (Expected files = paths the card text names ∪ the M2 `files` attribute, plus the lane branch's changed files for an in-flight card; a missing input on either side reads `unmeasured`, never `none`; the judgement is per set of in-flight cards — a card with no input drops out of the comparison unreported — and a path matches only as a whole normalized path, never by basename or prefix.)
- REQ-TCI-003 The presentation shall not refuse, delay beyond its time bound, reorder, edit, fold or drop any card, and any probe outside the queue shall run outside the queue's write lock under a time bound.
- REQ-TCI-004 Where `--dry-run` is passed, the add path shall print the same presentation, write nothing (file byte-identical, no id consumed, `last_seq` unchanged) and exit 0, reporting that a real add would refuse an exact duplicate.
- REQ-TCI-005 When the MCP `todo_add` tool or `moai gtd engage` admits a card, the admission shall carry the presentation (MCP: after the first `<id> <position>` line; engage: error stream) without recording a finding on engage and without changing the MCP first line.
- REQ-TCI-006 The lookup shall be a read-only path separate from the classifier, which, with the normalizer, the Jaccard measure and the threshold, shall keep behavior and tests, including the dropped-card skip.

**B — 카드 스키마 (M2)**
- REQ-TCI-007 Every card shall be able to carry optional issuance attributes (spawning card, origin, size estimate, expected files, drop reason), absence represented as absent rather than as an empty value, so a card carrying none is stored, serialized and exported exactly as before.
- REQ-TCI-008 When a database predating the change is opened, the engine shall add the storage for issuance attributes to both card tables and for dispositions to both finding tables without altering any row, and the pure reader shall return absent values without running schema changes.
- REQ-TCI-009 Archive and restore, and a merge of two queues, shall carry the attributes and each finding's disposition unchanged, and the in-memory, JSON and SQLite renderings shall agree (parity covers all the new storage).
- REQ-TCI-010 When a card is dropped, `drop` shall store the reason in the attribute while keeping the `[DROPPED — <reason>] ` prefix, and a card's closing time shall be readable through one accessor deriving it from the archive or drop stamp, answering "unknown" when neither exists.
- REQ-TCI-011 A finding shall be able to carry an optional disposition (accept, merge, reject), set only through an explicit operator verb, absent for existing findings, changing no card, relation or order.

**C — 관계 모델과 그래프 (M3)**
- REQ-TCI-012 Card relations shall be expressed in seven kinds (blocks, duplicates, parent-of, follow-up-of, merged-into, supersedes, relates-to) over Card, Component, File, Commit, Spec and Finding nodes, legacy relations mapped at read time with no stored row rewritten.
- REQ-TCI-013 When `moai todo relate` is asked to record a self-edge or a `blocks` or `supersedes` cycle, or `moai todo add` receives a second `--parent` or `--origin`, an unknown `--parent` or an out-of-set `--origin`, the verb shall refuse naming the kind or flag and the pair or value and leave the queue file byte-identical; a symmetric pair recorded in the opposite order shall create no second record; and `parent-of`, `follow-up-of` and `merged-into` shall be written by no relation verb.
- REQ-TCI-014 One resolver shall map a card id to its GTD item id and back, so a read shows findings with `gtd_relations`, adding no write path to `gtd_relations` and leaving `gtd organize` unchanged.
- REQ-TCI-015 When `moai todo trace <id> [--kind] [--depth]` runs, it shall print reachable nodes in deterministic order, terminate on cycles, write nothing and be available to a lane session.
- REQ-TCI-016 `moai graph build` shall derive card-to-file edges only from committed evidence — merge commits attributed to exactly one card and reachable from HEAD by any parent path, and the files they brought in — byte-identical across two builds over the same tree and reachable history, carrying no queue-private state, an absorb merge contributing no edge, with `graph check` noticing a change to that source.
- REQ-TCI-017 For every pre-change record, `list`, `list --json`, `why`, `export`, the pickup filter and the auto-rank near-duplicate path shall behave and render exactly as before.

**D — 묶음과 병합 (M4)**
- REQ-TCI-018 When cards are assigned to one bundle, the factory shall lease them to the bundle's lane one at a time in order, offering the next member after the previous local merge and withholding members from other lanes, the fleet-wide serial slot keeping its meaning.
- REQ-TCI-019 When an operator runs `moai todo merge <into> <from>`, it shall append `<from>`'s text to `<into>`, record merged-into, drop `<from>` with its reason, refuse picked, already-merged, closed-target or cyclic pairs, and be uncallable from the analyser, `analyze`, `relate` and lane sessions.
- REQ-TCI-020 While a hub file from the embedded hub-file list (produced by the baseline step) appears in the explicitly recorded `files` attribute of two open cards, the factory shall order them as a bundle chain so the second is leased only after the first has merged, the keep-set continuing to read no file overlap and no shipped code path reading a SPEC directory or reports path.

**E — 규칙과 웹 (M5, M6)**
- REQ-TCI-021 Where the M5 gate holds (a commit recording the t1453 landing merge (a merge whose subject line does not name an absorb of develop into that card's own branch) is reachable from HEAD through any parent path, read with positive controls), the card shall add a path-scoped issuance rule stating as mechanism only a size floor and ceiling, the follow-up rule, a derivation-depth cap, an in-flight cap and an issuance checklist, write each numeric value only in a local-only rule taken from the baseline record, make the rule reachable from the always-loaded dispatch rule, and stay within budget with every mirror guard green and no card id, SPEC path, report path or measured value in any template copy.
- REQ-TCI-022 Where the gate does not hold at M5 start, the card shall leave the six rule files and copies unedited, deliver a ready-to-apply draft under the tracked path `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/` with insertion anchors, and record the gate output and the follow-up card text in `progress.md`.
- REQ-TCI-023 When `/todo?view=graph` is requested, the console shall render a bounded, server-side, GET-only relation graph of live, dropped and archived cards from the unified read seam, with no write, lock, network fetch, embedded assets and all-locale strings.
- REQ-TCI-024 The existing `/todo` table, sorts, detail pane and live refresh shall render and behave exactly as before.

## 수용 기준 — 24 (출시 차단 20, 조건부 1: -021, 회귀 가드 3: -007, -009, -023)

각 출시 차단 기준은 `acceptance.md` 의 RED-now 원장 행(핀 `ad02a5677`)과 green path(명령과 통과 출력의 모양)를 가진다.

- AC-TCI-001 (REQ-001) **Given** M0 가 끝난 트리 **When** `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` 를 읽는다 **Then** `baseline.md` 가 추적되고 `figure:`·`command:`·`tree:` 줄 수가 같고 35 이상이며 git 쪽 헤드라인이 재현되고, 기준선 커밋이 자기 커밋으로 모든 제품 변경 커밋보다 앞섬을 커밋 그래프 명령 다섯이 보인다(증인 3: 두 질의의 `--full-history` SHA 목록이 같고 비어 있지 않다 — 개수가 아니다; 증인 5: `B^` 에서 닿는 이력의 카드 id 제품 커밋 0개; 제품 경로는 `internal cmd pkg scripts .claude .codex`).
- AC-TCI-002 (REQ-002) **Given** live·dropped·보관 이웃과 진행 중 레인 카드를 가진 큐 **When** `add` **Then** stdout 은 `t<id> <pos>\n` 뿐이고 stderr 에 3개 이하의 이웃 줄이 id·상태·점수·척도·앞 60자와 함께 나오며, 공유 경로가 있으면 겹침 줄이 진행 중 카드 id·레인·경로와 함께 나온다.
- AC-TCI-003 (REQ-002, -006) **Given** 출처별 고정 입력과 진행 중 레인 카드 둘 **When** 조회 **Then** 이웃은 보관·dropped 를 포함하고 구성요소·완료 SPEC 이 규칙대로 나오며, 겹침은 공유 경로마다 항목(양성), 입력이 있는 비교에서 겹침이 없으면 `none`(측정됨; 입력 없는 진행 중 카드가 섞여도 같고, 파일 이름·접두사만 같은 경로는 겹침이 아니다), 비교할 입력이 하나도 없으면 `unmeasured` 다.
- AC-TCI-004 (REQ-003) **Given** 실패하거나 느린 제시 출처 **When** `add` **Then** admit 결과가 같고 탐침은 락 밖에서 돌며 상한 초과는 `unmeasured (time bound)` 다.
- AC-TCI-005 (REQ-004) **Given** 큐 파일 F **When** `add --dry-run` **Then** F 바이트 동일, `last_seq` 불변, 종료 0.
- AC-TCI-006 (REQ-005) **Given** 이웃이 있는 큐 **When** MCP `todo_add`/`gtd engage` **Then** MCP 첫 줄 뒤에 제시, 이웃이 없으면 CLI stdout 과 같고 engage 는 소견을 기록하지 않는다.
- AC-TCI-007 (REQ-006, 회귀 가드) **Given** 핀 **When** 작업 뒤 **Then** `backlog_analysis.go` diff 가 비고 고정 시험이 통과한다.
- AC-TCI-008 (REQ-007~009) **Given** 옛 DB 와 새 속성·처분 **When** 열기·순수 읽기·보관/복원·큐 병합 **Then** 카드 표 둘과 finding 표 둘 모두 컬럼 추가, 행 불변, DDL 없는 읽기, 동결 튜플 고정, 왕복 동일, parity 포함, 속성 없는 카드와 처분 없는 소견은 SQL NULL 로 저장.
- AC-TCI-009 (REQ-007, 회귀 가드) **Given** 새 속성 없는 카드 **When** `list --json` **Then** 골든과 바이트 동일.
- AC-TCI-010 (REQ-010) **Given** drop·done 카드 **When** `drop` 과 접근자 **Then** 사유가 속성과 접두사 모두에 있고 닫힘 시각은 스탬프에서, 없으면 "unknown".
- AC-TCI-011 (REQ-011) **Given** 소견과 처분 동사 **When** 처분 기록 **Then** 소견에만 저장되고 카드·순서·필터는 바이트 동일.
- AC-TCI-012 (REQ-012, -017) **Given** 옛 소견·GTD 관계 **When** 해석기와 옛 독자 **Then** 일곱 종류 매핑, 저장 행 불변, 옛 출력 동일.
- AC-TCI-013 (REQ-013) **Given** `relate`·`add` 의 위반 입력과 `relate` 에 준 `parent-of`·`follow-up-of`·`merged-into` **When** 쓰기 동사 **Then** 거절되고(종류·플래그 명시) 파일 바이트 동일, 대칭 쌍 정규화, 보관 카드를 부모로 지목하는 것은 통과.
- AC-TCI-014 (REQ-014) **Given** engage 된 항목과 안 된 카드 **When** 해석기 **Then** 양방향 왕복, GTD 쓰기 경로 없음.
- AC-TCI-015 (REQ-015) **Given** 부모 사슬과 순환 데이터 **When** `trace` **Then** 결정적, 깊이·종류 제한, 순환에서 종료, 쓰기 없음, 레인 허용.
- AC-TCI-016 (REQ-016) **Given** 카드 귀속 병합과 develop 흡수 병합이 있는 저장소 **When** `graph build` 두 번 **Then** `card-file` 간선, 바이트 동일, 큐 상태 없음, 두 번째 부모로 닿는 카드 병합의 간선 있음, 흡수 방향 병합은 간선 없음, `graph check` 가 새 병합에 stale.
- AC-TCI-017 (REQ-018) **Given** 묶음과 레인 둘 **When** `factory next` 반복 **Then** 한 레인이 순서대로 직렬 임대, 다른 레인은 못 받고, 비묶음 동작 동일.
- AC-TCI-018 (REQ-019) **Given** queued·picked 카드 **When** `merge` **Then** 본문 절·drop·`merged-into` 기록, picked·이미 병합·닫힌 대상·순환 거절, 레인·분석 경로에서 호출 불가.
- AC-TCI-019 (REQ-020) **Given** `files` 속성에 허브 경로를 명시한 열린 카드 둘과 임베드 허브 목록 **When** 레코드 생성·선택·목록 시험 **Then** `after` 체인, keep-set 은 겹침을 읽지 않는다, 적재가 SPEC·reports 경로를 읽지 않고(정적 + 작업 디렉터리 동작 검사), 임베드 목록이 기준선 사본·측정 명령과 일치하며 CI 에서 SHA 없음은 실패다.
- AC-TCI-020 (REQ-021, -022) **Given** 게이트 다섯 판독(어느 부모 경로 술어에서 병합 제목 줄에 `absorb` 가 든 흡수 방향 병합을 뺀 착지 후보 — 본문에만 든 낱말은 빼지 않고 `T ≥ 1` 인데 `T − A = 0` 이면 목록을 읽는다, SHA 고정과 같은 줄 흡수 점검, 본문 `absorb` 착지(t1439)를 포함한 양성 대조, 두 번째 부모 부등식 대조, 형태 점검) **When** M5 종료 **Then** 게이트가 가리킨 모드 하나만 성립한다(Mode A 규칙 두 파일, Mode B 추적되는 초안과 무편집).
- AC-TCI-021 (REQ-021, 조건부 Mode A) **Given** Mode A 커밋 **When** 크기·가드·도달성 **Then** `gtd.md` ≤ 40000자, 상시 로드 무증가, 미러·생성물·카탈로그 정합, 스텁 문장과 경로 고정 시험 등록.
- AC-TCI-022 (REQ-023) **Given** 각 상태·종류의 고정 큐 **When** `GET`/`POST /todo?view=graph` **Then** 200 과 노드·간선, POST 405·큐 파일 바이트 불변, 락을 쥔 채 요청해도 2초 안에 200(`TestTodoGraphViewDoesNotWaitOnQueueLock`; 락 파일 바이트·mtime 은 락 획득의 증거가 아니다)과 웹 비시험 소스에 락 토큰 없음, 상한 문구, 외부 참조(`http:`·`https:`·`//`, `url(`·`@import`·`srcset`)·새 JS 자산·로케일 누락 없음.
- AC-TCI-023 (REQ-024, 회귀 가드) **Given** `view` 없는 `GET /todo` **When** M6 이후 **Then** 응답이 M6 시작 기준과 바이트 동일.
- AC-TCI-024 (Amendments) **Given** sync 단계 **When** 대상 SPEC 개정 커밋 **Then** `amendment_of`·`## Amendments`·`prior_completed_sha` 일치, 조건부 행은 미적용 기록.

## 수정 대상 파일

M0: `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/`(신규). M1: `internal/cli/todo.go`, `todo_issuance.go`(신규), `mcp_todo.go`, `gtd.go`, `internal/kanban/backlog_issuance.go`(신규). M2: `internal/kanban/backlog_store.go`·`backlog_sqlite.go`·`backlog_migrate.go`·`backlog_schema_freeze_test.go`·`todo_queue_merge.go`·`todo_merge_procedure.go`, `internal/cli/todo.go`·`todo_drop.go`·`todo_claim.go`·`todo_export.go`. M3: `internal/kanban/backlog_relation.go`(신규), `internal/cli/todo_relate.go`·`todo_trace.go`(신규), `internal/graph/card_file.go`(신규)·`graph.go`·`meta.go`, `internal/cli/graph.go`. M4: `internal/homestate/factory.go`·`card_record.go`·`card_picked.go`·`card_transition.go`·`hub_files.go`(신규)·`hub_files.txt`(신규), `internal/cli/factory_card.go`·`todo_merge.go`(신규). M6: `internal/web/todo_queue_read.go`·`todo_view.go`·`screens.templ`(+생성물)·`assets.go`·`assets/`. M5(Mode A): `.claude/rules/moai/workflow/card-issuance.md`(신규), `.claude/rules/local/card-issuance-thresholds.md`(신규, 로컬 전용), `gtd.md`, `kanban-dispatch.md`, `kanban-dispatch-detail.md`, `.claude/agents/moai/sync-auditor.md`, `manager-todo.md` 와 각 템플릿 사본, `internal/template/workflow_rule_paths_pinned_test.go`, `internal/template/catalog.yaml`, 생성 `.codex/agents/moai/*.toml`, 그리고 압축이 카드 id 기준선 짝의 문장을 지울 때만 `internal/template/card_id_leak_test.go`. M5(Mode B): `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/`(신규).

## 범위 밖 (Exclusions)

- t1349 의 세 항목(`show`, `add` 대시 본문 파싱, list 기본 상한) — SPEC-TODO-SURFACE-POLISH-001 이 완료했다.
- t1453 절체 시점 편집, t1450 적재 범위 재설계, t1452 병합 창 변경, t1359 `gtd.md` 문서 정정.
- 분석기·`analyze`·`relate`·Jev·레인의 자동 카드 변형(접기·삭제·수정·순서 변경); 병합 동사는 운영자 호출 전용.
- 제시 경로의 모델 호출(의미 기반 판정은 기존 Jev 소비자에 남는다).
- 기존 1,174장의 소급 입력(부모·출처·크기·예상 파일), 직렬 슬롯·keep-set·임대 경로의 묶음 외 변경, 릴리스·CI·README·docs-site.
- 배포되는 사용자 프로젝트의 허브 목록·임계값 설정 키(출하 목록은 이 저장소의 측정이고 배포 규칙은 메커니즘만 싣는다).
