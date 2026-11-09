# group3-66acadc2 수정안 — MoAI RC1 GitHub 작업 구조와 실행 계획

- 대상: `references/chatgpt-space/page_66acadc28bdc8191875556986da72e6b.md` (읽기 전용 입력)
- 검토 근거: `review-group3-install-host-plan-run-release.md` §7·§8(X7)·§9.1
- 적용 tip: **604792666** (2026-11-09 재확인 — 검토 기준 `9bfd26d3b` 이후 `src/claim.rs`가 +45 변경되어 아래 근거는 전부 현재 tip에서 재인용)
- 우선순위: S1(§9.1 — 이 문서의 권위 방향 미갱신이 종합 최위험 항목)

## 수정 항목

| # | 위치 | 현재 | 교체 | 근거 (tip 604792666 재확인) |
| --- | --- | --- | --- | --- |
| 1 (S1, X7) | §8 "실행 소유권" 단락 — "초기 구현안은 한 조정 owner가 쓰는 작은 durable journal이다. 이미 있는 native/provider 조정 기능이 필요한 원자성을 실제 보장하면 이를 adapter로 사용할 수 있다." 및 직후 "GitHub assignee·댓글·status label은 이 결과를 보여 주는 projection이다." · "로컬 파일의 rename·lock은 여러 컴퓨터 전체의 CAS가 아니다. 여러 리더가 같은 Issue를 배차하려면 검증된 기존 provider의 조정 primitive가 필요하며 없으면 단일 리더를 유지하거나 배차를 차단한다." | journal이 primary이고 Issue 댓글·provider primitive는 projection/대체 경로로 읽힘 — RC1 실물과 권위 방향이 반대 | "RC1(2026-11-08 P08 결정): ⑤ Issue 댓글 claim(`moai-claim: session=<id> epoch=<n> lease_until=<utc>` 1개, 최소 live 댓글 id가 승자)이 **cross-host primary**이고, ① 같은 host의 저널(`<common dir>/.moai-cli/claims/` — append-only `journal.log` + `File::try_lock`)은 **같은 host 클레이먼트의 직렬화기**다. 위 단락의 journal-primary·provider-adapter 문장은 RC1 이전의 초기 구현안 역사로 표시하고, 다음 구현자가 저널을 primary로 읽지 않게 한다. 댓글 projection 표현은 'GitHub assignee·댓글·status label은 상태의 projection'으로 유지하되 claim의 권위는 댓글 id에 있음을 앞 문장에서 명시한다. 여러 컴퓨터 전체의 CAS는 제공하지 않는다는 경고는 유지(⑤도 최소 live 댓글 id 비교가 전부다)." | `src/claim.rs:1-5`(⑤ primary·① serializer 명시), `src/claim.rs:13-19`(댓글 1개 게시·최소 id 승·`Moai-Claim` 토큰 분기/트레일러/PR 본문 운반), `docs/decisions-2026-11-08.md:21`(P08), 검토 §8 X7 행 |
| 2 | §9 상태 모델(L1256) — "상태 모델은 unclaimed → assigned → running → verifying → pr_ready → in_review → integrated를 기본 설명으로 …" | `integrated`가 기본 흐름의 종착 상태처럼 읽힘 | 상태 모델 뒤에 한 문장 추가: "`integrated`는 RC1 claim 상태가 아니다 — 통합은 리더 게이트(독립 AUDIT verdict, PR head==audited SHA, `moai/audit` status)가 판정하고 claim record는 `pr_ready`/`in_review`에서 리더 인계로 끝난다. RC1 ACTIVE_STATES는 assigned·running·verifying·pr_ready·in_review·blocked·handoff_pending·uncertain의 8개다." | `src/claim.rs:48-58`(ACTIVE_STATES 8개 + 주석 "unclaimed, integrated, cancelled does not block a new acquire"), `docs/support-matrix.md`(리더 게이트 절차는 AGENTS.local.md § Lane loop integration gate와 동일) |
| 3 | §9 파일·workflow 실물명(L1310 전후) — fixture 디렉터리와 가칭으로 된 workflow 이름 | 가칭·fixture 네 디렉터리 표기 | 실물 경로로 교체: Issue forms `.github/ISSUE_TEMPLATE/moai-{feature,bugfix,investigation}.yml`, 라벨 원장 `.github/moai/labels.yml`(부속 `README.md`·`removed-go-generation-paths.txt`), workflow `.github/workflows/{rc-ci.yml,rc-label-sync.yml}`. fixture 전용 디렉터리는 삭제 | 현재 tip에서 `ls .github/ISSUE_TEMPLATE/`·`.github/moai/`·`.github/workflows/` 재확인(2026-11-09, 본 수정안 작성 시 실측), 검토 §7.6.3 |
| 4 | §9 통합 기록(구 L1256 표) — `integrated` 상태와 ruleset·`moai/audit` 미기재 | 통합 절차가 상태 표 밖에 흩어짐 | 통합 행을 상태 표에 추가: "integrated — RC1 claim 상태 아님(리더 게이트): ruleset `RC1 integration gate (v2026.11.0)`(PR 필수, `moai/audit` status, up-to-date strict, force-push·삭제 금지) + 독립 AUDIT verdict 첫 줄 대조" | AGENTS.local.md § Lane loop integration gate 절(현재 tip), 검토 §7.6.4 |
| 5 | "기존 자산 원본·190개 Claude 비교" 표(§3 전후) — "190개" 집계 기준 부재 | 수치의 집계 기준 미기재 | "190개"에 기준을 붙이거나 삭제: "190개는 `docs/github-inventory-rc1.md`(2026-11-08T07:56Z 기준)의 비교 행 수 — 이후 원격 변동은 재조회 전까지 미반영" (기준을 확인할 수 없으면 수치 삭제) | `docs/github-inventory-rc1.md`(2026-11-08T07:56Z 스냅샷 명시), 검토 §7.6.5·§7.7 |
| 6 | §11 "남은 선택은 정확한 대상 repo·… signing·release 운영이다"(L1663) | 옛 미결 목록 그대로 — decisions의 HOLD 목록과 불일치 | "미결(HOLD, docs/decisions-2026-11-08.md:94-96): P08은 claim 방식(⑤+①)만 결정·lease·heartbeat 수치 HOLD, P15는 ZCode 헤드리스만 결정·나머지 host 지원표 HOLD, P12의 OS/arch matrix·locale·release matrix, 그리고 P04·P05 나머지·P06·P08~P11·P14·P15는 HOLD 유지" | `docs/decisions-2026-11-08.md:94-96`(현재 tip에서 재인용), 검토 §7.6.6 |
| 7 | 시험 ID 행(§10 수용 기준) — GH-* ID들이 RC1 receipt와 대응 없음 | GH-* 시험군의 RC1 상태 미표기 | GH-* 행마다 RC1 상태 추가: "RC1 receipt 대응 없음(not_applicable — RC1의 E2E receipt는 AC-08C-1·2·5·7·8·HA-02만 존재, `docs/e2e-receipts/receipts.jsonl` 6 passed + 3 not_run; `docs/support-matrix.md` 지원 6행). AC-08C-3·4·6은 not_run(운영자 실측 대기)" | `docs/support-matrix.md`(지원 6행·not_run 표), `docs/e2e-receipts/receipts.jsonl`(9행 실측, 2026-11-09 읽음) |

## X11 공통 문단 (한 교체문 · 세 위치 적용)

X11의 66a 소속 항목은 없다(세 위치는 [83f]·[25f]·[487]·[dff] 소속) — 해당 문서들의 group3-*.md에 "한 교체문 · 세 위치" 표시로 반영한다.

## 비고

- 가설 단독 근거 0(AC-3): 위 근거는 전부 현재 tip 파일 관측값이다.
- 개인 절대 경로·비밀 0건(AC-3).
