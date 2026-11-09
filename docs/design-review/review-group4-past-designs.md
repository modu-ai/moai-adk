# 과거 설계 6건 검토 — 상위 §20 "이전 설계에서 남기는 교훈" 표 대조

- 대상: 상위 설계 `page_5490c4ee8c8481918b3fc3217401a3f4.md`(본문 sha256 6ee32452…e82b) §20 표(2244–2341행)와 그 표가 가리키는 과거 문서 6건(아래 각 절 머리에 사본 sha256 표기). 여섯 문서 모두 텍스트 사본 전문을 읽었다.
- 대조 기준: `docs/decisions-2026-11-08.md`, `origin/v2026.11.0` 로컬 remote-tracking tip `d253f4a81f70174d366a0ff72a525e2c22856daa`(2026-11-09T08:25:31+09:00; fetch 미실행, 앞선 상위 검토가 쓴 8ac440587에서 이동함). 파일 줄 번호는 모두 이 tip의 `git show` 결과다. Go 저장소 로컬 체크아웃(HEAD 4fa2aa9db2)에서는 과거 문서가 인용한 SHA의 존재만 `git cat-file -e`로 확인했다.
- 작성 시각 2026-11-08T23:31Z(`date -u` 측정). 읽기 전용 검토이며 쓰기는 이 파일 하나다.
- 인용 표기: `문서N:줄` 은 해당 과거 문서 사본의 줄, `상위:줄` 은 상위 설계 사본의 줄, 그 밖은 저장소 경로:줄. 예외: `AGENTS.md`는 `origin/v2026.11.0`에 추적되지 않아 로컬 작업 트리 사본(시스템 컨텍스트로 제공된 본문)을 인용했다.

## 요약

상위 §20 표는 여섯 문서를 한 줄씩 "남길 것 / 계승하지 않을 것"으로 압축했는데, 그 압축 자체에 네 종류의 오류가 있다. ① 비계승 열이 문서 1·4·5에서는 문서가 본문에서 명시적으로 막아 둔 결함을 그 문서의 주장인 양 적었다. ② 계승·비계승 항목의 귀속이 문서 사이에 엇갈린다(문서 5의 "cancel·배포 SHA"는 문서 3, 문서 6의 "Pele"는 문서 1, 문서 2의 "고정 반복 간격"은 문서 1, 문서 3의 "Claude 우선 Codex 후속"은 문서 4의 것). ③ 비계승으로 적은 조건 가운데 lane·lease·epoch·agent-N 명명·Issue별 PR은 RC1이 그대로 다시 썼고, 남겼다고 적은 "보고만으로 CI 재시작하지 않기"는 상위 §1–§18 본문 어디에도 없다. ④ 두 문서 이상이 공유하는 교훈 열 가지가 상위 본문에 빠져 있으며, 그중 다섯은 RC1이 운영 중에 다시 발견해 `AGENTS.local.md`·`pr-contract.md`로 적어 넣었다.

---

## 1. MoAI Factory 생산성 중심 전체 흐름 재설계 (`page_85f9b625…`, sha256 a8cc9a83…)

### 1.1 문서 개요
- 상태 기준 2026-11-02 18:37 KST(문서1:10), 11-07 운영 방향(AF-Flow) 하위 페이지로 보강(문서1:5–6). 비공개 설계 보고서이며 구현·설치 완료가 아님을 명시.
- 기준 commit: 공개 main 59464c5f(문서1:47), PR 1960 head 37be323f, PR 1955 head a81a3187(문서1:86·99) — 세 SHA 모두 Go 저장소에 존재 확인.
- 핵심: AF(배차·실행)와 AFF(통합·릴리스) 분리, worktree별 PR을 기본값으로 두지 않고 로컬 release-branch 후보에 모아 release PR 하나로 main 병합, release SHA 고정, 대기에는 이유와 해제 사건, 리더 5분 행동 루프, RC39 수직 검증 8단계, 생산성 지표 7종.
- 도식 6개(Pele 1.13.0 SVG)와 Mermaid 원문은 사본에 없다.

### 1.2 상위 §20 표와의 대조
- 계승 "독립 작업·통합 writer·고정 candidate, 보고만으로 CI 재시작하지 않기"(상위:2266): 문서1:8·161·1057에 실제로 있다. 그러나 "진행 기록 commit으로 HEAD를 움직이지 않는다"는 문장이 상위 §1–§18 본문에는 없다(`진행 기록|detached|HEAD 밖|보고서 커밋` 0건). 남겼다는 교훈이 표에만 있고 본문에 없다.
- 비계승 "작업별 PR 또는 release 단위 PR을 모든 프로젝트에 강제"(상위:2270): 문서1:18은 정반대다 — worktree별 PR을 기본값으로 삼지 않되 외부 협업·독립 릴리스·보호 정책이 필요한 변경은 기존 PR을 유지할 수 있다고 적었고, "모든 프로젝트에 강제"라는 문장은 없다(문서1:10은 ADK 내부 보고서임을 명시). 비계승 행이 문서가 하지 않은 주장을 귀속했다.
- 비계승으로 적었지만 상위 본문이 여전히 필요로 하는 것: 상위 §5 통합 행(상위:572–580)은 "최신 base로 exact candidate 재검증 → 허용된 merge/release"만 있고, 문서1:208의 세 완료(카드 local-integrated ≠ release 완료 ≠ 설치 완료)와 writer 종료→leader 인수(문서1:1059) 구분이 없다. RC1은 같은 구분을 `AGENTS.local.md:32`("its handoff is not PR completion")로 다시 적었다.
- 상위가 조용히 빠뜨린 교훈: (a) 검사 캐시 키 = SHA + test scope + toolchain + dependency lock + configuration digest(문서1:1059); 상위 §15 stale 행(상위:1638)에는 dependency lock이 없다 — RC1 #2029 toolchain 불일치(decisions:48)가 바로 이 항목. (b) 필수 검사 목록은 저장소의 실제 보호 설정에서 읽고 문서 이름으로 정하지 않는다(문서1:898) — 상위 0건. (c) 리더 한 회차 점검 8순서와 "읽기 실패는 unmeasured이지 hold도 PASS도 아님"(문서1:830) — 상위 0건, RC1 `AGENTS.local.md:35` lane-check routine이 재발명. (d) 카드 실행 lease와 integration window lease는 다른 자원(문서1:788) — 상위 0건. (e) 생산성 지표 7종과 baseline 선측정(문서1:1334–1422) — 상위 §16–17은 검색·성능 benchmark만 있고 워크플로 처리량 지표가 없다.

