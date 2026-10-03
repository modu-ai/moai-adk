# SPEC-TODO-CARD-ISSUANCE-001 — 조사

카드 t1454. 계획 시작 트리 `2de0a2cb6`(전체 SHA `2de0a2cb613b04765a1554f86685a3b48e0be806`, 로컬 develop 팁), 브랜치 `WT-card-issuance-overlap-graph`, 워크트리 `.claude/worktrees/t1454`.

이 문서는 세 개의 읽기 전용 렌즈가 모은 근거와, 이 계획 단계가 같은 트리에서 다시 잰 값을 묶는다. 읽는 법은 세 가지다.

- **재독 표지.** 인용한 모든 `파일:줄` 은 이 세션이 직접 다시 읽은 것이다. 렌즈 요약에서 확인하지 못한 줄은 버리거나 고쳤고, 고친 목록은 §2 에 있다.
- **귀속.** 숫자마다 (a) 이 세션이 이 트리에서 잰 값, (b) 카드가 인용한 다른 세션의 값(귀속만), (c) 재지 못한 값(공백, §10) 중 어느 것인지 적는다. 증거 없는 숫자는 쓰지 않았다.
- **큐 스냅숏.** 큐 DB(`~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db`)는 비공개이고 분 단위로 바뀐다. 큐 쪽 수치는 이 세션이 2026-10-03 11:45 에 복사한 스냅숏(`meta.last_seq` 1463, `schema_version` 2) 위에서 잰 값이다. 원본 DB 를 읽기 전용 플래그로 여는 시도는 샌드박스에서 실패했고(`unable to open database file`), 복사본은 열렸다. 복사본은 이 세션의 scratchpad 에 있고 추적되지 않는다.

## 1. 요약

- 카드의 git 쪽 헤드라인 수치(964장, 제품 줄 0 = 12%, 50줄 미만 = 32%, 공정 산출물 63%, 같은 주 파일 묶음 64개·148장·절약 84, 엄격형 39)는 이 트리에서 **전부 재현**됐다. 큐 쪽 수치는 큐가 자란 만큼만 드리프트했다.
- 어휘 겹침(token-set Jaccard)은 약한 변별자다. 기계 near 구간은 큐 전체 역사에서 한 번도 발화하지 않았고, 기록된 near-duplicate 107건은 전부 Jev 소견이다. M1 제시는 이 사실 위에서 설계해야 한다(§3.3).
- 예상 파일 입력은 거의 비어 있다. 열린 카드의 본문에 경로가 보이는 비율이 32%이고 정확히 같은 경로를 공유하는 쌍이 0이다. M2 의 명시 필드가 실질 입력이다(§3.3).
- 큐 카드 관계(`findings`)와 GTD 관계(`gtd_relations`)는 정말로 둘이고, 후자는 오늘 행이 0 이다(§5).
- `moai graph` 의 산출물은 추적되지 않는다(`.gitignore:344`). 카드의 "커밋된 edges" 전제는 낡았다(§5.3).
- 묶음의 걸림돌은 `assign --after` 가 **할당 간선**(T2)에서 이전 카드 병합을 요구한다는 점이다 — 이전 카드가 끝나기 전에는 다음 카드를 레인에 걸어 둘 수 없다(§5.4).
- M5 의 대상 파일은 이미 한도에 붙어 있고(gtd.md 40,037자, kanban-dispatch-detail.md 43,138자), 두 미러는 가드 없이 갈라져 있으며, t1453 이 같은 파일을 고칠 예정이다(§6).

## 2. 렌즈 요약에서 정정한 것

| 렌즈 요약의 서술 | 이 세션이 읽은 사실 | 처리 |
|---|---|---|
| `internal/cli/todo_runtime.go:27-36` 에 `rec.Runtime.Assignments` | 파일이 `internal/cli` 에 없다. 정의는 `internal/kanban/todo_runtime.go:15-34`(`TodoRuntime`, `TodoRuntimeAssignment`) | 위치 정정 |
| 큐 루트가 primary 체크아웃으로 해석됨(`todo_root.go:81`) | 그 파일이 없다. 해석은 `internal/cli/todo.go:74` 의 `resolveTodoQueueRoot` → `kanban.ResolveTodoQueueRootAdopting` | 위치 정정 |
| `internal/closure/gitio/gitio.go:131`, `internal/core/git/worktree.go:83` 의 워크트리 열거 도우미 | 해당 이름을 찾지 못했다 | 인용 삭제 |
| 카드→파일 선례 `cardChangedPaths` | `internal/cli/codex_review_scope.go:201` 에서 확인했다 | 유지 |
| "committed `edges.jsonl` 은 결정적이어야 하고 비공개 큐 상태를 싣지 않아야 한다" | `.moai/project/graph/` 는 `.gitignore:344` 로 추적 제외, `git ls-files` 로 센 `edges.jsonl` 은 0 | 전제 정정(§5.3) |
| 임대 경로의 `ErrPredecessorUnmerged` 처리 | `factoryNextClaimRefused`(`factory_card.go:654`)는 `ErrStaleVersion`·`ErrLeaseHolder` 만 경쟁으로 보고 나머지는 실제 오류로 반환한다 — `after` 힌트가 걸린 카드가 선택 호의 맨 앞에 서면 레인이 오류로 끝날 수 있다는 읽기 추정이고 **실행으로 확인하지 않았다** | 공백(§10) |
| `.moai/logs/rule-load-audit.jsonl` 로 40,000자 경고 실발화 확인 가능 | 이 트리에 파일이 없다 | 공백(§10) |

## 3. M0 기준선 반입 — 측정값

### 3.1 git 쪽 (카드 수치의 재현)

재현 방법: 카드의 scratchpad 스크립트(`03_git_collect.py`, `04_granularity.py` 와 `lib.py`)를 이 세션 scratchpad 로 복사해 **경로 상수만** 이 트리·스냅숏으로 바꿔 돌렸다. `fp.tsv` 는 `git log --first-parent --since=2026-08-13 --until=2026-10-02T20:25:00+09:00 --format='%H%x09%P%x09%cI%x09%s' 2de0a2cb6` 로 다시 만들었다(1,269행; 카드의 fp.tsv 는 1,262행이었고, 같은 `--since` 만 쓰면 1,300행이다 — 근사 일치). 카드 id 추출은 스크립트의 정규식 `(?<![A-Za-z0-9])t(\d{1,4})(?![0-9A-Za-z])` 를 제목에, 없으면 병합된 가지의 커밋 본문에 적용한다. 단순 grep 으로는 964 가 나오지 않는다.

| id | 값 | 카드 인용값 | 일치 |
|---|---|---|---|
| GB01 | 1,269 first-parent 커밋 중 카드 id 부착 1,151(가지 본문 대체 143) | — | 새 값 |
| GB02 | first-parent develop 커밋을 가진 카드 964(큐에 존재 867) | 964 | 동일 |
| GB03 | 제품 줄 [최소, Q1, 중앙, Q3, P90, 최대] = [0, 30, 161, 577, 1,604, 202,434] | 같은 표 | 동일 |
| GB04 | 제품 파일 [0, 1, 3, 8, 21, 732] | 같은 표 | 동일 |
| GB05 | 제품 줄 0: 111(12%) | 111(12%) | 동일 |
| GB06 | 제품 줄 50 미만: 305(32%) | 305(32%) | 동일 |
| GB07 | 제품 줄 150 미만: 468(49%) | 468(49%) | 동일 |
| GB08 | 제품 파일 3개 이하: 507(53%) | 507(53%) | 동일 |
| GB09 | 50줄 미만이면서 파일 3개 이하: 278(29%) | 278(29%) | 동일 |
| GB10 | 공정 산출물 줄 1,957,205 / 제품 줄 1,156,280 = 공정 비중 63% | 63% | 동일 |
| GB11 | 48시간 내 같은 제품 파일 집합 쌍 31(카드 49) | 31(49) | 동일 |
| GB12 | 48시간 내 파일 집합 Jaccard ≥ 0.5 쌍 90(카드 119, 12%) | 90(119) | 동일 |
| GB13 | 같은 주(主) 파일이 72시간 안에 이어진 묶음 64개, 카드 148장(15%), 절약 가능 레인 세션 84(9%) | 64·148·84 | 동일 |
| GB14 | 엄격형(같은 주 파일 + 한 카드가 다른 카드를 인용) 33개 하위 묶음, 절약 39(4%) | 39 | 동일 |
| GB15 | 허브 파일(만진 서로 다른 카드 수; 72시간 창 최대): `internal/template/catalog.yaml` 109(20), `internal/config/defaults.go` 52(12), `internal/cli/todo.go` 39(10), `.claude/rules/moai/workflow/kanban-dispatch.md` 35(10), `internal/config/types.go` 35(7), `internal/web/assets/i18n.js` 35(8) | 109/52/39/35 | 동일 |
| GB16 | 파생 카드 중 자식·부모 모두 착지한 321: 자식 발행이 부모 착지보다 앞선 215(67%), 부모 착지 후 24시간 내 91(28%); 같은 파일을 공유하고 ±48시간인 131(41%) | 321·67%·28% | 동일 |

