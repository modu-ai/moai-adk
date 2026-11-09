#!/usr/bin/env python3 -I
"""rc.2 Issue generator — leader-authored card specs → GitHub Issues + body mirrors.

Usage: python3 -I rc2_issue_gen.py [--dry-run]
Writes .moai-cli/reports/issue-bodies/<n>-<key>.md and prints a JSON map key→number.
Source of truth for scope: ~/Downloads/MoAI ADK 환골탈태 통합 최종 설계.md (22 sections) + decisions 2026-11-09 00:4x~01:2x.
"""
import json
import pathlib
import re
import subprocess
import sys

REPO = "modu-ai/moai-adk"
MILESTONE = "2026.11.0-rc.2"
TIP = "110b375c7"
DRY = "--dry-run" in sys.argv
OUT = pathlib.Path(".moai-cli/reports/issue-bodies")

COMMON = """
## 공통 규칙
- 한 Issue = 한 세션. 시작 전 `moai-claim: session=<id> epoch=1 lease_until=<utc+4h>` 댓글(리더 중계 또는 self-claim), 가장 낮은 live 댓글 id가 소유자. id 회신 전 커밋 금지, trailer `Moai-Claim: <issue>#<comment-id>`를 모든 커밋과 PR 본문에.
- branch `issue/<n>-<key>`, worktree `<primary>/.moai-cli/worktrees/<n>-<key>` (EnterWorktree/ExitWorktree(keep), 셸 cd 금지). base는 최신 `origin/v2026.11.0`.
- 완료 = check set 실측 exit 기록 + PR(`Refs #<n>`) + 영속 보고 → 독립 감사(agent-2) PASS → `moai/audit` → 리더 병합. 미실행 검사는 not_run으로 적는다.
- 설계 ↔ RC1 결정 충돌은 RC1 결정 우선(decisions 00:4x(11-09)). 개인 절대 경로·비밀 0.
"""

EPS = {
 "EP-A": ("자율 무중단 모드·Factory", "설계 §6 운영 + 종합 보고서 §10.4 + DR-02 SPEC(#2138, docs/design-review/autonomy-modes-spec.md)",
   "세션 안 `/moai --mode leader|agent [--autonomy full|semi]`로 리더·에이전트 모드에 들어가, 에이전트는 Issue를 self-claim해 끝까지 가고(PR 뒤 대기 없이 다음 Issue), 리더는 배차 대신 감독(300초 PR 점검·수리 재배차·배정 기록·`/goal` 안내)한다. 결정: D-1 론처 -l/-a 폐기 · D-2 300초 설정값 · D-3 Claude만 리더 · D-4 원격 Issue 정본/로컬 캐시 · lease 4h 수동 인계."),
 "EP-B": ("online-first 잔여 — SETUP·SYNC·협업·저장소", "설계 §3·§5·§6·§13; 대조표 §4 미발행 항목",
   "RC1이 남긴 online-first 공백: SYNC drift 4종, owner 4필드, SETUP의 승인된 초기 commit·Issue 발행, GitLab adapter 계약, sub-issue/dependency probe, audit-only GitHub App, webhook·MCP 조사, 홈 레이아웃 P09 이행."),
 "EP-C": ("Core Brain · Code Atlas", "설계 §7~§11; 결정: 묶음명 Code Atlas(충돌 확인 선행) + `@MOAI:` dual-read, budget 기본 256 KiB",
   "Code Map+Code Graph 묶음(Code Atlas), `@MOAI:` 태그 grammar와 `@MX:` dual-read, 온톨로지 schema, 명시 SPEC link, 증분 색인, 비벡터 검색·budget, 조회 op, 회귀 시험군. JSON 정본(TOON 도입 안 함, 01:2x)."),
 "EP-D": ("보존·아카이브·품질", "설계 §12·§16·§17; 결정: 자동 이동 허용(hash/readback 필수), terminal+90일",
   "archive 적격 판정·dry-run·자동 이동·restore·검색, ARCHIVE·PRIVACY 시험군, benchmark baseline, 비용 분해, 문서/소스/template/live 의미 parity."),
 "EP-E": ("릴리스·플랫폼", "설계 §18 '남은 결정'; 결정(01:1x): macOS+Linux x86_64/arm64, Windows 빌드+self-replace만, attestation+sha256(notarization 없음), SBOM=cargo-dist 산출물, tag ruleset `moai-cli-v*`",
   "cargo-dist 도입, Linux CI matrix, Windows 빌드·self-replace, attestation 실측, tag ruleset, RELEASE 시험군. release/deploy 자체는 여전히 사용자 승인 밖(카드는 준비까지)."),
 "EP-F": ("후순위 — Pi·Desktop·이미지·판단 backend", "설계 §14·F21·F20; 결정: Pi spike 뒤 채택 결정, 판단 backend는 조사 뒤",
   "설계가 'RC1 선행조건 아님'으로 둔 영역. 전부 발행하되 선행 카드(F1 spike, F6 조사)에 blocked. 비용 드는 외부 호출 0."),
 "EP-G": ("설계 문서 정합 적용(Space)", "design-docs-r2 #2140~#2143 산출물(docs/design-review/fixes·appendix) → ChatGPT Space 페이지 적용; 결정: 리더가 ego-browser로 적용, 페이지당 적용 전 사용자 확인",
   "에이전트는 Space를 편집할 수 없으므로 적용은 리더(ego-browser)가 하고, 적용 후 사본 sha256을 기록한다. TOON은 제외안 적용."),
}

