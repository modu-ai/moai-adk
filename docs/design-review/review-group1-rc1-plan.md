# "MoAI ADK 2026.11.0-rc.1 온라인 우선 최종 계획" 검토

- 대상: `references/chatgpt-space/page_43faf222414c81919de71f7d7830e26c.md` (ChatGPT Space 텍스트 사본, 2026-11-08T23:19:26Z 읽음, 본문 sha256 `7e13dcd0…f821`)
- 검토자: 리더 세션이 위임한 읽기 전용 검토 서브에이전트, 2026-11-09. 사본 전문(5,523행)을 직접 읽었다.
- 대조 기준: 상위 문서 `page_5490c4ee8c8481918b3fc3217401a3f4.md`(22절), 선행 검토 `review-page_5490c4ee-2026-11-08.md`(A1~A8·B1~B8), `gap-analysis-rc1-vs-design-2026-11-08.md`, `docs/decisions-2026-11-08.md`, 그리고 `origin/v2026.11.0` @ `9bfd26d3b1` (2026-11-09 08:20 KST). 선행 검토 두 편은 `8ac440587` 기준이었고 그 뒤 tip이 움직였다. decisions 행 번호는 작업 트리 사본(`git status`에서 `M`) 기준이며 origin 판과 89행 이후가 어긋나므로 시각 키(예: 00:2x)를 함께 적었다.
- 인용 규약: 이 사본은 같은 페이지를 두 번 담고 있다. 1~1,752행은 표가 평문으로 풀리고 §7 표 중간(1,725행)에서 끊긴 첫 렌더이고, 1,753~5,523행이 Mermaid 원문까지 포함한 완전본이다. 아래 모든 행 번호(L)는 완전본 기준이며, 선행 검토가 이미 다룬 항목은 그 ID(A1 등)로 가리키고 되풀이하지 않는다.

## 1. 문서 개요

이 문서는 RC1의 "메인 계획"이다. 사용자 수준에서 MoAI skill을 한 번 추가하면 Claude Code·Codex Desktop 안에서 자연어로 SETUP·UPDATE·PLAN~SYNC를 수행하고, 별도 MoAI 서버·계정·DB 없이 기존 GitHub/GitLab Issues를 업무 원장으로 쓰는 online-first 제품 경계를 확정한다(L1755-1769). 29개 절로, 제품 경계(§1~5), 기능 F01~F25(§6), API 크레딧·판단 backend(§7), GEARS·방법론·협업·저장소·worktree(§8~12), context·온톨로지·검증·보안·ADE(§13~17), Rust CLI 계약·Desktop·파일 배치(§18~20), Go 이행·릴리스·검증 순서(§21~26), 전수 목록·hook·profile·TOON·명령 전환(§27~29)을 다룬다.
작성일·기준 commit은 머리말에 없다. 날짜는 §7의 "2026년 11월 8일 공식 문서 기준"(L3326)에서, 기준 commit은 Go 기준선 `4fa2aa9db2…`(L5049·L5469)에서만 간접 확인된다. RC1 Rust tip은 인용하지 않는다.
상위 문서 §20(L2115-2116)은 "현재 제품 결정은 [이 문서], 상세 Core Brain과 기능별 문서를 통합한다. 최신 요구와 충돌하는 과거 선택은 역사로 남기며 현재 구현·승인으로 복원하지 않는다"고 적어, 이 문서를 계승 원본으로 삼되 충돌 시 상위 문서가 우선함을 선언한다. 따라서 아래 §3의 불일치는 이 문서가 '역사'로 남는 항목이다.

## 2. 내부 결함

