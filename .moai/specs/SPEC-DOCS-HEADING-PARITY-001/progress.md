# Progress: SPEC-DOCS-HEADING-PARITY-001 (card t1328)

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출물 세트(spec.md / plan.md / acceptance.md / progress.md) 작성 완료, status: draft.
- 사전 결정 테이블(spec.md §2)과 패리티 카운터 정의(spec.md §4)는 2026-09-29 본 워크트리(base develop `ec52b6e70`)에서 git 이력·카운터 출력을 재측정해 확정했다.
- 카드 t1328 인계 메모: run phase는 M1(기준선 기록)부터 착수하고, 증거는 `.moai/reports/t1328/`에 착지시킨다.

## §E.2 Run-phase Evidence

- 실행 주체: `hns-oss-docs-locale-translator-specialist` (카드 t1328, 워크트리 `.moai/worktrees/t1328`, base develop `ec52b6e70`, base HEAD `59f6d5072`, 브랜치 `WT-heading-parity`)
- M1 기준선: `.moai/reports/t1328/baseline.md` + `before-heading-tables.md` (§2 결정 재검증 — 근거 커밋 `5f4f199ef`/`9a53efd24`/`7ad954556`/`ce79ef7ca` 본 런 git show 재확인, BEFORE 카운터 agent-guide 22/30/30/30 · multi-llm 11/5/5/5 · tokenomics 12/13/12/11)
- M2 커밋 `211defda4` — tokenomics en/ja/zh → ko 12 구조: en "Model Tier Routing" 접어 Layer B 병합(en 단독 수리 문장 이월), ja `モデル別コンテキストしきい値` 신설 + 티어 라우팅 병합, zh 프롬프트 캐시·임계치 섹션 신설 + 티어 라우팅 병합. ko 불변. 고정 카운터 12/12/12/12
- M3 커밋 `471686950` — multi-llm en/ja/zh → ko 11 구조 전면 재유도(ko 5세대 스테일 해소, ja 퀵스타트 `moai glm setup` 누락 수리 포함). 고정 카운터 11/11/11/11
- M4 커밋 `329ad80b0` — agent-guide en/ja/zh 30→22 접기(ko 구조). REQ-006 인벤토리 diff + 섹션별 처분 기록 완비. 고정 카운터 22/22/22/22, 4로케일 헤딩 줄 위치 동일
- M5 커밋 `eeb939b27` — 래칫 프루닝 삭제 전용(`0 added / 3 deleted`), `comm -23`(신규 발산) 전체 트리 empty, `comm -13` empty
- M6 검증 게이트: build-clean PASS(hugo exit 0 · WARN/ERROR 0줄 · sitemap OK) · locale-parity PASS · version-sync PASS(배지 4로케일 v3.1.3 = hugo.toml; faq 🗿 4건은 업데이트 플로 예시로 분류 — m6-verify-gate.md §5) · Mermaid TD-only 무위반 · URL 블랙리스트 무적중 · 수정 9파일 이모지 0적중 · README 4파일 H2 12/12/12/12
- AC-DOCS-004 생존 grep (M2 종료 판정): en `concentrated on judgment work` 1 · en `Sonnet 4\.5 and earlier \(200K\)` 1 · ko `Sonnet 4\.5 이하 \(200K\)` 1 · ja/zh `Sonnet 4\.5.*[（(]200K[)）]` 각 1 — 전 패턴 1 이상 적중
- 증거 위치: `.moai/reports/t1328/` (baseline · before/after-heading-tables · m2/m3/m4 evidence · m6-verify-gate — `.moai/reports/*` gitignore 정책에 따라 디스크 전용, 커밋하지 않음)

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready — M1~M6 전 마일스톤 완료, must_pass 3차원(build-clean · locale-parity · version-sync) 모두 PASS
- 커밋 목록: `211defda4`(M2) → `471686950`(M3) → `329ad80b0`(M4) → `eeb939b27`(M5) → 본 커밋(progress 갱신)
- spec.md frontmatter status 전환(draft → in-progress)은 run-phase 스폰 전문가 권한 밖으로 판단해 수행하지 않음 — orchestrator/manager 레이어 처리 필요
- 감사 유의 사항: ① m6-verify-gate.md §5의 version-sync 분류(faq 업데이트 예시 4건)는 sync-auditor 재판정 대상 ② M4에서 파생에만 존재하던 사실 목록(m4-agent-guide-evidence.md §3)은 content-author 검토 대상(무처분 삭제 아님, ko 편집 금지 원칙 존중) ③ M3에서 GLM 구독 가격·무료 모델 콘텐츠는 ko 트리 미게재 확인 후 스테일 탈락 처분(m3-multi-llm-evidence.md §3)
- push 금지 준수 — publish는 orchestrator/human-gated

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete — 단일 sync 커밋으로 CHANGELOG [Unreleased] > Added 등재, spec.md frontmatter `implemented → completed` 전환, 본 §E.4 기록을 함께 실어 착지. `sync_commit_sha`는 커밋이 자기 해시를 인용할 수 없으므로 `pending-backfill-sync` 플레이스홀더로 적고 직후 커밋에서 백필(D3 면제).
- sync_commit_sha: "pending-backfill-sync"
- changelog_entry_position: [Unreleased] > Added 최상단 (B12 선방출 grep `grep -c 'SPEC-DOCS-HEADING-PARITY-001' CHANGELOG.md` = 0 확인 후 편입)
- b12_self_test_a: PASS — 선방출 grep 0건 (중복 편입 없음)
- b12_self_test_b: PASS — acceptance.md SSOT 기준 고유 AC 9건(AC-DOCS-001..009) = CHANGELOG 기재 수 9건 일치
- b12_self_test_c: PASS — CHANGELOG가 인용하는 문서 경로 9파일 + 래칫 베이스라인 파일 실존 확인(`ls` 재검증), 3페이지 × 4로케일 헤딩 카운터 `grep -rc '^#\{2,\} '`로 본 트리에서 재측정: agent-guide 22/22/22/22 · tokenomics 12/12/12/12 · multi-llm 11/11/11/11
- canary_compliance_check: n/a — 본 SPEC이 정의하는 선향 정책 없음
- mx_tag_validation: n/a — docs 전용 SPEC, Go 표면 변경 없음(RUN §E.3 참조)
- README/docs-site 결정: README 4파일은 이번 SPEC이 손대지 않았고 M6에서 H2 12/12/12/12 패리티 재확인 — CHANGELOG·README 추가 항목 없음. 인도물 자체가 docs-site 콘텐츠이므로 별도 문서 동기화 대상 없음.
- reviewer-attention 이월(run §E.3 ①): m6-verify-gate.md §5의 version-sync faq 🗿 4건 분류(업데이트 플로 예시)는 기록된 검토자 주의 항목으로 유지 — 실패 아님, sync-auditor 재판정 여지 명시
- AC 상태: 9/9 PASS (AC-TSS형 이월 편차 없음)
