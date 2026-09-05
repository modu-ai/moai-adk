# t484 run 원장 — t410 후속 4항목(F1·F5·D4·D6) 상환

- **브랜치**: `WT-t410-followups` (워크트리 `.claude/worktrees/t484`, 베이스 `a825183dd` = 당시 로컬 develop 팁)
- **운영 승인**: 리드 경유 운영자 kickoff 승인(2026-09-05) — 제안안 그대로 4건 상환 + F1 ①② 병행
- **성격**: 문서 전용 카드 — `internal/**`·`.claude/rules/**`·t410 워크트리 무접촉, `progress.md`는 읽기 전용 입력

## Phase 1 — 확인·재측정 (본 세션 직접 수행)

4개 라벨이 서로 다른 실제 항목임을 원문으로 확인, 넷 모두 트리 `a825183dd`에서 생존 재측정.
전제 무효화 검사: `git log --oneline 79c7a0e2f..a825183dd -- <ERA-H3 spec.md · DRIFT-CLOSE-BODY-001/ · drift.go · drift_index.go>` → **빈 출력** (t410 이후 대상 표면 무변경).
상세: 같은 디렉터리 `verdict.md` (Claim/Evidence/Baseline-attribution/Gaps/Residual-risk).
소비 구분: F2는 t410 sync가 소비(progress.md §E.4.1) · index.lock 캡처는 t485 소관 미접촉 · F3·F4·F6·F7·F8은 재량 잔여(배차 밖).

## Phase 2 — run (manager-spec 위임, opus)

산출 커밋 2개:

```
271f9bdb0 docs(SPEC-DRIFT-CLOSE-BODY-001): in-place amendment t484 — narrow REQ-DCB-002 (D4), document shape-A residual (F5), route cross-SPEC edits via manager-spec (F1), define Tier file population (D6)
c9b9a55dd docs(SPEC-ERA-H3-NARROWING-001): channel-attribution correction 0.5.2 (t484) — t410 0.5.1 edit was a manager-spec surface
```

- **D4**: REQ-DCB-002 "상태를 **반환했고**"로 좁힘(spec.md:100) + §4 오류 경로 H3 신설(:212, REQ-DCB-006 교차 참조)
- **F5**: §5.2 모양 A 꼬리 무제약 잔여 기록(:264, 코퍼스 18줄·실오해제 0건 실측 인용)
- **F1-②**: REQ-DCB-007(:110)·AC-DCB-007(:191 수행 채널 + 말미 근거 줄 강화) manager-spec 재위임 경유로 개정 — AC의 세 술어·grep 방언 주의·공허 방지 2조항 원문 보존 확인
- **D6**: plan.md §A 모집단 정의(소스 파일) + 범위 안 초과 시 동작 추가
- **F1-①**: ERA-H3 0.5.2 채널 귀속 정정 행 — 0.5.1 행 무편집, frontmatter는 version/updated 2필드만(diff +3 −2 실측)
- **amendment 장치**: frontmatter `status: in-progress` + `amendment_of: self` + version 0.4.0 + `## Amendments`(직전 완료 버전 0.3.0 · prior_completed_sha `c1a389036` · 사유 · 범위 · 간극 조정 기록)
- **간극 조정**: "REQ 2개+§4"(0.3.0 D1 행) vs 원장 스케치(REQ-DCB-002+§4) → 확정 편집 REQ는 REQ-DCB-002·REQ-DCB-007(F1-②는 0.3.0 시점 미상), REQ-DCB-006 문면 수정 불요 — `## Amendments`에 서술

## Phase 2 검증 (lane trust-but-verify, 본 세션 독립 실행)

- HISTORY 불변: `git diff -U0 a825183dd..HEAD -- <양쪽 spec.md> | grep -E '^[+-]\|'` → `+` 2행(0.4.0·0.5.2)뿐, 표 행 `-` 0건
- ERA-H3 전체 diff = frontmatter 2필드 + 신규 행 1개가 전부
- DRIFT spec.md 삭제 6줄 전부 검증: frontmatter 3 + REQ-DCB-002/007 치환 2 + AC-DCB-007 근거 줄 **강화 치환**(인용 규약과의 자기모순 해소 명시)
- lint: `/tmp/moai-t484 spec lint <DRIFT spec.md> --json` → `[]` rc=0 (단독 실행으로 0건 독립 확인) · ERA-H3는 advisory warning 8건 — 전부 편집 밖 95–100행(선존재 REQ 섹션), 본 편집이 만든 finding 아님

## Phase 3 — sync 재종결 (manager-docs 위임, sonnet)

```
76631690b docs(SPEC-DRIFT-CLOSE-BODY-001): sync-phase — 3-phase close (t484 amendment re-close, doc-only, no CHANGELOG: zero source changes)
bf08d54c1 chore(SPEC-DRIFT-CLOSE-BODY-001): backfill sync_commit_sha 76631690b (card t484)
```

- frontmatter `status: completed` 복귀(version 0.4.0 유지) · progress.md §E.4.2 신설(재종결 기록 + 실제 SHA 백필 완료)
- §E.3 의 선존재 `run_commit_sha: pending-backfill`(62·77행)은 run-phase 기록이라 미접촉 — 정확한 범위 판단
- CHANGELOG 미기입: 소스 0변경 (t481 선례와 동일)
- 종결 상태 lint 재확인: `[]` rc=0

## Gaps

- `--strict` lint 미실행(배차 밖) · ERA-H3 advisory 8건 미수리(제3 SPEC 편집이 되어 배차 초과)
- 원장·판정서의 `/tmp` 사본(`t484-lint.json` 등)은 휘발성 — 본 파일과 verdict.md 가 영구 증거
- 통합 창 미수행 — 리드 호명 대기

## Residual-risk

- 카드 진행 중 develop 이 `a825183dd` → `2e8057258+` 로 전진했으나, 전진분의 대상 표면 영향은 두 차례 측정(`a825183dd..develop` 경로 필터 log)으로 **0건** 확인 — 창 병합 시 충돌 예측 근거
- amend된 SPEC의 amendment 중 DRIFT 노출(정상 경로)은 재종결로 해소 — 현재 `completed`
