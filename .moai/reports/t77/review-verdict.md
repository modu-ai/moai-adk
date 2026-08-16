# t77 Review Verdict — PASS (운영자 대시보드 체크리스트 잔존)

- Card: t77 — GLM 과금 경로 실증 (측정 카드, 코드 변경 0 — verified: 커밋 `d9fddbca3`는 `.moai/reports/t77/` 5파일 추가만 포함)
- Criteria (dispatch-specified): 측정 설계의 타당성 + 증거 재현성
- Reviewer session: release-v311 (2acd4be4), 2026-08-17

## Claim (주장)

측정 설계가 타당하고, 증거가 재현 가능하며, 리포트 숫자가 원시 출력과 일치한다.

## Evidence (증거) — reviewer-executed re-verification

| Check | Command | Observed |
|---|---|---|
| Raw 대조 (런 A) | `grep -o '"modelUsage":{...}' raw/run-a.txt` | `glm-5.3` 단일 버킷: input 51053 / output 1471 / cacheRead 493504 / costUSD 0.53879... — 리포트 표와 소수점까지 일치 |
| claude-* 항목 | `grep -c '"claude-' raw/run-a.txt raw/run-b.txt` | `0` / `0` — 리포트 주장 일치 |
| 런 B / 산술 | `grep costUSD` + 수동 합산 | 0.525395 ✓; 0.4858 + 0.5388 + 0.5254 = 1.5500 ≈ **$1.55** ✓ |
| Probe input | `grep inputTokens raw/probe-print-passthrough.txt` | 97,126 — 체크리스트 "input ~80-97K" 상한의 출처 확인(80,282=런B, 97,126=probe) |
| 코드 0변경 | `git show --name-status d9fddbca3` | `.moai/reports/t77/**` 5파일 순수 추가 |
| Base | `git merge-base release/v3.1.1 WT-t77-t64` | `97daa3baf` ✓ |
| 비밀값 규율 | raw/report 통독 | 토큰 값 미출력, SET 여부만 기록 — 규율 준수 |
| 격리 규율 | report §환경 | `/tmp/t77t64-lab` 고립 실험장 — CLAUDE.local §13 준수 |

## 설계 타당성 판정

4겹 관측(코드 `setGLMEnv` → live 프로세스 env → modelUsage 집계 → Claude측 대조)은
"서브에이전트 = 같은 프로세스의 추가 API 호출"이라는 아키텍처 사실 위에서
과금 경로를 삼각측량하는 올바른 설계. z.ai 공개 usage API 부재(문서 열거) 하에
가능한 최강의 증거 체인. Gaps·잔여 위험이 정직하게 명시됨(5-섹션 형식 준수).

## Gaps (미검증 — 리뷰어 몫)

- 런 A/B의 **트랜스크립트** Agent 스폰 3회/런 주장: 트랜스크립트 파일 미커밋 —
  raw의 완료 결과·단일 glm 버킷으로 삼각측정될 뿐 직접 재검증은 안 됨.
- headless 런의 env는 인터랙티브 세션(pid 78144)에서 대표 채집 — 리포트 스스로
  "추론"이라 명시(동등성 근거: 동일 binary·cwd·config + 사용량 집계의 독립 뒷받침).

## Minor (cosmetic, non-blocking)

- 운영자 체크리스트의 "세션별 input ~80-97K"는 런A(51,053)를 포함하지 않는
  상한형 표현 — 대시보드 대조의 실질 키(총액 $1.55·glm 귀속)에는 무영향.

## Residual-risk (잔여 위험)

- **카드가 규정한 최종 판정(z.ai 콘솔 대시보드)은 운영자 체크리스트로 잔존** —
  리포트 §5의 기대치(합산 ≈ $1.55, 2026-08-17 04:00~04:15 KST)로 확인할 것.
  모순 관측 시 카드 지시에 따라 t77 retro-FAIL → t78/t79/t84~t86 전면 재검토.
- CC `modelUsage` 집계 ≠ z.ai 청구 항목 1:1 가능(캐시 단가) — 체크리스트 1번에서 대조.

**Verdict: PASS** (dispatch 기준 충족; 최종 판정은 문서화된 운영자 몫)

## 독립성 고지

본 리뷰 세션은 t77 작업을 수행한 세션과 동일 계보(/clear 이후) — 대화 메모리 없이
커밋된 증거만으로 전 판정을 재유도했음(신선판정 한계 명시).