### 1.3 내부 결함
- "리더의 5분 행동 루프"(문서1:794)는 제목이 확정 어조인데 본문(문서1:798)은 "운영 제안 주기"라 하고 수치 근거가 없다.
- 문서1:832 "현재 요청된 Go 순차 실행"의 Go가 언어인지 작업 종류인지 불명.
- 단일 integration writer가 누구인지(리더인지 전용 세션인지) 어디에도 지정하지 않았다(문서1:169·683).
- RC38·RC39가 어느 tag·commit인지 매핑이 없다(문서1:112·1077).
- 도식 "6개 모두 PNG 시각검수"(문서1:1429) 주장은 사본에 도식이 없어 검증할 수 없다(결함이 아니라 검증 불가).

### 1.4 RC1 결정·실물과의 관계
- 재채택: ExitWorktree(keep)→main 확인→/clear→다음 카드(문서1:208) → decisions:90, `AGENTS.local.md:33`. 알림·보고는 완료 증거가 아님(문서1:534) → `AGENTS.md` "Completion is read from evidence". lease 수치 미정(문서1:465) → `src/claim.rs:25–26` HOLD. 멱등 키·stable marker·늦은 epoch 거절(문서1:792) → `src/forge/outbox.rs:4–13`, `src/handoff.rs:18–20`.
- 역행(RC1이 문서 1과 반대로 감): Issue별 PR을 ruleset으로 강제하고 리더 문서 커밋까지 PR 경로(decisions:39, `AGENTS.local.md:39`, `docs/pr-contract.md`) — 문서1:18이 기본값에서 뺀 "worktree마다 PR" 모델이며, 로컬 선통합 release branch는 없다. 즉 상위 비계승 행과 RC1 실물이 서로 반대 방향이다.
- 두 완료 구분(문서1:208)은 `AGENTS.local.md:32`가 그대로 쓴다.

### 1.5 상위 §20 행 개선 제안
- 비계승 열을 "로컬 선통합 release-branch 후보 + 단일 integration writer 모델(문서1:161) — RC1은 Issue별 PR + 에이전틱 게이트로 대체"로 바꾸고, 계승 열의 "보고만으로 CI 재시작하지 않기"를 §15 stale 행에 실제 문장으로 넣을 것.

### 1.6 미확인·한계
- 하위 페이지 "AF-Flow 개념과 병합 운영 기준"(문서1:1434) 미읽음. Pele SVG 6개 미열람. 문서1:60이 인용한 `sync/delivery.md` 원문 미대조.

---

## 2. ADK Factory와 TODO 제어 루프 재설계 검토본 (`page_08f6df95…`, sha256 21621b31…)

### 2.1 문서 개요
- 갱신 기준 2026-10-28 06:36 KST(문서2:8), 실측 기준 10-28 03:15:31 KST clean source 4c4217a1(문서2:229, Go 저장소 존재 확인). 10-27 RFC·자문 이력 E0–E6를 시점별로 보존(문서2:903–991).
- 주제: 리더 제어 루프, 카드 임대·배차, 동일 파일 충돌·CI 실행 경로 부재·만료 임대와 전달 단계 혼동. 전면 교체가 아니라 기존 CAS·fencing·receipt 검증에 연결하는 최소 수리를 주장.
- 수용 기준 AC1–AC9(문서2:1583–1599)와 P0/P1/P2 적용 순서(문서2:586–672). Cloud 수리 보고·Codex 자문·Jev 응답은 모두 "독립 검증 아님"으로 표시.

### 2.2 상위 §20 표와의 대조
- 계승 "실행 용량·차단 이유·owner·증거 준비도 구분"(상위:2280): 문서2:449–513의 네 구분(카드 단계와 owner / 실행 권한과 용량 / 차단 이유 / 증거 준비도)과 정확히 맞고, 상위 §6:756(assignee·execution_owner·provider_session·integration_owner)·§15:1611(차단 행)이 이를 받는다.
- 비계승 "과거 SQLite/lane/cron 구조와 고정 반복 간격"(상위:2284): 문서 2는 SQLite를 현행 사실로 관측(문서2:252 readonly opener 644/600)했을 뿐 설계로 제안하지 않았고, cron은 오히려 "변화 없는 카드마다 polling cron을 증식시키지 않는다"(문서2:1127)고 막았다. "고정 반복 간격"은 문서 2에 없고 문서1:794(5분 루프)의 것이다 — 귀속 오류. lane 개념(카드와 lane의 수명 분리, 문서2:1131)은 RC1 `AGENTS.local.md:24`(agent-N 세션 lane)가 재채택했다 — 당시 Go Factory slot과 RC1 세션이라는 실체 차이는 있다.
- 비계승으로 적었지만 상위가 필요로 하는 것: 상위 §6:758 "단일 조정 owner가 쓰는 작은 durable journal"은 문서2:515 "durable inbox 중복 제거·읽기 전용 snapshot 대사"의 축약이다. 조회 무변경(문서2:532)은 상위 §15:1593과 일치.
- 상위가 빠뜨린 교훈: (a) CI_TRIGGER_UNREACHABLE — required context별로 현 head·event·ref·정책 revision에서 실행 trigger가 있는지 먼저 확인(문서2:142); 상위 0건. RC1 `docs/pr-contract.md`("기본 브랜치 템플릿만 적용, Closes는 기본 브랜치 대상에서만 해석")가 같은 교훈을 뒤늦게 체득했다. (b) answered ≠ verdict PASS ≠ policy accepted, archive ≠ verified done, merged ≠ deployed ≠ local sync(문서2:1280) — 상위 §15:1678은 일부만. (c) Stop stale ALLOW와 notify_when_idle의 idle은 완료가 아님(문서2:1077·1316) — 상위 0건. (d) candidate freeze 뒤 증거만 커밋해 head가 바뀌는 순환(문서2:169·559) — 상위 0건. (e) delivered·held·refused를 실행 성공으로 합치지 않음(문서2:1311) — 상위 0건. (f) 감사 재실행 예산을 findings fingerprint·입력 digest·ceiling HOLD에 묶음(문서2:557) — 상위 §15:1629에 있음(일치).

