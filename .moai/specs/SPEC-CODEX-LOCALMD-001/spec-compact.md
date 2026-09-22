# spec-compact.md — SPEC-CODEX-LOCALMD-001 (derived digest)

## 요구 모듈 (5)

1. **듀얼 적재·합성** (REQ-LMD-001, 003): `CLAUDE.local.md` + `AGENTS.local.md`를 CLAUDE first 순서와 provenance 헤더로 연결한 뒤 단일 `developer_instructions` config override로 전달한다.
2. **안전 판독·불변** (REQ-LMD-002, 010): 양쪽 파일 모두 symlink-follow 없는 safe open → same-descriptor fstat/read를 사용한다. 사용자 로컬 입력 두 파일은 모든 단계에서 불변이며 런처/LIVE는 `AGENTS.md`/`CLAUDE.md`도 수정하지 않는다. sync의 추적 `AGENTS.md` 설명 갱신은 별도 문서 변경이다.
3. **부재/빈 매트릭스** (REQ-LMD-004): 빈 파일은 부재와 같고, 비어 있지 않은 파일만 합성한다. 둘 다 없거나 비면 argv를 변경하지 않는다.
4. **충돌·실행 한도 방어** (REQ-LMD-005, 007): Codex 0.155.1이 수용하는 다섯 config 표기(`--config value`, `--config=value`, `-c value`, `-c=value`, `-cvalue`)가 `developer_instructions`를 중복 지정하면 거부한다. direct 최종 토큰과 spawn 최종 shell-quoted 명령을 각각 측정해 독립적으로 126,976-byte 기본 상한을 적용한다.
5. **균일성·무절단·신선 판독·표면 정합** (REQ-LMD-006, 008, 009): 각 body slice는 provenance-aware 추출 후 원본 bytes/hash와 일치한다. 헬프는 CLAUDE 공통 입력과 AGENTS Codex-specific 입력을 구분하며 status 행은 추가하지 않는다.

## 수용 기준 (12)

AC-LMD-001 (3 verb 동일) · 002 (spawn/-w/factory 동일) · 003 (순서·provenance) · 004 (absent/empty 8셀) · 005 (플랫폼별 safe open/fstat 및 경합) · 006 (추출 body slice hash) · 007 (parser 5형태 충돌) · 008 (direct/spawn 독립 상한) · 009 (로컬 입력 불변과 sync 문서 구분) · 010 (재판독) · 011 (공통/전용 help 구분) · 012 (120초 exact-command LIVE, response+log 양 nonce/source predicate).

## 수정 대상 파일

**Run phase (최소 in-scope)**:

- `internal/cli/codex_launcher.go` — 2-파일 합성, 5형태 충돌 검사, direct/spawn 최종 표현 guard, 헬프 카피
- `internal/cli/codex_contract.go` — `CLAUDE.local.md` 이름 상수; exactly-3 table 불변
- `internal/config/defaults.go` — 126,976-byte 기본 상한의 단일 원천
- `internal/cli/`의 플랫폼별 safe-open helper — 양쪽 파일의 no-follow open + same-descriptor fstat/read 또는 fail-closed 대체
- `internal/cli/codex_local_instructions_test.go` 및 필요한 플랫폼 fixture 테스트 — 코드 레벨 AC, parser 5형태, 경합, direct/spawn 경계와 quote 팽창
- `internal/template/templates/AGENTS.md.tmpl` — 특정 로컬 파일명을 재열거하지 않는 일반 문구 정확화

**Sync phase (manager-docs)**:

- `CHANGELOG.md`, `docs-site/content/{en,ko,ja,zh}/advanced/codex-dual-harness.md`, 리포 `AGENTS.md` §8

## Exclusions

- `moai codex status` 행·출력 계약 변경
- Factory broker / idle-wake (t1074/t1075 소관)
- `AGENTS.md`/`CLAUDE.md`의 run-phase 수정, symlink/import chain 생성
- `codexInstructionRelPathsFn` exactly-3 table 확장
- Codex CLI 외부 동작 변경
- sync-phase 문서의 run-phase 선실행
