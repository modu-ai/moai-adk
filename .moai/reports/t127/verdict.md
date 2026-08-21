# t127 — GLM 백엔드 레인 10 병렬 런타임 실측

Class B (plan 스킵), 측정 카드. 측정 환경은 이 세션 자체: 팩토리 lane-9
(session 85c47b97-647a-4c6d-bf31-1e68c7488722, GLM-5.3 백엔드), 2026-08-22
02:14–02:22 KST. 스폰 표본은 무도구 trivial 프롬프트 ("Return the single
word OK") — 부하 규율 [HARD] 준수: 전체 스위트 없음, 백그라운드 부하 없음,
스폰 스태거 준수(앵커 1개 반환 후 9개 동시), 동시 상한 10.

- branch: `WT-glm-lanes` (워크트리 t127)
- base: origin/main @ e7aeec088 (미병합 0인 깨끗한 트리에서 착수)
- 코드 변경 없음 — 측정 카드

## 측정 환경 실측

| 항목 | 값 | 근거 |
|---|---|---|
| 세션 모델 | GLM-5.3 (세션 상속, 모델 오버라이드 없음) | 런타임 환경 |
| 레인 캡 주입 | `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS=10` 세션 env 실측 | `env` 출력 |
| 캡 상수 | `internal/config/defaults.go:393` `DefaultLaneMaxConcurrentSubagents = 10` | grep |
| 주입 함수 | `internal/cli/kanban.go:260` `seedLaneAgentCap()` — 빈 값일 때 env 설정 | grep |
| 주변 동시 GLM 세션 | lane-5/6/7/8 + lead, 대부분 busy (ListAgents 02:15 관측) | 세션 레지스트리 |
| CPU | 16 logical (12 P-core) | `sysctl hw.ncpu` |

카드 전제 확인: 캡 주입은 백엔드 무관 env 경로로 이 세션에 실제로 살아 있다.

## 측정 항목 4개 — 판정

### 1. 동시 스폰 성공률 — unnamed 기준 10/10 (100%)

 unnamed(이름 없는) 스폰, 스태거 앵커 1 + 동시 9, 전원 정상 반환:

| 샘플 | 지속시간 ms |
|---|---|
| 앵커 (02:20:53 스폰) | 6,028 |
| u1 | 5,168 |
| u2 | 963 |
| u3 | 5,005 |
| u4 | 4,749 |
| u5 | 5,023 |
| u6 | 4,343 |
| u7 | 5,712 |
| u8 | 6,266 |
| u9 | 14,331 |

전원 결과 "OK", 0 tool uses. 9개 동시 배치(02:21:44) 벽시계 완주 ~15s
(최장 u9 14.3s). p50 ≈ 5.0s, 최장 14.3s.

**named(이름 붙인) 스폰은 10/10 수용됐으나 0/10 반환 — 별도 발견으로 분리**
(아래 "핵심 발견"). 카드 질문("10 병렬을 GLM이 견디는가")의 답은
unnamed 경로 기준 **예, 100%**.

### 2. 429 발생 — 0건

스폰 총 20회(명명 10 + 무명 10) + 프로브 메시지 1회 전체에서 429/API
에러 통지 0건. 단, 런타임이 내부 재시도한 429는 관측 불가(아래 Gaps).

### 3. 레인 세션 CW% — 기준치 15% of 1M, 압력 없음

- 02:14:09 유효 스냅샷(`.moai/state/context-usage.json`, session_id
  일치): `context_window_size: 1000000`, `tokens_used: 150000`,
  `raw_pct: 15`, `stage: none`, `band: large` — **#1574의 1M 선언이
  레인 세션에 살아 있음을 직접 확인** (raw effectiveWindow 무시 규칙 준수).
- 측정 직후(02:22:32) 스냅샷은 `tokens_used: 0` — writer_pid 교체
  (71472→15288) 후 리셋 아티팩트로 유효성 가드상 무효 판정, 제외.
- 해석: trivial 10병렬 스폰은 세션 컨텍스트에 스폰 메타데이터 수준만
  더한다(각 반환 수 KB 미만). 1M 기준 15%에서 10병렬 버스트 후에도
  handoff 임계(50%)에 절대 도달하지 않는다. CW 압력은 병렬 스폰이 아니라
  스폰당 반환물 크기(실제 레인 작업의 리포트 분량)가 지배.

### 4. 머신 load — 스폰 팬아웃은 로컬 부하를 만들지 않는다

16코어 기준 load average (1/5/15min):

| 시점 | load | 상황 |
|---|---|---|
| 02:14:08 | 16.74 / 20.10 / 15.43 | 기준치(주변 레인 활동) |
| 02:17:15 | 13.96 / 16.56 / 14.78 | named 10 동시 진행 중 |
| 02:21:44 | 8.57 / 12.86 / 13.63 | unnamed 9 팬아웃 순간 |
| 02:22:23 | 9.61 / 12.52 / 13.46 | 완주 직후 |

스폰 배치 전후 load 상승 없음 — trivial 스폰은 API-bound이며 로컬 CPU를
소비하지 않는다. 관측된 부하는 전부 주변 레인(5세션 busy)의 앰비언트.
2026-08-15의 load 413 사고(로컬 풀 스위트 병렬)와 스폰 팬아웃은 부하
계열이 다르다 — 스폰 상한 10 자체는 머신 보호 관점에서 안전.

## 핵심 발견 — named 스폰은 Agent Teams 팀원으로 전환되어 결과를 반환하지 않는다

CLAUDE.md §15의 "미해결 불일치"(한 세션은 named 워커 5명 정상 반환, 다른
동일 버전 세션은 출력 없는 인프로세스 팀원으로 전환)의 **실패 측면을 이
세션에서 완전 재현**했다. 이 카드의 측정 중 우연히 포획됐고, 팩토리 레인
운영에 직접 적용되는 발견이다.

증거 체인 (전부 이번 세션 관측):

1. named 스폰 10개의 spawn 결과는 "will receive instructions via
   mailbox" 형태 (unnamed의 "Async agent launched + agentId"와 다른 채널).
2. spawn-01 (02:14:35 스폰) 6분+ 무반환. 무도구 1왕복 작업이다.
3. 02:19:28 spawn-01에 SendMessage 생태 프로브 → 무응답.
4. SendMessage 결과에 `routing` 객체: `{"sender":"team-lead",
   "target":"@spawn-01"}` — 팀 네임스페이스가 이름을 가져감
   (kanban-dispatch.md [HARD] 시그널과 동일 패턴).
5. TaskStop 10회 전부 `task_type: "in_process_teammate"`로 분류됨 —
   런타임 스스로 서브에이전트가 아니라 팀원이라 명시.
6. `~/.claude/teams/`, `~/.claude/tasks/` 모두 빈 디렉터리 — 디스크
   상태 없는 세션 내 전환.
7. 턴 양보(60s) 후에도 무반환 — 실행 스케줄 가설도 기각.
8. 대조 unnamed 스폰: 동일 프롬프트로 963ms~14.3s 전원 반환.

**운영 규칙 제안**: 팩토리/칸반 레인에서 결과 회수가 필요한 팬아웃은
 unnamed 스폰으로. named 스폰은 주소 지정 가능하지만(SendMessage 수신은
 성공) 결과 채널이 없어 좀비가 된다. `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
 배포 기본값이 유지되는 한 유효.

## Evidence (5-Section 준수)

| Claim | Command | Observed |
|---|---|---|
| 캡 주입 | `env \| grep MAX_CONCURRENT` | `=10` |
| 캡 코드 앵커 | `grep -n DefaultLaneMaxConcurrentSubagents internal/config/defaults.go` | `:393 const … = 10` |
| seedLaneAgentCap | `grep -n seedLaneAgentCap internal/cli/kanban.go` | `:260` |
| unnamed 10/10 | 스폰별 완료 통지 `<duration_ms>` | 6028/5168/963/5005/4749/5023/4343/5712/6266/14331 |
| named 0/10 | 02:14:35–02:20:36 관측창 | 완료 통지 0, 프로브 무응답 |
| 팀원 분류 | TaskStop ×10 | 전부 `in_process_teammate` |
| CW 기준치 | `cat .moai/state/context-usage.json` (02:14:09) | 1,000,000창 / 150,000토큰 / 15% |
| load 4시점 | `uptime` ×4 | 위 표 |

Baseline-attribution: 전 항목 이번 실행, 이 세션, 이 머신에서 측정.
이전 세션 수치 인용 없음.

## Gaps (미관측)

- 무도구 trivial 스폰은 스폰/API 경로만 측정한다. 실제 레인 작업(다중
  tool round, 긴 반환물)의 지속 동시성·토큰 소비는 별도 — 단 이 카드의
  질문(10병렬 생존)에는 trivial이 정확히 맞는 탐침.
- 429 관측은 '표면화된 에러' 기준. 런타임 내부 재시도가 429를 삼켰다면
  관측 불가 — "429 없음"은 "관측된 429 없음"의 정확한 표현.
- 사후 CW 스냅샷 무효(writer 리셋) — CW 추이가 아니라 기준치 단일 판독.
- n=1 배치, 단일 레인, 단일 시간대. 통계가 아니라 생존 실측.
- named 전환의 근인(플래그 상호작용)은 미검증 — env는 런치 시 고정이라
  `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=0` 대조에는 신규 세션이 필요.

## Residual-risk

- named→팀원 전환은 업스트림 런타임 동작. 버전 업데이트로 언제든 양상이
  바뀔 수 있음 — §15 양면 증거의 한쪽면으로 기록.
- load 귀속은 '차이 없음' 기반 — 주변 레인 활동과의 완전 분리가 아니라
  스폰이 유의미한 델타를 만들지 않았다는 관측.
- u9의 14.3s(타 평균 3배)는 단일 표본 — 급행정지 판단 근거 아님.
