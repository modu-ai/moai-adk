# SPEC-TODO-CARD-ISSUANCE-001 — M0 기준선 기록

card-head: 4bc76ceb37f0c2c8e666c37686ed50da8b80b5f6 (측정을 시작한 시점의 카드 브랜치 끝)
measured-at: 2026-10-05 (큐 스냅숏 /tmp/t1454-m0/queue-snapshot.db, 11:17 KST 고정)
develop-tip: 4964d0796 (카드 브랜치가 흡수한 develop 팁 — git 이력 기반 figure 의 tree)
snapshot-sha256: 0169ff357eec31d5cfabb519bb1bcef43306b4fe79d253485ab560f0bd6e7d6a

측정 큐: live 129(dropped 90·hold 10·picked 8·queued 21) + archived 1102 = 1231 행, last_seq 1520.
측정 커밋: first-parent 1357 중 카드 id 부착 1239(가지 본문 대체 146).

## git 이력 기반 figure (기준 커밋 4964d0796)

figure: GB01
command: scripts/03_git_collect.py (fp.tsv = git log --first-parent --since=2026-08-13 --until=2026-10-05T12:00:00+09:00 --format='%H%x09%P%x09%cI%x09%s' 4964d0796)
output: 1357 first-parent 커밋 중 카드 id 부착 1239(가지 본문 대체 146)
tree: 4964d0796

figure: GB02
command: scripts/04_granularity.py (commits.json × 큐 스냅숏)
output: first-parent develop 커밋을 가진 카드 1045 중 큐 스냅숏에 존재 948; archived 1102 중 착지 커밋 발견 932(85%)
tree: 4964d0796
drift: research §3.1 GB02(964/큐 867) 대비 성장 — 카드·커밋 수가 모두 늘었고 10-04~10-05 배치가 착지한 것

figure: GB03
command: scripts/04_granularity.py
output: 제품 줄 [min, Q1, median, Q3, P90, max] = [0, 30, 163, 593, 1799, 202434]
tree: 4964d0796
drift: P90 1604 → 1799 (분포 성장)

figure: GB04
command: scripts/04_granularity.py
output: 제품 파일 [0, 1, 3, 9, 22, 880]
tree: 4964d0796
drift: P90 21 → 22

figure: GB05
command: scripts/04_granularity.py
output: 제품 줄 0(process-only delivery): 117 (11%)
tree: 4964d0796

figure: GB06
command: scripts/04_granularity.py
output: 제품 줄 50 미만: 331 (32%)
tree: 4964d0796

figure: GB07
command: scripts/04_granularity.py
output: 제품 줄 150 미만: 503 (48%)
tree: 4964d0796

figure: GB08
command: scripts/04_granularity.py
output: 제품 파일 3개 이하: 543 (52%)
tree: 4964d0796

figure: GB09
command: scripts/04_granularity.py
output: 50줄 미만이면서 파일 3개 이하: 299 (29%)
tree: 4964d0796

figure: GB10
command: scripts/04_granularity.py
output: 공정 산출물 줄 2030261 / 제품 줄 1405668 = 공정 비중 59%
tree: 4964d0796
drift: 63% → 59% (제품 쪽이 더 많이 늘었다)

figure: GB11
command: scripts/04_granularity.py (48h same-product-file-set pairs)
output: 48시간 내 같은 제품 파일 집합 쌍 33(카드 53)
tree: 4964d0796

figure: GB12
command: scripts/04_granularity.py (48h Jaccard>=0.5 pairs)
output: 48시간 내 파일 집합 Jaccard ≥ 0.5 쌍 102(카드 133, 13%)
tree: 4964d0796

figure: GB13
command: scripts/04_granularity.py (same primary file chained within 72h)
output: 같은 주(主) 파일이 72시간 안에 이어진 묶음 70개, 카드 163장(16%), 절약 가능 레인 세션 93(9%)
tree: 4964d0796

figure: GB14
command: scripts/04_granularity.py (stricter subset)
output: 엄격형(같은 주 파일 + 한 카드가 다른 카드를 인용) 하위 묶음 37개, 절약 45(4%)
tree: 4964d0796