### 2.3 내부 결함
- 사용자 홈 하위의 개인 절대 경로(…/moai-adk-go/.moai/reports/… 감사 원문 경로)를 본문에 노출(문서2:221) — 문서6:1205와 RC1 REQ-05A-08(`src/spec.rs:126`)이 금지하는 패턴.
- graph 변경량 임계 40(문서2:398·411)의 출처가 없고, "65분 preflight timeout"(문서2:1100)은 맥락 없이 등장.
- AC3 본문(문서2:1587) 안에 "Cloud가 … PASS를 보고했습니다"가 섞여 수용 기준과 진행 상태가 한 문단에 혼재.
- 로그 해시 7건(문서2:697–783)은 "실행자가 기록한 값이며 재해시 아님"(문서2:682)이라 재검증 불가 — 정직하게 표시했으므로 결함은 아니나 근거 등급은 낮다.
- 자문 모델 `gpt-6.1-terra / served_model=unknown`(문서2:1641), Fable 미호출(문서2:1624) → "삼자 합의 아님"(문서2:1680)으로 일관.
- 기능 검토 기준 Claude Code v2.1.311(문서2:933)은 RC1 관측 2.1.319(`docs/host-support-matrix.md:60`)와 다르다 — 시점 차이.

### 2.4 RC1 결정·실물과의 관계
- 재채택: lane 구조(agent-1~8, `AGENTS.local.md:24`); 리더의 짧은 reconcile 루프·lane 점검(문서2:797·1106 → `AGENTS.local.md:35`); 카드 owner 보존 + 만료 lease 자동 탈취 금지(문서2:1135 → `src/claim.rs:22`); 조회 무변경(문서2:532 → `docs/rust-product-path.md:17` "writes nothing"); 외부 부작용 intent→dispatched→confirmed|UNKNOWN(문서2:1255 → `src/forge/outbox.rs:11` proposed→publishing→confirmed|publish_unknown); 검사 0·PASS-WITH-DEBT 비승격(문서2:10 → decisions:49 Issue 닫기 기준).
- 비계승 대상인데 재채택: lane·lease·epoch·journal — `src/claim.rs:16` `moai-claim: session=<id> epoch=<n> lease_until=<utc>`, `src/worktree.rs:7` `--lock --reason "moai claim <task_key> e<epoch>"`. SQLite만 미채택(Cargo.toml 의존성 4개, DB 없음). cron 폴링도 없고 `x-poll-interval` 우선(decisions:54) — 문서2:1127과 같은 방향.
- 근거 없음: 문서2:142 CI trigger 사전 확인을 RC1 코드로 구현한 흔적은 grep 범위에서 없음(pr-contract.md는 문서 규칙).

### 2.5 상위 §20 행 개선 제안
- 비계승 열의 "고정 반복 간격"을 문서 1 행으로 옮기고, 문서 2 행은 "당시 Go Factory의 SQLite 운영 DB와 card_transition.go 수리 범위(문서2:1690)"로 한정; 계승 열에 "CI trigger 도달성 사전 확인(문서2:142)"과 "delivered ≠ done(문서2:1311)"을 추가.

### 2.6 미확인·한계
- E0 RFC HTML, E1/E2 md, Cloud 로그 7건 미열람(해시만 기록됨). Mermaid 도식 미확인. t-번호 카드(t1627 등)의 원장은 Go 운영 DB라 읽지 않았다.

---

## 3. MoAI ADK 3.2 독립 실행 코어와 통합 어댑터 참고 설계 버전 5 (`page_0914dfdc…`, sha256 58f91af5…)

### 3.1 문서 개요
- 작성 2026-11-01T05:12:49Z, 근거 HEAD 5c3f0ebf(문서3:9, Go 저장소 존재 확인). 보고서 v5(v2 Go/Rust 비교, v3 browser, v4 native supervisor 재구조화).
- 실행 순서 AF 완료 검증 → rc39 로컬 빌드·복사 → 새 AF로 3.2 개발(문서3:7·95). 3.2 = Go core + 내장 native supervisor + Codex/Claude host package + ego(lite)/Aside browser adapter + Orca/Herdr 선택 backend.
- 항목 00~17, DEV1–3, FUTURE1–2 전부 "상태 unverified"; held 카드 203개(165+38) 개별 검토 부록. 모든 구현·테스트 NOT_RUN 명시.

### 3.2 상위 §20 표와의 대조
- 계승 "adapter 호환·멱등 복구·단일 owner·사용자 환경 보호"(상위:2294): 문서3:275–306(02B), 356–382(04), 529–554(09)·880–899(15), 913–938(16)에 실제로 있고 상위 §3:353(3-way merge)·§4:462(독립 복사)·§6:760(단일 owner)·§15(복구)가 받는다.
- 비계승 "Go 기본·Claude 우선 Codex 후속·과거 릴리스 선후 관계"(상위:2298): "Go 유지가 기본값"(문서3:688)과 "AF→rc39→3.2"(문서3:7)는 맞다. "Claude 우선 Codex 후속"은 문서 3에 없다 — 문서3:7은 Codex/Claude 플러그인을 동시에 두고 문서3:285는 "Codex를 skills-only로 보는 설계는 출발점부터 틀린다"고 한다. 이 조건은 문서4:145·165의 사용자 지시다 — 귀속 오류.
- 비계승으로 적었지만 상위가 필요로 하는 것: 문서3:688–792(항목 11–13)의 전제는 "언어 전환은 pprof baseline·동일 fixture·총비용 비교 뒤"다. 상위 §1:55는 Rust를 선언하지만 이 전제를 만족했다는 근거가 없고, §16:1759는 "실측 전 몇 배 빠름을 약속하지 않는다"까지만 있다. 비계승으로 보낸 것은 "Go 기본" 결론인데 그 결론의 방법(측정 뒤 결정)까지 같이 사라졌다.
- 상위가 빠뜨린 교훈: (a) 분리 확대 중단 조건 6개(문서3:680 — provider 분기로 뒤덮임, trust 우회 필요, rollback이 사용자 수정을 잃음 등) — 상위 §18 남은 결정에 없음. (b) supervisor 상태 어휘와 "exit 0·heartbeat 종료만으로 성공 아님"(문서3:251·261) — 상위 §15:1678 일부. (c) 설치 상태 CORE_READY/REGISTERED/WORKER_READY/NEEDS_TRUST·AUTH·RELOAD/PENDING_HOST/UNSUPPORTED(문서3:330) — 상위 §3:297–298은 "발견·호출·smoketest·ready" 한 줄. (d) 실행 중 run이 있으면 update/uninstall이 조용히 프로세스를 죽이지 않고 drain(문서3:382) — 상위 §15 없음. (e) 성능 비교 4분할 startup/RSS/CPU-GC/프로세스-I/O(문서3:738–741) — 상위 §16:1761 benchmark 항목과 부분 중복, 옮길 가치 있음.

