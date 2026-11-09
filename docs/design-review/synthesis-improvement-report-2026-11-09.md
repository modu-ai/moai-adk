<!-- MoAI ADK 환골탈태 통합 최종 설계 — 전체 하위 문서 포함 개선점 종합 보고서 | source: ChatGPT Space page_5490c4ee + 하위 19 페이지 (references/chatgpt-space/page_*.md, 2026-11-08T23:00~23:19Z 수집) | written: 2026-11-09T00:0xZ (leader [a963ab]) | inputs: review-page_5490c4ee-2026-11-08.md · gap-analysis-rc1-vs-design-2026-11-08.md · review-group1-rc1-plan.md · review-group2-core-brain-sync-owner.md · review-group3-install-host-plan-run-release.md · review-group4-past-designs.md | RC1 대조 기준: 상위 검토·대조표 8ac440587, 그룹 1~3 9bfd26d3b, 그룹 4 d253f4a81; 현재 tip 654a1f9bb — 본 보고서는 재검증하지 않음 -->

# MoAI ADK 환골탈태 통합 최종 설계 — 하위 문서 전체 개선점 종합 보고서

## 0. 이 보고서의 성격

- 대상은 ChatGPT Space의 상위 설계 1편(22절, 2,494행)과 그 페이지가 링크한 하위 19편(합계 약 31,000행)이다. 20편 모두 텍스트 사본 전문을 읽었다(상위·대조표는 리더가, 하위 19편은 읽기 전용 검토 세션 4개가 분담).
- 이 문서는 **새 사실을 주장하지 않는다.** 일곱 검토 파일이 이미 확인한 항목을 합치고 순위를 매긴다. 각 항목 끝의 괄호는 출처 검토 파일과 절이다. 하위 검토가 "가설"(문자열 검색 0건)·"근거 없음"으로 적은 등급은 그대로 옮겼고 "미구현"으로 승격하지 않았다.
- RC1 소스 대조는 검토마다 기준 SHA가 다르다(머리말 주석). `src/claim.rs`·`src/handoff.rs`·`src/worktree.rs` 인용 행은 `9bfd26d3b` 기준이며 현재 tip `654a1f9bb`에서는 밀렸을 수 있다. `cargo test`·바이너리 실행·외부 URL 재조회는 어느 검토도 하지 않았다.
- 설계 방향의 옳고 그름은 판정하지 않는다. 문서 품질과 RC1 실물·결정 기록과의 정합성만 다룬다.

## 1. 한 줄 요약

설계의 원칙(근거 없는 PASS 금지, 승인 분리, 단일 writer, Go 격리, symlink 미계승)은 RC1이 실제로 구현한 방향과 같다. 고칠 것은 원칙이 아니라 **어휘·경로·인터페이스명의 분산**(홈 경로 이름 4종, 상태 어휘 5벌, 내부 명령 체계 4종, 수용 시험 ID 8체계)과 **RC1 결정이 문서에 되돌아가지 않은 것**(이미 결정된 항목이 "미결"로, RC1이 택한 claim 방식이 "금지"로, 범위 밖인 MCP·hook이 RC1 시험처럼)이다.

## 2. 우선 고칠 것 — 교차 상위 12건 (위험 순)