# key, EP, title, design_ref, scope(list), reqs(list), acs(list), blocked_by(list of keys or '#n'), decisions
CARDS = [
 # ---------------- EP-A ----------------
 ("A1","EP-A","`/moai --mode leader|agent [--autonomy full|semi]` skill 인자와 진입 절차","보고서 §10.4; autonomy-modes-spec.md §2",
  ["`skills/moai/SKILL.md`·`references/routes.json`에 `--mode` 인자(leader|agent), `--autonomy`(full|semi, leader만) 정의와 진입 workflow `workflows/mode-leader.md`·`mode-agent.md`",
   "역할 없는 세션에서 입력 시: 모드 배너 출력 → 현재 상태(열린 Issue·PR·assignments 스냅샷) 요약 → 모드별 첫 행동(leader: §A4 `/goal` 안내, agent: self-claim 1회)",
   "Codex는 `$moai --mode agent`만(D-3). `--mode leader`를 Codex에서 입력하면 거절 메시지+근거",
   "source-manifest.json sha 갱신(skills/moai 편집 규칙)"],
  ["REQ-A1-01: WHEN 사용자가 `/moai --mode leader` 또는 `--mode agent`를 입력하면 THE SKILL SHALL 해당 모드 workflow를 로드하고 다른 모드 지시를 무시한다.",
   "REQ-A1-02: WHERE host가 codex이고 `--mode leader`이면 THE SKILL SHALL 진입을 거절하고 D-3 결정을 인용한다."],
  ["AC-1: routes.json에 두 route 추가, `tests/` 라우팅 시험 2건(leader/agent) + codex 거절 1건 passed",
   "AC-2: 두 workflow 문서가 self-claim 7단계·감독 루프를 autonomy-modes-spec.md 절 번호로 참조(중복 서술 0)",
   "AC-3: source-manifest sha 일치 시험 passed, check set exit 기록"],
  [], "00:0x·01:0x(11-09) D-1~D-3"),
 ("A2","EP-A","에이전트 self-claim 루프 — `queue-next` op(원격 미배차 재계산 → journal+lock → claim 댓글 → 최소 id → projection)","autonomy-modes-spec.md §3; `src/claim.rs:1-23`",
  ["`src/main.rs` OPERATIONS에 `queue-next`(`queue.next`) 추가: milestone의 열린 Issue 중 claim 댓글 없는 것(원격 정본, D-4) → 우선순위(blocked 제외·EP 순) → 후보 1건 → 기존 `claim-acquire` 호출 → 승/패 결과 JSON",
   "패배(already_owned)는 journal에 쓰지 않는 현행 유지(claim.rs:906-911); 결과 JSON에 `lost_to: <comment-id>`만 반환(상태 추가 없음)",
   "승자가 label `moai:in-progress`·소유 댓글 `상태:` 줄 갱신 — projection 전용, 실패해도 claim 유효",
   "`assignments.json`은 읽기만(레인), 쓰기 금지"],
  ["REQ-A2-01: WHEN `queue-next`가 호출되면 THE SYSTEM SHALL 원격 Issue 목록을 다시 읽어 claim 댓글 없는 Issue만 후보로 삼는다.",
   "REQ-A2-02: WHEN 두 세션이 같은 Issue를 집으면 THE SYSTEM SHALL 가장 낮은 live 댓글 id를 승자로 하고 패자는 다음 후보로 넘어간다(자동 takeover 0).",
   "REQ-A2-03: THE SYSTEM SHALL label·assignee 변경 실패를 claim 실패로 취급하지 않는다."],
  ["AC-1: RecordedTransport fixture로 두 세션 경합 시험 — 승자 1·패자 다음 후보 이동·중복 claim 0",
   "AC-2: blocked 카드(본문 `Blocked-By:` 헤더)가 후보에서 제외되는 시험",
   "AC-3: 패자 journal 미기록 확인 시험(현행 보존)",
   "AC-4: check set exit, OPERATIONS 수 갱신(47→48) 문서 반영"],
  ["A1"], "00:3x(11-09) self-claim; D-4"),
 ("A3","EP-A","리더 감독 루프 — 300초 설정키, PR 점검, 수리 재배차 `assignee-hint`, assignments 스냅샷 커밋","autonomy-modes-spec.md §4; AGENTS.local.md § Assignment record and supervision(#2144)",
  ["설정 키 `leader.pr_check_interval_secs`(기본 300, 문서에 '사용자 지정값·측정값 아님') — `.moai-cli/project.yaml` profile 관리 키 추가",
   "`leader-tick` op: 열린 PR 상태(CI·moai/audit·충돌·BEHIND)·claim 댓글·assignments 대조 → 변경 요약 JSON; 변화 없는 PR은 보고 생략(폴링 증식 금지)",
   "수리 필요 PR: Issue 댓글에 `assignee-hint: agent-N`(원 작성자 아닌 레인) + 수리 Issue 본문 초안 생성(발행은 리더 승인)",
   "`assignments.json` 스냅샷 재생성(리더만 커밋)"],
  ["REQ-A3-01: WHEN 간격이 지나면 THE LEADER SHALL 열린 PR과 claim 댓글을 다시 읽고 변화가 있을 때만 보고한다.",
   "REQ-A3-02: WHEN PR이 FAIL verdict 또는 CI 실패이면 THE LEADER SHALL 원 작성자 아닌 레인에 assignee-hint를 남긴다(소유권은 claim 댓글로만)."],
  ["AC-1: fixture로 PR 상태 변화 유무에 따른 보고/생략 시험 2건","AC-2: assignee-hint 생성 시험(원 작성자 제외)","AC-3: profile 키 추가 후 `profile.rs` 관리 키 시험 갱신, check set exit"],
  ["A2"], "D-2 300초; 00:3x 리더 감독"),
 ("A4","EP-A","`/goal` 안내 생성 — Issue·PR 분석 → 완료 조건 문장 → 붙여넣기 블록(Claude 내장 /goal·Codex /goal)","autonomy-modes-spec.md §7; #2145 지원표 goal 행",
  ["리더 모드 진입 시: milestone 열린 Issue·PR을 읽어 완료 조건 문장 생성(예: 'milestone X의 Issue·PR 모두 병합·종료, #2016 제외') → 사용자에게 `/goal <문장>` 붙여넣기 블록 출력(모델이 대신 입력 불가 — 확인됨)",
   "에이전트 모드: 자기 카드 범위 goal 문장(예: '#<n> PR 병합 또는 blocked 보고') 출력; Codex는 `/goal` 존재 확인됨(learn.chatgpt.com) — 같은 블록 제공",
   "제외 규칙: 운영자 보유·keep-set Issue는 조건에서 빼고 사유를 적는다"],
  ["REQ-A4-01: WHEN 리더 모드가 시작되면 THE SKILL SHALL 현재 열린 Issue·PR로부터 완료 조건 문장을 만들고 붙여넣기 블록으로 제시한다.",
   "REQ-A4-02: THE SKILL SHALL `/goal`을 대신 실행했다고 주장하지 않는다."],
  ["AC-1: fixture 3종(열린 Issue만/PR만/keep-set 포함)에서 생성 문장 snapshot 시험","AC-2: 문서에 '모델 입력 불가, 붙여넣기' 명시 + 지원표 goal 행 참조","AC-3: check set exit"],
  ["A1","#2145"], "00:1x /goal 붙여넣기 안내"),
 ("A5","EP-A","론처 `-l`/`-a` deprecate — 경고 → 제거, `--name` 유지, 문서·지원표·AGENTS.local 갱신","`src/main.rs` Cli/HostLaunch; 00-F #2028; #2053",
  ["1단계(이 카드): `-l`/`-a` 사용 시 stderr 경고('deprecated: use `/moai --mode leader|agent` inside the session') + notes에 기록, 동작은 유지; `--name` 독립 옵션 유지",
   "문서: docs/user-entry-draft.md·host-support-matrix.md 론처 표를 '역할 없는 세션 + /moai --mode' 중심으로; AGENTS.local.md § Sessions and roles 갱신",
   "2단계(후속 카드 후보): 제거 시점은 EP-A A1~A3 병합 뒤"],
  ["REQ-A5-01: WHEN `-l` 또는 `-a`가 주어지면 THE LAUNCHER SHALL 경고를 출력하고 기존 동작을 유지한다.","REQ-A5-02: THE LAUNCHER SHALL 역할 없는 호출의 동작(#2053)을 바꾸지 않는다."],
  ["AC-1: `--dry-run` JSON notes에 deprecation 문구 시험 2건","AC-2: 역할 없는 호출 기존 시험 불변","AC-3: 문서 3곳 갱신, check set exit"],
  [], "D-1 폐기"),
 ("A6","EP-A","같은 session 값 claim 충돌 방지 — REQ-AM-04(session id에 세션 ref 포함, 동명 세션 adopt 금지)","autonomy-modes-spec.md REQ-AM-04; `src/claim.rs:829-845`",
  ["claim 요청의 `owner.id`에 세션 고유 ref(예: `agent-1-e29c15`)를 필수화하거나, 동일 session 문자열의 기존 claim을 adopt하지 않고 `uncertain`으로 반환하는 옵션 중 하나(레인 판단, 근거 기록)",
   "오늘 실례(agent-1 중복 세션)를 fixture로"],
  ["REQ-A6-01: WHEN 같은 session 값의 claim 댓글이 이미 있고 요청자가 그 댓글의 작성 세션임을 증명할 수 없으면 THE SYSTEM SHALL adopt하지 않는다."],
  ["AC-1: 동명 세션 fixture에서 adopt 0·uncertain 또는 ref 불일치 거절","AC-2: 기존 claim 시험 불변, check set exit"],
  [], "01:0x REQ-AM-04"),
 ("A7","EP-A","lease 기본 4h 상수·만료 표시·수동 인계 절차 문서","decisions 00:5x(11-09) P08 잔여; `src/claim.rs:25-28`",
  ["`lease_until` 미지정 시 기본 now+4h(상수 `DEFAULT_LEASE`), 명시값 우선; 만료는 `claim-show`에 `expired: true`로 표시만, 자동 takeover 0(현행)",
   "handoff 절차 문서: 만료 확인 → 리더가 epoch+1 인계 → 기록"],
  ["REQ-A7-01: WHEN lease_until이 없으면 THE SYSTEM SHALL 4시간 기본값을 쓴다.","REQ-A7-02: THE SYSTEM SHALL 만료만으로 소유권을 옮기지 않는다."],
  ["AC-1: 기본값 시험·명시값 우선 시험","AC-2: 만료 표시 시험 + takeover 0 기존 시험 불변","AC-3: check set exit"],
  [], "00:5x lease 4h"),
 ("A8","EP-A","Factory 문서 통합 — factory-dispatch '확정 보고서→카드', lane loop·감독 절 정리, 임시 절 제거","AGENTS.local.md(§ Lane loop, § Assignment record and supervision); 보고서 §10.2 F-E",
  ["AGENTS.local.md를 '세션·역할 / 에이전트 루프(self-claim) / 리더 감독 / 게이트' 4절로 재구성, 중복 문장 제거, 32 KiB 예산 확인",
   "docs/lane-loop-contract.md와 autonomy-modes-spec.md 상호 참조",
   "코드 변경 0"],
  ["REQ-A8-01: THE DOCS SHALL self-claim·감독·게이트 규칙을 각 한 곳에서만 정의하고 나머지는 참조한다."],
  ["AC-1: 규칙 중복 grep(같은 문장 2곳 이상) 0","AC-2: AGENTS.local.md+AGENTS.md 합계 < 32 KiB","AC-3: check set exit"],
  ["A1","A2","A3"], "00:4x AGENTS.local 추가 지시"),
 # ---------------- EP-B ----------------
 ("B1","EP-B","SYNC drift 7종 중 미구현 4종 — SPEC↔oracle 충돌·참조 단절·Issue 완료인데 증거 stale·명시 link 의미 미검토","설계 §5 drift 7분류; `skills/moai/workflows/sync.md:20`(현 3종)",
  ["sync 절차에 4종 판정 규칙·출력 어휘 추가(기존 3종과 대응표), 각 종별 fixture","코드 op 추가 여부는 레인 판단(절차+gh 유지 가능)"],
  ["REQ-B1-01: WHEN drift 7종 중 하나가 탐지되면 THE SYNC SHALL 종류·근거·제안 조치를 분리해 보고하고 승인 SPEC을 자동 약화하지 않는다."],
  ["AC-1: 4종 fixture 각 1건 탐지","AC-2: 상위 7종↔RC1 어휘 대응표(docs/design-review/appendix 참조)","AC-3: check set exit"],
  [], "RC1 결정 우선"),
 ("B2","EP-B","실행 소유권 4필드 — assignee / execution_owner / provider_session / integration_owner schema + claim record","설계 §6 L756; `src/claim.rs` record",
  ["claim record·handoff record에 4필드 추가(기존 `owner`는 execution_owner로 별칭), integration_owner는 리더 게이트 식별자","assignee는 표시 전용(lock 아님) 명시"],
  ["REQ-B2-01: THE RECORD SHALL 네 필드를 분리 저장하고 어느 것도 lock으로 해석하지 않는다."],
  ["AC-1: schema 시험 + 기존 claim 시험 불변","AC-2: 문서(protocol-v1.md) 갱신","AC-3: check set exit"],
  [], "RC1 06-A 유지"),
 ("B3","EP-B","SETUP의 '허용된 초기 local commit'과 '검토된 Issue 발행' 단계 — 승인 분리","설계 §3 L294-295·L319-321; `skills/moai/workflows/setup.md`(단계 없음)",
  ["setup 절차에 두 단계 추가: 각각 별도 승인(commit 승인 ≠ Issue 공개 승인), dry-run 출력 먼저","receipt에 두 효과 기록"],
  ["REQ-B3-01: WHEN 사용자가 초기 commit을 승인하면 THE SETUP SHALL commit만 만들고 push·Issue 발행은 하지 않는다.","REQ-B3-02: WHEN Issue 발행이 승인되면 THE SETUP SHALL 검토된 초안만 발행하고 readback으로 확인한다."],
  ["AC-1: fixture로 승인 없이 commit/발행 0","AC-2: readback 확인 시험","AC-3: check set exit"],
  [], "승인 분리 원칙"),
 ("B4","EP-B","GitLab Forge adapter 계약 — trait+fixture(MR·pipeline·note), 실제 구현은 blocked","설계 §6 Forge Adapter; `docs/rust-product-path.md:238`",
  ["Forge trait에 GitLab capability 선언(unknown/documented), fixture 스키마만","지원표에 GitLab 행 'documented, not_run'"],
  ["REQ-B4-01: THE ADAPTER SHALL GitLab을 capability 선언으로만 노출하고 실호출 경로를 만들지 않는다."],
  ["AC-1: trait 컴파일·fixture 로드 시험","AC-2: 지원표 행","AC-3: check set exit"],
  [], "01:1x GitLab 계약만(계정 없음)"),
 ("B5","EP-B","sub-issue·dependency capability probe + 본문 링크 fallback","설계 §6; `tests/fixtures/github/{sub-issues,dependencies}.json`",
  ["probe op: 저장소에서 sub-issue/dependency API 사용 가능 여부(권한·plan) 읽기 → documented/probed 기록","불가 시 Issue 본문 `Blocked-By:`/`Parent:` 링크 fallback 사용"],
  ["REQ-B5-01: WHEN probe가 실패하면 THE SYSTEM SHALL 본문 링크 fallback을 쓰고 API 쓰기를 시도하지 않는다."],
  ["AC-1: fixture 양쪽 경로 시험","AC-2: 한도 수치(100/8/50)는 공식 URL+확인일 또는 미확인 표기","AC-3: check set exit"],
  [], "01:1x probe 후 사용"),
 ("B6","EP-B","audit-only GitHub App — `moai/audit` status 출처 고정(운영자 App 등록 keep-set)","AGENTS.local.md 게이트 known gap; decisions 01:1x",
  ["App 권한 최소(statuses:write, pull_requests:read) 설계 문서 + 토큰 주입 경로(환경 변수 이름만, 값 저장 0)","status 게시 코드가 App 토큰이면 `context`에 출처 표시; 공유 계정이면 경고","App 등록·설치는 운영자 1회(keep-set) — 카드는 준비까지"],
  ["REQ-B6-01: WHEN App 토큰이 있으면 THE GATE SHALL status를 App으로 게시하고 없으면 경고 후 기존 경로."],
  ["AC-1: 두 경로 fixture 시험","AC-2: 비밀 값 로그 0","AC-3: check set exit"],
  [], "01:1x audit App"),
 ("B7","EP-B","webhook 수신 조사 — 서명 검증·dedup·역순 처리 요구와 service-free 전제의 양립 조건(구현 blocked)","설계 §6 L164; [0d45] L164",
  ["조사 문서: 수신자 없는 설계에서 webhook이 성립하는 조건(사용자 기존 수신 환경), GitHub 재전송 정책(공식 URL+확인일), polling과의 역할 분담","구현 카드는 조사 결론 뒤"],
  ["REQ-B7-01: THE REPORT SHALL 외부 사실마다 URL과 확인일을 붙이고 확인 못 한 것은 미확인으로 둔다."],
  ["AC-1: 외부 사실 전부 URL+날짜 또는 미확인","AC-2: 결론 1절(채택/보류)","AC-3: check set exit"],
  [], "조사 우선"),
 ("B8","EP-B","홈 레이아웃 P09 이행 — receipt/outbox/claim journal의 `~/.moai-cli/projects/<id>/` opt-in 이동","설계 §13; `src/apply.rs:29` swap point; decisions 78·86③",
  ["P09 결정안 문서(홈 트리 실물 정의) + opt-in 이동 op(dry-run→apply, 기존 프로젝트 쪽 경로는 읽기 호환)","기본값은 현행(프로젝트 쪽) 유지"],
  ["REQ-B8-01: WHEN 이동이 opt-in되면 THE SYSTEM SHALL 원본을 지우지 않고 복사·검증 후 포인터를 바꾼다."],
  ["AC-1: dry-run/apply fixture, 되돌리기 시험","AC-2: 기본 경로 불변 시험","AC-3: check set exit"],
  [], "RC1 결정 우선(프로젝트 쪽 기본)"),
 ("B9","EP-B","MCP parity 조사 — RC1 CLI op ↔ Go MCP 38 tool 대응표(구현 blocked)","설계 §1·§13; decisions 51(MCP 다음 버전)",
  ["대응표: 각 OPERATIONS op에 대응 MCP tool 유무, 입력/출력 schema 차이, structuredContent 요구","결론: rc.2 안 MCP 범위 제안(구현은 별도 카드)"],
  ["REQ-B9-01: THE REPORT SHALL 38 tool 전수를 다룬다."],
  ["AC-1: 38행 전수","AC-2: 결론 1절","AC-3: check set exit"],
  [], "MCP 다음 버전"),
 # ---------------- EP-C ----------------
 ("C1","EP-C","Code Atlas 이름 충돌 확인(패키지·상표·명령) + 명칭 확정 기록","설계 §9 L10 '확정 명칭으로 배포하지 않음'; decisions 00:5x",
  ["crates.io·npm·PyPI·GitHub org·상표 DB(공식 URL)에서 'code-atlas'/'Code Atlas' 충돌 조사, CLI 서브커맨드명 후보 비교","결론 문서 → 사용자 확정(리더 중계)"],
  ["REQ-C1-01: THE REPORT SHALL 조사 출처마다 URL·확인일을 적고 충돌 여부를 판정한다."],
  ["AC-1: 출처 ≥4, 전부 URL+날짜","AC-2: 대안 이름 ≥2","AC-3: check set exit"],
  [], "00:5x Code Atlas"),
 ("C2","EP-C","`@MOAI:` grammar + parser + `@MX:` dual-read 스캐너","설계 §9 L24(dual-read→normalization→진단→opt-in migration); Core Brain L3470-3475",
  ["`@MOAI:<KIND>[:<sub>] <args>` grammar 정의(상위 §9와 Core Brain 예시 통일 — DR-07 부록 참조), parser version 필드","스캐너: `@MX:`·`@MOAI:` 둘 다 읽어 같은 노드로 정규화, 쓰기 migration은 하지 않음"],
  ["REQ-C2-01: WHEN 파일에 두 namespace가 섞여 있으면 THE SCANNER SHALL 둘 다 읽고 중복을 진단만 한다.","REQ-C2-02: THE SCANNER SHALL 기존 파일을 자동 변경하지 않는다."],
  ["AC-1: grammar 시험(정상·오류 각 ≥5)","AC-2: dual-read fixture(15파일 `@MX:` 기준선) 변경 0","AC-3: check set exit"],
  ["C1"], "00:5x dual-read"),
 ("C3","EP-C","Go MX 자산 inventory·normalization·중복 진단(863 태그 기준선)","Core Brain L2852; `moai-adk-go@4fa2aa9 internal/navigator/sync/mx_bridge.go:54-63`",
  ["읽기 전용 inventory op: Go 기준선 태그 수·종류·파일 분포를 재현 명령과 함께 산출(863·28.8k·7 MB 수치 재측정)","중복·고아 태그 진단 보고(수정 0)"],
  ["REQ-C3-01: THE INVENTORY SHALL 재현 명령과 기준 commit을 출력에 포함한다."],
  ["AC-1: 수치 재측정 + 차이 설명","AC-2: Go 트리 무변경(해시)","AC-3: check set exit"],
  ["C2"], "Go 격리 유지"),
 ("C4","EP-C","온톨로지 노드/엣지 schema — basis·verification 분리, tombstone","설계 §8; Core Brain L741·L752-810·L2413-2491(세 벌 통합)",
  ["노드 8종·엣지 통합 목록(DR-07 부록 기준) JSON schema, `basis`(명시/추론)·`verification`(검증 상태) 분리 필드, tombstone 레코드","JSON 정본(TOON 없음)"],
  ["REQ-C4-01: THE SCHEMA SHALL 명시 link와 추론 link를 같은 필드에 섞지 않는다."],
  ["AC-1: schema 검증 시험(유효·무효 각 ≥5)","AC-2: tombstone 라운드트립","AC-3: check set exit"],
  [], "01:2x JSON 정본"),
 ("C5","EP-C","명시 SPEC link 추출 — REQ↔code↔test(`@MOAI:REQ`·테스트 이름·Issue 번호)","설계 §7·§8; 05-A spec.rs digest",
  ["추출기: 태그·테스트명·커밋 trailer에서 REQ↔파일↔테스트 삼각 link 생성(명시만, 추론 0)","출력은 C4 schema"],
  ["REQ-C5-01: THE EXTRACTOR SHALL 추론 link를 만들지 않는다."],
  ["AC-1: fixture 저장소에서 link precision 100%(명시만)","AC-2: 누락 사례 보고","AC-3: check set exit"],
  ["C2","C4"], ""),
 ("C6","EP-C","증분 색인 — generation 원자 교체, fresh/stale/partial/unavailable","설계 §10; `src/context/mod.rs` freshness",
  ["generation 디렉터리 temp+rename 교체, 질의 결과에 freshness 상태 4종, 변경 탐지(추가·삭제·새 파일·branch switch)","watcher 없음(polling/명령 기반)"],
  ["REQ-C6-01: WHEN 색인 중 실패하면 THE SYSTEM SHALL 이전 generation을 유지하고 partial을 표시한다."],
  ["AC-1: FRESHNESS 시험군 fixture(추가·삭제·새 파일·branch switch) 탐지","AC-2: 혼합 generation 0","AC-3: check set exit"],
  ["C4"], ""),
 ("C7","EP-C","비벡터 우선 검색 순서 + graph traversal","설계 §11 L3726 순서(exact ID→승인 계약→명시 관계→심볼→task 관련도→evidence→lexical)",
  ["순서 고정 구현, 각 단계 결과에 source citation","graph traversal 깊이 제한·budget 연동(C8)"],
  ["REQ-C7-01: THE SEARCH SHALL exact ID 일치를 다른 모든 단계보다 먼저 반환한다."],
  ["AC-1: RETRIEVAL fixture recall/precision 보고","AC-2: citation 정확도 시험","AC-3: check set exit"],
  ["C5","C6"], ""),
 ("C8","EP-C","retrieval budget — 기본 256 KiB, profile 재정의, host별 tokenizer 측정","설계 §11; decisions 01:1x budget; `src/context/mod.rs` bytes_div_4",
  ["기본값 상수 + `project.yaml` profile 키, 초과 시 omitted 기록(현행)","tokenizer 측정 카드: Claude/Codex 실제 토큰 수와 bytes/4 오차 표(실측은 운영자 환경 — not_run 허용, 표 틀만)"],
  ["REQ-C8-01: WHEN budget이 없으면 THE SYSTEM SHALL 256 KiB 기본을 쓰고 출처를 기록한다."],
  ["AC-1: 기본/재정의 시험","AC-2: 오차 표 틀 + 측정 절차","AC-3: check set exit"],
  [], "01:1x budget"),
 ("C9","EP-C","Code Map / Code Graph 조회 op","설계 §7; Core Brain L4302 context 선택",
  ["`atlas-map`·`atlas-graph` 읽기 op: 파일/모듈 지도, 심볼 관계 질의, 출력 JSON(C4 schema)","쓰기 0"],
  ["REQ-C9-01: THE OPS SHALL 읽기 전용이며 색인이 없으면 unavailable을 반환한다."],
  ["AC-1: fixture 질의 시험","AC-2: unavailable 경로","AC-3: check set exit"],
  ["C6","C7"], ""),
 ("C10","EP-C","회귀 시험군 ID GRAPH·FRESHNESS·RETRIEVAL fixture + eval corpus(30~50 task)","설계 §17; Core Brain L3703",
  ["eval corpus 정의(위치 `tests/fixtures/atlas/`, 구성 기준), 세 시험군 fixture와 통과 기준 수치(baseline 뒤 고정)"],
  ["REQ-C10-01: THE CORPUS SHALL 각 task의 ground truth와 출처를 기록한다."],
  ["AC-1: corpus ≥30 task","AC-2: 세 시험군 실행 결과 표","AC-3: check set exit"],
  ["C7"], ""),
 # ---------------- EP-D ----------------
 ("D1","EP-D","archive 적격 판정 — terminal + 비활성 90일, pending evidence·hold 제외, dry-run 보고","설계 §12; decisions 00:5x",
  ["`archive-plan` op(dry-run): 적격 목록·사유·제외 사유 JSON","수치는 profile 재정의 가능"],
  ["REQ-D1-01: THE PLAN SHALL pending evidence 또는 hold가 있는 항목을 제외한다."],
  ["AC-1: fixture 적격/제외 각 ≥3","AC-2: 보고 형식","AC-3: check set exit"],
  [], "00:5x archive"),
 ("D2","EP-D","archive 자동 이동(hash/readback 필수) + restore + 검색","설계 §12; decisions 00:5x 자동 이동 허용",
  ["`archive-apply`: 이동 전 hash → 이동 → readback 일치 확인 → 인덱스 갱신; 실패 시 원위치 유지","restore op, 아카이브 검색(비벡터)"],
  ["REQ-D2-01: WHEN readback이 불일치하면 THE SYSTEM SHALL 이동을 취소하고 원본을 유지한다."],
  ["AC-1: 성공/실패 fixture","AC-2: restore 라운드트립","AC-3: check set exit"],
  ["D1"], "00:5x"),
 ("D3","EP-D","ARCHIVE·PRIVACY 시험군 fixture","설계 §17",
  ["ARCHIVE: 활성 작업·pending·hold 이동 0, hash/readback, restore, 검색 보존","PRIVACY: 다른 project/ACL 누출 0, secret 전송 0, malicious document 지시 실행 0"],
  ["REQ-D3-01: THE FIXTURES SHALL 각 통과 기준을 자동 판정한다."],
  ["AC-1: 두 시험군 각 ≥4 시험","AC-2: 결과 표","AC-3: check set exit"],
  ["D2"], ""),
 ("D4","EP-D","benchmark 묶음 — p50/p95·RSS·index bytes/LOC baseline","설계 §16; P11",
  ["측정 harness(동일 fixture·조건), 결과 표 형식, 비교는 baseline 뒤에만(절감률 발표 금지)"],
  ["REQ-D4-01: THE REPORT SHALL 측정 환경·명령·반복 횟수를 포함한다."],
  ["AC-1: baseline 1회 산출물","AC-2: 환경 기록","AC-3: check set exit"],
  [], ""),
 ("D5","EP-D","비용 분해 보고(토큰·호출·시간)","설계 §16",
  ["run/verify receipt에서 비용 항목 집계 보고 op, 외부 backend 호출 0 유지"],
  ["REQ-D5-01: THE REPORT SHALL 추정값과 측정값을 구분 표기한다."],
  ["AC-1: fixture 집계","AC-2: 표기 구분","AC-3: check set exit"],
  ["D4"], ""),
 ("D6","EP-D","문서/소스/template/live 의미 parity 검사","설계 §18 6단계 '의미 parity'; 00-A baseline 분류",
  ["parity op: 템플릿↔생성물↔문서의 키·값 차이 보고(00-A 분류 재사용)"],
  ["REQ-D6-01: THE CHECK SHALL 차이를 종류별(누락/추가/값 상이)로 분리한다."],
  ["AC-1: fixture 차이 탐지","AC-2: 보고 형식","AC-3: check set exit"],
  [], ""),
 # ---------------- EP-E ----------------
 ("E1","EP-E","cargo-dist 도입 — sha256·SBOM 산출(결정 A, 미구현)","decisions 7; docs/release-checklist.md:29",
  ["`dist-workspace.toml` 설정, 로컬 `cargo dist build` 산출물 검증, SBOM 포함 설정; 실제 release 게시는 카드 범위 밖"],
  ["REQ-E1-01: THE BUILD SHALL 바이너리마다 sha256과 SBOM을 산출한다."],
  ["AC-1: 로컬 dist build 산출물 목록","AC-2: 체크리스트 갱신","AC-3: check set exit"],
  [], "01:1x SBOM=cargo-dist"),
 ("E2","EP-E","Linux x86_64/arm64 CI matrix + 지원표","decisions 01:1x OS matrix; `.github/workflows/rc-ci.yml`",
  ["rc-ci에 ubuntu x86_64·arm64 job 추가, macOS 전용 cfg 시험 분리, 지원표 행"],
  ["REQ-E2-01: THE CI SHALL 네 target에서 check set을 실행한다."],
  ["AC-1: CI 4 job success(run id)","AC-2: 지원표","AC-3: check set exit"],
  ["E1"], ""),
 ("E3","EP-E","Windows 빌드 + self-replace crate 구현(지원표 not_run)","decisions 12 self-replace 1.5.1; Cargo.toml 의존성 없음",
  ["self-replace 의존성 추가(버전 고정), Windows 실행 중 교체 경로 구현, CI windows build job(테스트는 not_run 표기 가능)"],
  ["REQ-E3-01: WHEN Windows에서 UPDATE가 실행되면 THE SYSTEM SHALL self-replace로 교체하고 실패 시 known-good을 유지한다."],
  ["AC-1: 컴파일·단위 시험(모의)","AC-2: 지원표 not_run 표기","AC-3: check set exit"],
  ["E1"], ""),
 ("E4","EP-E","attestation 생성·검증 실측 + quarantine 안내 문서","decisions 8 attestation 단독; release-checklist.md:34",
  ["GitHub attestation workflow(게시 없는 dry-run) + `gh attestation verify` 절차, macOS quarantine 해제 안내(notarization 없음)"],
  ["REQ-E4-01: THE DOCS SHALL checksum이 서명을 대체하지 않음을 명시한다."],
  ["AC-1: dry-run 산출물 또는 not_run 사유","AC-2: 안내 문서","AC-3: check set exit"],
  ["E1"], "01:1x notarization 없음"),
 ("E5","EP-E","tag ruleset `moai-cli-v*` 생성(사용자 승인 1회, keep-set)","decisions 9; release-checklist.md:62",
  ["ruleset JSON 초안 + 적용 절차, 적용은 운영자 승인 뒤 리더"],
  ["REQ-E5-01: THE RULESET SHALL `moai-cli-v*` 외 tag 생성을 막지 않는다(Go tag 격리)."],
  ["AC-1: 초안 JSON","AC-2: 적용 후 readback(승인 뒤)","AC-3: check set exit"],
  [], "keep-set"),
 ("E6","EP-E","RELEASE 시험군 — Go/RC feed 격리·중단 복구 fixture","설계 §17 RELEASE; tests/install_cli.rs 특성화",
  ["feed 오염 표 확장, 중단 복구(known-good) fixture, 지원표 연결"],
  ["REQ-E6-01: THE FIXTURES SHALL Go RC 채널 필터 위험(평문 v* tag)을 재현한다."],
  ["AC-1: 시험 ≥4","AC-2: 결과 표","AC-3: check set exit"],
  ["E4"], ""),
 # ---------------- EP-F ----------------
 ("F1","EP-F","Pi v1.5.0 pin spike — conformance·보안·취소·인증 4항 측정","설계 §14; decisions 00:5x",
  ["고정 버전 설치(격리 환경), 4항 측정 절차·결과 표, 채택/비채택 권고(결정은 사용자)"],
  ["REQ-F1-01: THE SPIKE SHALL 측정 환경·버전·명령을 기록하고 미측정 항목은 not_run으로 둔다."],
  ["AC-1: 4항 표","AC-2: 권고 1절","AC-3: check set exit"],
  [], "00:5x Pi spike"),
 ("F2","EP-F","Model Adapter conformance(MODEL 시험군)","설계 §14·§17",
  ["tool/schema/image/usage/overflow/stream error/abort conformance fixture 정의"],
  ["REQ-F2-01: THE SUITE SHALL 각 항목을 독립 판정한다."],
  ["AC-1: fixture 정의","AC-2: not_run 표기","AC-3: check set exit"],
  ["F1"], ""),
 ("F3","EP-F","Electron Desktop 보안 경계 설계 — IPC sender/schema, renderer credential 차단","설계 §14 L658",
  ["설계 문서만(코드 0)"],
  ["REQ-F3-01: THE DESIGN SHALL renderer가 credential에 접근하지 못하는 경계를 명시한다."],
  ["AC-1: 문서","AC-2: 시험군 DESKTOP 연결","AC-3: check set exit"],
  ["F1"], ""),
 ("F4","EP-F","DESKTOP·PI·BROWSER 시험군 fixture 정의","설계 §17",
  ["세 시험군 통과 기준을 fixture 틀로(실행은 not_run)"],
  ["REQ-F4-01: THE FIXTURES SHALL 실행 없이도 기준을 기계 판정 가능하게 적는다."],
  ["AC-1: 세 틀","AC-2: not_run","AC-3: check set exit"],
  ["F1","F3"], ""),
 ("F5","EP-F","이미지 생성 asset 계약 조사(F21) — 권리·비용·visual QA","설계 §20 F21; [e361] L157",
  ["조사 문서: 권리 확인 방법, 비용 경계, visual QA 분리(DOM/AX vs visual); 구현 blocked"],
  ["REQ-F5-01: THE REPORT SHALL 비용이 드는 호출을 하지 않는다."],
  ["AC-1: 문서","AC-2: 외부 사실 URL+확인일","AC-3: check set exit"],
  [], "권리·비용 HOLD"),
 ("F6","EP-F","판단 backend 조사 — Jev·OpenAI Decisions·Claude/Codex 구독 credit(URL+확인일, 비용 호출 0)","설계 §18 '남은 결정' 8항; [e361] L147-149; decisions 01:1x",
  ["표: backend별 현행 API·요금·구독 credit 자격·확인일·URL; 미확인은 미확인","Jev는 Go `internal/jev` 기준선 사실만"],
  ["REQ-F6-01: THE REPORT SHALL 과거 문구만으로 혜택을 확정하지 않는다."],
  ["AC-1: 표 전수 URL+날짜 또는 미확인","AC-2: 비용 호출 0 선언","AC-3: check set exit"],
  [], "01:1x 조사 카드"),
 ("F7","EP-F","typed 판단 backend 연동(native 기본 유지, 선택)","설계 §20 F20; `src/decision.rs` P10",
  ["F6 결론 뒤 연동 설계·구현; authority none 유지"],
  ["REQ-F7-01: THE SYSTEM SHALL native 판단을 기본으로 하고 backend 결과에 authority를 부여하지 않는다."],
  ["AC-1: 설계 승인","AC-2: fixture","AC-3: check set exit"],
  ["F6"], ""),
 # ---------------- EP-G ----------------
 ("G1","EP-G","Space 상위 페이지 적용 — fixes/top-5490c4ee.md 19항목 + 적용 후 사본 sha 확인","#2140 산출물; 결정 01:2x 리더 적용",
  ["리더: 항목별 요약 보고 → 사용자 확인 → ego-browser 적용 → 재수집 사본 sha256 기록(docs/design-review/applied.md)"],
  ["REQ-G1-01: THE LEADER SHALL 적용 전 사용자 확인을 받고 적용 후 사본으로 검증한다."],
  ["AC-1: 19항목 적용 체크표","AC-2: 사본 sha 기록","AC-3: 레인 작업 없음(리더)"],
  ["#2140"], "01:2x"),
 ("G2","EP-G","Space 하위 13편 적용 — group2·group3·plan 수정안","#2141·#2142 산출물",
  ["G1과 같은 절차, 페이지당 확인"],
  ["REQ-G2-01: 동일"],
  ["AC-1: 13편 체크표","AC-2: sha 기록","AC-3: 리더"],
  ["#2141","#2142"], "01:2x"),
 ("G3","EP-G","Space §20 표·부록 4종 적용","#2143 산출물",
  ["§20 6행 교체 + 부록 4종을 상위 페이지 부록으로"],
  ["REQ-G3-01: 동일"],
  ["AC-1: 체크표","AC-2: sha 기록","AC-3: 리더"],
  ["#2143"], "01:2x"),
 ("G4","EP-G","TOON 제외안 적용 — Core Brain L4436-4450·상위 §13 L1395","결정 01:2x 도입하지 않음, JSON 정본; #2142 136c 제외안",
  ["제외안 교체문 적용, FORMAT-* 시험 삭제, 상위 L1395 수정"],
  ["REQ-G4-01: 동일"],
  ["AC-1: 적용","AC-2: sha 기록","AC-3: 리더"],
  ["#2142"], "01:2x TOON"),
]