figure: GB15
command: scripts/04_granularity.py (hottest product files, distinct cards; max in any 72h window)
output: internal/template/catalog.yaml 121(20), internal/config/defaults.go 64(15), internal/cli/todo.go 42(10), internal/web/assets/i18n.js 41(10), .claude/rules/moai/workflow/kanban-dispatch.md 39(10), internal/config/types.go 41(9)
tree: 4964d0796

figure: GB16
command: scripts/06_followup_overlap.py
output: 파생 카드 중 자식·부모 모두 착지한 356: 자식 발행이 부모 착지보다 앞선 233(65%), 부모 착지 후 24시간 내 105(29%); 같은 제품 파일 공유·±48시간 157(44%)
tree: 4964d0796
drift: 연구값 321/67%/28%에서 모수 성장

## 큐 스냅숏 기반 figure (기준: card-head 4bc76ceb3 + snapshot 11:17 KST)

figure: QB01
command: scripts/01_inventory.py
output: 큐 행 1231 = live 129 + archived 1102, last_seq 1520
tree: 4bc76ceb3

figure: QB02
command: scripts/01_inventory.py
output: live 상태: dropped 90, hold 10, picked 8, queued 21 — 열린 카드(hold+picked+queued) 39
tree: 4bc76ceb3

figure: QB03
command: scripts/01_inventory.py
output: archived 1102: 상태 picked 680·queued 422, landing 값이 있는 것 110
tree: 4bc76ceb3

figure: QB04
command: scripts/05_dups_stale_lanes.py (findings section)
output: findings: live 16 + archived 137; (relation, source): near-duplicate/jev 133, contains/agent 10, absorbs/agent 5, conflicts/agent 3, replaces/agent 2; 서로 다른 (쌍, 관계) 153
tree: 4bc76ceb3

figure: QB05
command: scripts/02_origin_graph.py
output: 파생 카드(머리말이 이전 카드를 인용하고 파생 키워드 포함) 490 (39.8%)
tree: 4bc76ceb3

figure: QB06
command: scripts/02_origin_graph.py
output: 파생 깊이 분포 [(0,741),(1,300),(2,116),(3,51),(4,20),(5,3)]; depth ≥ 2 = 190 (15.4%), ≥ 3 = 74 (6.0%)
tree: 4bc76ceb3

figure: QB07
command: scripts/t1454_jaccard.py
output: 정규화 본문(drop 접두사 제거) 동일 그룹 61(카드 122); 저장된 그대로 53그룹(106장)
tree: 4bc76ceb3

figure: QB08
command: scripts/05_dups_stale_lanes.py (near-duplicate scores)
output: Jev near-duplicate 133건, 점수 [min 0.28, Q1 0.61, med 0.77, Q3 0.88, max 0.99], 0.8 이상 61건; 주체 카드 결말 done 47·queued 9·dropped 3·hold 2
tree: 4bc76ceb3

figure: QB09
command: scripts/01_inventory.py + 05 (field fill)
output: archived_at 보관 131 / 1102, dropped_at 1 / 90, picked_at 7 / 8 — 과거 카드의 닫힘 시각은 대부분 없다
tree: 4bc76ceb3

figure: QB10
command: scripts/05_dups_stale_lanes.py (lanes proxy)
output: 동시 진행 대리 지표 — 하루(UTC) 중앙 20·P90 56·최대 76; 3시간 구간 중앙 5·P90 16·최대 41; 스냅숏 시점 picked 8
tree: 4bc76ceb3

## 이 계획의 새 측정 (발행 시점 제시 설계 근거, 기준: card-head + snapshot)

figure: SB01
command: scripts/t1454_jaccard.py
output: 비-정확 근접 이웃이 기계 near 구간 [0.80,1.0) 에 드는 카드 0장 — ≥0.8 인 122장은 전부 정확히 1.0
tree: 4bc76ceb3

figure: SB02
command: scripts/t1454_issuance_sim.py (window ids >= t1)
output: 발행 시점 가정 top-1 비정확 Jaccard: median 0.102, P75 0.138, P90 0.230, P95 0.460, max 0.750; 최근 창(id ≥ t1300, 221장): median 0.095, P90 0.161, P95 0.250
tree: 4bc76ceb3

figure: SB03
command: scripts/t1454_issuance_sim.py
output: 표시 하한별 이웃 보유율(전체/최근 창): 0.6 → 1.9%/4.1% · 0.5 → 3.0%/4.1% · 0.4 → 5.8%/4.1% · 0.3 → 8.1%/4.1% · 0.2 → 11.2%/6.8%
tree: 4bc76ceb3