| # | 문제 | 어디에 있나 | 왜 위험한가 | 고칠 방법 | 출처 |
|---|---|---|---|---|---|
| S1 | **실행 소유권의 권위 방향이 RC1과 반대로 읽힌다.** 문서는 "journal이 authority, Issue 댓글은 projection / 댓글·label·assignee는 lock 대체물 아님"이라 적는데, RC1 P08 결정(decisions:21)은 ⑤ Issue 댓글 id fencing(가장 낮은 live 댓글 id 승)이 cross-host claim의 1차 수단이고 ① journal은 같은 host 직렬화 보조다(`src/claim.rs:1-23`) | 상위 §6 L758; [66a] L1270-1272; [0259] L69; [3d9] L167; [43fa] L3781 | 다음 구현자가 문서대로 journal을 primary로 만들면 RC1 06-A와 정반대가 된다. 네 검토가 독립적으로 같은 결론 | §6에 claim 댓글 프로토콜(`moai-claim: session= epoch= lease_until=`, 최소 id 승, `Moai-Claim: <issue>#<comment-id>` 토큰의 trailer·PR 전파, 만료 시 자동 takeover 금지)을 적고 L758·[66a] L1270·[0259] L69를 "GitHub에 원자 primitive가 없으므로 댓글 id 단조성을 가정한 eventual claim"으로 다시 쓴다. journal schema에 `claim_ref` 필드 추가 | 상위 A4; 대조 충돌 3; G1 R8; G2 ③; G3 X7·§9#1 |
| S2 | **"남은 결정"이 이미 결정됐다.** RC repository(같은 repo)·tag `moai-cli-v<version>`·attestation 단독·cargo-dist(A)·`~/.moai-cli/bin`·Windows self-replace는 decisions:7-12에 결정돼 있는데 문서는 미결로 둔다. 특히 [43fa] L5164 tag 후보 `v2026.11.0-rc.1`은 Go RC 채널 필터에 걸려 "업데이트 있음"으로 읽히는 바로 그 위험한 이름이다(`docs/rust-product-path.md:452`) | 상위 §18 L1951; [25f] L197·L271; [66a] L1663; [43fa] L5164·L5444 | 결정과 반대인 값이 "후보"로 남아 있으면 그대로 채택될 수 있다 | §18·[25f] §5·§10을 "결정됨(decisions 행) / 구현 상태(release-checklist.md)" 두 열로 다시 쓴다. 결정됐지만 미구현인 3건(cargo-dist 미도입·self-replace가 `Cargo.toml`에 없음·attestation 생성 실측 없음)을 명시 | 상위 A1; G1 R5; G3 X4·§9#2 |
| S3 | **홈·프로젝트 경로 이름이 네 가지.** 상위 `installs/<host>/<version>/`·`workspaces/<id>/{journal.jsonl,outbox}`, [83f] `<state_root>/skills/moai/<version>/`·`installations/`·`runs/`, [487]·[43fa] `projects/<id>/{registry.json,contracts,artifacts,indexes,runs,retention}`·`~/moai/worktrees`, Core Brain은 같은 트리를 두 벌로 그림. RC1 실물은 홈 `~/.moai-cli/installations/<id>.json`·`…/hosts/<host>/known-good/<hex>`·`bin/`, 프로젝트 `.moai-cli/{project.yaml,manifests/,receipts/,forge/,decisions/,evidence/,worktrees/}`, claim journal `<common dir>/.moai-cli/claims/` | 상위 §13; [83f] §1; [487] §3; [43fa] §20; Core Brain L710-733·L3802-3834 | 문서와 같은 이름은 [83f]의 `installations/`뿐. `~/moai/`(점 없음)는 Go에도 RC1에도 없다 | 경로를 한 곳(상위 §13)에서만 정의하고 "RC1 출하 경로 / 후속 목표(P09)" 두 열로 나눈다. 나머지 문서는 참조만. `<state_root>`·`<moai_home>`·`~/.moai-cli`·`workspace_root` 네 용어를 한 절에서 정의 | 상위 A2; G1 I1·P1·R1·R2; G2 ①§3·⑥#2; G3 X1·X3·X6 |
| S4 | **상태 어휘가 다섯 벌 이상, 대응표 없음.** [83f] 8단계 discovered…host_verified, [25f] 10개, [e08] Draft…Cancelled 10개, [3d9] Ready→Running→Review→Verified, [66a] 11개, [0259] 14개, [0d45] 효과 상태 3벌. RC1은 receipt `pending|committed|recovered`, admission `ready|stale|blocked`+사유 코드, claim ACTIVE 8+outcome 5, sync 7+4(+`needs_rebase`) | 거의 모든 하위 문서 | `blocked`·`stale`·`unknown`·`reconciling`이 기계마다 다른 뜻. `docs/protocol-v1.md`에 69 토큰 crosswalk가 있는데 어느 문서도 참조하지 않음 | 상태 기계별 정본 1벌을 정하고 RC1 어휘와의 대응표를 상위 부록으로 둔다. [66a]만 RC1 claim 상태와 거의 같으므로 그것을 기준으로 삼고 `integrated`는 "리더 게이트, RC1 상태 아님" 표시 | G2 ⑥#5; G3 X5·§9#5; G3 §4.2·§5.2 |
| S5 | **MCP·hook을 RC1 시험 대상처럼 서술.** RC1 결정은 Claude hook 0·Codex hook 0(decisions:66-67), MCP는 다음 버전(decisions:51). [dff] HA-05·07·08·09, [83f] HOOK-EVENT-PROBE, [c4f9] L204, [43fa] §28이 측정 불가인데 그 표시가 없다 | [dff] §5·§9; [83f] §6; [c4f9] L37·L204; [43fa] L5479-5485; [25f] RL-08 | "RC acceptance" 10건 중 RC1 receipt는 HA-02 1건. 나머지는 통과도 실패도 아닌데 독자는 알 수 없다 | 해당 절 머리에 "RC1 범위 밖(hook 0개·MCP 다음 버전·TOML 미처리)" 배너. 각 시험 행에 RC1 상태(passed receipt / not_run / not_applicable)와 `docs/support-matrix.md` 링크 | G3 §2.6#1-2·X8·§9#4; G1 R11·R14; G2 ④#6 |
| S6 | **Core Brain TOON 결론이 정면 충돌.** L4438 "조건부 채택·항목마다 다시 묻지 않음" + L4446 FORMAT-* 7개 시험 ↔ L4450 "TOON은 절대 도입하지 않는다. 설계에서 제외." 상위 §13 L1395는 전자 쪽 | Core Brain L4436-4450; 상위 §13 L1395 | 덧붙인 지시문으로 읽히며 작성 주체 불명. 제외가 결정이면 시험 7개와 상위 한 줄이 함께 죽는다 | 제외가 결정이면 L4436-4448을 "제외 결정(출처·일시)"로 바꾸고 FORMAT-*·상위 L1395를 지운다. 채택이면 L4450 삭제 | G2 ①§2·§6#1 |
| S7 | **상위 §20 "이전 설계 교훈" 표의 귀속 오류.** 비계승 열이 문서 1·4·5에서 그 문서가 본문에서 명시적으로 막은 결함을 문서 주장인 양 적었고, 계승·비계승 항목이 문서 사이에 엇갈린다(문서5↔3 cancel·배포 SHA, 문서6↔1 Pele, 문서2↔1 고정 반복 간격, 문서3↔4 Claude 우선). 비계승으로 보낸 lane·lease·epoch·agent-N·Issue별 PR은 RC1이 그대로 다시 썼다 | 상위 §20 L2244-2341 | 표가 "역사"로 보낸 것과 RC1 실물이 반대. "보고만으로 CI 재시작하지 않기"는 남긴다고 적었으나 본문 0건 | 각 비계승 행을 "문서가 실제로 제안했으나 채택하지 않은 선택지 + 문서 줄"로 고쳐 쓰고 "RC1에서 재채택됨" 열을 추가. 교훈 10건(§7.2)을 §1-§18 본문에 실제 문장으로 넣는다 | G4 요약·§7.3; 각 문서 §N.5 |
| S8 | **같은 F 번호가 다른 기능.** [83f] F20 API 예산·F22 저장·F23 릴리스·F24 Desktop·F25 호환 vs 상위 §20 F20 판단 backend·F22 ADE·F23 Desktop·F24 유지보수·F25 보관. [66a] L556이 경고만 하고 어느 쪽도 고치지 않음 | [83f] L941-1046; 상위 §20 L2407-2434 | 번호로 누락 검사를 하면 틀린다 | 한쪽으로 통일하고 [66a] L556 경고를 제거 | G3 §1.2·X13 |
| S9 | **단계 목록이 세 가지.** 상위 §2·§5는 UPDATE·READ가 없고(SETUP·PLAN·RUN·VERIFY·AUDIT·SYNC), [43fa] §1은 5단계, §5는 8 intent, §22는 7 workflows(READ 없음). RC1 `routes.json`은 READ·SETUP·UPDATE·PLAN·RUN·VERIFY·AUDIT·SYNC 8개 | 상위 §2·§5; [43fa] L1896·L2476·L5181; Core Brain L280 | 단계 이름이 흔들리면 수용 시험·상태 기계의 적용 범위가 흔들린다 | 상위 §2에 8개로 고정하고 모든 하위는 참조만. `read.md`가 RC1에 존재하므로 READ 포함 | 상위 A7; G1 I9·P4·§6#11; G2 ①§3 |
| S10 | **RC1 통합 게이트와 복구 메커니즘이 문서에 없다.** ruleset `RC1 integration gate`(PR 필수·status `moai/audit`·strict up-to-date), verdict 첫 줄 형식 `AUDIT #<n> (<key>) head <40hex> base <40hex> — verdict: PASS|FAIL`, 리더 status 게시 후 auto-merge, 독립성 표(fresh/fork/separate-session/same-session); known-good 사본·`version_before`·pending→committed→recovered receipt | 상위 §5 통합 행·§15; [0d45] L48; [c4f9] L10·L59; [66a] L1264; [43fa] L4104-4120; [0259]·[487] 복구 절 | "별도 통합 owner·merge 승인"만 있고 메커니즘이 없어 승인 분리 원칙과 RC1 게이트 수준 자율 병합(decisions:39) 사이 긴장이 보이지 않는다 | §5 통합 행에 required status + 독립 verdict + head==audited_sha 확인을 RC1 확정 방식으로 적고, §15에 known-good→recovery-plan/apply→recovered 한 행 추가 | 상위 A5·A6; 대조 충돌 1; G1 R9; G2 ②#2·④#2; G3 §7.4 |
| S11 | **내부 CLI 명령 체계가 문서마다 다르고 RC1과도 다르다.** [25f] `setup/project init/maintenance update/doctor/capabilities`, [e08] `spec plan/task status/skill show`, [3d9] `task submit/cancel, spec verify, artifacts list`, [43fa] §18 전부 vs RC1 OPERATIONS 47개(`project-plan/apply`, `update-plan/apply`, `plan-admit`, `run-verify`, `audit-reduce`, `claim-*`, `effect-*`…). 인터페이스명도 `result.admit`/`evidence.admit`/`verify.accept` vs RC1 `audit.reduce`만 | [25f] L114; [e08] L157-160; [3d9] L267-272; [43fa] L4394-4410; [0259] L160; [c4f9] L192; [e361] L165 | 공통 명령은 `capabilities`뿐인데 RC1에 없다. 참조 가치가 0 | `src/main.rs` OPERATIONS 표로 교체하거나 "RC1 미출하 계약안" 배너. 인터페이스명 대응표 1장 | G1 R4·§6#3; G2 ⑥#3; G3 X9 |
| S12 | **근거 없는 수치·확인일 없는 외부 사실.** "Hook 96개"(Go 실물 82), "190개 Claude 비교", "13 harness 소비자", "Codex 기본 14항목", "20개 후속 시험"(실제 18), §27 수치 다수(RC1 `baseline-rc1.md`와 기준·값 다름), 구독 크레딧 금액·OpenAI Decisions GA·vendor credit 정책이 URL·날짜 없이 사실로 | [83f] L1214; [66a] L408; [dff] L483·L942; [43fa] L3347-3390·L3590·L5469; [e361] L147-149 | 상위 §18이 미결로 둔 자리를 하위가 날짜 없이 채운 경우(vendor credit)는 결정처럼 읽힌다 | 검증된 수치(12 events·65.6KB·22.5KB·32 KiB·v2.1.277·라벨 26)와 미검증 수치를 표기로 구분. 외부 사실에 확인 날짜·URL·적용 상품을 붙이거나 미결로 되돌림 | G1 I5-I7; G2 ⑤#1; G3 §1.2·§2.2·§7.2 |