| # | 위치 | 결함 | 성격 |
|---|---|---|---|
| I1 | L4479·L4483·L4591 ↔ L4542 ↔ L5485 | 사용자 홈 상태 root를 `<state_root>`(§20 트리·표), `~/.moai-cli/`(§20 "기존 구조 이행"), `<moai_home>`(§28)의 세 이름으로 부른다. `workspace_root`(`~/moai/`, L4479)까지 합치면 홈 경로 용어가 넷이고 상호 정의가 없다 | 용어 미정의 |
| I2 | L2713 | ROUTING acceptance가 "description에 맞는 자연어 요청의 선택률을 측정"이라고만 적고 임계값·표본 집합·측정 host가 없다. 통과/실패를 판정할 수 없다 | 측정 불가 AC |
| I3 | L5298·L5311 ↔ L5345 | M3·M4 완료 조건("stale/false PASS·비용·비밀 누수 차단", "single-owner·중복 효과·실제 capability")이 정성 문장이고, L5345가 "정량 목표는 baseline 확보 뒤 고정"이라 하므로 현재 측정 가능한 완료 조건이 없다 | 측정 불가 AC |
| I4 | L2686 | HOST-INSTALL 통과 조건의 "지원 OS/host 조합"이 문서 어디에도 열거되지 않는다(§25 L5359가 OS/arch 범위를 미결로 둔다) | 미정의 집합 |
| I5 | L3590 | "현재 Decisions API는 실제 GA 제품이며 POST /v1/decisions, 지원 모델은 gpt-6-luna다"를 사실로 단정한다. 사본에는 URL이 없고 L3434의 "확인" 주체·시각도 없다 | 검증 불가 단정 |
| I6 | L3347-3390·L3394 | 구독별 API 크레딧 금액($100/$200/$500, Max $100, Ultra $200, seat당 $20/$100, 상한 $500)과 적용 제외 규칙이 숫자로 적혀 있으나 사본에는 출처 URL이 남아 있지 않다. 라이브 페이지에는 하이퍼링크가 있을 수 있으나 이 사본만으로는 재확인할 수 없다 | 출처 미제시 숫자 |
| I7 | L5469-5471 | §27의 수치(blob 17,478, 1,107행, 논리 SKILL 82, 지원 파일 631, hook 96, CLI 388행, MCP 38, package 90, workflow module 19, fixture 4,750, Go 파일 1,908·symbol 8,793, template 491=동일 423+차이 66+template-only 2, active-only 48)에 재현 명령이 없고 L5471 "별도 목록으로 제공"은 대상이 없는 참조다. 1,107만 RC1 `.moai-cli/manifests/legacy-asset-map.json`(`assets` 1,107건)과 decisions 18:5x로 뒷받침된다. RC1 `docs/baseline-rc1.md:69-72`는 같은 기준선에서 Go template `.claude` 512, 분류 합계 654(identical 479·baseline-only 89·template-only 52·differing 32·active-only 2)를 얻어 §27의 491/423/66/2/48과 기준도 값도 다르다 | 출처 미제시 숫자 |
| I8 | L5444 | §26은 검토 반영 결과라면서 "새 repo는 선택안"을 유지한다. 같은 §26 표가 '결정 경계'를 다루므로 결정 여부를 표시했어야 한다(RC1 결정은 §4 R5 참조) | 내부 시점 불일치 |
| I9 | L1896 ↔ L2476 ↔ L5181 | 기본 workflow를 §1은 `plan → run → verify → audit → sync`, §5는 intent 8종(READ·SETUP·UPDATE 포함), §22 저장소 구조는 `workflows/ setup·update·plan·run·verify·audit·sync`(READ 없음)로 적는다. 세 목록이 서로 다르다 | 단계 목록 불일치(상위 문서 A7과 같은 계열) |
| I10 | L4913-4943·L4967-4980·L5019-5031·L5036-5043 | `{{…}}` placeholder와 `"canonical_sha256": null`, `"capability_status": "unverified"`가 그대로 남아 있다. L4550·L4982·L5033이 "실제 확인값으로 채우기 전 활성화하지 않는다"고 명시하므로 결함이 아니라 의도된 템플릿이다. 다만 검증 fixture 경로가 없어 "schema·host fixture를 통과"(L4550)의 기준이 문서 안에 없다 | placeholder(명시됨) |
| I11 | L1757(첫 렌더 L7) | "Pele 도식"이 무엇인지 정의가 없다(도구명으로 보이나 설명·링크 없음). 사본 1,757행은 링크 제목이 "페이지 불러오는 중…"으로 치환되어 참조 대상을 잃었다 | 용어 미정의 / 사본 손상 |
| I12 | L3453 | Jev의 `choice, noul, score` 중 `noul`이 설명 없이 등장한다. "명제가 참일 확률"이라는 뒷말로 유추할 수 있을 뿐 정의가 아니다 | 용어 미정의 |

