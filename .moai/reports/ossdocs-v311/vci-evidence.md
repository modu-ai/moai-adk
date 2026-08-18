# ossdocs-v311 VCI Evidence — v3.1.1 문서 갭 착지 (oss-docs 2단계 + 추가 지시 4건)

Branch: WT-ossdocs-v311 @ base 4100d8767 (origin/main — release 전체 포함, 1bd9140eb 조상 확인)
Date: 2026-08-18
Session: db221a6c-e73f-4806-b60e-bc00af9ab6fa (run lane)

## 1. Claim (주장)

1. **본카드(13파일 × 4로케일 = 52파일)**: ko 드래프트 13파일(README.ko + 기존 9페이지 전체
   교체 + 신규 3페이지 graph/tokens/memory) 반영 → en/ja/zh 로케일 전문가 3인 순차 전면
   파생(agent-teams만 1문단 표적 미러 — cc234-align 직후 4로케일 동기화 페이지).
2. **구조 등록**: cli-reference `_meta.yaml` 4로케일에 graph/tokens/memory 3항목 추가 +
   `data/menu/main.yaml` 3항목(4로케일 이름, B6 선례 패턴). 리다이렉트 불필요(이동 없음).
3. **추가1 — 용어 정정(긴급)**: 드래프트의 "팩토리(모드)/factory mode" 전부 칸반 모드 소관의
   **"번호 붙은 워컌 런"(numbered-workers run)** 으로 치환 — ko 14곳·en 13곳 직접 수정,
   ja/zh는 파생 시 반영. "팩토리"는 `-f`/`--factory` 은퇴 각주·옛 이름 언급·코드 센티널
   `FACTORY_MODE_UNSUPPORTED_BACKEND`(internal/cli/factory.go:41 실재 상수 — 원문 유지)에만 남김.
   ※ 치환 지시를 드래프트 적용 전에 놓친 것이 발견되어 en 파생 후 전수 재치환+재검증으로 회복.
4. **추가2 — mermaid 감사 22파일**: fixes/ 전체 반영(감사 base와 동일 4100d8767, 13파일 세트와
   불일치 확인 후 적용). 감사 레시피 재실행: 잔여 `[/label` 패턴 0(매치는 전부 온전한 산문 링크),
   깨진 재작성 `["moai loop"](` 0·온전한 `[/moai loop](` 16파일 유지, hugo 무경고.