def slug(s):
    s = re.sub(r"[`/|\\\\]", " ", s)
    s = re.sub(r"[^0-9A-Za-z가-힣\s-]", "", s).strip().lower()
    return re.sub(r"\s+", "-", s)[:40]

def run(args):
    if DRY:
        print("DRY:", " ".join(args)[:160])
        return ""
    return subprocess.check_output(args, text=True).strip()

def create(title, body):
    url = run(["gh","issue","create","-R",REPO,"--milestone",MILESTONE,"--title",title,"--body",body])
    return int(url.rsplit("/",1)[-1]) if url else 0

nums = {}
OUT.mkdir(parents=True, exist_ok=True)
# EPs
for key,(name,ref,summary) in EPS.items():
    children = [c for c in CARDS if c[1]==key]
    title = f"[rc.2][{key}] {name}"
    body = f"# {title}\n\n## 범위\n{summary}\n\n## 설계 근거\n{ref}\n\n## 실행 카드 ({len(children)})\n" + "\n".join(f"- {c[0]} {c[2]}" for c in children) + f"\n\n## 완료 조건\n모든 실행 카드가 PASS·병합·종료. 카드 Issue 번호는 발행 뒤 이 본문에 리더가 채운다.\n\n기준 tip: {TIP}. 사용자 결정: docs/decisions-2026-11-08.md 00:4x~01:2x(11-09).\n{COMMON}"
    n = create(title, body)
    nums[key] = n
    (OUT/f"{n or 'DRY'}-rc2-{key}.md").write_text(body)
