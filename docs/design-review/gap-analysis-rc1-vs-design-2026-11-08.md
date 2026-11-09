# 설계 문서(환골탈태 통합 최종 설계) 대 RC1 `moai-cli` 현황 대조 보고

- 작성: 리더 세션이 위임한 읽기 전용 분석 서브에이전트, 2026-11-08 23:1xZ 완료(소요 약 7분, 도구 호출 37회). 리더가 내용을 편집하지 않고 보존함.
- 기준: `references/chatgpt-space/page_5490c4ee8c8481918b3fc3217401a3f4.md`(2,494행, 22절) ↔ `origin/v2026.11.0` @ `8ac440587` ↔ RC1 milestone Issue 목록 ↔ `docs/decisions-2026-11-08.md`.

## 1. 한 줄 요약

설계 문서의 1~6절·13절·15절(제품 경계·호스트 경계·PLAN/RUN/VERIFY/AUDIT/SYNC·Issues 협업·파일 구조·오류 복구)은 RC1 Issue 32장과 `src/` 모듈로 대부분 구현되거나 부분 구현되었고, 7~12절·14절·16절(Core Brain·Code Map/Graph·온톨로지·증분 색인·검색·보존/아카이브·Pi/Desktop·성능 측정)은 RC1 Issue가 하나도 없는 범위 밖 영역이며, "첫 RC gate"의 양 host 실제 로딩 절반(08-C T2)은 운영자 실측 대기로 명시적 `not_run` 상태다.

## 2. 절별 대조표

상태 어휘: 구현 / 부분 / 범위 밖 / 충돌 / 해당 없음. "근거 없음"은 확인하지 못한 항목. 텍스트 검색 0건은 가설이지 미구현 판정이 아니다(6절 참조).