5. **추가3 — 버전 동기화**: hugo.toml `v3.1-rc.2`→`v3.1.1`(+releaseDate 2026-08-18 — 하단
   플래그 참조), README 뱃지 ×4 v3.1.1, statusline 예시 `🗿 v3.1.1` ×4, 업데이트 예시
   `🗿 v3.1.0 ⬆️ v3.1.1` ×4, FAQ(ko·en)·moai-feedback(ko) 예시값 갱신. 역사 서술("v3.0.0부터
   은퇴/도입/기본값" 류)은 보존.
6. **추가3 — 하네스 지침**: hns-oss-docs-i18n-rules §7에 릴리즈 동기화 의무 추가,
   hns-oss-docs-verify §6 "Version-string sync" 검사 신설 + 스코어맵 `version-sync`
   (must_pass 1.0) 행 추가.
7. **추가4 — 백엔드 조합 추천**: 커밋 00af8d334의 부트스트랩 추천(리드 glm / plan cc /
   run glm / sync cc)을 README 칸반 절 + kanban-mode 페이지에 서브섹션으로 4로케일 추가
   (레인별 근거·`judge` 탈출구·429 계정 분산 팁·"다른 조합·통일 무방" 원칙). 세션명은
   plan/run/sync만 사용(t118 이전 이름 미사용), `-f` 미언급 준수. 헤딩 델타 대칭
   (README 75→76, kanban-mode 21→22 ×4로케일).
8. **ko 정본 보조 수정**: 오타 3건(族→계열, 충링→충돌, 우대우→우대 — 번역자 플래그 수용).
9. verify 레시피 전항목 PASS + 패리티 래칫 55→54(statusline 4로케일 수렴, baseline pruned).

## 2. Evidence (증거) — 전부 본 세션 실측

### 헤딩 패리티 (게이트 메트릭, 4로케일 최종값 — 오케스트레이터 전건 재측정)

| 페이지 | ko | en | ja | zh |
|---|---|---|---|---|
| README | 76 (H2 12) | 76 (12) | 76 (12) | 76 (12) |
| advanced/kanban-mode | 22 | 22 | 22 | 22 |
| cli-reference/launchers | 7 | 7 | 7 | 7 |
| guides/mcp-server | 20 | 20 | 20 | 20 |
| advanced/statusline | 13 | 13 | 13 | 13 |
| utility-commands/moai-todo | 8 | 8 | 8 | 8 |
| utility-commands/moai-clean | 19 | 19 | 19 | 19 |
| cli-reference/update | 32 | 32 | 32 | 32 |
| advanced/moai-web-console | 15 | 15 | 15 | 15 |
| cli-reference/graph (신규) | 5 | 5 | 5 | 5 |
| cli-reference/tokens (신규) | 4 | 4 | 4 | 4 |
| cli-reference/memory (신규) | 3 | 3 | 3 | 3 |
| agent-teams (표적 1문단) | 14 | 15 | 15 | 15 (기존 발산 유지, baseline 등재) |

### 정적 검증 (exit 1 = 무매치, 전부 PASS)

- `hugo --minify --gc` WARN/ERROR grep 무매치 · sitemap OK (수차례: 드래프트 후·mermaid 후·최종)
- URL 블랙리스트(content+README 4종) 0 · Mermaid LR/RL 0
- README H2 패리티 12/12/12/12
- 버전 게이트: `Release-v3.1.0|v3.1-rc` 0, 스테일 `🗿 v3.0.x/v3.1.0│` 표시 0
- 용어 게이트: 허용 위치(은퇴 각주·센티널·ExecutionFactory=Java API명) 외 팩토리/factory 0
- 이모지 diff 스캔: 추가 줄의 이모지는 전부 UI 글리프 문서화 클래스(statusline 렌더 출력
  예시·세그먼트 키 표의 글리프 셀 — ko 정본 표와 동일 위치) — 판정 통과

### 래칫

발산 집합 55→**54** (awk 전수, statusline 4로케일 13/13/13/13 수렴 → baseline 1행 prune).
신규 발산 0 (12 Mode-A 페이지 전부 ko와 동일).

### D2 — 커맨드 수 실측 (go build 후 `--help`)

```
$ go build -o /tmp/moai-cmdcount ./cmd/moai && /tmp/moai-cmdcount --help
원시 최상위 항목 52개 (completion·help·moai 자기자신·mcp/mcp-server·migrate/migration 등
내장·별칭·내부 포함). "전체 36개" 주장(README:658,690)의 원 집계 규칙은 #1406 시점 것으로
역추적 불가 — 계획서 지시("확정 전 임의 변경 금지")대로 36 유지, 실측 데이터를 리드 승격.
```

### 드래프트 안전성

`git diff --stat 1bd9140eb..4100d8767 -- <13 경로>` → 빈 (드래프트 기점 이후 main에서 대상
파일 무변경 → 전체 교체 무손실).

## 3. Baseline-attribution (baseline 귀속)

전 측정 본 워크트리(`.claude/worktrees/ossdocs-v311`, HEAD 4100d8767 + 작업 diff)에서 이번
실행. ko 드래프트는 release-update/oss-docs 1단계 감사 산물(reports/oss-docs-v311/drafts/,
13파일 전체본)을 원문 그대로 적용. mermaid fixes는 mermaid-audit 카드 산물 원문 적용.
"전" 헤딩 수는 드래프트 적용 전·번역 인계 전 각 시점에 실측.

## 4. Gaps (미검증)

- **PR 프리뷰 육안검증 미실행**(mermaid 감사 지시 사항): mmdc 부재로 로컬 렌더 불가 — PR
  개설 후 Vercel 프리뷰에서 worktree/_index.md 보드 다이어그램·goal.md 루프·대표 1페이지
  육안 확인이 **PR 단계 게이트**로 이관됨 (PR은 리드 승인 후 개설).
- ja·zh 번역의 모국어 심사(제3자) 미실시 — 번역 전문가 자체 검증 + 오케스트레이터 구조·
  식별자·용어 검증으로 갈음.
- README moai-web 스크린샷 alt "10개 설정 탭" vs 본문 11탭 (번역자 2인 플래그) — alt는 실제
  캡처 화면 서술일 수 있어 보존, 콘텐츠 소관 판단 이관.
- D2 집계 규칙 미확정(위) — 릴리즈 노트 시점 운영자 확정 필요.
- sessions.md en/ko에 ja/zh만 있는 `/clear` 노드 부재(mermaid 감사 관찰) — 기록만,
  콘텐츠 소관(다이어그램 내용 추가 창작은 이 카드 범위 밖).
- 5페이지 미감사(ast-grep·hooks-reference·worktree faq·agentic/worktrees·model-policy) —
  다음 라운드 후보로 카드 지시대로 범위 외.

## 5. Residual-risk (잔여 위험)

- **hugo.toml releaseDate=2026-08-18 임의 설정** — 실제 태그일과 다를 수 있음. 릴리즈
  하네스가 태그 시점에 정정해야 함 (i18n 규칙상 날짜는 릴리즈 프로세스 소관).
- 인용-사각형(F1) mermaid 처방은 코퍼스 증명 형태이나 로컬 렌더 미검증 — 프리뷰 육안이 최종.
- "36개" 커맨드 수가 릴리즈 시점에 실측과 어긋나면 README 4본 + 안내 문구 재수정 필요.
- 백엔드 조합 서브섹션은 00af8d334 부트스트랩 문구 기준 — 향후 세션명 개명(t118) 시
  4로케일 동시 갱신 필요.

## 변경 파일 개수 (커밋 시점 git status 기준 기재)

- docs-site content: ko 12+1(_meta) / en·ja·zh 각 12+1 / mermaid 22(중복 없음—불일치 세트) 
- README 4종 · hugo.toml · main.yaml · faq(ko,en) · moai-feedback(ko)
- baseline 1 · 하네스 스킬 2 · 본 증거
