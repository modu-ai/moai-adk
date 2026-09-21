# t500 판정서 — codex init→doctor 종단 테스트 + statusline 불변 + 가드 사각 + "41본" 정정

- 카드: t500 [C7 · 테스트] (class C, lens --deep)
- 브랜치: `WT-codex-e2e-guard` (워크트리 `.claude/worktrees/t500`) · 베이스 `ace1c5440` (= dispatch 시점 `origin/develop`)
- SPEC: `SPEC-CODEX-E2E-GUARD-001` (Tier M, test-only) — **completed** (3-phase close)
- 집행: lane-6 (세션 27fee58f) · kickoff 승인: 운영자 2026-09-07 (리드 경유 AskUserQuestion, 리드 대리 아님)

## 커밌 목록 (브랜치 위 순서)

| SHA | 주제 |
|---|---|
| `a3153c215` | plan 산출물 3종 + progress + plan-audit iter1/iter2 |
| `667509ac9` | run M1–M5 (e2e 신규 테스트 110줄, 가드 2→12본 확장, 뮤턴트 증명) |
| `2d98654af` | sync close (CHANGELOG + completed 전이 + §E.4) |
| `24aebc38b` | §E.3/§E.4 SHA 백필 |
| `1ad75e47a` | sync-audit F1 — CHANGELOG 41 모집단 문장 수리 (1줄) |

## Claim (주장)

1. **축1 (init→doctor 종단)**: 실제 init 커맨드 경로(위자드 seam=codex) → 밀폐된 doctor 판정(`checkCodexWiring`)으로 가는 테스트가 신설됐고 건강한 시스템에서 CheckOK + codex 발견 구절 부재를 단언한다 (claude-only 동반 음성 케이스 포함). 종단 테스트는 결함을 발견하지 않았다(시스템 건강).
2. **축2 (statusline 불변)**: 카드 전제 「부분집합 단언 테스트 부재」는 **반증**됐다 — `TestStatusLineDefaultSubsetOfAllowlist` 가 기존 존재. 소관은 뮤턴트 증명으로 이행: 기본 토큰 `git-branch`→`git-branchx` 1자 변형에서 RED, 복원 후 GREEN, 뮤턴트는 커밋 흔적 0. 정정 기록은 spec §F.1 1급 항목.
3. **축3 (가드 사각)**: `codexSpecFiles` 2본 → 12본 길이 고정 집합(`len != 12` Fatal), 빌드태그 가드 12본 전체 확장, exec 가드는 파일별 기대 첫인자 표(launcher=`req.Program`, mcp_codex=`binaryPath`, review_gate=`"git"`, 나머지 9본 제로 프리미티브 단정), 중립성은 go/ast 리터럴 스캔 + 긍정 대조 캐너리로 확장. 기존 단언 약화 0 (sync-audit diff 검증).
4. **부기 ("41본")**: 41 = `find internal -name '*codex*_test.go'` 전수 **테스트 파일** 모집단이고 추적 인용이 실재한다(`SPEC-CODEX-E2E-MEASURE-001` spec.md:40). 무테스트는 정확히 2본(빌드태그 픽스처) → 단언 표면 39. 레인 초기 「모집단 불일치」 결론은 스윕 범위 과소로 **폐기**(spec §F.2가 양쪽 모집단 병기). 레포 편집 없이 기록으로 종결.
5. **판정**: plan-audit iter2 **PASS 1.00** / sync-audit **PASS-with-debt 92.6** (차단 F1 = CHANGELOG 문장 오기 → `1ad75e47a`로 수리 완료) / AC 7/7 PASS / 프로덕션 코드 변경 0 (test-only 카드 준수).

## Evidence (증거 — 명령 + 관측, 전문 기록은 각 경로)

- run 전수: `progress.md §E.2` (명령+전문출력+HEAD 귀속 3조) + `.moai/state/verify/t500/run-evidence-20260907.md`(로컬 전용, 미추적)
- plan-audit: `.moai/reports/t500/plan-audit-iter1.md` (FAIL 0.94, D1) · `plan-audit-iter2.md` (PASS 1.00)
- sync-audit: `.moai/reports/t500/sync-audit.md` (4차원 97/96/88/90, 조화평균 92.6)
- 오케스트레이터 독립 재측정 (이 런, HEAD `1ad75e47a` 직전 시점): 커밋 파일 목록 일치(4파일) · `configtoml.go` 베이스 대비 델타 0 · e2e 셀렉터 `ok 1.372s` · 가드 배치 5종 전부 PASS · `codexwiring` `ok 0.598s`
- 뮤턴트 흔적 부재: `git log ace1c5440..HEAD -- internal/codexwiring/configtoml.go` 무출력 (sync-auditor 재현)

## Baseline-attribution (귀속)

전 측정은 이 런에 `.claude/worktrees/t500` 트리에서 수행. plan/run 측정은 `a3153c215`·`ace1c5440` 시점, 최종 재측정과 병합 대상 트리는 HEAD `1ad75e47a`(+증거 커밋). 리드의 독립 실측(HEAD `a3153c215`·미푸시 1 시점)과 본 보고 일치 확인됨.

## Gaps (명시적으로 관측하지 않은 것)

- cross-model `audit_multi`: 백엔드 0개 응답 → **inconclusive** 기록 (fail-open; 판정 권위는 세션 내 감사).
- 변경 **전** 커버리지 베이스라인 미측정 (§E.2.3 Gap 기록; 프로덕션 무변경이라 분모 동일 — 신규는 테스트 측 기여만).
- AC-CEG-004는 run에서 판정하지 않음 (sync 소관 — sync에서 PASS로 전도).
- D5 화장품 채무: acceptance.md RED-now 셀의 줄번호 3개가 수리 전 값(편집 시 Phase-1 skip 조건 3 소멸 — 미편집, 채무로 보존).
- `moai spec lint` ID 인자형 미지원: `moai spec lint SPEC-CODEX-E2E-GUARD-001` → ParseFailure(파일 경로로 해석). 경로형만 통과. **피드백 후보** — 이슈 발행은 리드·운영자 표면으로 이관.

## Residual-risk (관측에도 남는 위험)

- 리터럴 중립성 스캔 축소 3클래스(CLAUDE.local 인용·`.moai/reports` 상수·비ASCII)는 REQ-CEG-009 기록 근거로 수용 — 잔여 누출 클래스 위험은 수용된 설계 결정 (커맨드 표면 스캔 AC-CL-013은 전체 표면 유지).
- `internal/cli` 패키지 커버리지 80.7%는 기존 베이스라인(이 카드 회귀 아님).
- 로컬 green은 조기 신호 — 통합 판정은 develop push 후 원격 CI (darwin/windows 매트릭스 포함). F2(스코프 배치가 단일 패키지형)·F5(서버 바이너리 지연 e79c010b8)는 리드 참고.

## 증거 반출

증거는 전부 본 브랜치에 커밋(위 경로) — primary 미반출. 리드는 워크트리 `.claude/worktrees/t500` 직독 또는 병합 후 develop에서 판독 가능. sync는 창 전 워크트리 종결 완료(t342 회피).
