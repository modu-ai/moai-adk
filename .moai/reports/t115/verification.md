# t115 — 검증 리포트 (보드 6컬럼→5컬럼 문서 동기화)

카드: t115 — t113 착지분 정합 (보드 5컬럼, review는 sync 게이트 흡수)
워크트리: `.claude/worktrees/t115` (branch `WT-t115`, base `47963ec0b` = origin/release/v3.1.1)
커밋: `99b1e5638` (17 files, +297/-212)

## 1. Claim (주장)

1. 코드 사실 확정: 보드는 5컬럼(`internal/kanban/column.go` 닫힌 열거, "review removed per t113"), 동반 역할은 plan/run/sync 3개(`bootstrap.go` `CompanionRoles`, "t97 retired the review role"), 동반 이름은 bare role·run-id는 리드 전용(t56, `bootstrap.go` 네이밍 정책 주석).
2. README 4로케일의 칸반 절이 5컬럼 체계로 정합됐다 (구 6컬럼·review 세션·`<role>-<run-id>` 서설 제거).
3. docs-site kanban-mode 4로케일 + kanban-board-terms 4로케일이 같이 정합됐다 (다섯 컬럼·세 단계·mermaid review 노드 제거·t56 네이밍).
4. companion(kanban-dispatch-detail.md) Terminology의 t56 낡음 4곳이 정합됐고 템플릿 미러는 byte-identical, `make build` 완료. 6컬럼 잔존은 companion에 없었음(실측).
5. `kanban-five-sessions.png`가 새 SVG 소스에서 재생성됐다 (소스 없음 → 재작성 경로, SVG 소스 보존).

## 2. Evidence (증거)

코드 사실 (준거 측정):
- `internal/kanban/column.go:20-26` — `ColumnBacklog/Plan/Run/Sync/Done` 5상수, "review removed per t113"
- `internal/kanban/bootstrap.go:37` — `CompanionRoles = []string{"plan", "run", "sync"}`, "D1 (card t97) retired the review role"
- `internal/kanban/bootstrap.go:14-18` — t56 네이밍 정책("companion names carry no run id")

검증 명령·출력 (커밋 99b1e5638 트리):