### 3.2 큐 쪽 (스냅숏 위의 재측정)

측정 방법은 카드의 `01_inventory.py`, `02_origin_graph.py`, `05_dups_stale_lanes.py`, `06_followup_overlap.py` 를 같은 방식으로 돌린 것이다. `05` 가 읽는 `allc.tsv` 는 `git log --since=2026-08-13 --until=2026-10-02T20:28:00+09:00 --format='%h%x09%cI%x09%an%x09%x09%s' 2de0a2cb6`(8,521행; 카드는 8,461행)로 다시 만들었다.

| id | 값 | 카드 인용값 | 비고 |
|---|---|---|---|
| QB01 | 큐 행 1,174 = live 127 + 보관 1,047, `last_seq` 1,463 | 1,164 | 큐가 10행 자랐다 |
| QB02 | live 상태: dropped 90, hold 5, picked 16, queued 16 | — | 열린 카드(hold+picked+queued) 37 |
| QB03 | 보관 1,047: 상태 picked 646·queued 401, `landing` 값이 있는 것 110 | 110 | 착지 964 는 DB 가 아니라 git 에서 온 수치다 |
| QB04 | findings: live 9 + 보관 116; (relation, source): near-duplicate/jev 107(live 7, 보관 100), contains/agent 10, absorbs/agent 4, replaces/agent 2, conflicts/agent 2; **mechanical 소스 행 0** | 관계 118 | 서로 다른 (쌍, 관계) 125 |
| QB05 | 파생 카드(머리말이 이전 카드를 인용하고 파생 키워드 포함) 473(40.3%) | 468(40.2%) | 드리프트 |
| QB06 | 파생 깊이 분포 [(0, 701), (1, 292), (2, 110), (3, 49), (4, 19), (5, 3)]; depth ≥ 2 = 181(15.4%), ≥ 3 = 71(6.0%) | 178(15.3%) | 드리프트 |
| QB07 | 정규화 본문(drop 접두사 제거) 동일 그룹 61(카드 122); 저장된 그대로의 본문으로는 53그룹(106장) | 60그룹(120장) | 정규화 차이 |
| QB08 | Jev near-duplicate 107건, 점수 [최소 0.28, Q1 0.60, 중앙 0.77, Q3 0.87, 최대 0.99], 0.8 이상 47건, 그 주체 카드의 결말: done 40·dropped 3·queued 3·picked 1 | 44건 중 37 | 드리프트 |
| QB09 | 필드 채움: `spec_id` live 0·보관 7, `archived_at` 보관 76 / 1,047, `dropped_at` 1 / dropped 90, `classification` live 26 / 127·보관 55 / 1,047 | — | 과거 카드의 닫힘 시각은 대부분 없다 |
| QB10 | 동시 진행의 대리 지표(`05` 재실행의 레인 절): 비병합 커밋을 가진 서로 다른 카드 수 — 하루(UTC) 중앙 20·평균 26.7·P90 56·최대 76, **3시간 구간** 중앙 5·P90 16·최대 41; 스냅숏 시점 picked 16 | 하루 중앙 19, 3시간 구간 P90 16 | 카드 커밋을 동시 진행의 대리로 쓴 것이다 — 실제 레인 점유가 아니다 |

### 3.3 이 계획이 새로 잰 값 (발행 시점 제시의 설계 근거)

스크립트 원문은 부록 A 에 있다. 모두 위 스냅숏(1,174행) 위의 값이고 큐가 바뀌면 달라진다.

| id | 측정 | 값 |
|---|---|---|
| SB01 | 비-정확 근접 이웃이 기계 near 구간 [0.80, 1.0) 에 드는 카드 | **0장**. 이웃 점수 ≥ 0.8 인 122장은 전부 정확히 1.0(정규화 본문 일치) |
| SB02 | 발행 시점 가정(이전 카드만 비교, 정확 일치 제외)에서 가장 가까운 이웃 점수, token-set Jaccard | 중앙 0.103, P75 0.140, P90 0.239, P95 0.463, 최대 0.750 (카드 1,174장); 최근 창 id ≥ t1300(164장)은 중앙 0.098, P90 0.157, P95 0.250 |
| SB03 | 표시 하한별로 이웃이 하나라도 있는 비율(Jaccard), 전체 / 최근 창 | 0.6: 2.0% / 4.9% · 0.5: 3.1% / 4.9% · 0.4: 6.0% / 4.9% · 0.3: 8.4% / 4.9% · 0.2: 11.3% / 6.1% |
| SB04 | 같은 측정, idf 가중 Jaccard(문서 빈도는 전체 코퍼스) | 중앙 0.050, P90 0.162, P99 0.623; 하한 0.30: 5.7% / 4.9%, 0.20: 9.0% / 6.1%, 0.10: 15.5% / 9.1% |
| SB05 | 기록된 관계(125쌍 중 상대가 주체보다 먼저 발행된 114쌍)의 상대가 상위 3개 안에 드는 비율 | Jaccard: 하한 없음 59쌍(52%), 0.2 → 15(13%), 0.3 → 7(6%), 0.4 → 5(4%); idf 가중: 하한 없음 74(65%), 0.2 → 7(6%), 0.3 → 5(4%) |
| SB06 | 본문에 경로가 보이는 카드(`todoTriagePathPattern` 형) | 전체 530 / 1,174(45%); 열린 카드 12 / 37(32%); 넓은 정규식 13 / 37(35%); 경로가 있는 카드의 서로 다른 경로 수 중앙 1, 최대 8 (전체 기준) |
| SB07 | 열린 카드 쌍(666쌍)의 구성요소 키(경로 앞 두 마디) 공유 | 깊이 2: 24쌍, 깊이 3: 0쌍; 정확히 같은 경로를 공유하는 쌍 0 |
| SB08 | git 탐침 비용: 레인 브랜치 하나와 핀 사이 `git diff --name-only` | 한 번 약 0.25초(95개 파일), 로컬 `WT-` 브랜치 353개, 워크트리 74개 |
| SB09 | SPEC frontmatter 채움 | SPEC 디렉터리 1,029개 중 `card:` 가 있는 것 32개; `module:` 최빈값은 `internal/cli`(따옴표 49 + 무따옴표 42) |

해석. (1) SB01 은 기계 near 구간이 쓸모를 못 찾았음을 말한다. 기록된 near-duplicate 는 Jev 의 의미 판정이고, 그 점수(중앙 0.77)는 어휘 겹침이 아니라 모델의 확신이다. (2) SB02·SB03 은 하한 없이 상위 3개를 늘 보이면 열 번 중 아홉 번은 점수 0.1 안팎의 소음을 보인다는 뜻이다. (3) SB05 는 그 소음을 하한으로 막으면 의미 중복 상대의 대부분(0.3 에서 94%)을 놓친다는 뜻이다. 어휘 이웃은 "확실한 것"(정규화 일치, 보관·dropped 카드와의 일치)을 놓치지 않는 보조 신호이고 Jev 의 대체물이 아니다. (4) SB06·SB07 은 본문에서 경로를 뽑는 방식의 재현율이 3분의 1 수준이므로 예상 파일의 실제 입력은 M2 의 명시 필드여야 함을 말한다.

### 3.4 재현 절차와 스크립트 해시

