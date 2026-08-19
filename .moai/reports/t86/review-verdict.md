# t86 Review Verdict — PASS

- Card: t86 — `moai tokens record` per-pool token accounting seed (commit `dd060a191`, branch WT-t86)
- Lens: default 4-perspective + security attention on file/path handling (dispatch-specified)
- Reviewer session: release-v311 (2acd4be4), 2026-08-17

## Claim (주장)

측정 설계(surfaces)·스키마 정정·범위 확장(형제 subagent 집계)·fail-open 설계가 타당하고,
증거(프로브)가 실측 트랜스크립트와 재현 가능하며, path/glob/ledger 쓰기 표면에 차단 결함이 없다.

## Evidence (증거) — reviewer-executed

| Check | Command | Observed |
|---|---|---|
| Ride-along | `git log release/v3.1.1..WT-t86` | 정확히 2커밋(구현 + baseline 머지) — 몰래 타는 것 없음 |
| 테스트 9/9 | `go -C <t86> test ./internal/cli/ -run 'Tokens' -count=1 -v` → 6 PASS; `-run 'TestAggregateTranscriptUsage' -v` → 3 PASS | **9/9 PASS** (디스패치의 "9 tests"와 정확 일치) |
| vet | `go -C <t86> vet ./internal/cli/` | VET OK |
| **범위 확장 근거 독립 재측정** | `grep -c '"isSidechain": *true'` on 실제 `~/.claude/projects/` 트랜스크립트 | 메인 2개(44400e2f·a2045b55) 모두 **0줄**; 형제 `subagents/*.jsonl` 2파일에 **457줄 전량** — 파서의 messages.sidechain=298은 assistant 라인만 카운트라 정합 |
| 프로브 대조 | probe JSON vs surfaces.md 숫자 | claude: 238 assistant/cache_read 77,804,403/output 154,499 · glm: 378/298·풀 분할 — 전부 일치 |
| diff 전문 독회 | /tmp/t86_full.diff (951줄) | 아래 판정 근거 |

## 판정 — 디스패치가 명시한 4개 심사 지점

1. **범위 확장(형제 subagent 집계): 정당하고 적절히 경계 지어짐.** 근거(메인 0 / 형제 전량)를
   이 리뷰 세션이 독립 재측정으로 확증. 경계: 형제 glob은 해석된 transcript 경로에서 파생
   (`<dir>/<stem>/subagents/*.jsonl` — 원시 사용자 입력 아님), 동일 assistant+usage 필터로
   병합, 구버전 레이아웃(메인 인라인 sidechain)은 라인별 isSidechain 플래그가 이미 커버
   (코드 코멘트 명시). 잔여: 가상의 혼합 레이아웃(메인에 sidechain >0 && 형제 존재)에서는
   이중 카운트 가능 — 2.1.23x 실측상 발생 0, 저비용 가드(메인 sidechain>0이면 형제 skip) 후속 권장.
2. **스키마 정정(message 하위 model/usage): 확증.** transcriptLine 파싱 구조·fixture·
   프로브(파서가 실제 파일에서 모델별 숫자를 정확 산출) 삼각 지지.
3. **Fail-open 설계: 확인.** truncated tail·malformed 라인 → `skipped_lines` 카운트만
   (테스트로 단언), walk-up 폴백은 `findStateDir` 기존 관례 재사용 + cwd 폴백.
4. **record.context 무가드: v1로 수용하되 기록.** 오염이 이론이 아니라 **커밋된 프로브에서
   실측 재현**(두 레코드 모두 제3세션 bfd68514의 스냅샷 삽입). 수용 사유: 토큰 회계 축은
   무영향이고 레코드가 양쪽 session_id(레코드 것과 context 것)를 자기기술해 소비자가
   불일치를 검출 가능. follow-up: `snap.SessionID == session_id` 매치 가드 1줄 또는
   필드 의미 문서화("cwd-프로젝트 마지막 렌더 스냅샷").

## Findings (advisory — 전부 비차단)

1. `--session` uuid 미검증 glob 주입: `defaultTranscriptResolver`가 sessionID를 검증 없이
   glob 패턴에 삽입 — `/`·`..` 성분으로 패턴이 `~/.claude/projects/*/` 밖으로 이탈 가능.
   영향은 로컬 동일 사용자의 .jsonl 읽기(토큰 합계로 환원·경로 레저 노출)로 한정 —
   실권한 경계 없음. follow-up: uuid 포맷 검증(`^[0-9a-fA-F-]{36}$` 또는 `/`, `..` 거부).
2. 혼합 레이아웃 이중카운트 가드 (위 판정 1 잔여).
3. walk-up 폴백 stray-ledger: 프로젝트 밖 실행 시 `<cwd>/.moai/state`에 1파일 낙서 —
   운영자 통제 호출로 봉쇄됨, 기명.
4. 동시 레인 append의 라인 인터리빙 이론 가능(카드 종료 시점 배타적 호출로 실질 무시).

## Gaps (미검증)

- golangci-lint 0 이슈 주장은 run lane 재검증 인용(리뷰 세션 재실행 안 함 — 워크트리
  격리로 lint cwd 진입 불가). CI가 PR에서 판정.
- 테스트 격리 검증: `t.Chdir` 사용 테스트들이 직렬 실행 관례 하에서만 안전(기존 패키지
  관례와 동일 — `-race` 무결침은 t79 사례와 동일하게 internal/cli 전체가 아니라 패턴
  스코프로 검증됨).

## Residual-risk (잔여 위험)

- context 축의 오염(판정 4)이 v1 레코드에 남음 — 자기기술로 완화, 가드는 후속.
- 다중 moai 프로세스가 동일 ledger에 동시 append하는 극단 시나리오(위 4번).

**Verdict: PASS** — 통합 진행. advisory 4건은 후속 카드/가드 권장, 본 카드 범위 아님.