### 3.3 내부 결함
- 제품명 표기 불일치: "Egolite"(문서3:27·55·75) vs "ego(lite)"(문서3:810) — 같은 묶음의 문서5:519가 "AGoLite를 별도 제품으로 만들지 않는다"고 경고하는 바로 그 혼동.
- "참고 설계 v5"인데 미확정 선택 질문이 본문에 남아 있다(문서3:342–348, 플러그인 설치 범위 사용자/프로젝트/팀).
- 지원 OS·host 최소버전이 미정(문서3:49)인데 항목 08 매트릭스는 "각 지원 OS × host 버전"을 전제(문서3:507).
- 출처 23건 중 17건이 "SHA256 미확인·immutable revision 미확인"(문서3:1352–1390) — 동적 문서라 불가피하나 재현성이 낮다.
- held 203개 처분 제안의 근거가 "제목/회수 텍스트"(문서3:157) — 문서 스스로 등급을 낮게 표시.

### 3.4 RC1 결정·실물과의 관계
- 역행: Rust 전면 채택(`Cargo.toml` `moai-cli 2026.11.0-rc.1`, `docs/rc-isolation-decision.md`) — 문서3:712 "실제 전면 재작성은 승인하지도 실행하지도 않았다"를 뒤집음. `docs/rust-product-path.md` grep에서 pprof·benchmark·baseline 0건(추적 문서 1건 grep, 전수 탐색 아님).
- 재채택: 사용자 수정 충돌은 덮어쓰지 않고 diff 반환(문서3:374) → decisions:78 HOLD-B3, `docs/rust-product-path.md:137` "applying a conflict plan writes nothing"; 같은 버전/해시 no-op(문서3:373) → `docs/support-matrix.md` HA-02 "UPDATE 재실행 diff 0"; 선택한 항목만 등록(문서3:336) → decisions:66 `--target` 명시 필수; release set pin(문서3:407–409) → `rust-toolchain.toml` 1.94.1·Cargo `=` 고정; 이전 set rollback(문서3:415) → `src/recovery.rs:10–11` known-good; 온라인 원장 + 로컬 journal 비권위(문서3:640–641) → `AGENTS.md` Authority 2, `docs/pr-contract.md` "record 파일은 권위가 아니다"; effect receipt·idempotency_key(문서3:267) → `src/forge/outbox.rs`.
- 부분 재채택: "Claude 우선 Codex 후속"은 문서 4의 것이지만 RC1 launcher는 Claude만(decisions:26 "Codex 후속"), host 범위는 claude·codex 둘 다(`.moai-cli/project.yaml`).
- 미채택(RC1 범위 밖): native supervisor, Orca/Herdr, browser adapter — decisions:23 "원격·Orca/Hide RC1 제외".

### 3.5 상위 §20 행 개선 제안
- 비계승 열을 "Go 유지 결론과 AF→rc39→3.2 선후(문서3:7·688) — 단 '측정 뒤 언어 결정' 방법(문서3:784)은 §16에 계승"으로 고치고, "Claude 우선 Codex 후속"은 문서 4 행으로 옮길 것.

### 3.6 미확인·한계
- 하위 페이지 11개(문서3:1427–1437) 중 문서 4·5 외 미읽음. 조사 메모 S1~S22 원문·held 165/38 부록 미열람. 도식 없음.

---

## 4. MoAI ADK Project Plan Run Sync와 순정 기능 활용 종합 조사 (`page_5a4472e7…`, sha256 5d3d5b33…)

### 4.1 문서 개요
- 기준 원격 main 5c3f0ebf(문서4:177). 11-01 20:06 KST 사용자 실행 승인(문서4:117), 11-02 대기 병목 소스 조사 추가(문서4:4–22, 설치본 RC38 b1806498fe). 인용 SHA 5건(b1806498fe·2fd9455965·0afd4c4e7e·5c3f0ebf·db0a69698a) 모두 Go 저장소 존재 확인.
- 교차 단계 발견 9건(문서4:314–384), 판정·책임 분리표 6종(문서4:386–479), 순정 기능 대체 판정 "완전 동등 0건, 부분 대체"(문서4:494–629), 단계 간 수용시험 12건 전부 NOT_RUN(문서4:649–673).
- 정적 조사임을 반복 명시(문서4:20·175·187·790).

### 4.2 상위 §20 표와의 대조
- 계승 "producer/consumer schema·digest·상태 의미 parity"(상위:2308): 문서4:109(null vs boolean), 318–336(PASS-WITH-DEBT 유실, hash 알고리즘 불일치, audit_ready 신호 차이)에 있다 — 정확.
- 비계승 "정적 조사 후보를 전부 실기 재현 결함으로 표현"(상위:2312): 문서 4는 정반대로 쓴다 — "코드에 근거한 추론이다. 동적 재현은 아직 없다"(문서4:20), "정적 조사만 수행했다"(문서4:175), "실제 사용자 파일 소실이나 중복 카드 생성은 재현하지 않았다"(문서4:187), "P0/P1은 개선 순서이지 재현된 취약점 등급이 아니다"(문서4:790). 비계승 행이 문서가 스스로 막은 결함을 문서 탓으로 돌렸다.
- 비계승으로 적었지만 상위가 필요로 하는 것: 비계승 내용 자체가 문서에 없어 해당 없음. 반대로 문서4:390–479의 판정 6종(구현 결과 보고 / 독립 감사 verdict / gate admission / 전달 실행 / 교차 레인 통합 / 설계 정본)과 권한 한계 표는 상위 §6:756의 4필드보다 세밀하고, RC1 `AGENTS.md` Workflow 5–7(작성자 ≠ 감사자 ≠ 통합 리더)이 그대로 쓰는데 상위 본문에는 없다.
- 상위가 빠뜨린 교훈: (a) 티어만으로 main push/merge 권한을 넓히지 않는다(문서4:352·661) — 상위 §5:582는 "작은 수정은 최소 SPEC"만. (b) `&& … || echo SKIP`, bare `wait`, commit 뒤 abort를 "no changes"로 표현하는 반례(문서4:376) — 상위 없음; RC1 `.github/workflows/rc-ci.yml:47–78`의 `run` 래퍼가 exit를 모아 해결. (c) EXPECTED_RED와 모든 assertion failure 즉시 중단 문구의 충돌(문서4:382) — 상위 §5:584 "진짜 RED"만. (d) 감사자는 판정만 반환하고 허용된 publisher가 저장(문서4:368) — 상위 §15:1620과 부분 대응. (e) 순정 기능 "부분 대체" 판정표(문서4:515–629) — 상위 §4는 지원 경계만 있고 대체 판정이 없다. (f) Sync as-built rewrite는 승인 정체성 변경(문서4:342) — 상위 §5:602에 있음(일치).

