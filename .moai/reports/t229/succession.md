# t229 승계 판정 — lane-4 후속 세션 (2026-08-26)

- **카드**: t229 (Class B) · **SPEC**: SPEC-CODEX-VERDICT-SYNTH-001 (Tier S) · **브랜치**: `WT-audit-verdict-converge` @ `1a6c0fac0`
- **경위**: 원소유 세션(lane-2 계열, 세션 1063eee2) 종료로 픽이 고아가 되어 운영자 재픽(2026-08-26) → lane-4 후속 세션이 승계. 리드 실측: 브랜치에 run 종결 커밋 `1a6c0fac0` 존재.

## 판정: **이어받기 (continue)**

### 근거 — run 종결 주장을 판독으로 재검증 (verification-completeness 적용)

| 주장 | 재검증 (본 세션 관측, 트리 `1a6c0fac0`) | 결과 |
|---|---|---|
| 전 패키지 수트 초록 | `.moai/state/verify/t229-m4/pkg-cli.log` 판독 → `ok github.com/modu-ai/moai-adk/internal/cli 510.756s` (본문 `ok` 행 — 래퍼 종료코드 아님, `m4-close.md` §2 방침과 일치) | 일치 |
| mutant (e) 독립성 (DoD 211) | `mutant-e.log` 판독 → `--- FAIL: TestSynthesizeReviewOutput_AdoptsMostConservativeSignal` + K1·K2·K4 갈라짐, 나머지 16건 PASS — **실패가 관측된 검사** | 일치 |
| 뮤테이션 잔류 없음 | `shasum -a 256 -c .moai/state/verify/t229-m4/mcp_codex.sha256` → `internal/cli/mcp_codex.go: OK` (본 세션 직접 실행) | 일치 |
| AC 6/6 | progress.md §E.2/§E.3 판독 — AC-CVS-001..006 증인 매핑 존재, K3/K7 기대값 P-CONS 유도 그대로 (낮춤 없음 §E.3 명시) | 일치 |
| 정적 검사 | vet-darwin/vet-windows 로그 존재 (0바이트 = 출력 없음·rc 기록) | 일치 |

재실행 대신 증거 판독+해시 검증으로 재검증한 근거: full 수트는 510초 (로컬 재실행은 부하 규율 위반 — 전 패키지 판정은 PR CI 몫). 대상 테스트 재실행은 병합 후 실시 (아래 진행 기록 참조).

## 발견 — 카드↔SPEC 범위 갭 (운영자 결정: 별도 카드로 연기)

카드 본문의 추가 축("수렴 결과에 기여한 on-target 백엔드 수 노출, 참여자 2 미만이면 disagreement_flag false 금지 — lane-4 관측 2026-08-24 t235 감사, '범위에 포함' 문언")이 **SPEC 계약에 흡수된 적이 없음**:

- REQ-CVS-001..004 및 §E Out of Scope 어느 곳에도 무관 (§A.3의 실측 결함 집합은 G1/G3/G4 — 전부 백엔드 *내부* 판정 합성축)
- `internal/cli/mcp_convergence.go`에 참여자 수 관련 구현 흔적 0 (grep 실측)
- 추정 원인: 2026-08-24 카드 본문 append가 plan-phase(감사 iter1/iter2로 범위 확정)와 시간적으로 어긋남 — 계약 흡수 절차 없이 카드 문언만 갱신

**처분 (운영자 승인 2026-08-26)**: 현재 SPEC은 계약 범위(G1/G3/G4)대로 완결·종결하고, 참여자 수 축은 **별도 신규 카드**로 발행(리드 건의). 본 문서가 그 갭의 존재를 기록해 침묵 누락이 아니게 한다. 대표 mutant("참여자 수를 세되 여전히 false를 내보내는 구현")는 신규 카드의 AC 설계로 이관.

## 병합 기록

- 분기점 `f7eec06c7` (2026-08-23) — 이후 main 대거 전진(t207·t187·v3.1.3 풀·t269·t250·t259·t273·t274 등). `origin/main..HEAD` diff의 −19,409는 스테일 베이스 역방향 아티팩트 (카드 삭제가 아님).
- 병합 커밋·충돌 해소·병합 후 재검증: 아래 '진행' 절에 순차 기록.

## 진행 — 병합 후 재검증 (2026-08-26, 후속 세션 관측)

| 단계 | 관측 |
|---|---|
| 병합 | `4561f432c` (origin/main → WT-audit-verdict-converge, `--no-edit`) — 충돌 0, `internal/cli/` 마커 0건, 작업 트리 클린, divergence `0 17` |
| 카드 코드 생존 | `mcp_codex.go`에 `adoptConservativeVerdict`·`codexScoredVerdict`·`codexVerdictSignalsOf` 마커 10회; `codex_verdict_regression_test.go`·`codex_review_rpc_test.go` 존재 |
| 대상 테스트 (병합 트리) | 아래 코드블록의 명령 그대로 실행 → `ok github.com/modu-ai/moai-adk/internal/cli 1.034s` (본문 ok 행 직독). `-v` 재실행 `=== RUN` 34케이스 — 셀렉터 0매칭 아님 확인 |
| 전 패키지 판정 | CI 몫 (로컬 full 수트 재실행은 부하 규율 위반 — `m4-close.md` §2·CLAUDE.local §4) |

위 표의 대상 테스트 명령 (복사·실행용 원문 — 표 셀 안에서는 파이프를 이스케이프하면 복사 시 정규식이 리터럴 파이프로 변해 0매칭이 되므로 코드블록으로 둔다, PR #1663 CR):

```bash
go test ./internal/cli/ -run 'TestSynthesizeReviewOutput|TestConverge_|TestRunMultiAudit_|TestCodexTask_OutputText' -count=1
```
