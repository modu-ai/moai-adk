# 병렬 수정 충돌 방지 설계 (2026-11-09, 리더)

지시: ADK_PARALLEL_CONFLICT_PREVENTION_20261109_01 + ADK_PARALLEL_CONFLICT_PERFORMANCE_20261109_01 + ADK_CLAIM_AGENT_AND_M1_PRIORITY_20261109_01 §2-4(한 과제로 기록).
조사: 읽기 전용 탐색(origin/v2026.11.0 8091f9eaa → cf4e662af, `src/` 인용 행은 두 tip에서 동일).

## 1. 측정한 사실(기준선)

| 항목 | 값 | 측정 |
|---|---|---|
| 11-08 이후 merge 커밋 | 412 | `git log --merges --since=2026-11-08 origin/v2026.11.0` |
| 그중 PR 병합(first-parent) | 109 | 같은 범위 first-parent |
| 레인 동기화 merge("… into issue/…") | 304 | 제목 패턴 |
| 양쪽 부모와 다른 결과를 만든 레인 동기화(충돌 해소 상한) | 136 / 304 | `git diff-tree --cc --name-only` |
| 제목에 반복 해소가 드러난 merge(하한) | 9 | 제목 패턴(third, resolve, re-merge) |
| 양쪽 수정 파일 상위 | rust-product-path.md 109 · main.rs 70 · lib.rs 54 · source-manifest.json 21 · product_cli.rs 18 · claim.rs 7 · protocol-v1.md 6 · contract.rs 5 | 같은 방법 |
| 컴파일 의미 충돌(E0063) | 2건(#2218×#2223 ClaimRequest, #2233 GhTransport) | 감사 기록 |
| 열린 rc PR 중 main.rs 수정 | 6 / 10 | `gh pr diff --name-only` |
| 문서 drift | OPERATIONS 52 vs interfaces.md 50, rust-product-path.md 표 5개 누락 | `comm` |

not_run: 레인 유휴·claim·감사 대기시간의 정량값(장부 시각이 일부 `2xZ` 표기라 정밀 비교 불가) — 개선 후 같은 방법으로 다시 셀 항목은 위 표의 행들.

## 2. 충돌 지점(기능 하나당 손대는 곳)

| 지점 | 새 op 하나당 | 새 profile 키 하나당 |
|---|---|---|
| `src/main.rs` `enum Command`(:98) · `OPERATIONS`(:655) · `run_internal` match(:777) | 3 | - |
| `src/lib.rs` / `src/forge/mod.rs` `pub mod` 목록 | 1 | - |
| `src/forge/mod.rs:642` · `src/contract.rs:180` 오류 kind 공용 시험 배열 | 0-N | - |
| `src/profile.rs` MANAGED_KEYS·ADDITIVE_KEYS·필드·validate·주석·known 목록·v2() | - | 6 + 주석 |
| `docs/rust-product-path.md` 표·절, `interfaces.md` 개수·목록 | 2-3 | 1 |
| 구조체 리터럴(`..` 없음) — ClaimRequest(queue.rs:266 제품 코드), GhTransport, HttpRequest(14곳), Profile v2() | 필드 추가 시 E0063 | 동일 |

조용한 결함: `Command`에 variant를 넣고 OPERATIONS 튜플을 빠뜨리면 컴파일은 되고 실행 시 `unreachable!`(main.rs:772)로 panic — 시험 없음.

## 3. 대안 평가(비용·효과)와 채택

| 대안 | 비용 | 효과 | 판정 |
|---|---|---|---|
| OPERATIONS → `fn operation(&Command)` 전수 match | 낮음 | 누락이 panic→컴파일 오류 | **채택**(1단계) |
| clap 하위 명령 ↔ operation 대응 시험 | 낮음 | CI에서 누락 검출 | **채택** |
| 끝에 붙이기 대신 알파벳 위치 삽입 규칙 | 0 | #2226·#2233 같은 끝 충돌 분산 | **채택**(지침) |
| profile `known` 목록을 MANAGED_KEYS로 대체 + 주석 키 목록 삭제 + (key, additive) 단일 표 | 낮음, 동작 불변 | 6점 → 3점 | **채택** |
| 오류 kind 시험을 모듈별 자체 시험으로 | 낮음 | 공용 배열 추가 충돌 제거 | **채택** |
| 문서 drift 시험 + interfaces.md 손 개수 삭제 | 낮음 | 7개 누락 검출, 개수 줄 충돌 제거 | **#2246 D7에 합침**(진행 중, 중복 카드 없음) |
| ClaimRequest·GhTransport·HttpRequest 생성자/빌더 + tests/common 픽스처 | 중간 | 필드 추가 시 E0063 확산 방지 | **채택**(2단계) |
| 후보 병합 트리 검사(merge-tree 행렬) + 기준 SHA 변화 감지 | 낮음 | 충돌을 감사 전·병합 전에 발견 | **채택**: 게이트 스크립트에 이미 감사 head 기준 delta·상태 직전 head 재확인 반영, PR 등록 시 열린 PR 쌍 merge-tree 행렬 추가 |
| `inventory`/`linkme` 자동 등록 crate | 높음(새 의존성 = 사용자 결정) | OPERATIONS·dispatch 편집 제거 | **보류**(결정 필요) |
| 영역별 Subcommand flatten | 중간~높음, clap 4.6.1 동작 미검증 | op 추가가 영역 파일로 | **보류**(검증 후) |
| `.gitattributes merge=union` | 낮음 | .rs에 조용한 중복 위험 | **기각** |
| `..Default::default()` 리터럴 | 낮음 | 큰 오류를 조용한 기본값으로 바꿈 | **기각** |
| 전역 잠금·전체 직렬화 | - | 처리량 저하 | **기각**(사용자 조건) |

## 4. claim 선착순·--agent 강제(같은 과제)

기존 계약: AGENTS.md "가장 낮은 live 댓글 id 우선", lease 4h, 만료 자동 takeover 없음(operator 결정); 구현 #2161 queue-next(병합).
추가 설계:
- 댓글 필드: 레인 번호, session id(확인 못 하면 `unverified`), 이슈, 브랜치, 시각, epoch, lease, relay 여부.
- 순서 기준: 서버 등록 순서 = 댓글 id 오름차순(같은 id 불가 → 동률 없음). 등록 후 전체 댓글 재조회·정렬로 확인, 늦게 보인 경쟁 댓글은 재조회에서 판정.
- 쓰기 경계: commit·push 전에 `claim-verify`(소유권 재검증: 내 댓글이 최저 live id인가, lease 유효, 취소/인계 댓글 없음) — 실패 시 쓰기 거부. 댓글 자체를 원자적 잠금으로 주장하지 않음.
- 시간 만료만으로 활동 중 writer 강탈 금지; 인계는 명시적 handoff 댓글 또는 operator.
- 무효 댓글: 형식 불일치·위조(작성자 계정/세션 불일치)·중복 → 무시하고 보고.
- `--mode agent` 진입점(skills/moai mode-agent workflow + queue-next)에서 항상 적용; 문구만으로 완료 처리하지 않고 시험으로 확인.

## 5. 이슈·배정·순서

- 카드 A(코드, 1단계): main.rs operation 전수 match + 대응 시험 + 알파벳 삽입 규칙 / profile 편집점 축소 / 오류 kind 모듈별 시험. **통합 시점**: main.rs·profile.rs를 바꾸는 열린 PR(#2226 #2228 #2235 #2236 #2248 #2251)이 병합된 뒤 짧은 창 하나 — 그 전에는 설계·시험 준비만.
- 카드 B(코드, 2단계): 구조체 생성자/빌더 + tests/common.
- 카드 C(claim·--agent 강제): §4.
- #2246 D7: 문서 drift 시험(진행 중, agent-9) — A와 합류.
- 게이트: merge-tree 행렬 자동화(리더 도구).

## 6. 전후 비교 방법

같은 지표(§1 표)를 카드 A·B 병합 후 같은 길이의 관측 구간에서 다시 세어 비교. 대기시간은 장부 시각을 분 단위로 기록하도록 바꾼 뒤부터 비교(이전 구간은 not_run).
