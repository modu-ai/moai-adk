# t279 — t250(#1648) CR 연기 항목 트리아지 판정 표

- **Tree baseline**: `c9eed8ac6` (branch `WT-t250-followup`, origin/main 기준 — `6786c3fa4` #1648 squash 포함 확인)
- **Method**: `.moai/reports/t250/cr-round2-comments.md`의 CodeRabbit round-2 발견 69건 전수를 현재 트리와 대조. 3개 read-only 조사(verify-graph 20건 · verify-cli 32건 · verify-docs 17건)가 파일 단위 인용 근거와 함께 판정. 근거 파일: `verify-graph.md` / `verify-cli.md` / `verify-docs.md`.
- **Card premise 검증**: 카드의 "연기 31 + Minor 2 = 33 미처리" — 실측 still-valid 33건으로 **정확히 일치**. 잔여 36건 = already-fixed 33 + invalid-premise 2 + deferred-by-design 1.

## 요약

| 판정 | 건수 | 비고 |
|---|---|---|
| **채택** (t279 실행) | 29 | 테스트 폴리시 11 · 코드 폴리시 10 · SPEC 본문 3 · 문서·문구 4 · astx 쿼리 1 |
| **긴급 실행 완료** | 1 | main 레드 해소 재스탬프 — 첫 커밋 `52f7ba135` (§F5) |
| **후속** (별도 카드 후보 — 이 카드 범위 밖) | 5 | F5(스탬프 고아화 구조 대책) 포함, 리드·운영자 판단으로 |
| **기각** (전제 오류) | 2 | 반증 근거와 함께 스레드 close |
| **연기 유지** (기록된 deviation) | 1 | AC-GF-006 문서화 완료 |
| **이미 수정됨** (조치 불요, 스레드 정리만) | 33 | 수정 위치 인용과 함께 스레드 close |
| 계 | 69 (+긴급 1) | |

추가 확정: **Minor-2** — ① `mcp_code_tools_test.go:80` 무방비 `res.Content[0]` 인덱싱(에러 결과 시 패닉, 채택 #6에 흡수) ② `codequery.go` 잔여 퀵윈(스왑된 CR-ID 주석 `:153`/`:244`, 리터럴 8 `:323`, `mcp_server.go:506` "capped at 8" 과잉 주장 — 채택 #22b에 흡수).

## A. 채택 — t279 실행군 (29건)

### A-1 테스트 폴리시 (11건)

| # | comment-id | 파일 (현재 줄) | 내용 | 노트 |
|---|---|---|---|---|
| 1 | 3855001995 | internal/graph/citation_test.go (신규 TC) | hash-불일치 분기(citation.go:148-152) 테스트 부재 | RegionHash≠excerpt 내부정합 분기 |
| 2 | 3855002004 | internal/graph/codequery_test.go:42-45,146,175 | CGO 미지원 환경 skip 가드 | `!cgo`에서 t.Fatal → t.Skip |
| 3 | 3855002013 | internal/graph/codequery_test.go:255-259 | tree A 공허통과 가드(positive assertion) | ":B" 부재만 검사 → inA 긍정단언 추가 |
| 4 | 3855149281 | internal/graph/check_test.go:199,225 | itoa/padding 수제 헬퍼 → strconv | quoteJSON 반쪽은 이미 수정됨 |
| 5 | 3855001906 | internal/cli/graph_check_test.go:59-61,168,231 | provenance 픽스처 string 보간 → json.Marshal | Windows 역슬래시 경로 JSON 파괴 (실결함) |
| 6 | 3855001928 | internal/cli/mcp_code_tools_test.go:113,125,153,178,198 **+ :80** | 결과 shape 검사 헬퍼(toolText) 도입 + 무방비 인덱싱 6곳 | **Minor-2a 흡수** — :80 패닉 위험 |
| 7 | 3855001933 | internal/cli/mcp_code_tools_test.go:188 | Symbol/Via 매치 내용 단언 | 태그 정합성은 이미 맞음(대문자 wire) — 단언만 부재 |
| 8 | 3855001948 | internal/cli/mcp_code_tools_test.go (신규 TC) | 필수파라미터 거부 분기 + `..` 경로 케이스 테이블 | |
| 9 | 3855002099 | internal/hook/quality/gate_graph_freshness_test.go:113-132 | 3층 전부 fresh 픽스처 + "all layers fresh" notice 단언 | gate.go:1216 문자 무테스트 |
| 10 | 3855149237 | internal/cli/graph_refresh_test.go:152-185 | budget-overrun 테스트 결정성(주입 시임) | graph_refresh_cli.go 측 신규 주입점 |
| 11 | 3855002141 | internal/navigator/astx/*_test.go (5개 무태그) | CGO 의존 양성 테스트 `//go:build cgo` + `!cgo` 폴백 | M3 astx 그룹과 함께 실행 |

### A-2 코드 폴리시 (10건)

| # | comment-id | 파일 (현재 줄) | 내용 |
|---|---|---|---|
| 12 | 3855149289 | internal/graph/check.go:109-117 | 에러 경로에서도 layer reports 반환 (또는 문서 계약 정정) |
| 13 | 3855149309 | internal/graph/check.go:254-255 | sidecarAbsentReason → 상수 + 주석 정정 |
| 14 | 3855149315 | internal/graph/check.go:309,327 | MXIndexNeedsRefresh 임계값 caller 주입 |
| 15 | 3855149325 | internal/graph/meta.go:109-120 + check.go:348-369 | 지문 비교 규칙 단일 헬퍼 공유 |
| 16 | 3855149332 | internal/graph/symbol.go:31-34,94-97 | graph-build 오류 %w 래핑 (2곳) |
| 17 | 3855001978 | internal/cli/mcp_code_tools.go:26-36,81-87 | jsonToolResult/NewToolResultError → 패키지 toolJSON/toolErr |
| 18 | 3855149248 | internal/cli/graph_stamp.go:46-68 | CLI 경계 파일시스템 오류 경로 위생 |
| 19 | 3855149254 | internal/cli/graph.go:150 | 선택된 --edges 아티팩트 기준 갱신 판정 |
| 20 | 3855149357 | internal/mx/provenance.go:157-158 | gitOut "빈 출력 불가" 주장 주석 정정 |
| 21 | 3855001991 | internal/config/testdata/shipped_key_inventory.yaml:380-394 | graph_freshness 키 5개 R→W (reader: gate.go:167, pre_tool.go:840, graph_refresh_cli.go:53) |
| 22b | (잔여) | internal/graph/codequery.go:153,244,323 + mcp_server.go:506 | **Minor-2b** — 스왑된 CR-ID 주석 교정, 리터럴 8 상수화, "shared by MCP description" 과잉 주장 수정 |

### A-3 SPEC 본문 정정 — manager-spec 재위임 (3건 + close)

| # | comment-id | 파일 (현재 줄) | 내용 | 정정안 |
|---|---|---|---|---|
| 23 | 3855001867 | …/acceptance.md:128-132 | AC-GF-020 언어 한정 부재 | Then절에 비-Go 언어 declaration-set 예외 명시 (구현 = Go 전용 unicode.IsUpper) |
| 24 | 3855001874 | …/acceptance.md:151,191 | AC-GF-008 SHOULD vs 종결게이트 MUST 충돌 | §D.1 SHOULD행 → MUST행 승격 (mutant kill 이미 관측 — progress.md:75) |
| 25 | 3855001890 | …/spec.md:87 | REQ-GF-004 exit-2 미정의 | 제3 When절 추가: 비교불능 시스템오류 → exit 2, 무판정 보고 (구현·CI·docs와 정합) |
| — | (카드 과업 4) | …/progress.md + spec.md | §E.5(Mx) 저작 → `moai spec close` → sync_commit_sha `2fc4b40a6` backfill → status completed | pass-with-debt 2건(AC-GF-012/022)은 §E.3 기록 유지 — close 경로(`--backfill-only` 여부)는 run에서 CLI 전제검사 후 판정 |

### A-4 astx 쿼리 + 문서·문구 (5건)

| # | comment-id | 파일 (현재 줄) | 내용 |
|---|---|---|---|
| 26 | 3855001858 | .moai/project/codemaps/dependencies.md:77,116 | hook 요약 불릿에 `graph` 누락 (Mermaid edge와 불일치) |
| 27 | 3855001863 | .moai/reports/t250/m5-baseline.md:18-19 | 개발자 로컬 transcript 경로 → 저장소 상대 라벨로 치환 |
| 28 | 3855001901 | CHANGELOG.md:46 | 누적 MCP 도구 수 "21 to 24" → "25 to 28" (기계 셈 28, 직전 항목 :97이 25) |
| 29 | 3855149226 | docs-site/content/ko/cli-reference/graph.md:53 | "오래했으면" → "오래되었으면" |
| 30 | 3855002146 | internal/navigator/astx/queries/go.scm:18 | raw string import 경로 캡처 추가 (interpreted만 존재) |

## B. 후속 — 별도 카드 후보 (5건, 이 카드 범위 밖)

| # | comment-id | 내용 | 후속 사유 |
|---|---|---|---|
| F1 | 3855001854 | graph-freshness.yml actions SHA 핀 + persist-credentials | **저장소 전역 정책 결정 필요** — ci.yml:46 등 전 워크플로가 동일 @v7 관행, 단일 파일만 핀하면 관행 분기. zizmor 오류는 실재하나 소관이 이 카드가 아님 |
| F2 | 3855001962 | mcp_code_tools ctx 전파 | graph.* API에 ctx seam 부재 — 스레딩은 기능 변경 |
| F3 | 3855149371 | git 부재 시 fingerprint 앵커 | REQ-GF-003 commit-앵커 설계와 상충하는 의미 변경 — 현행 honest-absent가 설계 의도 |
| F4 | 3855002093 | symbol.Extract 스캔패키지 증거 분리 반환 | Heavy lift (반환 시그니처 변경 파급) |
| F5 | (리드 긴급 지시, 2026-08-26) | **머지 방식 × provenance 스탬프 상호작용 — squash 머지가 스탬프 커밋을 고아화** | 아래 §F5 상세 |
| F6 | 3855852818 (round-3+, 스레드 스윕에서 발견) | `resolveWithin` 부재파일≠탈출 혼동 — codequery.go:73-75·citation.go 해당 지점이 루트 내 부재 파일에 "escapes the project root" 보고 (resolveWithin :48-59의 `tErr != nil → ""` 경로) | t279 스레드 정리 중 실측 확인된 신규 still-valid Minor — SPEC 범위 동결(채택 29) 밖. 운영자·리드의 접기/후속 카드 판단 대기 |
| F7 | 3855795705 (round-3+, 스레드 스윕에서 발견) | graph_check.go:77-78·87-88 원시 cfgErr/err 텍스트 stderr 노출(경로 포함 가능) — M2 #18은 graph_stamp.go만 커버 | 동일 클래스의 미커버 지점 — F6과 동일 사유 |

## §F5 구조적 결함 기록 — squash 머지와 스탬프의 상호작용 (리드 긴급 지시 반영)

**사건**: #1648 squash 머지(`6786c3fa4`)가 t250 브랜치의 스탬프 커밋 `0d15864ae90b`(WT-graph-freshness 상의 origin/main 머지 커밋)를 고아화 — 커밋 객체가 브랜치에만 존재하고 main 히스토리에 없음. main의 graph-freshness 체크가 비교 불능 → 레드, 이후 모든 신규 PR이 상속(lane-4 #1662 실측, t274 완료 보고).

**본 워크트리 실측** (c9eed8ac6): `is-ancestor 0d15864ae HEAD` rc=1 (고아 확인). 로컬은 객체를 fetch해 보유해 codemaps가 fresh로 판정되나, CI의 체크아웃에는 객체가 없어 not-comparable(exit 2) — 환경에 따라 다른 결함 형태가 관측되는 점 자체가 스탬프의 브랜치 로컬성이 원인.

**응급 조치 (t279 첫 커밋 `52f7ba135`)**: main 도달 가능 커밋 `c9eed8ac6` 기준 재스탬프. 절차 실측: `mx scan --quiet` → `graph build` → `graph stamp codemaps` → **`graph build` 재실행**(스탬프가 provenance.json을 고쳐 edges 지문이 stale화 — 스탬프 후 빌드가 확정 순서) → `graph check` 3계층 전부 fresh, exit 0.

**t279 자체 제약 (재발 방지 — sync-phase 필수 준수)**: 이 PR의 마지막 스탬프는 `c9eed8ac6`을 지칭하는 현행 유지. **브랜치 HEAD에 재스탬프 금지** — squash 머지가 동일 고아화를 재현한다. M1-M3의 described-source 변경(~15-20파일)은 임계 40 이내로 흡수된다.

**근본 대책 제안** (후속 카드 F5 내용):
1. **CI 사전 병합 가드 (권장)**: graph-freshness 잡에 "스탬프 도달 가능성" 단계 추가 — `provenance.json`의 `commit_sha`가 PR base(origin/main)의 조상인지 검증, 아니면 레드. 고아 스탬프를 병합 전에 잡는 유일한 지점.
2. **스탬프 커밋 명시 모드**: `moai graph stamp codemaps --commit <main-ancestor>` — HEAD 무조건 명명 대신 병합 생존 커밋 지정 허용(ergonomic 보조).
3. 병합 후 main 재스탬프 루틴(매 그래프 관련 머지마다 chore PR — 무겁고 권장 안 함), merge-commit 방식 전환(저장소 관행 변경 — 기록만).

## C. 기각 — 전제 오류 (2건)

| # | comment-id | 내용 | 반증 |
|---|---|---|---|
| R1 | 3855149188 | moai-mcp-tools.md 도구 수 29/25로 수정 요구 | 문서 28/24 = catalog.go:41-70 실등록과 정확히 일치. CR이 codex군 5개를 6개로 오산 |
| R2 | 3855149192 | provenance.json 절대 tree_root 커밋 금지 요구 | check.go:153-161이 이 comment-id를 인용하며 의도 설계 문서화 — codemaps는 tracked artifact, TreeRoot 경합 금지가 설계. REQ-GF-003이 절대경로 명시 |

## D. 연기 유지 (1건)

| # | comment-id | 내용 | 근거 |
|---|---|---|---|
| D1 | 3855149345 | unconfigured(nil) skip notice 미발행 | progress.md §E.2 AC-GF-006 "Deviation recorded" — 기존 silent-pass 계약 보존, gate.go:1185-1189 근거 주석 |

## E. 이미 수정됨 (33건 — 조치 불요)

스레드 close 시 수정 위치 인용용 요약 (상세 근거는 verify-*.md):

- **internal/graph**: 3855002024 codequery.go:89-97 %w · 3855002033 codequery.go:248-257 absent 오류 · 3855002040 codequery.go:152-160 dedupe key · 3855002055 codequery.go:17-20 maxTraceDepth · 3855002059 codequery.go:195-205 edges 인덱스 · 3855002067 codequery.go:298-304 unicode.IsUpper · 3855002078 codequery.go:342-378 depth-0 구분자 · 3855002085 meta.go:69-77 EdgeCount · 3855149319 citation.go:124-141 경로 차단 · 3855149341 symbol/symbol.go:217-219 module-root local
- **internal/cli · mcp · mx · hook**: 3855001912 check.go:422-447 untracked 카운트 · 3855001919 graph_stamp_test.go:35-88 · 3855001953 mcp_code_tools.go:28,44,64 resolveToolProjectRoot · 3855001968 mcp_code_tools.go:63 GetInt 기본값 · 3855001981 mx_scan.go:108 · 3855002107 catalog_test.go:13 wantCatalogSize · 3855002115 check.go:153-172,389-416 범위 검증 · 3855002126 scanner.go:67-80 확장자 선검사 · 3855002133 scanner.go:339 · 3855149230 graph_check.go:71-79,136-146 exit 2 · 3855149264 mcp_server.go:505 WithInteger · 3855149353 provenance.go:124-126 + refresh.go:197-199 IsRegular · 3855149384 graph_stamp_test.go:35-88 커버리지 · 3855149390 refresh.go:82-85 digest 순방향 · 3855149392 refresh.go:84,97 RecordError · 3855149395 refresh.go:129 · 3855149412 refresh.go:170-176 · 3855149419 scanner_test.go:409-424
- **SPEC · 문서 · 보고서**: 3855001878 progress.md:108,214-216 + CHANGELOG:53-55 협소화 · 3855149200 m5-post.md:3-16 exact-set 재측정 · 3855149207 m5-post.md:19-21 28 도구합 · 3855149212 progress.md:224-228 정직 재진술 · 3855149215 docs-site 4로케일 MD014 해소

## PR #1648 스레드 정리 매핑 (sync-phase)

- 미해결 42개 스레드 ID: `pr1648-unresolved-threads.tsv` (comment-id → thread ID), 판정 조인: `thread-disposition.tsv`
- **스레드 스윕 재조정 (실측일 2026-08-27 — 기준 트리 `c9eed8ac6`는 2026-08-26 커밋이다. 실측은 커밋 이후 날짜에 이 트리를 대상으로 수행했으며, 실측일이 곧 리뷰 대상 커밋의 작성일은 아니다)**: 42개 중 3개(3855852818·3855852801·3855795705)는 round-2 덤프(69건) 밖의 **후속 라운드 댓글**이었다. 본문 확인 결과: 3855852801(mcp_code_tools_test.go:80-83 타입 단언)은 M1 #6의 toolText/toolTextShape 수정에 이미 흡수(병합 후 resolve) · 3855852818·3855795705는 still-valid 신규 발견 → F6·F7로 기록(즉시 resolve, 후속 안내)
- 정리 원칙: 채택분(24+1=25개) → t279 PR 머지 후 "fixed in t279 (<위치>)" resolve · 기각 2·연기 1·후속 4+F6·F7·이미 수정됨 8 = **17개 → 즉시 resolve**(반증·기록·수정 위치 인용)
- 목표: 미해결 42 → 17 (머지 후 → 0; Merge Risk High 가중 감소)

## 실행 구조 제안 (manager-spec 입력)

- **신규 SPEC 권장** (구 SPEC 증분 불가한 이유): 구 SPEC은 implemented + 감사 기록 동결 상태 — 22 AC 감사 흔적을 사후 교란하지 않고, 정정 3건+close를 포함한 본 정리 작업을 독립 전달로 명시. 구 SPEC 파일 정정은 신규 SPEC의 M4가 소관하며 version 1.1.0 → 1.2.0 승격.
- Tier M, 4 마일스톤: M1 테스트 폴리시 11건 · M2 코드 폴리시 10+Minor-2b건 · M3 astx 2건+문서·문구 4건 · M4 SPEC 본문 3건 정정 + §E.5/close/backfill
- M4는 manager-spec 재위임(구 SPEC 본문 소유권), M1-M3은 manager-develop
