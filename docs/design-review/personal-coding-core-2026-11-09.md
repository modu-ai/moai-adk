# 개인 맞춤 코딩 코어 — 설계 반영안 (ADK_PERSONAL_CODING_CORE_20261109_01)

- 지시 수신: 2026-11-09 01:1xZ. 근거 기록: `docs/decisions-2026-11-08.md` 01:1x~01:2x(11-09) 행(5단계 확정, 학습 루프 승인 수준, 이름 재검토 방식, MoAI Atlas 명칭).
- 대조 기준: `origin/v2026.11.0` a76135235. 설계 원본: 환골탈태 통합 최종 설계(22절) + 하위 19편 + 종합 보고서(`synthesis-improvement-report-2026-11-09.md`).
- 이 문서는 **로컬 설계 수정안**이다. ChatGPT Space 페이지 반영은 EP-G(#2159) 카드로 별도 기록하며, 두 가지를 같은 완료로 취급하지 않는다. 여기서 "구현됨/부분/설계만/제안"은 각각 소스 인용, 부분 인용, 설계 문서만, 이번 지시로 새로 제안된 것을 뜻한다.

## 1. 정체성 문장(제안)

MoAI ADK는 **개인에게 맞춰지는 AI 코딩 에이전트 도구**다. Claude Code·Codex 등 특정 호스트에 종속되지 않고, 프로젝트에 맞는 개발 워크플로우(5단계)·에이전트·스킬을 생성·갱신한다. 호스트는 발견·호출·native 기능 차이만 다루는 얇은 어댑터이며(설계 §2·§4, [dff] §1), 개인화의 핵심 산출물은 프로젝트의 `AGENTS.md`(공통) + `AGENTS.override.md`(프로젝트 맞춤)와 생성된 에이전트·스킬이다.

영향: 상위 설계 §1 "통합 결정과 제품 경계" 첫 표에 "제품 정체성" 행 추가(제안), §3 "개인 맞춤 프로젝트 시작"의 목적 문장 보강(제안).

## 2. 5단계 개발 과정 — 확정

| 단계 | 역할 | RC1 실물 | 상태 |
|---|---|---|---|
| PLAN | SPEC(GEARS)·승인 snapshot·task DAG admission | `src/spec.rs`, `src/plan.rs`, route PLAN(`skills/moai/references/routes.json`) | 구현됨 |
| RUN | 전략별 구현·evidence(`src/run.rs` STRATEGIES 6종) | `src/run.rs`, route RUN | 구현됨 |
| VERIFY | 검증 명령 argv 실행·FailureClass | `src/verify.rs`, route VERIFY | 구현됨 |
| AUDIT | 독립 감사·admission 6규칙 | `src/audit.rs`, route AUDIT | 구현됨 |
| SYNC | drift 분류·Issue 보고·통합 | `skills/moai/workflows/sync.md`, `src/forge/*`, route SYNC | 구현됨(절차+gh) |

- 사용자 확정(01:2x): **PLAN → RUN → VERIFY → AUDIT → SYNC**. READ·SETUP·UPDATE는 "진입·설치 단계"로 별도 분류하며 5단계에 넣지 않는다.
- 설계 수정안: 상위 §2·§5가 이미 이 다섯을 쓴다(종합 보고서 S9는 §2에 READ·UPDATE가 빠진 것을 지적 — 수정안은 "진입 3 + 개발 5"로 적는다). [43fa] §1 L1896 "plan → run → verify → audit → sync" 5단계 표기와 일치. Go 세대 plan/run/sync 3단계는 역사로 표시.
- 영향 카드: 없음(이름 변경 0). DR-04(#2140) 수정안 항목 S9의 교체문에 "진입 3 + 개발 5" 문장을 반영(EP-G 적용 시).

## 3. AGENTS 계층 — AGENTS.md(공통) + AGENTS.override.md(프로젝트 맞춤)

### 3.1 현재 실물
- `AGENTS.md`: 호스트 중립 공통 규칙 + 관리 구간(`moai-cli agents-apply`가 재생성, `src/agents.rs:1-6` 3-way 보존). 구현됨.
- `AGENTS.override.md`: RC1 코드에 **파일명 인지**는 있음 — `src/agents.rs:369`(이름 목록), `src/hosts/codex.rs:274-275`(Codex 전역 우선순위 override → AGENTS.md). **생성 경로는 없음**(`agents-plan/apply`는 AGENTS.md 관리 구간만 — `src/main.rs` OPERATIONS 표 기준, 가설 아님: 표 전수에 override 생성 op 부재).
- 호스트 공식 지원(02-C·02-B 계약 문서, 2026-11-08 확인): **Codex** — `AGENTS.override.md`를 전역·프로젝트 레벨에서 공식 로드(learn.chatgpt.com/docs/agent-configuration/agents-md). **Claude Code** — AGENTS.md 자체 로딩은 v2.1.277+이고 CLAUDE.md 변형이 없을 때만; `AGENTS.override.md`·`AGENTS.local.md`는 "Claude 동등물 아님"(`AGENTS.md:53`). 따라서 Claude에는 **어댑터가 필요**하다.

### 3.2 설계안(제안)
| 구분 | 내용 | 분류 |
|---|---|---|
| 생성 | `/moai project`(SETUP 진입) 끝에 `AGENTS.override.md`를 생성: 딥 인터뷰 profile(`src/profile.rs` 관리 키) + 저장소 조사 결과 → 프로젝트 규칙·워크플로우 조정·에이전트/스킬 목록. 관리 구간 마커로 감싸고 사용자 구간은 비움 | 제안 |
| Codex 배치 | 프로젝트 루트 `AGENTS.override.md` 그대로(공식 우선순위). 32 KiB 예산(`project_doc_max_bytes`)에 AGENTS.md와 합산 | 공식 지원 |
| Claude 배치 | 선택지 A: `CLAUDE.md`에 `@AGENTS.md`·`@AGENTS.override.md` import 줄(02-B doctor가 억제 조건 검사); B: `.claude/`에 projection 파일 생성(동기화 receipt). 어느 쪽이든 "자동 로드 가정 금지" — doctor로 확인 | 어댑터 필요(결정 항목 Q-A) |
| 기타 호스트 | capability 상태(unknown→documented→probed→certified)로만 표시, 지원표 행 | 설계 §2 |
| 보존 | 사용자 구간·기존 설정은 3-way 보존(`src/apply.rs`, `src/agents.rs` 기존 계약 재사용), 충돌은 diff 반환·쓰기 0 | 구현됨(재사용) |

## 4. 재귀 자기 개선 루프 — 설계 경계

사용자 확정(01:2x): **제안은 자동, 적용은 사용자 승인 1회.**

| 요소 | 설계 | 경계 |
|---|---|---|
| 입력 | 개발 사이클 증거만: receipt(`src/run.rs` evidence), 감사 verdict, drift 분류, 사용자 교정 기록(auto-memory `feedback_*`) | 대화 전문·비밀·개인 경로는 입력 금지(설계 §3 L33, REQ-05A-08) |
| 제안 생성 | `override-propose`(제안 op): 교훈 → override 관리 구간 diff + 근거(어느 receipt/verdict) + 검증 결과(문서 lint·예산·금지어) + 영향 범위 | 자동. 쓰기 0(제안 파일만 `.moai-cli/proposals/<run>.md`) |
| 적용 | `override-apply`: 사용자 승인 뒤 관리 구간만 교체, 이전 판 `.moai-cli/agents-history/<ts>.md` 보존, receipt 기록 | 승인 없이는 어떤 경로로도 적용 0; 사용자 구간은 **절대 자동 변경 0** |
| 롤백 | `override-rollback <ts>`: history에서 복원, receipt | 되돌릴 수 있는 경계 = 관리 구간 |
| 충돌 | 사용자가 같은 줄을 고쳤으면 제안은 `conflict`로 중단, diff 반환 | 3-way 규칙 재사용 |
| 권한 | 학습은 AGENTS 텍스트만 바꾼다 — 호스트 권한·hook·설정·HOLD는 손대지 않음 | "학습 ≠ 권한 확대·보호 우회" |
| 검증 | 제안마다: 예산(32 KiB), 비밀 0, 개인 경로 0, 금지 지시(권한 요구·우회) 0, 5단계 어휘 일치 | 기계 검사 + 감사 가능 |

분류: 전부 **제안**(구현 0). 저장 위치 이름(`proposals/`, `agents-history/`)은 홈 레이아웃 P09(#2175)와 함께 결정.

## 5. 프로젝트 맞춤 에이전트·스킬 구성

- 현재: `src/generation.rs`가 역할 파일을 호스트별 경로에 생성(`.claude/agents/<role>.md`, `.codex/agents/<role>.toml`), 외부 스킬 provenance는 03-B(`src/sources.rs`). 구현됨.
- 설계안(제안): `/moai project`에서 (1) 저장소 조사(언어·프레임워크·테스트 도구·CI) → (2) 최신 적합 자료 조사(공식 문서 URL+확인일을 `manifests/sources.json`에 기록 — 03-B 재사용) → (3) 에이전트·스킬 선택 생성 → (4) 호스트별 공식 경로 배치(Claude `.claude/skills|agents`, Codex `.agents/skills`·`.codex/agents`) → (5) doctor로 로딩 확인. "모든 호스트가 자동 로드한다고 가정하지 않음" — 지원표에 documented/probed/certified로만.
- 영향 카드: EP-B B3 #2170(SETUP 승인 분리)와 같은 진입 흐름 — **B3 소유자(agent-5)에게 수정 요청 댓글**로 "override 생성 단계는 EP-H 카드가 맡음, B3는 commit·Issue 발행 승인만" 경계 명시(중복 작성자 0).

## 6. 이름 재검토(ADK)

- 사용자 확정(01:2x): 조사 카드 1장 → 후보 3~5개(충돌 조사 URL+확인일) 보고, 확정은 사용자. **rename 0**, "Agentic Development Kit" 유지.
- 조사 범위 제안: 현재 풀이와 정체성의 거리, 후보별 패키지·도메인·상표 조사(404 = 미등록 관측일 뿐 충돌 0 아님), 기존 사용자 혼동 비용.

## 7. 다른 결정과의 연결

- MoAI Atlas(01:2x): Code Map/Graph 묶음 명칭만 확정. 패키지·CLI·태그 문법·상표는 미확정 — C2(#2178) 이후 카드에서 별도.
- 자율 모드(EP-A): 5단계는 에이전트 self-claim 카드 안의 작업 순서, 모드는 그 바깥의 운영 계층 — 충돌 없음.
- 보존 경계 불변: Go `moai`·`.moai`·`~/.moai`·`.claude/` 무변경, HOLD, 감사·exact-SHA·PR 보호 유지.

## 8. 영향 카드와 신규 카드(EP-H, 미발행)

| 카드 | 내용 | 겹침/소유 |
|---|---|---|
| H1 | `/moai project` → `AGENTS.override.md` 생성(관리 구간, profile 연동) | B3 #2170(agent-5)과 진입 흐름 공유 — 경계 댓글 |
| H2 | Claude 호스트 어댑터: CLAUDE.md import 또는 projection + doctor 확인(Q-A 결정 뒤) | 02-B 계약(#1992 종료) 확장 |
| H3 | 학습 루프 op 3종(propose/apply/rollback) + history/receipt | P09(#2175)와 저장 위치 공유 |
| H4 | 제안 검증기(예산·비밀·개인 경로·금지 지시·5단계 어휘) | 05-A REQ-05A-08 재사용 |
| H5 | 맞춤 에이전트·스킬 구성 조사 단계(공식 문서 URL+확인일 기록) | 03-B(`src/sources.rs`) 재사용 |
| H6 | ADK 이름 재검토 조사 | F6/C1과 같은 "조사 카드" 형식 |
| G-추가 | Space 반영: §1 정체성 행·§3 override 절·5단계 "진입 3+개발 5" 문장 | EP-G(#2159) 리더 |

## 9. 남은 결정(사용자)

- Q-A Claude 어댑터 방식: CLAUDE.md import 줄 vs `.claude/` projection(둘 다 공식 자동 로드 아님 → doctor 확인 필수).
- Q-B 학습 입력 범위: receipt·verdict·drift·교정 기록 외에 "세션 요약"을 넣을지(개인정보·비밀 위험 ↑).
- Q-C EP-H 발행 시점: #2213 병합 뒤 즉시 vs B3(#2170) 병합 뒤.

## 10. 한계

- 호스트 공식 문서는 2026-11-08 확인본(02-B·02-C 계약)에 의존, 이번에 재조회하지 않았다.
- 코드 인용은 origin tip a76135235 기준이며 `cargo` 실행·바이너리 검증은 하지 않았다.