figure: SB04
command: scripts/t1454_idf_sim.py
output: idf 가중 top-1: median 0.051, P90 0.162, P99 0.621; 하한별 보유율 0.30 → 5.6%, 0.20 → 8.8%, 0.10 → 15.4%
tree: 4bc76ceb3

figure: SB05
command: scripts/t1454_recall.py
output: 기록된 관계 153쌍 중 상대가 먼저 발행된 142쌍의 상위-3 재현율 — Jaccard: 무하한 50%, 0.2 → 11%, 0.3 → 5%, 0.4 → 4%; idf: 무하한 62%, 0.2 → 5%
tree: 4bc76ceb3

figure: SB06
command: scripts/t1454_paths.py
output: 본문에 경로가 보이는 카드 — 전체 554/1231 (45%), 열린 카드 17/39 (44%), 넓은 정규식 15/39 (38%); 경로 있는 카드의 서로 다른 경로 수 median 1, max 8
tree: 4bc76ceb3

figure: SB07
command: scripts/t1454_paths.py
output: 열린 카드 쌍(741쌍) 구성요소 키 공유 — 깊이 2: 13쌍, 깊이 3: 0쌍; 정확히 같은 경로를 공유하는 쌍 0
tree: 4bc76ceb3

figure: SB08
command: (M1 이 진행 중 레인 수·탐침 지연 분포를 재측정 — 계획 F.11)
output: 미측정 — git 탐침 비용은 카드 시절 약 0.25초/브랜치(95파일)였고 브랜치는 422로 늘었다
tree: 4bc76ceb3

figure: SB09
command: git grep -l '^card:' -- '.moai/specs/*/spec.md' | wc -l (와 ls .moai/specs | wc -l)
output: SPEC 디렉터리 1050 중 card: 가 있는 것 36
tree: 4bc76ceb3

## 개수 figure (측정 시각의 값 — 귀속 줄에 적지 않는다)

figure: CT01
command: git branch --list 'WT-*' | wc -l
output: 422
tree: (저장소 전역 ref — 귀속 없음, 측정 2026-10-05)

figure: CT02
command: git worktree list | wc -l
output: 149
tree: (저장소 전역 ref — 귀속 없음, 측정 2026-10-05)

figure: CT03
command: ls .moai/specs | wc -l
output: 1050
tree: (작업 트리 디렉터리 수 — 귀속 없음, 측정 2026-10-05)

figure: CT04
command: git grep -l '^card:' -- '.moai/specs/*/spec.md' | wc -l
output: 36
tree: (작업 트리 — 귀속 없음, 측정 2026-10-05)

## 임계값 (F.11 작업 기본값의 재도출 — 근거 figure 는 이 기록의 행)

| 항목 | 재도출 값 | 근거 |
|---|---|---|
| 유사 카드 상한 | 3 | 카드 본문이 정한 값(고정) |
| 표시 하한(token-set Jaccard) | 0.30 — 전체 8.1%, 최근 창 4.1%에 알림; 기록 관계 재현율 5%(0.3) | SB03, SB05 |
| 구성요소 깊이 | 경로 앞 두 마디 — 깊이 3 공유 쌍 0 | SB07 |
| 카드 크기 하한 트리거 | 예상 제품 줄 50 미만·제품 파일 3개 이하(29%) | GB09 |
| 카드 크기 상한 | 제품 줄 1799 초과 또는 제품 파일 22 초과(새 P90) | GB03, GB04 |
| 파생 깊이 상한 | 2 — depth ≥ 3 은 74장(6.0%) | QB06 |
| 동시 진행 한도 | 16 — 3시간 구간 P90(중앙 5, 최대 41) | QB10 |
| 허브 임계값 | 72시간 창 최대 겹침 ≥ 10 — catalog.yaml(20)·defaults.go(15)·todo.go(10)·i18n.js(10)·kanban-dispatch.md(10) 5파일 | GB15 |
| git 탐침 시간 상한 | M1 이 재측정해 정한다(브랜치 422로 증가) | SB08 |
| 웹 그래프 노드 상한 | 관계 명명 카드 214 + 열린 카드 39 의 합집합 규모에서 M6 가 정한다 | QB04, QB02 |
