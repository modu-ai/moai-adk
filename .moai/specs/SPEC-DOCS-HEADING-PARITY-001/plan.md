# Plan: SPEC-DOCS-HEADING-PARITY-001 (card t1328)

## A. Context

docs-site의 4로케일(ko/en/ja/zh) 페이지 중 헤딩 구조가 갈라진 표적 3페이지를 정렬한다. 원인은 페이지마다 다르다 — (1) `advanced/agent-guide.md`는 t1300 M1(`5f4f199ef`, 2026-09-29)의 4로케일 동시 재작성이 구조를 불균등하게 만들었고(ko 22 vs 타 30), (2) `multi-llm/_index.md`는 ko가 마지막 4로케일 터치(`ce79ef7ca`, 2026-09-12) 이후 5세대 단독 진행돼 en/ja/zh가 스테일이다(ko 11 vs 타 5), (3) `advanced/tokenomics-overview.md`는 en 단독 사실 수리 2건(`9a53efd24` t1115, `7ad954556` t1095)이 4로케일 흐름 사이에 끼어 있다(ko 12 / en 13 / ja 12 / zh 11).

판정 도구는 고정된 패리티 카운터(spec.md §4, `hns-oss-docs-verify` §4 recipe)다. 카드 본문의 수치와 레인의 naive grep 수치가 어긋났던 건 도구 차이였고, 카드 수치가 공식 카운터 출력과 정확히 일치함을 본 워크트리에서 재측정으로 확인했다.

## B. Known Issues

- `tokenomics-overview.md`의 en 단독 수리는 ja/zh에 미반영 상태다. 재유도는 ja/zh가 이 사실을 처음 갖게 되는 경로이기도 하다 — 나이브 ko 덮어쓰기로 en의 수리 문장이 사라지는 것이 본 SPEC이 막는 최대 위험이다.
- `agent-guide.md` 정렬은 en/ja/zh의 30→22 접기를 수반한다. ko에 없는 섹션이 en에만 있을 가능성을 REQ-006의 섹션 인벤토리 diff로 걸러낸다.
- 카운터는 코드 펜스 내부의 `## ` 패턴 적중 줄도 센다. 정렬 전후 반드시 같은 정의로 재측정해야 수치 비교가 성립한다.

## C. Pre-flight

- [ ] 작업 트리: 카드 워크트리(본 워크트리, base = develop `ec52b6e70`)에서만 문서 수정한다.
- [ ] spec.md §2 사전 결정 테이블을 run phase 착수 시 재확인한다. 표를 바꿔야 하면 orchestrator 경유 사용자 확인(D-NEW-1) 후에만 바꾼다.
- [ ] 정렬 착수 전 `.moai/reports/t1328/before-heading-tables.md`에 세 페이지 × 4로케일의 카운터 출력과 섹션 제목 목록을 기록한다.
- [ ] `hns-oss-docs-i18n-rules` 스킬을 스폰 전문가에게 로드 지시한다(모든 oss-docs 전문가 최우선 로드 규약).

## D. Constraints

- docs-site 본문 수정은 표적 3페이지 + 래칫 파일 `docs-site/.locale-parity-baseline` + (필요 시 최소한의) 내비게이션 설정으로 한정한다.
- 4로케일 수정은 같은 커밋 묶음으로 유지한다.
- Mermaid는 TD-only, 본문 이모지 금지, URL은 `adk.mo.ai.kr`만.
- 래칫 파일은 프루닝만 허용 — 줄 추가 금지(spec.md REQ-007).
- 스폰 전문가는 AskUserQuestion을 쓰지 않고 블로커 리포트로 돌아온다.

## E. Self-Verification

- [ ] E1: 세 표적 페이지가 고정 카운터 기준 ko == en == ja == zh (레시피 `comm -23` 출력 없음).
- [ ] E2: 래칫 파일 diff가 삭제 전용(프루닝)이고 수렴 3페이지의 줄이 빠졌는지 확인.
- [ ] E3: tokenomics en 단독 수리 생존 검사 — `9a53efd24`/`7ad954556`가 만든 문장이 재유도 후 4로케일 모두에 존재(grep 증거).
- [ ] E4: `hns-oss-docs-verify` 전체 레시피 — must_pass 4개 차원(build-clean, locale-parity, version-sync, content-fidelity 기준) 통과 출력.
- [ ] E5: 증거가 `.moai/reports/t1328/`에 착지(verdict + 전/후 테이블)했는지 확인.

## F. Milestones

우선순위 순이며, 결정 가역성 순으로 배열했다 — 사람이 다시 결정할 가능성이 큰 것(구조 접기·사실 이월)을 앞에 두고 기계적 마무리(래칫·게이트)를 뒤에 둔다.

