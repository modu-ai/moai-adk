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