카드의 scratchpad(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/061c4c0e-c21e-4332-8ccc-0d1d51432c49/scratchpad/cards/`)는 다른 세션의 세션 범위 `/tmp` 이므로 정리될 수 있다. 이 계획이 복사해 사용한 원본의 sha256 은 다음과 같다. M0 는 이 스크립트들을 `.moai/reports/t1454/baseline/` 로 복사하고, 원본이 사라졌다면 아래 §3.1~§3.3 의 방법 서술로 재구성한다.

| 파일 | sha256 |
|---|---|
| `lib.py` | `0853291051b813d4f3f69992a03508c3787f7b0115442c82d701ce761e06307b` |
| `01_inventory.py` | `bb246a9fecf7901d77e1115eb4b65a58c9fcccd4890cfec885dd46ee5d0ab755` |
| `02_origin_graph.py` | `ed455ed36fa5e5e1d1ae0a0e029eb68b6bb78bdbf0289cf28e1066139569cbaf` |
| `03_git_collect.py` | `e8d1ff03013856ee809504cc36ce33a3b4e0f739ebe7ce5d38d1fe87b79a3635` |
| `04_granularity.py` | `d5420903b3a2f1e82b8ba7b0f7d6ffe9c4fbc1bea488502c39f0eb488c8ad4e8` |
| `05_dups_stale_lanes.py` | `e21d0d469ac51b7a0b7dde95d09ddf4cc8f0391a7ddd86783eacc4381ddce69b` |
| `06_followup_overlap.py` | `78d8535a6cb4111e1256c0bdaf1708cf91fa893739d2b023b7fe2a3a93c77305` |

큐 쪽 집계에 쓴 읽기 전용 SQL(스냅숏 복사본 위에서 실행):

```sql
SELECT 'items', COUNT(*) FROM items;
SELECT 'items.'||state, COUNT(*) FROM items GROUP BY state;
SELECT 'archived_items', COUNT(*) FROM archived_items;
SELECT 'findings', COUNT(*) FROM findings;
SELECT 'archived_findings', COUNT(*) FROM archived_findings;
SELECT 'gtd_relations', COUNT(*) FROM gtd_relations;
SELECT 'findings.'||relation||'.'||source, COUNT(*) FROM findings GROUP BY relation, source;
SELECT 'archived_findings.'||relation||'.'||source, COUNT(*) FROM archived_findings GROUP BY relation, source;
SELECT 'archived.state.'||state, COUNT(*) FROM archived_items GROUP BY state;
SELECT 'archived.with_landing', COUNT(*) FROM archived_items WHERE landing IS NOT NULL;
SELECT 'meta.'||key, value FROM meta;
```

### 3.5 재현되지 않았거나 흔들린 것

- 큐 쪽 모든 수치는 스냅숏 시각의 값이다. 카드의 값과 다른 줄은 §3.2 비고에 이유를 적었다(대부분 큐 성장).
- `fp.tsv` 의 행 수(1,269 대 1,262)가 같지 않다. 같은 `--until` 을 줬음에도 7행 차이가 났다. 헤드라인 수치가 전부 일치했으므로 영향은 관측되지 않았지만 원인은 확인하지 못했다.
- `05` 의 "stale / blocked" 절과 `06` 의 일부 출력은 발행→착지 시간 같은 항목을 더 담는다. 이 계획의 기준에 쓰지 않아 옮기지 않았다.

## 4. 추가 경로와 카드 저장소 (렌즈 A, 재독)

**추가 진입점.** `moai todo add` 는 `internal/cli/todo.go:727-773` 의 `newTodoAddCmd` 가 `DisableFlagParsing: true`(:732) 로 자기 인자를 직접 파싱한다. `scanTodoAddArgs`(:661-714)는 `--pick`, `--force`, `--classification-file`, `--help` 만 알고(:675-690) 나머지 `--` 토큰은 `unknown flag` 로 거절하며(:693), 단일 대시 토큰은 본문으로 통과시킨다(:695-703). `cobra` 플래그 선언(:766-771)은 사용법 문구용일 뿐이다(주석 :724-726). 자연어 폴스루 `moai todo <단어…>` 는 루트 RunE 에서 `runTodoAddAppend(cmd, strings.Join(args, " "), false, todoCardDecider)`(:283-286)로 가며 **플래그를 받지 않는다**(주석 :775-778). `moai gtd add` 는 같은 서브커맨드 트리다(`internal/cli/gtd.go` 의 `NewGTDCommand` 가 `newTodoCmd()` 를 재사용). MCP `todo_add` 는 `internal/cli/mcp_todo.go:89-110` 의 `handleTodoAdd` 가 `runTodoAddAppendRoot(root, newBufferedCommand(out, errBuf), text, false, todoCardDecider)`(:107)를 부르고 `strings.TrimRight(out.String(), "\n")` 만 결과로 돌려준다 — `errBuf` 는 버려진다. 이 MCP 경로는 레인 거절(`factoryLaneRefusal`, :99 부근)을 먼저 적용한다.

**`runTodoAddAppendRoot`**(`todo.go:787-831`). 빈 본문 검사(:791), stale-store 공개는 stderr(:803), `todoStoreAt(root).Mutate`(:809) 안에서 `appendAnalyzedCard`(:811)와 분류 적용(:818)과 재정렬(:821), 끝으로 stdout 에 정확히 `"%s %d\n"`(:829). stdout 한 줄은 `TestTodoAdd_PrintsIDAndPosition`, 대시 본문 테스트들, `TestSD_AC014_MCPMatchesCLIWithProjectRoot`(`factory_m3_test.go`; MCP 도구 결과의 TextContent 를 이어 붙인 문자열이 CLI stdout 과 같아야 한다 — `sdCallTool` 이 모든 TextContent 를 이어 붙인다)가 고정한다.

**분석기.** `appendAnalyzedCard`(`internal/cli/todo_analysis.go:38-100`): `kanban.ClassifyCardText(text, rec.Items)`(:39)가 `BacklogMatchExact` 이고 `--force` 가 없으면 Mutate 콜백이 오류를 돌려줘 파일이 바이트 그대로 남는다(:40-43). `BacklogMatchNear` 는 `rec.AppendFindingOnce` 로 `near-duplicate`/`mechanical` 소견 하나를 남긴다(:74-81). Jev 근접 중복 탐침은 그 뒤(:96)이고 `workflow.jev.enabled` 가 켜졌을 때만 돈다(이 값은 이 세션이 읽지 않았다).

**분류기.** `internal/kanban/backlog_analysis.go`: 임계값 `BacklogNearDuplicateThreshold = 0.80`(:33), `NormalizeCardText`(:71, NFC·trim·공백 접기·소문자, 문장부호 유지), `TokenSetJaccard`(:86, 집합 기반), `ClassifyCardText`(:135)는 `State == BacklogStateDropped` 를 건너뛴다(:139; 이유는 함수 문서 주석 — dropped 는 거절 대상이 아니다). 보관 카드는 `rec.Items` 에 없고 `rec.Archived`(`BacklogArchiveEntry`, `backlog_store.go:328`)에 있다 — `LoadPure` 가 이미 메모리에 올려 둔다. dropped 카드 본문은 `[DROPPED — <reason>] ` 접두사를 달고(`todo_drop.go:122`), 사유는 이 접두사로만 남는다. `stripTodoDropMarker`(`todo_drop.go:214`)가 접두사를 벗긴다.

**`gtd engage`.** `internal/kanban/gtd_engage.go:136` 의 `store.addWithCardUUID(item.Content, &identity)` 는 `backlog_store.go:952` 에서 곧바로 추가하고 정렬만 한다 — 분석도 분류 결정기도 거치지 않는다(`appendAnalyzedCard` 는 cli 패키지에 있다). engage 입력에는 이미 `DryRun`(`gtd_engage.go:127`, `if in.DryRun { return EngageResult{Actionable: true}, nil }`)이 있다.

**재사용 가능한 도우미.** `todoTriageSymbols`(`internal/cli/todo_triage.go:439`)는 세 정규식(백틱 식별자, `dir/file.(go|md|yaml|sh)` 경로 :84, 호출형)이고 최대 `todoTriageMaxSymbols = 4`(:49)개만 낸다 — 경로 겹침 용도에는 상한이 작다. 구성요소 판별·예상 파일 추출·진행 중 레인과의 겹침을 내는 도우미는 없다. 레인 사실은 `rec.Runtime.Assignments`(`internal/kanban/todo_runtime.go:15-34`: `CardID`, `OwnerLabel`, `ReportedState`)에 있고 레인이 바꾼 파일은 git 에서만 나온다(`cardChangedPaths`, `internal/cli/codex_review_scope.go:201`이 선례). 완료 SPEC 은 `spec.ListDocs(baseDir)`(`internal/spec/listdocs.go:36`)가 `SPECFrontmatter`(`internal/spec/lint.go:546-562`: `ID, Title, Status, Module, Tags` 등)를 돌려준다. 카드→SPEC 링크는 `BacklogItem.SpecID` 하나뿐이고 채움률이 0.6%(7/1,174)다(QB09). `add` 시점 `--dry-run` 선례는 `auto-done --dry-run`(`internal/cli/todo_autodone.go` 의 `newTodoAutoDoneCmd`)이다.

**저장소.** SQLite `~/.moai/db/<project-key>/todo/backlog.db`(프로젝트 키 `moai-adk-go-1bd3d038`). 큐 루트는 primary 체크아웃으로 해석된다(`internal/cli/todo.go:74-76`). `BacklogItem`(`backlog_store.go:85-127`)에는 부모·출처·크기·예상 파일·drop 사유 필드가 없다. 스키마는 `items`, `archived_items`(position·archived_at·landing_verdict 추가), `findings`, `archived_findings`, `meta`다. 안전한 새 컬럼 경로는 `ensureSchema`(`backlog_sqlite.go:410-468`)와 `ensureColumn`(:599, `pragma_table_info` 로 존재 판정)이고 순서가 중요하다 — 버전 조정이 먼저, 가산 retrofit 이 나중(카드 t1310의 주석). 새 컬럼은 v1→v2 재구성 목록 `backlogItemsTableColumns`(:122)에 넣지 않는다. 새 컬럼이 닿는 곳: 쓰기 `INSERT INTO items(...)`(`backlog_migrate.go:586`)·`INSERT INTO archived_items(...)`(:469), 읽기 `readSnapshot`(:86)·`readArchive`(:243)와 `columnExpr`(:75; 없는 컬럼은 NULL 로 읽어 DDL 을 안 돌리는 순수 읽기를 지킨다), `assertBacklogParity`(:904), 동결 컬럼 튜플 `backlog_schema_freeze_test.go`(`wantItemsColumns` 와 `wantArchivedItemsColumns`, 각 문자열이 `lease_expires_at:TEXT:0:NULL` 로 끝난다), 큐 병합·이주 코드(`internal/kanban/todo_queue_merge.go`, `todo_merge_procedure.go`가 항목과 소견을 복사·재매핑). 선례: `classification` 한 컬럼이 JSON 으로 여러 속성을 나른다(`backlog_store.go:85-127` 의 `Classification` 주석). `list --json` 은 `todoJSONProjection`(`internal/cli/todo_claim.go:79-100` 부근, 임대 필드를 투영에서 뺀다)과 `omitempty` 로 골든 `internal/cli/testdata/ac_tst_012_golden.json`(859바이트)을 `TestTodoListJSON_GoldenByteIdentity`(`todo_json_golden_test.go:65`)가 고정한다.

**`done`/`undone`.** `ArchiveCard`(`backlog_store.go:397`)가 항목과 소견을 `rec.Archived` 로 옮기고 `RestoreCard`(:438)가 되돌린다. dropped 는 live `Items` 에 남는다. 소견은 카드와 함께 보관으로 이사하므로 한쪽 끝이 보관되면 그 소견은 live 큐에서 보이지 않는다. 락: `Mutate`(:876)가 교차 프로세스 `backlog.lock` 을 읽기-수정-쓰기 내내 쥔다. 느린 점검(git 탐침)은 Mutate 안에서 돌면 안 된다.

**특성 테스트(이 변경이 건드리면 안 되는 것).** `TestTodoAddRefusesExactDuplicate`, `TestTodoAddForceAdmitsAndRecords`, `TestTodoExactRefusalWorksWithoutAgent`, `TestTodoAddNearDuplicateRecordsOnly`, `TestTodoAnalysisNeverReordersQueue`, `TestTodoAnalyzeRerunIsIdempotent`, kanban `TestNormalizeCardText`/`TestTokenSetJaccard`/`TestClassifyCardText`, `TestTodoAdd_PrintsIDAndPosition`, `TestTodoAddLeadingDashText`, `TestTodoAddPick`, `TestTodoMultiWordFallthroughAdds`, `TestTodoAddDefaultDeciderPrintsNoNotice`, `TestTodoStaleStoreDisclosure_AddVerb`, `TestTodoWriteVerbs_CarryNoDisclosure`, `TestTodoConcurrentAdd_8Processes`, `TestTodoAddPick_OneLockedWrite`, `TestTodoListJSON_GoldenByteIdentity`, 스키마 동결 테스트(`TestTodoHistoryAddsNoSchemaChange`, `TestSchemaFreezeRecordsTransitionStamps`), `TestTodoListDefaultLimit`, `TestSD_AC014_MCPMatchesCLIWithProjectRoot`, 문서 정합 `todo_skill_doc_parity_test.go`·`todo_classify_doc_parity_test.go`. 이름은 렌즈 요약에서 가져왔고 이 세션이 `-list` 로 존재를 확인한 것은 `TestClassifyCardText`, `TestNormalizeCardText`, `TestTokenSetJaccard`, `TestTodoAddRefusesExactDuplicate`, `TestTodoAddNearDuplicateRecordsOnly`, `TestTodoAdd_PrintsIDAndPosition`, `TestSD_AC014_MCPMatchesCLIWithProjectRoot`, `TestTodoListJSON_GoldenByteIdentity` 여덟 개다(`acceptance.md` 맥락 행). 나머지는 실행 단계의 M1 시작 점검이 `-list` 로 확인한다.

**레인 가드.** `todoRefuseLaneMutation`(`internal/cli/todo.go:387-409`)은 레인 세션의 큐 변경을 거절하고 읽기 전용 동사 허용 목록 `todoLaneReadOnlyVerbs` 에 든 서브커맨드만 통과시킨다. 이 세션의 `moai todo add --help` 가 "refused — lane boundary" 로 끝났다(exit 1) — `add` 는 레인에서 `--dry-run` 이어도 거절된다. 새 읽기 동사 `trace` 는 이 목록에 넣어야 레인이 쓸 수 있고, `merge` 는 넣지 않는다.

## 5. 관계·그래프·레인 (렌즈 B, 재독)

### 5.1 두 관계 저장소

`moai gtd` 는 `moai todo` 동사 트리에 GTD 서브커맨드를 더한 것이다(`gtd.go` 의 `NewGTDCommand`: `cmd := newTodoCmd(); cmd.Use = "gtd"` 후 `AddCommand(newGTDCaptureCmd(), newGTDClarifyCmd(), newGTDOrganizeCmd(), newGTDReflectCmd(), newGTDEngageCmd(), newGTDAnswerCmd())`). `relate`/`unrelate`(`todo_relate.go`)는 큐 레코드의 `Findings` 배열에만 쓴다 — 머리 주석이 "카드 필드를 쓰지 않음을 코드 모양이 강제한다"고 적는다.

- 큐 카드 소견: `BacklogFinding{SubjectID, RelatedID, Relation, Source, Score, Note, At}`(`backlog_store.go:208-216`), 표 `findings`/`archived_findings`, UNIQUE 제약 없음(중복 제거는 응용 계층: `HasFindingTuple` :516, `AppendFindingOnce` :528, 키 {subject, related, relation, source}). 관계값: 기계 `duplicate-forced`·`near-duplicate`, 의미 `contains`·`absorbs`·`replaces`·`conflicts`·`blocks`·`depends`(`BacklogSemanticRelations`); 소스 `mechanical|agent|jev`. 순환 가드는 `blocks`/`depends` 만(`WaitsOnOf`, `FindingsBlocking`, `WaitsOnClosesCycle` — `backlog_store.go` 의 `:228-292` 범위).
- GTD 관계: 표 `gtd_relations`(`internal/kanban/backlog_gtd_schema.go:47-57`; 열 subject_id, object_id, kind, note, source, assertion_status, source_revision, policy_version; PK (subject_id, object_id, kind)), 종류 아홉(`gtd_relation.go:10-20`: depends_on, part_of, supported_by, related_to, supersedes, contains, absorbs, replaces, conflicts). `ValidateGTDRelation`(:48)은 쌍 단위(빈 id·자기 간선·모르는 종류·대상 존재)이고 `depends_on` 에만 DFS 순환 검사가 있다(`dependency_cycle`). 카디널리티 제약은 없다. 이 표는 GTD 항목(`gtd-<16hex>`)을 잇고 카드(`tN`)는 `gtd_items.card_id`(engage 때 설정)로만 닿는다. 스냅숏에서 `gtd_relations` 0행, `gtd_items` 0행이다. `moai gtd organize` 만 쓰고 `unrelate` 에 해당하는 동사는 없다.
- SPEC-RELATION-PICKUP-FILTER-001 §A.3 가 "섞지 말 것"이라고 적었고, SPEC-TODO-AUTO-PICK-001 §B.7·REQ-TAU-013 이 통합을 카드 t1454 로 예약했다.

### 5.2 `moai graph`

`edges.jsonl` 은 한 줄 한 간선, `(kind, source, target, line)` 정렬(`EdgeLess`, `internal/graph/graph.go:223`), 시각 정보 없음, "같은 트리에서 두 번 돌리면 바이트 동일"이 파일 머리 주석의 계약이다(`graph.go:1-20`). 간선 종류에 `report-milestone`, `milestone-card`(`report.go:36`, 카드 열에서 온다 — 본문의 `tNN` 을 카드 노드로 쓴다)가 있고 `reportEdges`(:58)가 만든다. `buildDocLayers`(:181)가 문서 층을 모으고 정렬은 :207 에서 한다. 신선도 판정 `checkEdges`(`check.go:666`)는 네 출처 집합(코드맵, MX 인덱스, specs, reports)의 지문을 비교한다(`SourceFingerprintsForEdges`, `meta.go:37`; `EdgesSourcesMoved` :154). 새 간선 출처를 더하면 지문 목록에 넣어야 신선도 판정이 알아챈다. CI 의 `.github/workflows/graph-freshness.yml` 이 `moai graph build` 후 `moai graph check` 를 돌린다 — CI 체크아웃에는 큐 DB 가 없다.

### 5.3 카드→파일 간선의 운반체

`.moai/project/graph/` 는 `.gitignore:344` 로 추적 제외이고(`.gitignore` 330-344 주석이 이유를 적는다: 십만 줄 규모라 `git add -A` 한 번에 검토 불가 diff 가 된다), `git ls-files | grep -c edges.jsonl` 은 0, `check.go:669` 도 "untracked derived artifact"라 부른다. 따라서 "큐 상태를 커밋된 산출물에 새기지 말 것"은 위반 대상이 이미 없고, 남는 제약은 (i) 같은 트리 동일 출력 계약, (ii) 큐가 없는 CI 와 있는 로컬의 산출물이 갈라지지 않을 것, (iii) 비공개 큐 상태가 `moai graph query` 소비자(MCP 그래프 도구는 `code-call` 간선만 읽는다)로 새지 않을 것이다. 선례: GTD 투영은 `~/.moai/db/<key>/todo/gtd-edges.jsonl` 에 모드 0600 비공개로 쓴다(`internal/graph/gtd_private.go:34` 의 `privateGTDDataName`). 착지 증거 쪽 도구: `GitLandedQuerier.Landed`(`internal/kanban/prlink_landed.go:337`)는 3치 답이고 의도적으로 SHA 를 돌려주지 않는다. `subjectAttribution`(:188)은 커밋 제목만으로 귀속하고(형태 `(tN)`, `type(tN):`, `Merge card tN`, `merge: tN`, 통합 병합), `LandedAttributions`(`autodone_scan.go:273`)가 카드 id → 첫 귀속 `LandedCommit{SHA, Subject, CommitTime}` 를 낸다. 카드→변경 파일을 내는 함수는 없다.

### 5.4 레인과 묶음

`moai factory assign` 은 `internal/cli/factory_card.go:1703` 부근에서 `--prefer`·`--after`·`--spec` 를 읽고 `db.RecordPicked`(:1739)와 `db.Transition(... CardAssigned ...)`(:1744)를 부른다. `HintAfter` 는 `internal/homestate/card_record.go:106` 의 `cards` 행 필드이고, 강제 위치는 **하나**다 — 할당 간선의 가드(`card_transition.go` 의 `case guardAssign:`, `if cur.HintAfter != ""` → `predecessorMerged`). `predecessorMerged`(`internal/homestate/card_picked.go:181`)는 이전 카드의 **팩토리 레코드**가 `merged-local`/`pushed`/`ci-green`/`done` 에 도달했을 것을 요구하고, 레코드가 없으면 `ErrUnknownPredecessor` 다. 귀결: 다음 카드는 이전 카드가 병합되기 전에는 레인에 **할당될 수 없다**(T2 가 거절한다). "직렬"은 레인별 순서가 아니라 전군 단일 슬롯이다 — `CardClassification.Mode`(serial 기본), `factorySerialSlotHeld`(`factory_card.go:229`), `factorySerialInFlightExcluding`(:244). 임대 선택 `factoryNextSelectAndLease`(:452)의 호(arm) (a) 이 레인에 할당된 카드, (b) 소유자 없는 운영자 pick 카드, (b2) 큐에서만 picked 인 카드(레코드 생성 후 임대), (c) 가장 높은 순위의 대기 카드를 순서대로 훑고, 호 (a)·(b)·(b2) 는 첫 후보에서 `factoryNextClaim` 으로 `return` 한다. `factoryNextClaim`(:625)의 거절 매핑 `factoryNextClaimRefused`(:654)는 `ErrStaleVersion`/`ErrLeaseHolder` 만 경쟁으로 본다. `factoryKeepSetRefusal`(:771)은 관계 저장소도 파일 겹침도 읽지 않는다(REQ-TAU-013). 관계는 `todo --auto` 픽업에서만 쓰이고 팩토리 임대 선택은 읽지 않는다. 팩토리 `cards` 표는 `internal/homestate/factory.go:45` 에서 만들어지고 컬럼 추가는 같은 파일의 `ALTER TABLE ... ADD COLUMN` 목록(:261-266, :337-338)으로 한다.

## 6. 규칙과 템플릿 (렌즈 C, 재독)

**크기(`wc -m`, 문자).** `.claude/rules/moai/workflow/kanban-dispatch.md` 28,092(상시 로드; 바이트 28,301), `kanban-dispatch-detail.md` 43,138(이미 한도 초과), `kanban-dispatch-mechanics.md` 18,767, `.claude/skills/moai/workflows/gtd.md` 40,037(한도 초과 37자), `.claude/agents/moai/sync-auditor.md` 24,458, `manager-todo.md` 5,313. 템플릿 사본: `kanban-dispatch.md` 27,771, `sync-auditor.md` 23,880, 나머지 넷은 로컬과 같은 크기(40,037·43,138·5,313; `cmp` 로 gtd.md 쌍은 바이트 동일, `kanban-dispatch.md` 쌍은 `differ: char 25113, line 181`).

**한도와 예산.** 훅의 문자 예산은 `internal/hook/instructions_loaded.go:103`(`const charBudget = 40000`, `utf8.RuneCount`, 초과 시 오류 반환)이다. 호출 쪽이 그 오류를 어떻게 다루는지는 따라가지 않았다. SPEC-INSTRUCTION-BUDGET-SCOPE-001(completed)의 §1 이 "`LoadReason` 는 게이트로 쓰이지 않는다, 따라서 `paths:` 한정 규칙도 글롭 적재마다 잰다"를 기록했고, 그 SPEC 이 정한 동반 파일 규칙은 "동반 패턴 ⊊ 부모 패턴, 여집합 비어 있지 않음"이며 `kanban-dispatch*.md` 글롭은 자기 매칭 이름 함정을 만든다. 상시 로드 토큰 가드 `AlwaysLoadedTokenBudget = 77600`(`internal/config/token_budget_guard.go:86`)의 여유는 이 세션이 다시 재지 않았다. `.claude/rules/moai/development/rule-authoring.md` 는 상시 로드 파일을 1,000바이트 넘게 키우는 편집에 크기·비용 진술을 요구한다. SPEC-TODO-AUTO-PICK-001 의 plan 은 `kanban-dispatch.md` 순증가 ≤ 0 을 걸었다(그 plan 문서의 상시 로드 성장 절).

**미러 가드.** `internal/template/card_id_leak_test.go:36-55` 의 기준선 목록은 `.claude/rules/moai/workflow/kanban-dispatch.md t1330`, `kanban-dispatch-detail.md t133·t224`, `.claude/skills/moai/workflows/gtd.md t696` 등 (경로, 카드 id) 쌍만 허용한다 — 템플릿 사본은 새 `t####` 를 얻으면 안 된다. `workflow_rule_paths_pinned_test.go:28-33` 은 `kanban-dispatch-detail.md` 의 `paths:` 를 `**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/gtd.md` 로 고정한다. `contract_mode_guided_test.go:679` 는 `kanban-dispatch.md` 를 "R" 로 분류한다. `jev_auto_exception_test.go:73-78` 은 `kanban-dispatch.md`, `gtd.md`, `manager-todo.md` 를 앵커로 둔다. `todo_skill_doc_parity_test.go` 는 `gtd.md` 의 findings-source 열거 문장(`` `source` is `mechanical` ``)과 템플릿 사본의 중립성(SPEC/REQ id·날짜·9자 이상 16진 단어 금지)을 지킨다. `rule_template_mirror_test.go` 의 `declaredForkedPairs`(:54)는 갈라진 쌍을 `mirror-fork: intentional` 표지와 이유로 선언하는 등록부이고 지금 세 쌍(`verification-claim-integrity.md`, `agent-common-protocol-reference.md`, `cross-session-messaging-detail.md`)만 있다 — **`kanban-dispatch.md` 와 `sync-auditor.md` 는 등록돼 있지 않고 바이트 동일 목록에도 없다**(이 세션이 `git grep` 으로 확인). 따라서 두 쌍의 분기는 가드 없이 서 있다.

**빌드와 생성물.** `Makefile` 의 `build` 대상(:34-37)은 `agents-emit-check commands-emit-check tool-policy-drift-check templ-generate` 후 `go run ./internal/template/scripts/gen-catalog-hashes.go --all` 를 돌린다. `sync-auditor.md` 와 `manager-todo.md` 편집은 `make agents-emit`(:38)으로 `.codex/agents/moai/*.toml`(저장소 루트와 `internal/template/templates/.codex/agents/moai/` 아래)을 다시 만들고 `internal/template/catalog.yaml` 해시(예: `sync-auditor` 항목 :132-136)를 갱신해야 한다.

**현재 문장.** 카드 등급 A/B/C 와 Class A "측정된 증거" 문구는 `kanban-dispatch.md` 의 "Card classes — not every card needs every column" 절에 있고 B·C 에는 크기 하한·상한이 없다. WIP 는 "one card per worktree"뿐이고 동시 진행 한도가 없다. `sync-auditor.md` 에는 blocking|optional 분류와 `PASS-WITH-DEBT` 토큰이 있고 구성요소별 부채 대장은 어디에도 없다. `manager-todo.md` 에는 후속 지적 언어가 없고, 파생 깊이 언어는 어디에도 없다("Derived, never invented" 는 본문 출처에 관한 문장이다, `gtd.md:423`). 병합 금지의 문면은 `gtd.md:63`(`absorbs` 행의 "`absorbs` does not absorb")와 `:114`("Analysis never folds one card into another…")이다.

## 7. 웹 (렌즈 C, 재독)

`/todo` 라우트는 `internal/web/app.go:205`(`mux.HandleFunc("/todo", a.handleTodo)`), 서버 렌더 Templ `templ Todo(vm ShellVM, q TodoVM)`(`screens.templ:297`), 생성물 `screens_templ.go` 는 커밋돼 있고 `make build` 가 `templ-generate` 를 돈다. `readTodoQueue`(`todo_queue_read.go:32`)가 `kanban.NewBacklogStore(...).LoadPure()`(:35)를 부르는 유일한 이음매다. 관계는 `rec.Findings` 만 읽어 `TodoItemVM`(`todo_view.go:56`: ID, Text, State, SpecID, Relations []string)의 미리 포맷한 문자열로 간다(`screens.templ:342-346`, 상세 :393-397). 보관 카드·`archived_findings`·`gtd_relations` 는 페이지가 읽지 않는다. 프런트엔드는 JS 프레임워크도 빌드 단계도 없고 자산을 `//go:embed assets/console.css assets/app.js assets/i18n.js assets/htmx.min.js assets/fonts assets/mascots`(`assets.go:24`)로 묶는다 — 새 자산은 이 목록에 들어가야 한다. `assets/i18n.js` 는 258,310바이트 단일 파일이고 i18n 거버넌스 테스트(`i18n_governance_test.go`)가 키를 요구한다. 실시간 갱신은 `events.go:35` 의 `"kanban": {".moai/state/todo"}` 감시 맵이다. 기존 테스트: `todo_route_test.go`(함수 6개), `todo_section_test.go`, `todo_sort_test.go`, `todo_queue_read_test.go`, `todo_queue_sorted_test.go`, `todo_hold_render_test.go`, 등.

## 8. 같은 파일을 만지는 다른 카드

- **t1448**(todo --auto 자율 선택): 이 트리에 착지(병합 `4315f0d0e`; `git merge-base --is-ancestor 4315f0d0e HEAD` exit 0). SPEC-TODO-AUTO-PICK-001 completed.
- **t1453**(github-flow 전환): 선택(picked), 브랜치 `WT-github-flow-default`(끝 `2a5f9c91c`), SPEC-GITHUB-FLOW-DEFAULT-001 in-progress. 그 plan 의 M4 는 "t1448 착지 뒤" 절체 시점에 `kanban-dispatch*`·`main-checkout-branch-guard*`·`worktree-integration*` 등 규칙 파일을 묶음으로 고친다. 핀 `2de0a2cb6` 과 그 브랜치 사이의 `git diff --stat` 는 이 SPEC 의 규칙 여섯 파일과 `todo.go` 에 대해 비어 있었다. 병합 여부는 `moai gtd pr t1453` 가 `no-link … picked` 로 답했고 `git rev-list --first-parent --count -E -i --grep='^merge[( :]+(card )?t1453' HEAD` 는 0 이다.
- **t1450**(상시 로드 규칙 다이어트): 대기. 역할 한정 규칙을 SessionStart 훅 주입으로 옮기는 안을 다룬다고 하나 이 세션은 그 카드의 본문 이상을 읽지 않았다 — 적재 범위가 바뀌면 M5 스텁 조항의 위치가 달라진다.
- **t1452**(병합 창 대기 단축): 대기. 통합 조항(`kanban-dispatch.md` 181행의 `moai worktree sweep` 문장 부근)을 고칠 가능성이 있으나 큐 문구가 파일명을 말하지 않아 겹침은 추정이다.
- **t1349**: 대기 상태 행이 남았지만 세 수리는 이미 착지했다(§2).
- **t1359**: 대기. `gtd.md` 문서 정정 두 건(`blocks`/`depends` 관계 행, `gtd.md:53` 의 옛 어휘 문장과 그 미러) — M3·M5 와 같은 파일이다.

## 9. 가장 가까운 SPEC (상태는 이 세션이 `status:` 줄로 확인)

completed: SPEC-TODO-ANALYSIS-001(기계 분석·소견·비수정 독트린), SPEC-TODO-ARCHIVE-QUERY-001, SPEC-TODO-SURFACE-POLISH-001, SPEC-TODO-CLASSIFY-DISPATCH-001(생성 시 분류 JSON 컬럼 — M2 의 스키마 선례, 직렬 슬롯의 기원), SPEC-RELATION-PICKUP-FILTER-001, SPEC-TODO-AUTO-PICK-001, SPEC-GTD-AUTONOMY-001, SPEC-TODO-CLAIM-LEASE-001(임대 컬럼 두 개 — 반대 방향 선례), SPEC-TODO-TRANSITION-STAMPS-001, SPEC-WEB-TODO-QUEUE-001, SPEC-JEV-CONSUMERS-001, SPEC-INSTRUCTION-BUDGET-SCOPE-001, SPEC-TODO-LANDING-ATTRIBUTION-001, SPEC-FACTORY-SELF-DISPATCH-001. in-progress: SPEC-KANBAN-PR-CARD-TRACEABILITY-001. 카드 발행 품질을 직접 다루는 SPEC 은 없다.

## 10. 공백 (측정하지 못했거나 확인하지 못한 것)

1. 어휘 이웃의 정밀도(정답 집합 부재).
2. "이미 덮는 완료 SPEC" 조회의 재현율·정밀도(`card:` 32/1,029, `spec_id` 7/1,174).
3. 진행 중 레인 겹침의 실제 입력 가용성 — 예상 파일 필드가 어느 카드에도 없다.
4. N개 진행 중 레인에 대한 제시 경로 전체 지연(0.25초는 브랜치 하나, 스냅숏 시점 picked 16개).
5. 카드 귀속 병합에서 변경 파일을 뽑는 `moai graph build` 의 시간.
6. 웹 그래프의 레이아웃·크기·브라우저 렌더 증거.
7. 기록 단계에서 같은 본문이 보관 카드와 충돌한 사례 수 — 정규화 일치 61그룹 중 첫 사본이 보관된 것은 31그룹이지만 **그 보관이 둘째 사본 발행 전이었는지는 확인하지 못했다**(`archived_at` 이 76/1,047 에만 있다).
8. t1453 이 규칙 파일을 고친다는 판단은 이동하는 ref(브랜치 끝)의 plan 문서를 읽은 것이다.
9. `ErrPredecessorUnmerged` 가 선택 호의 맨 앞 후보에서 오류로 끝나는지(§2) — 읽기 추정, 실행 확인 없음.
10. `gtd.md`(스킬 본문)가 40,000자 훅에 실제로 측정되는지, `workflow.jev.enabled` 의 현재 값, 상시 로드 토큰 여유(`.moai/logs/rule-load-audit.jsonl` 이 이 트리에 없다).
11. MCP `StructuredContent` 가 Claude Code 에서 어떻게 보이는지(D5 의 기본값은 이를 쓰지 않는다).
12. 구성요소의 정의(경로 앞 두 마디)는 작업 정의다 — 깊이 3 에서는 공유 쌍이 0이고 정답이 없다.
13. `hold`/레인 세션에서 `add --dry-run` 이 거절되는 것(§4)은 관찰했으나 그 정책이 의도인지는 확인하지 못했다.
14. MCP `todo_*` 핸들러 중 관계를 내는 표면, `todo export` 가 새 필드를 어떻게 다루는지, 훅·스크립트가 소견 JSON 을 읽는지는 추적하지 않았다.

## 부록 A — 이 계획의 측정 스크립트 (원문, 스냅숏 복사본 위에서 실행)

각 스크립트의 `DB` 상수는 이 세션 scratchpad 의 스냅숏 복사본을 가리킨다. M0 는 이 경로만 자기 스냅숏으로 바꿔 돌린다. 출력은 §3.3 에 옮겼다.

### A.1 코퍼스 전체 Jaccard 분포 (`t1454_jaccard.py`)

```python
import sqlite3, unicodedata, re, collections, statistics, sys
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
rows = []
for cid, text, st in con.execute("select id,text,state from items"):
    rows.append((cid, text, 'live:' + st))
for cid, text, st in con.execute("select id,text,state from archived_items"):
    rows.append((cid, text, 'archived'))


def norm(s):
    return ' '.join(unicodedata.normalize('NFC', s).split()).lower()


DROP = re.compile(r'^\[DROPPED\s*—\s*[^\]]*\]\s*')


def toks_stripped(text):
    return frozenset(norm(DROP.sub('', text)).split())


print('rows', len(rows))
g = collections.defaultdict(list)
for cid, text, st in rows:
    g[norm(DROP.sub('', text))].append(cid)
dups = [v for v in g.values() if len(v) > 1]
print('identical normalized text groups (drop prefix stripped):', len(dups), 'cards', sum(len(v) for v in dups))
g2 = collections.defaultdict(list)
for cid, text, st in rows:
    g2[norm(text)].append(cid)
dups2 = [v for v in g2.values() if len(v) > 1]
print('identical normalized text groups (as stored):', len(dups2), 'cards', sum(len(v) for v in dups2))
T = [(cid, toks_stripped(text), st) for cid, text, st in rows]
n = len(T)
best = []
top3 = []
for i in range(n):
    sc = []
    a = T[i][1]
    for j in range(n):
        if i == j:
            continue
        b = T[j][1]
        if not a or not b:
            continue
        inter = len(a & b)
        u = len(a) + len(b) - inter
        sc.append(inter / u)
    sc.sort(reverse=True)
    best.append(sc[0] if sc else 0)
    top3.append(sc[:3])


def q(vals, p):
    vals = sorted(vals)
    return vals[int(p * (len(vals) - 1))]


print('best-neighbor Jaccard: median %.3f P75 %.3f P90 %.3f P95 %.3f P99 %.3f max %.3f' % (
    q(best, .5), q(best, .75), q(best, .9), q(best, .95), q(best, .99), max(best)))
for th in (0.8, 0.7, 0.6, 0.5, 0.4, 0.3):
    c = sum(1 for b in best if b >= th)
    print('cards with a neighbor >= %.1f: %d (%.1f%%)' % (th, c, 100 * c / n))
print('cards with neighbor in [0.80,1.0):', sum(1 for b in best if 0.8 <= b < 1.0), ' ==1.0:', sum(1 for b in best if b >= 1.0))
for th in (0.5, 0.4, 0.3, 0.2):
    print('cards whose 3rd best neighbor >= %.1f: %d' % (th, sum(1 for t in top3 if len(t) == 3 and t[2] >= th)))
tc = [len(t[1]) for t in T]
print('tokens per card: median %d P90 %d max %d' % (statistics.median(tc), q(tc, .9), max(tc)))
```

### A.2 발행 시점 가정 — 이웃 점수 밀도 (`t1454_issuance_sim.py`)

```python
import sqlite3, unicodedata, re, statistics
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
rows = []
for cid, text, st, seq in con.execute("select id,text,state,seq from items"):
    rows.append((int(cid[1:]), cid, text, 'live:' + st))
for cid, text, st, seq in con.execute("select id,text,state,seq from archived_items"):
    rows.append((int(cid[1:]), cid, text, 'archived'))
rows.sort()


def norm(s):
    return ' '.join(unicodedata.normalize('NFC', s).split()).lower()


DROP = re.compile(r'^\[DROPPED\s*—\s*[^\]]*\]\s*')
toks = {r[1]: frozenset(norm(DROP.sub('', r[2])).split()) for r in rows}
normed = {r[1]: norm(DROP.sub('', r[2])) for r in rows}


def jac(a, b):
    if not a or not b:
        return 0.0
    i = len(a & b)
    return i / (len(a) + len(b) - i)


def q(vals, p):
    vals = sorted(vals)
    return vals[int(p * (len(vals) - 1))]


def sim(window_lo):
    res = []
    exact = 0
    for n, cid, text, st in rows:
        if n < window_lo:
            continue
        prior = [r for r in rows if r[0] < n]
        sc = []
        ex = False
        for _, pid, _, pst in prior:
            if normed[pid] == normed[cid]:
                ex = True
                continue
            sc.append((jac(toks[cid], toks[pid]), pid))
        sc.sort(reverse=True)
        res.append([s for s, _ in sc[:3]])
        exact += 1 if ex else 0
    return res, exact


for lo in (1, 1300):
    res, exact = sim(lo)
    n = len(res)
    print('window ids >= t%d: cards %d, with an exact normalized duplicate among EARLIER cards (live+archived+dropped): %d' % (lo, n, exact))
    top1 = [r[0] for r in res if r]
    print('  top-1 non-exact Jaccard: median %.3f P75 %.3f P90 %.3f P95 %.3f max %.3f' % (q(top1, .5), q(top1, .75), q(top1, .9), q(top1, .95), max(top1)))
    for th in (0.8, 0.6, 0.5, 0.4, 0.3, 0.2):
        c1 = sum(1 for r in res if r and r[0] >= th)
        c3 = sum(1 for r in res if len(r) == 3 and r[2] >= th)
        print('  floor %.1f: cards with >=1 neighbor %d (%.1f%%), with all 3 slots filled %d (%.1f%%)' % (th, c1, 100 * c1 / n, c3, 100 * c3 / n))
```

### A.3 idf 가중 밀도 (`t1454_idf_sim.py`)

```python
import sqlite3, unicodedata, re, math, collections
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
rows = []
for cid, text, st in con.execute("select id,text,state from items"):
    rows.append((int(cid[1:]), cid, text))
for cid, text, st in con.execute("select id,text,state from archived_items"):
    rows.append((int(cid[1:]), cid, text))
rows.sort()


def norm(s):
    return ' '.join(unicodedata.normalize('NFC', s).split()).lower()


DROP = re.compile(r'^\[DROPPED\s*—\s*[^\]]*\]\s*')
T = {r[1]: frozenset(norm(DROP.sub('', r[2])).split()) for r in rows}
N = len(T)
df = collections.Counter(x for s in T.values() for x in s)
idf = {x: math.log(N / v) for x, v in df.items()}
normed = {r[1]: norm(DROP.sub('', r[2])) for r in rows}


def wj(a, b):
    inter = a & b
    u = a | b
    den = sum(idf[x] for x in u)
    return sum(idf[x] for x in inter) / den if den else 0.0


def q(vals, p):
    vals = sorted(vals)
    return vals[int(p * (len(vals) - 1))]


for lo in (1, 1300):
    top1 = []
    for n, cid, text in rows:
        if n < lo:
            continue
        best = 0.0
        for pn, pid, _ in rows:
            if pn >= n or normed[pid] == normed[cid]:
                continue
            s = wj(T[cid], T[pid])
            if s > best:
                best = s
        top1.append(best)
    m = len(top1)
    print('idf-weighted, window ids >= t%d: cards %d: top-1 median %.3f P75 %.3f P90 %.3f P95 %.3f P99 %.3f max %.3f' % (
        lo, m, q(top1, .5), q(top1, .75), q(top1, .9), q(top1, .95), q(top1, .99), max(top1)))
    for th in (0.5, 0.4, 0.3, 0.2, 0.15, 0.1):
        c = sum(1 for v in top1 if v >= th)
        print('  floor %.2f: cards with >=1 neighbor %d (%.1f%%)' % (th, c, 100 * c / m))
```

### A.4 기록된 관계의 재현율 (`t1454_recall.py`)

```python
import sqlite3, unicodedata, re, math, collections
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
rows = []
for cid, text in con.execute("select id,text from items"):
    rows.append((int(cid[1:]), cid, text))
for cid, text in con.execute("select id,text from archived_items"):
    rows.append((int(cid[1:]), cid, text))
rows.sort()
num = {r[1]: r[0] for r in rows}


def norm(s):
    return ' '.join(unicodedata.normalize('NFC', s).split()).lower()


DROP = re.compile(r'^\[DROPPED\s*—\s*[^\]]*\]\s*')
T = {r[1]: frozenset(norm(DROP.sub('', r[2])).split()) for r in rows}
N = len(T)
df = collections.Counter(x for s in T.values() for x in s)
idf = {x: math.log(N / v) for x, v in df.items()}


def jac(a, b):
    i = len(a & b)
    u = len(a) + len(b) - i
    return i / u if u else 0.0


def wj(a, b):
    den = sum(idf[x] for x in a | b)
    return sum(idf[x] for x in a & b) / den if den else 0.0


pairs = []
for tbl in ('findings', 'archived_findings'):
    for s, r, rel, src in con.execute("select subject_id, related_id, relation, source from %s" % tbl):
        pairs.append((s, r, rel, src))
pairs = list({(s, r, rel, src) for s, r, rel, src in pairs})
print('recorded finding rows (distinct):', len(pairs), collections.Counter((rel, src) for _, _, rel, src in pairs))


def rank(measure, subject, floor):
    me = num[subject]
    sc = []
    for n, cid, _ in rows:
        if cid == subject or n >= me:
            continue
        sc.append((measure(T[subject], T[cid]), cid))
    sc.sort(reverse=True)
    return [c for s, c in sc[:3] if s >= floor]


usable = [(s, r, rel, src) for s, r, rel, src in pairs if s in num and r in num and num[r] < num[s]]
print('pairs where the related card is earlier than the subject (surfaceable at issuance):', len(usable), 'of', len(pairs))
for name, m in (('token-set Jaccard', jac), ('idf-weighted Jaccard', wj)):
    for floor in (0.0, 0.2, 0.3, 0.4, 0.5):
        hit = sum(1 for s, r, rel, src in usable if r in rank(m, s, floor))
        print('%s floor %.1f: recorded counterpart inside top-3: %d of %d (%.0f%%)' % (name, floor, hit, len(usable), 100 * hit / len(usable)))
```

### A.5 경로·구성요소 커버리지 (`t1454_paths.py`)

```python
import sqlite3, re, collections, itertools
DB = '/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e45eb4da-ebe2-4e77-9359-5646d0ac07d3/scratchpad/t1454-queue-copy.db'
con = sqlite3.connect(DB)
live = [(r[0], r[1], r[2]) for r in con.execute("select id,text,state from items")]
arch = [(r[0], r[1], 'archived') for r in con.execute("select id,text from archived_items")]
# the path regex the existing triage helper uses (todo_triage.go todoTriagePathPattern)
PATH = re.compile(r'\b([a-z_]+/[a-z0-9_./-]+\.(?:go|md|yaml|sh))\b')
# a wider form: any path-like token with an extension, to bound the narrow form's recall
WIDE = re.compile(r'(?<![A-Za-z0-9_./-])((?:[.A-Za-z0-9_-]+/)+[A-Za-z0-9_.-]+\.(?:go|md|yaml|yml|sh|json|js|toml|templ|py|tmpl))')


def comp(p, depth=2):
    return '/'.join(p.split('/')[:depth])


def stats(name, cards):
    n = len(cards)
    narrow = sum(1 for _, t, _ in cards if PATH.search(t))
    wide = sum(1 for _, t, _ in cards if WIDE.search(t))
    npaths = [len(set(WIDE.findall(t))) for _, t, _ in cards if WIDE.search(t)]
    npaths.sort()
    med = npaths[len(npaths) // 2] if npaths else 0
    print('%s: %d cards; narrow(todoTriagePathPattern) names >=1 path: %d (%.0f%%); wide path form: %d (%.0f%%); median distinct paths among those: %d; max %d' % (
        name, n, narrow, 100 * narrow / n, wide, 100 * wide / n, med, max(npaths) if npaths else 0))


stats('all cards', live + arch)
stats('live non-dropped', [c for c in live if c[2] != 'dropped'])
stats('live open (queued+picked+hold)', [c for c in live if c[2] in ('queued', 'picked', 'hold')])
open_ = [c for c in live if c[2] in ('queued', 'picked', 'hold')]
for depth in (2, 3):
    keys = {cid: {comp(p, depth) for p in set(WIDE.findall(t))} for cid, t, _ in open_}
    pairs = [(a, b) for a, b in itertools.combinations(keys, 2) if keys[a] & keys[b]]
    withkey = sum(1 for k in keys.values() if k)
    print('open cards depth-%d component keys: cards with >=1 key %d of %d; pairs sharing >=1 key %d of %d' % (depth, withkey, len(keys), len(pairs), len(keys) * (len(keys) - 1) // 2))
fk = {cid: set(WIDE.findall(t)) for cid, t, _ in open_}
fpairs = [(a, b, sorted(fk[a] & fk[b])) for a, b in itertools.combinations(fk, 2) if fk[a] & fk[b]]
print('open cards pairs sharing >=1 exact path:', len(fpairs))
for a, b, s in fpairs[:8]:
    print('  ', a, b, s[:3])
cnt = collections.Counter()
for _, t, _ in live + arch:
    for p in set(WIDE.findall(t)):
        cnt[p] += 1
print('most-named paths across all cards:', cnt.most_common(8))
```
