# C1 — 'code-atlas' / 'Code Atlas' 이름 충돌 조사 (rc.2 #2177)

결론: **충돌 확인 — 'code-atlas'를 RC1 배포명으로 사용할 수 없다.** crates.io에 같은 이름의 활성 패키지(코드 저장소 시각화 도구 — 의미 영역까지 동일)가 있고, GitHub org `code-atlas`가 실재하며, 이름 포함 리포지터리가 1,248개다. npm·PyPI는 비어 있으나 하나라도 점유된 이름으로의 배포는 REQ-C1-01 판정상 충돌이다. 대안 `moai-atlas`·`moai-code-atlas`는 세 레지스트리 전부 비어 있다(실측). 최종 명칭은 사용자 확정 대상(리더 중계)이다.

## 조사 근거 (URL + 확인일 2026-11-09)

| # | 출처 | URL | 확인일 | 결과 | 판정 |
| --- | --- | --- | --- | --- | --- |
| 1 | crates.io | `https://crates.io/api/v1/crates/code-atlas` (API, UA 부착 — UA 없으면 403) | 2026-11-09 | **200 — 패키지 존재**: `code-atlas` v0.2.0, updated 2026-08-24, 47 downloads, "Render local code repositories as architectural treemaps…" | **충돌** (코드 저장소 도구로 의미 영역 동일) |
| 2 | npm | `https://registry.npmjs.org/code-atlas` | 2026-11-09 | 404 — 미점유 | 충돌 없음 |
| 3 | PyPI | `https://pypi.org/pypi/code-atlas/json` | 2026-11-09 | 404 — 미점유 | 충돌 없음 |
| 4 | GitHub org | `https://api.github.com/users/code-atlas` (`gh api users/code-atlas`) | 2026-11-09 | 200 — **Organization `code-atlas` 존재** | **충돌** (org namespace 점유) |
| 5 | GitHub repos | `https://api.github.com/search/repositories?q=code-atlas+in:name` | 2026-11-09 | total_count **1248** — `league1991/CodeAtlasVsix` ★268·`ystemsrx/code-atlas` ★63 등 기존 프로젝트 다수 | **충돌** (이름 선점 광범위) |
| 6 | 상표 DB (WIPO Global Brand Database) | `https://branddb.wipo.int/en/quicksearch/brand?by=brand%20name&v=code%20atlas` | 2026-11-09 | HTTP 200 — 단, 결과는 JS 렌더링이라 curl로 열람 불가 | **not_run/운영자 실측 대기** — 상표 교차는 브라우저에서 재확인 필요 |

추가 관측: crates.io 검색 `q=code-atlas` total 2(`code-atlas`·`pheobe`); UA 없는 curl은 crates.io가 403으로 거부(정책 — UA 식별 필수).

## REQ-C1-01 판정

- crates.io 충돌(의미 영역 동일) + GitHub org 점유 + 선점 프로젝트 광범위 → **'code-atlas' 단독 명칭은 충돌**. npm·PyPI가 비어 있어도 레지스트리 간 분리 배포는 동일 이름의 혼동을 만든다.
- 상표 DB는 결과 열람이 이루어지지 않았다(위 6행) — "상표 충돌 없음"이라고 쓰지 않는다. 배포 전 운영자 브라우저 실측이 필요하다.

## 대안 이름 후보 (AC-2 — 실측 2026-11-09, 전부 404 = 미점유)

| 후보 | crates.io | npm | PyPI | 비고 |
| --- | --- | --- | --- | --- |
| `moai-atlas` | 404 | 404 | 404 | 짧고 제품 namespace(moai-)와 결합 — CLI 서브커맨드 `moai atlas`와 자연 대응 |
| `moai-code-atlas` | 404 | 404 | 404 | 기능 설명성 유지 + namespace 결합 |

CLI 서브커밴드 비교: `moai atlas`(짧음, 입력 최소) vs `moai code-atlas`(자기 설명적). 둘 다 moai-cli 바이너리 하위 서브커맨드라 레지스트리 충돌과 무관하나, 배포 패키지명과의 대응이 일관되려면 패키지명 접미사와 같은 형태를 권한다. 최종 선택은 사용자 확정(HOLD — 리더 중계).

## not_run · 한계

- 상표 DB 결과 열람: not_run(JS 렌더링) — 운영자 브라우저 실측 대기(위 6행).
- Apple App Store·Homebrew·winget 등 그 외 채널: 이 카드 범위 밖(08-D·배포 채널 결정 이후).
- 조사는 2026-11-09 단일 시점 실측이다 — 배포 직전 재확인 권장(레지스트리는 실시간으로 변한다).