```
$ grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md   → 12/12/12/12
$ grep -c '^```' ...                                                 → 42/42/42/42
$ wc -l ...                                                          → 755/755/755/755
$ grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' README*.md docs-site/content/ → 0건
$ grep -rn 'flowchart LR\|flowchart RL\|graph LR\|graph RL' README*.md docs-site/content/    → 0건
$ cd docs-site && hugo --minify --gc   → Total in 4288 ms, exit 0 (경고 0)
$ ls docs-site/public/sitemap.xml      → 존재 (572B)
$ diff <companion> <mirror>            → MIRROR-IDENTICAL (stub도 IDENTICAL)
$ wc -c companion                      → 13,575B (+111B vs 13,464B — 1,000B 성장 기술 게이트 미달)
```

잔존 스윕 (전 표면): `review-<run-id>`·`<role>-<run-id>`·six-column·여섯 칸·6列·六列 → 0건.

PNG 파이프라인 (소스 없음 → 재작성):
- SVG 소스 신규 작성 `assets/images/kanban-five-sessions.svg` (수치 레이아웃 테이블 주석 포함, moai-domain-svg-infographic 스킬 절차)
- lint: `check-svg.mjs` → **0 errors, 1 warning**(SVG030 :104 footnote — 휴리스틱이 직전 박스에 소속시킨 오판; 렌더 PNG에서 footnote 오버플로 없음 확인 → 노이즈로 트리아이 종결)
- 렌더: `render.mjs` → browser `/Applications/Google Chrome.app` **Chrome 151.0.7922.138** (headless=new), target 3200x1760(2x), **PNG IHDR 3200x1760 일치** verified, exit 0
- 배포: `assets/images/kanban-five-sessions.png` (203,681B) + `docs-site/static/images/profile/` 동일 사본
- 육안: 5열 보드·리드 밴드·동반 3행·review gate 배지·화살표 전부 의도 렌더링 확인

## 3. Baseline 귀속

모든 측정은 워크트리 `.claude/worktrees/t115` HEAD `99b1e5638`에서 이 세션이 직접 실행. 코드 사실은 base `47963ec0b`(origin/release/v3.1.1) 체크아웃의 실제 소스에서 grep/Read로 관측.

## 4. Gaps (미검증)

- **docs-site 배포 프리뷰 미확인**: hugo 빌드는 로컬 clean이나 Vercel 프리뷰 렌더링(실제 페이지에서 그림·mermaid 표시)은 push 후에만 관측 가능 — 통합 후 프리뷰 URL 확인 권장.
- **`-k --name plan` 실측 미실행**: t56 네이밍 정책은 코드 주석·타입으로 확인했으나 실제 런처 실행(`moai cc -k --name plan` 부트스트랩 안내 출력)은 dev 프로젝트에서 금지(CLAUDE.local.md §13)라 문서 서술은 코드 근거만으로 정합.
- **본선 README 외 번역 사이트**: 외부 커뮤니티 번역(있다면)의 6컬럼 서술은 범위 밖.
- **`internal/kanban/reconcile.go:44` 주석** "run vs review" 잔존 — reconcile 로직 주석으로 카드 범위(문서 동기화) 밖. 코드 주석 정리는 별도 카드 후보로 기록.

## 5. 잔여 위험

- 그림 파일명 `kanban-five-sessions.png`의 "five"가 이제 **다섯 칸 보드**를 가리키도록 재해석됨(세션 수는 4). 파일명 개명은 4로케일·docs-site 참조 전파 비용이 커서 유지 — 개명 원하면 후속 카드.
- docs-site ko 정본 체인상 이번 8페이지는 4로케일 동시 갱신(로케일 패리티 의무 준수)이나, ja/zh 문체 품질은 사람 검수 몫.
- PNG가 기존 터미널 스크린샷 계열(1MB)에서 다이어그램(204KB)으로 교체됨 — 문서상 맥락(alt)은 새 그림에 맞게 갱신했으나, "실화면 스크린샷"을 기대한 독자에겐 성격이 달라짐.

## 산출물

- README 4로케일 (칸반 절 + 핵심 기능 각주)
- docs-site: advanced/kanban-mode ×4, core-concepts/kanban-board-terms ×4
- 룰: `.claude/rules/moai/workflow/kanban-dispatch-detail.md` (+템플릿 미러, make build 반영)
- 이미지: `assets/images/kanban-five-sessions.svg` (신규 소스) + `.png` (재생성) + docs-site static 사본

---

# 부록 — MUST-FIX 재작업 (허브 판정 반영, 커밋 a916c8137)

허브 판정(MUST-FIX, review-verdict.md): 1차 갱신분은 전건 재측 일치했으나 **docs-site 홈 `content/<locale>/_index.md` ×4에 6컬럼 서술 잔존** — 1차 증거의 "잔존 스윕 전 표면 → 0건" 주장 기각. 원인: 1차 스윕이 파일 나열 방식이라 `_index.md`가 대상에 없었음(본인 분석, 판정문 R2와 일치).

## 반영 내역

- **R1 (판정문 MUST-FIX 본체)**: `_index.md` ×4 — 보드 문장 5컬럼화 + review-in-sync-gate 절, 터미널 블록 bare-role 3컴패니언, 히어로 alt "다섯 칸 보드 + 리드 + 세 동반", moai web 문장 "칸반 체인".
- **R2 (교정 스윕 패턴 적발 추가 표면)**:
  - `multi-llm/kanban-mode.md` ×4 — 세 컴패니언, 체인 `plan -> run -> sync`, bare-role 네이밍 + t56 run-id 설명, 부트스트랩 다이어그램 3명령, SessionStart 개수 서술.
  - `advanced/moai-web-console.md` ×4 — 체인 바를 네 역할 `lead → plan → run → sync`로 (코드 실측 근거: `internal/web/viewmodel_ops.go:46` `ChainRoles = ["lead","plan","run","sync"]`, review 없음 — 낡은 문서가 코드와 어긋났던 것).

## 재검증 (커밋 a916c8137 트리)

```
$ grep -rn 'six columns|six-column|여섯 칸|여섯칸|6列|六列|→ review →|review-<run-id>|review-abc123|<role>-<run-id>' \
    README×4 + docs-site/content 전체 + kanban-dispatch 룰 2 + 템플릿 미러   → 0건
$ cd docs-site && hugo --minify --gc → Total in 4301 ms, exit 0 (경고 0)
```

## 갱신 Gaps (추가)

- **README의 `plan → run → verify → sync` 프리셋 단계 표기(×4)**: R2 패턴 불일치·판정문 미지적. 코드에 verify 스테이지 상수·프리셋 정의 없음(탐색했으나 원천 미발견) — 멋대로 범위 확장하지 않고 허브 판단으로 넘김. multi-llm의 동일 표기는 컴패니언 수 서술과 불가분이라 세 컴패니언 정합과 함께 수정했음(위 반영 내역).
- 1차 증거의 스윕 클레임 기각 경위: 스윕 대상을 "건드린 파일 나열"로 한정한 설계 결함 — 재작업 스윕은 디렉터리 전체 대상 + `→ review →` 2차 안전망 패턴으로 교정.