## 3. 문서별 개선 요약 (20편)

각 행은 그 문서 검토의 "개선 제안(영향 순)" 상위 2~3건이다. 전체 목록은 출처 열의 파일·절을 연다.

| 문서 | 핵심 개선 | 잘된 점(유지) | 출처 |
|---|---|---|---|
| 상위 `5490c4ee` 환골탈태 통합 최종 설계 | A1 결정됨↔미결 분리, A2 §13 경로 실물화, A3 §18 순서를 §1 첫 gate 순으로, A4 claim 프로토콜, A5 통합 게이트, A6 known-good 복구, A7 단계 8개, A8 RC1 Rust 근거 소절 + tip SHA 고정; B1 GitLab "RC1 미인증", B2 OS 지원표, B5 §14 비범위 배너, B8 "RC1 gate 상태" 링크(#2014) | 원칙 전부가 RC1 감사에서 실제로 걸러낸 결함(#2099·#2104)과 같은 문장 | `review-page_5490c4ee-2026-11-08.md` A·B·C |
| `43fa` RC1 온라인 우선 최종 계획 | §20 파일 배치 실물화(R1·R2), tag 후보 교체(R5), §18 명령 47개로(R4), 홈 용어 4종 정의(I1), ROUTING·M3·M4 측정 가능하게(I2·I3), §7·§19·§23·§28·§4 비범위 배너, §27 수치 재현 명령 | Mermaid 원문 포함(도식 미확인 한계 없음), 첫 RC gate·Desktop 비선행·승인 분리 방향 일치 | `review-group1-rc1-plan.md` §2·§4·§6 |
| `136c` Core Brain 상세 설계 | TOON L4450 정리(S6), 저장 구조 한 벌 + 두 열, 노드·엣지·상태 어휘 한 표, drift 7종↔상위 7종↔RC1 3종 대응표, 시험 ID 통일, bootstrap과 `setup.md` 차이(commit·Issue 발행·receipt 미구현), 태그 구문 상위 §9와 통일, 중복 구간 L1239-2247 | Go 수치(90일·7일·863/28.8k/7 MB·LEGACY 독립 kind) 전부 확인; Claude/Codex 어댑터 표·3-way 보존 일치 | `review-group2-…md` ① |
| `0d45` SYNC와 Issues 보고 | 효과 상태 어휘 1벌 + `needs_rebase`, 통합 행에 RC1 게이트, outbox·tasks.md·promotions.json 경로 실물(`<project>/.moai-cli/forge/`), `sync.preview/apply`·`artifact.pin` 미구현 표시, L347 "발행 안 함" 갱신, L344-357 삽입 제거 | 상태 기계 = `sync.md:13`, drift 3분류 = `sync.md:20`, PATCH 금지 = `write.rs`, timeout→marker = `outbox.rs`, 승격 절차 = `promote.rs` 1:1 | `review-group2-…md` ② |
| `0259` 협업 실행 소유권과 복구 | S1 claim 프로토콜 삽입·L69 재서술, 상태 14개↔RC1 claim 상태 대응표·전이·active 집합, JSON 필드명 상위 §6과 통일, 인터페이스 `claim.*`/`effect.*`/`audit.reduce`로, worktree 삭제·cancel의 RC1 known gap, 제품 원칙(고정 agent 없음)↔개발 운영(고정 레인) 구분 | journal·`local-process` liveness·intent/confirm·독립 감사·write_scope 모두 RC1과 일치 | `review-group2-…md` ③ |
| `c4f9` 검증 감사와 보안 | preflight(`policy.authorize`)를 "RC1: host 위임, `permission_reference` 형태 검사만"으로, verdict 첫 줄·`moai/audit`·독립성 표 삽입, timeout을 FailureClass와 맞춤, 판정 어휘 1벌, 예외 필드 통일, hook 전제 수락(L204) 측정 불가 표시, 근거에 버전·날짜 | admission 6규칙 = `audit.rs:298,308,481-512` + `tests/audit_admission.rs` 6건 1:1; 비밀·공급망 일치 | `review-group2-…md` ④ |
| `e361` 판단 backend와 이미지 | vendor credit 정책 L147-149에 날짜·URL·상품 또는 미결로 환원, RC1 `decision.record`(authority none, P10 HOLD)를 현행으로, `verify.accept` 어휘 통일, benchmark fixture 위치·ground truth, 이미지 절 "RC1 Issue 없음" 배너, `max_cost_usd: 0` 정의 | Jev 기준선 L78 전부 Go에서 확인; native 기본·backend 선택 틀 = RC1 P10 | `review-group2-…md` ⑤ |
| `83f1` 홈 설치와 프로젝트 생성 | §1 홈 경로 실물화 + 후속(P09) 표시, §6 hook 절 RC1 배너, F20·F22-F25 이름 통일(S8), §7 시험마다 8단계 상태·RC1 대응 시험명, L1101 예시 키를 `profile.rs` 관리 키로, 사본 중복·§8 링크 복원 | symlink 금지·AGENTS 로딩 doctor·3-way UPDATE·known-good·외부 skill 설치 후 전체 읽기·TODO 승격 모두 구현 일치 | `review-group3-…md` §1 |
| `dff0` Host adapter와 hook 이식성 | §9 HA 행마다 RC1 상태 + `support-matrix.md`, hook·MCP 절 배너, "20개"→실제 ID, Codex Desktop↔CLI 구분·Desktop 실측 스크립트 부재 명시, 검증/미검증 수치 표기 구분 | v2.1.277·억제 파일 3종·32 KiB·65.6KB/22.5KB·MoAI-Easy 기본 전부 실물 확인 | `review-group3-…md` §2 |
| `25fa` 릴리스와 Desktop | §5·§10 "결정됨/구현 상태" 두 열(S2), L114 명령 실물화, crate 8개→단일 crate(RC1)+분리 후보, "P17" 정의 또는 삭제, §9 비범위 배너, RL-* 4개 시험에 RC1 대응·not_run | Go updater feed 조회 서술 = 특성화 시험 `go_updater_feed_contamination_table`; M1 양 host 필수 = `release blocked` gap #5; 독립 감사·SBOM HOLD 일치 | `review-group3-…md` §3 |
| `e087` PLAN과 GEARS SPEC 승인 계약 | §2 "snapshot 저장 안 함, 요청마다 재계산(`plan.rs:6`)" + PLAN-08 수정, 상태 10개↔`ready|stale|blocked`+사유 코드, digest "정렬 직렬화 임시·JCS 후속(P06)", §7 명령 `plan-admit`·`snapshot-describe`, GEARS where/while 복수·절 순서 | 4축 검사·GEARS 절 분해·EARS import·approval이 host 권한을 주지 않음 — 구현 일치 | `review-group3-…md` §4 |
| `3d91` RUN과 TDD DDD | §3에 RC1 소유권 실물(journal+lock, 댓글 id fencing, takeover 금지), L165 상태를 `RunOutcome`과 claim 상태로 분리, §8 `run-verify`·remote/CI 위임 "RC1 미구현", budget에 deadline 없음 반영, 외부 링크 7건 문장별 각주 | 전략 6종·`method` 필드·evidence 4 digest·producer·stale·RED 환경 분리 — `src/run.rs`와 1:1(7편 중 가장 정확) | `review-group3-…md` §5 |
| `487f` 파일 소유권과 worktree 수명 | L156-174 "RC1 실물 / 후속 목표" 분리, L359 조건부 링크 삭제(상위 §4 L462 인용), worktree 5단계→"RC1: add --lock·keep·삭제 금지 / 후속: retire·cleanup", ST-06에 RC1 known gap(receipt 소실), `~/moai/`→`~/.moai/` 정정 또는 "새 제안", `config.toml`↔`project.yaml` 역할표 | identity `--git-common-dir`·manifests tracked/receipt ignored·3-way·bare/separate-git-dir·Go 파일 3종 존재 확인 | `review-group3-…md` §6 |
| `66ac` RC1 GitHub 작업 구조와 실행 계획 | §8 원자성 경계를 RC1 ⑤+① 으로(S1), 머리에 "발행 완료(2026-11-08), 상태는 Issue 원장" 배너, §9 파일명·workflow 실물(`moai-*.yml`·`.github/moai/labels.yml`·`rc-label-sync.yml`), fixture 4디렉터리 삭제, L1256에 `integrated` 표시 + ruleset, "190개" 기준, L1663 미결→decisions HOLD 목록 | 32개 키로 실제 발행·완료된 유일한 "실행된" 문서; 라벨 26·claim 상태 어휘·Refs/Closes·takeover 금지 일치 | `review-group3-…md` §7 |
| `85f9` Factory 생산성 흐름 재설계 | 상위 §20 비계승 열을 "로컬 선통합 release-branch + 단일 integration writer(문서1:161) — RC1은 Issue별 PR+게이트로 대체"로, 계승 "보고만으로 CI 재시작 않기"를 §15 stale 행에 실제 문장으로; 검사 캐시 키에 dependency lock(#2029 교훈) | ExitWorktree(keep)→main→/clear 종료 순서·두 완료 구분·멱등 키는 RC1이 재채택 | `review-group4-…md` §1 |
| `08f6` Factory·TODO 제어 루프 재설계 | "고정 반복 간격"을 문서 1 행으로, 문서 2 행은 Go SQLite·card_transition 수리 범위로 한정; 계승에 "CI trigger 도달성 사전 확인(CI_TRIGGER_UNREACHABLE)"·"delivered ≠ done" 추가; 개인 절대 경로 노출(문서2:221) 제거 | 4구분(단계/권한·용량/차단 이유/증거 준비도)·조회 무변경·intent→confirmed|UNKNOWN은 상위·RC1에 정확히 반영 | `review-group4-…md` §2 |
| `0914` 3.2 독립 실행 코어 참고 설계 v5 | 비계승 열 "Go 유지 결론 + AF→rc39→3.2 — 단 '측정 뒤 언어 결정' 방법은 §16에 계승", "Claude 우선 Codex 후속"은 문서 4로 이동; 분리 확대 중단 조건 6개·설치 상태 어휘·drain·성능 4분할을 상위에 | 충돌 시 diff 반환·같은 해시 no-op·release set pin·known-good rollback·온라인 원장 비권위 journal은 RC1 재채택 | `review-group4-…md` §3 |
| `5a44` Plan Run Sync와 순정 기능 조사 | 비계승 열 "정적 조사를 실기 결함으로" 삭제(문서가 스스로 막음), "AF rc.39→Claude 중심→Codex 보류" 순서로 교체; 계승에 "판정 6종 분리표(문서4:390)"; Go 줄 번호에 commit 고정 | not_run→passed 승격 금지 기계 검사·정렬 digest·base/head 불일치 거부·argv 검증 명령은 RC1 재채택 | `review-group4-…md` §4 |
| `bb3c` 코어·제공자·브라우저 공식 계약 조사 | 계승 열을 "host별 계약 차이표 + documented/runtime_verified/unverified/unsupported 4분류 + browser capability 행렬"로, cancel·배포 SHA는 문서 3으로; 비계승 열 "Codex alpha 문법·2026-11-01 URL 재사용 않기"; hooks fail-open·설치됨≠신뢰됨≠로그인됨·bootstrap에 trust 우회 금지를 상위에 | Go 관측 permalink 고정(6편 중 최고); `.agents/skills`·`$moai` 변환·hooks 0·documented/not_run 분리 재채택 | `review-group4-…md` §5 |
| `0078` 팀 개발 워크플로와 전체 실행 로직 | 계승 열 "Issues 정본·FSM E01-E27·위임 상태·recover-release"로 구체화(Pele는 문서 1로); 비계승 "GLM 제외·@codex 조건부·담당 agent가 PR 후 종결까지 — RC1은 레인 이동+리더 통합 큐(decisions:77)"; config_revision·canonical block·링크 장부·own-only 작성자·trusted audit bundle을 상위에; 날짜·기준 commit 추가 | agent-N 명명·Issues 정본·SQLite 없음·PR≠DONE·force 해제 불채택·writer/auditor 분리·durable intent 순서는 RC1 재채택 | `review-group4-…md` §6 |

## 4. 교차 규칙 — 20편에 한꺼번에 적용할 편집 원칙

1. **정의는 한 곳, 나머지는 참조.** 경로(상위 §13), 단계 8개(상위 §2), 상태 어휘(부록 1표), 인터페이스·명령(부록 1표), 수용 시험 ID(상위 §17 15군 기준). 하위 문서가 같은 표를 다시 그리지 않는다.
2. **"RC1 출하 / 후속 목표" 두 열.** 경로·crate 구조·worktree 수명·홈 트리처럼 RC1이 다른 값으로 출하한 항목은 실물을 첫 열에, 설계 목표를 둘째 열에 둔다. 한 열만 있으면 다음 구현자가 둘 중 하나를 새로 만든다.
3. **"RC1 범위 밖" 배너.** Core Brain·Code Atlas·온톨로지·보존/아카이브·Pi/Desktop·MCP·hook·이미지·판단 backend·GitLab·webhook·remote 위임 절 머리에 한 줄. 선행조건으로 오해되지 않게.
4. **결정 기록 역참조.** 문서가 "미결"·"후보"·"검토 중"이라 적은 자리마다 `docs/decisions-2026-11-08.md` 행 또는 HOLD(P04-P15) 번호를 붙인다. 결정과 반대인 후보 값은 지운다.
5. **수치 등급 표기.** 검증된 수치(출처·명령 있음)와 미검증 수치를 시각적으로 구분하고, 외부 사실에는 확인 날짜·URL·적용 상품을 붙인다. 재현 명령 없는 집계는 삭제하거나 `baseline-rc1.md` 값으로 교체.
6. **시험마다 RC1 상태 한 줄.** 각 수용 시험 행에 passed receipt / not_run / not_applicable / 대응 시험명(`tests/*.rs`) — `docs/support-matrix.md`·`e2e-receipts/receipts.jsonl`을 가리킨다.
7. **사본 결함 정리.** 중복 구간([83f] L1258-2575, Core Brain L1239-2247, [43fa] 1-1752), "페이지 불러오는 중…" 링크 자리표시, 다른 Page 요약의 본문 삽입([0d45] L344-357), 세 문서에 붙여 넣은 같은 문단(X11)을 원본 Page에서 걷어낸다.
8. **RC1 Rust 근거 소절.** 상위 §22에 `src/*.rs`·`tests/*_cli.rs`·`skills/moai/workflows/*.md` 인용과 tip SHA를 고정한다. 현재는 Go 4fa2aa9만 인용.

## 5. 상위 §1-§18 본문에 빠진 교훈 10건 (과거 설계 두 편 이상이 공유)

그룹 4 §7.2 그대로. 다섯은 RC1이 운영 중 다시 발견해 `AGENTS.local.md`·`pr-contract.md`에 적어 넣었다 — 상위 본문에 있었다면 재발명이 아니었다.

1. 보고서·진행 기록 커밋이 head를 바꿔 검증을 stale로 만드는 순환 — 증거는 freeze 전에 포함하거나 HEAD 밖에(RC1: update-branch 뒤 delta 감사 규칙으로 체득).
2. 종료 순서 ExitWorktree(keep) → main 확인 → /clear → 다음 카드(RC1 decisions:90).
3. PR 이후 책임 귀속 — 레인 이동+리더 통합(RC1 선택, decisions:77) vs 담당자 종결까지. 상위 §6:792가 미결정.
4. 리더 제어 루프는 긴 작업을 안에서 기다리지 않고 대기 이유·다음 trigger를 기록(RC1 lane-check routine 재발명).
5. 필수 검사·CI trigger는 보호 설정과 현 head에서 읽고, 없으면 CI_TRIGGER_UNREACHABLE.
6. hooks·skill은 보안 경계가 아니며 host hook은 fail-open일 수 있다(RC1 P04=A와 정합).
7. 성능·생산성 수치는 baseline 뒤에만; 워크플로 처리량·토큰·재개 성공률 지표가 상위 §16에 없다.
8. 자동 설치·quarantine 제거·trust 우회를 bootstrap 기본값에 넣지 않는다(RC1 skip-permissions 왕복 decisions:47→80이 이 공백의 비용).
9. 개인 절대 경로·비밀을 증거 작성 시점에 제외(RC1 `src/spec.rs:126`).
10. 외부 SHA·버전·시각을 한 표로 고정.

## 6. 설계 ↔ RC1 결정 사이의 진짜 긴장 (문서 결함이 아닌 것)

대조표 §3 그대로. 설계 위반으로 단정하지 않고 "문서에 그 선택을 기록할 자리가 필요하다"로 읽는다.

| # | 설계 | RC1 결정 | 성격 |
|---|---|---|---|
| T1 | 승인 분리(commit·push·Issue·PR·merge·release 각각) | merge는 게이트(ruleset+`moai/audit`) 수준 승인 뒤 리더 auto-merge | 긴장 — 절차상 분리는 남고 merge의 사용자 승인이 게이트 수준으로 올라감 |
| T2 | journal·outbox·receipt는 홈 쪽 | 모두 프로젝트/clone 쪽, Git 제외(`src/apply.rs:29` P09 swap point) | 위치 충돌, 의도 일치 |
| T3 | label·ETag로 전역 CAS 불가 → 단일 owner journal | 댓글 id 단조성에 기댄 eventual claim + 자동 takeover 금지 | 가정에 기댄 선택(S1) |
| T4 | Go 격리·자동 이동 금지 | 04-D RC1 밖 이동으로 해소(decisions:91 ②) | 해소됨 |
| T5 | "G 코드" 특정 제품 미매핑 | 사용자 지정 ZCode 지원표 | 보조 관찰(사용자 지정) |

## 7. 첫 RC gate 상태와 문서의 관계

설계 §1 "Claude·Codex SETUP·UPDATE + service-free Issues end-to-end 실증, 완주를 미루지 않는다"의 현재 상태(대조표 §5): T1 결정적 fixture 5건 passed, **T2 실제 host 로딩 3건 not_run(#2014, 운영자 1회 실측 대기)**, 제품 경로의 실제 GitHub 쓰기 not_run(dogfooding 기록과 공식 receipt를 합치지 않음). `release blocked` 유지. 문서에는 이 gate 상태 링크(B8·R10)가 없어 독자가 "미충족이지만 조용히 미룬 것은 아님"을 알 수 없다.

## 8. 유지할 점 (일곱 검토 공통)

- 원칙 문장 — "검색 결과·모델 요약은 원문 권위를 대체하지 않는다", "not_run을 PASS로 승격 0", "exit 0·ACK·파일 존재 ≠ 완료", "epoch 기록 ≠ fencing", "결과 불명이면 재조회 우선" — 다섯 상세 문서가 같은 문장 수준으로 공유하고 RC1 outbox·handoff·audit가 그대로 구현했다.
- RC1 코드와 1:1인 절: [3d9] 전략표·evidence·method, [e08] 4축 검사·EARS import, [c4f9] admission 6규칙, [0d45] 상태 기계·PATCH 금지·timeout→marker·승격 절차, [dff] 로딩 조건·바이트 수치, [66a] claim 상태·라벨 26. 이들은 "RC1 상태 한 줄"만 보태면 된다.
- 호스트별 독립 복사·symlink 미계승·Go 격리·Issues 정본·새 서버 없음 — 설계와 결정이 같은 방향.

## 9. 미확인·한계 (종합)

- 20편 모두 텍스트 사본이라 Mermaid/Pele 도식과 하이퍼링크 URL을 보지 못했다([43fa]만 Mermaid 원문 포함). 외부 공식 문서(Claude·Codex·GitHub·GEARS·Electron·vendor 정책)는 어느 검토도 재조회하지 않았다.
- RC1 소스는 모듈 doc 주석·상수·operation 표·시험 함수명·grep으로 확인했다. `cargo test`·바이너리 실행·Issue 라이브 본문 전수 조회는 하지 않았다. 검색 0건은 가설이며 긍정 근거(`rust-product-path.md:238·360`, `decision.rs:6`, `main.rs` 표 전수, `verify.rs:52`, `apply.rs:25`)가 있는 곳만 "없음"이다.
- 검토 기준 SHA가 네 개(8ac440587·9bfd26d3b·d253f4a81, 현재 654a1f9bb). 그 사이 변경은 `src/claim.rs`(+45)·`src/handoff.rs`(+206/−47)·`src/worktree.rs`(+3)·문서 1행 — 해당 파일 행 번호만 밀렸을 수 있다. 본 종합은 재검증하지 않았다.
- Go 기준선 수치(hook 82, agents 22, `.claude/` 600, labels 26, 65,627·22,505 bytes, Codex hook 이벤트 12)는 단순 집계이며 문서가 쓴 집계 기준과 같다는 보장은 없다.
- 검토 파일 7개 중 그룹 1·2·3·4와 대조표는 서브에이전트 산출물이고 리더는 편집하지 않았다. 리더가 직접 쓴 것은 상위 검토와 이 종합이다.

## 10. 후속 운영 제안 — 보고서 확정 뒤 MoAI-Factory 모드로 처리 (사용자 지시 2026-11-09 00:xx 반영)

사용자 지시: 이 보고서를 아티팩트로 확인·컨펌하면, 리더가 GitHub Issue를 신규/수정으로 만들어 각 에이전트에게 배정하고 진행 상태·배정 기록을 관리한다. 이를 MoAI-Factory 모드의 기능으로 넣는다. 론처는 `moai cc/glm -l --leader`, `moai cc/glm/codex -a --agent`로 Claude Code·Codex CLI를 띄운다. 에이전트는 보고서를 `.md`로 읽는다.

### 10.1 이미 있는 것 (tip 654a1f9bb 확인)

- 론처: `moai-cli cc|glm|codex -l|--leader` / `-a|--agent` 가 `src/main.rs` `HostLaunch`(00-F #2028)로 구현돼 있다. 공통 옵션 `--name`(기본 `leader` / `agent-1`)·`--model`·`--effort`·`--no-advisor`·`--permissions`·`--dry-run`, codex는 `-s workspace-write -a on-request [-m <model>] [PROMPT]`로 뜬다(`src/session.rs:13`). 역할 없이 `cc|glm|codex`만 치면 평범한 host 세션이고 역할을 추측하지 않는다(#2053). 최상위 `moai-cli -l|-a [HOST]`도 있다. **사용자가 적은 형태 그대로이므로 론처 신규 작업은 없다.** 다만 `docs/user-entry-draft.md`·README에 이 형태가 "현행"으로 적혀 있는지는 확인하지 않았다(미확인).
- Codex 세션은 RC1에서 cross-session 채널(ListAgents/SendMessage)이 없다(`src/session.rs:199` note). Codex 에이전트에게는 Issue 댓글과 `.md` 파일이 유일한 전달 수단이다 — 사용자가 "agent들은 md로 읽고 처리"라 한 것과 맞는다.
- 리더 ↔ 에이전트 운영 규칙: 카드 배차·claim 댓글·epoch 인계·레인 상시 점검·"레인 대기 금지"·ExitWorktree(keep)는 `AGENTS.local.md` § Lane loop와 decisions(33·77·87·90)에 이미 있다. 배정 기록은 `.moai-cli/manifests/github-issues.json`(dispatch_journal·user_decisions·session_notes)과 Issue 소유 댓글에 적는 중이다.

### 10.2 새로 넣을 것 (제안, 사용자 컨펌 뒤 Issue로)

| # | 기능 | 내용 | 비고 |
|---|---|---|---|
| F-A | **보고서 → Issue 변환(신규/수정)** | 확정된 개선 보고서의 §2 S1-S12·§3 문서별 행을 리더가 Issue로 발행하거나 기존 Issue 본문을 수정한다. 한 Issue = 한 세션 원칙 유지. Issue 본문은 `.moai-cli/reports/issue-bodies/<n>-<key>.md` 거울 + "사용자 결정" 절. 설계 문서 수정 카드는 코드가 아니라 Space 페이지 편집이므로 수용 기준을 "해당 행이 고쳐진 사본 sha256 + 리더 읽기 확인"으로 둔다 | 신규 Issue 발행은 사용자 컨펌 뒤에만(승인 분리). 현재 RC1 milestone은 닫는 중이므로 별도 milestone(예: `design-docs-r2`) 제안 |
| F-B | **배정 기록** | 각 Issue에 `assigned_to: agent-N`, `claim: <issue>#<comment-id>`, `epoch`, `dispatched_at`, `stage`를 `github-issues.json` dispatch_journal에 기록(이미 쓰는 구조 확장). Issue 소유 댓글 `상태:` 줄과 동기 | 기록 파일은 로컬·Git 포함(`.moai-cli/manifests/` tracked 구간) |
| F-C | **진행 상태 관리** | 리더 루프가 `gh issue list --milestone` + 댓글 `상태:` 줄 + PR 상태를 읽어 Progress Board로 렌더하고, 유휴 레인에 다음 카드를 배차. 완료는 PR 병합·verdict 파일 읽기로만 판정(메시지·idle 알림 아님) | `AGENTS.local.md` lane-check routine을 Factory 모드 문서(`.claude/rules/moai/workflow/factory-dispatch*.md`)에 반영 |
| F-D | **에이전트 입력 = md** | 배차 메시지는 fixed-field block(card/spec/cmd/wt/evidence) + 읽을 `.md` 경로(이 보고서, 해당 검토 파일, Issue 본문 거울)만 전달. Codex 레인은 메시지 채널이 없으므로 첫 프롬프트(PROMPT 인자)에 경로를 넣는다 | `references/chatgpt-space/`는 현재 untracked — 에이전트가 읽으려면 tracked 위치(예: `docs/design-review/`)로 옮기는 카드가 먼저 필요 |
| F-E | **Factory 모드 문서화** | 위 F-A~F-D를 `factory-dispatch.md` § 카드 소스에 "확정 보고서로부터의 카드 생성"으로 추가하고, 론처 형태(`moai cc/glm -l`, `moai cc/glm/codex -a`)를 사용자 문서(`docs/user-entry-draft.md`)에 현행으로 적는다 | 문서 카드(Class A/B) |

### 10.3 처리 순서 (시간 추정 없음)

1. 사용자가 아티팩트로 이 보고서를 확인·컨펌(수정 지시 포함).
2. 리더가 F-D 선행 카드(보고서·검토 파일을 tracked 위치로 이동)를 먼저 Issue로 발행·배차.
3. 리더가 §2 S1-S12를 우선 Issue로, §3 문서별 행을 후속 Issue로 발행(신규) 또는 기존 Issue 본문 수정(#2014 등 관련 카드는 수정).
4. 유휴 레인에 claim 프로토콜대로 배차, `github-issues.json`에 배정 기록, Issue 댓글 `상태:` 유지.
5. 완료는 리더가 증거(Space 사본 sha256·PR·verdict)를 읽어 판정, Progress Board 보고.

### 10.4 자율 무중단 개발 모드 — `/moai --leader-mode` · `/moai --agent-mode` (사용자 지시 2026-11-09 00:0x, 설계 추가)

사용자 지시: `-l`/`-a` 없이 `moai cc|glm|codex`로 접속하고(론처는 `--advisor` 등 model client별 편의 설정만), 세션 안에서 `/moai --leader-mode` 또는 `/moai --agent-mode` 스킬을 입력하면 리더·에이전트 모드로 **자율 무중단 개발**에 들어간다. 이를 설계에 추가한다.

| 요소 | 설계 내용 | RC1 실물과의 관계 |
|---|---|---|
| 진입 | 론처는 역할을 모른다(`moai cc|glm|codex` = 평범한 host 세션, #2053과 일치). 역할은 세션 안에서 skill 인자로 켠다: `/moai --leader-mode [--autonomy full|semi]`, `/moai --agent-mode`. Codex 세션은 `$moai` skill 호출로 같은 인자를 받는다 | 현재 `-l`/`-a` 론처 플래그는 세션 이름·prompt만 다르고 역할 행동은 `AGENTS.local.md`·SessionStart 주입에 의존. 모드 skill은 그 주입을 명시적 진입으로 바꾸는 것. `-l`/`-a`는 호환용으로 남기거나 폐기 여부를 결정 항목으로 둔다 |
| 리더 의사결정 등급 | **완전자율(full)**: 결정 지점마다 먼저 "Jev 또는 자체 판단으로 결정 가능한가"를 확인하고 가능하면 리더가 결정·기록. 불가능(환경·운영자 보유·되돌릴 수 없는 외부 공유 효과)만 사람에게. **반자율(semi)**: 중요한 결정(Issue 신규 발행, HOLD 해제, 병합 정책 변경, release)은 사람 허가, 나머지는 리더 자율 | RC1 `auto-semantics.md` §6 결정 ladder·§9 gate 목록과 keep-set(환경 불가·운영자 보유·되돌릴 수 없는 외부 공유)이 그대로 쓰인다. Jev는 P10 HOLD·`decision.record`(authority none)이므로 "Jev 사용 여부 확인"은 Jev가 켜져 있고 질문형이 맞을 때만, 아니면 자체 판단 — 어느 쪽이든 `decisions/<run>.jsonl`에 기록 |
| 에이전트 루프 — **self-claim 기본** (사용자 지시 00:3x) | 에이전트는 리더 배차를 기다리지 않고 스스로 Issue를 집는다. 순서: ① 로컬 배정 파일(`.moai-cli/manifests/github-issues.json` dispatch_journal 또는 그 옆의 `assignments.json`)의 **미배차 Issue 번호 목록을 원격에서 다시 읽어 갱신**(읽기 전용 스냅샷) → ② 후보 1건 선택(우선순위·ready 조건) → ③ 같은 호스트 직렬화: `<git common dir>/.moai-cli/claims/` journal append + advisory lock → ④ Issue에 `moai-claim: session= epoch= lease_until=` 댓글 게시 → ⑤ 댓글 목록 재조회, **가장 낮은 live 댓글 id가 승자**(진 쪽은 로컬 기록을 `lost`로 닫고 다음 후보로) → ⑥ 승자가 label(`moai:in-progress` 등)·소유 댓글 `상태:` 줄을 갱신 — 이것은 **표시(projection)**, 소유권 근거가 아님 → ⑦ `EnterWorktree` → 구현·검사 → PR → 영속 보고 → `ExitWorktree(keep)` → **병합을 기다리지 않고** ①로 돌아가 반복. 목록이 비면 `idle: backlog empty <time>` 1회 기록 후 대기(폴링 증식 금지) | 동시성 판단: **맞다** — 단, 중재 수단은 로컬 파일이나 label이 아니라 **댓글 id 단조성**이어야 한다. 로컬 파일은 다른 PC의 세션이 볼 수 없어 lock이 될 수 없고(같은 호스트 직렬화·캐시 용도만), GitHub label 변경은 원자 CAS가 아니라 두 세션이 동시에 붙일 수 있다(설계 §6·RC1 모두 "label은 lock 아님"). 이 순서는 RC1 06-A `src/claim.rs:1-23`(① journal+lock, ⑤ 댓글 claim primary, 최소 id 승, 만료 자동 takeover 금지)이 이미 구현한 그대로이며, 새로 넣는 것은 "리더 배차 없이 에이전트가 ①~⑥을 직접 돈다"는 진입 경로와 모드 skill이다. `AGENTS.local.md` § Lane loop·decisions:77·90과 일치. 리더의 역할은 배차가 아니라 감독(5분 PR 점검·수리 재배차·배정 기록 대조·목록에 없는 Issue 보충) |
| 리더 PR 점검 루프 | 리더는 **5분마다** 열린 PR을 읽어(상태·CI·감사 verdict·충돌) 수리가 필요하면 수리 Issue/댓글을 만들어 **다른 에이전트**에게 수리 지시(원 작성자가 아닌 레인; 작성자≠감사자≠통합자 유지). PASS면 게이트(status `moai/audit` → auto-merge) 진행 | 리더 lane-check routine(decisions:87)·cron 5분 틱과 같다. 5분은 사용자 지정 주기이며 측정값이 아님을 문서에 적는다(과거 설계 "5분 루프" 수치 근거 없음 지적과 정합). `ScheduleWakeup`/cron 중 하나로 구현, 변화 없는 PR마다 폴링 증식 금지 |
| 리더 감독(배차 아님) | 리더는 Issue를 직접 배차하지 않는다. 대신 ⓐ 미배차 목록에 새 Issue를 보충(보고서→Issue 변환, 수리 Issue 발행), ⓑ 에이전트들의 claim 댓글·로컬 배정 파일을 대조해 배정 기록(F-B·F-C)을 갱신, ⓒ 대기 중인 에이전트가 있는데 목록이 비어 있으면 사유 통지, ⓓ 완료 판정은 증거(PR 병합·verdict 파일) 읽기로만 | 기존 규칙에서 "리더가 배차"만 "에이전트 self-claim + 리더 감독"으로 바뀜. 리더가 특정 Issue를 특정 에이전트에게 주고 싶을 때(예: 수리는 원 작성자 아닌 레인)는 Issue 본문/댓글에 `assignee-hint: agent-N`을 적고, 그래도 소유권은 그 에이전트의 claim 댓글이 성립해야 생긴다 |
| 리더의 `/goal` 안내 (사용자 지시 00:1x) | 리더 모드는 켜지면 먼저 열린 Issue·PR을 분석해 **완료 조건 문장**을 만들고(예: "milestone X의 Issue·PR 모두 병합·종료"), 사용자가 그대로 붙여 넣을 `/goal <목표>` 지시문을 안내한다. 사용자가 붙여 넣으면 Stop hook이 그 조건이 충족될 때까지 턴 종료를 막아 무중단 진행이 된다. 에이전트 모드도 같은 방식으로 자기 카드 범위의 goal 문장을 안내한다 | **자동 입력 가능 여부 확인 결과**: ① Claude Code 내장 `/goal`은 CLI 내장 명령이라 모델이 대신 입력할 수 없다(내장 명령은 skill이 아님 — 이 세션에서 사용자가 직접 `/goal`을 설정했고 리더는 해제도 못 했음). 따라서 **붙여넣기 지시문**으로 제공한다. ② MoAI 자체 `moai goal`(MCP `goal_arm`)은 리더가 프로그램으로 armed 가능(`goal-directive.md` "Arming is programmatic")하나 Go 세대 도구이고 RC1 Rust `routes.json`에는 goal route가 없다 — RC1에서는 ①의 지침 방식이 유일. ③ Codex CLI에 동등한 `/goal` 명령이 있는지는 **미확인** — 확인 전까지 Codex 에이전트에는 goal 문장을 첫 프롬프트(PROMPT 인자)·Issue 댓글의 지침으로 넘긴다. 후속 카드: Codex 문서 확인 뒤 지원표에 goal 행 추가 |
| 안전 경계(불변) | 두 모드 모두 ruleset·`moai/audit` 독립 감사·비밀 미노출·hook/권한 우회 금지·Go 상태 보존·release/deploy 금지(04-D 밖)를 그대로 지킨다. 완전자율도 keep-set 결정은 사람에게 | 사용자 결정 22:0x "HOLD만 해제" 범위와 동일 |

후속 카드 후보: (a) `/moai` skill에 `--leader-mode`·`--agent-mode`·`--autonomy` 인자 추가 + 모드별 진입 절차 문서(`skills/moai/workflows/`), (b) 리더 5분 PR 점검·수리 재배차 절차를 `factory-dispatch.md`에, (c) 에이전트 자율 수령(큐 lease) 경로와 `AGENTS.local.md` lane loop 통합, (d) 결정 등급(full/semi)과 Jev 사용 여부 확인 절차를 `auto-semantics.md` ladder에 연결, (e) 론처 `-l`/`-a` 존치 여부 결정.
