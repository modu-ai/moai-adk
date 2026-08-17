# t113 Evidence — Chain session 보드 review 컬럼 제거·sync 통합

Worktree: `.claude/worktrees/t113` (branch `WT-t113`, base = origin/release/v3.1.1 @ d169c4aec — t60 머지까지 포함, fast-forward)

## Claim (주장)

1. **ColumnReview 제거·5값 폐쇄 집합** — `internal/kanban/column.go`: `ColumnReview` 상수·`allColumns` 항목 제거, 6→5 (`backlog → plan → run → sync → done`), 헤더/생성자/에러 문자열 주석 전면 갱신 ("no sixth value"). `ParseColumn`은 이제 `review`를 거부.
2. **reconcile 짝 판정 갱신** — `internal/kanban/reconcile.go`: `pairingConsistent`의 StatusInProgress 분기 `run || review || sync` → `run || sync`; 호환 표 주석과 `statusDecidesMultipleColumns` 주문("run and review" → "run and sync") 갱신.
3. **보드 UI 역할 목록** — `internal/web/viewmodel_ops.go`: `ChainRoles`에서 `"review"` 제거 (`lead/plan/run/sync` 4역할), 278행 주석에서 review 문구 제거. 뷰 B(`pipelineColumns`)는 원래 review 없음 — 웹 스크린샷의 review 렌더는 ChainRoles(뷰 A)가 유일 소스였음을 실측으로 확인 (templ·i18n에 보드 review 문자열 grep 0건 — role명은 로케일 키가 아니므로 **i18n 4로케일 동반 대상 없음**).
4. **테스트 갱신 3건** — `column_test.go` (열거 6→5·순서·함수명, `review`/`reviewed`를 ParseColumn reject 목록에 추가, HasOwningSession true 목록), `admission_test.go` (round-trip fixture ColumnReview→ColumnSync), `reconcile_test.go` (in-progress 충돌쌍 run/review→run/sync).
5. **kanban-dispatch.md 6→3작업컬럼 정리** (stub+companion+t템플릿 미러, 전부 byte-identical):
   - stub: The board 5값·3 working·"review verdict absorbed by the sync gate" 문장; Scope 역할 목록에서 review 제거; 카드 클래스(A: plan만 skip / B: `run → sync` / C: 3 working columns); "it takes the review column" 문장 제거; B의 "skips plan, not review" → "not the sync gate's review"; dispatch 사이클 화살표; lens 필드 "sync dispatch"; CodeRabbit "does not leave review or sync"→"sync"; Boundaries "six-column state"→"board state".
   - companion: board 표 review 행 제거 + sync 행에 "Review verdict (lenses per card)" 명시; Terminology column 정의(5값·3 working·review는 sync 게이트 내부)·companion 예식에서 review-a1b2c3 제거; 카드 클래스 표; 사이클 다이어그램; Review lens selection 절 첫문장 sync gate화; 병렬성 문단 4→3; Class-A 문단 "Handing three sessions…through three columns".
   - Review lens 선택·CodeRabbit 판정 읽기는 **삭제하지 않고 sync 게이트의 파라미터/절차로 재배치** (D1: 판정=sync 흡수, 렌즈=sync 디스패치 파라미터).
6. **러더이더 1 [t107 규율 산문]** — stub에 "## Report milestones ↔ queue cards" [HARD] 절 신설 (~110토큰, t107 paste-ready에서 카드 id 제거·중립화); 절차·근거(파서·기계 쿼리·git 판정 절차)는 companion 동명 절로 배치 — 리드 지시 "규칙성은 stub, 절차·근거는 companion" 구조 적용.
7. **러더이더 2 [t107 소급 대조표]** — primary `.moai/reports/unified-board-design-20260817.md` §7 뒤 §8 앞에 "## Card Cross-Check" 절 삽입. **리드 2차 지시 스펙 반영**: S6 card칸 `t110` 단독 + 실측칸 "재발행 t84→t110 — 머지 6b44bdd2e"; 요약 "8개 전 매핑 완결(신규 0)"; 재발행 계보 주석(t84→t110·t105→t106·t38→t111); caveat에 파서 한계(card칸 → 화살표 금지)와 "not in live queue"의 완결/미발급 구분 불가 명시. 삽입 직전 primary 큐 재측정(t108·t113 queued)과 t59 머지(d0f946d62, 본 세션 관측) 반영.
8. **러더이더 3 [t59 인계]** — companion Terminology·본계의 "여섯 칸/6컬럼" 표현 전부 5값·3작업컬럼 기준으로 갱신 (5번에 포함).

