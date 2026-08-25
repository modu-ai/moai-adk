# plan — SPEC-DOCS-V313-CATCHUP-001

## §A Context

카드 t274 (Class C, Tier M): v3.1.3 릴리즈(#1602 계열, CHANGELOG 2026-08-24)에서 문서가 코드를 따라가지 못한 격차를 닫는다. CHANGELOG `[3.1.3]` 26항목(Added 13·Changed 4·Fixed 9)을 docs-site 4로케일 + README 4파일에 문서 존재 여부로 매핑한 격차 표가 spec.md §1에 검증 완료 상태로 있다: **D 4 · U 10 · N 4 · NA 8**, 항목 외 version SSOT 갭 3건(V1–V3).

- baseline 트리: `e07a6d0f4` (worktree t274, branch WT-v313-docs) — 격차 표의 모든 관측은 이 트리에서 이번 실행으로 얻었다.
- 4로케일 파일 파리티 100% (파일 목록 체크섬 `98d2b226e6569dd7b07a8ce9ee4d3e5c` ×4), README H2 12개 ×4 — 구조는 건전하고 격차는 콘텐츠 수준이다.
- 문서 규격은 Skill "hns-oss-docs-i18n-rules" (SSOT `.moai/docs/docs-site-i18n-rules.md`)와 Skill "hns-oss-docs-readme-sync"를 따른다.

## §B Known Issues

- **hugo.toml v3.1.2 스테일** (55–56행): t272 종결(2026-08-25) 시 지목된 잔여. README 배지는 v3.1.3인데 hugo.toml과 README 예시(491·766행)만 뒤처진 상태. 이 SPEC이 증상을 수정(V1–V3), 프로세스 원인은 별도 카드 권장 (spec.md §6).
- **ko `settings-json.md`의 `ANTHROPIC_` 부재 편차**: 배치 관측에서 en/ja/zh의 settings-json.md는 `ANTHROPIC_` 언급이 있으나 ko는 없다 (파일 목록 파리티와 무관한 콘텐츠 편차). 본 SPEC 범위 밖 관측 기록 — C4는 NA 판정이므로 이 편차를 이 SPEC에서 수리하지 않는다.
- **`moai-gate.md`가 코드를 앞서 있었음**: A13(typecheck 축)은 문서가 이미 4축을 서술 — #1592로 코드가 문서를 따라잡았다. 갱신 불필요(판정 D)이며, run-phase에서 문서를 되레 "수정"하지 않도록 한다.

## §C Pre-flight (착수 전 측정)

run-phase 착수 시 §1 격차 표의 셀을 전부 재관측한다 (REQ-DVC-008). 재관측 명령 세트 (전부 이 워크트리에서 실행):

1. `grep -n '^## \[' CHANGELOG.md` — `[3.1.3]` 177행 존재 + 항목 수 재확인 (Added 13 / Changed 4 / Fixed 9).
2. §1 각 행의 grep 재실행 — 예: `grep -rln 'todo\.enabled' docs-site/content README.md README.ko.md README.ja.md README.zh.md` → 현재 0파일. 셀 값이 plan-phase 관측과 달라지면(예: 병렬 세션이 문서화) 해당 행을 재판정하고 격차 표를 갱신한다.
3. verify 레시피 baseline: hugo build가 이미 경고 없이 통과하는지, Mermaid LR/RL·URL 블랙리스트·body-emoji 현재 카운트를 측정 — run 종료 게이트(REQ-DVC-007)의 경과 판정 기준선. 기존 경고가 이미 있다면 그 목록을 기록하고 이 SPEC diff가 만드는 새 경고만 게이트 대상으로 판정한다(기존 부채 수리는 무상 확장 금지).
4. hugo 빌드 환경 확인 (`hugo` 바이너리 존재 — verify 레시피 요건).

## §D Constraints

- 4로케일 동일 PR 의무 (locale-parity must_pass): ko 정본 → en/ja/zh 파생, 같은 PR.
- docs-site 규칙: Mermaid TD-only, 본문 emoji 금지(icon shortcode), URL `adk.moai.kr` 단독, 강조 마커 간격, `moai-brand.css` FROZEN.
- README 규칙: `README.ko.md` 정본, 공용 언어 스위처 헤더 계약, 섹션 순서 파리티 유지 (H2 12개 구조 변경 없음 — 기존 섹션 안에서 갱신).
- 버전 SSOT: `hugo.toml`이 단일 버전 표면 — 페이지·메뉴·README에 배포 프로세스 밖의 이중 버전 문자열을 하드코딩하지 않는다. "introduced in vX.Y.Z"류 역사 인용은 건드리지 않는다.
- 게시(푸시)는 human-gated — run-phase는 파일 편집까지만.

## §E Self-Verification (run-phase 보고 형식)

마일스톤별로 아래 표를 progress.md §E.2에 축적한다 (전부 이번 실행·이 트리 관측):

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-DVC-00x | PASS/FAIL | `<명령 원문>` | `<출력 원문>` |

검증 명령 카탈로그 (acceptance.md §D와 대응):
- U 항목 착지: 항목별 키워드 grep이 4로케일 각각 1+ 파일 매칭 (예: `grep -rln 'todo\.enabled' docs-site/content/{ko,en,ja,zh}/…`).
- version SSOT: `grep -n 'version\|releaseDate' docs-site/hugo.toml` → `v3.1.3`/`2026-08-24`; `grep -n '🗿 v3\.1' README.ko.md` 등 4파일.
- 범위 순结성: `git diff --stat e07a6d0f4 -- internal/ pkg/ cmd/ internal/template/templates/` → 0파일.
- 종료 게이트: hns-oss-docs-verify 레시피 7축 (warning-free hugo build, sitemap 존재, URL 블랙리스트 grep 0, Mermaid LR/RL grep 0, 4로케일 파일+섹션 파리티, README 4파일 헤딩 파리티, body-emoji 스캔 0).

## §F Milestones (의사결정 가역성 내림차순 — 검토 민감도 높은 것부터)

### M1 — [OPERATOR GATE] codex dual-harness 신규 페이지 승인 확정 (N: A1–A4)

가장 구조적인 결정부터. 신규 페이지(가칭 `advanced/codex-dual-harness.md`) 생성 여부를 operator에게 질의한다 — 내비게이션 설정(`_meta.yaml` ×4, `data/menu/main.yaml`, `menu.html` SVG case)은 structure-curator 소관이며 묵시적으로 가정하지 않는다. 대안: 기존 페이지 흡수(foundations 절 언급 수준) 또는 deferred(별도 카드). **승인·거부·지연 어느 쪽이든 M2–M3는 봉쇄되지 않는다.** 거부/지연 시 A1–A4를 "deferred — separate card"로 격차 표에 기록하고 M4는 축소된 범위로 진행한다.

### M2 — ko canonical 갱신 (U 10항목 + V1–V3)

docs-site ko 7페이지(moai-feedback, config-sections, skill-guide, update, profile-matrix, model-policy, multi-model-audit) + `hugo.toml` + `README.ko.md`(V2·V3 예시, A11 절)를 ko 정본으로 갱신한다. C1은 profile-matrix.md와 model-policy.md 양쪽 매트릭스 표를 모두 다시 쓴다 (한쪽만 고치면 페이지 간 모순이 남는다).

### M3 — en/ja/zh 파생 (U 항목 + version SSOT)

M2의 ko 정본에서 en/ja/zh 3로케일 ×7페이지 + README 3파일을 파생한다. Mermaid 방향·코드 블록·URL은 verbatim 보존. README 언어 스위처 헤더 계약 유지.

### M4 — (승인된 경우에만) 신규 페이지 ko 저작 + 4로케일 파생

M1 승인이 난 경우에만 실행. ko 저작 → `_meta.yaml`·`main.yaml`·`menu.html` 반영(Skill "hns-oss-docs-structure-map" 로드) → en/ja/zh 파생. 승인이 나지 않았으면 이 마일스톤은 건너뛰고 그 사실을 progress.md에 기록한다.

### M5 — 종료 게이트: hns-oss-docs-verify + 격차 표 폐쇄

verify 레시피 7축 실행(§E 카탈로그) — 전 축 통과 후 §1 격차 표의 각 U/N 행에 착지 증거(grep 결과)를 대입해 폐쇄 상태를 progress.md §E.2에 기록한다. NA 8항목은 근거와 함께 "문서화 안 함"으로 확정(REQ-DVC-006).

## §G Anti-Patterns

- **번역본에서 정본 수정**: en/ja/zh에서 발견한 ko 오류를 파생 로케일에서 고치는 행위 — 격차 표에 기록하고 ko에서 고친다.
- **로케일 일부만 커밋**: 4로케일 동일 PR 위반 (가장 흔한 파리티 실패 형태).
- **NA 항목 조용한 삭제**: 26항목에서 제외하면 "전수 조사" 주장이 깨진다 — 근거와 함께 표에 남긴다.
- **신규 페이지 조용한 추가**: 승인 없는 `_meta.yaml`/`main.yaml`/`menu.html` 변경은 REQ-DVC-003 위반.
- **버전 문자열 이중 하드코딩**: hugo.toml 외 표면에 배포와 무관한 버전/날짜를 박는 행위. 역사 인용("v3.1.1부터")은 예외가 아니다 — 그대로 둔다.
- **문서 드라이브바이 코드 수정**: 문서화 중 발견한 코드 문제를 같은 PR에서 고치는 행위 (REQ-DVC-005 위반) — 별도 카드로 보고.
- **A13 되돌리기**: moai-gate.md는 이미 옳다 — "typecheck이라는 단어가 없다"는 이유로 재작성하지 않는다.

## §H Cross-References

- spec.md §1 격차 표 (이 plan의 입력) · acceptance.md §D AC Matrix (검증 계약)
- Skill `hns-oss-docs-i18n-rules` (SSOT: `.moai/docs/docs-site-i18n-rules.md`) — 규칙 §1–§9
- Skill `hns-oss-docs-verify` — 종료 게이트 7축 레시피 (스크립트 `docs-i18n-check.sh`·`gen_menu.py`는 존재하지 않음 — 인라인 체크만 실행)
- Skill `hns-oss-docs-readme-sync` — README 4파일 동기화 절차
- Skill `hns-oss-docs-structure-map` — M4 신규 페이지 시 내비게이션 설정 스키마
- CHANGELOG.md 177–306행 (`[3.1.3]`) — 항목 원문
- t272 종결 메모리 (README 배지 v3.1.3 vs hugo.toml v3.1.2 잔여 지목, 2026-08-25)
- 카드 t274 (Class C, Tier M) — `moai todo` 큐