긍정 사항: 이 사본은 Mermaid 원문을 본문에 포함한다(L1983-1997, L2553-2563, L3215-3230, L3455-3468, L3597-3604, L3743-3757, L4430-4442, L4507-4514). 상위 문서 검토의 "도식 미확인" 한계가 여기에는 적용되지 않으며, 도식과 본문 사이에 모순은 발견하지 못했다.

## 3. 상위 문서와의 불일치

| # | 상위 문서 | 이 문서 | 내용 |
|---|---|---|---|
| P1 | §13 L1347-1383 | §20 L4459-4494 | 파일 배치가 서로 다르다. 상위: 저장소 `.moai-cli/{project.yaml, manifest.json, contracts/{spec-index.json,policy.json}, specs/, teams/}`, 홈 `~/.moai-cli/{config, installs/<host>/<version>, adapters, projects/<id>/{objects, manifests, brain, specs, workspaces/<id>/{overlay,attempts,journal.jsonl,outbox}}, cache}`. 이 문서: 저장소 `.moai-cli/{project.yaml, contracts.lock, policies/, specs/, workflows/, manifests/}`, 홈 `<state_root>/{skills/moai/<version>, projects/<id>/{registry.json, tasks.md, contracts/<digest>, artifacts/sha256, indexes, runs/<run-id>, workspace-state, retention}, cache}`. 두 설계 모두 RC1 실물과 다르다(§4 R1·R2; 상위 쪽은 A2) |
| P2 | §13 L1378-1382 | §20 L4491·L4682 | 실행 journal·outbox 위치가 상위는 `projects/<id>/workspaces/<workspace-id>/`, 이 문서는 `projects/<project-id>/runs/<run-id>/journal.jsonl, outbox/`다. 단위(workspace ↔ run)가 다르다 |
| P3 | §9 L1182-1184·§18 L1949 | §14 L4162-4174 | 상위는 Code Map+Graph 묶음명 `Code Atlas`와 새 태그 namespace `@MOAI:`를 추천하고 §18에서 미결로 둔다. 이 문서 §14는 `@MX` 6종(NOTE·WARN·ANCHOR·DEBT·TODO·LEGACY)만 쓰고 `@MOAI:`·Code Atlas를 언급하지 않는다. 상위 §20 원칙대로면 이 문서의 `@MX` 서술이 '역사'인데, 그 표시가 없다 |
| P4 | §2(SETUP·PLAN·RUN·VERIFY·AUDIT·SYNC) | §5 L2476 | 상위 §2 단계 목록에 UPDATE가 없고(A7), 이 문서는 READ·UPDATE를 포함한 8 intent를 쓴다. 이 문서 쪽이 RC1 `skills/moai/SKILL.md`(routes READ·SETUP·UPDATE·PLAN·RUN·VERIFY·AUDIT·SYNC)와 일치하므로 상위를 고칠 항목이다 |
| P5 | §12 L1318 | §20 L4524 | 상위는 "지금 scheduler를 켜지 않음"을 명시하는데, 이 문서 §20의 GC(mark·tombstone·유예·sweep)는 비활성 조건이나 RC 범위 표시가 없다. 상위 B4와 같은 계열 |
| P6 | §6 L756 | §10 L3761 | 상위는 `assignee / execution_owner / provider_session / integration_owner` 4필드를 분리하지만 이 문서 작업 레코드는 `owner` 한 필드와 attempt별 `provider session ID`만 둔다. 모순은 아니나 통합 책임(integration_owner)이 이 문서에는 없다 |

일치하여 유지할 점: 첫 RC gate(상위 §1 L114-118 ↔ 이 문서 §24 L5236), Desktop 비선행(상위 L109 ↔ L4426), 승인 분리(상위 L100 ↔ L3093 "sync가 자동 push/release는 아님"), symlink 공유 금지(L2201·L2387)는 같은 방향이다.

## 4. RC1 결정·실물과의 불일치