| 절 | 제목 | 설계가 요구하는 핵심 | RC1 상태 | 근거 |
|---|---|---|---|---|
| 1 | 통합 결정과 제품 경계 (L11) | 공통 skill 1개 진입, Rust moai-cli·Go 격리, 새 서버·계정 없음, Issues 정본, 승인 분리, Pi는 RC1 선행조건 아님, 첫 RC gate | 부분 — 경계 결정은 모두 일치; 승인 분리는 merge 자율화와 긴장(3절), 첫 RC gate 절반 not_run(5절) | `skills/moai/SKILL.md`; `docs/rc-isolation-decision.md`; decisions:24·39·89·91; #1975 본문 |
| 2 | 전체 구조와 책임 (L122) | 계층별 소유/비소유 책임, capability unknown→documented→probed→certified | 부분 — capability 상태 모델은 지원표에 채택; Core Brain·Model Adapter·Storage Adapter 계층은 미존재 | `docs/host-support-matrix.md` 상태 어휘 표; `docs/support-matrix.md` 머리 문단; #1993 #1994 |
| 3 | 개인 맞춤 프로젝트 시작 (L279) | 딥 인터뷰→profile→승인 범위 적용→허용된 초기 commit→검토된 Issues 발행→맞춤 agent/skill 생성, 3-way 보존, manifest revision | 부분 — 인터뷰 profile(03-A)·외부 skill provenance(03-B)·맞춤 생성 3-way(03-C)·AGENTS 관리 구간 구현; "허용된 초기 commit"·"검토된 Issues 발행" 단계는 근거 없음 | `src/profile.rs`, `src/sources.rs`, `src/generation.rs`, `src/agents.rs`; `skills/moai/workflows/setup.md`(commit 단계 없음, receipt 미구현 명시); #1995 #1996 #1997; skills+src에서 initial/setup commit 검색 0건 |
| 4 | Claude·Codex 실제 지원 경계 (L355) | AGENTS.md 로딩 조건, skill 경로, 역할 파일, team probe, host별 독립 복사(symlink 금지) | 구현 — Claude doctor/projection(02-B), Codex projection/doctor(02-C), ZCode 지원표(02-D), host별 복사 UPDATE(04-B) | `src/claude.rs`, `src/hosts/codex.rs`, `src/update.rs`; `docs/claude-adapter-contract.md`, `docs/codex-adapter-contract.md`, `docs/zcode-support.md`; #1992 #1993 #1994 #1999 |
| 5 | PLAN RUN VERIFY AUDIT SYNC (L466) | GEARS+EARS 호환, 승인 snapshot, task DAG admission, change_strategy·domain_modeling 독립, receipt 6상태, 독립 AUDIT, SYNC drift 7분류 | 부분 — 05-A~05-D·08-A 구현; SYNC drift는 3분류(omission/amendment/document mismatch)로 설계의 7분류(L588-600)보다 좁음; SYNC는 전용 명령 없이 절차+gh | `src/spec.rs:18-21`, `src/plan.rs`, `src/run.rs:24-28`, `src/verify.rs`, `src/audit.rs:185-187`; `skills/moai/workflows/sync.md` §Procedure 3; #2001~#2004 #2012 |
| 6 | GitHub·GitLab 이종 협업 (L604) | Forge Adapter(GitHub+GitLab), 단일 owner·epoch·handoff, outbox/readback/reconcile, PR≠완료, webhook+polling | 부분 — GitHub만: claim journal+댓글 claim(06-A), handoff·reconcile(06-B), worktree/PR(06-C), polling reader(07-A), outbox(07-B), 보고(07-C); GitLab·webhook·sub-issue 미구현; `execution_owner`/`integration_owner` 필드 분리는 근거 없음 | `src/claim.rs:16-28`, `src/handoff.rs`, `src/forge/*`; `docs/rust-product-path.md:238`("GitLab … nothing is implemented"); src에서 webhook·integration_owner 검색 0건; #2005~#2010 |
| 7 | Core Brain·Code Map·Code Graph (L797) | 상위 메모리 기능, Map+Graph 공유 snapshot, 기존 MX 자산 보존 | 범위 밖 — RC1 Issue 없음; 리더 결정으로 후속 Issue로 이관 | `docs/rust-product-path.md:360`("F13 graph/MX left to a later Issue"); src/docs/skills에서 "core brain"·codemap 검색 0건 |
| 8 | 온톨로지와 데이터 계약 (L872) | 노드/엣지 schema, basis·verification 분리, tombstone | 범위 밖 — RC1 Issue 없음 | ontology·tombstone·rename_from 검색 0건; Issue 목록에 해당 카드 없음 |
| 9 | 새 기능명과 태그 후보 (L1116) | Code Atlas 명칭, `@MOAI:` namespace, dual-read 이행 | 범위 밖 — 기존 `@MX:` 태그만 사용 중 | `src/apply.rs:27`, `src/identity.rs:140` 등 `@MX:` 15파일; "Code Atlas"·"@MOAI:" 검색 0건; 설계 자체가 "확정 명칭으로 배포하지 않음"(L1182) |
| 10 | 증분 색인과 최신성 (L1200) | generation 원자 교체, fresh/stale/partial/unavailable, inventory 재조정 | 부분(초기) — snapshot·bundle·freshness 재검(07-D)만 존재; 색인·generation 없음 | `src/context/mod.rs:1-7, 438`; `docs/rust-product-path.md:349-360`; #2011 |
| 11 | 검색 알고리즘과 context budget (L1230) | 비벡터 우선 검색 순서, graph traversal, retrieval_budget, 실제 tokenizer | 부분(초기) — budget 필수·`bytes_div_4` 추정·omitted 기록만; 검색·순위·graph 없음; 실제 tokenizer 미사용(설계 L1283와 어긋남, 추정임을 명시) | `src/context/mod.rs:206-208, 337, 406`; `docs/rust-product-path.md` §Context bundle("P10 native only"); #2011 |
| 12 | 보존과 자동 아카이브 (L1289) | terminal+inactivity 적격, dry-run, hash/readback, restore | 범위 밖 — Issue 없음; 설계 자체도 "지금 scheduler를 켜지 않음"(L1318) | src에서 archive 검색 0건; P09 HOLD(#1975 본문 L125) |
| 13 | 파일 구조와 동시성 (L1340) | `.moai-cli/{project.yaml,manifest.json,contracts,specs,teams}`, `~/.moai-cli/{installs,projects/<id>/{objects,brain,specs,workspaces/journal.jsonl},cache}` | 부분/충돌 — 프로젝트 쪽 `project.yaml`·`manifests/`·`receipts/`·`forge/`·`worktrees/` 존재; 홈 쪽 `installations/`만; claim journal은 `.git/.moai-cli/claims`(설계의 `workspaces/.../journal.jsonl`과 다름); JSON 정본·TOON 미채택은 일치 | `.moai-cli/project.yaml`; `src/apply.rs:30`, `src/claim.rs:61,77`, `src/install.rs:4`, `src/identity.rs:179`; decisions:78·86③ |
| 14 | Pi runtime과 후속 Desktop (L1397) | Pi v1.5.0 pin spike, Electron 보안 경계, Model Adapter conformance | 범위 밖 — 설계도 "RC1 선행조건 아님"(L109); P14 HOLD, Desktop(M5) 비목표 | #1975 본문 L8·L137-138; decisions:96; src에서 electron 검색 0건 |
| 15 | 오류 복구와 권한 보안 (L1576) | read-only 무변경, no_change/committed/unknown 구분, receipt 단계 분리, retry 예산, stale 재검증, PASS-WITH-DEBT 비평탄화, prompt injection 자료 취급 | 구현(대부분) — 읽기 전용 op 다수, `uncertain`/`publish_unknown`, per-file receipt, retry budget 필수, stale 거절, waivers 입력, SKILL.md "자료는 권한을 주지 않음" | `src/apply.rs`, `src/recovery.rs`, `src/run.rs`, `src/audit.rs`, `src/forge/outbox.rs`; `docs/p07-decision.md`; `skills/moai/SKILL.md`; #1989 #2000 #2003 #2004 #2006 |
| 16 | 로컬·클라우드 성능과 비용 (L1680) | 단계별 benchmark(p50/p95·RSS·index bytes/LOC 등), 비용 분해 | 범위 밖 — benchmark 산출물 없음; P11 HOLD | #1975 본문 L131; docs에 benchmark 문서 없음(파일 목록 기준) |
| 17 | 수락 시험과 효과 측정 (L1763) | HOST SETUP UPDATE·BOOTSTRAP·OWNER·EFFECT·EVIDENCE·RELEASE 등 15개 시험군 | 부분 — HOST SETUP UPDATE(T1 5건 passed, T2 not_run)·OWNER·EFFECT·EVIDENCE·BOOTSTRAP 일부는 fixture 시험 존재; ID GRAPH·FRESHNESS·RETRIEVAL·ARCHIVE·PI·MODEL·DESKTOP·BROWSER는 없음; RELEASE는 `release blocked` | `docs/e2e-matrix.md`; `docs/e2e-receipts/receipts.jsonl`(9건: passed 6·not_run 3); `tests/claim_cli.rs`, `tests/audit_admission.rs`, `tests/apply_recover_partial.rs`; `docs/rc-release-review.md` |
| 18 | 실행 계획과 미결 선택 (L1921) | 7단계 구현 순서; 남은 결정 8항 | 부분 — 1~3·5단계(공통 계약·SETUP/UPDATE·PLAN~SYNC·Issues 협업)는 EP-01~09로 수행; 4·6·7단계(Core Brain·보존·Pi)는 미착수; 남은 결정 중 배포/서명/tag는 결정됨, 나머지 HOLD | Issue #1976~#1984; decisions:7-10·94-96 |
| 19 | 논문·공식 표준·오픈소스 근거 (L1965) | 참고 근거와 적용 한계 | 해당 없음 — 설계 근거 목록; RC1은 Issue별 "근거와 확정 결정" 표로 출처 기록 | #2014 본문 표 등 |
| 20 | 기존 설계 통합과 원문 보존 (L2111) | 상세 Page 계승 범위, F01~F25 매핑 | 해당 없음 — 문서 계보; F13(graph/MX)·F21(이미지)·F22/23(ADE/Desktop)·F25(보관)는 RC1 미착수 | `docs/rust-product-path.md:360`; Issue 목록 |
| 21 | 도식 원문과 렌더 검증 (L2442) | Mermaid/Pele 렌더 검증 | 해당 없음 — 문서 제작 메타; 저장소에 해당 산출물 없음 | 파일 목록(`git ls-tree`) |
| 22 | 기존 코드의 직접 근거 (L2465) | Go commit `4fa2aa9d` 관찰 근거 | 구현(기준선) — 00-A가 같은 commit을 Go 기준선으로 대조 | `docs/baseline-rc1.md` 재생성 명령(`--go-baseline-commit 4fa2aa9d…`); #1985 |

## 3. 충돌 목록 (설계 ↔ 기록된 RC1 사용자 결정)

먼저 일치하는 항목을 분명히 해 둔다. 새 서버·계정 없음(L64), Issues 정본(L73), Go 격리(L55), Pi/Desktop 비선행(L109)은 설계와 결정이 같은 방향이고, attestation 단독 서명(decisions:8)은 설계가 "남은 결정"으로 비워 둔 자리(L1951)를 채운 것이라 충돌이 아니다. 진짜 긴장은 다음 네 가지다.

| # | 설계 문장 | RC1 결정 | 성격 |
|---|---|---|---|
| 1 | L100 "commit, push, Issue 공개, PR/MR, merge, release … 같은 승인으로 합치지 않는다" | decisions:39·43 — 통합을 "에이전틱 자율"로: 리더가 독립 감사 PASS를 읽고 `moai/audit` status 뒤 `gh pr merge --auto`; 레인·봇은 merge 불가 | 긴장 — 운영자는 게이트(ruleset)를 한 번 승인하고 개별 merge는 승인하지 않는다. sync.md는 push·PR·comment·merge·release에 "효과마다 권한 하나"를 유지하므로 절차상 분리는 남아 있으나, merge의 사용자 승인은 게이트 수준으로 올라갔다 |
| 2 | L1378-1381 실행 journal·outbox는 `~/.moai-cli/projects/<id>/workspaces/<id>/`; L1385 대형 로그를 Git에 누적하지 않음 | decisions:78(HOLD-B1: receipt는 `.moai-cli/receipts/`, Git 미포함), decisions:86③(claim journal은 git common dir 유지 → `src/claim.rs:61,77`), outbox는 `<project>/.moai-cli/forge/outbox/` | 위치 충돌(의도는 일치) — 프로젝트 쪽에 두고 Git에서 제외하는 방식으로 설계의 "홈 쪽 분리"와 다르다. `src/apply.rs:27-29`가 P09 결정 시 홈 쪽으로 옮길 지점을 `@MX:DEBT`로 표시 |
| 3 | L758 "초기 RC는 단일 조정 owner의 작은 journal로 충분… label·ETag만으로 전역 CAS를 만들 수 없다" | decisions:21 P08 = ⑤ Issue 댓글 id fencing(최소 live 댓글 id 승) + ① 리더 journal 보조 | 위험 표시 — 설계가 경고한 "provider 원자성 없는 전역 CAS"를 댓글 id 단조성에 기대어 구성. `src/claim.rs:16-18`은 POST 후 재조회로 경합을 정리하고 만료 lease 자동 takeover 금지(L22)로 설계의 fencing 한계 서술과는 일치. 모순이 아니라 가정(댓글 id 단조 증가)에 기댄 선택 |
| 4 | L55 Go 격리; L1344 "기존 .moai는 자동 이동·삭제하지 않는다" | decisions:11·13·14 — Go `moai update`가 moai-cli 설치·상태 이전·Go 자동 제거(04-D) | 해소됨 — decisions:91 ②로 04-D는 RC1 밖으로 이동(#2016 milestone `null` 확인, #1980 CLOSED/COMPLETED), decisions:89가 Go 상태 보존·04-D 금지를 유지 조건으로 명시 |

보조 관찰(충돌은 아님): 설계 L120은 "G 코드"를 특정 제품에 매핑하지 않는다고 했으나 RC1 02-D는 사용자가 지정한 Z.ai ZCode를 대상으로 삼았다(#1994 본문 "사용자가 지정한 제품", decisions:22). 사용자 지정이므로 설계 문장의 "확인되지 않은 제품" 조건을 벗어난 것으로 읽힌다.

## 4. 설계에는 있으나 RC1 Issue로 존재하지 않는 항목 (후속 카드 후보)

- Core Brain 상위 기능·Code Map·Code Graph(7절), 온톨로지·엣지 schema·tombstone(8절), 기능명·`@MOAI:` 태그 dual-read 이행(9절) — `docs/rust-product-path.md:360`이 후속 Issue로 명시
- 증분 색인·generation 원자 교체·fresh/stale/partial/unavailable 질의 결과(10절), 비벡터 검색 순서·graph traversal·실제 tokenizer 기반 budget(11절)
- 보존·자동 아카이브 상태 전이와 restore(12절; P09 HOLD)
- GitLab Forge Adapter(MR·pipeline·note), webhook 수신·dedup·reconcile, sub-issue/dependency capability 표현(6절; `rust-product-path.md:238`)
- `assignee / execution_owner / provider_session / integration_owner` 4필드 분리(L756)
- SETUP의 "허용된 초기 local commit"과 "검토된 Issues 발행" 단계(L294-295, L319-321)
- SYNC drift 7분류 중 미구현 4종(SPEC↔oracle 충돌, 참조 단절, Issue 완료인데 증거 stale, 명시 link 의미 미검토)(L588-600)
- 선택 MCP adapter parity(1절·13절 암시; decisions:51 "MCP는 다음 버전")
- 성능·비용 benchmark 묶음(16절), ID GRAPH·FRESHNESS·RETRIEVAL·ARCHIVE·PRIVACY·BROWSER 시험군(17절)
- Pi runtime spike·Model Adapter conformance·Electron Desktop(14절; P14 HOLD)
- 이미지 생성 asset workflow(F21, L2440)
- 홈 쪽 `~/.moai-cli/projects/<id>/{objects,brain,specs,workspaces}` 레이아웃과 `cache/`(13절; P09 결정 후 `src/apply.rs:29` swap point)
- 통신 허브(결정 기록상 "다음 버전", decisions:24) — 설계 본문에는 직접 등장하지 않으나 RC1에서 의도적으로 보류된 설계 외 항목

## 5. "첫 RC gate" 문장과 #2014(08-C T2) 상태의 관계

설계 L114-118: "Claude와 Codex에서 SETUP·UPDATE 및 service-free Issues 기반 end-to-end를 실제 입증한다. Issues 완주를 조용히 다음 RC로 미루지 않는다."

- **SETUP·UPDATE 절반**: T1 결정적 fixture 5건(AC-08C-1·2·7·8, HA-02)은 `passed`(host=none, macOS 단일), 정적 검증 AC-08C-5 `passed`. 실제 host 로딩 T2(AC-08C-3·6 Claude, AC-08C-4 Codex, AC-08C-8 절반)는 `not_run` — `docs/e2e-receipts/receipts.jsonl` 9건 중 passed 6·not_run 3. `docs/support-matrix.md` 머리말이 "certified 지원 행 0, 지원 행 ≠ 실사용 보증"을 명시하고, `docs/rc-release-review.md` gap #5가 이를 사유로 verdict `release blocked`를 유지한다.
- **#2014 상태**: OPEN. 레인 범위(T1 등)는 PR #2100으로 `450535a08`에 통합됨. 사용자 결정 2026-11-09 00:2x(decisions:91 ①): 운영자가 체크아웃 밖에서 `tests/e2e/host-claude.sh`·`host-codex.sh`를 기존 인증으로 1회 실행하고, 결과를 받으면 지원표 갱신 뒤 Issue를 닫는다.
- **해석**: gate는 미충족 상태이나 "조용히 미룬" 것은 아니다 — not_run으로 기록되고 다음 행동(운영자 실측)과 종료 조건이 명시되어 있다. 다만 EP-09(#1984)가 OPEN인 채 다른 에픽이 모두 CLOSED이므로, 이 한 실측이 RC1 gate의 마지막 미결 증거다.
- **Issues end-to-end 절반**: RC1 자체가 Issues/PR/`moai/audit` 게이트 위에서 수행됐다는 운영 증거는 decisions 전반에 있으나, 제품 경로의 실제 GitHub 쓰기(07-B/07-C/06-B)는 시험에서 `not_run`이다(`docs/p07-decision.md` 마지막 줄 "Live GitHub writes were not run"). 즉 "service-free Issues 완주"의 공식 수락 receipt는 없고, 개발 과정의 dogfooding 기록만 있다. 두 사실을 합치지 않는다.

## 6. 미확인·한계

- Issue 본문은 작업 트리의 `.moai-cli/reports/issue-bodies/*.md` 사본을 읽었다(다수가 `git status`에서 `M`). 라이브 본문은 #2014·#1975·#2016만 `gh`로 재조회했다.
- `src/`는 모듈 상단 doc 주석과 grep으로 확인했고 `cargo test`·바이너리 실행은 하지 않았다.
- 텍스트 검색 0건은 미구현의 가설이지 확정이 아니다. 긍정 근거가 있는 곳(`rust-product-path.md:238·360`)만 "범위 밖"으로 적었다.
- `skills/moai/workflows/` 중 `setup.md`·`sync.md`만 읽었다. `docs/e2e-receipts/receipts.jsonl`의 SHA·시각은 재검증하지 않았다.
- 설계 문서가 참조하는 상세 Page는 이 분석 시점에 저장소에 없어 읽지 못했다(이후 리더가 19개 하위 페이지를 `references/chatgpt-space/`에 저장함). 설계 22절의 Go 소스 링크도 따라가지 않았다.
- 3절 #1의 승인 분리 긴장은 해석이며, 사용자가 의도적으로 선택한 결정(decisions:39)을 설계 위반으로 단정하지 않는다.