### M1 — 사전 결정 확정 + 정렬 전 기준선 기록 (Priority High)

- spec.md §2 테이블과 근거 커밋을 run phase에서 재확인하고, `.moai/reports/t1328/`에 canonical-generation verdict 초안 + 정렬 전 헤딩 테이블(고정 카운터 출력)을 착지시킨다.
- 담당: orchestrator가 내용 없는 기록 단계로 직접 수행 가능하나, 판정이 들어가면 locale-translator 스폰이 수행.

### M2 — `advanced/tokenomics-overview.md` 정렬 (Priority High — 위험 최대 페이지를 먼저)

- ko 구조를 기준으로 en/ja/zh 재유도하되, `9a53efd24`(t1115 sync-audit 수리)와 `7ad954556`(t1095 Sonnet 4.5/4.6 1M 경계)의 en 단독 문장을 이월한다(REQ-004).
- 담당: ko-캐노닉 저자(content-author)가 ko 구조 확정 → locale-translator가 en/ja/zh 재유도(수리 문장 이월 포함).
- 내용 보존 검사(REQ-006): 섹션 인벤토리 전/후 diff + 수리 문장 grep 증거.

### M3 — `multi-llm/_index.md` 정렬 (Priority High)

- en/ja/zh를 ko 최신 구조로 전면 재유도한다(스테일 파생 — 보존할 단독 수정 없음을 §2.1에서 이력으로 확인).
- 담당: locale-translator (ko 원본 → 3로케일 파생).

### M4 — `advanced/agent-guide.md` 정렬 (Priority Medium)

- en/ja/zh를 ko 구조(22섹션)로 재유도한다. 접기 과정에서 en/ja/zh에만 있는 내용은 REQ-006 인벤토리 diff로 잡아 처분(병합/이동/ko-부재 확인)을 문서화한다.
- 담당: content-author가 ko 구조 재확인 → locale-translator 재유도.

### M5 — 래칫 갱신 (Priority Medium)

- 수렴한 3페이지 줄을 `docs-site/.locale-parity-baseline`에서 삭제한다(REQ-007 — 프루닝만). diff는 삭제 전용이어야 한다.

### M6 — 마감 검증 게이트 (Priority High)

- `hns-oss-docs-verify` 레시피 전체 실행: warning-free hugo build + sitemap, 4로케일 패리티(신규 발산 0), README 패리티 회귀 없음, TD-only Mermaid, URL 블랙리스트, 이모지 스캔, version-sync(REQ-008).
- 증거를 `.moai/reports/t1328/`에 착지하고 전/후 헤딩 테이블을 완성한다(REQ-005).

### Run-phase 라우팅 노트

- 파이프라인 스킬 `hns-oss-docs-run`(Runner, `.claude/workflows/hns-oss-docs-run.js` 보유)의 author → translate → verify 흐름이 이 카드의 기본 경로다. 스폰 형태(Runner 위임 vs 페이지별 전문가 스폰)는 orchestrator가 run 시점에 결정한다.
- 내비게이션 설정을 건드리는 조각만 structure-curator가 담당한다(REQ-009) — 발생하지 않으면 스폰하지 않는다.

## G. Anti-Patterns

- 나이브 ko 덮어쓰기: tokenomics의 en 단독 수리를 파괴하는 경로 — 절대 금지(REQ-004).
- 트리 합계로 패리티 판정: 반대 방향 편차가 상쇄된다 — 페이지별 비교만 유효(hns-oss-docs-verify §4 경고).
- 래칫에 줄 추가: "구조조정으로 발산이 남았다"는 새 부채 인정 — 이 SPEC 범위에서 금지.
- 카운터 혼용: naive `grep -c '^#'` 등 다른 정의로 재측정해 "수렴"을 주장 — 금지(REQ-002).

## H. Cross-References

- spec.md §2 (사전 결정 테이블), §4 (패리티 카운터 정의)
- `.claude/skills/hns-oss-docs-verify/SKILL.md` §4 (recipe 원문)
- `.claude/skills/hns-oss-docs-i18n-rules/SKILL.md` (4로케일 동일 PR, Mermaid TD-only, URL 규칙)
- `docs-site/.locale-parity-baseline` (래칫 파일, 헤더 문언 = 프루닝 전용 규약)
- 관련 SPEC: SPEC-DOCS-LOCALE-PARITY-REPAIR-001 (선행 패리티 수리), SPEC-AGENT-MODEL-INHERIT-DOCS-001 (t1300 — agent-guide 발산의 직접 원인)
- 증거 경로: `.moai/reports/t1328/` (run phase에서 생성)