# cards
for key,ep,title_,ref,scope,reqs,acs,blocked,dec in CARDS:
    title = f"[rc.2][{key}] {title_}"
    bl = ", ".join((f"#{nums[b]}" if b in nums else b) for b in blocked)
    body = f"# {title}\n\n상위: #{nums[ep]} ({ep}). 설계 근거: {ref}.\n\n## 범위\n" + "\n".join(f"- {s}" for s in scope) + "\n\n## 요구(GEARS)\n" + "\n".join(f"- {r}" for r in reqs) + "\n\n## 수용 기준\n" + "\n".join(f"- {a}" for a in acs) + f"\n\n## 선행·차단\nBlocked-By: {bl or '없음'}\n\n## 사용자 결정\n{dec or 'RC1 결정 우선(00:4x)'} — docs/decisions-2026-11-08.md\n\n## 증거\nPR 본문에 check set 명령·exit, 변경 파일, not_run 항목, claim 토큰. 기준 tip {TIP}.\n{COMMON}"
    n = create(title, body)
    nums[key] = n
    (OUT/f"{n or 'DRY'}-rc2-{key}-{slug(title_)}.md").write_text(body)
# fill EP child numbers
if not DRY:
    for key in EPS:
        children = [c for c in CARDS if c[1]==key]
        p = next(OUT.glob(f"{nums[key]}-rc2-{key}.md"))
        body = p.read_text()
        body = re.sub(r"## 실행 카드 \(\d+\)\n(?:- .*\n)+", "## 실행 카드 (%d)\n" % len(children) + "".join(f"- #{nums[c[0]]} {c[0]} {c[2]}\n" for c in children), body)
        p.write_text(body)
        run(["gh","issue","edit","-R",REPO,str(nums[key]),"--body-file",str(p)])
print(json.dumps(nums, ensure_ascii=False))
