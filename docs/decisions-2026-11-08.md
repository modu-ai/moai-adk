# 사용자 결정 기록 (2026-11-08)

리더 세션의 `AskUserQuestion` 응답으로 확정된 결정. 각 항목은 해당 Issue 본문의 "사용자 결정" 절과 `.moai-cli/manifests/github-issues.json`의 `user_decisions`에 같은 내용으로 기록된다. 설계 문서(P04~P15)의 미결 중 여기 적힌 범위만 결정되었고, 나머지는 HOLD다.

| 시각(KST) | 주제 | 결정 | 적용 Issue | 비고 |
| --- | --- | --- | --- | --- |
| 15:5x | 04-A 배포·설치 채널 | A: GitHub Releases + `dist`(cargo-dist) installer(shell/PowerShell). 일반 사용자에게 Rust 요구 없음 | #1998, #1980, #1999, #2015 | P12 중 배포 채널 부분 |
| 15:5x | 04-A artifact 서명·검증 | A: GitHub artifact attestation만(Sigstore Public Good, SLSA Build L2, `gh attestation verify`) | #1998, #2015 | P12 중 서명 부분 |
| 15:5x | 04-A release tag/feed namespace | B: 같은 repo `modu-ai/moai-adk`, 접두사 `moai-cli-v<version>`. Go checker(`v`/`go-v`)에 노출되지 않음 | #1998, #2015, #1975 | 기존 tag ruleset `v*` 불변성은 이 접두사를 덮지 않음 → 08-D HOLD |
| 15:5x | 04-A binary 경로 | A: `~/.moai-cli/bin/moai-cli`(installer가 PATH 안내) | #1998, #1999 | — |
| 16:0x | 마이그레이션 요구 | 기존 Go 사용자가 `moai update`(Go)를 실행하면 moai-cli(Rust) 설치·상태 마이그레이션·Go 제거가 되어야 한다 | #1998, #1980, #1999, 04-D | 설계 반영만, 지금 실행 없음 |
| 16:1x | Windows 실행 중 binary 교체 | A: self-replace 패턴(rename → 새 파일 배치 → 정리 프로세스 삭제; self-replace crate 1.5.1) | #1998, #1999 | P12 중 Windows 교체 부분 |
| 16:1x | 마이그레이션 브리지 위치 | B: 마지막 Go 릴리스는 moai-cli installer 실행 + `moai-cli migrate --from-go` 호출만; 마이그레이션·백업·Go 제거 로직은 Rust | 04-D | P05 한시 bridge 중 이 부분 |
| 16:1x | Go 제거 시점 | B: 자가 검증(버전·디스커버리·상태 복사 검증) 통과 시 자동 제거. 백업(`~/.moai` → `~/.moai-cli/backups/go-<date>/`)은 보관, 롤백 가능 | 04-D | 보존 원칙과의 긴장은 04-D에 롤백·보관 기간 HOLD로 기록 |
| 16:1x | 마이그레이션 Issue 소속 | A: 신규 `04-D Go 사용자 마이그레이션(bridge)` Issue를 EP-05 하위로 발행(선행 04-A·04-B) | 04-D | 사용자 승인에 따른 1건 발행 |
| 16:3x | 통신 허브 전송 조합(RC1) | H-B 파일 큐(내구층) + H-A 로컬 브로커(실시간층); H-C MCP 도구층은 후속 | #1987 #1990 #1992 #1993 #1994 #2005 #2006 #2007 #2008 #2010 #2011 | 보고서 §7 |
| 16:3x | 허브 수명 | L1 리더 세션 결합 | 위와 동일 | 보고서 §8 |
| 16:3x | 승인 중계 | C 사전 정책 + 예외만 중계 | 위와 동일 | 보고서 §3·§11 |
| 16:3x | 런처 UX(가정) | `moai` / `moai -l` / `moai -a <host>` / `moai cc`(alias); 기본값 후속 | 위와 동일 | 문법 확정 아님 |
| 16:4x | 04-D Go 자동 제거 범위(P13 재확인) | A: Go binary + Go 전용 캐시·feed 상태만 제거, `~/.moai` 데이터 보존, 자격 증명 복사 안 함 | #2016 | P13과 정합 |
| 16:4x | P08 claim 규약 | ⑤ Issue 댓글 id fencing + ① 리더 로컬 journal 보조; takeover는 운영자 승인 | #2005, #2006 | lease 수치는 HOLD |
| 16:4x | ZCode·ACP | ZCode 헤드리스만, ACP 후속 | #1994 | — |
| 16:4x | 원격·Orca/Hide | RC1 제외(후속) | #1994 | — |
| 16:4x | **통신 허브 전체** | **RC1 보류 → 다음 버전**. 위 허브 4행(전송·수명·승인·UX)은 다음 버전 적용용 기록 | 11개 Issue | 사용자: "일단 허브 보류, 기존 계획대로, 1번째 버전 빠르게 구현" |
| 17:0x | 개발 세션 모델 | 주 모델 `opus` effort `medium`, 고난이도 추론은 advisor `fable`(공식 문서 `--advisor`/`advisorModel` 확인 2026-11-08). 추후 설정에서 선택 | AGENTS.md, AGENTS.local.md, `moai-cli session` | 레인·리더 공통 |
| 17:0x | 개발 런처 | `moai-cli session --role leader|agent --name … [--advisor fable] [--dry-run]`(Claude만; Codex 후속). 제품 UX `moai -l/-a`는 방향만 | src/session.rs, docs/rust-product-path.md | — |
| 17:0x | AGENTS.* 재설계 | 재귀적 자가 생성 구조로 본문 재작성, `moai-cli agents-plan/apply`로 관리 구간 재생성(state unchanged, plan digest sha256:a8ad47c0…) | AGENTS.md, AGENTS.local.md | 다음 세션부터 이 절차로 계속 개선 |
| 17:1x | 공유 interface 소유자(중복 12묶음) | 상류 Issue 소유: tree digest→00-A, GitHub 읽기 adapter→00-B, `.moai-cli` 하위 경로(P09)·agents.rs 리팩터→01-B, exit code→01-C, host 설정 병합·toml_edit→02-B, 지원표 문서→02-C, redaction 모듈→03-B, PR 템플릿·write_scope→06-C, outbox marker→07-B, `.github/workflows`→08-D | 중복 묶음을 적은 모든 Issue | 비소유 Issue는 참조만 |
| 17:1x | 첫 배차 | 00-A #1985→agent-4, 00-B #1986→agent-3 (worktree `issue/<n>-<key>`, PR→v2026.11.0) | #1985 #1986 | agent-2 r7 재검토 병행 |
| 17:3x | 00-A HOLD-00A-1 수집기 위치 | A: Rust `moai-cli baseline-inventory` | #1985 | agent-4 blocked 보고로 상정 |
| 17:3x | 00-A HOLD-00A-3 산출물 추적 | `.moai-cli/manifests/baseline/` tracked + 개인 경로·secret 검사 | #1985 | — |
| 17:3x | 00-A AC-3 재정의 | Go 런타임 경로(`.moai/logs/`·`state/`·`db/`) 제외 + fixture에서 moai-cli 직접 쓰기 0 | #1985 | 전체 `.moai/` 전후 동일은 Go 훅 때문에 성립 불가(실측) |
| 17:3x | 레인 소유 댓글 게시 경로 | worktree 가드가 `gh`·Write를 막는 동안 **리더가 레인 보고를 대신 게시**(게시문에 "레인 agent-N 보고 중계" 명시). 가드는 변경하지 않음 | 모든 레인 | 독립성 약화는 알려진 trade-off |
| 16:5x | 레인 worktree 위치 | ~~리더 결정: `.claude/worktrees/`~~ → agent-4 이의(`.claude/` 보존 경계·00-A 측정 오염)로 철회. **사용자 결정 16:5x→17:0x: primary 안 `.moai-cli/worktrees/<issue>-<key>`를 Claude·Codex 공통 기본 위치로**, `.gitignore`에 `/.moai-cli/worktrees/` 추가(리더; 중첩 worktree는 `.git` 파일 때문에 ignore 없이는 untracked로 표시됨). Go 훅은 변경하지 않음. agent-3(`.claude/worktrees/issue-1986-00-B`)·agent-4(형제 디렉터리) 모두 이동 | 모든 레인 | Go `moai` 훅은 primary 밖 Edit/Write 거부(바이너리 내부 검사) |
| 17:1x | 00-B HOLD-00B-3 수집기 위치 | RC1 임시: `tests/` Python 유지; Rust 이전은 07-A #2008 후속 | #1986 | agent-2 조건부 FAIL 해소 조건 |
| 17:1x | 00-B HOLD-00B-2 API 버전 pin | RC1 임시: gh 기본(헤더 미전송) 유지; pin은 07-A에서 | #1986 | fixture 재기록 없음 |
| 17:1x | 00-A 범위(리더 확정, 소유자 결정의 귀결) | tree digest 소유 Issue로서 공용 `tree_digest(root, excludes)` 함수를 00-A PR에 포함(00A-AC-8); 05-C는 호출만 | #1985 #2003 | agent-2 r10 "주인 없는 산출물" 해소 |
| 17:1x | 인수 산출물 재배치 | workflows(rc-ci·rc-e2e)→08-D, PR 템플릿·pr-contract→06-C(REQ-06C-13), outbox.rs→07-B, generation 훅→01-B | #2015 #2007 #2009 #1989 | r11 |
| 17:3x | 통합 정책(에이전틱) | **모든 통합을 에이전틱 자율로**: ruleset `RC1 integration gate (v2026.11.0)`(id 26791615; PR 필수, 필수 status `moai/audit`, force-push·삭제 금지, bypass 없음). 리더가 독립 감사 PASS를 읽고 head SHA에 `moai/audit=success`를 올린 뒤 `gh pr merge --auto --merge`; head가 바뀌면 status 무효 → 재감사. 레인·봇은 merge 불가. 리더 문서 커밋도 PR 경로 | 모든 PR | 첫 적용 PR #2019 |
| 17:3x | Issue 원장 거울(`github-issues.json`·본문 사본) | 07-A/07-B 구현 때까지 그대로 사용(제품급 대체는 그 Issue에서) | #2008 #2009 | 리더 설명 뒤 결정 |
| 17:3x | 00-A AC-3 제외 범위 | "Go 훅 append 전용 `*.jsonl`" 규칙으로 제외(`.moai/logs/`·`state/`·`db/` + `.moai/evolution/telemetry/*.jsonl`·`harness/usage-log.jsonl`·`lessons-inbox.jsonl` 등); 제외 목록을 산출물에 기록; 검증 명령은 원장 3파일 기준 | #1985 | PR #2018 재감사 뒤 통합 |
| 17:4x | 00-E cargo 캐시 전략(리더 RC1 임시) | B 매번 `cargo fetch --locked` 뒤 `--offline`(의존성 4개, 액션 추가 없음); A actions/cache는 08-D 재평가. 소유 변경: rc-ci.yml 1차·PR 템플릿 1차 → 00-E #2020; 08-D는 release·rc-e2e·필수 체크 등록, 06-C는 렌더러·pr-contract | #2020 #2015 #2007 | agent-2 설계 검토 반영 |
| 17:3x | 통합 게이트 보강(agent-2 지적, 리더 적용) | `strict_required_status_checks_policy: true`(base 최신 아니면 merge 불가 → `gh pr update-branch` 뒤 재감사). verdict 첫 줄 형식 고정: `AUDIT #<n> (<key>) head <40hex> base <40hex> — verdict: PASS\|FAIL`(수식어 금지), 리더는 그 줄만으로 판정하고 head==PR head 확인. 알려진 빈틈: 모든 세션이 같은 GitHub 계정이라 status 출처(integration_id) 고정 불가 — 감사 전용 GitHub App은 후속(HOLD); status description/target_url에 감사 댓글 링크 기록으로 사후 대조 | 모든 PR | — |
| 17:3x | Go 세대 `.github` 제거·재구성 | 사용자 지시: untracked Go 세대 `.github` 23파일 제거(00-C AC-5가 moai-adk-go 템플릿과 23/23 동일 확인; sha256 기록은 리더 scratchpad), RC1 전용으로 재구성은 신규 Issue로 배차(workflows·PR 템플릿·labeler 선행 구현; 08-D·06-C는 그 위에 확장) | 신규 Issue, #2015 #2007 | 소유자 표 조정 |
| 17:3x | AGENTS 역할 분리 | `AGENTS.md` = moai-adk 기본 워크플로·규칙(제품 템플릿 성격), `AGENTS.local.md` = moai-adk 로컬 개발 규칙(역할·lane loop·통합 정책·보존 경계·구현 지도·검사 명령) | AGENTS.md, AGENTS.local.md | 재생성은 agents-plan/apply |

부분 결정 분리: P08은 claim 방식(⑤+①)만 결정, lease·heartbeat 수치 HOLD · P15는 ZCode 헤드리스 범위만 결정, 나머지 host 지원표 HOLD.

P12에서 아직 HOLD인 것: OS/arch matrix, locale, release matrix. P04(첫 RC 인증 순서), P05(package 경계·bridge 범위 나머지), P06, P08~P11, P14, P15는 HOLD 유지.
