# 상세 설계 5편 검토 — Core Brain · SYNC/Issues 보고 · 실행 소유권/복구 · 검증 감사/보안 · 판단 backend/이미지

- 대상(ChatGPT Space 텍스트 사본, 2026-11-08T23:19:26Z 읽음, `references/chatgpt-space/`):
  - ① `page_136c8203bf2881919eb1f543f72f87ea.md` "MoAI ADK Core Brain 상세 설계" (4,450행, sha256 747ac160…57cb)
  - ② `page_0d4538a27fb88191a19ba261bd5a3442.md` "MoAI SYNC와 Issues 보고 상세 설계" (357행, e0aa3ba6…ee82)
  - ③ `page_0259e59572808191bdb8fd302e86bd54.md` "MoAI 협업 실행 소유권과 복구 상세 설계" (204행, d5499f8b…5ae2)
  - ④ `page_c4f994be97548191b41945af0713ecc9.md` "MoAI 검증 감사와 보안 상세 설계" (228행, 92cc4b57…2da)
  - ⑤ `page_e3612b0b07888191ae58c58d66bdff4a.md` "MoAI 판단 backend와 이미지 생성 상세 설계" (209행, 0b3c4e0d…de72)
- 검토자: 리더가 위임한 읽기 전용 서브에이전트, 2026-11-08 23:3xZ. 다섯 사본 전문을 읽었다. 파일은 쓰지 않았고 이 문서 하나만 만들었다.
- 대조 기준: 상위 문서 `page_5490c4ee8c8481918b3fc3217401a3f4.md`(22절, 줄 번호는 사본 기준) · `origin/v2026.11.0` @ `9bfd26d3b19992bf1768967e4dc841b128e23bb8`(2026-11-09 08:20 KST; 선행 검토와 대조 보고는 `8ac440587` 기준이었고 그 뒤 PR #2129까지 들어왔다) · `docs/decisions-2026-11-08.md` · Go 기준선은 형제 저장소 `../moai-adk-go` 작업 트리 HEAD `4fa2aa9db2`(상위 §22가 고정한 commit과 같음)를 grep으로 확인.
- 선행 문서에서 이미 다룬 항목은 번호로만 가리킨다: 선행 검토 `review-page_5490c4ee-2026-11-08.md`의 A1~A8·B1~B8, 대조 보고 `gap-analysis-rc1-vs-design-2026-11-08.md`의 절별 행과 충돌 1~4.
- 표기: 자식 문서는 `L###`, 상위는 `§N L###`, RC1 실물은 `경로:줄`. "근거 없음"은 확인하지 못한 것, "가설"은 텍스트 검색 0건만 있는 것이다.

---

## ① MoAI ADK Core Brain 상세 설계

### 1. 문서 개요

개인 맞춤 bootstrap(프로젝트 설명 → 딥 인터뷰 → profile → 초기 commit → Issue 발행 → 맞춤 생성, L14-1066), Core Brain 메모리 구조와 노드·엣지 계약(L1068-2527), Code Map·Code Graph 재설계와 Go 소스 조사(L2529-3707), 검색·context 선택·budget(L3709-3762), 보존·아카이브·파일 구조·보안(L3763-4072), 논문·표준 대조(L4074-4264), 기존 context 실행 계약(L4270-4434), JSON/TOON 경계(L4436-4450)를 한 Page에 모았다. 여러 Page를 이어 붙인 모음집이며, 상위 §7~§12의 상세본이자 RC1 범위 밖 영역(대조 보고 7~12절)의 유일한 설계 기록이다.

### 2. 내부 결함

- **TOON 결론 충돌.** L4438 "이 적용 판단은 위임됐으며 항목마다 포맷 선택을 다시 묻지 않는다"(조건부 채택)와 L4450 "따라서, TOON은 절대 도입하지 않는다. 설계에서 제외."가 정면으로 어긋난다. L4446의 FORMAT-* 7개 후속 검증도 L4450과 양립할 수 없다. L4450은 작성 주체가 다른 덧붙임으로 읽힌다.
- **저장 구조 두 벌.** L710-733과 L3802-3834가 같은 `<state_root>/projects/<id>/`를 다르게 그린다: L729 `artifacts/` vs L3813 `artifacts/sha256/<digest>`, L731 `runs/<setup-run-id>/` vs L3825 `runs/<run-id>/`; `registry.json`(L3811)·`catalog/`(L3814)·`retention/`(L3830)·`policies/core-brain.yaml`(L3807)은 앞 구조에 없다. 어느 쪽이 정본인지 선언이 없다.
- **엣지 어휘 네 벌.** L752-810(derived_from·generated_from·targets_host·governed_by·verifies·tracks·supersedes/conflicts_with), L2413-2491(revision_of·generated_by·documents·references_requirement·observed_execution·reports_on·archived_as·included_in_context 등), L3004(declares…depends_on·changed_by), L4318(…references…). `depends_on`·`changed_by`는 L3004에만, `references`는 L4318에만 있고 통합 목록이 없다.
- **노드 목록 세 벌.** L741(ProjectProfileRevision 등 6종), L1204(Document…ContextBundle 8종), L2951-3000(Repository/Snapshot…Test/Issue). 포함 관계를 적지 않았다.
- **수락 시험 세 묶음의 ID 중첩.** L943-1065(13개), L3956-4072(12개), L4344-4424(이름 없는 8행). OFFLINE RESUME(L1051)과 OFFLINE RESTORE(L4048), CORE BRAIN TRACE(L1042)·TRACE SYNC(L4066)·"trace false link"(L4409)가 같은 시험인지 알 수 없다. 통과 조건이 "증거가 있음"(L1019), "상태가 설명 가능"(L3647)처럼 측정 명령이 없다.
- **ranking 순서 두 벌.** L3726(exact ID → 승인 계약 → 명시 관계 → 해석 심볼 → task 관련도 → 관측 evidence → lexical)과 L4302(task relevance → 직접 근거 → dependency → 실패 사례)가 다르다.
- **상태 기계 세 개가 어휘를 각자 정의.** SetupRun L854, context L4334, archive L3776-3786이 `blocked`·`stale`을 재사용하면서 교차 매핑이 없다.
- **태그 문법.** L3475 "REQ와 ISSUE 같은 새 하위행은 schema와 parser version을 정의한 뒤 추가한다"면서 L3470-3473 예시는 `@MOAI:SPEC:`·`@MOAI:REQ:` 두 하위행을 이미 쓴다. L2672 "LEGACY … 여섯 독립 종류" vs L2676 "protocol 문서의 LEGACY 하위행 설명과 맞춰야" — 불일치를 스스로 짚었으나 어느 쪽이 맞는지 판정하지 않았다(Go 실물은 4절 참조).
- **수량.** L3703 "대표 task 30~50개"는 제안 수량이라 밝혔으나 corpus 위치·구성 기준이 없다. L716 `specs/<SPEC-ID>.ref.json`과 L1104 "SPEC ID/revision 포인터"는 포인터 파일 schema가 없다.

### 3. 상위 문서와의 불일치

- **파일 구조.** 상위 §13 L1350-1356 `.moai-cli/{project.yaml, manifest.json, contracts/{spec-index.json, policy.json}, specs/, teams/}` vs 자식 L714-719 `{project.yaml, contracts.lock, specs/<ID>.ref.json, workflows/<id>/, teams/<id>.yaml, manifests/}`. 홈 쪽은 상위 L1367 `installs/<host>/<version>/`, L1378-1382 `projects/<id>/workspaces/<id>/{overlay, attempts, journal.jsonl, outbox/}`, `brain/generations/`, `objects/` vs 자식 L728-733·L3810-3834 `{artifacts, contracts, runs, indexes, workspace-state, catalog, retention}` — 겹치는 이름이 없다(A2는 상위↔RC1만 다뤘다).
- **drift 분류.** 상위 §5 L588-600 7종 vs 자식 L3266-3352 7종인데 구성이 다르다. 자식에 없는 상위 항목: "승인된 설계 변경이 SPEC·문서에 아직 반영되지 않음", "명시 link는 있으나 의미 미검토". 상위에 없는 자식 항목: "테스트 drift", "불확실". 대응표가 없다.
- **태그 구문.** 상위 §9 L1184 `@MOAI:REQ SPEC-X/REQ-007`(공백, 복합) vs 자식 L3472-3473 `@MOAI:SPEC: SPEC-PAYMENT-001` / `@MOAI:REQ: REQ-PAYMENT-003`(콜론, 분리).
- **단계 이름.** 자식 L28 흐름과 L280 "PLAN, RUN, VERIFY, AUDIT, SYNC"에 SETUP·UPDATE가 없다(A7과 같은 결함). L854 SetupRun이 사실상 SETUP인데 그 이름을 쓰지 않는다.
- **archive 설정 키.** 자식 L3768 `archive_after_days`를 이름까지 정했으나 상위 §12 L1318·§18 L1953은 기간·정책을 미결로 둔다.
- 일치(긍정): 상위 §3 "/moai 프로젝트 [설명]" = 자식 L26; 상위 L1182 Code Atlas 추천 = 자식 L3413; 상위 L1228 "7일 미만 ≠ 일치" = 자식 L2790·L3167; 상위 L1318 Go 90일 = 자식 L1098.

### 4. RC1 결정·실물과의 불일치 (`9bfd26d3b`)

- **bootstrap 흐름 L28·L359-383(초기 local commit, Issue 발행).** RC1 `skills/moai/workflows/setup.md` 절차 1~7에 commit·Issue 발행 단계가 없고 receipt도 "planned; not implemented"다(대조 보고 3행과 같음; 자식 줄 번호만 추가).
- **진입 alias L26 "프로젝트".** RC1 `skills/moai/references/routes.json`은 READ·SETUP·UPDATE·PLAN·RUN·VERIFY·AUDIT·SYNC 8개 route뿐이며 PROJECT_SETUP intent나 한글 alias가 없다(파일 전체 확인, 긍정 근거; B7 참조).
- **profile 상태 L252 "explicit/inferred/proposed".** RC1 `src/profile.rs:44-47`은 decision 단위 `source = repo | user | default`다. 필드 단위 상태가 아니고 어휘도 다르다.
- **`.moai-cli/` 하위 L714-719.** RC1 실물은 `project.yaml`(내용은 JSON), `manifests/{generation.json, github-issues.json, legacy-asset-map.json, baseline/…}`, `receipts/`(`src/apply.rs:22-24`), `forge/{outbox/, poll-state.json, promotions.json}`(`src/forge/outbox.rs:7`, `src/forge/mod.rs:6`), `decisions/<run>.jsonl`(`src/main.rs:474`), `evidence/<run>/`(`src/forge/report.rs:10`), `worktrees/`(`src/worktree.rs:12`), `workflows/<id>/`(`src/generation.rs:11-12`, 일치). `contracts.lock`·`specs/*.ref.json`·`teams/`·`policies/`는 없다.
- **홈 쪽 L728-733.** RC1은 `~/.moai-cli/installations/<id>.json`과 `…/hosts/<host>/known-good/<sha256>`(`src/install.rs:4`, `src/recovery.rs:11`), `~/.moai-cli/bin/moai-cli`(decisions 15:5x 행)뿐이다. `projects/<id>/` 계층은 P09 HOLD(`src/apply.rs:25` swap point).
- **Claude/Codex 어댑터 표 L543·L547·L556·L569·L573 — 일치(긍정).** `src/claude.rs:6`(v2.1.277), `src/hosts/codex.rs:11`(32 KiB), `src/update.rs:3-4`(`~/.claude/skills/moai`, `~/.agents/skills/moai`), `src/generation.rs:12-13`(`.claude/agents/<role>.md`, `.codex/agents/<role>.toml`).
- **3-way·보존 L254·L510-516 — 일치(긍정).** `src/agents.rs:1-6`, `src/apply.rs:42-43`, `src/generation.rs:1-2`.
- **AGENTS.local.md L712 "ignored".** RC1 `src/agents.rs:30-31`은 none/ignored/tracked 중 선택이고, RC1 저장소 자체는 tracked다(`git ls-tree`에 존재). "ignored"는 기본이 아니라 선택지다.
- **Go 기준선 수치 — 확인(긍정).** L1098·L2276 90일: `internal/cli/spec_archive.go:41,102`. L2790 7일: `internal/cli/mx_index.go:25`. L2852 863 태그·~28.8k·~7 MB·~16.4k: `internal/navigator/sync/mx_bridge.go:54-63`. L2672 LEGACY 독립 kind: `internal/mx/tag.go:25`, `internal/mx/scanner.go:36`.
- **Core Brain·Map·Graph·archive·tombstone·FTS.** RC1 Issue 없음 — `docs/rust-product-path.md:360` "F13 graph/MX left to a later Issue"(긍정 근거). context는 `src/context/mod.rs:1-7` snapshot·bundle·freshness만. L3730 "retrieval budget 계산" vs RC1 budget 필수 입력·기본값 없음(`rust-product-path.md:355`); L4282 `max_tokens: 6000` 예시와 달리 RC1은 `max_bytes`/`max_items`이고 토큰은 `bytes_div_4` 추정(`context/mod.rs:7`).
- **L4442 MCP structuredContent.** RC1은 MCP를 다음 버전으로 보냈다(decisions 18:2x 행 HOLD-C1).
- **TOON.** RC1 src에 0건(가설). 대조 보고 13행 "TOON 미채택은 일치"는 L4450과 방향이 같고 L4438과는 다르다.

### 5. 낡은 내용

- L1066·L2247·L4072·L4268 "이번 작업은 … 실제 설치·scaffold·commit·Issue 발행 … 수행하지 않았다"는 설계 시점 문장이다. RC1은 Issue #1975~#2129 발행, 70여 PR 통합, `~/.moai-cli/bin/moai-cli` 로컬 빌드(decisions 20:5x 행)까지 진행했고, 문서는 RC1 Rust 실물을 한 줄도 인용하지 않는다(A8과 동일).
- L2633 열 제목 "현재 구현에서 확인한 내용"의 "현재 구현"은 Go만 가리킨다. RC1 Rust가 별도로 진행 중이므로 그 사실을 적어야 한다.
- L595 "teams는 experimental이며 기본 비활성" 등 host 서술은 2026-11-08 확인이라 밝혀 양호하나 재확인 조건이 없다.
- L4444 TOON spec 4.4·Rust 구현 0.7.0 — 버전 고정은 양호하나 L4450으로 절 전체가 폐기 대상이다.

### 6. 개선 제안 (영향순)

1. L4450을 본문과 일치시킨다 — 제외가 결정이면 L4436-4448을 "제외 결정(출처·일시)"로 바꾸고 FORMAT-* 시험을 지우며 상위 §13 L1395도 함께 고친다.
2. 저장 구조를 한 벌로 통합하고 "RC1 출하 경로"(`manifests/`, `receipts/`, `forge/`, `decisions/`, `worktrees/`, `installations/`)와 "후속 목표"를 두 열로 나눈다(A2와 함께).
3. 노드·엣지·상태 어휘를 한 표에 모으고 각 절은 그 표를 참조만 한다.
4. drift 분류를 상위 §5 7종과 1:1 대응표로 만들고 RC1 3종(omission/amendment/document mismatch)과의 포함 관계를 적는다.
5. 수락 시험 ID를 하나의 체계로 통일하고 통과 조건을 명령·fixture 단위로 적는다.
6. bootstrap 흐름에 RC1 setup.md와의 차이(commit·Issue 발행·receipt 미구현)를 적는다.
7. 태그 구문 예시를 상위 §9와 한 가지로 맞춘다.
8. 중복 구간(L1239-2247)을 원본 Page에서 걷어내거나 사본에 중복 표식을 단다.

### 7. 미확인·한계

- 사본 L14-1228이 L1239-2247에 다시 나온다(공백 제외 158행만 다르고, 차이는 도식의 평문 노드 목록 vs Mermaid 원문). L1229-1233·L2280의 "페이지 불러오는 중…"은 링크 제목 자리다. 이 검토는 첫 출현 줄만 인용했다.
- 근거 목록(L115-119, L627, L701, L2743, L2848, L2856, L3120, L4426-4434, L4448)은 링크 제목만 남아 URL·버전을 확인할 수 없다.
- L2852의 34개/65개 SPEC 수치는 재확인하지 않았다(`mx_bridge.go:54-63`에서 863·28.8k·7 MB·16.4k만 확인).
- Claude teams 파일 경로(L608), Codex `agents.enabled`(L599) 등 host 내부 서술과 논문·표준 표(L4095-4260)의 버전·날짜는 외부에서 재확인하지 않았다.
- Go 확인은 작업 트리 grep이며 `git show`가 아니다.

---

## ② MoAI SYNC와 Issues 보고 상세 설계

### 1. 문서 개요

F17 통합·F18 SYNC·F19 보고의 계약이다. 입력/출력과 receipt JSON(L16-34), 통합·동기화 6단계(L40-50), 상태 기계와 재시도(L71-75), Issue·artifact 책임표(L81-156), polling/webhook(L162-164), 내부 인터페이스(L170-172), 수락 시험 8행(L178-254), TODO backend 선택과 승격 절차(L274-342)를 담고, 끝에 다른 Page 요약과 중복 문단(L344-357)이 붙어 있다. 다섯 문서 중 RC1 실물과 가장 가깝다.

### 2. 내부 결함

- **효과 상태 어휘 세 벌.** L18 "proposed, applied, not_required, blocked, unknown" / L29 예시 `"status": "proposed"` / L71 "ready → reconciling → document_proposed → authorized → publishing → confirmed → closed; needs_amendment, conflict, blocked, publish_unknown". `applied`↔`confirmed`, `unknown`↔`publish_unknown` 대응이 없다.
- **`closed_as` L156 "completed, cancelled, superseded".** GitHub `state_reason` 어휘와의 대응이 없고 "superseded"가 MoAI 어휘인지 provider 값인지 알 수 없다.
- **L162 "항상 켜진 서비스가 없다는 점을 UI에 표시".** UI 표면이 정의돼 있지 않다(RC1은 CLI와 skill뿐).
- **L170 "실제 Rust argv와 MCP tool 이름은 adapter schema로 생성".** 생성 규칙이 없고 MCP는 RC1 밖이다.
- **수락 시험 체계 불일치.** L178-254 8행은 ID가 없고 L284-342는 TASK-* ID를 쓴다. "수정 보존"(L191), "식별 유지"(L245)는 측정 명령이 없다.
- **중복·삽입.** L349-353이 L18 후반을 그대로 반복하고, L344-347은 다른 Page("MoAI RC1 GitHub 작업 구조")의 요약이 본문처럼 들어와 있다.
- **webhook 전제.** L164의 서명 검증·dedup·역순 처리는 상시 수신자를 전제하는데 L10 "자체 MoAI 서버 요구 안 함"과의 양립 조건이 "기존 수신 환경을 사용자가 선택한 경우" 한 구절뿐이다.

### 3. 상위 문서와의 불일치

- **drift 3분류 L44**(구현 누락→RUN, 요구 변경→amendment, 문서 누락→문서 제안) vs 상위 §5 L588-600 7분류. 상위의 "oracle 충돌", "참조 단절", "Issue 완료인데 증거 stale", "link 의미 미검토"가 자식의 어느 갈래로 가는지 없다.
- **outbox 위치 L150 "home의 작은 outbox"** vs 상위 §13 L1378-1382 `projects/<id>/workspaces/<workspace-id>/outbox/`. 둘 다 홈이지만 자식은 경로를 정하지 않는다.
- **tasks.md L274 `<state_root>/projects/{project_id}/tasks.md`** — 상위 §13 구조에 tasks.md가 없다. 상위 L1385 "Issue 본문을 두 번째 TODO 원장으로 복제하지 않는다"와 자식 L278-282 "로컬에는 stable identity·최소 복구 정보만"은 일치(긍정).
- 일치(긍정): 상위 §5 통합 행 "최신 base로 만든 exact candidate 재검증, readback" = 자식 L42·L50; 상위 §1 L100 승인 분리 = 자식 L8·L48.

### 4. RC1 결정·실물과의 불일치

- **상태 기계 L71 = RC1 `skills/moai/workflows/sync.md:13`과 같다.** RC1은 `needs_rebase`를 더 가진다. 자식이 한 상태 뒤처진다.
- **drift 3분류 L44 = `sync.md:20`(omission / amendment / document mismatch) 일치(긍정).** 상위 7종과는 둘 다 다르다.
- **집합 필드 PATCH 금지 L75 = `src/forge/write.rs:2-6` 일치(긍정).** ETag≠CAS도 write.rs:5 "read-before-write rather than If-Match"로 구현.
- **timeout→marker 재조회 L73 = `src/forge/outbox.rs:3-13`** (`publish_unknown`은 reconcile만이 `confirmed`로 옮김). 시험 `tests/forge_report_cli.rs:322 timeout_then_reconcile_creates_no_duplicate`, `tests/handoff_cli.rs:523 reconcile_uncertain_no_duplicate`(긍정).
- **요약 댓글 marker L111 = `src/forge/report.rs:3-5`** `<!-- moai:report run=<id> -->`. 사람 편집 보존 L75·L191 = report.rs:19-20, 시험 `human_edited_comment_is_not_overwritten`(긍정).
- **outbox 위치 L150 "home"** vs RC1 `<project>/.moai-cli/forge/outbox/<op-id>.json`(outbox.rs:7) — 프로젝트 쪽(대조 보고 충돌 2).
- **tasks.md L274 기본 경로** vs RC1 `src/forge/promote.rs:4` "path given by the caller; RC1 interim decision: no default location"(decisions 19:2x 행). 매핑은 `<project>/.moai-cli/forge/promotions.json`(`docs/p07-decision.md`).
- **인터페이스 L170.** `forge.read`·`forge.report`·`effects.reconcile`은 RC1 operation 표(`src/main.rs:623-635`)에 있고, `sync.preview`·`sync.apply`·`artifact.pin`은 없다 — SYNC는 "절차 + git/gh"(sync.md:7), artifact 업로드는 미구현(report.rs:25 "P09"). 표는 전수라 긍정 근거다.
- **승격 절차 L280 = promote.rs:3-19 일치(긍정)**: preview/apply 분리, digest 불일치 `stale`, `uncertain`은 재생성 금지, 선택 항목만 전송. "제목 유사도만으로 자동 병합 안 함" — RC1 Jaccard≥0.8은 후보 표시만(p07-decision).
- **RC1 통합 게이트 부재.** ruleset `RC1 integration gate`, 필수 status `moai/audit`, verdict 첫 줄 형식, 리더 auto-merge(decisions 17:3x 두 행, `sync.md` 8단계)가 자식에 없다. L48 "merge 각각 policy가 확인"과 에이전틱 자율 병합의 긴장은 대조 보고 충돌 1.
- **polling L162 = `src/forge/mod.rs:3-6`**(조건부 GET, `since`, Link, poll-state.json)과 `src/forge/policy.rs`(backoff, rate-limit) 일치. "jitter"는 근거 없음.
- **webhook·GitLab L164·L245.** GitLab은 `docs/rust-product-path.md:238` "nothing is implemented"(긍정); webhook은 src 0건(가설).
- **라이브 쓰기.** `docs/p07-decision.md` 마지막 "Live GitHub writes were not run" — 수락 "게시 성공 직후 연결 단절"은 녹화 transport로만 검증됐다.

### 5. 낡은 내용

- L347 "실제 GitHub 발행은 계획 검토 뒤 후속 지시 전까지 하지 않는다" — Issue #1975~#2129는 발행됐다(`.moai-cli/manifests/github-issues.json`, decisions 전반).
- L172 "worktree별 대형 반복 보고서 대신 Issue 요약과 별도 artifact로 이행" — RC1은 Issue 소유 댓글 1개(report.rs)에 더해 `.moai-cli/reports/issue-bodies/*.md` 사본 거울을 쓴다(decisions 17:3x "07-A/07-B 구현 때까지"). 거울 정책이 반영돼 있지 않다.
- L164 "GitHub 실패 webhook의 자동 재전송을 전제하지 않는다" — 외부 사실인데 확인 날짜가 없다. 근거 L260-268은 제목만 있다.

### 6. 개선 제안 (영향순)

1. 효과 상태 어휘를 L71 한 벌로 고정하고 L18·L29를 거기에 맞추며 RC1 `needs_rebase`를 추가한다.
2. L48 통합 행에 RC1 확정 게이트(필수 status + 독립 verdict 첫 줄 + head==audited_sha)를 적는다(A5와 함께).
3. outbox·tasks.md·promotions.json 경로를 RC1 실물(`<project>/.moai-cli/forge/`)로 고치거나 "P09 뒤 홈 이동"으로 표시한다.
4. `sync.preview/apply`·`artifact.pin`을 "미구현(절차로 대체)"로 표시한다.
5. drift 3분류와 상위 7분류의 대응표를 둔다.
6. 수락 시험 1표에 ID를 붙이고 RC1 시험명(`forge_report_cli`, `todo_promotion_cli`, `tests/fixtures/forge/labels_conflict.json`)을 연결하며 GitLab 행은 "RC1 미인증"으로 둔다(B1).
7. L344-357 삽입·중복 문단을 제거한다.
8. `closed_as`를 GitHub `state_reason`과 대응시킨다.

### 7. 미확인·한계

- 근거 5건의 URL·버전과 GitHub `state_reason` 어휘는 외부에서 재확인하지 않았다.
- jitter·webhook 부재는 src grep 0건에 기댄 가설이다.
- L344-347의 출처 Page는 읽지 않았다. `tests/todo_promotion_cli.rs`는 파일명만 확인했다.

---

## ③ MoAI 협업 실행 소유권과 복구 상세 설계

### 1. 문서 개요

F08 실행 경로·F09 dispatch·F10 복구·F11 native 협업이다. task/attempt JSON(L16-29), 실행 선택과 native 팀(L37-41), 상태 14개(L46-59), journal·중복 방지·fencing(L67-73), 독립 감사(L79-81), 복구 사례 7건(L96-154), 인터페이스 6개(L160), 수락 8항(L168-182), 세션 TODO와 소유권 분리(L200-202)를 담는다. 원칙은 RC1 claim·handoff 구현과 같은 방향이나, RC1이 택한 claim 수단은 적혀 있지 않다.

### 2. 내부 결함

- **상태 14개(L46-59)가 목록뿐이다.** 전이 규칙·종결 상태·"active" 집합이 없고 `accepted`와 `closed`, `unknown`과 `reconciling`의 관계가 없다.
- **L68 "provider가 실제 제공하는 coordinator·lease/fencing 계약을 이용"** — 무엇이 "실제 제공"인지 판정 기준이 없다. GitHub에 없다는 결론이면 다중 host 자동 takeover는 항상 비활성(L180)인데 그 귀결을 적지 않았다.
- **JSON 공백.** L21 `owner.kind` 값 목록(`local-process` 외), L22 `external_task` 구조, L27 `effects` 항목 schema가 없다.
- **L168 "두 merge owner를 거부"** — merge owner를 담는 필드가 JSON에 없다(상위의 `integration_owner`도 없다). L170 "fault injection"은 환경·도구가 없다.
- **L182 "모든 transport에서 동일하게 적용"** — L160의 transport 4종 각각에 대한 시험이 없다.
- **L162 "운영 큐·진행 작업은 자동 이동·삭제하지 않는다"** — Go 격리 원칙의 재진술이며 측정 기준이 없다.

### 3. 상위 문서와의 불일치

- **JSON 필드.** 상위 §6 L764-781 `{task_id, forge_ref{…}, attempt_id, owner_id, epoch, workspace_id, base_revision, spec_digest, policy_digest, write_scope, state, last_observed_at}` vs 자식 L16-29 `{schema_version, task_id, attempt, epoch, owner{kind,id}, external_task, base_revision, spec_digest, write_scope, state, effects, last_observed_at}`. `attempt`↔`attempt_id`, `owner`↔`owner_id`; 자식에 `forge_ref`·`workspace_id`·`policy_digest`가 없고 상위에 `external_task`·`effects`가 없다.
- **4필드 분리.** 상위 L756 assignee/execution_owner/provider_session/integration_owner — 자식 L10은 개념만 말하고 필드명을 쓰지 않는다.
- **상태.** 상위 §6에는 상태 목록이 없어 자식 L46-59가 유일한 정의인데, 상위 §5·§15의 `not_run`·`stale`과 연결이 없다.
- 일치(긍정): 상위 L758 "epoch는 fencing 자체가 아니다" = 자식 L71; 상위 L762 "timeout·heartbeat 만료만으로 사망 단정 금지" = 자식 L105-109·L180; 상위 L1836 수락(동시 writer 차단, 무응답 takeover 0, stale epoch 차단) = 자식 L168·L172·L180.

### 4. RC1 결정·실물과의 불일치

- **L69 "Issue comment, label, assignee와 ETag GET는 분산 lock 대체물이 아니다"** vs RC1 P08 결정 ⑤ — Issue 댓글 id fencing이 cross-host claim의 1차 수단(decisions 16:4x 행, `src/claim.rs:1-3,15-21`). 사용자가 택한 결정과의 긴장이며(대조 보고 충돌 3이 상위 L758에 대해 지적한 것과 같은 축) 자식 문장이 더 단정적이다. 합치점: 자동 takeover 금지(claim.rs:22-23, `src/handoff.rs:5-7`).
- **상태 어휘.** 자식 L46-59 vs RC1 `src/claim.rs:50-58` ACTIVE_STATES `assigned, running, verifying, pr_ready, in_review, blocked, handoff_pending, uncertain` + `unclaimed, integrated, cancelled`(claim.rs:48-49), `handed_off`(handoff.rs:43). 자식에 없는 것: pr_ready, in_review, handoff_pending, uncertain, integrated, handed_off. RC1에 없는 것: prepared, authorized, submitted, result_ready, cancel_requested, reconciling, accepted, needs_fix, closed.
- **journal L67 = claim.rs:5-13 일치(긍정)**: `journal.log` append+fsync, `<task>.json` temp+rename, `.lock` advisory. 위치는 RC1 `<git common dir>/.moai-cli/claims/`(claim.rs:6); 자식은 위치 미지정, 상위 §13은 홈.
- **owner.kind L21 `local-process` = handoff.rs:9-10 일치(긍정)**(같은 host의 local-process만 liveness probe).
- **외부 쓰기 전 intent L73 = handoff.rs:16-19** `effect_intend/effect_confirm`, 하위 epoch intent는 `stale_epoch` 거부 — 일치(긍정). "duplicate report·upload·PR 생성 시험" 중 report·PR은 있고(`forge_report_cli`, `pr_link`) upload는 미구현(report.rs:25).
- **fault injection L170.** RC1 시험 `claim_uncertain_when_reread_fails`, `claim_loser_stops_after_post_reread`, `handoff_stale_epoch_fenced`, `reconcile_uncertain_no_duplicate`(tests/claim_cli.rs, handoff_cli.rs), `timeout_bounds_runtime_even_with_grandchildren`(tests/run_verify.rs:549) — 녹화 transport·stub gh 기반이고 crash 주입은 없다(부분).
- **cancel/resume L61·L176** vs RC1 `src/verify.rs:52` "No cancellation path exists yet; recorded so a later one has a place". "unsupported를 명확히 표시"는 필드 예약 수준이다.
- **worktree 삭제 L132-136** vs RC1 `src/apply.rs:25-26` "Worktree deletion loses the receipt with the tree — a documented known gap". 복구 절차가 없다.
- **인터페이스 L160** `execution.prepare, dispatch.submit, dispatch.observe, owner.handoff, effects.reconcile, result.admit` vs RC1 operation 표(`src/main.rs:614-660`): `claim.acquire/show/release/handoff`, `effect.intend/confirm`, `effects.reconcile`(일치), `audit.reduce`, `run.verify`, `worktree.plan/apply`, `scope.check`. `dispatch.*`·`execution.prepare`·`result.admit`은 없다(표 전수). `owner.handoff`↔`claim.handoff` 이름이 다르다.
- **독립 감사 L79-81 = `src/audit.rs:1-9, 193-196` 일치(긍정)**(독립성 표, reviewer 편집 시 `candidate-changed`). 다만 RC1 독립성은 자가 신고다(`docs/rust-product-path.md:382`).
- **L8 "고정 전문 agent 집합을 만들지 않는다".** 제품은 일치(`src/generation.rs` 선택 생성). 개발 운영은 agent-1~8 고정 레인과 agent-2 고정 감사자(`AGENTS.local.md:24`)다. 제품/개발 구분이 필요하다.
- **write_scope L25·L41 = `src/scope.rs:1-8,15` 일치(긍정)**(겹침→`blocked(scope_conflict)`, 통합 순서 지정 시 허용, lockfile 등 기본 공유 경로).

### 5. 낡은 내용

- L162 "버전 전환은 read-only mapping, export, opt-in migration으로 시작" — 04-D 마이그레이션은 RC1 밖으로 이동(decisions 00:2x(11-09) 행 ②). 문장은 유효하나 결정 링크가 없다.
- 근거 L188-194(Claude Code on the web, Agent SDK, Codex developer commands, GitHub REST) 제목만 있고 날짜가 없다.
- L203 "MoAI SYNC와 Issues 보고 상세 설계"는 다음 Page 링크 제목이 본문 끝에 붙은 것이다(사본 artifact).

### 6. 개선 제안 (영향순)

1. claim 댓글 프로토콜(`moai-claim: session= epoch= lease_until=` 댓글, 최소 live 댓글 id 승, `Moai-Claim: <issue>#<id>` 토큰, 만료 시 운영자 판단)을 L67-71에 적고, L69를 "GitHub에 원자 primitive가 없으므로 댓글 id 단조성을 가정한 eventual claim + 자동 takeover 금지"로 고친다(A4와 함께).
2. 상태 14개를 RC1 claim 상태와 대응표로 만들고 전이·종결·active 집합을 정의한다.
3. JSON 필드명을 상위 §6과 통일한다(attempt_id, owner_id, workspace_id, policy_digest, forge_ref — 또는 상위를 고친다).
4. 인터페이스 이름을 RC1 operation(`claim.*`, `effect.*`, `effects.reconcile`, `audit.reduce`)에 맞추거나 대응표를 둔다.
5. worktree 삭제·cancel 행에 RC1 known gap(receipt 소실, cancel 경로 없음)을 적는다.
6. 제품 원칙(고정 agent 없음)과 개발 운영(고정 레인)을 구분하는 한 줄을 넣는다.
7. L170 fault injection을 RC1 시험명으로 구체화하고 crash 주입은 미시험으로 표시한다.

### 7. 미확인·한계

- GitHub에 원자 claim primitive가 없다는 전제는 자식·상위·RC1이 공유하나 외부 문서로 재확인하지 않았다.
- `tests/worktree_cli.rs`, `tests/claim_record_cli.rs`는 파일명만 확인했다.
- Claude Code on the web·Agent SDK remote task API의 cancel/resume 지원 여부는 확인하지 않았다.

---

## ④ MoAI 검증 감사와 보안 상세 설계

### 1. 문서 개요

F15 검증·F16 감사와 실행 전 권한 경계다. 입출력 계약과 receipt JSON(L16-29), preflight·보안 gate(L35-37), receipt admission 6규칙(L59-71), 비밀·공급망(L77-81), 실패·복구 7행(L100-186), 인터페이스 5개(L192), 기존 자산 재사용(L194), 수락 8항(L200-214), 근거 5건(L220-228). admission 논리는 RC1 `audit.rs`와 거의 1:1로 맞는 반면, preflight와 hook 전제는 RC1에 없다.

### 2. 내부 결함

- **L117 "check timeout → unknown 또는 failed 정책별 구분"** — 정책 기준이 없다. L65와 합쳐도 timeout이 어느 상태로 가는지 정해지지 않는다.
- **어휘 불일치.** L29 출력 어휘(passed, failed, not_run, blocked, stale, unknown)에 표의 `rejected`(L143)·`review missing`(L182)이 없다. L25 `"tool_version": "unverified"`는 상태값인지 자리표시자인지 알 수 없다.
- **L10 "높은 위험의 release gate"** — 위험 등급 기준이 없다(상위 §5 "위험에 맞게"와 같은 공백).
- **L79 "필요 권한, network destinations를 inventory"** — 정적 열람으로 어떻게 얻는지 없다.
- **예외 필드 두 벌.** L81 "severity·rule ID·근거·예외 소유자·유효기간" vs L210 "ID·범위·기간".
- **L194 "기존 wrapper의 binary 부재 exit0"** — 어느 wrapper(Go 훅)인지 명시가 없다.
- **L206 "이미지 metadata·download bundle로 새지 않는다"** — F21·F19 범위 기능을 전제한 수락 조건이다.

### 3. 상위 문서와의 불일치

- 일치(긍정): 출력 어휘 L29 = 상위 §5 VERIFY 행; L10 독립 AUDIT = 상위 §5; L59·L63 "자기 보고·exit 0로 완료 안 됨" = 상위 §15.
- **상위 §15 L1647 PASS-WITH-DEBT·debt ID** — 자식에 없다(waiver만 L210).
- **상위 §15 retry 행**(repair episode, failure fingerprint, 단일 retry owner, 비용 예산) — 자식 L126-134 "scope 안 수정, 재실행, 새 receipt"에 예산·fingerprint가 없다.
- **상위 §15 receipt 행** "생성→durable→전송→ACK→identity/digest 확인→gate 채택 분리" — 자식 L59-71은 admission만 다루고 전송·ACK 단계가 없다(SYNC 문서로 위임이면 명시 필요).
- **상위 §15 browser QA 행**(DOM/AX vs visual) — 자식에 없다.
- 인터페이스 L192 5개는 상위에 대응 절이 없다(상위는 단계명만).

### 4. RC1 결정·실물과의 불일치

- **admission 6규칙 L59-69 = `src/audit.rs:298`(REQ-05D-01 identity·attempt·timestamps), `:308`(REQ-05D-02 계약 대조→stale), `:481-512`(independence-missing→blocked, 의견≠증거→held) — 강한 일치(긍정).** 시험 `tests/audit_admission.rs:150 rejects_synthetic_and_foreign_evidence`, `:192 must_pass_matrix`, `:262 independence_recorded_and_enforced`, `:313 evidence_over_opinion`, `:360 waiver_preserves_original`, `:433 pr_head_binding`. L63 "exit 0 하나로 충족 주장 금지" = `tests/run_verify.rs:368 self_report_not_promoted`, audit.rs:5-6.
- **출력 어휘 L29 = `skills/moai/workflows/verify.md:5` 동일(긍정).**
- **preflight L35**(actor·target·network·비용·persistent access·destructive) — RC1에 없다: operation 표(`src/main.rs:614-660`)에 `policy.*`가 없고(전수), `permission_reference`는 "presence and shape only; caller authorization remains host-owned"(`docs/p07-decision.md` 끝, `skills/moai/SKILL.md`). `policy.authorize`는 RC1에서 host에 위임됐다.
- **보안 hook L37·L204.** RC1 제품은 hook 0개(`docs/codex-adapter-contract.md:10`, decisions 19:1x 02-B·02-C 행). 개발 중 가드는 Go 훅(decisions 20:3x 훅 범위 행). 수락 L204는 RC1에서 측정할 수 없다.
- **timeout L117** vs RC1 `src/verify.rs:18-23` FailureClass — timeout은 environment-failure(RED로 세지 않음), `tests/run_verify.rs:549`. RC1은 unknown이 아니라 환경 실패로 분류한다.
- **waiver L81·L210 = audit.rs:63-70,167**(`check:<index>`, quality만, security/secret/permission 불가, `expires_at`) — 대체로 일치. "rule ID·근거·예외 소유자"는 RC1에 없다(부분).
- **독립성 L10·L212.** RC1 `audit.rs:193-196`: Claude fresh subagent·separate-session = 독립, fork·Codex = 제한 기록, same-session = 불가(자가 신고, `rust-product-path.md:382`). 자식 L212 "동일 모델 재질문 ≠ 완전한 독립 보증"과 RC1 "fresh 같은 모델 = 독립"은 어휘 긴장이다(RC1은 "conservative until measured"로 표시).
- **비밀 L77 — 일치(긍정).** 토큰은 gh keyring/`GH_TOKEN`, 프로그램이 읽지 않음(`src/forge/transport.rs:3-4`); `src/redact.rs:8` "conservative shape filter, not a secret detector"는 L79 "scanner 결과만으로 안전 보증 안 함"과 같은 방향.
- **공급망 L79 = `src/sources.rs:4-14`**(license evidence, 고정 revision digest, `host_loaded:false`/`execution_verified:false` 분리) — 일치. "필요 권한·network destinations"는 기록하지 않는다(부분).
- **L194 "현행 MoAI의 … 코드를 재사용 후보 … Rust contract로 옮긴다"** vs RC1 원칙 "Go는 참고 자료, 기본 이식 아님"(`AGENTS.md` Standing principles 1항). "후보"라 충돌은 아니나 RC1은 연구 우선을 택했다.
- **RC1 verdict 첫 줄 형식**(`AUDIT #<n> (<key>) head <40hex> base <40hex> — verdict: PASS|FAIL`, decisions 17:3x 보강 행)과 `moai/audit` status — 자식에 없다.
- **verdict 저장.** 자식 L29 "증거 manifest" 출력 vs RC1 "the verdict is not stored anywhere (P09)"(`rust-product-path.md:386`, `audit.md:3`).

### 5. 낡은 내용

- 근거 L220-228(NIST SSDF, OWASP Agentic Top 10, Claude hooks, Codex hooks, AgentSpec)에 버전·확인 날짜가 없어 어느 판본을 가리키는지 알 수 없다.
- RC1이 2026-11-08에 확정한 게이트(ruleset, status, verdict 형식)와 독립성 표(H1)가 반영돼 있지 않다.
- L194의 Go 코드 재사용 전제는 RC1 결정 뒤 재검토 대상이다.

### 6. 개선 제안 (영향순)

1. preflight(`policy.authorize`)를 "RC1: host 권한에 위임, `permission_reference`는 형태 검사"로 표기하고 자체 구현은 후속으로 둔다.
2. RC1 verdict 첫 줄 형식·`moai/audit` status·독립성 표(fresh/fork/separate-session/same-session)를 L10·L59 뒤에 적는다.
3. timeout 판정을 RC1 FailureClass(environment-failure/test-failure/unclassified)와 맞추거나 차이를 명시한다.
4. 판정 어휘(L29)와 표 어휘(rejected, review missing)를 한 벌로 만든다.
5. 예외 필드(L81 vs L210)를 통일하고 RC1 waiver 필드와 대응시킨다.
6. hook 전제 수락 항목(L204)에 "RC1 hook 0개 → 측정 불가"를 표시한다.
7. 근거에 버전·확인 날짜를 붙인다.

### 7. 미확인·한계

- NIST SSDF·OWASP Agentic·AgentSpec·Claude/Codex hooks 문서는 재확인하지 않았다.
- `tests/audit_admission.rs`는 함수명만 확인했고 실행하지 않았다.
- `src/lib.rs`의 `SECRET_SHAPES` 목록은 읽지 않았다.

---

## ⑤ MoAI 판단 backend와 이미지 생성 상세 설계

### 1. 문서 개요

F04 intent 가속·F20 판단·F21 이미지다. 판단 경로 4종 표(L29-76), Jev 기준선 사실(L78-80), OpenAI Decisions(L82), decision JSON(L87-98), 판단 흐름 6단계(L106-116), 상태·fallback(L139-141), 구독 크레딧·비용(L147-151), 이미지 workflow(L157-159), 인터페이스·비교 시험(L165-167), 수락 8항(L173-187), 근거 9건. "native 기본, backend는 선택"이라는 틀은 RC1 P10 결정과 같다.

### 2. 내부 결함

- **vendor 정책을 날짜 없이 사실로 서술.** L147 "양사의 구독 연계 API credit은 실제 공식 정책이 있으나"와 L149의 구체 규칙(API-key Codex 제외, `claude -p`·Agent SDK 예외, annual/monthly, auto-reload)에 확인 일시·URL·적용 상품이 없다.
- **"OpenAI Decisions"(L55-63·L82)** — endpoint·문서 버전 등 제품 식별이 없다. L82 "모델 이름을 영구 상수로 고정하지 않는다"면서 제품명은 고정한다.
- **L96 `"max_cost_usd": 0`** — "무료만"인지 "미설정"인지 정의가 없다.
- **L139 상태 7개 + 독립 결과 6개** — `policy_checked` 뒤 종결 상태가 없다.
- **L165 `verify.accept`** — 이 문서에 정의가 없고 검증 문서는 `verify.execute`/`evidence.admit`을 쓴다.
- **L167 benchmark 지표**(p50/p95, false accept 등)에 fixture 위치·크기·ground truth 소유자가 없어 수락 L179·L181을 측정할 수 없다.
- **L157 "권리 있는 참조 자료"** — 권리 확인 방법이 없다.

### 3. 상위 문서와의 불일치

- **상위 §18 L1963 "Jev·Decisions … 과거 문구만으로 혜택을 확정하지 않는다"(미결)** vs 자식 L147-149가 정책 내용을 서술 — 상위가 미결로 둔 자리를 자식이 날짜 없이 채웠다.
- 일치(긍정): 상위 §20 L2407 "F20: native 기본·선택 typed 판단·현행 vendor 정책 확인" = 자식 L8; 상위 L2440 이미지 asset workflow = 자식 L157-159; 상위 §2 L269 = 자식 표; 상위 §15 "credential을 cloud model로 자동 전송 안 함" = 자식 L141·L183.
- 상위에 `intent.classify`·`media.generate` 등 인터페이스명이 없고, 상위 §3 router(F03)와 자식 F04 intent 가속의 경계가 적혀 있지 않다.

### 4. RC1 결정·실물과의 불일치

- **Jev 기준선 L78 — Go `4fa2aa9db2`에서 전부 확인(긍정).** `internal/jev/jev.go:1,5`("TypeSafe System One"), `:73`(`ModelID = "jev-latest"`), `internal/jev/cache.go:3`(alias와 실제 model id 분리), `internal/config/workflow_jev_test.go:13`(`workflow.jev.enabled` 기본 false), `internal/mcp/catalog.go:96`·`internal/cli/mcp_server.go:411`(`jev_ask` 등록). noul은 `cache_test.go:71`에서 확인; choice/score는 grep하지 않았다.
- **RC1 Rust 실물 — 일치(긍정).** `src/decision.rs:1-6` "judge is the running coding agent (native); a separate judging backend is P10 HOLD", `:35-45` Confidence{none, stated}·Authority{none}; `docs/rust-product-path.md:360` "P10 native only". 자식 L8·L80("권한 만들지 않음")·L100("confidence 만들어 넣지 않음")과 같다.
- **RC1에 없는 것.** 자식 L139 상태 기계, L96 budget, L97 fallback, L165 인터페이스 4개 — RC1은 `decision.record` 하나(`src/main.rs:647`). jev·image·media는 src 0건이며 `decision.rs:6`의 P10 HOLD 문장이 긍정 근거다. 이미지는 RC1 Issue가 없다(대조 보고 §4).
- **L175 "Jev 식별과 기존 연결을 migration inventory에 기록".** RC1 `.moai-cli/manifests/legacy-asset-map.json`에 `moai-jev-skill-suggestion`·`moai-ref-jev-question-design` SKILL 경로 4건이 있다(부분 충족; `jev.go` 자체의 연결 기록은 아니다).
- **L173 "key와 별도 API credit 없이 PLAN/RUN/검증 경로"** = `tests/workflow_e2e.rs:181 plan_run_verify_audit_admits_once_and_rejects_a_not_run_check`(네트워크 없음) — 일치(긍정).
- **L187 "출시 gate 밖이면 미지원/예정으로 표시".** RC1 지원표(`docs/support-matrix.md`)는 host 로딩만 다루고 backend/media 행이 없어 표시 자체가 없다.

### 5. 낡은 내용

- L147-149 vendor credit 정책 — 날짜 없음, 상위 §18이 미결로 둔 영역.
- L78 "고정 기준선 4fa2aa9db2"는 Go HEAD와 같아 유효하다.
- 근거 L193-209는 제목만 있고 "OpenAI Decisions" 문서 버전이 없다.

### 6. 개선 제안 (영향순)

1. L147-149를 "확인 날짜·URL·적용 상품"이 있는 표로 바꾸거나 상위 §18처럼 미결로 되돌린다.
2. RC1 `decision.record`(authority none, confidence none|stated, P10 HOLD)를 "현행"으로 적고 Jev·Decisions·budget·fallback은 "P10 결정 뒤"로 표시한다.
3. `verify.accept`를 검증 문서 어휘(`verify.execute`/`evidence.admit`/`audit.reduce`)로 통일한다.
4. 비교 시험 fixture의 위치·크기·ground truth 소유자를 적는다.
5. 이미지 workflow 절 머리에 "RC1 Issue 없음, F21 후속"을 둔다.
6. `max_cost_usd: 0`의 뜻을 정의한다.
7. 지원표에 backend/media 행을 넣을지 정한다(L187의 이행 수단).

### 7. 미확인·한계

- OpenAI Decisions·ChatGPT API credits·OpenAI spend limits·Claude subscription API credits·Claude monthly credits FAQ는 외부에서 확인하지 않았다.
- Jev choice/score 질문형은 grep하지 않았다(noul만 확인). `internal/jev/display_only_test.go`의 존재가 L80 "projection은 표시용"을 간접 뒷받침한다.
- Go 확인은 작업 트리 grep이다.

---

## ⑥ 다섯 문서 사이의 모순

| # | 주제 | 어긋나는 곳 | RC1 실물 |
|---|---|---|---|
| 1 | **drift 분류** | ② L44 3종 · ① L3266-3352 7종 · 상위 §5 L588-600 7종(구성 다름) | `sync.md:20` 3종(②와 동일) |
| 2 | **outbox 위치** | ② L150 "home의 작은 outbox" · ① L731 `<state_root>/projects/<id>/runs/<setup-run-id>/`·L3829 `runs/<run-id>/outbox/` · 상위 L1378-1382 `workspaces/<workspace-id>/outbox/` | `<project>/.moai-cli/forge/outbox/`(outbox.rs:7) — 네 곳이 모두 다르다 |
| 3 | **admission 인터페이스 이름** | ③ L160 `result.admit` · ④ L192 `evidence.admit` + `audit.reduce` · ⑤ L165 `verify.accept` | `audit.reduce`만(main.rs:648). 같은 식으로 ③ `owner.handoff`↔RC1 `claim.handoff`, ⑤ `decision.propose`↔RC1 `decision.record`. 합치: `effects.reconcile`(③ L160, ② L170, main.rs:635) |
| 4 | **같은 개념 다른 필드명** | candidate: ② L26 `candidate_tree` · ④ L21 `subject.tree` · ③ L23 `base_revision`(+patch). spec: ③ L24·② L25 `spec_digest` · ④ L22 `contracts.spec`. attempt: ③ L19·② L24 `attempt` · ④ `subject.attempt` · 상위 L772 `attempt_id` | `docs/protocol-v1.md`에 7개 상태 어휘 69 토큰의 crosswalk(HOLD-00D-1 미결)가 있으나 다섯 문서 어느 것도 참조하지 않는다 |
| 5 | **상태 어휘** | ③ L46-59(14) · ② L71(7+4) · ① L854·L4334·L3776 · ⑤ L139(7+6) — `blocked`·`unknown`·`stale`·`reconciling`이 기계마다 다른 뜻, 매핑 없음 | `claim.rs:50-58`, `sync.md:13`, `verify.md:5`가 각각 다른 집합 |
| 6 | **독립성의 필요조건** | ④ L212 "동일 모델 재질문 ≠ 완전한 독립" · ③ L81 "별도 model/provider는 비용·데이터 정책으로 선택" · ⑤ L167 "분리된 입력 패키지" | `audit.rs:193-196` fresh 같은 모델 = 독립(자가 신고, conservative) |
| 7 | **Issue 댓글의 역할** | ③ L69 "댓글은 lock 대체물 아님" · ② L111 "stable run marker가 있는 요약 댓글"(상태 운반체) — 목적이 달라 모순은 아니나 한 곳에 정리돼 있지 않다 | 댓글을 claim 1차 수단으로도 쓴다(claim.rs:15-21) — 세 용법이 공존 |
| 8 | **수락 시험 ID 체계** | ① 대문자 이름(L952 등) · ② TASK-*(L293) + 무ID 표(L178) · ③④⑤ 무ID 목록 | 상위 §17 15개 시험군과 대응 불가(B3) |
| 9 | **보존 장소** | ① L1198 "삭제 없이 누적·검토·검색"(archive) · ② L18·L355 "Issue에는 짧은 결론만, 상세 로그는 필요할 때만"(Issue/artifact) — 충돌은 아니나 역할 분담이 적혀 있지 않다. ① L3796 = ② L172 "worktree Markdown 복제 금지"는 일치(긍정) | RC1은 Issue 소유 댓글 1개 + 로컬 `evidence/<run>/`(report.rs:10) |

합치점(긍정): 다섯 문서 모두 "ACK·exit 0·파일 존재 ≠ 완료", "unknown→pass 금지", "epoch 기록 ≠ fencing", "결과 불명이면 재조회 우선"을 같은 문장 수준으로 공유하며, RC1 outbox·handoff·audit가 그대로 구현했다. 고칠 것은 원칙이 아니라 어휘·경로·인터페이스명의 분산이다.

---

## 공통 미확인·한계

- 다섯 사본 모두 근거 절이 링크 제목만 남아(innerText 추출) URL·버전·날짜를 확인할 수 없다. 인용이 틀렸다는 뜻이 아니라 검증할 수 없다는 뜻이다.
- RC1 Rust는 모듈 doc 주석·operation 표·시험 함수명·workflow 문서로 확인했고 `cargo test`나 바이너리 실행은 하지 않았다.
- src grep 0건(webhook, jitter, TOON, image)은 가설이며, 긍정 근거가 있는 곳(`rust-product-path.md:238·360`, `decision.rs:6`, `main.rs:614-660` 전수, `verify.rs:52`, `apply.rs:25`)만 "없음"으로 적었다.
- Go 기준선은 작업 트리 grep(HEAD `4fa2aa9db2`)이다. 상위 문서의 Mermaid 도식은 텍스트 사본에 없어 보지 못했다.
