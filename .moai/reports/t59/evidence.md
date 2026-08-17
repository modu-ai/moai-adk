# t59 Evidence — 칸반/팀 모드 용어 정의 + 문서화

Worktree: `.claude/worktrees/t59` (branch `WT-t59`, base `931e4138a`)

## Deliverables

1. **(1) 내부 용어집** — `kanban-dispatch-detail.md` (lazy companion, `paths:`-scoped) 맨 앞에 "Terminology — the board vocabulary" 절 추가. 9용어 (lane, card, column, backlog, lead, companion, run-id, worktree, dispatch) 각 한 줄 정의 + 예시 1개. lane/column 구별 문단 포함.
   - 로컬 `.claude/rules/moai/workflow/kanban-dispatch-detail.md` + 템플릿 `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` — byte-identical (`diff -q` MIRROR-IDENTICAL 관측).
   - always-loaded 본계 (`kanban-dispatch.md`) 미수정 — 예산 여유 ≤3토큰 경계(t104/t113) 존중.
   - 예시 식별자는 중립값(`a1b2c3`/`t0`) — 내부 운영 식별자(t59/tjv7iy) 미사용, template-neutrality 배려. `a1b2c3`은 본계가 이미 쓰는 예시 스타일과 동일.
2. **(2) docs-site 개념 페이지** — `docs-site/content/{ko,en,ja,zh}/core-concepts/kanban-board-terms.md` 신규 4파일 (ko 정본 → en/ja/zh 파생). 구조: 보드 다이어그램(mermaid TD) → 9용어 표 → lane vs column 절(mermaid TD) → 카드 한 장의 여정 → 관련 문서.
   - `_meta.yaml` 4로케일 등록 (title + weight 40), `data/menu/main.yaml` 사이드바 항목 추가 (trust-5 ↔ verification-claim-integrity 사이).
   - 기존 번역어 일관성 유지: 리드 세션/동반 세션(kanban-mode.md 관례), 링크는 사이트 내부 경로 + adk.mo.ai.kr만.
3. **(3) README.ko.md 요약** — "### 다섯 세션이 쓰는 말" 소절 (보드 규칙 문단 뒤, "보드를 눈으로 보기" 앞). 텍스트 그림(보드 흐름 + 레인 2개 병렬) + 9용어 한 줄 표 + docs-site 용어집 링크. **ko만** — en/ja/zh는 t47 재파생 체인에 맡김 (리드 범위 조정 확정).

## Claim / Evidence / Baseline-attribution

| Claim | Evidence (command → observed output) |
|---|---|
| hugo 빌드 warning-free | `hugo -s docs-site --minify --gc` → exit 0, 테이블 출력 (KO 179 / EN·JA·ZH 177 pages); WARN/ERROR grep 카운트 `0` |
| sitemap 생성 | `test -f docs-site/public/sitemap.xml` → `sitemap OK` |
| URL 블랙리스트 위반 0 | `grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content README.ko.md` → 매치 없음 |
| Mermaid TD-only 위반 0 | `grep -rn 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content` → 매치 없음 (신규 페이지 다이어그램 2개 모두 `flowchart TD`) |
| 신규 페이지 4로케일 섹션 패리티 | `grep -c '^#\{2,\} '` → ko/en/ja/zh 각 `5` — 균일, NEW divergence 0 |
| 신규 페이지 본문 이모지 0 | emoji 스캔 → 4파일 매치 없음 (README.ko 179행 💡는 기존 라인, 본 카드 변경분 아님·README는 docs-site 본문 규칙 범위 밖) |
| 템플릿 미러 byte-identical | `diff -q` 로컬↔템플릿 → `MIRROR-IDENTICAL` |
| mirror/parity/leak 테스트 통과 | `go test ./internal/template/ -run 'Mirror|Parity|Leak' -count=1` → `ok … 1.261s` |
| always-loaded 예산 무영향 | `go test ./internal/config/ -run TestAlwaysLoadedTokenBudget -count=1` → `ok … 0.438s` (paths-scoped 파일이라 예산 슬롯 밖 — 관측으로 확인) |
| 템플릿 재빌드 | `make build` → catalog.yaml 갱신 무변동(해시 동일), `go build … bin/moai` 완료 |
| README H2 패리티 무변동 | `grep -c '^## '` → ko 11 vs en/ja/zh 14 — **기존 격차 그대로** (t47이 지적한 ko 포크 부채; 본 카드 추가분은 H3 소절이라 카운트 불변) |

Baseline-attribution: 위 커맨드는 전부 본 워크트리(`.claude/worktrees/t59`, HEAD `931e4138a` + 본 카드 미커밋 변경)에서 이번 실행에 관측한 출력이다.

## Gaps (미검증)

- 전체 143×4페이지 래칫(`.locale-parity-baseline` 대조 awk 파이프라인)은 미실행 — 워크트리 guard가 복합 파이프라인 거부. 대체 관측: 신규 페이지 4개 균일(5/5/5/5) + `git status`로 기존 페이지 0개 수정 확인 → NEW divergence 발생 경로 부재. CI의 M6 4-locale 게이트가 최종 판정.
- 4로케일 페이지의 실제 렌더링(사이드바 노출·new-badge)은 로컬 hugo 빌드로 페이지 생성까지만 확인 — 브라우저상 육안 확인 안 함.
- docs-site 배포(preview/production)는 미실행 — push 금지(배차 지시), Vercel 배포는 human-gated.

## Residual-risk

- `t113`(보드 review 칸 제거·3컬럼 통합)이 착지하면 "여섯 칸" 서술(용어집·README·docs-site 4로케일 전부)이 함께 갱신돼야 한다 — t113이 kanban-dispatch.md 템플릿 미러 동반을 이미 명시하므로, t59 산물의 column 서술도 t113 편성 시 동반 갱신 권장.
- `t47`(README ko 신골격 승격 → en/ja/zh 재파생)이 진행되면 본 카드의 README ko 소절이 en/ja/zh 파생의 원천이 된다 — 소절 제목("다섯 세션이 쓰는 말")은 ko 골격 특유이므로 t47 파생 시 현지화 필요.
- ko `_meta.yaml`과 `main.yaml`의 기존 항목 순서 어긋남(constitution 위치 등)은 기존 부채(t32 범위) — 본 카드는 신규 항목만 verification 직전으로 양쪽 정렬.

## Rider (리드 판정 후 — zh 용어 통일)

허브 리뷰 판정 PASS(2026-08-17)에 advisory 1건이 러더이더로 채택됨: zh 용어집이 기존 `kanban-mode.zh.md` 용어와 어긋났던 것. 치환 전 기존 용어를 직접 관측해 리뷰 주장을 검증 (`grep -n '主导会话\|伴随会话' …/kanban-mode.zh.md` → L18·34·106·127·128·131·142·170에서 해당 용어 확인).

| 치환 | 발생 | 근거 |
|---|---|---|
| `主控` → `主导会话` | 5회 | kanban-mode.zh.md L18/L34/L106 등 기존 lead 역 번역 |
| `同伴会话` → `伴随会话` | 9회 | kanban-mode.zh.md L18/L127/L128 등 기존 companion 번역 |
| `同伴角色` → `伴随角色` | 2회 | 리드 지시 범위 밖이나 판정 원리(이중 용어 방치 불가)에 따라 같은 어근 통일 |

재검증(치환 후 이번 실행 관측): 잔여 구 용어 `grep -c '主控\|同伴'` → `0`; `hugo -s docs-site --minify --gc` → exit 0, WARN/ERROR 0라인 (빌드 영향 없다는 리뷰 주장 재현).