판정 어휘: 충돌(양쪽 확인) / 긴장(의도는 같으나 방식이 다름) / 구현 근거 미발견(검색 0건 — 가설이지 미구현 확정이 아님) / 근거 없음(확인 못 함).

| # | 이 문서 | RC1 결정·실물 | 판정 |
|---|---|---|---|
| R1 | L4108·L4481·L4578 worktree 기본 `~/moai/worktrees/<project>/<task>`, 주 checkout `~/moai/projects/<project>`; L4501 "기본 checkout·worktree 배치는 ~/moai/…" | decisions 16:5x→17:0x(34행): primary 안 `.moai-cli/worktrees/<issue>-<key>`가 Claude·Codex 공통 기본. `src/worktree.rs:12,347-348`, `.gitignore:361`(`/.moai-cli/worktrees/`), `docs/rust-product-path.md:290`. `~/moai/`는 src·docs·skills 검색 0건 | 충돌(기본 위치) |
| R2 | L4483-4494 홈 트리 `<state_root>/skills/moai/<version>/`, `projects/<id>/{registry.json, contracts, artifacts, indexes, runs, workspace-state, retention}`, `cache/` | RC1 홈 실물: `~/.moai-cli/installations/<install_id>.json`(`src/install.rs:4,340`), `installations/<id>/hosts/<host>.json`·`hosts/<host>/known-good/<hex>`(`src/update.rs:18,335`; `docs/rust-product-path.md:460,466`), `~/.moai-cli/bin/moai-cli`(decisions 15:5x 10행). 이 문서에는 `installations/`·`bin/`·known-good이 없고, RC1에는 `skills/<version>/`·`registry.json`·`runs/`가 없다 | 충돌(홈 배치) |
| R3 | L4464·L4770-4778 `.moai-cli/contracts.lock` tracked, "hash가 미정이면 승인 완료로 취급하지 않음" | `origin/v2026.11.0` 추적 트리의 `.moai-cli/`는 `manifests/`·`project.yaml`·`reports/`뿐. `contracts.lock`은 src·skills 검색 0건, `docs/github-plan-proposal.md:26,69,100`에 설계 인용만 있음 | 구현 근거 미발견 |
| R4 | L2567-2568·L4394-4410·L5521 내부 CLI `moai-cli setup`, `maintenance update`, `doctor`, `capabilities`, `skill show`, `project init --hosts … --apply-plan`, `project connect`, `spec plan/verify/sync`, `task submit/status/cancel`, `artifacts list`, `config explain` | `src/main.rs:613-660` OPERATIONS: `package-inspect`, `project-plan/apply`, `agents-plan/apply`, `apply-recover`, `baseline-inventory`, `identity`, `forge-read/report`, `plan-admit`, `claude-doctor`, `codex-doctor`, `claim-*`, `effect-*`, `run-verify`, `context-bundle/verify`, `audit-reduce`, `install-inspect`, `update-plan/apply`, `recovery-plan/apply`, `worktree-plan/apply`, `github-plan/apply`, `pr-*` 등 47개. 문서 이름은 하나도 출하되지 않았다. 문서가 스스로 "계약안"이라 하므로 허위는 아니나 참조 가치가 없다 | 낡음(이름 전부 상이) |
| R5 | L5164 "Git tag 후보는 v2026.11.0-rc.1, package version 2026.11.0-rc.1, 채널은 next"; L5444 "새 repo는 선택안" | decisions 15:5x(9행): 같은 repo, 접두사 `moai-cli-v<version>`; `docs/release-checklist.md:8,44`; `docs/rust-product-path.md:452`는 평문 `v2026.11.0-rc.1`이 Go RC 채널 필터에 걸려 "업데이트 있음"으로 읽힌다는 특성화 시험을 기록. 즉 이 문서의 후보가 바로 위험한 쪽이다. L5444의 "Go latest는 release가 있으면 prerelease보다 우선" 서술은 stable 채널에만 맞고 RC 채널 위험을 빠뜨린다 | 충돌(tag 접두사) |
| R6 | L5169-5178 OSS 구조 `crates/{moai-domain, moai-cli, moai-config, moai-spec, moai-evidence, moai-execution, moai-storage}` workspace | `Cargo.toml:1-4` 단일 package `moai-cli` 2026.11.0-rc.1, 의존성 4개(clap·serde·serde_json·sha2); decisions 18:5x(54행) "단일 crate `src/spec.rs`". 문서는 "책임 경계이지 독립 package 아님"(L5214)이라 완화하나 crate 분할 자체가 채택되지 않았다 | 낡음(구조) |
| R7 | L4460·L4739 `AGENTS.md` "기본 tracked"; L4461·L4752·L4940 `AGENTS.local.md` "Git 제외·ignored" | RC1 개발 저장소: `.gitignore:349-356`이 `/AGENTS.md`를 무시(기계별 생성)하고 "AGENTS.local.md is explicitly allowed to be tracked"; `AGENTS.local.md`가 추적 blob으로 존재(`git ls-tree`). decisions 17:3x(45행) AGENTS 역할 분리 | 긴장 — 이는 개발 저장소 자체의 사용자 결정이며 제품 템플릿 규칙과의 충돌로 단정하지 않는다. 제품 기본값이 RC1 템플릿에서 어떻게 생성되는지는 확인하지 않았다(근거 없음) |
| R8 | L3765 "RC 기본은 한 task의 단일 dispatcher와 명시적 handoff"; L3781 "Issue의 assignee·label·댓글을 원자적 lock으로 사용하지 않는다" | decisions 16:4x(21행) P08 = ⑤ Issue 댓글 id fencing 주 + ① 리더 journal 보조; `src/claim.rs:1-2,15-21` "⑤ comment claim is primary … smallest comment id wins". RC1은 댓글 id 단조성을 cross-host 조정에 쓴다 | 긴장 — gap-analysis §3 #3이 상위 L758에 대해 같은 평가를 했고, 이 문서는 "댓글"을 명시적으로 금지해 표현이 더 강하다. 모순이 아니라 가정에 기댄 선택이므로 문서에 그 선택을 기록할 자리가 필요하다 |
| R9 | L4104·L4110-4120 Landing 5단계(pin 확인→candidate→semantic conflict→정확 revision 검사→승인 범위·branch protection 확인) | decisions 17:3x(39·43행): ruleset `RC1 integration gate`(PR 필수, status `moai/audit`, strict up-to-date), verdict 첫 줄 형식, 리더가 status 게시 후 auto-merge. 이 문서에는 required status·독립 verdict·head==audited_sha 확인이 없다 | 긴장(상위 A5와 같은 계열) |
| R10 | L5236 RC1 필수 gate: 두 host 실제 발견·호출·SETUP·UPDATE + Issues 기반 online 완주 + native 판단 fallback | gap-analysis §5: 08-C T2 실제 host 로딩 `not_run`(receipts 9건 중 passed 6·not_run 3), 제품 경로의 실제 GitHub 쓰기 `not_run`(`docs/p07-decision.md`), `docs/support-matrix.md` 머리말 "certified 0". decisions 00:2x(11-09, 91행)가 운영자 1회 실측으로 종료 조건을 정했다 | 미충족 상태(문서 결함 아님) — 문서에 "RC1 gate 상태" 링크가 없다(상위 B8과 동일) |
| R11 | L5483-5485 §28 MoAI 관리 hook, `<moai_home>/hooks/packages/<version>/<host>/`·`hooks/manifest.json` | decisions 19:1x(66·67행) 02-B·02-C 모두 "hook 0개"; 04-B는 host 설정 key를 P04 HOLD로 둠(`docs/rust-product-path.md:460`) | 낡음/비범위 — RC1은 hook을 출하하지 않았다 |
| R12 | L2639 "CWD/상위의 CLAUDE.md·.claude/CLAUDE.md·CLAUDE.local.md가 있으면 기본 AGENTS 로딩을 억제할 수 있다" | decisions 19:4x(74행) #2044 FAIL 사유 "HOME/.claude/CLAUDE.md 억제 오판" — 억제 규칙의 범위(HOME 쪽 포함 여부)가 RC1 감사에서 쟁점이었다. #2044 본문은 읽지 않았다 | 근거 부분 — 저자가 02-B의 확정 규칙과 대조할 것 |
| R13 | L2769 ZCode 등은 "공식 확장 경로 확인 뒤 확대"; L4339 Orca "Claude Code·Codex·ZCode 지원" | decisions 16:4x(22·23행) ZCode 헤드리스만 RC1, Orca/Hide RC1 제외; 02-D(#1994) 지원표 ZCode 행을 `app`/`cli` 두 줄로(61행) | 일치(확대 보류)이나 RC1이 ZCode 지원표 행을 이미 만든 사실이 문서에 없다 |
| R14 | L1961·L2288 MCP "선택적 1급 adapter"; L5397-5405 | decisions 18:2x(51행) HOLD-C1 "CLI만, MCP는 다음 버전" | 일치(선택이므로 충돌 아님) — RC1 시점 표시만 없음 |
| R15 | L5154·L5519 "테스트 완료 후 기존 Go 설치를 제거하고 moai 명령으로 전환… 지금 제거하거나 전환하지 않는다" | decisions 16:0x~16:4x(11-15·20행) 04-D bridge 설계 뒤 00:2x(91행) 04-D RC1 밖 이동, Go 상태 무변경 유지 | 일치(지금 아님) — 04-D 결정 이력은 반영되지 않음 |

## 5. 낡은 내용

- §18 내부 CLI 이름 전부(L4394-4410)와 §5·§29의 `moai-cli project init`(L2567·L5521) — RC1 OPERATIONS 표(R4)로 대체됨.
- §22 tag 후보 `v2026.11.0-rc.1`·채널 `next`(L5164) — `moai-cli-v<version>`·prerelease 게시 절차(`docs/release-checklist.md:75`)로 대체됨(R5).
- §12·§20의 `~/moai/projects`·`~/moai/worktrees` 배치(L4108·L4479-4481·L4565-4586) — `.moai-cli/worktrees/`(R1).
- §20 홈 트리(L4483-4494) — `installations/…/hosts/…/known-good`(R2).
- §22 crate workspace(L5169-5178) — 단일 crate(R6).
- §24 M0~M5 단계표(L5251-5337) — RC1은 EP-01~09·Issue 32장(#1975 이하)으로 수행했고 M-번호를 쓰지 않는다(gap-analysis §2 18행). L1777 "9개 영역·32개 작업의 분해는 검토안이며 GitHub에 발행된 항목이 아니다"는 발행 전 문장이다.
- §26 L5444 "새 repo는 선택안" — 같은 repo로 결정(decisions 9행).
- §7 구독 크레딧·Decisions·Jev backend(L3326-3616) — decisions 16:4x(24행) 허브 보류, 51행 MCP 후속; RC1 Issue에 판단 backend 카드 없음(gap-analysis §2 7·11행 "범위 밖"). 문서 자체도 "mock adapter부터"(L3616)라 하므로 RC1 비범위 배너가 필요하다.
- §19 Desktop·§23 일일 공식 변경 검토(L4424-4450·L5226-5232) — P14 HOLD, RC1 Issue 없음.
- §4 Claude/Codex cloud adapter(L2101-2194) — RC1 Issue 없음(gap-analysis §4). 구현 상태는 근거 없음.
- §28 hook 소유권(L5479-5485) — RC1 hook 0개(R11).
- L1773 "사용자 확정 83항목, 제공 기능 25개, 구현 검증 19항목" — 체크리스트 페이지 수치이며 decisions 표의 결정 행과의 대응표가 없다.

## 6. 개선 제안 (영향 순)

1. §20 파일 배치(L4459-4494·L4565-4716)를 RC1 출하 경로(`~/.moai-cli/installations/<id>/hosts/<host>/known-good`, `~/.moai-cli/bin`, `<primary>/.moai-cli/worktrees/<issue>-<key>`, `.moai-cli/{project.yaml,manifests/,receipts/}`)로 고치거나 "RC1 출하 / 후속 목표" 두 열로 나눈다.
2. §22 L5164 tag 후보를 `moai-cli-v<version>`·`--prerelease --latest=false`로 바꾸고, L5444 서술에 RC 채널 특성화(`docs/rust-product-path.md:452`)를 연결한다.
3. §18 명령 목록(L4394-4410)을 `src/main.rs` OPERATIONS 47개로 교체하거나 "RC1 미출하 계약안" 배너를 단다.
4. `<state_root>`·`<moai_home>`·`~/.moai-cli`·`workspace_root`를 한 절에서 정의하고 나머지는 참조만 한다(L4479·L4483·L4542·L5485).
5. ROUTING(L2713)에 표본 집합·임계값·host, M3·M4(L5298·L5311)에 측정 지표 이름과 "baseline 확보 뒤 수치 고정" 조건을 명시해 측정 가능하게 만든다.
6. §7·§19·§23·§28·§4 머리에 "RC1 비범위, 선행조건 아님" 배너를 두고 decisions의 보류 행(24·51·96행)을 인용한다.
7. §27 수치(L5469-5471)에 재현 명령과 산출물 경로를 붙이거나, 1,107(`legacy-asset-map.json`)만 남기고 나머지는 RC1 `docs/baseline-rc1.md` 값으로 교체한다.
8. §10 L3781에 RC1 P08 선택(댓글 id fencing + 만료 lease 자동 takeover 금지)을 "가정에 기댄 선택"으로 기록하고, §12 Landing에 required status `moai/audit`·독립 verdict 첫 줄 형식·head 일치 확인을 추가한다.
9. §14 `@MX` 서술과 상위 §9 `@MOAI:`·Code Atlas 추천을 대조해 어느 쪽이 결정인지 적는다(상위 §18 L1949 미결 표시와 일관되게).
10. §24 L5236 옆에 "RC1 gate 상태" 링크(`docs/support-matrix.md`, `docs/e2e-receipts/receipts.jsonl`, #2014)를 두어 미충족 항목(T2 실측·실제 GitHub 쓰기)이 보이게 한다.
11. §1·§5·§22의 단계 목록(L1896·L2476·L5181)을 §5의 8 intent로 통일하고 READ를 workflows 목록에 포함한다(RC1 `skills/moai/workflows/read.md` 존재).
12. §5 L2639 CLAUDE.md 억제 규칙을 02-B 확정 규칙(#2044 재감사 결과)과 대조해 HOME 쪽 처리를 명시한다.

## 7. 미확인·한계

- 사본 중복(1~1,752행 첫 렌더, 1,753~5,523행 완전본)으로 행 번호는 완전본 기준만 유효하다. 첫 렌더는 §7 표에서 끊긴다.
- 사본에는 하이퍼링크 URL이 없다. §4·§5·§7·§11·§12의 "공식 근거" 문장(예: Codex cloud Experimental, jj 0.46.0·Git 2.42, GitHub Issue Artifacts preview, Decisions API GA)은 라이브 페이지에서만 확인할 수 있어 재검증하지 않았다.
- RC1 실물은 `origin/v2026.11.0` @ `9bfd26d3b1`의 트리를 `git show`·`git ls-tree`·`git grep`으로 읽었다. `cargo test`·바이너리 실행·Issue 본문 라이브 조회는 하지 않았다. 검색 0건은 가설이며 "구현 근거 미발견"으로만 적었다.
- `src/main.rs`는 600~680행, `src/claim.rs`는 1~40행만 읽었다. `src/install.rs`·`src/update.rs`·`src/worktree.rs`는 grep 결과 행만 확인했다.
- R7(AGENTS 추적 정책)은 개발 저장소의 `.gitignore`와 추적 blob만 확인했고, RC1 생성기가 사용자 프로젝트에 만드는 기본값(템플릿 쪽)은 확인하지 않았다.
- R12(#2044 CLAUDE.md 억제)는 decisions 한 행에 의존하며 PR 본문·감사 댓글은 읽지 않았다.
- 상위 문서는 §1·§9·§13·§18·§20만 다시 읽었고 나머지 절은 선행 검토 두 편의 기록을 따랐다.
- 이 검토는 문서 품질과 정합성 관점이며 설계 방향의 옳고 그름을 판정하지 않는다.
