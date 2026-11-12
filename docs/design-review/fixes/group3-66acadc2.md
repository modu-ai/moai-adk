# group3-66acadc2 수정안 — MoAI RC1 GitHub 작업 구조와 실행 계획

- 대상: `references/chatgpt-space/page_66acadc28bdc8191875556986da72e6b.md` (읽기 전용 입력)
- 검토 근거: `review-group3-install-host-plan-run-release.md` §7·§8(X7)·§9.1
- 적용 tip: **86762b8e1** (2026-11-10 재확인, #2207 — 처음 작성 기준은 604792666, 그때 검토 기준 `9bfd26d3b` 이후 `src/claim.rs` +45 변경 반영)
- 우선순위: S1(§9.1 — 이 문서의 권위 방향 미갱신이 종합 최위험 항목)

## 수정 항목

| # | 위치 | 현재 | 교체 | 근거 (tip 86762b8e1 재확인) |
| --- | --- | --- | --- | --- |
| 1 (S1, X7) | §8 "실행 소유권" 단락 — "초기 구현안은 한 조정 owner가 쓰는 작은 durable journal이다. 이미 있는 native/provider 조정 기능이 필요한 원자성을 실제 보장하면 이를 adapter로 사용할 수 있다." 및 직후 "GitHub assignee·댓글·status label은 이 결과를 보여 주는 projection이다." · "로컬 파일의 rename·lock은 여러 컴퓨터 전체의 CAS가 아니다. 여러 리더가 같은 Issue를 배차하려면 검증된 기존 provider의 조정 primitive가 필요하며 없으면 단일 리더를 유지하거나 배차를 차단한다." | journal이 primary이고 Issue 댓글·provider primitive는 projection/대체 경로로 읽힘 — RC1 실물과 권위 방향이 반대 | "RC1(2026-11-08 P08 결정, 2026-11-09 잔여 결정): ⑤ Issue 댓글 claim(`moai-claim: session=<id> epoch=<n> lease_until=<utc>` 1개, 최소 live 댓글 id가 승자)이 **cross-host primary**이고, ① 같은 host의 저널(`<common dir>/.moai-cli/claims/` — append-only `journal.log` + `File::try_lock`)은 **같은 host 클레이먼트의 직렬화기**다. lease는 생략 시 4시간, heartbeat는 없고, 만료돼도 자동 takeover는 없다(리더의 epoch+1 수동 인계만). 위 단락의 journal-primary·provider-adapter 문장은 RC1 이전의 초기 구현안 역사로 표시하고, 다음 구현자가 저널을 primary로 읽지 않게 한다. 댓글 projection 표현은 'GitHub assignee·댓글·status label은 상태의 projection'으로 유지하되 claim의 권위는 댓글 id에 있음을 앞 문장에서 명시한다. 여러 컴퓨터 전체의 CAS는 제공하지 않는다는 경고는 유지(⑤도 최소 live 댓글 id 비교가 전부다 — 'GitHub에 검증된 조정 primitive 없음' 전제)." | `src/claim.rs:1-3`(⑤ primary·① serializer 명시), `src/claim.rs:15-24`(댓글 1개 게시·최소 id 승·`Moai-Claim` 토큰 분기/트레일러/PR 본문 운반), `src/claim.rs:25-30`(자동 takeover 금지·lease 기본 4h), `docs/decisions-2026-11-08.md` L21(P08)·L102(00:5x 잔여), 검토 §8 X7 행 |
| 2 | §9 상태 모델(L1256) — "상태 모델은 unclaimed → assigned → running → verifying → pr_ready → in_review → integrated를 기본 설명으로 …" | `integrated`가 기본 흐름의 종착 상태처럼 읽힘 | 상태 모델 뒤에 한 문장 추가: "`integrated`는 RC1 claim 상태가 아니다 — 통합은 리더 게이트(독립 AUDIT verdict, PR head==audited SHA, `moai/audit` status)가 판정하고 claim record는 `pr_ready`/`in_review`에서 리더 인계로 끝난다. RC1 ACTIVE_STATES는 assigned·running·verifying·pr_ready·in_review·blocked·handoff_pending·uncertain의 8개다." | `src/claim.rs:56-66`(ACTIVE_STATES 8개 + 주석 "unclaimed, integrated, cancelled does not block a new acquire"), `AGENTS.local.md:73`(integration gate) |
| 3 | §9 파일·workflow 실물명(L1310 전후) — fixture 디렉터리와 가칭으로 된 workflow 이름 | 가칭·fixture 네 디렉터리 표기 | 실물 경로로 교체: Issue forms `.github/ISSUE_TEMPLATE/moai-{feature,bugfix,investigation}.yml`, 라벨 원장 `.github/moai/labels.yml`(부속 `README.md`·`removed-go-generation-paths.txt`), workflow `.github/workflows/{rc-ci.yml,rc-label-sync.yml}`. `rc-ci`는 Linux x86_64·arm64, macOS arm64·x86_64 네 target에서 check set을 돌리고 집계 job 하나가 `rc-ci` 결과를 낸다(필수 check는 아님). fixture 전용 디렉터리는 삭제 | 현재 tip에서 `ls .github/ISSUE_TEMPLATE/`·`.github/moai/`·`.github/workflows/` 재확인(2026-11-10), `.github/workflows/rc-ci.yml`(E2 #2194 matrix, #2312 집계 조건), 검토 §7.6.3 |
| 4 | §9 통합 기록(구 L1256 표) — `integrated` 상태와 ruleset·`moai/audit` 미기재 | 통합 절차가 상태 표 밖에 흩어짐 | 통합 행을 상태 표에 추가: "integrated — RC1 claim 상태 아님(리더 게이트): ruleset `RC1 integration gate (v2026.11.0)`(PR 필수, `moai/audit` status, up-to-date strict, force-push·삭제 금지) + 독립 AUDIT verdict 첫 줄 대조. 게이트가 브랜치를 갱신(update-branch)했다면 감사 head 대비 변경 줄이 같고, tip을 포함하고, rc-ci가 success이고, head를 다시 확인한 뒤에만 status·병합한다. `Moai-Claim` 줄은 레인이 쓴 커밋(merge 제외)과 PR 본문에 필수다. 감사 전 열린 PR은 레인당 1개, 전체 8개가 상한이다" | `AGENTS.local.md:73`(integration gate), `docs/decisions-2026-11-08.md` L39·L43(통합 정책·게이트 보강)·L123(06:02 claim 줄 범위·통합 동기화)·L125(06:2x WIP 상한·감사자 4명), 검토 §7.6.4 |
| 5 | "기존 자산 원본·190개 Claude 비교" 표(§3 전후) — "190개" 집계 기준 부재 | 수치의 집계 기준 미기재 | "190개"에 기준을 붙이거나 삭제: "190개는 `docs/github-inventory-rc1.md`(2026-11-08T07:56Z 기준)의 비교 행 수 — 이후 원격 변동은 재조회 전까지 미반영" (기준을 확인할 수 없으면 수치 삭제) | `docs/github-inventory-rc1.md`(2026-11-08T07:56Z 스냅샷 명시), 검토 §7.6.5·§7.7 |
| 6 | §11 "남은 선택은 정확한 대상 repo·… signing·release 운영이다"(L1663) | 옛 미결 목록 그대로 — decisions의 결정·HOLD 목록과 불일치 | "**결정됨**(2026-11-09): P08 잔여 — 수동 인계·lease 4h·heartbeat 없음(L102). P04 — 첫 RC 인증 조합은 macOS arm64 × Claude Code·Codex(L120). P12·서명 — OS matrix macOS(arm64·x86_64)+Linux(x86_64·arm64), Windows는 빌드+self-replace만, attestation+sha256(notarization 없음), SBOM은 cargo-dist 산출물, tag ruleset `moai-cli-v*` 카드(L107). P11 — retry max_attempts 3·context 256 KiB·retention 90일·비용 상한 미설정(L108). Forge 권한·fallback(L109). P10 — 조사 카드 뒤 구현, native 판단 유지(L110). TOON — 도입하지 않음, JSON 정본(L111). **HOLD 유지**(L130·L132): P15 나머지 host 지원표, P12 locale·release matrix, P04 프로젝트 단위 skill 복사 범위, P05 나머지, P06, P09, P14" | `docs/decisions-2026-11-08.md` L102·L107·L108·L109·L110·L111·L120(결정 행)·L130·L132(HOLD 목록), 검토 §7.6.6 |
| 7 | 시험 ID 행(§10 수용 기준) — GH-* ID들이 RC1 receipt와 대응 없음 | GH-* 시험군의 RC1 상태 미표기 | GH-* 행마다 RC1 상태 추가: "RC1 receipt 대응 없음(not_applicable — RC1의 E2E receipt는 AC-08C-1·2·5·7·8·HA-02만 존재, `docs/e2e-receipts/receipts.jsonl` 6 passed + 3 not_run; `docs/support-matrix.md` 지원 6행). AC-08C-3·4·6은 not_run(운영자 실측 대기)" | `docs/support-matrix.md`(지원 6행·not_run 표), `docs/e2e-receipts/receipts.jsonl`(9행, 2026-11-10 재확인 — 변화 없음) |

## X11 공통 문단 (한 교체문 · 세 위치 적용)

X11의 66a 소속 항목은 없다(세 위치는 [83f]·[25f]·[487]·[dff] 소속) — 해당 문서들의 group3-*.md에 "한 교체문 · 세 위치" 표시로 반영한다.

## 비고

- 가설 단독 근거 0(AC-3): 위 근거는 전부 tip 86762b8e1 파일 관측값이다.
- 개인 절대 경로·비밀 0건(AC-3).
- 기록 정리 대상(사용자 선택 아님): L132 문구는 "P08~P11 … HOLD 유지"라고 적지만 L102(P08)·L108(P11)·L110(P10)이 그 일부를 뒤에 결정했고, "P12에서 아직 HOLD인 것"에도 L107이 정한 OS/arch matrix가 남아 있다. 6번 교체문은 개별 결정 행을 따른다.

## 갱신 기록 (#2207, tip 86762b8e1)

- 6번: 인용한 "decisions:94-96"의 HOLD 목록은 L130·L132로 옮겨 갔고, 그 사이 P08 잔여·P04 첫 조합·P12·P11·P10·Forge·TOON이 결정됐다. 결정됨/HOLD 유지 두 묶음으로 다시 썼다.
- 4번: 06:02(L123)의 claim 줄 범위·update-branch 뒤 재확인 게이트와 06:2x(L125)의 WIP 상한을 더했다. 근거 줄을 `AGENTS.local.md:73`으로 명시했다.
- 1번: lease 4h·heartbeat 없음·수동 인계(L102)를 더했다. 줄 번호 이동: `src/claim.rs:13-19`→`15-24`.
- 2번: `src/claim.rs:48-58`→`56-66`.
- 3번: rc-ci의 4 target matrix와 집계 job(E2 #2194, #2312)을 적었다.

## 현재 tip 재확인 (#2207, 2026-11-12, tip `325cc2bf03`)

- 인용 대조(`86762b8e1` → `325cc2bf03`, 같은 줄 범위의 글자 비교): 6건 중 그대로 4, 줄만 이동 2, 내용 바뀜 0, 모호 0, 고정 tip에 없던 파일 0(뒤 base `8c5732615` 근거). 인용은 각자 적힌 고정 tip 기준이라 그대로 두고, 교체문에 옮겨지는 사실만 다시 확인했다.
- 교체문에서 바뀐 사실 없음.
- 교체문에 들어간 다음 사실은 tip에서 다시 확인했고 그대로다: claim lease 기본 4시간·heartbeat 없음 (`src/claim.rs` `DEFAULT_LEASE`), `Budget`은 `max_attempts` 하나·기본값 없음(`src/run.rs`), E2E receipt 9건(6 passed + 3 not_run, `docs/e2e-receipts/receipts.jsonl`), GitLab은 capability 선언만 (`src/forge/gitlab.rs`, 호출 경로 없음), hook 생성의 host 로딩 `not_run`(`docs/rust-product-path.md` `hooks.apply` 행), `capabilities` 명령 없음, Codex config CLI 없음(#2237 PR #2534 미병합), #2197·#2014 OPEN. OPERATIONS 전수는 지금 99개다(고정 근거 `86762b8e1`에서는 70개). 70을 적은 근거 열은 그 고정 tip 기준이며, 교체문에는 개수가 없다.