### 4.3 내부 결함
- `factory_wait.go:29`, `card_wait.go:121`, `factory_wait_test.go:34`(문서4:14–18)와 `codex-review-gate.sh:183`(문서4:109)의 줄 번호가 어느 commit 기준인지 없다 — 이동 좌표(레인 보고 인용).
- head가 네 개(RC38 b1806498fe 문서4:22, 후보 2fd9455965 문서4:159, AF HEAD 0afd4c4e7e 문서4:643, main 5c3f0ebf 문서4:177)인데 시점표가 없다.
- 조사 범위 수치(72/28/26/61/50/20, 문서4:698–754)의 출처 파일 6종(문서4:760–772)이 링크·해시 없이 이름만 있다 — 검증 불가.
- 사용자 음성 전사(문서4:129·151)를 승인 근거로 인용 — Sentinel 메시지 ID는 있으나 외부에서 확인할 수 없다.
- 문서4:145 "Codex 보류"와 문서4:167 "Codex 감사 도구 금지 아님"은 서로 일관(결함 아님).

### 4.4 RC1 결정·실물과의 관계
- 재채택: not_run→passed 승격 금지를 기계 검사로(문서4:273·665 → `schemas/protocol/v1/status-crosswalk.json`, `docs/protocol-v1.md:88` 00D-AC-3); 정렬 canonical digest와 알고리즘 기록(문서4:328 → `src/spec.rs:549–587` `SortedKeys`·`content_digest`, 미정 부분은 `@MX:DEBT`); 감사 base/head 불일치 거부(문서4:655 → `src/audit.rs:237–247` `PrBinding::Stale`); 검증 명령은 argv·shell 없음(문서4:376 반례 → `src/verify.rs:3·71–73`); 종료 순서(문서4:97·153 → decisions:90).
- 역행: "Claude Code 중심 3.2 후속, Codex 보류"(문서4:145·165) → RC1은 Codex adapter를 범위에 넣었다(02-C #1993 merge e324e0c72, decisions:74; `project.yaml` hosts claude·codex). launcher만 Claude 우선(decisions:26).
- 역행: 같은 owner가 수리→검증→push→review→통합 확인까지 책임(문서4:58) → RC1 lane은 PR 뒤 다음 카드로 이동(decisions:77).

### 4.5 상위 §20 행 개선 제안
- 비계승 열을 "당시 Go 지침 본문의 결함 목록(Phase 13 참조 등 문서4:185)과 'AF rc.39 선행 → Claude 중심 → Codex 보류' 순서(문서4:145)"로 바꾸고 현재 문구는 삭제; 계승 열에 "판정 6종 분리표(문서4:390)"를 추가.

### 4.6 미확인·한계
- 세부 보고서 7종(문서4:760–772)과 하위 페이지 7개 미열람. Sentinel 메시지 원문 미확인. Go 소스 줄 번호를 실제 commit에 대조하지 않음.

---

## 5. MoAI ADK 코어와 제공자 및 브라우저 공식 계약 조사 (`page_bb3cfe5d…`, sha256 ff8b348f…)

### 5.1 문서 개요
- 2026-11-01 04:16~05:10 UTC 조사 4건 보관(문서5:5·11·211·512·750). 기준 HEAD 5c3f0ebf(문서5:11, Go 저장소 존재 확인).
- ① core/provider 경계: 단계적 하이브리드(내부 adapter → 번들 계약 → 검증된 외부 plugin) 권고(문서5:17·135–147). ② Claude/Codex plugin·skill·command·agent·hook·MCP·권한·배포·업데이트 차이표(문서5:235–404)와 bootstrap 상태기계(문서5:444). ③ ego(lite)/Aside capability 행렬 9항목(문서5:552–721). ④ desktop/cloud provider 계약(문서5:746–899).
- 설치·실행·변경 전부 NOT_RUN 명시(문서5:211·476·512·899).

### 5.2 상위 §20 표와의 대조
- 계승 "DOM/AX/screenshot/trace/cancel별 capability, 배포 SHA 결합"(상위:2322): 행렬에는 탭·DOM·accessibility·screenshot·console·network/HAR·trace·viewport·session/profile이 있고(문서5:573–721) browser 행렬에 "cancel" 행은 없다(문서 5에서 cancel은 host adapter의 lifecycle 매핑·시험 항목으로만 등장, 문서5:155·197; browser capability로서의 cancel은 문서3:929 항목 16). "배포 SHA 결합"은 문서3:968–969(항목 17 evidence envelope·deployment identity)의 것이고 문서 5의 QA 계약(문서5:730–734)에는 없다. 계승 열의 절반이 문서 3 귀속.
- 비계승 "문서 지원을 실기 성공으로 표시, 특정 ADE를 필수 제품화"(상위:2326): 문서5:512 "'공식 확인'은 제품 문서상 지원이며 현 환경 실행 성공을 뜻하지 않는다", 문서5:523 documented/runtime_verified/unverified/unsupported 분리, 문서5:525 "screenshot 수집 또는 시각 QA 완료로 승격하지 않는다" — 비계승 행이 문서가 명시적으로 막은 결함을 귀속. "특정 ADE 필수 제품화"도 문서5:523 "두 제품 모두 MoAI가 실제 연결·검증한 어댑터는 아니다"와 반대이며, 그에 가까운 표현은 문서3:75("Egolite·Aside 공식 지원은 본 설계의 제품 목표").
- 비계승으로 적었지만 상위가 필요로 하는 것: 상위 §4:275 "unknown→documented→probed→certified"는 문서5:523 4분류를 재명명한 것으로, 비계승이 아니라 계승된 실체다.
- 상위가 빠뜨린 교훈: (a) hooks를 유일한 fail-closed 정책으로 두지 말 것, Codex MCP hook은 오류 때 block하지 않음(문서5:466–468, O5) — 상위 `fail-open|hook` 0건(§14:1566은 Pi 한정). (b) 설치됨 ≠ 신뢰됨 ≠ 로그인됨 ≠ 활성화됨(문서5:217)과 상태기계 CORE_INSTALLED→…→VERIFIED_READY(문서5:444) — 상위 §3:297 한 줄. (c) Claude commands/agents를 Codex에 그대로 복사하지 않고 `$moai-*` skill entry로 변환(문서5:298–302·319) — 상위 §4:432는 team 파일만. (d) Claude cloud/WSL plugin 미지원, desktop MCP precedence가 CLI와 다름(문서5:767–771) — 상위 §14 Desktop에 없음. (e) 양 host가 같은 파일을 동시에 편집하는 문제는 plugin 호환성과 별개(문서5:470) — 상위 §13 동시성에 없음. (f) `--dangerously-bypass-hook-trust`·`--yes`를 bootstrap 기본값에 쓰지 않음(문서5:424·438) — 상위 없음; RC1은 launcher 기본에 `--dangerously-skip-permissions`를 넣었다가(decisions:47) 철회(decisions:80·86)하는 왕복을 겪었다.

### 5.3 내부 결함
- 로컬 Codex 0.166.0-alpha.1(문서5:223)로 문법만 확인 — 문서 스스로 stable 일반화 금지(문서5:476)를 적었고, RC1 관측 0.163.0(`docs/host-support-matrix.md:78`)과 다르다(환경·시점 차이).
- C/O 출처 URL 14건이 모두 revision 미고정 "동적 공식 문서"(문서5:482–506·905–916) — 재현성 한계(문서3:1352 목록이 SHA256 미확인으로 인정).
- 조사 4건의 시각이 04:16·04:17·05:10으로 다른데 머리글(문서5:5)에는 단일 날짜만 있다.
- ego 상세 API는 구 캐시(references/api.md)만 있고 main raw URL은 404(문서5:740) — 미확인으로 정직 처리, 결함 아님.
- Go 코드 관측(문서5:28–78)은 permalink로 고정돼 있다 — 여섯 문서 중 출처 고정이 가장 좋다.

### 5.4 RC1 결정·실물과의 관계
- 재채택: Codex skill 경로 `.agents/skills`(문서5:281 → `docs/host-support-matrix.md:82`, `docs/rust-product-path.md:62`); commands→skill 변환·`$moai`(문서5:302 → `docs/user-entry-draft.md`, 상위 검토 B7); hooks 0개(P04=A, `docs/host-support-matrix.md:72`) — 문서5:468 "hooks 유일 gate 금지"와 같은 방향; documented/not_run 분리(문서5:523 → `docs/support-matrix.md` 머리글, `schemas/e2e-receipt.schema.json:34` result enum passed/failed/not_run/blocked, `docs/host-support-matrix.md:12` `documented` 정의, ZCode app/cli 두 줄 분리 decisions:61); hook trust·permission bypass 금지(문서5:438 → decisions:80).
- 미채택: plugin manifest/marketplace bootstrap(문서5:419–434) — RC1은 skill 폴더 직접 설치(`docs/rust-product-path.md:62`); browser adapter 전부 RC1 밖.
- 문서5:159 Forge adapters는 GitHub/GitLab 동시 서술 — RC1은 GitHub만(상위 검토 B1과 동일 지적).

### 5.5 상위 §20 행 개선 제안
- 계승 열을 "host별 plugin·skill·hook·MCP 계약 차이표(문서5:235)와 documented/runtime_verified/unverified/unsupported 4분류(문서5:523), browser capability 행렬(문서5:552)"로 바꾸고 cancel·배포 SHA는 문서 3 행으로; 비계승 열은 "Codex alpha CLI 문법과 2026-11-01 시점 URL을 현행 계약으로 재사용하지 않기"로 교체.

### 5.6 미확인·한계
- 공식 URL 14건 재방문 안 함. Codex/Claude 현재 버전의 plugin 명령 변동 미확인. 조사 메모 원본(provider-contracts.md·browser-contracts.md) 미열람.

---

## 6. MoAI ADK 팀 개발 워크플로와 전체 실행 로직 (`page_00788de2…`, sha256 da611198…)

### 6.1 문서 개요
- 날짜·기준 commit 없음(여섯 문서 중 유일). "사용자 최신 지시" 반영 이력만 있고, 10-27~11-01 원문은 하위 페이지 14개로 보관(문서6:2315–2329). 착수 등록 GitHub Issue #1948(문서6:16) — 소속 저장소 미표기.
- 내용: GitHub/GitLab Issues 업무 정본·SQLite 업무 큐 미도입(문서6:198), Coordinator/Agent/Control Service 역할(문서6:349–463), trusted config ref와 config_revision(문서6:471), 권한 R0–R3(문서6:625–685), own-only Issue 작성자 정책(문서6:693–759), 준비 상태 ACCEPTED→READY_ACKED(문서6:1021–1079), Issue 링크 장부(문서6:1199–1278), 담당 agent의 PR 후 종결 책임(문서6:1400), 업무 FSM E01–E27(문서6:1591–2153), 위임 상태 PREPARED→APPLIED(문서6:2174–2299), launch key·quiescence·recover-release(문서6:2307–2313).
- 수용시험 EFF01–06·AUTH01–08·PREP01–10·CONTEXT01–06·HANDOFF01–10 전부 NOT_RUN.

### 6.2 상위 §20 표와의 대조
- 계승 "과거 상세 계약·Pele 원문과 검토 기록 보존"(상위:2336): "Pele"는 문서6:483 한 번("실제 API나 Pele 코드가 아니다")뿐이고 도식 원문은 문서1:1429 소유 — 귀속 오류. 상세 계약의 실체(FSM E01–E27, 위임 상태, recover-release)를 계승 열이 지목하지 않는다. 상위 §6:760 handoff 문장과 §15:1611 차단 행이 문서6:2313·1402의 축약이다.
- 비계승 "당시 담당자·lane·DB·제공자 제한을 새 제품 전체에 적용"(상위:2340): "DB 제한"은 문서6:198 "SQLite 업무 큐를 도입하지 않는다"로 상위 §6:613·§13:1391과 같은 방향 — 비계승이 아니라 계승. "제공자 제한"(GLM 제외 문서6:11·202, @codex 조건부)은 RC1이 뒤집었다(`src/session.rs:24` HOSTS claude/glm/codex, decisions:47) — 비계승 판정이 맞다. "담당자·lane"은 RC1이 agent-1~n·`-a`로 그대로 채택(문서6:1406 → `src/session.rs:23`, decisions:47·76) — 비계승 대상이 계승됨.
- 비계승으로 적었지만 상위가 필요로 하는 것: 상위 §6:792 "PR 이후 수정 요청의 owner를 남긴다"는 문서6:1400(담당 agent가 종결까지)과 문서6:1400이 "대체했다"고 한 이전 계약(PR 인계 후 리더로 이전) 가운데 어느 쪽인지 정하지 않았다. RC1은 이전 계약 쪽이다(`AGENTS.local.md:32·34` "A lane never waits for the merge", 리더 큐가 수리·병합 소유). 상위가 이 결정을 명시해야 한다.
- 상위가 빠뜨린 교훈: (a) config_revision = trusted commit OID + 각 blob OID + canonical semantic digest, blob SHA 하나로 복수 파일 원자 스냅샷을 주장하지 않음(문서6:471) — 상위 §6 schema는 `policy_digest` 한 필드. (b) Issue canonical contract의 해시 포함/제외 필드와 canonicalization 규칙(문서6:913–915) — 상위 §6 schema에 없음(§8은 전문 미독). (c) leader 비차단 백그라운드 접수·준비 큐·READY ACK 전 launch 금지(문서6:996–1081) — 상위 없음. (d) 링크 장부: `.moai/report` vs `.moai/reports` 추정 금지, commit 고정 URL + 내용 hash, LINK_BLOCKED(문서6:1205–1278) — 상위 §12 보존에 없음. (e) own-only 작성자 정책, bot author, action별 권한 분리(문서6:693–759) — 상위 §6 권한 행은 "최소 권한, credential 분리"만. (f) end-to-end 효율 측정 계약, baseline 없으면 비교 불가 표시(문서6:36–100) — 상위 §16–17은 검색·성능 benchmark만. (g) auditor는 trusted audit bundle만 로드하고 candidate hooks/CLAUDE.md/MCP 자동 실행을 차단, `--bare` 강제 금지(문서6:1562) — 상위 §15:1656 prompt injection 행은 "자료 속 명령은 권한을 넓히지 못한다"까지.

### 6.3 내부 결함
- 참조 끊김: M01~M56 단일 manifest(문서6:312), CLI01–23(문서6:1556), "이전 ZIP 1.0.0"(문서6:1556)이 본문 사본에 없다.
- 문서6:913 canonical block은 댓글·상태·assignee를 제외해 안정 해시를 노리는데, 문서6:1276은 자주 바뀌는 링크 장부 색인을 "canonical context block"에 둔다 — 같은 블록이면 해시가 흔들리고 다른 블록이면 이름 충돌; 설명 없음.
- 문서6:202 "감사는 Claude 또는 Codex 중 하나만" 과 문서6:11 "3종 감사 선택안 대체"는 일관하지만, RC1 실제(Claude agent-2 독립 감사 + Codex review P2, decisions:74)와 다르다.
- `browser_selection.default_when_unset: "chrome"`(문서6:564)을 제품 설정으로 두지만 RC1·상위 어디에도 대응 항목이 없다(F22 ADE는 후속).
- Issue #1948(문서6:16)의 저장소 불명 — moai-adk RC1 번호대(1973~)보다 앞서 Go 저장소 Issue로 추정, 미확인.

### 6.4 RC1 결정·실물과의 관계
- 재채택: agent-1~n·`-a` 명명(문서6:1406 → `src/session.rs:23`, decisions:47·76); Issues 정본·SQLite 없음(문서6:198 → `AGENTS.md` Authority 2, `Cargo.toml`); Closes 키워드의 간접 종료를 효과로 계산(문서6:757 → `docs/pr-contract.md` "자동 종료 키워드는 쓰지 않는다"); PR ≠ DONE(문서6:208 → `docs/pr-contract.md` "PR 개설은 완료가 아니다", `AGENTS.local.md:32`); 증거 없는 force 해제 불채택(문서6:2313 → `src/claim.rs:22`, `src/handoff.rs:6–7` "expiry and takeover are always named, never automatic"); writer와 auditor 세션 분리(문서6:1562 → `AGENTS.local.md:24` agent-2 read-only); durable intent→send→ACK 순서(문서6:285·2182–2221 → `src/forge/outbox.rs:4–13` intent→response→observed); 예산·lease 수치는 실측 뒤(문서6:606 → decisions:94 lease HOLD); 개인 경로 금지(문서6:1205 → `src/spec.rs:126`).
- 역행: GLM 제외(문서6:202) → `moai glm` launcher(`src/session.rs:10·24`); "담당 agent가 PR 후 종결까지"(문서6:1400) → lane은 PR 후 다음 카드(decisions:77); "@codex는 설정된 경우만"(문서6:202) → Codex review 사용(decisions:74), 설정 여부 기록은 decisions에 없음 — 근거 없음.
- 근거 없음: leader 백그라운드 접수(문서6:1000), own-only 작성자 정책(문서6:693) — decisions·AGENTS.local.md·tracked 파일 grep에서 대응 없음(RC1 Issue는 전부 리더 발행이라 미발생).

### 6.5 상위 §20 행 개선 제안
- 계승 열을 "Issues 정본·FSM E01–E27·위임 상태 PREPARED→APPLIED·recover-release(문서6:1568–2313)"로 구체화하고 Pele는 문서 1 행으로; 비계승 열은 "GLM 제외·@codex 조건부(문서6:202)와 '담당 agent가 PR 후 종결까지'(문서6:1400) — RC1은 레인 이동 + 리더 통합 큐로 결정(decisions:77)"로 바꿔 결정을 드러낼 것.

### 6.6 미확인·한계
- 하위 페이지 14개(문서6:2315–2329) 미읽음. Issue #1948 소속 저장소 미확인. M01~M56·CLI01~23 원문 없음. "이전 ZIP 1.0.0" 미열람.

---

## 7. 문서 교차 — 두 문서 이상이 공유하지만 상위 본문에 없는 교훈

### 7.1 상위에 이미 있는 것(공로 인정)
- lease 만료 ≠ 소유권 이전 ≠ writer 종료: 상위:760 ← 문서1:456, 문서2:1135, 문서3:899, 문서6:697.
- intent → durable → 전송 → ACK → identity/digest → gate 채택 분리: 상위:1620 ← 문서2:546, 문서6:2182–2221.
- documented → probed → certified 단계: 상위:275 ← 문서2:937, 문서3:818, 문서5:523, 문서6:28.
- PASS-WITH-DEBT 비평탄화: 상위:1647 ← 문서2:10, 문서4:318.
- exit 0·ACK·자기 보고 ≠ 완료: 상위:1678 ← 문서2:799, 문서3:261, 문서6:2301.
- read-only 조회 무변경: 상위:1593 ← 문서2:532.
- 멱등 키·readback·불명 상태에서 중복 생성 금지: 상위:790 ← 문서1:792, 문서2:1255, 문서3:267, 문서6:1006.

### 7.2 상위 §1–§18에 없는 교훈(grep + §3–§6·§12–§13·§15–§18 전문 기준; §7–§11·§14·§19는 grep만)
1. 보고서·진행 기록 전용 커밋이 head를 바꿔 검증을 stale로 만드는 순환 — 증거는 freeze 전에 포함하거나 HEAD 밖에 둔다(문서1:20·1057; 문서2:169·559·1278). 상위:2266이 "남긴다"고 적었으나 본문 0건. RC1은 decisions:56(본문 사본 기준 변경)과 `AGENTS.local.md:39`(update-branch 뒤 delta 감사)로 유사한 stale 순환을 운영 중에 만났다.
2. 종료 순서 ExitWorktree(keep) → main 확인 → /clear → 다음 카드, dirty·미push·활성 검사 보존(문서1:208; 문서4:97·153) — 상위 0건. RC1 decisions:90·`AGENTS.local.md:33` 채택.
3. PR 이후 책임 귀속 — 레인 이동 + 리더 통합(문서1:208) vs 담당자가 종결까지(문서4:58, 문서6:1400) — 상위 §6:792는 미결정. RC1은 전자(decisions:77). 상위가 결정을 한 줄로 적어야 다음 구현자가 둘 중 하나를 새로 만들지 않는다.
4. 리더 제어 루프는 긴 작업을 안에서 기다리지 않고, 대기 이유·다음 trigger를 기록한다(문서1:798–830; 문서2:515·797·1106; 문서6:1000–1004) — 상위는 §15:1611 차단 행만 있고 리더 루프 자체가 없다. RC1 `AGENTS.local.md:35` lane-check routine이 재발명.
5. 필수 검사·CI trigger는 저장소 보호 설정과 현 head·event에서 읽고, 경로가 없으면 CI_TRIGGER_UNREACHABLE로 표시(문서1:898; 문서2:142) — 상위 0건. RC1 `docs/pr-contract.md`가 "기본 브랜치에서만 템플릿·Closes 해석"으로 체득.
6. hooks·skill은 보안 경계가 아니며 host hook은 fail-open일 수 있다(문서3:202·433–447; 문서5:466–468) — 상위 §14:1566은 Pi 한정. RC1 P04=A hooks 0개와 정합.
7. 성능·생산성 수치는 같은 조건 baseline 뒤에만, 측정 전 절감률 발표 금지(문서1:1422; 문서2:674·1474; 문서3:784; 문서6:98–100) — 상위 §16:1759는 비용·속도에 한정, 워크플로 처리량·토큰·재개 성공률 지표(문서1:1334–1420, 문서6:38–96)가 없다.
8. 자동 설치·quarantine 제거·trust 우회를 bootstrap 기본값에 넣지 않는다(문서3:938; 문서5:438·538) — 상위 §3:37 "전체 묶음 기본 설치 안 함"만. RC1 decisions:47→80의 skip-permissions 왕복이 이 공백의 비용.
9. 개인 절대 경로·비밀을 보고서·증거 작성 시점에 제외(문서3:33; 문서6:1205·1274; 문서2:221은 위반 사례) — 상위 §15:1665 공개 전송 행은 전송 시점만. RC1 `src/spec.rs:126`·decisions:57.
10. 외부 SHA·버전·시각을 한 표로 고정(문서3:69·483; 문서4:22·159·177·643; 문서5:11) — 문서들이 각자 경고하지만 상위 §22는 Go 4fa2aa9 하나(상위 검토 A8과 겹침).

### 7.3 §20 표의 구조적 결함(공통)
- 비계승 열이 문서 1·4·5에서 문서가 명시적으로 배제한 결함을 문서 귀속으로 적었다. 각 비계승 행은 "문서가 실제로 제안했으나 채택하지 않은 선택지 + 문서 줄"로 고쳐 쓰는 편이 맞다.
- 계승·비계승 항목의 교차 귀속: 문서5↔문서3(cancel·배포 SHA), 문서6↔문서1(Pele), 문서2↔문서1(고정 반복 간격), 문서3↔문서4(Claude 우선 Codex 후속).
- 비계승으로 적은 lane·lease·epoch·agent-N·Issue별 PR은 RC1이 다시 썼다. 표가 "역사"로 보낸 것과 RC1 실물이 어긋나므로 "RC1에서 재채택됨"을 행마다 표시하는 열을 추가할 가치가 있다.

## 8. 미확인·한계(전체)
- 상위 §7–§11·§14·§19·§21–§22는 전문을 읽지 않았다(§20과 grep만). 따라서 7.2의 "없음"은 §8 온톨로지·§14 Pi 절에 다른 표현으로 있을 가능성을 배제하지 못한다.
- 도식(Pele SVG·Mermaid)은 여섯 문서와 상위 모두 텍스트 사본에 없다.
- 하위 페이지(문서1:1, 문서3:11, 문서6:14)와 외부 URL은 열람하지 않았다.
- `origin/v2026.11.0`은 fetch하지 않은 로컬 remote-tracking tip 기준이다. Go 저장소 SHA는 존재만 확인했고 내용은 대조하지 않았다.
- grep 부재가 본문 부재는 아니다(동의어 가능). "근거 없음" 판정은 decisions·AGENTS.local.md·tracked 파일 grep 범위에 한정한다.
- 문서 3·4의 Go 소스 줄 번호 인용(`factory_wait.go:29` 등)은 실제 commit에 대조하지 않았다.
