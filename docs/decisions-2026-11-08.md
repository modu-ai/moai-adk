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

부분 결정 분리: P08은 claim 방식(⑤+①)만 결정, lease·heartbeat 수치 HOLD · P15는 ZCode 헤드리스 범위만 결정, 나머지 host 지원표 HOLD.

P12에서 아직 HOLD인 것: OS/arch matrix, locale, release matrix. P04(첫 RC 인증 순서), P05(package 경계·bridge 범위 나머지), P06, P08~P11, P14, P15는 HOLD 유지.