## Evidence (증거) — 이번 실행 관측

```
$ go build ./internal/kanban/ ./internal/web/ && go vet ./internal/kanban/ ./internal/web/  → rc=0
$ go test ./internal/kanban/ -count=1        → ok  15.662s
$ go test ./internal/web/ -count=1 -timeout 300s  → ok  2.607s
$ make build                                 → catalog.yaml 재생성(내용 HEAD와 동일 — git diff 0), bin/moai 빌드
$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -count=1  → ok 0.460s
$ go test ./internal/template/ -run 'Mirror|Parity|Leak' -count=1          → ok 1.086s
$ diff -q <본계 kanban-dispatch.md> <템플릿>       → 동일
$ diff -q <본계 kanban-dispatch-detail.md> <템플릿> → 동일
$ grep -rn "ColumnReview" internal/ --include="*.go"  → 0건 (잔여 없음)
```

하위호환 관측: `LoadBoard`(`internal/kanban/board.go:104`)는 JSON unmarshal만 하고 컬럼 값을 검증하지 않음 — 기존 보드 상태 파일에 `review` 값이 남아 있어도 로드는 되고 `reconcile.go:121`의 `Valid()` 체크에서 해당 카드가 inconsistent로 **보고**된다(파탄 아님). 마이그레이션 코드 불필요 — 낡은 값은 카드가 다음 칸으로 옮겨질 때 새 값으로 기록된다.

## Baseline-attribution

모든 커맨드 출력은 본 워크트리(HEAD d169c4aec + 본 카드 미커밋 변경)에서 이번 실행으로 관측. 대조표의 큐 실측은 primary 삽입 직전(2026-08-17 20:1x) `moai todo` 관측, 머지 SHA는 본 세션 통합 이력 + t107 evidence.md 기록 귀속.

## Gaps (미검증)

- **대조표 기계 검증 불가** — primary 설치 바이너리에 `graph` 명령 부재(`Unknown command "graph"` 관측): t107이 아직 release에 미머지. 형식 준수는 리드 지시 스펙(card칸 단일 id·계보 실측칸) 수동 적용. t107 머지 후 `moai graph build && moai graph query --milestones-no-card` 재실행 권장.
- 웹 보드 화면의 실제 렌더 확인(스크린샷 대조)은 로컬에서 안 함 — ChainRoles가 유일 소스임을 코드로 실측했으나 시각 확인은 아님. CI·운영 확인 몫.
- `go test ./internal/cli/`(kanban 소비처 중 일부) 미실행 — lane-local 규율(변경 패키지 중심)에 따라 kanban·web·config·template만. 전체 스위트는 CI.
- 대조표 삽입본의 커밋 — primary untracked 보고서 파일이라 커밋하지 않음(리드/운영자 소유 산물). 삽입 사실은 본 evidence와 리드 보고로 귀속.

## Residual-risk

- 5컬럼 열거가 남긴 서술 부채: docs-site kanban 관련 페이지(ko/en/ja/zh)·README.ko "여섯 칸"·t59 용어집 4로케일의 6컬럼 표기는 본 카드에서 **갱신하지 않았다** — 카드 원문 범위(kanban-dispatch.md·코드·보드 UI) 밖이며 docs-site/README는 별도 4로케일 의무 표면. 후속 카드 권장 (t98·t32와 같은 파일).
- kanban-mode.md(docs-site)의 "체인의 네 단계(plan → run → review → sync)" 서술도 동일 부채.
- 기존 보드 상태 파일의 review 잔값은 reconcile 보고에서 inconsistent로 드러나지만, 운영자가 방치하면 대시보드에 계속 표시될 수 있음 — 운영 안내 권장.
